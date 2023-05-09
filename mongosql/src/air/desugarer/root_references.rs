use crate::air::{
    desugarer::{Pass, Result},
    visitor::Visitor,
    Expression,
    Expression::*,
    Stage, Variable,
};

/// Desugars any Variable expressions that start with "ROOT" to omit the
/// ROOT reference and make them FieldRef expressions.
pub struct RootReferenceDesugarerPass;

impl Pass for RootReferenceDesugarerPass {
    fn apply(&self, pipeline: Stage) -> Result<Stage> {
        let mut visitor = RootReferenceDesugarerPassVisitor::default();
        Ok(visitor.visit_stage(pipeline))
    }
}

#[derive(Default)]
struct RootReferenceDesugarerPassVisitor;

impl RootReferenceDesugarerPassVisitor {
    fn desugar_variable(&self, v: Variable) -> Expression {
        let v_as_string = format!("{v}");

        if v_as_string.starts_with("ROOT.") {
            FieldRef(v_as_string.chars().skip(5).collect::<String>().into())
        } else {
            Variable(v)
        }
    }
}

impl Visitor for RootReferenceDesugarerPassVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        match node {
            Variable(v) => self.desugar_variable(v),
            _ => node.walk(self),
        }
    }
}
