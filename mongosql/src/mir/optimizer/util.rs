use crate::mir::{visitor_ref::VisitorRef, Expression, FieldPath};
use std::collections::HashSet;

/// A visitor that checks if an expression contains a subquery.
///
/// This struct implements the VisitorRef trait and traverses down the expressions
/// to determine if it contains any subquery-related expressions.
#[derive(Default)]
pub(crate) struct ContainsSubqueryVisitor {
    pub(crate) contains_subquery: bool,
}
impl VisitorRef for ContainsSubqueryVisitor {
    fn visit_expression(&mut self, expr: &Expression) {
        match expr {
            Expression::Subquery(_) => {
                self.contains_subquery = true;
            }
            Expression::Exists(_) => {
                self.contains_subquery = true;
            }
            Expression::SubqueryComparison(_) => {
                self.contains_subquery = true;
            }
            Expression::Array(e) => e.walk_ref(self),
            Expression::Cast(e) => e.walk_ref(self),
            Expression::DateFunction(e) => e.walk_ref(self),
            Expression::Document(e) => e.walk_ref(self),
            Expression::FieldAccess(e) => e.walk_ref(self),
            Expression::ComputedFieldAccess(e) => e.walk_ref(self),
            Expression::Is(e) => e.walk_ref(self),
            Expression::Like(e) => e.walk_ref(self),
            Expression::Literal(_) => (),
            Expression::Reference(e) => e.walk_ref(self),
            Expression::ScalarFunction(e) => e.walk_ref(self),
            Expression::SearchedCase(e) => e.walk_ref(self),
            Expression::SimpleCase(e) => e.walk_ref(self),
            Expression::TypeAssertion(e) => e.walk_ref(self),
            Expression::HigherOrderFunction(e) => e.walk_ref(self),
            Expression::Variable(_) => (),
            Expression::MqlIntrinsicFieldExistence(e) => e.walk_ref(self),
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
/// For example, for a field path "foo.a.b.c" this function inserts "foo.a.b.c", "foo.a.b" and "foo.a"
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
