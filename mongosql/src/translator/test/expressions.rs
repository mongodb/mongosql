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
                is_join: false,
            };
            let expected = $expected;
            let actual = translator.translate_expression(input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_translate_expression_with_schema_info {
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(mapping_registry = $mapping_registry:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{translator, mapping_registry::MqlMappingRegistry, catalog::Catalog, mir::schema::{CachedSchema, SchemaInferenceState}};

            // force the input
            let input = $input;
            #[allow(unused_mut, unused_assignments)]
            let mut mapping_registry = MqlMappingRegistry::default();
            $(mapping_registry = $mapping_registry;)?

            let catalog = Catalog::default();
            let schema_inference_state = SchemaInferenceState::empty(&catalog);
            let _ = input.schema(&schema_inference_state).unwrap();
            let translator = translator::MqlTranslator{
                mapping_registry,
                scope_level: 0u16,
                is_join: false,
            };
            let expected = $expected;
            let actual = translator.translate_expression(input);
            assert_eq!(expected, actual);
        }
    };
}

mod literal {
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

mod reference {
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
        expected = Ok(air::Expression::FieldRef("f".to_string().into())),
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
        expected = Ok(air::Expression::Variable("f".to_string().into())),
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

mod array {
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

mod document {
    use crate::unchecked_unique_linked_hash_map;
    use crate::{air, mapping_registry::*, mir, translator::Error};

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
        dollar_prefixed_key_becomes_set_field,
        expected = Ok(air::Expression::SetField(air::SetField {
            field: "$foo".to_string(),
            input: Box::new(air::Expression::Document(unchecked_unique_linked_hash_map!{})),
            value: Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
        })),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"$foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}.into()),
    );
    test_translate_expression!(
        key_containing_dot_becomes_set_field,
        expected = Ok(air::Expression::SetField(air::SetField {
            field: "foo.bar".to_string(),
            input: Box::new(air::Expression::Document(unchecked_unique_linked_hash_map!{})),
            value: Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
        })),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo.bar".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}.into(),
        ),
    );
    test_translate_expression!(
        multiple_keys_become_set_field_if_at_least_one_requires,
        expected = Ok(air::Expression::SetField(air::SetField {
            field: "y".to_string(),
            input: Box::new(air::Expression::SetField(air::SetField {
                field: "$foo".to_string(),
                input: Box::new(air::Expression::SetField(air::SetField {
                    field: "x".to_string(),
                    input: Box::new(air::Expression::SetField(air::SetField {
                        field: "foo.bar".to_string(),
                        input: Box::new(air::Expression::Document(unchecked_unique_linked_hash_map!{})),
                        value: Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    })),
                    value: Box::new(air::Expression::FieldRef("f.x".to_string().into())),
                })),
                value: Box::new(air::Expression::Literal(air::LiteralValue::Integer(2))),
            })),
            value: Box::new(air::Expression::Literal(air::LiteralValue::Integer(3))),
        })),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "foo.bar".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                "x".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("f", 0u16).into())),
                    field: "x".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                }),
                "$foo".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                "y".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(3).into()),
            }.into(),
        ),
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
        empty_key_disallowed,
        expected = Err(Error::InvalidDocumentKey("".to_string())),
        input = mir::Expression::Document(
            unchecked_unique_linked_hash_map! {"".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1).into())}.into()),
    );
}

mod date_function {
    use crate::{air, mir};

