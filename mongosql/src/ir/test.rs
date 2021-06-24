macro_rules! test_schema {
    ($func_name:ident, match $expected:pat, $input:expr,) => {
        test_schema!(
            $func_name,
            match $expected,
            $input,
            crate::schema::SchemaEnvironment::default(),
        );
    };
    ($func_name:ident, match $expected:pat, $input:expr, $schema_env:expr,) => {
        #[test]
        fn $func_name() {
            use crate::ir::schema::SchemaInferenceState;

            let input = $input;
            let schema_env = $schema_env;

            let state = SchemaInferenceState::new(0u16, schema_env);
            let actual = input.schema(&state);

            assert!(matches!(actual, $expected));
        }
    };
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

            let state = SchemaInferenceState::new(0u16, schema_env);
            let actual = input.schema(&state);

            assert_eq!(expected, actual);
        }
    };
}

mod schema {
    use crate::ir::binding_tuple::DatasourceName::Bottom;
    use crate::{
        ir::{schema::*, *},
        map,
        schema::*,
        set,
    };
    use lazy_static::lazy_static;

    lazy_static! {
        pub static ref TEST_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {
            "bar".into() => Schema::Atomic(Atomic::Integer),
                },
            required: set! {"bar".into()},
            additional_properties: false,
        });
    }

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
        Ok(Schema::Atomic(Atomic::Integer)),
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
        map! {("a", 0u16).into() => Schema::Atomic(Atomic::Null),},
    );

    // Array Literals
    test_schema!(
        array_literal_empty,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![])))),
        Expression::Array(vec![]),
    );
    test_schema!(
        array_literal_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Null)
        ])))),
        Expression::Array(vec![Expression::Literal(Literal::Null)]),
    );
    test_schema!(
        array_literal_two_nulls,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])))),
        Expression::Array(vec![
            Expression::Literal(Literal::Null),
            Expression::Literal(Literal::Null)
        ]),
    );
    test_schema!(
        array_literal_missing_to_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Null),
        ])))),
        Expression::Array(vec![Expression::Reference(("a", 0u16).into()),]),
        map! {("a", 0u16).into() => Schema::Missing,},
    );
    test_schema!(
        array_literal_with_nested_document_missing_preserved,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Document(Document {
                keys: map! {
                "bar".into() => Schema::Atomic(Atomic::String),
                    },
                required: set! {"bar".into()},
                additional_properties: false,
            })
        ])))),
        Expression::Array(vec![Expression::Document(map! {
            "foo".into() => Expression::Reference(("a", 0u16).into()),
            "bar".into() => Expression::Reference(("b", 0u16).into()),
        }),]),
        map! {
            ("a", 0u16).into() => Schema::Missing,
            ("b", 0u16).into() => Schema::Atomic(Atomic::String),
        },
    );
    test_schema!(
        array_literal_any_of_any_of_missing_to_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Document(Document {
                keys: map! {"b".into() =>
                    Schema::AnyOf(
                    vec![
                        Schema::AnyOf(
                        vec![
                            Schema::Missing,
                            Schema::Atomic(Atomic::Integer)
                        ]),
                        Schema::Atomic(Atomic::Double),
                    ]),
                },
                required: set! {},
                additional_properties: false,
            })
        ])))),
        Expression::Array(vec![Expression::Document(
            map! {"b".into() => Expression::Reference(("a", 0u16).into())},
        )]),
        map! {("a", 0u16).into() =>
        Schema::AnyOf(
        vec![
            Schema::AnyOf(
            vec![
                Schema::Missing,
                Schema::Atomic(Atomic::Integer)
            ]),
            Schema::Atomic(Atomic::Double),
        ]),},
    );
    test_schema!(
        array_of_array_of_literal_any_of_any_of_missing_to_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Array(Box::new(Schema::AnyOf(vec![Schema::AnyOf(vec![
                Schema::AnyOf(vec![
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer)
                ]),
                Schema::Atomic(Atomic::Double)
            ])]))),
            Schema::Array(Box::new(Schema::AnyOf(vec![Schema::AnyOf(vec![
                Schema::AnyOf(vec![
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer)
                ]),
                Schema::Atomic(Atomic::Double)
            ])])))
        ])))),
        Expression::Array(vec![
            Expression::Array(vec![Expression::Reference(("a", 0u16).into()),]),
            Expression::Array(vec![Expression::Reference(("a", 0u16).into()),]),
        ]),
        map! {("a", 0u16).into() =>
        Schema::AnyOf(
        vec![
            Schema::AnyOf(
            vec![
                Schema::Missing,
                Schema::Atomic(Atomic::Integer)
            ]),
            Schema::Atomic(Atomic::Double),
        ]),},
    );
    test_schema!(
        array_literal_null_or_string,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
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
            keys: map! {},
            required: set! {},
            additional_properties: false,
        })),
        Expression::Document(map! {}),
    );
    test_schema!(
        document_literal_all_required,
        Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::Atomic(Atomic::Null),
                "d".to_string() => Schema::Atomic(Atomic::Long),
            },
            required: set! {
                "a".to_string(),
                "b".to_string(),
                "c".to_string(),
                "d".to_string(),
            },
            additional_properties: false,
        })),
        Expression::Document(map! {
            "a".to_string() => Expression::Literal(Literal::String("Hello".to_string())),
            "b".to_string() => Expression::Literal(Literal::String("World".to_string())),
            "c".to_string() => Expression::Literal(Literal::Null),
            "d".to_string() => Expression::Literal(Literal::Long(42)),
        }),
    );
    test_schema!(
        document_literal_some_keys_may_or_must_satisfy_missing,
        Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::Atomic(Atomic::Null),
                "d".to_string() => Schema::AnyOf(vec![Schema::Atomic(Atomic::Null), Schema::Missing]),
            },
            required: set! {
                "a".to_string(),
                "c".to_string(),
            },
            additional_properties: false,
        })),
        Expression::Document(map! {
            "a".to_string() => Expression::Literal(Literal::String("Hello".to_string())),
            "b".to_string() => Expression::Reference(("b", 0u16).into()),
            "c".to_string() => Expression::Literal(Literal::Null),
            "d".to_string() => Expression::Reference(("a", 0u16).into()),
        }),
        map! {
            ("a", 0u16).into() => Schema::AnyOf(vec![Schema::Atomic(Atomic::Null), Schema::Missing]),
            ("b", 0u16).into() => Schema::Missing,
        },
    );

    // FieldAccess
    test_schema!(
        field_access_accessee_cannot_be_document,
        Err(Error::SchemaChecking {
            name: "FieldAccess",
            required: crate::schema::ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::Long),
        }),
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
        map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foof".to_string()},
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
        map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foof".to_string()},
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
        map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foo".to_string()},
                additional_properties: false,
            }
        ),},
    );
    test_schema!(
        field_access_field_must_any_of,
        Ok(Schema::AnyOf(
            vec! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer)}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        map! {("bar", 0u16).into() =>
            Schema::AnyOf(vec!{
            Schema::Document(
                Document {
                    keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                    required: set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Document(
                Document {
                    keys: map!{"foo".to_string() => Schema::Atomic(Atomic::Integer)},
                    required: set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
        })},
    );
    test_schema!(
        field_access_field_must_any_of_with_missing,
        Ok(Schema::AnyOf(
            vec! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer), Schema::Missing}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        map! {("bar", 0u16).into() =>
            Schema::AnyOf(vec!{
            Schema::Document(
                Document {
                    keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                    required: set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Document(
                Document {
                    keys: map!{"foo".to_string() => Schema::Atomic(Atomic::Integer)},
                    required: set!{"foo".to_string()},
                    additional_properties: false,
                }
            ),
            Schema::Atomic(Atomic::Integer),
        })},
    );

    // General function schema checking.
    test_schema!(
        arg_may_satisfy_schema_is_not_sufficient,
        Err(Error::SchemaChecking {
            name: "Pos",
            required: Schema::AnyOf(vec![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::AnyOf(vec![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Reference(("bar", 0u16).into())],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );

    // Unary functions.
    test_schema!(
        unary_pos,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        unary_neg,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Literal(Literal::Double(1.0))],
        }),
    );
    test_schema!(
        unary_pos_requires_one_arg,
        Err(Error::IncorrectArgumentCount {
            name: "Pos",
            required: 1,
            found: 0
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![],
        }),
    );
    test_schema!(
        unary_neg_requires_one_arg,
        Err(Error::IncorrectArgumentCount {
            name: "Neg",
            required: 1,
            found: 2
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Neg,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2))
            ],
        }),
    );

    // Arithmetic function type correctness.
    test_schema!(
        variadic_arg_arithmetic_no_args_returns_integer,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![],
        }),
    );
    test_schema!(
        variadic_arg_arithmetic_one_arg_returns_that_type,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![Expression::Literal(Literal::Double(1.0))],
        }),
    );
    test_schema!(
        arithmetic_null_takes_priority,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Null),
                Expression::Literal(Literal::Double(2.0)),
                Expression::Literal(Literal::Long(3))
            ],
        }),
    );
    test_schema!(
        arithmetic_missing_takes_priority_as_null_result,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Double(2.0)),
                Expression::Literal(Literal::Long(3)),
                Expression::Reference(("bar", 0u16).into())
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Missing},
    );
    test_schema!(
        arithmetic_decimal_takes_priority,
        Ok(Schema::Atomic(Atomic::Decimal)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Long(2)),
                Expression::Literal(Literal::Double(3.0))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Decimal)},
    );
    test_schema!(
        arithmetic_double_takes_priority,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(Literal::Double(1.0)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::Long(3))
            ],
        }),
    );
    test_schema!(
        arithmetic_long_takes_priority,
        Ok(Schema::Atomic(Atomic::Long)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Long(2)),
                Expression::Literal(Literal::Integer(3))
            ],
        }),
    );
    test_schema!(
        arithmetic_integer_takes_priority,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2))
            ],
        }),
    );

    // Arithmetic function errors.
    test_schema!(
        sub_requires_exactly_two_args,
        Err(Error::IncorrectArgumentCount {
            name: "Sub",
            required: 2,
            found: 1
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sub,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        div_requires_exactly_two_args,
        Err(Error::IncorrectArgumentCount {
            name: "Div",
            required: 2,
            found: 3
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Div,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::Integer(3))
            ],
        }),
    );
    test_schema!(
        fixed_arg_arithmetic_first_arg_must_be_number,
        Err(Error::SchemaChecking {
            name: "Sub",
            required: Schema::AnyOf(vec![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sub,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(2)),
            ],
        }),
    );
    test_schema!(
        fixed_arg_arithmetic_second_arg_must_be_number,
        Err(Error::SchemaChecking {
            name: "Div",
            required: Schema::AnyOf(vec![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::Boolean),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Div,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Boolean(true)),
            ],
        }),
    );
    test_schema!(
        variadic_arg_arithmetic_all_args_must_be_numbers,
        Err(Error::SchemaChecking {
            name: "Add",
            required: Schema::AnyOf(vec![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::Boolean),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::Integer(3)),
                Expression::Literal(Literal::Boolean(true)),
                Expression::Literal(Literal::Integer(4)),
            ]
        }),
    );

    // ComputedFieldAccess Function
    test_schema!(
        computed_field_access_requires_two_args,
        Err(Error::IncorrectArgumentCount {
            name: "ComputedFieldAccess",
            required: 2,
            found: 3
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Literal(Literal::Long(1)),
                Expression::Literal(Literal::Long(2)),
                Expression::Literal(Literal::Long(3))
            ],
        }),
    );
    test_schema!(
        computed_field_access_first_arg_must_not_be_document,
        Err(Error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::Long),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Literal(Literal::Long(1)),
                Expression::Literal(Literal::Long(2)),
            ],
        }),
    );
    test_schema!(
        computed_field_access_first_arg_may_be_document,
        Err(Error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(vec![ANY_DOCUMENT.clone(), Schema::Missing]),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("field".to_string())),
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(vec![ANY_DOCUMENT.clone(), Schema::Missing])},
    );
    test_schema!(
        computed_field_access_second_arg_must_not_be_string,
        Err(Error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: Schema::Atomic(Atomic::String),
            found: Schema::Atomic(Atomic::Long),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Long(42)),
            ],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        computed_field_access_second_arg_may_be_string,
        Err(Error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: Schema::Atomic(Atomic::String),
            found: Schema::AnyOf(vec![Schema::Atomic(Atomic::String), Schema::Missing]),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone(),
        ("baz", 0u16).into() => Schema::AnyOf(vec![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        computed_field_access_valid_args,
        Ok(Schema::Any),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("field".to_string())),
            ],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    // Datetime value scalar function.
    test_schema!(
        current_timestamp_no_arg,
        Ok(Schema::Atomic(Atomic::Date)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::CurrentTimestamp,
            args: vec![],
        }),
    );
    test_schema!(
        current_timestamp_integer_arg_should_be_removed_in_algebrization,
        Err(Error::IncorrectArgumentCount {
            name: "CurrentTimestamp",
            required: 0,
            found: 1
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::CurrentTimestamp,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );

    test_schema!(
        collection_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        Stage::Collection(Collection {
            db: "test2".into(),
            collection: "foo".into(),
        }),
    );

    test_schema!(
        empty_array_datasource_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(vec![])
            },
            min_size: 0,
            max_size: Some(0),
        }),
        Stage::Array(Array {
            array: vec![],
            alias: "foo".into(),
        }),
    );
    test_schema!(
        dual_array_datasource_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(
                    vec![
                        Schema::Document( Document {
                            keys: map!{},
                            required: set!{},
                            additional_properties: false,
                        })
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        Stage::Array(Array {
            array: vec![Expression::Document(map! {})],
            alias: "foo".into(),
        }),
    );
    test_schema!(
        literal_array_items_datasource_schema,
        Err(Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(vec![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ])
        }),
        Stage::Array(Array {
            array: vec![
                Expression::Literal(Literal::Integer(42)),
                Expression::Literal(Literal::Double(42f64))
            ],
            alias: "foo".into(),
        }),
    );
    test_schema!(
        single_document_array_datasource_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(
                    vec![
                        Schema::Document( Document {
                            keys: map!{"bar".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"bar".into()},
                            additional_properties: false,
                        })
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        Stage::Array(Array {
            array: vec![Expression::Document(map! {
                "bar".into() => Expression::Literal(Literal::Integer(1))
            })],
            alias: "foo".into(),
        }),
    );
    test_schema!(
        two_document_array_datasource_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(
                    vec![
                        Schema::Document( Document {
                            keys: map!{"bar".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"bar".into()},
                            additional_properties: false,
                        }),
                        Schema::Document( Document {
                            keys: map!{"car".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"car".into()},
                            additional_properties: false,
                        }),
                    ]
                ),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        Stage::Array(Array {
            array: vec![
                Expression::Document(map! {
                "bar".into() => Expression::Literal(Literal::Integer(1))
                }),
                Expression::Document(map! {
                "car".into() => Expression::Literal(Literal::Integer(1))
                })
            ],
            alias: "foo".into(),
        }),
    );

    test_schema!(
        project_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("bar1", 0u16).into() => ANY_DOCUMENT.clone(),
                ("bar2", 0u16).into() => ANY_DOCUMENT.clone(),
                ("bar3", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test2".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                ("bar1", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
                ("bar2", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
                ("bar3", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
            }
        }),
    );

    mod filter {
        use crate::{
            ir::{schema::Error, *},
            map,
            schema::{Atomic, ResultSet, Schema},
        };

        const TRUE: Expression = Expression::Literal(Literal::Boolean(true));
        lazy_static::lazy_static! {
            static ref TEST_SOURCE: Stage = Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into()
            });
        }

        test_schema!(
            boolean_condition_allowed,
            match Ok(_),
            Stage::Filter(Filter {
                source: Box::new(TEST_SOURCE.clone()),
                condition: TRUE,
            }),
        );
        test_schema!(
            null_condition_allowed,
            match Ok(_),
            Stage::Filter(Filter {
                source: Box::new(TEST_SOURCE.clone()),
                condition: Expression::Literal(Literal::Null),
            }),
        );
        test_schema!(
            missing_condition_allowed,
            match Ok(_),
            Stage::Filter(Filter {
                source: Box::new(TEST_SOURCE.clone()),
                condition: Expression::Reference(("m", 0u16).into()),
            }),
            map! {("m", 0u16).into() => Schema::Missing},
        );
        test_schema!(
            non_boolean_condition_disallowed,
            Err(Error::SchemaChecking {
                name: "filter condition",
                required: Schema::AnyOf(vec![
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing,
                ]),
                found: Schema::Atomic(Atomic::Integer),
            }),
            Stage::Filter(Filter {
                source: Box::new(TEST_SOURCE.clone()),
                condition: Expression::Literal(Literal::Integer(123)),
            }),
        );
        test_schema!(
            possible_non_boolean_condition_disallowed,
            Err(Error::SchemaChecking {
                name: "filter condition",
                required: Schema::AnyOf(vec![
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing,
                ]),
                found: Schema::Any,
            }),
            Stage::Filter(Filter {
                source: Box::new(TEST_SOURCE.clone()),
                condition: Expression::FieldAccess(FieldAccess {
                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                    field: "bar".into(),
                }),
            }),
        );
        test_schema!(
            source_fails_schema_check,
            match Err(Error::SchemaChecking {
                name: "array datasource items",
                ..
            }),
            Stage::Filter(Filter {
                source: Stage::Array(Array {
                    alias: "arr".into(),
                    array: vec![Expression::Literal(Literal::Null)],
                }).into(),
                condition: TRUE,
            }),
        );
        test_schema!(
            condition_fails_schema_check,
            Err(Error::DatasourceNotFoundInSchemaEnv(("abc", 0u16).into())),
            Stage::Filter(Filter {
                source: Box::new(TEST_SOURCE.clone()),
                condition: Expression::Reference(("abc", 0u16).into()),
            }),
        );
        test_schema!(
            min_size_reduced_to_zero_max_size_preserved,
            match Ok(ResultSet{
                min_size: 0,
                max_size: Some(1),
                ..
            }),
            Stage::Filter(Filter {
                condition: TRUE,
                source: Stage::Array(Array {
                    alias: "arr".into(),
                    array: vec![Expression::Document(map!{"a".into() => Expression::Literal(Literal::Null),})],
                }).into(),
            }),
        );
    }

    // Cast.
    test_schema!(
        cast_expr_to_same_type,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Null)),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        cast_expr_to_other_type,
        Ok(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            to: Type::Double,
            on_null: Box::new(Expression::Literal(Literal::Null)),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        cast_expr_to_other_type_with_on_null_and_on_error_set,
        Ok(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ])),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            to: Type::Double,
            on_null: Box::new(Expression::Literal(Literal::String("abc".to_string()))),
            on_error: Box::new(Expression::Literal(Literal::Boolean(true))),
        }),
    );
    test_schema!(
        cast_multi_type_expr_to_possible_type,
        Ok(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ])),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::Double,
            on_null: Box::new(Expression::Literal(Literal::String("abc".to_string()))),
            on_error: Box::new(Expression::Literal(Literal::Boolean(true))),
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );
    test_schema!(
        cast_multi_type_expr_to_impossible_type,
        Ok(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ])),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::String,
            on_null: Box::new(Expression::Literal(Literal::String("abc".to_string()))),
            on_error: Box::new(Expression::Literal(Literal::Boolean(true))),
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );
    test_schema!(
        cast_null_expr_to_type,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(Literal::Null)),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Null)),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        cast_null_expr_to_type_with_on_null_set,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(Literal::Null)),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Double(1.0))),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        cast_missing_expr_to_type,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Null)),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
        map! {("bar", 0u16).into() => Schema::Missing},
    );
    test_schema!(
        cast_missing_expr_to_type_with_on_null_set,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::Cast(CastExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Double(1.0))),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
        map! {("bar", 0u16).into() => Schema::Missing},
    );

    // TypeAssert.
    test_schema!(
        assert_expr_to_same_type,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::TypeAssertion(TypeAssertionExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            target_type: Type::Int32,
        }),
    );
    test_schema!(
        assert_multi_type_expr_to_possible_type,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::TypeAssertion(TypeAssertionExpr {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            target_type: Type::Double,
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(vec![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );
    test_schema!(
        assert_expr_to_impossible_type,
        Err(Error::SchemaChecking {
            name: "::!",
            required: Schema::Atomic(Atomic::String),
            found: Schema::Atomic(Atomic::Integer),
        }),
        Expression::TypeAssertion(TypeAssertionExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            target_type: Type::String,
        }),
    );

    // Limit and Offset
    test_schema!(
        limit_collection_datasource,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: Some(20),
        }),
        Stage::Limit(Limit {
            limit: 20,
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
        }),
    );
    test_schema!(
        limit_lt_num_docs,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(vec![
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone()
                ]),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        Stage::Limit(Limit {
            limit: 2,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        limit_gt_num_docs,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(vec![
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                ]),
            },
            min_size: 3,
            max_size: Some(3),
        }),
        Stage::Limit(Limit {
            limit: 10,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        offset_collection_datasource,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        Stage::Offset(Offset {
            offset: 20,
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
        }),
    );
    test_schema!(
        offset_lt_num_docs,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(vec![
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                ]),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        Stage::Offset(Offset {
            offset: 1,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        offset_gt_num_docs,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(vec![
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                    TEST_DOCUMENT_SCHEMA.clone(),
                ]),
            },
            min_size: 0,
            max_size: Some(0),
        }),
        Stage::Offset(Offset {
            offset: 10,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "bar".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
    // Exists subquery
    test_schema!(
        exists_uncorrelated,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::Exists(Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
        }))),
    );

    test_schema!(
        exists_correlated,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                (Bottom, 1u16).into() => Expression::Document(map! {
                    "a".into() => Expression::FieldAccess(FieldAccess{
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    })
                })
            }
        }))),
        map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
    );

    test_schema!(
        exists_invalid_expression,
        Err(Error::IncorrectArgumentCount {
            name: "Div",
            required: 2,
            found: 3
        }),
        Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                ("a", 0u16).into() =>
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(2)),
                            Expression::Literal(Literal::Integer(3))
                        ],
                    })
            }
        }))),
    );

    // Subquery Expression

    test_schema!(
        subquery_output_expr_violates_type_constraints,
        Err(Error::AccessMissingField("_2".into())),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "_2".into()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(map! {
                        "_1".into() => Expression::Literal(Literal::Integer(5))
                    })
                }
            }))
        }),
    );

    // Analogous SQL query: SELECT (SELECT foo.a FROM []) FROM foo
    test_schema!(
        correlated_subquery,
        Ok(Schema::AnyOf(vec![
            Schema::Atomic(Atomic::String),
            Schema::Missing
        ])),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Array(Array {
                    array: vec![],
                    alias: "bar".into(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        }),
        map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
    );

    test_schema!(
        subquery_output_expr_correlated_datasource,
        Err(Error::DatasourceNotFoundInSchemaEnv(("foo", 0u16).into())),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "a".into()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        }),
        map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [] AS foo)"
    test_schema!(
        uncorrelated_subquery_cardinality_is_zero,
        Ok(Schema::AnyOf(vec![Schema::AnyOf(vec![]), Schema::Missing])),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Array(Array {
                array: vec![],
                alias: "foo".into(),
            })),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_must_be_one,
        Ok(Schema::AnyOf(vec![Schema::Atomic(Atomic::Integer)])),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "a".into()
            })),
            subquery: Box::new(Stage::Array(Array {
                array: vec![Expression::Document(map! {
                    "a".into() => Expression::Literal(Literal::Integer(5))
                }),],
                alias: "foo".into(),
            })),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM foo)"
    test_schema!(
        subquery_cardinality_may_be_1,
        Err(Error::InvalidSubqueryCardinality),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}, {'a': 6}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_gt_one,
        Err(Error::InvalidSubqueryCardinality),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "a".into()
            })),
            subquery: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(5))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(6))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
}
