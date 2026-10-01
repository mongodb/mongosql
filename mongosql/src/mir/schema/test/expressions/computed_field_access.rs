use crate::{
    map,
    mir::{schema::Error as mir_error, *},
    schema::{ANY_DOCUMENT, Atomic, Schema},
    set, test_schema,
};

test_schema!(
    computed_field_access_first_arg_must_not_be_document,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "ComputedFieldAccess",
        required: ANY_DOCUMENT.clone().into(),
        found: Schema::Atomic(Atomic::Long).into(),
        var_cause: None,
    }),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Literal(LiteralValue::Long(1))),
        Box::new(Expression::Literal(LiteralValue::Long(2))),
    )),
);

test_schema!(
    computed_field_access_first_arg_may_be_document,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "ComputedFieldAccess",
        required: ANY_DOCUMENT.clone().into(),
        found: Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing]).into(),
        var_cause: None,
    }),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Reference(("bar", 0u16).into())),
        Box::new(Expression::Literal(LiteralValue::String(
            "field".to_string()
        ))),
    )),
    schema_env =
        map! {("bar", 0u16).into() => Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing])},
);

test_schema!(
    computed_field_access_first_arg_invalid_with_var_cause,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "ComputedFieldAccess",
        required: ANY_DOCUMENT.clone().into(),
        found: Schema::Atomic(Atomic::Integer).into(),
        var_cause: Some("this".to_string()),
    }),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Variable(Variable {
            name: "this".to_string(),
            is_nullable: false,
        })),
        Box::new(Expression::Literal(LiteralValue::String(
            "field".to_string()
        ))),
    )),
    variables = map! {
        "this" => Schema::Atomic(Atomic::Integer),
    },
);

test_schema!(
    computed_field_access_second_arg_must_not_be_string,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "ComputedFieldAccess",
        required: Schema::Atomic(Atomic::String).into(),
        found: Schema::Atomic(Atomic::Long).into(),
        var_cause: None,
    }),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Reference(("bar", 0u16).into())),
        Box::new(Expression::Literal(LiteralValue::Long(42))),
    )),
    schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
);

test_schema!(
    computed_field_access_second_arg_may_be_string,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "ComputedFieldAccess",
        required: Schema::Atomic(Atomic::String).into(),
        found: Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing]).into(),
        var_cause: None,
    }),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Reference(("bar", 0u16).into())),
        Box::new(Expression::Reference(("baz", 0u16).into())),
    )),
    schema_env = map! {
        ("bar", 0u16).into() => ANY_DOCUMENT.clone(),
        ("baz", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])
    },
);

test_schema!(
    computed_field_access_second_arg_invalid_with_var_cause,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "ComputedFieldAccess",
        required: Schema::Atomic(Atomic::String).into(),
        found: Schema::Atomic(Atomic::Long).into(),
        var_cause: Some("this".to_string()),
    }),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Reference(("bar", 0u16).into())),
        Box::new(Expression::Variable(Variable {
            name: "this".to_string(),
            is_nullable: false,
        })),
    )),
    schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    variables = map! {
        "this" => Schema::Atomic(Atomic::Long),
    },
);

test_schema!(
    computed_field_access_valid_args,
    expected = Ok(Schema::Any),
    input = Expression::ComputedFieldAccess(ComputedFieldAccess::new(
        Box::new(Expression::Reference(("bar", 0u16).into())),
        Box::new(Expression::Literal(LiteralValue::String(
            "field".to_string()
        ))),
    )),
    schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
);
