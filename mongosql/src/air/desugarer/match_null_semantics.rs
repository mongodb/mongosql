use crate::air::{
    self,
    desugarer::{Pass, Result},
    util::match_sql_to_mql_op,
    visitor::Visitor,
    Expression,
    Expression::*,
    FieldRef, LiteralValue, MQLOperator, MQLSemanticOperator, Match, SQLSemanticOperator, Stage,
};
use std::collections::BTreeMap;

/// Desugars Match stages with SQL operators such that they can partially
/// utilize indexes. Specifically, if a Match contains a SQL operator, a
/// separate Match stage is created with a constraint for the arguments
/// that ensures the values exist and are not null. This new Match is placed
/// before the original Match and the original is updated to use the MQL
/// version of the operator.
pub struct MatchDesugarerPass;

impl Pass for MatchDesugarerPass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage> {
        let mut visitor = MatchDesugarerPassVisitor;
        Ok(visitor.visit_stage(pipeline))
    }
}

#[derive(Default)]
struct MatchDesugarerPassVisitor;

impl MatchDesugarerPassVisitor {
    fn desugar_stage(&mut self, stage: &Match) -> Option<Stage> {
        self.generate_schema(&stage.expr).map(|new_expr| {
            Stage::Match(Match {
                source: stage.source.clone(),
                expr: Box::new(new_expr),
            })
        })
    }

    fn generate_schema(&mut self, expr: &Expression) -> Option<air::Expression> {
        let schema_vars = self
            // gather fields in expr that need to be null checked
            .gather_fields_for_null_semantic_constraints(expr)
            .into_values()
            // create a null-check expression for each field
            .map(|fr| {
                MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Gt,
                    args: vec![FieldRef(fr), Literal(LiteralValue::Null)],
                })
            })
            .collect::<Vec<Expression>>();

        match schema_vars.len() {
            0 => None,
            1 => schema_vars.first().cloned(),
            _ => Some(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::And,
                args: schema_vars,
            })),
        }
    }

    fn gather_fields_for_null_semantic_constraints(
        &mut self,
        expr: &Expression,
    ) -> BTreeMap<String, FieldRef> {
        let mut visitor = FieldRefGatherer::new();
        visitor.gather_field_refs(expr)
    }

    fn desugar_expr(expr: &Expression) -> Expression {
        match expr {
            SQLSemanticOperator(sql_op) => {
                let desugared_args = sql_op.args.iter().map(Self::desugar_expr).collect();
                if let Some(mql_op) = match_sql_to_mql_op(sql_op.op.clone()) {
                    MQLSemanticOperator(MQLSemanticOperator {
                        op: mql_op,
                        args: desugared_args,
                    })
                } else {
                    SQLSemanticOperator(SQLSemanticOperator {
                        op: sql_op.op.clone(),
                        args: desugared_args,
                    })
                }
            }
            _ => expr.clone(),
        }
    }
}

impl Visitor for MatchDesugarerPassVisitor {
    fn visit_stage(&mut self, stage: Stage) -> Stage {
        let node = match stage {
            Stage::Match(m) => {
                let desugared_stage = match self.desugar_stage(&m) {
                    Some(schema_stage) => schema_stage,
                    None => return Stage::Match(m),
                };
                let optimized_match_expr = Self::desugar_expr(&m.expr);
                Stage::Match(Match {
                    source: Box::new(desugared_stage),
                    expr: Box::new(optimized_match_expr),
                })
            }
            _ => stage,
        };
        node.walk(self)
    }
}

struct FieldRefGatherer {
    schema_vars: BTreeMap<String, FieldRef>,
}

impl FieldRefGatherer {
    fn new() -> Self {
        Self {
            schema_vars: BTreeMap::new(),
        }
    }

    fn gather_field_refs(&mut self, expr: &Expression) -> BTreeMap<String, FieldRef> {
        self.visit_expression(expr.clone());
        self.schema_vars.clone()
    }
}

impl Visitor for FieldRefGatherer {
    // Do not walk stages nested within expressions.
    fn visit_stage(&mut self, node: Stage) -> Stage {
        node
    }

    fn visit_expression(&mut self, expr: Expression) -> Expression {
        if let SQLSemanticOperator(sql_op) = expr.clone() {
            for arg in &sql_op.args {
                if let FieldRef(fr) = arg.clone() {
                    self.schema_vars.insert(format!("{fr}"), fr.clone());
                }
            }
        }
        expr.walk(self)
    }
}
