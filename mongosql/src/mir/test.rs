#![allow(clippy::redundant_pattern_matching)]
macro_rules! test_schema {
    ($func_name:ident, $(expected_error_code = $expected_error_code:literal,)? $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? input = $input:expr, $(schema_env = $schema_env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{mir::schema::SchemaInferenceState, schema::SchemaEnvironment, SchemaCheckingMode, catalog::Catalog};

            let input = $input;

            #[allow(unused_mut, unused_assignments)]
            let mut schema_env = SchemaEnvironment::default();
            $(schema_env = $schema_env;)?

            #[allow(unused_mut, unused_assignments)]
            let mut catalog = Catalog::default();
            $(catalog = $catalog;)?

            #[allow(unused_mut, unused_assignments)]
            let mut schema_checking_mode = SchemaCheckingMode::Strict;
            $(schema_checking_mode = $schema_checking_mode;)?

            let state = SchemaInferenceState::new(0u16, schema_env, &catalog, schema_checking_mode);
            let actual = input.schema(&state);

            $(assert!(matches!(actual, $expected_pat));)?
            $(assert_eq!($expected, actual);)?

            #[allow(unused_variables)]
             if let Err(e) = actual {
                 $(assert_eq!($expected_error_code, e.code()))?
             }
        }
    };
}

