use crate::{
    catalog::Namespace,
    map,
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::{Atomic, Document, ResultSet, Schema, ANY_DOCUMENT},
    set, test_schema, unchecked_unique_linked_hash_map,
};

mod collection {
    use super::*;

    test_schema!(
        collection_schema_no_catalog,
        expected_error_code = 1016,
        expected = Err(mir_error::CollectionNotFound("test2".into(), "foo".into())),
        input = Stage::Collection(Collection {
            db: "test2".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        collection_schema_namespace_not_in_catalog,
        expected_error_code = 1016,
        expected = Err(mir_error::CollectionNotFound("foo".into(), "baz".into())),
        input = Stage::Collection(Collection {
            db: "foo".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "foo".into(), collection: "bar".into()} => Schema::Atomic(Atomic::Integer)
        }),
    );

    test_schema!(
        collection_schema_namespace_in_catalog,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Integer),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Collection(Collection {
            db: "foo".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "foo".into(), collection: "bar".into()} => Schema::Atomic(Atomic::Integer)
        }),
    );
}

mod array {
    use super::*;

    test_schema!(
        empty_array_datasource_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![])
            },
            min_size: 0,
            max_size: Some(0),
        }),
        input = Stage::Array(ArraySource {
            array: vec![],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        dual_array_datasource_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        Schema::Document( Document {
                            keys: map!{},
                            required: set!{},
                            additional_properties: false,
                        })
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        input = Stage::Array(ArraySource {
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {}.into()
            )],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        literal_array_items_datasource_schema,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ])
        }),
        input = Stage::Array(ArraySource {
            array: vec![
                Expression::Literal(LiteralValue::Integer(42)),
                Expression::Literal(LiteralValue::Double(42f64))
            ],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        single_document_array_datasource_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        Schema::Document( Document {
                            keys: map!{"bar".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"bar".into()},
                            additional_properties: false,
                        })
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        input = Stage::Array(ArraySource {
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "bar".into() => Expression::Literal(LiteralValue::Integer(1))
                }
                .into()
            )],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        two_document_array_datasource_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        Schema::Document( Document {
                            keys: map!{"bar".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"bar".into()},
                            additional_properties: false,
                        }),
                        Schema::Document( Document {
                            keys: map!{"car".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"car".into()},
                            additional_properties: false,
                        }),
                    ]
                ),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        input = Stage::Array(ArraySource {
            array: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {
                    "bar".into() => Expression::Literal(LiteralValue::Integer(1))
                    }
                    .into()
                ),
                Expression::Document(
                    unchecked_unique_linked_hash_map! {
                    "car".into() => Expression::Literal(LiteralValue::Integer(1))
                    }
                    .into()
                )
            ],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );
}
