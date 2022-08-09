use crate::{
    ast::{self, CollectionSource, Datasource},
    catalog::{Catalog, Namespace},
    ir::{schema::SchemaCache, Collection, Expression, Project, Stage},
    map,
    schema::ANY_DOCUMENT,
};
use lazy_static::lazy_static;

macro_rules! test_algebrize {
    ($func_name:ident, method = $method:ident, $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? input = $ast:expr, $(source = $source:expr,)? $(env = $env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{
                algebrizer::{Algebrizer, Error},
                catalog::Catalog,
                SchemaCheckingMode,
            };

            #[allow(unused_mut, unused_assignments)]
            let mut catalog = Catalog::default();
            $(catalog = $catalog;)?

            #[allow(unused_mut, unused_assignments)]
            let mut schema_checking_mode = SchemaCheckingMode::Strict;
            $(schema_checking_mode = $schema_checking_mode;)?

            #[allow(unused_mut, unused_assignments)]
            let mut algebrizer = Algebrizer::new("test".into(), &catalog, 0u16, schema_checking_mode);
            $(algebrizer = Algebrizer::with_schema_env("test".into(), $env, &catalog, 1u16, schema_checking_mode);)?

            let res: Result<_, Error> = algebrizer.$method($ast $(, $source)?);
            $(assert!(matches!(res, $expected_pat));)?
            $(assert_eq!(res, $expected);)?
        }
    };
}

fn ir_source_foo() -> Stage {
    Stage::Project(Project {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into())
        },
        cache: SchemaCache::new(),
    })
}

fn ir_source_bar() -> Stage {
    Stage::Project(Project {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() => Expression::Reference(("bar", 0u16).into())
        },
        cache: SchemaCache::new(),
    })
}

fn catalog(ns: Vec<(&str, &str)>) -> Catalog {
    ns.into_iter()
        .map(|(db, c)| {
            (
                Namespace {
                    db: db.into(),
                    collection: c.into(),
                },
                ANY_DOCUMENT.clone(),
            )
        })
        .collect::<Catalog>()
}

lazy_static! {
    static ref AST_SOURCE_FOO: Datasource = Datasource::Collection(CollectionSource {
        database: Some("test".into()),
        collection: "foo".into(),
        alias: Some("foo".into()),
    });
    static ref AST_QUERY_FOO: ast::Query = ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
        },
        from_clause: Some(AST_SOURCE_FOO.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    });
    static ref AST_SOURCE_BAR: Datasource = Datasource::Collection(CollectionSource {
        database: Some("test".into()),
        collection: "bar".into(),
        alias: Some("bar".into()),
    });
    static ref AST_QUERY_BAR: ast::Query = ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
        },
        from_clause: Some(AST_SOURCE_BAR.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    });
}

mod expression {
    use crate::{
        ast,
        ir::{self, binding_tuple::Key, schema::SchemaCache},
        map, multimap,
        schema::{
            Atomic, Document, Schema, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, NUMERIC_OR_NULLISH,
            STRING_OR_NULLISH,
        },
        set, unchecked_unique_linked_hash_map,
    };

