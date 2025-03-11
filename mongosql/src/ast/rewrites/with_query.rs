use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};
use std::collections::BTreeMap;

/// Finds all Unary Expressions where the operator is not, and the expression is a comparison,
/// and rewrites as the negation of the internal comparison expression.
/// Example: NOT (a = b) becomes a <> b
pub struct WithQueryRewritePass;

impl Pass for WithQueryRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = WithQueryVisitor;
        Ok(visitor.visit_query(query))
    }
}

struct CollectionReplaceVisitor<'a> {
    theta: &'a BTreeMap<String, ast::Query>,
}

impl Visitor for CollectionReplaceVisitor<'_> {
    fn visit_datasource(&mut self, mut datasource: ast::Datasource) -> ast::Datasource {
        match datasource {
            ast::Datasource::Collection(ast::CollectionSource {
                ref mut database,
                ref mut collection,
                ref mut alias,
            }) => {
                // If the database is specified, this cannot be a NamedQuery from a WITH statement,
                // this is how a user can unshadow a namespace specifically.
                if database.is_some() {
                    return ast::Datasource::Collection(ast::CollectionSource {
                        database: std::mem::take(database),
                        collection: std::mem::take(collection),
                        alias: std::mem::take(alias),
                    });
                }
                // If the query is in theta, we replace the datasource with the query
                if let Some(query) = self.theta.get(collection) {
                    return ast::Datasource::Derived(ast::DerivedSource {
                        // This clone will always be necessary because we are moving the query out of
                        // the theta map, and we need to keep the query around in case it is used
                        // again.
                        query: Box::new(query.clone()),
                        alias: std::mem::take(alias).unwrap_or_else(|| std::mem::take(collection)),
                    });
                }
                // Otherwise, return the original datasource. The borrow checker makes us
                // reconstruct the struct, but at least we are avoiding cloning the strings.
                ast::Datasource::Collection(ast::CollectionSource {
                    database: std::mem::take(database),
                    collection: std::mem::take(collection),
                    alias: std::mem::take(alias),
                })
            }
            // a derived query could still have a use of a WITH query
            _ => datasource.walk(self),
        }
    }
}

/// The visitor that performs the rewrites for the `WithQueryRewritePass`.
#[derive(Default)]
struct WithQueryVisitor;

impl Visitor for WithQueryVisitor {
    fn visit_query(&mut self, query: ast::Query) -> ast::Query {
        match query {
            ast::Query::With(ast::WithQuery { queries, body }) => {
                let mut theta = BTreeMap::new();
                for query in queries {
                    let ast::NamedQuery { name, query } = query;
                    let query = CollectionReplaceVisitor { theta: &theta }.visit_query(query);
                    theta.insert(name, query);
                }
                CollectionReplaceVisitor { theta: &theta }.visit_query(*body)
            }
            // we currently only allow WITH queries at the top level, so we don't need to walk
            _ => query,
        }
    }
}
