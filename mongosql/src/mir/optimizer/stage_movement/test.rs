use crate::{
    catalog::Catalog,
    map,
    schema::{self, Schema},
    set,
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
                                        }
                                    )
                                },
                                required: set!{"a".to_string()},
                                additional_properties: false,
                            }
                        ),
                        "y".to_string() => schema::Schema::Atomic(schema::Atomic::Integer),
                        "z".to_string() => schema::Schema::Atomic(schema::Atomic::Integer),
                    },
                required: set!{"x".to_string()},
                additional_properties: false,
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
                                        }
                                    )
                                },
                                required: set!{"a".to_string()},
                                additional_properties: false,
                            }
                        ),
                    },
                required: set!{"x".to_string()},
                additional_properties: false,
            }
        )
    });
}

macro_rules! test_move_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                catalog::Catalog,
                map,
                mir::{
                    self,
                    binding_tuple::{BindingTuple, Key},
                    optimizer::stage_movement::StageMovementVisitor,
                    schema::{SchemaCache, SchemaCheckingMode, SchemaInferenceState},
                    visitor::Visitor,
                    Collection,
                    Expression::{self, *},
                    FieldAccess, Filter, Group, Join, JoinType, Limit, LiteralExpr,
                    LiteralValue::*,
                    Offset, Project, ReferenceExpr, Set, SetOperation, Sort, SortSpecification,
                    Stage, Unwind,
                },
                schema::SchemaEnvironment,
                set, unchecked_unique_linked_hash_map,
            };
            #[allow(unused)]
            let input = $input;
            let expected = $expected;
            let mut visitor = StageMovementVisitor {
                schema_state: SchemaInferenceState::new(
                    0,
                    SchemaEnvironment::new(),
                    &*CATALOG,
                    SchemaCheckingMode::Relaxed,
                ),
            };
            assert_eq!(expected, visitor.visit_stage(input));
        }
    };
}

