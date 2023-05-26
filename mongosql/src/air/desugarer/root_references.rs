use crate::air::{
    desugarer::{Pass, Result},
    visitor::Visitor,
    Expression,
    Expression::*,
    SortSpecification, Stage, Variable,
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

        let (v, has_root) = Self::remove_root_from_string(v_as_string);
        if has_root {
            FieldRef(v.into())
        } else {
            Variable(v.into())
        }
    }

    fn remove_root_from_string(var: String) -> (String, bool) {
        if var.starts_with("ROOT.") {
            (var.chars().skip(5).collect::<String>(), true)
        } else {
            (var, false)
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

    fn visit_sort_specification(&mut self, spec: SortSpecification) -> SortSpecification {
        match spec {
            SortSpecification::Asc(s) => SortSpecification::Asc(Self::remove_root_from_string(s).0),
            SortSpecification::Desc(s) => {
                SortSpecification::Desc(Self::remove_root_from_string(s).0)
            }
        }
    }
}
