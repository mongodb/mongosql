use crate::{
    catalog::{Catalog, Namespace},
    map, mir,
    mir::{
        optimizer::{dead_code_elimination::DeadCodeEliminator, Optimizer},
        schema::{SchemaCache, SchemaInferenceState},
        *,
    },
    schema::{Atomic, Document, Schema, SchemaEnvironment},
    set, SchemaCheckingMode,
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
            let actual = optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!(expected, actual);
        }
    };
}

fn collection_source(collection: &str) -> Box<Stage> {
    Box::new(Stage::Collection(Collection {
        db: "db".to_string(),
        collection: collection.to_string(),
        cache: SchemaCache::new(),
    }))
}

test_dead_code_elimination!(
    cannot_eliminate_non_project_source_for_group,
    expected = Stage::Group(Group {
        source: Box::new(Stage::Filter(Filter {
            source: collection_source("bar"),
            condition: Expression::Literal(LiteralExpr {
                value: LiteralValue::Boolean(true),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })),
        keys: vec![],
        aggregations: vec![],
        scope: 0u16,
        cache: SchemaCache::new(),
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Filter(Filter {
            source: collection_source("bar"),
            condition: Expression::Literal(LiteralExpr {
                value: LiteralValue::Boolean(true),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })),
        keys: vec![],
        aggregations: vec![],
        scope: 0u16,
        cache: SchemaCache::new(),
    })
);

test_dead_code_elimination!(
    eliminate_project_source_for_group,
    expected = Stage::Group(Group {
        source: collection_source("bar"),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".to_string(),
                expr: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                        field: "a".to_string(),
                        cache: SchemaCache::new(),
                    })],
                    cache: SchemaCache::new(),
                })
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                field: "b".to_string(),
                cache: SchemaCache::new(),
            })),
        ],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                    field: "c".to_string(),
                    cache: SchemaCache::new(),
                })),
            }),
        }],
        scope: 0u16,
        cache: SchemaCache::new(),
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Project(Project {
            source: collection_source("bar"),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".to_string(),
                expr: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".to_string(),
                        cache: SchemaCache::new(),
                    })],
                    cache: SchemaCache::new(),
                })
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "b".to_string(),
                cache: SchemaCache::new(),
            })),
        ],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "c".to_string(),
                    cache: SchemaCache::new(),
                })),
            }),
        }],
        scope: 0u16,
        cache: SchemaCache::new(),
    })
);

test_dead_code_elimination!(
    cannot_eliminate_project_source_for_group_if_not_all_sources_are_substitutable,
    expected = Stage::Group(Group {
        source: Box::new(Stage::Project(Project {
            source: collection_source("bar"),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".to_string(),
                expr: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".to_string(),
                        cache: SchemaCache::new(),
                    })],
                    cache: SchemaCache::new(),
                })
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("bad", 0u16).into())),
                field: "b".to_string(),
                cache: SchemaCache::new(),
            })),
        ],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("bad", 0u16).into())),
                    field: "c".to_string(),
                    cache: SchemaCache::new(),
                })),
            }),
        }],
        scope: 0u16,
        cache: SchemaCache::new(),
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Project(Project {
            source: collection_source("bar"),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".to_string(),
                expr: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".to_string(),
                        cache: SchemaCache::new(),
                    })],
                    cache: SchemaCache::new(),
                })
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("bad", 0u16).into())),
                field: "b".to_string(),
                cache: SchemaCache::new(),
            })),
        ],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                function: AggregationFunction::Avg,
                distinct: false,
                arg: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("bad", 0u16).into())),
                    field: "c".to_string(),
                    cache: SchemaCache::new(),
                })),
            }),
        }],
        scope: 0u16,
        cache: SchemaCache::new(),
    })
);
