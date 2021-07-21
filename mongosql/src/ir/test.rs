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

macro_rules! test_constant_fold {
    ($func_name:ident, $expected:expr, $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::ir::constant_folding::*;
            let input = $input;
            let expected = $expected;
            let actual = fold_constants(input);
            assert_eq!(actual, expected);
        }
    };
}

macro_rules! test_flatten_variadic_functions {
    ($func_name:ident, $expected:expr, $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::ir::flatten::*;
            let input = $input;
            let expected = $expected;
            let actual = flatten_variadic_functions(input);
            assert_eq!(actual, expected);
        }
    };
}

mod schema {
    use crate::{
        ir::{binding_tuple::DatasourceName::Bottom, schema::*, *},
        map,
        schema::*,
        set,
    };
    use lazy_static::lazy_static;

    lazy_static! {
        pub static ref TEST_DOCUMENT_A: Expression = Expression::Document(map! {
            "a".into() => Expression::Literal(Literal::Integer(1))
        });
        pub static ref TEST_DOCUMENT_SCHEMA_A: Schema = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into()},
            additional_properties: false,
        });
        pub static ref TEST_DOCUMENT_B: Expression = Expression::Document(map! {
            "b".into() => Expression::Literal(Literal::Integer(1))
        });
        pub static ref TEST_DOCUMENT_SCHEMA_B: Schema = Schema::Document(Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"b".into()},
            additional_properties: false,
        });
        pub static ref TEST_DOCUMENT_C: Expression = Expression::Document(map! {
            "c".into() => Expression::Literal(Literal::Integer(1))
        });
        pub static ref TEST_DOCUMENT_SCHEMA_C: Schema = Schema::Document(Document {
            keys: map! {
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"c".into()},
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
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![])))),
        Expression::Array(vec![]),
    );
    test_schema!(
        array_literal_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null)
        ])))),
        Expression::Array(vec![Expression::Literal(Literal::Null)]),
    );
    test_schema!(
        array_literal_two_nulls,
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
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
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
        ])))),
        Expression::Array(vec![Expression::Reference(("a", 0u16).into()),]),
        map! {("a", 0u16).into() => Schema::Missing,},
    );
    test_schema!(
        array_literal_with_nested_document_missing_preserved,
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
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
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {"b".into() =>
                    Schema::AnyOf(set![
                        Schema::AnyOf(set![
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
        Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Missing,
                Schema::Atomic(Atomic::Integer)
            ]),
            Schema::Atomic(Atomic::Double),
        ]),},
    );
    test_schema!(
        array_of_array_of_literal_any_of_any_of_missing_to_null,
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Array(Box::new(Schema::AnyOf(set![Schema::AnyOf(set![
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer)
                ]),
                Schema::Atomic(Atomic::Double)
            ])]))),
            Schema::Array(Box::new(Schema::AnyOf(set![Schema::AnyOf(set![
                Schema::AnyOf(set![
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
        Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Missing,
                Schema::Atomic(Atomic::Integer)
            ]),
            Schema::Atomic(Atomic::Double),
        ]),},
    );
    test_schema!(
        array_literal_null_or_string,
        Ok(Schema::Array(Box::new(Schema::AnyOf(set![
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
                "d".to_string() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]),
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
            ("a", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]),
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
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer)}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        map! {("bar", 0u16).into() =>
            Schema::AnyOf(set!{
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
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer), Schema::Missing}
        )),
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
        }),
        map! {("bar", 0u16).into() =>
            Schema::AnyOf(set!{
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
        Err(schema::Error::SchemaChecking {
            name: "Pos",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Reference(("bar", 0u16).into())],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        an_arg_that_may_be_nullish_manifests_as_null_in_final_schema,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Reference(("array_or_null", 0u16).into())],
        }),
        map! { ("array_or_null", 0u16).into() =>
        Schema::AnyOf(set![ANY_ARRAY.clone(), Schema::Atomic(Atomic::Null)]) },
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

    // Substring function.
    test_schema!(
        substring_requires_string_for_first_arg,
        Err(Error::SchemaChecking {
            name: "Substring",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::Integer(3))
            ],
        }),
    );
    test_schema!(
        substring_requires_integer_for_second_arg,
        Err(Error::SchemaChecking {
            name: "Substring",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::String("def".to_string())),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_schema!(
        substring_requires_integer_for_third_arg,
        Err(Error::SchemaChecking {
            name: "Substring",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::String("def".to_string()))
            ],
        }),
    );
    test_schema!(
        substring_with_start_arg,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_schema!(
        substring_with_start_and_length_args,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2))
            ],
        }),
    );
    test_schema!(
        substring_with_null_arg,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::Null),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2))
            ],
        }),
    );
    test_schema!(
        substring_with_potentially_null_arg,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1)),
                Expression::Reference(("integer_or_null", 0u16).into())
            ],
        }),
        map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)])},
    );

    // Like function type correctness
    test_schema!(
        like_first_arg_not_string_or_nullish_is_error,
        Err(Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::Like(LikeExpr {
            expr: Expression::Reference(("bar", 0u16).into()).into(),
            pattern: Expression::Literal(Literal::String("hello".into())).into(),
            escape: None,
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        like_second_arg_not_string_or_nullish_is_error,
        Err(Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::Like(LikeExpr {
            expr: Expression::Literal(Literal::String("hello".into())).into(),
            pattern: Expression::Reference(("bar", 0u16).into()).into(),
            escape: None,
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        like_must_be_string,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::Like(LikeExpr {
            expr: Expression::Reference(("bar", 0u16).into()).into(),
            pattern: Expression::Literal(Literal::String("hello".into())).into(),
            escape: None,
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        like_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::Like(LikeExpr {
            expr: Expression::Reference(("bar", 0u16).into()).into(),
            pattern: Expression::Literal(Literal::String("hello".into())).into(),
            escape: None,
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        like_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::Like(LikeExpr {
            expr: Expression::Reference(("bar", 0u16).into()).into(),
            pattern: Expression::Literal(Literal::String("hello".into())).into(),
            escape: None,
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        like_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::Like(LikeExpr {
            expr: Expression::Reference(("bar", 0u16).into()).into(),
            pattern: Expression::Literal(Literal::String("hello".into())).into(),
            escape: None,
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // And tests.
    test_schema!(
        and_first_arg_is_not_bool_is_error,
        Err(Error::SchemaChecking {
            name: "And",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        and_second_arg_is_not_bool_is_error,
        Err(Error::SchemaChecking {
            name: "And",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Literal(Literal::Boolean(true)),
                Expression::Reference(("bar", 0u16).into()),
            ],
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        and_must_be_bool,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );
    test_schema!(
        and_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        and_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );
    test_schema!(
        and_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // Or tests.
    test_schema!(
        or_first_arg_is_not_bool_is_error,
        Err(Error::SchemaChecking {
            name: "Or",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        or_second_arg_is_not_bool_is_error,
        Err(Error::SchemaChecking {
            name: "Or",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Literal(Literal::Boolean(true)),
                Expression::Reference(("bar", 0u16).into()),
            ],
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        or_must_be_bool,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );
    test_schema!(
        or_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        or_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );
    test_schema!(
        or_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // Not tests.
    test_schema!(
        not_arg_is_not_bool_is_error,
        Err(Error::SchemaChecking {
            name: "Not",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        not_must_be_bool,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );
    test_schema!(
        not_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        not_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );
    test_schema!(
        not_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // Trim function type correctness
    test_schema!(
        ltrim_must_be_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        ltrim_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        ltrim_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        ltrim_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        rtrim_must_be_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        rtrim_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        rtrim_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        rtrim_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
    test_schema!(
        btrim_must_be_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        btrim_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        btrim_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        btrim_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
    test_schema!(
        concat_must_be_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        concat_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        concat_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        concat_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("hello".into()))
            ],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        lower_must_be_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        lower_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        lower_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        lower_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        upper_must_be_string,
        Ok(Schema::Atomic(Atomic::String)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        upper_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        upper_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        upper_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        year_must_be_string,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        year_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        year_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        year_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        month_must_be_string,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        month_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        month_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        month_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        day_must_be_string,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        day_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        day_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        day_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        minute_must_be_string,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        minute_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        minute_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        minute_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        hour_must_be_string,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        hour_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        hour_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        hour_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        second_must_be_string,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Second,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        second_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Second,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        second_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Second,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        second_must_be_null,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Second,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
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
    test_schema!(
        arithmetic_decimal_takes_priority_in_any_of_must_be_numeric,
        Ok(Schema::Atomic(Atomic::Decimal)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {
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
        arithmetic_decimal_takes_priority_in_any_of_may_be_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {
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
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {
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
        arithmetic_decimal_takes_priority_in_any_of_may_be_missing,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {
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
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {
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

    // Comparison operators.
    test_schema!(
        comp_op_requires_exactly_two_args,
        Err(Error::IncorrectArgumentCount {
            name: "Lt",
            required: 2,
            found: 1
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lt,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        comp_op_requires_a_valid_comparison,
        Err(Error::InvalidComparison(
            "Lte",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lte,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::String("abc".to_string()))
            ],
        }),
    );
    test_schema!(
        comp_op_returns_boolean_schema_for_non_nullish_comparison,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2))
            ],
        }),
    );
    test_schema!(
        comp_op_returns_boolean_and_null_schema_for_potentially_nullish_comparison,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Gt,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Reference(("integer_or_null", 0u16).into()),
            ],
        }),
        map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        comp_op_returns_null_schema_for_nullish_comparison,
        Ok(Schema::Atomic(Atomic::Null)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Gte,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Null),
            ],
        }),
    );

    // Between function.
    test_schema!(
        between_requires_exactly_three_args,
        Err(Error::IncorrectArgumentCount {
            name: "Between",
            required: 3,
            found: 1
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        between_requires_a_valid_comparison_between_first_and_second_args,
        Err(Error::InvalidComparison(
            "Between",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(2)),
            ],
        }),
    );
    test_schema!(
        between_requires_a_valid_comparison_between_first_and_third_args,
        Err(Error::InvalidComparison(
            "Between",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::String("abc".to_string())),
            ],
        }),
    );
    test_schema!(
        between_returns_boolean_schema_for_non_nullish_comparisons,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Boolean)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::Long(3))
            ],
        }),
    );
    test_schema!(
        between_returns_boolean_and_null_schema_for_potentially_nullish_comparison,
        Ok(Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Boolean),
                Schema::Atomic(Atomic::Null)
            ]),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Boolean),
                Schema::Atomic(Atomic::Null)
            ]),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("long_or_null", 0u16).into()),
            ],
        }),
        map! {
            ("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)]),
            ("long_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Long), Schema::Atomic(Atomic::Null)])
        },
    );
    test_schema!(
        between_returns_null_schema_for_nullish_comparison,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Null),
                Expression::Literal(Literal::Null),
            ],
        }),
    );

    test_schema!(
        merge_object_args_must_be_documents,
        Err(Error::SchemaChecking {
            name: "MergeObjects",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![Expression::Literal(Literal::String("abc".to_string())),],
        }),
    );
    test_schema!(
        merge_objects_ok_to_be_one_any_document,
        Ok(ANY_DOCUMENT.clone()),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        merge_objects_not_ok_to_be_multiple_any_document,
        Err(Error::CannotMergeObjects(
            ANY_DOCUMENT.clone(),
            ANY_DOCUMENT.clone(),
            Satisfaction::May,
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        merge_object_args_must_have_disjoint_keys,
        Err(Error::CannotMergeObjects(
            Schema::Document(Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Integer) },
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Double) },
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Satisfaction::Must,
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("car", 0u16).into())
            ],
        }),
        map! {
            ("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
            ("car", 0u16).into() => Schema::Document(
            Document {
                keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
        },
    );
    test_schema!(
        merge_three_objects,
        Ok(Schema::Document(Document {
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
        })),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        }),
        map! {
            ("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"a".into()},
                additional_properties: false,
            }),
            ("baz", 0u16).into() => Schema::Document(
            Document {
                keys: map! {
                    "c".into() => Schema::Atomic(Atomic::Integer),
                    "d".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"d".into()},
                additional_properties: false,
            }),
            ("foo", 0u16).into() => Schema::Document(
            Document {
                keys: map! {
                    "e".into() => Schema::Atomic(Atomic::Integer),
                    "f".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"e".into()},
                additional_properties: false,
            }),
        },
    );
    test_schema!(
        merge_two_anyof_objects,
        Ok(Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::AnyOf(set![
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Integer),
                    ]),
                    Schema::Atomic(Atomic::Integer)
                ]),
                "b".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer),  Schema::Atomic(Atomic::Integer)]),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer),  Schema::Atomic(Atomic::Integer)]),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "a".into(), "e".into(),
            },
            additional_properties: false,
        })),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
        }),
        map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![
                Schema::Document(
                    Document {
                        keys: map! {
                           "a".into() => Schema::Atomic(Atomic::Integer),
                           "b".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"a".into()},
                    additional_properties: false,
                }),
                Schema::Document(
                    Document {
                        keys: map! {
                           "a".into() => Schema::Atomic(Atomic::Integer),
                           "b".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"a".into(), "b".into()},
                    additional_properties: false,
                }),
                Schema::Document(
                    Document {
                        keys: map! {
                           "a".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"a".into()},
                    additional_properties: false,
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
                }),
                Schema::Document(
                    Document {
                        keys: map! {
                           "e".into() => Schema::Atomic(Atomic::Integer),
                           "f".into() => Schema::Atomic(Atomic::Integer),
                        },
                    required: set! {"e".into()},
                    additional_properties: false,
                }),
            ]),
        },
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
            found: Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing]),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(Literal::String("field".to_string())),
            ],
        }),
        map! {("bar", 0u16).into() => Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing])},
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
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing]),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
        }),
        map! {("bar", 0u16).into() => ANY_DOCUMENT.clone(),
        ("baz", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
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

    // NullIf function.
    test_schema!(
        nullif_requires_two_args,
        Err(Error::IncorrectArgumentCount {
            name: "NullIf",
            required: 2,
            found: 1,
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        nullif_cannot_compare_numeric_with_non_numeric,
        Err(Error::InvalidComparison(
            "NullIf",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::String("abc".to_string()))
            ],
        }),
    );
    test_schema!(
        nullif_types_must_be_identical_if_non_numeric,
        Err(Error::InvalidComparison(
            "NullIf",
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::String)
        )),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(Literal::Boolean(true)),
                Expression::Literal(Literal::String("abc".to_string()))
            ],
        }),
    );
    test_schema!(
        nullif_args_cannot_be_potentially_comparable,
        Err(Error::InvalidComparison(
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
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
        }),
        map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)]),
            ("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)])
        },
    );
    test_schema!(
        nullif_identical_types,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::String("def".to_string()))
            ],
        }),
    );
    test_schema!(
        nullif_missing_type_upconverts_to_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Reference(("missing", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
            ],
        }),
        map! {
            ("missing", 0u16).into() => Schema::Missing,
        },
    );
    test_schema!(
        nullif_different_numerical_types_uses_first_arg_type,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Long(2))
            ],
        }),
    );
    test_schema!(
        nullif_multitype_numeric_args,
        Ok(Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ]),
            Schema::Atomic(Atomic::Null),
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
        }),
        map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)]),
            ("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Double), Schema::Atomic(Atomic::Decimal)])
        },
    );

    // Coalesce function.
    test_schema!(
        coalesce_requires_at_least_one_arg,
        Err(Error::IncorrectArgumentCount {
            name: "Coalesce",
            required: 1,
            found: 0,
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![],
        }),
    );
    test_schema!(
        coalesce_returns_all_arg_schemas_with_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
    );
    test_schema!(
        coalesce_upconverts_missing_to_null,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(Literal::Integer(1)),
                Expression::Reference(("missing", 0u16).into()),
            ],
        }),
        map! {
            ("missing", 0u16).into() => Schema::Missing,
        },
    );

    // Slice function.
    test_schema!(
        slice_requires_more_than_one_arg,
        Err(Error::IncorrectArgumentCount {
            name: "Slice",
            required: 2,
            found: 1,
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![Expression::Reference(("array", 0u16).into())],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_requires_fewer_than_four_arg,
        Err(Error::IncorrectArgumentCount {
            name: "Slice",
            required: 2,
            found: 4,
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
                Expression::Literal(Literal::Integer(3))
            ],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_two_args_requires_an_array_for_the_first_arg,
        Err(Error::SchemaChecking {
            name: "Slice",
            required: ANY_ARRAY.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1)),
            ],
        }),
    );
    test_schema!(
        slice_with_two_args_requires_an_integer_for_the_second_arg,
        Err(Error::SchemaChecking {
            name: "Slice",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::Long),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(Literal::Long(1)),
            ],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_three_args_requires_an_array_for_the_first_arg,
        Err(Error::SchemaChecking {
            name: "Slice",
            required: ANY_ARRAY.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2))
            ],
        }),
    );
    test_schema!(
        slice_with_three_args_requires_an_integer_for_the_second_arg,
        Err(Error::SchemaChecking {
            name: "Slice",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(Literal::String("abc".to_string())),
                Expression::Literal(Literal::Integer(1)),
            ],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_three_args_requires_an_integer_for_the_third_arg,
        Err(Error::SchemaChecking {
            name: "Slice",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
                Schema::Missing
            ]),
            found: Schema::Atomic(Atomic::String),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::String("abc".to_string())),
            ],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_length_arg,
        Ok(ANY_ARRAY.clone()),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
            ],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_start_and_length_arg,
        Ok(ANY_ARRAY.clone()),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(2)),
            ],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    // Size function.
    test_schema!(
        size_requires_one_arg,
        Err(Error::IncorrectArgumentCount {
            name: "Size",
            required: 1,
            found: 0,
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![],
        }),
    );
    test_schema!(
        size_requires_array_arg,
        Err(Error::SchemaChecking {
            name: "Size",
            required: Schema::AnyOf(set![
                ANY_ARRAY.clone(),
                Schema::Atomic(Atomic::Null),
                Schema::Missing,
            ]),
            found: Schema::Atomic(Atomic::Integer),
        }),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Literal(Literal::Integer(1))],
        }),
    );
    test_schema!(
        size_of_array,
        Ok(Schema::Atomic(Atomic::Integer)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Reference(("array", 0u16).into())],
        }),
        map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    // Datasource tests.
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
                ("foo", 0u16).into() => Schema::AnyOf(set![])
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
                ("foo", 0u16).into() => Schema::AnyOf(set![
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
            found: Schema::AnyOf(set![
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
                ("foo", 0u16).into() => Schema::AnyOf(set![
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
                ("foo", 0u16).into() => Schema::AnyOf(set![
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

    // Project.
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
            set,
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
                required: Schema::AnyOf(set![
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
                required: Schema::AnyOf(set![
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
        Ok(Schema::AnyOf(set![
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
        Ok(Schema::AnyOf(set![
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
        Ok(Schema::AnyOf(set![
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
        map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );
    test_schema!(
        cast_multi_type_expr_to_impossible_type,
        Ok(Schema::AnyOf(set![
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
        map! {("bar", 0u16).into() => Schema::AnyOf(set![
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
        map! {("bar", 0u16).into() => Schema::AnyOf(set![
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

    // Searched Case
    test_schema!(
        searched_case_when_branch_condition_must_be_boolean_or_nullish,
        Err(Error::SchemaChecking {
            name: "SearchedCase",
            required: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Boolean),
                Schema::Atomic(Atomic::Null),
                Schema::Missing,
            ]),
            found: Schema::Atomic(Atomic::Integer),
        }),
        Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(Literal::Integer(1))),
                then: Box::new(Expression::Literal(Literal::Integer(2))),
            }],
            else_branch: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        searched_case_with_no_when_branch_uses_else_branch,
        Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Long)])),
        Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![],
            else_branch: Box::new(Expression::Literal(Literal::Long(1))),
        }),
    );
    test_schema!(
        searched_case_multiple_when_branches,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(true))),
                    then: Box::new(Expression::Literal(Literal::Integer(1))),
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(true))),
                    then: Box::new(Expression::Literal(Literal::Long(2))),
                }
            ],
            else_branch: Box::new(Expression::Literal(Literal::Null)),
        }),
    );

    // Simple Case
    test_schema!(
        simple_case_when_branch_operand_must_be_comparable_with_case_operand,
        Err(Error::InvalidComparison(
            "SimpleCase",
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Integer),
        )),
        Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(Literal::String("abc".to_string()))),
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(Literal::Integer(1))),
                then: Box::new(Expression::Literal(Literal::Integer(2))),
            }],
            else_branch: Box::new(Expression::Literal(Literal::Null)),
        }),
    );
    test_schema!(
        simple_case_with_no_when_branch_uses_else_branch,
        Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Long)])),
        Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            when_branch: vec![],
            else_branch: Box::new(Expression::Literal(Literal::Long(2))),
        }),
    );
    test_schema!(
        simple_case_multiple_when_branches,
        Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null)
        ])),
        Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Integer(2))),
                    then: Box::new(Expression::Literal(Literal::Integer(3))),
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Long(4))),
                    then: Box::new(Expression::Literal(Literal::Long(5))),
                }
            ],
            else_branch: Box::new(Expression::Literal(Literal::Null)),
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
        match Ok(ResultSet {
            min_size: 2,
            max_size: Some(2),
            ..
        }),
        Stage::Limit(Limit {
            limit: 2,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        limit_gt_num_docs,
        match Ok(ResultSet {
            min_size: 3,
            max_size: Some(3),
            ..
        }),
        Stage::Limit(Limit {
            limit: 10,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(3))
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
        match Ok(ResultSet {
            min_size: 2,
            max_size: Some(2),
            ..
        }),
        Stage::Offset(Offset {
            offset: 1,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        offset_gt_num_docs,
        match Ok(ResultSet {
            min_size: 0,
            max_size: Some(0),
            ..
        }),
        Stage::Offset(Offset {
            offset: 10,
            source: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(3))
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
        Ok(Schema::AnyOf(set![
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
        Ok(Schema::AnyOf(set![Schema::AnyOf(set![]), Schema::Missing])),
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
        Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)])),
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

    // Set tests
    test_schema!(
        set_unionall_same_name_unioned,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
                    ]
                ),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        set_unionall_distinct_name_not_unioned,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
                ("bar", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "bar".into(),
            })),
        }),
    );
    test_schema!(
        set_union_same_name_unioned,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(2),
        }),
        Stage::Set(Set {
            operation: SetOperation::Union,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "foo".into(),
            })),
        }),
    );
    test_schema!(
        set_union_distinct_name_not_unioned,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone(), TEST_DOCUMENT_SCHEMA_A.clone()]),
                ("bar", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone(), TEST_DOCUMENT_SCHEMA_C.clone()]),
            },
            min_size: 1,
            max_size: Some(4),
        }),
        Stage::Set(Set {
            operation: SetOperation::Union,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone(), TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone(), TEST_DOCUMENT_C.clone(),],
                alias: "bar".into(),
            })),
        }),
    );
    test_schema!(
        set_union_of_two_empty_sets_has_min_and_max_size_0,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() =>
                        Schema::AnyOf(set![]),
                ("bar", 0u16).into() =>
                        Schema::AnyOf(set![]),
            },
            min_size: 0,
            max_size: Some(0),
        }),
        Stage::Set(Set {
            operation: SetOperation::Union,
            left: Box::new(Stage::Array(Array {
                array: vec![],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![],
                alias: "bar".into(),
            })),
        }),
    );

    // Join tests
    test_schema!(
        left_join,
        Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::AnyOf(set![
                                TEST_DOCUMENT_SCHEMA_B.clone()
                            ]
                        ),
                    ]
                ),
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "bar".into(),
            })),
            condition: Some(Expression::Literal(Literal::Boolean(false)))
        }),
    );
    test_schema!(
        cross_join,
        match Ok(ResultSet {
            min_size: 6,
            max_size: Some(6),
            ..
        }),
        Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![Expression::Document(map! {
                        "b".into() => Expression::Literal(Literal::Integer(5))
                    }),
                    Expression::Document(map! {
                        "b".into() => Expression::Literal(Literal::Integer(6))
                    }),
                ],
                alias: "bar".into(),
            })),
            condition: None
        }),
    );
    test_schema!(
        inner_join,
        match Ok(ResultSet {
            min_size: 0,
            max_size: Some(6),
            ..
        }),
        Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(Array {
                array: vec![
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(1))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(2))
                    }),
                    Expression::Document(map! {
                        "a".into() => Expression::Literal(Literal::Integer(3))
                    })
                ],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![Expression::Document(map! {
                        "b".into() => Expression::Literal(Literal::Integer(5))
                    }),
                    Expression::Document(map! {
                        "b".into() => Expression::Literal(Literal::Integer(6))
                    }),
                ],
                alias: "bar".into(),
            })),
            condition: Some(Expression::Literal(Literal::Boolean(false)))
        }),
    );
    test_schema!(
        inner_and_left_join,
        Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone(),
                    ]
                ),
                ("bar", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_B.clone(),
                    ]
                ),
                ("car", 0u16).into() => Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_C.clone()]),
                    ]
                ),
            },
            min_size: 1,
            max_size: Some(1),
        }),
        Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Join(Join {
                join_type: JoinType::Left,
                left: Box::new(Stage::Array(Array {
                    array: vec![TEST_DOCUMENT_B.clone()],
                    alias: "bar".into(),
                })),
                right: Box::new(Stage::Array(Array {
                    array: vec![TEST_DOCUMENT_C.clone()],
                    alias: "car".into(),
                })),
                condition: Some(Expression::Literal(Literal::Boolean(false)))
            })),
            condition: None
        }),
    );
    test_schema!(
        join_duplicate_datasource_names,
        Err(Error::DuplicateKeyError(("foo", 0u16).into())),
        Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "foo".into(),
            })),
            condition: None
        }),
    );
    test_schema!(
        invalid_join_condition,
        Err(Error::SchemaChecking {
            name: "join condition",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "bar".into(),
            })),
            condition: Some(Expression::Literal(Literal::Integer(5)))
        }),
    );
    test_schema!(
        join_condition_uses_left_datasource,
        match Ok(ResultSet {..}),
        Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(Array {
                array: vec![Expression::Document(map! {
                    "a".into() => Expression::Literal(Literal::Boolean(true))
                })],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_B.clone()],
                alias: "bar".into(),
            })),
            condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".to_string(),
                })),
                target_type: Type::Boolean
            }))
        }),
    );
    test_schema!(
        join_condition_uses_right_datasource,
        match Ok(ResultSet {..}),
        Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(Array {
                array: vec![TEST_DOCUMENT_A.clone()],
                alias: "foo".into(),
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![Expression::Document(map! {
                    "b".into() => Expression::Literal(Literal::Boolean(true))
                })],
                alias: "bar".into(),
            })),
            condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                    field: "b".to_string(),
                })),
                target_type: Type::Boolean
            }))
        }),
    );
    test_schema!(
        join_condition_uses_correlated_datasource,
        Ok(Schema::Atomic(Atomic::Boolean)),
        Expression::SubqueryExpression(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Left,
                    left: Box::new(Stage::Array(Array {
                        array: vec![TEST_DOCUMENT_B.clone()],
                        alias: "bar".into(),
                    })),
                    right: Box::new(Stage::Array(Array {
                        array: vec![TEST_DOCUMENT_C.clone()],
                        alias: "car".into(),
                    })),
                    condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                        expr: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".to_string(),
                        })),
                        target_type: Type::Boolean
                    }))
                }),),
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
                    "a".into() => Schema::Atomic(Atomic::Boolean)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
    );

    mod sort {
        use crate::{
            ir::{schema::*, *},
            map,
            schema::*,
            set,
        };
        test_schema!(
            comparable_schemas,
            Ok(ResultSet {
                schema_env: map! {
                    ("bar", 0u16).into() => ANY_DOCUMENT.clone(),
                },
                min_size: 0,
                max_size: None,
            }),
            Stage::Sort(Sort {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into()
                })),
                specs: vec![
                    SortSpecification::Asc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                    }))),
                    SortSpecification::Desc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "b".into(),
                    })))
                ]
            }),
            map! {
                ("foo", 0u16).into() => Schema::Document( Document{
                    keys: map! {
                        "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)]),
                        "b".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "a".into(), "b".into() },
                    additional_properties: false,
                }),
            },
        );
        test_schema!(
            incomparable_schemas,
            Err(Error::SortKeyNotSelfComparable(
                1,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String)
                ]),
            )),
            Stage::Sort(Sort {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into()
                })),
                specs: vec![
                    SortSpecification::Asc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                    }))),
                    SortSpecification::Asc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "b".into(),
                    })))
                ]
            }),
            map! {
                ("foo", 0u16).into() => Schema::Document( Document{
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)]),
                    },
                    required: set! {"a".into(), "b".into()},
                    additional_properties: false,
                }),
            },
        );
        test_schema!(
            mix_comparable_and_incomparable_schemas,
            Err(Error::SortKeyNotSelfComparable(
                0,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Null)
                ]),
            )),
            Stage::Sort(Sort {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into()
                })),
                specs: vec![SortSpecification::Asc(Box::new(Expression::FieldAccess(
                    FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                    }
                ))),]
            }),
            map! {
                ("foo", 0u16).into() => Schema::Document( Document{
                    keys: map! {
                        "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)]),
                    },
                    required: set! {"a".into(), "b".into(), "c".into()},
                    additional_properties: false,
                }),
            },
        );
    }
}

