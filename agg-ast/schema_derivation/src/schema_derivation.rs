use crate::{
    get_schema_for_path_mut, promote_missing, remove_field, schema_difference,
    schema_for_type_numeric, schema_for_type_str, Error, Result,
};
use agg_ast::definitions::{
    Expression, LiteralValue, Ref, Stage, TaggedOperator, UntaggedOperator, UntaggedOperatorName,
};
use mongosql::{
    map,
    schema::{
        Atomic, Document, Satisfaction, Schema, ANY_DOCUMENT, DATE_OR_NULLISH, EMPTY_DOCUMENT,
        INTEGER_LONG_OR_NULLISH, INTEGER_OR_NULLISH, INTEGRAL, NULLISH, NULLISH_OR_UNDEFINED,
        NUMERIC, NUMERIC_OR_NULLISH,
    },
    set,
};
use std::collections::{BTreeMap, BTreeSet};

#[allow(dead_code)]
pub(crate) trait DeriveSchema {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema>;
}

#[allow(dead_code)]
#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub(crate) struct ResultSetState<'a> {
    pub catalog: &'a BTreeMap<String, Schema>,
    pub variables: BTreeMap<String, Schema>,
    pub result_set_schema: Schema,
    // the null_behavior field allows us to keep track of what behavior we are expecting to be exhibited
    // by the rows returned by this query. This comes up in both normal schema derivation, where something like
    // $eq: [null, {$op: ...}] can influence the values returned by the operator), as well as in match schema derivation
    // where more broadly things like null field references or a falsifiable return type (e.g. {$eq: [{$op: ...}, 0])
    // may influcence they types of values the underlying result_set_schema can contain.
    pub null_behavior: Satisfaction,
}

impl DeriveSchema for Stage {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        match *self {
            Stage::AddFields(_) => todo!(),
            Stage::AtlasSearchStage(_) => todo!(),
            Stage::Bucket(_) => todo!(),
            Stage::BucketAuto(_) => todo!(),
            Stage::Collection(_) => todo!(),
            Stage::Count(_) => todo!(),
            Stage::Densify(_) => todo!(),
            Stage::Documents(_) => todo!(),
            Stage::EquiJoin(_) => todo!(),
            Stage::Facet(_) => todo!(),
            Stage::Fill(_) => todo!(),
            Stage::GeoNear(_) => todo!(),
            Stage::GraphLookup(_) => todo!(),
            Stage::Group(_) => todo!(),
            Stage::Join(_) => todo!(),
            Stage::Limit(_) => todo!(),
            Stage::Lookup(_) => todo!(),
            Stage::Match(ref m) => m.derive_schema(state),
            Stage::Project(_) => todo!(),
            Stage::Redact(_) => todo!(),
            Stage::ReplaceWith(_) => todo!(),
            Stage::Sample(_) => todo!(),
            Stage::SetWindowFields(_) => todo!(),
            Stage::Skip(_) => todo!(),
            Stage::Sort(_) => todo!(),
            Stage::SortByCount(_) => todo!(),
            Stage::UnionWith(_) => todo!(),
            Stage::Unset(_) => todo!(),
            Stage::Unwind(_) => todo!(),
        }
    }
}

fn derive_schema_for_literal(literal_value: &LiteralValue) -> Result<Schema> {
    match literal_value {
        LiteralValue::Binary(_) => Ok(Schema::Atomic(Atomic::BinData)),
        LiteralValue::Boolean(_) => Ok(Schema::Atomic(Atomic::Boolean)),
        LiteralValue::DateTime(_) => Ok(Schema::Atomic(Atomic::Date)),
        LiteralValue::DbPointer(_) => Ok(Schema::Atomic(Atomic::DbPointer)),
        LiteralValue::Decimal128(_) => Ok(Schema::Atomic(Atomic::Decimal)),
        LiteralValue::Double(_) => Ok(Schema::Atomic(Atomic::Double)),
        LiteralValue::Int32(_) => Ok(Schema::Atomic(Atomic::Integer)),
        LiteralValue::Int64(_) => Ok(Schema::Atomic(Atomic::Long)),
        LiteralValue::JavaScriptCode(_) => Ok(Schema::Atomic(Atomic::Javascript)),
        LiteralValue::JavaScriptCodeWithScope(_) => Ok(Schema::Atomic(Atomic::JavascriptWithScope)),
        LiteralValue::MaxKey => Ok(Schema::Atomic(Atomic::MaxKey)),
        LiteralValue::MinKey => Ok(Schema::Atomic(Atomic::MinKey)),
        LiteralValue::Null => Ok(Schema::Atomic(Atomic::Null)),
        LiteralValue::ObjectId(_) => Ok(Schema::Atomic(Atomic::ObjectId)),
        LiteralValue::RegularExpression(_) => Ok(Schema::Atomic(Atomic::Regex)),
        LiteralValue::String(_) => Ok(Schema::Atomic(Atomic::String)),
        LiteralValue::Symbol(_) => Ok(Schema::Atomic(Atomic::Symbol)),
        LiteralValue::Timestamp(_) => Ok(Schema::Atomic(Atomic::Timestamp)),
        LiteralValue::Undefined => Ok(Schema::Atomic(Atomic::Undefined)),
    }
}