macro_rules! test_retain {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::mir::schema::retain;
            let expected = $expected;
            let actual = retain(&$input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_max_numeric {
    ($func_name:ident, expected = $expected:expr, input1 = $input1:expr, input2 = $input2:expr) => {
        #[test]
        fn $func_name() {
            use crate::mir::schema::max_numeric;
            let expected = $expected;
            let actual = max_numeric(&$input1, &$input2);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_user_error_messages {
    ($func_name:ident, input = $input:expr, expected = $expected:expr) => {
        #[test]
        fn $func_name() {
            use crate::{mir::schema::Error, usererror::UserError};

            let user_message = $input.user_message();

            if let Some(message) = user_message {
                assert_eq!($expected, message)
            }
        }
    };
}

mod schema {
    use crate::{
        catalog::*,
        map,
        mir::{
            binding_tuple::DatasourceName::Bottom,
            schema::Error as mir_error,
            schema::{CachedSchema, SchemaCache},
            *,
        },
        schema::*,
        set, unchecked_unique_linked_hash_map,
        usererror::UserError,
        util::{mir_collection, mir_field_access},
    };
    use lazy_static::lazy_static;

    fn test_document_a() -> Expression {
        Expression::Document(
            unchecked_unique_linked_hash_map! {
                "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
            }
            .into(),
        )
    }

    fn test_document_b() -> Expression {
        Expression::Document(
            unchecked_unique_linked_hash_map! {
                "b".into() => Expression::Literal(LiteralValue::Integer(1).into())
            }
            .into(),
        )
    }

    fn test_document_c() -> Expression {
        Expression::Document(
            unchecked_unique_linked_hash_map! {
                "c".into() => Expression::Literal(LiteralValue::Integer(1).into())
            }
            .into(),
        )
    }

    lazy_static! {
        pub static ref TEST_DOCUMENT_SCHEMA_A: Schema = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into()},
            additional_properties: false,
        });
        pub static ref TEST_DOCUMENT_SCHEMA_B: Schema = Schema::Document(Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"b".into()},
            additional_properties: false,
        });
        pub static ref TEST_DOCUMENT_SCHEMA_C: Schema = Schema::Document(Document {
            keys: map! {
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"c".into()},
            additional_properties: false,
        });
        pub static ref TEST_DOCUMENT_SCHEMA_S: Schema = Schema::Document(Document {
            keys: map! {
                "s".into() => Schema::Atomic(Atomic::String),
            },
            required: set! {"s".into()},
            additional_properties: false,
        });
    }

    test_schema!(
        literal_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::Literal(LiteralValue::Null.into()),
    );
    test_schema!(
        literal_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Literal(LiteralValue::Boolean(true).into()),
    );
    test_schema!(
        literal_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::Literal(LiteralValue::String("foobar".to_string()).into()),
    );
    test_schema!(
        literal_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::Literal(LiteralValue::Integer(5).into()),
    );
    test_schema!(
        literal_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = Expression::Literal(LiteralValue::Long(6).into()),
    );
    test_schema!(
        literal_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::Literal(LiteralValue::Double(7.0).into()),
    );
    test_schema!(
        reference_does_not_exist_in_schema_env,
        expected_error_code = 1000,
        expected = Err(mir_error::DatasourceNotFoundInSchemaEnv(("a", 0u16).into())),
        input = Expression::Reference(("a", 0u16).into()),
    );
    test_schema!(
        reference_exists_in_schema_env,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::Reference(("a", 0u16).into()),
        schema_env = map! {("a", 0u16).into() => Schema::Atomic(Atomic::Null),},
    );

    // Array Literals
    test_schema!(
        array_literal_empty,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![])))),
        input = Expression::Array(vec![].into()),
    );
    test_schema!(
        array_literal_null,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null)
        ])))),
        input = Expression::Array(vec![Expression::Literal(LiteralValue::Null.into())].into()),
    );
    test_schema!(
        array_literal_two_nulls,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])))),
        input = Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Literal(LiteralValue::Null.into())
            ]
            .into()
        ),
    );
    test_schema!(
        array_literal_missing_to_null,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
        ])))),
        input = Expression::Array(vec![Expression::Reference(("a", 0u16).into()),].into()),
        schema_env = map! {("a", 0u16).into() => Schema::Missing,},
    );
    test_schema!(
        array_literal_with_nested_document_missing_preserved,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                "bar".into() => Schema::Atomic(Atomic::String),
                    },
                required: set! {"bar".into()},
                additional_properties: false,
            })
        ])))),
        input = Expression::Array(
            vec![Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "foo".into() => Expression::Reference(("a", 0u16).into()),
                    "bar".into() => Expression::Reference(("b", 0u16).into()),
                }
                .into()
            ),]
            .into()
        ),
        schema_env = map! {
            ("a", 0u16).into() => Schema::Missing,
            ("b", 0u16).into() => Schema::Atomic(Atomic::String),
        },
    );
    test_schema!(
        array_literal_any_of_any_of_missing_to_null,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
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
        input = Expression::Array(vec![Expression::Document(
            unchecked_unique_linked_hash_map! {"b".into() => Expression::Reference(("a", 0u16).into())}.into())].into()),
        schema_env = map! {("a", 0u16).into() =>
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
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
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
        input = Expression::Array(
            vec![
                Expression::Array(vec![Expression::Reference(("a", 0u16).into()),].into()),
                Expression::Array(vec![Expression::Reference(("a", 0u16).into()),].into()),
            ]
            .into()
        ),
        schema_env = map! {("a", 0u16).into() =>
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
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::String),
        ])))),
        input = Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Literal(LiteralValue::String("hello".to_string()).into()),
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Literal(LiteralValue::String("world".to_string()).into()),
            ]
            .into()
        ),
    );

    // Document Literal
    test_schema!(
        document_literal_empty,
        expected = Ok(Schema::Document(Document {
            keys: map! {},
            required: set! {},
            additional_properties: false,
        })),
        input = Expression::Document(unchecked_unique_linked_hash_map! {}.into()),
    );
    test_schema!(
        document_literal_all_required,
        expected = Ok(Schema::Document(Document {
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
        input = Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => Expression::Literal(LiteralValue::String("Hello".to_string()).into()),
            "b".to_string() => Expression::Literal(LiteralValue::String("World".to_string()).into()),
            "c".to_string() => Expression::Literal(LiteralValue::Null.into()),
            "d".to_string() => Expression::Literal(LiteralValue::Long(42).into()),
        }.into()),
    );
    test_schema!(
        document_literal_some_keys_may_or_must_satisfy_missing,
        expected = Ok(Schema::Document(Document {
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
        input = Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => Expression::Literal(LiteralValue::String("Hello".to_string()).into()),
            "b".to_string() => Expression::Reference(("b", 0u16).into()),
            "c".to_string() => Expression::Literal(LiteralValue::Null.into()),
            "d".to_string() => Expression::Reference(("a", 0u16).into()),
        }.into()),
        schema_env = map! {
            ("a", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]),
            ("b", 0u16).into() => Schema::Missing,
        },
    );

    // FieldAccess
    test_schema!(
        field_access_accessee_cannot_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "FieldAccess",
            required: crate::schema::ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::Long),
        }),
        input = Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Literal(LiteralValue::Long(1).into())),
            field: "foo".to_string(),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        field_access_field_must_not_exist_not_in_document,
        expected_error_code = 1007,
        expected = Err(mir_error::AccessMissingField("foo".to_string())),
        input = Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foof".to_string()},
                additional_properties: false,
            }
        ),},
    );
    test_schema!(
        field_access_field_may_exist,
        expected = Ok(Schema::Any),
        input = Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foof".to_string()},
                additional_properties: true,
            }
        ),},
    );
    test_schema!(
        field_access_field_must_exist,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Document(
            Document {
                keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foo".to_string()},
                additional_properties: false,
            }
        ),},
    );
    test_schema!(
        field_access_field_must_any_of,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer)}
        )),
        input = Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() =>
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
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer), Schema::Missing}
        )),
        input = Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("bar", 0u16).into())),
            field: "foo".to_string(),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() =>
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

    // OptimizedMatchExists
    test_schema!(
        optimized_match_exists_is_always_boolean,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::OptimizedMatchExists(OptimizedMatchExists {
            field_access: FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "x".to_string(),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new()
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "x".to_string() => Schema::AnyOf(set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null), Schema::Missing}),
                },
                required: set! {},
                additional_properties: false,
            })
        },
    );

    // General function schema checking.
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Reference(("bar", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        an_arg_that_may_be_nullish_manifests_as_null_in_final_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Reference(("array_or_null", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        schema_env = map! { ("array_or_null", 0u16).into() =>
        Schema::AnyOf(set![ANY_ARRAY.clone(), Schema::Atomic(Atomic::Null)]) },
    );

    // Unary functions.
    test_schema!(
        unary_pos,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        unary_neg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![Expression::Literal(LiteralValue::Double(1.0).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        unary_pos_requires_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Pos",
            required: 1,
            found: 0
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pos,
            args: vec![],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        unary_neg_requires_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Neg",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Neg,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Substring function.
    test_schema!(
        substring_requires_string_for_first_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Substring",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        substring_requires_integer_for_second_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Substring",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::String("def".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        substring_requires_integer_for_third_arg,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Substring",
            required: INTEGER_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::String("def".to_string()).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        substring_with_start_arg,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        substring_with_start_and_length_args,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        substring_with_null_arg,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        substring_with_potentially_null_arg,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Reference(("integer_or_null", 0u16).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)])},
    );

    // Like function type correctness
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

    // And tests.
    test_schema!(
        and_first_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "And",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Literal(LiteralValue::Boolean(true).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        and_must_be_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );
    test_schema!(
        and_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        and_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );
    test_schema!(
        and_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // Or tests.
    test_schema!(
        or_first_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Or",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Literal(LiteralValue::Boolean(true).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        or_must_be_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );
    test_schema!(
        or_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        or_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );
    test_schema!(
        or_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Or,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // Not tests.
    test_schema!(
        not_arg_is_not_bool_is_error,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Not",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: NUMERIC_OR_NULLISH.clone(),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => NUMERIC_OR_NULLISH.clone()},
    );
    test_schema!(
        not_must_be_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean)},
    );
    test_schema!(
        not_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        not_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Boolean), Schema::Missing])},
    );
    test_schema!(
        not_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Not,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    // Trim function type correctness
    test_schema!(
        ltrim_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        ltrim_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        ltrim_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        ltrim_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::LTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        rtrim_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        rtrim_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        rtrim_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        rtrim_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::RTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
    test_schema!(
        btrim_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        btrim_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        btrim_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        btrim_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BTrim,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );
    test_schema!(
        concat_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        concat_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        concat_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        concat_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Concat,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello".into()).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        lower_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        lower_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        lower_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        lower_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lower,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        upper_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::String)},
    );
    test_schema!(
        upper_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        upper_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        upper_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Upper,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        year_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        year_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        year_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        year_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Year,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        month_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        month_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        month_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        month_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Month,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        day_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        day_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        day_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        day_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Day,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        minute_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        minute_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        minute_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        minute_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Minute,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    test_schema!(
        hour_must_be_string,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Date)},
    );
    test_schema!(
        hour_may_be_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        hour_may_be_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Date), Schema::Missing])},
    );
    test_schema!(
        hour_must_be_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Hour,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Null)},
    );

    //AggregationFunction schema checking.
    test_schema!(
        max_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Max".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Max,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        min_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Min".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Min,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        distinct_sum_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Sum DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_count_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Count DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Count,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_first_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "First DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::First,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_last_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Last DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Last,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_stddevpop_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "StddevPop DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_stddevsamp_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "StddevSamp DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_max_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Max DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Max,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        distinct_min_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Min DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Min,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        count_star_is_int,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ])),
        input = AggregationExpr::CountStar(false),
    );
    test_schema!(
        count_expr_is_int,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Count,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );

    test_schema!(
        arg_to_sum_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Sum",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        sum_of_int_and_long_is_int_long,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
    test_schema!(
        sum_of_int_and_double_is_int_double,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );
    test_schema!(
        sum_of_int_and_decimal_is_int_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        arg_to_mergedocuments_must_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "MergeDocuments",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::MergeDocuments,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        mergedocuments_returns_field_schema,
        expected = Ok(Schema::Document(Document {
            keys: map! {"foo".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: true,
        })),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::MergeDocuments,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Document(Document {
            keys: map!{"foo".into() => Schema::Atomic(Atomic::Integer)},
            required: set!{},
            additional_properties: true,
        })},
    );

    test_schema!(
        arg_to_avg_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Avg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        second_may_be_null,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
    test_schema!(
        avg_of_int_and_decimal_is_double_and_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
    test_schema!(
        avg_of_decimal_and_missing_is_decimal_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Missing,
        ])},
    );
    test_schema!(
        avg_of_long_and_null_is_double_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null),
        ])},
    );
    test_schema!(
        avg_of_decimal_long_and_null_is_decimal_double_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double)
            ],),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null),
        ])},
    );
    test_schema!(
        avg_of_integer_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
        ])},
    );
    test_schema!(
        avg_of_decimal_is_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        arg_to_stddev_pop_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "StddevPop",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        stddev_pop_of_integer_and_long_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double),),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
    test_schema!(
        stddev_pop_of_integer_and_decimal_is_double_and_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
    test_schema!(
        stddev_pop_of_decimal_is_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        arg_to_stddev_samp_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "StddevSamp",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );
    test_schema!(
        stddev_samp_of_integer_and_long_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double),),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
    test_schema!(
        stddev_samp_of_integer_and_decimal_is_double_and_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
    test_schema!(
        stddev_samp_of_decimal_is_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        add_to_array_has_array_schema,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])))),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::AddToArray,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        add_to_set_has_array_schema,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])))),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::AddToArray,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        first_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::First,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        last_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Last,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        min_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Min,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        max_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Max,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    // Arithmetic function type correctness.
    test_schema!(
        variadic_arg_arithmetic_no_args_returns_integer,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        variadic_arg_arithmetic_one_arg_returns_that_type,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![Expression::Literal(LiteralValue::Double(1.0).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        arithmetic_null_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Literal(LiteralValue::Double(2.0).into()),
                Expression::Literal(LiteralValue::Long(3).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        arithmetic_missing_takes_priority_as_null_result,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Double(2.0).into()),
                Expression::Literal(LiteralValue::Long(3).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Missing},
    );
    test_schema!(
        arithmetic_decimal_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Long(2).into()),
                Expression::Literal(LiteralValue::Double(3.0).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Atomic(Atomic::Decimal)},
    );
    test_schema!(
        arithmetic_double_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(LiteralValue::Double(1.0).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Long(3).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        arithmetic_long_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Long(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        arithmetic_integer_takes_priority,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        arithmetic_results_in_any_numeric,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mul,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
            distinct: false,
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Atomic(Atomic::Double),
            ("bar", 0u16).into() =>Schema::Atomic(Atomic::String),
        },
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );

    // Abs tests
    test_schema!(
        abs_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Abs",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Abs,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Abs,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );

    // Ceil tests
    test_schema!(
        ceil_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Ceil",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Ceil,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Ceil,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );

    // Degrees tests
    test_schema!(
        degrees_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Degrees",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Degrees,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Degrees,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        degrees_returns_double_schema_for_integer_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Degrees,
            args: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        degrees_returns_double_schema_for_long_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Degrees,
            args: vec![Expression::Literal(LiteralValue::Long(2).into())],
            cache: SchemaCache::new(),
        }),
    );

    // Floor tests
    test_schema!(
        floor_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Floor",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Floor,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Floor,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );

    // Log tests
    test_schema!(
        log_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Log",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Log,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Log,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Log,
            args: vec![
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Mod tests
    test_schema!(
        mod_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Mod",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mod,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mod,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Mod,
            args: vec![
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Pow tests
    test_schema!(
        pow_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Pow",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pow,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pow,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Pow,
            args: vec![
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Round tests
    test_schema!(
        round_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Round",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Round,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Round,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Round,
            args: vec![
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Cos tests
    test_schema!(
        cos_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Cos",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Cos,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Cos,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        cos_returns_double_schema_for_integer_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Cos,
            args: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
    );

    // Sin tests
    test_schema!(
        sin_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Sin",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sin,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sin,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        sin_returns_double_schema_for_double_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sin,
            args: vec![Expression::Literal(LiteralValue::Double(2.8).into())],
            cache: SchemaCache::new(),
        }),
    );

    // Tan tests
    test_schema!(
        tan_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Tan",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Tan,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Tan,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        tan_returns_double_schema_for_long_arg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sin,
            args: vec![Expression::Literal(LiteralValue::Long(2).into())],
            cache: SchemaCache::new(),
        }),
    );

    // Radians tests
    test_schema!(
        radians_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Radians",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Radians,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Radians,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );

    // Sqrt tests
    test_schema!(
        sqrt_requires_exactly_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Sqrt",
            required: 1,
            found: 2
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sqrt,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sqrt,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
    );

    // Arithmetic function errors.
    test_schema!(
        sub_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Sub",
            required: 2,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sub,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        div_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Div",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Div,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Sub,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Div,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Boolean(true).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into()),
                Expression::Literal(LiteralValue::Boolean(true).into()),
                Expression::Literal(LiteralValue::Integer(4).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Comparison operators.
    test_schema!(
        comp_op_requires_exactly_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Lt",
            required: 2,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lt,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        comp_op_requires_a_valid_comparison,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "Lte",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lte,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        comp_op_returns_boolean_schema_for_non_nullish_comparison,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        comp_op_returns_boolean_and_null_schema_for_potentially_nullish_comparison,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Gt,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Reference(("integer_or_null", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)])},
    );
    test_schema!(
        comp_op_returns_null_schema_for_nullish_comparison,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Gte,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Null.into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Between function.
    test_schema!(
        between_requires_exactly_three_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Between",
            required: 3,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        between_requires_a_valid_comparison_between_first_and_second_args,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "Between",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        between_requires_a_valid_comparison_between_first_and_third_args,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "Between",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        between_returns_boolean_schema_for_non_nullish_comparisons,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Boolean)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Long(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("long_or_null", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Literal(LiteralValue::Null.into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        merge_object_args_must_be_documents,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "MergeObjects",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::String),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![Expression::Literal(
                LiteralValue::String("abc".to_string()).into()
            ),],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        merge_objects_ok_to_be_one_any_document,
        expected = Ok(ANY_DOCUMENT.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![Expression::Reference(("bar", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("bar", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );
    test_schema!(
        merge_object_args_must_have_disjoint_keys,
        expected_error_code = 1006,
        expected = Err(mir_error::CannotMergeObjects(
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("car", 0u16).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
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
        })),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
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
        expected = Ok(Schema::Document(Document {
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("foo", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "ComputedFieldAccess",
            required: 2,
            found: 3
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Literal(LiteralValue::Long(1).into()),
                Expression::Literal(LiteralValue::Long(2).into()),
                Expression::Literal(LiteralValue::Long(3).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        computed_field_access_first_arg_must_not_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::Long),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Literal(LiteralValue::Long(1).into()),
                Expression::Literal(LiteralValue::Long(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        computed_field_access_first_arg_may_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "ComputedFieldAccess",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![ANY_DOCUMENT.clone(), Schema::Missing]),
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("field".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::Long(42).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Reference(("baz", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone(),
        ("baz", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::String), Schema::Missing])},
    );
    test_schema!(
        computed_field_access_valid_args,
        expected = Ok(Schema::Any),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Reference(("bar", 0u16).into()),
                Expression::Literal(LiteralValue::String("field".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    // Datetime value scalar function.
    test_schema!(
        current_timestamp_no_arg,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::CurrentTimestamp,
            args: vec![],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        current_timestamp_integer_arg_should_be_removed_in_algebrization,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "CurrentTimestamp",
            required: 0,
            found: 1
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::CurrentTimestamp,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );

    // NullIf function.
    test_schema!(
        nullif_requires_two_args,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "NullIf",
            required: 2,
            found: 1,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        nullif_cannot_compare_numeric_with_non_numeric,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "NullIf",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        nullif_types_must_be_identical_if_non_numeric,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "NullIf",
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::String)
        )),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(LiteralValue::Boolean(true).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::String("def".to_string()).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        nullif_missing_type_upconverts_to_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Null),
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Reference(("missing", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Long(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::NullIf,
            args: vec![
                Expression::Reference(("foo", 0u16).into()),
                Expression::Reference(("bar", 0u16).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)]),
            ("bar", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Double), Schema::Atomic(Atomic::Decimal)])
        },
    );

    // Coalesce function.
    test_schema!(
        coalesce_requires_at_least_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Coalesce",
            required: 1,
            found: 0,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        coalesce_returns_first_non_nullish_arg_omitting_null_and_missing,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1).into()),
                Expression::Literal(LiteralValue::Double(2.0).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! { ("null_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]) },
    );
    test_schema!(
        coalesce_returns_up_to_first_non_nullish_arg_omitting_null_and_missing,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1).into()),
                Expression::Literal(LiteralValue::Double(2.0).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Reference(("any", 0u16).into()),
                Expression::Reference(("missing", 0u16).into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("any", 0u16).into() => Schema::Any,
            ("missing", 0u16).into() => Schema::Missing,
            ("null_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing])
        },
    );
    test_schema!(
        coalesce_with_any_schema_and_non_nullish_arg_yields_non_nullish_schema,
        expected = Ok(NON_NULLISH.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Reference(("any", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("any", 0u16).into() => Schema::Any},
    );
    test_schema!(
        coalesce_with_only_nullish_args_yields_only_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(LiteralValue::Null.into()),
                Expression::Reference(("missing", 0u16).into()),
                Expression::Reference(("null_or_missing", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Reference(("integer_or_null", 0u16).into()),
                Expression::Reference(("long_or_missing", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("integer_or_null", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)]),
            ("long_or_missing", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Long), Schema::Missing])
        },
    );

    // Slice function.
    test_schema!(
        slice_requires_more_than_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Slice",
            required: 2,
            found: 1,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![Expression::Reference(("array", 0u16).into())],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
                Expression::Literal(LiteralValue::Integer(3).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into())
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_length_arg,
        expected = Ok(ANY_ARRAY.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );
    test_schema!(
        slice_with_start_and_length_arg,
        expected = Ok(ANY_ARRAY.clone()),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Reference(("array", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(2).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    // Split function.
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
            args: vec![Expression::Literal(
                LiteralValue::String("a-b-c".to_string()).into()
            )],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(5).into()),
                Expression::Literal(LiteralValue::String("-".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::String("a-b-c".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(5).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::String("a-b-c".to_string()).into()),
                Expression::Literal(LiteralValue::String("-".to_string()).into()),
                Expression::Literal(LiteralValue::String("1".to_string()).into()),
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::String("a-b-c".to_string()).into()),
                Expression::Literal(LiteralValue::String("".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into()),
            ],
            cache: SchemaCache::new(),
        }),
    );

    // Size function.
    test_schema!(
        size_requires_one_arg,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Size",
            required: 1,
            found: 0,
        }),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![],
            cache: SchemaCache::new(),
        }),
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
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        size_of_array,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Reference(("array", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        schema_env = map! { ("array", 0u16).into() => ANY_ARRAY.clone() },
    );

    // Datasource tests.
    test_schema!(
        collection_schema_no_catalog,
        expected_error_code = 1016,
        expected = Err(mir_error::CollectionNotFound("test2".into(), "foo".into())),
        input = Stage::Collection(Collection {
            db: "test2".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        collection_schema_namespace_not_in_catalog,
        expected_error_code = 1016,
        expected = Err(mir_error::CollectionNotFound("foo".into(), "baz".into())),
        input = Stage::Collection(Collection {
            db: "foo".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "foo".into(), collection: "bar".into()} => Schema::Atomic(Atomic::Integer)
        }),
    );

    test_schema!(
        collection_schema_namespace_in_catalog,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Integer),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Collection(Collection {
            db: "foo".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "foo".into(), collection: "bar".into()} => Schema::Atomic(Atomic::Integer)
        }),
    );

    test_schema!(
        empty_array_datasource_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::AnyOf(set![])
            },
            min_size: 0,
            max_size: Some(0),
        }),
        input = Stage::Array(ArraySource {
            array: vec![],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        dual_array_datasource_schema,
        expected = Ok(ResultSet {
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
        input = Stage::Array(ArraySource {
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {}.into()
            )],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        literal_array_items_datasource_schema,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ])
        }),
        input = Stage::Array(ArraySource {
            array: vec![
                Expression::Literal(LiteralValue::Integer(42).into()),
                Expression::Literal(LiteralValue::Double(42f64).into())
            ],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        single_document_array_datasource_schema,
        expected = Ok(ResultSet {
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
        input = Stage::Array(ArraySource {
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "bar".into() => Expression::Literal(LiteralValue::Integer(1).into())
                }
                .into()
            )],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        two_document_array_datasource_schema,
        expected = Ok(ResultSet {
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
        input = Stage::Array(ArraySource {
            array: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {
                    "bar".into() => Expression::Literal(LiteralValue::Integer(1).into())
                    }
                    .into()
                ),
                Expression::Document(
                    unchecked_unique_linked_hash_map! {
                    "car".into() => Expression::Literal(LiteralValue::Integer(1).into())
                    }
                    .into()
                )
            ],
            alias: "foo".into(),
            cache: SchemaCache::new(),
        }),
    );

    // Project.
    test_schema!(
        project_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar1", 0u16).into() => ANY_DOCUMENT.clone(),
                ("bar2", 0u16).into() => ANY_DOCUMENT.clone(),
                ("bar3", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test2".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar1", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
                ("bar2", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
                ("bar3", 0u16).into() =>
                    Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test2".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    mod filter {
        use crate::schema::Document;
        use crate::{
            catalog::Namespace,
            map,
            mir::{schema::CachedSchema, schema::Error as mir_error, schema::SchemaCache, *},
            schema::{Atomic, ResultSet, Schema, ANY_DOCUMENT},
            set, unchecked_unique_linked_hash_map,
            usererror::UserError,
        };

        fn true_mir() -> Expression {
            Expression::Literal(LiteralValue::Boolean(true).into())
        }

        fn test_source() -> Stage {
            Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: schema::SchemaCache::new(),
            })
        }

        test_schema!(
            boolean_condition_allowed,
            expected_pat = Ok(_),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: true_mir(),
                cache: SchemaCache::new(),
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            null_condition_allowed,
            expected_pat = Ok(_),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::Literal(LiteralValue::Null.into()),
                cache: SchemaCache::new(),
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            missing_condition_allowed,
            expected_pat = Ok(_),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::Reference(("m", 0u16).into()),
                cache: SchemaCache::new(),
            }),
            schema_env = map! {("m", 0u16).into() => Schema::Missing},
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            non_boolean_condition_disallowed,
            expected_error_code = 1002,
            expected = Err(mir_error::SchemaChecking {
                name: "filter condition",
                required: Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing,
                ]),
                found: Schema::Atomic(Atomic::Integer),
            }),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::Literal(LiteralValue::Integer(123).into()),
                cache: SchemaCache::new(),
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            possible_non_boolean_condition_disallowed,
            expected_error_code = 1002,
            expected = Err(mir_error::SchemaChecking {
                name: "filter condition",
                required: Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing,
                ]),
                found: Schema::Any,
            }),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::FieldAccess(FieldAccess {
                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                    field: "bar".into(),
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            source_fails_schema_check,
            expected_pat = Err(mir_error::SchemaChecking {
                name: "array datasource items",
                ..
            }),
            input = Stage::Filter(Filter {
                source: Stage::Array(ArraySource {
                    alias: "arr".into(),
                    array: vec![Expression::Literal(LiteralValue::Null.into())],
                    cache: SchemaCache::new(),
                })
                .into(),
                condition: true_mir(),
                cache: SchemaCache::new(),
            }),
        );
        test_schema!(
            condition_fails_schema_check,
            expected_error_code = 1000,
            expected = Err(mir_error::DatasourceNotFoundInSchemaEnv(
                ("abc", 0u16).into()
            )),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::Reference(("abc", 0u16).into()),
                cache: SchemaCache::new(),
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            min_size_reduced_to_zero_max_size_preserved,
            expected_pat = Ok(ResultSet{
                min_size: 0,
                max_size: Some(1),
                ..
            }),
            input = Stage::Filter(Filter {
                condition: true_mir(),
                source: Stage::Array(ArraySource {
                    alias: "arr".into(),
                    array: vec![Expression::Document(unchecked_unique_linked_hash_map!{"a".into() => Expression::Literal(LiteralValue::Null.into()),}.into())],
                    cache: SchemaCache::new(),
                }).into(),
                cache: SchemaCache::new(),
            }),
        );

        test_schema!(
            optimized_fields_in_filter_stage_ensure_those_fields_are_non_null_in_the_result_set,
            expected = Ok(ResultSet {
                min_size: 0,
                max_size: None,
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "non_nullable".to_string() => Schema::Atomic(Atomic::Integer),
                            "nullable_a".to_string() => Schema::Atomic(Atomic::Integer),
                            "nullable_b".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set! {"non_nullable".to_string(), "nullable_a".to_string(), "nullable_b".to_string()},
                        additional_properties: false,
                    })
                },
            }),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    args: vec![
                        Expression::OptimizedMatchExists(OptimizedMatchExists {
                            field_access: FieldAccess {
                                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                field: "nullable_a".to_string(),
                                cache: SchemaCache::new(),
                            },
                            cache: SchemaCache::new(),
                        }),
                        Expression::OptimizedMatchExists(OptimizedMatchExists {
                            field_access: FieldAccess {
                                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                field: "nullable_b".to_string(),
                                cache: SchemaCache::new(),
                            },
                            cache: SchemaCache::new(),
                        }),
                    ],
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                    keys: map! {
                        "non_nullable".to_string() => Schema::Atomic(Atomic::Integer),
                        "nullable_a".to_string() => Schema::Atomic(Atomic::Integer),
                        "nullable_b".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
                    },
                    required: set! {"non_nullable".to_string()},
                    additional_properties: false,
                }),
            }),
        );
    }

    // Cast.
    test_schema!(
        cast_expr_to_same_type,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Literal(LiteralValue::Integer(1).into())),
            to: Type::Int32,
            on_null: Box::new(Expression::Literal(LiteralValue::Null.into())),
            on_error: Box::new(Expression::Literal(LiteralValue::Null.into())),
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
            cache: SchemaCache::new(),
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Missing},
    );

    // TypeAssert.
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

    // Searched Case
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
            }],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        searched_case_with_no_when_branch_uses_else_branch,
        expected = Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Long)])),
        input = Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![],
            else_branch: Box::new(Expression::Literal(LiteralValue::Long(1).into())),
            cache: SchemaCache::new(),
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
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                    then: Box::new(Expression::Literal(LiteralValue::Long(2).into())),
                }
            ],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
        }),
    );

    // Simple Case
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
            }],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
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
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Long(4).into())),
                    then: Box::new(Expression::Literal(LiteralValue::Long(5).into())),
                }
            ],
            else_branch: Box::new(Expression::Literal(LiteralValue::Null.into())),
            cache: SchemaCache::new(),
        }),
    );

    // Limit and Offset
    test_schema!(
        limit_collection_datasource,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: Some(20),
        }),
        input = Stage::Limit(Limit {
            limit: 20,
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );
    test_schema!(
        limit_lt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 2,
            max_size: Some(2),
            ..
        }),
        input = Stage::Limit(Limit {
            limit: 2,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        limit_gt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 3,
            max_size: Some(3),
            ..
        }),
        input = Stage::Limit(Limit {
            limit: 10,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        offset_collection_datasource,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Offset(Offset {
            offset: 20,
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );
    test_schema!(
        offset_lt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 2,
            max_size: Some(2),
            ..
        }),
        input = Stage::Offset(Offset {
            offset: 1,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        offset_gt_num_docs,
        expected_pat = Ok(ResultSet {
            min_size: 0,
            max_size: Some(0),
            ..
        }),
        input = Stage::Offset(Offset {
            offset: 10,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    // Exists subquery
    test_schema!(
        exists_uncorrelated,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Exists(
            Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            }))
            .into()
        ),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    test_schema!(
        exists_correlated,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".into() => Expression::FieldAccess(FieldAccess{
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    })
                }.into())
            },
            cache: SchemaCache::new(),
        })).into()),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    test_schema!(
        exists_invalid_expression,
        expected_error_code = 1001,
        expected = Err(mir_error::IncorrectArgumentCount {
            name: "Div",
            required: 2,
            found: 3
        }),
        input = Expression::Exists(
            Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("a", 0u16).into() =>
                        Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Div,
                            args: vec![
                                Expression::Literal(LiteralValue::Integer(1).into()),
                                Expression::Literal(LiteralValue::Integer(2).into()),
                                Expression::Literal(LiteralValue::Integer(3).into())
                            ],
                        cache: SchemaCache::new(),
                        })
                },
                cache: SchemaCache::new(),
            }))
            .into()
        ),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Subquery Expression

    test_schema!(
        subquery_output_expr_violates_type_constraints,
        expected_error_code = 1007,
        expected = Err(mir_error::AccessMissingField("_2".into())),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "_2".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "_1".into() => Expression::Literal(LiteralValue::Integer(5).into())
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Analogous SQL query: SELECT (SELECT foo.a FROM []) FROM foo
    test_schema!(
        correlated_subquery,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::String),
            Schema::Missing
        ])),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![],
                    alias: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
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
        expected_error_code = 1000,
        expected = Err(mir_error::DatasourceNotFoundInSchemaEnv(
            ("foo", 0u16).into()
        )),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
            }),
        },
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [] AS foo)"
    test_schema!(
        uncorrelated_subquery_cardinality_is_zero,
        expected = Ok(Schema::AnyOf(set![Schema::AnyOf(set![]), Schema::Missing])),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_must_be_one,
        expected = Ok(Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)])),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(5).into())
                    }
                    .into()
                ),],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM foo)"
    test_schema!(
        subquery_cardinality_may_be_1,
        expected_error_code = 1008,
        expected = Err(mir_error::InvalidSubqueryCardinality),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}, {'a': 6}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_gt_one,
        expected_error_code = 1008,
        expected = Err(mir_error::InvalidSubqueryCardinality),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(6).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    // Subquery Comparisons
    test_schema!(
        uncorrelated_subquery_comparison_known_type,
        expected = Ok(Schema::Atomic(Atomic::Boolean).clone()),
        input = Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5).into())),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                    field: "a".into(),
                    cache: SchemaCache::new(),
                })),
                subquery: Box::new(Stage::Array(ArraySource {
                    array: vec![Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    )],
                    alias: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
        }),
    );

    test_schema!(
        uncorrelated_subquery_comparison_possibly_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null)
        ])),
        input =
            Expression::SubqueryComparison(SubqueryComparison {
                operator: SubqueryComparisonOp::Eq,
                modifier: SubqueryModifier::All,
                argument: Box::new(Expression::Literal(LiteralValue::Integer(5).into())),
                subquery_expr: SubqueryExpr {
                    output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    })),
                    subquery: Box::new(Stage::Array(ArraySource {
                        array: vec![
                        Expression::Document(unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }.into()),
                        Expression::Document(unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }.into())
                    ],
                        alias: "foo".into(),
                        cache: SchemaCache::new(),
                    })),
                    cache: SchemaCache::new(),
                },
                cache: SchemaCache::new(),
            }),
    );

    test_schema!(
        incomparable_argument_and_output_expr,
        expected_error_code = 1005,
        expected = Err(mir_error::InvalidComparison(
            "subquery comparison",
            Schema::Atomic(Atomic::String),
            Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
        )),
        input = Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(
                LiteralValue::String("abc".into()).into()
            )),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                    field: "a".into(),
                    cache: SchemaCache::new(),
                })),
                subquery: Box::new(Stage::Array(ArraySource {
                    array: vec![Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    )],
                    alias: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
        }),
    );

    // Set tests
    test_schema!(
        set_unionall_same_name_unioned,
        expected = Ok(ResultSet {
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
        input = Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        set_unionall_distinct_name_not_unioned,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone()]),
                ("bar", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone()]),
            },
            min_size: 2,
            max_size: Some(2),
        }),
        input = Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        set_union_same_name_unioned,
        expected = Ok(ResultSet {
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
        input = Stage::Set(Set {
            operation: SetOperation::Union,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        set_union_distinct_name_not_unioned,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_A.clone(), TEST_DOCUMENT_SCHEMA_A.clone()]),
                ("bar", 0u16).into() =>
                        Schema::AnyOf(set![TEST_DOCUMENT_SCHEMA_B.clone(), TEST_DOCUMENT_SCHEMA_C.clone()]),
            },
            min_size: 1,
            max_size: Some(4),
        }),
        input = Stage::Set(Set {
            operation: SetOperation::Union,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a(), test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b(), test_document_c(),],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        set_union_of_two_empty_sets_has_min_and_max_size_0,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() =>
                        Schema::AnyOf(set![]),
                ("bar", 0u16).into() =>
                        Schema::AnyOf(set![]),
            },
            min_size: 0,
            max_size: Some(0),
        }),
        input = Stage::Set(Set {
            operation: SetOperation::Union,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    // EquiJoin tests
    test_schema!(
        left_equijoin,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::AnyOf(set![
                            Schema::Missing,
                            TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
            },
            min_size: 1,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Left,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            from: mir_collection("bar"),
            local_field: mir_field_access("foo", "a"),
            foreign_field: mir_field_access("bar", "a"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );
    test_schema!(
        inner_equijoin,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => TEST_DOCUMENT_SCHEMA_A.clone(),
                ("foo", 0u16).into() => Schema::AnyOf(set![ TEST_DOCUMENT_SCHEMA_A.clone() ]),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            from: mir_collection("bar"),
            local_field: mir_field_access("foo", "a"),
            foreign_field: mir_field_access("bar", "a"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );
    test_schema!(
        equijoin_fields_not_comparable,
        expected_error_code = 1005,
        expected = Err(schema::Error::InvalidComparison(
            "equijoin comparison",
            Schema::AnyOf(set![Schema::Atomic(Atomic::Integer),]),
            Schema::Atomic(Atomic::String),
        )),
        input = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            from: mir_collection("bar"),
            local_field: mir_field_access("foo", "a"),
            foreign_field: mir_field_access("bar", "s"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_S.clone(),
        }),
    );

    // LateralJoin tests
    test_schema!(
        left_lateral_join,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => Schema::AnyOf(set![
                            Schema::Missing,
                            TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
                ("foo", 0u16).into() => Schema::AnyOf(set![
                        TEST_DOCUMENT_SCHEMA_A.clone()
                    ]
                ),
            },
            min_size: 1,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Left,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            subquery: mir_collection("bar"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );
    test_schema!(
        inner_lateral_join,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("bar", 0u16).into() => TEST_DOCUMENT_SCHEMA_A.clone(),
                ("foo", 0u16).into() => Schema::AnyOf(set![ TEST_DOCUMENT_SCHEMA_A.clone() ]),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            subquery: mir_collection("bar"),
            cache: SchemaCache::new(),
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test_db".into(), collection: "foo".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
            Namespace {db: "test_db".into(), collection: "bar".into()} => TEST_DOCUMENT_SCHEMA_A.clone(),
        }),
    );

    // Join tests
    test_schema!(
        left_join,
        expected = Ok(ResultSet {
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
        input = Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::Literal(LiteralValue::Boolean(false).into())),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        cross_join,
        expected_pat = Ok(ResultSet {
            min_size: 6,
            max_size: Some(6),
            ..
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(6).into())
                        }
                        .into()
                    ),
                ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: None,
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        inner_join,
        expected_pat = Ok(ResultSet {
            min_size: 0,
            max_size: Some(6),
            ..
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(2).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(3).into())
                        }
                        .into()
                    )
                ],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(5).into())
                        }
                        .into()
                    ),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "b".into() => Expression::Literal(LiteralValue::Integer(6).into())
                        }
                        .into()
                    ),
                ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::Literal(LiteralValue::Boolean(false).into())),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        inner_and_left_join,
        expected = Ok(ResultSet {
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
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Join(Join {
                join_type: JoinType::Left,
                left: Box::new(Stage::Array(ArraySource {
                    array: vec![test_document_b()],
                    alias: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                right: Box::new(Stage::Array(ArraySource {
                    array: vec![test_document_c()],
                    alias: "car".into(),
                    cache: SchemaCache::new(),
                })),
                condition: Some(Expression::Literal(LiteralValue::Boolean(false).into())),
                cache: SchemaCache::new(),
            })),
            condition: None,
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        join_duplicate_datasource_names,
        expected_error_code = 1009,
        expected = Err(mir_error::DuplicateKey(("foo", 0u16).into())),
        input = Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            condition: None,
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        invalid_join_condition,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "join condition",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::Literal(LiteralValue::Integer(5).into())),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        join_condition_uses_left_datasource,
        expected_pat = Ok(ResultSet { .. }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Boolean(true).into())
                    }
                    .into()
                )],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_b()],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".to_string(),
                    cache: SchemaCache::new(),
                })),
                target_type: Type::Boolean,
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        join_condition_uses_right_datasource,
        expected_pat = Ok(ResultSet { .. }),
        input = Stage::Join(Join {
            join_type: JoinType::Left,
            left: Box::new(Stage::Array(ArraySource {
                array: vec![test_document_a()],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "b".into() => Expression::Literal(LiteralValue::Boolean(true).into())
                    }
                    .into()
                )],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                    field: "b".to_string(),
                    cache: SchemaCache::new(),
                })),
                target_type: Type::Boolean,
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        join_condition_uses_correlated_datasource,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Join(Join {
                    join_type: JoinType::Left,
                    left: Box::new(Stage::Array(ArraySource {
                        array: vec![test_document_b()],
                        alias: "bar".into(),
                        cache: SchemaCache::new(),
                    })),
                    right: Box::new(Stage::Array(ArraySource {
                        array: vec![test_document_c()],
                        alias: "car".into(),
                        cache: SchemaCache::new(),
                    })),
                    condition: Some(Expression::TypeAssertion(TypeAssertionExpr {
                        expr: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".to_string(),
                            cache: SchemaCache::new(),
                        })),
                        target_type: Type::Boolean,
                        cache: SchemaCache::new(),
                    })),
                    cache: SchemaCache::new(),
                }),),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
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
            catalog::Namespace,
            map,
            mir::{
                schema::Error as mir_error,
                schema::{CachedSchema, SchemaCache},
                *,
            },
            schema::*,
            set,
            usererror::UserError,
        };

        test_schema!(
            comparable_schemas,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("bar", 0u16).into() => ANY_DOCUMENT.clone(),
                },
                min_size: 0,
                max_size: None,
            }),
            input = Stage::Sort(Sort {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                specs: vec![
                    SortSpecification::Asc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))),
                    SortSpecification::Desc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "b".into(),
                        cache: SchemaCache::new(),
                    })))
                ],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document( Document{
                    keys: map! {
                        "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)]),
                        "b".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "a".into(), "b".into() },
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            incomparable_schemas,
            expected_error_code = 1010,
            expected = Err(mir_error::SortKeyNotSelfComparable(
                1,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String)
                ]),
            )),
            input = Stage::Sort(Sort {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                specs: vec![
                    SortSpecification::Asc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))),
                    SortSpecification::Asc(Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "b".into(),
                        cache: SchemaCache::new(),
                    })))
                ],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document( Document{
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)]),
                    },
                    required: set! {"a".into(), "b".into()},
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            mix_comparable_and_incomparable_schemas,
            expected_error_code = 1010,
            expected = Err(mir_error::SortKeyNotSelfComparable(
                0,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Null)
                ]),
            )),
            input = Stage::Sort(Sort {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                specs: vec![SortSpecification::Asc(Box::new(Expression::FieldAccess(
                    FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }
                ))),],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document( Document{
                    keys: map! {
                        "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)]),
                    },
                    required: set! {"a".into(), "b".into(), "c".into()},
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
    }

    mod group_by {
        use crate::{
            catalog::Namespace,
            map,
            mir::{
                binding_tuple::Key,
                schema::Error as mir_error,
                schema::{CachedSchema, SchemaCache},
                *,
            },
            schema::*,
            set,
            usererror::UserError,
        };
        fn group_stage_refs_only() -> Stage {
            Stage::Group(Group {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: schema::SchemaCache::new(),
                })),
                keys: vec![group_aliased_ref(), group_non_aliased_ref()],
                aggregations: vec![AliasedAggregation {
                    alias: "agg".to_string(),
                    agg_expr: AggregationExpr::CountStar(false),
                }],
                cache: schema::SchemaCache::new(),
                scope: 0,
            })
        }
        fn group_aliased_ref() -> OptionallyAliasedExpr {
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "A".into(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into(),
                    cache: SchemaCache::new(),
                }),
            })
        }
        fn group_non_aliased_ref() -> OptionallyAliasedExpr {
            OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "b".into(),
                cache: SchemaCache::new(),
            }))
        }

        test_schema!(
            key_schemas_are_all_self_comparable,
            expected_pat = Ok(_),
            input = group_stage_refs_only(),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "a".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Double)]),
                        "b".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "a".into(), "b".into() },
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            key_schemas_not_all_self_comparable,
            expected_error_code = 1011,
            expected = Err(mir_error::GroupKeyNotSelfComparable(
                1,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String)
                ]),
            )),
            input = group_stage_refs_only(),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)]),
                    },
                    required: set! {"a".into(), "b".into()},
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            aliased_field_access_mapped_to_bottom_datasource,
            expected = Ok(ResultSet {
                schema_env: map! {
                    Key::bot(0u16) => Schema::Document(Document {
                        keys: map! {
                            "A".into() => Schema::Atomic(Atomic::String),
                        },
                        required: set! { "A".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = Stage::Group(Group {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                keys: vec![group_aliased_ref()],
                aggregations: vec![],
                cache: SchemaCache::new(),
                scope: 0,
            }),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "a".into(), },
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            non_aliased_field_access_mapped_to_reference,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "b".into() => Schema::Atomic(Atomic::String),
                        },
                        required: set! { "b".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = Stage::Group(Group {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                keys: vec![group_non_aliased_ref()],
                aggregations: vec![],
                cache: SchemaCache::new(),
                scope: 0,
            }),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "b".into(), },
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            all_literal_keys_results_in_max_size_1,
            expected_pat = Ok(ResultSet {
                max_size: Some(1),
                ..
            }),
            input = Stage::Group(Group {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "a".into(),
                        expr: Expression::Literal(LiteralValue::Integer(1).into()),
                    }),
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "b".into(),
                        expr: Expression::Literal(LiteralValue::String("abc".into()).into()),
                    })
                ],
                aggregations: vec![],
                cache: SchemaCache::new(),
                scope: 0,
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            mix_literal_key_and_non_literal_key_results_in_no_max_size,
            expected_pat = Ok(ResultSet { max_size: None, .. }),
            input = Stage::Group(Group {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                keys: vec![
                    OptionallyAliasedExpr::Aliased(AliasedExpr {
                        alias: "literal".into(),
                        expr: Expression::Literal(LiteralValue::Integer(1).into()),
                    }),
                    OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                ],
                aggregations: vec![],
                cache: SchemaCache::new(),
                scope: 0,
            }),
            schema_env = map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "a".into() },
                    additional_properties: false,
                }),
            },
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
        test_schema!(
            aliased_key_and_multiple_agg_functions_all_correctly_unioned_under_bottom_datasource,
            expected = Ok(ResultSet {
                schema_env: map! {
                    Key::bot(0u16) => Schema::Document(Document {
                        keys: map! {
                            "A".into() => Schema::Atomic(Atomic::Boolean),
                            "B".into() => Schema::Atomic(Atomic::String),
                            "literal".into() => Schema::Atomic(Atomic::Integer),
                        },
                        required: set! {
                            "A".into(),
                            "B".into(),
                            "literal".into(),
                        },
                        additional_properties: false,
                    })
                },
                min_size: 0,
                max_size: Some(1),
            }),
            input = Stage::Group(Group {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "literal".into(),
                    expr: Expression::Literal(LiteralValue::Integer(1).into()),
                })],
                aggregations: vec![
                    AliasedAggregation {
                        alias: "A".to_string(),
                        agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                            function: AggregationFunction::First,
                            distinct: false,
                            arg: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        }),
                    },
                    AliasedAggregation {
                        alias: "B".to_string(),
                        agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                            function: AggregationFunction::First,
                            distinct: false,
                            arg: Expression::Literal(LiteralValue::String("abc".into()).into())
                                .into(),
                        }),
                    },
                ],
                cache: SchemaCache::new(),
                scope: 0,
            }),
            catalog = Catalog::new(map! {
                Namespace {db: "test".into(), collection: "bar".into()} => ANY_DOCUMENT.clone(),
            }),
        );
    }

    mod unwind {
        use crate::{
            catalog::{Catalog, Namespace},
            map,
            mir::{
                schema::Error,
                schema::{CachedSchema, SchemaCache},
                Collection, Expression, FieldAccess, Join, JoinType, Stage, Unwind,
            },
            schema::{Atomic, Document, ResultSet, Schema},
            set,
            usererror::UserError,
        };

        /// Most tests use the same source, path, and cache for the Unwind stage.
        /// Typically, only the index and outer parameters change, so this helper
        /// allows us to easily and concisely construct test Unwinds.
        fn make_unwind(index: Option<String>, outer: bool) -> Stage {
            Stage::Unwind(Unwind {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                path: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "arr".into(),
                    cache: SchemaCache::new(),
                })),
                index,
                outer,
                cache: SchemaCache::new(),
            })
        }

        /// Most tests use the same collection source and need to specify the
        /// collection schema for the test to work. This helper allows easy
        /// definition of that collection schema.
        fn make_catalog(s: Schema) -> Catalog {
            Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => s,
            })
        }

        mod no_index {
            use super::*;

            mod outer_false {
                use super::*;

                test_schema!(
                    path_result_schema_is_inner_schema_when_path_must_be_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::String),
                                },
                                required: set! { "arr".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_any_of_inner_schemas_and_other_schemas_when_path_may_be_array_but_never_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::Double),
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Integer),
                                    ]),
                                },
                                required: set! { "arr".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                                Schema::Atomic(Atomic::Integer),
                            ]),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_inner_schema_when_path_may_be_array_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::String),
                                },
                                required: set! { "arr".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing,
                            ]),
                        },
                        required: set! {},
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_unchanged_when_path_not_array_and_not_nullish,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::Integer),
                                },
                                required: set! { "arr".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::Integer),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_non_nullish_when_path_not_array_but_may_be_nullish,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::Integer),
                                },
                                required: set! { "arr".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::Integer),
                                Schema::Atomic(Atomic::Null),
                            ]),
                        },
                        required: set! {},
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_retains_path_schema_nullability_from_within_array_regardless_of_outer_false,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set![
                                        Schema::Atomic(Atomic::String),
                                        Schema::Atomic(Atomic::Null)
                                    ]),
                                },
                                required: set! { "arr".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::AnyOf(set![
                                Schema::Atomic(Atomic::String),
                                Schema::Atomic(Atomic::Null)
                            ]))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );
            }

            mod outer_true {
                use super::*;

                test_schema!(
                    path_result_schema_is_any_of_inner_schema_or_missing_when_path_must_be_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::String),
                                        Schema::Missing,
                                    ]),
                                },
                                required: set! { },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_any_of_inner_schemas_or_other_schemas_or_missing_when_path_may_be_array_but_never_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::Double),
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Integer),
                                        Schema::Missing,
                                    ]),
                                },
                                required: set! { },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                                Schema::Atomic(Atomic::Integer),
                            ]),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_any_of_inner_schema_null_or_missing_when_path_may_be_array_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::String),
                                        Schema::Atomic(Atomic::Null),
                                        Schema::Missing,
                                    ]),
                                },
                                required: set! { },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing,
                            ]),
                        },
                        required: set! { },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    path_result_schema_is_unchanged_when_path_not_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::Integer),
                                        Schema::Missing,
                                    ]),
                                },
                                required: set! { },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(None, true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::Integer),
                        },
                        required: set! {},
                        additional_properties: false,
                    })),
                );
            }

            test_schema!(
                correlated_path_disallowed,
                expected_error_code = 1000,
                expected = Err(Error::DatasourceNotFoundInSchemaEnv(("bar", 0u16).into())),
                input = Stage::Unwind(Unwind {
                    source: Box::new(Stage::Collection(Collection {
                        db: "test".into(),
                        collection: "foo".into(),
                        cache: SchemaCache::new(),
                    })),
                    path: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
                        field: "arr".into(),
                        cache: SchemaCache::new(),
                    })),
                    index: None,
                    outer: false,
                    cache: SchemaCache::new(),
                }),
                schema_env = map! {
                    ("bar", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    }),
                },
                catalog = make_catalog(Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                })),
            );
        }

        mod index_does_not_conflict {
            use super::*;

            mod outer_false {
                use super::*;

                test_schema!(
                    index_result_schema_is_not_nullable_when_path_must_be_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::String),
                                    "idx".into() => Schema::Atomic(Atomic::Long),
                                },
                                required: set! { "arr".into(), "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    index_result_schema_is_nullable_when_path_may_be_array_but_never_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set![
                                        Schema::Atomic(Atomic::Double),
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Integer),
                                    ]),
                                    "idx".into() => Schema::AnyOf(set![
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Null)
                                    ]),
                                },
                                required: set! { "arr".into(), "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                                Schema::Atomic(Atomic::Integer),
                            ]),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    index_result_schema_is_not_nullable_when_path_may_be_array_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::String),
                                    "idx".into() => Schema::Atomic(Atomic::Long),
                                },
                                required: set! { "arr".into(), "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing,
                            ]),
                        },
                        required: set! {},
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    index_result_schema_is_exactly_null_when_path_not_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::Integer),
                                    "idx".into() => Schema::Atomic(Atomic::Null),
                                },
                                required: set! { "arr".into(), "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), false),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::Integer),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );
            }

            mod outer_true {
                use super::*;

                test_schema!(
                    index_result_schema_is_nullable_when_path_must_be_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::String),
                                        Schema::Missing,
                                    ]),
                                    "idx".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Null),
                                    ]),
                                },
                                required: set! { "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    index_result_schema_is_nullable_when_path_may_be_array_but_never_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set![
                                        Schema::Atomic(Atomic::Double),
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Integer),
                                        Schema::Missing,
                                    ]),
                                    "idx".into() => Schema::AnyOf(set![
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Null)
                                    ]),
                                },
                                required: set! { "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                                Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                                Schema::Atomic(Atomic::Integer),
                            ]),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    index_result_schema_is_nullable_when_path_may_be_array_null_or_missing,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::String),
                                        Schema::Atomic(Atomic::Null),
                                        Schema::Missing,
                                    ]),
                                    "idx".into() => Schema::AnyOf(set! [
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Null),
                                    ]),
                                },
                                required: set! { "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing,
                            ]),
                        },
                        required: set! {},
                        additional_properties: false,
                    })),
                );

                test_schema!(
                    index_result_schema_is_exactly_null_when_path_not_array,
                    expected = Ok(ResultSet {
                        schema_env: map! {
                            ("foo", 0u16).into() => Schema::Document(Document {
                                keys: map! {
                                    "arr".into() => Schema::Atomic(Atomic::Integer),
                                    "idx".into() => Schema::Atomic(Atomic::Null),
                                },
                                required: set! { "arr".into(), "idx".into() },
                                additional_properties: false,
                            }),
                        },
                        min_size: 0,
                        max_size: None,
                    }),
                    input = make_unwind(Some("idx".into()), true),
                    catalog = make_catalog(Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::Integer),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    })),
                );
            }

            test_schema!(
                index_is_put_at_top_level_even_when_path_is_nested,
                expected = Ok(ResultSet {
                    schema_env: map! {
                        ("foo", 0u16).into() => Schema::Document(Document {
                            keys: map! {
                                "a".into() => Schema::Document(Document {
                                    keys: map! {
                                        "b".into() => Schema::Document(Document {
                                            keys: map! {
                                                "arr".into() => Schema::Atomic(Atomic::String),
                                            },
                                            required: set! { "arr".into() },
                                            additional_properties: false,
                                        }),
                                    },
                                    required: set! { "b".into() },
                                    additional_properties: false,
                                }),
                                "idx".into() => Schema::Atomic(Atomic::Long),
                            },
                            required: set! { "a".into(), "idx".into() },
                            additional_properties: false,
                        }),
                    },
                    min_size: 0,
                    max_size: None,
                }),
                input = Stage::Unwind(Unwind {
                    source: Box::new(Stage::Collection(Collection {
                        db: "test".into(),
                        collection: "foo".into(),
                        cache: SchemaCache::new(),
                    })),
                    path: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                field: "a".into(),
                                cache: SchemaCache::new(),
                            })),
                            field: "b".into(),
                            cache: SchemaCache::new(),
                        })),
                        field: "arr".into(),
                        cache: SchemaCache::new(),
                    })),
                    index: Some("idx".into()),
                    outer: false,
                    cache: SchemaCache::new(),
                }),
                catalog = make_catalog(Schema::Document(Document {
                    keys: map! {
                        "a".into() => Schema::Document(Document {
                            keys: map! {
                                "b".into() => Schema::Document(Document {
                                    keys: map! {
                                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                                    },
                                    required: set! { "arr".into() },
                                    additional_properties: false,
                                })
                            },
                            required: set! { "b".into() },
                            additional_properties: false,
                        }),
                    },
                    required: set! { "a".into() },
                    additional_properties: false,
                })),
            );

            test_schema!(
                index_conflicts_are_only_checked_against_path_datasource,
                expected = Ok(ResultSet {
                    schema_env: map! {
                        ("foo", 0u16).into() => Schema::Document(Document {
                            keys: map! {
                                "arr".into() => Schema::Atomic(Atomic::String),
                                "idx".into() => Schema::Atomic(Atomic::Long),
                            },
                            required: set! { "arr".into(), "idx".into() },
                            additional_properties: false,
                        }),
                        ("bar", 0u16).into() => Schema::Document(Document {
                            keys: map! {
                                "idx".into() => Schema::Atomic(Atomic::String),
                            },
                            required: set! { "idx".into() },
                            additional_properties: false,
                        }),
                    },
                    min_size: 0,
                    max_size: None,
                }),
                input = Stage::Unwind(Unwind {
                    source: Box::new(Stage::Join(Join {
                        join_type: JoinType::Inner,
                        left: Box::new(Stage::Collection(Collection {
                            db: "test".into(),
                            collection: "foo".into(),
                            cache: SchemaCache::new(),
                        })),
                        right: Box::new(Stage::Collection(Collection {
                            db: "test".into(),
                            collection: "bar".into(),
                            cache: SchemaCache::new(),
                        })),
                        condition: None,
                        cache: SchemaCache::new(),
                    })),
                    path: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "arr".into(),
                        cache: SchemaCache::new(),
                    })),
                    index: Some("idx".into()),
                    outer: false,
                    cache: SchemaCache::new(),
                }),
                catalog = Catalog::new(map! {
                    Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        },
                        required: set! { "arr".into() },
                        additional_properties: false,
                    }),
                    Namespace {db: "test".into(), collection: "bar".into()} => Schema::Document(Document {
                        keys: map! {
                            "idx".into() => Schema::Atomic(Atomic::String),
                        },
                        required: set! { "idx".into() },
                        additional_properties: false,
                    }),
                }),
            );
        }

        mod index_may_conflict {
            use super::*;
            test_schema!(
                error_in_strict_mode,
                expected_error_code = 1014,
                expected = Err(Error::UnwindIndexNameConflict("idx".into())),
                input = make_unwind(Some("idx".into()), false),
                catalog = make_catalog(Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        "idx".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "arr".into(), "idx".into() },
                    additional_properties: false,
                })),
            );

            test_schema!(
                succeed_in_relaxed_mode,
                expected = Ok(ResultSet {
                    schema_env: map! {
                        ("foo", 0u16).into() => Schema::Document(Document {
                            keys: map! {
                                "arr".into() => Schema::Atomic(Atomic::String),
                                "idx".into() => Schema::Atomic(Atomic::Long),
                            },
                            required: set! { "arr".into(), "idx".into() },
                            additional_properties: false,
                        }),
                    },
                    min_size: 0,
                    max_size: None,
                }),
                input = make_unwind(Some("idx".into()), false),
                catalog = make_catalog(Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        "idx".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                })),
                schema_checking_mode = SchemaCheckingMode::Relaxed,
            );
        }

        mod index_must_conflict {
            use super::*;
            test_schema!(
                error_in_strict_mode,
                expected_error_code = 1014,
                expected = Err(Error::UnwindIndexNameConflict("idx".into())),
                input = make_unwind(Some("idx".into()), false),
                catalog = make_catalog(Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        "idx".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "arr".into(), "idx".into() },
                    additional_properties: false,
                })),
            );

            test_schema!(
                error_in_relaxed_mode,
                expected_error_code = 1014,
                expected = Err(Error::UnwindIndexNameConflict("idx".into())),
                input = make_unwind(Some("idx".into()), false),
                catalog = make_catalog(Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        "idx".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "arr".into(), "idx".into() },
                    additional_properties: false,
                })),
                schema_checking_mode = SchemaCheckingMode::Relaxed,
            );
        }
    }
}

