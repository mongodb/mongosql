use super::*;

mod map {
    use super::*;

    // MAP([1], 1)
    test_algebrize!(
        no_variables,
        method = algebrize_expression,
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                is_nullable: false,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            })),
    );

    // MAP(foo.a, 1)
    test_algebrize!(
        nullable,
        method = algebrize_expression,
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                        key: ("foo", 0u16).into(),
                    })),
                    field: "a".to_string(),
                    is_nullable: true,
                })),
                f: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                is_nullable: true,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            })),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    // MAP([1], this)
    test_algebrize!(
        with_this_variable,
        method = algebrize_expression,
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: false,
                })),
                is_nullable: false,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            })),
    );

    // MAP([1, NULL], this)
    test_algebrize!(
        with_nullable_this_variable,
        method = algebrize_expression,
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        mir::Expression::Literal(mir::LiteralValue::Null),
                    ]
                })),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: true,
                })),
                is_nullable: false,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![
                    ast::Expression::Literal(ast::Literal::Integer(1)),
                    ast::Expression::Literal(ast::Literal::Null),
                ])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            })),
    );

    // MAP([1], this + this)
    test_algebrize!(
        with_this_variable_multiple_times,
        method = algebrize_expression,
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Add,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }
                )),
                is_nullable: false,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Identifier("this".into())),
                    }
                ))),
            })),
    );

    // MAP([1], this + foo.this)
    test_algebrize!(
        with_this_variable_and_field_access,
        method = algebrize_expression,
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Add,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                    key: ("foo", 0u16).into(),
                                })),
                                field: "this".to_string(),
                                is_nullable: false,
                            }),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }
                )),
                is_nullable: false,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("foo".into())),
                            subpath: "this".to_string(),
                        })),
                    }
                ))),
            })),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "this".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! { "this".to_string() },
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        invalid_array_arg,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Map",
            cause: HigherOrderFunctionErrorCause::ArrayArg,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            })),
    );

    test_algebrize!(
        invalid_function_arg,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Map",
            cause: HigherOrderFunctionErrorCause::FunctionArg,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("foo".into())),
                            subpath: "this".to_string(),
                        })),
                    }
                ))),
            })),
    );
}

mod filter {
    use super::*;

    // FILTER([1], true)
    test_algebrize!(
        no_variables,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Boolean(true)
                ))),
            }
        )),
    );

    // FILTER(foo.a, 1)
    test_algebrize!(
        nullable,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                        key: ("foo", 0u16).into(),
                    })),
                    field: "a".to_string(),
                    is_nullable: true,
                })),
                f: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                is_nullable: true,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Boolean(true)
                ))),
            }
        )),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    // FILTER([true], this)
    test_algebrize!(
        with_this_variable,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Boolean(true))]
                })),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: false,
                })),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Boolean(true)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            }
        )),
    );

    // FILTER([true, NULL], this)
    test_algebrize!(
        with_nullable_this_variable,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                        mir::Expression::Literal(mir::LiteralValue::Null),
                    ]
                })),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: true,
                })),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![
                    ast::Expression::Literal(ast::Literal::Boolean(true)),
                    ast::Expression::Literal(ast::Literal::Null),
                ])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            }
        )),
    );

    // FILTER([1], this = this)
    test_algebrize!(
        with_this_variable_multiple_times,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Eq,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }
                )),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
                        right: Box::new(ast::Expression::Identifier("this".into())),
                    }
                ))),
            }
        )),
    );

    // FILTER([1], this = foo.this)
    test_algebrize!(
        with_this_variable_and_field_access,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Eq,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                    key: ("foo", 0u16).into(),
                                })),
                                field: "this".to_string(),
                                is_nullable: false,
                            }),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }
                )),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
                        right: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("foo".into())),
                            subpath: "this".to_string(),
                        })),
                    }
                ))),
            }
        )),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "this".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! { "this".to_string() },
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        invalid_array_arg,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Filter",
            cause: HigherOrderFunctionErrorCause::ArrayArg,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Boolean(true)
                ))),
            }
        )),
    );

    test_algebrize!(
        invalid_function_arg,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Filter",
            cause: HigherOrderFunctionErrorCause::FunctionArg,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
                        right: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("foo".into())),
                            subpath: "this".to_string(),
                        })),
                    }
                ))),
            }
        )),
    );
}

mod reduce {
    use super::*;