impl DeriveSchema for Expression {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        state.result_set_schema = promote_missing(&state.result_set_schema);
        match self {
            Expression::Array(ref a) => {
                let array_schema = a
                    .iter()
                    .map(|e| {
                        e.derive_schema(state)
                            .map(|schema| schema.upconvert_missing_to_null())
                    })
                    .collect::<Result<BTreeSet<_>>>()?;
                let array_schema = match array_schema.len() {
                    0 => Schema::Unsat,
                    1 => array_schema.into_iter().next().unwrap(),
                    _ => Schema::AnyOf(array_schema),
                };
                Ok(Schema::Array(Box::new(array_schema)))
            }
            Expression::Document(d) => {
                let (mut keys, mut required) = (BTreeMap::new(), BTreeSet::new());
                for (key, e) in d.iter() {
                    let key_schema = e.derive_schema(state)?;
                    match key_schema.satisfies(&Schema::Missing) {
                        Satisfaction::Not => {
                            required.insert(key.clone());
                            keys.insert(key.clone(), key_schema);
                        }
                        Satisfaction::May => {
                            keys.insert(key.clone(), key_schema);
                        }
                        Satisfaction::Must => (),
                    }
                }
                Ok(Schema::Document(Document {
                    keys,
                    required,
                    ..Default::default()
                }))
            }
            Expression::Literal(ref l) => derive_schema_for_literal(l),
            Expression::Ref(Ref::FieldRef(f)) => {
                let path = f.split(".").map(|s| s.to_string()).collect::<Vec<String>>();
                let schema = get_schema_for_path_mut(&mut state.result_set_schema, path);
                match schema {
                    Some(schema) => Ok(schema.clone()),
                    // Unknown fields actually have the Schema Missing, while unknown variables are
                    // an error.
                    None => Ok(Schema::Missing),
                }
            }
            Expression::Ref(Ref::VariableRef(v)) => match v.as_str() {
                "REMOVE" => Ok(Schema::Missing),
                "ROOT" => Ok(state.result_set_schema.clone()),
                v => match state.variables.get(v) {
                    Some(schema) => Ok(schema.clone()),
                    None => Err(Error::UnknownReference(v.into())),
                },
            },
            Expression::TaggedOperator(op) => op.derive_schema(state),
            Expression::UntaggedOperator(op) => op.derive_schema(state),
        }
    }
}

/// This helper gets the maximal satisfaction of a list of expressions for a given type. This is primarily useful
/// for determining if _any_ argument must be null, or may be null, which can determine the output of an operator.
/// This has similar implications for Decimal, which affects many math ops.
fn arguments_schema_satisfies(
    args: &[&Expression],
    state: &mut ResultSetState,
    schema: &Schema,
) -> Result<Satisfaction> {
    let mut satisfaction = Satisfaction::Not;
    for arg in args.iter() {
        let arg_schema = arg.derive_schema(state)?.upconvert_missing_to_null();
        match (arg_schema.satisfies(schema), satisfaction) {
            (Satisfaction::May, Satisfaction::Not) => {
                satisfaction = Satisfaction::May;
            }
            (Satisfaction::Must, _) => {
                satisfaction = Satisfaction::Must;
            }
            _ => {}
        };
    }
    Ok(satisfaction)
}

/// handle_null_satisfaction captures the behavior of operators that are nullable if any of the arguments are null
/// by checking the nullability of an input schema, and applying that to the default schema.
fn handle_null_satisfaction(
    args: Vec<&Expression>,
    state: &mut ResultSetState,
    non_null_type: Schema,
) -> Result<Schema> {
    match arguments_schema_satisfies(&args, state, &NULLISH)? {
        Satisfaction::Not => Ok(non_null_type),
        Satisfaction::May => Ok(Schema::simplify(&Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            non_null_type
        )))),
        Satisfaction::Must => Ok(Schema::Atomic(Atomic::Null)),
    }
}

