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

        // Store all substar expressions in a vector and all non-substar expressions in a single map,
        // where the alias is the key and the expression is the value. There's no need to rewrite
        // SelectValuesExpressions or queries of the form `select *`.
        let mut v: Vec<ast::DocumentPair> = Vec::new();
        let mut substar_exprs: Vec<SelectValuesExpression> = vec![];
        let mut rewrite = true;
        match node.clone() {
            SelectBody::Standard(exprs) => {
                for expr in exprs {
                    match expr {
                        SelectExpression::Substar(substar) => {
                            substar_exprs.push((SelectValuesExpression::Substar(substar)).clone())
                        }
                        SelectExpression::Aliased(aliased) => {
                            match aliased.alias {
                                None => self.error = Some(Error::NoAliasForSelectExpression),
                                Some(alias) => {
                                    v.push(ast::DocumentPair {
                                        key: alias,
                                        value: aliased.expr,
                                    });
                                }
                            };
                        }
                        SelectExpression::Star => rewrite = false,
                    }
                }
            }
            SelectBody::Values(_) => {
                // Even though all SELECT queries should be rewritten to SELECT VALUE(S) for the
                // algebrizer, we should return an error if the top-level SELECT in a subquery
                // expression is a SELECT VALUE(S).
                if !self.allow_select_value {
                    self.error = Some(Error::SubqueryWithSelectValue)
                }
                rewrite = false
            }
        };

        if !rewrite {
            return node;
        };

        // Return a vector of SelectValuesExpressions that includes the map of non-substar expressions
        // followed by the substar expressions.
        let mut array = vec![];
        if !v.is_empty() {
            array.push(SelectValuesExpression::Expression(Expression::Document(v)));
        };
        array.append(&mut substar_exprs);
        SelectBody::Values(array)
    }
}
