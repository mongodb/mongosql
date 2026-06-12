use crate::air::Stage::{EquiJoin, Join};
use crate::air::{
    self,
    desugarer::{Pass, Result},
    util::sql_op_to_mql_op,
    visitor::Visitor,
    Expression,
    Expression::*,
    LetVariable, LiteralValue, MqlOperator, MqlSemanticOperator,
    SqlOperator::*,
    Stage, SwitchCase,
};
use crate::make_cond_expr;
use bson::binary::Vector;
use std::env::args;

/// Desugars any Sql operators that require Sql null semantics into their
/// corresponding Mql operators wrapped in operations to null-check the
/// arguments.
pub struct SqlNullSemanticsOperatorsDesugarerPass;

impl Pass for SqlNullSemanticsOperatorsDesugarerPass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage> {
        let is_in_match_context: bool = match pipeline {
            air::Stage::Match(_) => true,
            _ => false,
        };

        Ok(pipeline.walk(&mut SqlNullSemanticsOperatorsDesugarerVisitor))
    }
}

#[derive(Default)]
struct SqlNullSemanticsMatchStageDesugarerVisitor;

impl Visitor for SqlNullSemanticsMatchStageDesugarerVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        // [TODO] Implement desugaring of SQL operators in $match stages, which is more complicated than
        // desugaring them in other stages because we want to preserve the ability to use indexes for
        // simple cases. This likely involves only desugaring SQL operators when they are used in a
        // way that prevents index usage, e.g. by appearing inside an $or or being compared to a non-constant.
        node
    }
}

#[derive(Default)]
struct SqlNullSemanticsOperatorsDesugarerVisitor;

impl SqlNullSemanticsOperatorsDesugarerVisitor {
    fn literal_check_args(
        let_vars: Vec<LetVariable>,
        op: MqlOperator,
        lit_val: LiteralValue,
    ) -> Expression {
        let args = let_vars
            .into_iter()
            .map(|let_var| {
                MqlSemanticOperator(air::MqlSemanticOperator {
                    op,
                    args: vec![Variable(let_var.name.into()), Literal(lit_val.clone())],
                })
            })
            .collect::<Vec<Expression>>();
        match args.len() {
            1 => args[0].clone(),
            _ => MqlSemanticOperator(air::MqlSemanticOperator {
                op: MqlOperator::Or,
                args,
            }),
        }
    }

