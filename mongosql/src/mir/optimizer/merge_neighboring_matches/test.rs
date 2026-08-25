mod merge_neighboring_matches_tests {
    use crate::util::mir_field_path;
    use crate::{
        map,
        mir::{
            binding_tuple::DatasourceName::Bottom, schema::SchemaCache, Expression::*,
            LiteralValue::*,
        },
        set, unchecked_unique_linked_hash_map, util,
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

    macro_rules! test_merge_neighboring_matches_no_op {
        ($func_name:ident, input = $input:expr,) => {
            test_merge_neighboring_matches!($func_name, expected = $input, input = $input,);
        };
    }

    test_merge_neighboring_matches_no_op!(
        simple_filter_no_merge,
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
                force_mql_semantics: false,
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
        two_match_filters_get_merged,
        expected = Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
            source: Box::new(Stage::Sentinel),
            condition: MatchQuery::Logical(MatchLanguageLogical {
                op: MatchLanguageLogicalOp::And,
                args: vec![
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("foo", vec!["x"])),
                        arg: LiteralValue::Integer(42),
                        cache: Default::default(),
                    }),
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("bar", vec!["y"])),
                        arg: LiteralValue::Integer(43),
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        input = Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
            source: Box::new(Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(
                MatchFilter {
                    source: Box::new(Stage::Sentinel),
                    condition: MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("foo", vec!["x"])),
                        arg: LiteralValue::Integer(42),
                        cache: Default::default(),
                    }),
                    cache: Default::default(),
                }
            )))),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["y"])),
                arg: LiteralValue::Integer(43),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
    );

    test_merge_neighboring_matches_no_op!(
        two_non_adjacent_match_filters_not_merged,
        input = Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(
                    MatchFilter {
                        source: Box::new(Stage::Sentinel),
                        condition: MatchQuery::Comparison(MatchLanguageComparison {
                            function: MatchLanguageComparisonOp::Eq,
                            input: Some(mir_field_path("foo", vec!["x"])),
                            arg: LiteralValue::Integer(1),
                            cache: Default::default(),
                        }),
                        cache: SchemaCache::new(),
                    }
                )))),
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
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["y"])),
                arg: LiteralValue::Integer(3),
                cache: Default::default(),
            }),
            cache: SchemaCache::new(),
        }))),
    );

    // Two Match Filters: Filter(A), Filter(B) - Project(C) followed by a project, followed by two more match filters
    // Filter(C), Filter(D)
    // Expect it to become Filter(A,B), Project(C), Filter(C,D)
    test_merge_neighboring_matches!(
        nested_match_filters_merged_through_project,
        expected = Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(
                    MatchFilter {
                        source: Box::new(Stage::Sentinel),
                        condition: MatchQuery::Logical(MatchLanguageLogical {
                            op: MatchLanguageLogicalOp::And,
                            args: vec![
                                MatchQuery::Comparison(MatchLanguageComparison {
                                    function: MatchLanguageComparisonOp::Eq,
                                    input: Some(mir_field_path("foo", vec!["a"])),
                                    arg: LiteralValue::Integer(1),
                                    cache: Default::default(),
                                }),
                                MatchQuery::Comparison(MatchLanguageComparison {
                                    function: MatchLanguageComparisonOp::Eq,
                                    input: Some(mir_field_path("foo", vec!["b"])),
                                    arg: LiteralValue::Integer(2),
                                    cache: Default::default(),
                                }),
                            ],
                            cache: SchemaCache::new(),
                        }),
                        cache: SchemaCache::new(),
                    }
                )))),
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
            condition: MatchQuery::Logical(MatchLanguageLogical {
                op: MatchLanguageLogicalOp::And,
                args: vec![
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("bar", vec!["c"])),
                        arg: LiteralValue::Integer(3),
                        cache: Default::default(),
                    }),
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("bar", vec!["d"])),
                        arg: LiteralValue::Integer(4),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        input = Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
            source: Box::new(Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(
                MatchFilter {
                    source: Box::new(Stage::Project(Project {
                        is_add_fields: false,
                        source: Box::new(Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(
                            MatchFilter {
                                source: Box::new(Stage::MqlIntrinsic(MqlStage::MatchFilter(
                                    Box::new(MatchFilter {
                                        source: Box::new(Stage::Sentinel),
                                        condition: MatchQuery::Comparison(
                                            MatchLanguageComparison {
                                                function: MatchLanguageComparisonOp::Eq,
                                                input: Some(mir_field_path("foo", vec!["a"])),
                                                arg: LiteralValue::Integer(1),
                                                cache: Default::default(),
                                            }
                                        ),
                                        cache: Default::default(),
                                    })
                                ))),
                                condition: MatchQuery::Comparison(MatchLanguageComparison {
                                    function: MatchLanguageComparisonOp::Eq,
                                    input: Some(mir_field_path("foo", vec!["b"])),
                                    arg: LiteralValue::Integer(2),
                                    cache: Default::default(),
                                }),
                                cache: Default::default(),
                            }
                        )))),
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
                    condition: MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("bar", vec!["c"])),
                        arg: LiteralValue::Integer(3),
                        cache: Default::default(),
                    }),
                    cache: Default::default(),
                }
            )))),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["d"])),
                arg: LiteralValue::Integer(4),
                cache: Default::default(),
            }),
            cache: Default::default(),
        }))),
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
                force_mql_semantics: false,
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
                force_mql_semantics: false,
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
                    args: vec![Literal(Integer(1)), Literal(Integer(2))],
                    force_mql_semantics: false,
                    is_nullable: false
                }),
                cache: SchemaCache::new(),
            })),
            condition: Literal(Integer(3)),
            cache: SchemaCache::new(),
        }),
    );

    test_merge_neighboring_matches_no_op!(
        non_adjacent_filters_not_merged,
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
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

    test_merge_neighboring_matches_no_op!(
        subquery_filter_at_start_not_merged,
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![],
                    alias: "foo".into(),
                    cache: SchemaCache::new()
                })),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![
                        *util::mir_field_access("__bot__", "x", false),
                        Literal(Integer(10)),
                    ],
                    force_mql_semantics: false,
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Subquery(SubqueryExpr {
                output_expr: util::mir_field_access("__bot__", "y", true),
                subquery: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "bar".into(),
                        cache: SchemaCache::new()
                    })),
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            *util::mir_field_access("__bot__", "z", false),
                            Literal(Integer(5)),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }),
                    cache: SchemaCache::new(),
                })),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_merge_neighboring_matches!(
        subquery_filter_not_at_start_gets_merged,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Sort(Sort {
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![],
                    alias: "foo".into(),
                    cache: SchemaCache::new()
                })),
                specs: set![SortSpecification::Asc(util::mir_field_path(
                    "foo",
                    vec!["a"]
                ))],
                cache: SchemaCache::new(),
            })),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Gt,
                        args: vec![
                            *util::mir_field_access("__bot__", "x", false),
                            Literal(Integer(10)),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }),
                    Subquery(SubqueryExpr {
                        output_expr: util::mir_field_access("__bot__", "y", true),
                        subquery: Box::new(Stage::Filter(Filter {
                            source: Box::new(Stage::Array(ArraySource {
                                array: vec![],
                                alias: "bar".into(),
                                cache: SchemaCache::new()
                            })),
                            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                function: ScalarFunction::Eq,
                                args: vec![
                                    *util::mir_field_access("__bot__", "z", false),
                                    Literal(Integer(5)),
                                ],
                                force_mql_semantics: false,
                                is_nullable: false,
                            }),
                            cache: SchemaCache::new(),
                        })),
                        is_nullable: true,
                    }),
                ],
                force_mql_semantics: false,
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Sort(Sort {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "foo".into(),
                        cache: SchemaCache::new()
                    })),
                    specs: set![SortSpecification::Asc(util::mir_field_path(
                        "foo",
                        vec!["a"]
                    ))],
                    cache: SchemaCache::new(),
                })),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![
                        *util::mir_field_access("__bot__", "x", false),
                        Literal(Integer(10)),
                    ],
                    force_mql_semantics: false,
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            condition: Subquery(SubqueryExpr {
                output_expr: util::mir_field_access("__bot__", "y", true),
                subquery: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "bar".into(),
                        cache: SchemaCache::new()
                    })),
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            *util::mir_field_access("__bot__", "z", false),
                            Literal(Integer(5)),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }),
                    cache: SchemaCache::new(),
                })),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );
}
