use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};
use std::collections::HashMap;

pub struct GroupBySelectAliasRewritePass;

impl Pass for GroupBySelectAliasRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = GroupBySelectAliasVisitor::default();
        let query = visitor.visit_query(query);
        Ok(query)
    }
}

#[derive(Default)]
struct GroupBySelectAliasVisitor;

impl Visitor for GroupBySelectAliasVisitor {
    fn visit_query(&mut self, query: ast::Query) -> ast::Query {
        // Gather information about select expression aliases and identifier
        // group keys.
        let mut alias_matcher = MatchingAliasFinder::default();
        let query = query.walk(&mut alias_matcher);

        // Rewrite group keys and select expressions as needed given the
        // information gathered above.
        let mut rewriter = alias_matcher.rewriter();
        query.walk(&mut rewriter)
    }
}

#[derive(Default)]
struct MatchingAliasFinder {
    aliased_select_exprs: HashMap<String, ast::AliasedExpr>,
    group_key_identifiers: Vec<String>,
    skip_subqueries: bool,
}

impl MatchingAliasFinder {
    fn rewriter(self) -> Rewriter {
        let mut select_exprs_by_group_key_ident = HashMap::new();

        for gk in self.group_key_identifiers.into_iter() {
            if let Some(ae) = self.aliased_select_exprs.get(&gk) {
                select_exprs_by_group_key_ident.insert(gk.clone(), ae.clone());
            }
        }

        Rewriter {
            select_exprs_by_group_key_ident,
            skip_subqueries: false,
        }
    }
}

impl Visitor for MatchingAliasFinder {
    fn visit_query(&mut self, subquery: ast::Query) -> ast::Query {
        if self.skip_subqueries {
            // Don't walk into subqueries
            return subquery;
        }
        let q = subquery.walk(self);
        self.skip_subqueries = true;
        q
    }
    fn visit_select_expression(
        &mut self,
        select_expr: ast::SelectExpression,
    ) -> ast::SelectExpression {
        if let ast::SelectExpression::Expression(ast::OptionallyAliasedExpr::Aliased(ref ae)) =
            select_expr
        {
            self.aliased_select_exprs
                .insert(ae.alias.clone(), ae.clone());
        };
        select_expr
    }

    fn visit_group_by_clause(&mut self, group_by: ast::GroupByClause) -> ast::GroupByClause {
        // Do not perform this rewrite if there are any aggregations.
        if !group_by.aggregations.is_empty() {
            return group_by;
        }

        for expr in group_by.keys.iter() {
            if let ast::OptionallyAliasedExpr::Unaliased(ast::Expression::Identifier(ident)) = expr
            {
                self.group_key_identifiers.push(ident.clone());
            }
        }
        group_by
    }
}

struct Rewriter {
    select_exprs_by_group_key_ident: HashMap<String, ast::AliasedExpr>,
    skip_subqueries: bool,
}

impl Visitor for Rewriter {
    fn visit_query(&mut self, subquery: ast::Query) -> ast::Query {
        if self.skip_subqueries {
            // Don't walk into subqueries
            return subquery;
        }
        let q = subquery.walk(self);
        self.skip_subqueries = true;
        q
    }
    fn visit_select_expression(
        &mut self,
        select_expr: ast::SelectExpression,
    ) -> ast::SelectExpression {
        match select_expr {
            ast::SelectExpression::Expression(ast::OptionallyAliasedExpr::Aliased(
                ast::AliasedExpr { alias, .. },
            )) if self.select_exprs_by_group_key_ident.contains_key(&alias) => {
                ast::SelectExpression::Expression(ast::OptionallyAliasedExpr::Unaliased(
                    ast::Expression::Identifier(alias),
                ))
            }
            _ => select_expr,
        }
    }

    fn visit_group_by_clause(&mut self, group_by: ast::GroupByClause) -> ast::GroupByClause {
        // Do not perform this rewrite if there are any aggregations.
        if !group_by.aggregations.is_empty() {
            return group_by;
        }
        let keys = group_by
            .keys
            .into_iter()
            .map(|expr| match expr {
                ast::OptionallyAliasedExpr::Unaliased(ast::Expression::Identifier(ref ident)) => {
                    if let Some(ae) = self.select_exprs_by_group_key_ident.get(ident) {
                        ast::OptionallyAliasedExpr::Aliased(ae.clone())
                    } else {
                        expr
                    }
                }
                _ => expr,
            })
            .collect();

        ast::GroupByClause {
            keys,
            aggregations: group_by.aggregations,
        }
    }
}
