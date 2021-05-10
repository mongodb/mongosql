macro_rules! linked_hash_map(
    { $($key:expr => $value:expr),+ $(,)?} => {
        {
            let mut m = ::linked_hash_map::LinkedHashMap::new();
            $(
                m.insert($key, $value);
            )+
            m
        }
     };
);

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

            assert_eq!(expected, actual);
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

    // Array Literals
    test_schema!(
        array_literal_empty,
        Ok(Schema::Array(Box::new(Schema::Any))),
        Expression::Array(vec![]),
    );
    test_schema!(
        array_literal_null,
        Ok(Schema::Array(Box::new(Schema::OneOf(vec![
            Schema::Atomic(Atomic::Null)
        ])))),
        Expression::Array(vec![Expression::Literal(Literal::Null)]),
    );
    test_schema!(
        array_literal_two_nulls,
        Ok(Schema::Array(Box::new(Schema::OneOf(vec![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])))),
        Expression::Array(vec![
            Expression::Literal(Literal::Null),
            Expression::Literal(Literal::Null)
        ]),
    );
    test_schema!(
        array_literal_null_or_string,
        Ok(Schema::Array(Box::new(Schema::OneOf(vec![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::String),
        ])))),
        Expression::Array(vec![
            Expression::Literal(Literal::Null),
            Expression::Literal(Literal::String("hello".to_string())),
            Expression::Literal(Literal::Null),
            Expression::Literal(Literal::String("world".to_string())),
        ]),
    );

    // Document Literal
    test_schema!(
        document_literal_empty,
        Ok(Schema::Document(Document {
            keys: b_tree_map! {},
            required: b_tree_set! {},
            additional_properties: false,
        })),
        Expression::Document(::linked_hash_map::LinkedHashMap::new()),
    );
    test_schema!(
        document_literal_all_required,
        Ok(Schema::Document(Document {
            keys: b_tree_map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::Atomic(Atomic::Null),
                "d".to_string() => Schema::Atomic(Atomic::Long),
            },
            required: b_tree_set! {
                "a".to_string(),
                "b".to_string(),
                "c".to_string(),
                "d".to_string(),
            },
            additional_properties: false,
        })),
        Expression::Document(linked_hash_map! {
            "a".to_string() => Expression::Literal(Literal::String("Hello".to_string())),
            "b".to_string() => Expression::Literal(Literal::String("World".to_string())),
            "c".to_string() => Expression::Literal(Literal::Null),
            "d".to_string() => Expression::Literal(Literal::Long(42)),
        }),
    );
    test_schema!(
        document_literal_some_keys_may_or_must_satisfy_missing,
        Ok(Schema::Document(Document {
            keys: b_tree_map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::Atomic(Atomic::Null),
                "d".to_string() => Schema::OneOf(vec![Schema::Atomic(Atomic::Null), Schema::Missing]),
            },
            required: b_tree_set! {
                "a".to_string(),
                "c".to_string(),
            },
            additional_properties: false,
        })),
        Expression::Document(linked_hash_map! {
            "a".to_string() => Expression::Literal(Literal::String("Hello".to_string())),
            "b".to_string() => Expression::Reference(("b", 0u16).into()),
            "c".to_string() => Expression::Literal(Literal::Null),
            "d".to_string() => Expression::Reference(("a", 0u16).into()),
        }),
        hash_map! {
            ("a", 0u16).into() => Schema::OneOf(vec![Schema::Atomic(Atomic::Null), Schema::Missing]),
            ("b", 0u16).into() => Schema::Missing,
        },
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