    test_algebrize!(
        null,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(ir::LiteralValue::Null.into())),
        input = ast::Expression::Literal(ast::Literal::Null),
    );
    test_algebrize!(
        expr_true,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(
            ir::LiteralValue::Boolean(true).into()
        )),
        input = ast::Expression::Literal(ast::Literal::Boolean(true)),
    );
    test_algebrize!(
        expr_false,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(
            ir::LiteralValue::Boolean(false).into()
        )),
        input = ast::Expression::Literal(ast::Literal::Boolean(false)),
    );
    test_algebrize!(
        string,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(
            ir::LiteralValue::String("hello!".into()).into()
        )),
        input = ast::Expression::Literal(ast::Literal::String("hello!".into())),
    );
    test_algebrize!(
        int,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(
            ir::LiteralValue::Integer(42).into()
        )),
        input = ast::Expression::Literal(ast::Literal::Integer(42)),
    );
    test_algebrize!(
        long,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(ir::LiteralValue::Long(42).into())),
        input = ast::Expression::Literal(ast::Literal::Long(42)),
    );
    test_algebrize!(
        double,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Literal(
            ir::LiteralValue::Double(42f64).into()
        )),
        input = ast::Expression::Literal(ast::Literal::Double(42f64)),
    );

    test_algebrize!(
        empty_array,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Array(vec![].into())),
        input = ast::Expression::Array(vec![]),
    );
    test_algebrize!(
        nested_array,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Array(
            vec![ir::Expression::Array(
                vec![
                    ir::Expression::Literal(ir::LiteralValue::Long(42).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                ]
                .into()
            )]
            .into()
        )),
        input = ast::Expression::Array(vec![ast::Expression::Array(vec![
            ast::Expression::Literal(ast::Literal::Long(42)),
            ast::Expression::Literal(ast::Literal::Integer(42)),
        ])]),
    );

    test_algebrize!(
        empty_document,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Document(
            unchecked_unique_linked_hash_map! {}.into()
        )),
        input = ast::Expression::Document(multimap! {}),
    );
    test_algebrize!(
        nested_document,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "foo2".into() => ir::Expression::Document(
                    unchecked_unique_linked_hash_map!{"nested".into() => ir::Expression::Literal(ir::LiteralValue::Integer(52).into())}
                .into()),
                "bar2".into() => ir::Expression::Literal(ir::LiteralValue::Integer(42).into())
            }
        .into())),
        input = ast::Expression::Document(multimap! {
                    "foo2".into() => ast::Expression::Document(
                        multimap!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                    ),
                    "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
        }),
    );
    test_algebrize!(
        qualified_ref_in_current_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 1u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("foo".into())),
            subpath: "a".into(),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        qualified_ref_in_super_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("foo".into())),
            subpath: "a".into(),
        }),
        env = map! {
            ("bar", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_ref_may_exist_in_current_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 1u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Identifier("a".into()),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_ref_must_exist_in_current_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 1u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Identifier("a".into()),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_ref_may_exist_only_in_super_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Identifier("a".into()),
        env = map! {
            ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_ref_must_exist_in_super_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Identifier("a".into()),
        env = map! {
            ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_ref_must_exist_in_super_scope_bot_source,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(Key::bot(0u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Identifier("a".into()),
        env = map! {
            ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
            Key::bot(0u16) => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_ref_may_and_must_exist_in_two_sources,
        method = algebrize_expression,
        expected = Err(Error::AmbiguousField("a".into())),
        input = ast::Expression::Identifier("a".into()),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
            Key::bot(1u16) => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_subpath_in_current_and_super_must_exist_in_current,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("test", 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            field: "c".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        env = map! {
            ("test", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
            ("super_test", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_subpath_in_current_and_super_may_exist_is_ambiguous,
        method = algebrize_expression,
        expected = Err(Error::AmbiguousField("a".into())),
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        env = map! {
            ("test", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("super_test", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_subpath_in_super_scope,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("super_test", 0u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            field: "c".into(),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        env = map! {
            ("test", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("super_test", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{"c".into()},
                        additional_properties: false,
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        qualified_ref_prefers_super_datasource_to_local_field,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        // test MongoSQL: SELECT (SELECT foo.a FROM bar) FROM foo => (foo.a)
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("foo".into())),
            subpath: "a".into(),
        }),
        env = map! {
            ("bar", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "foo".into() => Schema::Document( Document {
                        keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic( Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        qualified_ref_to_local_field,
        method = algebrize_expression,
        expected = Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("bar", 1u16).into())),
                field: "foo".into(),
                cache: SchemaCache::new(),
            })),
            field: "a".into(),
            cache: SchemaCache::new(),
        })),
        //test MongoSQL: SELECT (SELECT bar.foo.a FROM bar) FROM foo => (bar.foo.a)
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("bar".into())),
                subpath: "foo".into(),
            })),
            subpath: "a".into(),
        }),
        env = map! {
            ("bar", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "foo".into() => Schema::Document( Document {
                        keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic( Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        unqualified_reference_and_may_contain_sub_and_must_contain_outer_is_ambiguous,
        method = algebrize_expression,
        expected = Err(Error::AmbiguousField("a".into())),
        input = ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        env = map! {
            ("test", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{},
                additional_properties: false,
            }),
            ("super_test", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Document( Document {
                        keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                        required: set!{},
                        additional_properties: false,
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        ref_does_not_exist,
        method = algebrize_expression,
        expected = Err(Error::FieldNotFound("bar".into())),
        input = ast::Expression::Identifier("bar".into()),
    );

    test_algebrize!(
        add_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Add,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Add,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        add_wrong_types,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Add,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        sub_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Sub,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Sub,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        sub_wrong_types,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Sub",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Sub,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        div_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Div,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Div,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        div_wrong_types,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Div",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Div,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        mul_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Mul,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Mul,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        mul_wrong_types,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Mul",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Mul,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        concat_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Concat,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::String("42".into()).into()),
                    ir::Expression::Literal(ir::LiteralValue::String("42".into()).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            op: ast::BinaryOp::Concat,
            right: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
        }),
    );
    test_algebrize!(
        concat_wrong_types,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Concat",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Concat,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        neg_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Neg,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Integer(42).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        neg_wrong_type,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Neg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Boolean),
        })),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        }),
    );

    test_algebrize!(
        pos_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Pos,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Integer(42).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Pos,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        pos_wrong_type,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Pos",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Boolean),
        })),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Pos,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        }),
    );

    test_algebrize!(
        standard_scalar_function,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Lower,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::String("hello".into()).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Lower,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("hello".into(),)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );
    test_algebrize!(
        ltrim,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::LTrim,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::String("hello".into()).into()),
                    ir::Expression::Literal(ir::LiteralValue::String("hello world".into()).into())
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Leading,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello world".into()
            ))),
        }),
    );
    test_algebrize!(
        rtrim,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::RTrim,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::String("world".into()).into()),
                    ir::Expression::Literal(ir::LiteralValue::String("hello world".into()).into())
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Trailing,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(
                "world".into()
            ))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello world".into()
            ))),
        }),
    );
    test_algebrize!(
        btrim,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::BTrim,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::String(" ".into()).into()),
                    ir::Expression::Literal(
                        ir::LiteralValue::String(" hello world ".into()).into()
                    )
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(" ".into()))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(
                " hello world ".into()
            ))),
        }),
    );
    test_algebrize!(
        trim_arg_must_be_string_or_null,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "BTrim",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(" ".into()))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        trim_escape_must_be_string,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "BTrim",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(" ".into()))),
        }),
    );

    test_algebrize!(
        extract_year,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Year,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Year,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_month,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Month,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Month,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_day,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Day,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Day,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_hour,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Hour,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Hour,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_minute,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Minute,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Minute,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_second,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Second,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Second,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_day_of_year,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::DayOfYear,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::DayOfYear,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_iso_week,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::IsoWeek,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::IsoWeek,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_iso_weekday,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::IsoWeekday,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        cache: SchemaCache::new(),
                    }
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::IsoWeekday,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_must_be_date,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Second",
            required: DATE_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Second,
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        unsupported_extract_date_part,
        method = algebrize_expression,
        expected = Err(Error::InvalidExtractDatePart(ast::DatePart::Quarter)),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Quarter,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        dateadd,
        method = algebrize_expression,
        expected = Ok(ir::Expression::DateFunction(ir::DateFunctionApplication {
            function: ir::DateFunction::Add,
            date_part: ir::DatePart::Quarter,
            args: vec![
                ir::Expression::Literal(ir::LiteralExpr {
                    value: ir::LiteralValue::Integer(5),
                    cache: SchemaCache::new()
                }),
                ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
            ],
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Add,
            date_part: ast::DatePart::Quarter,
            args: vec![
                ast::Expression::Literal(ast::Literal::Integer(5)),
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })
            ],
        }),
    );
    test_algebrize!(
        datediff,
        method = algebrize_expression,
        expected = Ok(ir::Expression::DateFunction(ir::DateFunctionApplication {
            function: ir::DateFunction::Diff,
            date_part: ir::DatePart::Week,
            args: vec![
                ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
                ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
                ir::Expression::Literal(ir::LiteralExpr {
                    value: ir::LiteralValue::String("sunday".to_string()),
                    cache: SchemaCache::new()
                })
            ],
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Diff,
            date_part: ast::DatePart::Week,
            args: vec![
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::Literal(ast::Literal::String("sunday".to_string()))
            ],
        }),
    );
    test_algebrize!(
        datetrunc,
        method = algebrize_expression,
        expected = Ok(ir::Expression::DateFunction(ir::DateFunctionApplication {
            function: ir::DateFunction::Trunc,
            date_part: ir::DatePart::Year,
            args: vec![
                ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
                ir::Expression::Literal(ir::LiteralExpr {
                    value: ir::LiteralValue::String("sunday".to_string()),
                    cache: SchemaCache::new()
                })
            ],
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Trunc,
            date_part: ast::DatePart::Year,
            args: vec![
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::Literal(ast::Literal::String("sunday".to_string()))
            ],
        }),
    );
    test_algebrize!(
        unsupported_date_function_date_part,
        method = algebrize_expression,
        expected = Err(Error::InvalidDateFunctionDatePart(ast::DatePart::IsoWeek)),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Trunc,
            date_part: ast::DatePart::IsoWeek,
            args: vec![
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::Literal(ast::Literal::String("sunday".to_string()))
            ],
        }),
    );

    test_algebrize!(
        searched_case,
        method = algebrize_expression,
        expected = Ok(ir::Expression::SearchedCase(ir::SearchedCaseExpr {
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(
                    ir::LiteralValue::Boolean(true).into()
                )),
                then: Box::new(ir::Expression::Literal(
                    ir::LiteralValue::String("bar".into()).into()
                )),
            }],
            else_branch: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("foo".into()).into()
            )),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                then: Box::new(ast::Expression::Literal(ast::Literal::String("bar".into()))),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::String(
                "foo".into()
            )))),
        }),
    );
    test_algebrize!(
        searched_case_no_else,
        method = algebrize_expression,
        expected = Ok(ir::Expression::SearchedCase(ir::SearchedCaseExpr {
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(
                    ir::LiteralValue::Boolean(true).into()
                )),
                then: Box::new(ir::Expression::Literal(
                    ir::LiteralValue::String("bar".into()).into()
                )),
            }],
            else_branch: Box::new(ir::Expression::Literal(ir::LiteralValue::Null.into())),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                then: Box::new(ast::Expression::Literal(ast::Literal::String("bar".into()))),
            }],
            else_branch: None,
        }),
    );
    test_algebrize!(
        searched_case_when_condition_is_not_bool,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "SearchedCase",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::String("foo".into()))),
                then: Box::new(ast::Expression::Literal(ast::Literal::String("bar".into()))),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::String(
                "foo".into()
            )))),
        }),
    );

    test_algebrize!(
        simple_case,
        method = algebrize_expression,
        expected = Ok(ir::Expression::SimpleCase(ir::SimpleCaseExpr {
            expr: Box::new(ir::Expression::Literal(ir::LiteralValue::Integer(1).into())),
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(ir::LiteralValue::Integer(2).into())),
                then: Box::new(ir::Expression::Literal(
                    ir::LiteralValue::String("bar".into()).into()
                )),
            }],
            else_branch: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("foo".into()).into()
            )),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                then: Box::new(ast::Expression::Literal(ast::Literal::String("bar".into()))),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::String(
                "foo".into()
            )))),
        }),
    );
    test_algebrize!(
        simple_case_no_else,
        method = algebrize_expression,
        expected = Ok(ir::Expression::SimpleCase(ir::SimpleCaseExpr {
            expr: Box::new(ir::Expression::Literal(ir::LiteralValue::Integer(1).into())),
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(ir::LiteralValue::Integer(2).into())),
                then: Box::new(ir::Expression::Literal(
                    ir::LiteralValue::String("bar".into()).into()
                )),
            }],
            else_branch: Box::new(ir::Expression::Literal(ir::LiteralValue::Null.into())),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                then: Box::new(ast::Expression::Literal(ast::Literal::String("bar".into()))),
            }],
            else_branch: None,
        }),
    );
    test_algebrize!(
        simple_case_operand_and_when_operand_not_comparable,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::InvalidComparison(
            "SimpleCase",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::String("foo".into()))),
                then: Box::new(ast::Expression::Literal(ast::Literal::String("bar".into()))),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::String(
                "baz".into()
            )))),
        }),
    );

    test_algebrize!(
        cast_full,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Cast(ir::CastExpr {
            expr: Box::new(ir::Expression::Literal(
                ir::LiteralValue::Integer(42).into()
            )),
            to: ir::Type::String,
            on_null: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("was_null".into()).into()
            )),
            on_error: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("was_error".into()).into()
            )),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Cast(ast::CastExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            to: ast::Type::String,
            on_null: Some(Box::new(ast::Expression::Literal(ast::Literal::String(
                "was_null".into()
            )))),
            on_error: Some(Box::new(ast::Expression::Literal(ast::Literal::String(
                "was_error".into()
            )))),
        }),
    );
    test_algebrize!(
        cast_simple,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Cast(ir::CastExpr {
            expr: Box::new(ir::Expression::Literal(
                ir::LiteralValue::Integer(42).into()
            )),
            to: ir::Type::String,
            on_null: Box::new(ir::Expression::Literal(ir::LiteralValue::Null.into())),
            on_error: Box::new(ir::Expression::Literal(ir::LiteralValue::Null.into())),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Cast(ast::CastExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            to: ast::Type::String,
            on_null: None,
            on_error: None,
        }),
    );

    test_algebrize!(
        type_assert_success,
        method = algebrize_expression,
        expected = Ok(ir::Expression::TypeAssertion(ir::TypeAssertionExpr {
            expr: Box::new(ir::Expression::Literal(
                ir::LiteralValue::Integer(42).into()
            )),
            target_type: ir::Type::Int32,
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::Type::Int32,
        }),
    );
    test_algebrize!(
        type_assert_fail,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "::!",
            required: Schema::Atomic(Atomic::String),
            found: Schema::Atomic(Atomic::Integer)
        })),
        input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::Type::String,
        }),
    );

    test_algebrize!(
        is_success,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Is(ir::IsExpr {
            expr: Box::new(ir::Expression::Literal(
                ir::LiteralValue::Integer(42).into()
            )),
            target_type: ir::TypeOrMissing::Type(ir::Type::Int32),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Is(ast::IsExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
        }),
    );
    test_algebrize!(
        is_recursive_failure,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String)
        })),
        input = ast::Expression::Is(ast::IsExpr {
            expr: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            })),
            target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
        }),
    );

    test_algebrize!(
        like_success_with_pattern,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Like(ir::LikeExpr {
            expr: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("42".into()).into()
            )),
            pattern: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("42".into()).into()
            )),
            escape: Some("foo".into()),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            escape: Some("foo".into()),
        }),
    );
    test_algebrize!(
        like_success_no_pattern,
        method = algebrize_expression,
        expected = Ok(ir::Expression::Like(ir::LikeExpr {
            expr: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("42".into()).into()
            )),
            pattern: Box::new(ir::Expression::Literal(
                ir::LiteralValue::String("42".into()).into()
            )),
            escape: None,
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            escape: None,
        }),
    );
    test_algebrize!(
        like_expr_must_be_string,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer)
        })),
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            escape: Some(" ".into()),
        }),
    );
    test_algebrize!(
        like_pattern_must_be_string,
        method = algebrize_expression,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer)
        })),
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            escape: Some(" ".into()),
        }),
    );

    test_algebrize!(
        log_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Log,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(100).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Log,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(100)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        round_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Round,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Round,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(10)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        cos_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Cos,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Integer(10).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Cos,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );
    test_algebrize!(
        sin_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Sin,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Integer(10).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Sin,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );
    test_algebrize!(
        tan_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Tan,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Integer(10).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Tan,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );
    test_algebrize!(
        radians_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Radians,
                args: vec![ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Radians,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );
    test_algebrize!(
        sqrt_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Sqrt,
                args: vec![ir::Expression::Literal(ir::LiteralValue::Integer(4).into()),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Sqrt,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(4)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        abs_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Abs,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Integer(10).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Abs,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        ceil_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Ceil,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Double(1.5).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Ceil,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Double(1.5)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        degrees_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Degrees,
                args: vec![ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Degrees,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        floor_unary_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Floor,
                args: vec![ir::Expression::Literal(
                    ir::LiteralValue::Double(1.5).into()
                ),],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Floor,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Double(1.5)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        mod_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Mod,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Mod,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(10)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        pow_bin_op,
        method = algebrize_expression,
        expected = Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Pow,
                args: vec![
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                    ir::Expression::Literal(ir::LiteralValue::Integer(10).into()),
                ],
                cache: SchemaCache::new(),
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Pow,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(10)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );
}

mod aggregation {
    use crate::{
        ast, ir, map, multimap,
        schema::{Atomic, Schema, ANY_DOCUMENT, NUMERIC_OR_NULLISH},
        unchecked_unique_linked_hash_map,
    };
    test_algebrize!(
        count_star,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::CountStar(false)),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        count_distinct_star_is_error,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(
            ir::schema::Error::CountDistinctStarNotSupported
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize!(
        count_all_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Count,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        count_distinct_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Count,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize!(
        count_distinct_expr_argument_not_self_comparable_is_error,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(
            ir::schema::Error::AggregationArgumentMustBeSelfComparable(
                "Count DISTINCT".into(),
                Schema::Any
            )
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Identifier("foo".into())]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
        env = map! {
            ("d", 1u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        sum_star_is_error,
        method = algebrize_aggregation,
        expected = Err(Error::StarInNonCount),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        sum_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Sum,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        sum_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Sum,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize!(
        sum_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Sum",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        avg_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Avg,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        avg_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Avg,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        avg_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Avg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        stddevpop_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevPop,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        stddevpop_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevPop,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize!(
        stddevpop_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "StddevPop",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        stddevsamp_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevSamp,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        stddevsamp_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevSamp,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize!(
        stddevsamp_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "StddevSamp",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        addtoarray_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToArray,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        addtoarray_distinct_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToArray,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        addtoset_expr_is_addtoarray_distinct_in_ir,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToSet,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        addtoset_distinct_expr_is_addtoarray_in_ir,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToSet,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        first_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::First,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::First,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        first_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::First,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::First,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        last_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Last,
                distinct: false,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Last,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        last_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Last,
                distinct: true,
                arg: ir::Expression::Literal(ir::LiteralValue::Integer(42).into()).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Last,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        mergedocuments_expr,
        method = algebrize_aggregation,
        expected = Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::MergeDocuments,
                distinct: false,
                arg: Box::new(ir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                        "b".into() => ir::Expression::Literal(ir::LiteralValue::Integer(42).into()),
                    }
                    .into()
                ))
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::MergeDocuments,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            })]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        mergedocuments_argument_must_be_document,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "MergeDocuments",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        input = ast::FunctionExpr {
            function: ast::FunctionName::MergeDocuments,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
}

mod select_clause {
    use super::catalog;
    use crate::{
        ast,
        ir::{self, binding_tuple::Key, schema::SchemaCache},
        map, multimap,
        schema::ANY_DOCUMENT,
        unchecked_unique_linked_hash_map,
    };

    fn source() -> ir::Stage {
        ir::Stage::Collection(ir::Collection {
            db: "test".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        })
    }

    test_algebrize!(
        select_distinct_not_allowed,
        method = algebrize_select_clause,
        expected = Err(Error::DistinctSelect),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::Distinct,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Identifier("foo".into())
            ),]),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        select_duplicate_bot,
        method = algebrize_select_clause,
        expected = Err(Error::DuplicateKey(Key::bot(1u16))),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
                ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
            ]),
        },
        source = source(),
        env = map! {},
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_duplicate_doc_key_a,
        method = algebrize_select_clause,
        expected = Err(Error::DuplicateDocumentKey("a".into())),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                },)
            ),]),
        },
        source = source(),
        env = map! {},
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_bot_and_double_substar,
        method = algebrize_select_clause,
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(source()),
            expression: map! {
                ("bar", 1u16).into() => ir::Expression::Reference(("bar", 1u16).into()),
                Key::bot(1u16) => ir::Expression::Document(unchecked_unique_linked_hash_map!{}.into()),
                ("foo", 1u16).into() => ir::Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Substar("bar".into()),
                ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
                ast::SelectValuesExpression::Substar("foo".into()),
            ]),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            ("bar", 1u16).into() => ANY_DOCUMENT.clone(),
        },
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_value_expression_must_be_document,
        method = algebrize_select_clause,
        expected = Err(Error::SchemaChecking(
            crate::ir::schema::Error::SchemaChecking {
                name: "project datasource",
                required: ANY_DOCUMENT.clone(),
                found: crate::schema::Schema::Atomic(crate::schema::Atomic::String),
            }
        )),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Literal(ast::Literal::String("foo".into()))
            ),]),
        },
        source = source(),
        env = map! {},
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_duplicate_substar,
        method = algebrize_select_clause,
        expected = Err(Error::DuplicateKey(("foo", 1u16).into())),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Substar("foo".into()),
                ast::SelectValuesExpression::Substar("foo".into()),
            ]),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_substar_body,
        method = algebrize_select_clause,
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(source()),
            expression: map! {
                ("foo", 1u16).into() => ir::Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(
                vec![ast::SelectValuesExpression::Substar("foo".into()),]
            ),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
        catalog = catalog(vec![("test", "baz")]),
    );
}

