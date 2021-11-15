macro_rules! test_schema {
    ($func_name:ident, $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? input = $input:expr, $(schema_env = $schema_env:expr,)? $(catalog = $catalog:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{ir::schema::SchemaInferenceState, schema::SchemaEnvironment, catalog::Catalog};

            let input = $input;

            #[allow(unused_mut, unused_assignments)]
            let mut schema_env = SchemaEnvironment::default();
            $(schema_env = $schema_env;)?

            #[allow(unused_mut, unused_assignments)]
            let mut catalog = Catalog::default();
            $(catalog = $catalog;)?

            let state = SchemaInferenceState::new(0u16, schema_env, &catalog);
            let actual = input.schema(&state);

            $(assert!(matches!(actual, $expected_pat));)?
            $(assert_eq!(actual, $expected);)?
        }
    };
}

macro_rules! test_constant_fold {
    ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
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
    ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
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

macro_rules! test_retain {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::ir::schema::retain;
            let expected = $expected;
            let actual = retain(&$input);
            assert_eq!(actual, expected);
        }
    };
}

macro_rules! test_max_numeric {
    ($func_name:ident, expected = $expected:expr, input1 = $input1:expr, input2 = $input2:expr) => {
        #[test]
        fn $func_name() {
            use crate::ir::schema::max_numeric;
            let expected = $expected;
            let actual = max_numeric(&$input1, &$input2);
            assert_eq!(actual, expected);
        }
    };
}

mod schema {
    use crate::{
        catalog::*,
        ir::{
            binding_tuple::DatasourceName::Bottom,
            schema::{CachedSchema, Error as ir_error, SchemaCache},
            *,
        },
        map,
        schema::*,
        set, unchecked_unique_linked_hash_map,
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
        expected = Err(ir_error::DatasourceNotFoundInSchemaEnv(("a", 0u16).into())),
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::AccessMissingField("foo".to_string())),
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

