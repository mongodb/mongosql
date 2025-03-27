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
        use crate::air::{self, agg_ast::from_test::default_source};
        use agg_ast::definitions as agg_ast;

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
            air::{self, agg_ast::from_test::default_source},
            map, unchecked_unique_linked_hash_map,
        };
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            empty,
            expected = air::Stage::Documents(air::Documents { array: vec![] }),
            input = agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![]))
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
            input = agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![map! {
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1))
            }]))
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
            input = agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![
                map! {
                    "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1))
                },
                map! {
                    "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(2))
                },
                map! {
                    "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(3))
                },
            ]))
        );
    }

    mod project {
        use crate::{
            air::{self, agg_ast::from_test::default_source},
            map, unchecked_unique_linked_hash_map,
        };
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            empty,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {},
            }),
            input = agg_ast::Stage::Project(agg_ast::ProjectStage { items: map! {} })
        );

        test_from_stage!(
            singleton_exclusion,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::ProjectItem::Exclusion,
                },
            }),
            input = agg_ast::Stage::Project(agg_ast::ProjectStage {
                items: map! {
                    "a".to_string() => agg_ast::ProjectItem::Exclusion,
                }
            })
        );

        test_from_stage!(
            singleton_inclusion,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::ProjectItem::Inclusion,
                },
            }),
            input = agg_ast::Stage::Project(agg_ast::ProjectStage {
                items: map! {
                    "a".to_string() => agg_ast::ProjectItem::Inclusion,
                }
            })
        );

        test_from_stage!(
            singleton_assignment,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                },
            }),
            input = agg_ast::Stage::Project(agg_ast::ProjectStage {
                items: map! {
                    "a".to_string() => agg_ast::ProjectItem::Assignment(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1))),
                }
            })
        );

        test_from_stage!(
            multiple_elements,
            expected = air::Stage::Project(air::Project {
                source: Box::new(default_source()),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    "b".to_string() => air::ProjectItem::Exclusion,
                    "c".to_string() => air::ProjectItem::Inclusion,
                },
            }),
            input = agg_ast::Stage::Project(agg_ast::ProjectStage {
                items: map! {
                    "a".to_string() => agg_ast::ProjectItem::Assignment(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1))),
                    "b".to_string() => agg_ast::ProjectItem::Exclusion,
                    "c".to_string() => agg_ast::ProjectItem::Inclusion,
                }
            })
        );
    }

    mod replace_with {
        use crate::air::{self, agg_ast::from_test::default_source};
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            simple,
            expected = air::Stage::ReplaceWith(air::ReplaceWith {
                source: Box::new(default_source()),
                new_root: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Stage::ReplaceWith(agg_ast::ReplaceStage::Expression(
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)
            ))
        );
    }

    mod match_stage {
        use crate::air::{self, agg_ast::from_test::default_source};
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            expr,
            expected = air::Stage::Match(air::Match::ExprLanguage(air::ExprLanguage {
                source: Box::new(default_source()),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Boolean(true))),
            })),
            input = agg_ast::Stage::Match(agg_ast::MatchStage {
                expr: vec![agg_ast::MatchExpression::Expr(agg_ast::MatchExpr {
                    expr: Box::new(agg_ast::Expression::Literal(
                        agg_ast::LiteralValue::Boolean(true)
                    )),
                })]
            })
        );
    }

    mod limit_skip {
        use crate::air::{self, agg_ast::from_test::default_source};
        use agg_ast::definitions as agg_ast;

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
            air::{self, agg_ast::from_test::default_source},
            map,
        };
        use agg_ast::definitions as agg_ast;

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

    mod unwind {
        use crate::air::{self, agg_ast::from_test::default_source};
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            unwind_field_path,
            expected = air::Stage::Unwind(air::Unwind {
                source: Box::new(default_source()),
                path: air::Expression::FieldRef(
                    "eca58228-b657-498a-b76e-f48a9161a404".to_string().into()
                ),
                index: None,
                outer: false
            }),
            input = agg_ast::Stage::Unwind(agg_ast::Unwind::FieldPath(agg_ast::Expression::Ref(
                agg_ast::Ref::FieldRef("eca58228-b657-498a-b76e-f48a9161a404".to_string())
            )))
        );

        test_from_stage!(
            unwind_document_no_options,
            expected = air::Stage::Unwind(air::Unwind {
                source: Box::new(default_source()),
                path: air::Expression::FieldRef("array".to_string().into()),
                index: None,
                outer: false
            }),
            input = agg_ast::Stage::Unwind(agg_ast::Unwind::Document(agg_ast::UnwindExpr {
                path: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                    "array".to_string()
                ))),
                include_array_index: None,
                preserve_null_and_empty_arrays: None
            }))
        );

        test_from_stage!(
            unwind_document_all_options,
            expected = air::Stage::Unwind(air::Unwind {
                source: Box::new(default_source()),
                path: air::Expression::FieldRef("array".to_string().into()),
                index: Some("i".to_string()),
                outer: true
            }),
            input = agg_ast::Stage::Unwind(agg_ast::Unwind::Document(agg_ast::UnwindExpr {
                path: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                    "array".to_string()
                ))),
                include_array_index: Some("i".to_string()),
                preserve_null_and_empty_arrays: Some(true)
            }))
        );
    }

    mod join {
        use crate::{
            air::{self, agg_ast::from_test::default_source},
            map, unchecked_unique_linked_hash_map,
        };
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            inner_join,
            expected = air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: Box::new(default_source()),
                right: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "bar".to_string()
                })),
                let_vars: None,
                condition: None
            }),
            input = agg_ast::Stage::Join(Box::new(agg_ast::Join {
                database: None,
                collection: Some("bar".to_string()),
                let_body: None,
                join_type: agg_ast::JoinType::Inner,
                pipeline: vec![],
                condition: None
            }))
        );

        test_from_stage!(
            left_join_with_db,
            expected = air::Stage::Join(air::Join {
                join_type: air::JoinType::Left,
                left: Box::new(default_source()),
                right: Box::new(air::Stage::Collection(air::Collection {
                    db: "db".to_string(),
                    collection: "bar".to_string()
                })),
                let_vars: None,
                condition: None
            }),
            input = agg_ast::Stage::Join(Box::new(agg_ast::Join {
                database: Some("db".to_string()),
                collection: Some("bar".to_string()),
                let_body: None,
                join_type: agg_ast::JoinType::Left,
                pipeline: vec![],
                condition: None
            }))
        );

        test_from_stage!(
            join_with_no_collection_and_pipeline,
            expected = air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: Box::new(default_source()),
                right: Box::new(air::Stage::Documents(air::Documents {
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
                })),
                let_vars: None,
                condition: None
            }),
            input = agg_ast::Stage::Join(Box::new(agg_ast::Join {
                database: None,
                collection: None,
                let_body: None,
                join_type: agg_ast::JoinType::Inner,
                pipeline: vec![agg_ast::Stage::Documents(agg_ast::Documents::Literals(
                    vec![
                        map! {"a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1)) },
                        map! {"a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(2)) },
                        map! {"a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(3)) },
                    ]
                ))],
                condition: None
            }))
        );

        test_from_stage!(
            join_with_let_vars_and_condition,
            expected = air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: Box::new(default_source()),
                right: Box::new(air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string()
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "_id".to_string() => air::ProjectItem::Exclusion,
                        "x".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    }
                })),
                let_vars: Some(vec![air::LetVariable {
                    name: "x".to_string(),
                    expr: Box::new(air::Expression::FieldRef("x".to_string().into()))
                }]),
                condition: Some(air::Expression::SQLSemanticOperator(
                    air::SQLSemanticOperator {
                        op: air::SQLOperator::Eq,
                        args: vec![
                            air::Expression::Variable("x".to_string().into()),
                            air::Expression::FieldRef("x".to_string().into())
                        ]
                    }
                ))
            }),
            input = agg_ast::Stage::Join(Box::new(agg_ast::Join {
                database: None,
                collection: Some("bar".to_string()),
                let_body: Some(map! {
                    "x".to_string() => agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("x".to_string()))
                }),
                join_type: agg_ast::JoinType::Inner,
                pipeline: vec![agg_ast::Stage::Project(agg_ast::ProjectStage {
                    items: map! {
                        "_id".to_string() => agg_ast::ProjectItem::Exclusion,
                        "x".to_string() => agg_ast::ProjectItem::Assignment(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1))),
                    }
                })],
                condition: Some(agg_ast::Expression::UntaggedOperator(
                    agg_ast::UntaggedOperator {
                        op: agg_ast::UntaggedOperatorName::SQLEq,
                        args: vec![
                            agg_ast::Expression::Ref(agg_ast::Ref::VariableRef("x".to_string())),
                            agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("x".to_string())),
                        ]
                    }
                ))
            }))
        );

        test_from_stage!(
            nested_join,
            expected = air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: Box::new(default_source()),
                right: Box::new(air::Stage::Join(air::Join {
                    join_type: air::JoinType::Inner,
                    left: Box::new(air::Stage::Collection(air::Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string()
                    })),
                    right: Box::new(air::Stage::Join(air::Join {
                        join_type: air::JoinType::Inner,
                        left: Box::new(air::Stage::Collection(air::Collection {
                            db: "test".to_string(),
                            collection: "baz".to_string(),
                        })),
                        right: Box::new(air::Stage::Collection(air::Collection {
                            db: "test".to_string(),
                            collection: "car".to_string()
                        })),
                        let_vars: None,
                        condition: None
                    })),
                    let_vars: None,
                    condition: None
                })),
                let_vars: None,
                condition: None
            }),
            input = agg_ast::Stage::Join(Box::new(agg_ast::Join {
                database: None,
                collection: Some("bar".to_string()),
                let_body: None,
                join_type: agg_ast::JoinType::Inner,
                pipeline: vec![agg_ast::Stage::Join(Box::new(agg_ast::Join {
                    database: None,
                    collection: Some("baz".to_string()),
                    join_type: agg_ast::JoinType::Inner,
                    let_body: None,
                    pipeline: vec![agg_ast::Stage::Join(Box::new(agg_ast::Join {
                        database: None,
                        collection: Some("car".to_string()),
                        join_type: agg_ast::JoinType::Inner,
                        let_body: None,
                        pipeline: vec![],
                        condition: None
                    }))],
                    condition: None
                }))],
                condition: None
            }))
        );
    }

    mod lookup {
        use crate::{
            air::{self, agg_ast::from_test::default_source},
            unchecked_unique_linked_hash_map,
            util::{ROOT, ROOT_NAME},
        };
        use agg_ast::definitions as agg_ast;

        test_from_stage!(
            empty,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                let_vars: None,
                pipeline: Box::new(default_source()),
                as_var: "simple".to_string()
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup::Subquery(agg_ast::SubqueryLookup {
                from: None,
                let_body: None,
                pipeline: vec![],
                as_var: "simple".to_string()
            }))
        );

        test_from_stage!(
            lookup_from_collection,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                let_vars: None,
                pipeline: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "coll".to_string()
                })),
                as_var: "collection".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup::Subquery(agg_ast::SubqueryLookup {
                from: Some(agg_ast::LookupFrom::Collection("coll".to_string())),
                let_body: None,
                pipeline: vec![],
                as_var: "collection".to_string()
            }))
        );

        test_from_stage!(
            lookup_from_namespace,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                let_vars: None,
                pipeline: Box::new(air::Stage::Collection(air::Collection {
                    db: "db".to_string(),
                    collection: "coll".to_string()
                })),
                as_var: "namespace".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup::Subquery(agg_ast::SubqueryLookup {
                from: Some(agg_ast::LookupFrom::Namespace(agg_ast::Namespace {
                    db: "db".to_string(),
                    coll: "coll".to_string()
                })),
                let_body: None,
                pipeline: vec![],
                as_var: "namespace".to_string()
            }))
        );

        test_from_stage!(
            lookup_with_let_vars,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                let_vars: Some(vec![
                    air::LetVariable {
                        name: "vfoo_a".to_string(),
                        expr: Box::new(air::Expression::FieldRef("foo_a".to_string().into()))
                    },
                    air::LetVariable {
                        name: "vfoo_b".to_string(),
                        expr: Box::new(air::Expression::FieldRef("foo_b".to_string().into()))
                    }
                ]),
                pipeline: Box::new(default_source()),
                as_var: "simple".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup::Subquery(agg_ast::SubqueryLookup {
                from: None,
                let_body: Some(
                    vec![
                        (
                            "vfoo_a".to_string(),
                            agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("foo_a".to_string()))
                        ),
                        (
                            "vfoo_b".to_string(),
                            agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("foo_b".to_string()))
                        )
                    ]
                    .into_iter()
                    .collect()
                ),
                pipeline: vec![],
                as_var: "simple".to_string()
            }))
        );

        test_from_stage!(
            lookup_with_pipeline,
            expected = air::Stage::Lookup(air::Lookup {
                source: Box::new(default_source()),
                let_vars: None,
                pipeline: Box::new(air::Stage::Skip(air::Skip {
                    source: Box::new(air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Project(air::Project {
                            source: Box::new(air::Stage::Project(air::Project {
                                source: Box::new(default_source()),
                                specifications: unchecked_unique_linked_hash_map! {
                                    "_id".to_string() => air::ProjectItem::Exclusion,
                                    "baz".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                                }
                            })),
                            specifications: unchecked_unique_linked_hash_map! {
                                "_id".to_string() => air::ProjectItem::Exclusion,
                                "baz".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("baz".to_string().into())),
                            }
                        })),
                        specifications: unchecked_unique_linked_hash_map! {
                            "__bot.a".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("baz.a".to_string().into())),
                            "_id".to_string() => air::ProjectItem::Exclusion,
                        }
                    })),
                    skip: 1
                })),
                as_var: "simple".to_string(),
            }),
            input = agg_ast::Stage::Lookup(agg_ast::Lookup::Subquery(agg_ast::SubqueryLookup {
                from: None,
                let_body: None,
                pipeline: vec![
                    agg_ast::Stage::Project(agg_ast::ProjectStage {
                        items: vec![
                            ("_id".to_string(), agg_ast::ProjectItem::Exclusion),
                            (
                                "baz".to_string(),
                                agg_ast::ProjectItem::Assignment(agg_ast::Expression::Ref(
                                    agg_ast::Ref::VariableRef(ROOT_NAME.to_string())
                                ))
                            ),
                        ]
                        .into_iter()
                        .collect()
                    }),
                    agg_ast::Stage::Project(agg_ast::ProjectStage {
                        items: vec![
                            ("_id".to_string(), agg_ast::ProjectItem::Exclusion),
                            (
                                "baz".to_string(),
                                agg_ast::ProjectItem::Assignment(agg_ast::Expression::Ref(
                                    agg_ast::Ref::FieldRef("baz".to_string())
                                ))
                            ),
                        ]
                        .into_iter()
                        .collect()
                    }),
                    agg_ast::Stage::Project(agg_ast::ProjectStage {
                        items: vec![
                            ("_id".to_string(), agg_ast::ProjectItem::Exclusion),
                            (
                                "__bot.a".to_string(),
                                agg_ast::ProjectItem::Assignment(agg_ast::Expression::Ref(
                                    agg_ast::Ref::FieldRef("baz.a".to_string())
                                ))
                            ),
                        ]
                        .into_iter()
                        .collect()
                    }),
                    agg_ast::Stage::Skip(1)
                ],
                as_var: "simple".to_string()
            }))
        );
    }

    mod group {
        use crate::{
            air::{self, agg_ast::from_test::default_source},
            map,
            schema::Satisfaction,
        };
        use agg_ast::definitions as agg_ast;

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
            group_with_single_acc_possibly_doc_must,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![],
                aggregations: vec![air::AccumulatorExpr {
                    alias: "acc".to_string(),
                    function: air::AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(air::Expression::FieldRef("a".to_string().into())),
                    arg_is_possibly_doc: Satisfaction::Must,
                }]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {
                "acc".to_string() => agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLSum(
                    agg_ast::SQLAccumulator {
                        distinct: true,
                        var: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))),
                        arg_is_possibly_doc: Some("must".to_string()),
                    }))
                }
            })
        );

        test_from_stage!(
            group_with_single_acc_possibly_doc_may,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![],
                aggregations: vec![air::AccumulatorExpr {
                    alias: "acc".to_string(),
                    function: air::AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(air::Expression::FieldRef("a".to_string().into())),
                    arg_is_possibly_doc: Satisfaction::May,
                }]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {
                    "acc".to_string() => agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLSum(
                        agg_ast::SQLAccumulator {
                            distinct: true,
                            var: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))),
                            arg_is_possibly_doc: Some("may".to_string()),
                        }
                    ))
                }
            })
        );

        test_from_stage!(
            group_with_single_acc_possibly_doc_not,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![],
                aggregations: vec![air::AccumulatorExpr {
                    alias: "acc".to_string(),
                    function: air::AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(air::Expression::FieldRef("a".to_string().into())),
                    arg_is_possibly_doc: Satisfaction::Not,
                }]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {
                "acc".to_string() => agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLSum(
                    agg_ast::SQLAccumulator {
                        distinct: true,
                        var: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))),
                        arg_is_possibly_doc: Some("not".to_string()),
                    }))
                }
            })
        );

        test_from_stage!(
            group_with_keys_and_multiple_acc,
            expected = air::Stage::Group(air::Group {
                source: Box::new(default_source()),
                keys: vec![air::NameExprPair {
                    name: "a".to_string(),
                    expr: air::Expression::FieldRef("a".to_string().into())
                }],
                aggregations: vec![
                    air::AccumulatorExpr {
                        alias: "acc_one".to_string(),
                        function: air::AggregationFunction::Sum,
                        distinct: true,
                        arg: Box::new(air::Expression::FieldRef("a".to_string().into())),
                        arg_is_possibly_doc: Satisfaction::Not,
                    },
                    air::AccumulatorExpr {
                        alias: "acc_two".to_string(),
                        function: air::AggregationFunction::Avg,
                        distinct: true,
                        arg: Box::new(air::Expression::FieldRef("b".to_string().into())),
                        arg_is_possibly_doc: Satisfaction::Not,
                    },
                ]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Document(map! {
                    "a".to_string() => agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))
                },),
                aggregations: map! {
                    "acc_one".to_string() => agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLSum(
                        agg_ast::SQLAccumulator {
                            distinct: true,
                            var: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))),
                            arg_is_possibly_doc: None,
                        },
                    )),
                    "acc_two".to_string() => agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLAvg(
                        agg_ast::SQLAccumulator {
                            distinct: true,
                            var: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("b".to_string()))),
                            arg_is_possibly_doc: None,
                        },
                    )),
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
                    distinct: false,
                    arg: Box::new(air::Expression::FieldRef("a".to_string().into())),
                    arg_is_possibly_doc: Satisfaction::Not,
                }]
            }),
            input = agg_ast::Stage::Group(agg_ast::Group {
                keys: agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
                aggregations: map! {
                    "acc".to_string() => agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                        op: agg_ast::UntaggedOperatorName::AddToSet,
                        args: vec![agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))],
                    })
                }
            })
        );
    }
}