    // REDUCE([1], 1, 1)
    test_algebrize!(
        no_variables,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            }
        )),
    );

    // REDUCE(foo.a, 1, 1)
    test_algebrize!(
        nullable_because_of_array,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                        key: ("foo", 0u16).into(),
                    })),
                    field: "a".to_string(),
                    is_nullable: true,
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                is_nullable: true,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            }
        )),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    // REDUCE([1], 1, this)
    test_algebrize!(
        with_this_variable,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: false,
                })),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            }
        )),
    );

    // REDUCE([1, NULL], 1, this)
    test_algebrize!(
        with_nullable_this_variable,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        mir::Expression::Literal(mir::LiteralValue::Null),
                    ]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: true,
                })),
                is_nullable: true,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![
                    ast::Expression::Literal(ast::Literal::Integer(1)),
                    ast::Expression::Literal(ast::Literal::Null),
                ])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            }
        )),
    );

    // REDUCE([NULL], 1, value)
    test_algebrize!(
        with_value_variable,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Null)]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "value".to_string(),
                    is_nullable: false,
                })),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Null
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "value".into()
                ))),
            }
        )),
    );

    // REDUCE([1], null, value)
    test_algebrize!(
        with_nullable_value_variable_because_of_init_value,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "value".to_string(),
                    is_nullable: true,
                })),
                is_nullable: true,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ),])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Null)),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "value".into()
                ))),
            }
        )),
    );

    // REDUCE([1], 1, value + null)
    test_algebrize!(
        with_nullable_value_variable_because_of_function_argument,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Add,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "value".to_string(),
                                is_nullable: true,
                            }),
                            mir::Expression::Literal(mir::LiteralValue::Null),
                        ],
                        force_mql_semantics: false,
                        is_nullable: true,
                    }
                )),
                is_nullable: true,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ),])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("value".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Literal(ast::Literal::Null)),
                    }
                ))),
            }
        )),
    );

    // REDUCE([1], field, this + value + this + value)
    test_algebrize!(
        with_variables_multiple_times,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                init_value: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                        key: ("foo", 0u16).into(),
                    })),
                    field: "field".to_string(),
                    is_nullable: true,
                })),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Add,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                function: mir::ScalarFunction::Add,
                                args: vec![
                                    mir::Expression::Variable(mir::Variable {
                                        name: "value".to_string(),
                                        is_nullable: true,
                                    }),
                                    mir::Expression::ScalarFunction(
                                        mir::ScalarFunctionApplication {
                                            function: mir::ScalarFunction::Add,
                                            args: vec![
                                                mir::Expression::Variable(mir::Variable {
                                                    name: "this".to_string(),
                                                    is_nullable: false,
                                                }),
                                                mir::Expression::Variable(mir::Variable {
                                                    name: "value".to_string(),
                                                    is_nullable: true,
                                                }),
                                            ],
                                            force_mql_semantics: false,
                                            is_nullable: true,
                                        }
                                    )
                                ],
                                force_mql_semantics: false,
                                is_nullable: true,
                            })
                        ],
                        force_mql_semantics: false,
                        is_nullable: true,
                    }
                )),
                is_nullable: true,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "field".to_string(),
                })),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                            left: Box::new(ast::Expression::Identifier("value".into())),
                            op: ast::BinaryOp::Add,
                            right: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                                left: Box::new(ast::Expression::Identifier("this".into())),
                                op: ast::BinaryOp::Add,
                                right: Box::new(ast::Expression::Identifier("value".into())),
                            })),
                        })),
                    }
                ))),
            }
        )),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "field".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! { },
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    // REDUCE([1], 1, this + foo.this + value + this.value)
    test_algebrize!(
        with_variable_and_field_accesses_of_same_names,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Add,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                function: mir::ScalarFunction::Add,
                                args: vec![
                                    mir::Expression::FieldAccess(mir::FieldAccess {
                                        expr: Box::new(mir::Expression::Reference(
                                            mir::ReferenceExpr {
                                                key: ("foo", 0u16).into(),
                                            }
                                        )),
                                        field: "this".to_string(),
                                        is_nullable: false,
                                    }),
                                    mir::Expression::ScalarFunction(
                                        mir::ScalarFunctionApplication {
                                            function: mir::ScalarFunction::Add,
                                            args: vec![
                                                mir::Expression::Variable(mir::Variable {
                                                    name: "value".to_string(),
                                                    is_nullable: false,
                                                }),
                                                mir::Expression::FieldAccess(mir::FieldAccess {
                                                    expr: Box::new(mir::Expression::Reference(
                                                        mir::ReferenceExpr {
                                                            key: ("foo", 0u16).into(),
                                                        }
                                                    )),
                                                    field: "value".to_string(),
                                                    is_nullable: false,
                                                }),
                                            ],
                                            force_mql_semantics: false,
                                            is_nullable: false,
                                        }
                                    )
                                ],
                                force_mql_semantics: false,
                                is_nullable: false,
                            })
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }
                )),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                            left: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                                expr: Box::new(ast::Expression::Identifier("foo".into())),
                                subpath: "this".to_string(),
                            })),
                            op: ast::BinaryOp::Add,
                            right: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                                left: Box::new(ast::Expression::Identifier("value".into())),
                                op: ast::BinaryOp::Add,
                                right: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                                    subpath: "value".to_string(),
                                })),
                            })),
                        })),
                    }
                ))),
            }
        )),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "this".into() => Schema::Atomic(Atomic::Integer),
                    "value".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! { "this".to_string(), "value".to_string() },
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        invalid_array_arg,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::ArrayArg,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            }
        )),
    );

    test_algebrize!(
        invalid_init_value,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::InitialValue,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".into())),
                    subpath: "a".to_string(),
                })),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ))),
            }
        )),
    );

    test_algebrize!(
        invalid_function_arg,
        method = algebrize_expression,
        expected = Err(Error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::FunctionArg,
            error: Box::new(Error::FieldNotFound(
                "foo".to_string(),
                None,
                ClauseType::Unintialized,
                0u16
            )),
        }),
        expected_error_code = 3035,
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("foo".into())),
                            subpath: "this".to_string(),
                        })),
                    }
                ))),
            }
        )),
    );
}

