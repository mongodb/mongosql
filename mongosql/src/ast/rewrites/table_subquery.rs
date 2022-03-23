use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};

/// Rewrites IN and NOT IN expressions with a subquery on the RHS
/// to subquery comparison expressions.
pub struct TableSubqueryRewritePass;

impl Pass for TableSubqueryRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        Ok(query.walk(&mut TableSubqueryRewriteVisitor))
    }
}

/// The visitor that performs the rewrites for the `TableSubqueryRewritePass`.
struct TableSubqueryRewriteVisitor;

impl TableSubqueryRewriteVisitor {}

impl Visitor for TableSubqueryRewriteVisitor {
    fn visit_expression(&mut self, e: ast::Expression) -> ast::Expression {
        use ast::*;
        match e {
            Expression::Binary(ref b) if b.op == BinaryOp::In || b.op == BinaryOp::NotIn => {
                if let Expression::Subquery(s) = &*b.right {
                    let (op, quantifier) = match b.op {
                        BinaryOp::In => (ComparisonOp::Eq, SubqueryQuantifier::Any),
                        _ => (ComparisonOp::Neq, SubqueryQuantifier::All),
                    };
                    Expression::SubqueryComparison(SubqueryComparisonExpr {
                        expr: b.left.clone(),
                        op,
                        quantifier,
                        subquery: s.clone(),
                    })
                } else {
                    e.walk(self)
                }
            }
            _ => e.walk(self),
        }
    }
}
