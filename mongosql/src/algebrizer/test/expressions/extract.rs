use super::*;

test_algebrize!(
    extract_year,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Year,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Year,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_month,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Month,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Month,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_day,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Day,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Day,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_hour,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Hour,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Hour,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_minute,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Minute,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Minute,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_second,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Second,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Second,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_millsecond,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Millisecond,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Millisecond,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_day_of_year,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::DayOfYear,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::DayOfYear,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_iso_week,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::IsoWeek,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::IsoWeek,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_day_of_week,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::DayOfWeek,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::DayOfWeek,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize!(
    extract_iso_weekday,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::IsoWeekday,
            args: vec![mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }
            ),],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::IsoWeekday,
        arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CurrentTimestamp,
            args: ast::FunctionArguments::Args(vec![]),
            set_quantifier: Some(ast::SetQuantifier::All)
        })),
    }),
);

test_algebrize_expr_and_schema_check!(
    extract_must_be_date,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Second",
        required: DATE_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Second,
        arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
    }),
);

test_algebrize!(
    extract_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::ScalarFunction(
        mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Year,
            args: vec![mir::Expression::Literal(mir::LiteralValue::DateTime(
                "2019-08-11T17:54:14.692Z"
                    .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                    .unwrap()
                    .into(),
            ))],
            is_nullable: false,
        }
    )),
    input = ast::Expression::Extract(ast::ExtractExpr {
        extract_spec: ast::DatePart::Year,
        arg: Box::new(ast::Expression::StringConstructor(
            "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
        )),
    }),
);
