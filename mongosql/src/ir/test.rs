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

            let state = SchemaInferenceState::new(0u16, schema_env);
            let actual = input.schema(&state);

            assert_eq!(expected, actual);
        }
    };
}

mod schema {
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
        array_literal_any_of_one_of_missing_to_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Document(Document {
                keys: map! {"b".into() =>
                    Schema::AnyOf(
                    vec![
                        Schema::OneOf(
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
            Schema::OneOf(
            vec![
                Schema::Missing,
                Schema::Atomic(Atomic::Integer)
            ]),
            Schema::Atomic(Atomic::Double),
        ]),},
    );
    test_schema!(
        array_of_array_of_literal_any_of_one_of_missing_to_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(vec![
            Schema::Array(Box::new(Schema::AnyOf(vec![Schema::AnyOf(vec![
                Schema::OneOf(vec![
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer)
                ]),
                Schema::Atomic(Atomic::Double)
            ])]))),
            Schema::Array(Box::new(Schema::AnyOf(vec![Schema::AnyOf(vec![
                Schema::OneOf(vec![
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
            Schema::OneOf(
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
                "d".to_string() => Schema::OneOf(vec![Schema::Atomic(Atomic::Null), Schema::Missing]),
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
            ("a", 0u16).into() => Schema::OneOf(vec![Schema::Atomic(Atomic::Null), Schema::Missing]),
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
        field_access_field_must_one_of,
        Ok(Schema::AnyOf(
            vec! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer)}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        map! {("bar", 0u16).into() =>
            Schema::OneOf(vec!{
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
        Expression::Function(FunctionApplication {
            function: Function::Pos,
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
        Expression::Function(FunctionApplication {
            function: Function::Pos,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        unary_neg,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::Function(FunctionApplication {
            function: Function::Pos,
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
        Expression::Function(FunctionApplication {
            function: Function::Pos,
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
        Expression::Function(FunctionApplication {
            function: Function::Neg,
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
        Expression::Function(FunctionApplication {
            function: Function::Add,
            args: vec![],
        }),
    );
    test_schema!(
        variadic_arg_arithmetic_one_arg_returns_that_type,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::Function(FunctionApplication {
            function: Function::Mul,
            args: vec![Expression::Literal(Literal::Double(1.0))],
        }),
    );
    test_schema!(
        arithmetic_null_takes_priority,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Function(FunctionApplication {
            function: Function::Mul,
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
        Expression::Function(FunctionApplication {
            function: Function::Mul,
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
        Expression::Function(FunctionApplication {
            function: Function::Add,
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
        Expression::Function(FunctionApplication {
            function: Function::Mul,
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
        Expression::Function(FunctionApplication {
            function: Function::Add,
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
        Expression::Function(FunctionApplication {
            function: Function::Mul,
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
        Expression::Function(FunctionApplication {
            function: Function::Sub,
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
        Expression::Function(FunctionApplication {
            function: Function::Div,
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
        Expression::Function(FunctionApplication {
            function: Function::Sub,
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
        Expression::Function(FunctionApplication {
            function: Function::Div,
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
        Expression::Function(FunctionApplication {
            function: Function::Add,
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
        Err(Error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::Long),
        }),
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
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
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
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
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
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
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
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
        Expression::Function(FunctionApplication {
            function: Function::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("field".to_string())),
            ],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        collection_schema,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Any,
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
                ("bar1", 0u16).into() => Schema::Any,
                ("bar2", 0u16).into() => Schema::Any,
                ("bar3", 0u16).into() => Schema::Any,
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

    // Cast.
    test_schema!(
        cast_expr_to_same_type,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::Cast(CastExpression {
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
        Expression::Cast(CastExpression {
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
        Expression::Cast(CastExpression {
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
        Expression::Cast(CastExpression {
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
        Expression::Cast(CastExpression {
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
        Expression::Cast(CastExpression {
            expr: Box::new(Expression::Literal(Literal::Null)),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Null)),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        cast_null_expr_to_type_with_on_null_set,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::Cast(CastExpression {
            expr: Box::new(Expression::Literal(Literal::Null)),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(Literal::Double(1.0))),
            on_error: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        cast_missing_expr_to_type,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Cast(CastExpression {
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
        Expression::Cast(CastExpression {
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
        Expression::TypeAssertion(TypeAssertionExpression {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            target_type: Type::Int32,
        }),
    );
    test_schema!(
        assert_multi_type_expr_to_possible_type,
        Ok(Schema::Atomic(Atomic::Double)),
        Expression::TypeAssertion(TypeAssertionExpression {
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
        Expression::TypeAssertion(TypeAssertionExpression {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            target_type: Type::String,
        }),
    );

    // Limit and Offset
    test_schema!(
        limit_collection_datasource,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Any,
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
                ("foo", 0u16).into() => Schema::Any,
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
}
