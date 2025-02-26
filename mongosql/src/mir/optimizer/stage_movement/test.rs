use crate::{
    catalog::Catalog,
    map,
    mir::{DocumentExpr, SubqueryComparisonOp, SubqueryModifier},
    schema::{self, Schema},
    set,
    util::mir_project_collection,
};
use lazy_static::lazy_static;

lazy_static! {
    static ref CATALOG: Catalog = Catalog::new(map! {
        ("foo", "bar").into() => Schema::Document(
            schema::Document {
                keys:
                    map! {
                        "date0".to_string() => schema::Schema::Atomic(schema::Atomic::Date),
                        "x".to_string() => schema::Schema::Document(
                            schema::Document {
                                keys: map! {
                                    "a".to_string() => schema::Schema::Document(
                                        schema::Document {
                                            keys: map! {
                                                "b".to_string() => schema::Schema::Atomic(schema::Atomic::Double),
                                                "c".to_string() => schema::Schema::Atomic(schema::Atomic::Double),
                                            },
                                            required: set!{"b".to_string(), "c".to_string()},
                                            additional_properties: false,
                                            ..Default::default()
                                            }
                                    )
                                },
                                required: set!{"a".to_string()},
                                additional_properties: false,
                                ..Default::default()
                                }
                        ),
                        "y".to_string() => schema::Schema::Atomic(schema::Atomic::Integer),
                        "z".to_string() => schema::Schema::Atomic(schema::Atomic::Integer),
                    },
                required: set!{"x".to_string()},
                additional_properties: false,
                ..Default::default()
                }
        ),
        ("foo", "bar2").into() => Schema::Document(
            schema::Document {
                keys:
                    map! {
                        "date0".to_string() => schema::Schema::Atomic(schema::Atomic::Date),
                        "x".to_string() => schema::Schema::Document(
                            schema::Document {
                                keys: map! {
                                    "a".to_string() => schema::Schema::Document(
                                        schema::Document {
                                            keys: map! {
                                                "b".to_string() => schema::Schema::Atomic(schema::Atomic::Double),
                                                "c".to_string() => schema::Schema::Atomic(schema::Atomic::Double),
                                            },
                                            required: set!{"b".to_string(), "c".to_string()},
                                            additional_properties: false,
                                            ..Default::default()
                                            }
                                    )
                                },
                                required: set!{"a".to_string()},
                                additional_properties: false,
                                ..Default::default()
                                }
                        ),
                    },
                required: set!{"x".to_string()},
                additional_properties: false,
                ..Default::default()
                }
        )
    });
}

macro_rules! test_move_stage {
    ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                catalog::Catalog,
                map,
                mir::{
                    self,
                    binding_tuple::{BindingTuple, Key},
                    optimizer::stage_movement::StageMovementOptimizer,
                    schema::{SchemaCache, SchemaCheckingMode, SchemaInferenceState},
                    visitor::Visitor,
                    ArraySource, Collection, Derived, EquiJoin,
                    Expression::{self, *},
                    FieldAccess, FieldPath, Filter, Group, Join, JoinType, LateralJoin, Limit,
                    LiteralValue,
                    LiteralValue::*,
                    MQLStage, MatchFilter, MatchLanguageComparison, MatchLanguageComparisonOp,
                    MatchLanguageLogical, MatchLanguageLogicalOp, MatchQuery, Offset, Project,
                    ReferenceExpr, ScalarFunction, ScalarFunctionApplication, Set, SetOperation,
                    Sort, SortSpecification, Stage, SubqueryComparison, SubqueryExpr,
                    TypeAssertionExpr, Unwind,
                },
                schema::SchemaEnvironment,
                set, unchecked_unique_linked_hash_map,
                util::{
                    mir_collection, mir_field_access, mir_field_access_multi_part, mir_field_path,
                },
            };
            #[allow(unused)]
            let input = $input;
            let expected = $expected;
            let (actual, actual_changed) = StageMovementOptimizer::move_stages(
                input,
                &SchemaInferenceState::new(
                    0,
                    SchemaEnvironment::new(),
                    &*CATALOG,
                    SchemaCheckingMode::Relaxed,
                ),
            );
            assert_eq!($expected_changed, actual_changed);
            assert_eq!(expected, actual);
        }
    };
    (ignore = $ignore:expr, $func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr,) => {
        #[ignore = $ignore]
        #[test]
        fn $func_name() {
            println!("I'm an ignored test being run anyway");
            assert!(true);
        }
    };
}

macro_rules! test_move_stage_no_op {
    ($func_name:ident, $input:expr) => {
        test_move_stage! { $func_name, expected = $input, expected_changed = false, input = $input, }
    };
}

