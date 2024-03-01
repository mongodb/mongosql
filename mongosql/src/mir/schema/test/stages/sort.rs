use crate::{
    catalog::Namespace,
    map,
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::*,
    set, test_schema,
};

test_schema!(
    comparable_schemas,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("bar", 0u16).into() => ANY_DOCUMENT.clone(),
        },
        min_size: 0,
        max_size: None,
    }),
    input = Stage::Sort(Sort {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        specs: vec![
            SortSpecification::Asc(FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["a".to_string()],
                is_nullable: false,
            }),
            SortSpecification::Desc(FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["b".to_string()],
                is_nullable: false,
            }),
        ],
        cache: SchemaCache::new(),
    }),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document( Document{
            keys: map! {
                "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)]),
                "b".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "a".into(), "b".into() },
            additional_properties: false,
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    incomparable_schemas,
    expected_error_code = 1010,
    expected = Err(mir_error::SortKeyNotSelfComparable(
        1,
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ]),
    )),
    input = Stage::Sort(Sort {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        specs: vec![
            SortSpecification::Asc(FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["a".to_string()],
                is_nullable: false,
            }),
            SortSpecification::Asc(FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["b".to_string()],
                is_nullable: false,
            }),
        ],
        cache: SchemaCache::new(),
    }),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document( Document{
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
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    mix_comparable_and_incomparable_schemas,
    expected_error_code = 1010,
    expected = Err(mir_error::SortKeyNotSelfComparable(
        0,
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ]),
    )),
    input = Stage::Sort(Sort {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        specs: vec![SortSpecification::Asc(FieldPath {
            key: ("foo", 0u16).into(),
            fields: vec!["a".to_string()],
            is_nullable: false,
        },)],
        cache: SchemaCache::new(),
    }),
    schema_env = map! {
        ("foo", 0u16).into() => Schema::Document( Document{
            keys: map! {
                "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)]),
            },
            required: set! {"a".into(), "b".into(), "c".into()},
            additional_properties: false,
            ..Default::default()
            }),
    },
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
    }),
);
