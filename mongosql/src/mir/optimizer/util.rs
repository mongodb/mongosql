use crate::mir::{visitor::Visitor, Expression, FieldPath};
use std::collections::HashSet;

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

/// insert_field_path_and_all_ancestors is a helper function for gathering field uses. Given a
/// FieldPath, which may or may not include multiple components, this function includes the path
/// and all ancestor paths in the provided mutable set of paths. This is important since this
/// function is used by the use_def_analysis "field_uses" method, which is used to determine whether
/// a stage can be moved above another. If we did not include ancestors in this list, it would be
/// possible to erroneously move a stage above another stage that defines the ancestor field.
///
/// For example, for field path "foo.a.b.c" this function inserts "foo.a.b.c", "foo.a.b" and "foo.a"
/// assuming "foo" is the FieldPath "key" and ["a", "b", "c"] are the fields.
pub(crate) fn insert_field_path_and_all_ancestors(
    field_uses: &mut HashSet<FieldPath>,
    fp: FieldPath,
) {
    let mut fields = fp.fields.clone();
    while !fields.is_empty() {
        field_uses.insert(FieldPath {
            key: fp.key.clone(),
            fields: fields.clone(),
            // We need to assume nullability for each field up the chain based on the final
            // nullability because we have no other information here. Fortunately, the FieldPaths
            // produced by this function are never used for their nullability data, just for their
            // names.
            is_nullable: fp.is_nullable,
        });
        fields.pop();
    }
}
