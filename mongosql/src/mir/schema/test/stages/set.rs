use crate::{
    map,
    mir::{
        schema::{
            test::{
                test_document_a, test_document_b, TEST_DOCUMENT_SCHEMA_A, TEST_DOCUMENT_SCHEMA_B,
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
