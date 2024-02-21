use crate::ast::{
    visitors::*, BinaryExpr, BinaryOp, CollectionSource, ComparisonOp, Datasource, DocumentPair,
    Expression::*, JoinSource, JoinType, Literal::*, OptionallyAliasedExpr, Query, SelectBody,
    SelectClause, SelectExpression, SelectQuery, SetQuantifier, SubpathExpr, UnaryExpr, UnaryOp,
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
macro_rules! test_subpath_fields_ast {
    ($test_name:ident, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $test_name() {
            let actual = get_subpath_fields($input);
            dbg!(&$expected);
            dbg!(&actual);
            assert_eq!($expected, actual);
        }
    };
}

macro_rules! build_select_query {
    ($body:expr) => {
        Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: $body,
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            order_by_clause: None,
            having_clause: None,
            limit: None,
            offset: None,
        })
    };
}

mod subpath_field_tests {
    use super::*;

    test_subpath_fields_ast!(
        empty_query,
        expected = Vec::<Vec<&str>>::new(),
        input = build_select_query!(SelectBody::Standard(vec![])),
    );

    test_subpath_fields_ast!(
        basic_subpath,
        expected = vec![vec!["a", "b"]],
        input = build_select_query!(SelectBody::Standard(vec![SelectExpression::Expression(
            OptionallyAliasedExpr::Unaliased(Subpath(SubpathExpr {
                expr: Box::new(Identifier("a".to_string(),)),
                subpath: "b".to_string(),
            },),),
        ),])),
    );

    test_subpath_fields_ast!(
        nested_subpath,
        expected = vec![vec!["a", "b", "c"]],
        input = build_select_query!(SelectBody::Standard(vec![SelectExpression::Expression(
            OptionallyAliasedExpr::Unaliased(Subpath(SubpathExpr {
                expr: Box::new(Subpath(SubpathExpr {
                    expr: Box::new(Identifier("a".to_string(),)),
                    subpath: "b".to_string(),
                },)),
                subpath: "c".to_string(),
            },),),
        ),])),
    );

    test_subpath_fields_ast!(
        complex_nested_subpaths,
        expected = vec![vec!["a", "b"], vec!["c", "d"], vec!["e", "f"]],
        input = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Subpath(SubpathExpr {
                        expr: Box::new(Identifier("a".to_string(),)),
                        subpath: "b".to_string(),
                    },),),
                ),]),
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Cross,
                left: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "employees".to_string(),
                    alias: Some("e".to_string(),),
                },)),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "departments".to_string(),
                    alias: Some("d".to_string(),),
                },)),
                condition: Some(Binary(BinaryExpr {
                    left: Box::from(Subpath(SubpathExpr {
                        expr: Box::new(Identifier("c".to_string(),)),
                        subpath: "d".to_string(),
                    },)),
                    op: BinaryOp::Comparison(ComparisonOp::Eq,),
                    right: Box::from(Subpath(SubpathExpr {
                        expr: Box::new(Identifier("e".to_string(),)),
                        subpath: "f".to_string(),
                    },)),
                },),),
            },),),
            where_clause: None,
            group_by_clause: None,
            order_by_clause: None,
            having_clause: None,
            limit: None,
            offset: None
        }),
    );
}
