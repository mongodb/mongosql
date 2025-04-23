use crate::mir;

#[test]
fn select_and_order_by_column_not_in_select() {
    use crate::{
        algebrizer::{Algebrizer, ClauseType},
        ast,
        catalog::Catalog,
        map,
        mir::{
            binding_tuple::Key, schema::SchemaCache, Collection, Expression, FieldAccess, Project,
            Sort, Stage,
        },
        schema, set, unchecked_unique_linked_hash_map, SchemaCheckingMode,
    };
    use agg_ast::definitions::Namespace;
    let select = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
            ast::Expression::Document(vec![
                ast::DocumentPair {
                    key: "_id".into(),
                    value: ast::Expression::Identifier("_id".into()),
                },
                ast::DocumentPair {
                    key: "a".into(),
                    value: ast::Expression::Identifier("a".into()),
                },
                ast::DocumentPair {
                    key: "c".into(),
                    value: ast::Expression::Binary(ast::BinaryExpr {
                        left: ast::Expression::Identifier("b".into()).into(),
                        op: ast::BinaryOp::Add,
                        right: ast::Expression::Literal(ast::Literal::Integer(42)).into(),
                    }),
                },
            ]),
        )]),
    };

    let order_by = Some(ast::OrderByClause {
        sort_specs: vec![
            ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Identifier("b".into())),
                direction: ast::SortDirection::Asc,
            },
            ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Identifier("c".into())),
                direction: ast::SortDirection::Asc,
            },
        ],
    });

    let source = Stage::Project(Project {
        source: Stage::Collection(Collection {
            db: "testquerydb".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })
        .into(),
        expression: map! {
            ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into()),
        },
        is_add_fields: false,
        cache: SchemaCache::new(),
    });

    let expected = Ok(Stage::Project(Project {
        source: Stage::Sort(Sort {
            source: Stage::Project(Project {
                source: Stage::Project(Project {
                    source: Stage::Collection(Collection {
                        db: "testquerydb".into(),
                        collection: "foo".into(),
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    expression: map! {
                        ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into()),
                    },
                    is_add_fields: false,
                    cache: SchemaCache::new(),
                })
                .into(),
                expression: map! {
                    Key::bot(0) => Expression::Document(unchecked_unique_linked_hash_map! {
                        "_id".into() => Expression::FieldAccess(FieldAccess {
                            expr: Expression::Reference(("foo", 0u16).into()).into(),
                            field: "_id".into(),
                            is_nullable: false,
                        }),
                        "a".into() => Expression::FieldAccess(FieldAccess {
                            expr: Expression::Reference(("foo", 0u16).into()).into(),
                            field: "a".into(),
                            is_nullable: false,
                        }),
                        "c".into() => Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Add,
                            args: vec![
                                Expression::FieldAccess(FieldAccess {
                                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                                    field: "b".into(),
                                    is_nullable: false,
                                }),
                                Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],
                            is_nullable: false,
                        }),
                    }.into())
                },
                is_add_fields: true,
                cache: SchemaCache::new(),
            })
            .into(),
            specs: vec![
                mir::SortSpecification::Asc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["b".into()],
                    is_nullable: false,
                }),
                mir::SortSpecification::Asc(mir::FieldPath {
                    key: Key::bot(0u16),
                    fields: vec!["c".into()],
                    is_nullable: false,
                }),
            ],
            cache: SchemaCache::new(),
        })
        .into(),
        expression: map! {
            Key::bot(0) => Expression::Document(unchecked_unique_linked_hash_map! {
                "_id".into() => Expression::FieldAccess(FieldAccess {
                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                    field: "_id".into(),
                    is_nullable: false,
                }),
                "a".into() => Expression::FieldAccess(FieldAccess {
                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                    field: "a".into(),
                    is_nullable: false,
                }),
                "c".into() => Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Add,
                    args: vec![
                        Expression::FieldAccess(FieldAccess {
                            expr: Expression::Reference(("foo", 0u16).into()).into(),
                            field: "b".into(),
                            is_nullable: false,
                        }),
                        Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false,
                }),
            }.into())
        },
        is_add_fields: false,
        cache: SchemaCache::new(),
    }));

    let catalog = vec![("testquerydb", "foo")]
        .into_iter()
        .map(|(db, c)| {
            (
                Namespace {
                    database: db.into(),
                    collection: c.into(),
                },
                schema::Schema::Document(schema::Document {
                    keys: map! {
                        "_id".into() => schema::Schema::Atomic(schema::Atomic::ObjectId),
                        "a".into() => schema::Schema::Atomic(schema::Atomic::Integer),
                        "b".into() => schema::Schema::Atomic(schema::Atomic::Integer),
                    },
                    required: set! {"_id".into(), "a".into(), "b".into()},
                    additional_properties: false,
                    jaccard_index: None,
                }),
            )
        })
        .collect::<Catalog>();

    let algebrizer = Algebrizer::new(
        "test",
        &catalog,
        0u16,
        SchemaCheckingMode::Strict,
        false,
        ClauseType::Unintialized,
    );
    assert_eq!(
        expected,
        algebrizer.algebrize_select_and_order_by_clause(select, order_by, source,),
    );
}
