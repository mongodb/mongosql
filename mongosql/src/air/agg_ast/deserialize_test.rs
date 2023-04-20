macro_rules! test_deserialize_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use super::TestStage;

            let input = $input;
            let s: TestStage = serde_yaml::from_str(&input).unwrap();

            assert_eq!($expected, s.stage)
        }
    };
}

macro_rules! test_deserialize_expr {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use super::TestExpr;

            let input = $input;
            let e: TestExpr = serde_yaml::from_str(&input).unwrap();

            assert_eq!($expected, e.expr)
        }
    };
}

mod stage_test {
    use crate::air::agg_ast::ast_definitions::Stage;
    use serde::Deserialize;

    #[derive(Debug, PartialEq, Deserialize)]
    struct TestStage {
        stage: Stage,
    }

    mod documents {
        use crate::{
            air::agg_ast::ast_definitions::{Expression, LiteralValue, Stage, StringOrRef},
            map,
        };

        test_deserialize_stage!(
            empty,
            expected = Stage::Documents(vec![]),
            input = r#"stage: {"$documents": []}"#
        );

        test_deserialize_stage!(
            singleton,
            expected = Stage::Documents(vec![
                map! {"a".to_string() => Expression::Literal(LiteralValue::Integer(1)) }
            ]),
            input = r#"stage: {"$documents": [{"a": 1}]}"#
        );

