use crate::{
    map,
    mir::{schema::Error as mir_error, *},
    schema::{
        Atomic, Document, Satisfaction, Schema, ANY_ARRAY, ANY_DOCUMENT, BOOLEAN_OR_NULLISH,
        INTEGER_OR_NULLISH, NON_NULLISH, NUMERIC_OR_NULLISH, STRING_OR_NULLISH,
    },
    set, test_schema,
};

mod substring {
    use super::*;

    test_schema!(
        substring_requires_string_for_first_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Substring",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
    );

    test_schema!(
        substring_requires_integer_for_second_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Substring",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::String("def".to_string())),
                Expression::Literal(LiteralValue::Integer(1))
            ],
        )),
    );

    test_schema!(
        substring_requires_integer_for_third_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Substring",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::String("def".to_string()))
            ],
        )),
    );

    test_schema!(
        substring_with_start_arg,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1))
            ],
        )),
    );

    test_schema!(
        substring_with_start_and_length_args,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        substring_with_null_arg,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        substring_with_potentially_null_arg,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Reference(("integer_or_null", 0u16).into())
            ],
        )),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)])},
    );
}

mod and {
    use super::*;

    test_schema!(
        and_first_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "And",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );

    test_schema!(
        and_second_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "And",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Reference(("bar", 0u16).into()),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );

    test_schema!(
        and_must_be_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );

    test_schema!(
        and_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        and_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );

