use super::*;

test_algebrize!(
    cast_full,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Cast(mir::CastExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        to: mir::Type::String,
        on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "was_null".into(),
        ))),
        on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "was_error".into(),
        ))),
        is_nullable: false,
    })),
    input = ast::Expression::Cast(ast::CastExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        to: ast::Type::String,
        on_null: Some(Box::new(ast::Expression::StringConstructor(
            "was_null".into(),
        ))),
        on_error: Some(Box::new(ast::Expression::StringConstructor(
            "was_error".into(),
        ))),
    }),
);

test_algebrize!(
    cast_simple,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Cast(mir::CastExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        to: mir::Type::String,
        on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true,
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
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        target_type: mir::Type::Int32,
    })),
    input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        target_type: ast::Type::Int32,
    }),
);

test_algebrize_expr_and_schema_check!(
    type_assert_fail,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "::!",
        required: Schema::Atomic(Atomic::String).into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        target_type: ast::Type::String,
    }),
);

test_algebrize!(
    type_assert_ext_json_string_does_not_convert_if_target_type_is_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "{\"$numberInt\": \"42\"}".to_string()
        ))),
        target_type: mir::Type::String,
    })),
    input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
        expr: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"42\"}".to_string()
        )),
        target_type: ast::Type::String,
    }),
);

test_algebrize!(
    type_assert_ext_json_string_converts_if_target_type_is_not_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        target_type: mir::Type::Int32,
    })),
    input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
        expr: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"42\"}".to_string()
        )),
        target_type: ast::Type::Int32,
    }),
);

test_algebrize!(
    is_success,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Is(mir::IsExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
    })),
    input = ast::Expression::Is(ast::IsExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
    }),
);

test_algebrize_expr_and_schema_check!(
    is_recursive_failure,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Add",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Is(ast::IsExpr {
        expr: Box::new(ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            op: ast::BinaryOp::Add,
            right: Box::new(ast::Expression::StringConstructor("a".into())),
        })),
        target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
    }),
);

test_algebrize!(
    is_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Is(mir::IsExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
    })),
    input = ast::Expression::Is(ast::IsExpr {
        expr: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"42\"}".to_string()
        )),
        target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
    }),
);
