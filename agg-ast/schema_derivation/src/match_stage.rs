use agg_ast::{
    definitions::{
        Expression, MatchBinaryOp, MatchExpr, MatchExpression, MatchField, MatchLogical, MatchNot,
        MatchNotExpression, Ref, UntaggedOperator,
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

fn wrap_in_zero_check(expr: Expression) -> Expression {
    Expression::UntaggedOperator(UntaggedOperator {
        op: "$eq".to_string(),
        args: vec![
            expr,
            Expression::Literal(agg_ast::definitions::LiteralValue::Int32(0)),
        ],
    })
}

fn wrap_in_null_check(expr: Expression) -> Expression {
    Expression::UntaggedOperator(UntaggedOperator {
        op: "$lte".to_string(),
        args: vec![
            expr,
            Expression::Literal(agg_ast::definitions::LiteralValue::Null),
        ],
    })
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
            ref_expr @ Expression::Ref(_) => Expression::UntaggedOperator(UntaggedOperator {
                op: "$or".to_string(),
                args: vec![
                    wrap_in_null_check(ref_expr.clone()),
                    wrap_in_zero_check(ref_expr.clone()),
                ],
            }),
            Expression::TaggedOperator(_t) => todo!(),
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
                        "$not" => todo!(),
                        "$cond" => todo!(),
                        _ => self.clone(),
                    }
                }
                _ => self.clone(),
            },
            MatchExpression::Logical(_logical) => todo!(),
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
            MatchExpression::Logical(_logical) => todo!(),
            MatchExpression::Field(field) => field.get_negation(),
            MatchExpression::Misc(_misc) => todo!(),
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
