use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};
use linked_hash_map::LinkedHashMap;

/// Finds all expressions of the form `<expr> IN (<e1>, <e2>, ...)` and rewrites them to `<expr> = ANY (SELECT _1 FROM [{'_1': <e1>}, {'_1': <e2>}, ...] AS _arr)`.
pub struct InTupleRewritePass;

impl Pass for InTupleRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = InTupleRewriteVisitor;
        Ok(query.walk(&mut visitor))
    }
}

/// The visitor that performs the rewrites for the `InTupleRewritePass`.
struct InTupleRewriteVisitor;

impl Visitor for InTupleRewriteVisitor {
    fn visit_expression(&mut self, node: ast::Expression) -> ast::Expression {
        use ast::*;

        // Visit children, rewriting if appropriate.
        let node = node.walk(self);

        // If `node` is an IN expression, then bind it to `in_expr`, else return the original `node` unmodified.
        let in_expr = if let Expression::Binary(
            ref
            in_expr
            @
            BinaryExpr {
                left: _,
                right: _,
                op: BinaryOp::In,
            },
        ) = node
        {
            in_expr.clone()
        } else {
            return node;
        };

        // If the right side of `in_expr` is a tuple, then bind its elements to `tuple_elems`, else return the original `node` unmodified.
        let tuple_elems = if let Expression::Tuple(elems) = *in_expr.right {
            elems
        } else {
            return node;
        };

        // Build the AST for a FROM clause representing `SELECT _1`.
        let select_clause = SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpr {
                expr: Expression::Identifier("_1".to_string()),
                alias: None,
            })]),
        };

        // Build a Vec of documents `{'_1': <expr>}` for each `<expr>` in `tuple_elems`.
        let array = tuple_elems
            .into_iter()
            .map(|expr| {
                let mut map = LinkedHashMap::new();
                map.insert("_1".to_string(), expr);
                Expression::Document(map)
            })
            .collect();

        // Build the AST for a FROM clause representing `FROM <array> AS _arr`.
        let from_clause = Some(Datasource::Array(ArraySource {
            array,
            alias: "_arr".to_string(),
        }));

        // Assemble `select_clause` and `from_clause` into a query AST.
        let subquery = Query::Select(SelectQuery {
            select_clause,
            from_clause,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        });

        // Build the AST for a subquery comparison representing `<left> = ANY <subquery>`.
        // Return it as a replacement for the original `IN` expression.
        Expression::SubqueryComparison(SubqueryComparisonExpr {
            expr: in_expr.left,
            op: BinaryOp::Eq,
            quantifier: SubqueryQuantifier::Any,
            subquery: Box::new(subquery),
        })
    }
}

/// Finds all one-element tuples like (x) or (1+2) and replaces them with the underlying expression.
/// Intended to be used after `InTupleRewritePass` to avoid rewriting one-element `IN` tuples.
pub struct SingleTupleRewritePass;

impl Pass for SingleTupleRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        Ok(query.walk(&mut SingleTupleRewriteVisitor))
    }
}

/// The visitor that performs the rewrites for `SingleTupleRewritePass`.
/// Also used in `pretty_print_test`.
pub struct SingleTupleRewriteVisitor;

impl Visitor for SingleTupleRewriteVisitor {
    fn visit_expression(&mut self, e: ast::Expression) -> ast::Expression {
        use ast::*;
        match e {
            Expression::Tuple(mut t) if t.len() == 1 => self.visit_expression(t.pop().unwrap()),
            _ => e.walk(self),
        }
    }
}
