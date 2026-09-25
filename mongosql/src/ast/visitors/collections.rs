use crate::ast::*;

#[derive(Default)]
struct CollectionVisitor {
    collections: Vec<CollectionSource>,
}

impl visitor_ref::VisitorRef for CollectionVisitor {
    fn visit_collection_source(&mut self, node: &CollectionSource) {
        self.collections.push(node.clone());
    }
}

pub fn get_collection_sources(query: &Query) -> Vec<CollectionSource> {
    let mut visitor = CollectionVisitor::default();
    query.walk_ref(&mut visitor);
    visitor.collections
}
