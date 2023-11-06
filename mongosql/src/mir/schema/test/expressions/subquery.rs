use crate::{
    catalog::Namespace,
    map,
    mir::{
        binding_tuple::DatasourceName::Bottom,
        schema::{Atomic, Document, Error as mir_error, SchemaCache},
        ArraySource, Collection, Expression, FieldAccess, LiteralValue, Project, ScalarFunction,
        ScalarFunctionApplication, Stage, SubqueryComparison, SubqueryComparisonOp, SubqueryExpr,
        SubqueryModifier,
    },
    schema::{Schema, ANY_DOCUMENT},
    set, test_schema, unchecked_unique_linked_hash_map,
};

mod exists {
    use super::*;

    test_schema!(
        exists_uncorrelated,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Exists(
            Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            }))
            .into()
        ),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    test_schema!(
        exists_correlated,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Exists(
            Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            }))
            .into()
        ),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    test_schema!(
        exists_invalid_expression,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Div",
            required: 2,
            found: 3
        }),
        input = Expression::Exists(
            Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("a", 0u16).into() =>
                        Expression::ScalarFunction(ScalarFunctionApplication::new(ScalarFunction::Div,vec![
                                Expression::Literal(LiteralValue::Integer(1)),
                                Expression::Literal(LiteralValue::Integer(2)),
                                Expression::Literal(LiteralValue::Integer(3))
                            ],))
                },
                cache: SchemaCache::new(),
            }))
            .into()
        ),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );
}

mod subquery_expr {
    use super::*;

    test_schema!(
        subquery_output_expr_violates_type_constraints,
        expected_error_code = 1007,
        expected = Err(mir_error::AccessMissingField(
            "_2".into(),
            Some(vec!["_1".to_string()])
        )),
        input = Expression::Subquery(SubqueryExpr::new(
            Box::new(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference((Bottom, 1u16).into())),
                "_2".into(),
            ))),
            Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "_1".into() => Expression::Literal(LiteralValue::Integer(5))
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
        )),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Analogous SQL query: SELECT (SELECT foo.a FROM []) FROM foo
    test_schema!(
        correlated_subquery,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Missing
        ])),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference((Bottom, 1u16).into())),
                "a".into(),
            ))),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![],
                    alias: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
    );

    test_schema!(
        subquery_output_expr_correlated_datasource,
        expected_error_code = 1000,
        expected = Err(mir_error::DatasourceNotFoundInSchemaEnv(
            ("foo", 0u16).into()
        )),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference(("foo", 0u16).into())),
                "a".into(),
            ))),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [] AS foo)"
    test_schema!(
        uncorrelated_subquery_cardinality_is_zero,
        expected = Ok(Schema::AnyOf(set![Schema::AnyOf(set![]), Schema::Missing])),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_must_be_one,
        expected = Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)])),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference(("foo", 1u16).into())),
                "a".into(),
            ))),
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(5))
                    }
                    .into()
                ),],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM foo)"
    test_schema!(
        subquery_cardinality_may_be_1,
        expected_error_code = 1008,
        expected = Err(mir_error::InvalidSubqueryCardinality),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}, {'a': 6}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_gt_one,
        expected_error_code = 1008,
        expected = Err(mir_error::InvalidSubqueryCardinality),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                Box::new(Expression::Reference(("foo", 1u16).into())),
                "a".into(),
            ))),
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5))
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(6))
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        }),
    );
}

mod subquery_comparison {
    use super::*;

    test_schema!(
        uncorrelated_subquery_comparison_known_type,
        expected = Ok(Schema::Atomic(Atomic::Boolean).clone()),
        input = Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 1u16).into())),
                    "a".into(),
                ))),
                subquery: Box::new(Stage::Array(ArraySource {
                    array: vec![Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5))
                        }
                        .into()
                    )],
                    alias: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: false,
        }),
    );

    test_schema!(
        uncorrelated_subquery_comparison_possibly_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 1u16).into())),
                    "a".into(),
                ))),
                subquery: Box::new(Stage::Array(ArraySource {
                    array: vec![
                        Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "a".into() => Expression::Literal(LiteralValue::Integer(5))
                            }
                            .into()
                        ),
                        Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "b".into() => Expression::Literal(LiteralValue::Integer(5))
                            }
                            .into()
                        )
                    ],
                    alias: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: true,
        }),
    );

    test_schema!(
        incomparable_argument_and_output_expr,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "subquery comparison",
            Schema::Atomic(Atomic::String),
            Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
        )),
        input = Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(LiteralValue::String("abc".into()))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess::new(
                    Box::new(Expression::Reference(("foo", 1u16).into())),
                    "a".into(),
                ))),
                subquery: Box::new(Stage::Array(ArraySource {
                    array: vec![Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5))
                        }
                        .into()
                    )],
                    alias: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: true,
        }),
    );
}
