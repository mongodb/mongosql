use crate::{
    map,
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::{Atomic, Schema},
    set, test_schema,
};

mod cast {
    use super::*;

    test_schema!(
        cast_expr_to_same_type,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(LiteralValue::Null.into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
            is_nullable: true,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        cast_expr_to_other_type,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            to: Type::Double,
            on_null: Box::new(Expression::Literal(LiteralValue::Null.into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
            is_nullable: true,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        cast_expr_to_other_type_with_on_null_and_on_error_set,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ])),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            to: Type::Double,
            on_null: Box::new(Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )),
            on_error: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
            is_nullable: false,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        cast_multi_type_expr_to_possible_type,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ])),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::Double,
            on_null: Box::new(Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )),
            on_error: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
            is_nullable: false,
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );

    test_schema!(
        cast_multi_type_expr_to_impossible_type,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ])),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::String,
            on_null: Box::new(Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )),
            on_error: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
            is_nullable: false,
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );

    test_schema!(
        cast_null_expr_to_type,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Null.into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(LiteralValue::Null.into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
            is_nullable: true,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        cast_null_expr_to_type_with_on_null_set,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Null.into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(LiteralValue::Double(1.0).into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
            is_nullable: true,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        cast_missing_expr_to_type,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(LiteralValue::Null.into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
            is_nullable: true,
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Missing},
    );

    test_schema!(
        cast_missing_expr_to_type_with_on_null_set,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(LiteralValue::Double(1.0).into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
            is_nullable: true,
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Missing},
    );
}

mod type_assert {
    use super::*;

    test_schema!(
        assert_expr_to_same_type,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::TypeAssertion(TypeAssertionExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            target_type: Type::Int32,
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        assert_multi_type_expr_to_possible_type,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::TypeAssertion(TypeAssertionExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            target_type: Type::Double,
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );

    test_schema!(
        assert_expr_to_impossible_type,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "::!",
            required: Schema::Atomic(Atomic::String),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::TypeAssertion(TypeAssertionExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            target_type: Type::String,
            cache: SchemaCache::new(),
        }),
    );
}