test_move_stage_no_op!(
    limit_does_not_move_above_filter,
    Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false,
                }
            ),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    })
);

test_move_stage!(
    move_limit_above_projects,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Project(Project {
            is_add_fields: false,
            source: Stage::Limit(Limit {
                source: mir_collection("foo", "bar"),
                limit: 42,
                cache: SchemaCache::new(),
            })
            .into(),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    }.into()
                )
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        expression: BindingTuple(map! {
            Key::bot(0) => mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(43)),
                }.into()
            )
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Limit(Limit {
        source: Stage::Project(Project {
            is_add_fields: false,
            source: Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    )
                }),
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(43)),
                    }.into()
                )
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 42,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_offsets_above_projects,
    expected = Stage::Project(Project {
                            is_add_fields: false,
        source: Stage::Project(Project {
                            is_add_fields: false,
            source: Stage::Offset(Offset {
                source: Stage::Offset(Offset {
                    source: mir_collection("foo", "bar"),
                    offset: 42,
                    cache: SchemaCache::new(),
                }).into(),
                offset: 44,
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    }.into()
                )
            }),
            cache: SchemaCache::new(),
        }).into(),
        expression: BindingTuple(map! {
            Key::bot(0) => mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(43)),
                }.into()
            )
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Offset(Offset {
        source: Stage::Offset(Offset {
            source: Stage::Project(Project {
                            is_add_fields: false,
                source: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        Key::bot(0) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        )
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(43)),
                        },
                    })
                }),
                cache: SchemaCache::new(),
            })
            .into(),
            offset: 42,
            cache: SchemaCache::new(),
        }).into(),
        offset: 44,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_above_projects,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Project(Project {
            is_add_fields: false,
            source: Stage::Filter(Filter {
                source: mir_collection("foo", "bar"),
                condition: Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Lt,
                        args: vec![
                            mir::Expression::Document(
                                unchecked_unique_linked_hash_map! {
                                    "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                                }.into()
                            ),
                            mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        ],
                        is_nullable: false,
                    }
                ),
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::named("bar", 0u16) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    }.into()
                ),
            }),
            cache: SchemaCache::new(),
        }).into(),
        expression: BindingTuple(map! {
            Key::bot(0) => mir::Expression::Reference(("bar", 0u16).into()),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Project(Project {
                        is_add_fields: false,
            source: Stage::Project(Project {
                        is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("bar", 0u16).into()),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lt,
                args: vec![
                    mir::Expression::Reference(Key::bot(0u16).into()),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                ],
                is_nullable: false,
            }
        ),
        cache: SchemaCache::new(),
    }),
);

test_move_stage_no_op!(
    do_not_move_filter_above_opaque_defines,
    Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                path: mir_field_path("__bot__", vec!["x"]),
                index: None,
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        *mir_field_access("__bot__", "x", true),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: true,
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    })
);

test_move_stage_no_op!(
    do_not_move_sort_above_project_when_nonsubstitutable_complex_expression_is_used,
    Stage::Limit(Limit {
        source: Stage::Sort(Sort {
            source: Stage::Project(Project {
                            is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    Key::bot(0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "y".to_string() => mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                function: mir::ScalarFunction::Add,
                                args: vec![],
                                is_nullable: true,
                            }),
                       }.into()
                    ),
               }),
               cache: SchemaCache::new(),
           }).into(),
           specs: vec![
               SortSpecification::Asc(mir_field_path("__bot__", vec!["y"])),
           ],
           cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    })
);