test_move_stage!(
    move_offsets_above_projects,
    expected = Stage::Project(Project {
        source: Stage::Project(Project {
            source: Stage::Offset(Offset {
                source: Stage::Offset(Offset {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    offset: 44,
                    cache: SchemaCache::new(),
                }).into(),
                offset: 42,
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    },
                    cache: SchemaCache::new(),
                })
            }),
            cache: SchemaCache::new(),
        }).into(),
        expression: BindingTuple(map! {
            Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {
                    "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(43).into()),
                },
                cache: SchemaCache::new(),
            })
        }),
        cache: SchemaCache::new(),
    }),
    input = Stage::Offset(Offset {
        source: Stage::Offset(Offset {
            source: Stage::Project(Project {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                        })
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(43).into()),
                        },
                    cache: SchemaCache::new(),
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
    expected = Stage::Limit(Limit {
        source: Stage::Project(Project {
            source: Stage::Project(Project {
                source: Stage::Filter(Filter {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                    },
                                    cache: SchemaCache::new(),
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            ],
                            cache: SchemaCache::new(),
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
                }),
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("bar", 0u16).into()),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Project(Project {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
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
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    do_not_move_filter_above_opaque_defines,
    expected = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                path: Expression::FieldAccess( FieldAccess {
                    expr: Expression::Reference( ReferenceExpr {
                        key: Key::bot(0u16),
                        cache: SchemaCache::new(),
                    }).into(),
                    field: "x".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                index: None,
                outer: false,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::bot(0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                path: Expression::FieldAccess( FieldAccess {
                    expr: Expression::Reference( ReferenceExpr {
                        key: Key::bot(0u16),
                        cache: SchemaCache::new(),
                    }).into(),
                    field: "x".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                index: None,
                outer: false,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::bot(0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    do_not_move_sort_above_project_when_nonsubstitutable_complex_expression_is_used,
    expected = Stage::Limit(Limit {
        source: Stage::Sort(Sort {
            source: Stage::Project(Project {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::bot(0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "y".to_string() => mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                function: mir::ScalarFunction::Add,
                                args: vec![],
                                cache: SchemaCache::new(),
                            }),
                       },
                       cache: SchemaCache::new(),
                   }),
               }),
               cache: SchemaCache::new(),
           }).into(),
           specs: vec![
               SortSpecification::Asc(
                   mir::Expression::FieldAccess(mir::FieldAccess {
                       expr: mir::Expression::Reference(Key::bot(0u16).into()).into(),
                       field: "y".to_string(),
                       cache: SchemaCache::new(),
                   }).into(),
               ),
           ],
           cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    input = Stage::Limit(Limit {
        source: Stage::Sort(Sort {
            source: Stage::Project(Project {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    Key::bot(0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "y".to_string() => mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                function: mir::ScalarFunction::Add,
                                args: vec![],
                                cache: SchemaCache::new(),
                            }),
                       },
                       cache: SchemaCache::new(),
                   }),
               }),
               cache: SchemaCache::new(),
           }).into(),
           specs: vec![
               SortSpecification::Asc(
                   mir::Expression::FieldAccess(mir::FieldAccess {
                       expr: mir::Expression::Reference(Key::bot(0u16).into()).into(),
                       field: "y".to_string(),
                       cache: SchemaCache::new(),
                   }).into(),
               ),
           ],
           cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_sort_above_project_when_substitutable_complex_expression_is_used,
    expected = Stage::Limit(Limit {
        source: Stage::Project(Project {
            source: Stage::Sort(Sort {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                })
                .into(),
                specs: vec![SortSpecification::Asc(
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: mir::Expression::Reference(("bar", 0u16).into()).into(),
                        field: "y".to_string(),
                        cache: SchemaCache::new(),
                    })
                    .into(),
                ),],
                cache: SchemaCache::new(),
            })
            .into(),
            expression: BindingTuple(map! {
                Key::bot(0u16) => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "y".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("bar", 0u16).into())),
                            field: "y".to_string(),
                            cache: SchemaCache::new(),
                        }),
                    },
                    cache: SchemaCache::new(),
                }),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    input = Stage::Limit(Limit {
        source: Stage::Sort(Sort {
            source: Stage::Project(Project {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                })
                .into(),
                expression: BindingTuple(map! {
                     Key::bot(0u16) => mir::Expression::Document(mir::DocumentExpr {
                         document: unchecked_unique_linked_hash_map! {
                             "y".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                                 expr: Box::new(mir::Expression::Reference(("bar", 0u16).into())),
                                 field: "y".to_string(),
                                 cache: SchemaCache::new(),
                             }),
                        },
                        cache: SchemaCache::new(),
                    }),
                }),
                cache: SchemaCache::new(),
            })
            .into(),
            specs: vec![SortSpecification::Asc(
                mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: mir::Expression::Reference(Key::bot(0u16).into()).into(),
                    field: "y".to_string(),
                    cache: SchemaCache::new(),
                })
                .into(),
            ),],
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_sorts_above_project_without_reordering_sort_to_sort,
    expected = Stage::Limit(Limit {
        source: Stage::Project(Project {
            source: Stage::Sort(Sort {
                source: Stage::Sort(Sort {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    specs: vec![
                        SortSpecification::Asc(
                            mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: mir::Expression::FieldAccess(mir::FieldAccess {
                                    expr: mir::Expression::Reference(Key::named("bar", 0u16).into()).into(),
                                    field: "x".to_string(),
                                    cache: SchemaCache::new(),
                                }).into(),
                                field: "a".to_string(),
                                cache: SchemaCache::new(),
                            }).into(),
                        ),
                    ],
                    cache: SchemaCache::new(),
                }).into(),
                specs: vec![
                    SortSpecification::Desc(
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: mir::Expression::Reference(Key::named("bar", 0u16).into()).into(),
                                field: "x".to_string(),
                                cache: SchemaCache::new(),
                            }).into(),
                            field: "b".to_string(),
                            cache: SchemaCache::new(),
                        }).into(),
                    ),
                ],
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::named("bar", 0u16) => mir::Expression::FieldAccess(mir::FieldAccess {
                     expr: mir::Expression::Reference(Key::named("bar", 0u16).into()).into(),
                     field: "x".to_string(),
                     cache: SchemaCache::new(),
                }),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    input = Stage::Limit(Limit {
        source: Stage::Sort(Sort {
            source: Stage::Sort(Sort {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::FieldAccess(mir::FieldAccess {
                             expr: mir::Expression::Reference(Key::named("bar", 0u16).into()).into(),
                             field: "x".to_string(),
                             cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                specs: vec![
                    SortSpecification::Asc(
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: mir::Expression::Reference(Key::named("bar", 0u16).into()).into(),
                            field: "a".to_string(),
                            cache: SchemaCache::new(),
                        }).into(),
                    ),
                ],
                cache: SchemaCache::new(),
            })
            .into(),
            specs: vec![
                SortSpecification::Desc(
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: mir::Expression::Reference(Key::named("bar", 0u16).into()).into(),
                        field: "b".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                ),
            ],
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filters_above_project_without_reordering_filter_to_filter,
    expected = Stage::Limit(Limit {
        source: Stage::Project(Project {
            source: Stage::Filter(Filter {
                source: Stage::Filter(Filter {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                    },
                                    cache: SchemaCache::new(),
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(54).into()),
                            ],
                            cache: SchemaCache::new(),
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                condition: Expression::ScalarFunction(
                   mir::ScalarFunctionApplication {
                       function: mir::ScalarFunction::Lt,
                       args: vec![
                           mir::Expression::Document(mir::DocumentExpr {
                               document: unchecked_unique_linked_hash_map! {
                                   "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                               },
                               cache: SchemaCache::new(),
                           }),
                           mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                       ],
                       cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            }).into(),
            expression: BindingTuple(map! {
                Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    },
                    cache: SchemaCache::new(),
                }),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Filter(Filter {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                 condition: Expression::ScalarFunction(
                     mir::ScalarFunctionApplication {
                         function: mir::ScalarFunction::Lt,
                         args: vec![
                             mir::Expression::Reference(Key::named("bar", 0u16).into()),
                             mir::Expression::Literal(mir::LiteralValue::Integer(54).into()),
                         ],
                         cache: SchemaCache::new(),
                     }
                 ),
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_right_above_join_and_projects,
    expected = Stage::Limit(Limit {
        source: Stage::Join( Join {
            left: Stage::Project(Project {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                source: Stage::Filter(Filter {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar2".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                    },
                                    cache: SchemaCache::new(),
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            ],
                            cache: SchemaCache::new(),
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
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
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Join( Join {
                left: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar2".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
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
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_left_above_join_and_projects,
    expected = Stage::Limit(Limit {
        source: Stage::Join( Join {
            left: Stage::Project(Project {
                source: Stage::Filter(Filter {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                    },
                                    cache: SchemaCache::new(),
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            ],
                            cache: SchemaCache::new(),
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar2".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
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
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Join( Join {
                left: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar2".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
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
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
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
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                source: Stage::Filter(Filter {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar2".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                    },
                                    cache: SchemaCache::new(),
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            ],
                            cache: SchemaCache::new(),
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
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
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Set( Set {
                left: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar2".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                operation: SetOperation::UnionAll,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar2", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
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
                source: Stage::Filter(Filter {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    condition: Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Lt,
                            args: vec![
                                mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                    },
                                    cache: SchemaCache::new(),
                                }),
                                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            ],
                            cache: SchemaCache::new(),
                        }
                    ),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
                }),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Project(Project {
                source: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar2".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                expression: BindingTuple(map! {
                    // In any real query, "bar" will be bound to a Document, but we just use a
                    // Literal for simplicity.
                    Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                        },
                        cache: SchemaCache::new(),
                    }),
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
    input = Stage::Limit(Limit {
        source: Stage::Filter(Filter {
            source: Stage::Set( Set {
                left: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        Key::named("bar", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar2".to_string(),
                        cache: SchemaCache::new(),
                    }).into(),
                    expression: BindingTuple(map! {
                        // In any real query, "bar" will be bound to a Document, but we just use a
                        // Literal for simplicity.
                        Key::named("bar2", 0u16) => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                            },
                            cache: SchemaCache::new(),
                        }),
                    }),
                    cache: SchemaCache::new(),
                }).into(),
                operation: SetOperation::UnionAll,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                    ],
                    cache: SchemaCache::new(),
                }
            ),
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);
