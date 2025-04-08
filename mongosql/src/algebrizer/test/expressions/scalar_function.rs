use super::*;

test_algebrize!(
    standard_scalar_function,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lower,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "hello".into(),
            )),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Lower,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "hello".into(),
        )]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    lower_scalar_function_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lower,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"1\"}".into(),
            )),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Lower,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".into(),
        )]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    replace,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Replace,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Replace,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor(" hello world ".to_string()),
            ast::Expression::StringConstructor("wo".to_string()),
            ast::Expression::StringConstructor("wowow".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    replace_null_one,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Replace,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null),
                mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Replace,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::Literal(ast::Literal::Null),
            ast::Expression::StringConstructor("wo".to_string()),
            ast::Expression::StringConstructor("wowow".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    replace_null_two,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Replace,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                mir::Expression::Literal(mir::LiteralValue::Null),
                mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Replace,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor(" hello world ".to_string()),
            ast::Expression::Literal(ast::Literal::Null),
            ast::Expression::StringConstructor("wowow".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    replace_null_three,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Replace,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                mir::Expression::Literal(mir::LiteralValue::Null),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Replace,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor(" hello world ".to_string()),
            ast::Expression::StringConstructor("wo".to_string()),
            ast::Expression::Literal(ast::Literal::Null),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    replace_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Replace,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"100\"}".into(),
                )),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"100\"}".into(),
                )),
                mir::Expression::Literal(mir::LiteralValue::Null),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Replace,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"100\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"100\"}".to_string()),
            ast::Expression::Literal(ast::Literal::Null),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize_expr_and_schema_check!(
    replace_args_must_be_string_or_null,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Replace",
        required: STRING_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Replace,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::Literal(ast::Literal::Integer(42)),
            ast::Expression::Literal(ast::Literal::Integer(42)),
            ast::Expression::Literal(ast::Literal::Integer(42)),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    log_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Log,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(100)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: false,
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
    log_bin_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Log,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(100)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Log,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"100\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    round_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Round,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: true,
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
    round_bin_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Round,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Round,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    cos_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Cos,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Cos,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            10
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    cos_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Cos,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Cos,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"10\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    sin_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sin,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Sin,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            10
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    sin_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sin,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Sin,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"10\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    tan_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Tan,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Tan,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            10
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    tan_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Tan,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Tan,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"10\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    radians_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Radians,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Radians,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            1
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    radians_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Radians,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Radians,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    sqrt_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sqrt,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(4)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Sqrt,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            4
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    sqrt_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sqrt,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(4)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Sqrt,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"4\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    abs_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Abs,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Abs,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            10
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    abs_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Abs,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Abs,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"10\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    ceil_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Ceil,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Ceil,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Double(
            1.5
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    ceil_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Ceil,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Ceil,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberDouble\": \"1.5\"}".to_string()
        )]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    degrees_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Degrees,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Degrees,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            1
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    degrees_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Degrees,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Degrees,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    floor_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Floor,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Floor,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Double(
            1.5
        )),]),
        set_quantifier: Some(ast::SetQuantifier::All),
    }),
);

test_algebrize!(
    floor_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Floor,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Floor,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberDouble\": \"1.5\"}".to_string()
        )]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    mod_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mod,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: true,
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
    mod_bin_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mod,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Mod,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string())
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    pow_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pow,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: false,
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

test_algebrize!(
    pow_bin_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pow,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                mir::Expression::Literal(mir::LiteralValue::Integer(10)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Pow,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string())
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    split_only_implicit_converts_third_arg,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Split,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Split,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    substring_implicit_converts_last_two_args,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Substring,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Substring,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    nullif_two_strings_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::NullIf,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::NullIf,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    nullif_first_arg_string_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::NullIf,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::NullIf,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ast::Expression::Literal(ast::Literal::Integer(1)),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    nullif_second_arg_string_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::NullIf,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::NullIf,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::Literal(ast::Literal::Integer(1)),
            ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    coalesce,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication::new(
            mir::ScalarFunction::Coalesce,
            vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::Integer(2)),
            ],
        )
    ),),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Coalesce,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::Literal(ast::Literal::Integer(1)),
            ast::Expression::Literal(ast::Literal::Integer(2)),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    coalesce_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication::new(
            mir::ScalarFunction::Coalesce,
            vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::Integer(2)),
            ],
        )
    ),),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Coalesce,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\":\"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\":\"2\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    size_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Size,
            args: vec![mir::Expression::Array(
                vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
            ),],
            is_nullable: false,
        }
    ),),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Size,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Array(vec![
            ast::Expression::Literal(ast::Literal::Integer(1)),
        ])]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    size_unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Size,
            args: vec![mir::Expression::Array(
                vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
            ),],
            is_nullable: false,
        }
    ),),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Size,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "[1]".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    slice_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Slice,
            args: vec![
                mir::Expression::Array(
                    vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
                ),
                mir::Expression::Literal(mir::LiteralValue::Integer(0)),
            ],

            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Slice,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::Array(vec![ast::Expression::Literal(ast::Literal::Integer(1)),]),
            ast::Expression::Literal(ast::Literal::Integer(0)),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    slice_bin_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Slice,
            args: vec![
                mir::Expression::Array(
                    vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
                ),
                mir::Expression::Literal(mir::LiteralValue::Integer(0)),
            ],

            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Slice,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("[1]".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\":\"0\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    day_of_week,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::DayOfWeek,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::DayOfWeek,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            1
        )),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    day_of_week_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::DayOfWeek,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::DayOfWeek,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\":\"1\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    bit_length_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BitLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::BitLength,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            1
        )),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    bit_length_unary_op_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BitLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\":\"1\"}".to_string()
            )),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::BitLength,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\":\"1\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    char_length_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::CharLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::CharLength,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            1
        )),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    char_length_unary_op_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::CharLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\":\"1\"}".to_string()
            )),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::CharLength,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\":\"1\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    octet_length_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::OctetLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::OctetLength,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            1
        )),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    octet_length_unary_op_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::OctetLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\":\"1\"}".to_string()
            )),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::OctetLength,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\":\"1\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    position_binary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Position,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".to_string())),
                mir::Expression::Literal(mir::LiteralValue::String("world".to_string())),
            ],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Position,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("hello".to_string()),
            ast::Expression::StringConstructor("world".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    position_binary_op_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Position,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\":\"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\":\"2\"}".to_string()
                )),
            ],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Position,
        args: ast::FunctionArguments::Args(vec![
            ast::Expression::StringConstructor("{\"$numberInt\":\"1\"}".to_string()),
            ast::Expression::StringConstructor("{\"$numberInt\":\"2\"}".to_string()),
        ]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    upper_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Upper,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "hello".to_string()
            )),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Upper,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "hello".to_string()
        ),]),
        set_quantifier: None,
    }),
);

test_algebrize!(
    upper_unary_op_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Upper,
            args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\":\"1\"}".to_string()
            )),],
            is_nullable: false
        }
    )),
    input = ast::Expression::Function(ast::FunctionExpr {
        function: ast::FunctionName::Upper,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
            "{\"$numberInt\":\"1\"}".to_string()
        ),]),
        set_quantifier: None,
    }),
);