mod shadowing_variables {
    use super::*;

    test_algebrize!(
        map_shadows_outer_this,
        method = algebrize_expression,
        expression_context = ExpressionContext::default().with_variables(&mut map! {
            // Note that the outer "this" is not nullable, but the expectation is that it is
            // shadowed by the Map's `this` variable which is nullable.
            "this" => Schema::Atomic(Atomic::Integer),
        }),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Null)]
                })),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: true,
                })),
                is_nullable: false,
            })
        )),
        input =
            ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Null
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            })),
    );

    test_algebrize!(
        filter_shadows_outer_this,
        method = algebrize_expression,
        expression_context = ExpressionContext::default().with_variables(&mut map! {
            // Note that the outer "this" is not a boolean, which is invalid in the context of the
            // Filter's function argument. The expectation is that it is  shadowed by the Filter's
            // `this` variable which is a boolean.
            "this" => Schema::Atomic(Atomic::Integer),
        }),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Boolean(true))]
                })),
                f: Box::new(mir::Expression::Variable(mir::Variable {
                    name: "this".to_string(),
                    is_nullable: false,
                })),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Filter(
            ast::FilterExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Boolean(true)
                )])),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Identifier(
                    "this".into()
                ))),
            }
        )),
    );

    test_algebrize!(
        reduce_shadows_outer_this_and_value,
        method = algebrize_expression,
        expression_context = ExpressionContext::default().with_variables(&mut map! {
            // Note that the outer "this" and "value" are not numeric, but the expectation is that
            // they are shadowed by the Reduce's `this` and `value` variables which are numeric.
            "this" => Schema::Atomic(Atomic::String),
            "value" => Schema::Atomic(Atomic::String),
        }),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Add,
                        args: vec![
                            mir::Expression::Variable(mir::Variable {
                                name: "this".to_string(),
                                is_nullable: false,
                            }),
                            mir::Expression::Variable(mir::Variable {
                                name: "value".to_string(),
                                is_nullable: false,
                            }),
                        ],
                        force_mql_semantics: false,
                        is_nullable: false,
                    }
                )),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                    ast::BinaryExpr {
                        left: Box::new(ast::Expression::Identifier("this".into())),
                        op: ast::BinaryOp::Add,
                        right: Box::new(ast::Expression::Identifier("value".into())),
                    }
                ))),
            }
        )),
    );

    // This test demonstrates how shadowing works for variables in nested higher order functions. At the
    // time of writing, MongoSQL does not support user-provided variable names. Instead, it requires use
    // of the variables `this` and `value` in higher order functions. This test demonstrates that in the
    // context of a nested higher order function, the variables `this` and `value` refer to the most
    // local higher order function that defines them.
    //
    // This is a very large test, but it is useful to cover all the shadowing rules in one test that
    // represents a real query. The previous tests are smaller, more targeted tests that demonstrate
    // shadowing per higher order function type.
    //
    // REDUCE(
    //   [1],
    //   1,
    //   SIZE(
    //     MAP(
    //       ['1'],
    //       CAST(
    //         this || '0' AS INT,  // `this` refers to the STRING elements of the MAP array arg since
    //                              // MAP defines (overwrites) the `this` from REDUCE
    //         0 ON NULL,
    //         0 ON ERROR
    //       )
    //       + value // `value` refers to the INTEGER value of the REDUCE accumulated result since MAP
    //               // does not define (overwrite) `value`
    //       + SIZE(
    //           FILTER(
    //             [false],
    //             this // `this` refers to the BOOLEAN elements of the FILTER array arg since FILTER
    //                  // defines (overwrites) `this` from MAP
    //             OR value::BOOL // `value` refers to the INTEGER value of the REDUCE accumulated
    //                            // result since FILTER does not define (overwrite) `value`
    //           )
    //         )
    //     )
    //   )
    //   + this  // `this` refers to the INTEGER elements of the REDUCE array arg
    //   + value // `value` refers to the INTEGER value of the REDUCE accumulated result
    //   + REDUCE(
    //       ["1"],
    //       "1",
    //       CAST(
    //         this::INT + value::INT // `this` and `value` refer to the STRING elements of the
    //         AS STRING              // nested REDUCE
    //       )
    //     )::INT
    // )
    test_algebrize!(
        shadowing_variables,
        method = algebrize_expression,
        expression_context = ExpressionContext::default(),
        expected = Ok(mir::Expression::HigherOrderFunction(
            mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                    array: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))]
                })),
                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                f: Box::new(mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Add,
                    args: vec![
                        mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Add,
                            args: vec![
                                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                    function: mir::ScalarFunction::Add,
                                    args: vec![
                                        mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                            function: mir::ScalarFunction::Size,
                                            args: vec![
                                                mir::Expression::HigherOrderFunction(mir::HigherOrderFunctionApplication::Map(mir::MapExpr {
                                                    array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                                                        array: vec![mir::Expression::Literal(mir::LiteralValue::String("1".to_string()))]
                                                    })),
                                                    f: Box::new(mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                                        function: mir::ScalarFunction::Add,
                                                        args: vec![
                                                            mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                                                function: mir::ScalarFunction::Add,
                                                                args: vec![
                                                                    mir::Expression::Cast(mir::CastExpr {
                                                                        expr: Box::new(mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                                                            function: mir::ScalarFunction::Concat,
                                                                            args: vec![
                                                                                mir::Expression::Variable(mir::Variable {
                                                                                    name: "this".to_string(),
                                                                                    is_nullable: false,
                                                                                }),
                                                                                mir::Expression::Literal(mir::LiteralValue::String("0".to_string())),
                                                                            ],
                                                                            force_mql_semantics: false,
                                                                            is_nullable: false,
                                                                        })),
                                                                        to: mir::Type::Int32,
                                                                        on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                                                                        on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                                                                        is_nullable: false,
                                                                    }),
                                                                    mir::Expression::Variable(mir::Variable {
                                                                        name: "value".to_string(),
                                                                        is_nullable: false,
                                                                    }),
                                                                ],
                                                                force_mql_semantics: false,
                                                                is_nullable: false,
                                                            }),
                                                            mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                                                function: mir::ScalarFunction::Size,
                                                                args: vec![
                                                                    mir::Expression::HigherOrderFunction(mir::HigherOrderFunctionApplication::Filter(mir::FilterExpr {
                                                                        array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                                                                            array: vec![mir::Expression::Literal(mir::LiteralValue::Boolean(false))]
                                                                        })),
                                                                        f: Box::new(mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                                                            function: mir::ScalarFunction::Or,
                                                                            args: vec![
                                                                                mir::Expression::Variable(mir::Variable {
                                                                                    name: "this".to_string(),
                                                                                    is_nullable: false,
                                                                                }),
                                                                                mir::Expression::Cast(mir::CastExpr {
                                                                                    expr: Box::new(mir::Expression::Variable(mir::Variable {
                                                                                        name: "value".to_string(),
                                                                                        is_nullable: false,
                                                                                    })),
                                                                                    to: mir::Type::Boolean,
                                                                                    on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(false))),
                                                                                    on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(false))),
                                                                                    is_nullable: false,
                                                                                }),
                                                                            ],
                                                                            force_mql_semantics: false,
                                                                            is_nullable: false,
                                                                        })),
                                                                        is_nullable: false,
                                                                    })),
                                                                ],
                                                                force_mql_semantics: false,
                                                                is_nullable: false,
                                                            }),
                                                        ],
                                                        force_mql_semantics: false,
                                                        is_nullable: false,
                                                    })),
                                                    is_nullable: false,
                                                }))
                                            ],
                                            force_mql_semantics: false,
                                            is_nullable: false,
                                        }),
                                        mir::Expression::Variable(mir::Variable {
                                            name: "this".to_string(),
                                            is_nullable: false,
                                        })
                                    ],
                                    force_mql_semantics: false,
                                    is_nullable: false,
                                }),
                                mir::Expression::Variable(mir::Variable {
                                    name: "value".to_string(),
                                    is_nullable: false,
                                }),
                            ],
                            force_mql_semantics: false,
                            is_nullable: false,
                        }),
                        mir::Expression::Cast(mir::CastExpr {
                            expr: Box::new(mir::Expression::HigherOrderFunction(mir::HigherOrderFunctionApplication::Reduce(mir::ReduceExpr {
                                array: Box::new(mir::Expression::Array(mir::ArrayExpr {
                                    array: vec![mir::Expression::Literal(mir::LiteralValue::String("1".to_string()))]
                                })),
                                init_value: Box::new(mir::Expression::Literal(mir::LiteralValue::String("1".to_string()))),
                                f: Box::new(mir::Expression::Cast(mir::CastExpr {
                                    expr: Box::new(mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                        function: mir::ScalarFunction::Add,
                                        args: vec![
                                            mir::Expression::Cast(mir::CastExpr {
                                                expr: Box::new(mir::Expression::Variable(mir::Variable {
                                                    name: "this".to_string(),
                                                    is_nullable: false,
                                                })),
                                                to: mir::Type::Int32,
                                                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                                                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                                                is_nullable: false,
                                            }),
                                            mir::Expression::Cast(mir::CastExpr {
                                                expr: Box::new(mir::Expression::Variable(mir::Variable {
                                                    name: "value".to_string(),
                                                    is_nullable: false,
                                                })),
                                                to: mir::Type::Int32,
                                                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                                                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                                                is_nullable: false,
                                            }),
                                        ],
                                        force_mql_semantics: false,
                                        is_nullable: false,
                                    })),
                                    to: mir::Type::String,
                                    on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::String("0".to_string()))),
                                    on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::String("0".to_string()))),
                                    is_nullable: false,
                                })),
                                is_nullable: false,
                            }))),
                            to: mir::Type::Int32,
                            on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                            on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(0))),
                            is_nullable: false,
                        }),
                    ],
                    force_mql_semantics: false,
                    is_nullable: false,
                })),
                is_nullable: false,
            })
        )),
        input = ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
            ast::ReduceExpr {
                array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1),
                )])),
                init_value: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(ast::BinaryExpr {
                    left: Box::new(ast::Expression::Binary(
                        ast::BinaryExpr {
                            left: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                                left: Box::new(ast::Expression::Function(ast::FunctionExpr {
                                    function: ast::FunctionName::Size,
                                    args: ast::FunctionArguments::Args(vec![
                                        ast::Expression::HigherOrderFunction(
                                            ast::HigherOrderFunctionExpr::Map(ast::MapExpr {
                                                array: Box::new(ast::Expression::Array(vec![
                                                    ast::Expression::StringConstructor("1".to_string()),
                                                ])),
                                                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(ast::BinaryExpr {
                                                    left: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                                                        left: Box::new(ast::Expression::Cast(ast::CastExpr {
                                                            expr: Box::new(ast::Expression::Binary(
                                                                ast::BinaryExpr {
                                                                    left: Box::new(ast::Expression::Identifier(
                                                                        "this".into(),
                                                                    )),
                                                                    op: ast::BinaryOp::Concat,
                                                                    right: Box::new(
                                                                        ast::Expression::StringConstructor(
                                                                            "0".to_string()
                                                                        )
                                                                    ),
                                                                }
                                                            )),
                                                            to: ast::Type::Int32,
                                                            on_null: Some(ast::Expression::Literal(
                                                                ast::Literal::Integer(0)
                                                            ).into()),
                                                            on_error: Some(ast::Expression::Literal(
                                                                ast::Literal::Integer(0)
                                                            ).into()),
                                                        })),
                                                        op: ast::BinaryOp::Add,
                                                        right: Box::new(ast::Expression::Identifier(
                                                            "value".into()
                                                        )),
                                                    })),
                                                    op: ast::BinaryOp::Add,
                                                    right: Box::new(ast::Expression::Function(ast::FunctionExpr {
                                                        function: ast::FunctionName::Size,
                                                        args: ast::FunctionArguments::Args(vec![
                                                            ast::Expression::HigherOrderFunction(
                                                                ast::HigherOrderFunctionExpr::Filter(
                                                                    ast::FilterExpr {
                                                                        array: Box::new(ast::Expression::Array(vec![ast::Expression::Literal(
                                                                            ast::Literal::Boolean(false)
                                                                        )])),
                                                                        f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Binary(
                                                                            ast::BinaryExpr {
                                                                                left: Box::new(ast::Expression::Identifier(
                                                                                    "this".into()
                                                                                )),
                                                                                op: ast::BinaryOp::Or,
                                                                                right: Box::new(ast::Expression::Cast(ast::CastExpr {
                                                                                    expr: Box::new(ast::Expression::Identifier(
                                                                                        "value".into()
                                                                                    )),
                                                                                    to: ast::Type::Boolean,
                                                                                    on_null: Some(ast::Expression::Literal(
                                                                                        ast::Literal::Boolean(false)
                                                                                    ).into()),
                                                                                    on_error: Some(ast::Expression::Literal(
                                                                                        ast::Literal::Boolean(false)
                                                                                    ).into()),
                                                                                })),
                                                                            }
                                                                        ))),
                                                                    }))
                                                        ]),
                                                        set_quantifier: None,
                                                    })),
                                                }))),
                                            }))
                                    ]),
                                    set_quantifier: None,
                                })),
                                op: ast::BinaryOp::Add,
                                right: Box::new(ast::Expression::Identifier("this".into())),
                            })),
                            op: ast::BinaryOp::Add,
                            right: Box::new(ast::Expression::Identifier("value".into()))
                        }
                    )),
                    op: ast::BinaryOp::Add,
                    right: Box::new(ast::Expression::Cast(ast::CastExpr {
                        expr: Box::new(ast::Expression::HigherOrderFunction(ast::HigherOrderFunctionExpr::Reduce(
                            ast::ReduceExpr {
                                array: Box::new(ast::Expression::Array(vec![ast::Expression::StringConstructor(
                                    "1".to_string(),
                                )])),
                                init_value: Box::new(ast::Expression::StringConstructor("1".to_string())),
                                f: Box::new(ast::FunctionArgument::Expr(ast::Expression::Cast(ast::CastExpr {
                                    expr: Box::new(ast::Expression::Binary(
                                        ast::BinaryExpr {
                                            left: Box::new(ast::Expression::Cast(ast::CastExpr {
                                                expr: Box::new(ast::Expression::Identifier("this".into())),
                                                to: ast::Type::Int32,
                                                on_null: Some(ast::Expression::Literal(
                                                    ast::Literal::Integer(0)
                                                ).into()),
                                                on_error: Some(ast::Expression::Literal(
                                                    ast::Literal::Integer(0)
                                                ).into()),
                                            })),
                                            op: ast::BinaryOp::Add,
                                            right: Box::new(ast::Expression::Cast(ast::CastExpr {
                                                expr: Box::new(ast::Expression::Identifier(
                                                    "value".into()
                                                )),
                                                to: ast::Type::Int32,
                                                on_null: Some(ast::Expression::Literal(
                                                    ast::Literal::Integer(0)
                                                ).into()),
                                                on_error: Some(ast::Expression::Literal(
                                                    ast::Literal::Integer(0)
                                                ).into()),
                                            })),
                                        }
                                    )),
                                    to: ast::Type::String,
                                    on_null: Some(ast::Expression::StringConstructor("0".to_string()).into()),
                                    on_error: Some(ast::Expression::StringConstructor("0".to_string()).into()),
                                }))),
                            }
                        ))),
                        to: ast::Type::Int32,
                        on_null: Some(ast::Expression::Literal(ast::Literal::Integer(0)).into()),
                        on_error: Some(ast::Expression::Literal(ast::Literal::Integer(0)).into()),
                    })),
                }))),
            }
        )),
    );
}
