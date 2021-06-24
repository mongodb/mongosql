use crate::ir::{Collection, Stage};
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
            let algebrizer = Algebrizer::new("test".into(), 1u16).with_merged_mappings($env);
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
            let algebrizer = Algebrizer::new("test".into(), 1u16).with_merged_mappings($env);
            let expected: Result<_, Error> = $expected;
            let res: Result<_, Error> = algebrizer.$method($ast);
            assert_eq!(expected, res);
        }
    };
}

mod expression {
    use crate::{
        ast,
        ir::{self, binding_tuple::Key},
        map,
        schema::{Atomic, Document, Schema},
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
            db: "bar".into(),
            collection: "bar".into(),
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
        expected = Err(Error::DuplicateDatasource("select clause", Key::bot(1u16))),
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
        expected = Err(Error::DuplicateDatasource(
            "select clause",
            ("foo", 1u16).into()
        )),
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
    use crate::{
        ast, ir, map,
        schema::{Atomic, Schema, ANY_DOCUMENT},
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
            found: Schema::AnyOf(vec![Schema::Atomic(Atomic::Integer)]),
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
            found: Schema::AnyOf(vec![Schema::Atomic(Atomic::Null)]),
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
}

lazy_static! {
    static ref TEST_SOURCE: Stage = Stage::Collection(Collection {
        db: "test".into(),
        collection: "foo".into()
    });
}

mod limit_or_offset_clause {
    use crate::{algebrizer::test::TEST_SOURCE, ast, ir, map};

    test_algebrize!(
        limit_set,
        algebrize_limit_clause,
        Ok(ir::Stage::Limit(ir::Limit {
            source: Box::new(TEST_SOURCE.clone()),
            limit: 42_u64
        })),
        Some(42_u32),
        TEST_SOURCE.clone()
    );
    test_algebrize!(
        limit_unset,
        algebrize_limit_clause,
        Ok(TEST_SOURCE.clone()),
        None,
        TEST_SOURCE.clone()
    );
    test_algebrize!(
        offset_set,
        algebrize_offset_clause,
        Ok(ir::Stage::Offset(ir::Offset {
            source: Box::new(TEST_SOURCE.clone()),
            offset: 3_u64
        })),
        Some(3_u32),
        TEST_SOURCE.clone()
    );
    test_algebrize!(
        offset_unset,
        algebrize_offset_clause,
        Ok(TEST_SOURCE.clone()),
        None,
        TEST_SOURCE.clone()
    );
    test_algebrize!(
        limit_and_offset,
        algebrize_select_query,
        Ok(ir::Stage::Limit(ir::Limit {
            source: Box::new(ir::Stage::Offset(ir::Offset {
                source: Box::new(ir::Stage::Project(ir::Project {
                    source: Box::new(TEST_SOURCE.clone()),
                    expression: map! {
                        ("foo", 0u16).into() => ir::Expression::Reference(("foo", 0u16).into())
                    }
                })),
                offset: 3
            })),
            limit: 10
        })),
        ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Collection(ast::CollectionSource {
                database: Some("test".into()),
                collection: "foo".into(),
                alias: Some("foo".into()),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(10_u32),
            offset: Some(3_u32)
        },
    );
}

mod filter_clause {
    use super::TEST_SOURCE;
    use crate::{ast, ir};

    const TRUE_IR: ir::Expression = ir::Expression::Literal(ir::Literal::Boolean(true));
    const TRUE_AST: ast::Expression = ast::Expression::Literal(ast::Literal::Boolean(true));

    test_algebrize!(
        simple,
        algebrize_filter_clause,
        Ok(ir::Stage::Filter(ir::Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: TRUE_IR,
        })),
        Some(TRUE_AST),
        TEST_SOURCE.clone(),
    );
    test_algebrize!(
        none,
        algebrize_filter_clause,
        Ok(TEST_SOURCE.clone()),
        None,
        TEST_SOURCE.clone(),
    );
}
