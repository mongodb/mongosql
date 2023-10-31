use crate::{
    catalog::Catalog,
    map,
    mir::{
        optimizer::{match_null_filtering::MatchNullFilteringOptimizer, Optimizer},
        schema::{SchemaCache, SchemaCheckingMode, SchemaInferenceState},
        Derived, ExistsExpr, Expression, FieldAccess, FieldExistence, Filter, LiteralExpr,
        LiteralValue, MQLExpression, Project, ScalarFunction, ScalarFunctionApplication, Stage,
        SubqueryExpr,
    },
    unchecked_unique_linked_hash_map,
    util::mir_collection,
};

macro_rules! test_match_null_filtering {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let catalog = Catalog::new(map! {});
            // Update the input schema cache so the optimizer actually has schema info to use
            let state = SchemaInferenceState::new(
                0u16,
                crate::schema::SchemaEnvironment::default(),
                &catalog,
                SchemaCheckingMode::Relaxed,
            );

            // Create actual optimized stage and assert it matches expected
            let optimizer = &MatchNullFilteringOptimizer;
            let actual = optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!(expected, actual);
        }
    };
}

fn field_access_expr(
    collection: &str,
    field: Vec<&str>,
    ref_scope: u16,
    is_nullable: bool,
) -> Expression {
    field.into_iter().fold(
        Expression::Reference((collection, ref_scope).into()),
        |acc, field_part| {
            Expression::FieldAccess(FieldAccess {
                expr: Box::new(acc),
                field: field_part.to_string(),
                cache: SchemaCache::new(),
                is_nullable,
            })
        },
    )
}

fn field_existence_expr(collection: &str, field: Vec<&str>, ref_scope: u16) -> Expression {
    let field_access = match field_access_expr(collection, field, ref_scope, true) {
        Expression::FieldAccess(fa) => fa,
        _ => unreachable!(),
    };
    Expression::MQLIntrinsic(MQLExpression::FieldExistence(FieldExistence {
        field_access,
        cache: SchemaCache::new(),
    }))
}

mod all_fields_always_nullable {
    use super::*;

