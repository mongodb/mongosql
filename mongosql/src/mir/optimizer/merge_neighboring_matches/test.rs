mod merge_neighboring_matches_tests {
    use crate::{
        map,
        mir::{
            binding_tuple::DatasourceName::Bottom, schema::SchemaCache, Expression::*,
            LiteralValue::*,
        },
        unchecked_unique_linked_hash_map,
    };

    macro_rules! test_merge_neighboring_matches {
        ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
            #[test]
            fn $func_name() {
                use crate::mir::{
                    optimizer::merge_neighboring_matches::MergeNeighboringMatchesOptimizer, *,
                };
                let input = $input;
                let expected = $expected;
                let (actual, _) =
                    MergeNeighboringMatchesOptimizer::merge_neighboring_matches(input);
                assert_eq!(expected, actual);
            }
        };
    }

    test_merge_neighboring_matches!(
        simple_filter_no_merge,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: Literal(Integer(42)),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: Literal(Integer(42)),
            cache: SchemaCache::new(),
        }),
    );

    test_merge_neighboring_matches!(
        two_filters_get_merged,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![Literal(Integer(1)), Literal(Integer(2)),],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![],
                    alias: "foo".into(),
                    cache: SchemaCache::new()
                })),
                condition: Literal(Integer(1)),
                cache: SchemaCache::new(),
            })),
            condition: Literal(Integer(2)),
            cache: SchemaCache::new(),
        }),
    );

    test_merge_neighboring_matches!(
        three_filters_merged_into_one_and,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Literal(Integer(1)),
                    Literal(Integer(2)),
                    Literal(Integer(3)),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "foo".into(),
                        cache: SchemaCache::new()
                    })),
                    condition: Literal(Integer(1)),
                    cache: SchemaCache::new(),
                })),
                condition: Literal(Integer(2)),
                cache: SchemaCache::new(),
            })),
            condition: Literal(Integer(3)),
            cache: SchemaCache::new(),
        }),
    );
    test_merge_neighboring_matches!(
        filter_appended_to_existing_and,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Literal(Integer(1)),
                    Literal(Integer(2)),
                    Literal(Integer(3)),
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![],
                    alias: "foo".into(),
                    cache: SchemaCache::new()
                })),
                condition: ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![Literal(Integer(1)), Literal(Integer(2)),],
                    is_nullable: false
                }),
                cache: SchemaCache::new(),
            })),
            condition: Literal(Integer(3)),
            cache: SchemaCache::new(),
        }),
    );

    test_merge_neighboring_matches!(
        non_adjacent_filters_not_merged,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "foo".into(),
                        cache: SchemaCache::new()
                    })),
                    condition: Literal(Integer(1)),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "c".to_string() =>
                            Expression::Literal(LiteralValue::Integer(1),),
                        },
                    }),
                },
                cache: SchemaCache::new(),
            })),
            condition: Literal(Integer(3)),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "foo".into(),
                        cache: SchemaCache::new()
                    })),
                    condition: Literal(Integer(1)),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 0u16).into() => Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "c".to_string() =>
                            Expression::Literal(LiteralValue::Integer(1),),
                        },
                    }),
                },
                cache: SchemaCache::new(),
            })),
            condition: Literal(Integer(3)),
            cache: SchemaCache::new(),
        }),
    );
}
