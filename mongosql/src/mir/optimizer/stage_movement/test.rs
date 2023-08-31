use crate::{
    catalog::Catalog,
    map, mir,
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

fn mir_reference(name: &str) -> mir::Expression {
    mir::Expression::Reference(mir::ReferenceExpr {
        key: if name == "__bot__" {
            mir::binding_tuple::Key::bot(0)
        } else {
            mir::binding_tuple::Key::named(name, 0)
        },
        cache: mir::schema::SchemaCache::new(),
    })
}

fn mir_field_access(ref_name: &str, field_name: &str) -> mir::Expression {
    mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir_reference(ref_name)),
        field: field_name.to_string(),
        cache: mir::schema::SchemaCache::new(),
    })
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
                path: mir::FieldPath {
                    key: Key::bot(0u16),
                    fields: vec!["x".to_string()],
                    cache: SchemaCache::new(),
                },
                index: None,
                outer: false,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir_field_access("__bot__", "x"),
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
                path: mir::FieldPath {
                    key: Key::bot(0u16),
                    fields: vec!["x".to_string()],
                    cache: SchemaCache::new(),
                },
                index: None,
                outer: false,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir_field_access("__bot__", "x"),
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
                   mir::FieldPath {
                       key: Key::bot(0u16),
                       fields: vec!["y".to_string()],
                       cache: SchemaCache::new(),
                   }
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
                   mir::FieldPath {
                       key: Key::bot(0u16),
                       fields: vec!["y".to_string()],
                       cache: SchemaCache::new(),
                   }
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
                specs: vec![SortSpecification::Asc(mir::FieldPath {
                    key: Key::named("bar", 0u16),
                    fields: vec!["y".to_string()],
                    cache: SchemaCache::new(),
                }),],
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
            specs: vec![SortSpecification::Asc(mir::FieldPath {
                key: Key::bot(0u16),
                fields: vec!["y".to_string()],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_sorts_above_project_will_not_reorder_sort_to_sort,
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
                            mir::FieldPath {
                                key: Key::named("bar", 0u16),
                                fields: vec!["x".to_string(), "a".to_string()],
                                cache: SchemaCache::new(),
                            }
                        ),
                    ],
                    cache: SchemaCache::new(),
                }).into(),
                specs: vec![
                    SortSpecification::Desc(
                        mir::FieldPath {
                            key: Key::named("bar", 0u16),
                            fields: vec!["x".to_string(), "b".to_string()],
                            cache: SchemaCache::new(),
                        }
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
                        mir::FieldPath {
                            key: Key::named("bar", 0u16),
                            fields: vec!["a".to_string()],
                            cache: SchemaCache::new(),
                        }
                    ),
                ],
                cache: SchemaCache::new(),
            })
            .into(),
            specs: vec![
                SortSpecification::Desc(
                    mir::FieldPath {
                         key: Key::named("bar", 0u16),
                         fields: vec!["b".to_string()],
                         cache: SchemaCache::new(),
                    }
                ),
            ],
            cache: SchemaCache::new(),
        }).into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_sorts_above_project_but_not_above_group,
    expected = Stage::Limit(Limit {
        source: Stage::Project(Project {
            source: Stage::Sort(Sort {
                source: Stage::Group(Group {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    cache: SchemaCache::new(),
                    keys: vec![],
                    aggregations: vec![],
                    scope: 0
                })
                .into(),
                specs: vec![SortSpecification::Desc(mir::FieldPath {
                    key: Key::named("bar", 0u16),
                    fields: vec!["x".to_string(), "b".to_string()],
                    cache: SchemaCache::new(),
                },),],
                cache: SchemaCache::new(),
            })
            .into(),
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
            source: Stage::Project(Project {
                source: Stage::Group(Group {
                    source: Stage::Collection(Collection {
                        db: "foo".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    cache: SchemaCache::new(),
                    keys: vec![],
                    aggregations: vec![],
                    scope: 0
                })
                .into(),
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
            specs: vec![SortSpecification::Desc(mir::FieldPath {
                key: Key::named("bar", 0u16),
                fields: vec!["b".to_string()],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        })
        .into(),
        limit: 10,
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filters_above_project_will_reorder_filter_to_filter,
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
                               mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                           ],
                           cache: SchemaCache::new(),
                    }),
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
    move_filter_under_join_into_none_on,
    expected = Stage::Join(Join {
        left: Stage::Collection(Collection {
            db: "foo".to_string(),
            collection: "bar".to_string(),
            cache: SchemaCache::new(),
        })
        .into(),
        right: Stage::Collection(Collection {
            db: "foo".to_string(),
            collection: "bar2".to_string(),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Some(Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        })),
        join_type: JoinType::Inner,
        cache: SchemaCache::new(),
    }),
    input = Stage::Filter(Filter {
        source: Stage::Join(Join {
            left: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "bar".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            right: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "bar2".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            condition: None,
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
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_filter_under_join_into_some_on,
    expected = Stage::Join(Join {
        left: Stage::Collection(Collection {
            db: "foo".to_string(),
            collection: "bar".to_string(),
            cache: SchemaCache::new(),
        })
        .into(),
        right: Stage::Collection(Collection {
            db: "foo".to_string(),
            collection: "bar2".to_string(),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Some(Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::And,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Reference(Key::named("bar", 0u16).into()),
                        mir::Expression::Reference(Key::named("bar2", 0u16).into()),
                    ],
                    cache: SchemaCache::new(),
                })
            ],
            cache: SchemaCache::new(),
        })),
        join_type: JoinType::Inner,
        cache: SchemaCache::new(),
    }),
    input = Stage::Filter(Filter {
        source: Stage::Join(Join {
            left: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "bar".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            right: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "bar2".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Some(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
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
            cache: SchemaCache::new(),
        }),
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
                                mir::Expression::FieldAccess(mir::FieldAccess {
                                    expr: mir::Expression::Document(mir::DocumentExpr {
                                        document: unchecked_unique_linked_hash_map! {
                                            "x".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                                        },
                                        cache: SchemaCache::new(),
                                    }).into(),
                                    field: "x".into(),
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
                        mir_field_access("bar", "x"),
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

test_move_stage!(
    move_two_filters_under_join_one_into_none_on_and_other_filter_above_join_because_filters_can_reorder,
    expected = Stage::Join(Join {
        left: Stage::Filter(Filter {
            source: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "bar".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Eq,
                args: vec![
                    mir::Expression::Reference(Key::named("bar", 0u16).into()),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        right: Stage::Collection(Collection {
            db: "foo".to_string(),
            collection: "bar2".to_string(),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Some(Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Reference(Key::named("bar2", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        })),
        join_type: JoinType::Inner,
        cache: SchemaCache::new(),
    }),
    input = Stage::Filter(Filter {
        source: Stage::Filter(Filter {
            source: Stage::Join(Join {
                left: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string(),
                    cache: SchemaCache::new(),
                })
                .into(),
                right: Stage::Collection(Collection {
                    db: "foo".to_string(),
                    collection: "bar2".to_string(),
                    cache: SchemaCache::new(),
                })
                .into(),
                condition: None,
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
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Eq,
            args: vec![
                mir::Expression::Reference(Key::named("bar", 0u16).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
            ],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
);

test_move_stage!(
    move_later_filter_above_unwind_that_prohibits_movement_of_earlier_filter,
    expected = Stage::Filter(Filter {
        source: Stage::Unwind(Unwind {
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
                                mir::Expression::Literal(mir::LiteralValue::Integer(41).into()),
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
            path: mir::FieldPath {
                key: Key::bot(0u16),
                fields: vec!["x".to_string()],
                cache: SchemaCache::new(),
            },
            index: None,
            outer: false,
            cache: SchemaCache::new(),
        })
        .into(),
        condition: Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lt,
                args: vec![
                    mir_field_access("__bot__", "x"),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }
        ),
        cache: SchemaCache::new(),
    }),
    input = Stage::Filter(Filter {
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
                path: mir::FieldPath {
                    key: Key::bot(0u16),
                    fields: vec!["x".to_string()],
                    cache: SchemaCache::new(),
                },
                index: None,
                outer: false,
                cache: SchemaCache::new(),
            })
            .into(),
            condition: Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir_field_access("__bot__", "x"),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
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
                    mir::Expression::Literal(mir::LiteralValue::Integer(41).into()),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }
        ),
        cache: SchemaCache::new(),
    }),
);
