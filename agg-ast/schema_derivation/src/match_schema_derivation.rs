use crate::{
    array_element_schema_or_error, get_or_create_schema_for_path_mut, maybe_any_of,
    negative_normalize::{NegativeNormalize, DECIMAL_ZERO},
    promote_missing, schema_difference, schema_for_bson, schema_for_type_str, DeriveSchema, Error,
    Result, ResultSetState,
};
use agg_ast::definitions::{
    DateAdd, DateDiff, DateExpression, DateFromParts, DateFromString, DateSubtract, DateToParts,
    DateToString, DateTrunc, Expression, Filter, Let, Map, MatchBinaryOp, MatchExpr,
    MatchExpression, MatchField, MatchLogical, MatchMisc, MatchNotExpression, MatchStage, Median,
    NArrayOp, Reduce, Ref, RegexAggExpression, Replace, SortArray, Switch, TaggedOperator, Trim,
    UntaggedOperator, Zip,
};
use bson::Bson;
use mongosql::{
    json_schema::Schema as JsonSchema,
    schema::{
        Atomic, Document, Satisfaction, Schema, BITS_APPLICABLE, DATE_COERCIBLE,
        DATE_COERCIBLE_OR_NULLISH, GEO, INTEGER_LONG_OR_NULLISH, NULLISH, NUMERIC,
        NUMERIC_OR_NULLISH, STRING_OR_NULLISH, UNFOLDED_ANY,
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
            get_or_create_schema_for_path_mut(&mut state.result_set_schema, path)
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
            expr.match_derive_schema(state)?;
        }
        Ok(state.result_set_schema.clone())
    }
}

pub(crate) trait MatchConstrainSchema {
    // match_derive_schema does not need to return Schema because it only applies constraints
    // to already existing Schema. It also does not need to return a Result because it is infallible
    // (modulo a panic that can only result due to programmer error).
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()>;
}

impl MatchConstrainSchema for MatchExpression {
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()> {
        match self {
            MatchExpression::Expr(e) => e.match_derive_schema(state),
            MatchExpression::Misc(m) => m.match_derive_schema(state),
            MatchExpression::Logical(l) => l.match_derive_schema(state),
            MatchExpression::Field(f) => f.match_derive_schema(state),
        }
    }
}