    test_schema!(
        and_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod or {
    use super::*;

    test_schema!(
        or_first_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Or",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );

    test_schema!(
        or_second_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Or",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Reference(("bar", 0u16).into()),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );

    test_schema!(
        or_must_be_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );

    test_schema!(
        or_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        or_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );

    test_schema!(
        or_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod not {
    use super::*;

    test_schema!(
        not_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Not",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );

    test_schema!(
        not_must_be_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );

    test_schema!(
        not_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        not_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );

    test_schema!(
        not_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod trim {
    use super::*;

    test_schema!(
        ltrim_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::LTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );

    test_schema!(
        ltrim_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::LTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        ltrim_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::LTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        ltrim_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::LTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        rtrim_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::RTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );

    test_schema!(
        rtrim_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::RTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        rtrim_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::RTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        rtrim_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::RTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        btrim_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );

    test_schema!(
        btrim_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        btrim_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        btrim_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BTrim,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod concat {
    use super::*;

    test_schema!(
        concat_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );

    test_schema!(
        concat_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        concat_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        concat_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod lower {
    use super::*;

    test_schema!(
        lower_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lower,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );

    test_schema!(
        lower_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lower,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        lower_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lower,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        lower_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lower,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod upper {
    use super::*;

    test_schema!(
        upper_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Upper,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );

    test_schema!(
        upper_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Upper,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        upper_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Upper,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        upper_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Upper,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod year {
    use super::*;

    test_schema!(
        year_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Year,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );

    test_schema!(
        year_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Year,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        year_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Year,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );

    test_schema!(
        year_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Year,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod month {
    use super::*;

    test_schema!(
        month_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Month,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );

    test_schema!(
        month_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Month,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        month_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Month,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );

    test_schema!(
        month_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Month,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod day {
    use super::*;

    test_schema!(
        day_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Day,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );

    test_schema!(
        day_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Day,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        day_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Day,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );

    test_schema!(
        day_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Day,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod minute {
    use super::*;

    test_schema!(
        minute_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Minute,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );

    test_schema!(
        minute_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Minute,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        minute_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Minute,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );

    test_schema!(
        minute_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Minute,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod hour {
    use super::*;

    test_schema!(
        hour_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Hour,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );

    test_schema!(
        hour_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Hour,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        hour_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Hour,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );

    test_schema!(
        hour_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Hour,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
}

mod arithmetic {
    use super::*;

    test_schema!(
        variadic_arg_arithmetic_no_args_returns_integer,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![],
        )),
    );

    test_schema!(
        variadic_arg_arithmetic_one_arg_returns_that_type,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![Expression::Literal(LiteralValue::Double(1.0))],
        )),
    );

    test_schema!(
        arithmetic_null_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Double(2.0)),
                Expression::Literal(LiteralValue::Long(3))
            ],
        )),
    );

    test_schema!(
        arithmetic_missing_takes_priority_as_null_result,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Double(2.0)),
                Expression::Literal(LiteralValue::Long(3)),
                Expression::Reference(("bar", 0u16).into())
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Missing},
    );

    test_schema!(
        arithmetic_decimal_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Double(3.0))
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Decimal)},
    );

    test_schema!(
        arithmetic_double_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Double(1.0)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Long(3))
            ],
        )),
    );

    test_schema!(
        arithmetic_long_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
    );

    test_schema!(
        arithmetic_integer_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        arithmetic_results_in_any_numeric,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() =>Schema::AnyOf(set![
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
            ]),
            ("foo", 0u16).into() =>Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Integer),
            ]),

        },
    );

    test_schema!(
        arithmetic_decimal_double_takes_priority_in_any_of_must_be_numeric,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
            ]),
            ("foo", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ]),
            ("baz", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        arithmetic_decimal_double_takes_priority_in_any_of_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
            ]),
            ("foo", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
            ]),
            ("baz", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        arithmetic_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
            ]),
            ("foo", 0u16).into() =>
                Schema::Atomic(Atomic::Null),
            ("baz", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        arithmetic_decimal_double_takes_priority_in_any_of_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
            ]),
            ("foo", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Missing,
            ]),
            ("baz", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        arithmetic_must_be_missing,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
            ]),
            ("foo", 0u16).into() =>
                Schema::Missing,
            ("baz", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        arithmetic_nested_anyof,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Integer),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() =>Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Integer),
                Schema::AnyOf(set! [
                    Schema::Atomic(Atomic::Double),
                    Schema::AnyOf(set! [
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Long)
                    ]),
                ]),
            ]),
            ("foo", 0u16).into() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Long),
            ]),

        },
    );

    test_schema!(
        arithmetic_anyof_and_atomic,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() =>Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Long),
            ]),
            ("foo", 0u16).into() => Schema::Atomic(Atomic::Double),
        },
    );

    test_schema!(
        arithmetic_atomic_and_anyof,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Atomic(Atomic::Double),
            ("bar", 0u16).into() =>Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Long),
            ]),
        },
    );

    test_schema!(
        arithmetic_any_returns_all_numerics_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("foo", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Any,
        },
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );

    test_schema!(
        arithmetic_filter_out_non_numerics,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Atomic(Atomic::Double),
            ("bar", 0u16).into() =>Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::DbPointer),
                Schema::Atomic(Atomic::Timestamp),
            ]),
        },
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );

    test_schema!(
        arithmetic_numeric_and_non_numeric_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Add",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Atomic(Atomic::Double),
            ("bar", 0u16).into() =>Schema::Atomic(Atomic::String),
        },
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );

    mod errors {
        use super::*;

        test_schema!(
            sub_requires_exactly_two_args,
            expected_error_code = 1001,
            expected = Err(mir_error::IncorrectArgumentCount {
                name: "Sub",
                required: 2,
                found: 1
            }),
            input = Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Sub,
                vec![Expression::Literal(LiteralValue::Integer(1))],
            )),
        );

        test_schema!(
            div_requires_exactly_two_args,
            expected_error_code = 1001,
            expected = Err(mir_error::IncorrectArgumentCount {
                name: "Div",
                required: 2,
                found: 3
            }),
            input = Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Div,
                vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(3))
                ],
            )),
        );

        test_schema!(
            fixed_arg_arithmetic_first_arg_must_be_number,
            expected_error_code = 1002,
            expected = Err(mir_error::SchemaChecking {
                name: "Sub",
                required: Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing
                ]),
                found: Schema::Atomic(Atomic::String),
            }),
            input = Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Sub,
                vec![
                    Expression::Literal(LiteralValue::String("abc".to_string())),
                    Expression::Literal(LiteralValue::Integer(2)),
                ],
            )),
        );

        test_schema!(
            fixed_arg_arithmetic_second_arg_must_be_number,
            expected_error_code = 1002,
            expected = Err(mir_error::SchemaChecking {
                name: "Div",
                required: Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing
                ]),
                found: Schema::Atomic(Atomic::Boolean),
            }),
            input = Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Div,
                vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Boolean(true)),
                ],
            )),
        );

        test_schema!(
            variadic_arg_arithmetic_all_args_must_be_numbers,
            expected_error_code = 1002,
            expected = Err(mir_error::SchemaChecking {
                name: "Add",
                required: Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing
                ]),
                found: Schema::Atomic(Atomic::Boolean),
            }),
            input = Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Add,
                vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(3)),
                    Expression::Literal(LiteralValue::Boolean(true)),
                    Expression::Literal(LiteralValue::Integer(4)),
                ],
            )),
        );
    }
}

