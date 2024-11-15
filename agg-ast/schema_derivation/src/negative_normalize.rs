use agg_ast::{
    definitions::{
        Expression, Let, LiteralValue, MatchBinaryOp, MatchExpr, MatchExpression, MatchField,
        MatchLogical, MatchMisc, MatchNot, MatchNotExpression, MatchRegex, Ref, TaggedOperator,
        UntaggedOperator,
    },
    map,
};
use bson::Bson;
use std::{collections::HashSet, sync::LazyLock};
static DECIMAL_ZERO: LazyLock<bson::Decimal128> = LazyLock::new(|| "0.0".parse().unwrap());

#[allow(dead_code)]
pub(crate) trait NegativeNormalize<T> {
    fn get_negation(&self) -> T;
    #[allow(dead_code)]
    fn get_negative_normal_form(&self) -> T;
}

macro_rules! wrap_in_check {
    ($op:expr, $expr:expr, $val:expr) => {
        Expression::UntaggedOperator(UntaggedOperator {
            op: $op.to_string(),
            args: vec![$expr, $val],
        })
    };
}

macro_rules! wrap_in_zero_check {
    ($expr:expr) => {
        wrap_in_check!("$eq", $expr, Expression::Literal(LiteralValue::Int32(0)))
    };
}

macro_rules! wrap_in_null_or_missing_check {
    ($expr:expr) => {
        wrap_in_check!("$lte", $expr, Expression::Literal(LiteralValue::Null))
    };
}

macro_rules! wrap_in_false_check {
    ($expr:expr) => {
        wrap_in_check!(
            "$eq",
            $expr,
            Expression::Literal(LiteralValue::Boolean(false))
        )
    };
}

impl NegativeNormalize<Expression> for Expression {
    fn get_negative_normal_form(&self) -> Self {
        self.clone()
    }

