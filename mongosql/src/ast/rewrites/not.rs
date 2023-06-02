use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};

/// Finds all Unary Expressions where the operator is not, and the expression is a comparison,
/// and rewrites as the negation of the internal comparison expression.
/// Example: NOT (a = b) becomes a <> b
pub struct NotComparisonRewritePass;

impl Pass for NotComparisonRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = NotComparisonVisitor::default();
        Ok(query.walk(&mut visitor))
    }
}

impl ast::ComparisonOp {
    fn negation(self) -> ast::ComparisonOp {
        match self {
            ast::ComparisonOp::Eq => ast::ComparisonOp::Neq,
            ast::ComparisonOp::Neq => ast::ComparisonOp::Eq,
            ast::ComparisonOp::Lt => ast::ComparisonOp::Gte,
            ast::ComparisonOp::Lte => ast::ComparisonOp::Gt,
            ast::ComparisonOp::Gt => ast::ComparisonOp::Lte,
            ast::ComparisonOp::Gte => ast::ComparisonOp::Lt,
        }
    }
}

/// The visitor that performs the rewrites for the `NotComparisonRewritePass`.
#[derive(Default)]
struct NotComparisonVisitor;

impl Visitor for NotComparisonVisitor {
    fn visit_expression(&mut self, node: ast::Expression) -> ast::Expression {
        use ast::*;
        let node = match node.clone() {
            Expression::Unary(UnaryExpr {
                op: UnaryOp::Not,
                expr,
            }) => match *expr {
                Expression::Binary(BinaryExpr {
                    left,
                    op: BinaryOp::Comparison(comparison_op),
                    right,
                }) => Expression::Binary(BinaryExpr {
                    left,
                    op: BinaryOp::Comparison(comparison_op.negation()),
                    right,
                }),
                _ => node,
            },
            _ => node,
        };
        node.walk(self)
    }
}
