use crate::{
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::{Atomic, Schema},
    set, test_schema,
};

mod searched {
    use super::*;

    test_schema!(
        searched_case_when_branch_condition_must_be_boolean_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "SearchedCase",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Boolean),
                Schema::Atomic(Atomic::Null),
                Schema::Missing,
            ]),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
                then: Box::new(Expression::Literal(LiteralValue::Integer(2).into())),
                is_nullable: false,
            }],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
    );

    test_schema!(
        searched_case_with_no_when_branch_uses_else_branch,
        expected = Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Long)])),
        input = Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![],
            else_branch: Box::new(Expression::Literal(LiteralValue::Long(1).into())),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
    );

    test_schema!(
        searched_case_multiple_when_branches,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                    then: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                    then: Box::new(Expression::Literal(LiteralValue::Long(2).into())),
                    is_nullable: false,
                }
            ],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
    );
}

mod simple {
    use super::*;

    test_schema!(
        simple_case_when_branch_operand_must_be_comparable_with_case_operand,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "SimpleCase",
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Integer),
        )),
        input = Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )),
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
                then: Box::new(Expression::Literal(LiteralValue::Integer(2).into())),
                is_nullable: false,
            }],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
    );

    test_schema!(
        simple_case_with_no_when_branch_uses_else_branch,
        expected = Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Long)])),
        input = Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            when_branch: vec![],
            else_branch: Box::new(Expression::Literal(LiteralValue::Long(2).into())),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
    );

    test_schema!(
        simple_case_multiple_when_branches,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Integer(2).into())),
                    then: Box::new(Expression::Literal(LiteralValue::Integer(3).into())),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Long(4).into())),
                    then: Box::new(Expression::Literal(LiteralValue::Long(5).into())),
                    is_nullable: false,
                }
            ],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
    );
}
