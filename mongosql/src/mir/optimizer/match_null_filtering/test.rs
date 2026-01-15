use crate::{
    catalog::Catalog,
    map,
    mir::{
        optimizer::{match_null_filtering::MatchNullFilteringOptimizer, Optimizer},
        schema::{SchemaCache, SchemaCheckingMode, SchemaInferenceState},
        Derived, ExistsExpr, Expression, FieldAccess, Filter, LiteralValue, Project,
        ScalarFunction, ScalarFunctionApplication, Stage, SubqueryExpr,
    },
    unchecked_unique_linked_hash_map,
    util::mir_collection,
};

macro_rules! test_match_null_filtering {
    ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr) => {
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
            let (actual, actual_changed) =
                optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!($expected_changed, actual_changed);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_match_no_filtering_no_op {
    ($func_name:ident, $input:expr,) => {
        test_match_null_filtering! { $func_name, expected = $input, expected_changed = false, input = $input }
    };
}

fn field_access_expr(
    collection: &str,
    field: Vec<&str>,
    ref_scope: u16,
    is_nullable: Vec<bool>,
) -> Expression {
    field.into_iter().enumerate().fold(
        Expression::Reference((collection, ref_scope).into()),
        |acc, (idx, field_part)| {
            Expression::FieldAccess(FieldAccess {
                expr: Box::new(acc),
                field: field_part.to_string(),
                is_nullable: *is_nullable.get(idx).unwrap(),
            })
        },
    )
}

fn field_existence_expr(
    collection: &str,
    field: Vec<&str>,
    ref_scope: u16,
    is_nullable: Vec<bool>,
) -> Expression {
    let field_access = match field_access_expr(collection, field, ref_scope, is_nullable) {
        Expression::FieldAccess(fa) => fa,
        _ => unreachable!(),
    };
    Expression::MqlIntrinsicFieldExistence(field_access)
}

mod all_fields_always_nullable {
    use super::*;

    test_match_null_filtering!(
        ignore_non_null_literals,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                    Expression::Literal(LiteralValue::Integer(1)),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                    Expression::Literal(LiteralValue::Integer(1),),
                ]
            )),
            cache: SchemaCache::new(),
        })
    );

    test_match_no_filtering_no_op!(
        expected_to_not_include_null_filters_for_simple_top_level_or,
        Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Null,),
                        ],
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Eq,
                        vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Null,)
                        ]
                    ))
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_match_no_filtering_no_op!(
        expected_to_not_include_null_filter_for_or_predicates_complex_top_level_or,
        Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::String("B".to_string()),),
                        ],
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Gt,
                        vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Integer(10),)
                        ]
                    ))
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_match_no_filtering_no_op!(
        expect_nested_ors_do_not_include_null_filters,
        Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            // 1 = 1 AND (b > 10 OR a = 'B')
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::And,
                vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1),),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ],
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        // Represents the following OR case: b > 10 OR a = 'B'"
                        function: ScalarFunction::Or,
                        args: vec![
                            Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                                    Expression::Literal(LiteralValue::String("B".to_string()),),
                                ],
                                is_nullable: false,
                            }),
                            Expression::ScalarFunction(ScalarFunctionApplication::new(
                                ScalarFunction::Gt,
                                vec![
                                    field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                                    Expression::Literal(LiteralValue::Integer(10),)
                                ]
                            ))
                        ],
                        is_nullable: true,
                    }),
                ]
            )),
            cache: SchemaCache::new(),
        }),
    );

    test_match_null_filtering!(
        null_filters_added_for_fields_not_in_or_expressions,
        // nullable_field1 AND (nullable_field2 OR nullable_field3)
        expected = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Or,
                        args: vec![field_existence_expr(
                            "foo",
                            vec!["nullable_b"],
                            0u16,
                            vec![true]
                        ),],
                        is_nullable: true,
                    })
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            // nullable_field1 AND (nullable_field2 OR nullable_field3)
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::And,
                vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::String("B".to_string()),),
                        ],
                        is_nullable: true,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Or,
                        args: vec![
                            Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                                    Expression::Literal(LiteralValue::String("B".to_string()),),
                                ],
                                is_nullable: true,
                            }),
                            Expression::ScalarFunction(ScalarFunctionApplication::new(
                                ScalarFunction::Gt,
                                vec![
                                    field_access_expr("foo", vec!["nullable_c"], 0u16, vec![true]),
                                    Expression::Literal(LiteralValue::Integer(10),)
                                ]
                            ))
                        ],
                        is_nullable: true,
                    })
                ]
            )),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        do_not_update_is_nullable_for_exprs_with_literal_null_args,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                    Expression::Literal(LiteralValue::Null),
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                    Expression::Literal(LiteralValue::Null,),
                ]
            )),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        intrinsic_null_args_do_not_impact_nullable_updating_of_siblings,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(LiteralValue::Null,),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Integer(3)),
                        ],
                        is_nullable: false,
                    }),
                ],
                is_nullable: true
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    Expression::Literal(LiteralValue::Null,),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Lt,
                        vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Integer(3)),
                        ],
                    )),
                ],
            )),
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
                        field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                        field_existence_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                    ],
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                ]
            )),
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
                        field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                        field_existence_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                    ],
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
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                        ],
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                        ],
                        is_nullable: false,
                    })
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Eq,
                        vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                        ]
                    )),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Eq,
                        vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                        ]
                    ))
                ],
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
                        field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                        field_existence_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                    ],
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
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ],
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ],
                        is_nullable: false,
                    }),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Eq,
                        vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ]
                    )),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Eq,
                        vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ]
                    )),
                ],
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
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Gt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ],
                        is_nullable: false,
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Integer(100),),
                        ],
                        is_nullable: false,
                    }),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Gt,
                        vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Integer(1),),
                        ]
                    )),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Lt,
                        vec![
                            field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                            Expression::Literal(LiteralValue::Integer(100),),
                        ]
                    )),
                ],
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
                        field_existence_expr(
                            "foo",
                            vec!["doc_a", "nested"],
                            0u16,
                            vec![false, true]
                        ),
                        field_existence_expr(
                            "foo",
                            vec!["doc_b", "nested"],
                            0u16,
                            vec![false, true]
                        ),
                    ],
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["doc_a", "nested"], 0u16, vec![false, false]),
                    field_access_expr("foo", vec!["doc_b", "nested"], 0u16, vec![false, false]),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    field_access_expr("foo", vec!["doc_a", "nested"], 0u16, vec![false, true]),
                    field_access_expr("foo", vec!["doc_b", "nested"], 0u16, vec![false, true]),
                ]
            )),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        nested_filter_stage,
        expected = Stage::Project(Project {
            is_add_fields: false,
            source: mir_collection("db", "foo"),
            expression: map! {
                ("foo", 0u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "sub".to_string() => Expression::Subquery(SubqueryExpr {
                        output_expr: Box::new(field_access_expr("foo", vec!["nullable_a"], 1u16, vec![true])),
                        subquery: Box::new(Stage::Filter(Filter {
                            source: Box::new(Stage::Filter(Filter {
                                source: mir_collection("db", "foo"),
                                condition: field_existence_expr("foo", vec!["nullable_a"], 1u16, vec![true]),
                                cache: SchemaCache::new(),
                            })),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr("foo", vec!["nullable_a"], 1u16, vec![false]),
                                    Expression::Literal(LiteralValue::Integer(1)),
                                ],
                                is_nullable: false,
                            }),
                            cache: SchemaCache::new(),
                        })),
                        is_nullable: true,
                    })
                }.into()),
            },
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Project(Project {
            is_add_fields: false,
            source: mir_collection("db", "foo"),
            expression: map! {
                ("foo", 0u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "sub".to_string() => Expression::Subquery(SubqueryExpr {
                        output_expr: Box::new(field_access_expr("foo", vec!["nullable_a"], 1u16, vec![true])),
                        subquery: Box::new(Stage::Filter(Filter {
                            source:  mir_collection("db", "foo"),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                                ScalarFunction::Eq,
                                vec![
                                    field_access_expr("foo", vec!["nullable_a"], 1u16, vec![true]),
                                    Expression::Literal(LiteralValue::Integer(1),),
                                ])),
                            cache: SchemaCache::new(),
                        })),
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
                        condition: field_existence_expr(
                            "foo",
                            vec!["nullable_b"],
                            0u16,
                            vec![true]
                        ),
                        cache: SchemaCache::new(),
                    })),
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                            Expression::Literal(LiteralValue::Integer(100)),
                        ],
                        is_nullable: false,
                    }),
                    cache: SchemaCache::new(),
                })),
                condition: field_existence_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![false]),
                    Expression::Literal(LiteralValue::Integer(1)),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Eq,
                    vec![
                        field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                        Expression::Literal(LiteralValue::Integer(100),),
                    ]
                )),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    field_access_expr("foo", vec!["nullable_a"], 0u16, vec![true]),
                    Expression::Literal(LiteralValue::Integer(1),),
                ]
            )),
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
                            condition: field_existence_expr(
                                "nested",
                                vec!["nullable_field"],
                                2u16,
                                vec![true]
                            ),
                            cache: SchemaCache::new(),
                        })),
                        condition: Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Eq,
                            args: vec![
                                // Note that the top-level Stage is a Derived, which increases
                                // the scope-level to 1. Then the Exists expr increases it to 2.
                                field_access_expr("foo", vec!["nullable_a"], 1u16, vec![true]),
                                field_access_expr(
                                    "nested",
                                    vec!["nullable_field"],
                                    2u16,
                                    vec![false]
                                ),
                            ],
                            is_nullable: true,
                        }),
                        cache: SchemaCache::new(),
                    })),
                }),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Derived(Derived {
            source: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "foo"),
                condition: Expression::Exists(ExistsExpr {
                    stage: Box::new(Stage::Filter(Filter {
                        source: mir_collection("db", "nested"),
                        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                            ScalarFunction::Eq,
                            vec![
                                field_access_expr("foo", vec!["nullable_a"], 1u16, vec![true]),
                                field_access_expr(
                                    "nested",
                                    vec!["nullable_field"],
                                    2u16,
                                    vec![true]
                                ),
                            ]
                        )),
                        cache: SchemaCache::new(),
                    })),
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
                condition: field_existence_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, vec![false]),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, vec![false]),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Eq,
                vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, vec![false]),
                    field_access_expr("foo", vec!["nullable_b"], 0u16, vec![true]),
                ]
            )),
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
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, vec![false]),
                    field_access_expr("foo", vec!["non_nullable_b"], 0u16, vec![false]),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        expected_changed = false,
        input = Stage::Filter(Filter {
            source: mir_collection("db", "foo"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr("foo", vec!["non_nullable_a"], 0u16, vec![false]),
                    field_access_expr("foo", vec!["non_nullable_b"], 0u16, vec![false]),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        })
    );
}
