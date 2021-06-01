macro_rules! test_algebrize {
    ($func_name:ident, $method:ident, $expected:expr, $ast:expr $(,)?) => {
        #[test]
        fn $func_name() {
            use crate::{
                algebrizer::{Algebrizer, Error},
                ast,
            };
            let algebrizer = Algebrizer::new("test".into(), 0u16);
            let expected: Result<_, Error> = $expected;
            let res: Result<_, Error> = algebrizer.$method($ast);
            assert_eq!(expected, res);
        }
    };
}

mod expression {
    use crate::{ir, map};
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
}

mod from_clause {
    use crate::{
        ir, map,
        schema::{Atomic, Schema, ANY_DOCUMENT},
    };

    test_algebrize!(
        collection_must_have_alias,
        algebrize_from_clause,
        Err(Error::CollectionMustHaveAlias),
        ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: None,
        }),
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
        ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: Some("bar".into()),
        }),
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
        ast::Datasource::Collection(ast::CollectionSource {
            database: Some("test2".into()),
            collection: "foo".into(),
            alias: Some("bar".into()),
        }),
    );
    test_algebrize!(
        empty_array,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![],
            alias: "bar".into(),
        }),),
        ast::Datasource::Array(ast::ArraySource {
            array: vec![],
            alias: "bar".into(),
        }),
    );
    test_algebrize!(
        dual,
        algebrize_from_clause,
        Ok(ir::Stage::Array(ir::Array {
            array: vec![ir::Expression::Document(map! {})],
            alias: "_dual".into(),
        }),),
        ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {},)],
            alias: "_dual".into(),
        }),
    );
    test_algebrize!(
        int_array,
        algebrize_from_clause,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(vec![Schema::Atomic(Atomic::Integer)]),
        })),
        ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Integer(42))],
            alias: "bar".into(),
        }),
    );
    test_algebrize!(
        null_array,
        algebrize_from_clause,
        Err(Error::SchemaChecking(ir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(vec![Schema::Atomic(Atomic::Null)]),
        })),
        ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Null)],
            alias: "bar".into(),
        })
    );
    test_algebrize!(
        array_datasource_must_be_constant,
        algebrize_from_clause,
        Err(Error::ArrayDatasourceMustBeLiteral),
        ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {
                "foo".into() => ast::Expression::Identifier("foo".into()),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "bar".into(),
        }),
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
        ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(map! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "bar".into(),
        }),
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
        ast::Datasource::Array(ast::ArraySource {
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
        }),
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
        ast::Datasource::Array(ast::ArraySource {
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
        }),
    );
}
