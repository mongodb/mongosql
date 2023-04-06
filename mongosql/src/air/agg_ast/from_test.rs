use crate::air;

fn default_source() -> air::Stage {
    air::Stage::Collection(air::Collection {
        db: "test".to_string(),
        collection: "default".to_string(),
    })
}

macro_rules! test_from_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let source = default_source();

            let actual = air::Stage::from((source, input));

            assert_eq!(expected, actual)
        }
    };
}

macro_rules! test_from_expr {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let actual = air::Expression::from(input);

            assert_eq!(expected, actual)
        }
    };
}

mod stage {
    mod collection {
        use crate::air::{
            self,
            agg_ast::{ast_definitions as agg_ast, from_test::default_source},
        };

        test_from_stage!(
            simple,
            expected = air::Stage::Collection(air::Collection {
                db: "test".to_string(),
                collection: "c".to_string()
            }),
            input = agg_ast::Stage::Collection(agg_ast::Collection {
                db: "test".to_string(),
                collection: "c".to_string(),
            })
        );
    }

    mod documents {
        use crate::{
            air::{
                self,
                agg_ast::{ast_definitions as agg_ast, from_test::default_source},
            },
            map, unchecked_unique_linked_hash_map,
        };

        test_from_stage!(
            empty,
            expected = air::Stage::Documents(air::Documents { array: vec![] }),
            input = agg_ast::Stage::Documents(vec![])
        );

        test_from_stage!(
            singleton,
            expected = air::Stage::Documents(air::Documents {
                array: vec![air::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                    }
                )],
            }),
            input = agg_ast::Stage::Documents(vec![map! {
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))
            }])
        );

        test_from_stage!(
            multiple_elements,
            expected = air::Stage::Documents(air::Documents {
                array: vec![
                    air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                    }),
                    air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2))
                    }),
                    air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(3))
                    }),
                ],
            }),
            input = agg_ast::Stage::Documents(vec![
                map! {
                    "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))
                },
                map! {
                    "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(2))
                },
                map! {
                    "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(3))
                },
            ])
        );
    }

    mod project {
        use crate::{
            air::{
                self,
                agg_ast::{ast_definitions as agg_ast, from_test::default_source},
            },
            map, unchecked_unique_linked_hash_map,
        };

        test_from_stage!(
            empty,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {},
            }),
            input = agg_ast::Stage::Project(map! {})
        );

        test_from_stage!(
            singleton,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                },
            }),
            input = agg_ast::Stage::Project(map! {
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1)),
            })
        );

        test_from_stage!(
            multiple_elements,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                    "b".to_string() => air::Expression::Literal(air::LiteralValue::Long(2)),
                    "c".to_string() => air::Expression::Literal(air::LiteralValue::Boolean(true)),
                },
            }),
            input = agg_ast::Stage::Project(map! {
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1)),
                "b".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Long(2)),
                "c".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Boolean(true)),
            })
        );
    }

    mod replace_with {
        use crate::air::{
            self,
            agg_ast::{ast_definitions as agg_ast, from_test::default_source},
        };

        test_from_stage!(
            simple,
            expected = air::Stage::ReplaceWith(air::ReplaceWith {
                source: Box::new(default_source()),
                new_root: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Stage::ReplaceWith(agg_ast::Expression::Literal(
                agg_ast::LiteralValue::Null
            ))
        );
    }

    mod match_stage {
        use crate::air::{
            self,
            agg_ast::{ast_definitions as agg_ast, from_test::default_source},
        };

        test_from_stage!(
            expr,
            expected = air::Stage::Match(air::Match {
                source: Box::new(default_source()),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Boolean(true))),
            }),
            input = agg_ast::Stage::Match(agg_ast::MatchExpression::Expr(agg_ast::MatchExpr {
                expr: Box::new(agg_ast::Expression::Literal(
                    agg_ast::LiteralValue::Boolean(true)
                )),
            }))
        );

        test_from_stage!(
            non_expr,
            expected = air::Stage::Match(air::Match {
                source: Box::new(default_source()),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Boolean(false))),
            }),
            input = agg_ast::Stage::Match(agg_ast::MatchExpression::NonExpr(
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Boolean(false))
            ))
        );
    }

    mod limit_skip {
        use crate::air::{
            self,
            agg_ast::{ast_definitions as agg_ast, from_test::default_source},
        };

        test_from_stage!(
            limit,
            expected = air::Stage::Limit(air::Limit {
                source: Box::new(default_source()),
                limit: 10,
            }),
            input = agg_ast::Stage::Limit(10)
        );

        test_from_stage!(
            skip,
            expected = air::Stage::Skip(air::Skip {
                source: Box::new(default_source()),
                skip: 20,
            }),
            input = agg_ast::Stage::Skip(20)
        );
    }

    mod sort {
        use crate::{
            air::{
                self,
                agg_ast::{ast_definitions as agg_ast, from_test::default_source},
            },
            map,
        };

        test_from_stage!(
            empty,
            expected = air::Stage::Sort(air::Sort {
                source: Box::new(default_source()),
                specs: vec![],
            }),
            input = agg_ast::Stage::Sort(map! {})
        );

        test_from_stage!(
            singleton,
            expected = air::Stage::Sort(air::Sort {
                source: Box::new(default_source()),
                specs: vec![air::SortSpecification::Asc("a".to_string())],
            }),
            input = agg_ast::Stage::Sort(map! {
                "a".to_string() => 1
            })
        );

        test_from_stage!(
            multiple_elements,
            expected = air::Stage::Sort(air::Sort {
                source: Box::new(default_source()),
                specs: vec![
                    air::SortSpecification::Asc("a".to_string()),
                    air::SortSpecification::Desc("b".to_string()),
                    air::SortSpecification::Asc("c".to_string()),
                ],
            }),
            input = agg_ast::Stage::Sort(map! {
                "c".to_string() => 1,
                "a".to_string() => 1,
                "b".to_string() => -1,
            })
        );
    }
}