impl MatchConstrainSchema for MatchLogical {
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()> {
        match self {
            MatchLogical::And(exprs) => {
                for expr in exprs.iter() {
                    expr.match_derive_schema(state)?;
                }
            }
            MatchLogical::Or(exprs) => {
                let mut states = Vec::new();
                for expr in exprs.iter() {
                    let mut state = state.clone();
                    expr.match_derive_schema(&mut state)?;
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
        Ok(())
    }
}

impl MatchConstrainSchema for MatchField {
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()> {
        self.ops.iter().for_each(|(op, b)| {
            match_derive_schema_for_op(&self.field, *op, b, state);
        });
        Ok(())
    }
}

impl MatchConstrainSchema for MatchMisc {
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()> {
        match self {
            MatchMisc::Regex(_) => Ok(()),
            // We could probably do something for MatchElement, but it's rather tedious, and doing
            // nothing is conservatively correct.
            MatchMisc::Element(_) => Ok(()),
            MatchMisc::Where(_) => Ok(()),
            MatchMisc::JsonSchema(s) => {
                let Bson::Document(ref d) = s.schema else {
                    return Err(Error::InvalidJsonSchema);
                };
                let schema: Schema = JsonSchema::from_document(d)
                    .map_err(|_| Error::InvalidJsonSchema)?
                    .try_into()
                    .map_err(|_| Error::InvalidJsonSchema)?;
                state.result_set_schema = state.result_set_schema.intersection(&schema);
                Ok(())
            }
            MatchMisc::Text(_) => Ok(()),
            MatchMisc::Comment(_) => Ok(()),
        }
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
            get_or_create_schema_for_path_mut(&mut state.result_set_schema, path)
        }
        Ref::VariableRef(v) => {
            // variables are treated as paths, so if we have a . separated path, we need to split it
            // and get the nested schema from the variable map
            if v.contains(".") {
                let path: Vec<String> = v.as_str().split('.').map(|s| s.to_string()).collect();
                state
                    .variables
                    .get_mut(&path[0])
                    .and_then(|doc| get_or_create_schema_for_path_mut(doc, path[1..].to_vec()))
            } else {
                state.variables.get_mut(v)
            }
        }
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
            intersect_if_exists(reference, state, BITS_APPLICABLE.clone());
        }
        MatchBinaryOp::Near
        | MatchBinaryOp::GeoWithin
        | MatchBinaryOp::NearSphere
        | MatchBinaryOp::GeoIntersects => {
            intersect_if_exists(reference, state, GEO.clone());
        }
    }
}

impl MatchConstrainSchema for Expression {
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()> {
        fn match_date_derive_common(
            d: &Expression,
            timezone: &Option<Box<Expression>>,
            state: &mut ResultSetState,
        ) -> Result<()> {
            let (date_schema, tz_schema) = match (state.null_behavior, timezone) {
                (Satisfaction::Not, _) => (DATE_COERCIBLE.clone(), Schema::Atomic(Atomic::String)),
                (Satisfaction::May, _) => {
                    (DATE_COERCIBLE_OR_NULLISH.clone(), STRING_OR_NULLISH.clone())
                }
                (Satisfaction::Must, Some(_)) => {
                    (DATE_COERCIBLE_OR_NULLISH.clone(), STRING_OR_NULLISH.clone())
                }
                (Satisfaction::Must, None) => (NULLISH.clone(), Schema::Any),
            };
            if let Expression::Ref(reference) = d {
                intersect_if_exists(reference, state, date_schema);
            } else {
                d.match_derive_schema(state)?;
            }
            if let Some(e) = timezone {
                if let Expression::Ref(reference) = e.as_ref() {
                    intersect_if_exists(reference, state, tz_schema);
                } else {
                    e.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_date_expression(
            d: &DateExpression,
            state: &mut ResultSetState,
        ) -> Result<()> {
            let DateExpression { date, timezone } = d;
            match_date_derive_common(date, timezone, state)
        }

        fn match_derive_date_to_parts(d: &DateToParts, state: &mut ResultSetState) -> Result<()> {
            let DateToParts { date, timezone, .. } = d;
            match_date_derive_common(date, timezone, state)
        }

        fn match_derive_date_from_parts(
            d: &DateFromParts,
            state: &mut ResultSetState,
        ) -> Result<()> {
            macro_rules! handle_date_field_schema {
                ( $e:expr, $schema:expr ) => {
                    if let Some(e) = $e {
                        if let Expression::Ref(reference) = e.as_ref() {
                            intersect_if_exists(reference, state, $schema);
                        } else {
                            e.match_derive_schema(state)?;
                        }
                    }
                };
            }
            let (most_part_schema, tz_schema) = match state.null_behavior {
                Satisfaction::Not => (NUMERIC.clone(), Schema::Atomic(Atomic::String)),
                Satisfaction::May | Satisfaction::Must => {
                    (NUMERIC_OR_NULLISH.clone(), STRING_OR_NULLISH.clone())
                }
            };
            // Since there are so many fields and all of them are optional (ish), we are not going
            // to do the MUST be NULLISH we see in other date operators because it's combimatorially
            // expensive for little gain in precision. It's very rare that we would be in a
            // situation where DateFromParts has exactly one field defined and it must be NULLISH.
            let DateFromParts {
                iso_week,
                iso_week_year,
                iso_day_of_week,
                year,
                month,
                day,
                hour,
                minute,
                second,
                millisecond,
                timezone,
            } = d;
            // Note that the reason we use macros here is that the clone will be lazy and only occur,
            // if the expression is Some and is a Reference.
            for e in [
                iso_week,
                iso_week_year,
                iso_day_of_week,
                year,
                month,
                day,
                hour,
                minute,
                second,
                millisecond,
            ]
            .into_iter()
            .flatten()
            {
                if let Expression::Ref(ref reference) = e.as_ref() {
                    intersect_if_exists(reference, state, most_part_schema.clone());
                } else {
                    e.match_derive_schema(state)?;
                }
            }
            handle_date_field_schema!(timezone, tz_schema);
            Ok(())
        }

        fn match_derive_date_from_string(
            d: &DateFromString,
            state: &mut ResultSetState,
        ) -> Result<()> {
            macro_rules! handle_date_field_schema {
                ( $e:expr, $schema:expr ) => {
                    if let Some(e) = $e {
                        if let Expression::Ref(reference) = e.as_ref() {
                            intersect_if_exists(reference, state, $schema);
                        } else {
                            e.match_derive_schema(state)?;
                        }
                    }
                };
            }
            let DateFromString {
                date_string,
                format,
                timezone,
                on_error,
                on_null,
            } = d;

            let string_schema = match (state.null_behavior, on_null) {
                (Satisfaction::Not, None) => Schema::Atomic(Atomic::String),
                (Satisfaction::Must, None)
                    if format.is_none() && timezone.is_none() && on_error.is_none() =>
                {
                    NULLISH.clone()
                }
                _ => STRING_OR_NULLISH.clone(),
            };
            if let Expression::Ref(reference) = date_string.as_ref() {
                intersect_if_exists(reference, state, string_schema.clone());
            } else {
                date_string.match_derive_schema(state)?;
            }
            handle_date_field_schema!(format, string_schema.clone());
            handle_date_field_schema!(timezone, string_schema);
            Ok(())
        }

        fn match_derive_date_to_string(d: &DateToString, state: &mut ResultSetState) -> Result<()> {
            macro_rules! handle_date_field_schema {
                ( $e:expr, $schema:expr ) => {
                    if let Some(e) = $e {
                        if let Expression::Ref(reference) = e.as_ref() {
                            intersect_if_exists(reference, state, $schema);
                        } else {
                            e.match_derive_schema(state)?;
                        }
                    }
                };
            }
            let DateToString {
                date,
                format,
                timezone,
                on_null,
            } = d;

            let (date_schema, string_schema) = match (state.null_behavior, on_null) {
                // if we have an onNull expression, we actually can't be specific on the
                // nullability of Schemas because it's possible for onNull to be NULL or the
                // other expressions to be NULL as long as they aren't NULL for the same document.
                (Satisfaction::Not, Some(_)) => {
                    // This may not be completely precise, but it is conservative.
                    (DATE_COERCIBLE_OR_NULLISH.clone(), STRING_OR_NULLISH.clone())
                }
                (Satisfaction::Not, None) => {
                    (DATE_COERCIBLE.clone(), Schema::Atomic(Atomic::String))
                }
                (Satisfaction::Must, None) if format.is_none() && timezone.is_none() => {
                    (NULLISH.clone(), /*will be unused*/ Schema::Any)
                }
                _ => (DATE_COERCIBLE_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()),
            };
            if let Expression::Ref(reference) = date.as_ref() {
                intersect_if_exists(reference, state, date_schema);
            } else {
                date.match_derive_schema(state)?;
            }
            handle_date_field_schema!(format, string_schema.clone());
            handle_date_field_schema!(timezone, string_schema);
            Ok(())
        }

        fn match_derive_string_operation(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            let schema = match state.null_behavior {
                Satisfaction::Not => Schema::Atomic(Atomic::String),
                Satisfaction::May => STRING_OR_NULLISH.clone(),
                Satisfaction::Must => {
                    if u.args.len() == 1 {
                        NULLISH.clone()
                    } else {
                        STRING_OR_NULLISH.clone()
                    }
                }
            };
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    intersect_if_exists(reference, state, schema.clone());
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_and(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            let mut initial_schema = state.result_set_schema.clone();
            loop {
                for arg in u.args.iter() {
                    arg.match_derive_schema(state)?;
                }
                if initial_schema == state.result_set_schema {
                    break;
                }
                initial_schema = state.result_set_schema.clone();
            }
            Ok(())
        }

        fn match_derive_or(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            // because the conditions of $or are not additive, we need to create a fresh copy of the incoming result set schema for
            // each. This avoids us applying the constraints of one operand to another.
            let mut schema: Option<Schema> = None;
            for arg in u.args.iter() {
                let mut tmp_state = state.clone();
                tmp_state.null_behavior = Satisfaction::Not;
                arg.match_derive_schema(&mut tmp_state)?;
                match schema {
                    None => schema = Some(tmp_state.result_set_schema),
                    Some(s) => schema = Some(s.union(&tmp_state.result_set_schema)),
                }
            }
            if let Some(schema) = schema {
                state.result_set_schema = schema;
            }
            Ok(())
        }

        fn match_derive_eq(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            let null_behavior = state.null_behavior;
            if u.args.len() == 2 {
                // we first check each argument's schema using Satisfaction::May to get the full schema these operators
                // can return. For example, two operators may only overlap if they both are null, in which case we'd use null_behavior=Must
                state.null_behavior = Satisfaction::May;
                let lhs_schema = u.args[0].derive_schema(state)?;
                let rhs_schema = u.args[1].derive_schema(state)?;
                let mut schema_intersection = lhs_schema.intersection(&rhs_schema);
                // this covers the fact that numerics are all comparable in equality (eg, an integer can equal a decimal)
                if schema_intersection.satisfies(&NUMERIC.clone()) != Satisfaction::Not {
                    schema_intersection = schema_intersection.union(&NUMERIC.clone());
                }
                state.null_behavior = schema_intersection.satisfies(&NULLISH.clone());
                for arg in u.args.iter() {
                    match arg {
                        Expression::Ref(reference) => {
                            intersect_if_exists(reference, state, schema_intersection.clone());
                        }
                        _ => {
                            arg.match_derive_schema(state)?;
                        }
                    };
                }
                state.null_behavior = null_behavior;
            }
            Ok(())
        }

        fn match_derive_ne(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            // we can only constrain the schema for ne with unitary types, for example ne null
            if u.args.len() == 2 {
                let lhs_schema = u.args[0].derive_schema(state)?;
                let rhs_schema = u.args[1].derive_schema(state)?;
                match (&u.args[0], &u.args[1], &lhs_schema, &rhs_schema) {
                    (
                        Expression::Ref(left_ref),
                        Expression::Ref(right_ref),
                        Schema::Atomic(left_atomic),
                        Schema::Atomic(right_atomic),
                    ) => {
                        // if we have an ne that is strictly unsatisfiable, set the schemas for these fields to be unsat
                        if left_atomic == right_atomic
                            && is_unitary_schema(&Schema::Atomic(*left_atomic))
                        {
                            intersect_if_exists(left_ref, state, Schema::Unsat);
                            intersect_if_exists(right_ref, state, Schema::Unsat);
                        }
                    }
                    (Expression::Ref(reference), _, _, Schema::Atomic(a))
                    | (_, Expression::Ref(reference), Schema::Atomic(a), _) => {
                        if is_unitary_schema(&Schema::Atomic(*a)) {
                            result_set_schema_difference(
                                reference,
                                state,
                                set!(Schema::Atomic(*a)),
                            );
                        }
                    }
                    (left, right, _, _) => {
                        // $ne does not tell us anything if we do not have unitary schemas, so we recurse on each sub-expression
                        // with null_behavior=May.
                        let null_behavior = state.null_behavior;
                        state.null_behavior = Satisfaction::May;
                        left.match_derive_schema(state)?;
                        right.match_derive_schema(state)?;
                        state.null_behavior = null_behavior;
                    }
                }
            }
            Ok(())
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
                    get_or_create_schema_for_path_mut(&mut state.result_set_schema, path)
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

        fn get_comparison_nullability(op: UntaggedOperatorName, schema: Schema) -> Satisfaction {
            // if we get a comparison that is <= null, this is null or missing. Technically, < null is only missing.
            // but for the purposes of constraining null we will keep these the same
            if (op == UntaggedOperatorName::Lt || op == UntaggedOperatorName::Lte)
                && schema.satisfies(&NULLISH.clone()) == Satisfaction::Must
            {
                Satisfaction::Must
            }
            // similarly, >= null would exclude missing but include null, but for the sake of constraining the schema
            // we will ignore gte to be safe since we don't handle missing separately.
            else if op == UntaggedOperatorName::Gt
                && schema.satisfies(&NULLISH.clone()) == Satisfaction::Not
            {
                Satisfaction::Not
            } else {
                Satisfaction::May
            }
        }

        fn match_derive_comparison(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            let null_behavior = state.null_behavior;
            if u.args.len() == 2 {
                let lhs_schema = u.args[0].derive_schema(state)?;
                let rhs_schema = u.args[1].derive_schema(state)?;
                if let Expression::Ref(reference) = &u.args[0] {
                    constrain_schema_for_comparison_reference(reference, u.op, state, rhs_schema);
                } else {
                    state.null_behavior = get_comparison_nullability(u.op, rhs_schema);
                    u.args[0].match_derive_schema(state)?;
                    state.null_behavior = null_behavior;
                }
                // we invert the operator, so that we can treat this field reference as the LHS of the comparison (in order to reuse the helper);
                // for example, if we want to constrain bar in the comparison foo < $bar, we can treat it as $bar > foo.
                let op = match u.op {
                    UntaggedOperatorName::Lt => UntaggedOperatorName::Gt,
                    UntaggedOperatorName::Lte => UntaggedOperatorName::Gte,
                    UntaggedOperatorName::Gt => UntaggedOperatorName::Lt,
                    UntaggedOperatorName::Gte => UntaggedOperatorName::Lte,
                    _ => return Ok(()),
                };
                if let Expression::Ref(reference) = &u.args[1] {
                    constrain_schema_for_comparison_reference(reference, op, state, lhs_schema);
                } else {
                    state.null_behavior = get_comparison_nullability(op, lhs_schema);
                    u.args[1].match_derive_schema(state)?;
                    state.null_behavior = null_behavior;
                }
            }
            Ok(())
        }

        fn match_derive_numeric(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            for arg in u.args.iter() {
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
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_add(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(
                                reference,
                                state,
                                NUMERIC.clone().union(&Schema::Atomic(Atomic::Date)),
                            );
                        }
                        // we cannot infer that if $add must be null, a reference must be null, because it returns null if _any_
                        // argument is null. Thus, {$add: [$foo, null]} would yield null but tell us nothing about $foo. We could,
                        // in the future, examine if all other args cannot be null and then enforce something on foo.
                        Satisfaction::May | Satisfaction::Must => {
                            intersect_if_exists(
                                reference,
                                state,
                                NUMERIC_OR_NULLISH
                                    .clone()
                                    .union(&Schema::Atomic(Atomic::Date)),
                            );
                        }
                    }
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_binary_size(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            if u.args.is_empty() {
                return Err(Error::NotEnoughArguments(u.op.to_string()));
            }
            if let Expression::Ref(ref reference) = u.args[0] {
                match state.null_behavior {
                    Satisfaction::Not => intersect_if_exists(
                        reference,
                        state,
                        Schema::AnyOf(set![
                            Schema::Atomic(Atomic::String),
                            Schema::Atomic(Atomic::BinData),
                        ]),
                    ),
                    Satisfaction::May => intersect_if_exists(
                        reference,
                        state,
                        Schema::AnyOf(set![
                            Schema::Atomic(Atomic::String),
                            Schema::Atomic(Atomic::BinData),
                            Schema::Missing,
                            Schema::Atomic(Atomic::Null)
                        ]),
                    ),
                    Satisfaction::Must => {
                        intersect_if_exists(reference, state, NULLISH.clone());
                    }
                }
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            Ok(())
        }

        fn match_derive_index_of(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if u.args.len() < 2 {
                return Err(Error::NotEnoughArguments(u.op.to_string()));
            }
            let (input_schema, find_schema, number_schema) = match state.null_behavior {
                Satisfaction::Not => (
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Integer),
                ),
                Satisfaction::May | Satisfaction::Must => {
                    // find schema can never be null, but input schema can be. Fun
                    (
                        STRING_OR_NULLISH.clone(),
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Integer),
                    )
                }
            };
            macro_rules! handle_arg {
                ($arg: expr, $sch: expr) => {
                    if let Expression::Ref(ref reference) = $arg {
                        intersect_if_exists(reference, state, $sch);
                    } else {
                        $arg.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(u.args[0], input_schema);
            handle_arg!(u.args[1], find_schema);
            if u.args.len() == 3 {
                handle_arg!(u.args[2], number_schema);
            } else if u.args.len() == 4 {
                // I don't like repeating the code, but it avoids a clone when there are only 3
                // args
                handle_arg!(u.args[2], number_schema.clone());
                handle_arg!(u.args[3], number_schema);
            }
            Ok(())
        }

        fn match_derive_substr(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if u.args.len() < 3 {
                return Err(Error::NotEnoughArguments(u.op.to_string()));
            }

            let (string_schema, number_schema) = match state.null_behavior {
                Satisfaction::Not => (Schema::Atomic(Atomic::String), NUMERIC.clone()),
                Satisfaction::May | Satisfaction::Must => {
                    (STRING_OR_NULLISH.clone(), NUMERIC_OR_NULLISH.clone())
                }
            };
            macro_rules! handle_arg {
                ($arg: expr, $sch: expr) => {
                    if let Expression::Ref(ref reference) = $arg {
                        intersect_if_exists(reference, state, $sch);
                    } else {
                        $arg.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(u.args[0], string_schema);
            handle_arg!(u.args[1], number_schema.clone());
            handle_arg!(u.args[2], number_schema);
            Ok(())
        }

        fn match_derive_subtract(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if u.args.len() == 2 {
                if let Expression::Ref(reference) = &u.args[0] {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(
                                reference,
                                state,
                                NUMERIC.clone().union(&Schema::Atomic(Atomic::Date)),
                            );
                        }
                        // we cannot infer that if $add must be null, a reference must be null, because it returns null if _any_
                        // argument is null. Thus, {$add: [$foo, null]} would yield null but tell us nothing about $foo. We could,
                        // in the future, examine if all other args cannot be null and then enforce something on foo.
                        Satisfaction::May | Satisfaction::Must => {
                            intersect_if_exists(
                                reference,
                                state,
                                NUMERIC_OR_NULLISH
                                    .clone()
                                    .union(&Schema::Atomic(Atomic::Date)),
                            );
                        }
                    }
                } else {
                    u.args[0].match_derive_schema(state)?;
                }
                if let Expression::Ref(reference) = &u.args[1] {
                    // iff the first argument can be a date, it is possible the second argument is a date, since subtract [<date>, <date>]
                    // is valid. Otherwise, the second argument must be numeric, becuase subtract [<int>, <date>] is invalid.
                    let can_be_date = u.args[0]
                        .derive_schema(state)?
                        .satisfies(&Schema::Atomic(Atomic::Date))
                        != Satisfaction::Not;
                    match state.null_behavior {
                        Satisfaction::Not => {
                            if can_be_date {
                                intersect_if_exists(
                                    reference,
                                    state,
                                    NUMERIC.clone().union(&Schema::Atomic(Atomic::Date)),
                                );
                            } else {
                                intersect_if_exists(reference, state, NUMERIC.clone());
                            }
                        }
                        // we cannot infer that if $add must be null, a reference must be null, because it returns null if _any_
                        // argument is null. Thus, {$add: [$foo, null]} would yield null but tell us nothing about $foo. We could,
                        // in the future, examine if all other args cannot be null and then enforce something on foo.
                        Satisfaction::May | Satisfaction::Must => {
                            if can_be_date {
                                intersect_if_exists(
                                    reference,
                                    state,
                                    NUMERIC_OR_NULLISH
                                        .clone()
                                        .union(&Schema::Atomic(Atomic::Date)),
                                );
                            } else {
                                intersect_if_exists(reference, state, NUMERIC_OR_NULLISH.clone());
                            }
                        }
                    }
                } else {
                    u.args[1].match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_object_to_array(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => intersect_if_exists(
                            reference,
                            state,
                            Schema::Document(Document::default()),
                        ),
                        Satisfaction::May => intersect_if_exists(
                            reference,
                            state,
                            Schema::AnyOf(set!(
                                Schema::Document(Document::default()),
                                Schema::Missing,
                                Schema::Atomic(Atomic::Null)
                            )),
                        ),
                        Satisfaction::Must => {
                            intersect_if_exists(reference, state, NULLISH.clone());
                        }
                    }
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_let(l: &Let, state: &mut ResultSetState) -> Result<()> {
            let mut variables = state.variables.clone();
            for (var, expression) in l.vars.iter() {
                let schema = expression.derive_schema(state)?;
                state.variables.insert(var.to_string(), schema);
            }
            l.inside.match_derive_schema(state)?;
            for (var, expression) in l.vars.iter() {
                match expression {
                    Expression::Ref(Ref::FieldRef(field_ref)) => {
                        if let Some(v) = state.variables.get_mut(var) {
                            let path = field_ref
                                .as_str()
                                .split('.')
                                .map(|s| s.to_string())
                                .collect();
                            if let Some(f) = get_or_create_schema_for_path_mut(
                                &mut state.result_set_schema,
                                path,
                            ) {
                                *f = v.clone();
                            }
                        }
                    }
                    Expression::Ref(Ref::VariableRef(var_ref)) => {
                        if let Some(v) = state.variables.get(var) {
                            variables.insert(var_ref.clone(), v.clone());
                        }
                    }
                    expr => expr.match_derive_schema(state)?,
                }
            }
            state.variables = variables;
            Ok(())
        }

        fn match_derive_max_min(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            result_set_schema_difference(
                                reference,
                                state,
                                set!(Schema::Atomic(Atomic::Null), Schema::Missing),
                            );
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(
                                reference,
                                state,
                                Schema::AnyOf(set!(
                                    Schema::Atomic(Atomic::Null),
                                    Schema::Array(Box::new(Schema::Any)),
                                    Schema::Missing
                                )),
                            );
                        }
                        Satisfaction::May => {}
                    }
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_switch(s: &Switch, state: &mut ResultSetState) -> Result<()> {
            for case in s.branches.iter() {
                if let Expression::Ref(reference) = case.case.as_ref() {
                    intersect_if_exists(reference, state, Schema::Atomic(Atomic::Boolean));
                } else {
                    case.case.match_derive_schema(state)?;
                }
                case.then.match_derive_schema(state)?;
            }
            s.default.match_derive_schema(state)
        }

        fn match_derive_bit_ops(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            for arg in u.args.iter() {
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
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_is_number(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(reference, state, NUMERIC.clone());
                    }
                    Satisfaction::Must => {
                        result_set_schema_difference(
                            reference,
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
            } else {
                for arg in u.args.iter() {
                    arg.match_derive_schema(state)?
                }
            }
            Ok(())
        }

        fn match_derive_range(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    intersect_if_exists(reference, state, NUMERIC.clone());
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_round(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if u.args.is_empty() {
                return Err(Error::NotEnoughArguments(u.op.to_string()));
            }

            if let Expression::Ref(ref reference) = u.args[0] {
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
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            if u.args.len() > 1 {
                if let Expression::Ref(ref reference) = u.args[1] {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(
                                reference,
                                state,
                                Schema::AnyOf(set!(
                                    Schema::Atomic(Atomic::Integer),
                                    Schema::Atomic(Atomic::Long)
                                )),
                            );
                        }
                        Satisfaction::May => {
                            intersect_if_exists(reference, state, INTEGER_LONG_OR_NULLISH.clone());
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(reference, state, NULLISH.clone());
                        }
                    };
                } else {
                    u.args[1].match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        fn match_derive_numeric_conversion(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            if u.args.is_empty() {
                return Err(Error::NotEnoughArguments(u.op.to_string()));
            }
            if let Expression::Ref(ref reference) = u.args[0] {
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
                        intersect_if_exists(reference, state, numeric_convertible);
                    }
                    Satisfaction::May => {
                        intersect_if_exists(
                            reference,
                            state,
                            numeric_convertible.union(&NULLISH.clone()),
                        );
                    }
                    Satisfaction::Must => {
                        intersect_if_exists(reference, state, NULLISH.clone());
                    }
                };
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            Ok(())
        }

        fn match_derive_replace(r: &Replace, state: &mut ResultSetState) -> Result<()> {
            let arg_schema = match state.null_behavior {
                Satisfaction::Not => Schema::Atomic(Atomic::String),
                Satisfaction::May | Satisfaction::Must => STRING_OR_NULLISH.clone(),
            };
            macro_rules! handle_arg {
                ($input:expr, $sch:expr) => {
                    if let Expression::Ref(ref reference) = $input {
                        intersect_if_exists(reference, state, $sch);
                    } else {
                        $input.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(*r.input, arg_schema.clone());
            handle_arg!(*r.find, arg_schema.clone());
            handle_arg!(*r.replacement, arg_schema);
            Ok(())
        }

        fn match_derive_non_nullable_regex(
            r: &RegexAggExpression,
            state: &mut ResultSetState,
        ) -> Result<()> {
            // $regexMatch and $regexFindAll cannot ever be null, so the null_behavior is irrelevant
            let (string_schema, regex_schema) = (
                Schema::Atomic(Atomic::String),
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Regex)
                ]),
            );
            macro_rules! handle_arg {
                ($input:expr, $sch:expr) => {
                    if let Expression::Ref(ref reference) = $input {
                        intersect_if_exists(reference, state, $sch);
                    } else {
                        $input.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(*r.input, string_schema.clone());
            handle_arg!(*r.regex, regex_schema);
            // options can always be NULL without producing NULL as an output, so
            // this behavior is regardless of null_behavior
            if let Some(options) = &r.options {
                handle_arg!(*options.as_ref(), STRING_OR_NULLISH.clone());
            }
            Ok(())
        }

        fn match_derive_regex_find(
            r: &RegexAggExpression,
            state: &mut ResultSetState,
        ) -> Result<()> {
            // $regexFind can be null, so the null_behavior is relevant
            let (string_schema, regex_schema) = match state.null_behavior {
                Satisfaction::Not => (
                    Schema::Atomic(Atomic::String),
                    Schema::AnyOf(set![
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Regex)
                    ]),
                ),
                Satisfaction::May | Satisfaction::Must => (
                    STRING_OR_NULLISH.clone(),
                    Schema::AnyOf(set![
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Regex),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    ]),
                ),
            };
            macro_rules! handle_arg {
                ($input:expr, $sch:expr) => {
                    if let Expression::Ref(ref reference) = $input {
                        intersect_if_exists(reference, state, $sch);
                    } else {
                        $input.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(*r.input, string_schema.clone());
            handle_arg!(*r.regex, regex_schema);
            if let Some(options) = &r.options {
                // options can be NULLISH no matter the current null_behavior
                handle_arg!(*options.as_ref(), STRING_OR_NULLISH.clone());
            }
            Ok(())
        }

        fn match_derive_to_string(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if u.args.is_empty() {
                return Err(Error::NotEnoughArguments(u.op.to_string()));
            }
            if let Expression::Ref(reference) = &u.args[0] {
                let schema = match state.null_behavior {
                    Satisfaction::Not => Schema::AnyOf(set! {
                        Schema::Atomic(Atomic::BinData),
                        Schema::Atomic(Atomic::Boolean),
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Decimal),
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::ObjectId),
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Date),
                    }),
                    Satisfaction::May => Schema::AnyOf(set! {
                        Schema::Atomic(Atomic::BinData),
                        Schema::Atomic(Atomic::Boolean),
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Decimal),
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::ObjectId),
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Date),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    }),
                    Satisfaction::Must => NULLISH.clone(),
                };
                intersect_if_exists(reference, state, schema);
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            Ok(())
        }

        fn match_derive_trim(t: &Trim, state: &mut ResultSetState) -> Result<()> {
            let arg_schema = match state.null_behavior {
                Satisfaction::Not => Schema::Atomic(Atomic::String),
                Satisfaction::May | Satisfaction::Must => STRING_OR_NULLISH.clone(),
            };
            macro_rules! handle_arg {
                ($input:expr, $sch:expr) => {
                    if let Expression::Ref(ref reference) = $input {
                        intersect_if_exists(reference, state, $sch);
                    } else {
                        $input.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(*t.input, arg_schema.clone());
            if let Some(chars) = &t.chars {
                handle_arg!(*chars.as_ref(), arg_schema);
            }
            Ok(())
        }

        /// match_derive_first_last constrains the result_set_schema for any field refs in
        /// $first or $last operators. These operators work on arrays, and can only return
        /// null if the field reference itself is null or if the array contains null values.
        fn match_derive_first_last(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)));
                    }
                    Satisfaction::May | Satisfaction::Must => {
                        intersect_if_exists(
                            reference,
                            state,
                            Schema::AnyOf(set!(
                                Schema::Array(Box::new(Schema::Any)),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing
                            )),
                        );
                    }
                };
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_array_op generically constrains the result_set_schema for the group
        /// of array operators that can only return null when the input is null, and can
        /// only return a value when the input is an array (examples: $concatArrays)
        fn match_derive_array_op(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(
                                reference,
                                state,
                                Schema::Array(Box::new(Schema::Any)),
                            );
                        }
                        Satisfaction::May => {
                            intersect_if_exists(
                                reference,
                                state,
                                NULLISH.clone().union(&Schema::Array(Box::new(Schema::Any))),
                            );
                        }
                        Satisfaction::Must => {
                            intersect_if_exists(reference, state, NULLISH.clone());
                        }
                    }
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        /// match_derive_zip constrains the result_set_schema for $zip operators in match.
        /// the inputs is a vec of arrays, but if any element of this vec is null, the whole
        /// operator can return null. Defaults can be any of an array or nullish
        fn match_derive_zip(z: &Zip, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Array(v) = z.inputs.as_ref() {
                v.iter().for_each(|input| {
                    if let Expression::Ref(reference) = input {
                        let mut schema = Schema::Array(Box::new(Schema::Any));
                        if state.null_behavior != Satisfaction::Not {
                            schema = schema.union(&NULLISH.clone());
                        }
                        intersect_if_exists(reference, state, schema);
                    }
                });
            }
            if let Some(ref a) = z.defaults {
                if let Expression::Array(v) = a.as_ref() {
                    v.iter().for_each(|input| {
                        if let Expression::Ref(reference) = input {
                            intersect_if_exists(
                                reference,
                                state,
                                Schema::AnyOf(set!(
                                    Schema::Array(Box::new(Schema::Any)),
                                    Schema::Atomic(Atomic::Null),
                                    Schema::Missing
                                )),
                            )
                        }
                    });
                }
            }
            Ok(())
        }

        /// match_derive_non_nullish_array_ops constrains the result_set_schema for
        /// operators that only deal with arrays, independent of the expected matching type.
        fn match_derive_non_nullish_array_ops(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            for arg in u.args.iter() {
                if let Expression::Ref(reference) = arg {
                    intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)));
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        /// match_derive_n_array_op constraints the result_set_schema for operators that
        /// take a fixed number of elements from an array, such as $firstN. The input must
        /// be an array, and the number must be numeric, regargless of nullability.
        fn match_derive_n_array_op(n: &NArrayOp, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = n.input.as_ref() {
                intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)));
            } else {
                n.input.match_derive_schema(state)?;
            }
            if let Expression::Ref(reference) = n.n.as_ref() {
                intersect_if_exists(reference, state, NUMERIC.clone());
            } else {
                n.input.match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_index_of_array constrains the result_set_schema for $indexOfArray
        /// operators. The first arg represents the array, and and must be null if the
        /// operator returns null. The third and fourth args represent the bounds to search
        /// and must be numeric.
        fn match_derive_index_of_array(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)))
                    }
                    Satisfaction::May => intersect_if_exists(
                        reference,
                        state,
                        Schema::AnyOf(set!(
                            Schema::Array(Box::new(Schema::Any)),
                            Schema::Atomic(Atomic::Null),
                            Schema::Missing
                        )),
                    ),
                    Satisfaction::Must => intersect_if_exists(reference, state, NULLISH.clone()),
                }
            }
            u.args[1].match_derive_schema(state)?;
            for arg in u.args[2..].iter() {
                if let Expression::Ref(reference) = arg {
                    intersect_if_exists(
                        reference,
                        state,
                        Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Integer),
                            Schema::Atomic(Atomic::Long)
                        )),
                    );
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        /// match_derive_all_elements_true constrains the result_set_schema for the
        /// $allElementsTrue operator. If this operator must return true, then we can
        /// remove null and missing from the types the array can contain.
        fn match_derive_all_elements_true(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            match &u.args[0] {
                Expression::Array(a) => {
                    for expr in a {
                        if state.null_behavior == Satisfaction::Not {
                            if let Expression::Ref(reference) = expr {
                                result_set_schema_difference(
                                    reference,
                                    state,
                                    set!(Schema::Atomic(Atomic::Null), Schema::Missing),
                                );
                            } else {
                                expr.match_derive_schema(state)?;
                            }
                        } else {
                            expr.match_derive_schema(state)?;
                        }
                    }
                }
                Expression::Ref(reference) => match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)));
                    }
                    Satisfaction::May | Satisfaction::Must => {
                        intersect_if_exists(
                            reference,
                            state,
                            Schema::AnyOf(set!(
                                Schema::Array(Box::new(Schema::Any)),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing
                            )),
                        );
                    }
                },
                expr => {
                    expr.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        /// match_derive_in constrains the result_set_schema for $in operators. The second
        /// arg must be an array, while the first arg can be anything (the search expr)
        fn match_derive_in(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            u.args[0].match_derive_schema(state)?;
            if let Expression::Ref(reference) = &u.args[1] {
                intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)));
            } else {
                u.args[1].match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_array_elem_at constrains the result_set_schema for $arrayElemAt
        /// operators. The first argument is the array, and the second is the index; either
        /// can be nullish if null is allowable.
        fn match_derive_array_elem_at(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                let mut schema = Schema::Array(Box::new(Schema::Any));
                if state.null_behavior != Satisfaction::Not {
                    schema = schema.union(&NULLISH.clone());
                }
                intersect_if_exists(reference, state, schema);
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            if let Expression::Ref(reference) = &u.args[1] {
                let schema = match state.null_behavior {
                    Satisfaction::Not => NUMERIC.clone(),
                    Satisfaction::May | Satisfaction::Must => NUMERIC_OR_NULLISH.clone(),
                };
                intersect_if_exists(reference, state, schema);
            } else {
                u.args[1].match_derive_schema(state)?;
            }
            Ok(())
        }

        fn match_derive_date_arithmetic(
            start_date: &Expression,
            unit: &Expression,
            amount: &Expression,
            timezone: Option<&Expression>,
            state: &mut ResultSetState,
        ) -> Result<()> {
            macro_rules! handle_arg {
                ($exp:expr, $basic_schema:expr) => {
                    if let Expression::Ref(ref reference) = $exp {
                        match state.null_behavior {
                            Satisfaction::Not => {
                                intersect_if_exists(reference, state, $basic_schema);
                            }
                            Satisfaction::May | Satisfaction::Must => {
                                let schema = $basic_schema.union(&NULLISH.clone());
                                intersect_if_exists(reference, state, schema);
                            }
                        }
                    } else {
                        $exp.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(
                start_date,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Date),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::Timestamp),
                ])
            );
            handle_arg!(unit, Schema::Atomic(Atomic::String));
            handle_arg!(amount, NUMERIC.clone());
            if let Some(ref timezone) = timezone {
                handle_arg!(timezone, Schema::Atomic(Atomic::String));
            }
            Ok(())
        }

        fn match_derive_date_add(d: &DateAdd, state: &mut ResultSetState) -> Result<()> {
            let tz = d.timezone.as_ref().map(|tz| tz.as_ref());
            match_derive_date_arithmetic(
                d.start_date.as_ref(),
                d.unit.as_ref(),
                d.amount.as_ref(),
                tz,
                state,
            )
        }

        fn match_derive_date_subtract(d: &DateSubtract, state: &mut ResultSetState) -> Result<()> {
            let tz = d.timezone.as_ref().map(|tz| tz.as_ref());
            match_derive_date_arithmetic(
                d.start_date.as_ref(),
                d.unit.as_ref(),
                d.amount.as_ref(),
                tz,
                state,
            )
        }

        fn match_derive_date_diff(d: &DateDiff, state: &mut ResultSetState) -> Result<()> {
            macro_rules! handle_arg {
                ($exp:expr, $basic_schema:expr) => {
                    if let Expression::Ref(ref reference) = $exp {
                        match state.null_behavior {
                            Satisfaction::Not => {
                                intersect_if_exists(reference, state, $basic_schema);
                            }
                            Satisfaction::May | Satisfaction::Must => {
                                let schema = $basic_schema.union(&NULLISH.clone());
                                intersect_if_exists(reference, state, schema);
                            }
                        }
                    } else {
                        $exp.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(
                d.start_date.as_ref(),
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Date),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::Timestamp),
                ])
            );
            handle_arg!(
                d.end_date.as_ref(),
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Date),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::Timestamp),
                ])
            );
            handle_arg!(d.unit.as_ref(), Schema::Atomic(Atomic::String));
            if let Some(ref timezone) = d.timezone {
                handle_arg!(timezone.as_ref(), Schema::Atomic(Atomic::String));
            }
            if let Some(ref start_of_week) = d.start_of_week {
                handle_arg!(start_of_week.as_ref(), Schema::Atomic(Atomic::String));
            }
            Ok(())
        }

        fn match_derive_date_trunc(d: &DateTrunc, state: &mut ResultSetState) -> Result<()> {
            macro_rules! handle_arg {
                ($exp:expr, $basic_schema:expr) => {
                    if let Expression::Ref(ref reference) = $exp {
                        match state.null_behavior {
                            Satisfaction::Not => {
                                intersect_if_exists(reference, state, $basic_schema);
                            }
                            Satisfaction::May | Satisfaction::Must => {
                                let schema = $basic_schema.union(&NULLISH.clone());
                                intersect_if_exists(reference, state, schema);
                            }
                        }
                    } else {
                        $exp.match_derive_schema(state)?;
                    }
                };
            }
            handle_arg!(
                d.date.as_ref(),
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Date),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::Timestamp),
                ])
            );
            handle_arg!(d.unit.as_ref(), Schema::Atomic(Atomic::String));
            if let Some(ref bin_size) = d.bin_size {
                handle_arg!(bin_size.as_ref(), NUMERIC.clone());
            }
            if let Some(ref timezone) = d.timezone {
                handle_arg!(timezone.as_ref(), Schema::Atomic(Atomic::String));
            }
            if let Some(ref start_of_week) = d.start_of_week {
                handle_arg!(start_of_week.as_ref(), Schema::Atomic(Atomic::String));
            }
            Ok(())
        }

        /// match_derive_array_to_object constrains the result_set_schema for $arrayToObject
        /// oeprators. For this operator to return a non-null value, the input must be an
        /// array of arrays.
        fn match_derive_array_to_object(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(
                            reference,
                            state,
                            Schema::Array(Box::new(Schema::Array(Box::new(Schema::Any)))),
                        );
                    }
                    Satisfaction::May => {
                        intersect_if_exists(
                            reference,
                            state,
                            Schema::AnyOf(set!(
                                Schema::Array(Box::new(Schema::Array(Box::new(Schema::Any)))),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing
                            )),
                        );
                    }
                    Satisfaction::Must => {
                        intersect_if_exists(reference, state, NULLISH.clone());
                    }
                }
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_merge_objects constrains the result_set_schema for $mergeObjects
        /// operators. The input must always be either a singular document or an array
        /// of documents.
        fn match_derive_merge_objects(
            u: &UntaggedOperator,
            state: &mut ResultSetState,
        ) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                intersect_if_exists(
                    reference,
                    state,
                    Schema::AnyOf(set!(
                        Schema::Array(Box::new(Schema::Document(Document::any()))),
                        Schema::Document(Document::any()),
                    )),
                );
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_sort_array constrains the result_set_schema for $sortArray
        /// operators. This functions the same as match_derive_array_op, however this is a
        /// tagged operator so needs its own implementation.
        fn match_derive_sort_array(s: &SortArray, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = s.input.as_ref() {
                match state.null_behavior {
                    Satisfaction::Not => {
                        intersect_if_exists(reference, state, Schema::Array(Box::new(Schema::Any)))
                    }
                    Satisfaction::May => intersect_if_exists(
                        reference,
                        state,
                        Schema::AnyOf(set!(
                            Schema::Array(Box::new(Schema::Any)),
                            Schema::Atomic(Atomic::Null),
                            Schema::Missing
                        )),
                    ),
                    Satisfaction::Must => intersect_if_exists(reference, state, NULLISH.clone()),
                }
            } else {
                s.input.match_derive_schema(state)?;
            }
            Ok(())
        }

        macro_rules! derive_map_filter_input {
            ($input:expr, $state:expr, $inside:expr) => {
                if let Expression::Ref(reference) = $input.input.as_ref() {
                    match $state.null_behavior {
                        Satisfaction::Not => intersect_if_exists(
                            reference,
                            $state,
                            Schema::Array(Box::new(Schema::Any)),
                        ),
                        Satisfaction::May => intersect_if_exists(
                            reference,
                            $state,
                            Schema::AnyOf(set!(
                                Schema::Array(Box::new(Schema::Any)),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing
                            )),
                        ),
                        Satisfaction::Must => {
                            intersect_if_exists(reference, $state, NULLISH.clone())
                        }
                    }
                } else {
                    $input.input.match_derive_schema($state)?;
                }
                let variable_name = $input._as.clone().unwrap_or("this".to_string());
                let variable_schema = $input.input.derive_schema($state)?;
                $state
                    .variables
                    .insert(variable_name.clone(), variable_schema);
                $inside.match_derive_schema($state)?;
                $state.variables.remove(&variable_name);
            };
        }

        /// match_derive_filter constrains the result_set_schema for $filter operators. The
        /// input must be an array to return a non-null value, or nullish to return null.
        /// We then add the inputs schema to the variable environment before recursing on
        /// the other arguments to this operator.
        fn match_derive_filter(f: &Filter, state: &mut ResultSetState) -> Result<()> {
            derive_map_filter_input!(f, state, f.cond);
            if let Some(limit) = f.limit.as_ref() {
                limit.match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_map constrains the result_set_schema for $map. It functions the
        /// same as $filter, checking the input and recursing on the additional args.
        fn match_derive_map(m: &Map, state: &mut ResultSetState) -> Result<()> {
            derive_map_filter_input!(m, state, m.inside);
            Ok(())
        }

        /// match_derive_median constrains the result_set_schema for $median. The input must
        /// be an array to return a non-null value, but can be an array or nullish to return
        /// null.
        fn match_derive_median(m: &Median, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = m.input.as_ref() {
                let mut schema = Schema::Array(Box::new(Schema::Any));
                if state.null_behavior != Satisfaction::Not {
                    schema = schema.union(&NULLISH.clone());
                }
                intersect_if_exists(reference, state, schema);
            } else {
                m.input.match_derive_schema(state)?;
            }
            Ok(())
        }

        /// match_derive_reduce constrains the result_set_schema for $reduce. The input must
        /// be an array to return a non-null value, but can be an array or nullish to return
        /// null. Other args can be expressions and so are recursed on.
        fn match_derive_reduce(r: &Reduce, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = r.input.as_ref() {
                let mut schema = Schema::Array(Box::new(Schema::Any));
                if state.null_behavior != Satisfaction::Not {
                    schema = schema.union(&NULLISH.clone());
                }
                intersect_if_exists(reference, state, schema);
            } else {
                r.input.match_derive_schema(state)?;
            }
            // if the input is null, there is no array to iterate over, so we can safely return
            let input_schema = r.input.derive_schema(state)?;
            if input_schema.satisfies(&NULLISH.clone()) == Satisfaction::Must {
                return Ok(());
            }
            let array_element_schema = array_element_schema_or_error!(input_schema, r.input);
            let initial_schema = r.initial_value.derive_schema(state)?;
            let initial_variables = state.variables.clone();
            state
                .variables
                .insert("this".to_string(), array_element_schema);
            state.variables.insert("value".to_string(), initial_schema);
            r.initial_value.match_derive_schema(state)?;
            r.inside.match_derive_schema(state)?;
            state.variables = initial_variables;
            Ok(())
        }

        /// match_derive_slice constrains the result_set_schema for $slice. The input must
        /// be an array to return a non-null value, but can be an array or nullish to return
        /// null. Other args represent bounds, and are thus constrained on as numeric.
        fn match_derive_slice(u: &UntaggedOperator, state: &mut ResultSetState) -> Result<()> {
            if let Expression::Ref(reference) = &u.args[0] {
                let mut schema = Schema::Array(Box::new(Schema::Any));
                if state.null_behavior != Satisfaction::Not {
                    schema = schema.union(&NULLISH.clone());
                }
                intersect_if_exists(reference, state, schema);
            } else {
                u.args[0].match_derive_schema(state)?;
            }
            for arg in u.args[1..].iter() {
                if let Expression::Ref(reference) = arg {
                    match state.null_behavior {
                        Satisfaction::Not => {
                            intersect_if_exists(reference, state, NUMERIC.clone());
                        }
                        Satisfaction::May | Satisfaction::Must => {
                            intersect_if_exists(reference, state, NUMERIC_OR_NULLISH.clone());
                        }
                    }
                } else {
                    arg.match_derive_schema(state)?;
                }
            }
            Ok(())
        }

        use agg_ast::definitions::UntaggedOperatorName;
        let null_behavior = state.null_behavior;
        match self {
            Expression::TaggedOperator(t) => match t {
                TaggedOperator::Let(l) => match_derive_let(l, state)?,
                TaggedOperator::Switch(s) => match_derive_switch(s, state)?,
                TaggedOperator::Year(d)
                | TaggedOperator::Month(d)
                | TaggedOperator::IsoWeekYear(d)
                | TaggedOperator::Week(d)
                | TaggedOperator::IsoWeek(d)
                | TaggedOperator::DayOfWeek(d)
                | TaggedOperator::IsoDayOfWeek(d)
                | TaggedOperator::DayOfYear(d)
                | TaggedOperator::DayOfMonth(d)
                | TaggedOperator::Hour(d)
                | TaggedOperator::Minute(d)
                | TaggedOperator::Second(d)
                | TaggedOperator::Millisecond(d) => match_derive_date_expression(d, state)?,
                TaggedOperator::DateToParts(d) => match_derive_date_to_parts(d, state)?,
                TaggedOperator::DateFromParts(d) => match_derive_date_from_parts(d, state)?,
                TaggedOperator::DateFromString(d) => match_derive_date_from_string(d, state)?,
                TaggedOperator::DateToString(d) => match_derive_date_to_string(d, state)?,
                TaggedOperator::Zip(z) => match_derive_zip(z, state)?,
                TaggedOperator::FirstN(n)
                | TaggedOperator::LastN(n)
                | TaggedOperator::MaxNArrayElement(n)
                | TaggedOperator::MinNArrayElement(n) => match_derive_n_array_op(n, state)?,
                TaggedOperator::SortArray(s) => match_derive_sort_array(s, state)?,
                TaggedOperator::Filter(f) => match_derive_filter(f, state)?,
                TaggedOperator::Map(m) => match_derive_map(m, state)?,
                TaggedOperator::Median(m) => match_derive_median(m, state)?,
                TaggedOperator::Reduce(r) => match_derive_reduce(r, state)?,
                TaggedOperator::ReplaceOne(r) | TaggedOperator::ReplaceAll(r) => {
                    match_derive_replace(r, state)?
                }
                TaggedOperator::RegexFind(r) => match_derive_regex_find(r, state)?,
                TaggedOperator::Regex(r) | TaggedOperator::RegexFindAll(r) => {
                    match_derive_non_nullable_regex(r, state)?
                }
                TaggedOperator::Trim(t) | TaggedOperator::LTrim(t) | TaggedOperator::RTrim(t) => {
                    match_derive_trim(t, state)?
                }
                TaggedOperator::DateAdd(d) => match_derive_date_add(d, state)?,
                TaggedOperator::DateDiff(d) => match_derive_date_diff(d, state)?,
                TaggedOperator::DateSubtract(d) => match_derive_date_subtract(d, state)?,
                TaggedOperator::DateTrunc(d) => match_derive_date_trunc(d, state)?,
                _ => {}
            },

            Expression::UntaggedOperator(u) => match u.op {
                // logical ops
                UntaggedOperatorName::And => match_derive_and(u, state)?,
                UntaggedOperatorName::Or => match_derive_or(u, state)?,
                // comparison ops
                UntaggedOperatorName::Eq => match_derive_eq(u, state)?,
                UntaggedOperatorName::Cmp | UntaggedOperatorName::Ne => match_derive_ne(u, state)?,
                UntaggedOperatorName::Gt
                | UntaggedOperatorName::Gte
                | UntaggedOperatorName::Lt
                | UntaggedOperatorName::Lte => match_derive_comparison(u, state)?,
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
                | UntaggedOperatorName::Floor => match_derive_numeric(u, state)?,
                // array ops
                UntaggedOperatorName::First | UntaggedOperatorName::Last => {
                    match_derive_first_last(u, state)?
                }
                UntaggedOperatorName::ConcatArrays | UntaggedOperatorName::ReverseArray => {
                    match_derive_array_op(u, state)?
                }
                UntaggedOperatorName::AnyElementTrue
                | UntaggedOperatorName::IsArray
                | UntaggedOperatorName::SetDifference
                | UntaggedOperatorName::SetEquals
                | UntaggedOperatorName::SetIntersection
                | UntaggedOperatorName::SetIsSubset
                | UntaggedOperatorName::SetUnion
                | UntaggedOperatorName::Size => match_derive_non_nullish_array_ops(u, state)?,
                UntaggedOperatorName::IndexOfArray => match_derive_index_of_array(u, state)?,
                UntaggedOperatorName::AllElementsTrue => match_derive_all_elements_true(u, state)?,
                UntaggedOperatorName::In => match_derive_in(u, state)?,
                UntaggedOperatorName::ArrayElemAt => match_derive_array_elem_at(u, state)?,
                UntaggedOperatorName::ArrayToObject => match_derive_array_to_object(u, state)?,
                UntaggedOperatorName::MergeObjects => match_derive_merge_objects(u, state)?,
                UntaggedOperatorName::Slice => match_derive_slice(u, state)?,
                // misc ops
                UntaggedOperatorName::Add => match_derive_add(u, state)?,
                UntaggedOperatorName::Subtract => match_derive_subtract(u, state)?,
                UntaggedOperatorName::Sum => {}
                UntaggedOperatorName::ObjectToArray => match_derive_object_to_array(u, state)?,
                UntaggedOperatorName::Max | UntaggedOperatorName::Min => {
                    match_derive_max_min(u, state)?
                }
                UntaggedOperatorName::BitAnd
                | UntaggedOperatorName::BitNot
                | UntaggedOperatorName::BitOr
                | UntaggedOperatorName::BitXor => match_derive_bit_ops(u, state)?,
                UntaggedOperatorName::IsNumber => match_derive_is_number(u, state)?,
                UntaggedOperatorName::Range => match_derive_range(u, state)?,
                UntaggedOperatorName::Round => match_derive_round(u, state)?,
                UntaggedOperatorName::ToInt
                | UntaggedOperatorName::ToDouble
                | UntaggedOperatorName::ToDecimal
                | UntaggedOperatorName::ToLong => match_derive_numeric_conversion(u, state)?,
                UntaggedOperatorName::Concat
                | UntaggedOperatorName::Strcasecmp
                | UntaggedOperatorName::StrLenCP
                | UntaggedOperatorName::StrLenBytes
                | UntaggedOperatorName::ToObjectId
                | UntaggedOperatorName::ToUpper
                | UntaggedOperatorName::ToLower => match_derive_string_operation(u, state)?,
                UntaggedOperatorName::Substr
                | UntaggedOperatorName::SubstrCP
                | UntaggedOperatorName::SubstrBytes => match_derive_substr(u, state)?,
                UntaggedOperatorName::BinarySize => match_derive_binary_size(u, state)?,
                UntaggedOperatorName::ToString => match_derive_to_string(u, state)?,
                UntaggedOperatorName::IndexOfCP | UntaggedOperatorName::IndexOfBytes => {
                    match_derive_index_of(u, state)?
                }
                _ => {}
            },
            _ => {}
        }
        state.null_behavior = null_behavior;
        Ok(())
    }
}

impl MatchConstrainSchema for MatchExpr {
    fn match_derive_schema(&self, state: &mut ResultSetState) -> Result<()> {
        self.expr.match_derive_schema(state)
    }
}