mod arithmetic_retain {
    use crate::{schema::*, set};

    test_retain!(
        atomics_retain_dominant,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Integer),
        ]
    );

    test_retain!(
        atomic_anyof_retains_any_gte_atomic,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::Atomic(Atomic::Double),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ])
        ]
    );

    test_retain!(
        anyof_atomic_retains_any_gte_atomic,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ]),
            Schema::Atomic(Atomic::Long),
        ]
    );

    test_retain!(
        anyof_anyof_retains_dominant_result_of_all_possible_pairs,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ]),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        ]
    );

    test_retain!(
        anyof_atomic_anyof_retains_dominant_result_of_all_possible_pairs,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ]),
            Schema::Atomic(Atomic::Double),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        ]
    );
}

mod max_numeric {
    use crate::schema::*;
    test_max_numeric!(
        integer_returned,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input1 = Schema::Atomic(Atomic::Integer),
        input2 = Schema::Atomic(Atomic::Integer)
    );

    test_max_numeric!(
        long_takes_priority_over_integer,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input1 = Schema::Atomic(Atomic::Long),
        input2 = Schema::Atomic(Atomic::Integer)
    );

    test_max_numeric!(
        double_takes_priority_over_long,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input1 = Schema::Atomic(Atomic::Long),
        input2 = Schema::Atomic(Atomic::Double)
    );

