use crate::ast::{
    self,
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
};

/// Finds all positional sort keys and replaces them with the select expression they reference.
pub struct PositionalSortKeyRewritePass;

impl Pass for PositionalSortKeyRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = PositionalSortKeyRewriteVisitor::default();
        let rewritten = query.walk(&mut visitor);
        match visitor.error {
            Some(err) => Err(err),
            None => Ok(rewritten),
        }
    }
}

/// The visitor that performs the rewrites for the `PositionalSortKeyRewritePass`.
struct PositionalSortKeyRewriteVisitor {
    select_exprs: Option<Vec<ast::SelectExpression>>,
    error: Option<Error>,
}

impl PositionalSortKeyRewriteVisitor {
    /// If `key` is a positional sort key, then attempt to replace it with the appropriate reference.
    /// Otherwise, return `key` unmodified.
    /// If an error is encountered while replacing a positional sort key, return `key` unmodified and set `self.error`.
    fn replace_sort_key(&mut self, key: ast::SortKey) -> ast::SortKey {
        use ast::*;

        // Bind `position` if `key` is a position sort key.
        let position = if let SortKey::Positional(p) = key {
            p as usize
        } else {
            return key;
        };

        // Error if position is zero, because sort keys are 1-indexed.
        if position == 0 {
            self.error = Some(Error::PositionalSortKeyOutOfRange(0));
            return key;
        }

        // If `self.select_exprs` is `None`, then we saw a `SELECT VALUE`.
        if self.select_exprs.is_none() {
            self.error = Some(Error::PositionalSortKeyWithSelectValue);
            return key;
        }

        // Return the alias contained in the referenced select expr, erroring if the select expr or its alias don't exist.
        let alias = match self.select_exprs.as_ref().unwrap().get(position - 1) {
            None => Err(Error::PositionalSortKeyOutOfRange(position)),
            Some(SelectExpression::Aliased(AliasedExpr {
                expr: _,
                alias: Some(alias),
            })) => Ok(alias),
            Some(_) => Err(Error::NoAliasForSortKeyAtPosition(position)),
        };
        match alias {
            Ok(alias) => SortKey::Simple(Expression::Identifier(alias.clone())),
            Err(err) => {
                self.error = Some(err);
                key
            }
        }
    }
}

impl Default for PositionalSortKeyRewriteVisitor {
    fn default() -> Self {
        Self {
            select_exprs: None,
            error: None,
        }
    }
}

impl Visitor for PositionalSortKeyRewriteVisitor {
    fn visit_select_body(&mut self, node: ast::SelectBody) -> ast::SelectBody {
        use ast::*;

        // Walk the children first, rewriting if appropriate.
        let node = node.walk(self);

        // It is important that we set `self.select_exprs` even if it is `None` so that we don't carry over state from subqueries.
        self.select_exprs = match node {
            SelectBody::Standard(ref exprs) => Some(exprs.clone()),
            SelectBody::Values(_) => None,
        };

        node
    }

    fn visit_order_by_clause(&mut self, node: ast::OrderByClause) -> ast::OrderByClause {
        use ast::*;

        // Try to write each positional sort spec to a reference.
        // The correctness of this function depends on the `ORDER BY` clause being visited after `SELECT`.
        let sort_specs = node
            .sort_specs
            .into_iter()
            .map(|spec| SortSpec {
                key: self.replace_sort_key(spec.key),
                ..spec
            })
            .collect();
        OrderByClause { sort_specs }
    }
}
