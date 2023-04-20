use crate::air::{
    self,
    desugarer::{Pass, Result},
    util::sql_op_to_mql_op,
    visitor::Visitor,
    Expression,
    Expression::*,
    LetVariable, LiteralValue, MQLOperator,
    SQLOperator::*,
};

/// Desugars any SQL operators that require SQL null semantics into their
/// corresponding MQL operators wrapped in operations to null-check the
/// arguments.
pub struct SQLNullSemanticsOperatorsDesugarerPass;

impl Pass for SQLNullSemanticsOperatorsDesugarerPass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage> {
        Ok(pipeline.walk(&mut SQLNullSemanticsOperatorsDesugarerVisitor))
    }
}

#[derive(Default)]
struct SQLNullSemanticsOperatorsDesugarerVisitor;

impl SQLNullSemanticsOperatorsDesugarerVisitor {
    fn null_check_cond_if_statement(let_vars: Vec<LetVariable>) -> Expression {
        let args = let_vars
            .into_iter()
            .map(|let_var| {
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op: MQLOperator::Lte,
                    args: vec![
                        Variable(air::Variable {
                            parent: None,
                            name: let_var.name,
                        }),
                        Literal(LiteralValue::Null),
                    ],
                })
            })
            .collect::<Vec<Expression>>();
        match args.len() {
            1 => args[0].clone(),
            _ => MQLSemanticOperator(air::MQLSemanticOperator {
                op: MQLOperator::Or,
                args,
            }),
        }
    }

    fn desugar_sql_and(&mut self, sql_operator: air::SQLSemanticOperator) -> Expression {
        let mut sql_operator_args_eq_false_check: Vec<Expression> = Vec::new();
        let mut let_vars: Vec<LetVariable> = Vec::new();

        for (let_vars_idx, expr) in sql_operator.args.into_iter().enumerate() {
            let let_var = LetVariable {
                name: format!("desugared_sqlAnd_input{}", let_vars_idx),
                expr: Box::new(expr),
            };
            let_vars.push(let_var.clone());
            sql_operator_args_eq_false_check.push(MQLSemanticOperator(air::MQLSemanticOperator {
                op: MQLOperator::Eq,
                args: vec![
                    Variable(air::Variable {
                        parent: None,
                        name: let_var.name,
                    }),
                    Literal(LiteralValue::Boolean(false)),
                ],
            }));
        }

        // If any of the arguments are false, return false. Otherwise, if any of the arguments are null, return null. Otherwise, return true.
        let cond = MQLSemanticOperator(air::MQLSemanticOperator {
            op: MQLOperator::Cond,
            args: vec![
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op: MQLOperator::Or,
                    args: sql_operator_args_eq_false_check,
                }),
                Literal(LiteralValue::Boolean(false)),
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op: MQLOperator::Cond,
                    args: vec![
                        SQLNullSemanticsOperatorsDesugarerVisitor::null_check_cond_if_statement(
                            let_vars.clone(),
                        ),
                        Literal(LiteralValue::Null),
                        Literal(LiteralValue::Boolean(true)),
                    ],
                }),
            ],
        });

        Let(air::Let {
            vars: let_vars,
            inside: Box::new(cond),
        })
    }

    fn desugar_sql_or(&mut self, sql_operator: air::SQLSemanticOperator) -> Expression {
        let mut sql_operator_args_eq_true_check: Vec<Expression> = Vec::new();
        let mut let_vars: Vec<LetVariable> = Vec::new();

        for (let_vars_idx, expr) in sql_operator.args.into_iter().enumerate() {
            let let_var = LetVariable {
                name: format!("desugared_sqlOr_input{}", let_vars_idx),
                expr: Box::new(expr),
            };
            let_vars.push(let_var.clone());
            sql_operator_args_eq_true_check.push(MQLSemanticOperator(air::MQLSemanticOperator {
                op: MQLOperator::Eq,
                args: vec![
                    Variable(air::Variable {
                        parent: None,
                        name: let_var.name,
                    }),
                    Literal(LiteralValue::Boolean(true)),
                ],
            }));
        }

        // If any of the arguments are true, return true. Otherwise, if any of the arguments are null, return null. Otherwise, return false.
        let cond = MQLSemanticOperator(air::MQLSemanticOperator {
            op: MQLOperator::Cond,
            args: vec![
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op: MQLOperator::Or,
                    args: sql_operator_args_eq_true_check,
                }),
                Literal(LiteralValue::Boolean(true)),
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op: MQLOperator::Cond,
                    args: vec![
                        SQLNullSemanticsOperatorsDesugarerVisitor::null_check_cond_if_statement(
                            let_vars.clone(),
                        ),
                        Literal(LiteralValue::Null),
                        Literal(LiteralValue::Boolean(false)),
                    ],
                }),
            ],
        });

        Let(air::Let {
            vars: let_vars,
            inside: Box::new(cond),
        })
    }

    fn desugar_sql_op(&mut self, sql_operator: air::SQLSemanticOperator) -> Expression {
        let op_name = "sql".to_string() + &format!("{:?}", sql_operator.op);

        let mut mql_operator_args: Vec<Expression> = Vec::new();
        let mut let_vars: Vec<LetVariable> = Vec::new();

        for (let_vars_idx, expr) in sql_operator.args.clone().into_iter().enumerate() {
            let let_var = LetVariable {
                name: format!("desugared_{}_input{}", op_name, let_vars_idx),
                expr: Box::new(expr),
            };
            let_vars.push(let_var.clone());
            mql_operator_args.push(Variable(air::Variable {
                parent: None,
                name: let_var.name,
            }));
        }

        let let_inside = MQLSemanticOperator(air::MQLSemanticOperator {
            op: MQLOperator::Cond,
            args: vec![
                SQLNullSemanticsOperatorsDesugarerVisitor::null_check_cond_if_statement(
                    let_vars.clone(),
                ),
                Literal(LiteralValue::Null),
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op: sql_op_to_mql_op(sql_operator.op).unwrap(),
                    args: mql_operator_args,
                }),
            ],
        });

        Let(air::Let {
            vars: let_vars,
            inside: Box::new(let_inside),
        })
    }
}

impl Visitor for SQLNullSemanticsOperatorsDesugarerVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        let node = match node {
            SQLSemanticOperator(sql_operator) => match sql_operator.op {
                And => self.desugar_sql_and(sql_operator),
                Or => self.desugar_sql_or(sql_operator),
                Eq | IndexOfCP | Lt | Lte | Gt | Gte | Ne | Not | Size | StrLenBytes | StrLenCP
                | SubstrCP | ToLower | ToUpper => self.desugar_sql_op(sql_operator),
                _ => SQLSemanticOperator(sql_operator),
            },
            _ => node,
        };
        node.walk(self)
    }
}
