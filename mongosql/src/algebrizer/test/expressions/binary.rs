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

mod in_operator {
    use super::*;
    test_algebrize!(
        in_operator_is_transformed_to_in_scalar_function,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                        ],
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("a".to_string())),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::Literal(ast::Literal::Integer(2)),
                ast::Expression::Literal(ast::Literal::Integer(3)),
            ])),
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
        not_in_operator_is_transformed_to_in_scalar_function,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::NotIn,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                        ],
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("a".to_string())),
            op: ast::BinaryOp::NotIn,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::Literal(ast::Literal::Integer(2)),
                ast::Expression::Literal(ast::Literal::Integer(3)),
            ])),
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
        in_operator_translates_array_with_one_element,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![mir::Expression::Literal(mir::LiteralValue::Boolean(true)),],
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("a".to_string())),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![ast::Expression::Literal(
                ast::Literal::Boolean(true)
            )])),
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
        in_operator_converts_type_when_left_is_extended_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(3))
                        ],
                    }),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::Literal(ast::Literal::Integer(2)),
                ast::Expression::Literal(ast::Literal::Integer(3))
            ])),
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
        in_operator_enables_conversion_when_lhs_is_extended_json_and_rhs_is_tuple,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(3))
                        ],
                    }),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::Literal(ast::Literal::Integer(2)),
                ast::Expression::Literal(ast::Literal::Integer(3))
            ])),
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

    // Case: LHS is not a StringConstructor, RHS Tuple is all StringConstructors, LHS schema is Date.
    // The RHS strings should be ITC-converted to DateTime values.
    test_algebrize!(
        in_operator_converts_rhs_strings_when_lhs_is_date_field,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "d".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::DateTime(
                                "2019-08-11T17:54:14.692Z"
                                    .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                                    .unwrap()
                                    .into(),
                            )),
                            mir::Expression::Literal(mir::LiteralValue::DateTime(
                                "2020-01-01T00:00:00Z"
                                    .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                                    .unwrap()
                                    .into(),
                            )),
                        ],
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("d".to_string())),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::StringConstructor(
                    "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
                ),
                ast::Expression::StringConstructor(
                    "{\"$date\":\"2020-01-01T00:00:00Z\"}".to_string()
                ),
            ])),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "d".into() => Schema::Atomic(Atomic::Date),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    // Case: LHS is not a StringConstructor, RHS Tuple is all StringConstructors, LHS schema is String.
    // Neither side needs conversion — RHS strings should stay as strings.
    test_algebrize!(
        in_operator_no_conversion_when_lhs_is_string_field_and_rhs_is_string_constructors,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "s".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::String("hello".into())),
                            mir::Expression::Literal(mir::LiteralValue::String("world".into())),
                        ],
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("s".to_string())),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::StringConstructor("hello".to_string()),
                ast::Expression::StringConstructor("world".to_string()),
            ])),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "s".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    // Case: LHS is a StringConstructor, RHS Tuple is all StringConstructors.
    // Both sides are strings — no conversion for either.
    test_algebrize!(
        in_operator_no_conversion_when_both_sides_are_string_constructors,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("hello".into())),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::String("hello".into())),
                            mir::Expression::Literal(mir::LiteralValue::String("world".into())),
                        ],
                    }),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::StringConstructor("hello".to_string())),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::StringConstructor("hello".to_string()),
                ast::Expression::StringConstructor("world".to_string()),
            ])),
        }),
    );

    // Case: LHS is not a StringConstructor, RHS Tuple has no StringConstructors.
    // No conversion needed — both sides algebrize with ITC = false.
    test_algebrize!(
        in_operator_no_conversion_when_neither_side_has_string_constructors,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "n".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                        ],
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Identifier("n".to_string())),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::Literal(ast::Literal::Integer(2)),
                ast::Expression::Literal(ast::Literal::Integer(3)),
            ])),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "n".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );
    // Case: LHS is a Literal Integer, RHS Tuple has a mix of Literal integers and
    // StringConstructor ($numberInt) values. Only the StringConstructors are
    // ITC-converted; plain integer literals pass through unchanged.
    test_algebrize!(
        in_operator_translates_rhs_on_per_element_basis,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::In,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Array(mir::ArrayExpr {
                        array: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(4)),
                        ],
                    }),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Literal(ast::Literal::Integer(10))),
            op: ast::BinaryOp::In,
            right: Box::new(ast::Expression::Tuple(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::StringConstructor("{\"$numberInt\": \"2\"}".to_string()),
                ast::Expression::Literal(ast::Literal::Integer(3)),
                ast::Expression::StringConstructor("{\"$numberInt\": \"4\"}".to_string()),
            ])),
        }),
    );
}
