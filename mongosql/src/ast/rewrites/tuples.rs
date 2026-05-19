use crate::ast::{
    rewrites::{Pass, Result},
    visitor::Visitor,
    *,
};

/// Finds all one-element tuples like (x) or (1+2) and replaces them with the underlying expression.
/// Intended to be used after `InTupleRewritePass` to avoid rewriting one-element `IN` tuples.
pub struct SingleTupleRewritePass;

impl Pass for SingleTupleRewritePass {
    fn apply(&self, query: Query) -> Result<Query> {
        Ok(query.walk(&mut SingleTupleRewriteVisitor))
    }
}

/// The visitor that performs the rewrites for `SingleTupleRewritePass`.
/// Also used in `pretty_print_test`.
pub struct SingleTupleRewriteVisitor;

impl Visitor for SingleTupleRewriteVisitor {
    fn visit_expression(&mut self, e: Expression) -> Expression {
        match e {
            // For IN/NotIn, the RHS tuple is an array operand and must not be
            // unwrapped, even if it contains only one element. Walk the LHS and each tuple
            // element individually so nested single-element tuples inside them are still
            // collapsed, but preserve the outer tuple structure on the RHS.
            Expression::Binary(BinaryExpr { left, op, right })
                if op == BinaryOp::In || op == BinaryOp::NotIn =>
            {
                let left = Box::new(self.visit_expression(*left));
                let right = Box::new(match *right {
                    Expression::Tuple(elems) => Expression::Tuple(
                        elems
                            .into_iter()
                            .map(|e| self.visit_expression(e))
                            .collect(),
                    ),
                    other => self.visit_expression(other),
                });
                Expression::Binary(BinaryExpr { left, op, right })
            }
            Expression::Tuple(mut t) if t.len() == 1 => self.visit_expression(t.pop().unwrap()),
            _ => e.walk(self),
        }
    }
}