mod expression {
    mod literal {
        use crate::air;
        use agg_ast::definitions as agg_ast;

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
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(10))
        );

        test_from_expr!(
            long,
            expected = air::Expression::Literal(air::LiteralValue::Long(12000000000)),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Int64(12000000000))
        );

        test_from_expr!(
            double,
            expected = air::Expression::Literal(air::LiteralValue::Double(14.3)),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::Double(14.3))
        );
    }

    mod string_or_ref {
        use crate::air;
        use agg_ast::definitions as agg_ast;

        test_from_expr!(
            string,
            expected = air::Expression::Literal(air::LiteralValue::String("s".to_string())),
            input = agg_ast::Expression::Literal(agg_ast::LiteralValue::String("s".to_string()))
        );

        test_from_expr!(
            empty_field_ref,
            expected = air::Expression::FieldRef("".to_string().into()),
            input = agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("".to_string()))
        );

        test_from_expr!(
            simple_field_ref,
            expected = air::Expression::FieldRef("a".to_string().into()),
            input = agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string()))
        );

        test_from_expr!(
            nested_field_ref,
            expected = air::Expression::FieldRef("a.b.c".to_string().into()),
            input = agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a.b.c".to_string()))
        );

        test_from_expr!(
            simple_variable,
            expected = air::Expression::Variable("v".to_string().into()),
            input = agg_ast::Expression::Ref(agg_ast::Ref::VariableRef("v".to_string()))
        );

        test_from_expr!(
            nested_variable,
            expected = air::Expression::Variable("x.y.z".to_string().into()),
            input = agg_ast::Expression::Ref(agg_ast::Ref::VariableRef("x.y.z".to_string()))
        );
    }

    mod array {
        use crate::air;
        use agg_ast::definitions as agg_ast;

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
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1)),
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(2)),
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(3)),
            ])
        );
    }

    mod document {
        use crate::{air, map, unchecked_unique_linked_hash_map};
        use agg_ast::definitions as agg_ast;

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
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(1)),
                "b".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(2)),
                "c".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(3)),
            })
        );
    }

    mod tagged_operators {
        use crate::{air, map, unchecked_unique_linked_hash_map};
        use agg_ast::definitions as agg_ast;

        test_from_expr!(
            get_field,
            expected = air::Expression::GetField(air::GetField {
                field: "a".to_string(),
                input: Box::new(air::Expression::FieldRef("d".to_string().into()))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::GetField(
                agg_ast::GetField {
                    field: "a".to_string(),
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "d".to_string()
                    )))
                }
            ))
        );

        test_from_expr!(
            set_field,
            expected = air::Expression::SetField(air::SetField {
                field: "a".to_string(),
                input: Box::new(air::Expression::FieldRef("d".to_string().into())),
                value: Box::new(air::Expression::Literal(air::LiteralValue::Null))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SetField(
                agg_ast::SetField {
                    field: "a".to_string(),
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "d".to_string()
                    ))),
                    value: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null))
                }
            ))
        );

        test_from_expr!(
            unset_field,
            expected = air::Expression::UnsetField(air::UnsetField {
                field: "a".to_string(),
                input: Box::new(air::Expression::FieldRef("d".to_string().into()))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::UnsetField(
                agg_ast::UnsetField {
                    field: "a".to_string(),
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "d".to_string()
                    )))
                }
            ))
        );

        test_from_expr!(
            switch,
            expected = air::Expression::Switch(air::Switch {
                branches: vec![
                    air::SwitchCase {
                        case: Box::new(air::Expression::FieldRef("a".to_string().into())),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    },
                    air::SwitchCase {
                        case: Box::new(air::Expression::FieldRef("b".to_string().into())),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(2))),
                    },
                ],
                default: Box::new(air::Expression::Literal(air::LiteralValue::Integer(3)))
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Switch(
                agg_ast::Switch {
                    branches: vec![
                        agg_ast::SwitchCase {
                            case: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                                "a".to_string()
                            ))),
                            then: Box::new(agg_ast::Expression::Literal(
                                agg_ast::LiteralValue::Int32(1)
                            )),
                        },
                        agg_ast::SwitchCase {
                            case: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                                "b".to_string()
                            ))),
                            then: Box::new(agg_ast::Expression::Literal(
                                agg_ast::LiteralValue::Int32(2)
                            )),
                        },
                    ],
                    default: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(
                        3
                    )))
                }
            ))
        );

        test_from_expr!(
            let_expr,
            expected = air::Expression::Let(air::Let {
                vars: vec![air::LetVariable {
                    name: "v".to_string(),
                    expr: Box::new(air::Expression::FieldRef("a".to_string().into())),
                },],
                inside: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Let(
                agg_ast::Let {
                    vars: map! {
                        "v".to_string() => agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    },
                    inside: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );

        test_from_expr!(
            convert,
            expected = air::Expression::Convert(air::Convert {
                input: Box::new(air::Expression::FieldRef("a".to_string().into())),
                to: air::Type::Int32,
                on_null: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Convert(
                agg_ast::Convert {
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "a".to_string()
                    ))),
                    to: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::String(
                        "int".to_string()
                    ))),
                    format: None,
                    on_null: Some(Box::new(agg_ast::Expression::Literal(
                        agg_ast::LiteralValue::Null
                    ))),
                    on_error: Some(Box::new(agg_ast::Expression::Literal(
                        agg_ast::LiteralValue::Null
                    ))),
                }
            ))
        );

        test_from_expr!(
            sql_convert,
            expected = air::Expression::SqlConvert(air::SqlConvert {
                input: Box::new(air::Expression::FieldRef("a".to_string().into())),
                to: air::SqlConvertTargetType::Array,
                on_null: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLConvert(
                agg_ast::SQLConvert {
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "a".to_string()
                    ))),
                    to: "array".to_string(),
                    on_null: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                    on_error: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );

        test_from_expr!(
            like_with_escape,
            expected = air::Expression::Like(air::Like {
                expr: Box::new(air::Expression::FieldRef("s".to_string().into())),
                pattern: Box::new(air::Expression::Literal(air::LiteralValue::String(
                    "pat".to_string()
                ))),
                escape: Some('e')
            }),
            input =
                agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Like(agg_ast::Like {
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "s".to_string()
                    ))),
                    pattern: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::String(
                        "pat".to_string()
                    ))),
                    escape: Some('e'),
                }))
        );

        test_from_expr!(
            like_without_escape,
            expected = air::Expression::Like(air::Like {
                expr: Box::new(air::Expression::FieldRef("s".to_string().into())),
                pattern: Box::new(air::Expression::Literal(air::LiteralValue::String(
                    "pat".to_string()
                ))),
                escape: None
            }),
            input =
                agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Like(agg_ast::Like {
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "s".to_string()
                    ))),
                    pattern: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::String(
                        "pat".to_string()
                    ))),
                    escape: None,
                }))
        );

        test_from_expr!(
            sql_divide,
            expected = air::Expression::SqlDivide(air::SqlDivide {
                dividend: Box::new(air::Expression::FieldRef("a".to_string().into())),
                divisor: Box::new(air::Expression::FieldRef("b".to_string().into())),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SQLDivide(
                agg_ast::SQLDivide {
                    dividend: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "a".to_string()
                    ))),
                    divisor: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "b".to_string()
                    ))),
                    on_error: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
                }
            ))
        );

        test_from_expr!(
            reduce,
            expected = air::Expression::Reduce(air::Reduce {
                input: Box::new(air::Expression::FieldRef("a".to_string().into())),
                init_value: Box::new(air::Expression::FieldRef("b".to_string().into())),
                inside: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Reduce(
                agg_ast::Reduce {
                    input: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "a".to_string()
                    ))),
                    initial_value: Box::new(agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                        "b".to_string()
                    ))),
                    inside: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Null)),
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
                        "x".to_string() => air::ProjectItem::Inclusion,
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Subquery(
                agg_ast::Subquery {
                    db: Some("foo".to_string()),
                    collection: Some("bar".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                    }),
                    output_path: Some(vec!["x".to_string()]),
                    pipeline: vec![
                        agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![])),
                        agg_ast::Stage::Project(agg_ast::ProjectStage {
                            items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                        })
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
                        "x".to_string() => air::ProjectItem::Inclusion,
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::Subquery(
                agg_ast::Subquery {
                    db: Some("foo".to_string()),
                    collection: Some("bar".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                    }),
                    output_path: Some(vec!["x".to_string()]),
                    pipeline: vec![agg_ast::Stage::Project(agg_ast::ProjectStage {
                        items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                    })]
                }
            ))
        );

        test_from_expr!(
            sql_subquery_comparison_sql_op,
            expected = air::Expression::SubqueryComparison(
                air::SubqueryComparison {
                    op: air::SubqueryComparisonOp::Eq,
                    op_type: air::SubqueryComparisonOpType::Sql,
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
                                "x".to_string() => air::ProjectItem::Inclusion,
                            }
                        }).into()
                    }.into()
                }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryComparison(
                   agg_ast::SubqueryComparison {
                       op: "sqlEq".to_string(),
                       modifier: "all".to_string(),
                       arg: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))),
                       subquery: agg_ast::Subquery {
                           db: Some("foo".to_string()),
                           collection: Some("bar".to_string()),
                           let_bindings: Some(map! {
                               "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                           }),
                           output_path: Some(vec!["x".to_string()]),
                           pipeline: vec![
                                agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![])),
                               agg_ast::Stage::Project(
                                   agg_ast::ProjectStage {
                                        items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                                   }
                               )
                           ]
                       }.into()
                   }
               ))
        );

        test_from_expr!(
            sql_subquery_comparison_mql_op,
            expected = air::Expression::SubqueryComparison(
                air::SubqueryComparison {
                    op: air::SubqueryComparisonOp::Eq,
                    op_type: air::SubqueryComparisonOpType::Mql,
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
                                "x".to_string() => air::ProjectItem::Inclusion,
                            }
                        }).into()
                    }.into()
                }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryComparison(
                   agg_ast::SubqueryComparison {
                       op: "eq".to_string(),
                       modifier: "all".to_string(),
                       arg: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))),
                       subquery: agg_ast::Subquery {
                           db: Some("foo".to_string()),
                           collection: Some("bar".to_string()),
                           let_bindings: Some(map! {
                               "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                           }),
                           output_path: Some(vec!["x".to_string()]),
                           pipeline: vec![
                                agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![])),
                               agg_ast::Stage::Project(
                                   agg_ast::ProjectStage {
                                        items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                                   }
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
                    op_type: air::SubqueryComparisonOpType::Mql,
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
                                "x".to_string() => air::ProjectItem::Inclusion,
                            }
                        }).into()
                    }.into()
                }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryComparison(
                   agg_ast::SubqueryComparison {
                       op: "ne".to_string(),
                       modifier: "any".to_string(),
                       arg: Box::new(agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))),
                       subquery: agg_ast::Subquery {
                           db: Some("foo".to_string()),
                           collection: Some("bar2".to_string()),
                           let_bindings: Some(map! {
                               "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                           }),
                           output_path: Some(vec!["x".to_string()]),
                           pipeline: vec![
                               agg_ast::Stage::Project(
                                   agg_ast::ProjectStage {
                                        items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                                   }
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
                        "x".to_string() => air::ProjectItem::Inclusion,
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryExists(
                agg_ast::SubqueryExists {
                    db: Some("foo".to_string()),
                    collection: Some("bar".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                    }),
                    pipeline: vec![
                        agg_ast::Stage::Documents(agg_ast::Documents::Literals(vec![])),
                        agg_ast::Stage::Project(agg_ast::ProjectStage {
                            items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                        })
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
                        "x".to_string() => air::ProjectItem::Inclusion
                    }
                })
                .into()
            }),
            input = agg_ast::Expression::TaggedOperator(agg_ast::TaggedOperator::SubqueryExists(
                agg_ast::SubqueryExists {
                    db: Some("foo2".to_string()),
                    collection: Some("bar2".to_string()),
                    let_bindings: Some(map! {
                        "z".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(42))
                    }),
                    pipeline: vec![agg_ast::Stage::Project(agg_ast::ProjectStage {
                        items: map! {"x".to_string() => agg_ast::ProjectItem::Inclusion}
                    })]
                }
            ))
        );
    }

    mod untagged_operators {
        use crate::air;
        use agg_ast::definitions as agg_ast;
        use linked_hash_map::LinkedHashMap;
        use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;

        test_from_expr!(
            sql_op_one_arg,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Pos,
                args: vec![air::Expression::FieldRef("a".to_string().into())]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::SQLPos,
                args: vec![agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                    "a".to_string()
                ))],
            })
        );

        test_from_expr!(
            sql_op_multiple_args,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Eq,
                args: vec![
                    air::Expression::FieldRef("a".to_string().into()),
                    air::Expression::FieldRef("b".to_string().into()),
                ]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::SQLEq,
                args: vec![
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("b".to_string())),
                ],
            })
        );

        test_from_expr!(
            mql_op_one_arg,
            expected = air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                op: air::MQLOperator::Size,
                args: vec![air::Expression::FieldRef("a".to_string().into())]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::Size,
                args: vec![agg_ast::Expression::Ref(agg_ast::Ref::FieldRef(
                    "a".to_string()
                ))],
            })
        );

        test_from_expr!(
            mql_op_multiple_args,
            expected = air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                op: air::MQLOperator::Lte,
                args: vec![
                    air::Expression::FieldRef("a".to_string().into()),
                    air::Expression::FieldRef("b".to_string().into()),
                ]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::Lte,
                args: vec![
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("b".to_string())),
                ],
            })
        );

        test_from_expr!(
            dollar_literal_doc_becomes_document,
            expected = air::Expression::Document(UniqueLinkedHashMap::new()),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::Literal,
                args: vec![agg_ast::Expression::Document(LinkedHashMap::new())],
            })
        );

        test_from_expr!(
            dollar_literal_literal_becomes_literal,
            expected = air::Expression::Literal(air::LiteralValue::Integer(5)),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::Literal,
                args: vec![agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(
                    5
                ))],
            })
        );

        test_from_expr!(
            dollar_literal_string_becomes_literal,
            expected = air::Expression::Literal(air::LiteralValue::String("a".to_string())),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::Literal,
                args: vec![agg_ast::Expression::Literal(agg_ast::LiteralValue::String(
                    "a".to_string()
                ))],
            })
        );

        test_from_expr!(
            dollar_sqlis_becomes_is,
            expected = air::Expression::Is(air::Is {
                expr: Box::new(air::Expression::FieldRef("a".to_string().into())),
                target_type: air::TypeOrMissing::Type(air::Type::Int32),
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::SQLIs,
                args: vec![
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    agg_ast::Expression::Literal(agg_ast::LiteralValue::String("int".to_string())),
                ],
            })
        );

        test_from_expr!(
            dollar_is_becomes_is,
            expected = air::Expression::Is(air::Is {
                expr: Box::new(air::Expression::FieldRef("a".to_string().into())),
                target_type: air::TypeOrMissing::Missing,
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::SQLIs,
                args: vec![
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    agg_ast::Expression::Literal(agg_ast::LiteralValue::String(
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
                op: agg_ast::UntaggedOperatorName::NullIf,
                args: vec![agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(
                    1
                ))]
            })
        );

        test_from_expr!(
            coalesce,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Coalesce,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(1))]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::Coalesce,
                args: vec![agg_ast::Expression::Literal(agg_ast::LiteralValue::Int32(
                    1
                ))]
            })
        );

        test_from_expr!(
            sql_between,
            expected = air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                op: air::SQLOperator::Between,
                args: vec![
                    air::Expression::FieldRef("a".to_string().into()),
                    air::Expression::FieldRef("b".to_string().into()),
                    air::Expression::FieldRef("c".to_string().into()),
                ]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::SQLBetween,
                args: vec![
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("b".to_string())),
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("c".to_string())),
                ],
            })
        );

        test_from_expr!(
            mql_between,
            expected = air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                op: air::MQLOperator::Between,
                args: vec![
                    air::Expression::FieldRef("a".to_string().into()),
                    air::Expression::FieldRef("b".to_string().into()),
                    air::Expression::FieldRef("c".to_string().into()),
                ]
            }),
            input = agg_ast::Expression::UntaggedOperator(agg_ast::UntaggedOperator {
                op: agg_ast::UntaggedOperatorName::MQLBetween,
                args: vec![
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("a".to_string())),
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("b".to_string())),
                    agg_ast::Expression::Ref(agg_ast::Ref::FieldRef("c".to_string())),
                ],
            })
        );
    }
}