impl DeriveSchema for TaggedOperator {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        macro_rules! derive_window_func {
            ($input:expr) => {{
                let input_schema = $input.input.derive_schema(state)?;
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::Null));
                if input_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) != Satisfaction::Not {
                    types.insert(Schema::Atomic(Atomic::Decimal));
                }
                if input_schema.satisfies(&Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                ))) != Satisfaction::Not
                {
                    types.insert(Schema::Atomic(Atomic::Double));
                }
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }};
        }
        macro_rules! derive_date_addition {
            ($input:expr) => {{
                let args = vec![
                    $input.amount.as_ref(),
                    $input.start_date.as_ref(),
                    $input
                        .timezone
                        .as_ref()
                        .map_or(&Expression::Literal(LiteralValue::Boolean(true)), |x| {
                            x.as_ref()
                        }),
                    $input.unit.as_ref(),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }};
        }
        macro_rules! optional_arg_or_truish {
            ($input:expr) => {{
                $input
                    .as_ref()
                    .map_or(&Expression::Literal(LiteralValue::Boolean(true)), |x| {
                        x.as_ref()
                    })
            }};
        }
        match self {
            TaggedOperator::Convert(c) => match c.to.as_ref() {
                Expression::Literal(LiteralValue::String(s)) => Ok(schema_for_type_str(s.as_str())),
                Expression::Literal(LiteralValue::Double(d)) => {
                    Ok(schema_for_type_numeric(*d as i32))
                }
                Expression::Literal(LiteralValue::Int32(i)) => Ok(schema_for_type_numeric(*i)),
                Expression::Literal(LiteralValue::Int64(i)) => {
                    Ok(schema_for_type_numeric(*i as i32))
                }
                Expression::Literal(LiteralValue::Decimal128(d)) => {
                    let decimal_string = d.to_string();
                    let decimal_as_double = decimal_string
                        .parse::<f64>()
                        .map_err(|_| Error::InvalidConvertTypeValue(decimal_string))?;
                    Ok(schema_for_type_numeric(decimal_as_double as i32))
                }
                _ => unreachable!(),
            },
            TaggedOperator::DenseRank(_)
            | TaggedOperator::DocumentNumber(_)
            | TaggedOperator::Rank(_) => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ))),
            TaggedOperator::Derivative(d) => derive_window_func!(d),
            TaggedOperator::ExpMovingAvg(e) => derive_window_func!(e),
            TaggedOperator::Median(m) => handle_null_satisfaction(
                vec![m.input.as_ref()],
                state,
                Schema::Atomic(Atomic::Double),
            ),
            TaggedOperator::Percentile(p) => handle_null_satisfaction(
                vec![p.input.as_ref()],
                state,
                Schema::Atomic(Atomic::Double),
            ),
            TaggedOperator::RegexFind(_) => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Null),
                Schema::Document(Document {
                    keys: map! {
                        "match".to_string() => Schema::Atomic(Atomic::String),
                        "idx".to_string() => Schema::Atomic(Atomic::Integer),
                        "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
                    },
                    required: set!(),
                    ..Default::default()
                })
            ))),
            TaggedOperator::RegexFindAll(_) => {
                Ok(Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "match".to_string() => Schema::Atomic(Atomic::String),
                        "idx".to_string() => Schema::Atomic(Atomic::Integer),
                        "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
                    },
                    required: set!(),
                    ..Default::default()
                }))))
            }
            TaggedOperator::LTrim(t) | TaggedOperator::RTrim(t) | TaggedOperator::Trim(t) => {
                handle_null_satisfaction(
                    vec![t.input.as_ref(), optional_arg_or_truish!(t.chars)],
                    state,
                    Schema::Atomic(Atomic::String),
                )
            }
            TaggedOperator::DayOfWeek(d)
            | TaggedOperator::DayOfMonth(d)
            | TaggedOperator::DayOfYear(d)
            | TaggedOperator::IsoDayOfWeek(d)
            | TaggedOperator::IsoWeek(d)
            | TaggedOperator::IsoWeekYear(d)
            | TaggedOperator::Week(d)
            | TaggedOperator::Month(d)
            | TaggedOperator::Year(d)
            | TaggedOperator::Hour(d)
            | TaggedOperator::Minute(d)
            | TaggedOperator::Second(d)
            | TaggedOperator::Millisecond(d) => handle_null_satisfaction(
                vec![d.date.as_ref(), optional_arg_or_truish!(d.timezone)],
                state,
                Schema::Atomic(Atomic::Integer),
            ),
            TaggedOperator::DateFromParts(d) => {
                let args = vec![
                    optional_arg_or_truish!(d.year),
                    optional_arg_or_truish!(d.month),
                    optional_arg_or_truish!(d.day),
                    optional_arg_or_truish!(d.hour),
                    optional_arg_or_truish!(d.minute),
                    optional_arg_or_truish!(d.second),
                    optional_arg_or_truish!(d.millisecond),
                    optional_arg_or_truish!(d.iso_day_of_week),
                    optional_arg_or_truish!(d.iso_week),
                    optional_arg_or_truish!(d.iso_week_year),
                    optional_arg_or_truish!(d.timezone),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }
            TaggedOperator::DateFromString(d) => {
                let nullable_args = vec![
                    d.date_string.as_ref(),
                    optional_arg_or_truish!(d.format),
                    optional_arg_or_truish!(d.timezone),
                ];
                let on_null_schema = d
                    .on_null
                    .as_ref()
                    .map(|x| x.derive_schema(state))
                    .unwrap_or(Ok(Schema::Atomic(Atomic::Null)))?;
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::Date));
                match arguments_schema_satisfies(
                    &nullable_args,
                    state,
                    &Schema::Atomic(Atomic::Null),
                )? {
                    Satisfaction::Must => {
                        return Ok(on_null_schema);
                    }
                    Satisfaction::May => {
                        types.insert(on_null_schema);
                    }
                    _ => {}
                };
                if let Some(error_schema) = d.on_error.as_ref().map(|x| x.derive_schema(state)) {
                    types.insert(error_schema?);
                }
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }
            TaggedOperator::DateToParts(d) => {
                let args = vec![d.date.as_ref(), optional_arg_or_truish!(d.timezone)];
                match d.iso8601 {
                    Some(true) => handle_null_satisfaction(
                        args,
                        state,
                        Schema::Document(Document {
                            keys: map! {
                                "isoWeekYear".to_string() => Schema::Atomic(Atomic::Integer),
                                "isoWeek".to_string() => Schema::Atomic(Atomic::Integer),
                                "isoDayOfWeek".to_string() => Schema::Atomic(Atomic::Integer),
                                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                                "second".to_string() => Schema::Atomic(Atomic::Integer),
                                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            required: set!(
                                "isoWeekYear".to_string(),
                                "isoWeek".to_string(),
                                "isoDayOfWeek".to_string(),
                                "hour".to_string(),
                                "minute".to_string(),
                                "second".to_string(),
                                "millisecond".to_string()
                            ),
                            ..Default::default()
                        }),
                    ),
                    _ => handle_null_satisfaction(
                        args,
                        state,
                        Schema::Document(Document {
                            keys: map! {
                                "year".to_string() => Schema::Atomic(Atomic::Integer),
                                "month".to_string() => Schema::Atomic(Atomic::Integer),
                                "day".to_string() => Schema::Atomic(Atomic::Integer),
                                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                                "second".to_string() => Schema::Atomic(Atomic::Integer),
                                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            required: set!(
                                "year".to_string(),
                                "month".to_string(),
                                "day".to_string(),
                                "hour".to_string(),
                                "minute".to_string(),
                                "second".to_string(),
                                "millisecond".to_string()
                            ),
                            ..Default::default()
                        }),
                    ),
                }
            }
            TaggedOperator::DateToString(d) => {
                let nullable_args = vec![
                    d.date.as_ref(),
                    optional_arg_or_truish!(d.format),
                    optional_arg_or_truish!(d.timezone),
                ];
                let on_null_schema = d
                    .on_null
                    .as_ref()
                    .map(|x| x.derive_schema(state))
                    .unwrap_or(Ok(Schema::Atomic(Atomic::Null)))?;
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::String));
                match arguments_schema_satisfies(
                    &nullable_args,
                    state,
                    &Schema::Atomic(Atomic::Null),
                )? {
                    Satisfaction::Must => {
                        return Ok(on_null_schema);
                    }
                    Satisfaction::May => {
                        types.insert(on_null_schema);
                    }
                    _ => {}
                };
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }
            TaggedOperator::DateAdd(d) => derive_date_addition!(d),
            TaggedOperator::DateSubtract(d) => derive_date_addition!(d),
            TaggedOperator::DateDiff(d) => {
                let args = vec![
                    d.start_date.as_ref(),
                    d.end_date.as_ref(),
                    optional_arg_or_truish!(d.timezone),
                    d.unit.as_ref(),
                    optional_arg_or_truish!(d.start_of_week),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }
            TaggedOperator::DateTrunc(d) => {
                let args = vec![
                    d.date.as_ref(),
                    d.unit.as_ref(),
                    optional_arg_or_truish!(d.timezone),
                    optional_arg_or_truish!(d.bin_size),
                    optional_arg_or_truish!(d.start_of_week),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }
            TaggedOperator::SortArray(s) => s.input.derive_schema(state),
            TaggedOperator::Let(l) => {
                // we create a copy of the underlying result set state, then add the vars to that.
                // this allows us to temporarily overwrite any variables from the top level if they are defined in
                // both places, and result set state remains unchanged for future operations.
                let mut variables = state.variables.clone();
                let mut let_state_variables = l
                    .vars
                    .iter()
                    .map(|(key, value)| {
                        value
                            .derive_schema(state)
                            .map(|schema| (key.clone(), schema))
                    })
                    .collect::<Result<BTreeMap<String, Schema>>>()?;
                variables.append(&mut let_state_variables);
                let mut let_state = ResultSetState {
                    result_set_schema: state.result_set_schema.clone(),
                    catalog: state.catalog,
                    variables,
                    null_behavior: Satisfaction::Not,
                };
                l.inside.derive_schema(&mut let_state)
            }
            TaggedOperator::GetField(g) => {
                let mut input_schema = g.input.derive_schema(state)?;
                let field_schema =
                    get_schema_for_path_mut(&mut input_schema, vec![g.field.clone()]);
                match field_schema {
                    None => Ok(Schema::Missing),
                    Some(schema) => Ok(schema.clone()),
                }
            }
            TaggedOperator::SetField(s) => {
                // set field does not update the underlying result set schema, but rather, gets the schema
                // of the input, modifies that, and returns it. Thus, we copy the input schema and modify that accordingly.
                let mut input_schema = s.input.derive_schema(state)?;
                let value_schema = s.value.derive_schema(state)?;
                let field_schema =
                    get_schema_for_path_mut(&mut input_schema, vec![s.field.clone()]);
                match field_schema {
                    // if we are setting a new field, add it in appropriately, unless its missing (no-op)
                    None => {
                        if value_schema != Schema::Missing {
                            let new_field = Schema::Document(Document {
                                keys: map! {
                                    s.field.clone() => value_schema,
                                },
                                required: set!(s.field.clone()),
                                ..Default::default()
                            });
                            Ok(input_schema.union(&new_field))
                        } else {
                            Ok(input_schema)
                        }
                    }
                    // if we are handling a new field, check first if the schema is missing (could either be
                    // cause by setting to missing, or setting ot $$REMOVE). Remove it or set it to the new type accordingly
                    Some(field_schema) => {
                        match value_schema {
                            Schema::Missing => {
                                remove_field(&mut input_schema, vec![s.field.clone()]);
                            }
                            _ => {
                                *field_schema = value_schema;
                            }
                        }
                        Ok(input_schema)
                    }
                }
            }
            TaggedOperator::UnsetField(u) => {
                // note: this is functionally the same as $setField with Schema::Missing or $$REMOVE
                let mut input_schema = u.input.derive_schema(state)?;
                remove_field(&mut input_schema, vec![u.field.clone()]);
                Ok(input_schema)
            }
            TaggedOperator::Accumulator(_) => todo!(),
            TaggedOperator::Bottom(b) => b.output.derive_schema(state),
            TaggedOperator::BottomN(b) => {
                Ok(Schema::Array(Box::new(b.output.derive_schema(state)?)))
            }
            TaggedOperator::Cond(c) => {
                let then_schema = c.then.derive_schema(state)?;
                let else_schema = c.r#else.derive_schema(state)?;
                Ok(Schema::simplify(&Schema::AnyOf(
                    set! {then_schema, else_schema},
                )))
            }
            TaggedOperator::Filter(_) => todo!(),
            TaggedOperator::FirstN(_) => todo!(),
            TaggedOperator::Function(_) => todo!(),
            TaggedOperator::Integral(_) => todo!(),
            TaggedOperator::LastN(_) => todo!(),
            TaggedOperator::Like(_) => todo!(),
            TaggedOperator::Map(m) => {
                let var = m._as.clone();
                let var = var.unwrap_or("this".to_string());
                let input_schema = m.input.derive_schema(state)?;
                let array_schema = match input_schema {
                    Schema::Array(a) => *a,
                    _ => {
                        return Err(Error::InvalidExpressionForField(
                            format!("{:?}", m.input),
                            "input",
                        ))
                    }
                };
                let mut new_state = state.clone();
                let mut variables = state.variables.clone();
                variables.insert(var, array_schema);
                new_state.variables = variables;
                Ok(
                    Schema::Array(Box::new(m.inside.derive_schema(&mut new_state)?))
                        .upconvert_missing_to_null(),
                )
            }
            // Unfortunately, unlike $max and $min, $maxN and $minN cannot
            // reduce the scope of the result Schema beyond Array(InputSchema)
            // because doing so would require knowing all the data.
            TaggedOperator::MaxNArrayElement(m) => {
                Ok(Schema::Array(Box::new(m.input.derive_schema(state)?)))
            }
            TaggedOperator::MinNArrayElement(m) => {
                Ok(Schema::Array(Box::new(m.input.derive_schema(state)?)))
            }
            TaggedOperator::Reduce(r) => {
                let input_schema = r.input.derive_schema(state)?;
                let array_schema = match input_schema {
                    Schema::Array(a) => *a,
                    _ => {
                        return Err(Error::InvalidExpressionForField(
                            format!("{:?}", r.input),
                            "input",
                        ))
                    }
                };
                let initial_schema = r.initial_value.derive_schema(state)?;
                let mut new_state = state.clone();
                let mut variables = state.variables.clone();
                variables.insert("this".to_string(), array_schema);
                variables.insert("value".to_string(), initial_schema);
                new_state.variables = variables;
                r.inside.derive_schema(&mut new_state)
            }
            TaggedOperator::Regex(_) => Ok(Schema::Atomic(Atomic::Integer)),
            TaggedOperator::ReplaceAll(_) => todo!(),
            TaggedOperator::ReplaceOne(_) => todo!(),
            TaggedOperator::Shift(_) => todo!(),
            TaggedOperator::Subquery(_) => todo!(),
            TaggedOperator::SubqueryComparison(_) => todo!(),
            TaggedOperator::SubqueryExists(_) => todo!(),
            TaggedOperator::Switch(_) => todo!(),
            TaggedOperator::Top(t) => t.output.derive_schema(state),
            TaggedOperator::TopN(t) => Ok(Schema::Array(Box::new(t.output.derive_schema(state)?))),
            TaggedOperator::Zip(z) => {
                let inputs = match z.inputs.as_ref() {
                    Expression::Array(a) => a,
                    exp => {
                        return Err(Error::InvalidExpressionForField(
                            format!("{:?}", exp),
                            "inputs",
                        ))
                    }
                };
                let mut array_schema = Schema::Unsat;
                for input in inputs.iter() {
                    let input_schema = input.derive_schema(state)?;
                    array_schema = array_schema.union(&input_schema);
                }
                if let Some(defaults) = z.defaults.as_ref() {
                    let defaults_schema = defaults.derive_schema(state)?;
                    if matches!(defaults_schema, Schema::Array(_)) {
                        array_schema = array_schema.union(&defaults_schema);
                    }
                }
                Ok(Schema::Array(Box::new(array_schema)))
            }
            TaggedOperator::SQLConvert(_) | TaggedOperator::SQLDivide(_) => {
                Err(Error::InvalidTaggedOperator(self.clone()))
            }
        }
    }
}