    /// Transforms SQL `x IN (a, b, c)` into MQL that correctly handles null semantics.
    ///
    /// SQL's IN operator has three-valued logic:
    ///   - `true`  — LHS matches at least one non-null RHS element
    ///   - `null`  — LHS is null, OR no non-null match was found but a null exists in RHS
    ///   - `false` — LHS is non-null and does not match any element (all comparisons are non-null)
    ///
    /// The generated MQL structure is:
    /// ```text
    /// $let { lhs, lhs_is_null, rhs }
    ///   $reduce(
    ///     input: $map(rhs) → per-element true/null/false via $switch
    ///     init:  false
    ///     body:  SQL-OR accumulation (true > null > false)
    ///   )
    /// ```
    fn desugar_sql_in(&mut self, sql_operator: air::SqlSemanticOperator) -> Expression {
        const LHS_VAR: &str = "desugared_sqlIn_input0";
        const LHS_NULL_VAR: &str = "desugared_sqlIn_input0_is_nullish";
        const RHS_VAR: &str = "desugared_sqlIn_input1";
        const THIS_NULL_VAR: &str = "desugared_this_is_nullish";

        assert_eq!(
            sql_operator.args.len(),
            2,
            "desugar_sql_in: expected exactly 2 args (lhs, rhs array), got {}",
            sql_operator.args.len()
        );
        let mut args = sql_operator.args.into_iter();
        let lhs = args.next().unwrap();
        let rhs = args.next().unwrap();

        // Pre-compute whether LHS is null using the original expression — not the $let variable
        // for LHS — because sibling $let variables cannot reference each other in MQL; they are
        // evaluated in parallel, not sequentially.
        //
        // Null detection trick: `$lte([x, null])` is true iff x is null or missing.
        // This works because null/missing sort below all other BSON types in MQL's comparison
        // order, so only a null/missing value satisfies `x <= null`.
        let lhs_null_check = MqlSemanticOperator(air::MqlSemanticOperator {
            op: MqlOperator::Lte,
            args: vec![lhs.clone(), Literal(LiteralValue::Null)],
        });

        // Bind three variables for the duration of the expression:
        //   - LHS value (evaluated once)
        //   - LHS nullish flag (pre-computed to avoid re-evaluating the original expression)
        //   - RHS array (evaluated once, iterated by $map)
        let outer_vars = vec![
            LetVariable {
                name: LHS_VAR.into(),
                expr: Box::new(lhs),
            },
            LetVariable {
                name: LHS_NULL_VAR.into(),
                expr: Box::new(lhs_null_check),
            },
            LetVariable {
                name: RHS_VAR.into(),
                expr: Box::new(rhs),
            },
        ];

        // Per-element comparison: for each element in the RHS array (accessible as $$this
        // inside $map), produce a three-valued result:
        //   - true  — both LHS and the current RHS element are non-null and equal
        //   - null  — either side is null (the comparison result is "unknown")
        //   - false — both are non-null but unequal
        //
        // The inner $let binds the RHS element's nullish state so we don't recompute it
        // across the two branches that reference it.
        let this_null_check = MqlSemanticOperator(air::MqlSemanticOperator {
            op: MqlOperator::Lte,
            args: vec![Variable("this".into()), Literal(LiteralValue::Null)],
        });

        // Case 1 → true: both sides are non-null AND equal.
        // Null guards come before $eq because MQL's $eq(null, null) returns true,
        // which would incorrectly match two null values under SQL semantics.
        let branch_true = SwitchCase {
            case: Box::new(MqlSemanticOperator(air::MqlSemanticOperator {
                op: MqlOperator::And,
                args: vec![
                    MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::Not,
                        args: vec![Variable(LHS_NULL_VAR.into())],
                    }),
                    MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::Not,
                        args: vec![Variable(THIS_NULL_VAR.into())],
                    }),
                    MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::Eq,
                        args: vec![Variable(LHS_VAR.into()), Variable("this".into())],
                    }),
                ],
            })),
            then: Box::new(Literal(LiteralValue::Boolean(true))),
        };

        // Case 2 → null: at least one side is null, so equality is indeterminate.
        // Only reached when branch_true didn't fire (no non-null match was found).
        let branch_null = SwitchCase {
            case: Box::new(MqlSemanticOperator(air::MqlSemanticOperator {
                op: MqlOperator::Or,
                args: vec![
                    Variable(LHS_NULL_VAR.into()),
                    Variable(THIS_NULL_VAR.into()),
                ],
            })),
            then: Box::new(Literal(LiteralValue::Null)),
        };

        // Wrap the $switch in a $let that binds THIS_NULL_VAR once; both branch_true and
        // branch_null reference it, so computing it here avoids evaluating $lte($$this, null) twice.
        // Default branch (false) fires when both sides are non-null but unequal.
        let per_elem_expr = Let(air::Let {
            vars: vec![LetVariable {
                name: THIS_NULL_VAR.into(),
                expr: Box::new(this_null_check),
            }],
            inside: Box::new(Switch(air::Switch {
                branches: vec![branch_true, branch_null],
                default: Box::new(Literal(LiteralValue::Boolean(false))),
            })),
        });

        // Apply per_elem_expr to every RHS element, producing an array of three-valued results,
        // e.g. [true, null, false]. as_name: None means each element is accessible as $$this.
        let map_expr = Map(air::Map {
            input: Box::new(Variable(RHS_VAR.into())),
            as_name: None,
            inside: Box::new(per_elem_expr),
        });

        // Accumulate per-element results using SQL-OR semantics: true > null > false.
        // Starting from false (the identity value), each element can only upgrade the result.
        //
        // Logic per step ($$value = accumulator, $$this = current mapped element):
        //   if  $or([$$value, $$this]) is truthy → true   (a match was found)
        //   elif either is null                  → null   (uncertainty preserved)
        //   else                                 → false
        //
        // MQL's $or treats any non-false, non-null value as truthy, so a mapped result
        // of `true` triggers the first branch; `null` and `false` fall through to the
        // null check.
        let acc_value = Variable("value".into());
        let acc_this = Variable("this".into());

        let null_cond = make_cond_expr!(
            MqlSemanticOperator(air::MqlSemanticOperator {
                op: MqlOperator::Or,
                args: vec![
                    MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::Lte,
                        args: vec![acc_value.clone(), Literal(LiteralValue::Null)],
                    }),
                    MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::Lte,
                        args: vec![acc_this.clone(), Literal(LiteralValue::Null)],
                    }),
                ],
            }),
            Literal(LiteralValue::Null),
            Literal(LiteralValue::Boolean(false))
        );

        let reduce_body = make_cond_expr!(
            MqlSemanticOperator(air::MqlSemanticOperator {
                op: MqlOperator::Or,
                args: vec![acc_value, acc_this],
            }),
            Literal(LiteralValue::Boolean(true)),
            null_cond
        );

        Let(air::Let {
            vars: outer_vars,
            inside: Box::new(Reduce(air::Reduce {
                input: Box::new(map_expr),
                init_value: Box::new(Literal(LiteralValue::Boolean(false))),
                inside: Box::new(reduce_body),
            })),
        })
    }

    fn desugar_sql_and(&mut self, sql_operator: air::SqlSemanticOperator) -> Expression {
        let mut let_vars: Vec<LetVariable> = Vec::new();

        let mut literal_null_found: Option<Expression> = None;

        for (let_vars_idx, expr) in sql_operator.args.into_iter().enumerate() {
            // Due to constant folding in the mir, Null is the only possible Literal expr can be.
            if let Literal(LiteralValue::Null) = expr {
                literal_null_found = Some(Literal(LiteralValue::Null));
            } else {
                let_vars.push(LetVariable {
                    name: format!("desugared_sqlAnd_input{let_vars_idx}"),
                    expr: Box::new(expr),
                });
            }
        }

        let false_check_cond_else_statement = literal_null_found.map_or(
            make_cond_expr!(
                Self::literal_check_args(let_vars.clone(), MqlOperator::Lte, LiteralValue::Null),
                Literal(LiteralValue::Null),
                Literal(LiteralValue::Boolean(true))
            ),
            |x| x,
        );

        // If any of the arguments are false, return false.
        // Otherwise, if any of the arguments are null, return null. Otherwise, return true.
        let cond = make_cond_expr!(
            Self::literal_check_args(
                let_vars.clone(),
                MqlOperator::Eq,
                LiteralValue::Boolean(false)
            ),
            Literal(LiteralValue::Boolean(false)),
            false_check_cond_else_statement
        );

        Let(air::Let {
            vars: let_vars,
            inside: Box::new(cond),
        })
    }

    fn desugar_sql_or_in_match_context(
        &mut self,
        sql_operator: air::SqlSemanticOperator,
    ) -> Expression {
        // 1. For each field reference in the $or's arguments, update it to be a conjunction of itself and an explicit null check, e.g. `a` → `a AND a >= null`.
        // This ensures that we'll get index usage because mongodb will know if the field is nullable or not.
        let mut transformed_or_expression: Vec<Expression> = sql_operator
            .args
            .into_iter()
            .map(|expr| match expr {
                FieldRef(fieldRef) => {
                    let null_check_expr = MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::Gte,
                        args: vec![
                            Variable(air::Variable {
                                parent: None,
                                name: fieldRef.name.clone(),
                            }),
                            Literal(LiteralValue::Null),
                        ],
                    });

                    MqlSemanticOperator(air::MqlSemanticOperator {
                        op: MqlOperator::And,
                        args: vec![FieldRef(fieldRef), null_check_expr],
                    })
                }
                _ => expr,
            })
            .collect();

        MqlSemanticOperator(air::MqlSemanticOperator {
            op: MqlOperator::Or,
            args: transformed_or_expression,
        })
    }

    fn desugar_sql_or(&mut self, sql_operator: air::SqlSemanticOperator) -> Expression {
        let mut let_vars: Vec<LetVariable> = Vec::new();

        let mut literal_null_found: Option<Expression> = None;

        for (let_vars_idx, expr) in sql_operator.args.into_iter().enumerate() {
            // Due to constant folding in the mir, Null is the only possible Literal expr can be.
            if let Literal(LiteralValue::Null) = expr {
                literal_null_found = Some(Literal(LiteralValue::Null));
            } else {
                let_vars.push(LetVariable {
                    name: format!("desugared_sqlOr_input{let_vars_idx}"),
                    expr: Box::new(expr),
                });
            }
        }

        let true_check_cond_else_statement = literal_null_found.map_or(
            make_cond_expr!(
                Self::literal_check_args(let_vars.clone(), MqlOperator::Lte, LiteralValue::Null),
                Literal(LiteralValue::Null),
                Literal(LiteralValue::Boolean(false))
            ),
            |x| x,
        );

        // If any of the arguments are true, return true.
        // Otherwise, if any of the arguments are null, return null. Otherwise, return false.
        let cond = make_cond_expr!(
            Self::literal_check_args(
                let_vars.clone(),
                MqlOperator::Eq,
                LiteralValue::Boolean(true)
            ),
            Literal(LiteralValue::Boolean(true)),
            true_check_cond_else_statement
        );

        Let(air::Let {
            vars: let_vars,
            inside: Box::new(cond),
        })
    }

    fn desugar_sql_op(&mut self, sql_operator: air::SqlSemanticOperator) -> Expression {
        let op_name = "sql".to_string() + &format!("{:?}", sql_operator.op);

        let mut mql_operator_args: Vec<Expression> = Vec::new();
        let mut let_vars: Vec<LetVariable> = Vec::new();

        for (let_vars_idx, expr) in sql_operator.args.clone().into_iter().enumerate() {
            // The mir optimizer ensures we will never have null literals as arguments to
            // any of these operators.
            let mql_operator_arg = if matches!(expr, Literal(_)) {
                expr
            } else {
                let let_var = LetVariable {
                    name: format!("desugared_{op_name}_input{let_vars_idx}"),
                    expr: Box::new(expr),
                };
                let_vars.push(let_var.clone());
                Variable(let_var.name.into())
            };

            mql_operator_args.push(mql_operator_arg);
        }

        let mql_op = MqlSemanticOperator(air::MqlSemanticOperator {
            op: sql_op_to_mql_op(sql_operator.op).unwrap(),
            args: mql_operator_args,
        });
        if let_vars.is_empty() {
            mql_op
        } else {
            Let(air::Let {
                vars: let_vars.clone(),
                inside: Box::new(make_cond_expr!(
                    Self::literal_check_args(let_vars, MqlOperator::Lte, LiteralValue::Null),
                    Literal(LiteralValue::Null),
                    mql_op
                )),
            })
        }
    }
}

