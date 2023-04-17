macro_rules! test_translate_expression {
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(mapping_registry = $mapping_registry:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{translator, mapping_registry::MqlMappingRegistry};

            // force the input
            let input = $input;
            #[allow(unused_mut, unused_assignments)]
            let mut mapping_registry = MqlMappingRegistry::default();
            $(mapping_registry = $mapping_registry;)?

            let translator = translator::MqlTranslator{
                mapping_registry,
                scope_level: 0u16,
            };
            let expected = $expected;
            let actual = translator.translate_expression(input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_translate_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{air, mir, translator};
            let mut translator = translator::MqlTranslator::new();
            let expected = $expected;
            let actual = translator.translate_stage($input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_translate_plan {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::{air, mir, translator};
            let mut translator = translator::MqlTranslator::new();
            let expected = $expected;
            let actual = translator.translate_plan($input);
            assert_eq!(expected, actual);
        }
    };
}

mod literal_expression {
    use crate::{air, mir};
    test_translate_expression!(
        null,
        expected = Ok(air::Expression::Literal(air::LiteralValue::Null)),
        input = mir::Expression::Literal(mir::LiteralValue::Null.into()),
    );
    test_translate_expression!(
        boolean,
        expected = Ok(air::Expression::Literal(air::LiteralValue::Boolean(true))),
        input = mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
    );
    test_translate_expression!(
        integer,
        expected = Ok(air::Expression::Literal(air::LiteralValue::Integer(1))),
        input = mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
    );
    test_translate_expression!(
        string,
        expected = Ok(air::Expression::Literal(air::LiteralValue::String(
            "foo".to_string()
        ))),
        input = mir::Expression::Literal(mir::LiteralValue::String("foo".to_string()).into()),
    );
    test_translate_expression!(
        long,
        expected = Ok(air::Expression::Literal(air::LiteralValue::Long(2))),
        input = mir::Expression::Literal(mir::LiteralValue::Long(2).into()),
    );
    test_translate_expression!(
        double,
        expected = Ok(air::Expression::Literal(air::LiteralValue::Double(3.0))),
        input = mir::Expression::Literal(mir::LiteralValue::Double(3.0).into()),
    );
}

mod document_expression {
    use crate::unchecked_unique_linked_hash_map;
    use crate::{air, mir, translator::Error};
    test_translate_expression!(
        empty,
        expected = Ok(air::Expression::Document(
            unchecked_unique_linked_hash_map! {}
        )),
        input = mir::Expression::Document(unchecked_unique_linked_hash_map! {}.into()),
    );
    test_translate_expression!(
        non_empty,
        expected = Ok(air::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))}
        )),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),}
        .into()),
    );
    test_translate_expression!(
        nested,
        expected = Ok(air::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                "bar".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "baz".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2))
                }),
            }
        )),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                "bar".to_string() => mir::Expression::Document(unchecked_unique_linked_hash_map!{
                    "baz".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(2).into())
                }.into()),
            }
            .into()
        ),
    );
    test_translate_expression!(
        dollar_prefixed_key_disallowed,
        expected = Err(Error::InvalidDocumentKey("$foo".to_string())),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"$foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}.into()),
    );
    test_translate_expression!(
        key_containing_dot_disallowed,
        expected = Err(Error::InvalidDocumentKey("foo.bar".to_string())),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo.bar".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}.into(),
        ),
    );
    test_translate_expression!(
        empty_key_disallowed,
        expected = Err(Error::InvalidDocumentKey("".to_string())),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}.into()),
    );
}

mod array_expression {
    use crate::{air, mir};
    test_translate_expression!(
        empty,
        expected = Ok(air::Expression::Array(vec![])),
        input = mir::Expression::Array(vec![].into()),
    );
    test_translate_expression!(
        non_empty,
        expected = Ok(air::Expression::Array(vec![air::Expression::Literal(
            air::LiteralValue::String("abc".to_string())
        )])),
        input = mir::Expression::Array(
            vec![mir::Expression::Literal(
                mir::LiteralValue::String("abc".into()).into()
            )]
            .into()
        ),
    );
    test_translate_expression!(
        nested,
        expected = Ok(air::Expression::Array(vec![
            air::Expression::Literal(air::LiteralValue::Null),
            air::Expression::Array(vec![air::Expression::Literal(air::LiteralValue::Null)])
        ])),
        input = mir::Expression::Array(
            vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Array(
                    vec![mir::Expression::Literal(mir::LiteralValue::Null.into())].into()
                )
            ]
            .into()
        ),
    );
}

mod scalar_function_expression {
    use crate::{air, mir, unchecked_unique_linked_hash_map};