        test_deserialize_stage!(
            multiple_elements,
            expected = Stage::Documents(vec![
                map! {
                    "a".to_string() => Expression::Literal(LiteralValue::Integer(1)),
                    "b".to_string() => Expression::Literal(LiteralValue::Integer(2)),
                },
                map! {
                    "a".to_string() => Expression::StringOrRef(StringOrRef::String("yes".to_string())),
                    "b".to_string() => Expression::Literal(LiteralValue::Null),
                },
                map! {
                    "a".to_string() => Expression::Document(map! {
                        "b".to_string() => Expression::Document(map! {
                            "c".to_string() => Expression::Literal(LiteralValue::Boolean(true)),
                        }),
                    }),
                },
            ]),
            input = r#"stage: {"$documents": [
                                {"a": 1, "b": 2},
                                {"a": "yes", "b": null},
                                {"a": {"b": {"c": true}}}
            ]}"#
        );
    }

    mod project {
        use crate::{
            air::agg_ast::ast_definitions::{
                Expression, LiteralValue, Stage, StringOrRef, UntaggedOperator,
            },
            map,
        };

        test_deserialize_stage!(
            empty,
            expected = Stage::Project(map! {}),
            input = r#"stage: {"$project": {}}"#
        );

        test_deserialize_stage!(
            singleton,
            expected = Stage::Project(
                map! { "_id".to_string() => Expression::Literal(LiteralValue::Integer(0)) }
            ),
            input = r#"stage: {"$project": {"_id": 0}}"#
        );

        test_deserialize_stage!(
            multiple_elements,
            expected = Stage::Project(map! {
                "_id".to_string() => Expression::Literal(LiteralValue::Integer(0)),
                "foo".to_string() => Expression::StringOrRef(StringOrRef::Variable("ROOT".to_string())),
                "bar".to_string() => Expression::StringOrRef(StringOrRef::FieldRef("bar".to_string())),
                "a".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                    op: "$add".to_string(),
                    args: vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                    ]
                }),
            }),
            input = r#"stage: {"$project": {
                                "_id": 0,
                                "foo": "$$ROOT",
                                "bar": "$bar",
                                "a": {"$add": [1, 2]},
            }}"#
        );
    }

    mod replace_with {
        use crate::air::agg_ast::ast_definitions::{
            Expression, Stage, StringOrRef, UntaggedOperator,
        };

        test_deserialize_stage!(
            simple,
            expected = Stage::ReplaceWith(Expression::StringOrRef(StringOrRef::FieldRef(
                "a".to_string()
            ))),
            input = r#"stage: {"$replaceWith": "$a"}"#
        );

        test_deserialize_stage!(
            complex,
            expected = Stage::ReplaceWith(Expression::UntaggedOperator(UntaggedOperator {
                op: "$mergeObjects".to_string(),
                args: vec![
                    Expression::StringOrRef(StringOrRef::Variable("ROOT".to_string())),
                    Expression::StringOrRef(StringOrRef::FieldRef("as".to_string())),
                ]
            })),
            input = r#"stage: {"$replaceWith": {"$mergeObjects": ["$$ROOT", "$as"]}}"#
        );
    }

    mod match_stage {
        use crate::air::agg_ast::ast_definitions::MatchExpr;
        use crate::{
            air::agg_ast::ast_definitions::{
                Expression, LiteralValue, MatchExpression, Stage, StringOrRef, UntaggedOperator,
            },
            map,
        };

        test_deserialize_stage!(
            non_expr,
            expected = Stage::Match(MatchExpression::NonExpr(Expression::Document(map! {
                "a".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                    op: "$exists".to_string(),
                    args: vec![Expression::Literal(LiteralValue::Boolean(true))]
                })
            }))),
            input = r#"stage: {"$match": {"a": {"$exists": true}}}"#
        );

        test_deserialize_stage!(
            expr,
            expected = Stage::Match(MatchExpression::Expr(MatchExpr {
                expr: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                    op: "$sqlEq".to_string(),
                    args: vec![
                        Expression::StringOrRef(StringOrRef::FieldRef("a".to_string())),
                        Expression::StringOrRef(StringOrRef::FieldRef("b".to_string())),
                    ]
                }))
            })),
            input = r#"stage: {"$match": {"$expr": {"$sqlEq": ["$a", "$b"]}}}"#
        );
    }

    mod limit_skip {
        use crate::air::agg_ast::ast_definitions::Stage;

        test_deserialize_stage!(
            limit,
            expected = Stage::Limit(10),
            input = r#"stage: {"$limit": 10}"#
        );

        test_deserialize_stage!(
            skip,
            expected = Stage::Skip(100),
            input = r#"stage: {"$skip": 100}"#
        );
    }

    mod sort {
        use crate::{air::agg_ast::ast_definitions::Stage, map};

        test_deserialize_stage!(
            empty,
            expected = Stage::Sort(map! {}),
            input = r#"stage: {"$sort": {}}"#
        );

        test_deserialize_stage!(
            singleton,
            expected = Stage::Sort(map! { "a".to_string() => 1 }),
            input = r#"stage: {"$sort": {"a": 1}}"#
        );

        test_deserialize_stage!(
            multiple_elements,
            expected = Stage::Sort(map! { "a".to_string() => 1, "b".to_string() => -1 }),
            input = r#"stage: {"$sort": {"a": 1, "b": -1}}"#
        );
    }
}

mod expression_test {
    use crate::air::agg_ast::ast_definitions::Expression;
    use serde::Deserialize;

    #[derive(Debug, PartialEq, Deserialize)]
    struct TestExpr {
        expr: Expression,
    }

    mod literal {
        use crate::air::agg_ast::ast_definitions::{Expression, LiteralValue};

        test_deserialize_expr!(
            null,
            expected = Expression::Literal(LiteralValue::Null),
            input = r#"expr: null"#
        );

        test_deserialize_expr!(
            boolean_true,
            expected = Expression::Literal(LiteralValue::Boolean(true)),
            input = r#"expr: true"#
        );

        test_deserialize_expr!(
            boolean_false,
            expected = Expression::Literal(LiteralValue::Boolean(false)),
            input = r#"expr: false"#
        );

        test_deserialize_expr!(
            int,
            expected = Expression::Literal(LiteralValue::Integer(1)),
            input = r#"expr: 1"#
        );

        test_deserialize_expr!(
            long,
            expected = Expression::Literal(LiteralValue::Long(2147483648)),
            input = r#"expr: 2147483648"#
        );

        test_deserialize_expr!(
            double,
            expected = Expression::Literal(LiteralValue::Double(1.5)),
            input = r#"expr: 1.5"#
        );
    }