mod constant_folding {
    use crate::ir::definitions::*;
    lazy_static::lazy_static! {
        static ref TEST_SOURCE: Stage = Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into()
        });
    }
    test_constant_fold!(
        literal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Integer(1)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Integer(1)),
        }),
    );
    test_constant_fold!(
        or_simple,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Boolean(false))
                ]
            }),
        }),
    );
    test_constant_fold!(
        and_simple,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Boolean(true))
                ]
            }),
        }),
    );
    test_constant_fold!(
        true_and_nulls_is_null,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Null),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                ]
            }),
        }),
    );
    test_constant_fold!(
        null_and_null_is_null,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Null),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                ]
            }),
        }),
    );
    test_constant_fold!(
        true_and_nulls_and_ref_is_null_and_ref,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(Literal::Null),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
    );
    test_constant_fold!(
        false_or_nulls_or_ref_is_null_or_ref,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(Literal::Null),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
    );
    test_constant_fold!(
        null_or_null_is_null,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Null),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                ]
            }),
        }),
    );
    test_constant_fold!(
        false_or_nulls_is_null,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Null),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Boolean(false)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        or_with_true_literal_is_true,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
    );
    test_constant_fold!(
        or_empty,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![]
            }),
        }),
    );
    test_constant_fold!(
        and_with_false_literal_is_false,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(Literal::Boolean(true)),
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Null),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
    );
    test_constant_fold!(
        and_empty,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![]
            }),
        }),
    );
    test_constant_fold!(
        add_simple,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Integer(3)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(1)),
                ]
            })
        }),
    );
    test_constant_fold!(
        add_empty,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Integer(0)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![]
            })
        }),
    );
    test_constant_fold!(
        add_to_zero_is_zero,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Long(0)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(-1)),
                    Expression::Literal(Literal::Long(1)),
                    Expression::Literal(Literal::Long(-1)),
                ]
            })
        }),
    );
    test_constant_fold!(
        add_constant_ref_is_constant_ref,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Double(3.0)),
                    Expression::Reference(("a", 0u16).into()),
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Double(1.0)),
                    Expression::Literal(Literal::Double(1.0)),
                    Expression::Literal(Literal::Double(1.0)),
                    Expression::Reference(("a", 0u16).into()),
                ]
            })
        }),
    );
    test_constant_fold!(
        add_zero_ref_is_ref,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Reference(("a", 0u16).into())
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(-1)),
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Reference(("a", 0u16).into()),
                ]
            })
        }),
    );
    test_constant_fold!(
        mul_simple,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Long(8)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(Literal::Long(2)),
                    Expression::Literal(Literal::Long(2)),
                    Expression::Literal(Literal::Long(2)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        mul_empty,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Integer(1)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![]
            })
        }),
    );
    test_constant_fold!(
        mul_to_one_is_one,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Double(1.0)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Double(0.5)),
                    Expression::Literal(Literal::Double(2.0)),
                ]
            })
        }),
    );
    test_constant_fold!(
        mul_one_ref_is_ref,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Reference(("a", 0u16).into())
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(Literal::Double(2.0)),
                    Expression::Literal(Literal::Double(0.5)),
                    Expression::Reference(("a", 0u16).into()),
                ]
            })
        }),
    );
    test_constant_fold!(
        arithmetic_null_arg,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Null),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(Literal::Integer(2)),
                            Expression::Literal(Literal::Long(2)),
                            Expression::Literal(Literal::Double(2.0)),
                            Expression::Literal(Literal::Null),
                        ]
                    }),
                    Expression::Reference(("a", 0u16).into()),
                ]
            })
        }),
    );
    test_constant_fold!(
        add_different_num_types,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(2)),
                    Expression::Literal(Literal::Long(4)),
                    Expression::Literal(Literal::Double(6.0))
                ],
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Long(2)),
                    Expression::Literal(Literal::Long(2)),
                    Expression::Literal(Literal::Double(3.0)),
                    Expression::Literal(Literal::Double(3.0))
                ],
            }),
        }),
    );
    test_constant_fold!(
        mul_different_num_types,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(Literal::Integer(4)),
                    Expression::Literal(Literal::Long(9)),
                    Expression::Literal(Literal::Double(16.0))
                ],
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(Literal::Integer(2)),
                    Expression::Literal(Literal::Integer(2)),
                    Expression::Literal(Literal::Long(3)),
                    Expression::Literal(Literal::Long(3)),
                    Expression::Literal(Literal::Double(4.0)),
                    Expression::Literal(Literal::Double(4.0))
                ],
            }),
        }),
    );
    test_constant_fold!(
        is_less_than,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Boolean(true)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_less_than,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    Expression::Literal(Literal::Boolean(false)),
                    Expression::Literal(Literal::Boolean(false)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_less_than_equal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(Literal::Integer(0)),
                    Expression::Literal(Literal::Integer(0)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_less_than_equal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(0)),
                ]
            })
        }),
    );
    test_constant_fold!(
        is_greater_than,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gt,
                args: vec![
                    Expression::Literal(Literal::Long(1)),
                    Expression::Literal(Literal::Long(0)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_greater_than,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gt,
                args: vec![
                    Expression::Literal(Literal::Long(0)),
                    Expression::Literal(Literal::Long(0)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_greater_than_equal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gte,
                args: vec![
                    Expression::Literal(Literal::Double(1.0)),
                    Expression::Literal(Literal::Double(1.0)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_greater_than_equal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gte,
                args: vec![
                    Expression::Literal(Literal::Double(0.0)),
                    Expression::Literal(Literal::Double(1.0)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_equal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(Literal::Long(1)),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_equal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(Literal::Long(0)),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_nequal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(Literal::Long(0)),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_nequal,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(Literal::Long(1)),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        compare_different_datatypes,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        compare_null_is_null,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Null),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(Literal::Null),
                    Expression::Literal(Literal::Long(1)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_between,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(0)),
                    Expression::Literal(Literal::Integer(2)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        is_not_between,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(false)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(2)),
                    Expression::Reference(("foo", 1u16).into())
                ]
            }),
        }),
    );
    test_constant_fold!(
        fold_comparison_nested,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::Literal(Literal::Boolean(true)),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(2)),
                        ]
                    }),
                    Expression::Literal(Literal::Boolean(true)),
                ]
            }),
        }),
    );
    test_constant_fold!(
        fold_between_args,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Reference(("foo", 1u16).into()),
                    Expression::Literal(Literal::Integer(0)),
                    Expression::Literal(Literal::Integer(3)),
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Reference(("foo", 1u16).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(-1)),
                        ]
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(2)),
                        ]
                    }),
                ]
            }),
        }),
    );
}

