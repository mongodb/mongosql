use crate::{
    ast::{self, CollectionSource, Datasource},
    ir::{Collection, Expression, Project, Stage},
    map,
};
use lazy_static::lazy_static;

macro_rules! test_algebrize {
    ($func_name:ident, $method:ident, $expected:expr, $ast:expr $(,)?) => {
        #[test]
        fn $func_name() {
            use crate::algebrizer::{Algebrizer, Error};
            let algebrizer = Algebrizer::new("test".into(), 0u16);
            let expected: Result<_, Error> = $expected;
            let res: Result<_, Error> = algebrizer.$method($ast);
            assert_eq!(expected, res);
        }
    };
    ($func_name:ident, $method:ident, $expected:expr, $ast:expr, $source:expr $(,)?) => {
        #[test]
        fn $func_name() {
            use crate::algebrizer::{Algebrizer, Error};
            let algebrizer = Algebrizer::new("test".into(), 0u16);
            let expected: Result<_, Error> = $expected;
            let res: Result<_, Error> = algebrizer.$method($ast, $source);
            assert_eq!(expected, res);
        }
    };
    ($func_name:ident, $method:ident, source = $source:expr, expected = $expected:expr, input = $ast:expr, env = $env:expr $(,)?) => {
        #[test]
        fn $func_name() {
            use crate::algebrizer::{Algebrizer, Error};
            let algebrizer = Algebrizer::with_schema_env("test".into(), $env, 1u16);
            let expected: Result<_, Error> = $expected;
            let res: Result<_, Error> = algebrizer.$method($ast, $source);
            assert_eq!(expected, res);
        }
    };
}

macro_rules! test_algebrize_with_env {
    ($func_name:ident, $method:ident, $expected:expr, $ast:expr, $env:expr $(,)?) => {
        #[test]
        fn $func_name() {
            use crate::{
                algebrizer::{Algebrizer, Error},
                ast,
            };
            let algebrizer = Algebrizer::with_schema_env("test".into(), $env, 1u16);
            let expected: Result<_, Error> = $expected;
            let res: Result<_, Error> = algebrizer.$method($ast);
            assert_eq!(expected, res);
        }
    };
}