    test_translate_expression!(
        dateadd,
        expected = Ok(air::Expression::DateFunction(
            air::DateFunctionApplication {
                function: air::DateFunction::Add,
                unit: air::DatePart::Year,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(5)),
                    air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                        op: air::SQLOperator::CurrentTimestamp,
                        args: vec![],
                    }),
                ],
            }
        )),
        input = mir::Expression::DateFunction(mir::DateFunctionApplication {
            function: mir::DateFunction::Add,
            date_part: mir::DatePart::Year,
            args: vec![
                mir::Expression::Literal(mir::LiteralExpr {
                    value: mir::LiteralValue::Integer(5),
                    cache: mir::schema::SchemaCache::new()
                }),
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: mir::schema::SchemaCache::new(),
                }),
            ],
            cache: mir::schema::SchemaCache::new()
        }),
    );
    test_translate_expression!(
        datediff,
        expected = Ok(air::Expression::DateFunction(
            air::DateFunctionApplication {
                function: air::DateFunction::Diff,
                unit: air::DatePart::Year,
                args: vec![
                    air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                        op: air::SQLOperator::CurrentTimestamp,
                        args: vec![],
                    }),
                    air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                        op: air::SQLOperator::CurrentTimestamp,
                        args: vec![],
                    }),
                    air::Expression::Literal(air::LiteralValue::String("sunday".to_string())),
                ],
            }
        )),
        input = mir::Expression::DateFunction(mir::DateFunctionApplication {
            function: mir::DateFunction::Diff,
            date_part: mir::DatePart::Year,
            args: vec![
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: mir::schema::SchemaCache::new(),
                }),
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: mir::schema::SchemaCache::new(),
                }),
                mir::Expression::Literal(mir::LiteralExpr {
                    value: mir::LiteralValue::String("sunday".to_string()),
                    cache: mir::schema::SchemaCache::new()
                }),
            ],
            cache: mir::schema::SchemaCache::new()
        }),
    );
    test_translate_expression!(
        datetrunc,
        expected = Ok(air::Expression::DateFunction(
            air::DateFunctionApplication {
                function: air::DateFunction::Trunc,
                unit: air::DatePart::Year,
                args: vec![
                    air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                        op: air::SQLOperator::CurrentTimestamp,
                        args: vec![],
                    }),
                    air::Expression::Literal(air::LiteralValue::String("sunday".to_string())),
                ],
            }
        )),
        input = mir::Expression::DateFunction(mir::DateFunctionApplication {
            function: mir::DateFunction::Trunc,
            date_part: mir::DatePart::Year,
            args: vec![
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::CurrentTimestamp,
                    args: vec![],
                    cache: mir::schema::SchemaCache::new(),
                }),
                mir::Expression::Literal(mir::LiteralExpr {
                    value: mir::LiteralValue::String("sunday".to_string()),
                    cache: mir::schema::SchemaCache::new()
                }),
            ],
            cache: mir::schema::SchemaCache::new()
        }),
    );
}

mod scalar_function {
    use crate::{air, mir, unchecked_unique_linked_hash_map};

