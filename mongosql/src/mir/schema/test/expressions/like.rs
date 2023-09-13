use crate::{
    map,
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::{Atomic, Schema, NUMERIC_OR_NULLISH, STRING_OR_NULLISH},
    set, test_schema,
};

test_schema!(
    like_first_arg_not_string_or_nullish_is_error,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "Like",
        required: STRING_OR_NULLISH.clone(),
        found: NUMERIC_OR_NULLISH.clone(),
    }),
    input = Expression::Like(LikeExpr {
        expr: Expression::Reference(("bar", 0u16).into()).into(),
        pattern: Expression::Literal(LiteralValue::String("hello".into()).into()).into(),
        escape: None,
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
);

test_schema!(
    like_second_arg_not_string_or_nullish_is_error,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "Like",
        required: STRING_OR_NULLISH.clone(),
        found: NUMERIC_OR_NULLISH.clone(),
    }),
    input = Expression::Like(LikeExpr {
        expr: Expression::Literal(LiteralValue::String("hello".into()).into()).into(),
        pattern: Expression::Reference(("bar", 0u16).into()).into(),
        escape: None,
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
);

test_schema!(
    like_must_be_string,
    expected = Ok(Schema::Atomic(Atomic::Boolean)),
    input = Expression::Like(LikeExpr {
        expr: Expression::Reference(("bar", 0u16).into()).into(),
        pattern: Expression::Literal(LiteralValue::String("hello".into()).into()).into(),
        escape: None,
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
);

test_schema!(
    like_may_be_null,
    expected = Ok(Schema::AnyOf(set![
        Schema::Atomic(Atomic::Boolean),
        Schema::Atomic(Atomic::Null)
    ])),
    input = Expression::Like(LikeExpr {
        expr: Expression::Reference(("bar", 0u16).into()).into(),
        pattern: Expression::Literal(LiteralValue::String("hello".into()).into()).into(),
        escape: None,
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
);

test_schema!(
    like_may_be_missing,
    expected = Ok(Schema::AnyOf(set![
        Schema::Atomic(Atomic::Boolean),
        Schema::Atomic(Atomic::Null)
    ])),
    input = Expression::Like(LikeExpr {
        expr: Expression::Reference(("bar", 0u16).into()).into(),
        pattern: Expression::Literal(LiteralValue::String("hello".into()).into()).into(),
        escape: None,
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
);

test_schema!(
    like_must_be_null,
    expected = Ok(Schema::Atomic(Atomic::Null)),
    input = Expression::Like(LikeExpr {
        expr: Expression::Reference(("bar", 0u16).into()).into(),
        pattern: Expression::Literal(LiteralValue::String("hello".into()).into()).into(),
        escape: None,
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
);
