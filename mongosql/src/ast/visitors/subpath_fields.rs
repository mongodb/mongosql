use crate::ast::*;

#[derive(Default)]
struct SubpathFieldVisitor {
    subpath_fields: Vec<Vec<String>>,
}

impl visitor_ref::VisitorRef for SubpathFieldVisitor {
    fn visit_subpath_expr(&mut self, node: &SubpathExpr) {
        let mut entries = vec![];
        self.collect_subpath_entries(&node.expr, &mut entries);
        entries.push(node.subpath.clone());
        self.subpath_fields.push(entries);
    }
}

impl SubpathFieldVisitor {
    #[allow(clippy::only_used_in_recursion)]
    /// collect_subpath_entries recursively collects subpath parts
    fn collect_subpath_entries(&mut self, expr: &Expression, entries: &mut Vec<String>) {
        match expr {
            Expression::Subpath(subpath_expr) => {
                self.collect_subpath_entries(&subpath_expr.expr, entries);
                entries.push(subpath_expr.subpath.clone());
            }
            Expression::Identifier(ident) => {
                entries.push(ident.clone());
            }
            _ => (),
        }
    }
}

pub fn get_subpath_fields(query: Query) -> Vec<Vec<String>> {
    let mut visitor = SubpathFieldVisitor::default();
    query.walk_ref(&mut visitor);
    visitor.subpath_fields
}
