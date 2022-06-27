use crate::ast::*;
// LiteralVisitor is an ast visitor that reports if a given expression
// is literal. A literal expression cannot contain an Identifier.
struct LiteralVisitor {
    is_literal: bool,
}

impl LiteralVisitor {
    fn new() -> LiteralVisitor {
        LiteralVisitor { is_literal: true }
    }
}

impl visitor::Visitor for LiteralVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        match node {
            Expression::Identifier(_) => {
                self.is_literal = false;
                node
            }
            _ => node.walk(self),
        }
    }
}

// is_literal returns if a given Expression is a literal, meaning
// that it does not contain an Identifier.
pub fn is_literal(node: Expression) -> (Expression, bool) {
    let mut visitor = LiteralVisitor::new();
    let out = node.walk(&mut visitor);
    (out, visitor.is_literal)
}

// returns if all the passed Expressions are literal.
pub fn are_literal(ve: Vec<Expression>) -> (Vec<Expression>, bool) {
    if let (Expression::Array(ve), array_is_literal) = is_literal(Expression::Array(ve)) {
        (ve, array_is_literal)
    } else {
        unreachable!()
    }
}