    test_translate_expression!(
        concat,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Concat,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Concat,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        pos,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Pos,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pos,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        neg,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Neg,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Neg,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        add,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Add,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Add,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        sub,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Subtract,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sub,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        mul,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Multiply,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mul,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        div,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Divide,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Div,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        lt,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Lt,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        lte,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Lte,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lte,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        eq,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Eq,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Eq,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        ne,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Ne,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Neq,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        gt,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Gt,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        gte,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Gte,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gte,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        between,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Between,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Between,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        not,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Not,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Not,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        and,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::And,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::And,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        or,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Or,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Or,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        computed_field_access,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::ComputedFieldAccess,
                args: vec![
                    air::Expression::Document(
                        unchecked_unique_linked_hash_map! {"foo".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))}
                    ),
                    air::Expression::Literal(air::LiteralValue::String("foo".into())),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::ComputedFieldAccess,
            args: vec![
                mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {"foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),}
                .into()),
                mir::Expression::Literal(mir::LiteralValue::String("foo".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        null_if,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::NullIf,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::NullIf,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        coalesce,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Coalesce,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Coalesce,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        slice,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Slice,
                args: vec![
                    air::Expression::Array(vec![air::Expression::Literal(
                        air::LiteralValue::String("abc".to_string())
                    )]),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Slice,
            args: vec![
                mir::Expression::Array(
                    vec![mir::Expression::Literal(
                        mir::LiteralValue::String("abc".into()).into()
                    )]
                    .into()
                ),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        size,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Size,
                args: vec![air::Expression::Literal(air::LiteralValue::Null)],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Size,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into())],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        position,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::IndexOfCP,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Position,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        char_length,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::StrLenCP,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::CharLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        octet_length,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::StrLenBytes,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::OctetLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        bit_length,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::BitLength,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BitLength,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        abs,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Abs,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Abs,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        ceil,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Ceil,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Ceil,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        floor,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Floor,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Floor,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        log,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Log,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Log,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        modulo,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Mod,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mod,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        pow,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Pow,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pow,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        radians,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::DegreesToRadians,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Radians,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        degrees,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::RadiansToDegrees,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Degrees,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        round,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Round,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Round,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        cos,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Cos,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Cos,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        sin,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Sin,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sin,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        tan,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Tan,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Tan,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        sqrt,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Sqrt,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sqrt,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        substring,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::SubstrCP,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Substring,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        upper,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::ToUpper,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Upper,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        lower,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::ToLower,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lower,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        trim,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Trim,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::String("h".into())),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        ltrim,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::LTrim,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::String("h".into())),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::LTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        rtrim,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::RTrim,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::String("h".into())),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::RTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        split,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Split,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Split,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        current_time_stamp,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::CurrentTimestamp,
                args: vec![],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::CurrentTimestamp,
            args: vec![],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        year,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Year,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Year,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        month,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Month,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Month,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        day,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::DayOfMonth,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Day,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        hour,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Hour,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Hour,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        minute,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Minute,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Minute,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        second,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Second,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Second,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        week,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Week,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Week,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        day_of_year,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::DayOfYear,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::DayOfYear,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        iso_week,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::IsoWeek,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::IsoWeek,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        iso_week_day,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::IsoDayOfWeek,
                args: vec![air::Expression::Literal(air::LiteralValue::Null),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::IsoWeekday,
            args: vec![mir::Expression::Literal(mir::LiteralValue::Null.into()),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        merge_objects,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::MergeObjects,
                args: vec![
                    air::Expression::Document(
                        unchecked_unique_linked_hash_map! {"foo".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))}
                    ),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::MergeObjects,
            args: vec![
                mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {"foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),}
                .into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod cast_expression {
    use crate::{air, mir};

    test_translate_expression!(
        cast_expression_basic,
        expected = Ok(air::Expression::Convert(air::Convert {
            input: air::Expression::Literal(air::LiteralValue::String(
                "2012-12-20T12:12:12Z".to_string()
            ))
            .into(),
            to: air::Type::Datetime,
            on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
            on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
        })),
        input = mir::Expression::Cast(mir::CastExpr {
            expr: mir::Expression::Literal(
                mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
            )
            .into(),
            to: mir::Type::Datetime,
            on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
            on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        cast_expression_array,
        expected = Ok(air::Expression::SqlConvert(air::SqlConvert {
            input: air::Expression::Literal(air::LiteralValue::String(
                "2012-12-20T12:12:12Z".to_string()
            ))
            .into(),
            to: air::SqlConvertTargetType::Array,
            on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
            on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
        })),
        input = mir::Expression::Cast(mir::CastExpr {
            expr: mir::Expression::Literal(
                mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
            )
            .into(),
            to: mir::Type::Array,
            on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
            on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        cast_expression_doc,
        expected = Ok(air::Expression::SqlConvert(air::SqlConvert {
            input: air::Expression::Literal(air::LiteralValue::String(
                "2012-12-20T12:12:12Z".to_string()
            ))
            .into(),
            to: air::SqlConvertTargetType::Document,
            on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
            on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
        })),
        input = mir::Expression::Cast(mir::CastExpr {
            expr: mir::Expression::Literal(
                mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
            )
            .into(),
            to: mir::Type::Document,
            on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
            on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod type_assertion_expression {
    use crate::{air, mir};

    test_translate_expression!(
        type_assertion_expression_basic,
        expected = Ok(air::Expression::Literal(air::LiteralValue::String(
            "2012-12-20T12:12:12Z".to_string()
        ))),
        input = mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
            expr: mir::Expression::Literal(
                mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
            )
            .into(),
            target_type: mir::Type::Datetime,
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod reference_expression {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir,
        translator::Error,
    };
    test_translate_expression!(
        not_found,
        expected = Err(Error::ReferenceNotFound(("f", 0u16).into())),
        input = mir::Expression::Reference(("f", 0u16).into()),
    );

    test_translate_expression!(
        found_field_ref,
        expected = Ok(air::Expression::FieldRef(air::FieldRef {
            parent: None,
            name: "f".to_string()
        })),
        input = mir::Expression::Reference(("f", 0u16).into()),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );

    test_translate_expression!(
        found_variable,
        expected = Ok(air::Expression::Variable(air::Variable {
            parent: None,
            name: "f".to_string()
        })),
        input = mir::Expression::Reference(("f", 0u16).into()),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::Variable),
            );
            mr
        },
    );
}

mod like_expression {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir,
    };
    test_translate_expression!(
        like_expr,
        expected = Ok(air::Expression::Like(air::Like {
            expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                name: "input".to_string(),
                parent: None
            })),
            pattern: Box::new(air::Expression::FieldRef(air::FieldRef {
                name: "pattern".to_string(),
                parent: None
            })),
            escape: Some("escape".to_string())
        })),
        input = mir::Expression::Like(mir::LikeExpr {
            expr: mir::Expression::Reference(("input", 0u16).into()).into(),
            pattern: mir::Expression::Reference(("pattern", 0u16).into()).into(),
            escape: Some("escape".to_string()),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("input", 0u16),
                MqlMappingRegistryValue::new("input".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("pattern", 0u16),
                MqlMappingRegistryValue::new("pattern".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
}

mod field_access_expression {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir, unchecked_unique_linked_hash_map,
    };

    test_translate_expression!(
        from_reference,
        expected = Ok(air::Expression::FieldRef(air::FieldRef {
            name: "sub".to_string(),
            parent: Some(Box::new(air::FieldRef {
                name: "f".to_string(),
                parent: None
            }))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Reference(("f", 0u16).into()).into(),
            field: "sub".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
    test_translate_expression!(
        from_field_access,
        expected = Ok(air::Expression::FieldRef(air::FieldRef {
            name: "sub2".to_string(),
            parent: Some(Box::new(air::FieldRef {
                name: "sub1".to_string(),
                parent: Some(Box::new(air::FieldRef {
                    name: "f".to_string(),
                    parent: None
                }))
            }))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            field: "sub2".to_string(),
            expr: mir::Expression::FieldAccess(mir::FieldAccess {
                expr: mir::Expression::Reference(("f", 0u16).into()).into(),
                field: "sub1".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
    test_translate_expression!(
        from_non_reference_expr_with_nesting,
        expected = Ok(air::Expression::GetField(air::GetField {
            field: "sub".to_string(),
            input: Box::new(air::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
                }
            ))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}
            .into())
            .into(),
            field: "sub".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        from_non_reference_expr_without_nesting,
        expected = Ok(air::Expression::GetField(air::GetField {
            field: "sub".to_string(),
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "f".to_string()
            )))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Literal(mir::LiteralValue::String("f".into()).into()).into(),
            field: "sub".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        dollar_prefixed_field,
        expected = Ok(air::Expression::GetField(air::GetField {
            field: "$sub".to_string(),
            input: Box::new(air::Expression::FieldRef(air::FieldRef {
                name: "f".to_string(),
                parent: None,
            })),
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Reference(("f", 0u16).into()).into(),
            field: "$sub".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
    test_translate_expression!(
        field_contains_dollar,
        expected = Ok(air::Expression::FieldRef(air::FieldRef {
            name: "s$ub".to_string(),
            parent: Some(Box::new(air::FieldRef {
                name: "f".to_string(),
                parent: None
            }))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Reference(("f", 0u16).into()).into(),
            field: "s$ub".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
    test_translate_expression!(
        field_contains_dot,
        expected = Ok(air::Expression::GetField(air::GetField {
            field: "s.ub".to_string(),
            input: Box::new(air::Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "f".to_string()
            }))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Reference(("f", 0u16).into()).into(),
            field: "s.ub".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
    test_translate_expression!(
        empty_field_in_field_access,
        expected = Ok(air::Expression::GetField(air::GetField {
            field: "".to_string(),
            input: Box::new(air::Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "f".to_string()
            }))
        })),
        input = mir::Expression::FieldAccess(mir::FieldAccess {
            expr: mir::Expression::Reference(("f", 0u16).into()).into(),
            field: "".to_string(),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
}

mod simple_case_expression {
    use crate::{air, mir};

    test_translate_expression!(
        integer_simple_case,
        expected = Ok(air::Expression::Let(air::Let {
            vars: vec![air::LetVariable {
                name: "target".to_string(),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(3)))
            }],
            inside: Box::new(air::Expression::Switch(air::Switch {
                branches: vec![
                    air::SwitchCase {
                        case: Box::new(air::Expression::SQLSemanticOperator(
                            air::SQLSemanticOperator {
                                op: air::SQLOperator::Eq,
                                args: vec![
                                    air::Expression::Variable(air::Variable {
                                        parent: None,
                                        name: "target".to_string()
                                    }),
                                    air::Expression::Literal(air::LiteralValue::Integer(4))
                                ]
                            }
                        )),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(5))),
                    },
                    air::SwitchCase {
                        case: Box::new(air::Expression::SQLSemanticOperator(
                            air::SQLSemanticOperator {
                                op: air::SQLOperator::Eq,
                                args: vec![
                                    air::Expression::Variable(air::Variable {
                                        parent: None,
                                        name: "target".to_string()
                                    }),
                                    air::Expression::Literal(air::LiteralValue::Integer(5))
                                ]
                            }
                        )),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(6))),
                    },
                ],
                default: Box::new(air::Expression::Literal(air::LiteralValue::Integer(7)))
            }))
        })),
        input = mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(3).into()
            )),
            when_branch: vec![
                mir::WhenBranch {
                    when: Box::new(mir::Expression::Literal(
                        mir::LiteralValue::Integer(4).into()
                    )),
                    then: Box::new(mir::Expression::Literal(
                        mir::LiteralValue::Integer(5).into()
                    )),
                },
                mir::WhenBranch {
                    when: Box::new(mir::Expression::Literal(
                        mir::LiteralValue::Integer(5).into()
                    )),
                    then: Box::new(mir::Expression::Literal(
                        mir::LiteralValue::Integer(6).into()
                    )),
                },
            ],
            else_branch: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(7).into()
            )),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod searched_case_expression {
    use crate::{air, mir};

    test_translate_expression!(
        integer_searched_case,
        expected = Ok(air::Expression::Switch(air::Switch {
            branches: vec![
                air::SwitchCase {
                    case: Box::new(air::Expression::SQLSemanticOperator(
                        air::SQLSemanticOperator {
                            op: air::SQLOperator::Eq,
                            args: vec![
                                air::Expression::Literal(air::LiteralValue::Integer(4)),
                                air::Expression::Literal(air::LiteralValue::Integer(4))
                            ]
                        }
                    )),
                    then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(5))),
                },
                air::SwitchCase {
                    case: Box::new(air::Expression::SQLSemanticOperator(
                        air::SQLSemanticOperator {
                            op: air::SQLOperator::Eq,
                            args: vec![
                                air::Expression::Literal(air::LiteralValue::Integer(5)),
                                air::Expression::Literal(air::LiteralValue::Integer(5))
                            ]
                        }
                    )),
                    then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(6))),
                },
            ],
            default: Box::new(air::Expression::Literal(air::LiteralValue::Integer(7)))
        })),
        input = mir::Expression::SearchedCase(mir::SearchedCaseExpr {
            when_branch: vec![
                mir::WhenBranch {
                    when: Box::new(mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Eq,
                            args: vec![
                                mir::Expression::Literal(mir::LiteralExpr {
                                    value: mir::LiteralValue::Integer(4),
                                    cache: mir::schema::SchemaCache::new()
                                }),
                                mir::Expression::Literal(mir::LiteralExpr {
                                    value: mir::LiteralValue::Integer(4),
                                    cache: mir::schema::SchemaCache::new()
                                })
                            ],
                            cache: mir::schema::SchemaCache::new(),
                        }
                    )),
                    then: Box::new(mir::Expression::Literal(
                        mir::LiteralValue::Integer(5).into()
                    )),
                },
                mir::WhenBranch {
                    when: Box::new(mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Eq,
                            args: vec![
                                mir::Expression::Literal(mir::LiteralExpr {
                                    value: mir::LiteralValue::Integer(5),
                                    cache: mir::schema::SchemaCache::new()
                                }),
                                mir::Expression::Literal(mir::LiteralExpr {
                                    value: mir::LiteralValue::Integer(5),
                                    cache: mir::schema::SchemaCache::new()
                                })
                            ],
                            cache: mir::schema::SchemaCache::new(),
                        }
                    )),
                    then: Box::new(mir::Expression::Literal(
                        mir::LiteralValue::Integer(6).into()
                    )),
                },
            ],
            else_branch: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(7).into()
            )),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod set_stage {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::UnionWith(air::UnionWith {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "foo".into(),
                    collection: "a".into(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "a".to_string() => translator::ROOT.clone(),
                },
            })),
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "bar".into(),
                    collection: "b".into(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "b".to_string() => translator::ROOT.clone(),
                },
            })),
        })),
        input = mir::Stage::Set(mir::Set {
            operation: mir::SetOperation::UnionAll,
            left: mir::Stage::Collection(mir::Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            right: mir::Stage::Collection(mir::Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod subquery_expression {
    use crate::{
        air, map,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir::{self, binding_tuple::DatasourceName::Bottom},
        unchecked_unique_linked_hash_map,
    };

    test_translate_expression!(
        uncorrelated,
        expected = Ok(air::Expression::Subquery(air::Subquery {
            let_bindings: vec![],
            output_path: vec!["foo".to_string()],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                },
            })),
        })),
        input = mir::Expression::Subquery(mir::SubqueryExpr {
            output_expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test".to_string(),
                collection: "foo".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        correlated,
        expected = Ok(air::Expression::Subquery(air::Subquery {
            let_bindings: vec![air::LetVariable {
                name: "vfoo_0".to_string(),
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "foo".to_string(),
                })),
            },],
            output_path: vec!["__bot".to_string(), "a".to_string()],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "bar".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                    },
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Variable(air::Variable {
                            parent: Some(Box::new(air::Variable {
                                parent: None,
                                name: "vfoo_0".to_string(),
                            })),
                            name: "a".to_string()
                        })
                    }),
                },
            })),
        })),
        input = mir::Expression::Subquery(mir::SubqueryExpr {
            output_expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference((Bottom, 1u16).into())),
                field: "a".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })),
            subquery: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test".to_string(),
                    collection: "bar".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                                field: "a".to_string(),
                                cache: mir::schema::SchemaCache::new(),
                            }),
                        },
                        cache: mir::schema::SchemaCache::new(),
                    })
                },
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("foo", 0u16),
                MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );

    test_translate_stage!(
        unqualified_correlated_reference,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "foo".to_string(),
                        collection: "schema_coll".to_string(),
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "schema_coll".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                    },
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "q".to_string() => air::Expression::FieldRef(air::FieldRef{parent: None, name: "schema_coll".to_string()}),
                },
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "bar".to_string() => air::Expression::Subquery(air::Subquery {
                        let_bindings: vec![air::LetVariable {
                            name: "vq_0".to_string(),
                            expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                                parent: None,
                                name: "q".to_string(),
                            })),
                        },],
                        output_path: vec!["__bot".to_string(), "bar".to_string()],
                        pipeline: Box::new(air::Stage::Limit(air::Limit {
                            source: Box::new(air::Stage::Project(air::Project {
                                source: Box::new(air::Stage::Project(air::Project {
                                    source: Box::new(air::Stage::Project(air::Project {
                                        source: Box::new(air::Stage::Collection(air::Collection {
                                            db: "foo".to_string(),
                                            collection: "schema_foo".to_string(),
                                        })),
                                        specifications: unchecked_unique_linked_hash_map! {
                                            "schema_foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                                        },
                                    })),
                                    specifications: unchecked_unique_linked_hash_map! {
                                        "q".to_string() => air::Expression::FieldRef(air::FieldRef {parent: None, name: "schema_foo".to_string()}),
                                    },
                                })),
                                specifications: unchecked_unique_linked_hash_map! {
                                    "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map! {
                                        "bar".to_string() => air::Expression::Variable(air::Variable {
                                            parent: Some(Box::new(air::Variable {
                                                parent: None,
                                                name: "vq_0".to_string(),
                                            })),
                                            name: "bar".to_string()
                                        })
                                    }),
                                },
                            })),
                            limit: 1,
                        })),
                    })
                }),
            },
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "foo".to_string(),
                    collection: "schema_coll".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    ("q".to_string(), 0u16).into() => mir::Expression::Reference(("schema_coll".to_string(), 0u16).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: map! {
                (Bottom, 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "bar".to_string() => mir::Expression::Subquery(mir::SubqueryExpr {
                            output_expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference((Bottom, 1u16).into())),
                                field: "bar".to_string(),
                                cache: mir::schema::SchemaCache::new(),
                            })),
                            subquery: Box::new(mir::Stage::Limit(mir::Limit {
                                source: Box::new(mir::Stage::Project(mir::Project {
                                    source: Box::new(mir::Stage::Project(mir::Project {
                                        source: Box::new(mir::Stage::Collection(mir::Collection {
                                            db: "foo".to_string(),
                                            collection: "schema_foo".to_string(),
                                            cache: mir::schema::SchemaCache::new(),
                                        })),
                                        expression: map! {
                                            ("q", 1u16).into() => mir::Expression::Reference(("schema_foo", 1u16).into()),
                                        },
                                        cache: mir::schema::SchemaCache::new(),
                                    })),
                                    expression: map! {
                                        (Bottom, 1u16).into() => mir::Expression::Document(mir::DocumentExpr {
                                            document: unchecked_unique_linked_hash_map! {
                                                "bar".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                                                    expr: Box::new(mir::Expression::Reference(("q", 0u16).into())),
                                                    field: "bar".to_string(),
                                                    cache: mir::schema::SchemaCache::new(),
                                                }),
                                            },
                                            cache: mir::schema::SchemaCache::new(),
                                        })
                                    },
                                    cache: mir::schema::SchemaCache::new(),
                                })),
                                limit: 1,
                                cache: mir::schema::SchemaCache::new(),
                            })),
                            cache: mir::schema::SchemaCache::new(),
                        }),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })
            },
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_expression!(
        invalid_output_path,
        expected = Err(crate::translator::Error::SubqueryOutputPathNotFieldRef),
        input = mir::Expression::Subquery(mir::SubqueryExpr {
            output_expr: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                value: mir::LiteralValue::Integer(1),
                cache: mir::schema::SchemaCache::new(),
            })),
            subquery: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test".to_string(),
                collection: "foo".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod documents_stage {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        non_empty,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Documents(air::Documents {
                array: vec![air::Expression::Literal(air::LiteralValue::Boolean(false))],
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
            },
        })),
        input = mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Literal(
                mir::LiteralValue::Boolean(false).into()
            )],
            alias: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        empty,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Documents(air::Documents { array: vec![] })),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => translator::ROOT.clone(),
            },
        })),
        input = mir::Stage::Array(mir::ArraySource {
            array: vec![],
            alias: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod filter_stage {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        basic,
        expected = Ok(air::Stage::Match(air::Match {
            source: air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Documents(air::Documents { array: vec![] })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                },
            })
            .into(),
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
        })),
        input = mir::Stage::Filter(mir::Filter {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: mir::schema::SchemaCache::new()
            })),
            condition: mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod limit_stage {
    use crate::{translator, unchecked_unique_linked_hash_map};
    use translator::Error;

    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::Limit(air::Limit {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".into(),
                    collection: "col".into(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "col".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                },
            })),
            limit: 1i64,
        })),
        input = mir::Stage::Limit(mir::Limit {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "col".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            limit: 1,
            cache: mir::schema::SchemaCache::new(),
        })
    );
    test_translate_stage!(
        out_of_i64_range,
        expected = Err(Error::LimitOutOfI64Range(u64::MAX)),
        input = mir::Stage::Limit(mir::Limit {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "col".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            limit: u64::MAX,
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod collection {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        collection,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => translator::ROOT.clone(),
            },
        })),
        input = mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod projection_stage {
    use crate::{map, unchecked_unique_linked_hash_map};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        project,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::Expression::FieldRef(
                    air::FieldRef {
                        parent: None,
                        name: "foo".to_string()
                    }),
                "bar".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
            }
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_with_user_bot_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "___bot".to_string() => air::Expression::FieldRef(
                    air::FieldRef {
                        parent: None,
                        name: "foo".to_string()
                    }),
                // reordered because BindingTuple uses BTreeMap
                "____bot".to_string() => air::Expression::Literal(air::LiteralValue::Integer(4)),
                "__bot".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2)),
                "_bot".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
            }
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("__bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                Key::named("_bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                Key::named("____bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(4).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod unwind_stage {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage! {
        unwind,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
            path: Box::new(air::Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "foo".to_string()
            })),
            index: None,
            outer: false,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(mir::Stage::Collection(mir::Collection{
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            path: Box::new(mir::Expression::Reference(("foo",0u16).into())),
            index: None,
            outer: false,
            cache: mir::schema::SchemaCache::new(),
        })
    }

    test_translate_stage! {
        unwind_outer,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
            path: Box::new(air::Expression::FieldRef(air::FieldRef {
                parent: Some(Box::new(air::FieldRef {
                    parent: None,
                    name: "foo".to_string()
                })),
                name: "bar".to_string()
            })),
            index: None,
            outer: true,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(mir::Stage::Collection(mir::Collection{
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            path: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: mir::Expression::Reference(("foo",0u16).into()).into(),
                field: "bar".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            index: None,
            outer: true,
            cache: mir::schema::SchemaCache::new(),
        })
    }
    test_translate_stage! {
        unwind_index,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
            path: Box::new(air::Expression::FieldRef(air::FieldRef {
                parent: Some(Box::new(air::FieldRef {
                    parent: None,
                    name: "foo".to_string()
                })),
                name: "bar".to_string()
            })),
            index: Some("i".to_string()),
            outer: true,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(mir::Stage::Collection(mir::Collection{
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            path: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: mir::Expression::Reference(("foo",0u16).into()).into(),
                field: "bar".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            index: Some("i".into()),
            outer: true,
            cache: mir::schema::SchemaCache::new(),
        })
    }
}

mod group_stage {
    use crate::{translator::Error, unchecked_unique_linked_hash_map};
    use mongosql_datastructures::binding_tuple::Key;

    test_translate_stage!(
        group_count_star,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: air::Stage::Project(air::Project {
                    source: air::Stage::Collection(air::Collection {
                        db: "test_db".into(),
                        collection: "foo".into()
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                    }
                }).into(),
                keys: vec![
                    air::NameExprPair {
                        name: "x_key".into(),
                        expr: air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo".to_string()
                        })
                    },
                ],
                aggregations: vec![
                    // Count(*) is traslated as Count(1).
                    air::AccumulatorExpr {
                        alias: "c_distinct".into(),
                        function: air::AggregationFunction::Count,
                        distinct: true,
                        arg: air::Expression::Literal(air::LiteralValue::Integer(1)).into(),
                    },
                    air::AccumulatorExpr {
                        alias: "c_nondistinct".into(),
                        function: air::AggregationFunction::Count,
                        distinct: false,
                        arg: air::Expression::Literal(air::LiteralValue::Integer(1)).into(),
                    },
                ]
            }).into(),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "x_key".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: Some(air::FieldRef { parent: None, name: "_id".into() }.into()),
                        name: "x_key".into()
                    }),
                    "c_distinct".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "c_distinct".into()
                    }),
                    "c_nondistinct".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "c_nondistinct".into()
                    })
                }),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "x_key".into(),
                    expr: mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    })
                }),
            ],
            aggregations: vec![
                mir::AliasedAggregation {
                    alias: "c_distinct".into(),
                    agg_expr: mir::AggregationExpr::CountStar(true),
                },
                mir::AliasedAggregation {
                    alias: "c_nondistinct".into(),
                    agg_expr: mir::AggregationExpr::CountStar(false),
                },
            ],
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        group_normal_operators,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: air::Stage::Project(air::Project {
                    source: air::Stage::Collection(air::Collection {
                        db: "test_db".into(),
                        collection: "foo".into()
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                    }
                }).into(),
                keys: vec![
                    air::NameExprPair {
                        name: "x_key".into(),
                        expr: air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo".to_string()
                        })
                    },
                ],
                aggregations: vec![
                    air::AccumulatorExpr {
                        alias: "max_distinct".into(),
                        function: air::AggregationFunction::Max,
                        distinct: true,
                        arg: air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo".into(),
                        })
                        .into()
                    },
                    air::AccumulatorExpr {
                        alias: "min_nondistinct".into(),
                        function: air::AggregationFunction::Min,
                        distinct: false,
                        arg: air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo".into(),
                        })
                        .into()
                    }
                ]
            }).into(),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "x_key".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: Some(air::FieldRef { parent: None, name: "_id".into() }.into()),
                        name: "x_key".into()
                    }),
                    "max_distinct".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "max_distinct".into()
                    }),
                    "min_nondistinct".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "min_nondistinct".into()
                    })
                }),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "x_key".into(),
                    expr: mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    })
                }),
            ],
            aggregations: vec![
                mir::AliasedAggregation {
                    alias: "max_distinct".into(),
                    agg_expr: mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
                        function: mir::AggregationFunction::Max,
                        distinct: true,
                        arg: mir::Expression::Reference(mir::ReferenceExpr {
                            key: Key::named("foo", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })
                        .into(),
                    }),
                },
                mir::AliasedAggregation {
                    alias: "min_nondistinct".into(),
                    agg_expr: mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
                        function: mir::AggregationFunction::Min,
                        distinct: false,
                        arg: mir::Expression::Reference(mir::ReferenceExpr {
                            key: Key::named("foo", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })
                        .into(),
                    }),
                },
            ],
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        group_key_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: air::Stage::Project(air::Project {
                    source: air::Stage::Collection(air::Collection {
                        db: "test_db".into(),
                        collection: "foo".into()
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                    }
                }).into(),
                keys: vec![
                    air::NameExprPair {
                        name: "__unaliasedKey2".into(),
                        expr: air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo".to_string()
                        })
                    },
                    air::NameExprPair {
                        name: "___unaliasedKey2".into(),
                        expr: air::Expression::FieldRef(air::FieldRef {
                            parent: Some(air::FieldRef {
                                parent: None,
                                name: "foo".into(),
                            }.into()),
                            name: "x".into(),
                        })
                    },
                ],
                aggregations: vec![]
            }).into(),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "x".to_string() => air::Expression::FieldRef(
                        air::FieldRef {
                            parent: Some(air::FieldRef { parent: None, name: "_id".into() }.into()),
                            name: "___unaliasedKey2".to_string()
                        }
                    )
                }),
                "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "__unaliasedKey2".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: Some(air::FieldRef { parent: None, name: "_id".into() }.into()),
                        name: "__unaliasedKey2".into()
                    })
                }),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__unaliasedKey2".into(),
                    expr: mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    })
                }),
                mir::OptionallyAliasedExpr::Unaliased(mir::Expression::FieldAccess(
                    mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: Key::named("foo", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                        field: "x".into(),
                        cache: mir::schema::SchemaCache::new(),
                    }
                )),
            ],
            aggregations: vec![],
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        aggregation_alias_id_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: air::Stage::Project(air::Project {
                    source: air::Stage::Collection(air::Collection {
                        db: "test_db".into(),
                        collection: "foo".into()
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                    }
                }).into(),
                keys: vec![
                    air::NameExprPair {
                        name: "__unaliasedKey2".into(),
                        expr: air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "foo".to_string()
                        })
                    },
                    air::NameExprPair {
                        name: "___unaliasedKey2".into(),
                        expr: air::Expression::FieldRef(air::FieldRef {
                            parent: Some(air::FieldRef {
                                parent: None,
                                name: "foo".into(),
                            }.into()),
                            name: "x".into(),
                        })
                    },
                ],
                aggregations: vec![
                    air::AccumulatorExpr {
                        alias: "__id".into(),
                        function: air::AggregationFunction::Count,
                        distinct: false,
                        arg: air::Expression::Literal(air::LiteralValue::Integer(1)).into(),
                    },
                ]
            }).into(),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "x".to_string() => air::Expression::FieldRef(
                        air::FieldRef {
                            parent: Some(air::FieldRef { parent: None, name: "_id".into() }.into()),
                            name: "___unaliasedKey2".to_string()
                        }
                    )
                }),
                "__bot".to_string() => air::Expression::Document(unchecked_unique_linked_hash_map!{
                    "__unaliasedKey2".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: Some(air::FieldRef { parent: None, name: "_id".into() }.into()),
                        name: "__unaliasedKey2".into()
                    }),
                    "_id".to_string() => air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "__id".into()
                    })
                }),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__unaliasedKey2".into(),
                    expr: mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    })
                }),
                mir::OptionallyAliasedExpr::Unaliased(mir::Expression::FieldAccess(
                    mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: Key::named("foo", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                        field: "x".into(),
                        cache: mir::schema::SchemaCache::new(),
                    }
                )),
            ],
            aggregations: vec![
                mir::AliasedAggregation {
                    alias: "_id".into(),
                    agg_expr: mir::AggregationExpr::CountStar(false),
                },
            ],
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        unaliased_group_key_with_no_datasource_is_error,
        expected = Err(Error::InvalidGroupKey),
        input = mir::Stage::Group(mir::Group {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![mir::OptionallyAliasedExpr::Unaliased(
                mir::Expression::Reference(mir::ReferenceExpr {
                    key: Key::named("foo", 0u16),
                    cache: mir::schema::SchemaCache::new(),
                })
            ),],
            aggregations: vec![],
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod join_stage {
    use crate::{air, mir, unchecked_unique_linked_hash_map};

    fn input_collection(collection_name: &str) -> Box<mir::Stage> {
        Box::new(mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: collection_name.into(),
            cache: mir::schema::SchemaCache::new(),
        }))
    }

    fn transformed_collection(collection_name: &str) -> Box<air::Stage> {
        Box::new(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".into(),
                collection: collection_name.into(),
            })),
            specifications: unchecked_unique_linked_hash_map! {
                collection_name.to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
            },
        }))
    }

    test_translate_stage!(
        join_without_condition,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Inner,
            left: transformed_collection("foo"),
            right: transformed_collection("bar"),
            let_vars: None,
            condition: None
        })),
        input = mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Inner,
            left: input_collection("foo"),
            right: input_collection("bar"),
            condition: None,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        join_with_condition,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Left,
            left: transformed_collection("foo"),
            right: transformed_collection("bar"),
            let_vars: Some(vec![air::LetVariable {
                name: "vfoo_0".to_string(),
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "foo".to_string()
                })),
            }]),
            condition: Some(air::Expression::SQLSemanticOperator(
                air::SQLSemanticOperator {
                    op: air::SQLOperator::Eq,
                    args: vec![
                        air::Expression::Variable(air::Variable {
                            parent: None,
                            name: "vfoo_0".to_string()
                        }),
                        air::Expression::FieldRef(air::FieldRef {
                            parent: None,
                            name: "bar".to_string()
                        }),
                    ]
                }
            ))
        })),
        input = mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Left,
            left: input_collection("foo"),
            right: input_collection("bar"),
            condition: Some(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Eq,
                    args: vec![
                        mir::Expression::Reference(("foo", 0u16).into()),
                        mir::Expression::Reference(("bar", 0u16).into())
                    ],
                    cache: mir::schema::SchemaCache::new(),
                }
            )),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        let_binding_name_conflict_appends_underscores_for_uniqueness,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Inner,
            left: Box::new(air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: transformed_collection("Foo"),
                right: transformed_collection("foo"),
                condition: None,
                let_vars: None,
            })),
            right: transformed_collection("bar"),
            let_vars: Some(vec![
                air::LetVariable {
                    name: "vfoo_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "Foo".to_string()
                    })),
                },
                air::LetVariable {
                    name: "vfoo_0_".to_string(),
                    expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                        parent: None,
                        name: "foo".to_string()
                    })),
                }
            ]),
            condition: Some(air::Expression::Literal(air::LiteralValue::Boolean(true))),
        })),
        input = mir::Stage::Join(mir::Join {
            condition: Some(mir::Expression::Literal(
                mir::LiteralValue::Boolean(true).into()
            )),
            left: mir::Stage::Join(mir::Join {
                condition: None,
                left: mir::Stage::Collection(mir::Collection {
                    db: "test_db".to_string(),
                    collection: "Foo".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })
                .into(),
                right: mir::Stage::Collection(mir::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })
                .into(),
                join_type: mir::JoinType::Inner,
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            right: mir::Stage::Collection(mir::Collection {
                db: "test_db".to_string(),
                collection: "bar".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            join_type: mir::JoinType::Inner,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        test_translate_array,
        expected = Ok(air::Stage::Join(air::Join {
            condition: None,
            left: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                )
            })),
            right: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Documents(air::Documents {
                    array: vec![
                        air::Expression::Literal(air::LiteralValue::Integer(1)),
                        air::Expression::Literal(air::LiteralValue::Integer(1))
                    ]
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "arr".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                )
            })),
            let_vars: None,
            join_type: air::JoinType::Left,
        })),
        input = mir::Stage::Join(mir::Join {
            condition: None,
            left: mir::Stage::Collection(mir::Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            right: mir::Stage::Array(mir::ArraySource {
                array: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1).into())
                ],
                alias: "arr".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            join_type: mir::JoinType::Left,
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod sort_stage {
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    use crate::{map, unchecked_unique_linked_hash_map};

    test_translate_stage!(
        sort_stage_multi_key_reference,
        expected = Ok(air::Stage::Sort(air::Sort {
                source: air::Stage::Project(air::Project {
                    source: air::Stage::Project(air::Project {
                        source: air::Stage::Collection(air::Collection {
                            db: "test_db".to_string(),
                            collection: "foo".to_string(),
                        }).into(),
                        specifications: unchecked_unique_linked_hash_map!{
                            "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                        }
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map!{
                        "__bot".to_string() => air::Expression::FieldRef(air::FieldRef{ parent: None, name: "foo".to_string()}),
                        "bar".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                        "baz".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2)),
                    }
                }).into(),
           specs: vec![
                air::SortSpecification::Desc("bar".to_string()),
                air::SortSpecification::Asc("baz".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                    Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                    Key::named("baz", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                }),
            cache: mir::schema::SchemaCache::new(),
        }).into(),
            specs: vec![
                mir::SortSpecification::Desc(
                    mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("bar", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    })
                    .into()
                ),
                mir::SortSpecification::Asc(
                    mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("baz", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    })
                    .into()
                )
            ],
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        sort_stage_multi_key_field_access,
        expected = Ok(air::Stage::Sort(air::Sort {
            source: air::Stage::Project(air::Project {
                source: air::Stage::Collection(air::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                })
                .into(),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })
            .into(),

            specs: vec![
                air::SortSpecification::Desc("foo.bar".to_string()),
                air::SortSpecification::Asc("foo.baz".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            specs: vec![
                mir::SortSpecification::Desc(
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: mir::Expression::Reference(("foo", 0u16).into()).into(),
                        field: "bar".into(),
                        cache: mir::schema::SchemaCache::new()
                    })
                    .into()
                ),
                mir::SortSpecification::Asc(
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: mir::Expression::Reference(("foo", 0u16).into()).into(),
                        field: "baz".into(),
                        cache: mir::schema::SchemaCache::new()
                    })
                    .into()
                )
            ],
            cache: mir::schema::SchemaCache::new()
        })
    );

    test_translate_stage!(
        sort_stage_nested_key,
        expected = Ok(air::Stage::Sort(air::Sort {
            source: air::Stage::Project(air::Project {
                source: air::Stage::Collection(air::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                })
                .into(),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })
            .into(),

            specs: vec![
                air::SortSpecification::Desc("foo.bar.quz".to_string()),
                air::SortSpecification::Asc("foo.baz.fizzle.bazzle".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            specs: vec![
                mir::SortSpecification::Desc(
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: mir::Expression::Reference(("foo", 0u16).into()).into(),
                            field: "bar".into(),
                            cache: mir::schema::SchemaCache::new()
                        })
                        .into(),
                        field: "quz".into(),
                        cache: mir::schema::SchemaCache::new()
                    })
                    .into(),
                ),
                mir::SortSpecification::Asc(
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: mir::Expression::Reference(("foo", 0u16).into()).into(),
                                field: "baz".into(),
                                cache: mir::schema::SchemaCache::new(),
                            })
                            .into(),
                            field: "fizzle".into(),
                            cache: mir::schema::SchemaCache::new(),
                        })
                        .into(),
                        field: "bazzle".into(),
                        cache: mir::schema::SchemaCache::new(),
                    })
                    .into()
                )
            ],
            cache: mir::schema::SchemaCache::new()
        })
    );
}

