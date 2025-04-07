use crate::{
    ast, map,
    mir::{self, binding_tuple::Key},
    multimap,
    schema::{
        Atomic, Document, Schema, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, NUMERIC_OR_NULLISH,
        STRING_OR_NULLISH,
    },
    set, test_algebrize, test_algebrize_expr_and_schema_check, unchecked_unique_linked_hash_map,
    usererror::UserError,
};

pub mod between;
pub mod binary;
pub mod date_function;
pub mod extract;
pub mod identifier_and_subpath;
pub mod literal;
pub mod scalar_function;
pub mod trim;
pub mod unary;

mod case {
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
}

mod type_operators {
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
}

mod like {
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
}