lazy_static! {
    static ref IR_SOURCE_FOO: Stage = Stage::Project(Project {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into()
        })),
        expression: map! {
            ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into())
        }
    });
    static ref IR_SOURCE_BAR: Stage = Stage::Project(Project {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into()
        })),
        expression: map! {
            ("bar", 0u16).into() => Expression::Reference(("bar", 0u16).into())
        }
    });
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
        ir::{self, binding_tuple::Key},
        map,
        schema::{
            Atomic, Document, Schema, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, NUMERIC_OR_NULLISH,
            STRING_OR_NULLISH,
        },
        set,
    };

    test_algebrize!(
        null,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::Null)),
        ast::Expression::Literal(ast::Literal::Null),
    );
    test_algebrize!(
        expr_true,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::Boolean(true))),
        ast::Expression::Literal(ast::Literal::Boolean(true)),
    );
    test_algebrize!(
        expr_false,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::Boolean(false))),
        ast::Expression::Literal(ast::Literal::Boolean(false)),
    );
    test_algebrize!(
        string,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::String(
            "hello!".into()
        ))),
        ast::Expression::Literal(ast::Literal::String("hello!".into())),
    );
    test_algebrize!(
        int,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::Integer(42))),
        ast::Expression::Literal(ast::Literal::Integer(42)),
    );
    test_algebrize!(
        long,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::Long(42))),
        ast::Expression::Literal(ast::Literal::Long(42)),
    );
    test_algebrize!(
        double,
        algebrize_expression,
        Ok(ir::Expression::Literal(ir::Literal::Double(42f64))),
        ast::Expression::Literal(ast::Literal::Double(42f64)),
    );

    test_algebrize!(
        empty_array,
        algebrize_expression,
        Ok(ir::Expression::Array(vec![])),
        ast::Expression::Array(vec![]),
    );
    test_algebrize!(
        nested_array,
        algebrize_expression,
        Ok(ir::Expression::Array(vec![ir::Expression::Array(vec![
            ir::Expression::Literal(ir::Literal::Long(42)),
            ir::Expression::Literal(ir::Literal::Integer(42)),
        ])])),
        ast::Expression::Array(vec![ast::Expression::Array(vec![
            ast::Expression::Literal(ast::Literal::Long(42)),
            ast::Expression::Literal(ast::Literal::Integer(42)),
        ])]),
    );

    test_algebrize!(
        empty_document,
        algebrize_expression,
        Ok(ir::Expression::Document(map! {})),
        ast::Expression::Document(map! {}),
    );
    test_algebrize!(
        nested_document,
        algebrize_expression,
        Ok(ir::Expression::Document(map! {
            "foo2".into() => ir::Expression::Document(
                map!{"nested".into() => ir::Expression::Literal(ir::Literal::Integer(52))},
            ),
            "bar2".into() => ir::Expression::Literal(ir::Literal::Integer(42))
        },)),
        ast::Expression::Document(map! {
                    "foo2".into() => ast::Expression::Document(
                        map!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                    ),
                    "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
        }),
    );
    test_algebrize_with_env!(
        qualified_ref_in_current_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 1u16).into())),
            field: "a".into(),
        })),
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("foo".into())),
            subpath: "a".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        qualified_ref_in_super_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
        })),
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("foo".into())),
            subpath: "a".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        unqualified_ref_may_exist_in_current_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 1u16).into())),
            field: "a".into(),
        })),
        ast::Expression::Identifier("a".into()),
        map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
            }),
        },
    );
    test_algebrize_with_env!(
        unqualified_ref_must_exist_in_current_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 1u16).into())),
            field: "a".into(),
        })),
        ast::Expression::Identifier("a".into()),
        map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_algebrize_with_env!(
        unqualified_ref_may_exist_only_in_super_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
        })),
        ast::Expression::Identifier("a".into()),
        map! {
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
    test_algebrize_with_env!(
        unqualified_ref_must_exist_in_super_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
        })),
        ast::Expression::Identifier("a".into()),
        map! {
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
    test_algebrize_with_env!(
        unqualified_ref_must_exist_in_super_scope_bot_source,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(Key::bot(0u16))),
            field: "a".into(),
        })),
        ast::Expression::Identifier("a".into()),
        map! {
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
    test_algebrize_with_env!(
        unqualified_ref_may_and_must_exist_in_two_sources,
        algebrize_expression,
        Err(Error::AmbiguousField("a".into())),
        ast::Expression::Identifier("a".into()),
        map! {
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
    test_algebrize_with_env!(
        unqualified_subpath_in_current_and_super_must_exist_in_current,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("test", 1u16).into())),
                field: "a".into(),
            })),
            field: "c".into(),
        })),
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        unqualified_subpath_in_current_and_super_may_exist_is_ambiguous,
        algebrize_expression,
        Err(Error::AmbiguousField("a".into())),
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        unqualified_subpath_in_super_scope,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("super_test", 0u16).into())),
                field: "a".into(),
            })),
            field: "c".into(),
        })),
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        qualified_ref_prefers_super_datasource_to_local_field,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
            field: "a".into(),
        })),
        // test MongoSQL: SELECT (SELECT foo.a FROM bar) FROM foo => (foo.a)
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("foo".into())),
            subpath: "a".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        qualified_ref_to_local_field,
        algebrize_expression,
        Ok(ir::Expression::FieldAccess(ir::FieldAccess {
            expr: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("bar", 1u16).into())),
                field: "foo".into(),
            })),
            field: "a".into(),
        })),
        //test MongoSQL: SELECT (SELECT bar.foo.a FROM bar) FROM foo => (bar.foo.a)
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("bar".into())),
                subpath: "foo".into(),
            })),
            subpath: "a".into(),
        }),
        map! {
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
    test_algebrize_with_env!(
        unqualified_reference_and_may_contain_sub_and_must_contain_outer_is_ambiguous,
        algebrize_expression,
        Err(Error::AmbiguousField("a".into())),
        ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            subpath: "c".into(),
        }),
        map! {
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
        algebrize_expression,
        Err(Error::FieldNotFound("bar".into())),
        ast::Expression::Identifier("bar".into()),
    );

    test_algebrize!(
        add_bin_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Add,
                args: vec![
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                ],
            }
        )),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Add,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        add_wrong_types,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Add,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        sub_bin_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Sub,
                args: vec![
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                ],
            }
        )),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Sub,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        sub_wrong_types,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Sub",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Sub,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        div_bin_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Div,
                args: vec![
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                ],
            }
        )),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Div,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        div_wrong_types,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Div",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Div,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        mul_bin_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Mul,
                args: vec![
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                    ir::Expression::Literal(ir::Literal::Integer(42)),
                ],
            }
        )),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Mul,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        mul_wrong_types,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Mul",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Mul,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        concat_bin_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Concat,
                args: vec![
                    ir::Expression::Literal(ir::Literal::String("42".into())),
                    ir::Expression::Literal(ir::Literal::String("42".into())),
                ],
            }
        )),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            op: ast::BinaryOp::Concat,
            right: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
        }),
    );
    test_algebrize!(
        concat_wrong_types,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Concat",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            op: ast::BinaryOp::Concat,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        neg_un_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Neg,
                args: vec![ir::Expression::Literal(ir::Literal::Integer(42)),],
            }
        )),
        ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        neg_wrong_type,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Neg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Boolean),
        })),
        ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        }),
    );

    test_algebrize!(
        pos_un_op,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Pos,
                args: vec![ir::Expression::Literal(ir::Literal::Integer(42)),],
            }
        )),
        ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Pos,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );
    test_algebrize!(
        pos_wrong_type,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Pos",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Boolean),
        })),
        ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Pos,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        }),
    );

    test_algebrize!(
        standard_scalar_function,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Lower,
                args: vec![ir::Expression::Literal(ir::Literal::String("hello".into())),],
            }
        )),
        ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Lower,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("hello".into(),)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        })
    );
    test_algebrize!(
        ltrim,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::LTrim,
                args: vec![
                    ir::Expression::Literal(ir::Literal::String("hello".into())),
                    ir::Expression::Literal(ir::Literal::String("hello world".into()))
                ]
            }
        )),
        ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Leading,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello".into()
            ))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello world".into()
            ))),
        })
    );
    test_algebrize!(
        rtrim,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::RTrim,
                args: vec![
                    ir::Expression::Literal(ir::Literal::String("world".into())),
                    ir::Expression::Literal(ir::Literal::String("hello world".into()))
                ]
            }
        )),
        ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Trailing,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(
                "world".into()
            ))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(
                "hello world".into()
            ))),
        })
    );
    test_algebrize!(
        btrim,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::BTrim,
                args: vec![
                    ir::Expression::Literal(ir::Literal::String(" ".into())),
                    ir::Expression::Literal(ir::Literal::String(" hello world ".into()))
                ]
            }
        )),
        ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(" ".into()))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(
                " hello world ".into()
            ))),
        })
    );
    test_algebrize!(
        trim_arg_must_be_string_or_null,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "BTrim",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::String(" ".into()))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        })
    );
    test_algebrize!(
        trim_escape_must_be_string,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "BTrim",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            arg: Box::new(ast::Expression::Literal(ast::Literal::String(" ".into()))),
        })
    );

    test_algebrize!(
        extract_year,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Year,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![]
                    }
                ),]
            }
        )),
        ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::ExtractSpec::Year,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_day,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Day,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![]
                    }
                ),]
            }
        )),
        ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::ExtractSpec::Day,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_hour,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Hour,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![]
                    }
                ),]
            }
        )),
        ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::ExtractSpec::Hour,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_minute,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Minute,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![]
                    }
                ),]
            }
        )),
        ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::ExtractSpec::Minute,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_second,
        algebrize_expression,
        Ok(ir::Expression::ScalarFunction(
            ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Second,
                args: vec![ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::CurrentTimestamp,
                        args: vec![]
                    }
                ),]
            }
        )),
        ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::ExtractSpec::Second,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );
    test_algebrize!(
        extract_must_be_date,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Second",
            required: DATE_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        })),
        ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::ExtractSpec::Second,
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        searched_case,
        algebrize_expression,
        Ok(ir::Expression::SearchedCase(ir::SearchedCaseExpr {
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(ir::Literal::Boolean(true))),
                then: Box::new(ir::Expression::Literal(ir::Literal::String("bar".into()))),
            }],
            else_branch: Box::new(ir::Expression::Literal(ir::Literal::String("foo".into()))),
        })),
        ast::Expression::Case(ast::CaseExpr {
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
        algebrize_expression,
        Ok(ir::Expression::SearchedCase(ir::SearchedCaseExpr {
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(ir::Literal::Boolean(true))),
                then: Box::new(ir::Expression::Literal(ir::Literal::String("bar".into()))),
            }],
            else_branch: Box::new(ir::Expression::Literal(ir::Literal::Null)),
        })),
        ast::Expression::Case(ast::CaseExpr {
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
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "SearchedCase",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::Expression::Case(ast::CaseExpr {
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
        algebrize_expression,
        Ok(ir::Expression::SimpleCase(ir::SimpleCaseExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::Integer(1))),
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(ir::Literal::Integer(2))),
                then: Box::new(ir::Expression::Literal(ir::Literal::String("bar".into()))),
            }],
            else_branch: Box::new(ir::Expression::Literal(ir::Literal::String("foo".into()))),
        })),
        ast::Expression::Case(ast::CaseExpr {
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
        algebrize_expression,
        Ok(ir::Expression::SimpleCase(ir::SimpleCaseExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::Integer(1))),
            when_branch: vec![ir::WhenBranch {
                when: Box::new(ir::Expression::Literal(ir::Literal::Integer(2))),
                then: Box::new(ir::Expression::Literal(ir::Literal::String("bar".into()))),
            }],
            else_branch: Box::new(ir::Expression::Literal(ir::Literal::Null)),
        })),
        ast::Expression::Case(ast::CaseExpr {
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
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::InvalidComparison(
            "SimpleCase",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        ast::Expression::Case(ast::CaseExpr {
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
        algebrize_expression,
        Ok(ir::Expression::Cast(ir::CastExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::Integer(42))),
            to: ir::Type::String,
            on_null: Box::new(ir::Expression::Literal(ir::Literal::String(
                "was_null".into()
            ))),
            on_error: Box::new(ir::Expression::Literal(ir::Literal::String(
                "was_error".into()
            ))),
        })),
        ast::Expression::Cast(ast::CastExpr {
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
        algebrize_expression,
        Ok(ir::Expression::Cast(ir::CastExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::Integer(42))),
            to: ir::Type::String,
            on_null: Box::new(ir::Expression::Literal(ir::Literal::Null)),
            on_error: Box::new(ir::Expression::Literal(ir::Literal::Null)),
        })),
        ast::Expression::Cast(ast::CastExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            to: ast::Type::String,
            on_null: None,
            on_error: None,
        }),
    );

    test_algebrize!(
        type_assert_success,
        algebrize_expression,
        Ok(ir::Expression::TypeAssertion(ir::TypeAssertionExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::Integer(42))),
            target_type: ir::Type::Int32,
        })),
        ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::Type::Int32,
        }),
    );
    test_algebrize!(
        type_assert_fail,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "::!",
            required: Schema::Atomic(Atomic::String),
            found: Schema::Atomic(Atomic::Integer)
        })),
        ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::Type::String,
        }),
    );

    test_algebrize!(
        is_success,
        algebrize_expression,
        Ok(ir::Expression::Is(ir::IsExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::Integer(42))),
            target_type: ir::TypeOrMissing::Type(ir::Type::Int32),
        })),
        ast::Expression::Is(ast::IsExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
        }),
    );
    test_algebrize!(
        is_recursive_failure,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String)
        })),
        ast::Expression::Is(ast::IsExpr {
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
        algebrize_expression,
        Ok(ir::Expression::Like(ir::LikeExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::String("42".into()))),
            pattern: Box::new(ir::Expression::Literal(ir::Literal::String("42".into()))),
            escape: Some("foo".into()),
        })),
        ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            escape: Some("foo".into()),
        }),
    );
    test_algebrize!(
        like_success_no_pattern,
        algebrize_expression,
        Ok(ir::Expression::Like(ir::LikeExpr {
            expr: Box::new(ir::Expression::Literal(ir::Literal::String("42".into()))),
            pattern: Box::new(ir::Expression::Literal(ir::Literal::String("42".into()))),
            escape: None,
        })),
        ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            escape: None,
        }),
    );
    test_algebrize!(
        like_expr_must_be_string,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer)
        })),
        ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            escape: Some(" ".into()),
        }),
    );
    test_algebrize!(
        like_pattern_must_be_string,
        algebrize_expression,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer)
        })),
        ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::String("42".into()))),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            escape: Some(" ".into()),
        }),
    );
}

