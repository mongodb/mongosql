use crate::air::{
    self,
    desugarer::{Pass, Result},
    util::match_sql_to_mql_op,
    visitor::Visitor,
};
use lazy_static::lazy_static;
use linked_hash_map::LinkedHashMap;
use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;
use std::collections::BTreeSet;

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

lazy_static! {
    static ref EXISTS_OPERATOR: air::Expression =
        air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
            op: air::MQLOperator::Exists,
            args: vec![air::Expression::Literal(air::LiteralValue::Boolean(true))],
        });
    static ref TYPE_OPERATOR: air::Expression =
        air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
            op: air::MQLOperator::Type,
            args: vec![air::Expression::Literal(air::LiteralValue::String(
                "null".to_string()
            ))],
        });
    static ref NOT_NULL_OPERATOR: air::Expression =
        air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
            op: air::MQLOperator::Not,
            args: vec![TYPE_OPERATOR.clone()],
        });
}

#[derive(Default)]
struct MatchDesugarerPassVisitor;

impl MatchDesugarerPassVisitor {
    fn desugar_stage(&mut self, stage: &air::Match) -> Option<air::Stage> {
        self.generate_schema(&stage.expr).map(|new_expr| {
            air::Stage::Match(air::Match {
                source: stage.source.clone(),
                expr: Box::new(new_expr),
            })
        })
    }

    fn generate_schema(&mut self, expr: &air::Expression) -> Option<air::Expression> {
        let schema_vars = self.gather_vars_for_null_semantic_constraints(expr);

        if schema_vars.is_empty() {
            None
        } else {
            let mut constraints = Vec::new();
            for field_name in schema_vars.iter() {
                let mut constraint_exists = LinkedHashMap::new();
                constraint_exists.insert(field_name.clone(), EXISTS_OPERATOR.clone());
                let constraint_exists_entry = UniqueLinkedHashMap::from(constraint_exists);
                constraints.push(air::Expression::Document(constraint_exists_entry));

                let mut constraint_not = LinkedHashMap::new();
                constraint_not.insert(field_name.clone(), NOT_NULL_OPERATOR.clone());
                let constraint_not_entry = UniqueLinkedHashMap::from(constraint_not);
                constraints.push(air::Expression::Document(constraint_not_entry));
            }
            Some(air::Expression::MQLSemanticOperator(
                air::MQLSemanticOperator {
                    op: air::MQLOperator::And,
                    args: constraints,
                },
            ))
        }
    }

    fn gather_vars_for_null_semantic_constraints(
        &mut self,
        expr: &air::Expression,
    ) -> BTreeSet<String> {
        let mut visitor = FieldRefGatherer::new();
        visitor.run(expr);
        visitor.schema_vars
    }

    fn desugar_sql_ops_in_match(&mut self, m: &air::Match) -> air::Match {
        let expr = self.desugar_expr(&m.expr);
        air::Match {
            source: m.source.clone(),
            expr: Box::new(expr),
        }
    }

    #[allow(clippy::only_used_in_recursion)]
    fn desugar_expr(&mut self, expr: &air::Expression) -> air::Expression {
        match expr {
            air::Expression::SQLSemanticOperator(sql_op) => {
                let desugared_args = sql_op
                    .args
                    .iter()
                    .map(|arg| self.desugar_expr(arg))
                    .collect();
                if let Some(mql_op) = match_sql_to_mql_op(sql_op.op.clone()) {
                    air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                        op: mql_op,
                        args: desugared_args,
                    })
                } else {
                    air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
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
    fn visit_stage(&mut self, stage: air::Stage) -> air::Stage {
        let node = match stage {
            air::Stage::Match(m) => {
                let desugared_stage = match self.desugar_stage(&m) {
                    Some(schema_stage) => schema_stage,
                    None => return air::Stage::Match(m),
                };
                let optimized_match = self.desugar_sql_ops_in_match(&m);
                air::Stage::Match(air::Match {
                    source: Box::new(desugared_stage),
                    expr: optimized_match.expr,
                })
            }
            _ => stage,
        };
        node.walk(self)
    }
}

struct FieldRefGatherer {
    schema_vars: BTreeSet<String>,
}

impl FieldRefGatherer {
    fn new() -> Self {
        Self {
            schema_vars: BTreeSet::new(),
        }
    }
    fn run(&mut self, expr: &air::Expression) {
        self.visit_expression(expr.clone());
    }
    fn handle_refs(&mut self, expression: &air::Expression) {
        if let air::Expression::FieldRef(field_ref) = expression {
            self.schema_vars.insert(format!("{field_ref}"));
        }
    }
}

impl Visitor for FieldRefGatherer {
    fn visit_expression(&mut self, expr: air::Expression) -> air::Expression {
        if let air::Expression::SQLSemanticOperator(sql_op) = expr.clone() {
            for arg in &sql_op.args {
                match arg {
                    air::Expression::FieldRef(_) => {
                        self.handle_refs(arg);
                    }
                    air::Expression::Array(arr) => {
                        for elem in arr {
                            self.handle_refs(elem);
                        }
                    }
                    _ => {}
                }
            }
        }
        expr.walk(self)
    }
}