mod abs {
    use super::*;

    test_schema!(
        abs_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Abs",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Abs,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        abs_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Abs",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Abs,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );
}

mod ceil {
    use super::*;

    test_schema!(
        ceil_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Ceil",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Ceil,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        ceil_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Ceil",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Ceil,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );
}

mod degrees {
    use super::*;

    test_schema!(
        degrees_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Degrees",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Degrees,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        degrees_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Degrees",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Degrees,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );

    test_schema!(
        degrees_returns_double_schema_for_integer_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Degrees,
            vec![Expression::Literal(LiteralValue::Integer(2))],
        )),
    );

    test_schema!(
        degrees_returns_double_schema_for_long_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Degrees,
            vec![Expression::Literal(LiteralValue::Long(2))],
        )),
    );
}

mod floor {
    use super::*;

    test_schema!(
        floor_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Floor",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Floor,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        floor_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Floor",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Floor,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );
}

mod log {
    use super::*;

    test_schema!(
        log_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Log",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Log,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
    );

    test_schema!(
        log_first_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Log",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Log,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        )),
    );

    test_schema!(
        log_second_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Log",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Log,
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
            ],
        )),
    );
}

mod mod_func {
    use super::*;

    test_schema!(
        mod_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Mod",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mod,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
    );

    test_schema!(
        mod_first_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Mod",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mod,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        )),
    );

    test_schema!(
        mod_second_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Mod",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mod,
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
            ],
        )),
    );
}

mod pow {
    use super::*;

    test_schema!(
        pow_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Pow",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pow,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
    );

    test_schema!(
        pow_first_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Pow",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pow,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        )),
    );

    test_schema!(
        pow_second_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Pow",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pow,
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
            ],
        )),
    );
}

mod round {
    use super::*;

    test_schema!(
        round_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Round",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Round,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
    );

    test_schema!(
        round_first_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Round",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Round,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        )),
    );

    test_schema!(
        round_second_arg_must_be_integral_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Round",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Round,
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
            ],
        )),
    );
}

mod cos {
    use super::*;

    test_schema!(
        cos_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Cos",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Cos,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        cos_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Cos",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Cos,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );

    test_schema!(
        cos_returns_double_schema_for_integer_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Cos,
            vec![Expression::Literal(LiteralValue::Integer(2))],
        )),
    );
}

mod sin {
    use super::*;

    test_schema!(
        sin_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Sin",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sin,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        sin_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Sin",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sin,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );

    test_schema!(
        sin_returns_double_schema_for_double_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sin,
            vec![Expression::Literal(LiteralValue::Double(2.8))],
        )),
    );
}

mod tan {
    use super::*;

    test_schema!(
        tan_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Tan",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Tan,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        tan_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Tan",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Tan,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );

    test_schema!(
        tan_returns_double_schema_for_long_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sin,
            vec![Expression::Literal(LiteralValue::Long(2))],
        )),
    );
}

mod radians {
    use super::*;

    test_schema!(
        radians_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Radians",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Radians,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        radians_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Radians",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Radians,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );
}

mod sqrt {
    use super::*;

    test_schema!(
        sqrt_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Sqrt",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sqrt,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        sqrt_arg_must_be_number_or_nullish,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Sqrt",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sqrt,
            vec![Expression::Literal(LiteralValue::String("abc".to_string()))],
        )),
    );
}

