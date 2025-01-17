use crate::{
    get_schema_for_path_mut, maybe_any_of,
    negative_normalize::{NegativeNormalize, DECIMAL_ZERO},
    promote_missing, schema_difference, schema_for_bson, schema_for_type_str, DeriveSchema, Result,
    ResultSetState,
};
use agg_ast::definitions::{
    Expression, MatchBinaryOp, MatchExpr, MatchExpression, MatchField, MatchLogical,
    MatchNotExpression, MatchStage, Ref, UntaggedOperator,
};
use bson::Bson;
use mongosql::{
    map,
    schema::{
        Atomic, Document, Satisfaction, Schema, FALSIFIABLE_TYPES, INTEGER_LONG_OR_NULLISH,
        NULLISH, NUMERIC, NUMERIC_OR_NULLISH, UNFOLDED_ANY,
    },
    set,
};
use std::collections::BTreeSet;

fn is_unitary_schema(schema: &Schema) -> bool {
    matches!(
        schema,
        Schema::Atomic(Atomic::Null)
            | Schema::Missing
            | Schema::Atomic(Atomic::MinKey)
            | Schema::Atomic(Atomic::MaxKey)
            | Schema::Atomic(Atomic::Undefined)
    )
}

macro_rules! bits_schema {
    () => {
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::BinData),
        ])
    };
}

macro_rules! geo_schema {
    () => {
        Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                    "type".to_string() => Schema::Atomic(Atomic::String),
                    "coordinates".to_string() => Schema::Array(Box::new(NUMERIC.clone())),
                },
                required: set!["coordinates".to_string()],
                additional_properties: false,
                jaccard_index: None,
            }),
            Schema::Array(Box::new(NUMERIC.clone())),
        ])
    };
}

// is_exists_true_bson returns true if the BSON value is considered to truthy in exist in a match
// operation.
fn is_exists_true_bson(bson: &Bson) -> bool {
    match bson {
        Bson::Null => false,
        Bson::Boolean(b) => *b,
        Bson::Int32(i) => *i != 0,
        Bson::Int64(i) => *i != 0,
        Bson::Double(d) => *d != 0.0,
        Bson::Decimal128(d) => *d != *DECIMAL_ZERO,
        _ => true,
    }
}

// schema_for_match_bson_literal generates the proper schema for a Bson object used in a $match.
// This is a bit complicated in that mongodb will allow a 1 integer, for instance, to match all
// numeric types but something like 1.5 can only match double and decimal128, but a 1.0 Decimal128
// or Double can match an integer 1.
fn schema_for_match_bson_literal(bson: &Bson, include_missing: bool) -> Schema {
    // helper to handle doubles
    let schema_for_match_double = |d: f64| {
        if d.fract() == 0.0 {
            NUMERIC.clone()
        } else {
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            ])
        }
    };
    match bson {
        Bson::Int32(_) | Bson::Int64(_) => NUMERIC.clone(),
        Bson::Double(d) => schema_for_match_double(*d),
        Bson::Decimal128(d) => {
            let d = d.to_string().parse::<f64>();
            match d {
                Ok(d) => schema_for_match_double(d),
                // If this can't be parsed into a double, we can clearly only match Decimal128
                Err(_) => Schema::Atomic(Atomic::Decimal),
            }
        }
        // Missing is considered equivalent to Null in $match
        Bson::Null => {
            if include_missing {
                Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing])
            } else {
                Schema::Atomic(Atomic::Null)
            }
        }
        b => schema_for_bson(b),
    }
}