    test_max_numeric!(
        double_takes_priority_over_integer,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input1 = Schema::Atomic(Atomic::Integer),
        input2 = Schema::Atomic(Atomic::Double)
    );

    test_max_numeric!(
        decimal_takes_priority_over_double,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input1 = Schema::Atomic(Atomic::Decimal),
        input2 = Schema::Atomic(Atomic::Double)
    );

    test_max_numeric!(
        decimal_takes_priority_over_long,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input1 = Schema::Atomic(Atomic::Decimal),
        input2 = Schema::Atomic(Atomic::Long)
    );

    test_max_numeric!(
        decimal_takes_priority_over_integer,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input1 = Schema::Atomic(Atomic::Decimal),
        input2 = Schema::Atomic(Atomic::Integer)
    );
}

mod user_error_messages {

    mod schema_checking {
        use crate::{
            schema::{
                Atomic, Schema, ANY_DOCUMENT, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH,
                NUMERIC_OR_NULLISH, STRING_OR_NULLISH,
            },
            set,
        };

        test_user_error_messages! {
            operation_needs_nullable_numeric_type,
            input = Error::SchemaChecking{
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            },
            expected = "Incorrect argument type for `Add`. Required: nullable numeric type. Found: string."
        }

        test_user_error_messages! {
            operation_needs_nullable_string_type,
            input = Error::SchemaChecking{
                name: "Concat",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            },
            expected = "Incorrect argument type for `Concat`. Required: nullable string. Found: int."
        }