mod comparison {
    use super::*;

    test_schema!(
        comp_op_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Lt",
            required: 2,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lt,
            vec![Expression::Literal(LiteralValue::Integer(1))],
        )),
    );

    test_schema!(
        comp_op_requires_a_valid_comparison,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "Lte",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lte,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::String("abc".to_string()))
            ],
        )),
    );

    test_schema!(
        comp_op_returns_boolean_schema_for_non_nullish_comparison,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Eq,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        comp_op_returns_boolean_and_null_schema_for_potentially_nullish_comparison,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Reference(("integer_or_null", 0u16).into()),
            ],
        )),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)])},
    );

    test_schema!(
        comp_op_returns_null_schema_for_nullish_comparison,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gte,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Null),
            ],
        )),
    );

    test_schema!(
    extended_json_error_oid,
    expected_error_code = 1017,
    expected = Err(mir_error::ExtJsonComparison("MongoSQL does not support direct comparisons with extended JSON. Try `= CAST(\"5ca4bbcea2dd94ee58162a6a\" as objectId)`.".to_string())),
    input = Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::Lte,
        args: vec![
            Expression::Literal(LiteralValue::Integer(1)),
            Expression::Literal(
                LiteralValue::String(r#"{"$oid":"5ca4bbcea2dd94ee58162a6a"}"#.to_string())
            )
        ],
        is_nullable: true,
    }),
);

    test_schema!(
        extended_json_error_timestamp,
        expected_error_code = 1017,
        expected = Err(mir_error::ExtJsonComparison(
            "MongoSQL does not support direct comparisons with extended JSON.".to_string()
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::String(
                    r#"{"$timestamp":{"t":1565545664,"i":1}}"#.to_string()
                ))
            ],
            is_nullable: true,
        }),
    );

    test_schema!(
        extended_json_error_number_double,
        expected_error_code = 1017,
        expected = Err(mir_error::ExtJsonComparison(
            "MongoSQL does not support direct comparisons with extended JSON. Try `= 55.55`."
                .to_string()
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                Expression::Literal(LiteralValue::Double(55.55)),
                Expression::Literal(LiteralValue::String(
                    r#"{"$numberDouble":"55.55"}"#.to_string()
                ))
            ],
            is_nullable: true,
        }),
    );

    test_schema!(
        extended_json_error_number_decimal,
        expected_error_code = 1017,
        expected = Err(mir_error::ExtJsonComparison(
            "MongoSQL does not support direct comparisons with extended JSON. Try `= 55.55`."
                .to_string()
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                Expression::Literal(LiteralValue::Double(55.55)),
                Expression::Literal(LiteralValue::String(
                    r#"{"$numberDecimal":"55.55"}"#.to_string()
                ))
            ],
            is_nullable: true,
        }),
    );
}

mod between {
    use super::*;

    test_schema!(
        between_requires_exactly_three_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Between",
            required: 3,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![Expression::Literal(LiteralValue::Integer(1))],
        )),
    );

    test_schema!(
        between_requires_a_valid_comparison_between_first_and_second_args,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "Between",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        )),
    );

    test_schema!(
        between_requires_a_valid_comparison_between_first_and_third_args,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "Between",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
            ],
        )),
    );

    test_schema!(
        between_returns_boolean_schema_for_non_nullish_comparisons,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Boolean)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Long(3))
            ],
        )),
    );

    test_schema!(
        between_returns_boolean_and_null_schema_for_potentially_nullish_comparison,
        expected = Ok(Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Boolean),
                Schema::Atomic(Atomic::Null)
            ]),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Boolean),
                Schema::Atomic(Atomic::Null)
            ]),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("long_or_null", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)]),
            ("long_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Long), Schema::Atomic(Atomic::Null)])
        },
    );

    test_schema!(
        between_returns_null_schema_for_nullish_comparison,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
            ],
        )),
    );
}

mod merge_objects {
    use super::*;