    fn get_negation(&self) -> Self {
        match self {
            // For now, we will simply avoid handling negation for literals, arrays, and documents.
            // This does potentially limit precision, for example, in cases where we have a $not of a $and/$or.
            // However, we won't be inspecting literal values in those cases anyways, so properly inverting these values
            // would have no effect.
            Expression::Array(_) | Expression::Document(_) | Expression::Literal(_) => self.clone(),
            // to negate a field reference, we should assert that it has falsish behavior
            Expression::Ref(_)
            | Expression::TaggedOperator(TaggedOperator::GetField(_))
            | Expression::TaggedOperator(TaggedOperator::Reduce(_))
            | Expression::TaggedOperator(TaggedOperator::SetField(_))
            | Expression::TaggedOperator(TaggedOperator::Switch(_))
            | Expression::TaggedOperator(TaggedOperator::UnsetField(_)) => {
                Expression::UntaggedOperator(UntaggedOperator {
                    op: "$or".to_string(),
                    args: vec![
                        wrap_in_null_or_missing_check!(self.clone()),
                        wrap_in_zero_check!(self.clone()),
                        wrap_in_false_check!(self.clone()),
                    ],
                })
            }
            Expression::TaggedOperator(t) => match t {
                // The following operators may evaluate to null, so the negation simply asserts
                // that they are less than or equal to null. Although they should never evaluate to
                // missing, we still use the $lte operator for the negation since there are some
                // MQL operators that can evaluate to missing under certain circumstances. At time
                // of writing, none of these operators behave that way but that doesn't preclude
                // them from ever behaving that way so $lte is defensive against such updates in
                // future MongoDB versions.
                TaggedOperator::DateAdd(_)
                | TaggedOperator::DateDiff(_)
                | TaggedOperator::DateFromParts(_)
                | TaggedOperator::DateFromString(_)
                | TaggedOperator::DateSubtract(_)
                | TaggedOperator::DateToParts(_)
                | TaggedOperator::DateToString(_)
                | TaggedOperator::DateTrunc(_)
                | TaggedOperator::Filter(_)
                | TaggedOperator::LTrim(_)
                | TaggedOperator::Map(_)
                | TaggedOperator::RegexFind(_)
                | TaggedOperator::RegexFindAll(_)
                | TaggedOperator::ReplaceAll(_)
                | TaggedOperator::ReplaceOne(_)
                | TaggedOperator::RTrim(_)
                | TaggedOperator::SortArray(_)
                | TaggedOperator::Trim(_)
                | TaggedOperator::Zip(_)
                // these operators will never return nullish values -- instead, they may return empty string,
                // or array, which are truish. Wrapping with null will ensure if they are negated, the schema
                // will evaluate to Unsat
                | TaggedOperator::FirstN(_)
                | TaggedOperator::LastN(_)
                | TaggedOperator::MaxNArrayElement(_)
                | TaggedOperator::MinNArrayElement(_)
                 => wrap_in_null_or_missing_check!(self.clone()),
                // The following operators may evaluate to the falsy values null or 0, so the
                // negation asserts that equality to any of those values.
                TaggedOperator::DayOfMonth(_)
                | TaggedOperator::DayOfWeek(_)
                | TaggedOperator::DayOfYear(_)
                | TaggedOperator::Hour(_)
                | TaggedOperator::Millisecond(_)
                | TaggedOperator::Minute(_)
                | TaggedOperator::Month(_)
                | TaggedOperator::Second(_)
                | TaggedOperator::Week(_)
                | TaggedOperator::Year(_)
                | TaggedOperator::IsoDayOfWeek(_)
                | TaggedOperator::IsoWeek(_)
                | TaggedOperator::IsoWeekYear(_)
                | TaggedOperator::Median(_)
                | TaggedOperator::Percentile(_) => Expression::UntaggedOperator(UntaggedOperator {
                    op: "$or".to_string(),
                    args: vec![
                        wrap_in_null_or_missing_check!(self.clone()),
                        wrap_in_zero_check!(self.clone()),
                    ],
                }),
                // This operator never evaluates to null, only ever true or false. So the negation
                // asserts that it is false.
                TaggedOperator::Regex(_) => wrap_in_false_check!(self.clone()),
                // to get the negation of $let we will simply negate the in statement
                TaggedOperator::Let(l) => {
                    let negated_inside = l.inside.get_negation();
                    Expression::TaggedOperator(TaggedOperator::Let(Let {
                        vars: l.vars.clone(),
                        inside: Box::new(negated_inside)
                    }))
                }
                _ => todo!(),
            },
            Expression::UntaggedOperator(u) => {
                let (op, args) = match u.op.as_str() {
                    "$eq" => ("$ne", u.args.clone()),
                    "$ne" => ("$eq", u.args.clone()),
                    "$lt" => ("$gte", u.args.clone()),
                    "$lte" => ("$gt", u.args.clone()),
                    "$gt" => ("$lte", u.args.clone()),
                    "$gte" => ("$lt", u.args.clone()),
                    "$and" => {
                        let args: Vec<Expression> =
                            u.args.iter().map(|x| x.get_negation()).collect();
                        ("$or", args)
                    }
                    "$or" => {
                        let args: Vec<Expression> =
                            u.args.iter().map(|x| x.get_negation()).collect();
                        ("$and", args)
                    }
                    // The following operators may evaluate to null, so the negation simply asserts
                    // that they are less than or equal to null. See the TaggedOperators section
                    // for further explanation.
                    "$arrayToObject" | "$objectToArray" | "$reverseArray" | "$toDate"
                    | "$toObjectId" | "$toString" | "$tsSecond" | "$tsIncrement" | "$concat"
                    | "$concatArrays" | "$setDifference" | "$setIntersection" | "$setUnion"
                    | "$slice" | "$split"
                    // these operators will never return nullish values -- instead, they may return empty string,
                    // or array, which are truish. Wrapping with null will ensure if they are negated, the schema
                    // will evaluate to Unsat
                    | "$meta" | "$mergeObjects" | "$rand" | "$range" | "$substr" | "$substrBytes" | "$substrCP" | "$toHashedIndexKey" | "$toLower" | "$toUpper" | "$type" => return wrap_in_null_or_missing_check!(self.clone()),
                    // The following operators may evaluate to the falsy values null or 0, so the
                    // negation asserts that equality to any of those values.
                    "$abs" | "$acos" | "$acosh" | "$asin" | "$asinh" | "$atan" | "$atan2"
                    | "$atanh" | "$avg" | "$cos" | "$cosh" | "$degreesToRadians" | "$divide"
                    | "$exp" | "$ln" | "$log" | "$log10" | "$mod" | "$multiply" | "$pow"
                    | "$radiansToDegrees" | "$sin" | "$sinh" | "$sqrt" | "$tan" | "$tanh"
                    | "$trunc" | "$ceil" | "$floor" | "$indexOfArray" | "$indexOfBytes"
                    | "$indexOfCP" | "$toInt" | "$add" | "$subtract" | "$arrayElemAt"
                    | "$binarySize" | "$bitAnd" | "$bitNot" | "$bitOr" | "$bitXor"
                    | "$bsonSize" | "$covariancePop" | "$covarianceSamp" | "$stdDevPop"
                    | "$stdDevSamp" | "$round" | "$toDecimal" | "$toDouble" | "$toLong" => {
                        let null_check = wrap_in_null_or_missing_check!(self.clone());
                        let zero_check = wrap_in_zero_check!(self.clone());
                        ("$or", vec![null_check, zero_check])
                    }
                    // The following operators may evaluate to the falsy values missing, null, 0, or
                    // false, so the negation asserts equality to any of those values.
                    "$first" | "$ifNull" | "$last" | "$literal" | "$max" | "$min" | "$setField"
                    | "$reduce" | "$switch" => {
                        let null_check = wrap_in_null_or_missing_check!(self.clone());
                        let zero_check = wrap_in_zero_check!(self.clone());
                        let false_check = wrap_in_false_check!(self.clone());
                        ("$or", vec![null_check, zero_check, false_check])
                    }
                    // the following operators negation depends on the underlying documents -- thus,
                    // for the sake of schema derivation, they function the same way negated as they do normally
                    "$allElementsTrue" | "$anyElementTrue" | "$cmp" | "$in" | "$size"
                    | "$strLenBytes" | "$strLenCP" | "$strcasecmp" | "$setEquals"
                    | "$setIsSubset" | "$sum" => (u.op.as_str(), u.args.clone()),
                    // toBool is the only untagged op that is boolean or nullish
                    "$toBool" => {
                        let null_check = wrap_in_null_or_missing_check!(self.clone());
                        let false_check = wrap_in_false_check!(self.clone());
                        ("$or", vec![null_check, false_check])
                    }
                    // The following operators only evaluate to boolean, so the negation simply asserts
                    // that they are false
                    "$isArray" | "$isNumber" => (
                        "$eq",
                        vec![
                            self.clone(),
                            Expression::Literal(LiteralValue::Boolean(false)),
                        ],
                    ),
                    // the negation of not(X) is X, so we short circuit here and just return X
                    "$not" => return u.args[0].clone(),
                    op => unreachable!("Cannot negate unknown untagged operator: {op}"),
                };
                Expression::UntaggedOperator(UntaggedOperator {
                    op: op.to_string(),
                    args,
                })
            }
        }
    }
}