        test_user_error_messages! {
            operation_needs_nullable_boolean_type,
            input = Error::SchemaChecking{
                name: "SearchedCase",
                required: BOOLEAN_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            },
            expected = "Incorrect argument type for `SearchedCase`. Required: nullable boolean. Found: string."
        }

        test_user_error_messages! {
            operation_needs_nullable_date_type,
            input = Error::SchemaChecking{
                name: "Second",
                required: DATE_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            },
            expected = "Incorrect argument type for `Second`. Required: nullable date. Found: int."
        }

        test_user_error_messages! {
            array_datasource_has_wrong_type,
            input = Error::SchemaChecking{
                name: "array datasource items",
                required: ANY_DOCUMENT.clone(),
                found: Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
            },
            expected = "Incorrect argument type for `array datasource items`. Required: object type. Found: int."
        }
    }

    mod invalid_comparison {
        use crate::schema::{Atomic, Schema};

        test_user_error_messages! {
            invalid_comparison,
            input = Error::InvalidComparison(
                "Lte",
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ),
            expected = "Invalid use of `Lte` due to incomparable types: `int` cannot be compared to `string`."
        }
    }

    mod sort_key_comparable {
        use crate::{
            schema::{Atomic, Schema},
            set,
        };

        test_user_error_messages! {
            sort_key_not_comparable_any,
            input = Error::SortKeyNotSelfComparable(0, Schema::Any),
            expected = "Cannot sort by key because `any type` can't be compared against itself."
        }

        test_user_error_messages! {
            sort_key_not_comparable_other_types,
            input = Error::SortKeyNotSelfComparable(0,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String)
                ])
            ),
            expected = "Cannot sort by key because `polymorphic type` can't be compared against itself."
        }
    }

    mod group_key_comparable {
        use crate::{
            schema::{Atomic, Schema},
            set,
        };

        test_user_error_messages! {
            group_key_not_comparable_any,
            input = Error::GroupKeyNotSelfComparable(0, Schema::Any),
            expected = "Cannot group by key because `any type` can't be compared against itself."
        }

        test_user_error_messages! {
            group_key_not_comparable_other_types,
            input = Error::GroupKeyNotSelfComparable(
                1,
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String)
                ]),
            ),
            expected = "Cannot group by key because `polymorphic type` can't be compared against itself."
        }
    }
}