test_move_stage!(
    move_sort_above_project_when_substitutable_complex_expression_is_used,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["y"]))],
            cache: SchemaCache::new(),
        })
        .into(),
        expression: BindingTuple(map! {
            Key::bot(0u16) => mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "y".to_string() => *mir_field_access("bar", "y", true),
                }.into()
            ),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Sort(Sort {
        source: Stage::Project(Project {
            is_add_fields: false,
            source: mir_collection("foo", "bar"),
            expression: BindingTuple(map! {
                 Key::bot(0u16) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                         "y".to_string() => *mir_field_access("bar", "y", true),
                    }.into()
                ),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        specs: vec![SortSpecification::Asc(mir_field_path("__bot__", vec!["y"]))],
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_limit_above_project_when_substitutable_complex_expression_is_used,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Limit(Limit {
            source: mir_collection("foo", "bar"),
            limit: 10,
            cache: SchemaCache::new(),
        })
        .into(),
        expression: BindingTuple(map! {
            Key::bot(0u16) => mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "y".to_string() => *mir_field_access("bar", "y", true),
                }.into()
            ),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Limit(Limit {
        source: Stage::Project(Project {
            is_add_fields: false,
            source: mir_collection("foo", "bar"),
            expression: BindingTuple(map! {
                 Key::bot(0u16) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                         "y".to_string() => *mir_field_access("bar", "y", true),
                    }.into()
                ),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_match_filter_above_project_when_substitutable_complex_expression_is_used,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["y"])),
                arg: LiteralValue::Integer(43),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))
        .into(),
        expression: BindingTuple(map! {
            Key::bot(0u16) => mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "y".to_string() => *mir_field_access("bar", "y", true),
                }.into()
            ),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Stage::Project(Project {
            is_add_fields: false,
            source: mir_collection("foo", "bar"),
            expression: BindingTuple(map! {
                 Key::bot(0u16) => mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                         "y".to_string() => *mir_field_access("bar", "y", true),
                    }.into()
                ),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("__bot__", vec!["y"])),
            arg: LiteralValue::Integer(43),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    move_sorts_above_project_will_not_reorder_sort_to_sort,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Sort(Sort {
            source: Stage::Sort(Sort {
                source: mir_collection("foo", "bar"),
                specs: vec![SortSpecification::Asc(mir_field_path(
                    "bar",
                    vec!["x", "a"]
                ))],
                cache: SchemaCache::new(),
            })
            .into(),
            specs: vec![SortSpecification::Desc(mir_field_path(
                "bar",
                vec!["x", "b"],
            ))],
            cache: SchemaCache::new(),
        })
        .into(),
        expression: BindingTuple(map! {
            Key::named("bar", 0u16) => *mir_field_access("bar", "x", true),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Sort(Sort {
        source: Stage::Sort(Sort {
            source: Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => *mir_field_access("bar", "x", true),
                }),
                cache: SchemaCache::new(),
            })
            .into(),
            specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["a"],))],
            cache: SchemaCache::new(),
        })
        .into(),
        specs: vec![SortSpecification::Desc(mir_field_path("bar", vec!["b"],))],
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_sorts_above_project_but_not_above_group,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Sort(Sort {
            source: Stage::Group(Group {
                source: mir_collection("foo", "bar"),
                cache: SchemaCache::new(),
                keys: vec![],
                aggregations: vec![],
                scope: 0
            })
            .into(),
            specs: vec![SortSpecification::Desc(mir_field_path(
                "bar",
                vec!["x", "b"],
            ))],
            cache: SchemaCache::new(),
        })
        .into(),
        expression: BindingTuple(map! {
            Key::named("bar", 0u16) => *mir_field_access("bar", "x", true),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Sort(Sort {
        source: Stage::Project(Project {
            is_add_fields: false,
            source: Stage::Group(Group {
                source: mir_collection("foo", "bar"),
                cache: SchemaCache::new(),
                keys: vec![],
                aggregations: vec![],
                scope: 0
            })
            .into(),
            expression: BindingTuple(map! {
                Key::named("bar", 0u16) => *mir_field_access("bar", "x", true),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        specs: vec![SortSpecification::Desc(mir_field_path("bar", vec!["b"],))],
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filters_above_project_will_reorder_filter_to_filter,
    expected = Stage::Project(Project {
            is_add_fields: false,
            source: Stage::Filter(Filter {
            source: Stage::Filter(Filter {
                source: mir_collection("foo", "bar"),
                condition: Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Lt,
                        args: vec![
                            mir::Expression::Document(
                                unchecked_unique_linked_hash_map! {
                                    "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                                }.into()
                            ),
                            mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        ],
                        is_nullable: false,
                }),
                cache: SchemaCache::new(),
            }).into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                        mir::Expression::Literal(mir::LiteralValue::Integer(54)),
                    ],
                    is_nullable: false,
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        expression: BindingTuple(map! {
            Key::named("bar", 0u16) => mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                }.into()
            ),
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Filter(Filter {
            source: Stage::Project(Project {
                        is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
                condition: Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(54)),
                    ],
                    is_nullable: false,
                },
                ),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lt,
                args: vec![
                    mir::Expression::Reference(Key::named("bar", 0u16).into()),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                ],
                is_nullable: false,
            },
        ),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    only_moving_filter_above_filter_does_not_register_as_a_change,
    expected = Stage::Filter(Filter {
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::Literal(LiteralValue::Boolean(false)),
            cache: SchemaCache::new(),
        })),
        condition: Expression::Literal(LiteralValue::Boolean(true)),
        cache: SchemaCache::new(),
    }),
    expected_changed = false,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::Literal(LiteralValue::Boolean(true)),
            cache: SchemaCache::new(),
        })),
        condition: Expression::Literal(LiteralValue::Boolean(false)),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_right_above_join_and_projects,
    expected = Stage::Limit(Limit {
        source: Stage::Join( Join {
            left: Stage::Project(Project {
                            is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                            is_add_fields: false,
                source: Stage::Filter(Filter {
                    source: mir_collection("foo", "bar2"),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(
                                    unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                                    }.into()
                                ),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],
                            is_nullable: false,
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            condition: None,
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Join( Join {
                left: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar2"),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                condition: None,
                join_type: JoinType::Inner,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar2", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false,
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_under_join_into_none_on,
    expected = Stage::Join(Join {
        left: mir_collection("foo", "bar"),
        right: mir_collection("foo", "bar2"),
        condition: Some(Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
            is_nullable: true,
        })),
        join_type: JoinType::Inner,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Join(Join {
            left: mir_collection("foo", "bar"),
            right: mir_collection("foo", "bar2"),
            condition: None,
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication::new(
            mir::ScalarFunction::Lt,
            vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage_no_op!(
    do_not_move_filter_under_left_join_into_on_when_filter_uses_rhs,
    Stage::Filter(Filter {
        source: Stage::Join(Join {
            left: mir_collection("foo", "bar"),
            right: mir_collection("foo", "bar2"),
            condition: None,
            join_type: JoinType::Left,
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication::new(
            mir::ScalarFunction::Lt,
            vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
        )),
        cache: SchemaCache::new(),
    })
);

test_move_stage!(
    move_filter_under_left_join_into_lhs_when_filter_does_not_use_rhs,
    expected = Stage::Join(Join {
        left: Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lt,
                args: vec![
                    mir::Expression::Reference(Key::named("bar", 0u16).into()),
                    mir::Expression::Reference(Key::named("bar", 0u16).into()),
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        right: mir_collection("foo", "bar2"),
        condition: None,
        join_type: JoinType::Left,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Join(Join {
            left: mir_collection("foo", "bar"),
            right: mir_collection("foo", "bar2"),
            condition: None,
            join_type: JoinType::Left,
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication::new(
            mir::ScalarFunction::Lt,
            vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_under_join_into_some_on,
    expected = Stage::Join(Join {
        left: mir_collection("foo", "bar"),
        right: mir_collection("foo", "bar2"),
        condition: Some(Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::And,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Reference(Key::named("bar2", 0u16).into()),
                    ],
                    is_nullable: false,
                })
            ],
            is_nullable: false,
        })),
        join_type: JoinType::Inner,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Join(Join {
            left: mir_collection("foo", "bar"),
            right: mir_collection("foo", "bar2"),
            condition: Some(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
            is_nullable: false,
        }),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_left_above_join_and_projects,
    expected = Stage::Limit(Limit {
        source: Stage::Join( Join {
            left: Stage::Project(Project {
                            is_add_fields: false,
                source: Stage::Filter(Filter {
                    source: mir_collection("foo", "bar"),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::FieldAccess(mir::FieldAccess {
                                    expr: mir::Expression::Document(
                                        unchecked_unique_linked_hash_map! {
                                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                                        }.into()
                                    ).into(),
                                    field: "x".into(),
                                    is_nullable: true,
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],
                            is_nullable: true,
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                            is_add_fields: false,
                source: mir_collection("foo", "bar2"),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()),
                }),
                cache: SchemaCache::new(),
            }).into(),
            condition: None,
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Join( Join {
                left: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar2"),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                condition: None,
                join_type: JoinType::Inner,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        *mir_field_access("bar", "x", true),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: true,
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_right_above_set_and_projects,
    expected = Stage::Limit(Limit {
        source: Stage::Set( Set {
            left: Stage::Project(Project {
                            is_add_fields: false,
                source: mir_collection("foo", "bar"),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                            is_add_fields: false,
                source: Stage::Filter(Filter {
                    source: mir_collection("foo", "bar2"),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(
                                    unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                                    }.into()
                                ),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],
                            is_nullable: false,
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            operation: SetOperation::UnionAll,
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Set( Set {
                left: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar2"),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                operation: SetOperation::UnionAll,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar2", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false
            }),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_left_above_set_and_projects,
    expected = Stage::Limit(Limit {
        source: Stage::Set( Set {
            left: Stage::Project(Project {
                            is_add_fields: false,
                source: Stage::Filter(Filter {
                    source: mir_collection("foo", "bar"),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication::new(
                            mir::ScalarFunction::Lt,
                            vec![
                                mir::Expression::Document(
                                    unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                                    }.into()
                                ),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],
                        )
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                            is_add_fields: false,
                source: mir_collection("foo", "bar2"),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            operation: SetOperation::UnionAll,
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Set(Set {
                left: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar2"),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                operation: SetOperation::UnionAll,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication::new(mir::ScalarFunction::Lt,vec![
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],)
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_into_derived_query,
    expected = Stage::Derived(Derived {
        source: Box::new(Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Filter(Filter {
                    source: mir_collection("foo", "bar"),
                    condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Lt,
                        args: vec![
                            Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("bar", 1u16).into())),
                                field: "y".to_string(),
                                is_nullable: true,
                            }),
                            Expression::Literal(LiteralValue::Integer(100)),
                        ],
                        is_nullable: false,
                    }),
                    cache: SchemaCache::new(),
                })),
                expression: BindingTuple(map! {
                    Key::named("bar", 1u16) => Expression::Reference(("bar", 1u16).into()),
                }),
                cache: SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::named("bar", 0u16) => Expression::Reference(("bar", 1u16).into()),
            }),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::Derived(Derived {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        Key::named("bar", 1u16) => Expression::Reference(("bar", 1u16).into()),
                    }),
                    cache: SchemaCache::new(),
                })),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => Expression::Reference(("bar", 1u16).into()),
                }),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                *mir_field_access("bar", "y", true),
                Expression::Literal(LiteralValue::Integer(100)),
            ],
            is_nullable: false,
        }),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_two_filters_under_join_one_into_none_on_and_other_filter_above_join_because_filters_can_reorder,
    expected = Stage::Join(Join {
        left: Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(mir::ScalarFunctionApplication::new(
                mir::ScalarFunction::Eq,
                vec![
                    mir::Expression::Reference(Key::named("bar", 0u16).into()),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                ],
            )),
            cache: SchemaCache::new(),
        })
        .into(),
        right: mir_collection("foo", "bar2"),
        condition: Some(Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
            is_nullable: true
        })),
        join_type: JoinType::Inner,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Filter(Filter {
            source: Stage::Join(Join {
                left: mir_collection("foo", "bar"),
                right: mir_collection("foo", "bar2"),
                condition: None,
                join_type: JoinType::Inner,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(mir::ScalarFunctionApplication::new(mir::ScalarFunction::Lt,vec![
                    mir::Expression::Reference(Key::named("bar", 0u16).into()),
                    mir::Expression::Reference(Key::named("bar2", 0u16).into()),
                ],)),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication::new(mir::ScalarFunction::Eq,vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            ],)),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_above_equijoin,
    expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
        join_type: JoinType::Inner,
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Gt,
                vec![
                    *mir_field_access("bar", "y", true),
                    Expression::Literal(LiteralValue::Integer(0)),
                ],
            )),
            cache: SchemaCache::new(),
        })),
        from: mir_collection("foo", "bar2"),
        local_field: Box::new(mir_field_path("bar", vec!["date0"])),
        foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            from: mir_collection("foo", "bar2"),
            local_field: Box::new(mir_field_path("bar", vec!["date0"])),
            foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access("bar", "y", true),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage_no_op!(
    cannot_move_filter_above_right_side_of_equijoin,
    Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            from: mir_collection("foo", "bar2"),
            local_field: Box::new(mir_field_path("bar", vec!["date0"])),
            foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Gt,
            args: vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
            is_nullable: true,
        }),
        cache: SchemaCache::new(),
    })
);

test_move_stage!(
    move_match_filter_above_equijoin,
    expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
        join_type: JoinType::Inner,
        source: Box::new(Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["y"])),
                arg: LiteralValue::Integer(43),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        from: mir_collection("foo", "bar2"),
        local_field: Box::new(mir_field_path("bar", vec!["date0"])),
        foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            from: mir_collection("foo", "bar2"),
            local_field: Box::new(mir_field_path("bar", vec!["date0"])),
            foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar", vec!["y"])),
            arg: LiteralValue::Integer(43),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage_no_op!(
    cannot_move_match_filter_above_right_side_of_equijoin,
    Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            from: mir_collection("foo", "bar2"),
            local_field: Box::new(mir_field_path("bar", vec!["date0"])),
            foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar2", vec!["x", "a", "b"],)),
            arg: LiteralValue::Double(2.4),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }))
);

test_move_stage!(
    move_sort_above_equijoin,
    expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
        join_type: JoinType::Inner,
        source: Box::new(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["y"]))],
            cache: SchemaCache::new(),
        })),
        from: mir_collection("foo", "bar2"),
        local_field: Box::new(mir_field_path("bar", vec!["date0"])),
        foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::Sort(Sort {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            from: mir_collection("foo", "bar2"),
            local_field: Box::new(mir_field_path("bar", vec!["date0"])),
            foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
            cache: SchemaCache::new(),
        }))),
        specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["y"]))],
        cache: SchemaCache::new(),
    }),
);

test_move_stage_no_op!(
    cannot_move_sort_above_right_side_of_equijoin,
    Stage::Sort(Sort {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            from: mir_collection("foo", "bar2"),
            local_field: Box::new(mir_field_path("bar", vec!["date0"])),
            foreign_field: Box::new(mir_field_path("bar2", vec!["date0"])),
            cache: SchemaCache::new(),
        }))),
        specs: vec![SortSpecification::Asc(mir_field_path(
            "bar2",
            vec!["x", "a", "b"],
        ))],
        cache: SchemaCache::new(),
    })
);