impl NegativeNormalize<MatchExpression> for MatchExpression {
    fn get_negative_normal_form(&self) -> Self {
        match self {
            MatchExpression::Expr(MatchExpr { expr }) => match *expr.clone() {
                Expression::UntaggedOperator(untagged_operator) => {
                    match untagged_operator.op.as_str() {
                        "$not" => MatchExpression::Expr(MatchExpr {
                            expr: Box::new(untagged_operator.args[0].get_negation()),
                        }),
                        "$cond" => todo!(),
                        _ => self.clone(),
                    }
                }
                _ => self.clone(),
            },
            MatchExpression::Logical(logical) => logical.get_negative_normal_form(),
            MatchExpression::Field(field) => field.get_negative_normal_form(),
            MatchExpression::Misc(_misc) => todo!(),
        }
    }

    fn get_negation(&self) -> Self {
        match self {
            MatchExpression::Expr(MatchExpr { expr }) => match **expr {
                Expression::TaggedOperator(_) | Expression::UntaggedOperator(_) => {
                    MatchExpression::Expr(MatchExpr {
                        expr: Box::new(expr.get_negation()),
                    })
                }
                // a $match where the expression is a value or field ref should fail in deserialization
                _ => unreachable!("Cannot negate match on non operator expression"),
            },
            MatchExpression::Logical(logical) => logical.get_negation(),
            MatchExpression::Field(field) => field.get_negation(),
            MatchExpression::Misc(_misc) => todo!(),
        }
    }
}

