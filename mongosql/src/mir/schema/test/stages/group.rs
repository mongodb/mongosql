use crate::{
    catalog::Namespace,
    map,
    mir::{
        schema::{Atomic, Document, Error as mir_error, Group, ResultSet, SchemaCache},
        AggregationExpr, AggregationFunction, AggregationFunctionApplication, AliasedAggregation,
        AliasedExpr, Collection, Expression, FieldAccess, LiteralValue, OptionallyAliasedExpr,
        Stage,
    },
    schema::{Schema, ANY_DOCUMENT},
    set, test_schema,
};
use mongosql_datastructures::binding_tuple::Key;

fn group_stage_refs_only() -> Stage {
    Stage::Group(Group {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        keys: vec![group_aliased_ref(), group_non_aliased_ref()],
        aggregations: vec![AliasedAggregation {
            alias: "agg".to_string(),
            agg_expr: AggregationExpr::CountStar(false),
        }],
        cache: SchemaCache::new(),
        scope: 0,
    })
}

fn group_aliased_ref() -> OptionallyAliasedExpr {
    OptionallyAliasedExpr::Aliased(AliasedExpr {
        alias: "A".into(),
        expr: Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        }),
    })
}

fn group_non_aliased_ref() -> OptionallyAliasedExpr {
    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
        field: "b".into(),
        cache: SchemaCache::new(),
    }))
}

test_schema!(
    key_schemas_are_all_self_comparable,
    expected_pat = Ok(ResultSet { .. }),
    input = group_stage_refs_only(),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Double)]),
                "b".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "a".into(), "b".into() },
            additional_properties: false,
        }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    key_schemas_not_all_self_comparable,
    expected_error_code = 1011,
    expected = Err(mir_error::GroupKeyNotSelfComparable(
        1,
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ]),
    )),
    input = group_stage_refs_only(),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::String),
                "b".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)]),
            },
            required: set! {"a".into(), "b".into()},
            additional_properties: false,
        }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    aliased_field_access_mapped_to_bottom_datasource,
    expected = Ok(ResultSet {
        schema_env: map! {
            Key::bot(0u16) => Schema::Document(Document {
                keys: map! {
                    "A".into() => Schema::Atomic(Atomic::String),
                },
                required: set! { "A".into() },
                additional_properties: false,
            }),
        },
        min_size: 0,
        max_size: None,
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        keys: vec![group_aliased_ref()],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "a".into(), },
            additional_properties: false,
        }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    non_aliased_field_access_mapped_to_reference,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String),
                },
                required: set! { "b".into() },
                additional_properties: false,
            }),
        },
        min_size: 0,
        max_size: None,
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        keys: vec![group_non_aliased_ref()],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document(Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "b".into(), },
            additional_properties: false,
        }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    all_literal_keys_results_in_max_size_1,
    expected_pat = Ok(ResultSet {
        max_size: Some(1),
        ..
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "a".into(),
                expr: Expression::Literal(LiteralValue::Integer(1).into()),
            }),
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "b".into(),
                expr: Expression::Literal(LiteralValue::String("abc".into()).into()),
            })
        ],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    mix_literal_key_and_non_literal_key_results_in_no_max_size,
    expected_pat = Ok(ResultSet { max_size: None, .. }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        keys: vec![
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "literal".into(),
                expr: Expression::Literal(LiteralValue::Integer(1).into()),
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            }))
        ],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "a".into() },
            additional_properties: false,
        }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    aliased_key_and_multiple_agg_functions_all_correctly_unioned_under_bottom_datasource,
    expected = Ok(ResultSet {
        schema_env: map! {
            Key::bot(0u16) => Schema::Document(Document {
                keys: map! {
                    "A".into() => Schema::Atomic(Atomic::Boolean),
                    "B".into() => Schema::Atomic(Atomic::String),
                    "literal".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "A".into(),
                    "B".into(),
                    "literal".into(),
                },
                additional_properties: false,
            })
        },
        min_size: 0,
        max_size: Some(1),
    }),
    input = Stage::Group(Group {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
            alias: "literal".into(),
            expr: Expression::Literal(LiteralValue::Integer(1).into()),
        })],
        aggregations: vec![
            AliasedAggregation {
                alias: "A".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::First,
                    distinct: false,
                    arg: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                }),
            },
            AliasedAggregation {
                alias: "B".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::First,
                    distinct: false,
                    arg: Expression::Literal(LiteralValue::String("abc".into()).into()).into(),
                }),
            },
        ],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);
