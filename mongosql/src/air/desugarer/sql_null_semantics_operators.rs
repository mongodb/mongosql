use crate::air::{
    self,
    desugarer::{Pass, Result},
    util::sql_op_to_mql_op,
    visitor::Visitor,
    Expression,
    Expression::*,
    LetVariable, LiteralValue, MqlOperator, MqlSemanticOperator,
    SqlOperator::*,
    SwitchCase,
};
use crate::make_cond_expr;

/// Desugars any Sql operators that require Sql null semantics into their
/// corresponding Mql operators wrapped in operations to null-check the
/// arguments.
pub struct SqlNullSemanticsOperatorsDesugarerPass;

impl Pass for SqlNullSemanticsOperatorsDesugarerPass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage> {
        Ok(pipeline.walk(&mut SqlNullSemanticsOperatorsDesugarerVisitor))
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

        let false_check_cond_else_statement = literal_null_found.unwrap_or(make_cond_expr!(
            Self::literal_check_args(let_vars.clone(), MqlOperator::Lte, LiteralValue::Null),
            Literal(LiteralValue::Null),
            Literal(LiteralValue::Boolean(true))
        ));

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

        let true_check_cond_else_statement = literal_null_found.unwrap_or(make_cond_expr!(
            Self::literal_check_args(let_vars.clone(), MqlOperator::Lte, LiteralValue::Null),
            Literal(LiteralValue::Null),
            Literal(LiteralValue::Boolean(false))
        ));

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

    /// Null-guards a `$match` predicate's `FieldRef`/`Variable` comparison operands.
    ///
    /// For a binary comparison (`Lt`/`Lte`/`Gt`/`Gte`/`Eq`/`Ne`), any `FieldRef` or `Variable`
    /// operand is guarded by a preceding `$gt(<operand>, null)` check. This
    ///  keeps the comparison index-eligible. Compound expressions like function calls
    /// are not index-eligible, so those operators are left untouched.
    /// Also recurses into `And`/`Or` (applying the above to each argument).
    /// We only rewrite if the LHS and RHS are simple (FieldRef/Variable/Literal) - otherwise we defer to the main desugarer pass to handle the compound operand.
    fn null_guard_match_predicates(&mut self, node: Expression) -> Expression {
        match node {
            SqlSemanticOperator(sql_operator) => {
                match sql_operator.op {
                    // For comparison operators we want to wrap the expression in an $and expression
                    // that checks for null operands. Ultimately producing an expression like:
                    // and ( ...gt(<operand>, null) for any field refs or variables, <original operation> )
                    Lt | Lte | Gt | Gte | Eq | Ne => {
                        // We use is_simple to determine if we can convert the comparison to MQL.
                        // Literals, FieldRefs and variables could potentially use an index so we consider those.
                        // Scalar functions and other complex expressions don't qualify, so we just exclude them here.
                        let is_simple =
                            |e: &Expression| matches!(e, Literal(_) | FieldRef(_) | Variable(_));

                        // Only FieldRef/Variable operands are worth guarding here - that's what
                        // keeps the comparison index-eligible.
                        let requires_null_guard =
                            |e: &Expression| matches!(e, FieldRef(_) | Variable(_));

                        if !sql_operator.args.iter().all(is_simple) {
                            return SqlSemanticOperator(sql_operator);
                        }

                        let lhs = sql_operator.args[0].clone();
                        let rhs = sql_operator.args[1].clone();

                        if matches!(lhs, Literal(_)) && matches!(rhs, Literal(_)) {
                            // If both sides are literals, we don't need to guard them - just return the comparison as-is.
                            return MqlSemanticOperator(air::MqlSemanticOperator {
                                op: sql_op_to_mql_op(sql_operator.op).unwrap(),
                                args: sql_operator.args,
                            });
                        }

                        // Rebuild the comparison itself as raw MQL only if both sides
                        let comparison = MqlSemanticOperator(air::MqlSemanticOperator {
                            op: sql_op_to_mql_op(sql_operator.op).unwrap(),
                            args: sql_operator.args,
                        });

                        let and_args: Vec<Expression> = [&lhs, &rhs]
                            .into_iter()
                            .filter(|expr| requires_null_guard(expr))
                            .cloned()
                            .map(|operand| {
                                // Return an expression that checks if the operand is null
                                MqlSemanticOperator(air::MqlSemanticOperator {
                                    op: MqlOperator::Gt,
                                    args: vec![operand, Literal(LiteralValue::Null)],
                                })
                            })
                            // Append the original comparison to the end of the list of null guards
                            .chain(std::iter::once(comparison))
                            .collect();

                        MqlSemanticOperator(air::MqlSemanticOperator {
                            op: MqlOperator::And,
                            args: and_args,
                        })
                    }
                    And | Or => {
                        // For logical operators we want to recursively call this function on the arguments.
                        // This function only ever runs on a $match's $expr boolean predicate, where MQL's $and/$or already treat
                        // null and false identically (both falsy) - exactly matching SQL's 3 value null semantics, so this is safe.
                        let new_args: Vec<Expression> = sql_operator
                            .args
                            .into_iter()
                            .map(|arg| self.null_guard_match_predicates(arg))
                            .collect();

                        MqlSemanticOperator(air::MqlSemanticOperator {
                            op: sql_op_to_mql_op(sql_operator.op).unwrap(),
                            args: new_args,
                        })
                    }
                    // Any other operator is left for the main desugarer pass to handle.
                    _ => SqlSemanticOperator(sql_operator),
                }
            }
            // No other statements are index eligible, so we don't guard them. Allow desugarer to do the rest of the work.
            _ => node,
        }
    }
}

