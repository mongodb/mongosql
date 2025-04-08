use super::*;

test_algebrize!(
    dateadd,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::DateFunction(
        mir::DateFunctionApplication {
            function: mir::DateFunction::Add,
            is_nullable: false,
            date_part: mir::DatePart::Quarter,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(5),),
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }),
            ],
        }
    )),
    input = ast::Expression::DateFunction(ast::DateFunctionExpr {
        function: ast::DateFunctionName::Add,
        date_part: ast::DatePart::Quarter,
        args: vec![
            ast::Expression::Literal(ast::Literal::Integer(5)),
            ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })
        ],
    }),
);

test_algebrize!(
    datediff,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::DateFunction(
        mir::DateFunctionApplication {
            function: mir::DateFunction::Diff,
            is_nullable: false,
            date_part: mir::DatePart::Week,
            args: vec![
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }),
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }),
                mir::Expression::Literal(mir::LiteralValue::String("sunday".to_string()),)
            ],
        }
    )),
    input = ast::Expression::DateFunction(ast::DateFunctionExpr {
        function: ast::DateFunctionName::Diff,
        date_part: ast::DatePart::Week,
        args: vec![
            ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            }),
            ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            }),
            ast::Expression::StringConstructor("sunday".to_string()),
        ],
    }),
);

test_algebrize!(
    datetrunc,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::DateFunction(
        mir::DateFunctionApplication {
            function: mir::DateFunction::Trunc,
            is_nullable: false,
            date_part: mir::DatePart::Year,
            args: vec![
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    is_nullable: false,
                }),
                mir::Expression::Literal(mir::LiteralValue::String("sunday".to_string()),)
            ],
        }
    )),
    input = ast::Expression::DateFunction(ast::DateFunctionExpr {
        function: ast::DateFunctionName::Trunc,
        date_part: ast::DatePart::Year,
        args: vec![
            ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            }),
            ast::Expression::StringConstructor("sunday".to_string()),
        ],
    }),
);

test_algebrize!(
    date_function_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::DateFunction(
        mir::DateFunctionApplication {
            function: mir::DateFunction::Add,
            is_nullable: false,
            date_part: mir::DatePart::Quarter,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(5),),
                mir::Expression::Literal(mir::LiteralValue::DateTime(
                    "2019-08-11T17:54:14.692Z"
                        .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                        .unwrap()
                        .into(),
                ))
            ],
        }
    )),
    input = ast::Expression::DateFunction(ast::DateFunctionExpr {
        function: ast::DateFunctionName::Add,
        date_part: ast::DatePart::Quarter,
        args: vec![
            ast::Expression::StringConstructor("{\"$numberInt\": \"5\"}".to_string()),
            ast::Expression::StringConstructor(
                "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
            )
        ],
    }),
);
