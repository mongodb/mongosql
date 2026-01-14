use crate::{air, mir};

macro_rules! test_translate_match_query {
    ($func_name:ident, expected = $expected: expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{
                air,
                mapping_registry::{MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
                mir,
                options::SqlOptions,
                translator::{
                    self,
                    test::match_query::{air_field_input, mir_field_input},
                },
            };

            let mut mapping_registry = MqlMappingRegistry::default();
            mapping_registry.insert(
                ("f", 0u16),
                MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
            );

            let translator = translator::MqlTranslator {
                mapping_registry,
                scope_level: 0u16,
                is_join: false,
                sql_options: SqlOptions::default(),
            };

            let input = $input;
            let expected = $expected;
            let actual = translator.translate_match_query(input);
            assert_eq!(expected, actual);
        }
    };
}

fn air_field_input() -> Option<air::FieldRef> {
    Some("f.a".to_string().into())
}

fn mir_field_input() -> Option<mir::FieldPath> {
    Some(mir::FieldPath::new(
        ("f", 0u16).into(),
        vec!["a".to_string()],
    ))
}

mod logical {
    mod or {
        test_translate_match_query!(
            empty,
            expected = Ok(air::MatchQuery::Or(vec![])),
            input = mir::MatchQuery::Logical(mir::MatchLanguageLogical {
                op: mir::MatchLanguageLogicalOp::Or,
                args: vec![],
                cache: mir::schema::SchemaCache::new(),
            })
        );

        test_translate_match_query!(
            single,
            expected = Ok(air::MatchQuery::Or(vec![air::MatchQuery::Comparison(
                air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Gt,
                    input: air_field_input(),
                    arg: air::LiteralValue::Integer(1),
                }
            )])),
            input = mir::MatchQuery::Logical(mir::MatchLanguageLogical {
                op: mir::MatchLanguageLogicalOp::Or,
                args: vec![mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                    function: mir::MatchLanguageComparisonOp::Gt,
                    input: mir_field_input(),
                    arg: mir::LiteralValue::Integer(1),
                    cache: mir::schema::SchemaCache::new(),
                })],
                cache: mir::schema::SchemaCache::new(),
            })
        );

        test_translate_match_query!(
            multiple,
            expected = Ok(air::MatchQuery::Or(vec![
                air::MatchQuery::Comparison(air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Gt,
                    input: air_field_input(),
                    arg: air::LiteralValue::Integer(1),
                }),
                air::MatchQuery::Comparison(air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Lt,
                    input: air_field_input(),
                    arg: air::LiteralValue::Integer(10),
                }),
            ])),
            input = mir::MatchQuery::Logical(mir::MatchLanguageLogical {
                op: mir::MatchLanguageLogicalOp::Or,
                args: vec![
                    mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                        function: mir::MatchLanguageComparisonOp::Gt,
                        input: mir_field_input(),
                        arg: mir::LiteralValue::Integer(1),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                    mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                        function: mir::MatchLanguageComparisonOp::Lt,
                        input: mir_field_input(),
                        arg: mir::LiteralValue::Integer(10),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                ],
                cache: mir::schema::SchemaCache::new(),
            })
        );
    }

    mod and {
        test_translate_match_query!(
            empty,
            expected = Ok(air::MatchQuery::And(vec![])),
            input = mir::MatchQuery::Logical(mir::MatchLanguageLogical {
                op: mir::MatchLanguageLogicalOp::And,
                args: vec![],
                cache: mir::schema::SchemaCache::new(),
            })
        );

        test_translate_match_query!(
            single,
            expected = Ok(air::MatchQuery::And(vec![air::MatchQuery::Comparison(
                air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Gt,
                    input: air_field_input(),
                    arg: air::LiteralValue::Integer(1),
                }
            )])),
            input = mir::MatchQuery::Logical(mir::MatchLanguageLogical {
                op: mir::MatchLanguageLogicalOp::And,
                args: vec![mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                    function: mir::MatchLanguageComparisonOp::Gt,
                    input: mir_field_input(),
                    arg: mir::LiteralValue::Integer(1),
                    cache: mir::schema::SchemaCache::new(),
                })],
                cache: mir::schema::SchemaCache::new(),
            })
        );

        test_translate_match_query!(
            multiple,
            expected = Ok(air::MatchQuery::And(vec![
                air::MatchQuery::Comparison(air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Gt,
                    input: air_field_input(),
                    arg: air::LiteralValue::Integer(1),
                }),
                air::MatchQuery::Comparison(air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Lt,
                    input: air_field_input(),
                    arg: air::LiteralValue::Integer(10),
                }),
            ])),
            input = mir::MatchQuery::Logical(mir::MatchLanguageLogical {
                op: mir::MatchLanguageLogicalOp::And,
                args: vec![
                    mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                        function: mir::MatchLanguageComparisonOp::Gt,
                        input: mir_field_input(),
                        arg: mir::LiteralValue::Integer(1),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                    mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                        function: mir::MatchLanguageComparisonOp::Lt,
                        input: mir_field_input(),
                        arg: mir::LiteralValue::Integer(10),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                ],
                cache: mir::schema::SchemaCache::new(),
            })
        );
    }
}

