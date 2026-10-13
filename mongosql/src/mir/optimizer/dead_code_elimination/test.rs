use crate::{
    SchemaCheckingMode,
    catalog::Catalog,
    map, mir,
    mir::{
        optimizer::{Optimizer, dead_code_elimination::DeadCodeEliminator},
        schema::{SchemaCache, SchemaInferenceState},
        *,
    },
    schema::{Atomic, Document, Satisfaction, Schema, SchemaEnvironment},
    set, unchecked_unique_linked_hash_map,
    util::{mir_collection, mir_field_access, mir_field_access_multi_part},
};
use agg_ast::definitions::Namespace;
use lazy_static::lazy_static;

lazy_static! {
    static ref CATALOG: Catalog = Catalog::new(map! {
        Namespace {database: "db".into(), collection: "bar".into()} => Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Double),
            },
            required: set! {},
            additional_properties: true,
            ..Default::default()
        }),
        Namespace {database: "db".into(), collection: "other".into()} => Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::Integer),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
                "z".to_string() => Schema::Atomic(Atomic::Double),
            },
            required: set! {},
            additional_properties: true,
            ..Default::default()
        }),
        Namespace {database: "db".into(), collection: "nested".into()} => Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                        "c".to_string() => Schema::Atomic(Atomic::Integer),
                        "x".to_string() => Schema::Document(Document {
                            keys: map! {
                                "y".to_string() => Schema::Atomic(Atomic::Integer),
                                "z".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            required: set! {},
                            additional_properties: true,
                            ..Default::default()
                        })
                    },
                    required: set! {},
                    additional_properties: true,
                    ..Default::default()
                }),
            },
            required: set! {},
            additional_properties: true,
            ..Default::default()
        }),
    });
}

macro_rules! test_dead_code_elimination {
    ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let state = SchemaInferenceState::new(
                0u16,
                SchemaEnvironment::default(),
                &*CATALOG,
                map! {},
                SchemaCheckingMode::Relaxed,
            );

            let optimizer = &DeadCodeEliminator;
            let (actual, actual_changed) =
                optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!($expected_changed, actual_changed);
            assert_eq!(expected, actual);
        }
    };
}
macro_rules! test_dead_code_elimination_no_op {
    ($func_name:ident, $input:expr) => {
        test_dead_code_elimination! { $func_name, expected = $input, expected_changed = false, input = $input }
    };
}

mod group {
    use super::*;

    fn aliased_field_access(alias: &str, field_access: Expression) -> OptionallyAliasedExpr {
        OptionallyAliasedExpr::Aliased(AliasedExpr {
            alias: alias.to_string(),
            expr: field_access,
        })
    }

    fn unaliased_field_access(key_name: &str, field_name: &str) -> OptionallyAliasedExpr {
        OptionallyAliasedExpr::Unaliased(*mir_field_access(key_name, field_name, true))
    }

    fn unaliased_field_access_multi_part(
        key_name: &str,
        field_names: Vec<&str>,
    ) -> OptionallyAliasedExpr {
        OptionallyAliasedExpr::Unaliased(*mir_field_access_multi_part(key_name, field_names, true))
    }