    test_match_null_filtering!(
        ignore_literals,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        multiple_field_refs,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        field_existence_expr("foo", vec!["nullable_a"], 0u16),
                        field_existence_expr("foo", vec!["nullable_b"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, false),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, true),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        multiple_field_refs_with_nested_operators_should_all_flip_no_null,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        field_existence_expr("foo", vec!["nullable_a"], 0u16),
                        field_existence_expr("foo", vec!["nullable_b"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, false),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, false),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, true),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, true),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    })
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        fields_extracted_from_nested_ops,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        field_existence_expr("foo", vec!["nullable_a"], 0u16),
                        field_existence_expr("foo", vec!["nullable_b"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, false),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, true),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        duplicate_field_refs_not_filtered_multiple_times,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Gt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(100),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Gt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(100),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        nested_field_refs_with_same_name_but_different_parents_both_filtered,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        field_existence_expr("foo", vec!["doc_a", "nested"], 0u16),
                        field_existence_expr("foo", vec!["doc_b", "nested"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["doc_a", "nested"], 0u16, false),
                    field_access_expr("foo", vec!["doc_b", "nested"], 0u16, false),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["doc_a", "nested"], 0u16, true),
                    field_access_expr("foo", vec!["doc_b", "nested"], 0u16, true),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        nested_filter_stage,
        expected = Stage::Project(Project {
            source: mir_collection("db", "foo"),
            expression: map! {
                ("foo", 0u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "sub".to_string() => Expression::Subquery(SubqueryExpr {
                        output_expr: Box::new(field_access_expr("foo", vec!["nullable_a"], 1u16, true)),
                        subquery: Box::new(Stage::Filter(Filter {
                            source: Box::new(Stage::Filter(Filter {
                                source: mir_collection("db", "foo"),
                                condition: field_existence_expr("foo", vec!["nullable_a"], 1u16),
                                cache: SchemaCache::new(),
                            })),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr("foo", vec!["nullable_a"], 1u16, false),
                                    Expression::Literal(LiteralExpr {
                                        value: LiteralValue::Integer(1),
                                        cache: SchemaCache::new(),
                                    }),
                                ],
                                cache: SchemaCache::new(),
                                is_nullable: false,
                            }),
                            cache: SchemaCache::new(),
                        })),
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    })
                }.into()),
            },
            cache: SchemaCache::new(),
        }),
        input = Stage::Project(Project {
            source: mir_collection("db", "foo"),
            expression: map! {
                ("foo", 0u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "sub".to_string() => Expression::Subquery(SubqueryExpr {
                        output_expr: Box::new(field_access_expr("foo", vec!["nullable_a"], 1u16, true)),
                        subquery: Box::new(Stage::Filter(Filter {
                            source:  mir_collection("db", "foo"),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr("foo", vec!["nullable_a"], 1u16, true),
                                    Expression::Literal(LiteralExpr {
                                        value: LiteralValue::Integer(1),
                                        cache: SchemaCache::new(),
                                    }),
                                ],
                                cache: SchemaCache::new(),
                        is_nullable: true,
                            }),
                            cache: SchemaCache::new(),
                        })),
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    })
                }.into()),
            },
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        multiple_match_stages_each_get_their_own_filter,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Filter(Filter {
                        source: mir_collection("db", "foo"),
                        condition: field_existence_expr("foo", vec!["nullable_b"], 0u16),
                        cache: SchemaCache::new(),
                    })),
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, false),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(100),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    }),
                    cache: SchemaCache::new(),
                })),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, false),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Eq,
                    args: vec![
                        field_access_expr("foo", vec!["nullable_b"], 0u16, true),
                        Expression::Literal(LiteralExpr {
                            value: LiteralValue::Integer(100),
                            cache: SchemaCache::new(),
                        }),
                    ],
                    cache: SchemaCache::new(),
                    is_nullable: true,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, true),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );
    test_match_null_filtering!(
        derived_stages_impact_scope_level,
        expected = Stage::Derived(Derived {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::Exists(ExistsExpr {
                    stage: Box::new(Stage::Filter(Filter {
                        source: Box::new(Stage::Filter(Filter {
                            source: mir_collection("db", "nested"),
                            condition: field_existence_expr("nested", vec!["nullable_field"], 2u16),
                            cache: SchemaCache::new(),
                        })),
                        condition: Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Eq,
                            args: vec![
                                // Note that the top-level Stage is a Derived, which increases
                                // the scope-level to 1.
                                field_access_expr("foo", vec!["nullable_a"], 1u16, false),
                                field_access_expr("nested", vec!["nullable_field"], 2u16, false),
                            ],
                            cache: SchemaCache::new(),
                            is_nullable: false,
                        }),
                        cache: SchemaCache::new(),
                    })),
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        input = Stage::Derived(Derived {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::Exists(ExistsExpr {
                    stage: Box::new(Stage::Filter(Filter {
                        source: mir_collection("db", "nested"),
                        condition: Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Eq,
                            args: vec![
                                field_access_expr("foo", vec!["nullable_a"], 1u16, true),
                                field_access_expr("nested", vec!["nullable_field"], 2u16, true),
                            ],
                            cache: SchemaCache::new(),
                            is_nullable: true,
                        }),
                        cache: SchemaCache::new(),
                    })),
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })
    );
}

mod mixed_field_nullability {
    use super::*;

    test_match_null_filtering!(
        multiple_field_refs_only_nullable_fields_are_filtered,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: field_existence_expr("foo", vec!["nullable_b"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, false),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, false),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, false),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, true),
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        no_nullable_fields_does_not_create_filter,
        expected = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, false),
                    field_access_expr("foo", vec!["non_nullable_b"], 0u16, false),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, false),
                    field_access_expr("foo", vec!["non_nullable_b"], 0u16, false),
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        })
    );
}
