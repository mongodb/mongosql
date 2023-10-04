use crate::ast::{
    self,
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
};

/// Rewrites non-VALUE SELECT to SELECT VALUE.
pub struct SelectRewritePass;

impl Pass for SelectRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = SelectRewriteVisitor::default();
        let rewritten = query.walk(&mut visitor);
        match visitor.error {
            Some(err) => Err(err),
            None => Ok(rewritten),
        }
    }
}

/// The visitor that performs the rewrites for the `SelectRewritePass`.
struct SelectRewriteVisitor {
    allow_select_value: bool,
    error: Option<Error>,
}

impl Default for SelectRewriteVisitor {
    fn default() -> Self {
        Self {
            allow_select_value: true,
            error: None,
        }
    }
}

impl SelectRewriteVisitor {}

impl Visitor for SelectRewriteVisitor {
    fn visit_datasource(&mut self, node: ast::Datasource) -> ast::Datasource {
        self.allow_select_value = true;
        node.walk(self)
    }

    fn visit_expression(&mut self, node: ast::Expression) -> ast::Expression {
        // We disallow SELECT VALUE(S) when walking into subquery expressions and comparisons.
        // We don't need to disallow EXISTS subqueries from using a top-level SELECT VALUE(S)
        // since those expressions just measure the cardinality of the subquery's result set.
        match node {
            ast::Expression::Subquery(_) | ast::Expression::SubqueryComparison(_) => {
                self.allow_select_value = false
            }
            ast::Expression::Exists(_) => self.allow_select_value = true,
            _ => {}
        }
        node.walk(self)
    }

    /// visit_select_body assumes that every expression in a standard select expression list
    /// has an alias.
    fn visit_select_body(&mut self, node: ast::SelectBody) -> ast::SelectBody {
        use ast::*;
        let node = node.walk(self);

        // Store a vector of select expressions, preserving select order. Consecutive non-substar
        // select expressions will be added to the same document. There's no need to rewrite
        // SelectValuesExpressions or queries of the form `select *`.
        let mut non_substar_exprs: Vec<ast::DocumentPair> = Vec::new();
        match node.clone() {
            SelectBody::Standard(exprs) => {
                let mut ordered_select_expressions = vec![];
                for expr in exprs {
                    match expr {
                        SelectExpression::Substar(substar) => {
                            // in order to preserve ordering, add a document expression with all the preceding
                            // non-substar expressions, then add the substar to the select vector
                            if !non_substar_exprs.is_empty() {
                                ordered_select_expressions.push(
                                    SelectValuesExpression::Expression(Expression::Document(
                                        non_substar_exprs,
                                    )),
                                );
                                non_substar_exprs = Vec::new();
                            };
                            ordered_select_expressions
                                .push((SelectValuesExpression::Substar(substar)).clone());
                        }
                        SelectExpression::Expression(expr) => {
                            match expr {
                                ast::OptionallyAliasedExpr::Unaliased(_) => {
                                    self.error = Some(Error::NoAliasForSelectExpression)
                                }
                                ast::OptionallyAliasedExpr::Aliased(AliasedExpr {
                                    expr,
                                    alias,
                                }) => {
                                    non_substar_exprs.push(ast::DocumentPair {
                                        key: alias,
                                        value: expr,
                                    });
                                }
                            };
                        }
                        SelectExpression::Star => return node,
                    }
                }
                // add any remaining non-substar expressions to a document and place that in the select list
                if !non_substar_exprs.is_empty() {
                    ordered_select_expressions.push(SelectValuesExpression::Expression(
                        Expression::Document(non_substar_exprs),
                    ));
                };
                SelectBody::Values(ordered_select_expressions)
            }
            SelectBody::Values(_) => {
                // Even though all SELECT queries should be rewritten to SELECT VALUE(S) for the
                // algebrizer, we should return an error if the top-level SELECT in a subquery
                // expression is a SELECT VALUE(S).
                if !self.allow_select_value {
                    self.error = Some(Error::SubqueryWithSelectValue)
                }
                node
            }
        }
    }
}
