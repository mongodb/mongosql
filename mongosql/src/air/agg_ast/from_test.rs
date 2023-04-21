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

            let source = Some(default_source());

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

    mod lookup {
        use crate::{
            air::{
                self,
                agg_ast::{ast_definitions as agg_ast, from_test::default_source},
            },
            unchecked_unique_linked_hash_map,
        };

        test_from_stage!(
            empty,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                from_db: None,
                from_coll: None,
                let_vars: None,
                pipeline: Box::new(default_source()),
                as_var: "simple".to_string()
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup {
                from: None,
                let_body: None,
                pipeline: vec![],
                as_var: "simple".to_string()
            })
        );

        test_from_stage!(
            lookup_from_collection,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                from_db: None,
                from_coll: Some("coll".to_string()),
                let_vars: None,
                pipeline: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "coll".to_string()
                })),
                as_var: "collection".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup {
                from: Some(agg_ast::LookupFrom::Collection("coll".to_string())),
                let_body: None,
                pipeline: vec![],
                as_var: "collection".to_string()
            })
        );

        test_from_stage!(
            lookup_from_namespace,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                from_db: Some("db".to_string()),
                from_coll: Some("coll".to_string()),
                let_vars: None,
                pipeline: Box::new(air::Stage::Collection(air::Collection {
                    db: "db".to_string(),
                    collection: "coll".to_string()
                })),
                as_var: "namespace".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup {
                from: Some(agg_ast::LookupFrom::Namespace(agg_ast::Namespace {
                    db: "db".to_string(),
                    coll: "coll".to_string()
                })),
                let_body: None,
                pipeline: vec![],
                as_var: "namespace".to_string()
            })
        );

        test_from_stage!(
            lookup_with_let_vars,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                from_db: None,
                from_coll: None,
                let_vars: Some(vec![
                    air::LetVariable {
                        name: "vfoo_a".to_string(),
                        expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo_a".to_string()
                        }))
                    },
                    air::LetVariable {
                        name: "vfoo_b".to_string(),
                        expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo_b".to_string()
                        }))
                    }
                ]),
                pipeline: Box::new(default_source()),
                as_var: "simple".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup {
                from: None,
                let_body: Some(
                    vec![
                        (
                            "vfoo_a".to_string(),
                            agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                                "foo_a".to_string()
                            ))
                        ),
                        (
                            "vfoo_b".to_string(),
                            agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                                "foo_b".to_string()
                            ))
                        )
                    ]
                    .into_iter()
                    .collect()
                ),
                pipeline: vec![],
                as_var: "simple".to_string()
            })
        );

        test_from_stage!(
            lookup_with_pipeline,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                from_db: None,
                from_coll: None,
                let_vars: None,
                pipeline: Box::new(air::Stage::Skip(air::Skip {
                    source: Box::new(air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Project(air::Project {
                            source: Box::new(air::Stage::Project(air::Project {
                                source: Box::new(default_source()),
                                specifications: unchecked_unique_linked_hash_map! {
                                    "_id".to_string() => air::Expression::Literal(air::LiteralValue::Integer(0)),
                                    "baz".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                                }
                            })),
                            specifications: unchecked_unique_linked_hash_map! {
                                "_id".to_string() => air::Expression::Literal(air::LiteralValue::Integer(0)),
                                "baz".to_string() => air::Expression::FieldRef(air::FieldRef{parent: None, name: "baz".to_string()}),
                            }
                        })),
                        specifications: unchecked_unique_linked_hash_map! {
                            "__bot.a".to_string() => air::Expression::FieldRef(air::FieldRef {
                                parent: Some(Box::new(air::FieldRef{parent: None, name: "baz".to_string()})),
                                name: "a".to_string(),
                            }),
                            "_id".to_string() => air::Expression::Literal(air::LiteralValue::Integer(0)),
                        }
                    })),
                    skip: 1
                })),
                as_var: "simple".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup {
                from: None,
                let_body: None,
                pipeline: vec![
                    agg_ast::Stage::Project(
                        vec![
                            (
                                "_id".to_string(),
                                agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(0))
                            ),
                            (
                                "baz".to_string(),
                                agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::Variable(
                                    "ROOT".to_string()
                                ))
                            ),
                        ]
                        .into_iter()
                        .collect()
                    ),
                    agg_ast::Stage::Project(
                        vec![
                            (
                                "_id".to_string(),
                                agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(0))
                            ),
                            (
                                "baz".to_string(),
                                agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                                    "baz".to_string()
                                ))
                            ),
                        ]
                        .into_iter()
                        .collect()
                    ),
                    agg_ast::Stage::Project(
                        vec![
                            (
                                "_id".to_string(),
                                agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(0))
                            ),
                            (
                                "__bot.a".to_string(),
                                agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef(
                                    "baz.a".to_string()
                                ))
                            ),
                        ]
                        .into_iter()
                        .collect()
                    ),
                    agg_ast::Stage::Skip(1)
                ],
                as_var: "simple".to_string()
            })
        );
    }

    mod group {
        use crate::{
            air::{
                self,
                agg_ast::{ast_definitions as agg_ast, from_test::default_source},
            },
            map,
        };

        test_from_stage!(
            group_null_id_no_acc,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![],
                aggregations: vec![]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {}
            })
        );

        test_from_stage!(
            group_with_single_acc,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![],
                aggregations: vec![air::AccumulatorExpr {
                    alias: "acc".to_string(),
                    function: air::AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    }))
                }]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {
                    "acc".to_string() => agg_ast::GroupAccumulator {
                        function: "$sqlSum".to_string(),
                        expr: agg_ast::GroupAccumulatorExpr::SqlAccumulator {
                            distinct: true,
                            var: Box::new(agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("a".to_string())))
                        }
                    }
                }
            })
        );

        test_from_stage!(
            group_with_keys_and_multiple_acc,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![air::NameExprPair {
                    name: "a".to_string(),
                    expr: air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    })
                }],
                aggregations: vec![
                    air::AccumulatorExpr {
                        alias: "acc_one".to_string(),
                        function: air::AggregationFunction::Sum,
                        distinct: true,
                        arg: Box::new(air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "a".to_string()
                        }))
                    },
                    air::AccumulatorExpr {
                        alias: "acc_two".to_string(),
                        function: air::AggregationFunction::Avg,
                        distinct: true,
                        arg: Box::new(air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "b".to_string()
                        }))
                    },
                ]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Document(map! {
                    "a".to_string() => agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("a".to_string()))
                },),
                aggregations: map! {
                    "acc_one".to_string() => agg_ast::GroupAccumulator {
                        function: "$sqlSum".to_string(),
                        expr: agg_ast::GroupAccumulatorExpr::SqlAccumulator {
                            distinct: true,
                            var: Box::new(agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("a".to_string())))
                        },
                    },
                    "acc_two".to_string() => agg_ast::GroupAccumulator {
                        function: "$sqlAvg".to_string(),
                        expr: agg_ast::GroupAccumulatorExpr::SqlAccumulator {
                            distinct: true,
                            var: Box::new(agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("b".to_string())))
                        },
                    },
                }
            })
        );

        test_from_stage!(
            group_with_non_sql_acc,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![],
                aggregations: vec![air::AccumulatorExpr {
                    alias: "acc".to_string(),
                    function: air::AggregationFunction::AddToSet,
                    distinct: true,
                    arg: Box::new(air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "a".to_string()
                    }))
                }]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {
                    "acc".to_string() => agg_ast::GroupAccumulator {
                        function: "$addToSet".to_string(),
                        expr: agg_ast::GroupAccumulatorExpr::NonSqlAccumulator(agg_ast::Expression::StringOrRef(agg_ast::StringOrRef::FieldRef("a".to_string()))),
                    }
                }
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
            map, unchecked_unique_linked_hash_map,
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
            expected = air::Expression::SqlDivide(air::SqlDivide {
                dividend: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "a".to_string()
                })),
                divisor: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "b".to_string()
                })),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
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

        test_from_expr!(
            sql_subquery,
            expected = air::Expression::Subquery(air::Subquery {
                let_bindings: vec![air::LetVariable {
                    name: "z".to_string(),
                    expr: air::Expression::Literal(air::LiteralValue::Integer(42)).into()
                },],
                output_path: vec!["x".to_string()],
                pipeline: air::Stage::Project(air::Project {
                    source: air::Stage::Documents(air::Documents { array: vec![] }).into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "x".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Subquery(
                agg_ast::Subquery {
                    db: Some("foo".to_string()),
                    collection: Some("bar".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))
                    }),
                    output_path: Some(vec!["x".to_string()]),
                    pipeline: vec![
                        agg_ast::Stage::Documents(vec![]),
                        agg_ast::Stage::Project(
                            map! {"x".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))}
                        )
                    ]
                }
            ))
        );
        test_from_expr!(
            sql_subquery_rootless,
            expected = air::Expression::Subquery(air::Subquery {
                let_bindings: vec![air::LetVariable {
                    name: "z".to_string(),
                    expr: air::Expression::Literal(air::LiteralValue::Integer(42)).into()
                },],
                output_path: vec!["x".to_string()],
                pipeline: air::Stage::Project(air::Project {
                    source: air::Stage::Collection(air::Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string()
                    })
                    .into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "x".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Subquery(
                agg_ast::Subquery {
                    db: Some("foo".to_string()),
                    collection: Some("bar".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))
                    }),
                    output_path: Some(vec!["x".to_string()]),
                    pipeline: vec![agg_ast::Stage::Project(
                        map! {"x".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))}
                    )]
                }
            ))
        );

        test_from_expr!(
            sql_subquery_comparison,
            expected = air::Expression::SubqueryComparison(
                air::SubqueryComparison {
                    op: air::SubqueryComparisonOp::Eq,
                    modifier: air::SubqueryModifier::All,
                    arg: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
                    subquery: air::Subquery {
                        let_bindings: vec![air::LetVariable {
                            name: "z".to_string(),
                            expr: air::Expression::Literal(air::LiteralValue::Integer(42)).into()
                        },],
                        output_path: vec!["x".to_string()],
                        pipeline: air::Stage::Project(air::Project {
                            source: air::Stage::Documents(air::Documents { array: vec![] }).into(),
                            specifications: unchecked_unique_linked_hash_map! {
                                "x".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                            }
                        }).into()
                    }.into()
                }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryComparison(
                   agg_ast::SubqueryComparison {
                       op: "eq".to_string(),
                       modifier: "all".to_string(),
                       arg: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))),
                       subquery: agg_ast::Subquery {
                           db: Some("foo".to_string()),
                           collection: Some("bar".to_string()),
                           let_bindings: Some(map! {
                               "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))
                           }),
                           output_path: Some(vec!["x".to_string()]),
                           pipeline: vec![
                               agg_ast::Stage::Documents(vec![]),
                               agg_ast::Stage::Project(
                                   map! {"x".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))}
                               )
                           ]
                       }.into()
                   }
               ))
        );
        test_from_expr!(
            sql_subquery_comparison_rootless,
            expected = air::Expression::SubqueryComparison(
                air::SubqueryComparison {
                    op: air::SubqueryComparisonOp::Neq,
                    modifier: air::SubqueryModifier::Any,
                    arg: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
                    subquery: air::Subquery {
                        let_bindings: vec![air::LetVariable {
                            name: "z".to_string(),
                            expr: air::Expression::Literal(air::LiteralValue::Integer(42)).into()
                        },],
                        output_path: vec!["x".to_string()],
                        pipeline: air::Stage::Project(air::Project {
                            source: air::Stage::Collection(air::Collection { db: "foo".to_string(), collection: "bar2".to_string() }).into(),
                            specifications: unchecked_unique_linked_hash_map! {
                                "x".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                            }
                        }).into()
                    }.into()
                }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryComparison(
                   agg_ast::SubqueryComparison {
                       op: "ne".to_string(),
                       modifier: "any".to_string(),
                       arg: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))),
                       subquery: agg_ast::Subquery {
                           db: Some("foo".to_string()),
                           collection: Some("bar2".to_string()),
                           let_bindings: Some(map! {
                               "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))
                           }),
                           output_path: Some(vec!["x".to_string()]),
                           pipeline: vec![
                               agg_ast::Stage::Project(
                                   map! {"x".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))}
                               )
                           ]
                       }.into()
                   }
               ))
        );

        test_from_expr!(
            sql_subquery_exists,
            expected = air::Expression::SubqueryExists(air::SubqueryExists {
                let_bindings: vec![air::LetVariable {
                    name: "z".to_string(),
                    expr: air::Expression::Literal(air::LiteralValue::Integer(42)).into()
                },],
                pipeline: air::Stage::Project(air::Project {
                    source: air::Stage::Documents(air::Documents { array: vec![] }).into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "x".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryExists(
                agg_ast::SubqueryExists {
                    db: Some("foo".to_string()),
                    collection: Some("bar".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))
                    }),
                    pipeline: vec![
                        agg_ast::Stage::Documents(vec![]),
                        agg_ast::Stage::Project(
                            map! {"x".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))}
                        )
                    ]
                }
            ))
        );
        test_from_expr!(
            sql_subquery_exists_rootless,
            expected = air::Expression::SubqueryExists(air::SubqueryExists {
                let_bindings: vec![air::LetVariable {
                    name: "z".to_string(),
                    expr: air::Expression::Literal(air::LiteralValue::Integer(42)).into()
                },],
                pipeline: air::Stage::Project(air::Project {
                    source: air::Stage::Collection(air::Collection {
                        db: "foo2".to_string(),
                        collection: "bar2".to_string()
                    })
                    .into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "x".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryExists(
                agg_ast::SubqueryExists {
                    db: Some("foo2".to_string()),
                    collection: Some("bar2".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))
                    }),
                    pipeline: vec![agg_ast::Stage::Project(
                        map! {"x".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(1))}
                    )]
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
