use crate::{
    get_schema_for_path_mut,
    negative_normalize::{NegativeNormalize, DECIMAL_ZERO},
    schema_for_bson, schema_for_type_str, DeriveSchema, Result, ResultSetState,
};
use agg_ast::definitions::{
    Expression, MatchBinaryOp, MatchExpr, MatchExpression, MatchField, MatchLogical,
    MatchNotExpression, MatchStage, Ref, UntaggedOperator,
};
use bson::Bson;
use mongosql::{
    map,
    schema::{
        Atomic, Document, Satisfaction, Schema, FALSIFIABLE_TYPES, NULLISH, NUMERIC,
        NUMERIC_OR_NULLISH,
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

macro_rules! maybe_any_of {
    ($schemas:expr) => {
        if $schemas.len() == 1 {
            $schemas.into_iter().next().unwrap()
        } else {
            Schema::AnyOf($schemas)
        }
    };
}

macro_rules! any_expand {
    () => {
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String),
            Schema::Document(Document::any()),
            Schema::Array(Box::new(Schema::Any)),
            Schema::Atomic(Atomic::BinData),
            Schema::Atomic(Atomic::Undefined),
            Schema::Atomic(Atomic::ObjectId),
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Date),
            Schema::Missing,
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Regex),
            Schema::Atomic(Atomic::DbPointer),
            Schema::Atomic(Atomic::Javascript),
            Schema::Atomic(Atomic::Symbol),
            Schema::Atomic(Atomic::JavascriptWithScope),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::MinKey),
            Schema::Atomic(Atomic::MaxKey),
        ])
    };
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

