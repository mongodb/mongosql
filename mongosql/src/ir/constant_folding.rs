use crate::ir::{self, visitor::Visitor, Expression};

struct ConstantFoldExprVisitor;

impl Default for ConstantFoldExprVisitor {
    fn default() -> Self {
        Self {}
    }
}

impl Visitor for ConstantFoldExprVisitor {
    fn visit_expression(&mut self, e: ir::Expression) -> ir::Expression {
        let e = e.walk(self);
        match e {
            Expression::Array(_) => e,
            Expression::Cast(_) => e,
            Expression::Document(_) => e,
            Expression::Exists(_) => e,
            Expression::FieldAccess(_) => e,
            Expression::Is(_) => e,
            Expression::Like(_) => e,
            Expression::Literal(_) => e,
            Expression::Reference(_) => e,
            Expression::ScalarFunction(_) => e,
            Expression::SearchedCase(_) => e,
            Expression::SimpleCase(_) => e,
            Expression::SubqueryComparison(_) => e,
            Expression::SubqueryExpression(_) => e,
            Expression::TypeAssertion(_) => e,
        }
    }

    fn visit_stage(&mut self, st: ir::Stage) -> ir::Stage {
        use ir::definitions::*;
        let st = st.walk(self);
        match st {
            Stage::Array(_) => st,
            Stage::Collection(_) => st,
            Stage::Filter(_) => st,
            Stage::Group(_) => st,
            Stage::Join(_) => st,
            Stage::Limit(_) => st,
            Stage::Offset(_) => st,
            Stage::Project(_) => st,
            Stage::Set(_) => st,
            Stage::Sort(_) => st,
        }
    }
}

pub fn fold_constants(st: ir::Stage) -> ir::Stage {
    let mut cf = ConstantFoldExprVisitor::default();
    cf.visit_stage(st)
}
