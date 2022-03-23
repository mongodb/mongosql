use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};

/// Finds all expressions of the form `<expr> IN (<e1>, <e2>, ...)` and rewrites them to `<expr> = ANY (SELECT _1 FROM [{'_1': <e1>}, {'_1': <e2>}, ...] AS _arr)`.
pub struct InTupleRewritePass;

impl Pass for InTupleRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = InTupleRewriteVisitor::default();
        Ok(query.walk(&mut visitor))
    }
}

/// The visitor that performs the rewrites for the `InTupleRewritePass`.
#[derive(Default)]
struct InTupleRewriteVisitor;

impl Visitor for InTupleRewriteVisitor {
    fn visit_expression(&mut self, node: ast::Expression) -> ast::Expression {
        use ast::*;

        // Visit children, rewriting if appropriate.
        let node = node.walk(self);

        // If `node` is an IN or NOT IN expression, then bind it to `expr`, else return the
        // original `node` unmodified.
        let expr = match node {
            Expression::Binary(ref expr)
                if expr.op == BinaryOp::In || expr.op == BinaryOp::NotIn =>
            {
                expr.clone()
            }
            _ => return node,
        };

        // If the right side of `expr` is a tuple, then bind its elements to `tuple_elems`,
        // else return the original `node` unmodified.
        let tuple_elems = if let Expression::Tuple(elems) = *expr.right {
            elems
        } else {
            return node;
        };

        // Build the AST for a FROM clause representing `SELECT _1`.
        let select_clause = SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Expression(
                OptionallyAliasedExpr::Unaliased(Expression::Identifier("_1".to_string())),
            )]),
        };

        // Build a Vec of documents `{'_1': <expr>}` for each `<expr>` in `tuple_elems`.
        let array = tuple_elems
            .into_iter()
            .map(|expr| {
                Expression::Document(vec![ast::DocumentPair {
                    key: "_1".to_string(),
                    value: expr,
                }])
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

        // Build the AST for a subquery comparison representing `<left> = ANY <subquery>`
        // or `<left> <> ALL <subquery>`. Return it as a replacement for the original
        // `IN` or `NOT IN` expression.
        let (op, quantifier) = match expr.op {
            BinaryOp::In => (ComparisonOp::Eq, SubqueryQuantifier::Any),
            _ => (ComparisonOp::Neq, SubqueryQuantifier::All),
        };
        Expression::SubqueryComparison(SubqueryComparisonExpr {
            expr: expr.left,
            op,
            quantifier,
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
