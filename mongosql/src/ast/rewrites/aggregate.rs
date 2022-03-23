use crate::ast::{
    self,
    pretty_print::PrettyPrint,
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
};
use linked_hash_map::LinkedHashMap;

pub struct AggregateRewritePass;

impl Pass for AggregateRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        // First, check for improper usage of aggregation functions
        // and return the error if one is found.
        let mut visitor = AggregateUsageCheckVisitor::default();
        let query = visitor.visit_query(query);
        if visitor.error.is_some() {
            return Err(visitor.error.unwrap());
        }

        // If there's no error from the first visitor, use the second
        // visitor to safely create aliases for any aggregation functions.
        let query = AggregateAliasingVisitor::default().visit_query(query);

        // Use the third visitor to rewrite aggregation functions that
        // use the ALL set quantifier.
        Ok(AggregateSetQuantifierVisitor::default().visit_query(query))
    }
}

///
/// A visitor that tracks improper usage of aggregation functions.
///
/// Can result in three possible errors:
///   - `Error::AggregationFunctionInGroupByKeyList`
///   - `Error::AggregateInGroupByAggListNotAliased`
///   - `Error::AggregationFunctionInGroupByAggListAndElsewhere`
///
#[derive(Default)]
pub struct AggregateUsageCheckVisitor {
    pub error: Option<Error>,

    pub in_group_by_key_list: bool,
    pub in_group_by_agg_func_list: bool,

    pub num_group_by_agg_funcs: u32,
    pub num_non_group_by_agg_funcs: u32,
}

impl AggregateUsageCheckVisitor {
    // `has_invalid_agg_mix` checks if there's an invalid mix of aggregation functions.
    //
    // An invalid mix occurs when at least one aggregation function is specified
    // in a `GROUP BY` clause's aggregation function list, along with at least one
    // aggregation function anywhere else.
    //
    // Examples:
    //   - SELECT * ... GROUP BY x                                     =>  valid
    //   - SELECT * ... GROUP BY x AGGREGATE SUM(x)                    =>  valid
    //   - SELECT SUM(x) ... GROUP BY x                                =>  valid
    //   - SELECT * ... GROUP BY x HAVING SUM(x) > 0                   =>  valid
    //   - SELECT SUM(x) ... GROUP BY x AGGREGATE SUM(x)               =>  invalid
    //   - SELECT * ... GROUP BY x AGGREGATE SUM(x) HAVING SUM(x) > 0  =>  invalid
    fn has_invalid_agg_mix(&self) -> bool {
        self.num_group_by_agg_funcs > 0 && self.num_non_group_by_agg_funcs > 0
    }
}

impl Visitor for AggregateUsageCheckVisitor {
    fn visit_query(&mut self, subquery: ast::Query) -> ast::Query {
        // Don't recurse into subqueries if there is an existing error.
        if self.error.is_some() {
            return subquery;
        }

        // Create a new visitor for each subquery walk, and copy
        // any errors from the subquery walk to the current visitor.
        let mut subquery_visitor = AggregateUsageCheckVisitor::default();
        let subquery = subquery.walk(&mut subquery_visitor);
        self.error = subquery_visitor.error;

        subquery
    }

    fn visit_select_query(&mut self, node: ast::SelectQuery) -> ast::SelectQuery {
        // First walk all of the clauses in the select query.
        let node = node.walk(self);

        // Return if we already have an error from walking the select query.
        if self.error.is_some() {
            return node;
        }

        // Check for an invalid mix of aggregation functions.
        if self.has_invalid_agg_mix() {
            self.error = Some(Error::AggregationFunctionInGroupByAggListAndElsewhere);
        }

        node
    }

    // Sets boolean values in the visitor accordingly before walking the `GROUP BY` clause's lists.
    //
    // This allows `visit_expression` to discern whether an aggregation function occurs inside a:
    // (1) `GROUP BY` clause's key list,
    // (2) `GROUP BY` clause's aggregation function list,
    // (3) or elsewhere.
    //
    // This is required for three error cases (see `has_invalid_agg_mix` above, and
    // `visit_aliased_expr` and `visit_expression` below).
    fn visit_group_by_clause(&mut self, node: ast::GroupByClause) -> ast::GroupByClause {
        use ast::*;

        self.in_group_by_key_list = true;
        let keys = node
            .keys
            .into_iter()
            .map(|vec_x| self.visit_optionally_aliased_expr(vec_x))
            .collect::<Vec<_>>();
        self.in_group_by_key_list = false;

        self.in_group_by_agg_func_list = true;
        let aggregations = node
            .aggregations
            .into_iter()
            .map(|vec_x| self.visit_aliased_expr(vec_x))
            .collect::<Vec<_>>();
        self.in_group_by_agg_func_list = false;

        GroupByClause { keys, aggregations }
    }

    fn visit_aliased_expr(&mut self, a: ast::AliasedExpr) -> ast::AliasedExpr {
        if self.error.is_some() {
            return a;
        }
        a.walk(self)
    }

    fn visit_expression(&mut self, e: ast::Expression) -> ast::Expression {
        use ast::*;
        match e {
            Expression::Function(ref f) if f.function.is_aggregation_function() => {
                if self.error.is_some() {
                    return e;
                }

                // It is not valid to specify an aggregation function in a `GROUP BY` key list.
                if self.in_group_by_key_list {
                    self.error = Some(Error::AggregationFunctionInGroupByKeyList);
                    return e;
                }

                // Record whether the function is in a `GROUP BY` aggregation function list or not.
                if self.in_group_by_agg_func_list {
                    self.num_group_by_agg_funcs += 1;
                } else {
                    self.num_non_group_by_agg_funcs += 1;
                }

                e
            }
            _ => e.walk(self),
        }
    }
}

