use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
    CollectionSource,
};

/// Adds aliases to any AliasedExprs or CollectionSources that are missing one. Since we can't
/// distinguish between top-level field references and SubpathExprs during syntactic rewrites, we
/// will skip generating aliases for any GROUP BY key that resembles a top-level field reference, i.e.
/// any expression with exactly one dot.
pub struct AddAliasRewritePass;

impl Pass for AddAliasRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = AddAliasRewriteVisitor::default();
        Ok(query.walk(&mut visitor))
    }
}

/// The visitor that performs the rewrites for the `AddAliasRewritePass`.
struct AddAliasRewriteVisitor {
    is_group_by: bool,
    counter: i32,
}

impl AddAliasRewriteVisitor {
    /// update_alias will generate an alias for the given Expression.  If the expression is a
    /// simple or compound identifier, set the alias to the identifier sans its qualifiers, if any.
    /// Otherwise, generate an alias of the form _<counter>.
    fn update_alias(&mut self, expr: ast::Expression) -> ast::OptionallyAliasedExpr {
        use ast::*;
        let new_alias = match expr {
            Expression::Subpath(SubpathExpr {
                expr: _,
                ref subpath,
            }) => subpath.to_string(),
            Expression::Identifier(ref id) => id.to_string(),
            _ => format!("_{}", self.counter),
        };
        OptionallyAliasedExpr::Aliased(AliasedExpr {
            expr,
            alias: new_alias,
        })
    }

    /// update_group_by_alias will generate an alias of the form _groupKey<counter> for the given
    /// Expression if it is (1) missing an alias and (2) in a GroupByClause where the relevant
    /// GROUP BY key is a non-reference expression. If both conditions are false, the Expression
    /// will be returned without an alias.
    fn update_group_by_alias(&mut self, expr: ast::Expression) -> ast::OptionallyAliasedExpr {
        use ast::*;
        match expr {
            Expression::Subpath(SubpathExpr {
                expr: ref sub_expr,
                ref subpath,
            }) => {
                // Don't generate aliases for GROUP BY references with a single dot;
                // these expressions only get disambiguated during algebrization
                match **sub_expr {
                    Expression::Identifier(_) => ast::OptionallyAliasedExpr::Unaliased(expr),
                    _ => ast::OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: subpath.to_string(),
                        expr,
                    }),
                }
            }
            Expression::Identifier(_) => ast::OptionallyAliasedExpr::Unaliased(expr),
            _ => ast::OptionallyAliasedExpr::Aliased(AliasedExpr {
                expr,
                alias: format!("_groupKey{}", self.counter),
            }),
        }
    }
}

impl Default for AddAliasRewriteVisitor {
    fn default() -> Self {
        Self {
            is_group_by: false,
            counter: 1,
        }
    }
}

impl Visitor for AddAliasRewriteVisitor {
    fn visit_select_query(&mut self, node: ast::SelectQuery) -> ast::SelectQuery {
        // Create a new visitor to ensure that the counter remains accurate in case of subqueries
        let mut sub_visitor = AddAliasRewriteVisitor::default();
        node.walk(&mut sub_visitor)
    }

    fn visit_collection_source(&mut self, node: ast::CollectionSource) -> ast::CollectionSource {
        let node = node.walk(self);
        let coll = node.collection.as_str();
        match node.alias {
            Some(_) => node,
            None => CollectionSource {
                database: node.database,
                collection: coll.to_string(),
                alias: Some(coll.to_string()),
            },
        }
    }

    fn visit_group_by_clause(&mut self, node: ast::GroupByClause) -> ast::GroupByClause {
        self.counter = 1;
        self.is_group_by = true;
        let node = node.walk(self);
        self.is_group_by = false;
        node
    }

    fn visit_optionally_aliased_expr(
        &mut self,
        node: ast::OptionallyAliasedExpr,
    ) -> ast::OptionallyAliasedExpr {
        let node = node.walk(self);
        let ret = match node {
            ast::OptionallyAliasedExpr::Aliased(_) => node,
            ast::OptionallyAliasedExpr::Unaliased(e) => {
                if self.is_group_by {
                    self.update_group_by_alias(e)
                } else {
                    self.update_alias(e)
                }
            }
        };
        self.counter += 1;
        ret
    }
}
