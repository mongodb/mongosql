mod flatten_node {
    use crate::mir::{self, schema::SchemaCache, *};
    macro_rules! test_flatten_variadic_functions {
        ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
            #[test]
            fn $func_name() {
                use crate::mir::optimizer::flatten_variadics::FlattenVariadicFunctionsOptimizer;
                let input = $input;
                let expected = $expected;
                let (actual, _) =
                    FlattenVariadicFunctionsOptimizer::flatten_variadic_functions(input);
                assert_eq!(expected, actual);
            }
        };
    }

    fn test_source() -> Stage {
        Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })
    }

    test_flatten_variadic_functions!(
        flatten_simple,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Add,
                        vec![
                            Expression::Literal(LiteralValue::Integer(1)),
                            Expression::Literal(LiteralValue::Integer(2))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_ignores_different_funcs,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Mul,
                        vec![
                            Expression::Literal(LiteralValue::Integer(1)),
                            Expression::Literal(LiteralValue::Integer(2))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Mul,
                        vec![
                            Expression::Literal(LiteralValue::Integer(1)),
                            Expression::Literal(LiteralValue::Integer(2))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_nested_multiple_levels,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(4))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
                            ScalarFunction::Add,
                            vec![
                                Expression::Literal(LiteralValue::Integer(1)),
                                Expression::Literal(LiteralValue::Integer(2)),
                                Expression::Literal(LiteralValue::Integer(4))
                            ],
                        ))],
                        is_nullable: false,
                    })
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_multiple_funcs,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::Literal(LiteralValue::Integer(4)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2)),
                            Expression::Literal(LiteralValue::Integer(1)),
                            Expression::Literal(LiteralValue::Integer(3)),
                            Expression::Literal(LiteralValue::Integer(1))
                        ],
                        is_nullable: false,
                    }),
                    Expression::Literal(LiteralValue::Integer(1))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Add,
                            args: vec![
                                Expression::Literal(LiteralValue::Integer(3)),
                                Expression::Literal(LiteralValue::Integer(4)),
                                Expression::ScalarFunction(ScalarFunctionApplication {
                                    function: ScalarFunction::Mul,
                                    args: vec![
                                        Expression::ScalarFunction(ScalarFunctionApplication::new(
                                            ScalarFunction::Mul,
                                            vec![
                                                Expression::Literal(LiteralValue::Integer(2)),
                                                Expression::Literal(LiteralValue::Integer(1))
                                            ],
                                        )),
                                        Expression::ScalarFunction(ScalarFunctionApplication::new(
                                            ScalarFunction::Mul,
                                            vec![
                                                Expression::Literal(LiteralValue::Integer(3)),
                                                Expression::Literal(LiteralValue::Integer(1))
                                            ],
                                        ))
                                    ],
                                    is_nullable: false,
                                })
                            ],
                            is_nullable: false,
                        })],
                        is_nullable: false,
                    },),
                    Expression::Literal(LiteralValue::Integer(1))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        flatten_not_necessary,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Sub,
                        vec![
                            Expression::Literal(LiteralValue::Integer(2)),
                            Expression::Literal(LiteralValue::Integer(1))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Sub,
                        vec![
                            Expression::Literal(LiteralValue::Integer(2)),
                            Expression::Literal(LiteralValue::Integer(1))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_order_matters,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("foo".to_string())),
                    Expression::Literal(LiteralValue::String("bar".to_string())),
                    Expression::Literal(LiteralValue::String("baz".to_string()))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("foo".to_string())),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Concat,
                        vec![
                            Expression::Literal(LiteralValue::String("bar".to_string())),
                            Expression::Literal(LiteralValue::String("baz".to_string()))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        add_flattened,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(1))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Add,
                        vec![
                            Expression::Literal(LiteralValue::Integer(2)),
                            Expression::Literal(LiteralValue::Integer(1))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        mul_flattened,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(1))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Mul,
                        vec![
                            Expression::Literal(LiteralValue::Integer(2)),
                            Expression::Literal(LiteralValue::Integer(1))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        and_flattened,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::And,
                        vec![
                            mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                            mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        or_flattened,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Or,
                        vec![
                            mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                            mir::Expression::Literal(mir::LiteralValue::Boolean(true))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        concat_flattened,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("a".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("b".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("c".into()))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("a".into())),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Concat,
                        vec![
                            mir::Expression::Literal(mir::LiteralValue::String("b".into())),
                            mir::Expression::Literal(mir::LiteralValue::String("c".into()))
                        ],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
}