/// get_input_schema will take in the arguments of an untagged operator and return all possible types
/// of any input. there are two primary uses for this -- first, it allows us to easily determine if any
/// argument is null or nullable without inspecting the whole list. It similarly allows us to work with
/// numerics more easily, where (amongst other things) Decimal is often handled differently.
fn get_input_schema(args: &[&Expression], state: &mut ResultSetState) -> Result<Schema> {
    let x = args
        .iter()
        .map(|e| {
            e.derive_schema(state)
                .map(|schema| schema.upconvert_missing_to_null())
        })
        .collect::<Result<BTreeSet<_>>>()?;
    Ok(Schema::simplify(&Schema::AnyOf(x)))
}

/// get_decimal_double_or_nullish handles one of the most common cases for math untagged operators,
/// which is that if any of the inputs is decimal, the operator returns a decimal; if there are any
/// other numeric types, they will return a double; and the operator is nullable, so must handle Null satisfaction.
fn get_decimal_double_or_nullish(
    args: Vec<&Expression>,
    state: &mut ResultSetState,
) -> Result<Schema> {
    let decimal_satisfaction =
        arguments_schema_satisfies(&args, state, &Schema::Atomic(Atomic::Decimal))?;
    let numeric_satisfaction = arguments_schema_satisfies(
        &args,
        state,
        &Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        )),
    )?;
    let schema = match (decimal_satisfaction, numeric_satisfaction) {
        (Satisfaction::Must, _) | (Satisfaction::May, Satisfaction::Not) => {
            Schema::Atomic(Atomic::Decimal)
        }
        (_, Satisfaction::Must) | (Satisfaction::Not, Satisfaction::May) => {
            Schema::Atomic(Atomic::Double)
        }
        (Satisfaction::May, Satisfaction::May) => Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal)
        )),
        _ => Schema::Atomic(Atomic::Null),
    };
    handle_null_satisfaction(args, state, schema)
}

