use crate::ast::{
    visitors::*, BinaryExpr, BinaryOp, DocumentPair, Expression::*, Literal::*, UnaryExpr, UnaryOp,
};

macro_rules! test_visitors {
    ($test_name:ident, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $test_name() {
            let input = $input;
            let expected = $expected;

            let (_, actual) = are_literal(input);

            assert_eq!(expected, actual);
        }
    };
}

mod are_literal_tests {
    use super::*;

    test_visitors!(
        nexted_expr_without_identifiers_is_literal,
        expected = true,
        input = vec![Array(vec![Document(vec![DocumentPair {
            key: "a".into(),
            value: Unary(UnaryExpr {
                op: UnaryOp::Neg,
                expr: Box::new(Literal(Integer(1)))
            }),
        }])])],
    );

    test_visitors!(
        nexted_expr_with_identifiers_is_non_literal,
        expected = false,
        input = vec![Array(vec![Document(vec![DocumentPair {
            key: "a".into(),
            value: Identifier("b".into()),
        }])])],
    );

    test_visitors!(
        multiple_expressions_without_identifiers_is_literal,
        expected = true,
        input = vec![
            Binary(BinaryExpr {
                left: Box::new(Literal(Integer(4))),
                op: BinaryOp::Add,
                right: Box::new(Binary(BinaryExpr {
                    left: Box::new(Literal(Integer(5))),
                    op: BinaryOp::Add,
                    right: Box::new(Literal(Integer(6))),
                })),
            }),
            Binary(BinaryExpr {
                left: Box::new(Literal(Integer(4))),
                op: BinaryOp::Add,
                right: Box::new(Binary(BinaryExpr {
                    left: Box::new(Literal(Integer(5))),
                    op: BinaryOp::Mul,
                    right: Box::new(Unary(UnaryExpr {
                        op: UnaryOp::Neg,
                        expr: Box::new(Literal(Integer(1)))
                    })),
                })),
            }),
        ],
    );

    test_visitors!(
        multiple_expressions_with_identifiers_is_non_literal,
        expected = false,
        input = vec![
            Binary(BinaryExpr {
                left: Box::new(Literal(Integer(4))),
                op: BinaryOp::Add,
                right: Box::new(Binary(BinaryExpr {
                    left: Box::new(Literal(Integer(5))),
                    op: BinaryOp::Add,
                    right: Box::new(Identifier("1".into())),
                })),
            }),
            Binary(BinaryExpr {
                left: Box::new(Literal(Integer(4))),
                op: BinaryOp::Add,
                right: Box::new(Binary(BinaryExpr {
                    left: Box::new(Literal(Integer(5))),
                    op: BinaryOp::Mul,
                    right: Box::new(Unary(UnaryExpr {
                        op: UnaryOp::Neg,
                        expr: Box::new(Literal(Integer(1)))
                    })),
                })),
            }),
        ],
    );

    test_visitors!(empty_vector_is_literal, expected = true, input = vec![],);

    test_visitors!(
        top_level_identifier_is_non_literal,
        expected = false,
        input = vec![Identifier("a".into())],
    );
}