    mod string_or_ref {
        use crate::air::agg_ast::ast_definitions::{Expression, StringOrRef};

        test_deserialize_expr!(
            string,
            expected = Expression::StringOrRef(StringOrRef::String("yes".to_string())),
            input = r#"expr: "yes""#
        );

        test_deserialize_expr!(
            simple_field_ref,
            expected = Expression::StringOrRef(StringOrRef::FieldRef("a".to_string())),
            input = r#"expr: "$a""#
        );

        test_deserialize_expr!(
            nested_field_ref,
            expected = Expression::StringOrRef(StringOrRef::FieldRef("a.b.c".to_string())),
            input = r#"expr: "$a.b.c""#
        );

        test_deserialize_expr!(
            variable,
            expected = Expression::StringOrRef(StringOrRef::Variable("v".to_string())),
            input = r#"expr: "$$v""#
        );
    }

    mod array {
        use crate::air::agg_ast::ast_definitions::{Expression, LiteralValue, StringOrRef};

        test_deserialize_expr!(
            empty,
            expected = Expression::Array(vec![]),
            input = r#"expr: []"#
        );

        test_deserialize_expr!(
            singleton,
            expected = Expression::Array(vec![Expression::Literal(LiteralValue::Integer(1))]),
            input = r#"expr: [1]"#
        );