mod from_clause {
    use super::{catalog, ir_source_bar, ir_source_foo, AST_SOURCE_BAR, AST_SOURCE_FOO};
    use crate::{
        ast::{self, JoinSource},
        ir::{self, binding_tuple::Key, schema::SchemaCache, JoinType},
        map, multimap,
        schema::{Atomic, Document, Schema, ANY_DOCUMENT},
        set, unchecked_unique_linked_hash_map,
    };

    fn ir_array_source() -> ir::Stage {
        ir::Stage::Array(ir::ArraySource {
            array: vec![ir::Expression::Document(ir::DocumentExpr {
                document: unchecked_unique_linked_hash_map! {"a".to_string() => ir::Expression::Document(ir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map!{
                        "b".to_string() => ir::Expression::Literal(ir::LiteralExpr {
                            value: ir::LiteralValue::Integer(5),
                            cache: SchemaCache::new()
                        })
                    },
                    cache: SchemaCache::new()
                })},
                cache: SchemaCache::new(),
            })],
            alias: "arr".to_string(),
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
        from_clause_must_exist,
        method = algebrize_from_clause,
        expected = Err(Error::NoFromClause),
        input = None,
    );
    test_algebrize!(
        collection_must_have_alias,
        method = algebrize_from_clause,
        expected = Err(Error::CollectionMustHaveAlias),
        input = Some(ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: None,
        })),
    );
    test_algebrize!(
        basic_collection,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Collection(ir::Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    ir::Expression::Reference(("foo", 0u16).into())
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
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Collection(ir::Collection {
                db: "test2".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    ir::Expression::Reference(("foo", 0u16).into())
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
        empty_array,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Array(ir::ArraySource {
            array: vec![],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        }),),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        dual,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Array(ir::ArraySource {
            array: vec![ir::Expression::Document(
                unchecked_unique_linked_hash_map! {}.into()
            )],
            alias: "_dual".into(),
            cache: SchemaCache::new(),
        }),),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {},)],
            alias: "_dual".into(),
        })),
    );
    test_algebrize!(
        int_array,
        method = algebrize_from_clause,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Integer(42))],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        null_array,
        method = algebrize_from_clause,
        expected = Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Null)]),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Null)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        array_datasource_must_be_literal,
        method = algebrize_from_clause,
        expected = Err(Error::ArrayDatasourceMustBeLiteral),
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
        expected = Ok(ir::Stage::Array(ir::ArraySource {
            array: vec![ir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "foo".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                    "bar".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                }
                .into()
            )],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        }),),
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
        expected = Ok(ir::Stage::Array(ir::ArraySource {
            array: vec![
                ir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                    "bar".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                }.into()),
                ir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo2".into() => ir::Expression::Literal(ir::LiteralValue::Integer(41).into()),
                    "bar2".into() => ir::Expression::Literal(ir::LiteralValue::Integer(42).into())
                }.into())
            ],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        }),),
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
        expected = Ok(ir::Stage::Array(ir::ArraySource {
            array: vec![
                ir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                    "bar".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                }.into()),
                ir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "foo2".into() => ir::Expression::Document(
                        unchecked_unique_linked_hash_map!{"nested".into() => ir::Expression::Literal(ir::LiteralValue::Integer(52).into())}
                    .into()),
                    "bar2".into() => ir::Expression::Literal(ir::LiteralValue::Integer(42).into())
                }.into())
            ],
            alias: "bar".into(),
            cache: SchemaCache::new(),
        }),),
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
        expected = Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Left,
            left: Box::new(ir_source_foo()),
            right: Box::new(ir_source_bar()),
            condition: Some(ir::Expression::Literal(
                ir::LiteralValue::Boolean(true).into()
            )),
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
        expected = Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Left,
            left: Box::new(ir_source_bar()),
            right: Box::new(ir_source_foo()),
            condition: Some(ir::Expression::Literal(
                ir::LiteralValue::Boolean(true).into()
            )),
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
        expected = Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Inner,
            left: Box::new(ir_source_foo()),
            right: Box::new(ir_source_bar()),
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
        expected = Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Inner,
            left: Box::new(ir_source_foo()),
            right: Box::new(ir_source_bar()),
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
        invalid_join_condition,
        method = algebrize_from_clause,
        expected = Err(Error::FieldNotFound("x".into())),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                })],
                alias: "foo".into()
            })),
            right: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                })],
                alias: "bar".into()
            })),
            condition: Some(ast::Expression::Identifier("x".into())),
        })),
    );
    test_algebrize!(
        derived_single_datasource,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Derived(ir::Derived {
            source: Box::new(ir::Stage::Project(ir::Project {
                source: Box::new(ir::Stage::Array(ir::ArraySource {
                    array: vec![ir::Expression::Document(
                        unchecked_unique_linked_hash_map! {"foo".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                             "bar".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                        }
                    .into())],
                    alias: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {("d", 0u16).into() =>
                    ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::MergeObjects,
                        args: vec![
                            ir::Expression::Reference(("bar", 1u16).into())
                        ],
                        cache: SchemaCache::new(),
                    }
                    )
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
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
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        derived_multiple_datasources,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Derived(ir::Derived {
            source: Box::new(ir::Stage::Project(ir::Project {
                source: Box::new(ir::Stage::Project(ir::Project {
                    source: Box::new(ir::Stage::Array(ir::ArraySource {
                        array: vec![ir::Expression::Document(
                            unchecked_unique_linked_hash_map! {"foo".into() => ir::Expression::Literal(
                                ir::LiteralValue::Integer(1).into()
                                ),
                                "bar".into() => ir::Expression::Literal(
                                    ir::LiteralValue::Integer(1).into())
                            }
                            .into()
                        )],
                        alias: "bar".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        Key::bot(1u16) => ir::Expression::Document(
                            unchecked_unique_linked_hash_map!{"baz".into() => ir::Expression::Literal(ir::LiteralValue::String("hello".into()).into())}
                        .into()),
                        ("bar", 1u16).into() =>
                            ir::Expression::Reference(("bar", 1u16).into())
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! { ("d", 0u16).into() =>
                    ir::Expression::ScalarFunction(
                        ir::ScalarFunctionApplication {
                            function: ir::ScalarFunction::MergeObjects,
                            args:
                                vec![
                                    ir::Expression::Reference(Key::bot(1u16).into()),
                                    ir::Expression::Reference(("bar", 1u16).into()),
                                ],
                            cache: SchemaCache::new(),
                        }
                    )
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![
                        ast::SelectValuesExpression::Substar("bar".into()),
                        ast::SelectValuesExpression::Expression(ast::Expression::Document(
                            multimap! {
                                "baz".into() => ast::Expression::Literal(ast::Literal::String("hello".into()))
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
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        derived_join_datasources_distinct_keys_succeeds,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Derived(ir::Derived {
            source: Box::new(ir::Stage::Project(ir::Project {
                source: ir::Stage::Join(ir::Join {
                    join_type: ir::JoinType::Inner,
                    left: ir::Stage::Array(ir::ArraySource {
                        array: vec![ir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                            "foo1".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                            "bar1".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                                                        }
                        .into())],
                        alias: "bar1".into(),
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    right: ir::Stage::Array(ir::ArraySource {
                        array: vec![ir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                            "foo2".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                            "bar2".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                                                        }
                        .into())],
                        alias: "bar2".into(),
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    condition: None,
                    cache: SchemaCache::new(),
                })
                .into(),
                expression: map! {("d", 0u16).into() =>
                    ir::Expression::ScalarFunction(
                        ir::ScalarFunctionApplication {
                            function: ir::ScalarFunction::MergeObjects,
                            args: vec![ir::Expression::Reference(("bar1", 1u16).into()),
                                       ir::Expression::Reference(("bar2", 1u16).into())],
                            cache: SchemaCache::new(),
                        }
                    )
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
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
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        join_condition_referencing_non_correlated_fields,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Join(ir::Join {
            join_type: ir::JoinType::Left,
            left: Box::new(ir::Stage::Array(ir::ArraySource {
                array: vec![ir::Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())}
                .into())],
                alias: "foo".to_string(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(ir::Stage::Array(ir::ArraySource {
                array: vec![ir::Expression::Document(
                    unchecked_unique_linked_hash_map! {"b".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(4).into())}
                .into())],
                alias: "bar".to_string(),
                cache: SchemaCache::new(),
            })),
            condition: Some(ir::Expression::ScalarFunction(
                ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::Eq,
                    args: vec![
                        ir::Expression::FieldAccess(ir::FieldAccess {
                            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                            field: "a".to_string(),
                            cache: SchemaCache::new(),
                        }),
                        ir::Expression::FieldAccess(ir::FieldAccess {
                            expr: Box::new(ir::Expression::Reference(("bar", 0u16).into())),
                            field: "b".to_string(),
                            cache: SchemaCache::new(),
                        })
                    ],
                    cache: SchemaCache::new(),
                }
            )),
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
        expected = Err(Error::DerivedDatasouceOverlappingKeys(
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "foo1".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "foo1".into()
                },
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "foo2".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "foo2".into()
                },
                additional_properties: false,
            }),
            crate::schema::Satisfaction::Must,
        )),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
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
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        flatten_simple,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir_array_source()),
            expression: map! {
                ("arr", 0u16).into() => ir::Expression::Document(ir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!{
                    "a_b".to_string() => ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                            expr: Box::new(ir::Expression::Reference(ir::ReferenceExpr {
                                key: ("arr", 0u16).into(),
                                cache: SchemaCache::new()
                            })),
                            field: "a".to_string(), cache: SchemaCache::new()
                        })),
                        field: "b".to_string(), cache: SchemaCache::new()
                    })},
                cache: SchemaCache::new()
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
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Array(ir::ArraySource {
                array: vec![ir::Expression::Document(ir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => ir::Expression::Document(ir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map!{
                                "b".to_string() => ir::Expression::Literal(ir::LiteralExpr {
                                    value: ir::LiteralValue::Integer(5),
                                    cache: SchemaCache::new()
                                })
                            },
                            cache: SchemaCache::new()
                        }),
                        "x".to_string() => ir::Expression::Document(ir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                                "y".to_string() => ir::Expression::Literal(ir::LiteralExpr {
                                    value: ir::LiteralValue::Integer(8),
                                    cache: SchemaCache::new()
                                })
                            },
                            cache: SchemaCache::new()
                        })
                    },
                    cache: SchemaCache::new()
                })],
                alias: "arr".to_string(),
                cache: SchemaCache::new()
            })),
            expression: map! {
                ("arr", 0u16).into() => ir::Expression::Document(ir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_b".to_string() => ir::Expression::FieldAccess(ir::FieldAccess {
                            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                                expr: Box::new(ir::Expression::Reference(ir::ReferenceExpr {
                                    key: ("arr", 0u16).into(),
                                    cache: SchemaCache::new()
                                })),
                                field: "a".to_string(),
                                cache: SchemaCache::new()
                            })),
                            field: "b".to_string(),
                            cache: SchemaCache::new()
                        }),
                        "x_y".to_string() => ir::Expression::FieldAccess(ir::FieldAccess {
                            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                                expr: Box::new(ir::Expression::Reference(ir::ReferenceExpr {
                                    key: ("arr", 0u16).into(),
                                    cache: SchemaCache::new()
                                })),
                                field: "x".to_string(),
                                cache: SchemaCache::new()
                            })),
                            field: "y".to_string(),
                            cache: SchemaCache::new()
                        })
                    },
                    cache: SchemaCache::new()
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
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(ast_array_source()),
            options: vec![ast::FlattenOption::Depth(1), ast::FlattenOption::Depth(2)]
        })),
    );
    test_algebrize!(
        flatten_polymorphic_non_document_schema_array_source,
        method = algebrize_from_clause,
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Array(ir::ArraySource {
                array: vec![
                    ir::Expression::Document(ir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => ir::Expression::Document(ir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map!{
                                "b".to_string() => ir::Expression::Document(ir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map!{
                                        "c".to_string() => ir::Expression::Literal(ir::LiteralExpr {
                                            value: ir::LiteralValue::Integer(5),
                                            cache: SchemaCache::new()
                                        })},
                                    cache: SchemaCache::new()
                                })},
                            cache: SchemaCache::new()
                        })},
                        cache: SchemaCache::new()
                    }),
                    ir::Expression::Document(ir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => ir::Expression::Document(ir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map!{
                                "b".to_string() => ir::Expression::Document(ir::DocumentExpr {
                                    document: unchecked_unique_linked_hash_map!{
                                        "c".to_string() => ir::Expression::Literal(ir::LiteralExpr {
                                            value: ir::LiteralValue::String("hello".to_string()),
                                            cache: SchemaCache::new()
                                        })},
                                    cache: SchemaCache::new()
                                })},
                            cache: SchemaCache::new()
                        })},
                        cache: SchemaCache::new()
                    })
                ],
                alias: "arr".to_string(),
                cache: SchemaCache::new()
            })),
            expression: map! {
            ("arr", 0u16).into() => ir::Expression::Document(ir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!{
                    "a_b_c".to_string() => ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                                expr: Box::new(ir::Expression::Reference(ir::ReferenceExpr {
                                    key: ("arr", 0u16).into(),
                                    cache: SchemaCache::new()
                                })),
                                field: "a".to_string(),
                                cache: SchemaCache::new()
                            })),
                            field: "b".to_string(),
                            cache: SchemaCache::new()
                        })),
                        field: "c".to_string(),
                        cache: SchemaCache::new()
                    })},
                cache: SchemaCache::new()
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
                        multimap!{"c".into() => ast::Expression::Literal(ast::Literal::String("hello".to_string()))},
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
    mod unwind {
        use super::*;
        use crate::catalog::{Catalog, Namespace};

        /// Most tests use the same collection source and need to specify the
        /// collection schema for the test to work. This helper allows easy
        /// definition of that collection schema.
        fn make_catalog(s: Schema) -> Catalog {
            Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => s,
            })
        }

        test_algebrize!(
            simple,
            method = algebrize_from_clause,
            expected = Ok(ir::Stage::Unwind(ir::Unwind {
                source: Box::new(ir_source_foo()),
                path: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                    field: "arr".into(),
                    cache: SchemaCache::new(),
                })),
                index: None,
                outer: false,
                cache: SchemaCache::new()
            })),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Identifier(
                    "arr".into()
                ))]
            })),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set! {"arr".into()},
                additional_properties: false,
            })),
        );
        test_algebrize!(
            all_opts,
            method = algebrize_from_clause,
            expected = Ok(ir::Stage::Unwind(ir::Unwind {
                source: Box::new(ir_source_foo()),
                path: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                    field: "arr".into(),
                    cache: SchemaCache::new(),
                })),
                index: Some("i".into()),
                outer: true,
                cache: SchemaCache::new()
            })),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![
                    ast::UnwindOption::Path(ast::Expression::Identifier("arr".into())),
                    ast::UnwindOption::Index("i".into()),
                    ast::UnwindOption::Outer(true),
                ]
            })),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set! {"arr".into()},
                additional_properties: false,
            })),
        );
        test_algebrize!(
            compound_path,
            method = algebrize_from_clause,
            expected = Ok(ir::Stage::Unwind(ir::Unwind {
                source: Box::new(ir_source_foo()),
                path: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                        field: "doc".into(),
                        cache: SchemaCache::new(),
                    })),
                    field: "arr".into(),
                    cache: SchemaCache::new(),
                })),
                index: None,
                outer: false,
                cache: SchemaCache::new()
            })),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
                    ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("doc".into())),
                        subpath: "arr".into()
                    }
                ))]
            })),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "doc".into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                        },
                        required: set!{"arr".into()},
                        additional_properties: false,
                    }),
                },
                required: set! {"doc".into()},
                additional_properties: false,
            })),
        );
        test_algebrize!(
            duplicate_opts,
            method = algebrize_from_clause,
            expected = Err(Error::DuplicateUnwindOption(ast::UnwindOption::Path(
                ast::Expression::Identifier("dup".into())
            ))),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![
                    ast::UnwindOption::Path(ast::Expression::Identifier("arr".into())),
                    ast::UnwindOption::Path(ast::Expression::Identifier("dup".into())),
                ]
            })),
            catalog = make_catalog(ANY_DOCUMENT.clone()),
        );
        test_algebrize!(
            missing_path,
            method = algebrize_from_clause,
            expected = Err(Error::NoUnwindPath),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![]
            })),
            catalog = make_catalog(ANY_DOCUMENT.clone()),
        );
        test_algebrize!(
            invalid_path,
            method = algebrize_from_clause,
            expected = Err(Error::InvalidUnwindPath),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
                    ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Document(vec![ast::DocumentPair {
                            key: "arr".into(),
                            value: ast::Expression::Array(vec![
                                ast::Expression::Literal(ast::Literal::Integer(1)),
                                ast::Expression::Literal(ast::Literal::Integer(2)),
                                ast::Expression::Literal(ast::Literal::Integer(3))
                            ])
                        }])),
                        subpath: "arr".into()
                    }
                )),]
            })),
            catalog = make_catalog(ANY_DOCUMENT.clone()),
        );
        test_algebrize!(
            correlated_path_disallowed,
            method = algebrize_from_clause,
            expected = Err(Error::FieldNotFound("bar".into())),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
                    ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("bar".into())),
                        subpath: "arr".into()
                    }
                )),]
            })),
            env = map! {
                ("bar", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                    },
                    required: set!{ "arr".into() },
                    additional_properties: false,
                }),
            },
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set! {"arr".into()},
                additional_properties: false,
            })),
        );
    }
}

