use super::*;

test_algebrize!(
    neg_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Neg,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Unary(ast::UnaryExpr {
        op: ast::UnaryOp::Neg,
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize_expr_and_schema_check!(
    neg_wrong_type,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Neg",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Boolean).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Unary(ast::UnaryExpr {
        op: ast::UnaryOp::Neg,
        expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
    }),
);

test_algebrize!(
    pos_unary_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pos,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Unary(ast::UnaryExpr {
        op: ast::UnaryOp::Pos,
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize_expr_and_schema_check!(
    pos_wrong_type,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Pos",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Boolean).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Unary(ast::UnaryExpr {
        op: ast::UnaryOp::Pos,
        expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
    }),
);

test_algebrize!(
    unary_op_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Neg,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Unary(ast::UnaryExpr {
        op: ast::UnaryOp::Neg,
        expr: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"42\"}".to_string()
        )),
    }),
);
