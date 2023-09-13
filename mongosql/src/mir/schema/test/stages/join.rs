use crate::{
    catalog::Namespace,
    map,
    mir::{
        schema::{
            test::{
                test_document_a, test_document_b, test_document_c, TEST_DOCUMENT_SCHEMA_A,
                TEST_DOCUMENT_SCHEMA_B, TEST_DOCUMENT_SCHEMA_C, TEST_DOCUMENT_SCHEMA_S,
            },
            Error as mir_error, SchemaCache,
        },
        *,
    },
    schema::{Atomic, Document, ResultSet, Schema, BOOLEAN_OR_NULLISH},
    set, test_schema, unchecked_unique_linked_hash_map,
    util::{mir_collection, mir_field_access},
};
use mongosql_datastructures::binding_tuple::DatasourceName::Bottom;

mod equijoin {
    use super::*;

    test_schema!(
        left_equijoin,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::AnyOf(set![
                            Schema::Missing,
                            TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
            },
            min_size: 1,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Left,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            from: mir_collection("bar"),
            local_field: mir_field_access("foo", "a"),
            foreign_field: mir_field_access("bar", "a"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );

    test_schema!(
        inner_equijoin,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => TEST_DOCUMENT_SCHEMA_A.clone(),
                ("foo", 0u16).into() => Schema::AnyOf(set![ TEST_DOCUMENT_SCHEMA_A.clone() ]),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            from: mir_collection("bar"),
            local_field: mir_field_access("foo", "a"),
            foreign_field: mir_field_access("bar", "a"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );

    test_schema!(
        equijoin_fields_not_comparable,
        expected_error_code = 1005,
        expected = Err(schema::Error::InvalidComparison(
            "equijoin comparison",
            Schema::AnyOf(set![Schema::Atomic(Atomic::Integer),]),
            Schema::Atomic(Atomic::String),
        )),
        input = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            from: mir_collection("bar"),
            local_field: mir_field_access("foo", "a"),
            foreign_field: mir_field_access("bar", "s"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_S.clone(),
        }),
    );
}

mod lateral {
    use super::*;

    test_schema!(
        left_lateral_join,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::AnyOf(set![
                            Schema::Missing,
                            TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
            },
            min_size: 1,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            subquery: mir_collection("bar"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );

    test_schema!(
        inner_lateral_join,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => TEST_DOCUMENT_SCHEMA_A.clone(),
                ("foo", 0u16).into() => Schema::AnyOf(set![ TEST_DOCUMENT_SCHEMA_A.clone() ]),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            subquery: mir_collection("bar"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );
}

mod standard {
    use super::*;

    test_schema!(
        left_join,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::AnyOf(set![
                                TEST_DOCUMENT_SCHEMA_B.clone()
                            ]
                        ),
                    ]
                ),
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
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
            condition: Some(Expression::Literal(LiteralValue::Boolean(false).into())),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        cross_join,
        expected_pat = Ok(ResultSet {
            min_size: 6,
            max_size: Some(6),
            ..
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
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
            right: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(6).into())
                        }
                        .into()
                    ),
                ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: None,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        inner_join,
        expected_pat = Ok(ResultSet {
            min_size: 0,
            max_size: Some(6),
            ..
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
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
            right: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(6).into())
                        }
                        .into()
                    ),
                ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::Literal(LiteralValue::Boolean(false).into())),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        inner_and_left_join,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone(),
                    ]
                ),
                ("bar", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_B.clone(),
                    ]
                ),
                ("car", 0u16).into() => Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_C.clone()]),
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Join(Join {
                join_type: JoinType::Left,
                left: Box::new(Stage::Array(ArraySource {
                    array: vec![test_document_b()],
                    alias: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                right: Box::new(Stage::Array(ArraySource {
                    array: vec![test_document_c()],
                    alias: "car".into(),
                    cache: SchemaCache::new(),
                })),
                condition: Some(Expression::Literal(LiteralValue::Boolean(false).into())),
                cache: SchemaCache::new(),
            })),
            condition: None,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        join_duplicate_datasource_names,
        expected_error_code = 1009,
        expected = Err(mir_error::DuplicateKey(("foo", 0u16).into())),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
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
            condition: None,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        invalid_join_condition,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "join condition",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
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
            condition: Some(Expression::Literal(LiteralValue::Integer(5).into())),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        join_condition_uses_left_datasource,
        expected_pat = Ok(ResultSet { .. }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Boolean(true).into())
                    }
                    .into()
                )],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".to_string(),
                    cache: SchemaCache::new(),
                })),
                target_type: Type::Boolean,
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        join_condition_uses_right_datasource,
        expected_pat = Ok(ResultSet { .. }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "b".into() => Expression::Literal(LiteralValue::Boolean(true).into())
                    }
                    .into()
                )],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                    field: "b".to_string(),
                    cache: SchemaCache::new(),
                })),
                target_type: Type::Boolean,
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        join_condition_uses_correlated_datasource,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Left,
                    left: Box::new(Stage::Array(ArraySource {
                        array: vec![test_document_b()],
                        alias: "bar".into(),
                        cache: SchemaCache::new(),
                    })),
                    right: Box::new(Stage::Array(ArraySource {
                        array: vec![test_document_c()],
                        alias: "car".into(),
                        cache: SchemaCache::new(),
                    })),
                    condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                        expr: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".to_string(),
                            cache: SchemaCache::new(),
                        })),
                        target_type: Type::Boolean,
                        cache: SchemaCache::new(),
                    })),
                    cache: SchemaCache::new(),
                }),),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Boolean)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
    );
}