impl NegativeNormalize<MatchExpression> for MatchLogical {
    fn get_negative_normal_form(&self) -> MatchExpression {
        macro_rules! negative_normal_form_logical {
            ($args:expr, $output_sym:path) => {{
                let args: Vec<_> = $args.iter().map(|x| x.get_negative_normal_form()).collect();
                if args.len() == 1 {
                    args.into_iter().next().unwrap()
                } else {
                    MatchExpression::Logical($output_sym(args))
                }
            }};
        }
        match self {
            MatchLogical::And(and) => {
                negative_normal_form_logical!(and, MatchLogical::And)
            }
            MatchLogical::Or(or) => {
                negative_normal_form_logical!(or, MatchLogical::Or)
            }
            MatchLogical::Nor(nor) => {
                let args = nor
                    .iter()
                    .map(|expr| expr.get_negation())
                    .collect::<Vec<MatchExpression>>();
                negative_normal_form_logical!(args, MatchLogical::And)
            }
            MatchLogical::Not(ref not) => match not.expr {
                MatchNotExpression::Regex(_) => {
                    MatchExpression::Logical(MatchLogical::Not(not.clone()))
                }
                MatchNotExpression::Query(ref ops) => {
                    let args = ops
                        .iter()
                        .map(|(op, b)| negate_binary_operator(&not.field, op, b))
                        .collect::<Vec<MatchExpression>>();
                    if args.len() == 1 {
                        args.into_iter().next().unwrap()
                    } else {
                        MatchExpression::Logical(MatchLogical::Or(args))
                    }
                }
            },
        }
    }

    fn get_negation(&self) -> MatchExpression {
        macro_rules! negate_logical {
            ($args:expr, $output_sym:path) => {{
                let args: Vec<_> = $args.iter().map(|x| x.get_negation()).collect();
                if args.len() == 1 {
                    args.into_iter().next().unwrap()
                } else {
                    MatchExpression::Logical($output_sym(args))
                }
            }};
        }
        match self {
            MatchLogical::And(and) => {
                negate_logical!(and, MatchLogical::Or)
            }
            MatchLogical::Or(or) => {
                negate_logical!(or, MatchLogical::And)
            }
            MatchLogical::Nor(nor) => MatchExpression::Logical(MatchLogical::Or(nor.clone())),
            MatchLogical::Not(ref not) => match not.expr {
                MatchNotExpression::Regex(ref b) => {
                    let (pattern, options) = if let Bson::Document(d) = b {
                        (
                            d.get("$regex").cloned().unwrap_or(Bson::Null),
                            d.get("$options").cloned(),
                        )
                    } else {
                        (b.clone(), None)
                    };
                    MatchExpression::Misc(MatchMisc::Regex(MatchRegex {
                        field: not.field.clone(),
                        pattern,
                        options,
                    }))
                }
                MatchNotExpression::Query(ref ops) => MatchExpression::Field(MatchField {
                    field: not.field.clone(),
                    ops: ops.clone(),
                }),
            },
        }
    }
}

// negate_exists_bson computes the inverse of the exists operator for a given BSON value that is
// used in an $exists query. The $exists operator is true for all values except for null, false,
// numeric 0, and undefined.
fn negate_exists_bson(bson: &Bson) -> Bson {
    match bson {
        Bson::Null => Bson::Boolean(true),
        Bson::Undefined => Bson::Boolean(true),
        Bson::Boolean(b) => Bson::Boolean(!b),
        Bson::Int32(i) => Bson::Boolean(*i == 0),
        Bson::Int64(i) => Bson::Boolean(*i == 0),
        Bson::Double(d) => Bson::Boolean(*d == 0.0),
        Bson::Decimal128(d) => Bson::Boolean(*d == *DECIMAL_ZERO),
        // All other types count as true in exists, so the negation is false.
        _ => Bson::Boolean(false),
    }
}

// negate_type_bson computes the set complement Bson::Array of bson type names from those provided
// in a Bson object that is either a Bson::Array(["type1", "type2", ...]) or a
// Bson::String("type").
fn negate_type_bson(bson: &Bson) -> Bson {
    let non_types = match bson {
        Bson::String(s) => {
            let mut set = HashSet::new();
            set.insert(s.as_str());
            set
        }
        Bson::Array(values) => values
            .iter()
            .map(|v| v.as_str().unwrap())
            .collect::<HashSet<_>>(),
        _ => unreachable!(),
    };
    Bson::Array(
        [
            "double",
            "string",
            "object",
            "array",
            "binData",
            "undefined",
            "objectId",
            "bool",
            "date",
            "null",
            "regex",
            "dbPointer",
            "javascript",
            "symbol",
            "javascriptWithScope",
            "int",
            "timestamp",
            "long",
            "decimal",
            "minKey",
            "maxKey",
        ]
        .iter()
        .filter(|t| !non_types.contains(*t))
        .map(|t| Bson::String(t.to_string()))
        .collect(),
    )
}