mod aggregation {
    use crate::{
        ast, ir, map,
        schema::{Atomic, Schema, ANY_DOCUMENT, NUMERIC_OR_NULLISH},
    };
    test_algebrize!(
        count_star,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::CountStar(false)),
        ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        count_distinct_star_is_error,
        algebrize_aggregation,
        Err(Error::SchemaChecking(
            ir::schema::Error::CountDistinctStarNotSupported
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );
    test_algebrize!(
        count_all_expr_basic_test,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Count,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        count_distinct_expr_basic_test,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Count,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );
    test_algebrize_with_env!(
        count_distinct_expr_argument_not_self_comparable_is_error,
        algebrize_aggregation,
        Err(Error::SchemaChecking(
            ir::schema::Error::AggregationArgumentMustBeSelfComparable(
                "Count DISTINCT".into(),
                Schema::Any
            )
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Identifier("foo".into())]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
        map! {
            ("d", 1u16).into() => ANY_DOCUMENT.clone(),
        }
    );
    test_algebrize!(
        sum_star_is_error,
        algebrize_aggregation,
        Err(Error::StarInNonCount),
        ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        sum_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Sum,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        sum_distinct_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Sum,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );
    test_algebrize!(
        sum_argument_must_be_numeric,
        algebrize_aggregation,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Sum",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        avg_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Avg,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        avg_distinct_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Avg,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        avg_argument_must_be_numeric,
        algebrize_aggregation,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "Avg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        stddevpop_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevPop,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        stddevpop_distinct_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevPop,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );
    test_algebrize!(
        stddevpop_argument_must_be_numeric,
        algebrize_aggregation,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "StddevPop",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        stddevsamp_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevSamp,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        stddevsamp_distinct_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::StddevSamp,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );
    test_algebrize!(
        stddevsamp_argument_must_be_numeric,
        algebrize_aggregation,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "StddevSamp",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        addtoarray_expr_basic_test,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::AddToArray,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        addtoarray_distinct_expr_basic_test,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::AddToArray,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        addtoset_expr_is_addtoarray_distinct_in_ir,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::AddToSet,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        addtoset_distinct_expr_is_addtoarray_in_ir,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::AddToArray,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::AddToSet,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        first_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::First,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::First,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        first_distinct_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::First,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::First,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        last_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Last,
                distinct: false,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Last,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        last_distinct_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Last,
                distinct: true,
                arg: ir::Expression::Literal(ir::Literal::Integer(42)).into(),
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::Last,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }
    );

    test_algebrize!(
        mergedocuments_expr,
        algebrize_aggregation,
        Ok(ir::AggregationExpr::Function(
            ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::MergeDocuments,
                distinct: false,
                arg: Box::new(ir::Expression::Document(map! {
                    "a".into() => ir::Expression::Literal(ir::Literal::Integer(42)),
                    "b".into() => ir::Expression::Literal(ir::Literal::Integer(42)),
                },))
            }
        )),
        ast::FunctionExpr {
            function: ast::FunctionName::MergeDocuments,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Document(map! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            })]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
    test_algebrize!(
        mergedocuments_argument_must_be_document,
        algebrize_aggregation,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "MergeDocuments",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        ast::FunctionExpr {
            function: ast::FunctionName::MergeDocuments,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::String("42".into())
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }
    );
}

mod select_clause {
    use crate::{
        ast,
        ir::{self, binding_tuple::Key},
        map,
        schema::ANY_DOCUMENT,
    };
    use lazy_static::lazy_static;