mod offset_stage {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::Skip(air::Skip {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
            skip: 10,
        })),
        input = mir::Stage::Offset(mir::Offset {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })),
            offset: 10,
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod translate_plan {
    use crate::{map, unchecked_unique_linked_hash_map};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_plan!(
        project_with_user_bot_conflict,
        expected = Ok(
            air::Stage::ReplaceWith(air::ReplaceWith {
                source: air::Stage::Project(air::Project {
                    source: air::Stage::Project(air::Project {
                        source: air::Stage::Collection(air::Collection {
                            db: "test_db".to_string(),
                            collection: "foo".to_string(),
                        }).into(),
                        specifications: unchecked_unique_linked_hash_map!{
                            "foo".to_string() => air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}),
                        }
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map!{
                        "___bot".to_string() => air::Expression::FieldRef(air::FieldRef{ parent: None, name: "foo".to_string()}),
                        "____bot".to_string() => air::Expression::Literal(air::LiteralValue::Integer(4)),
                        "__bot".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2)),
                        "_bot".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                    }
                }).into(),
                new_root: air::Expression::UnsetField(air::UnsetField {
                    field: "___bot".to_string(),
                    input: air::Expression::SetField(air::SetField {
                        field: "".to_string(),
                        input: air::Expression::Variable(air::Variable{parent: None, name: "ROOT".to_string()}).into(),
                        value: air::Expression::FieldRef(air::FieldRef { parent: None, name: "___bot".to_string() }).into(),
                    }).into()
                }).into()
            })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("__bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                Key::named("_bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                Key::named("____bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(4).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod translate_is {
    use crate::{air, mir};

    test_translate_expression!(
        is_number,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Number,
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Number,
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_missing,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Missing,
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Missing,
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_array,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Array),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Array),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_bindata,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::BinData),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::BinData),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_bool,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Boolean),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Boolean),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_datetime,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Datetime),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Datetime),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_dbpointer,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::DbPointer),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::DbPointer),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_decimal128,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Decimal128),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Decimal128),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_document,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Document),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Document),
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        is_type_double,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Double),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Double),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_int32,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Int32),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_int64,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Int64),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Int64),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_javascript,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Javascript),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Javascript),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_javascript_with_scope,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::JavascriptWithScope),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::JavascriptWithScope),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_max_key,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::MaxKey),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::MaxKey),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_min_key,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::MinKey),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::MinKey),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_null,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Null),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Null),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_object_id,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::ObjectId),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::ObjectId),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_regular_expression,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::RegularExpression),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::RegularExpression),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_string,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::String),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::String),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_symbol,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Symbol),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Symbol),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_timestamp,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Timestamp),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Timestamp),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        is_type_undefined,
        expected = Ok(air::Expression::Is(air::Is {
            expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            target_type: air::TypeOrMissing::Type(air::Type::Undefined),
        })),
        input = mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(42).into()
            )),
            target_type: mir::TypeOrMissing::Type(mir::Type::Undefined),
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}
