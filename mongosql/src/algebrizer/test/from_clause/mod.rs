use super::{catalog, mir_source_bar, mir_source_foo, AST_SOURCE_BAR, AST_SOURCE_FOO};
use crate::{
    ast::{self, JoinSource},
    catalog::Catalog,
    map,
    mir::{self, binding_tuple::Key, schema::SchemaCache, JoinType},
    multimap,
    schema::{Atomic, Document, Schema, ANY_DOCUMENT},
    set, unchecked_unique_linked_hash_map,
    usererror::UserError,
};
use agg_ast::definitions::Namespace;

fn mir_array_source() -> mir::Stage {
    mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {"a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map!{
                        "b".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(5),)
                    },
                })},
            })],
            alias: "arr".to_string(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("arr", 0u16).into() => mir::Expression::Reference(("arr", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })
}

fn ast_array_source() -> ast::Datasource {
    ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Document(
                        multimap!{"b".into() => ast::Expression::Literal(ast::Literal::Integer(5))},
                    ),
        })],
        alias: "arr".to_string(),
    })
}

test_algebrize!(
    basic_collection,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() =>
                mir::Expression::Reference(("foo", 0u16).into())
        },
        cache: SchemaCache::new(),
    },),),
    input = Some(ast::Datasource::Collection(ast::CollectionSource {
        database: None,
        collection: "foo".into(),
        alias: Some("bar".into()),
    })),
    catalog = catalog(vec![("test", "foo")]),
);
test_algebrize!(
    qualified_collection,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: "test2".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() =>
                mir::Expression::Reference(("foo", 0u16).into())
        },
        cache: SchemaCache::new(),
    }),),
    input = Some(ast::Datasource::Collection(ast::CollectionSource {
        database: Some("test2".into()),
        collection: "foo".into(),
        alias: Some("bar".into()),
    })),
    catalog = catalog(vec![("test2", "foo")]),
);
test_algebrize!(
    collection_and_alias_contains_dot,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: "test2".into(),
            collection: "foo.bar".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("foo.bar", 0u16).into() =>
                mir::Expression::Reference(("foo.bar", 0u16).into())
        },
        cache: SchemaCache::new(),
    }),),
    input = Some(ast::Datasource::Collection(ast::CollectionSource {
        database: Some("test2".into()),
        collection: "foo.bar".into(),
        alias: Some("foo.bar".into()),
    })),
    catalog = catalog(vec![("test2", "foo.bar")]),
);
test_algebrize!(
    collection_and_alias_starts_with_dollar,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: "test2".into(),
            collection: "$foo".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("$foo", 0u16).into() =>
                mir::Expression::Reference(("$foo", 0u16).into())
        },
        cache: SchemaCache::new(),
    }),),
    input = Some(ast::Datasource::Collection(ast::CollectionSource {
        database: Some("test2".into()),
        collection: "$foo".into(),
        alias: Some("$foo".into()),
    })),
    catalog = catalog(vec![("test2", "$foo")]),
);
test_algebrize!(
    empty_array,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Array(mir::ArraySource {
            array: vec![],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    dual,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Document(
                unchecked_unique_linked_hash_map! {}.into(),
            )],
            alias: "_dual".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("_dual", 0u16).into() => mir::Expression::Reference(("_dual", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Document(multimap! {},)],
        alias: "_dual".into(),
    })),
);
test_algebrize!(
    int_array,
    method = algebrize_from_clause,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "array datasource items",
        required: ANY_DOCUMENT.clone().into(),
        found: Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]).into(),
    })),
    expected_error_code = 1002,
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Literal(ast::Literal::Integer(42))],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    null_array,
    method = algebrize_from_clause,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "array datasource items",
        required: ANY_DOCUMENT.clone().into(),
        found: Schema::AnyOf(set![Schema::Atomic(Atomic::Null)]).into(),
    })),
    expected_error_code = 1002,
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Literal(ast::Literal::Null)],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    array_datasource_must_be_literal,
    method = algebrize_from_clause,
    expected = Err(Error::ArrayDatasourceMustBeLiteral),
    expected_error_code = 3004,
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Document(multimap! {
            "foo".into() => ast::Expression::Identifier("foo".into()),
            "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
        },)],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    single_document_array,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                }
                .into(),
            )],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Document(multimap! {
            "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
        },)],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    two_document_array,
    method = algebrize_from_clause,
    expected =
        Ok(mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![
                mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                }.into()),
                mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(41)),
                    "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42))
                }.into())
            ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![
            ast::Expression::Document(multimap! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            }),
            ast::Expression::Document(multimap! {
                "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(41)),
                "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            },)
        ],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    two_document_with_nested_document_array,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
        source: Box::new(mir::Stage::Array(mir::ArraySource {
            array: vec![
                mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                }.into()),
                mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo2".into() => mir::Expression::Document(
                        unchecked_unique_linked_hash_map!{"nested".into() => mir::Expression::Literal(mir::LiteralValue::Integer(52))}
                    .into()),
                    "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42))
                }.into())
            ],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Array(ast::ArraySource {
        array: vec![
            ast::Expression::Document(multimap! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            }),
            ast::Expression::Document(multimap! {
                "foo2".into() => ast::Expression::Document(
                    multimap!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                ),
                "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            },)
        ],
        alias: "bar".into(),
    })),
);
test_algebrize!(
    left_join,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Join(mir::Join {
        join_type: JoinType::Left,
        left: Box::new(mir_source_foo()),
        right: Box::new(mir_source_bar()),
        condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Left,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: Some(ast::Expression::Literal(ast::Literal::Boolean(true)))
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    right_join,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Join(mir::Join {
        join_type: JoinType::Left,
        left: Box::new(mir_source_bar()),
        right: Box::new(mir_source_foo()),
        condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Right,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: Some(ast::Expression::Literal(ast::Literal::Boolean(true)))
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    left_outer_join_without_condition,
    method = algebrize_from_clause,
    expected = Err(Error::NoOuterJoinCondition),
    expected_error_code = 3019,
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Left,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: None
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    right_outer_join_without_condition,
    method = algebrize_from_clause,
    expected = Err(Error::NoOuterJoinCondition),
    expected_error_code = 3019,
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Right,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: None
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    inner_join,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Join(mir::Join {
        join_type: JoinType::Inner,
        left: Box::new(mir_source_foo()),
        right: Box::new(mir_source_bar()),
        condition: None,
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Inner,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: None
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    cross_join,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Join(mir::Join {
        join_type: JoinType::Inner,
        left: Box::new(mir_source_foo()),
        right: Box::new(mir_source_bar()),
        condition: None,
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Cross,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: None
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    join_on_one_true,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Filter(mir::Filter {
        source: Box::new(mir::Stage::Join(mir::Join {
            join_type: JoinType::Inner,
            left: Box::new(mir_source_foo()),
            right: Box::new(mir_source_bar()),
            condition: None,
            cache: SchemaCache::new(),
        })),
        condition: mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Cross,
        left: Box::new(AST_SOURCE_FOO.clone()),
        right: Box::new(AST_SOURCE_BAR.clone()),
        condition: Some(ast::Expression::Literal(ast::Literal::Integer(1)))
    })),
    catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
);
test_algebrize!(
    invalid_join_condition,
    method = algebrize_from_clause,
    expected = Err(Error::FieldNotFound(
        "x".into(),
        Some(vec!["bar".into(), "foo".into()]),
        ClauseType::From,
        0u16,
    )),
    expected_error_code = 3008,
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Cross,
        left: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            })],
            alias: "foo".into(),
        })),
        right: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            })],
            alias: "bar".into(),
        })),
        condition: Some(ast::Expression::Identifier("x".into())),
    })),
);
test_algebrize!(
    join_condition_must_have_boolean_schema,
    method = algebrize_from_clause,
    expected_pat = Err(Error::SchemaChecking(_)),
    expected_error_code = 1002,
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Cross,
        left: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            })],
            alias: "foo".into(),
        })),
        right: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            })],
            alias: "bar".into(),
        })),
        condition: Some(ast::Expression::Literal(ast::Literal::Integer(42))),
    })),
);
test_algebrize!(
    derived_single_datasource,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Derived(mir::Derived {
        source: Box::new(mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::Project(mir::Project {
                is_add_fields: false,
                source: Box::new(mir::Stage::Array(mir::ArraySource {
                    array: vec![mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        }
                        .into()
                    )],
                    alias: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("bar", 1u16).into() => mir::Expression::Reference(("bar", 1u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("d", 0u16).into() => mir::Expression::Reference(("bar", 1u16).into())
            },
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Derived(ast::DerivedSource {
        query: Box::new(ast::Query::Select(Box::new(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                },)],
                alias: "bar".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }))),
        alias: "d".into(),
    })),
);
test_algebrize!(
    derived_multiple_datasources,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Derived(mir::Derived {
        source: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
            source: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
                source: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
                    source: Box::new(mir::Stage::Array(mir::ArraySource {
                        array: vec![mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                                "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            }
                        .into())],
                        alias: "bar".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        ("bar", 1u16).into() => mir::Expression::Reference(("bar", 1u16).into()),
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    Key::bot(1u16) => mir::Expression::Document(
                        unchecked_unique_linked_hash_map!{"baz".into() => mir::Expression::Literal(mir::LiteralValue::String("hello".into()))}
                    .into()),
                    ("bar", 1u16).into() =>
                        mir::Expression::Reference(("bar", 1u16).into())
                },
                cache: SchemaCache::new(),
            })),
            expression: map! { ("d", 0u16).into() =>
                mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::MergeObjects,
                        args:
                            vec![
                                mir::Expression::Reference(Key::bot(1u16).into()),
                                mir::Expression::Reference(("bar", 1u16).into()),
                            ],
                        is_nullable: false,
                    }
                )
            },
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Derived(ast::DerivedSource {
        query: Box::new(ast::Query::Select(Box::new(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![
                    ast::SelectValuesExpression::Substar("bar".into()),
                    ast::SelectValuesExpression::Expression(ast::Expression::Document(
                        multimap! {
                            "baz".into() => ast::Expression::StringConstructor("hello".into())
                        }
                    )),
                ]),
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                },)],
                alias: "bar".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }))),
        alias: "d".into(),
    })),
);
test_algebrize!(
    derived_join_datasources_distinct_keys_succeeds,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Derived(mir::Derived {
        source: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
            source: mir::Stage::Join(mir::Join {
                join_type: mir::JoinType::Inner,
                left: mir::Stage::Project(mir::Project {
                        is_add_fields: false,
                    source: Box::new(mir::Stage::Array(mir::ArraySource {
                        array: vec![mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                            "foo1".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            "bar1".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                                                        }
                        .into())],
                        alias: "bar1".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        ("bar1", 1u16).into() => mir::Expression::Reference(("bar1", 1u16).into()),
                    },
                    cache: SchemaCache::new(),
                })
                .into(),
                right: mir::Stage::Project(mir::Project {
                        is_add_fields: false,
                    source: Box::new(mir::Stage::Array(mir::ArraySource {
                        array: vec![mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                            "foo2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                                                        }
                        .into())],
                        alias: "bar2".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        ("bar2", 1u16).into() => mir::Expression::Reference(("bar2", 1u16).into()),
                    },
                    cache: SchemaCache::new(),
                })
                .into(),
                condition: None,
                cache: SchemaCache::new(),
            })
            .into(),
            expression: map! {("d", 0u16).into() =>
                mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::MergeObjects,
                        args: vec![mir::Expression::Reference(("bar1", 1u16).into()),
                                   mir::Expression::Reference(("bar2", 1u16).into())],
                        is_nullable: false,
                    }
                )
            },
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Derived(ast::DerivedSource {
        query: Box::new(ast::Query::Select(Box::new(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star,]),
            },
            from_clause: Some(ast::Datasource::Join(JoinSource {
                join_type: ast::JoinType::Inner,
                left: ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "foo1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        "bar1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    },)],
                    alias: "bar1".into(),
                })
                .into(),
                right: ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    },)],
                    alias: "bar2".into(),
                })
                .into(),
                condition: None,
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }))),
        alias: "d".into(),
    })),
);
test_algebrize!(
    join_condition_referencing_non_correlated_fields,
    method = algebrize_from_clause,
    expected = Ok(
        mir::Stage::Join(mir::Join {
        join_type: mir::JoinType::Left,
        left: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1))}
                .into())],
                alias: "foo".to_string(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        right: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {"b".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(4))}
                .into())],
                alias: "bar".to_string(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        condition: Some(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Eq,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                        field: "a".to_string(),
                        is_nullable: false,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("bar", 0u16).into())),
                        field: "b".to_string(),
                        is_nullable: false,
                    })
                ],
                is_nullable: false,
            })
        ),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Join(JoinSource {
        join_type: ast::JoinType::Left,
        left: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "foo".into(),
        })),
        right: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(4)),
            },)],
            alias: "bar".into(),
        })),
        condition: Some(ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("a".to_string())),
            op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
            right: Box::new(ast::Expression::Identifier("b".to_string())),
        }))
    })),
);
test_algebrize!(
    derived_join_datasources_overlapped_keys_fails,
    method = algebrize_from_clause,
    expected = Err(Error::DerivedDatasourceOverlappingKeys(
        Schema::Document(Document {
            keys: map! {
                "bar".into() => Schema::Atomic(Atomic::Integer),
                "foo1".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "bar".into(),
                "foo1".into(),
            },
            additional_properties: false,
            ..Default::default()
        })
        .into(),
        Schema::Document(Document {
            keys: map! {
                "bar".into() => Schema::Atomic(Atomic::Integer),
                "foo2".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "bar".into(),
                "foo2".into(),
            },
            additional_properties: false,
            ..Default::default()
        })
        .into(),
        "d".into(),
        crate::schema::Satisfaction::Must,
    )),
    expected_error_code = 3016,
    input = Some(ast::Datasource::Derived(ast::DerivedSource {
        query: Box::new(ast::Query::Select(Box::new(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star,]),
            },
            from_clause: Some(ast::Datasource::Join(JoinSource {
                join_type: ast::JoinType::Inner,
                left: ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "foo1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    },)],
                    alias: "bar1".into(),
                })
                .into(),
                right: ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    },)],
                    alias: "bar2".into(),
                })
                .into(),
                condition: None,
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }))),
        alias: "d".into(),
    })),
);
test_algebrize!(
    flatten_simple,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir_array_source()),
        expression: map! {
            ("arr", 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
            document: unchecked_unique_linked_hash_map!{
                "a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: ("arr", 0u16).into(),
                        })),
                        field: "a".to_string(),
                        is_nullable: false,
                    })),
                    field: "b".to_string(),
                    is_nullable: false,
                })},
        })},
        cache: SchemaCache::new()
    })),
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(ast_array_source()),
        options: vec![]
    })),
);
test_algebrize!(
    flatten_array_source_multiple_docs,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map!{
                                "b".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(5),)
                            },
                        }),
                        "x".to_string() => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "y".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(8),)
                            },
                        })
                    },
                })],
                alias: "arr".to_string(),
                cache: SchemaCache::new()
            })),
            expression: map! {
                ("arr", 0u16).into() => mir::Expression::Reference(("arr", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("arr", 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {
                    "a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("arr", 0u16).into(),
                            })),
                            field: "a".to_string(),
                            is_nullable: false,
                        })),
                        field: "b".to_string(),
                        is_nullable: false,

                    }),
                    "x_y".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("arr", 0u16).into(),
                            })),
                            field: "x".to_string(),
                            is_nullable: false,
                        })),
                        field: "y".to_string(),
                        is_nullable: false,
                    })
                },
            })
        },
        cache: SchemaCache::new()
    })),
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Document(
                            multimap!{"b".into() => ast::Expression::Literal(ast::Literal::Integer(5))},
                        ),
                "x".into() => ast::Expression::Document(
                            multimap!{"y".into() => ast::Expression::Literal(ast::Literal::Integer(8))},
                        ),
            })],
            alias: "arr".to_string()
        })),
        options: vec![]
    })),
);
test_algebrize!(
    flatten_duplicate_options,
    method = algebrize_from_clause,
    expected = Err(Error::DuplicateFlattenOption(ast::FlattenOption::Depth(2))),
    expected_error_code = 3024,
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(ast_array_source()),
        options: vec![ast::FlattenOption::Depth(1), ast::FlattenOption::Depth(2)]
    })),
);
test_algebrize!(
    flatten_polymorphic_non_document_schema_array_source,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![
                    mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map!{
                                "b".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map!{
                                        "c".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(5),)},
                                })},
                        })},
                    }),
                    mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map!{
                                "b".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map!{
                                        "c".to_string() => mir::Expression::Literal(mir::LiteralValue::String("hello".to_string()),)},
                                })},
                        })},
                    })
                ],
                alias: "arr".to_string(),
                cache: SchemaCache::new()
            })),
            expression: map! {
                ("arr", 0u16).into() => mir::Expression::Reference(("arr", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        expression: map! {
        ("arr", 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
            document: unchecked_unique_linked_hash_map!{
                "a_b_c".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("arr", 0u16).into(),
                            })),
                            field: "a".to_string(),
                            is_nullable: false,
                        })),
                        field: "b".to_string(),
                        is_nullable: false,
                    })),
                    field: "c".to_string(),
                    is_nullable: false,
                })},
        })},
        cache: SchemaCache::new()
    })),
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Document(
                    multimap!{"b".into() => ast::Expression::Document(
                    multimap!{"c".into() => ast::Expression::Literal(ast::Literal::Integer(5))},
                )},
                )}),
                ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Document(
                    multimap!{"b".into() => ast::Expression::Document(
                    multimap!{"c".into() => ast::Expression::StringConstructor("hello".to_string())},
                )},
                )}),
            ],
            alias: "arr".to_string()
        })),
        options: vec![]
    })),
);
test_algebrize!(
    flatten_polymorphic_object_schema_array_source,
    method = algebrize_from_clause,
    expected = Err(Error::PolymorphicObjectSchema("a".to_string())),
    expected_error_code = 3026,
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(5))}),
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Document(
                        multimap!{"b".into() => ast::Expression::Literal(ast::Literal::Integer(6))},
                    )
                }),
            ],
            alias: "arr".to_string()
        })),
        options: vec![]
    })),
);
test_algebrize!(
    flattening_polymorphic_objects_other_than_just_null_or_missing_polymorphism_causes_error,
    method = algebrize_from_clause,
    expected = Err(Error::PolymorphicObjectSchema("a".to_string())),
    expected_error_code = 3026,
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![]
    })),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set!{
                    Schema::Document(Document {
                        keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                        ..Default::default()
                    }),
                    Schema::Atomic(Atomic::Integer)
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    }),
);

