use crate::{
    map,
    mir::{
        AggregationExpr, AggregationFunction, AggregationFunctionApplication, AliasedAggregation,
        AliasedExpr, Collection, Expression, FieldAccess, LiteralValue, OptionallyAliasedExpr,
        Stage,
        schema::{Atomic, Document, Error as mir_error, Group, ResultSet, SchemaCache},
    },
    schema::{ANY_DOCUMENT, Satisfaction, Schema},
    set, test_schema,
    util::{mir_collection, mir_field_access, mir_field_access_multi_part},
};
use agg_ast::definitions::Namespace;
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
        expr: Expression::FieldAccess(FieldAccess::new(
            Box::new(Expression::Reference(("foo", 0u16).into())),
            "a".into(),
        )),
    })
}

fn group_non_aliased_ref() -> OptionallyAliasedExpr {
    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
        Box::new(Expression::Reference(("foo", 0u16).into())),
        "b".into(),
    )))
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
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
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
        ])
        .into(),
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
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
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
                ..Default::default()
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
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
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
                ..Default::default()
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
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    multipart_non_aliased_field_access_mapped_to_reference,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Document(Document {
                        keys: map! {
                            "b".into() => Schema::Atomic(Atomic::String),
                        },
                        required: set! { "b".into() },
                        additional_properties: false,
                        ..Default::default()
                    }),
                },
                required: set! { "a".into() },
                additional_properties: false,
                ..Default::default()
            }),
        },
        min_size: 0,
        max_size: None,
    }),
    input = Stage::Group(Group {
        source: mir_collection("test", "foo"),
        keys: vec![OptionallyAliasedExpr::Unaliased(
            *mir_field_access_multi_part("foo", vec!["a", "b"], false)
        )],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "b".into() },
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set! { "a".into(), },
            additional_properties: false,
            ..Default::default()
        }),
    }),
);

test_schema!(
    multiple_multipart_and_single_part_non_aliased_field_access_mapped_to_references,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Document(Document {
                        keys: map! {
                            "b".into() => Schema::Atomic(Atomic::String),
                            "c".into() => Schema::Atomic(Atomic::String),
                        },
                        required: set! { "b".into(), "c".into() },
                        additional_properties: false,
                        ..Default::default()
                    }),
                    "x".into() => Schema::Document(Document {
                        keys: map! {
                            "y".into() => Schema::Atomic(Atomic::String),
                        },
                        required: set! { "y".into() },
                        additional_properties: false,
                        ..Default::default()
                    }),
                    "single".into() => Schema::Atomic(Atomic::String),
                },
                required: set! { "a".into(), "x".into(), "single".into() },
                additional_properties: false,
                ..Default::default()
            }),
        },
        min_size: 0,
        max_size: None,
    }),
    input = Stage::Group(Group {
        source: mir_collection("test", "foo"),
        keys: vec![
            OptionallyAliasedExpr::Unaliased(*mir_field_access_multi_part(
                "foo",
                vec!["a", "b"],
                false
            )),
            OptionallyAliasedExpr::Unaliased(*mir_field_access_multi_part(
                "foo",
                vec!["a", "c"],
                false
            )),
            OptionallyAliasedExpr::Unaliased(*mir_field_access_multi_part(
                "foo",
                vec!["x", "y"],
                false
            )),
            OptionallyAliasedExpr::Unaliased(*mir_field_access("foo", "single", false)),
        ],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Atomic(Atomic::String),
                        "c".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "b".into(), "c".into() },
                    additional_properties: false,
                    ..Default::default()
                }),
                "x".into() => Schema::Document(Document {
                    keys: map! {
                        "y".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "y".into() },
                    additional_properties: false,
                    ..Default::default()
                }),
                "single".into() => Schema::Atomic(Atomic::String),
                "omitted".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "a".into(), "x".into(), "single".into(), "omitted".into() },
            additional_properties: false,
            ..Default::default()
        }),
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
                expr: Expression::Literal(LiteralValue::Integer(1)),
            }),
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "b".into(),
                expr: Expression::Literal(LiteralValue::String("abc".into())),
            })
        ],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
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
                expr: Expression::Literal(LiteralValue::Integer(1)),
            }),
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference(("foo", 0u16).into())),
                "a".into(),
            )))
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
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
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
                ..Default::default()
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
            expr: Expression::Literal(LiteralValue::Integer(1)),
        })],
        aggregations: vec![
            AliasedAggregation {
                alias: "A".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::First,
                    distinct: false,
                    arg: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    arg_is_possibly_doc: Satisfaction::Not,
                }),
            },
            AliasedAggregation {
                alias: "B".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::First,
                    distinct: false,
                    arg: Expression::Literal(LiteralValue::String("abc".into())).into(),
                    arg_is_possibly_doc: Satisfaction::Not,
                }),
            },
        ],
        cache: SchemaCache::new(),
        scope: 0,
    }),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);
