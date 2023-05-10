mod flatten_node {
    use crate::mir::{schema::SchemaCache, *};

    macro_rules! test_flatten_variadic_functions {
        ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
            #[test]
            fn $func_name() {
                use crate::mir::optimizer::flatten_variadics::FlattenVariadicFunctionsOptimizer;
                let input = $input;
                let expected = $expected;
                let actual = FlattenVariadicFunctionsOptimizer::flatten_variadic_functions(input);
                assert_eq!(actual, expected);
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(4).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Add,
                            args: vec![
                                Expression::Literal(LiteralValue::Integer(1).into()),
                                Expression::Literal(LiteralValue::Integer(2).into()),
                                Expression::Literal(LiteralValue::Integer(4).into())
                            ],
                            cache: SchemaCache::new(),
                        })],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(4).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(3).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
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
                                Expression::Literal(LiteralValue::Integer(3).into()),
                                Expression::Literal(LiteralValue::Integer(4).into()),
                                Expression::ScalarFunction(ScalarFunctionApplication {
                                    function: ScalarFunction::Mul,
                                    args: vec![
                                        Expression::ScalarFunction(ScalarFunctionApplication {
                                            function: ScalarFunction::Mul,
                                            args: vec![
                                                Expression::Literal(
                                                    LiteralValue::Integer(2).into()
                                                ),
                                                Expression::Literal(
                                                    LiteralValue::Integer(1).into()
                                                )
                                            ],
                                            cache: SchemaCache::new(),
                                        }),
                                        Expression::ScalarFunction(ScalarFunctionApplication {
                                            function: ScalarFunction::Mul,
                                            args: vec![
                                                Expression::Literal(
                                                    LiteralValue::Integer(3).into()
                                                ),
                                                Expression::Literal(
                                                    LiteralValue::Integer(1).into()
                                                )
                                            ],
                                            cache: SchemaCache::new(),
                                        })
                                    ],
                                    cache: SchemaCache::new(),
                                })
                            ],
                            cache: SchemaCache::new(),
                        })],
                        cache: SchemaCache::new(),
                    },),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::String("foo".to_string()).into()),
                    Expression::Literal(LiteralValue::String("bar".to_string()).into()),
                    Expression::Literal(LiteralValue::String("baz".to_string()).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("foo".to_string()).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Concat,
                        args: vec![
                            Expression::Literal(LiteralValue::String("bar".to_string()).into()),
                            Expression::Literal(LiteralValue::String("baz".to_string()).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );
}
