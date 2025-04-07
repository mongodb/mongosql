use super::*;

test_algebrize!(
    ltrim,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::LTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into())),
                mir::Expression::Literal(mir::LiteralValue::String("hello world".into()))
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Trim(ast::TrimExpr {
        trim_spec: ast::TrimSpec::Leading,
        trim_chars: Box::new(ast::Expression::StringConstructor("hello".into())),
        arg: Box::new(ast::Expression::StringConstructor("hello world".into())),
    }),
);

test_algebrize!(
    rtrim,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::RTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("world".into())),
                mir::Expression::Literal(mir::LiteralValue::String("hello world".into()))
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Trim(ast::TrimExpr {
        trim_spec: ast::TrimSpec::Trailing,
        trim_chars: Box::new(ast::Expression::StringConstructor("world".into())),
        arg: Box::new(ast::Expression::StringConstructor("hello world".into())),
    }),
);

test_algebrize!(
    btrim,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(" ".into())),
                mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into()))
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Trim(ast::TrimExpr {
        trim_spec: ast::TrimSpec::Both,
        trim_chars: Box::new(ast::Expression::StringConstructor(" ".into())),
        arg: Box::new(ast::Expression::StringConstructor(" hello world ".into())),
    }),
);

test_algebrize_expr_and_schema_check!(
    trim_arg_must_be_string_or_null,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "BTrim",
        required: STRING_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Trim(ast::TrimExpr {
        trim_spec: ast::TrimSpec::Both,
        trim_chars: Box::new(ast::Expression::StringConstructor(" ".into())),
        arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize_expr_and_schema_check!(
    trim_escape_must_be_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "BTrim",
        required: STRING_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Trim(ast::TrimExpr {
        trim_spec: ast::TrimSpec::Both,
        trim_chars: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        arg: Box::new(ast::Expression::StringConstructor(" ".into())),
    }),
);
