use super::*;

test_algebrize!(
    searched_case,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "bar".into(),
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "foo".into(),
        ))),
        is_nullable: false,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: None,
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
    }),
);

test_algebrize!(
    searched_case_no_else,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "bar".into(),
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: None,
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: None,
    }),
);

test_algebrize_expr_and_schema_check!(
    searched_case_when_condition_is_not_bool,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "SearchedCase",
        required: BOOLEAN_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::Expression::Case(ast::CaseExpr {
        expr: None,
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::StringConstructor("foo".into())),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
    }),
);

test_algebrize!(
    searched_case_nullable_then_expression_is_nullable,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "foo".into(),
        ))),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: None,
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            then: Box::new(ast::Expression::Literal(ast::Literal::Null)),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
    }),
);

test_algebrize!(
    searched_case_nullable_else_expression_is_nullable,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: None,
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            then: Box::new(ast::Expression::Literal(ast::Literal::Null)),
        }],
        else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Null))),
    }),
);

test_algebrize!(
    simple_case,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "bar".into(),
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "foo".into(),
        ))),
        is_nullable: false,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
    }),
);

test_algebrize!(
    simple_case_no_else,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "bar".into(),
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: None,
    }),
);

test_algebrize_expr_and_schema_check!(
    simple_case_operand_and_when_operand_not_comparable,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::SchemaChecking(
        mir::schema::Error::InvalidComparison(
            "SimpleCase",
            Schema::Atomic(Atomic::Integer).into(),
            Schema::Atomic(Atomic::String).into(),
        )
    )),
    expected_error_code = 1005,
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::StringConstructor("foo".into())),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor("baz".into()))),
    }),
);

test_algebrize!(
    simple_case_nullable_then_expression_is_nullable,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "foo".into(),
        ))),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
            then: Box::new(ast::Expression::Literal(ast::Literal::Null)),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
    }),
);

test_algebrize!(
    simple_case_nullable_else_expression_is_nullable,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "bar".into(),
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
            then: Box::new(ast::Expression::StringConstructor("bar".into())),
        }],
        else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Null))),
    }),
);

test_algebrize!(
    simple_case_expr_not_string_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"2\"}".to_string()
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "{\"$numberInt\": \"3\"}".to_string()
        ))),
        is_nullable: false,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
            then: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
        }],
        else_branch: Some(Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"3\"}".to_string()
        ))),
    }),
);

test_algebrize!(
    simple_case_string_expr_does_not_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "hello".to_string()
        ))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"2\"}".to_string()
            ))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(3))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(4))),
        is_nullable: false,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::StringConstructor(
            "hello".to_string()
        ))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
            then: Box::new(ast::Expression::Literal(ast::Literal::Integer(3))),
        }],
        else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(4)))),
    }),
);

test_algebrize!(
    simple_case_where_all_strings_does_not_implicit_convert_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
        expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
            "{\"$numberInt\": \"1\"}".to_string()
        ))),
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"2\"}".to_string()
            ))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(3))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(4))),
        is_nullable: false,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: Some(Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"1\"}".to_string()
        ))),
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
            then: Box::new(ast::Expression::Literal(ast::Literal::Integer(3))),
        }],
        else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(4)))),
    }),
);

test_algebrize!(
    searched_case_implicit_converts_when_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
        when_branch: vec![mir::WhenBranch {
            when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
            then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"2\"}".to_string()
            ))),
            is_nullable: false,
        }],
        else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
        is_nullable: true,
    })),
    input = ast::Expression::Case(ast::CaseExpr {
        expr: None,
        when_branch: vec![ast::WhenBranch {
            when: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            then: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
        }],
        else_branch: None,
    }),
);