mod type_op {
    test_translate_match_query!(
        missing,
        expected = Ok(air::MatchQuery::Type(air::MatchLanguageType {
            input: air_field_input(),
            target_type: air::TypeOrMissing::Missing,
        })),
        input = mir::MatchQuery::Type(mir::MatchLanguageType {
            input: mir_field_input(),
            target_type: mir::TypeOrMissing::Missing,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        number,
        expected = Ok(air::MatchQuery::Type(air::MatchLanguageType {
            input: air_field_input(),
            target_type: air::TypeOrMissing::Number,
        })),
        input = mir::MatchQuery::Type(mir::MatchLanguageType {
            input: mir_field_input(),
            target_type: mir::TypeOrMissing::Number,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        atomic,
        expected = Ok(air::MatchQuery::Type(air::MatchLanguageType {
            input: air_field_input(),
            target_type: air::TypeOrMissing::Type(air::Type::String),
        })),
        input = mir::MatchQuery::Type(mir::MatchLanguageType {
            input: mir_field_input(),
            target_type: mir::TypeOrMissing::Type(mir::Type::String),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod regex {
    test_translate_match_query!(
        simple,
        expected = Ok(air::MatchQuery::Regex(air::MatchLanguageRegex {
            input: air_field_input(),
            regex: "abc".into(),
            options: "ix".into(),
        })),
        input = mir::MatchQuery::Regex(mir::MatchLanguageRegex {
            input: mir_field_input(),
            regex: "abc".into(),
            options: "ix".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod elem_match {
    test_translate_match_query!(
        simple,
        expected = Ok(air::MatchQuery::ElemMatch(air::ElemMatch {
            input: air_field_input().unwrap(),
            condition: Box::new(air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gte,
                input: None,
                arg: air::LiteralValue::Integer(1),
            })),
        })),
        input = mir::MatchQuery::ElemMatch(mir::ElemMatch {
            input: mir_field_input().unwrap(),
            condition: Box::new(mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                function: mir::MatchLanguageComparisonOp::Gte,
                input: None,
                arg: mir::LiteralValue::Integer(1),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod comp {
    test_translate_match_query!(
        lt,
        expected = Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Lt,
            input: air_field_input(),
            arg: air::LiteralValue::Integer(1),
        })),
        input = mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
            function: mir::MatchLanguageComparisonOp::Lt,
            input: mir_field_input(),
            arg: mir::LiteralValue::Integer(1),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        lte,
        expected = Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Lte,
            input: air_field_input(),
            arg: air::LiteralValue::Integer(1),
        })),
        input = mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
            function: mir::MatchLanguageComparisonOp::Lte,
            input: mir_field_input(),
            arg: mir::LiteralValue::Integer(1),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        ne,
        expected = Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Ne,
            input: air_field_input(),
            arg: air::LiteralValue::Integer(1),
        })),
        input = mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
            function: mir::MatchLanguageComparisonOp::Ne,
            input: mir_field_input(),
            arg: mir::LiteralValue::Integer(1),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        eq,
        expected = Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Eq,
            input: air_field_input(),
            arg: air::LiteralValue::Integer(1),
        })),
        input = mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
            function: mir::MatchLanguageComparisonOp::Eq,
            input: mir_field_input(),
            arg: mir::LiteralValue::Integer(1),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        gt,
        expected = Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Gt,
            input: air_field_input(),
            arg: air::LiteralValue::Integer(1),
        })),
        input = mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
            function: mir::MatchLanguageComparisonOp::Gt,
            input: mir_field_input(),
            arg: mir::LiteralValue::Integer(1),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_match_query!(
        gte,
        expected = Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Gte,
            input: air_field_input(),
            arg: air::LiteralValue::Integer(1),
        })),
        input = mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
            function: mir::MatchLanguageComparisonOp::Gte,
            input: mir_field_input(),
            arg: mir::LiteralValue::Integer(1),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod match_path {
    use crate::{
        air,
        mapping_registry::{MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
        mir,
        options::SqlOptions,
        translator::{self, Result},
    };

    #[test]
    fn multi_part_match_path() {
        let mut mapping_registry = MqlMappingRegistry::default();
        mapping_registry.insert(
            ("f", 0u16),
            MqlMappingRegistryValue::new("f".to_string(), MqlReferenceType::FieldRef),
        );

        let translator = translator::MqlTranslator {
            mapping_registry,
            scope_level: 0u16,
            is_join: false,
            sql_options: SqlOptions::default(),
        };

        let input = mir::FieldPath::new(("f", 0u16).into(), vec!["x".to_string(), "y".to_string()]);

        let expected: Result<Option<air::FieldRef>> = Ok(Some("f.x.y".to_string().into()));
        let actual = translator.translate_field_path(input);
        assert_eq!(expected, actual);
    }
}
