use crate::ast::{
    rewrites::{Pass, Result},
    visitor::Visitor,
    *,
};

/// Finds all expressions of the form `<expr> [NOT] IN (<e1>, <e2>, ...)` and rewrites them to either
/// `<expr> (<> | =) <e1> [(AND | OR) <expr> (<> | =) <e2>, ...]` or
/// `<expr> [NOT] IN (SELECT _1 FROM [{'_1': <e1>}, {'_1': <e2>}, ...] AS _arr)`
/// depending on the type of expression `<expr>` is. For simple field references, we rewrite to the former;
/// for all other expression types, we rewrite to the latter.
pub struct InTupleRewritePass;

impl Pass for InTupleRewritePass {
    fn apply(&self, query: Query) -> Result<Query> {
        let mut visitor = InTupleRewriteVisitor::default();
        Ok(query.walk(&mut visitor))
    }
}

/// The visitor that performs the rewrites for the `InTupleRewritePass`.
#[derive(Default)]
struct InTupleRewriteVisitor;

impl InTupleRewriteVisitor {
    fn create_binary_expr_comparison(
        left: Expression,
        right: Expression,
        tree_op: BinaryOp,
    ) -> Expression {
        let comparison_op = if tree_op == BinaryOp::Or {
            BinaryOp::Comparison(ComparisonOp::Eq)
        } else {
            BinaryOp::Comparison(ComparisonOp::Neq)
        };

        Expression::Binary(BinaryExpr {
            left: Box::new(left),
            op: comparison_op,
            right: Box::new(right),
        })
    }

    fn rewrite_to_comparisons(
        lhs: Expression,
        op: BinaryOp,
        tuple_elems: Vec<Expression>,
    ) -> Expression {
        let tree_op = if op == BinaryOp::In {
            BinaryOp::Or
        } else {
            BinaryOp::And
        };

        tuple_elems
            .into_iter()
            .map(|rhs| Self::create_binary_expr_comparison(lhs.clone(), rhs, tree_op)) // this creates a Vec<Expression> where the expressions are the comparisons `lhs (<>|=) rhs`
            .reduce(|acc, comp| {
                Expression::Binary(BinaryExpr {
                    left: Box::new(acc),
                    op: tree_op,
                    right: Box::new(comp),
                })
            }) // this will reduce the Vec<Expression> into an Expression::Binary that does the ORs or ANDs.
            .unwrap()
    }

    fn is_simple_field_ref_expr(lhs: Expression) -> bool {
        match lhs {
            Expression::Identifier(_) => true,
            Expression::Subpath(sp) => Self::is_simple_field_ref_expr(*sp.expr),
            Expression::Access(ae) => {
                Self::is_simple_field_ref_expr(*ae.expr)
                    && matches!(*ae.subfield, Expression::Literal(Literal::String(_)))
            }
            _ => false,
        }
    }

    fn rewrite_to_subquery(
        lhs: Expression,
        op: BinaryOp,
        tuple_elems: Vec<Expression>,
    ) -> Expression {
        // Build the AST for a SELECT clause representing `SELECT _1`.
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
                Expression::Document(vec![DocumentPair {
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

        Expression::Binary(BinaryExpr {
            left: Box::new(lhs),
            op,
            right: Box::new(Expression::Subquery(Box::new(subquery))),
        })
    }
}

impl Visitor for InTupleRewriteVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
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

        // If the left side of `expr` is a simple field ref, then rewrite to the appropriate
        // boolean/comparison operations. Otherwise, rewrite to a subquery.
        if Self::is_simple_field_ref_expr(*expr.left.clone()) {
            Self::rewrite_to_comparisons(*expr.left, expr.op, tuple_elems)
        } else {
            Self::rewrite_to_subquery(*expr.left, expr.op, tuple_elems)
        }
    }
}

/// Finds all one-element tuples like (x) or (1+2) and replaces them with the underlying expression.
/// Intended to be used after `InTupleRewritePass` to avoid rewriting one-element `IN` tuples.
pub struct SingleTupleRewritePass;

impl Pass for SingleTupleRewritePass {
    fn apply(&self, query: Query) -> Result<Query> {
        Ok(query.walk(&mut SingleTupleRewriteVisitor))
    }
}

/// The visitor that performs the rewrites for `SingleTupleRewritePass`.
/// Also used in `pretty_print_test`.
pub struct SingleTupleRewriteVisitor;

impl Visitor for SingleTupleRewriteVisitor {
    fn visit_expression(&mut self, e: Expression) -> Expression {
        match e {
            Expression::Tuple(mut t) if t.len() == 1 => self.visit_expression(t.pop().unwrap()),
            _ => e.walk(self),
        }
    }
}

#[cfg(test)]
mod is_simple_field_ref_expr_test {
    use crate::ast::rewrites::tuples::InTupleRewriteVisitor;
    use crate::ast::*;

    macro_rules! test_is_simple_field_ref_expr {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let expected = $expected;
                let actual = InTupleRewriteVisitor::is_simple_field_ref_expr($input);
                assert_eq!(expected, actual);
            }
        };
    }

    test_is_simple_field_ref_expr!(
        simple_field_ref,
        expected = true,
        input = Expression::Identifier("foo".to_string())
    );

    test_is_simple_field_ref_expr!(
        not_simple_field_ref,
        expected = false,
        input = Expression::Literal(Literal::String("foo".to_string()))
    );

    test_is_simple_field_ref_expr!(
        simple_field_ref_recursive_check,
        expected = true,
        input = Expression::Access(AccessExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Identifier("foo2".to_string())),
                subpath: "bar".to_string(),
            })),
            subfield: Box::new(Expression::Literal(Literal::String("foo1".to_string()))),
        })
    );

    test_is_simple_field_ref_expr!(
        not_simple_field_ref_recursive_check,
        expected = false,
        input = Expression::Access(AccessExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Literal(Literal::String("foo1".to_string()))),
                subpath: "bar".to_string(),
            })),
            subfield: Box::new(Expression::Literal(Literal::String("foo1".to_string()))),
        })
    );

    test_is_simple_field_ref_expr!(
        access_subfield_is_not_string,
        expected = false,
        input = Expression::Access(AccessExpr {
            expr: Box::new(Expression::Identifier("foo".to_string())),
            subfield: Box::new(Expression::Literal(Literal::Integer(32))),
        })
    );
}