    test_schema!(
        merge_objects_args_must_be_documents,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "MergeObjects",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![Expression::Literal(LiteralValue::String("abc".to_string())),],
        )),
    );

    test_schema!(
        merge_objects_ok_to_be_one_any_document,
        expected = Ok(ANY_DOCUMENT.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![Expression::Reference(("bar", 0u16).into()),],
        )),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        merge_objects_not_ok_to_be_multiple_any_document,
        expected_error_code = 1006,
        expected = Err(mir_error::CannotMergeObjects(
            ANY_DOCUMENT.clone(),
            ANY_DOCUMENT.clone(),
            Satisfaction::May,
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        merge_objects_args_must_have_disjoint_keys,
        expected_error_code = 1006,
        expected = Err(mir_error::CannotMergeObjects(
            Schema::Document(Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Integer) },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Double) },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Satisfaction::Must,
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("car", 0u16).into())
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
                }),
            ("car", 0u16).into() => Schema::Document(
            Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
                }),
        },
    );

    test_schema!(
        merge_three_objects,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "a".into(), "d".into(), "e".into(),
            },
            additional_properties: false,
            ..Default::default()
        })),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
                }),
            ("baz", 0u16).into() => Schema::Document(
            Document {
                keys: map! {
                    "c".into() => Schema::Atomic(Atomic::Integer),
                    "d".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"d".into()},
                additional_properties: false,
                ..Default::default()
                }),
            ("foo", 0u16).into() => Schema::Document(
            Document {
                keys: map! {
                    "e".into() => Schema::Atomic(Atomic::Integer),
                    "f".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"e".into()},
                additional_properties: false,
                ..Default::default()
                }),
        },
    );

    test_schema!(
        merge_two_anyof_objects,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() =>  Schema::AnyOf(set!{
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Double),
                }),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "a".into(), "e".into(),
            },
            additional_properties: false,
            ..Default::default()
        })),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        )),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![
                Schema::Document(
                    Document {
                        keys: map! {
                           "a".into() => Schema::Atomic(Atomic::Integer),
                           "b".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"a".into()},
                    additional_properties: false,
                    ..Default::default()
                    }),
                Schema::Document(
                    Document {
                        keys: map! {
                           "a".into() => Schema::Atomic(Atomic::Integer),
                           "b".into() => Schema::Atomic(Atomic::Double),
                        },
                    required: set! {"a".into(), "b".into()},
                    additional_properties: false,
                    ..Default::default()
                    }),
                Schema::Document(
                    Document {
                        keys: map! {
                           "a".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"a".into()},
                    additional_properties: false,
                    ..Default::default()
                    }),
            ]),
            ("bar", 0u16).into() => Schema::AnyOf(set![
                Schema::Document(
                    Document {
                        keys: map! {
                           "c".into() => Schema::Atomic(Atomic::Integer),
                           "d".into() => Schema::Atomic(Atomic::Integer),
                           "e".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"e".into(), "d".into()},
                    additional_properties: false,
                    ..Default::default()
                    }),
                Schema::Document(
                    Document {
                        keys: map! {
                           "e".into() => Schema::Atomic(Atomic::Integer),
                           "f".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"e".into()},
                    additional_properties: false,
                    ..Default::default()
                    }),
            ]),
        },
    );
}

mod computed_field_access {
    use super::*;

    test_schema!(
        computed_field_access_requires_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "ComputedFieldAccess",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::ComputedFieldAccess,
            vec![
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Long(3))
            ],
        )),
    );

    test_schema!(
        computed_field_access_first_arg_must_not_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::Long),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::ComputedFieldAccess,
            vec![
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Long(2)),
            ],
        )),
    );

    test_schema!(
        computed_field_access_first_arg_may_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing]),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::ComputedFieldAccess,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("field".to_string())),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing])},
    );

    test_schema!(
        computed_field_access_second_arg_must_not_be_string,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: Schema::Atomic(Atomic::String),
            found: Schema::Atomic(Atomic::Long),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::ComputedFieldAccess,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Long(42)),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        computed_field_access_second_arg_may_be_string,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: Schema::Atomic(Atomic::String),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing]),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::ComputedFieldAccess,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone(),
        ("baz", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );

    test_schema!(
        computed_field_access_valid_args,
        expected = Ok(Schema::Any),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::ComputedFieldAccess,
            vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("field".to_string())),
            ],
        )),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
}

