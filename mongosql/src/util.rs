use crate::ast;

#[macro_export]
macro_rules! map {
	($($key:expr => $val:expr),* $(,)?) => {
		std::iter::Iterator::collect(std::array::IntoIter::new([
			$({
				($key, $val)
			},)*
		]))
	};
}

#[macro_export]
macro_rules! set {
	($($val:expr),* $(,)?) => {
		std::iter::Iterator::collect(std::array::IntoIter::new([
			$({
				($val)
			},)*
		]))
	};
}

// LiteralVisitor is an ast visitor that reports if a given expression
// is a literal. A literal can consist only of Arrays, Documents, and
// scalar Literals.
struct LiteralVisitor {
    is_literal: bool,
}

impl LiteralVisitor {
    fn new() -> LiteralVisitor {
        LiteralVisitor { is_literal: true }
    }
}

impl ast::visitor::Visitor for LiteralVisitor {
    fn visit_expression(&mut self, node: ast::Expression) -> ast::Expression {
        match node {
            ast::Expression::Literal(_) => node,
            ast::Expression::Array(_) | ast::Expression::Document(_) => node.walk(self),
            _ => {
                self.is_literal = false;
                node
            }
        }
    }
}

// is_literal returns if a given ast::Expression is a literal, meaning
// that it consists only of Arrays, Documents, and scalar Literals.
pub fn is_literal(node: ast::Expression) -> (ast::Expression, bool) {
    let mut visitor = LiteralVisitor::new();
    let out = node.walk(&mut visitor);
    (out, visitor.is_literal)
}

// returns if all the passed Expressions are literal.
pub fn are_literal(ve: Vec<ast::Expression>) -> (Vec<ast::Expression>, bool) {
    if let (ast::Expression::Array(ve), array_is_literal) = is_literal(ast::Expression::Array(ve)) {
        (ve, array_is_literal)
    } else {
        unreachable!()
    }
}