// negate_binary_operator computes the negation of a binary operator in a MatchField that
// has a field name, a binary operator, and a BSON value, there may be multiple operators in an
// implicit conjunction, e.g. {x: {$lt: 10, $gt: 5}}, but that is handled below; here,
// {x: {$lt: 10}} and {x: {$gt: 5}} are handled in two separate calls to this function.
fn negate_binary_operator(field: &Ref, op: &MatchBinaryOp, b: &Bson) -> MatchExpression {
    macro_rules! simple_negate {
        ($field:expr, $neg_op:expr, $b:expr) => {
            MatchExpression::Field(MatchField {
                field: $field.clone(),
                ops: map! {$neg_op => $b.clone()},
            })
        };
    }
    macro_rules! logical_not_negate {
        ($field:expr, $op:expr, $b:expr) => {
            MatchExpression::Logical(MatchLogical::Not(MatchNot {
                field: $field.clone(),
                expr: MatchNotExpression::Query(map! {$op => b.clone()}),
            }))
        };
    }
    // function_negate is kept separate from simple_negate to avoid double cloning
    macro_rules! funcion_negate {
        ($field:expr, $op:expr, $f:expr, $b:expr) => {
            MatchExpression::Field(MatchField {
                field: $field.clone(),
                ops: map! {$op => $f($b)},
            })
        };
    }
    match op {
        MatchBinaryOp::Eq => simple_negate!(field, MatchBinaryOp::Ne, b),
        MatchBinaryOp::Ne => simple_negate!(field, MatchBinaryOp::Eq, b),
        MatchBinaryOp::Lt => simple_negate!(field, MatchBinaryOp::Gte, b),
        MatchBinaryOp::Lte => simple_negate!(field, MatchBinaryOp::Gt, b),
        MatchBinaryOp::Gt => simple_negate!(field, MatchBinaryOp::Lte, b),
        MatchBinaryOp::Gte => simple_negate!(field, MatchBinaryOp::Lt, b),
        MatchBinaryOp::In => simple_negate!(field, MatchBinaryOp::Nin, b),
        MatchBinaryOp::Nin => simple_negate!(field, MatchBinaryOp::In, b),
        MatchBinaryOp::Exists => {
            funcion_negate!(field, MatchBinaryOp::Exists, negate_exists_bson, b)
        }
        MatchBinaryOp::Type => funcion_negate!(field, MatchBinaryOp::Type, negate_type_bson, b),
        MatchBinaryOp::Size => logical_not_negate!(field, MatchBinaryOp::Size, b),
        MatchBinaryOp::Mod => logical_not_negate!(field, MatchBinaryOp::Mod, b),
        MatchBinaryOp::BitsAnySet => simple_negate!(field, MatchBinaryOp::BitsAllClear, b),
        MatchBinaryOp::BitsAnyClear => simple_negate!(field, MatchBinaryOp::BitsAllSet, b),
        MatchBinaryOp::BitsAllSet => simple_negate!(field, MatchBinaryOp::BitsAnyClear, b),
        MatchBinaryOp::BitsAllClear => simple_negate!(field, MatchBinaryOp::BitsAnySet, b),
        // We actually could expand this to a $nor of the values, but it would not tell us
        // any useful schema information.
        MatchBinaryOp::All => logical_not_negate!(field, MatchBinaryOp::All, b),
        MatchBinaryOp::GeoIntersects => logical_not_negate!(field, MatchBinaryOp::GeoIntersects, b),
        MatchBinaryOp::GeoWithin => logical_not_negate!(field, MatchBinaryOp::GeoWithin, b),
        MatchBinaryOp::Near => logical_not_negate!(field, MatchBinaryOp::Near, b),
        MatchBinaryOp::NearSphere => logical_not_negate!(field, MatchBinaryOp::NearSphere, b),
    }
}

impl NegativeNormalize<MatchExpression> for MatchField {
    fn get_negative_normal_form(&self) -> MatchExpression {
        MatchExpression::Field(self.clone())
    }

    fn get_negation(&self) -> MatchExpression {
        let ops = self
            .ops
            .iter()
            .map(|(op, b)| negate_binary_operator(&self.field, op, b))
            .collect::<Vec<MatchExpression>>();
        if ops.len() == 1 {
            ops.into_iter().next().unwrap()
        } else {
            MatchExpression::Logical(MatchLogical::Or(ops))
        }
    }
}