// schema_for_type_expr generates a schema for a type expression used in a $type match operation.
fn schema_for_type_expr(bson: &Bson) -> Schema {
    match bson {
        Bson::String(s) => schema_for_type_str(s),
        Bson::Array(values) => {
            let schemas = values
                .iter()
                .map(|v| match v {
                    Bson::String(s) => schema_for_type_str(s),
                    _ => unreachable!(),
                })
                .collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        // this should never happen because this should result in a statically unacceptable query
        _ => Schema::Any,
    }
}

// schema_for_array_bson_values generates a schema for an array of BSON values used in a match
// operation. This currently only happens in the cases of $in, $nin, and $all. While $type takes
// an array, it is array of type name strings and must be handled separately.
fn schema_for_array_bson_values(bson: &Bson, include_missing: bool) -> Schema {
    match bson {
        Bson::Array(values) => {
            let schemas = values
                .iter()
                .map(|x| schema_for_match_bson_literal(x, include_missing))
                .collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        // this should never happen because this should result in a statically unacceptable query
        _ => Schema::Any,
    }
}

// result_set_schema_difference wraps schema_difference to operate on field refs and variables within the
// result set state. This is specifically for applying constraints via match.
fn result_set_schema_difference(
    reference: &agg_ast::definitions::Ref,
    state: &mut ResultSetState,
    to_remove: BTreeSet<Schema>,
) {
    let ref_schema: Option<&mut Schema> = match reference {
        agg_ast::definitions::Ref::FieldRef(reference) => {
            let path = reference
                .as_str()
                .split('.')
                .map(|s| s.to_string())
                .collect();
            get_schema_for_path_mut(&mut state.result_set_schema, path)
        }
        agg_ast::definitions::Ref::VariableRef(v) => state.variables.get_mut(v),
    };
    if let Some(schema) = ref_schema {
        schema_difference(schema, to_remove);
    }
}

impl DeriveSchema for MatchStage {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        state.result_set_schema = promote_missing(&state.result_set_schema);
        for expr in self.expr.iter() {
            let expr = expr.get_negative_normal_form();
            expr.match_derive_schema(state);
        }
        Ok(state.result_set_schema.clone())
    }
}

trait MatchConstrainSchema {
    // match_derive_schema does not need to return Schema because it only applies constraints
    // to already existing Schema. It also does not need to return a Result because it is infallible
    // (modulo a panic that can only result due to programmer error).
    fn match_derive_schema(&self, state: &mut ResultSetState);
}

impl MatchConstrainSchema for MatchExpression {
    fn match_derive_schema(&self, state: &mut ResultSetState) {
        match self {
            MatchExpression::Expr(e) => e.match_derive_schema(state),
            MatchExpression::Misc(_) => todo!(),
            MatchExpression::Logical(l) => l.match_derive_schema(state),
            MatchExpression::Field(f) => f.match_derive_schema(state),
        }
    }
}

impl MatchConstrainSchema for MatchLogical {
    fn match_derive_schema(&self, state: &mut ResultSetState) {
        match self {
            MatchLogical::And(exprs) => {
                for expr in exprs.iter() {
                    expr.match_derive_schema(state);
                }
            }
            MatchLogical::Or(exprs) => {
                let mut states = Vec::new();
                for expr in exprs.iter() {
                    let mut state = state.clone();
                    expr.match_derive_schema(&mut state);
                    states.push(state);
                }
                let mut schema = Schema::Unsat;
                for state in states.into_iter() {
                    schema = schema.union(&state.result_set_schema);
                }
                state.result_set_schema = schema;
            }
            MatchLogical::Nor(_) => {
                panic!(
                    "found $nor in match_derive_schema, this means negative normalization did not occur"
                )
            }
            MatchLogical::Not(n) => {
                // The only operator left with $not after negative normalization that can inform
                // types is $type. We originally had the idea of just negating the type set for
                // $type, but there is no type name for missing, so that does not handle missing
                // correctly. All other operators left with $not currently do not have an affect on
                // schema. For instance geo ops will fail to match if the field is an incorrect geo
                // coordinate or if it is _any other type_.
                if let MatchNotExpression::Query(ref ops) = n.expr {
                    // ops must be length one after negative normalization
                    if ops.len() == 1 {
                        let (op, b) = ops.iter().next().unwrap();
                        if let MatchBinaryOp::Type = op {
                            let to_remove_schema = schema_for_type_expr(b);
                            match to_remove_schema {
                                Schema::AnyOf(schemas) => {
                                    result_set_schema_difference(&n.field, state, schemas);
                                }
                                _ => result_set_schema_difference(
                                    &n.field,
                                    state,
                                    set![to_remove_schema],
                                ),
                            }
                        }
                    }
                }
            }
        }
    }
}

impl MatchConstrainSchema for MatchField {
    fn match_derive_schema(&self, state: &mut ResultSetState) {
        self.ops.iter().for_each(|(op, b)| {
            match_derive_schema_for_op(&self.field, *op, b, state);
        });
    }
}

// this function simply applies the intersection to a field if it exists, otherwise does nothing..
// this is useful for applying the constraints of certain match operations to the overall result set schema
fn intersect_if_exists(reference: &Ref, state: &mut ResultSetState, input_schema: Schema) {
    let ref_schema: Option<&mut Schema> = match reference {
        Ref::FieldRef(reference) => {
            let path = reference
                .as_str()
                .split('.')
                .map(|s| s.to_string())
                .collect();
            get_schema_for_path_mut(&mut state.result_set_schema, path)
        }
        Ref::VariableRef(v) => state.variables.get_mut(v),
    };
    if let Some(schema) = ref_schema {
        *schema = schema.intersection(&input_schema);
    }
}

// match_derive_schema_for_op derives the schema for a single match operation, since a given field
// may be nested above multiple match operations. This is a helper function for MatchField::match_derive_schema.
fn match_derive_schema_for_op(
    reference: &Ref,
    op: MatchBinaryOp,
    b: &Bson,
    state: &mut ResultSetState,
) {
    match op {
        MatchBinaryOp::Eq | MatchBinaryOp::Gte | MatchBinaryOp::Lte => {
            let schema = schema_for_match_bson_literal(b, true);
            intersect_if_exists(reference, state, schema);
        }
        MatchBinaryOp::Gt | MatchBinaryOp::Lt => {
            let schema = schema_for_match_bson_literal(b, false);
            intersect_if_exists(reference, state, schema);
        }
        // Ne actually does not tell us anything about Schema except for types that are
        // unitary-valued: Null, MinKey, MaxKey, Undefined.
        MatchBinaryOp::Ne => {
            let schema = schema_for_match_bson_literal(b, true);
            if is_unitary_schema(&schema) {
                result_set_schema_difference(reference, state, set![schema]);
            } else if let Schema::AnyOf(schemas) = schema {
                let to_remove = schemas.into_iter().filter(is_unitary_schema).collect();
                result_set_schema_difference(reference, state, to_remove);
            }
        }
        MatchBinaryOp::Exists => {
            if is_exists_true_bson(b) {
                result_set_schema_difference(reference, state, set![Schema::Missing]);
            } else {
                intersect_if_exists(reference, state, Schema::Missing);
            }
        }
        MatchBinaryOp::Type => {
            let schema = schema_for_type_expr(b);
            intersect_if_exists(reference, state, schema);
        }
        MatchBinaryOp::In => {
            let schema = schema_for_array_bson_values(b, true);
            intersect_if_exists(reference, state, schema);
        }
        MatchBinaryOp::Nin => {
            let schema = schema_for_array_bson_values(b, true);
            if is_unitary_schema(&schema) {
                result_set_schema_difference(reference, state, set![schema]);
            } else if let Schema::AnyOf(schemas) = schema {
                let to_remove = schemas
                    .iter()
                    .cloned()
                    .flat_map(|schema| {
                        if let Schema::AnyOf(schemas) = schema {
                            schemas
                        } else {
                            set! {schema}
                        }
                    })
                    .filter(is_unitary_schema)
                    .collect();
                result_set_schema_difference(reference, state, to_remove);
            }
        }
        MatchBinaryOp::Mod => {
            intersect_if_exists(reference, state, NUMERIC.clone());
        }
        MatchBinaryOp::Size => {
            let schema = Schema::Array(Box::new(Schema::Any));
            intersect_if_exists(reference, state, schema);
        }
        MatchBinaryOp::All => {
            let schema = Schema::Array(Box::new(schema_for_array_bson_values(b, false)));
            intersect_if_exists(reference, state, schema);
        }
        MatchBinaryOp::BitsAllClear
        | MatchBinaryOp::BitsAnyClear
        | MatchBinaryOp::BitsAllSet
        | MatchBinaryOp::BitsAnySet => {
            intersect_if_exists(reference, state, bits_schema!());
        }
        MatchBinaryOp::Near
        | MatchBinaryOp::GeoWithin
        | MatchBinaryOp::NearSphere
        | MatchBinaryOp::GeoIntersects => {
            intersect_if_exists(reference, state, geo_schema!());
        }
    }
}

impl MatchConstrainSchema for Expression {
    fn match_derive_schema(&self, state: &mut ResultSetState) {
        // unwrap_or_return is used to simply move on if derive_schema returns an error. Becuase the schema
        // derivation currently does not verify the correctness of pipelines, this results in us simply _not_
        // narrowing the result set schema for a $match, limiting precision but not correctness.
        macro_rules! unwrap_or_return {
            ( $e:expr ) => {
                match $e {
                    Ok(x) => x,
                    Err(_) => return,
                }
            };
        }

        fn match_derive_and(u: &UntaggedOperator, state: &mut ResultSetState) {
            let mut initial_schema = state.result_set_schema.clone();
            loop {
                u.args.iter().for_each(|arg| {
                    arg.match_derive_schema(state);
                });
                if initial_schema == state.result_set_schema {
                    break;
                }
                initial_schema = state.result_set_schema.clone();
            }
        }

        fn match_derive_or(u: &UntaggedOperator, state: &mut ResultSetState) {
            let schema: Option<Schema> = u.args.iter().fold(None, |schema, arg| {
                // because the conditions of $or are not additive, we need to create a fresh copy of the incoming result set schema for
                // each. This avoids us applying the constraints of one operand to another.
                let mut tmp_state = state.clone();
                tmp_state.null_behavior = Satisfaction::Not;
                arg.match_derive_schema(&mut tmp_state);
                match schema {
                    None => Some(tmp_state.result_set_schema),
                    Some(schema) => Some(schema.union(&tmp_state.result_set_schema)),
                }
            });
            if let Some(schema) = schema {
                state.result_set_schema = schema;
            }
        }

        fn match_derive_eq(u: &UntaggedOperator, state: &mut ResultSetState) {
            if u.args.len() == 2 {
                let lhs_schema = unwrap_or_return!(u.args[0].derive_schema(state));
                let rhs_schema = unwrap_or_return!(u.args[1].derive_schema(state));
                let mut schema_intersection = lhs_schema.intersection(&rhs_schema);
                // this covers the fact that numerics are all comparable in equality (eg, an integer can equal a decimal)
                if schema_intersection.satisfies(&NUMERIC.clone()) != Satisfaction::Not {
                    schema_intersection = schema_intersection.union(&NUMERIC.clone());
                }
                // falsifiable types include numbers, nullish, and boolean. evaluate if we must, may, or cannot be any nullish
                // type to determine the set of values the operands can take on.
                state.null_behavior = schema_intersection.satisfies(&FALSIFIABLE_TYPES.clone());
                u.args.iter().for_each(|arg| match arg {
                    Expression::Ref(reference) => {
                        intersect_if_exists(reference, state, schema_intersection.clone());
                    }
                    _ => {
                        arg.match_derive_schema(state);
                    }
                });
            }
        }

        fn match_derive_ne(u: &UntaggedOperator, state: &mut ResultSetState) {
            // we can only constrain the schema for ne with unitary types, for example ne null
            if u.args.len() == 2 {
                let lhs_schema = unwrap_or_return!(u.args[0].derive_schema(state));
                let rhs_schema = unwrap_or_return!(u.args[1].derive_schema(state));
                match (u.args[0].clone(), u.args[1].clone(), lhs_schema, rhs_schema) {
                    (
                        Expression::Ref(left_ref),
                        Expression::Ref(right_ref),
                        Schema::Atomic(left_atomic),
                        Schema::Atomic(right_atomic),
                    ) => {
                        // if we have an ne that is strictly unsatisfiable, set the schemas for these fields to be unsat
                        if left_atomic == right_atomic
                            && is_unitary_schema(&Schema::Atomic(left_atomic))
                        {
                            intersect_if_exists(&left_ref, state, Schema::Unsat);
                            intersect_if_exists(&right_ref, state, Schema::Unsat);
                        }
                    }
                    (Expression::Ref(reference), _, _, Schema::Atomic(a))
                    | (_, Expression::Ref(reference), Schema::Atomic(a), _) => {
                        if is_unitary_schema(&Schema::Atomic(a)) {
                            result_set_schema_difference(
                                &reference,
                                state,
                                set!(Schema::Atomic(a)),
                            );
                        }
                    }
                    _ => {}
                }
            }
        }

        // this function is a helper for the comparison functions $lt, $lte, $gt, $gte. It is invoked
        // when we have a field reference to constrain that fields schemas. It assumes the field ref is the
        // lhs of the operation, and the input_schema is the schema of the rhs. That is, for a field "foo", with
        // operator $gt, and input schema bar, we are constraining the types of foo according to $foo > bar.
        fn constrain_schema_for_comparison_reference(
            reference: &Ref,
            op: UntaggedOperatorName,
            state: &mut ResultSetState,
            input_schema: Schema,
        ) {
            let ref_schema: Option<&mut Schema> = match reference {
                Ref::FieldRef(reference) => {
                    let path = reference
                        .as_str()
                        .split('.')
                        .map(|s| s.to_string())
                        .collect();
                    get_schema_for_path_mut(&mut state.result_set_schema, path)
                }
                Ref::VariableRef(v) => state.variables.get_mut(v),
            };
            // the limit is the maximum schema ($lt, $lte) or minimum schema ($gt, $gte) that the
            // given field reference can take on, given the schema we are comparing it to.
            let limit = match input_schema.clone() {
                Schema::Any => match op {
                    UntaggedOperatorName::Lt | UntaggedOperatorName::Lte => {
                        Schema::Atomic(Atomic::MaxKey)
                    }
                    _ => Schema::Atomic(Atomic::MinKey),
                },
                Schema::AnyOf(a) => match op {
                    UntaggedOperatorName::Lt | UntaggedOperatorName::Lte => a
                        .iter()
                        .max()
                        .unwrap_or(&Schema::Atomic(Atomic::MaxKey))
                        .clone(),
                    _ => a
                        .iter()
                        .min()
                        .unwrap_or(&Schema::Atomic(Atomic::MinKey))
                        .clone(),
                },
                _ => input_schema.clone(),
            };
            if let Some(schema) = ref_schema {
                if schema == &mut Schema::Any {
                    *schema = UNFOLDED_ANY.clone();
                }
                // use the limit to constrain which types the field reference can take on
                let mut updated_schema = match schema.clone() {
                    Schema::AnyOf(a) => Schema::AnyOf(
                        a.into_iter()
                            .filter(|x| match op {
                                UntaggedOperatorName::Lt => {
                                    x < &limit || (x == &limit && !is_unitary_schema(x))
                                }
                                UntaggedOperatorName::Lte => x <= &limit,
                                UntaggedOperatorName::Gt => {
                                    x > &limit || (x == &limit && !is_unitary_schema(x))
                                }
                                UntaggedOperatorName::Gte => x >= &limit,
                                _ => true,
                            })
                            .collect(),
                    ),
                    s => {
                        if (s < limit
                            && (op == UntaggedOperatorName::Lt || op == UntaggedOperatorName::Lte))
                            || (s > limit
                                && (op == UntaggedOperatorName::Gt
                                    || op == UntaggedOperatorName::Gte))
                            || (s == limit
                                && (op == UntaggedOperatorName::Lte
                                    || op == UntaggedOperatorName::Gte
                                    || !is_unitary_schema(&s)))
                        {
                            s.clone()
                        } else {
                            Schema::Unsat
                        }
                    }
                };
                // because all numerics are comparable, we must include them even if they are greater than the max type of the rhs
                if updated_schema.intersection(&NUMERIC.clone()) != Schema::Unsat {
                    updated_schema = updated_schema.union(&NUMERIC.clone().intersection(schema));
                }
                *schema = updated_schema;
            }
        }

        fn match_derive_comparison(u: &UntaggedOperator, state: &mut ResultSetState) {
            if u.args.len() == 2 {
                let lhs_schema = unwrap_or_return!(u.args[0].derive_schema(state));
                let rhs_schema = unwrap_or_return!(u.args[1].derive_schema(state));
                if let Expression::Ref(reference) = &u.args[0] {
                    constrain_schema_for_comparison_reference(reference, u.op, state, rhs_schema);
                } else {
                    u.args[0].match_derive_schema(state);
                }
                if let Expression::Ref(reference) = &u.args[1] {
                    // we invert the operator, so that we can treat this field reference as the LHS of the comparison (in order to reuse the helper);
                    // for example, if we want to constrain bar in the comparison foo < $bar, we can treat it as $bar > foo.
                    let op = match u.op {
                        UntaggedOperatorName::Lt => UntaggedOperatorName::Gt,
                        UntaggedOperatorName::Lte => UntaggedOperatorName::Gte,
                        UntaggedOperatorName::Gt => UntaggedOperatorName::Lt,
                        UntaggedOperatorName::Gte => UntaggedOperatorName::Lte,
                        _ => return,
                    };
                    constrain_schema_for_comparison_reference(reference, op, state, lhs_schema);
                } else {
                    u.args[1].match_derive_schema(state);
                }
            }
        }

        fn match_derive_numeric(u: &UntaggedOperator, state: &mut ResultSetState) {
            u.args.iter().for_each(|arg| {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(reference, state, NUMERIC.clone());
                        }
                        Satisfaction::May => {
                            intersect_if_exists(reference, state, NUMERIC_OR_NULLISH.clone());
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(reference, state, NULLISH.clone());
                        }
                    };
                }
            });
        }

        fn match_derive_bit_ops(u: &UntaggedOperator, state: &mut ResultSetState) {
            u.args.iter().for_each(|arg| {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(
                                reference,
                                state,
                                Schema::AnyOf(set!(
                                    Schema::Atomic(Atomic::Integer),
                                    Schema::Atomic(Atomic::Long),
                                )),
                            );
                        }
                        Satisfaction::May => {
                            intersect_if_exists(reference, state, INTEGER_LONG_OR_NULLISH.clone());
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(reference, state, NULLISH.clone());
                        }
                    }
                }
            });
        }

        fn match_derive_is_number(u: &UntaggedOperator, state: &mut ResultSetState) {
            if let Expression::Ref(reference) = u.args[0].clone() {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(&reference, state, NUMERIC.clone());
                    }
                    Satisfaction::Must => {
                        result_set_schema_difference(
                            &reference,
                            state,
                            set!(
                                Schema::Atomic(Atomic::Decimal),
                                Schema::Atomic(Atomic::Double),
                                Schema::Atomic(Atomic::Integer),
                                Schema::Atomic(Atomic::Long),
                            ),
                        );
                    }
                    _ => {}
                };
            }
        }

        fn match_derive_range(u: &UntaggedOperator, state: &mut ResultSetState) {
            u.args.iter().for_each(|arg| {
                if let Expression::Ref(reference) = arg {
                    intersect_if_exists(reference, state, NUMERIC.clone());
                }
            });
        }

        fn match_derive_round(u: &UntaggedOperator, state: &mut ResultSetState) {
            if let Expression::Ref(reference) = u.args[0].clone() {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(&reference, state, NUMERIC.clone());
                    }
                    Satisfaction::May => {
                        intersect_if_exists(&reference, state, NUMERIC_OR_NULLISH.clone());
                    }
                    Satisfaction::Must => {
                        intersect_if_exists(&reference, state, NULLISH.clone());
                    }
                };
            }
            if u.args.len() > 1 {
                if let Expression::Ref(reference) = u.args[1].clone() {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(
                                &reference,
                                state,
                                Schema::AnyOf(set!(
                                    Schema::Atomic(Atomic::Integer),
                                    Schema::Atomic(Atomic::Long)
                                )),
                            );
                        }
                        Satisfaction::May => {
                            intersect_if_exists(&reference, state, INTEGER_LONG_OR_NULLISH.clone());
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(&reference, state, NULLISH.clone());
                        }
                    };
                }
            }
        }

        fn match_derive_numeric_conversion(u: &UntaggedOperator, state: &mut ResultSetState) {
            if let Expression::Ref(reference) = u.args[0].clone() {
                let numeric_convertible = Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::String),
                ));
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(&reference, state, numeric_convertible);
                    }
                    Satisfaction::May => {
                        intersect_if_exists(
                            &reference,
                            state,
                            numeric_convertible.union(&NULLISH.clone()),
                        );
                    }
                    Satisfaction::Must => {
                        intersect_if_exists(&reference, state, NULLISH.clone());
                    }
                };
            }
        }

        use agg_ast::definitions::UntaggedOperatorName;
        let null_behavior = state.null_behavior;
        match self {
            Expression::TaggedOperator(_t) => todo!(),
            Expression::UntaggedOperator(u) => match u.op {
                // logical ops
                UntaggedOperatorName::And => match_derive_and(u, state),
                UntaggedOperatorName::Or => match_derive_or(u, state),
                // comparison ops
                UntaggedOperatorName::Eq => match_derive_eq(u, state),
                UntaggedOperatorName::Cmp | UntaggedOperatorName::Ne => match_derive_ne(u, state),
                UntaggedOperatorName::Gt
                | UntaggedOperatorName::Gte
                | UntaggedOperatorName::Lt
                | UntaggedOperatorName::Lte => match_derive_comparison(u, state),
                // numeric ops
                UntaggedOperatorName::Abs
                | UntaggedOperatorName::Acos
                | UntaggedOperatorName::Acosh
                | UntaggedOperatorName::Asin
                | UntaggedOperatorName::Asinh
                | UntaggedOperatorName::Atan
                | UntaggedOperatorName::Atan2
                | UntaggedOperatorName::Atanh
                | UntaggedOperatorName::Cos
                | UntaggedOperatorName::Cosh
                | UntaggedOperatorName::DegreesToRadians
                | UntaggedOperatorName::Divide
                | UntaggedOperatorName::Exp
                | UntaggedOperatorName::Ln
                | UntaggedOperatorName::Log
                | UntaggedOperatorName::Log10
                | UntaggedOperatorName::Mod
                | UntaggedOperatorName::Multiply
                | UntaggedOperatorName::Pow
                | UntaggedOperatorName::RadiansToDegrees
                | UntaggedOperatorName::Sin
                | UntaggedOperatorName::Sinh
                | UntaggedOperatorName::Sqrt
                | UntaggedOperatorName::Tan
                | UntaggedOperatorName::Tanh
                | UntaggedOperatorName::Trunc
                | UntaggedOperatorName::Ceil
                | UntaggedOperatorName::Floor => match_derive_numeric(u, state),
                UntaggedOperatorName::BitAnd
                | UntaggedOperatorName::BitNot
                | UntaggedOperatorName::BitOr
                | UntaggedOperatorName::BitXor => match_derive_bit_ops(u, state),
                UntaggedOperatorName::IsNumber => match_derive_is_number(u, state),
                UntaggedOperatorName::Range => match_derive_range(u, state),
                UntaggedOperatorName::Round => match_derive_round(u, state),
                UntaggedOperatorName::ToInt
                | UntaggedOperatorName::ToDouble
                | UntaggedOperatorName::ToDecimal
                | UntaggedOperatorName::ToLong => match_derive_numeric_conversion(u, state),
                _ => todo!(),
            },
            _ => {}
        }
        state.null_behavior = null_behavior;
    }
}

impl MatchConstrainSchema for MatchExpr {
    fn match_derive_schema(&self, state: &mut ResultSetState) {
        self.expr.match_derive_schema(state);
    }
}