mod expression {
    mod literal {
        use crate::air::{self, agg_ast::ast_definitions as agg_ast};

        test_from_expr!(
            null,
            expected = air::Expression::Literal(air::LiteralValue::Null),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)
        );

        test_from_expr!(
            boolean,
            expected = air::Expression::Literal(air::LiteralValue::Boolean(true)),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Boolean(true))
        );

        test_from_expr!(
            int,
            expected = air::Expression::Literal(air::LiteralValue::Integer(10)),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(10))
        );

        test_from_expr!(
            long,
            expected = air::Expression::Literal(air::LiteralValue::Long(200)),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Long(200))
        );

        test_from_expr!(
            double,
            expected = air::Expression::Literal(air::LiteralValue::Double(14.3)),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Double(14.3))
        );
    }

    mod string_or_ref {
        use crate::air::{self, agg_ast::ast_definitions as agg_ast};

        test_from_expr!(
            string,
            expected = air::Expression::Literal(air::LiteralValue::String("s".to_string())),
            input = agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::String("s".to_string()))
        );

        test_from_expr!(
            empty_field_ref,
            expected = air::Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "".to_string(),
            }),
            input =
                agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("".to_string()))
        );

        test_from_expr!(
            simple_field_ref,
            expected = air::Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "a".to_string(),
            }),
            input =
                agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("a".to_string()))
        );

        test_from_expr!(
            nested_field_ref,
            expected = air::Expression::FieldRef(air::FieldRef {
                parent: Some(Box::new(air::FieldRef {
                    parent: Some(Box::new(air::FieldRef {
                        parent: None,
                        name: "a".to_string(),
                    })),
                    name: "b".to_string(),
                })),
                name: "c".to_string(),
            }),
            input = agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                "a.b.c".to_string()
            ))
        );

        test_from_expr!(
            simple_variable,
            expected = air::Expression::Variable(air::Variable {
                parent: None,
                name: "v".to_string()
            }),
            input =
                agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::Variable("v".to_string()))
        );

        test_from_expr!(
            nested_variable,
            expected = air::Expression::Variable(air::Variable {
                parent: Some(Box::new(air::Variable {
                    parent: Some(Box::new(air::Variable {
                        parent: None,
                        name: "x".to_string()
                    })),
                    name: "y".to_string()
                })),
                name: "z".to_string()
            }),
            input = agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::Variable(
                "x.y.z".to_string()
            ))
        );
    }

    mod array {
        use crate::air::{self, agg_ast::ast_definitions as agg_ast};

        test_from_expr!(
            empty,
            expected = air::Expression::Array(vec![]),
            input = agg_ast::Expression::Array(vec![])
        );

        test_from_expr!(
            singleton,
            expected =
                air::Expression::Array(vec![air::Expression::Literal(air::LiteralValue::Null)]),
            input = agg_ast::Expression::Array(vec![agg_ast::Expression::Literal(
                agg_ast::LiteralValue::Null
            )])
        );

        test_from_expr!(
            multiple_elements,
            expected = air::Expression::Array(vec![
                air::Expression::Literal(air::LiteralValue::Integer(1)),
                air::Expression::Literal(air::LiteralValue::Integer(2)),
                air::Expression::Literal(air::LiteralValue::Integer(3)),
            ]),
            input = agg_ast::Expression::Array(vec![
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1)),
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(2)),
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(3)),
            ])
        );
    }

    mod document {
        use crate::{
            air::{self, agg_ast::ast_definitions as agg_ast},
            map, unchecked_unique_linked_hash_map,
        };

        test_from_expr!(
            empty,
            expected = air::Expression::Document(unchecked_unique_linked_hash_map! {}),
            input = agg_ast::Expression::Document(map! {})
        );

        test_from_expr!(
            singleton,
            expected = air::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => air::Expression::Literal(air::LiteralValue::Null)
            }),
            input = agg_ast::Expression::Document(map! {
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)
            })
        );

        test_from_expr!(
            multiple_elements,
            expected = air::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                "b".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2)),
                "c".to_string() => air::Expression::Literal(air::LiteralValue::Integer(3)),
            }),
            input = agg_ast::Expression::Document(map! {
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1)),
                "b".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(2)),
                "c".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(3)),
            })
        );
    }

    mod tagged_operators {
        use crate::{
            air::{self, agg_ast::ast_definitions as agg_ast},
            map,
        };

        test_from_expr!(
            get_field,
            expected = air::Expression::GetField(air::GetField {
                field: "a".to_string(),
                input: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "d".to_string()
                }))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::GetField(
                agg_ast::GetField {
                    field: "a".to_string(),
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("d".to_string())
                    ))
                }
            ))
        );

        test_from_expr!(
            set_field,
            expected = air::Expression::SetField(air::SetField {
                field: "a".to_string(),
                input: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "d".to_string(),
                })),
                value: Box::new(air::Expression::Literal(air::LiteralValue::Null))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SetField(
                agg_ast::SetField {
                    field: "a".to_string(),
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("d".to_string())
                    )),
                    value: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null))
                }
            ))
        );

        test_from_expr!(
            unset_field,
            expected = air::Expression::UnsetField(air::UnsetField {
                field: "a".to_string(),
                input: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "d".to_string()
                }))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::UnsetField(
                agg_ast::UnsetField {
                    field: "a".to_string(),
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("d".to_string())
                    ))
                }
            ))
        );

        test_from_expr!(
            switch,
            expected = air::Expression::Switch(air::Switch {
                branches: vec![
                    air::SwitchCase {
                        case: Box::new(air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "a".to_string(),
                        })),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    },
                    air::SwitchCase {
                        case: Box::new(air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "b".to_string(),
                        })),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(2))),
                    },
                ],
                default: Box::new(air::Expression::Literal(air::LiteralValue::Integer(3)))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Switch(
                agg_ast::Switch {
                    branches: vec![
                        agg_ast::SwitchCase {
                            case: Box::new(agg_ast::Expression::StringOrRef(
                                agg_ast::StringOrRef::FieldRef("a".to_string())
                            )),
                            then: Box::new(agg_ast::Expression::Literal(
                                agg_ast::LiteralValue::Integer(1)
                            )),
                        },
                        agg_ast::SwitchCase {
                            case: Box::new(agg_ast::Expression::StringOrRef(
                                agg_ast::StringOrRef::FieldRef("b".to_string())
                            )),
                            then: Box::new(agg_ast::Expression::Literal(
                                agg_ast::LiteralValue::Integer(2)
                            )),
                        },
                    ],
                    default: Box::new(agg_ast::Expression::Literal(
                        agg_ast::LiteralValue::Integer(3)
                    ))
                }
            ))
        );

        test_from_expr!(
            let_expr,
            expected = air::Expression::Let(air::Let {
                vars: vec![air::LetVariable {
                    name: "v".to_string(),
                    expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    })),
                },],
                inside: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Let(
                agg_ast::Let {
                    vars: map! {
                        "v".to_string() => agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("a".to_string())),
                    },
                    inside: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );

        test_from_expr!(
            convert,
            expected = air::Expression::Convert(air::Convert {
                input: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string(),
                })),
                to: air::Type::Int32,
                on_null: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Convert(
                agg_ast::Convert {
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("a".to_string())
                    )),
                    to: "int".to_string(),
                    on_null: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                    on_error: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );

        test_from_expr!(
            sql_convert,
            expected = air::Expression::SqlConvert(air::SqlConvert {
                input: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string(),
                })),
                to: air::SqlConvertTargetType::Array,
                on_null: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SqlConvert(
                agg_ast::SqlConvert {
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("a".to_string())
                    )),
                    to: "array".to_string(),
                    on_null: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                    on_error: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );

        test_from_expr!(
            like_with_escape,
            expected = air::Expression::Like(air::Like {
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "s".to_string(),
                })),
                pattern: Box::new(air::Expression::Literal(air::LiteralValue::String(
                    "pat".to_string()
                ))),
                escape: Some("e".to_string())
            }),
            input =
                agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Like(agg_ast::Like {
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("s".to_string())
                    )),
                    pattern: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::String("pat".to_string())
                    )),
                    escape: Some("e".to_string()),
                }))
        );

        test_from_expr!(
            like_without_escape,
            expected = air::Expression::Like(air::Like {
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "s".to_string(),
                })),
                pattern: Box::new(air::Expression::Literal(air::LiteralValue::String(
                    "pat".to_string()
                ))),
                escape: None
            }),
            input =
                agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Like(agg_ast::Like {
                    input: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("s".to_string())
                    )),
                    pattern: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::String("pat".to_string())
                    )),
                    escape: None,
                }))
        );

        test_from_expr!(
            sql_divide,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Divide,
                args: vec![
                    air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    }),
                    air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "b".to_string()
                    }),
                ]
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SqlDivide(
                agg_ast::SqlDivide {
                    dividend: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("a".to_string())
                    )),
                    divisor: Box::new(agg_ast::Expression::StringOrRef(
                        agg_ast::StringOrRef::FieldRef("b".to_string())
                    )),
                    on_error: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );
    }

    mod untagged_operators {
        use crate::air::{self, agg_ast::ast_definitions as agg_ast};

        test_from_expr!(
            sql_op_one_arg,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Pos,
                args: vec![air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string()
                })]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$sqlPos".to_string(),
                args: vec![agg_ast::Expression::StringOrRef(
                    agg_ast::StringOrRef::FieldRef("a".to_string())
                )],
            })
        );

        test_from_expr!(
            sql_op_multiple_args,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Eq,
                args: vec![
                    air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    }),
                    air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "b".to_string()
                    }),
                ]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$sqlEq".to_string(),
                args: vec![
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                        "a".to_string()
                    )),
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                        "b".to_string()
                    )),
                ],
            })
        );

        test_from_expr!(
            mql_op_one_arg,
            expected = air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                op: air::MQLOperator::Size,
                args: vec![air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string()
                })]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$size".to_string(),
                args: vec![agg_ast::Expression::StringOrRef(
                    agg_ast::StringOrRef::FieldRef("a".to_string())
                )],
            })
        );

        test_from_expr!(
            mql_op_multiple_args,
            expected = air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                op: air::MQLOperator::Lte,
                args: vec![
                    air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    }),
                    air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "b".to_string()
                    }),
                ]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$lte".to_string(),
                args: vec![
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                        "a".to_string()
                    )),
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                        "b".to_string()
                    )),
                ],
            })
        );

        test_from_expr!(
            dollar_literal_literal_becomes_literal,
            expected = air::Expression::Literal(air::LiteralValue::Integer(5)),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$literal".to_string(),
                args: vec![agg_ast::Expression::Literal(
                    agg_ast::LiteralValue::Integer(5)
                )],
            })
        );

        test_from_expr!(
            dollar_literal_string_becomes_literal,
            expected = air::Expression::Literal(air::LiteralValue::String("a".to_string())),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$literal".to_string(),
                args: vec![agg_ast::Expression::StringOrRef(
                    agg_ast::StringOrRef::String("a".to_string())
                )],
            })
        );

        test_from_expr!(
            dollar_sqlis_becomes_is,
            expected = air::Expression::Is(air::Is {
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string(),
                })),
                target_type: air::TypeOrMissing::Type(air::Type::Int32),
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$sqlIs".to_string(),
                args: vec![
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                        "a".to_string()
                    )),
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::String(
                        "int".to_string()
                    )),
                ],
            })
        );

        test_from_expr!(
            dollar_is_becomes_is,
            expected = air::Expression::Is(air::Is {
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string(),
                })),
                target_type: air::TypeOrMissing::Missing,
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$sqlIs".to_string(),
                args: vec![
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                        "a".to_string()
                    )),
                    agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::String(
                        "missing".to_string()
                    )),
                ],
            })
        );

        test_from_expr!(
            null_if,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::NullIf,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(1))]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$nullIf".to_string(),
                args: vec![agg_ast::Expression::Literal(
                    agg_ast::LiteralValue::Integer(1)
                )]
            })
        );

        test_from_expr!(
            coalesce,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Coalesce,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(1))]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: "$coalesce".to_string(),
                args: vec![agg_ast::Expression::Literal(
                    agg_ast::LiteralValue::Integer(1)
                )]
            })
        );
    }
}