impl Visitor for SqlNullSemanticsOperatorsDesugarerVisitor {
    fn visit_match(&mut self, node: air::Match) -> air::Match {
        match node {
            air::Match::ExprLanguage(air::ExprLanguage { source, expr }) => {
                let source = self.visit_stage(*source);
                // Apply the null-guard rewrite directly to the raw $expr before it ever
                // reaches visit_expression.
                let expr = self.null_guard_match_predicates(*expr);
                let expr = self.visit_expression(expr);
                air::Match::ExprLanguage(air::ExprLanguage {
                    source: Box::new(source),
                    expr: Box::new(expr),
                })
            }
            air::Match::MatchLanguage(ml) => air::Match::MatchLanguage(ml.walk(self)),
        }
    }

    fn visit_expression(&mut self, node: Expression) -> Expression {
        let node = match node {
            SqlSemanticOperator(sql_operator) => match sql_operator.op {
                And => self.desugar_sql_and(sql_operator),
                Or => self.desugar_sql_or(sql_operator),
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
mod match_predicate_null_guard_tests {
    use super::*;

    fn visitor() -> SqlNullSemanticsOperatorsDesugarerVisitor {
        SqlNullSemanticsOperatorsDesugarerVisitor
    }

    fn field(name: &str) -> Expression {
        FieldRef(name.into())
    }

    fn int(v: i32) -> Expression {
        Literal(LiteralValue::Integer(v))
    }

    fn sql_op(op: air::SqlOperator, args: Vec<Expression>) -> Expression {
        SqlSemanticOperator(air::SqlSemanticOperator { op, args })
    }

    fn mql_op(op: MqlOperator, args: Vec<Expression>) -> Expression {
        MqlSemanticOperator(air::MqlSemanticOperator { op, args })
    }

    /// `$gt([<field>, null])` — true iff the field is non-null and non-missing,
    /// because null and missing sort below all other BSON types.
    fn null_guard(name: &str) -> Expression {
        mql_op(
            MqlOperator::Gt,
            vec![field(name), Literal(LiteralValue::Null)],
        )
    }

    #[test]
    fn comparison_with_field_and_literal_gets_one_null_guard() {
        let input = sql_op(air::SqlOperator::Lt, vec![field("a"), int(0)]);
        let expected = mql_op(
            MqlOperator::And,
            vec![
                null_guard("a"),
                mql_op(MqlOperator::Lt, vec![field("a"), int(0)]),
            ],
        );

        let actual = visitor().null_guard_match_predicates(input);

        assert_eq!(expected, actual);
    }

    #[test]
    fn comparison_with_two_fields_gets_a_guard_for_each() {
        let input = sql_op(air::SqlOperator::Gt, vec![field("a"), field("b")]);
        let expected = mql_op(
            MqlOperator::And,
            vec![
                null_guard("a"),
                null_guard("b"),
                mql_op(MqlOperator::Gt, vec![field("a"), field("b")]),
            ],
        );

        let actual = visitor().null_guard_match_predicates(input);

        assert_eq!(expected, actual);
    }

    #[test]
    fn comparison_with_literal_lhs_guards_only_the_field() {
        let input = sql_op(air::SqlOperator::Gte, vec![int(100), field("a")]);
        let expected = mql_op(
            MqlOperator::And,
            vec![
                null_guard("a"),
                mql_op(MqlOperator::Gte, vec![int(100), field("a")]),
            ],
        );

        let actual = visitor().null_guard_match_predicates(input);

        assert_eq!(expected, actual);
    }

    #[test]
    fn comparison_with_two_literals_is_left_unguarded() {
        let input = sql_op(air::SqlOperator::Ne, vec![int(1), int(2)]);
        let expected = mql_op(MqlOperator::Ne, vec![int(1), int(2)]);

        let actual = visitor().null_guard_match_predicates(input);

        assert_eq!(expected, actual);
    }

    #[test]
    fn comparison_with_two_compound_operands_defers_entirely() {
        // Neither side is a FieldRef/Variable/Literal, so there is nothing cheap to guard here;
        // defer the whole comparison, unchanged, to desugar_sql_op.
        let lhs = sql_op(air::SqlOperator::StrLenCP, vec![field("a")]);
        let rhs = sql_op(air::SqlOperator::StrLenCP, vec![field("b")]);
        let input = sql_op(air::SqlOperator::Ne, vec![lhs, rhs]);

        let actual = visitor().null_guard_match_predicates(input.clone());

        assert_eq!(input, actual);
    }

    #[test]
    fn logical_operator_recurses_into_each_operand() {
        let input = sql_op(
            air::SqlOperator::Or,
            vec![
                sql_op(air::SqlOperator::Lt, vec![field("a"), int(10)]),
                sql_op(air::SqlOperator::Gt, vec![field("b"), int(20)]),
            ],
        );
        let expected = mql_op(
            MqlOperator::Or,
            vec![
                mql_op(
                    MqlOperator::And,
                    vec![
                        null_guard("a"),
                        mql_op(MqlOperator::Lt, vec![field("a"), int(10)]),
                    ],
                ),
                mql_op(
                    MqlOperator::And,
                    vec![
                        null_guard("b"),
                        mql_op(MqlOperator::Gt, vec![field("b"), int(20)]),
                    ],
                ),
            ],
        );

        let actual = visitor().null_guard_match_predicates(input);

        assert_eq!(expected, actual);
    }
    #[test]
    fn logical_operator_mixes_guarded_and_deferred_branches() {
        // a > 5 AND CHAR_LENGTH(b) <> 0 - one branch is a simple field comparison (guarded), the
        // other's operand is compound (deferred to desugar_sql_op). And/Or must recurse into
        // each branch independently rather than treating the whole thing uniformly.
        let compound = sql_op(air::SqlOperator::StrLenCP, vec![field("b")]);
        let input = sql_op(
            air::SqlOperator::And,
            vec![
                sql_op(air::SqlOperator::Gt, vec![field("a"), int(5)]),
                sql_op(air::SqlOperator::Ne, vec![compound.clone(), int(0)]),
            ],
        );
        let expected = mql_op(
            MqlOperator::And,
            vec![
                mql_op(
                    MqlOperator::And,
                    vec![
                        null_guard("a"),
                        mql_op(MqlOperator::Gt, vec![field("a"), int(5)]),
                    ],
                ),
                sql_op(air::SqlOperator::Ne, vec![compound, int(0)]),
            ],
        );

        let actual = visitor().null_guard_match_predicates(input);

        assert_eq!(expected, actual);
    }

    #[test]
    fn switch_is_returned_unchanged() {
        // A Switch (CASE expression) is never index-eligible, so comparisons nested in its
        // case/then branches are bailed on entirely rather than guarded - the whole Switch is
        // left untouched for the later visit_expression walk to desugar normally.
        let input = Switch(air::Switch {
            branches: vec![SwitchCase {
                case: Box::new(sql_op(air::SqlOperator::Lt, vec![field("a"), int(5)])),
                then: Box::new(sql_op(air::SqlOperator::Gt, vec![field("b"), int(10)])),
            }],
            default: Box::new(int(0)),
        });

        let actual = visitor().null_guard_match_predicates(input.clone());

        assert_eq!(input, actual);
    }

    #[test]
    fn unhandled_operator_is_returned_unchanged() {
        let input = sql_op(air::SqlOperator::Size, vec![field("a")]);

        let actual = visitor().null_guard_match_predicates(input.clone());

        assert_eq!(input, actual);
    }

    #[test]
    fn non_sql_operator_expression_is_returned_unchanged() {
        let input = field("a");

        let actual = visitor().null_guard_match_predicates(input.clone());

        assert_eq!(input, actual);
    }
}
