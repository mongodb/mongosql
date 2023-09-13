mod datasource; // mir::Stage::{Array, Collection}
mod filter; // mir::Stage::Filter
mod group; // mir::Stage::Group
mod join; // mir::Stage::{Join, MqlIntrinsic::{EquiJoin, LateralJoin}}
mod set; // mir::Stage::Set
mod sort; // mir::Stage::Sort
mod unwind; // mir::Stage::Unwind

mod limit {
    use crate::{
        catalog::Namespace,
        map,
        mir::{schema::SchemaCache, *},
        schema::{ResultSet, ANY_DOCUMENT},
        test_schema, unchecked_unique_linked_hash_map,
    };

    test_schema!(
        limit_collection_datasource,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: Some(20),
        }),
        input = Stage::Limit(Limit {
            limit: 20,
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    test_schema!(
        limit_lt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 2,
            max_size: Some(2),
            ..
        }),
        input = Stage::Limit(Limit {
            limit: 2,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        limit_gt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 3,
            max_size: Some(3),
            ..
        }),
        input = Stage::Limit(Limit {
            limit: 10,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
}

mod project {
    use crate::{
        catalog::Namespace,
        map,
        mir::{schema::SchemaCache, *},
        schema::{ResultSet, ANY_DOCUMENT},
        test_schema,
    };

    test_schema!(
        project_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar1", 0u16).into() => ANY_DOCUMENT.clone(),
                ("bar2", 0u16).into() => ANY_DOCUMENT.clone(),
                ("bar3", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test2".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar1", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
                ("bar2", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
                ("bar3", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test2".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );
}

mod offset {
    use crate::{
        catalog::Namespace,
        map,
        mir::{schema::SchemaCache, *},
        schema::{ResultSet, ANY_DOCUMENT},
        test_schema, unchecked_unique_linked_hash_map,
    };

    test_schema!(
        offset_collection_datasource,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Offset(Offset {
            offset: 20,
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    test_schema!(
        offset_lt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 2,
            max_size: Some(2),
            ..
        }),
        input = Stage::Offset(Offset {
            offset: 1,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        offset_gt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 0,
            max_size: Some(0),
            ..
        }),
        input = Stage::Offset(Offset {
            offset: 10,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
}