    lazy_static! {
        static ref SOURCE: ir::Stage = ir::Stage::Collection(ir::Collection {
            db: "test".into(),
            collection: "baz".into(),
        });
    }

    test_algebrize!(
        select_distinct_not_allowed,
        algebrize_select_clause,
        source = SOURCE.clone(),
        expected = Err(Error::DistinctSelect),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::Distinct,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Identifier("foo".into())
            ),]),
        },
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        select_duplicate_bot,
        algebrize_select_clause,
        source = SOURCE.clone(),
        expected = Err(Error::DuplicateKey(Key::bot(1u16))),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Expression(ast::Expression::Document(map! {},)),
                ast::SelectValuesExpression::Expression(ast::Expression::Document(map! {},)),
            ]),
        },
        env = map! {},
    );
    test_algebrize!(
        select_bot_and_double_substar,
        algebrize_select_clause,
        source = SOURCE.clone(),
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(SOURCE.clone()),
            expression: map! {
                ("bar", 1u16).into() => ir::Expression::Reference(("bar", 1u16).into()),
                Key::bot(1u16) => ir::Expression::Document(map!{}),
                ("foo", 1u16).into() => ir::Expression::Reference(("foo", 0u16).into()),
            }
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Substar("bar".into()),
                ast::SelectValuesExpression::Expression(ast::Expression::Document(map! {},)),
                ast::SelectValuesExpression::Substar("foo".into()),
            ]),
        },
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            ("bar", 1u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        select_value_expression_must_be_document,
        algebrize_select_clause,
        source = SOURCE.clone(),
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
        env = map! {},
    );
    test_algebrize!(
        select_duplicate_substar,
        algebrize_select_clause,
        source = SOURCE.clone(),
        expected = Err(Error::DuplicateKey(("foo", 1u16).into())),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Substar("foo".into()),
                ast::SelectValuesExpression::Substar("foo".into()),
            ]),
        },
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        select_substar_body,
        algebrize_select_clause,
        source = SOURCE.clone(),
        expected = Ok(ir::Stage::Project(ir::Project {
            source: Box::new(SOURCE.clone()),
            expression: map! {
                ("foo", 1u16).into() => ir::Expression::Reference(("foo", 0u16).into()),
            }
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(
                vec![ast::SelectValuesExpression::Substar("foo".into()),]
            ),
        },
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
    );
}

mod from_clause {
    use super::{AST_SOURCE_BAR, AST_SOURCE_FOO, IR_SOURCE_BAR, IR_SOURCE_FOO};
    use crate::{
        ast::{self, JoinSource},
        ir::{self, binding_tuple::Key, JoinType},
        map,
        schema::{Atomic, Document, Schema, ANY_DOCUMENT},
        set,
    };