test_move_stage!(
    move_sort_above_join,
    expected = Stage::Join(Join {
        join_type: JoinType::Inner,
        left: Box::new(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["y"]))],
            cache: SchemaCache::new(),
        })),
        right: mir_collection("foo", "bar2"),
        condition: None,
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Sort(Sort {
        source: Box::new(Stage::Join(Join {
            join_type: JoinType::Inner,
            left: mir_collection("foo", "bar"),
            right: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
            condition: None,
        })),
        specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["y"]))],
        cache: SchemaCache::new(),
    }),
);

test_move_stage_no_op!(
    cannot_move_sort_above_right_side_of_join,
    Stage::Sort(Sort {
        source: Box::new(Stage::Join(Join {
            join_type: JoinType::Inner,
            left: mir_collection("foo", "bar2"),
            right: mir_collection("foo", "bar"),
            cache: SchemaCache::new(),
            condition: None,
        })),
        specs: vec![SortSpecification::Asc(mir_field_path("bar", vec!["y"]))],
        cache: SchemaCache::new(),
    })
);

test_move_stage!(
    move_filter_above_lateral_inner_join_if_only_left_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Gt,
                vec![
                    *mir_field_access("bar", "y", false),
                    Expression::Literal(LiteralValue::Integer(0)),
                ],
            )),
            cache: SchemaCache::new(),
        })),
        subquery: mir_collection("foo", "bar2"),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access("bar", "y", false),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_into_lateral_join_subquery_if_only_right_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: mir_collection("foo", "bar"),
        subquery: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar2"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Gt,
                vec![
                    *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                    Expression::Literal(LiteralValue::Integer(0)),
                ],
            )),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_into_lateral_inner_join_subquery_if_both_datasources_are_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: mir_collection("foo", "bar"),
        subquery: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar2"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Gt,
                vec![
                    *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                    *mir_field_access("bar", "y", false),
                ],
            )),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                *mir_field_access("bar", "y", false),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_match_filter_above_lateral_inner_join_if_only_left_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: Box::new(Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["y"])),
                arg: LiteralValue::Integer(43),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        subquery: mir_collection("foo", "bar2"),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar", vec!["y"])),
            arg: LiteralValue::Integer(43),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    move_match_filter_into_lateral_inner_join_if_right_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: mir_collection("foo", "bar"),
        subquery: Box::new(Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar2"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
                arg: LiteralValue::Integer(101),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
            arg: LiteralValue::Integer(101),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    move_match_filter_into_lateral_inner_join_subquery_if_both_datasources_are_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: mir_collection("foo", "bar"),
        subquery: Box::new(Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar2"),
            condition: MatchQuery::Logical(MatchLanguageLogical {
                op: MatchLanguageLogicalOp::And,
                args: vec![
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Gt,
                        input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
                        arg: LiteralValue::Integer(24),
                        cache: SchemaCache::new(),
                    }),
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Gt,
                        input: Some(mir_field_path("bar", vec!["y"])),
                        arg: LiteralValue::Integer(25),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Logical(MatchLanguageLogical {
            op: MatchLanguageLogicalOp::And,
            args: vec![
                MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
                    arg: LiteralValue::Integer(24),
                    cache: SchemaCache::new(),
                }),
                MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(mir_field_path("bar", vec!["y"])),
                    arg: LiteralValue::Integer(25),
                    cache: SchemaCache::new(),
                }),
            ],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    cannot_move_filter_above_lateral_left_join_if_only_left_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Left,
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Gt,
                vec![
                    *mir_field_access("bar", "y", false),
                    Expression::Literal(LiteralValue::Integer(0)),
                ],
            )),
            cache: SchemaCache::new(),
        })),
        subquery: mir_collection("foo", "bar2"),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access("bar", "y", false),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    cannot_move_filter_into_lateral_left_join_subquery_if_only_right_datasource_is_used,
    expected = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        )),
        cache: SchemaCache::new(),
    }),
    expected_changed = false,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    cannot_move_filter_into_lateral_left_join_subquery_if_both_datasources_are_used,
    expected = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                *mir_field_access("bar", "y", false),
            ],
        )),
        cache: SchemaCache::new(),
    }),
    expected_changed = false,
    input = Stage::Filter(Filter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                *mir_field_access_multi_part("bar2", vec!["x", "a", "b"], false),
                *mir_field_access("bar", "y", false),
            ],
        )),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_match_filter_above_lateral_left_join_if_only_left_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Left,
        source: Box::new(Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("bar", vec!["y"])),
                arg: LiteralValue::Integer(43),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        subquery: mir_collection("foo", "bar2"),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar", vec!["y"])),
            arg: LiteralValue::Integer(43),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    cannot_move_match_filter_into_lateral_left_join_if_right_datasource_is_used,
    expected = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
            arg: LiteralValue::Integer(101),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
    expected_changed = false,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
            arg: LiteralValue::Integer(101),
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    cannot_move_match_filter_into_lateral_left_join_subquery_if_both_datasources_are_used,
    expected = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Logical(MatchLanguageLogical {
            op: MatchLanguageLogicalOp::And,
            args: vec![
                MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
                    arg: LiteralValue::Integer(24),
                    cache: SchemaCache::new(),
                }),
                MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(mir_field_path("bar", vec!["y"])),
                    arg: LiteralValue::Integer(25),
                    cache: SchemaCache::new(),
                }),
            ],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
    expected_changed = false,
    input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: Box::new(Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: mir_collection("foo", "bar"),
            subquery: mir_collection("foo", "bar2"),
            cache: SchemaCache::new(),
        }))),
        condition: MatchQuery::Logical(MatchLanguageLogical {
            op: MatchLanguageLogicalOp::And,
            args: vec![
                MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(mir_field_path("bar2", vec!["x", "a", "b"])),
                    arg: LiteralValue::Integer(24),
                    cache: SchemaCache::new(),
                }),
                MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(mir_field_path("bar", vec!["y"])),
                    arg: LiteralValue::Integer(25),
                    cache: SchemaCache::new(),
                }),
            ],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
);

