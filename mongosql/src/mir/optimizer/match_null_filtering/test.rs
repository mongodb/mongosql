use crate::{
    catalog::{Catalog, Namespace},
    map,
    mir::{
        optimizer::{match_null_filtering::MatchNullFilteringOptimizer, Optimizer},
        schema::{CachedSchema, SchemaCache, SchemaCheckingMode, SchemaInferenceState},
        Collection, Expression, FieldAccess, Filter, LiteralExpr, LiteralValue,
        OptimizedMatchExists, Project, ScalarFunction, ScalarFunctionApplication, Stage,
        SubqueryExpr,
    },
    schema::{Atomic, Document, Schema, SchemaEnvironment, INTEGER_OR_NULLISH},
    set, unchecked_unique_linked_hash_map,
};
use lazy_static::lazy_static;

lazy_static! {
    /// A catalog to use for describing the schema of the `input` for a test. It
    /// defines the schema for the default collection source "foo". It defines 2
    /// nullable fields, 2 non-nullable fields, and 2 nullable nested fields. We
    /// can use this same catalog for all tests.
    static ref CATALOG: Catalog = Catalog::new(map! {
        Namespace {db: "db".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "nullable_a".into() => INTEGER_OR_NULLISH.clone(),
                "nullable_b".into() => INTEGER_OR_NULLISH.clone(),
                "non_nullable_a".into() => Schema::Atomic(Atomic::Integer),
                "non_nullable_b".into() => Schema::Atomic(Atomic::Integer),
                "doc_a".into() => Schema::Document(Document {
                    keys: map! {
                        "nested".into() => INTEGER_OR_NULLISH.clone(),
                    },
                    required: set! {},
                    additional_properties: false,
                }),
                "doc_b".into() => Schema::Document(Document {
                    keys: map! {
                        "nested".into() => INTEGER_OR_NULLISH.clone(),
                    },
                    required: set! {},
                    additional_properties: false,
                }),
            },
            required: set! { "non_nullable_a".to_string(), "non_nullable_b".to_string() },
            additional_properties: false,
        })
    });
}

macro_rules! test_match_null_filtering {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            // Update the input schema cache so the optimizer actually has schema info to use
            let state = SchemaInferenceState::new(
                0u16,
                SchemaEnvironment::default(),
                &*CATALOG,
                SchemaCheckingMode::Relaxed,
            );
            let _ = input.schema(&state);

            // Create actual optimized stage and assert it matches expected
            let optimizer = &MatchNullFilteringOptimizer;
            let actual = optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!(expected, actual);
        }
    };
}

fn collection_source() -> Box<Stage> {
    Box::new(Stage::Collection(Collection {
        db: "db".to_string(),
        collection: "foo".to_string(),
        cache: SchemaCache::new(),
    }))
}

fn field_access_expr(field: Vec<&str>, ref_scope: u16) -> Expression {
    field.into_iter().fold(
        Expression::Reference(("foo", ref_scope).into()),
        |acc, field_part| {
            Expression::FieldAccess(FieldAccess {
                expr: Box::new(acc),
                field: field_part.to_string(),
                cache: SchemaCache::new(),
            })
        },
    )
}

fn optimized_match_exists_expr(field: Vec<&str>, ref_scope: u16) -> Expression {
    let field_access = match field_access_expr(field, ref_scope) {
        Expression::FieldAccess(fa) => fa,
        _ => unreachable!(),
    };
    Expression::OptimizedMatchExists(OptimizedMatchExists {
        field_access,
        cache: SchemaCache::new(),
    })
}

mod all_fields_always_nullable {
    use super::*;

