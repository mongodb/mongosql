use crate::mir::visitor::Visitor;
use crate::mir::Expression;

/// A visitor that checks if an expression contains a subquery.
///
/// This struct implements the Visitor trait and traverses down the expressions
/// to determine if it contains any subquery-related expressions.
#[derive(Default)]
pub(crate) struct ContainsSubqueryVisitor {
    pub(crate) contains_subquery: bool,
}
impl Visitor for ContainsSubqueryVisitor {
    fn visit_expression(&mut self, expr: Expression) -> Expression {
        match expr {
            Expression::Subquery(e) => {
                self.contains_subquery = true;
                Expression::Subquery(e)
            }
            Expression::Exists(e) => {
                self.contains_subquery = true;
                Expression::Exists(e)
            }
            Expression::SubqueryComparison(e) => {
                self.contains_subquery = true;
                Expression::SubqueryComparison(e)
            }
            Expression::Array(e) => Expression::Array(e.walk(self)),
            Expression::Cast(e) => Expression::Cast(e.walk(self)),
            Expression::DateFunction(e) => Expression::DateFunction(e.walk(self)),
            Expression::Document(e) => Expression::Document(e.walk(self)),
            Expression::FieldAccess(e) => Expression::FieldAccess(e.walk(self)),
            Expression::Is(e) => Expression::Is(e.walk(self)),
            Expression::Like(e) => Expression::Like(e.walk(self)),
            Expression::Literal(e) => Expression::Literal(e),
            Expression::Reference(e) => Expression::Reference(e.walk(self)),
            Expression::ScalarFunction(e) => Expression::ScalarFunction(e.walk(self)),
            Expression::SearchedCase(e) => Expression::SearchedCase(e.walk(self)),
            Expression::SimpleCase(e) => Expression::SimpleCase(e.walk(self)),
            Expression::TypeAssertion(e) => Expression::TypeAssertion(e.walk(self)),
            Expression::MQLIntrinsicFieldExistence(e) => {
                Expression::MQLIntrinsicFieldExistence(e.walk(self))
            }
        }
    }
}