    test_algebrize!(
        from_clause_must_exist,
        algebrize_from_clause,
        Err(Error::NoFromClause),
        None,
    );
    test_algebrize!(
        collection_must_have_alias,
        algebrize_from_clause,
        Err(Error::CollectionMustHaveAlias),
        Some(ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: None,
        })),
    );
    test_algebrize!(
        basic_collection,
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Collection(ir::Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    ir::Expression::Reference(("foo", 0u16).into())
            }
        },),),
        Some(ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: Some("bar".into()),
        })),
    );
    test_algebrize!(
        qualified_collection,
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Collection(ir::Collection {
                db: "test2".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    ir::Expression::Reference(("foo", 0u16).into())
            }
        }),),
        Some(ast::Datasource::Collection(ast::CollectionSource {
            database: Some("test2".into()),
            collection: "foo".into(),
            alias: Some("bar".into()),
        })),
    );
    test_algebrize!(
        empty_array,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![],
            alias: "bar".into(),
        }),),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        dual,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![ir::Expression::Document(map! {})],
            alias: "_dual".into(),
        }),),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {},)],
            alias: "_dual".into(),
        })),
    );
    test_algebrize!(
        int_array,
        algebrize_from_clause,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
        })),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Integer(42))],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        null_array,
        algebrize_from_clause,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Null)]),
        })),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Null)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        array_datasource_must_be_constant,
        algebrize_from_clause,
        Err(Error::ArrayDatasourceMustBeLiteral),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {
                "foo".into() => ast::Expression::Identifier("foo".into()),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        single_document_array,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![ir::Expression::Document(map! {
                "foo".into() => ir::Expression::Literal(ir::Literal::Integer(1)),
                "bar".into() => ir::Expression::Literal(ir::Literal::Integer(1))
            })],
            alias: "bar".into(),
        }),),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        two_document_array,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![
                ir::Expression::Document(map! {
                    "foo".into() => ir::Expression::Literal(ir::Literal::Integer(1)),
                    "bar".into() => ir::Expression::Literal(ir::Literal::Integer(1))
                }),
                ir::Expression::Document(map! {
                    "foo2".into() => ir::Expression::Literal(ir::Literal::Integer(41)),
                    "bar2".into() => ir::Expression::Literal(ir::Literal::Integer(42))
                },)
            ],
            alias: "bar".into(),
        }),),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(map! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                }),
                ast::Expression::Document(map! {
                    "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(41)),
                    "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                },)
            ],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        two_document_with_nested_document_array,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![
                ir::Expression::Document(map! {
                    "foo".into() => ir::Expression::Literal(ir::Literal::Integer(1)),
                    "bar".into() => ir::Expression::Literal(ir::Literal::Integer(1))
                }),
                ir::Expression::Document(map! {
                    "foo2".into() => ir::Expression::Document(
                        map!{"nested".into() => ir::Expression::Literal(ir::Literal::Integer(52))},
                    ),
                    "bar2".into() => ir::Expression::Literal(ir::Literal::Integer(42))
                },)
            ],
            alias: "bar".into(),
        }),),
        Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(map! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                }),
                ast::Expression::Document(map! {
                    "foo2".into() => ast::Expression::Document(
                        map!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                    ),
                    "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                },)
            ],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        left_join,
        algebrize_from_clause,
        Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Left,
            left: Box::new(IR_SOURCE_FOO.clone()),
            right: Box::new(IR_SOURCE_BAR.clone()),
            condition: Some(ir::Expression::Literal(ir::Literal::Boolean(true)))
        })),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Left,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: Some(ast::Expression::Literal(ast::Literal::Boolean(true)))
        }))
    );
    test_algebrize!(
        right_join,
        algebrize_from_clause,
        Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Left,
            left: Box::new(IR_SOURCE_BAR.clone()),
            right: Box::new(IR_SOURCE_FOO.clone()),
            condition: Some(ir::Expression::Literal(ir::Literal::Boolean(true)))
        })),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Right,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: Some(ast::Expression::Literal(ast::Literal::Boolean(true)))
        }))
    );
    test_algebrize!(
        left_outer_join_without_condition,
        algebrize_from_clause,
        Err(Error::NoOuterJoinCondition),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Left,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        }))
    );
    test_algebrize!(
        right_outer_join_without_condition,
        algebrize_from_clause,
        Err(Error::NoOuterJoinCondition),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Right,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        }))
    );
    test_algebrize!(
        inner_join,
        algebrize_from_clause,
        Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Inner,
            left: Box::new(IR_SOURCE_FOO.clone()),
            right: Box::new(IR_SOURCE_BAR.clone()),
            condition: None
        })),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Inner,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        }))
    );
    test_algebrize!(
        cross_join,
        algebrize_from_clause,
        Ok(ir::Stage::Join(ir::Join {
            join_type: JoinType::Inner,
            left: Box::new(IR_SOURCE_FOO.clone()),
            right: Box::new(IR_SOURCE_BAR.clone()),
            condition: None
        })),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        }))
    );
    test_algebrize!(
        invalid_join_condition,
        algebrize_from_clause,
        Err(Error::FieldNotFound("bar".into())),
        Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: Some(ast::Expression::Identifier("bar".into())),
        }))
    );

    test_algebrize!(
        derived_single_datasource,
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Array(ir::Array {
                array: vec![ir::Expression::Document(
                    map! {"foo".into() => ir::Expression::Literal(ir::Literal::Integer(1)),
                         "bar".into() => ir::Expression::Literal(ir::Literal::Integer(1))
                    }
                )],
                alias: "bar".into()
            })),
            expression: map! {("d", 0u16).into() =>
                ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::MergeObjects,
                    args: vec![
                        ir::Expression::Reference(("bar", 1u16).into())
                    ]
                }
                )
            },
        })),
        Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
                },
                from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(map! {
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
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Project(ir::Project {
                source: Box::new(ir::Stage::Array(ir::Array {
                    array: vec![ir::Expression::Document(
                        map! {"foo".into() => ir::Expression::Literal(
                            ir::Literal::Integer(1)
                            ),
                            "bar".into() => ir::Expression::Literal(
                                ir::Literal::Integer(1))
                        }
                    )],
                    alias: "bar".into()
                })),
                expression: map! {
                    Key::bot(1u16) => ir::Expression::Document(
                        map!{"baz".into() => ir::Expression::Literal(ir::Literal::String("hello".into()))}
                    ),
                    ("bar", 1u16).into() =>
                        ir::Expression::Reference(("bar", 1u16).into())
                }
            })),
            expression: map! { ("d", 0u16).into() =>
                ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::MergeObjects,
                        args:
                            vec![
                                ir::Expression::Reference(Key::bot(1u16)),
                                ir::Expression::Reference(("bar", 1u16).into()),
                            ],
                    }
                )
            }
        })),
        Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![
                        ast::SelectValuesExpression::Substar("bar".into()),
                        ast::SelectValuesExpression::Expression(ast::Expression::Document(map! {
                            "baz".into() => ast::Expression::Literal(ast::Literal::String("hello".into()))
                        })),
                    ]),
                },
                from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(map! {
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
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: ir::Stage::Join(ir::Join {
                join_type: ir::JoinType::Inner,
                left: ir::Stage::Array(ir::Array {
                    array: vec![ir::Expression::Document(map! {
                    "foo1".into() => ir::Expression::Literal(ir::Literal::Integer(1)),
                    "bar1".into() => ir::Expression::Literal(ir::Literal::Integer(1))
                                                })],
                    alias: "bar1".into()
                })
                .into(),
                right: ir::Stage::Array(ir::Array {
                    array: vec![ir::Expression::Document(map! {
                    "foo2".into() => ir::Expression::Literal(ir::Literal::Integer(1)),
                    "bar2".into() => ir::Expression::Literal(ir::Literal::Integer(1))
                                                })],
                    alias: "bar2".into()
                })
                .into(),
                condition: None
            })
            .into(),
            expression: map! {("d", 0u16).into() =>
                ir::Expression::ScalarFunction(
                    ir::ScalarFunctionApplication {
                        function: ir::ScalarFunction::MergeObjects,
                        args: vec![ir::Expression::Reference(("bar1", 1u16).into()),
                                   ir::Expression::Reference(("bar2", 1u16).into())]
                    }
                )
            }
        })),
        Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star,]),
                },
                from_clause: Some(ast::Datasource::Join(JoinSource {
                    join_type: ast::JoinType::Inner,
                    left: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(map! {
                            "foo1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                            "bar1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        },)],
                        alias: "bar1".into(),
                    })
                    .into(),
                    right: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(map! {
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
        derived_join_datasources_overlapped_keys_fails,
        algebrize_from_clause,
        Err(Error::DerivedDatasouceOverlappingKeys(
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
        Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star,]),
                },
                from_clause: Some(ast::Datasource::Join(JoinSource {
                    join_type: ast::JoinType::Inner,
                    left: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(map! {
                            "foo1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                            "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        },)],
                        alias: "bar1".into(),
                    })
                    .into(),
                    right: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(map! {
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
}

mod limit_or_offset_clause {
    use super::{AST_SOURCE_FOO, IR_SOURCE_FOO};
    use crate::{ast, ir};

    test_algebrize!(
        limit_set,
        algebrize_limit_clause,
        Ok(ir::Stage::Limit(ir::Limit {
            source: Box::new(IR_SOURCE_FOO.clone()),
            limit: 42_u64
        })),
        Some(42_u32),
        IR_SOURCE_FOO.clone()
    );
    test_algebrize!(
        limit_unset,
        algebrize_limit_clause,
        Ok(IR_SOURCE_FOO.clone()),
        None,
        IR_SOURCE_FOO.clone()
    );
    test_algebrize!(
        offset_set,
        algebrize_offset_clause,
        Ok(ir::Stage::Offset(ir::Offset {
            source: Box::new(IR_SOURCE_FOO.clone()),
            offset: 3_u64
        })),
        Some(3_u32),
        IR_SOURCE_FOO.clone()
    );
    test_algebrize!(
        offset_unset,
        algebrize_offset_clause,
        Ok(IR_SOURCE_FOO.clone()),
        None,
        IR_SOURCE_FOO.clone()
    );
    test_algebrize!(
        limit_and_offset,
        algebrize_select_query,
        Ok(ir::Stage::Limit(ir::Limit {
            source: Box::new(ir::Stage::Offset(ir::Offset {
                source: Box::new(IR_SOURCE_FOO.clone()),
                offset: 3
            })),
            limit: 10
        })),
        ast::SelectQuery {
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
    );
}

mod set_query {
    use super::{AST_QUERY_BAR, AST_QUERY_FOO, IR_SOURCE_BAR, IR_SOURCE_FOO};
    use crate::{ast, ir};

    test_algebrize!(
        union_distinct_not_allowed,
        algebrize_set_query,
        Err(Error::DistinctUnion),
        ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::Union,
            right: Box::new(AST_QUERY_BAR.clone()),
        }
    );
    test_algebrize!(
        basic,
        algebrize_set_query,
        Ok(ir::Stage::Set(ir::Set {
            operation: ir::SetOperation::UnionAll,
            left: Box::new(IR_SOURCE_FOO.clone()),
            right: Box::new(IR_SOURCE_BAR.clone()),
        })),
        ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::UnionAll,
            right: Box::new(AST_QUERY_BAR.clone()),
        }
    );
}

mod filter_clause {
    use super::IR_SOURCE_FOO;
    use crate::{ast, ir};

    const TRUE_IR: ir::Expression = ir::Expression::Literal(ir::Literal::Boolean(true));
    const TRUE_AST: ast::Expression = ast::Expression::Literal(ast::Literal::Boolean(true));

    test_algebrize!(
        simple,
        algebrize_filter_clause,
        Ok(ir::Stage::Filter(ir::Filter {
            source: Box::new(IR_SOURCE_FOO.clone()),
            condition: TRUE_IR,
        })),
        Some(TRUE_AST),
        IR_SOURCE_FOO.clone(),
    );
    test_algebrize!(
        none,
        algebrize_filter_clause,
        Ok(IR_SOURCE_FOO.clone()),
        None,
        IR_SOURCE_FOO.clone(),
    );
}

mod order_by_clause {
    use crate::{
        ast, ir, map,
        schema::{Atomic, Document, Schema},
        set,
    };
    use lazy_static::lazy_static;

    lazy_static! {
        static ref SOURCE: ir::Stage = ir::Stage::Collection(ir::Collection {
            db: "test".into(),
            collection: "baz".into(),
        });
    }

    test_algebrize!(
        asc_and_desc,
        algebrize_order_by_clause,
        source = SOURCE.clone(),
        expected = Ok(ir::Stage::Sort(ir::Sort {
            source: Box::new(SOURCE.clone()),
            specs: vec![
                ir::SortSpecification::Asc(Box::new(ir::Expression::FieldAccess(
                    ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                        field: "a".to_string()
                    }
                ))),
                ir::SortSpecification::Desc(Box::new(ir::Expression::FieldAccess(
                    ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("foo", 0u16).into())),
                        field: "b".to_string()
                    }
                )))
            ]
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
    );

    test_algebrize!(
        sort_key_from_source,
        algebrize_order_by_clause,
        Ok(ir::Stage::Sort(ir::Sort {
            source: Box::new(ir::Stage::Array(ir::Array {
                array: vec![ir::Expression::Document(map! {
                    "a".into() => ir::Expression::Literal(ir::Literal::Integer(1))
                })],
                alias: "arr".into(),
            })),
            specs: vec![ir::SortSpecification::Asc(Box::new(
                ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                    field: "a".to_string()
                })
            )),]
        })),
        Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })),
                direction: ast::SortDirection::Asc
            },],
        }),
        ir::Stage::Array(ir::Array {
            array: vec![ir::Expression::Document(map! {
                "a".into() => ir::Expression::Literal(ir::Literal::Integer(1))
            })],
            alias: "arr".into(),
        }),
    );
}