    test_translate_expression_with_schema_info!(
        concat_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Concat,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::String("world".into())),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Concat,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::String("world".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        concat_nullish,
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

    test_translate_expression_with_schema_info!(
        pos_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Pos,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(19)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pos,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Integer(19).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        pos_nullish,
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

    test_translate_expression_with_schema_info!(
        neg_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Neg,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(32)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Neg,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Integer(32).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        neg_nullish,
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

    test_translate_expression_with_schema_info!(
        add_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Add,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Add,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        add_nullish,
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

    test_translate_expression_with_schema_info!(
        sub_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Subtract,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sub,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );
    test_translate_expression!(
        addition_with_more_than_two_operands,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Add,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(1)),
                    air::Expression::Literal(air::LiteralValue::Integer(2)),
                    air::Expression::Literal(air::LiteralValue::Integer(3)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Add,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(3).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        sub_nullish,
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

    test_translate_expression_with_schema_info!(
        mul_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Multiply,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mul,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        subtraction_with_more_than_2_operands,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Subtract,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(1)),
                    air::Expression::Literal(air::LiteralValue::Integer(2)),
                    air::Expression::Literal(air::LiteralValue::Integer(3)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sub,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(3).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        mul_nullish,
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

    test_translate_expression_with_schema_info!(
        div_no_nullish,
        expected = Ok(air::Expression::SqlDivide(air::SqlDivide {
            dividend: Box::new(air::Expression::Literal(air::LiteralValue::Integer(32))),
            divisor: Box::new(air::Expression::Literal(air::LiteralValue::Integer(20))),
            on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Div,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(20).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        multiplication_with_more_than_2_operands,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Multiply,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(1)),
                    air::Expression::Literal(air::LiteralValue::Integer(2)),
                    air::Expression::Literal(air::LiteralValue::Integer(3)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mul,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(2).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(3).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        div_nullish,
        expected = Ok(air::Expression::SqlDivide(air::SqlDivide {
            dividend: Box::new(air::Expression::Literal(air::LiteralValue::Integer(32))),
            divisor: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Div,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        lt_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Lt,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        lt_nullish,
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

    test_translate_expression_with_schema_info!(
        lte_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Lte,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lte,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        lte_nullish,
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

    test_translate_expression_with_schema_info!(
        neq_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Ne,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Neq,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        neq_nullish,
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

    test_translate_expression_with_schema_info!(
        eq_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Eq,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Eq,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        eq_nullish,
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

    test_translate_expression_with_schema_info!(
        gt_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Gt,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gt,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        gt_nullish,
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

    test_translate_expression_with_schema_info!(
        gte_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Gte,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Gte,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        gte_nullish,
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

    test_translate_expression_with_schema_info!(
        between_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Between,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                    air::Expression::Literal(air::LiteralValue::Integer(32)),
                    air::Expression::Literal(air::LiteralValue::Integer(19)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Between,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(32).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(19).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        between_nullish,
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

    test_translate_expression_with_schema_info!(
        not_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Not,
                args: vec![air::Expression::Literal(air::LiteralValue::Boolean(false)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Not,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Boolean(false).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        not_nullish,
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

    test_translate_expression_with_schema_info!(
        and_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::And,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Boolean(false)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::And,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        and_nullish,
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

    test_translate_expression_with_schema_info!(
        or_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Or,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Boolean(false)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Or,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        or_nullish,
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

    test_translate_expression_with_schema_info!(
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

    test_translate_expression_with_schema_info!(
        null_if_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::NullIf,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Boolean(false)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::NullIf,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        null_if_nullish,
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

    test_translate_expression_with_schema_info!(
        coalesce_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Coalesce,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Boolean(true)),
                    air::Expression::Literal(air::LiteralValue::Boolean(false)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Coalesce,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Boolean(true).into()),
                mir::Expression::Literal(mir::LiteralValue::Boolean(false).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        coalesce_nullish,
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

    test_translate_expression_with_schema_info!(
        slice_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Slice,
                args: vec![
                    air::Expression::Array(vec![air::Expression::Literal(
                        air::LiteralValue::String("abc".to_string())
                    )]),
                    air::Expression::Literal(air::LiteralValue::Integer(0)),
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
                mir::Expression::Literal(mir::LiteralValue::Integer(0).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        slice_nullish,
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

    test_translate_expression_with_schema_info!(
        size_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Size,
                args: vec![air::Expression::Array(vec![air::Expression::Literal(
                    air::LiteralValue::String("abc".to_string())
                )]),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Size,
            args: vec![mir::Expression::Array(
                vec![mir::Expression::Literal(
                    mir::LiteralValue::String("abc".into()).into()
                )]
                .into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        size_nullish,
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

    test_translate_expression_with_schema_info!(
        position_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::IndexOfCP,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("world".into())),
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Position,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::String("world".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        position_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::IndexOfCP,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Null),
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
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

    test_translate_expression_with_schema_info!(
        char_length_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::StrLenCP,
                args: vec![air::Expression::Literal(air::LiteralValue::String(
                    "hello".into()
                )),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::CharLength,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::String("hello".into()).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        char_length_nullish,
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

    test_translate_expression_with_schema_info!(
        octet_length_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::StrLenBytes,
                args: vec![air::Expression::Literal(air::LiteralValue::String(
                    "hello".into()
                )),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::OctetLength,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::String("hello".into()).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        octet_length_nullish,
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

    test_translate_expression_with_schema_info!(
        bit_length_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::BitLength,
                args: vec![air::Expression::Literal(air::LiteralValue::String(
                    "hello".into()
                )),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BitLength,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::String("hello".into()).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        bit_length_nullish,
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

    test_translate_expression_with_schema_info!(
        abs_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Abs,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Abs,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        abs_nullish,
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

    test_translate_expression_with_schema_info!(
        ceil_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Ceil,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Ceil,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        ceil_nullish,
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

    test_translate_expression_with_schema_info!(
        floor_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Floor,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Floor,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        floor_nullish,
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

    test_translate_expression_with_schema_info!(
        log_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Log,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Log,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        log_nullish,
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

    test_translate_expression_with_schema_info!(
        mod_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Mod,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Mod,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        mod_nullish,
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

    test_translate_expression_with_schema_info!(
        pow_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Pow,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Pow,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        pow_nullish,
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

    test_translate_expression_with_schema_info!(
        radians_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::DegreesToRadians,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Radians,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        radians_nullish,
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

    test_translate_expression_with_schema_info!(
        round_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Round,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                    air::Expression::Literal(air::LiteralValue::Double(3.5)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Round,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
                mir::Expression::Literal(mir::LiteralValue::Double(3.5).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        round_nullish,
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

    test_translate_expression_with_schema_info!(
        cos_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Cos,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Cos,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        cos_nullish,
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

    test_translate_expression_with_schema_info!(
        sin_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Sin,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sin,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        sin_nullish,
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

    test_translate_expression_with_schema_info!(
        tan_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Tan,
                args: vec![air::Expression::Literal(air::LiteralValue::Double(3.5)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Tan,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Double(3.5).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        tan_nullish,
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

    test_translate_expression_with_schema_info!(
        substring_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::SubstrCP,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::Integer(1)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Substring,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        substring_nullish,
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

    test_translate_expression_with_schema_info!(
        upper_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::ToUpper,
                args: vec![air::Expression::Literal(air::LiteralValue::String(
                    "hello".into()
                )),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Upper,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::String("hello".into()).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        upper_nullish,
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

    test_translate_expression_with_schema_info!(
        lower_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::ToLower,
                args: vec![air::Expression::Literal(air::LiteralValue::String(
                    "hello".into()
                )),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Lower,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::String("hello".into()).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        lower_nullish,
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

    test_translate_expression_with_schema_info!(
        trim_no_nullish,
        expected = Ok(air::Expression::Trim(air::Trim {
            op: air::TrimOperator::Trim,
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "h".into()
            ))),
            chars: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "hello".into()
            ))),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        trim_nullish,
        expected = Ok(air::Expression::Trim(air::Trim {
            op: air::TrimOperator::Trim,
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "h".into()
            ))),
            chars: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::BTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        ltrim_no_nullish,
        expected = Ok(air::Expression::Trim(air::Trim {
            op: air::TrimOperator::LTrim,
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "h".into()
            ))),
            chars: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "hello".into()
            ))),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::LTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        ltrim_nullish,
        expected = Ok(air::Expression::Trim(air::Trim {
            op: air::TrimOperator::LTrim,
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "h".into()
            ))),
            chars: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::LTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        rtrim_no_nullish,
        expected = Ok(air::Expression::Trim(air::Trim {
            op: air::TrimOperator::RTrim,
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "h".into()
            ))),
            chars: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "hello".into()
            ))),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::RTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        rtrim_nullish,
        expected = Ok(air::Expression::Trim(air::Trim {
            op: air::TrimOperator::RTrim,
            input: Box::new(air::Expression::Literal(air::LiteralValue::String(
                "h".into()
            ))),
            chars: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
        })),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::RTrim,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::Null.into()),
                mir::Expression::Literal(mir::LiteralValue::String("h".into()).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        split_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Split,
                args: vec![
                    air::Expression::Literal(air::LiteralValue::String("hello".into())),
                    air::Expression::Literal(air::LiteralValue::String("l".into())),
                    air::Expression::Literal(air::LiteralValue::Integer(1)),
                ],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Split,
            args: vec![
                mir::Expression::Literal(mir::LiteralValue::String("hello".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::String("l".into()).into()),
                mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
            ],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        split_nullish,
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

    test_translate_expression_with_schema_info!(
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

    test_translate_expression_with_schema_info!(
        year_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Year,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Year,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        year_nullish,
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

    test_translate_expression_with_schema_info!(
        month_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Month,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Month,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        month_nullish,
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

    test_translate_expression_with_schema_info!(
        day_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::DayOfMonth,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Day,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        day_nullish,
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

    test_translate_expression_with_schema_info!(
        hour_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Hour,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Hour,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        hour_nullish,
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

    test_translate_expression_with_schema_info!(
        minute_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Minute,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Minute,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        minute_nullish,
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

    test_translate_expression_with_schema_info!(
        second_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Second,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Second,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        second_nullish,
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

    test_translate_expression_with_schema_info!(
        week_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Week,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Week,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        week_nullish,
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

    test_translate_expression_with_schema_info!(
        day_of_year_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::DayOfYear,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::DayOfYear,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        day_of_year_nullish,
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

    test_translate_expression_with_schema_info!(
        iso_week_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::IsoWeek,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::IsoWeek,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        iso_week_nullish,
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

    test_translate_expression_with_schema_info!(
        iso_week_day_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::IsoDayOfWeek,
                args: vec![air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String(
                        "2012-12-20T12:12:12Z".to_string()
                    ))
                    .into(),
                    to: air::Type::Datetime,
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                }),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::IsoWeekday,
            args: vec![mir::Expression::Cast(mir::CastExpr {
                expr: mir::Expression::Literal(
                    mir::LiteralValue::String("2012-12-20T12:12:12Z".to_string()).into()
                )
                .into(),
                to: mir::Type::Datetime,
                on_error: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                on_null: mir::Expression::Literal(mir::LiteralValue::Null.into()).into(),
                cache: mir::schema::SchemaCache::new(),
            })],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        iso_week_day_nullish,
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

    test_translate_expression_with_schema_info!(
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

    test_translate_expression_with_schema_info!(
        sqrt_nullish,
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

    test_translate_expression_with_schema_info!(
        sqrt_no_nullish,
        expected = Ok(air::Expression::SQLSemanticOperator(
            air::SQLSemanticOperator {
                op: air::SQLOperator::Sqrt,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(4)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Sqrt,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Integer(4).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression_with_schema_info!(
        degrees_nullish,
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

    test_translate_expression_with_schema_info!(
        degrees_no_nullish,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::RadiansToDegrees,
                args: vec![air::Expression::Literal(air::LiteralValue::Integer(30)),],
            }
        )),
        input = mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Degrees,
            args: vec![mir::Expression::Literal(
                mir::LiteralValue::Integer(30).into()
            ),],
            cache: mir::schema::SchemaCache::new(),
        }),
    );
}

mod cast {
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

mod searched_case {
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

mod simple_case {
    use crate::{air, mir};

    test_translate_expression_with_schema_info!(
        integer_simple_case_no_nullish,
        expected = Ok(air::Expression::Let(air::Let {
            vars: vec![air::LetVariable {
                name: "target".to_string(),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(3)))
            }],
            inside: Box::new(air::Expression::Switch(air::Switch {
                branches: vec![
                    air::SwitchCase {
                        case: Box::new(air::Expression::MQLSemanticOperator(
                            air::MQLSemanticOperator {
                                op: air::MQLOperator::Eq,
                                args: vec![
                                    air::Expression::Variable("target".to_string().into()),
                                    air::Expression::Literal(air::LiteralValue::Integer(4))
                                ]
                            }
                        )),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(5))),
                    },
                    air::SwitchCase {
                        case: Box::new(air::Expression::MQLSemanticOperator(
                            air::MQLSemanticOperator {
                                op: air::MQLOperator::Eq,
                                args: vec![
                                    air::Expression::Variable("target".to_string().into()),
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

    test_translate_expression_with_schema_info!(
        integer_simple_case_nullish_expr,
        expected = Ok(air::Expression::Let(air::Let {
            vars: vec![air::LetVariable {
                name: "target".to_string(),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Null))
            }],
            inside: Box::new(air::Expression::Switch(air::Switch {
                branches: vec![
                    air::SwitchCase {
                        case: Box::new(air::Expression::SQLSemanticOperator(
                            air::SQLSemanticOperator {
                                op: air::SQLOperator::Eq,
                                args: vec![
                                    air::Expression::Variable("target".to_string().into()),
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
                                    air::Expression::Variable("target".to_string().into()),
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
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Null.into())),
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

    test_translate_expression_with_schema_info!(
        integer_simple_case_nullish_when,
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
                                    air::Expression::Variable("target".to_string().into()),
                                    air::Expression::Literal(air::LiteralValue::Null)
                                ]
                            }
                        )),
                        then: Box::new(air::Expression::Literal(air::LiteralValue::Integer(5))),
                    },
                    air::SwitchCase {
                        case: Box::new(air::Expression::MQLSemanticOperator(
                            air::MQLSemanticOperator {
                                op: air::MQLOperator::Eq,
                                args: vec![
                                    air::Expression::Variable("target".to_string().into()),
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
                    when: Box::new(mir::Expression::Literal(mir::LiteralValue::Null.into())),
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

    test_translate_expression_with_schema_info!(
        nested_case_expression_translation,
        expected = Ok(air::Expression::Let(air::Let {
            vars: vec![air::LetVariable {
                name: "target".to_string(),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
            },],
            inside: Box::new(air::Expression::Switch(air::Switch {
                branches: vec![air::SwitchCase {
                    case: Box::new(air::Expression::MQLSemanticOperator(
                        air::MQLSemanticOperator {
                            op: air::MQLOperator::Eq,
                            args: vec![
                                air::Expression::Variable(air::Variable {
                                    parent: None,
                                    name: "target".to_string(),
                                }),
                                air::Expression::Let(air::Let {
                                    vars: vec![air::LetVariable {
                                        name: "target".to_string(),
                                        expr: Box::new(air::Expression::Literal(
                                            air::LiteralValue::Integer(2)
                                        )),
                                    },],
                                    inside: Box::new(air::Expression::Switch(air::Switch {
                                        branches: vec![air::SwitchCase {
                                            case: Box::new(air::Expression::MQLSemanticOperator(
                                                air::MQLSemanticOperator {
                                                    op: air::MQLOperator::Eq,
                                                    args: vec![
                                                        air::Expression::Variable(air::Variable {
                                                            parent: None,
                                                            name: "target".to_string(),
                                                        }),
                                                        air::Expression::Literal(
                                                            air::LiteralValue::Integer(5)
                                                        ),
                                                    ],
                                                },
                                            )),
                                            then: Box::new(air::Expression::Literal(
                                                air::LiteralValue::Integer(2)
                                            )),
                                        },],
                                        default: Box::new(air::Expression::Literal(
                                            air::LiteralValue::Integer(1)
                                        )),
                                    })),
                                })
                            ],
                        }
                    )),
                    then: Box::new(air::Expression::Literal(air::LiteralValue::String(
                        "YES".to_string()
                    ))),
                },],
                default: Box::new(air::Expression::Literal(air::LiteralValue::String(
                    "NO".to_string()
                ))),
            })),
        })),
        input = mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                value: mir::LiteralValue::Integer(1),
                cache: mir::schema::SchemaCache::new(),
            })),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
                    expr: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                        value: mir::LiteralValue::Integer(2),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    when_branch: vec![mir::WhenBranch {
                        when: Box::new(mir::Expression::Literal(
                            mir::LiteralValue::Integer(5).into()
                        )),
                        then: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                            value: mir::LiteralValue::Integer(2),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                    }],
                    else_branch: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                        value: mir::LiteralValue::Integer(1),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    cache: mir::schema::SchemaCache::new(),
                })),
                then: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                    value: mir::LiteralValue::String("YES".to_string()),
                    cache: mir::schema::SchemaCache::new(),
                })),
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralExpr {
                value: mir::LiteralValue::String("NO".to_string()),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new()
        }),
    );
}

mod type_assertion {
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

mod is {
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

mod like {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir,
    };
    test_translate_expression!(
        like_expr,
        expected = Ok(air::Expression::Like(air::Like {
            expr: Box::new(air::Expression::FieldRef("input".to_string().into())),
            pattern: Box::new(air::Expression::FieldRef("pattern".to_string().into())),
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

mod field_access {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir, unchecked_unique_linked_hash_map,
    };

    test_translate_expression!(
        from_reference,
        expected = Ok(air::Expression::FieldRef("f.sub".to_string().into())),
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
        expected = Ok(air::Expression::FieldRef("f.sub1.sub2".to_string().into())),
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
            input: Box::new(air::Expression::FieldRef("f".to_string().into())),
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
        expected = Ok(air::Expression::FieldRef("f.s$ub".to_string().into())),
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
            input: Box::new(air::Expression::FieldRef("f".to_string().into()))
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
            input: Box::new(air::Expression::FieldRef("f".to_string().into()))
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

mod subquery {
    use crate::{
        air, map,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir::{self, binding_tuple::DatasourceName::Bottom},
        translator::utils::ROOT,
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
                    "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                },
            })),
        })),
        input = mir::Expression::Subquery(mir::SubqueryExpr {
            output_expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 1u16).into() => mir::Expression::Reference(("foo", 1u16).into()),
                },
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
                expr: Box::new(air::Expression::FieldRef("foo".to_string().into())),
            },],
            output_path: vec!["__bot".to_string(), "a".to_string()],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "bar".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Variable("vfoo_0.a".to_string().into()),
                    })),
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

    test_translate_expression!(
        datasource_names_normalized_and_conflicts_avoided_in_let_bindings,
        expected = Ok(air::Expression::Subquery(air::Subquery {
            let_bindings: vec![
                air::LetVariable {
                    name: "vfoo_coll_ß_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef("Foo coll-ß".to_string().into())),
                },
                air::LetVariable {
                    name: "vfoo_coll_ß_0_".to_string(),
                    expr: Box::new(air::Expression::FieldRef("foo_coll_ß".to_string().into())),
                },
            ],
            output_path: vec!["__bot".to_string(), "a".to_string()],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "bar".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                            op: air::SQLOperator::Eq,
                            args: vec![
                                air::Expression::Variable("vfoo_coll_ß_0.a".to_string().into()),
                                air::Expression::Variable("vfoo_coll_ß_0_.a".to_string().into()),
                            ],
                        }),
                    })),
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
                            "a".to_string() => mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                function: mir::ScalarFunction::Eq,
                                args: vec![
                                    mir::Expression::FieldAccess(mir::FieldAccess{
                                        expr: Box::new(mir::Expression::Reference(("Foo coll-ß", 0u16).into())),
                                        field: "a".into(),
                                        cache: mir::schema::SchemaCache::new(),
                                    }),
                                    mir::Expression::FieldAccess(mir::FieldAccess{
                                        expr: Box::new(mir::Expression::Reference(("foo_coll_ß", 0u16).into())),
                                        field: "a".into(),
                                        cache: mir::schema::SchemaCache::new(),
                                    }),
                                ],
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
                ("Foo coll-ß", 0u16),
                MqlMappingRegistryValue::new("Foo coll-ß".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("foo_coll_ß", 0u16),
                MqlMappingRegistryValue::new("foo_coll_ß".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );

    // This test verifies that we are using the datasource's MQL field name
    // in the output path. We create an MQL field name that doesn't match the
    // corresponding datasource by forcing a naming conflict in the project
    // stage. The translation engine could never actually produce a query like
    // this one, though, since a subquery expression's degree must be exactly 1.
    test_translate_expression!(
        use_datasource_mql_name,
        expected = Ok(air::Expression::Subquery(air::Subquery {
            let_bindings: vec![],
            output_path: vec!["___bot".to_string()],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "__bot".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "___bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(42))),
                }
            })),
        })),
        input = mir::Expression::Subquery(mir::SubqueryExpr {
            output_expr: Box::new(mir::Expression::Reference((Bottom, 1u16).into())),
            subquery: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test".to_string(),
                    collection: "__bot".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => mir::Expression::Reference(("__bot", 1u16).into()),
                    ("__bot", 1u16).into() => mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        output_path_contains_dot,
        expected = Ok(air::Expression::Subquery(air::Subquery {
            let_bindings: vec![],
            output_path: vec!["foo".to_string(), "a.b".to_string()],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                },
            })),
        })),
        input = mir::Expression::Subquery(mir::SubqueryExpr {
            output_expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                field: "a.b".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })),
            subquery: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 1u16).into() => mir::Expression::Reference(("foo", 1u16).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        }),
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

mod subquery_comparison {
    use crate::{
        air, map,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir::{self, binding_tuple::DatasourceName::Bottom},
        translator::utils::ROOT,
        unchecked_unique_linked_hash_map,
    };

    test_translate_expression!(
        uncorrelated,
        expected = Ok(air::Expression::SubqueryComparison(
            air::SubqueryComparison {
                op: air::SubqueryComparisonOp::Eq,
                modifier: air::SubqueryModifier::Any,
                arg: Box::new(air::Expression::Literal(air::LiteralValue::Integer(5))),
                subquery: Box::new(air::Subquery {
                    let_bindings: vec![],
                    output_path: vec!["foo".to_string()],
                    pipeline: Box::new(air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Collection(air::Collection {
                            db: "test".to_string(),
                            collection: "foo".to_string(),
                        })),
                        specifications: unchecked_unique_linked_hash_map! {
                            "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                        },
                    })),
                }),
            }
        )),
        input = mir::Expression::SubqueryComparison(mir::SubqueryComparison {
            operator: mir::SubqueryComparisonOp::Eq,
            modifier: mir::SubqueryModifier::Any,
            argument: Box::new(mir::Expression::Literal(
                mir::LiteralValue::Integer(5).into()
            )),
            subquery_expr: mir::SubqueryExpr {
                output_expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                subquery: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test".to_string(),
                        collection: "foo".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {
                        ("foo", 1u16).into() => mir::Expression::Reference(("foo", 1u16).into()),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })),
                cache: mir::schema::SchemaCache::new(),
            },
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_expression!(
        correlated,
        expected = Ok(air::Expression::SubqueryComparison(
            air::SubqueryComparison {
                op: air::SubqueryComparisonOp::Gt,
                modifier: air::SubqueryModifier::All,
                arg: Box::new(air::Expression::FieldRef("x".to_string().into())),
                subquery: Box::new(air::Subquery {
                    let_bindings: vec![
                        air::LetVariable {
                            name: "vfoo_0".to_string(),
                            expr: Box::new(air::Expression::FieldRef("foo".to_string().into())),
                        },
                        air::LetVariable {
                            name: "vx_0".to_string(),
                            expr: Box::new(air::Expression::FieldRef("x".to_string().into())),
                        },
                    ],
                    output_path: vec!["__bot".to_string(), "a".to_string()],
                    pipeline: Box::new(air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Collection(air::Collection {
                            db: "test".to_string(),
                            collection: "bar".to_string(),
                        })),
                        specifications: unchecked_unique_linked_hash_map! {
                            "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                                "a".to_string() => air::Expression::Variable("vfoo_0.a".to_string().into()),
                            })),
                        },
                    })),
                })
            }
        )),
        input = mir::Expression::SubqueryComparison(mir::SubqueryComparison {
            operator: mir::SubqueryComparisonOp::Gt,
            modifier: mir::SubqueryModifier::All,
            argument: Box::new(mir::Expression::Reference(("x", 0u16).into())),
            subquery_expr: mir::SubqueryExpr {
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
            },
            cache: mir::schema::SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("foo", 0u16),
                MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("x", 0u16),
                MqlMappingRegistryValue::new("x".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
    );
}

mod subquery_exists {
    use crate::{
        air, map,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir::{self, binding_tuple::DatasourceName::Bottom},
        translator::utils::ROOT,
        unchecked_unique_linked_hash_map,
    };

    test_translate_expression!(
        subquery_exists_uncorrelated,
        expected = Ok(air::Expression::SubqueryExists(air::SubqueryExists {
            let_bindings: vec![],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                },
            })),
        })),
        input = mir::Expression::Exists(
            Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 1u16).into() => mir::Expression::Reference(("foo", 1u16).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            }))
            .into()
        ),
    );
    test_translate_expression!(
        subquery_exists_correlated,
        expected = Ok(air::Expression::SubqueryExists(air::SubqueryExists {
            let_bindings: vec![air::LetVariable {
                name: "vfoo_0".to_string(),
                expr: Box::new(air::Expression::FieldRef("foo".to_string().into())),
            },],
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test".to_string(),
                    collection: "bar".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::Expression::Variable("vfoo_0.a".to_string().into()),
                    })),
                },
            })),
        })),
        input = mir::Expression::Exists(Box::new(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test".into(),
                collection: "bar".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: map! {
                (Bottom, 1u16).into() => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".into() => mir::Expression::FieldAccess(mir::FieldAccess{
                        expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: mir::schema::SchemaCache::new(),
                    })
                }.into())
            },
            cache: mir::schema::SchemaCache::new(),
        })).into()),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("foo", 0u16), MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef));
            mr
        },
    );
}

mod optimized_match_exists {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistryValue, MqlReferenceType},
        mir,
    };

    test_translate_expression!(
        basic,
        expected = Ok(air::Expression::MQLSemanticOperator(
            air::MQLSemanticOperator {
                op: air::MQLOperator::Gt,
                args: vec![
                    air::Expression::FieldRef("foo.x".to_string().into()),
                    air::Expression::Literal(air::LiteralValue::Null),
                ],
            }
        )),
        input = mir::Expression::OptimizedMatchExists(mir::OptimizedMatchExists {
            field_access: mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                field: "x".to_string(),
                cache: mir::schema::SchemaCache::new(),
            },
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
}