mod limit_or_offset_clause {
    use super::{catalog, ir_source_foo, AST_SOURCE_FOO};
    use crate::{ast, ir, ir::schema::SchemaCache};

    test_algebrize!(
        limit_set,
        method = algebrize_limit_clause,
        expected = Ok(ir::Stage::Limit(ir::Limit {
            source: Box::new(ir_source_foo()),
            limit: 42_u64,
            cache: SchemaCache::new(),
        })),
        input = Some(42_u32),
        source = ir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        limit_unset,
        method = algebrize_limit_clause,
        expected = Ok(ir_source_foo()),
        input = None,
        source = ir_source_foo(),
    );
    test_algebrize!(
        offset_set,
        method = algebrize_offset_clause,
        expected = Ok(ir::Stage::Offset(ir::Offset {
            source: Box::new(ir_source_foo()),
            offset: 3_u64,
            cache: SchemaCache::new(),
        })),
        input = Some(3_u32),
        source = ir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        offset_unset,
        method = algebrize_offset_clause,
        expected = Ok(ir_source_foo()),
        input = None,
        source = ir_source_foo(),
    );
    test_algebrize!(
        limit_and_offset,
        method = algebrize_select_query,
        expected = Ok(ir::Stage::Limit(ir::Limit {
            source: Box::new(ir::Stage::Offset(ir::Offset {
                source: Box::new(ir_source_foo()),
                offset: 3,
                cache: SchemaCache::new(),
            })),
            limit: 10,
            cache: SchemaCache::new(),
        })),
        input = ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(AST_SOURCE_FOO.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(10_u32),
            offset: Some(3_u32)
        },
        catalog = catalog(vec![("test", "foo")]),
    );
}

