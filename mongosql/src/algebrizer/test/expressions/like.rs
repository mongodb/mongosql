use super::*;

test_algebrize!(
    like_success_with_pattern,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Like(mir::LikeExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "42".into(),
        ))),
        pattern: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "42".into(),
        ))),
        escape: Some('f'),
    })),
    input = ast::Expression::Like(ast::LikeExpr {
        expr: Box::new(ast::Expression::StringConstructor("42".into())),
        pattern: Box::new(ast::Expression::StringConstructor("42".into())),
        escape: Some('f'),
    }),
);

test_algebrize!(
    like_success_no_pattern,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Like(mir::LikeExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "42".into(),
        ))),
        pattern: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "42".into(),
        ))),
        escape: None,
    })),
    input = ast::Expression::Like(ast::LikeExpr {
        expr: Box::new(ast::Expression::StringConstructor("42".into())),
        pattern: Box::new(ast::Expression::StringConstructor("42".into())),
        escape: None,
    }),
);

test_algebrize_expr_and_schema_check!(
    like_expr_must_be_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Like",
        required: STRING_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Like(ast::LikeExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        pattern: Box::new(ast::Expression::StringConstructor("42".into())),
        escape: None,
    }),
);

test_algebrize_expr_and_schema_check!(
    like_pattern_must_be_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Like",
        required: STRING_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Like(ast::LikeExpr {
        expr: Box::new(ast::Expression::StringConstructor("42".into())),
        pattern: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        escape: Some(' '),
    }),
);
