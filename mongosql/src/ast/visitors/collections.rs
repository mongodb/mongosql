use crate::ast::*;

#[derive(Default)]
struct CollectionVisitor {
    collections: Vec<CollectionSource>,
}

impl visitor::Visitor for CollectionVisitor {
    fn visit_collection_source(&mut self, node: CollectionSource) -> CollectionSource {
        self.collections.push(node.clone());
        node
    }
}

pub fn get_collection_sources(query: Query) -> Vec<CollectionSource> {
    let mut visitor = CollectionVisitor::default();
    query.walk(&mut visitor);
    visitor.collections
}