impl Visitor for SqlNullSemanticsOperatorsDesugarerVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        let node = match node {
            SqlSemanticOperator(sql_operator) => match sql_operator.op {
                And => self.desugar_sql_and(sql_operator),
                Or => self.desugar_sql_or_in_match_context(sql_operator),
                Eq | IndexOfCP | Lt | Lte | Gt | Gte | Ne | Not | Size | StrLenBytes | StrLenCP
                | SubstrCP | ToLower | ToUpper => self.desugar_sql_op(sql_operator),
                In => self.desugar_sql_in(sql_operator),
                NotIn => {
                    // Rewrite as Not(In(...)) so each operator's null-semantics desugaring
                    // fires independently: desugar_sql_in handles the IN, then desugar_sql_op
                    // wraps the result in a null-aware $cond guard for NOT.
                    // Using plain MqlOperator::Not here would give $not(null) = true, which
                    // violates SQL three-valued logic (NOT NULL must remain NULL).
                    return self.visit_expression(SqlSemanticOperator(air::SqlSemanticOperator {
                        op: Not,
                        args: vec![SqlSemanticOperator(air::SqlSemanticOperator {
                            op: In,
                            args: sql_operator.args,
                        })],
                    }));
                }
                _ => SqlSemanticOperator(sql_operator),
            },
            _ => node,
        };
        node.walk(self)
    }
}