mod flatten_node {
    use crate::ir::*;
    lazy_static::lazy_static! {
        static ref TEST_SOURCE: Stage = Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into()
        });
    }
    test_flatten_variadic_functions!(
        flatten_simple,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(2))
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(2))
                        ]
                    })
                ]
            }),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_ignores_different_funcs,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(2))
                        ]
                    })
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(2))
                        ]
                    })
                ]
            }),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_nested_multiple_levels,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::Literal(Literal::Integer(1)),
                    Expression::Literal(Literal::Integer(2)),
                    Expression::Literal(Literal::Integer(4))
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Add,
                            args: vec![
                                Expression::Literal(Literal::Integer(1)),
                                Expression::Literal(Literal::Integer(2)),
                                Expression::Literal(Literal::Integer(4))
                            ]
                        })]
                    })
                ]
            }),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_multiple_funcs,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(Literal::Integer(3)),
                    Expression::Literal(Literal::Integer(4)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(Literal::Integer(2)),
                            Expression::Literal(Literal::Integer(1)),
                            Expression::Literal(Literal::Integer(3)),
                            Expression::Literal(Literal::Integer(1))
                        ]
                    }),
                    Expression::Literal(Literal::Integer(1))
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Add,
                            args: vec![
                                Expression::Literal(Literal::Integer(3)),
                                Expression::Literal(Literal::Integer(4)),
                                Expression::ScalarFunction(ScalarFunctionApplication {
                                    function: ScalarFunction::Mul,
                                    args: vec![
                                        Expression::ScalarFunction(ScalarFunctionApplication {
                                            function: ScalarFunction::Mul,
                                            args: vec![
                                                Expression::Literal(Literal::Integer(2)),
                                                Expression::Literal(Literal::Integer(1))
                                            ]
                                        }),
                                        Expression::ScalarFunction(ScalarFunctionApplication {
                                            function: ScalarFunction::Mul,
                                            args: vec![
                                                Expression::Literal(Literal::Integer(3)),
                                                Expression::Literal(Literal::Integer(1))
                                            ]
                                        })
                                    ]
                                })
                            ]
                        })]
                    },),
                    Expression::Literal(Literal::Integer(1))
                ]
            }),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_not_necessary,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(Literal::Integer(5)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(Literal::Integer(2)),
                            Expression::Literal(Literal::Integer(1))
                        ]
                    })
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(Literal::Integer(5)),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(Literal::Integer(2)),
                            Expression::Literal(Literal::Integer(1))
                        ]
                    })
                ]
            }),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_order_matters,
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(Literal::String("foo".to_string())),
                    Expression::Literal(Literal::String("bar".to_string())),
                    Expression::Literal(Literal::String("baz".to_string()))
                ]
            }),
        }),
        Stage::Filter(Filter {
            source: Box::new(TEST_SOURCE.clone()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(Literal::String("foo".to_string())),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Concat,
                        args: vec![
                            Expression::Literal(Literal::String("bar".to_string())),
                            Expression::Literal(Literal::String("baz".to_string()))
                        ]
                    })
                ]
            }),
        }),
    );
}
