use crate::{
    map,
    mir::{
        schema::{
            test::{
                test_document_a, test_document_b, test_document_c, TEST_DOCUMENT_SCHEMA_A,
                TEST_DOCUMENT_SCHEMA_B, TEST_DOCUMENT_SCHEMA_C,
            },
            SchemaCache,
        },
        *,
    },
    schema::{ResultSet, Schema},
    set, test_schema,
};

test_schema!(
    set_unionall_same_name_unioned,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
                ]
            ),
        },
        min_size: 2,
        max_size: Some(2),
    }),
    input = Stage::Set(Set {
        operation: SetOperation::UnionAll,
        left: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_a()],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        right: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_b()],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    }),
);

test_schema!(
    set_unionall_distinct_name_not_unioned,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() =>
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
            ("bar", 0u16).into() =>
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
        },
        min_size: 2,
        max_size: Some(2),
    }),
    input = Stage::Set(Set {
        operation: SetOperation::UnionAll,
        left: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_a()],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        right: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_b()],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    }),
);

test_schema!(
    set_union_same_name_unioned,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
                ]
            ),
        },
        min_size: 1,
        max_size: Some(2),
    }),
    input = Stage::Set(Set {
        operation: SetOperation::Union,
        left: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_a()],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        right: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_b()],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    }),
);

test_schema!(
    set_union_distinct_name_not_unioned,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() =>
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone(), TEST_DOCUMENT_SCHEMA_A.clone()]),
            ("bar", 0u16).into() =>
                    Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone(), TEST_DOCUMENT_SCHEMA_C.clone()]),
        },
        min_size: 1,
        max_size: Some(4),
    }),
    input = Stage::Set(Set {
        operation: SetOperation::Union,
        left: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_a(), test_document_a()],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        right: Box::new(Stage::Array(ArraySource {
            array: vec![test_document_b(), test_document_c(),],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    }),
);

test_schema!(
    set_union_of_two_empty_sets_has_min_and_max_size_0,
    expected = Ok(ResultSet {
        schema_env: map! {
            ("foo", 0u16).into() =>
                    Schema::AnyOf(set![]),
            ("bar", 0u16).into() =>
                    Schema::AnyOf(set![]),
        },
        min_size: 0,
        max_size: Some(0),
    }),
    input = Stage::Set(Set {
        operation: SetOperation::Union,
        left: Box::new(Stage::Array(ArraySource {
            array: vec![],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        })),
        right: Box::new(Stage::Array(ArraySource {
            array: vec![],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    }),
);