#[cfg(test)]
mod tests {
    use crate::air;
    use crate::air::visitor::Visitor;

    #[test]
    fn it_transforms_or_statement_to_conjunction_with_null_checks_for_field_refs() {
        let mut visitor = super::SqlNullSemanticsOperatorsDesugarerVisitor::default();
        let or_expression = air::Expression::SqlSemanticOperator(air::SqlSemanticOperator {
            op: air::SqlOperator::Or,
            args: vec![
                air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string(),
                }),
                air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "b".to_string(),
                }),
                air::Expression::Literal(air::LiteralValue::Boolean(true)),
            ],
        });

        let expected_transformed_expression =
            air::Expression::MqlSemanticOperator(air::MqlSemanticOperator {
                op: air::MqlOperator::Or,
                args: vec![
                    air::Expression::MqlSemanticOperator(air::MqlSemanticOperator {
                        op: air::MqlOperator::And,
                        args: vec![
                            air::Expression::FieldRef(air::FieldRef {
                                parent: None,
                                name: "a".to_string(),
                            }),
                            air::Expression::MqlSemanticOperator(air::MqlSemanticOperator {
                                op: air::MqlOperator::Gte,
                                args: vec![
                                    air::Expression::Variable(air::Variable {
                                        parent: None,
                                        name: "a".to_string(),
                                    }),
                                    air::Expression::Literal(air::LiteralValue::Null),
                                ],
                            }),
                        ],
                    }),
                    air::Expression::MqlSemanticOperator(air::MqlSemanticOperator {
                        op: air::MqlOperator::And,
                        args: vec![
                            air::Expression::FieldRef(air::FieldRef {
                                parent: None,
                                name: "b".to_string(),
                            }),
                            air::Expression::MqlSemanticOperator(air::MqlSemanticOperator {
                                op: air::MqlOperator::Gte,
                                args: vec![
                                    air::Expression::Variable(air::Variable {
                                        parent: None,
                                        name: "b".to_string(),
                                    }),
                                    air::Expression::Literal(air::LiteralValue::Null),
                                ],
                            }),
                        ],
                    }),
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                ],
            });
        let actual_transformed_expression = visitor.visit_expression(or_expression);
        assert_eq!(
            expected_transformed_expression, actual_transformed_expression,
            "{}",
            "Transformation Test"
        );
    }
}