mod current_timestamp {
    use super::*;

    test_schema!(
        current_timestamp_no_arg,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::CurrentTimestamp,
            vec![],
        )),
    );

    test_schema!(
        current_timestamp_integer_arg_should_be_removed_in_algebrization,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "CurrentTimestamp",
            required: 0,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::CurrentTimestamp,
            vec![Expression::Literal(LiteralValue::Integer(1))],
        )),
    );
}

mod nullif {
    use super::*;

    test_schema!(
        nullif_requires_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "NullIf",
            required: 2,
            found: 1,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![Expression::Literal(LiteralValue::Integer(1))],
        )),
    );

    test_schema!(
        nullif_cannot_compare_numeric_with_non_numeric,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "NullIf",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::String("abc".to_string()))
            ],
        )),
    );

    test_schema!(
        nullif_types_must_be_identical_if_non_numeric,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "NullIf",
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::String)
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::String("abc".to_string()))
            ],
        )),
    );

    test_schema!(
        nullif_args_cannot_be_potentially_comparable,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "NullIf",
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String)
            ]),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String)
            ])
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
        )),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)]),
            ("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)])
        },
    );

    test_schema!(
        nullif_identical_types,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::String("def".to_string()))
            ],
        )),
    );

    test_schema!(
        nullif_missing_type_upconverts_to_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Reference(("missing", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
        )),
        schema_env = map! {
            ("missing", 0u16).into() => Schema::Missing,
        },
    );

    test_schema!(
        nullif_different_numerical_types_uses_first_arg_type,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Long(2))
            ],
        )),
    );

    test_schema!(
        nullif_multitype_numeric_args,
        expected = Ok(Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ]),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
        )),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)]),
            ("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Double), Schema::Atomic(Atomic::Decimal)])
        },
    );
}

mod coalesce {
    use super::*;

    test_schema!(
        coalesce_requires_at_least_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Coalesce",
            required: 1,
            found: 0,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![],
        )),
    );

    test_schema!(
        coalesce_returns_first_non_nullish_arg_omitting_null_and_missing,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("null_or_missing", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Double(2.0)),
            ],
        )),
        schema_env = map! { ("null_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]) },
    );

    test_schema!(
        coalesce_returns_up_to_first_non_nullish_arg_omitting_null_and_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Double(2.0)),
            ],
        )),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)]),
            ("null_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing])
        },
    );

    test_schema!(
        coalesce_with_any_schema_and_nullish_args_yields_non_nullish_with_null,
        expected = Ok(Schema::simplify(&Schema::AnyOf(set![
            NON_NULLISH.clone(),
            Schema::Atomic(Atomic::Null),
        ]))),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("any", 0u16).into()),
                Expression::Reference(("missing", 0u16).into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
            ],
        )),
        schema_env = map! {("any", 0u16).into() => Schema::Any,
            ("missing", 0u16).into() => Schema::Missing,
            ("null_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing])
        },
    );

    test_schema!(
        coalesce_with_any_schema_and_non_nullish_arg_yields_non_nullish_schema,
        expected = Ok(NON_NULLISH.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Reference(("any", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1))
            ],
        )),
        schema_env = map! {("any", 0u16).into() => Schema::Any},
    );

    test_schema!(
        coalesce_with_only_nullish_args_yields_only_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("missing", 0u16).into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
            ],
        )),
        schema_env = map! {("missing", 0u16).into() => Schema::Missing,
            ("null_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing])
        },
    );

    test_schema!(
        coalesce_with_all_potentially_nullish_args_includes_null_and_omits_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Long),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("long_or_missing", 0u16).into()),
            ],
        )),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)]),
            ("long_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Long), Schema::Missing])
        },
    );
}

mod slice {
    use super::*;

