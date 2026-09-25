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

impl visitor_ref::VisitorRef for LiteralVisitor {
    fn visit_identifier_expr(&mut self, _: &IdentifierExpr) {
        self.is_literal = false;
    }
}

// is_literal returns if a given Expression is a literal, meaning
// that it does not contain an Identifier.
pub fn is_literal(node: &Expression) -> bool {
    let mut visitor = LiteralVisitor::new();
    node.walk_ref(&mut visitor);
    visitor.is_literal
}

// returns if all the passed Expressions are literal.
pub fn are_literal(ve: &Vec<Expression>) -> bool {
    ve.iter().all(|e| is_literal(e))
}