mod set_query {
    use super::{catalog, ir_source_bar, ir_source_foo, AST_QUERY_BAR, AST_QUERY_FOO};
    use crate::{ast, ir, ir::schema::SchemaCache};

    test_algebrize!(
        union_distinct_not_allowed,
        method = algebrize_set_query,
        expected = Err(Error::DistinctUnion),
        input = ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::Union,
            right: Box::new(AST_QUERY_BAR.clone()),
        },
    );
    test_algebrize!(
        basic,
        method = algebrize_set_query,
        expected = Ok(ir::Stage::Set(ir::Set {
            operation: ir::SetOperation::UnionAll,
            left: Box::new(ir_source_foo()),
            right: Box::new(ir_source_bar()),
            cache: SchemaCache::new(),
        })),
        input = ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::UnionAll,
            right: Box::new(AST_QUERY_BAR.clone()),
        },
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
}

mod filter_clause {
    use super::{catalog, ir_source_foo};
    use crate::{ast, ir, ir::schema::SchemaCache};

    fn true_ir() -> ir::Expression {
        ir::Expression::Literal(ir::LiteralValue::Boolean(true).into())
    }
    const TRUE_AST: ast::Expression = ast::Expression::Literal(ast::Literal::Boolean(true));

    test_algebrize!(
        simple,
        method = algebrize_filter_clause,
        expected = Ok(ir::Stage::Filter(ir::Filter {
            source: Box::new(ir_source_foo()),
            condition: true_ir(),
            cache: SchemaCache::new(),
        })),
        input = Some(TRUE_AST),
        source = ir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        none,
        method = algebrize_filter_clause,
        expected = Ok(ir_source_foo()),
        input = None,
        source = ir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
}

mod order_by_clause {
    use super::catalog;
    use crate::{
        ast, ir,
        ir::schema::SchemaCache,
        map,
        schema::{Atomic, Document, Schema},
        set, unchecked_unique_linked_hash_map,
    };