mod group_by_clause {
    use crate::{ast, ir, map};
    use lazy_static::lazy_static;

    lazy_static! {
        // ARRAY DATASOURCE

        // [{"a" : 1}] AS arr
        static ref IR_ARRAY_SOURCE: ir::Stage = ir::Stage::Array(ir::Array {
            array: vec![ir::Expression::Document(map! {
                "a".into() => ir::Expression::Literal(ir::Literal::Integer(1))
            })],
            alias: "arr".into(),
        });

        // GROUP BY KEYS

        // arr.a AS key
        static ref IR_FIELD_ACCESS: ir::AliasedExpression = ir::AliasedExpression {
            alias: Some("key".to_string()),
            inner: ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                field: "a".to_string()
            })
        };
        static ref AST_SUBPATH: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                subpath: "a".to_string()
            }),
            alias: Some("key".to_string()),
        };

        // 1 AS literal
        static ref IR_LITERAL_KEY: ir::AliasedExpression = ir::AliasedExpression {
            alias: Some("literal".into()),
            inner: ir::Expression::Literal(ir::Literal::Integer(1)),
        };
        static ref AST_LITERAL_KEY: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Literal(ast::Literal::Integer(1)),
            alias: Some("literal".into()),
        };

        // a + 1 AS complex_expr
        static ref IR_FIELD_ACCESS_COMPLEX_EXPR: ir::AliasedExpression = ir::AliasedExpression {
            alias: Some("complex_expr".into()),
            inner: ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Add,
                args: vec![
                    ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                        field: "a".to_string()
                    }),
                    ir::Expression::Literal(ir::Literal::Integer(1))
                ]
            })
        };
        static ref AST_SUBPATH_COMPLEX_EXPR: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))
            }),
            alias: Some("complex_expr".into()),
        };

        // AGGREGATION FUNCTIONS

        // AVG(DISTINCT arr.a) AS agg1
        static ref IR_AGG_1_ARRAY: ir::AliasedAggregation = ir::AliasedAggregation {
            alias: "agg1".to_string(),
            inner: ir::AggregationExpr::Function(ir::AggregationFunctionApplication {
                function: ir::AggregationFunction::Avg,
                arg: Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(("arr", 0u16).into())),
                    field: "a".to_string()
                })),
                distinct: true,
            })
        };
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
            alias: Some("agg1".to_string()),
        };

        // COUNT(*) AS agg2
        static ref IR_AGG_2: ir::AliasedAggregation = ir::AliasedAggregation {
            alias: "agg2".to_string(),
            inner: ir::AggregationExpr::CountStar(false),
        };
        static ref AST_AGG_2: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Count,
                args: ast::FunctionArguments::Star,
                set_quantifier: None
            }),
            alias: Some("agg2".to_string()),
        };
    }

    // Successful tests.

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE AVG(DISTINCT arr.a) AS agg1, COUNT(*) AS agg2
    test_algebrize!(
        group_by_key_with_aggregation_array_source,
        algebrize_group_by_clause,
        Ok(ir::Stage::Group(ir::Group {
            source: Box::new(IR_ARRAY_SOURCE.clone()),
            keys: vec![IR_FIELD_ACCESS.clone()],
            aggregations: vec![IR_AGG_1_ARRAY.clone(), IR_AGG_2.clone()],
        })),
        Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![AST_AGG_1_ARRAY.clone(), AST_AGG_2.clone()],
        }),
        IR_ARRAY_SOURCE.clone(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY 1
    test_algebrize!(
        group_by_key_is_literal,
        algebrize_group_by_clause,
        Ok(ir::Stage::Group(ir::Group {
            source: Box::new(IR_ARRAY_SOURCE.clone()),
            keys: vec![IR_LITERAL_KEY.clone()],
            aggregations: vec![],
        })),
        Some(ast::GroupByClause {
            keys: vec![AST_LITERAL_KEY.clone()],
            aggregations: vec![],
        }),
        IR_ARRAY_SOURCE.clone(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY a + 1
    test_algebrize!(
        group_by_key_is_complex_expression,
        algebrize_group_by_clause,
        Ok(ir::Stage::Group(ir::Group {
            source: Box::new(IR_ARRAY_SOURCE.clone()),
            keys: vec![IR_FIELD_ACCESS_COMPLEX_EXPR.clone()],
            aggregations: vec![],
        })),
        Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH_COMPLEX_EXPR.clone()],
            aggregations: vec![],
        }),
        IR_ARRAY_SOURCE.clone(),
    );

    // Error tests.

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE 42 AS agg
    test_algebrize!(
        group_by_key_with_non_function_aggregation_expression,
        algebrize_group_by_clause,
        Err(Error::NonAggregationInPlaceOfAggregation(0)),
        Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![ast::AliasedExpr {
                expr: ast::Expression::Literal(ast::Literal::Integer(42)),
                alias: Some("agg".to_string()),
            },],
        }),
        IR_ARRAY_SOURCE.clone(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*)
    test_algebrize!(
        group_by_key_with_non_aliased_aggregation_function,
        algebrize_group_by_clause,
        Err(Error::AggregationFunctionHasNoAlias(0)),
        Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![ast::AliasedExpr {
                expr: ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::Count,
                    args: ast::FunctionArguments::Star,
                    set_quantifier: None
                }),
                alias: None,
            },],
        }),
        IR_ARRAY_SOURCE.clone(),
    );
}