    test_match_null_filtering!(
        ignore_literals,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: collection_source(),
                condition: optimized_match_exists_expr(vec!["nullable_a"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["nullable_a"], 0u16),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["nullable_a"], 0u16),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        multiple_field_refs,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: collection_source(),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        optimized_match_exists_expr(vec!["nullable_a"], 0u16),
                        optimized_match_exists_expr(vec!["nullable_b"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["nullable_a"], 0u16),
                    field_access_expr(vec!["nullable_b"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["nullable_a"], 0u16),
                    field_access_expr(vec!["nullable_b"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        fields_extracted_from_nested_ops,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: collection_source(),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        optimized_match_exists_expr(vec!["nullable_a"], 0u16),
                        optimized_match_exists_expr(vec!["nullable_b"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr(vec!["nullable_a"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr(vec!["nullable_b"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr(vec!["nullable_a"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr(vec!["nullable_b"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        duplicate_field_refs_not_filtered_multiple_times,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: collection_source(),
                condition: optimized_match_exists_expr(vec!["nullable_a"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Gt,
                        args: vec![
                            field_access_expr(vec!["nullable_a"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            field_access_expr(vec!["nullable_a"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(100),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Gt,
                        args: vec![
                            field_access_expr(vec!["nullable_a"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(1),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            field_access_expr(vec!["nullable_a"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(100),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        nested_field_refs_with_same_name_but_different_parents_both_filtered,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: collection_source(),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        optimized_match_exists_expr(vec!["doc_a", "nested"], 0u16),
                        optimized_match_exists_expr(vec!["doc_b", "nested"], 0u16),
                    ],
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["doc_a", "nested"], 0u16),
                    field_access_expr(vec!["doc_b", "nested"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["doc_a", "nested"], 0u16),
                    field_access_expr(vec!["doc_b", "nested"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        nested_filter_stage,
        expected = Stage::Project(Project {
            source: collection_source(),
            expression: map! {
                ("foo", 0u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "sub".to_string() => Expression::Subquery(SubqueryExpr {
                        output_expr: Box::new(field_access_expr(vec!["nullable_a"], 1u16)),
                        subquery: Box::new(Stage::Filter(Filter {
                            source: Box::new(Stage::Filter(Filter {
                                source: collection_source(),
                                condition: optimized_match_exists_expr(vec!["nullable_a"], 1u16),
                                cache: SchemaCache::new(),
                            })),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr(vec!["nullable_a"], 1u16),
                                    Expression::Literal(LiteralExpr {
                                        value: LiteralValue::Integer(1),
                                        cache: SchemaCache::new(),
                                    }),
                                ],
                                cache: SchemaCache::new(),
                            }),
                            cache: SchemaCache::new(),
                        })),
                        cache: SchemaCache::new(),
                    })
                }.into()),
            },
            cache: SchemaCache::new(),
        }),
        input = Stage::Project(Project {
            source: collection_source(),
            expression: map! {
                ("foo", 0u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "sub".to_string() => Expression::Subquery(SubqueryExpr {
                        output_expr: Box::new(field_access_expr(vec!["nullable_a"], 1u16)),
                        subquery: Box::new(Stage::Filter(Filter {
                            source:  collection_source(),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    field_access_expr(vec!["nullable_a"], 1u16),
                                    Expression::Literal(LiteralExpr {
                                        value: LiteralValue::Integer(1),
                                        cache: SchemaCache::new(),
                                    }),
                                ],
                                cache: SchemaCache::new(),
                            }),
                            cache: SchemaCache::new(),
                        })),
                        cache: SchemaCache::new(),
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
                        source: collection_source(),
                        condition: optimized_match_exists_expr(vec!["nullable_b"], 0u16),
                        cache: SchemaCache::new(),
                    })),
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            field_access_expr(vec!["nullable_b"], 0u16),
                            Expression::Literal(LiteralExpr {
                                value: LiteralValue::Integer(100),
                                cache: SchemaCache::new(),
                            }),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    cache: SchemaCache::new(),
                })),
                condition: optimized_match_exists_expr(vec!["nullable_a"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["nullable_a"], 0u16),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: collection_source(),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Eq,
                    args: vec![
                        field_access_expr(vec!["nullable_b"], 0u16),
                        Expression::Literal(LiteralExpr {
                            value: LiteralValue::Integer(100),
                            cache: SchemaCache::new(),
                        }),
                    ],
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["nullable_a"], 0u16),
                    Expression::Literal(LiteralExpr {
                        value: LiteralValue::Integer(1),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
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
                source: collection_source(),
                condition: optimized_match_exists_expr(vec!["nullable_b"], 0u16),
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["non_nullable_a"], 0u16),
                    field_access_expr(vec!["nullable_b"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["non_nullable_a"], 0u16),
                    field_access_expr(vec!["nullable_b"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );

    test_match_null_filtering!(
        no_nullable_fields_does_not_create_filter,
        expected = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["non_nullable_a"], 0u16),
                    field_access_expr(vec!["non_nullable_b"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: collection_source(),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    field_access_expr(vec!["non_nullable_a"], 0u16),
                    field_access_expr(vec!["non_nullable_b"], 0u16),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
    );
}