    fn source() -> ir::Stage {
        ir::Stage::Collection(ir::Collection {
            db: "test".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        })
    }

    test_algebrize!(
        asc_and_desc,
        method = algebrize_order_by_clause,
        expected = Ok(ir::Stage::Sort(ir::Sort {
            source: Box::new(source()),
            specs: vec![
                ir::SortSpecification::Asc(Box::new(ir::Expression::FieldAccess(
                    ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                        field: "a".to_string(),
                        cache: SchemaCache::new(),
                    }
                ))),
                ir::SortSpecification::Desc(Box::new(ir::Expression::FieldAccess(
                    ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                        field: "b".to_string(),
                        cache: SchemaCache::new(),
                    }
                )))
            ],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::OrderByClause {
            sort_specs: vec![
                ast::SortSpec {
                    key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("foo".to_string())),
                        subpath: "a".to_string()
                    })),
                    direction: ast::SortDirection::Asc
                },
                ast::SortSpec {
                    key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("foo".to_string())),
                        subpath: "b".to_string()
                    })),
                    direction: ast::SortDirection::Desc
                }
            ],
        }),
        source = source(),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
        catalog = catalog(vec![("test", "baz")]),
    );

    test_algebrize!(
        sort_key_from_source,
        method = algebrize_order_by_clause,
        expected = Ok(ir::Stage::Sort(ir::Sort {
            source: Box::new(ir::Stage::Array(ir::ArraySource {
                array: vec![ir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                    }
                    .into()
                )],
                alias: "arr".into(),
                cache: SchemaCache::new(),
            })),
            specs: vec![ir::SortSpecification::Asc(Box::new(
                ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                    field: "a".to_string(),
                    cache: SchemaCache::new(),
                })
            )),],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })),
                direction: ast::SortDirection::Asc
            },],
        }),
        source = ir::Stage::Array(ir::ArraySource {
            array: vec![ir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                }
                .into()
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod group_by_clause {
    use crate::{ast, ir, ir::schema::SchemaCache, unchecked_unique_linked_hash_map};
    use lazy_static::lazy_static;

    // ARRAY DATASOURCE
    // [{"a" : 1}] AS arr
    fn ir_array_source() -> ir::Stage {
        ir::Stage::Array(ir::ArraySource {
            array: vec![ir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
                }
                .into(),
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })
    }
    // GROUP BY KEYS
    // arr.a AS key
    fn ir_field_access() -> ir::OptionallyAliasedExpr {
        ir::OptionallyAliasedExpr::Aliased(ir::AliasedExpr {
            alias: "key".to_string(),
            expr: ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                field: "a".to_string(),
                cache: SchemaCache::new(),
            }),
        })
    }
    // 1 AS literal
    fn ir_literal_key() -> ir::OptionallyAliasedExpr {
        ir::OptionallyAliasedExpr::Aliased(ir::AliasedExpr {
            alias: "literal".into(),
            expr: ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
        })
    }

    // a + 1 as complex_expr
    fn ir_field_access_complex_expr() -> ir::OptionallyAliasedExpr {
        ir::OptionallyAliasedExpr::Aliased(ir::AliasedExpr {
            alias: "complex_expr".into(),
            expr: ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Add,
                args: vec![
                    ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                        field: "a".to_string(),
                        cache: SchemaCache::new(),
                    }),
                    ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                ],
                cache: SchemaCache::new(),
            }),
        })
    }
    // AVG(DISTINCT arr.a) AS agg1
    fn ir_agg_1_array() -> ir::AliasedAggregation {
        ir::AliasedAggregation {
            alias: "agg1".to_string(),
            agg_expr: ir::AggregationExpr::Function(ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Avg,
                arg: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                    field: "a".to_string(),
                    cache: SchemaCache::new(),
                })),
                distinct: true,
            }),
        }
    }
    // COUNT(*) AS agg2
    fn ir_agg_2() -> ir::AliasedAggregation {
        ir::AliasedAggregation {
            alias: "agg2".to_string(),
            agg_expr: ir::AggregationExpr::CountStar(false),
        }
    }

    lazy_static! {
        // GROUP BY KEYS
        static ref AST_SUBPATH: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
            expr: ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                subpath: "a".to_string()
            }),
            alias: "key".to_string(),
        });

        // 1 AS literal
        static ref AST_LITERAL_KEY: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
            expr: ast::Expression::Literal(ast::Literal::Integer(1)),
            alias: "literal".into(),
        });

        // a + 1 AS complex_expr
        static ref AST_SUBPATH_COMPLEX_EXPR: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
            expr: ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))
            }),
            alias: "complex_expr".into(),
        });

        // AGGREGATION FUNCTIONS

        // AVG(DISTINCT arr.a) AS agg1
        static ref AST_AGG_1_ARRAY: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Avg,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                        subpath: "a".to_string()
                    })
                ]),
                set_quantifier: Some(ast::SetQuantifier::Distinct),
            }),
            alias: "agg1".to_string(),
        };

        // COUNT(*) AS agg2
        static ref AST_AGG_2: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Count,
                args: ast::FunctionArguments::Star,
                set_quantifier: None
            }),
            alias: "agg2".to_string(),
        };
    }

    // Successful tests.

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE AVG(DISTINCT arr.a) AS agg1, COUNT(*) AS agg2
    test_algebrize!(
        group_by_key_with_aggregation_array_source,
        method = algebrize_group_by_clause,
        expected = Ok(ir::Stage::Group(ir::Group {
            source: Box::new(ir_array_source()),
            keys: vec![ir_field_access()],
            aggregations: vec![ir_agg_1_array(), ir_agg_2()],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![AST_AGG_1_ARRAY.clone(), AST_AGG_2.clone()],
        }),
        source = ir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY 1
    test_algebrize!(
        group_by_key_is_literal,
        method = algebrize_group_by_clause,
        expected = Ok(ir::Stage::Group(ir::Group {
            source: Box::new(ir_array_source()),
            keys: vec![ir_literal_key()],
            aggregations: vec![],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::GroupByClause {
            keys: vec![AST_LITERAL_KEY.clone()],
            aggregations: vec![],
        }),
        source = ir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY a + 1
    test_algebrize!(
        group_by_key_is_complex_expression,
        method = algebrize_group_by_clause,
        expected = Ok(ir::Stage::Group(ir::Group {
            source: Box::new(ir_array_source()),
            keys: vec![ir_field_access_complex_expr()],
            aggregations: vec![],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH_COMPLEX_EXPR.clone()],
            aggregations: vec![],
        }),
        source = ir_array_source(),
    );

    // Error tests.

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE 42 AS agg
    test_algebrize!(
        group_by_key_with_non_function_aggregation_expression,
        method = algebrize_group_by_clause,
        expected = Err(Error::NonAggregationInPlaceOfAggregation(0)),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![ast::AliasedExpr {
                expr: ast::Expression::Literal(ast::Literal::Integer(42)),
                alias: "agg".to_string(),
            },],
        }),
        source = ir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key, arr.a AS key
    test_algebrize!(
        group_by_keys_must_have_unique_aliases,
        method = algebrize_group_by_clause,
        expected = Err(Error::DuplicateDocumentKey("key".into())),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone(), AST_SUBPATH.clone()],
            aggregations: vec![],
        }),
        source = ir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*) AS a, COUNT(*) AS a
    test_algebrize!(
        group_by_aggregations_must_have_unique_aliases,
        method = algebrize_group_by_clause,
        expected = Err(Error::DuplicateDocumentKey("a".into())),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![
                ast::AliasedExpr {
                    expr: ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::Count,
                        args: ast::FunctionArguments::Star,
                        set_quantifier: None
                    }),
                    alias: "a".into(),
                },
                ast::AliasedExpr {
                    expr: ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::Count,
                        args: ast::FunctionArguments::Star,
                        set_quantifier: None
                    }),
                    alias: "a".into(),
                },
            ],
        }),
        source = ir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*) AS key
    test_algebrize!(
        group_by_aliases_must_be_unique_across_keys_and_aggregates,
        method = algebrize_group_by_clause,
        expected = Err(Error::DuplicateDocumentKey("key".into())),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![ast::AliasedExpr {
                expr: ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::Count,
                    args: ast::FunctionArguments::Star,
                    set_quantifier: None
                }),
                alias: "key".into(),
            },],
        }),
        source = ir_array_source(),
    );
}