        test_deserialize_expr!(
            multiple_elements,
            expected = Expression::Array(vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::StringOrRef(StringOrRef::String("yes".to_string())),
                Expression::Array(vec![
                    Expression::Literal(LiteralValue::Boolean(true)),
                    Expression::Literal(LiteralValue::Double(4.1)),
                ]),
            ]),
            input = r#"expr: [1, "yes", [true, 4.1]]"#
        );
    }

    mod document {
        use crate::{
            air::agg_ast::ast_definitions::{Expression, LiteralValue, StringOrRef},
            map,
        };

        test_deserialize_expr!(
            empty,
            expected = Expression::Document(map! {}),
            input = r#"expr: {}"#
        );

        test_deserialize_expr!(
            singleton,
            expected = Expression::Document(
                map! {"a".to_string() => Expression::Literal(LiteralValue::Integer(1))}
            ),
            input = r#"expr: {"a": 1}"#
        );

        test_deserialize_expr!(
            multiple_elements,
            expected = Expression::Document(map! {
                "a".to_string() => Expression::Literal(LiteralValue::Integer(1)),
                "b".to_string() => Expression::StringOrRef(StringOrRef::String("two".to_string())),
                "c".to_string() => Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Boolean(false))
                }),
            }),
            input = r#"expr: {"a": 1, "b": "two", "c": {"x": false}}"#
        );

        test_deserialize_expr!(
            similar_to_op_but_no_dollarx,
            expected = Expression::Document(map! {
                "notOp".to_string() => Expression::Array(vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(3)),
                ])
            }),
            input = r#"expr: {"notOp": [1, 2, 3]}"#
        );
    }

    mod tagged_operators {
        use crate::{
            air::agg_ast::ast_definitions::{
                Convert, Expression, GetField, Let, Like, LiteralValue, SetField, SqlConvert,
                SqlDivide, Stage, StringOrRef, Subquery, SubqueryComparison, SubqueryExists,
                Switch, SwitchCase, TaggedOperator, UnsetField,
            },
            map,
        };

        test_deserialize_expr!(
            get_field,
            expected = Expression::TaggedOperator(TaggedOperator::GetField(GetField {
                field: "x".to_string(),
                input: Box::new(Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Integer(1))
                }))
            })),
            input = r#"expr: {"$getField": {"field": "x", "input": {"x": 1}}}"#
        );

        test_deserialize_expr!(
            set_field,
            expected = Expression::TaggedOperator(TaggedOperator::SetField(SetField {
                field: "x".to_string(),
                input: Box::new(Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Integer(1))
                })),
                value: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "new".to_string()
                )))
            })),
            input = r#"expr: {"$setField": {"field": "x", "input": {"x": 1}, "value": "new"}}"#
        );

        test_deserialize_expr!(
            unset_field,
            expected = Expression::TaggedOperator(TaggedOperator::UnsetField(UnsetField {
                field: "x".to_string(),
                input: Box::new(Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Integer(1))
                }))
            })),
            input = r#"expr: {"$unsetField": {"field": "x", "input": {"x": 1}}}"#
        );

        test_deserialize_expr!(
            switch,
            expected = Expression::TaggedOperator(TaggedOperator::Switch(Switch {
                branches: vec![
                    SwitchCase {
                        case: Box::new(Expression::StringOrRef(StringOrRef::FieldRef(
                            "a".to_string()
                        ))),
                        then: Box::new(Expression::Literal(LiteralValue::Integer(10))),
                    },
                    SwitchCase {
                        case: Box::new(Expression::StringOrRef(StringOrRef::FieldRef(
                            "b".to_string()
                        ))),
                        then: Box::new(Expression::Literal(LiteralValue::Integer(20))),
                    },
                ],
                default: Box::new(Expression::Literal(LiteralValue::Null))
            })),
            input = r#"expr: {"$switch": {
                                "branches": [
                                    {"case": "$a", "then": 10},
                                    {"case": "$b", "then": 20},
                                ],
                                "default": null
            }}"#
        );

        test_deserialize_expr!(
            let_expr,
            expected = Expression::TaggedOperator(TaggedOperator::Let(Let {
                vars: map! {
                    "a".to_string() => Expression::Literal(LiteralValue::Integer(1)),
                    "b".to_string() => Expression::Literal(LiteralValue::Integer(2)),
                },
                inside: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "body".to_string()
                )))
            })),
            input = r#"expr: {"$let": {
                                "vars": {"a": 1, "b": 2},
                                "in": "body"
            }}"#
        );

        test_deserialize_expr!(
            sql_convert,
            expected = Expression::TaggedOperator(TaggedOperator::SqlConvert(SqlConvert {
                input: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "1".to_string()
                ))),
                to: "int".to_string(),
                on_null: Box::new(Expression::Literal(LiteralValue::Null)),
                on_error: Box::new(Expression::Literal(LiteralValue::Null)),
            })),
            input = r#"expr: {"$sqlConvert": {
                                "input": "1",
                                "to": "int",
                                "onNull": null,
                                "onError": null
            }}"#
        );

        test_deserialize_expr!(
            convert,
            expected = Expression::TaggedOperator(TaggedOperator::Convert(Convert {
                input: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "1".to_string()
                ))),
                to: "int".to_string(),
                on_null: Box::new(Expression::Literal(LiteralValue::Null)),
                on_error: Box::new(Expression::Literal(LiteralValue::Null)),
            })),
            input = r#"expr: {"$convert": {
                                "input": "1",
                                "to": "int",
                                "onNull": null,
                                "onError": null
            }}"#
        );

        test_deserialize_expr!(
            like_with_escape,
            expected = Expression::TaggedOperator(TaggedOperator::Like(Like {
                input: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "x*yz".to_string()
                ))),
                pattern: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "x!*.*".to_string()
                ))),
                escape: Some("!".to_string())
            })),
            input = r#"expr: {"$like": {
                                "input": "x*yz",
                                "pattern": "x!*.*",
                                "escape": "!"
            }}"#
        );

        test_deserialize_expr!(
            like_without_escape,
            expected = Expression::TaggedOperator(TaggedOperator::Like(Like {
                input: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "x*yz".to_string()
                ))),
                pattern: Box::new(Expression::StringOrRef(StringOrRef::String(
                    "x!*.*".to_string()
                ))),
                escape: None
            })),
            input = r#"expr: {"$like": {
                                "input": "x*yz",
                                "pattern": "x!*.*"
            }}"#
        );

        test_deserialize_expr!(
            sql_divide,
            expected = Expression::TaggedOperator(TaggedOperator::SqlDivide(SqlDivide {
                dividend: Box::new(Expression::StringOrRef(StringOrRef::FieldRef(
                    "a".to_string()
                ))),
                divisor: Box::new(Expression::Literal(LiteralValue::Integer(2))),
                on_error: Box::new(Expression::Literal(LiteralValue::Null)),
            })),
            input = r#"expr: {"$sqlDivide": {
                                "dividend": "$a",
                                "divisor": 2,
                                "onError": null
            }}"#
        );

        test_deserialize_expr!(
            sql_subquery,
            expected = Expression::TaggedOperator(TaggedOperator::Subquery(Subquery {
                db: Some("foo".to_string()),
                collection: Some("bar".to_string()),
                let_bindings: None,
                output_path: Some(vec!["x".to_string()]),
                pipeline: vec![Stage::Project(
                    map! {"x".to_string() => Expression::Literal(LiteralValue::Integer(1))}
                )]
            })),
            input = r#"expr: {"$subquery": {
                            "db": "foo",
                            "collection": "bar",
                            "outputPath": ["x"],
                            "pipeline": [
                              {
                                "$project": {
                                  "x": 1
                                }
                              }
                            ]
                          }}"#
        );

        test_deserialize_expr!(
            sql_subquery_comparison,
            expected = Expression::TaggedOperator(TaggedOperator::SubqueryComparison(
                SubqueryComparison {
                    op: "eq".to_string(),
                    modifier: "all".to_string(),
                    arg: Box::new(Expression::Literal(LiteralValue::Integer(42))),
                    subquery: Subquery {
                        db: Some("foo".to_string()),
                        collection: Some("bar".to_string()),
                        let_bindings: None,
                        output_path: Some(vec!["x".to_string()]),
                        pipeline: vec![
                            Stage::Documents(vec![]),
                            Stage::Project(
                                map! {"x".to_string() => Expression::Literal(LiteralValue::Integer(1))}
                            )
                        ]
                    }.into()
                }
            )),
            input = r#"expr: {"$subqueryComparison": {
                            "op": "eq",
                            "modifier": "all",
                            "arg": 42,
                            "subquery": {
                                "db": "foo",
                                "collection": "bar",
                                "outputPath": ["x"],
                                "pipeline": [
                                    {"$documents": []},
                                    {
                                        "$project": {
                                            "x": 1
                                        }
                                    }
                                ]
                          }}}"#
        );

        test_deserialize_expr!(
            sql_subquery_exists,
            expected = Expression::TaggedOperator(TaggedOperator::SubqueryExists(SubqueryExists {
                db: Some("foo".to_string()),
                collection: Some("bar".to_string()),
                let_bindings: None,
                pipeline: vec![Stage::Project(
                    map! {"x".to_string() => Expression::Literal(LiteralValue::Integer(1))}
                )]
            })),
            input = r#"expr: {"$subqueryExists": {
                            "db": "foo",
                            "collection": "bar",
                            "pipeline": [
                              {
                                "$project": {
                                  "x": 1
                                }
                              }
                            ]
                          }}"#
        );
    }

    mod untagged_operators {
        use crate::air::agg_ast::ast_definitions::{Expression, StringOrRef, UntaggedOperator};

        test_deserialize_expr!(
            one_argument_non_array,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: "$sqlSqrt".to_string(),
                args: vec![Expression::StringOrRef(StringOrRef::FieldRef(
                    "x".to_string()
                ))]
            }),
            input = r#"expr: {"$sqlSqrt": "$x"}"#
        );

        test_deserialize_expr!(
            one_argument,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: "$sqlSqrt".to_string(),
                args: vec![Expression::StringOrRef(StringOrRef::FieldRef(
                    "x".to_string()
                ))]
            }),
            input = r#"expr: {"$sqlSqrt": ["$x"]}"#
        );

        test_deserialize_expr!(
            multiple_arguments,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: "$add".to_string(),
                args: vec![
                    Expression::StringOrRef(StringOrRef::FieldRef("x".to_string())),
                    Expression::StringOrRef(StringOrRef::FieldRef("y".to_string())),
                    Expression::StringOrRef(StringOrRef::FieldRef("z".to_string())),
                ]
            }),
            input = r#"expr: {"$add": ["$x", "$y", "$z"]}"#
        );
    }
}
