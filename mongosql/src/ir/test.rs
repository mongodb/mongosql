macro_rules! test_schema {
    ($func_name:ident, $expected:expr, $input:expr,) => {
        test_schema!(
            $func_name,
            $expected,
            $input,
            crate::schema::SchemaEnvironment::default(),
        );
    };
    ($func_name:ident, $expected:expr, $input:expr, $schema_env:expr,) => {
        #[test]
        fn $func_name() {
            use crate::ir::schema::SchemaInferenceState;

            let expected = $expected;
            let input = $input;
            let schema_env = $schema_env;

            let state = SchemaInferenceState::from(&schema_env);
            let actual = input.schema(&state);

            assert_eq!(actual, expected);
        }
    };
}

mod schema {
    use crate::{
        ir::{schema::*, *},
        schema::*,
    };
    use common_macros::{b_tree_map, b_tree_set, hash_map};

    test_schema!(
        literal_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Literal(Literal::Null),
    );
    test_schema!(
        literal_bool,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::Literal(Literal::Boolean(true)),
    );
    test_schema!(
        literal_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::Literal(Literal::String("foobar".to_string())),
    );
    test_schema!(
        literal_int,
        Ok(Schema::Atomic(Atomic::Int)),
        Expression::Literal(Literal::Integer(5)),
    );
    test_schema!(
        literal_long,
        Ok(Schema::Atomic(Atomic::Long)),
        Expression::Literal(Literal::Long(6)),
    );
    test_schema!(
        literal_double,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::Literal(Literal::Double(7.0)),
    );
    test_schema!(
        reference_does_not_exist_in_schema_env,
        Err(Error::DatasourceNotFoundInSchemaEnv(("a", 0u16).into())),
        Expression::Reference(("a", 0u16).into()),
    );
    test_schema!(
        reference_exists_in_schema_env,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Reference(("a", 0u16).into()),
        hash_map! {("a", 0u16).into() => Schema::Atomic(Atomic::Null),},
    );

    // FieldAccess
    test_schema!(
        field_access_accessee_cannot_be_document,
        Err(Error::SchemaChecking(
            "FieldAccess",
            Schema::Atomic(Atomic::Long),
            crate::schema::ANY_DOCUMENT.clone(),
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Literal(Literal::Long(1))),
            field: "foo".to_string(),
        }),
    );
    test_schema!(
        field_access_field_must_not_exist_not_in_document,
        Err(Error::AccessMissingField("foo".to_string())),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        hash_map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: b_tree_map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
                required: b_tree_set!{"foof".to_string()},
                additional_properties: false,
            }
        ),},
    );
    test_schema!(
        field_access_field_may_exist,
        Ok(Schema::Any),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        hash_map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: b_tree_map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
                required: b_tree_set!{"foof".to_string()},
                additional_properties: true,
            }
        ),},
    );
    test_schema!(
        field_access_field_must_exist,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        hash_map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                required: b_tree_set!{"foo".to_string()},
                additional_properties: false,
            }
        ),},
    );
    test_schema!(
        field_access_field_must_one_of,
        Ok(Schema::AnyOf(
            vec! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Int)}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        hash_map! {("bar", 0u16).into() =>
            Schema::OneOf(vec!{
            Schema::Document(
                Document {
                    keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                    required: b_tree_set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Document(
                Document {
                    keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::Int)},
                    required: b_tree_set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
        })},
    );
    test_schema!(
        field_access_field_must_any_of_with_missing,
        Ok(Schema::AnyOf(
            vec! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Int), Schema::Missing}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        hash_map! {("bar", 0u16).into() =>
            Schema::AnyOf(vec!{
            Schema::Document(
                Document {
                    keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                    required: b_tree_set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Document(
                Document {
                    keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::Int)},
                    required: b_tree_set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Atomic(Atomic::Int),
        })},
    );

    // ComputedFieldAccess Function
    test_schema!(
        computed_field_access_requires_two_args,
        Err(Error::IncorrectArgumentCount("ComputedFieldAccess", 2, 3)),
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
            args: vec![
                Expression::Literal(Literal::Long(1)),
                Expression::Literal(Literal::Long(2)),
                Expression::Literal(Literal::Long(3))
            ],
        }),
    );
    test_schema!(
        computed_field_access_first_arg_must_not_be_document,
        Err(Error::SchemaChecking(
            "ComputedFieldAccess",
            Schema::Atomic(Atomic::Long),
            crate::schema::ANY_DOCUMENT.clone(),
        )),
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
            args: vec![
                Expression::Literal(Literal::Long(1)),
                Expression::Literal(Literal::Long(2)),
            ],
        }),
    );
    test_schema!(
        computed_field_access_second_arg_must_not_be_string,
        Err(Error::SchemaChecking(
            "ComputedFieldAccess",
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::String),
        )),
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Long(42)),
            ],
        }),
        hash_map! {("bar", 0u16).into() =>
            Schema::OneOf(vec!{
            Schema::Document(
                Document {
                    keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                    required: b_tree_set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Document(
                Document {
                    keys: b_tree_map!{"foo".to_string() => Schema::Atomic(Atomic::Int)},
                    required: b_tree_set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
        })},
    );
}