    test_schema!(
        slice_requires_more_than_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Slice",
            required: 2,
            found: 1,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![Expression::Reference(("array", 0u16).into())],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        slice_requires_fewer_than_four_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Slice",
            required: 2,
            found: 4,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        slice_with_two_args_requires_an_array_for_the_first_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Slice",
            required: ANY_ARRAY.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
        )),
    );

    test_schema!(
        slice_with_two_args_requires_an_integer_for_the_second_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Slice",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::Long),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        slice_with_three_args_requires_an_array_for_the_first_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Slice",
            required: ANY_ARRAY.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );

    test_schema!(
        slice_with_three_args_requires_an_integer_for_the_second_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Slice",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::String("abc".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        slice_with_three_args_requires_an_integer_for_the_third_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Slice",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::String("abc".to_string())),
            ],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        slice_with_length_arg,
        expected = Ok(ANY_ARRAY.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        slice_with_start_and_length_arg,
        expected = Ok(ANY_ARRAY.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Slice,
            vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
}

mod split {
    use super::*;

    test_schema!(
        split_requires_three_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Split",
            required: 3,
            found: 1,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Split,
            args: vec![Expression::Literal(LiteralValue::String(
                "a-b-c".to_string()
            ))],
            is_nullable: true,
        }),
    );

    test_schema!(
        split_requires_string_or_nullish_for_first_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Split",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Split,
            args: vec![
                Expression::Literal(LiteralValue::Integer(5)),
                Expression::Literal(LiteralValue::String("-".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }),
    );

    test_schema!(
        split_requires_string_or_nullish_for_second_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Split",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Split,
            args: vec![
                Expression::Literal(LiteralValue::String("a-b-c".to_string())),
                Expression::Literal(LiteralValue::Integer(5)),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }),
    );

    test_schema!(
        split_requires_integer_or_nullish_for_third_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Split",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Split,
            args: vec![
                Expression::Literal(LiteralValue::String("a-b-c".to_string())),
                Expression::Literal(LiteralValue::String("-".to_string())),
                Expression::Literal(LiteralValue::String("1".to_string())),
            ],
            is_nullable: true,
        }),
    );

    test_schema!(
        split_returns_null_if_empty_delimiter,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Split,
            args: vec![
                Expression::Literal(LiteralValue::String("a-b-c".to_string())),
                Expression::Literal(LiteralValue::String("".to_string())),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
            is_nullable: true,
        }),
    );
}

mod size {
    use super::*;

    test_schema!(
        size_requires_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Size",
            required: 1,
            found: 0,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Size,
            vec![],
        )),
    );

    test_schema!(
        size_requires_array_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Size",
            required: Schema::AnyOf(set![
                ANY_ARRAY.clone(),
                Schema::Atomic(Atomic::Null),
                Schema::Missing,
            ]),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Size,
            vec![Expression::Literal(LiteralValue::Integer(1))],
        )),
    );

    test_schema!(
        size_of_array,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Size,
            vec![Expression::Reference(("array", 0u16).into())],
        )),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    test_schema!(
        an_arg_that_may_be_nullish_manifests_as_null_in_final_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Size,
            vec![Expression::Reference(("array_or_null", 0u16).into())],
        )),
        schema_env = map! { ("array_or_null", 0u16).into() =>
        Schema::AnyOf(set![ANY_ARRAY.clone(), Schema::Atomic(Atomic::Null)]) },
    );
}

mod pos {
    use super::*;

    test_schema!(
        unary_pos,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pos,
            vec![Expression::Literal(LiteralValue::Integer(1))],
        )),
    );

    test_schema!(
        unary_pos_requires_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Pos",
            required: 1,
            found: 0
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pos,
            vec![],
        )),
    );

    test_schema!(
        arg_may_satisfy_schema_is_not_sufficient,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Pos",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pos,
            vec![Expression::Reference(("bar", 0u16).into())],
        )),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
}

mod neg {
    use super::*;

    test_schema!(
        unary_neg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pos,
            vec![Expression::Literal(LiteralValue::Double(1.0))],
        )),
    );

    test_schema!(
        unary_neg_requires_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Neg",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Neg,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        )),
    );
}
