mod flatten_node {
    use crate::mir;
    use crate::mir::{schema::SchemaCache, *};
    macro_rules! test_flatten_variadic_functions {
        ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
            #[test]
            fn $func_name() {
                use crate::mir::optimizer::flatten_variadics::FlattenVariadicFunctionsOptimizer;
                let input = $input;
                let expected = $expected;
                let actual = FlattenVariadicFunctionsOptimizer::flatten_variadic_functions(input);
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
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
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
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
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(4).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
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
                            is_nullable: false,
                        })],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                        is_nullable: false,
                    }),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
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
                                            is_nullable: false,
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
                                            is_nullable: false,
                                        })
                                    ],
                                    cache: SchemaCache::new(),
                                    is_nullable: false,
                                })
                            ],
                            cache: SchemaCache::new(),
                            is_nullable: false,
                        })],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    },),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
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
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::String("foo".to_string()).into()),
                    Expression::Literal(LiteralValue::String("bar".to_string()).into()),
                    Expression::Literal(LiteralValue::String("baz".to_string()).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
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
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::And,
                        args: vec![
                            mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
                            mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Or,
                        args: vec![
                            mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
                            mir::Expression::Literal(mir::LiteralValue::Boolean(true).into())
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
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
                    mir::Expression::Literal(mir::LiteralValue::String("a".into()).into()),
                    mir::Expression::Literal(mir::LiteralValue::String("b".into()).into()),
                    mir::Expression::Literal(mir::LiteralValue::String("c".into()).into())
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("a".into()).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Concat,
                        args: vec![
                            mir::Expression::Literal(mir::LiteralValue::String("b".into()).into()),
                            mir::Expression::Literal(mir::LiteralValue::String("c".into()).into())
                        ],
                        cache: SchemaCache::new(),
                        is_nullable: false,
                    })
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
}