mod subquery {
    use crate::{
        ast,
        ir::{binding_tuple::DatasourceName, *},
        map,
        schema::{Atomic, Document, Schema},
        set,
    };
    use lazy_static::lazy_static;
    lazy_static! {
        static ref AST_ARRAY: ast::Datasource = ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
            },)],
            alias: "arr".into()
        });
        static ref IR_ARRAY: Stage = Stage::Array(Array {
            array: vec![Expression::Document(map! {
                "a".into() => Expression::Literal(Literal::Integer(1))
            })],
            alias: "arr".into()
        });
    }
    test_algebrize!(
        uncorrelated_exists,
        algebrize_expression,
        Ok(Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(IR_ARRAY.clone()),
            expression: map! {
                (DatasourceName::Bottom, 1u16).into() => Expression::Document(map!{
                    "a".into() => Expression::Literal(Literal::Integer(1))
                })
            }
        })))),
        ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(map! {
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
    test_algebrize_with_env!(
        correlated_exists,
        algebrize_expression,
        Ok(Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(IR_ARRAY.clone()),
            expression: map! {
                (DatasourceName::Bottom, 2u16).into() => Expression::Document(map!{
                    "b_0".into() => Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                        field: "b".into()
                    })
                })
            }
        })))),
        ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(map! {
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
        map! {
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
        algebrize_expression,
        Ok(Expression::Exists(Box::new(Stage::Array(Array {
            array: vec![
                Expression::Document(map! {"a".into() => Expression::Literal(Literal::Integer(1))}),
                Expression::Document(map! {"a".into() => Expression::Literal(Literal::Integer(2))})
            ],
            alias: "arr".into()
        })))),
        ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(map! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    },),
                    ast::Expression::Document(map! {
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
        algebrize_expression,
        Ok(Expression::Exists(Box::new(Stage::Array(Array {
            array: vec![Expression::Document(map! {
                "a".to_string() => Expression::Literal(Literal::Integer(1)),
                "b".to_string() => Expression::Literal(Literal::Integer(2))
            })],
            alias: "arr".to_string()
        })))),
        ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(map! {
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
        algebrize_expression,
        Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(IR_ARRAY.clone()),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        })),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(map! {
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
    test_algebrize_with_env!(
        correlated_subquery_expr,
        algebrize_expression,
        Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
                field: "b_0".to_string()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(IR_ARRAY.clone()),
                expression: map! {
                    (DatasourceName::Bottom, 2u16).into() => Expression::Document(map!{
                        "b_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                            field: "b".into()
                        })
                    })
                }
            }))
        })),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(map! {
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
        map! {
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
        algebrize_expression,
        Err(Error::InvalidSubqueryDegree),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
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
        algebrize_expression,
        Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                field: "a".to_string()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(IR_ARRAY.clone()),
                expression: map! {
                    ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into())
                }
            }))
        })),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
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
        algebrize_expression,
        Err(Error::InvalidSubqueryDegree),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(map! {
                        "a_0".into() => ast::Expression::Identifier("a".into()),
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                ),])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(map! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    },),
                    ast::Expression::Document(map! {
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
        algebrize_expression,
        Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                field: "a".to_string()
            })),
            subquery: Box::new(IR_ARRAY.clone())
        })),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
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
        algebrize_expression,
        Err(Error::InvalidSubqueryDegree),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(map! {
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
        algebrize_expression,
        Err(Error::InvalidSubqueryDegree),
        ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                    ast::SubstarExpr {
                        datasource: "arr".into()
                    }
                )])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(map! {
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
}