impl DeriveSchema for UntaggedOperator {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        // numeric_filter is a helper that takes a schema and a set of schemas. If the schema is
        // numeric it will retain all the numeric schemas from the set of schemas, otherwise it
        // will return that original schema. This is useful for $min and $max.
        let numeric_filter = |schema: Schema, schemas: BTreeSet<Schema>| {
            if schema.satisfies(&NUMERIC) == Satisfaction::Must {
                let out_schemas = schemas
                    .into_iter()
                    .filter(|s| {
                        matches!(
                            s,
                            Schema::Atomic(
                                Atomic::Integer | Atomic::Long | Atomic::Double | Atomic::Decimal
                            )
                        )
                    })
                    .collect::<BTreeSet<_>>();
                if out_schemas.len() == 1 {
                    out_schemas.into_iter().next().unwrap()
                } else {
                    Schema::AnyOf(out_schemas)
                }
            } else {
                schema
            }
        };
        let mut args = self.args.iter().collect();
        match self.op {
            // no-ops
            UntaggedOperatorName::Abs | UntaggedOperatorName::Ceil | UntaggedOperatorName::Floor | UntaggedOperatorName::ReverseArray | UntaggedOperatorName::Round | UntaggedOperatorName::SampleRate | UntaggedOperatorName::Slice
            | UntaggedOperatorName::Trunc
            // We cannot know anything about the Schema change from the set difference, since it only
            // removes values not types. The best we can do is keep the lhs Schema, which may
            // be overly broad.
            | UntaggedOperatorName::SetDifference => self.args[0].derive_schema(state),
            // operators returning constants
            UntaggedOperatorName::AllElementsTrue | UntaggedOperatorName::AnyElementTrue | UntaggedOperatorName::And | UntaggedOperatorName::Eq | UntaggedOperatorName::Gt | UntaggedOperatorName::Gte | UntaggedOperatorName::In
            | UntaggedOperatorName::IsArray | UntaggedOperatorName::IsNumber | UntaggedOperatorName::Lt | UntaggedOperatorName::Lte | UntaggedOperatorName::Not | UntaggedOperatorName::Ne | UntaggedOperatorName::Or
            | UntaggedOperatorName::SetEquals | UntaggedOperatorName::SetIsSubset => Ok(Schema::Atomic(Atomic::Boolean)),
            UntaggedOperatorName::BinarySize | UntaggedOperatorName::Cmp | UntaggedOperatorName::Strcasecmp | UntaggedOperatorName::StrLenBytes | UntaggedOperatorName::StrLenCP => {
                Ok(Schema::Atomic(Atomic::Integer))
            }
            UntaggedOperatorName::Count => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ))),
            UntaggedOperatorName::Range => Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
            UntaggedOperatorName::Rand => Ok(Schema::Atomic(Atomic::Double)),
            UntaggedOperatorName::Substr | UntaggedOperatorName::SubstrBytes | UntaggedOperatorName::SubstrCP | UntaggedOperatorName::ToLower | UntaggedOperatorName::ToUpper | UntaggedOperatorName::Type => {
                Ok(Schema::Atomic(Atomic::String))
            }
            UntaggedOperatorName::ToHashedIndexKey => Ok(Schema::Atomic(Atomic::Long)),
            // Ops that return a constant schema but must handle nullability
            UntaggedOperatorName::BsonSize | UntaggedOperatorName::IndexOfArray | UntaggedOperatorName::IndexOfBytes | UntaggedOperatorName::IndexOfCP | UntaggedOperatorName::Size | UntaggedOperatorName::ToInt => {
                handle_null_satisfaction(
                    args, state,
                    Schema::Atomic(Atomic::Integer),
                )
            }
            UntaggedOperatorName::Concat | UntaggedOperatorName::Split | UntaggedOperatorName::ToString => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::String),
            ),
            UntaggedOperatorName::TSIncrement | UntaggedOperatorName::TSSecond | UntaggedOperatorName::ToLong => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Long),
            ),
            UntaggedOperatorName::ToBool => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Boolean),
            ),
            UntaggedOperatorName::ToDate => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Date),
            ),
            UntaggedOperatorName::ToDecimal => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Decimal),
            ),
            UntaggedOperatorName::ToDouble => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Double),
            ),
            UntaggedOperatorName::ToObjectId => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::ObjectId),
            ),
            // these operators can only return a decimal (if the input is a decimal), double for any other numeric input, or nullish.
            UntaggedOperatorName::Acos | UntaggedOperatorName::Acosh | UntaggedOperatorName::Asin | UntaggedOperatorName::Asinh | UntaggedOperatorName::Atan | UntaggedOperatorName::Atan2 | UntaggedOperatorName::Atanh | UntaggedOperatorName::Avg
            | UntaggedOperatorName::Cos | UntaggedOperatorName::Cosh | UntaggedOperatorName::DegreesToRadians | UntaggedOperatorName::Divide | UntaggedOperatorName::Exp | UntaggedOperatorName::Ln | UntaggedOperatorName::Log
            | UntaggedOperatorName::Log10 | UntaggedOperatorName::RadiansToDegrees | UntaggedOperatorName::Sin | UntaggedOperatorName::Sinh | UntaggedOperatorName::Sqrt | UntaggedOperatorName::Tan | UntaggedOperatorName::Tanh =>
                get_decimal_double_or_nullish(args, state)
            ,
            // if any of the args are long, long; otherwise int. Int, long only possible types
            UntaggedOperatorName::BitAnd | UntaggedOperatorName::BitNot | UntaggedOperatorName::BitOr | UntaggedOperatorName::BitXor => {
                let non_null_schema = match arguments_schema_satisfies(&args, state, &Schema::Atomic(Atomic::Long))? {
                    Satisfaction::Must => Schema::Atomic(Atomic::Long),
                    Satisfaction::May => Schema::AnyOf(set!(
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::Integer),
                    )),
                    _ => Schema::Atomic(Atomic::Integer),
                };
                Ok(match state.null_behavior {
                    Satisfaction::Not => non_null_schema,
                    Satisfaction::May => Schema::simplify(&Schema::AnyOf(set!(Schema::Atomic(Atomic::Null), non_null_schema))),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null)
                })
            }
            UntaggedOperatorName::Add => {
                let input_schema = get_input_schema(&args, state)?;
                if input_schema.satisfies(&INTEGER_OR_NULLISH) == Satisfaction::Must {
                    // If both are (nullable) Ints, the result is (nullable) Int or Long
                    handle_null_satisfaction(args, state, INTEGRAL.clone())
                } else if input_schema.satisfies(&INTEGER_LONG_OR_NULLISH) == Satisfaction::Must {
                    // If both are (nullable) Ints or Longs, the result is (nullable) Long
                    handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Long))
                } else if input_schema.satisfies(&Schema::Atomic(Atomic::Date)) == Satisfaction::May {
                    // TODO: this is not exactly correct. Consider { $add: ["$date_or_int", "$int"] }
                    //  - this could be a Date when the first is a date
                    //  - but could also be an Int when the first is an int
                    //  - I think we need to analyze the args separately to get the most accurate schema.
                    handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
                } else {
                    get_decimal_double_or_nullish(args, state)
                }
            }
            UntaggedOperatorName::Subtract => {
                let input_schema = get_input_schema(&args, state)?;
                if input_schema.satisfies(&DATE_OR_NULLISH) == Satisfaction::Must {
                    // If both are (nullable) Dates, the result is (nullable) Long
                    handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Long))
                } else if input_schema.satisfies(&DATE_OR_NULLISH) == Satisfaction::May && input_schema.satisfies(&NUMERIC_OR_NULLISH) == Satisfaction::May {
                    // If only one is a (nullable) Date and the other is a (nullable) number,
                    // the result is a (nullable) Date
                    // TODO: what about { $subtract: ["$date", "$date_or_int"] }
                    //   - this could be a Long when it is Date-Date
                    //   - this could be a Date when it is Date-Int
                    handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
                } else if input_schema.satisfies(&INTEGER_OR_NULLISH) == Satisfaction::Must {
                    handle_null_satisfaction(args, state, INTEGRAL.clone())
                } else if input_schema.satisfies(&INTEGER_LONG_OR_NULLISH) == Satisfaction::Must {
                    handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Long))
                } else {
                    get_decimal_double_or_nullish(args, state)
                }
            }
            // int + int -> int or long; int + long, long + long -> long,
            UntaggedOperatorName::Multiply => {
                let input_schema = get_input_schema(&args, state)?;
                if input_schema.satisfies(&Schema::Atomic(Atomic::Integer)) == Satisfaction::Must {
                    Ok(INTEGRAL.clone())
                } else if input_schema.satisfies(&INTEGRAL) == Satisfaction::Must {
                    Ok(Schema::Atomic(Atomic::Long))
                } else {
                    get_decimal_double_or_nullish(args, state)
                }
            }
            // window function operators
            UntaggedOperatorName::CovariancePop | UntaggedOperatorName::CovarianceSamp | UntaggedOperatorName::StdDevPop | UntaggedOperatorName::StdDevSamp => {
                let input_schema = get_input_schema(&args, state)?;
                // window function operators can return null, even if the data is not null, based on the window
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::Null));
                if input_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) != Satisfaction::Not {
                    types.insert(Schema::Atomic(Atomic::Decimal));
                }
                // double for any numeric other than Decimal
                if input_schema.satisfies(&Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                ))) != Satisfaction::Not
                {
                    types.insert(Schema::Atomic(Atomic::Double));
                }
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }
            // pow will return the maximal numeric type of its inputs; integrals are lumped together
            // because an int ^ int can return a long
            UntaggedOperatorName::Pow => {
                let input_schema = get_input_schema(&args, state)?;
                let schema = if input_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) != Satisfaction::Not {
                    Schema::Atomic(Atomic::Decimal)
                } else if input_schema.satisfies(&Schema::Atomic(Atomic::Double))
                    != Satisfaction::Not
                {
                    Schema::Atomic(Atomic::Double)
                } else if input_schema.satisfies(&Schema::Atomic(Atomic::Long)) != Satisfaction::Not {
                    Schema::Atomic(Atomic::Long)
                } else {
                    Schema::AnyOf(set!(
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Long),
                    ))
                };
                handle_null_satisfaction(args, state, schema)
            }
            // mod returns the maximal numeric type of its inputs
            UntaggedOperatorName::Mod => {
                let input_schema = get_input_schema(&args, state)?;
                let schema = if input_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) != Satisfaction::Not {
                    Schema::Atomic(Atomic::Decimal)
                } else if input_schema.satisfies(&Schema::Atomic(Atomic::Double))
                    != Satisfaction::Not
                {
                    Schema::Atomic(Atomic::Double)
                } else if input_schema.satisfies(&Schema::Atomic(Atomic::Long)) != Satisfaction::Not
                {
                    Schema::Atomic(Atomic::Long)
                } else {
                    Schema::Atomic(Atomic::Integer)
                };
                handle_null_satisfaction(args, state, schema)
            }
            UntaggedOperatorName::ArrayElemAt => {
                let input_schema = self.args[0].derive_schema(state)?;
                match input_schema {
                    Schema::Array(a) => Ok(a.as_ref().clone()),
                    _ => {
                        if input_schema.satisfies(&NULLISH_OR_UNDEFINED) == Satisfaction::Must {
                            Ok(Schema::Atomic(Atomic::Null))
                        } else {
                            Err(Error::InvalidType(input_schema, 0usize))
                        }
                    }
                }
            }
            UntaggedOperatorName::ArrayToObject => {
                // We could only know the keys, if we have the entire array.
                // We may consider making this more precise for array literals.
                Ok(Schema::Document(Document::any()))
            }
            UntaggedOperatorName::ConcatArrays | UntaggedOperatorName::SetUnion => {
                let mut array_schema = Schema::Unsat;
                for (i, arg) in self.args.iter().enumerate() {
                    let schema = arg.derive_schema(state)?;
                    match schema {
                        Schema::Array(a) => array_schema = array_schema.union(a.as_ref()),
                        _ => return Err(Error::InvalidType(schema, i)),
                    };
                }
                Ok(Schema::Array(Box::new(array_schema)))
            }
            UntaggedOperatorName::SetIntersection => {
                if args.is_empty() {
                    return Ok(Schema::Array(Box::new(Schema::Unsat)));
                }
                let mut array_schema = match args.remove(0_usize).derive_schema(state)?{
                    Schema::Array(a) => *a,
                    Schema::Missing => return Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
                    schema => return Err(Error::InvalidType(schema, 0)),
                };
                for (i, arg) in self.args.iter().enumerate() {
                    let schema = arg.derive_schema(state)?;
                    match schema {
                        Schema::Array(a) => {
                            // If the array_schema MAY satisify numeric, we need to augment the rhs
                            // of the intersection with the entire numeric Schema set because 42.0
                            // as a double is considered equivalent to 42 as an integer in mongo,
                            // meaning that intersection of [42.0] and [42] is [42.0]. Note
                            // specifically that Mongo retains the lhs value when there is
                            // equivalent numeric values in the rhs. This is why we pull out the
                            // first Schema individually before the loop.
                            let a = if a.satisfies(&NUMERIC) >= Satisfaction::May {
                                a.union(&NUMERIC)
                            } else {
                                *a
                            };
                            array_schema = array_schema.intersection(&a)
                        }
                        Schema::Missing => return Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
                        _ => return Err(Error::InvalidType(schema, i+1)),
                    };
                }
                Ok(Schema::Array(Box::new(array_schema)))
            }
            UntaggedOperatorName::Locf => {
                self.args[0].derive_schema(state)
            }
            UntaggedOperatorName::Max => {
                let schema = self.args[0].derive_schema(state)?;
                let schema = match schema {
                    Schema::AnyOf(a) => {
                        // Unsat should be impossible, since we should never see AnyOf(empty_set)
                        let schema = a.iter().max().unwrap_or(&Schema::Unsat);
                        numeric_filter(schema.clone(), a)
                    }
                    _ => schema,
                };
                Ok(schema)
            }
            UntaggedOperatorName::Min => {
                let schema = self.args[0].derive_schema(state)?;
                let schema = match schema {
                    Schema::AnyOf(a) => {
                        // Unsat should be impossible, since we should never see AnyOf(empty_set)
                        let schema = a.iter().min().unwrap_or(&Schema::Unsat);
                        numeric_filter(schema.clone(), a)
                    }
                    _ => schema,
                };
                Ok(schema)
            }
            UntaggedOperatorName::AddToSet | UntaggedOperatorName::Push => {
                let schema = self.args[0].derive_schema(state)?;
                Ok(Schema::Array(Box::new(schema)))
            }
            UntaggedOperatorName::IfNull => {
                // Note that $ifNull is variadic, not binary. It returns the first
                // non-null argument. If all arguments are nullish, it returns the
                // last argument unmodified (meaning, even if the last argument is
                // null or missing, that value is returned). Therefore, the schema
                // for this expression is the union of all argument schemas up to
                // the first non-nullish argument (minus the nullish types).
                //
                // If all arguments up to the last one are nullish, then it is the
                // union of all argument schemas. The schema retains any nullish
                // types that the last argument may satisfy.
                let mut schema = Schema::Unsat;
                let last_elem_idx = args.len() - 1;
                for (i, arg) in args.into_iter().enumerate() {
                    let arg_schema = arg.derive_schema(state)?;
                    if i == last_elem_idx {
                        // If we get to the last element, we do not want to remove
                        // nullish types from this argument's schema since $ifNull
                        // returns this value no matter what.
                        schema = schema.union(&arg_schema);
                        break;
                    }

                    match arg_schema.satisfies(&NULLISH) {
                        // If this argument is never nullish, include this schema
                        // and break.
                        Satisfaction::Not => {
                            schema = schema.union(&arg_schema);
                            break;
                        }
                        // If this argument may be nullish, retain the non-nullish
                        // types only.
                        Satisfaction::May => {
                            schema = schema.union(&arg_schema.subtract_nullish());
                        }
                        // If this argument must be nullish, ignore it
                        Satisfaction::Must => {}
                    }
                }

                Ok(schema)
            }
            UntaggedOperatorName::MergeObjects => {
                // $mergeObjects ignores nullish arguments. If all arguments are
                // nullish, then the result is the empty document schema. It is
                // tempting to simply use document schema union to represent the
                // result schema for this operator, however that is not exactly
                // correct. $mergeObjects uses the last value for a key if it
                // appears multiple times. See the tests for examples.
                let arg_schemas: Result<Vec<Document>> = args.iter().filter_map(|arg| {
                    let arg_schema = arg.derive_schema(state);
                    match arg_schema {
                        Err(e) => Some(Err(e)),
                        Ok(arg_schema) => {
                            fn retain_only_doc_schemas(sch: Schema) -> Option<Schema> {
                                match sch {
                                    Schema::Unsat => None,
                                    Schema::Missing => None,
                                    Schema::Atomic(_) => None,
                                    Schema::Array(_) => None,
                                    Schema::Document(_) => Some(sch),
                                    Schema::Any => Some(ANY_DOCUMENT.clone()),
                                    Schema::AnyOf(ao) => {
                                        // Retain only the Document schemas in this AnyOf and
                                        // union them all together. The presence of any types
                                        // other than Document implies that any "required"
                                        // fields in any Document schemas are not required in
                                        // the resulting schema. This is achieved by starting
                                        // the fold with EMPTY_DOCUMENT. If the AnyOf only
                                        // contains Document schemas, then some fields in the
                                        // result schema may actually be required. This is
                                        // achieved by starting the fold with the first schema
                                        // from the AnyOf.
                                        if ao.is_empty() {
                                            return None
                                        }
                                        let init_doc_schema = if ao.iter().all(|s| matches!(s, Schema::Document(_))) {
                                            // At this point, we know ao is non-empty and contains
                                            // only document schemas.
                                            ao.first().unwrap().clone()
                                        } else {
                                            EMPTY_DOCUMENT.clone()
                                        };

                                        Some(ao.into_iter()
                                            .filter_map(retain_only_doc_schemas)
                                            .fold(init_doc_schema, Schema::document_union))
                                    }
                                }
                            }

                            match retain_only_doc_schemas(arg_schema) {
                                Some(Schema::Document(d)) => Some(Ok(d)),
                                _ => None,
                            }
                        }
                    }
                }).collect();

                Ok(Schema::simplify(&Schema::Document(arg_schemas?
                    .into_iter()
                    .fold(Document::empty(), |acc, arg_schema| {
                        // Generally, mergeObjects retains the last value seen
                        // for a key. Therefore, we iterate through the keys
                        // of this argument and insert them and their schemas.
                        let mut keys = acc.keys;
                        for (arg_key, mut arg_key_schema) in arg_schema.keys {
                            let current_key_schema = keys.get(&arg_key);
                            let schema_to_insert = if let Some(current_key_schema) = current_key_schema {
                                if arg_key_schema.satisfies(&Schema::Missing) == Satisfaction::May {
                                    // If this key already appears in the accumulated schema _and_
                                    // this argument's schema for this key is possibly missing, then
                                    // we cannot simply overwrite the accumulated schema for this
                                    // key. This is because in the case the later document's value
                                    // for this key is missing, the earlier document's value will be
                                    // returned. Therefore, we must union the accumulated schema and
                                    // this argument's schema for this key. See the tests for an
                                    // example.
                                    schema_difference(&mut arg_key_schema, set!{Schema::Missing});
                                    arg_key_schema.union(current_key_schema)
                                } else {
                                    arg_key_schema
                                }
                            } else {
                                arg_key_schema
                            };

                            keys.insert(arg_key, schema_to_insert);
                        }

                        // All required keys must still be required.
                        let mut required = acc.required;
                        required.extend(arg_schema.required);

                        // If any Document allows additional properties, the result
                        // must also allow additional_properties.
                        let additional_properties = acc.additional_properties || arg_schema.additional_properties;

                        Document {
                            keys,
                            required,
                            additional_properties,
                            ..Default::default()
                        }
                    })
                )))
            }
            UntaggedOperatorName::ObjectToArray => {
                let document_value_types = self.args.iter().try_fold(Schema::Unsat, |schema , arg| {
                    let arg_schema = arg.derive_schema(state)?;
                    Ok(match schema {
                        Schema::Unsat => arg_schema,
                        schema => schema.union(&arg_schema)
                    })
                })?;
                let array_type = Schema::Array(Box::new(Schema::Document(Document { keys: map! {
                        "k".to_string() => Schema::Atomic(Atomic::String),
                        "v".to_string() => document_value_types
                    },
                    ..Default::default()
                })));
                Ok(handle_null_satisfaction(vec![args[0]], state, array_type)?)
            }
            _ => Err(Error::InvalidUntaggedOperator(self.op.into())),
        }
    }
}