test_algebrize!(
    flattening_polymorphic_objects_with_just_null_polymorphism_works,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir_source_foo()),
        expression: map! {("foo", 0u16).into() => mir::Expression::Document(
            mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {"a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess{
                    expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: ("foo", 0u16).into(),
                        })),
                        field: "a".to_string(),
                        is_nullable: true,
                    })),

                    field: "b".to_string(),
                    is_nullable: true,
                }
            )}},
        )},
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![]
    })),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set!{
                    Schema::Document(Document {
                        keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                        ..Default::default()
                    }),
                    Schema::Atomic(Atomic::Null),
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    }),
);

test_algebrize!(
    flattening_polymorphic_objects_with_just_missing_polymorphism_works,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir_source_foo()),
        expression: map! {("foo", 0u16).into() => mir::Expression::Document(
            mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {"a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess{
                    expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: ("foo", 0u16).into(),
                        })),
                        field: "a".to_string(),
                        is_nullable: true,
                    })),

                    field: "b".to_string(),
                    is_nullable: true,
                }
            )}},
        )},
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![]
    })),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set!{
                    Schema::Document(Document {
                        keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                        ..Default::default()
                    }),
                    Schema::Missing,
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    }),
);

test_algebrize!(
    flattening_polymorphic_objects_with_just_null_and_missing_polymorphism_works,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir_source_foo()),
        expression: map! {("foo", 0u16).into() => mir::Expression::Document(
            mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {"a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess{
                    expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: ("foo", 0u16).into(),
                        })),
                        field: "a".to_string(),
                        is_nullable: true,
                    })),

                    field: "b".to_string(),
                    is_nullable: true,
                }
            )}},
        )},
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Datasource::Flatten(ast::FlattenSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![]
    })),
    catalog = Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set!{
                    Schema::Document(Document {
                        keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                        ..Default::default()
                    }),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing,
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    }),
);
#[cfg(test)]
mod unwind;
