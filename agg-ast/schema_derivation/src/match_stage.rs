use agg_ast::definitions::{Expression, MatchExpr, MatchExpression, UntaggedOperator};

#[allow(dead_code)]
pub(crate) trait NegativeNormalize {
    fn get_negation(&self) -> Self;
    #[allow(dead_code)]
    fn get_negative_normal_form(&self) -> Self;
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

impl NegativeNormalize for Expression {
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

impl NegativeNormalize for MatchExpression {
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
            MatchExpression::Field(_field) => todo!(),
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
            MatchExpression::Field(_field) => todo!(),
            MatchExpression::Misc(_misc) => todo!(),
        }
    }
}