    // General function schema checking.
    test_schema!(
        arg_may_satisfy_schema_is_not_sufficient,
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::AggregationArgumentMustBeSelfComparable(
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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

    // Arithmetic function errors.
    test_schema!(
        sub_requires_exactly_two_args,
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::CannotMergeObjects(
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
        expected = Err(ir_error::CannotMergeObjects(
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        coalesce_returns_all_arg_schemas_with_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::String("abc".to_string()).into()),
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
        }),
    );
    test_schema!(
        coalesce_upconverts_missing_to_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ])),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Coalesce,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Reference(("missing", 0u16).into()),
            ],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("missing", 0u16).into() => Schema::Missing,
        },
    );

    // Slice function.
    test_schema!(
        slice_requires_more_than_one_arg,
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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

    // Size function.
    test_schema!(
        size_requires_one_arg,
        expected = Err(ir_error::IncorrectArgumentCount {
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
        expected = Err(ir_error::SchemaChecking {
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
        collection_schema,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Collection(Collection {
            db: "test2".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
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
        expected = Err(ir_error::SchemaChecking {
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
    );

    test_schema!(
        namespace_in_catalog,
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
        namespace_not_in_catalog,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("baz", 0u16).into() => crate::schema::ANY_DOCUMENT.clone()
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Collection(Collection {
            db: "foo".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "foo".into(), collection: "bar".into()} => Schema::Atomic(Atomic::Integer)
        }),
    );

    mod filter {
        use crate::{
            ir::{schema::CachedSchema, schema::Error as ir_error, schema::SchemaCache, *},
            map,
            schema::{Atomic, ResultSet, Schema},
            set, unchecked_unique_linked_hash_map,
        };

        fn true_ir() -> Expression {
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
                condition: true_ir(),
                cache: SchemaCache::new(),
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
        );
        test_schema!(
            non_boolean_condition_disallowed,
            expected = Err(ir_error::SchemaChecking {
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
        );
        test_schema!(
            possible_non_boolean_condition_disallowed,
            expected = Err(ir_error::SchemaChecking {
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
        );
        test_schema!(
            source_fails_schema_check,
            expected_pat = Err(ir_error::SchemaChecking {
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
                condition: true_ir(),
                cache: SchemaCache::new(),
            }),
        );
        test_schema!(
            condition_fails_schema_check,
            expected = Err(ir_error::DatasourceNotFoundInSchemaEnv(
                ("abc", 0u16).into()
            )),
            input = Stage::Filter(Filter {
                source: Box::new(test_source()),
                condition: Expression::Reference(("abc", 0u16).into()),
                cache: SchemaCache::new(),
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
                condition: true_ir(),
                source: Stage::Array(ArraySource {
                    alias: "arr".into(),
                    array: vec![Expression::Document(unchecked_unique_linked_hash_map!{"a".into() => Expression::Literal(LiteralValue::Null.into()),}.into())],
                    cache: SchemaCache::new(),
                }).into(),
                cache: SchemaCache::new(),
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::SchemaChecking {
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
        expected = Err(ir_error::InvalidComparison(
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
    );

    test_schema!(
        exists_invalid_expression,
        expected = Err(ir_error::IncorrectArgumentCount {
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
    );

    // Subquery Expression

    test_schema!(
        subquery_output_expr_violates_type_constraints,
        expected = Err(ir_error::AccessMissingField("_2".into())),
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
        expected = Err(ir_error::DatasourceNotFoundInSchemaEnv(
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
        expected = Err(ir_error::InvalidSubqueryCardinality),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );

    // Analogous SQL query: "SELECT (SELECT * FROM [{'a': 5}, {'a': 6}] AS foo)"
    test_schema!(
        subquery_expression_cardinality_gt_one,
        expected = Err(ir_error::InvalidSubqueryCardinality),
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
        expected = Err(ir_error::InvalidComparison(
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
        expected = Err(ir_error::DuplicateKey(("foo", 0u16).into())),
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
        expected = Err(ir_error::SchemaChecking {
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
            ir::{
                schema::{CachedSchema, Error as ir_error, SchemaCache},
                *,
            },
            map,
            schema::*,
            set,
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
        );
        test_schema!(
            incomparable_schemas,
            expected = Err(ir_error::SortKeyNotSelfComparable(
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
        );
        test_schema!(
            mix_comparable_and_incomparable_schemas,
            expected = Err(ir_error::SortKeyNotSelfComparable(
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
        );
    }

    mod group_by {
        use crate::{
            ir::{
                binding_tuple::Key,
                schema::Error as ir_error,
                schema::{CachedSchema, SchemaCache},
                *,
            },
            map,
            schema::*,
            set,
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
        );
        test_schema!(
            key_schemas_not_all_self_comparable,
            expected = Err(ir_error::GroupKeyNotSelfComparable(
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
            }),
        );
    }
}

mod constant_folding {
    use crate::{
        ir::{definitions::*, schema::SchemaCache},
        unchecked_unique_linked_hash_map,
    };

    fn test_source() -> Stage {
        Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })
    }

    test_constant_fold!(
        literal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        or_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        and_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        true_and_nulls_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_and_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        true_and_nulls_and_ref_is_null_and_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        false_or_nulls_or_ref_is_null_or_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_or_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        false_or_nulls_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        or_with_true_literal_is_true,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        or_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        and_with_false_literal_is_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        and_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(3).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_to_zero_is_zero,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Long(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(-1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(-1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_constant_ref_is_constant_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Double(3.0).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_zero_ref_is_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Reference(("a", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(-1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Long(8).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_to_one_is_one,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Double(1.0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Double(0.5).into()),
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_one_ref_is_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Reference(("a", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                    Expression::Literal(LiteralValue::Double(0.5).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        arithmetic_null_arg,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Long(2).into()),
                            Expression::Literal(LiteralValue::Double(2.0).into()),
                            Expression::Literal(LiteralValue::Null.into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_different_num_types,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Long(4).into()),
                    Expression::Literal(LiteralValue::Double(6.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Double(3.0).into()),
                    Expression::Literal(LiteralValue::Double(3.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_different_num_types,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(4).into()),
                    Expression::Literal(LiteralValue::Long(9).into()),
                    Expression::Literal(LiteralValue::Double(16.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Long(3).into()),
                    Expression::Literal(LiteralValue::Long(3).into()),
                    Expression::Literal(LiteralValue::Double(4.0).into()),
                    Expression::Literal(LiteralValue::Double(4.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        sub_ref_by_zero_is_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Reference(("a", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::Long(0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        sub_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Long(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        sub_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Long(-1).into()),
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Long(1).into()),
                            Expression::Literal(LiteralValue::Long(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Double(3.0).into()),
                            Expression::Literal(LiteralValue::Double(1.0).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        div_zero_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Div,
                args: vec![
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::Long(0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        div_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Div,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Long(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        div_ref_by_one_is_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Reference(("a", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Div,
                args: vec![
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::Long(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        div_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(-1).into()),
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(LiteralValue::Long(-2).into()),
                            Expression::Literal(LiteralValue::Long(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(LiteralValue::Double(2.0).into()),
                            Expression::Literal(LiteralValue::Double(1.0).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_less_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_less_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_less_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_less_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_greater_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gt,
                args: vec![
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_greater_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gt,
                args: vec![
                    Expression::Literal(LiteralValue::Long(0).into()),
                    Expression::Literal(LiteralValue::Long(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_greater_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gte,
                args: vec![
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_greater_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gte,
                args: vec![
                    Expression::Literal(LiteralValue::Double(0.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(0).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_nequal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(0).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_nequal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        compare_different_datatypes,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        compare_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_between,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_between,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        fold_comparison_nested,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        fold_between_args,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Reference(("foo", 1u16).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Integer(3).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Reference(("foo", 1u16).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(-1).into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        pos_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Pos,
                args: vec![Expression::Literal(LiteralValue::Integer(2).into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        neg_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(-2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neg,
                args: vec![Expression::Literal(LiteralValue::Integer(2).into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        not_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Not,
                args: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        upper_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("AABBCC".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Upper,
                args: vec![Expression::Literal(
                    LiteralValue::String("aaBBcC".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        lower_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("aabbcc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lower,
                args: vec![Expression::Literal(
                    LiteralValue::String("aaBBcC".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        lower_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lower,
                args: vec![Expression::Literal(LiteralValue::Null.into()),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        btrim_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("AABBCC".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::String("aAABBCCa".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        ltrim_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("AABBCCa".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::LTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::String("aAABBCCa".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        rtrim_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("aAABBCC".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::RTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::String("aAABBCCa".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        btrim_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::Null.into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_nested,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("hello".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into()),
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_multi_codepoint_char,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("🇷🇺ááá".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("ááá🇷🇺ááá".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_negative_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("world".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                    Expression::Literal(LiteralValue::Integer(-1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_negative_start_with_smaller_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(-6).into()),
                    Expression::Literal(LiteralValue::Integer(5).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_negative_start_with_larger_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("hello".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(-6).into()),
                    Expression::Literal(LiteralValue::Integer(11).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_start_larger_than_string_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(20).into()),
                    Expression::Literal(LiteralValue::Integer(4).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_end_larger_than_string_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("world".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                    Expression::Literal(LiteralValue::Integer(20).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("hello world".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world".to_string()).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_with_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::String("hello world2".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world2".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::String("hello".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        char_length_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(11).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::CharLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("hello world".to_string()).into()
                )],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        char_length_multi_codepoint,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(14).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::CharLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("ááá🇷🇺ááá".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        octet_length_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(11).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::OctetLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("hello world".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        octet_length_multi_codepoint,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(26).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::OctetLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("ááá🇷🇺ááá".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        bit_length_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(88).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BitLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("hello world".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        bit_length_multi_codepoint,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(208).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BitLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("ááá🇷🇺ááá".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        array_size_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Size,
                args: vec![Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(0).into()),
                        Expression::Literal(LiteralValue::Integer(0).into())
                    ]
                    .into()
                )],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        array_size_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Size,
                args: vec![Expression::Array(vec![].into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        coalesce_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Coalesce,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        coalesce_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Coalesce,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        coalesce_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Coalesce,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        merge_objects_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                "b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
            .into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        merge_objects_reference,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        merge_objects_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {}.into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        merge_objects_combine_early_docs,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_if_args_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::NullIf,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_if_args_unequal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::NullIf,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        computed_field_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Literal(LiteralValue::String("a".into()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        computed_field_missing,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Literal(LiteralValue::String("b".into()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Literal(LiteralValue::String("b".into()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![Expression::Literal(LiteralValue::Integer(2).into())].into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_negative_length_no_start,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_positive_length_no_start,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_negative_start_within_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![Expression::Literal(LiteralValue::Integer(2).into())].into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-2).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_negative_start_outside_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-5).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_negative_len_longer_than_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-5).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_positive_len_longer_than_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(5).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_positive_len_longer_than_array_no_start,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(5).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_start_larger_than_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(vec![].into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_with_pos_neg_length_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::Literal(LiteralValue::Integer(-1).into())
                ],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                to: Type::Boolean,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_mismatched_types,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                to: Type::String,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                to: Type::String,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_array_as_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![Expression::Literal(LiteralValue::Boolean(true).into())].into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Array(
                    vec![Expression::Literal(LiteralValue::Boolean(true).into())].into()
                )
                .into(),
                to: Type::Array,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_non_array_as_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("error".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Integer(0).into()).into(),
                to: Type::Array,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_document_as_document,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
            .into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                .into())
                .into(),
                to: Type::Document,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_non_document_as_document,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("error".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Integer(0).into()).into(),
                to: Type::Document,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("null".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Null.into()).into(),
                to: Type::Array,
                on_null: Expression::Literal(LiteralValue::String("null".into()).into()).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::Literal(LiteralValue::Integer(1).into()).into(),
                target_type: TypeOrMissing::Type(Type::Int32),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::Literal(LiteralValue::String("a".into()).into()).into(),
                target_type: TypeOrMissing::Type(Type::Double),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_number,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::Literal(LiteralValue::Double(1.0).into()).into(),
                target_type: TypeOrMissing::Number,
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_nested,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Concat,
                    args: vec![
                        Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                        Expression::Literal(LiteralValue::String("world".to_string()).into()),
                    ],
                    cache: SchemaCache::new(),
                })
                .into(),
                target_type: TypeOrMissing::Type(Type::String),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("then 2".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(3).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 2".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_ref_ahead,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 2".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 2".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_prune_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then a".into()).into()).into()
                },],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(3).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then a".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_else,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("else".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(3).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_keep_branches,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("then true".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(false).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then false".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then true".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_ref_ahead,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then true".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then true".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_prune_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into()).into()).into()
                },],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(false).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then false".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::String("false".into()).into())
                            .into(),
                        then: Expression::Literal(
                            LiteralValue::String("then false string".into()).into()
                        )
                        .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_else,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("else".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(false).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then false".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        field_access_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into())}
                .into())
                .into(),
                field: "a".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        field_access_missing_field,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into())}
                .into())
                .into(),
                field: "b".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into())}
                .into())
                .into(),
                field: "b".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        field_access_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                field: "a".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                field: "a".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        offset_simple,
        expected = test_source(),
        input = Stage::Offset(Offset {
            source: Box::new(test_source()),
            offset: 0u64,
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        offset_nonzero,
        expected = Stage::Offset(Offset {
            source: Box::new(test_source()),
            offset: 1u64,
            cache: SchemaCache::new(),
        }),
        input = Stage::Offset(Offset {
            source: Box::new(test_source()),
            offset: 1u64,
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        filter_simple,
        expected = test_source(),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::Literal(LiteralValue::Boolean(true).into()),
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        filter_non_literal,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::Reference(("a", 0u16).into()),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::Reference(("a", 0u16).into()),
            cache: SchemaCache::new(),
        }),
    );
}

mod flatten_node {
    use crate::ir::{schema::SchemaCache, *};
    fn test_source() -> Stage {
        Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })
    }

    test_flatten_variadic_functions!(
        flatten_simple,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_ignores_different_funcs,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_nested_multiple_levels,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(4).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Add,
                            args: vec![
                                Expression::Literal(LiteralValue::Integer(1).into()),
                                Expression::Literal(LiteralValue::Integer(2).into()),
                                Expression::Literal(LiteralValue::Integer(4).into())
                            ],
                            cache: SchemaCache::new(),
                        })],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_multiple_funcs,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(3).into()),
                    Expression::Literal(LiteralValue::Integer(4).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(3).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Add,
                            args: vec![
                                Expression::Literal(LiteralValue::Integer(3).into()),
                                Expression::Literal(LiteralValue::Integer(4).into()),
                                Expression::ScalarFunction(ScalarFunctionApplication {
                                    function: ScalarFunction::Mul,
                                    args: vec![
                                        Expression::ScalarFunction(ScalarFunctionApplication {
                                            function: ScalarFunction::Mul,
                                            args: vec![
                                                Expression::Literal(
                                                    LiteralValue::Integer(2).into()
                                                ),
                                                Expression::Literal(
                                                    LiteralValue::Integer(1).into()
                                                )
                                            ],
                                            cache: SchemaCache::new(),
                                        }),
                                        Expression::ScalarFunction(ScalarFunctionApplication {
                                            function: ScalarFunction::Mul,
                                            args: vec![
                                                Expression::Literal(
                                                    LiteralValue::Integer(3).into()
                                                ),
                                                Expression::Literal(
                                                    LiteralValue::Integer(1).into()
                                                )
                                            ],
                                            cache: SchemaCache::new(),
                                        })
                                    ],
                                    cache: SchemaCache::new(),
                                })
                            ],
                            cache: SchemaCache::new(),
                        })],
                        cache: SchemaCache::new(),
                    },),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_flatten_variadic_functions!(
        flatten_not_necessary,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(1).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );
    test_flatten_variadic_functions!(
        flatten_order_matters,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("foo".to_string()).into()),
                    Expression::Literal(LiteralValue::String("bar".to_string()).into()),
                    Expression::Literal(LiteralValue::String("baz".to_string()).into())
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("foo".to_string()).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Concat,
                        args: vec![
                            Expression::Literal(LiteralValue::String("bar".to_string()).into()),
                            Expression::Literal(LiteralValue::String("baz".to_string()).into())
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }),
    );
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
