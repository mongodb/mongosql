use super::*;

test_algebrize!(
    add_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Add,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        op: ast::BinaryOp::Add,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize_expr_and_schema_check!(
    add_wrong_types,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Add",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor("hello".into())),
        op: ast::BinaryOp::Add,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    sub_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sub,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        op: ast::BinaryOp::Sub,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize_expr_and_schema_check!(
    sub_wrong_types,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Sub",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor("hello".into())),
        op: ast::BinaryOp::Sub,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    div_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Div,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Double(42.5)),
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Double(42.5))),
        op: ast::BinaryOp::Div,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    cast_div_result_of_two_integers_to_integer,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Cast(mir::CastExpr {
        expr: Box::new(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Div,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42))
                ],
                is_nullable: true
            }
        )),
        to: mir::Type::Int32,
        on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true
    })),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        op: ast::BinaryOp::Div,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    cast_div_result_of_long_and_integer_to_long,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Cast(mir::CastExpr {
        expr: Box::new(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Div,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Long(42)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(42))
                ],
                is_nullable: true
            }
        )),
        to: mir::Type::Int64,
        on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true
    })),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Long(42))),
        op: ast::BinaryOp::Div,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);
test_algebrize!(
    cast_implicit_converts_expr_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Cast(mir::CastExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
        to: mir::Type::String,
        on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "{\"$numberInt\": \"1\"}".to_string()
        ))),
        on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "{\"$numberInt\": \"2\"}".to_string()
        ))),
        is_nullable: false,
    })),
    input = ast::Expression::Cast(ast::CastExpr {
        expr: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"42\"}".to_string()
        )),
        to: ast::Type::String,
        on_null: Some(Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        ))),
        on_error: Some(Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"2\"}".to_string()
        ))),
    }),
);

test_algebrize_expr_and_schema_check!(
    div_wrong_types,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Div",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor("hello".into())),
        op: ast::BinaryOp::Div,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    mul_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mul,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        op: ast::BinaryOp::Mul,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize_expr_and_schema_check!(
    mul_wrong_types,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Mul",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor("hello".into())),
        op: ast::BinaryOp::Mul,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    concat_bin_op,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Concat,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("42".into())),
                mir::Expression::Literal(mir::LiteralValue::String("42".into())),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor("42".into())),
        op: ast::BinaryOp::Concat,
        right: Box::new(ast::Expression::StringConstructor("42".into())),
    }),
);

test_algebrize_expr_and_schema_check!(
    concat_wrong_types,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Concat",
        required: STRING_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor("hello".into())),
        op: ast::BinaryOp::Concat,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    eq_bool_and_int,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Eq,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
    }),
);

test_algebrize!(
    gt_bool_and_int,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(0))),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Gt),
        right: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
    }),
);

test_algebrize!(
    and_bool_and_int,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::And,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(0))),
        op: ast::BinaryOp::And,
        right: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
    }),
);

test_algebrize!(
    or_int_and_int,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Or,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Literal(ast::Literal::Integer(0))),
        op: ast::BinaryOp::Or,
        right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
    }),
);

test_algebrize!(
    add_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Add,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
        op: ast::BinaryOp::Add,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
);

test_algebrize!(
    and_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::And,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
        op: ast::BinaryOp::And,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
);

test_algebrize!(
    div_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Div,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Double(4.0)),
                mir::Expression::Literal(mir::LiteralValue::Double(2.0)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberDouble\": \"4.0\"}".to_string()
        )),
        op: ast::BinaryOp::Div,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberDouble\": \"2.0\"}".to_string()
        )),
    }),
);

test_algebrize!(
    mul_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mul,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                mir::Expression::Literal(mir::LiteralValue::Integer(3)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"2\"}".to_string()
        )),
        op: ast::BinaryOp::Mul,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"3\"}".to_string()
        )),
    }),
);

test_algebrize!(
    or_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Or,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"0\"}".to_string()
        )),
        op: ast::BinaryOp::Or,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
);

test_algebrize!(
    sub_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sub,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(7)),
                mir::Expression::Literal(mir::LiteralValue::Integer(6)),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"7\"}".to_string()
        )),
        op: ast::BinaryOp::Sub,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"6\"}".to_string()
        )),
    }),
);

test_algebrize!(
    concat_does_not_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Concat,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
        op: ast::BinaryOp::Concat,
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
);

test_algebrize!(
    comp_with_two_strings_does_not_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Eq,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            ],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
);

test_algebrize!(
    comp_with_left_that_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                    field: "a".into(),
                    is_nullable: true,
                }),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Gt),
        right: Box::new(ast::Expression::Identifier("a".to_string())),
    }),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    comp_with_left_that_does_not_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                    field: "a".into(),
                    is_nullable: true,
                }),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Lt),
        right: Box::new(ast::Expression::Identifier("a".to_string())),
    }),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::String),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    comp_with_right_that_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gte,
            args: vec![
                mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                    field: "a".into(),
                    is_nullable: true,
                }),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Identifier("a".to_string())),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Gte),
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    comp_with_right_that_does_not_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lte,
            args: vec![
                mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                    field: "a".into(),
                    is_nullable: true,
                }),
                mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            ],
            is_nullable: true,
        }
    )),
    input = ast::Expression::Binary(ast::BinaryExpr {
        left: Box::new(ast::Expression::Identifier("a".to_string())),
        op: ast::BinaryOp::Comparison(ast::ComparisonOp::Lte),
        right: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        )),
    }),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::String),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);
