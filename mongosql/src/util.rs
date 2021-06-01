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

// ConstantVisitor is a simple ast visitor that reports
// if a given expression can be evaluated at compile time:
// meaning it can contain no Indentifiers or Subqueries.
struct ConstantVisitor {
    is_constant: bool,
}

#[allow(dead_code)]
impl ConstantVisitor {
    fn new() -> ConstantVisitor {
        ConstantVisitor { is_constant: true }
    }
}

impl ast::visitor::Visitor for ConstantVisitor {
    fn visit_expression(&mut self, node: ast::Expression) -> ast::Expression {
        match node {
            ast::Expression::Identifier(_)
            | ast::Expression::Exists(_)
            | ast::Expression::Subquery(_)
            | ast::Expression::SubqueryComparison(_) => {
                self.is_constant = false;
                // Do not walk if we have found a non-constant.
                node
            }
            _ => node.walk(self),
        }
    }
}

// is_constant returns if a given ast::Expression is a constant, meaning
// that it could be entirely evaluated to a value at compile time.
#[allow(dead_code)]
pub fn is_constant(node: ast::Expression) -> (ast::Expression, bool) {
    let mut visitor = ConstantVisitor::new();
    let out = node.walk(&mut visitor);
    (out, visitor.is_constant)
}

// returns if all the passed Expressions are constant.
#[allow(dead_code)]
pub fn are_constant(ve: Vec<ast::Expression>) -> (Vec<ast::Expression>, bool) {
    if let (ast::Expression::Array(ve), array_is_constant) = is_constant(ast::Expression::Array(ve))
    {
        (ve, array_is_constant)
    } else {
        unreachable!()
    }
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
