macro_rules! test_translate_expression {
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(mapping_registry = $mapping_registry:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{translator, mapping_registry::MqlMappingRegistry};

            #[allow(unused_mut, unused_assignments)]
            let mut mapping_registry = MqlMappingRegistry::default();
            $(mapping_registry = $mapping_registry;)?

            let translator = translator::MqlTranslator{
                mapping_registry,
                scope_level: 0u16,
            };
            let expected = $expected;
            let actual = translator.translate_expression($input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_translate_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::{air, mir, translator};
            let mut translator = translator::MqlTranslator::new();
            let expected = $expected;
            let actual = translator.translate_stage($input);
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

mod reference_expression {
    use crate::{air, mir, translator::Error};
    test_translate_expression!(
        not_found,
        expected = Err(Error::ReferenceNotFound(("f", 0u16).into())),
        input = mir::Expression::Reference(("f", 0u16).into()),
    );

    test_translate_expression!(
        found,
        expected = Ok(air::Expression::FieldRef(air::FieldRef {
            parent: None,
            name: "f".to_string()
        })),
        input = mir::Expression::Reference(("f", 0u16).into()),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
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
                "foo".into() => air::Expression::Variable("ROOT".into()),
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
                "foo".into() => translator::ROOT.clone(),
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
                    "foo".into() => air::Expression::Variable("ROOT".to_string()),
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
                "foo".into() => translator::ROOT.clone(),
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
                    "foo".to_string() =>  air::Expression::Variable("ROOT".to_string())
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
                    "foo".to_string() =>  air::Expression::Variable("ROOT".to_string())
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