test_move_stage!(
    move_later_filter_above_unwind_that_prohibits_movement_of_earlier_filter,
    expected = Stage::Filter(Filter {
        source: Stage::Unwind(Unwind {
            source: Stage::Project(Project {
                            is_add_fields: false,
                source: Stage::Filter(Filter {
                    source: mir_collection("foo", "bar"),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication::new(mir::ScalarFunction::Lt,vec![
                                mir::Expression::Literal(mir::LiteralValue::Integer(41)),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],)
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        }.into()
                    ),
                }),
                cache: SchemaCache::new(),
            }).into(),
            path: mir_field_path("__bot__", vec!["x"]),
            index: None,
            outer: false,
            cache: SchemaCache::new(),
            is_prefiltered: false,
        })
        .into(),
        condition: Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lt,
                args: vec![
                    *mir_field_access("__bot__", "x", true),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                ],
                is_nullable: true,
            }
        ),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Project(Project {
                            is_add_fields: false,
                    source: mir_collection("foo", "bar"),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            }.into()
                        ),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                path: mir_field_path("__bot__", vec!["x"]),
                index: None,
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        *mir_field_access("__bot__", "x", true),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: true,
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        condition: Expression::ScalarFunction(
            mir::ScalarFunctionApplication::new(mir::ScalarFunction::Lt,vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(41)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                ],)
        ),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    filter_with_subquery_does_not_move_to_start_of_pipeline,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Stage::Filter(Filter {
            source: Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "foo"),
                expression: BindingTuple(map! {
                    Key::bot(0) => Document(
                        unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("__bot__", "x", false),
                        }.into()
                    )
                }),
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Subquery(SubqueryExpr {
                output_expr: Box::from(FieldAccess(FieldAccess {
                    expr: Box::from(Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "b".to_string() => *mir_field_access("__bot__", "a", false),
                        }
                    })),
                    field: "y".to_string(),
                    is_nullable: true,
                })),
                subquery: Box::new(Stage::Filter(Filter {
                    source: mir_collection("db", "bar"),
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            FieldAccess(FieldAccess {
                                expr: Box::from(Document(DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "b".to_string() => *mir_field_access("__bot__", "a", false),
                                    }
                                })),
                                field: "z".to_string(),
                                is_nullable: false,
                            }),
                            Literal(Integer(5)),
                        ],
                        is_nullable: false,
                    }),
                    cache: SchemaCache::new(),
                })),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        expression: BindingTuple(map! {
            Key::bot(0) => Document(
                unchecked_unique_linked_hash_map! {
                    "b".to_string() => *mir_field_access("__bot__", "a", false),
                }.into()
            )
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Stage::Project(Project {
            is_add_fields: false,
            source: Stage::Project(Project {
                is_add_fields: false,
                source: mir_collection("db", "foo"),
                expression: BindingTuple(map! {
                    Key::bot(0) => Document(
                        unchecked_unique_linked_hash_map! {
                            "a".to_string() => *mir_field_access("__bot__", "x", false),
                        }.into()
                    )
                }),
                cache: SchemaCache::new(),
            })
            .into(),
            expression: BindingTuple(map! {
                Key::bot(0) => Document(
                    unchecked_unique_linked_hash_map! {
                        "b".to_string() => *mir_field_access("__bot__", "a", false),
                    }.into()
                )
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Subquery(SubqueryExpr {
            output_expr: mir_field_access("__bot__", "y", true),
            subquery: Box::new(Stage::Filter(Filter {
                source: mir_collection("db", "bar"),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Eq,
                    args: vec![
                        *mir_field_access("__bot__", "z", false),
                        Literal(Integer(5)),
                    ],
                    is_nullable: false,
                }),
                cache: SchemaCache::new(),
            })),
            is_nullable: true,
        }),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    subquery_filter_with_field_existence_and_comparison_does_not_move_to_start,
    expected = Stage::Filter(Filter {
        source: mir_project_collection(None, "nullable_fields", None, None),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::SubqueryComparison(SubqueryComparison {
                    operator: SubqueryComparisonOp::Eq,
                    modifier: SubqueryModifier::Any,
                    argument: mir_field_access("nullable_fields", "b", true),
                    subquery_expr: SubqueryExpr {
                        output_expr: mir_field_access("__bot__", "b", false),
                        subquery: Box::new(Stage::Project(Project {
                            is_add_fields: false,
                            source: Box::new(Stage::Filter(Filter {
                                source: mir_collection("test_db", "non_nullable_fields"),
                                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                    function: ScalarFunction::Lt,
                                    args: vec![
                                        *mir_field_access("non_nullable_fields", "a", false),
                                        Expression::Literal(LiteralValue::Integer(300)),
                                    ],
                                    is_nullable: false,
                                }),
                                cache: SchemaCache::new(),
                            })),
                            expression: BindingTuple(map! {
                                Key::bot(1) => Expression::Document(DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "b".to_string() => *mir_field_access("non_nullable_fields", "b", false),
                                    }
                                }),
                            }),
                            cache: SchemaCache::new(),
                        })),
                        is_nullable: true,
                    },
                    is_nullable: true,
                }),
                Expression::MQLIntrinsicFieldExistence(FieldAccess {
                    expr: Box::new(Expression::Reference(ReferenceExpr {
                        key: Key::named("nullable_fields", 0),
                    })),
                    field: "a".to_string(),
                    is_nullable: true,
                }),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![
                        *mir_field_access("nullable_fields", "a", false),
                        Expression::Literal(LiteralValue::Integer(100)),
                    ],
                    is_nullable: false,
                }),
            ],
            is_nullable: true,
        }),
        cache: SchemaCache::new(),
    }),
    expected_changed = false,
    input = Stage::Filter(Filter {
        source: mir_project_collection(None, "nullable_fields", None, None),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::SubqueryComparison(SubqueryComparison {
                    operator: SubqueryComparisonOp::Eq,
                    modifier: SubqueryModifier::Any,
                    argument: mir_field_access("nullable_fields", "b", true),
                    subquery_expr: SubqueryExpr {
                        output_expr: mir_field_access("__bot__", "b", false),
                        subquery: Box::new(Stage::Project(Project {
                            is_add_fields: false,
                            source: Box::new(Stage::Filter(Filter {
                                source: mir_collection("test_db", "non_nullable_fields"),
                                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                                    function: ScalarFunction::Lt,
                                    args: vec![
                                        *mir_field_access("non_nullable_fields", "a", false),
                                        Expression::Literal(LiteralValue::Integer(300)),
                                    ],
                                    is_nullable: false,
                                }),
                                cache: SchemaCache::new(),
                            })),
                            expression: BindingTuple(map! {
                                Key::bot(1) => Expression::Document(DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "b".to_string() => *mir_field_access("non_nullable_fields", "b", false),
                                    }
                                }),
                            }),
                            cache: SchemaCache::new(),
                        })),
                        is_nullable: true,
                    },
                    is_nullable: true,
                }),
                Expression::MQLIntrinsicFieldExistence(FieldAccess {
                    expr: Box::new(Expression::Reference(ReferenceExpr {
                        key: Key::named("nullable_fields", 0),
                    })),
                    field: "a".to_string(),
                    is_nullable: true,
                }),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![
                        *mir_field_access("nullable_fields", "a", false),
                        Expression::Literal(LiteralValue::Integer(100)),
                    ],
                    is_nullable: false,
                }),
            ],
            is_nullable: true,
        }),
        cache: SchemaCache::new(),
    }),
);