// promote_missing adds Schema::Missing to any non-required key in a Document schema. This is
// necessary for our operations to work correctly, since they operate on paths, leaving no way
// to check or modify required. We will rely on Schema::simplify to lower Schema::Missing back to
// removing from required, since Schema::Missing cannot be serialized.
fn promote_missing(schema: &Schema) -> Schema {
    // It would be much more efficient to do this in place, but we can't do that because of
    // BTreeSets. At some point we may want to consider moving to Vec, which would have no
    // effect on serialization, but we would have to be more careful to remove duplicates and
    // define an ordering.
    match schema {
        Schema::AnyOf(schemas) => {
            let schemas = schemas.iter().map(promote_missing).collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        Schema::Array(schema) => Schema::Array(Box::new(promote_missing(schema))),
        Schema::Document(doc) => {
            let mut doc = doc.clone();
            for (key, schema) in doc.keys.iter_mut() {
                if !doc.required.contains(key) {
                    *schema = schema.union(&Schema::Missing);
                    doc.required.insert(key.clone());
                }
            }
            Schema::Document(doc)
        }
        _ => schema.clone(),
    }
}

// this function gets the maximum type of a given schema, which is mostly meaningful for AnyOfs. This allows us to handle comparison
// operators that use total ordering, such as $lte in $expr.
fn max_type_for_schema(schema: Schema) -> Schema {
    match schema {
        Schema::Unsat
        | Schema::Array(_)
        | Schema::Atomic(_)
        | Schema::Document(_)
        | Schema::Missing
        | Schema::Any => schema,
        Schema::AnyOf(a) => a.iter().fold(Schema::Unsat, |local_max, anyof_element| {
            if &local_max < anyof_element {
                anyof_element.clone()
            } else {
                local_max
            }
        }),
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
                            let path = n.field.as_str().split('.').map(|s| s.to_string()).collect();
                            let to_remove_schema = schema_for_type_expr(b);
                            let field_schema =
                                get_schema_for_path_mut(&mut state.result_set_schema, path);
                            let field_schema = match field_schema {
                                Some(field_schema) => field_schema,
                                None => {
                                    return;
                                }
                            };
                            match to_remove_schema {
                                Schema::AnyOf(schemas) => {
                                    schema_difference(field_schema, schemas);
                                }
                                _ => schema_difference(field_schema, set![to_remove_schema]),
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
        let path: Vec<_> = self
            .field
            .as_str()
            .split('.')
            .map(|s| s.to_string())
            .collect();
        self.ops.iter().for_each(|(op, b)| {
            match_derive_schema_for_op(path.clone(), *op, b, state);
        });
    }
}

// schema_difference removes a set of Schema from another Schema. This differs from
// Schema::intersection in that it does not use two Schemas as operands. Part of this is that
// schema_difference only ever happens with Atomic Schemas (and Missing, which is rather isomorphic
// to Atomic) and this is expedient. If we ever need to expand this to more complex Schemas, it may
// make sense to make this a real operator on two Schemas in the schema module.
//
// Note that this could also be achieved by complementing the Schema to be removed and intersecting
// it with the Schema to be modified, but this would be quite a bit less efficient.
fn schema_difference(schema: &mut Schema, to_remove: BTreeSet<Schema>) {
    match schema {
        Schema::Any => {
            *schema = any_expand!();
            schema_difference(schema, to_remove);
        }
        Schema::AnyOf(schemas) => {
            let any_of_schemas = schemas
                .difference(&to_remove)
                .cloned()
                .collect::<BTreeSet<_>>();
            *schema = maybe_any_of!(any_of_schemas);
        }
        _ => (),
    }
}

// this function simply applies the intersection to a field if it exists, otherwise does nothing..
// this is useful for applying the constraints of certain match operations to the overall result set schema
fn intersect_if_exists(path: Vec<String>, state: &mut ResultSetState, schema: Schema) {
    if let Some(field_schema) = get_schema_for_path_mut(&mut state.result_set_schema, path) {
        *field_schema = schema.intersection(field_schema);
    }
}

// match_derive_schema_for_op derives the schema for a single match operation, since a given field
// may be nested above multiple match operations. This is a helper function for MatchField::match_derive_schema.
fn match_derive_schema_for_op(
    path: Vec<String>,
    op: MatchBinaryOp,
    b: &Bson,
    state: &mut ResultSetState,
) {
    macro_rules! schema_difference {
        ($path:expr, $state:expr, $schemas:expr) => {
            let field_schema = get_schema_for_path_mut(&mut $state.result_set_schema, $path);
            match field_schema {
                Some(field_schema) => {
                    schema_difference(field_schema, $schemas);
                }
                None => (),
            }
        };
    }
    match op {
        MatchBinaryOp::Eq | MatchBinaryOp::Gte | MatchBinaryOp::Lte => {
            let schema = schema_for_match_bson_literal(b, true);
            intersect_if_exists(path, state, schema);
        }
        MatchBinaryOp::Gt | MatchBinaryOp::Lt => {
            let schema = schema_for_match_bson_literal(b, false);
            intersect_if_exists(path, state, schema);
        }
        // Ne actually does not tell us anything about Schema except for types that are
        // unitary-valued: Null, MinKey, MaxKey, Undefined.
        MatchBinaryOp::Ne => {
            let schema = schema_for_match_bson_literal(b, true);
            if is_unitary_schema(&schema) {
                schema_difference!(path, state, set![schema]);
            } else if let Schema::AnyOf(schemas) = schema {
                let to_remove = schemas.into_iter().filter(is_unitary_schema).collect();
                schema_difference!(path, state, to_remove);
            }
        }
        MatchBinaryOp::Exists => {
            if is_exists_true_bson(b) {
                schema_difference!(path, state, set![Schema::Missing]);
            } else {
                intersect_if_exists(path, state, Schema::Missing);
            }
        }
        MatchBinaryOp::Type => {
            let schema = schema_for_type_expr(b);
            intersect_if_exists(path, state, schema);
        }
        MatchBinaryOp::In => {
            let schema = schema_for_array_bson_values(b, true);
            intersect_if_exists(path, state, schema);
        }
        MatchBinaryOp::Nin => {
            let schema = schema_for_array_bson_values(b, true);
            if is_unitary_schema(&schema) {
                schema_difference!(path, state, set![schema]);
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
                schema_difference!(path, state, to_remove);
            }
        }
        MatchBinaryOp::Mod => {
            intersect_if_exists(path, state, NUMERIC.clone());
        }
        MatchBinaryOp::Size => {
            let schema = Schema::Array(Box::new(Schema::Any));
            intersect_if_exists(path, state, schema);
        }
        MatchBinaryOp::All => {
            let schema = Schema::Array(Box::new(schema_for_array_bson_values(b, false)));
            intersect_if_exists(path, state, schema);
        }
        MatchBinaryOp::BitsAllClear
        | MatchBinaryOp::BitsAnyClear
        | MatchBinaryOp::BitsAllSet
        | MatchBinaryOp::BitsAnySet => {
            intersect_if_exists(path, state, bits_schema!());
        }
        MatchBinaryOp::Near
        | MatchBinaryOp::GeoWithin
        | MatchBinaryOp::NearSphere
        | MatchBinaryOp::GeoIntersects => {
            intersect_if_exists(path, state, geo_schema!());
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
                // falsish types include numbers, nullish, and boolean. evaluate if we must, may, or cannot be any nullish
                // type to determine the set of values the operands can take on.
                state.null_behavior = schema_intersection.satisfies(&FALSIFIABLE_TYPES.clone());
                u.args.iter().for_each(|arg| match arg {
                    Expression::Ref(Ref::FieldRef(r)) => {
                        let path = r.as_str().split('.').map(|s| s.to_string()).collect();
                        intersect_if_exists(path, state, schema_intersection.clone());
                    }
                    _ => {
                        arg.match_derive_schema(state);
                    }
                });
            }
        }

        fn match_derive_lte(u: &UntaggedOperator, state: &mut ResultSetState) {
            if u.args.len() == 2 {
                if let Expression::Ref(Ref::FieldRef(r)) = &u.args[0] {
                    let rhs_schema = unwrap_or_return!(u.args[1].derive_schema(state));
                    let max_rhs_schema = max_type_for_schema(rhs_schema);
                    let path = r.as_str().split('.').map(|s| s.to_string()).collect();
                    if let Some(field_schema) =
                        get_schema_for_path_mut(&mut state.result_set_schema, path)
                    {
                        if field_schema == &mut Schema::Any {
                            *field_schema = any_expand!();
                        }
                        let mut updated_schema = match field_schema.clone() {
                            Schema::AnyOf(a) => Schema::AnyOf(
                                a.into_iter().filter(|x| x <= &max_rhs_schema).collect(),
                            ),
                            s => {
                                if s <= max_rhs_schema {
                                    s.clone()
                                } else {
                                    Schema::Unsat
                                }
                            }
                        };
                        // because all numerics are comparable, we must include them even if they are greater than the max type of the rhs
                        if updated_schema.intersection(&NUMERIC.clone()) != Schema::Unsat {
                            updated_schema =
                                updated_schema.union(&NUMERIC.clone().intersection(field_schema));
                        }
                        *field_schema = updated_schema;
                    }
                } else {
                    u.args[0].match_derive_schema(state);
                    u.args[1].match_derive_schema(state);
                }
            }
        }

        fn match_derive_numeric(u: &UntaggedOperator, state: &mut ResultSetState) {
            u.args.iter().for_each(|arg| {
                if let Expression::Ref(Ref::FieldRef(r)) = arg {
                    let path = r.as_str().split('.').map(|s| s.to_string()).collect();
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(path, state, NUMERIC.clone());
                        }
                        Satisfaction::May => {
                            intersect_if_exists(path, state, NUMERIC_OR_NULLISH.clone());
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(path, state, NULLISH.clone());
                        }
                    };
                }
            });
        }

        use agg_ast::definitions::UntaggedOperatorName;
        let null_behavior = state.null_behavior;
        match self {
            Expression::TaggedOperator(_t) => todo!(),
            Expression::UntaggedOperator(u) => match u.op {
                // logical ops
                UntaggedOperatorName::Or => match_derive_or(u, state),
                // comparison ops
                UntaggedOperatorName::Eq => match_derive_eq(u, state),
                UntaggedOperatorName::Lte => match_derive_lte(u, state),
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