///
/// A visitor that creates aliases for and de-duplicates aggregation functions defined
/// outside of a `GROUP BY` clause.
///
/// This visitor does not alias or de-duplicate top-level aggregation functions defined
/// in a `GROUP BY` clause's aggregation function list. This is because the
/// `AggregateUsageCheckVisitor` ensures that these functions are already aliased.
///
/// Alias ordering is determined by either the `SELECT` clause or the `HAVING` clause, whichever
/// is visited first. Generally, the `SELECT` clause is expected to be visited first.
///
pub struct AggregateAliasingVisitor {
    pub next_agg_id: u32,
    pub agg_funcs: LinkedHashMap<String, ast::AliasedExpr>,
    pub in_group_by_agg_func_list: bool,
}

impl Default for AggregateAliasingVisitor {
    fn default() -> Self {
        Self {
            next_agg_id: 1,
            agg_funcs: LinkedHashMap::new(),
            in_group_by_agg_func_list: false,
        }
    }
}

impl Visitor for AggregateAliasingVisitor {
    fn visit_query(&mut self, subquery: ast::Query) -> ast::Query {
        subquery.walk(&mut AggregateAliasingVisitor::default())
    }

    fn visit_select_query(&mut self, node: ast::SelectQuery) -> ast::SelectQuery {
        use ast::*;

        // Walk all of the clauses in the select query to search for aggregation functions.
        let node = node.walk(self);

        // Return the original select query if no aggregation functions were found.
        if self.agg_funcs.is_empty() {
            return node;
        }

        // If a `GROUP BY` clause already exists, use those existing keys.
        // Otherwise, create a new key list only containing the `NULL` literal.
        let keys = match node.group_by_clause {
            Some(g) => g.keys,
            None => vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                expr: Expression::Literal(Literal::Null),
                alias: "_groupKey1".to_string(),
            })],
        };

        // Return a select query containing a new `GROUP BY` clause with the aggregation function aliases.
        SelectQuery {
            group_by_clause: Some(GroupByClause {
                keys,
                aggregations: self.agg_funcs.iter().map(|(_, v)| v.clone()).collect(),
            }),
            ..node
        }
    }

    // Sets a boolean value in the visitor accordingly before walking the `GROUP BY`
    // clause's aggregation function list.
    //
    // This allows `visit_expression` to discern whether or not an aggregation function
    // occurs inside a `GROUP BY` clause's aggregation function list.
    fn visit_group_by_clause(&mut self, node: ast::GroupByClause) -> ast::GroupByClause {
        use ast::*;

        let keys = node
            .keys
            .into_iter()
            .map(|vec_x| self.visit_optionally_aliased_expr(vec_x))
            .collect::<Vec<_>>();

        self.in_group_by_agg_func_list = true;
        let aggregations = node
            .aggregations
            .into_iter()
            .map(|vec_x| self.visit_aliased_expr(vec_x))
            .collect::<Vec<_>>();
        self.in_group_by_agg_func_list = false;

        GroupByClause { keys, aggregations }
    }

    fn visit_expression(&mut self, e: ast::Expression) -> ast::Expression {
        use ast::*;
        match e {
            Expression::Function(f) if f.function.is_aggregation_function() => {
                // Walk first, in case the function's arguments contain any subquery expressions.
                let f = f.walk(self);

                // Do not alias top-level aggregation functions in a `GROUP BY` clause's aggregation
                // function list, since the `AggregateUsageCheckVisitor` ensures they are already aliased.
                if self.in_group_by_agg_func_list {
                    return Expression::Function(f);
                }

                // If the aggregation function has already been seen, get the existing alias from `agg_funcs`.
                // Otherwise, create a new alias for the function and store it in `agg_funcs`.
                // In both cases, replace the function with an Identifier containing the alias.
                let func_key = f.pretty_print().unwrap();
                match self.agg_funcs.get(&func_key) {
                    // We can safely unwrap the alias here because any value retrieved
                    // from `agg_funcs` would have been previously inserted with an alias.
                    Some(x) => Expression::Identifier(x.alias.clone()),
                    None => {
                        let new_agg_alias = format!("_agg{}", self.next_agg_id);
                        self.next_agg_id += 1;
                        self.agg_funcs.insert(
                            func_key,
                            AliasedExpr {
                                expr: Expression::Function(f),
                                alias: new_agg_alias.clone(),
                            },
                        );
                        Expression::Identifier(new_agg_alias)
                    }
                }
            }
            _ => e.walk(self),
        }
    }
}

/// Rewrites SetQuantifier::All aggregation functions into unmodified
/// aggregation functions.
#[derive(Default)]
pub struct AggregateSetQuantifierVisitor;

impl Visitor for AggregateSetQuantifierVisitor {
    fn visit_function_expr(&mut self, f: ast::FunctionExpr) -> ast::FunctionExpr {
        use ast::{FunctionExpr, SetQuantifier};
        // Walk first in case there are nested aggregation functions.
        let f = f.walk(self);
        if f.function.is_aggregation_function() {
            match f.set_quantifier {
                Some(SetQuantifier::All) => FunctionExpr {
                    function: f.function,
                    args: f.args,
                    set_quantifier: None,
                },
                _ => f,
            }
        } else {
            f
        }
    }
}
