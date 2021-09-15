use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};

/// Finds each select query without a FROM clause and rewrites it to have an explicit `FROM [{}] AS _dual`.
pub struct ImplicitFromRewritePass;

impl Pass for ImplicitFromRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = ImplicitFromRewriteVisitor;
        Ok(query.walk(&mut visitor))
    }
}

/// The visitor that performs the rewrites for the `ImplicitFromRewritePass`.
struct ImplicitFromRewriteVisitor;

impl Visitor for ImplicitFromRewriteVisitor {
    fn visit_select_query(&mut self, node: ast::SelectQuery) -> ast::SelectQuery {
        use ast::*;
        let node = node.walk(self);
        let dual_source = Datasource::Array(ArraySource {
            array: vec![Expression::Document(Vec::new())],
            alias: "_dual".to_string(),
        });
        SelectQuery {
            from_clause: node.from_clause.or(Some(dual_source)),
            ..node
        }
    }
}