mod subquery {
    use super::catalog;
    use crate::{
        ast,
        ir::{binding_tuple::DatasourceName, schema::SchemaCache, *},
        map, multimap,
        schema::{Atomic, Document, Schema},
        set, unchecked_unique_linked_hash_map,
    };
    use lazy_static::lazy_static;

    fn ir_array() -> Stage {
        Stage::Array(ArraySource {
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                }
                .into(),
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })
    }
    lazy_static! {
        static ref AST_ARRAY: ast::Datasource = ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
            },)],
            alias: "arr".into()
        });
    }
    test_algebrize!(
        uncorrelated_exists,
        method = algebrize_expression,
        expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(ir_array()),
            expression: map! {
                (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                    "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                }.into())
            },
            cache: SchemaCache::new(),
        })).into())),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        correlated_exists,
        method = algebrize_expression,
        expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(ir_array()),
            expression: map! {
                (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                    "b_0".into() => Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        cache: SchemaCache::new(),
                    })
                }.into())
            },
            cache: SchemaCache::new(),
        })).into())),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"b".to_string()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize!(
        exists_cardinality_gt_1,
        method = algebrize_expression,
        expected = Ok(Expression::Exists(Box::new(Stage::Array(ArraySource {
            array: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                .into()),
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                .into())
            ],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })).into())),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    },),
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                    },)
                ],
                alias: "arr".into()
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        exists_degree_gt_1,
        method = algebrize_expression,
        expected = Ok(Expression::Exists(
            Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".to_string() => Expression::Literal(LiteralValue::Integer(1).into()),
                        "b".to_string() => Expression::Literal(LiteralValue::Integer(2).into())
                    }
                    .into()
                )],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }))
            .into()
        )),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                },),],
                alias: "arr".into()
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        uncorrelated_subquery_expr,
        method = algebrize_expression,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(ir_array()),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        correlated_subquery_expr,
        method = algebrize_expression,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
                field: "b_0".to_string(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(ir_array()),
                expression: map! {
                    (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "b_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                            field: "b".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"b".to_string()},
                additional_properties: false,
            })
        },
    );
    test_algebrize!(
        degree_zero_unsat_output,
        method = algebrize_expression,
        expected = Err(Error::InvalidSubqueryDegree),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![],
                alias: "arr".into()
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        substar_degree_eq_1,
        method = algebrize_expression,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                field: "a".to_string(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(ir_array()),
                expression: map! {
                    ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                    ast::SubstarExpr {
                        datasource: "arr".into()
                    }
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        select_values_degree_gt_1,
        method = algebrize_expression,
        expected = Err(Error::InvalidSubqueryDegree),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into()),
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                ),])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    },),
                    ast::Expression::Document(multimap! {
                        "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                    },)
                ],
                alias: "arr".into()
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        star_degree_eq_1,
        method = algebrize_expression,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                field: "a".to_string(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(ir_array()),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        select_star_degree_gt_1,
        method = algebrize_expression,
        expected = Err(Error::InvalidSubqueryDegree),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                })],
                alias: "arr".into()
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        substar_degree_gt_1,
        method = algebrize_expression,
        expected = Err(Error::InvalidSubqueryDegree),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                    ast::SubstarExpr {
                        datasource: "arr".into()
                    }
                )])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                })],
                alias: "arr".into()
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        uncorrelated_subquery_comparison_all,
        method = algebrize_expression,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5).into())),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "a_0".to_string(),
                    cache: SchemaCache::new(),
                })),
                subquery: Box::new(Stage::Project(Project {
                    source: Box::new(ir_array()),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                                field: "a".into(),
                                cache: SchemaCache::new(),
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(5))),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::All,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        uncorrelated_subquery_comparison_any,
        method = algebrize_expression,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::Any,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5).into())),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "a_0".to_string(),
                    cache: SchemaCache::new(),
                })),
                subquery: Box::new(Stage::Project(Project {
                    source: Box::new(ir_array()),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                                field: "a".into(),
                                cache: SchemaCache::new(),
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(5))),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::Any,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        argument_from_super_scope,
        method = algebrize_expression,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "b".to_string(),
                cache: SchemaCache::new(),
            })),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
                    field: "a_0".to_string(),
                    cache: SchemaCache::new(),
                })),
                subquery: Box::new(Stage::Project(Project {
                    source: Box::new(ir_array()),
                    expression: map! {
                        (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 2u16).into())),
                                field: "a".into(),
                                cache: SchemaCache::new(),
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Identifier("b".into())),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::All,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"b".to_string()},
                additional_properties: false,
            })
        },
    );
    test_algebrize!(
        argument_only_evaluated_in_super_scope,
        method = algebrize_expression,
        expected = Err(Error::FieldNotFound("a".into())),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::All,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        potentially_missing_column,
        method = algebrize_expression,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "x".to_string(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Limit(Limit {
                source: Box::new(Stage::Project(Project {
                    source: Box::new(Stage::Project(Project {
                        source: Box::new(Stage::Collection(Collection {
                            db: "test".to_string(),
                            collection: "bar".to_string(),
                            cache: SchemaCache::new(),
                        })),
                        expression: map! {
                            (DatasourceName::Named("bar".to_string()), 1u16).into() => Expression::Reference(("bar".to_string(), 1u16).into())
                        },
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "x".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("bar", 1u16).into())),
                                field: "x".into(),
                                cache: SchemaCache::new(),
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                limit: 1,
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "x".into() => ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("bar".into())),
                            subpath: "x".to_string()
                        })
                    })
                )])
            },
            from_clause: Some(ast::Datasource::Collection(ast::CollectionSource {
                database: None,
                collection: "bar".to_string(),
                alias: Some("bar".to_string()),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(1),
            offset: None,
        }))),
        catalog = catalog(vec![("test", "bar")]),
    );
}

mod schema_checking_mode {
    use super::catalog;
    use crate::{
        ast,
        ir::{schema::SchemaCache, *},
        Schema,
    };

    test_algebrize!(
        comparison_fails_in_strict_mode,
        method = algebrize_order_by_clause,
        expected = Err(Error::SchemaChecking(
            schema::Error::SortKeyNotSelfComparable(0, Schema::Any)
        )),
        input = Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Identifier("a".into())),
                direction: ast::SortDirection::Asc
            }]
        }),
        source = Stage::Collection(Collection {
            db: "".into(),
            collection: "test".into(),
            cache: SchemaCache::new(),
        }),
        catalog = catalog(vec![("", "test")]),
    );

    test_algebrize!(
        comparison_passes_in_relaxed_mode,
        method = algebrize_order_by_clause,
        expected_pat = Ok(_),
        input = Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Identifier("a".into())),
                direction: ast::SortDirection::Asc
            }]
        }),
        source = Stage::Collection(Collection {
            db: "".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        }),
        catalog = catalog(vec![("", "foo")]),
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );
}