    fn aliased_aggregation(alias: &str, arg_field_access: Box<Expression>) -> AliasedAggregation {
        AliasedAggregation {
            alias: alias.to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: arg_field_access,
                arg_is_possibly_doc: Satisfaction::Not,
            }),
        }
    }

    test_dead_code_elimination_no_op!(
        cannot_swap_non_project_source_for_group,
        Stage::Group(Group {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "bar"),
                condition: Expression::Literal(LiteralValue::Boolean(true)),
                cache: SchemaCache::new(),
            })),
            keys: vec![],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination_no_op!(
        cannot_swap_project_source_for_group_if_not_all_references_are_substitutable,
        Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                aliased_field_access("a", *mir_field_access("foo", "a", true)),
                unaliased_field_access("bad", "b"),
            ],
            aggregations: vec![aliased_aggregation(
                "agg",
                Box::new(*mir_field_access("bad", "c", true))
            )],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination_no_op!(
        cannot_swap_project_source_for_group_if_unaliased_key_is_impure,
        Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(
                FieldAccess::new(
                    Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::MergeObjects,
                        args: vec![],
                        is_nullable: false,
                    })),
                    "a".to_string(),
                )
            ))],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources.
    test_dead_code_elimination!(
        swap_simple_project_source_for_group_with_single_unaliased_group_key,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![unaliased_field_access("bar", "a")],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![unaliased_field_access("foo", "a")],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources.
    test_dead_code_elimination!(
        swap_simple_project_source_for_group_with_multiple_unaliased_group_keys_from_same_source,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![
                    unaliased_field_access("bar", "a"),
                    unaliased_field_access("bar", "b"),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                unaliased_field_access("foo", "a"),
                unaliased_field_access("foo", "b"),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources.
    test_dead_code_elimination!(
        swap_simple_project_source_for_group_with_multiple_unaliased_group_keys_from_different_sources,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    unaliased_field_access("bar", "a"),
                    unaliased_field_access("bar", "b"),
                    unaliased_field_access("other", "x"),
                    unaliased_field_access("other", "y"),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                ("foo2", 0u16).into() => Expression::Reference(("other", 0u16).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                    ("foo2", 0u16).into() => Expression::Reference(("other", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                unaliased_field_access("foo", "a"),
                unaliased_field_access("foo", "b"),
                unaliased_field_access("foo2", "x"),
                unaliased_field_access("foo2", "y"),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources. In this case, the Project stage
    // originally output a literal Document expression. After the swap, all unaliased group key
    // datasources are retained as keys in the Project expression, but only fields that are
    // referenced as unaliased group keys are retained in the Document literal for each datasource.
    // This is because the other fields from the Document literal will not be emitted by the Group
    // stage.
    test_dead_code_elimination!(
        swap_project_source_with_doc_literal_for_group_with_single_unaliased_group_key_retain_only_field_that_is_key,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(
                    FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a".to_string() => *mir_field_access("bar", "a", true),
                                "b".to_string() => *mir_field_access("bar", "b", true),
                                "c".to_string() => *mir_field_access("bar", "c", true),
                            },
                        })),
                        "b".to_string(),
                    )
                ))],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "b".to_string() => *mir_field_access("bar", "b", true),
                    },
                }),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("bar", "a", true),
                            "b".to_string() => *mir_field_access("bar", "b", true),
                            "c".to_string() => *mir_field_access("bar", "c", true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![unaliased_field_access("bar", "b")],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources. In this case, the Project stage
    // originally output a literal Document expression. After the swap, all unaliased group key
    // datasources are retained as keys in the Project expression, but only fields that are
    // referenced as unaliased group keys are retained in the Document literal for each datasource.
    // This is because the other fields from the Document literal will not be emitted by the Group
    // stage. This test specifically demonstrates that this works even when the unaliased group key
    // maps to a multipart field path.
    // Equivalent query:
    // SELECT * FROM FLATTEN(nested AS flattened WITH DEPTH => 1) GROUP BY flattened.a_b
    test_dead_code_elimination!(
        swap_project_source_with_doc_literal_for_group_with_single_unaliased_group_key_that_maps_to_multipart_field,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "nested"),
                keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                            "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                            "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                        },
                    })),
                    field: "a_b".to_string(),
                    is_nullable: true,
                }))],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                    },
                })
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "nested"),
                expression: map! {
                    ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                            "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                            "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![unaliased_field_access("flattened", "a_b")],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // Equivalent query:
    // SELECT *
    // FROM FLATTEN(FLATTEN(nested AS flattened WITH DEPTH => 1) WITH DEPTH => 1)
    // GROUP BY flattened.a_x_y
    //
    // This test shows the first step: moving the Group stage above the first Project stage (the
    // outer FLATTEN). The following test picks up from that state and shows how the Group stage
    // can be moved above the second Project stage (the inner FLATTEN).
    test_dead_code_elimination!(
        swap_flatten_like_project_source_with_doc_literal_for_group_with_single_unaliased_group_key_that_maps_to_multipart_field,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: mir_collection("db", "nested"),
                    expression: map! {
                        ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                            },
                        })
                    },
                    cache: SchemaCache::new(),
                })),
                keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(
                    FieldAccess {
                        expr: Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a_b".to_string() => *mir_field_access("flattened", "a_b", true),
                                "a_c".to_string() => *mir_field_access("flattened", "a_c", true),
                                "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                                "a_x_z".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "z"], true),
                            },
                        })),
                        field: "a_x_y".to_string(),
                        is_nullable: true,
                    }
                ))],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                    },
                })
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: mir_collection("db", "nested"),
                    expression: map! {
                        ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                            },
                        })
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a_b".to_string() => *mir_field_access("flattened", "a_b", true),
                            "a_c".to_string() => *mir_field_access("flattened", "a_c", true),
                            "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                            "a_x_z".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "z"], true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![unaliased_field_access("flattened", "a_x_y")],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // Equivalent query:
    // SELECT *
    // FROM FLATTEN(FLATTEN(nested AS flattened WITH DEPTH => 1) WITH DEPTH => 1)
    // GROUP BY flattened.a_x_y
    //
    // This test shows the second step: moving the Group stage above the second Project stage (the
    // inner FLATTEN). The previous test shows how the Group stage initially moved above the outer
    // FLATTEN.
    test_dead_code_elimination!(
        swap_flatten_like_project_source_with_doc_literal_for_group_with_single_unaliased_multipart_group_key,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Group(Group {
                    source: mir_collection("db", "nested"),
                    keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Document(DocumentExpr {
                                document: unchecked_unique_linked_hash_map! {
                                    "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                    "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                    "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                                },
                            })),
                            field: "a_x".to_string(),
                            is_nullable: true,
                        })),
                        field: "y".to_string(),
                        is_nullable: true,
                    }))],
                    aggregations: vec![],
                    scope: 0u16,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                    },
                })
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: mir_collection("db", "nested"),
                    expression: map! {
                        ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                            },
                        })
                    },
                    cache: SchemaCache::new(),
                })),
                keys: vec![unaliased_field_access_multi_part(
                    "flattened",
                    vec!["a_x", "y"]
                )],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                    },
                })
            },
            cache: SchemaCache::new(),
        })
    );

    // Equivalent query:
    // SELECT *
    // FROM FLATTEN(FLATTEN(nested AS flattened WITH DEPTH => 1) WITH DEPTH => 1)
    // GROUP BY flattened.a_x_y, flattened.a_x_z
    //
    // This test shows the second step: moving the Group stage above the second Project stage (the
    // inner FLATTEN). It assumes the first step succeeds as normal since the more interesting
    // aspect to test is this second move.
    test_dead_code_elimination!(
        swap_flatten_like_project_source_with_doc_literal_for_group_with_multiple_unaliased_multipart_group_keys,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Group(Group {
                    source: mir_collection("db", "nested"),
                    keys: vec![
                        OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Document(DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                        "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                        "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                                    },
                                })),
                                field: "a_x".to_string(),
                                is_nullable: true,
                            })),
                            field: "y".to_string(),
                            is_nullable: true,
                        })),
                        OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Document(DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                        "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                        "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                                    },
                                })),
                                field: "a_x".to_string(),
                                is_nullable: true,
                            })),
                            field: "z".to_string(),
                            is_nullable: true,
                        })),
                    ],
                    aggregations: vec![],
                    scope: 0u16,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                        "a_x_z".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "z"], true),
                    },
                })
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: mir_collection("db", "nested"),
                    expression: map! {
                        ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a_b".to_string() => *mir_field_access_multi_part("nested", vec!["a", "b"], true),
                                "a_c".to_string() => *mir_field_access_multi_part("nested", vec!["a", "c"], true),
                                "a_x".to_string() => *mir_field_access_multi_part("nested", vec!["a", "x"], true),
                            },
                        })
                    },
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    unaliased_field_access_multi_part("flattened", vec!["a_x", "y"]),
                    unaliased_field_access_multi_part("flattened", vec!["a_x", "z"]),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("flattened", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_x_y".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "y"], true),
                        "a_x_z".to_string() => *mir_field_access_multi_part("flattened", vec!["a_x", "z"], true),
                    },
                })
            },
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources. In this case, the Project stage
    // originally output a literal Document expression. After the swap, all unaliased group key
    // datasources are retained as keys in the Project expression, but only fields that are
    // referenced as unaliased group keys are retained in the Document literal for each datasource.
    // This is because the other fields from the Document literal will not be emitted by the Group
    // stage.
    test_dead_code_elimination!(
        swap_project_source_with_doc_literal_for_group_with_multiple_unaliased_group_keys_from_same_source_retain_all_fields_that_are_keys,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a".to_string() => *mir_field_access("bar", "a", true),
                                "b".to_string() => *mir_field_access("bar", "b", true),
                                "c".to_string() => *mir_field_access("bar", "c", true),
                            },
                        })),
                        "b".to_string(),
                    ))),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a".to_string() => *mir_field_access("bar", "a", true),
                                "b".to_string() => *mir_field_access("bar", "b", true),
                                "c".to_string() => *mir_field_access("bar", "c", true),
                            },
                        })),
                        "c".to_string(),
                    ))),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "b".to_string() => *mir_field_access("bar", "b", true),
                        "c".to_string() => *mir_field_access("bar", "c", true),
                    },
                }),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("bar", "a", true),
                            "b".to_string() => *mir_field_access("bar", "b", true),
                            "c".to_string() => *mir_field_access("bar", "c", true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                unaliased_field_access("bar", "b"),
                unaliased_field_access("bar", "c"),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An unaliased group key is emitted under key's datasource name, therefore when the Project is
    // swapped it retains all unaliased group key datasources. In this case, the Project stage
    // originally output literal Document expressions. After the swap, all unaliased group key
    // datasources are retained as keys in the Project expression, but only fields that are
    // referenced as unaliased group keys are retained in the Document literal for each datasource.
    // This is because the other fields from the Document literals will not be emitted by the Group
    // stage.
    test_dead_code_elimination!(
        swap_project_source_with_doc_literal_for_group_with_multiple_unaliased_group_keys_from_different_sources_retain_all_fields_that_are_keys,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a".to_string() => *mir_field_access("bar", "a", true),
                                "b".to_string() => *mir_field_access("bar", "b", true),
                                "c".to_string() => *mir_field_access("bar", "c", true),
                            },
                        })),
                        "b".to_string(),
                    ))),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a".to_string() => *mir_field_access("bar", "a", true),
                                "b".to_string() => *mir_field_access("bar", "b", true),
                                "c".to_string() => *mir_field_access("bar", "c", true),
                            },
                        })),
                        "c".to_string(),
                    ))),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => *mir_field_access("other", "x", true),
                                "y".to_string() => *mir_field_access("other", "y", true),
                                "z".to_string() => *mir_field_access("other", "z", true),
                            },
                        })),
                        "x".to_string(),
                    ))),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => *mir_field_access("other", "x", true),
                                "y".to_string() => *mir_field_access("other", "y", true),
                                "z".to_string() => *mir_field_access("other", "z", true),
                            },
                        })),
                        "z".to_string(),
                    ))),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "b".to_string() => *mir_field_access("bar", "b", true),
                        "c".to_string() => *mir_field_access("bar", "c", true),
                    },
                }),
                ("other", 0u16).into() => Expression::Document(DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "x".to_string() => *mir_field_access("other", "x", true),
                        "z".to_string() => *mir_field_access("other", "z", true),
                    },
                }),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("bar", "a", true),
                            "b".to_string() => *mir_field_access("bar", "b", true),
                            "c".to_string() => *mir_field_access("bar", "c", true),
                        },
                    }),
                    ("other", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => *mir_field_access("other", "x", true),
                            "y".to_string() => *mir_field_access("other", "y", true),
                            "z".to_string() => *mir_field_access("other", "z", true),
                        },
                    }),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                unaliased_field_access("bar", "b"),
                unaliased_field_access("bar", "c"),
                unaliased_field_access("other", "x"),
                unaliased_field_access("other", "z"),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An aliased group key is emitted under the Bottom datasource, therefore when the Project is
    // swapped it must include the Bottom datasource.
    test_dead_code_elimination!(
        swap_simple_project_source_for_group_with_single_aliased_group_key,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![aliased_field_access(
                    "a",
                    *mir_field_access("bar", "a", true)
                )],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![aliased_field_access(
                "a",
                *mir_field_access("foo", "a", true)
            )],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An aliased group key is emitted under the Bottom datasource, therefore when the Project is
    // swapped it must include the Bottom datasource.
    test_dead_code_elimination!(
        swap_simple_project_source_for_group_with_multiple_aliased_group_keys_from_same_source,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![
                    aliased_field_access("a", *mir_field_access("bar", "a", true)),
                    aliased_field_access("b", *mir_field_access("bar", "b", true)),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                aliased_field_access("a", *mir_field_access("foo", "a", true)),
                aliased_field_access("b", *mir_field_access("foo", "b", true)),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An aliased group key is emitted under the Bottom datasource, therefore when the Project is
    // swapped it must include the Bottom datasource.
    test_dead_code_elimination!(
        swap_simple_project_source_for_group_with_multiple_aliased_group_keys_from_different_sources,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    aliased_field_access("a", *mir_field_access("bar", "a", true)),
                    aliased_field_access("b", *mir_field_access("bar", "b", true)),
                    aliased_field_access("n", *mir_field_access("other", "x", true)),
                    aliased_field_access("m", *mir_field_access("other", "y", true)),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                    ("foo2", 0u16).into() => Expression::Reference(("other", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                aliased_field_access("a", *mir_field_access("foo", "a", true)),
                aliased_field_access("b", *mir_field_access("foo", "b", true)),
                aliased_field_access("n", *mir_field_access("foo2", "x", true)),
                aliased_field_access("m", *mir_field_access("foo2", "y", true)),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An aliased group key is emitted under the Bottom datasource, therefore when the Project is
    // swapped it must include the Bottom datasource. In this case, it does not matter that the
    // original Project outputs a Document literal since the Group stage will only emit bindings
    // under the Bottom datasource.
    test_dead_code_elimination!(
        swap_project_source_with_doc_literal_for_group_with_single_aliased_group_key,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "a".to_string(),
                    expr: Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Document(DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "a".to_string() => *mir_field_access("bar", "a", true),
                                "b".to_string() => *mir_field_access("bar", "b", true),
                                "c".to_string() => *mir_field_access("bar", "c", true),
                            },
                        })),
                        "a".to_string(),
                    ))
                }),],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("bar", "a", true),
                            "b".to_string() => *mir_field_access("bar", "b", true),
                            "c".to_string() => *mir_field_access("bar", "c", true),
                        },
                    })
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![aliased_field_access(
                "a",
                *mir_field_access("bar", "a", true)
            )],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    // An aliased group key is emitted under the Bottom datasource, therefore when the Project is
    // swapped it must include the Bottom datasource. In this case, it does not matter that the
    // original Project outputs Document literals since the Group stage will only emit bindings
    // under the Bottom datasource.
    test_dead_code_elimination!(
        swap_project_source_with_doc_literal_for_group_with_multiple_aliased_group_keys,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "a".to_string(),
                        expr: Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Document(DocumentExpr {
                                document: unchecked_unique_linked_hash_map! {
                                    "a".to_string() => *mir_field_access("bar", "a", true),
                                    "b".to_string() => *mir_field_access("bar", "b", true),
                                    "c".to_string() => *mir_field_access("bar", "c", true),
                                },
                            })),
                            "a".to_string(),
                        ))
                    }),
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "b".to_string(),
                        expr: Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Document(DocumentExpr {
                                document: unchecked_unique_linked_hash_map! {
                                    "a".to_string() => *mir_field_access("bar", "a", true),
                                    "b".to_string() => *mir_field_access("bar", "b", true),
                                    "c".to_string() => *mir_field_access("bar", "c", true),
                                },
                            })),
                            "b".to_string(),
                        ))
                    }),
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "n".to_string(),
                        expr: Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Document(DocumentExpr {
                                document: unchecked_unique_linked_hash_map! {
                                    "x".to_string() => *mir_field_access("other", "x", true),
                                    "y".to_string() => *mir_field_access("other", "y", true),
                                    "z".to_string() => *mir_field_access("other", "z", true),
                                },
                            })),
                            "x".to_string(),
                        ))
                    }),
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "m".to_string(),
                        expr: Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Document(DocumentExpr {
                                document: unchecked_unique_linked_hash_map! {
                                    "x".to_string() => *mir_field_access("other", "x", true),
                                    "y".to_string() => *mir_field_access("other", "y", true),
                                    "z".to_string() => *mir_field_access("other", "z", true),
                                },
                            })),
                            "y".to_string(),
                        ))
                    }),
                ],
                aggregations: vec![],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Inner,
                    left: mir_collection("db", "bar"),
                    right: mir_collection("db", "other"),
                    condition: None,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("bar", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("bar", "a", true),
                            "b".to_string() => *mir_field_access("bar", "b", true),
                            "c".to_string() => *mir_field_access("bar", "c", true),
                        },
                    }),
                    ("other", 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => *mir_field_access("other", "x", true),
                            "y".to_string() => *mir_field_access("other", "y", true),
                            "z".to_string() => *mir_field_access("other", "z", true),
                        },
                    }),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                aliased_field_access("a", *mir_field_access("bar", "a", true)),
                aliased_field_access("b", *mir_field_access("bar", "b", true)),
                aliased_field_access("n", *mir_field_access("other", "x", true)),
                aliased_field_access("m", *mir_field_access("other", "y", true)),
            ],
            aggregations: vec![],
            scope: 0u16,
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination!(
        swap_project_source_for_group_retains_bottom_datasource_when_aggregations_present,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Group(Group {
                    source: mir_collection("db", "bar"),
                    keys: vec![OptionallyAliasedExpr::Unaliased(*mir_field_access(
                        "bar", "b", true
                    )),],
                    aggregations: vec![AliasedAggregation {
                        alias: "agg".to_string(),
                        agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                            function: AggregationFunction::Avg,
                            distinct: false,
                            arg: mir_field_access("bar", "c", true),
                            arg_is_possibly_doc: Satisfaction::Not,
                        }),
                    }],
                    scope: 0u16,
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                    binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Group(Group {
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: mir_collection("db", "bar"),
                    expression: map! {
                        ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                    },
                    cache: SchemaCache::new(),
                })),
                keys: vec![OptionallyAliasedExpr::Unaliased(*mir_field_access(
                    "foo", "b", true
                )),],
                aggregations: vec![AliasedAggregation {
                    alias: "agg".to_string(),
                    agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                        function: AggregationFunction::Avg,
                        distinct: false,
                        arg: mir_field_access("foo", "c", true),
                        arg_is_possibly_doc: Satisfaction::Not,
                    }),
                }],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("bar", 0u16).into()),
                binding_tuple::Key::bot(0) => Expression::Reference(binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        })
    );
}

mod project {
    use super::*;

    test_dead_code_elimination_no_op!(
        cannot_eliminate_non_project_source_for_project,
        Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "bar"),
                condition: Expression::Literal(LiteralValue::Boolean(true),),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination_no_op!(
        cannot_eliminate_non_add_fields_source_for_project,
        Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination_no_op!(
        cannot_eliminate_any_source_for_add_fields,
        Stage::Project(Project {
            is_add_fields: true,
            source: Box::new(Stage::Project(Project {
                is_add_fields: true,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination_no_op!(
        cannot_eliminate_if_project_and_add_fields_define_different_fields,
        Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: true,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo2", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })
    );

    test_dead_code_elimination!(
        remove_add_fields_before_project_that_defines_same_fields,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: mir_collection("db", "bar"),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: true,
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })
    );
}
