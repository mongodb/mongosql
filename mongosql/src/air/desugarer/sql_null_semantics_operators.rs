use crate::air::{
    self,
    desugarer::{Pass, Result},
    util::sql_op_to_mql_op,
    visitor::Visitor,
    Expression,
    Expression::*,
    LetVariable, LiteralValue, MQLOperator, MQLSemanticOperator,
    SQLOperator::*,
};
use crate::make_cond_expr;

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
    fn literal_check_args(
        let_vars: Vec<LetVariable>,
        op: MQLOperator,
        lit_val: LiteralValue,
    ) -> Expression {
        let args = let_vars
            .into_iter()
            .map(|let_var| {
                MQLSemanticOperator(air::MQLSemanticOperator {
                    op,
                    args: vec![Variable(let_var.name.into()), Literal(lit_val.clone())],
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
                Self::literal_check_args(let_vars.clone(), MQLOperator::Lte, LiteralValue::Null),
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
                MQLOperator::Eq,
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

    fn desugar_sql_or(&mut self, sql_operator: air::SQLSemanticOperator) -> Expression {
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
                Self::literal_check_args(let_vars.clone(), MQLOperator::Lte, LiteralValue::Null),
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
                MQLOperator::Eq,
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

    fn desugar_sql_op(&mut self, sql_operator: air::SQLSemanticOperator) -> Expression {
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

        let mql_op = MQLSemanticOperator(air::MQLSemanticOperator {
            op: sql_op_to_mql_op(sql_operator.op).unwrap(),
            args: mql_operator_args,
        });
        if let_vars.is_empty() {
            mql_op
        } else {
            Let(air::Let {
                vars: let_vars.clone(),
                inside: Box::new(make_cond_expr!(
                    Self::literal_check_args(let_vars, MQLOperator::Lte, LiteralValue::Null),
                    Literal(LiteralValue::Null),
                    mql_op
                )),
            })
        }
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
