use crate::{get_schema_for_path_mut, Error, Result};
use agg_ast::definitions::{
    Expression, LiteralValue, Ref, Stage, TaggedOperator, UntaggedOperator,
};
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema, NULLISH},
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
    pub variables: &'a BTreeMap<String, Schema>,
    pub result_set_schema: Schema,
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
        LiteralValue::Undefined => Err(Error::InvalidLiteralType),
    }
}

impl DeriveSchema for Expression {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        match self {
            Expression::Array(ref a) => Ok(Schema::Array(Box::new(Schema::AnyOf(
                a.iter()
                    .map(|e| {
                        e.derive_schema(state)
                            .map(|schema| schema.upconvert_missing_to_null())
                    })
                    .collect::<Result<BTreeSet<_>>>()?,
            )))),
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
                    None => Err(Error::UnknownReference(f.clone())),
                }
            }
            Expression::Ref(Ref::VariableRef(v)) => match state.variables.get(v) {
                Some(schema) => Ok(schema.clone()),
                None => Err(Error::UnknownReference(v.clone())),
            },
            Expression::TaggedOperator(op) => op.derive_schema(state),
            Expression::UntaggedOperator(op) => op.derive_schema(state),
        }
    }
}

/// This helper gets the maximal satisfaction of a list of expressions for a given type. This is primarily useful
/// for determininig if _any_ argument must be null, or may be null, which can determine the output of an operator.
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
            TaggedOperator::Accumulator(_) => todo!(),
            TaggedOperator::Bottom(_) => todo!(),
            TaggedOperator::BottomN(_) => todo!(),
            TaggedOperator::Convert(_) => todo!(),
            TaggedOperator::Filter(_) => todo!(),
            TaggedOperator::FirstN(_) => todo!(),
            TaggedOperator::Function(_) => todo!(),
            TaggedOperator::GetField(_) => todo!(),
            TaggedOperator::Integral(_) => todo!(),
            TaggedOperator::LastN(_) => todo!(),
            TaggedOperator::Let(_) => todo!(),
            TaggedOperator::Like(_) => todo!(),
            TaggedOperator::Map(_) => todo!(),
            TaggedOperator::MaxNArrayElement(_) => todo!(),
            TaggedOperator::MinNArrayElement(_) => todo!(),
            TaggedOperator::Reduce(_) => todo!(),
            TaggedOperator::Regex(_) => todo!(),
            TaggedOperator::ReplaceAll(_) => todo!(),
            TaggedOperator::ReplaceOne(_) => todo!(),
            TaggedOperator::SetField(_) => todo!(),
            TaggedOperator::Shift(_) => todo!(),
            TaggedOperator::Subquery(_) => todo!(),
            TaggedOperator::SubqueryComparison(_) => todo!(),
            TaggedOperator::SubqueryExists(_) => todo!(),
            TaggedOperator::Switch(_) => todo!(),
            TaggedOperator::Top(_) => todo!(),
            TaggedOperator::TopN(_) => todo!(),
            TaggedOperator::UnsetField(_) => todo!(),
            TaggedOperator::Zip(_) => todo!(),
            TaggedOperator::SqlConvert(_) | TaggedOperator::SqlDivide(_) => {
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
        let args = self.args.iter().collect();
        match self.op.as_str() {
            // no-ops
            "$abs" | "$ceil" | "$floor" | "$reverseArray" | "$round" | "$sampleRate" | "$slice"
            | "$trunc" => self.args[0].derive_schema(state),
            // operators returning constants
            "$allElementsTrue" | "$anyElementTrue" | "$and" | "$eq" | "$gt" | "$gte" | "$in"
            | "$isArray" | "$isNumber" | "$lt" | "$lte" | "$not" | "$ne" | "$or"
            | "$regexMatch" | "$setEquals" | "$setIsSubset" => Ok(Schema::Atomic(Atomic::Boolean)),
            "$binarySize" | "$cmp" | "$strcasecmp" | "$strLenBytes" | "$strLenCP" => {
                Ok(Schema::Atomic(Atomic::Integer))
            }
            "$count" => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ))),
            "$range" => Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
            "$rand" => Ok(Schema::Atomic(Atomic::Double)),
            "$substr" | "$substrBytes" | "$substrCP" | "$toLower" | "$toUpper" | "$type" => {
                Ok(Schema::Atomic(Atomic::String))
            }
            "$toHashedIndexKey" => Ok(Schema::Atomic(Atomic::Long)),
            // Ops that return a constant schema but must handle nullability
            "$bsonSize" | "$indexOfArray" | "$indexOfBytes" | "$indexOfCP" | "$size" | "$toInt" => {
                handle_null_satisfaction(
                    args, state,
                    Schema::Atomic(Atomic::Integer),
                )
            }
            "$concat" | "$split" | "$toString" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::String),
            ),
            "$tsIncrement" | "$tsSecond" | "$toLong" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Long),
            ),
            "$toBool" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Boolean),
            ),
            "$toDate" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Date),
            ),
            "$toDecimal" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Decimal),
            ),
            "$toDouble" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Double),
            ),
            "$toObjectid" => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::ObjectId),
            ),
            // these operators can only return a decimal (if the input is a decimal), double for any other numeric input, or nullish.
            "$acos" | "$acosh" | "$asin" | "$asinh" | "$atan" | "$atan2" | "$atanh" | "$avg"
            | "$cos" | "$cosh" | "$degressToRadians" | "$divide" | "$exp" | "$ln" | "$log"
            | "$log10" | "$radiansToDegrees" | "$sin" | "$sinh" | "$sqrt" | "$tan" | "$tanh" =>
                get_decimal_double_or_nullish(args, state)
            ,
            // if any of the args are long, long; otherwise int. Int, long only possilbe types
            "$bitAnd" | "$bitNot" | "$bitOr" | "$bitXor" => {
                Ok(
                    match arguments_schema_satisfies(&args, state, &Schema::Atomic(Atomic::Long))? {
                        Satisfaction::Must => Schema::Atomic(Atomic::Long),
                        Satisfaction::May => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Long),
                            Schema::Atomic(Atomic::Integer),
                        )),
                        _ => Schema::Atomic(Atomic::Integer),
                    },
                )
            }
            // int + int -> int or long; int + long, long + long -> long,
            "$multiply" => {
                let input_schema = get_input_schema(&args, state)?;
                let integral_types = Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long)
                ));
                if input_schema.satisfies(&Schema::Atomic(Atomic::Integer)) == Satisfaction::Must {
                    Ok(integral_types)
                } else if input_schema.satisfies(&integral_types) == Satisfaction::Must {
                    Ok(Schema::Atomic(Atomic::Long))
                } else {
                    get_decimal_double_or_nullish(args, state)
                }
            }
            // window function operators
            "$covariancePop" | "$covarianceSamp" | "$stdDevPop" | "$stdDevSamp" => {
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
            "$pow" => {
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
            "$mod" => {
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
            _ => Err(Error::InvalidUntaggedOperator(self.op.clone())),
        }
    }
}
