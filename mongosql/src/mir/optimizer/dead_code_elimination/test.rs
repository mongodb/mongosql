use crate::{
    catalog::{Catalog, Namespace},
    map, mir,
    mir::{
        optimizer::{dead_code_elimination::DeadCodeEliminator, Optimizer},
        schema::{SchemaCache, SchemaInferenceState},
        *,
    },
    schema::{Atomic, Document, Schema, SchemaEnvironment},
    set, unchecked_unique_linked_hash_map,
    util::mir_collection,
    SchemaCheckingMode,
};
use lazy_static::lazy_static;

lazy_static! {
    static ref CATALOG: Catalog = Catalog::new(map! {
        Namespace {db: "db".into(), collection: "bar".into()} => Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Double),
            },
            required: set! {},
            additional_properties: true,
            ..Default::default()
            }),
    });
}

macro_rules! test_dead_code_elimination {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let state = SchemaInferenceState::new(
                0u16,
                SchemaEnvironment::default(),
                &*CATALOG,
                SchemaCheckingMode::Relaxed,
            );

            let optimizer = &DeadCodeEliminator;
            let (actual, _) = optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!(expected, actual);
        }
    };
}

test_dead_code_elimination!(
    cannot_eliminate_non_project_source_for_group,
    expected = Stage::Group(Group {
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("db", "bar"),
            condition: Expression::Literal(LiteralValue::Boolean(true),),
            cache: SchemaCache::new(),
        })),
        keys: vec![],
        aggregations: vec![],
        scope: 0u16,
        cache: SchemaCache::new(),
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Filter(Filter {
            source: mir_collection("db", "bar"),
            condition: Expression::Literal(LiteralValue::Boolean(true),),
            cache: SchemaCache::new(),
        })),
        keys: vec![],
        aggregations: vec![],
        scope: 0u16,
        cache: SchemaCache::new(),
    })
);

test_dead_code_elimination!(
    swap_group_and_project,
    expected = Stage::Project(Project {
        source: Box::new(Stage::Project(Project {
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "a".to_string(),
                        expr: Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Reference(("bar", 0u16).into())),
                            "a".to_string(),
                        )),
                    }),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Reference(("bar", 0u16).into())),
                        "b".to_string(),
                    ))),
                ],
                aggregations: vec![AliasedAggregation {
                    alias: "agg".to_string(),
                    agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                        function: AggregationFunction::Avg,
                        distinct: false,
                        arg: Box::new(Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Reference(("bar", 0u16).into())),
                            "c".to_string(),
                        ))),
                    }),
                }],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                mir::binding_tuple::Key::bot(0) => mir::Expression::Reference(mir::binding_tuple::Key::bot(0).into()),
            },
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
        },
        cache: SchemaCache::new(),
    }),
    input = Stage::Project(Project {
        source: Box::new(Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "a".to_string(),
                    expr: Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Reference(("foo", 0u16).into())),
                        "a".to_string(),
                    ))
                }),
                OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 0u16).into())),
                    "b".to_string(),
                ))),
            ],
            aggregations: vec![AliasedAggregation {
                alias: "agg".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: Box::new(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Reference(("foo", 0u16).into())),
                        "c".to_string(),
                    ))),
                }),
            }],
            scope: 0u16,
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })
);

test_dead_code_elimination!(
    swap_group_and_project_outer_project_agg_only,
    expected = Stage::Project(Project {
        source: Box::new(Stage::Project(Project {
            source: Box::new(Stage::Group(Group {
                source: mir_collection("db", "bar"),
                keys: vec![
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "a".to_string(),
                        expr: Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Reference(("bar", 0u16).into())),
                            "a".to_string(),
                        ))
                    }),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Reference(("bar", 0u16).into())),
                        "b".to_string(),
                    ))),
                ],
                aggregations: vec![AliasedAggregation {
                    alias: "agg".to_string(),
                    agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                        function: AggregationFunction::Avg,
                        distinct: false,
                        arg: Box::new(Expression::FieldAccess(FieldAccess::new(
                            Box::new(Expression::Reference(("bar", 0u16).into())),
                            "c".to_string(),
                        ))),
                    }),
                }],
                scope: 0u16,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                mir::binding_tuple::Key::bot(0) => mir::Expression::Reference(mir::binding_tuple::Key::bot(0).into()),
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        expression: map! {
            mir::binding_tuple::Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!(
                    "agg".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr{
                            key: mir::binding_tuple::Key::bot(0),
                        })),
                        field: "_agg1".to_string(),
                        is_nullable: false,
                    })
                ),
            })
        },
        cache: SchemaCache::new(),
    }),
    input = Stage::Project(Project {
        source: Box::new(Stage::Group(Group {
            source: Box::new(Stage::Project(Project {
                source: mir_collection("db", "bar"),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "a".to_string(),
                    expr: Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Reference(("foo", 0u16).into())),
                        "a".to_string(),
                    ))
                }),
                OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 0u16).into())),
                    "b".to_string(),
                ))),
            ],
            aggregations: vec![AliasedAggregation {
                alias: "agg".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: Box::new(Expression::FieldAccess(FieldAccess::new(
                        Box::new(Expression::Reference(("foo", 0u16).into())),
                        "c".to_string(),
                    ))),
                }),
            }],
            scope: 0u16,
            cache: SchemaCache::new(),
        })),
        expression: map! {
            mir::binding_tuple::Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!(
                    "agg".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr{
                            key: mir::binding_tuple::Key::bot(0),
                        })),
                        field: "_agg1".to_string(),
                        is_nullable: false,
                    })
                ),
            })
        },
        cache: SchemaCache::new(),
    })
);

test_dead_code_elimination!(
    cannot_eliminate_project_source_for_group_if_not_all_sources_are_substitutable,
    expected = Stage::Group(Group {
        source: Box::new(Stage::Project(Project {
            source: mir_collection("db", "bar"),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".to_string(),
                expr: Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 0u16).into())),
                    "a".to_string(),
                ))
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference(("bad", 0u16).into())),
                "b".to_string(),
            ))),
        ],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: Box::new(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("bad", 0u16).into())),
                    "c".to_string(),
                ))),
            }),
        }],
        scope: 0u16,
        cache: SchemaCache::new(),
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Project(Project {
            source: mir_collection("db", "bar"),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".to_string(),
                expr: Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 0u16).into())),
                    "a".to_string(),
                ))
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference(("bad", 0u16).into())),
                "b".to_string(),
            ))),
        ],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: Box::new(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("bad", 0u16).into())),
                    "c".to_string(),
                ))),
            }),
        }],
        scope: 0u16,
        cache: SchemaCache::new(),
    })
);
