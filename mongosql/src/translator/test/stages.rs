macro_rules! test_translate_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{air, mir, options, translator, util};
            let mut translator = translator::MqlTranslator::new(options::SqlOptions::default());
            let expected = $expected;
            let actual = translator.translate_stage($input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_translate_plan {
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(options = $options:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{air, options, translator};
            #[allow(unused_mut, unused_assignments)]
            let mut options = options::SqlOptions::default();
            $(options = $options;)?
            let mut translator = translator::MqlTranslator::new(options);
            let expected = $expected;
            let actual = translator.translate_plan($input);
            assert_eq!(expected, actual);
        }
    };
}

mod filter {
    test_translate_stage!(
        basic,
        expected = Ok(air::Stage::Match(air::Match::ExprLanguage(
            air::ExprLanguage {
                source: util::air_documents_stage(vec![]),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            }
        ))),
        input = mir::Stage::Filter(mir::Filter {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            condition: mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod project {
    use crate::{map, translator::Error, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        project,
        expected = Ok(air::Stage::Project(air::Project {
            source: util::air_collection_stage("test_db", "foo"),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                "bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1)))
            }
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: util::mir_collection("test_db", "foo"),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        add_fields,
        expected = Ok(air::Stage::AddFields(air::AddFields {
            source: util::air_collection_stage("test_db", "foo"),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => ROOT.clone(),
                "bar".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1))
            }
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: true,
            source: util::mir_collection("test_db", "foo"),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_key_contains_dots,
        expected = Ok(*util::air_project_collection(
            None,
            "foo.bar.baz",
            Some("foo_bar_baz"),
        )),
        input = *util::mir_project_collection(None, "foo.bar.baz", None, None)
    );

    test_translate_stage!(
        project_key_starts_with_dollar,
        expected = Ok(*util::air_project_collection(None, "$foo", Some("_foo"))),
        input = *util::mir_project_collection(None, "$foo", None, None)
    );

    test_translate_stage!(
        project_key_is_empty,
        expected = Err(Error::InvalidProjectField),
        input = *util::mir_project_collection(None, "", None, None)
    );

    test_translate_stage!(
        project_key_starts_with_special_characters_conflicts_with_other_project_key,
        expected = Ok(air::Stage::Project(air::Project {
            source: util::air_collection_stage("test_db", "$foo.bar"),
            specifications: unchecked_unique_linked_hash_map! {
                "__foo_bar".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                "_foo_bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
            }
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: util::mir_collection("test_db", "$foo.bar"),
            expression: BindingTuple(map! {
                Key::named("$foo.bar", 0u16) => mir::Expression::Reference(("$foo.bar", 0u16).into()),
                Key::named("_foo_bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_with_user_bot_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: util::air_collection_stage("test_db", "foo"),
            specifications: unchecked_unique_linked_hash_map! {
                "___bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                // reordered because BindingTuple uses BTreeMap
                "____bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(4))),
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
                "_bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
            }
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: util::mir_collection("test_db", "foo"),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("__bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                Key::named("_bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                Key::named("____bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(4)),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        select_values_non_literal_document_expr_correctness_test,
        expected = Ok(air::Stage::Project(air::Project {
            source: util::air_project_collection(Some("mydb"), "foo", Some("t1")),
            specifications: unchecked_unique_linked_hash_map!(
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("t1.b".to_string().into())),
            ),
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: util::mir_project_collection(Some("mydb"), "foo", Some("t1"), None),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                    expr: util::mir_field_access("t1", "b", true),
                    target_type: mir::Type::Document,
                }),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod group {
    use crate::{translator::Error, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::Key;

    test_translate_stage!(
        group_count_star,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: util::air_collection_stage("test_db", "foo"),
                keys: vec![air::NameExprPair {
                    name: "x_key".into(),
                    expr: ROOT.clone()
                },],
                aggregations: vec![
                    // Count(*) is translated as Count(ROOT).
                    air::AccumulatorExpr {
                        alias: "c_distinct".into(),
                        function: air::AggregationFunction::Count,
                        distinct: true,
                        arg: ROOT.clone().into(),
                    },
                    air::AccumulatorExpr {
                        alias: "c_nondistinct".into(),
                        function: air::AggregationFunction::Count,
                        distinct: false,
                        arg: ROOT.clone().into(),
                    },
                ]
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "x_key".to_string() => air::Expression::FieldRef("_id.x_key".to_string().into()),
                    "c_distinct".to_string() => air::Expression::FieldRef("c_distinct".to_string().into()),
                    "c_nondistinct".to_string() => air::Expression::FieldRef("c_nondistinct".to_string().into()),
                })),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: util::mir_collection("test_db", "foo"),
            keys: vec![mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                alias: "x_key".into(),
                expr: mir::Expression::Reference(mir::ReferenceExpr {
                    key: Key::named("foo", 0u16),
                })
            }),],
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
            scope: 0,
        })
    );

    test_translate_stage!(
        group_normal_operators,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: util::air_collection_stage("test_db", "foo"),
                keys: vec![air::NameExprPair {
                    name: "x_key".into(),
                    expr: ROOT.clone()
                },],
                aggregations: vec![
                    air::AccumulatorExpr {
                        alias: "max_distinct".into(),
                        function: air::AggregationFunction::Max,
                        distinct: true,
                        arg: Box::new(ROOT.clone()),
                    },
                    air::AccumulatorExpr {
                        alias: "min_nondistinct".into(),
                        function: air::AggregationFunction::Min,
                        distinct: false,
                        arg: Box::new(ROOT.clone()),
                    }
                ]
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "x_key".to_string() => air::Expression::FieldRef("_id.x_key".to_string().into()),
                    "max_distinct".to_string() => air::Expression::FieldRef("max_distinct".to_string().into()),
                    "min_nondistinct".to_string() => air::Expression::FieldRef("min_nondistinct".to_string().into()),
                })),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: util::mir_collection("test_db", "foo"),
            keys: vec![mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                alias: "x_key".into(),
                expr: mir::Expression::Reference(mir::ReferenceExpr {
                    key: Key::named("foo", 0u16),
                })
            }),],
            aggregations: vec![
                mir::AliasedAggregation {
                    alias: "max_distinct".into(),
                    agg_expr: mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
                        function: mir::AggregationFunction::Max,
                        distinct: true,
                        arg: mir::Expression::Reference(mir::ReferenceExpr {
                            key: Key::named("foo", 0u16),
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
                        })
                        .into(),
                    }),
                },
            ],
            cache: mir::schema::SchemaCache::new(),
            scope: 0,
        })
    );

    test_translate_stage!(
        group_key_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: util::air_project_collection(None, "foo", None),
                keys: vec![
                    air::NameExprPair {
                        name: "__unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo".to_string().into()),
                    },
                    air::NameExprPair {
                        name: "___unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo.x".to_string().into()),
                    },
                ],
                aggregations: vec![]
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "x".to_string() => air::Expression::FieldRef("_id.___unaliasedKey2".to_string().into()),
                })),
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "__unaliasedKey2".to_string() => air::Expression::FieldRef("_id.__unaliasedKey2".to_string().into()),
                })),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: util::mir_project_collection(None, "foo", None, None),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__unaliasedKey2".into(),
                    expr: mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                    })
                }),
                mir::OptionallyAliasedExpr::Unaliased(*util::mir_field_access("foo", "x", true)),
            ],
            aggregations: vec![],
            cache: mir::schema::SchemaCache::new(),
            scope: 0,
        })
    );

    test_translate_stage!(
        aggregation_alias_id_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: util::air_project_collection(None, "foo", None),
                keys: vec![
                    air::NameExprPair {
                        name: "__unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo".to_string().into()),
                    },
                    air::NameExprPair {
                        name: "___unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo.x".to_string().into())
                    },
                ],
                aggregations: vec![air::AccumulatorExpr {
                    alias: "__id".into(),
                    function: air::AggregationFunction::Count,
                    distinct: false,
                    arg: ROOT.clone().into(),
                },]
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "x".to_string() => air::Expression::FieldRef("_id.___unaliasedKey2".to_string().into()),
                })),
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "__unaliasedKey2".to_string() => air::Expression::FieldRef("_id.__unaliasedKey2".to_string().into()),
                    "_id".to_string() => air::Expression::FieldRef("__id".to_string().into()),
                })),
            }
        })),
        input = mir::Stage::Group(mir::Group {
            source: util::mir_project_collection(None, "foo", None, None),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__unaliasedKey2".into(),
                    expr: mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                    })
                }),
                mir::OptionallyAliasedExpr::Unaliased(*util::mir_field_access("foo", "x", true)),
            ],
            aggregations: vec![mir::AliasedAggregation {
                alias: "_id".into(),
                agg_expr: mir::AggregationExpr::CountStar(false),
            },],
            cache: mir::schema::SchemaCache::new(),
            scope: 0,
        })
    );

    test_translate_stage!(
        unaliased_group_key_with_no_datasource_is_error,
        expected = Err(Error::InvalidGroupKey),
        input = mir::Stage::Group(mir::Group {
            source: util::mir_collection("test_db", "foo"),
            keys: vec![mir::OptionallyAliasedExpr::Unaliased(
                mir::Expression::Reference(mir::ReferenceExpr {
                    key: Key::named("foo", 0u16),
                })
            ),],
            aggregations: vec![],
            cache: mir::schema::SchemaCache::new(),
            scope: 0,
        })
    );
}

mod limit {
    use crate::translator;
    use translator::Error;

    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::Limit(air::Limit {
            source: util::air_collection_stage("test_db", "col"),
            limit: 1i64,
        })),
        input = mir::Stage::Limit(mir::Limit {
            source: util::mir_collection("test_db", "col"),
            limit: 1,
            cache: mir::schema::SchemaCache::new(),
        })
    );
    test_translate_stage!(
        out_of_i64_range,
        expected = Err(Error::LimitOutOfI64Range(u64::MAX)),
        input = mir::Stage::Limit(mir::Limit {
            source: util::mir_collection("test_db", "col"),
            limit: u64::MAX,
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod offset {
    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::Skip(air::Skip {
            source: util::air_collection_stage("test_db", "foo"),
            skip: 10,
        })),
        input = mir::Stage::Offset(mir::Offset {
            source: util::mir_collection("test_db", "foo"),
            offset: 10,
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod sort {
    use crate::{map, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        sort_stage_multi_key_reference,
        expected = Ok(air::Stage::Sort(air::Sort {
                source: air::Stage::Project(air::Project {
                    source: util::air_collection_stage("test_db", "foo"),
                    specifications: unchecked_unique_linked_hash_map!{
                        "__bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                        "bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                        "baz".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
                    }
                }).into(),
           specs: vec![
                air::SortSpecification::Desc("bar.a".to_string()),
                air::SortSpecification::Asc("baz.a".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: mir::Stage::Project(mir::Project {
            is_add_fields: false,
                source: util::mir_collection("test_db", "foo"),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                    Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    Key::named("baz", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                }),
            cache: mir::schema::SchemaCache::new(),
        }).into(),
            specs: vec![
                mir::SortSpecification::Desc(util::mir_field_path("bar", vec!["a"])),
                mir::SortSpecification::Asc(util::mir_field_path("baz", vec!["a"]))
            ],
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        sort_stage_multi_key_field_access,
        expected = Ok(air::Stage::Sort(air::Sort {
            source: util::air_project_collection(None, "foo", None),
            specs: vec![
                air::SortSpecification::Desc("foo.bar".to_string()),
                air::SortSpecification::Asc("foo.baz".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: util::mir_project_collection(None, "foo", None, None),
            specs: vec![
                mir::SortSpecification::Desc(util::mir_field_path("foo", vec!["bar"])),
                mir::SortSpecification::Asc(util::mir_field_path("foo", vec!["baz"]))
            ],
            cache: mir::schema::SchemaCache::new()
        })
    );

    test_translate_stage!(
        sort_stage_nested_key,
        expected = Ok(air::Stage::Sort(air::Sort {
            source: util::air_project_collection(None, "foo", None),
            specs: vec![
                air::SortSpecification::Desc("foo.bar.quz".to_string()),
                air::SortSpecification::Asc("foo.baz.fizzle.bazzle".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: util::mir_project_collection(None, "foo", None, None),
            specs: vec![
                mir::SortSpecification::Desc(util::mir_field_path("foo", vec!["bar", "quz"])),
                mir::SortSpecification::Asc(util::mir_field_path(
                    "foo",
                    vec!["baz", "fizzle", "bazzle"],
                ))
            ],
            cache: mir::schema::SchemaCache::new()
        })
    );
}

mod collection {
    test_translate_stage!(
        collection,
        expected = Ok(air::Stage::Collection(air::Collection {
            db: "test_db".into(),
            collection: "foo".into(),
        })),
        input = mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        collection_with_dot,
        expected = Ok(air::Stage::Collection(air::Collection {
            db: "test_db".into(),
            collection: "foo.bar".into(),
        })),
        input = mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: "foo.bar".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod array {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        non_empty,
        expected = Ok(air::Stage::Documents(air::Documents {
            array: vec![air::Expression::Literal(air::LiteralValue::Boolean(false))],
        })),
        input = mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Literal(mir::LiteralValue::Boolean(false))],
            alias: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        empty,
        expected = Ok(air::Stage::Documents(air::Documents { array: vec![] })),
        input = mir::Stage::Array(mir::ArraySource {
            array: vec![],
            alias: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        basic_array_datasource_with_multiple_documents_correctness_test,
        expected = Ok(air::Stage::Documents(air::Documents {
            array: vec![
                air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(1)),
                }),
                air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2)),
                }),
                air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".to_string() => air::Expression::Literal(air::LiteralValue::Integer(3)),
                }),
            ],
        })),
        input = mir::Stage::Array(mir::ArraySource {
            array: vec![
                mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Literal(
                            mir::LiteralValue::Integer(1),
                        ),
                    },
                }),
                mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Literal(
                            mir::LiteralValue::Integer(2),
                        ),
                    },
                }),
                mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Literal(
                            mir::LiteralValue::Integer(3),
                        ),
                    },
                }),
            ],
            alias: "arr".to_string(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod join {
    use crate::{map, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        join_without_condition,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Inner,
            left: util::air_collection_stage("test_db", "foo"),
            right: util::air_collection_stage("test_db", "bar"),
            let_vars: None,
            condition: None
        })),
        input = mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Inner,
            left: util::mir_collection("test_db", "foo"),
            right: util::mir_collection("test_db", "bar"),
            condition: None,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        left_join_with_condition,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Left,
            left: util::air_collection_stage("test_db", "foo"),
            right: util::air_collection_stage("test_db", "bar"),
            let_vars: Some(vec![air::LetVariable {
                name: "vfoo_0".to_string(),
                expr: Box::new(ROOT.clone()),
            }]),
            condition: Some(air::Expression::SQLSemanticOperator(
                air::SQLSemanticOperator {
                    op: air::SQLOperator::Eq,
                    args: vec![
                        air::Expression::Variable("vfoo_0".to_string().into()),
                        ROOT.clone(),
                    ]
                }
            ))
        })),
        input = mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Left,
            left: util::mir_collection("test_db", "foo"),
            right: util::mir_collection("test_db", "bar"),
            condition: Some(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication::new(
                    mir::ScalarFunction::Eq,
                    vec![
                        mir::Expression::Reference(("foo", 0u16).into()),
                        mir::Expression::Reference(("bar", 0u16).into())
                    ],
                )
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
                left: util::air_project_collection(None, "Foo", None),
                right: util::air_project_collection(None, "foo", None),
                condition: None,
                let_vars: None,
            })),
            right: util::air_collection_stage("test_db", "bar"),
            let_vars: Some(vec![
                air::LetVariable {
                    name: "vfoo_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef("Foo".to_string().into())),
                },
                air::LetVariable {
                    name: "vfoo_0_".to_string(),
                    expr: Box::new(air::Expression::FieldRef("foo".to_string().into())),
                }
            ]),
            condition: Some(air::Expression::Literal(air::LiteralValue::Boolean(true))),
        })),
        input = mir::Stage::Join(mir::Join {
            condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            left: mir::Stage::Join(mir::Join {
                condition: None,
                left: util::mir_project_collection(None, "Foo", None, None),
                right: util::mir_project_collection(None, "foo", None, None),
                join_type: mir::JoinType::Inner,
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            right: util::mir_collection("test_db", "bar"),
            join_type: mir::JoinType::Inner,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        test_translate_array,
        expected = Ok(air::Stage::Join(air::Join {
            condition: None,
            left: util::air_collection_stage("mydb", "col"),
            right: util::air_documents_stage(vec![
                air::Expression::Literal(air::LiteralValue::Integer(1)),
                air::Expression::Literal(air::LiteralValue::Integer(1))
            ]),
            let_vars: None,
            join_type: air::JoinType::Left,
        })),
        input = mir::Stage::Join(mir::Join {
            condition: None,
            left: util::mir_collection("mydb", "col"),
            right: mir::Stage::Array(mir::ArraySource {
                array: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1))
                ],
                alias: "arr".to_string(),
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            join_type: mir::JoinType::Left,
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        joins_retain_aliases_for_left_and_right,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Join(air::Join {
                left: util::air_project_collection(None, "foo", Some("t1")),
                right: util::air_project_collection(None, "bar", Some("t2")),
                join_type: air::JoinType::Inner,
                let_vars: Some(vec![air::LetVariable {
                    name: "vt1_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef("t1".to_string().into())),
                }]),
                condition: Some(air::Expression::SQLSemanticOperator(
                    air::SQLSemanticOperator {
                        op: air::SQLOperator::Eq,
                        args: vec![
                            air::Expression::Variable("vt1_0.a".to_string().into()),
                            air::Expression::FieldRef("t2.b".to_string().into())
                        ]
                    }
                )),
            })),
            specifications: unchecked_unique_linked_hash_map!(
                "bar".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("t2".to_string().into())),
                "foo".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("t1".to_string().into())),
            )
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::Join(mir::Join {
                join_type: mir::JoinType::Inner,
                left: util::mir_project_collection(None, "foo", Some("t1"), None),
                right: util::mir_project_collection(None, "bar", Some("t2"), None),
                condition: Some(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Eq,
                        args: vec![
                            *util::mir_field_access("t1", "a", true),
                            *util::mir_field_access("t2", "b", true),
                        ],
                        is_nullable: true,
                    }
                )),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::named("foo", 0u16) => mir::Expression::Reference(("t1".to_string(), 0u16).into()),
                Key::named("bar", 0u16) => mir::Expression::Reference(("t2".to_string(), 0u16).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        ensure_let_variables_start_with_lowercase_letters_not_underscore_or_uppercase,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Inner,
            left: Box::new(air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: util::air_project_collection(None, "foo", Some("Foo")),
                right: util::air_project_collection(None, "bar", Some("_bar")),
                condition: None,
                let_vars: None,
            })),
            right: util::air_collection_stage("test_db", "bar"),
            let_vars: Some(vec![
                air::LetVariable {
                    name: "vfoo_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef("Foo".to_string().into())),
                },
                air::LetVariable {
                    name: "v_bar_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef("_bar".to_string().into())),
                }
            ]),
            condition: Some(air::Expression::Literal(air::LiteralValue::Boolean(true))),
        })),
        input = mir::Stage::Join(mir::Join {
            condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            left: mir::Stage::Join(mir::Join {
                condition: None,
                left: util::mir_project_collection(None, "foo", Some("Foo"), None),
                right: util::mir_project_collection(None, "bar", Some("_bar"), None),
                join_type: mir::JoinType::Inner,
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            right: util::mir_collection("test_db", "bar"),
            join_type: mir::JoinType::Inner,
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod equijoin {
    use crate::translator::Error;

    test_translate_stage!(
        inner_join,
        expected = Ok(air::Stage::EquiJoin(air::EquiJoin {
            join_type: air::JoinType::Inner,
            source: util::air_project_collection(None, "foo", None),
            from: util::air_collection_raw("test_db", "bar"),
            local_field: "foo.a".into(),
            foreign_field: "a".into(),
            as_name: "bar".to_string(),
        })),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::EquiJoin(mir::EquiJoin {
            join_type: mir::JoinType::Inner,
            source: util::mir_project_collection(None, "foo", None, None),
            from: util::mir_project_collection(None, "bar", None, None),
            local_field: Box::new(util::mir_field_path("foo", vec!["a"])),
            foreign_field: Box::new(util::mir_field_path("bar", vec!["a"])),
            cache: mir::schema::SchemaCache::new(),
        }))
    );

    test_translate_stage!(
        left_join,
        expected = Ok(air::Stage::EquiJoin(air::EquiJoin {
            join_type: air::JoinType::Left,
            source: util::air_project_collection(None, "foo", None),
            from: util::air_collection_raw("test_db", "bar"),
            local_field: "foo.a".into(),
            foreign_field: "a".into(),
            as_name: "x".to_string(),
        })),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::EquiJoin(mir::EquiJoin {
            join_type: mir::JoinType::Left,
            source: util::mir_project_collection(None, "foo", None, None),
            from: util::mir_project_collection(None, "bar", Some("x"), None),
            local_field: Box::new(util::mir_field_path("foo", vec!["a"])),
            foreign_field: Box::new(util::mir_field_path("x", vec!["a"])),
            cache: mir::schema::SchemaCache::new(),
        }))
    );

    test_translate_stage!(
        from_must_be_collection,
        expected = Err(Error::ExpectedCollection),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::EquiJoin(mir::EquiJoin {
            join_type: mir::JoinType::Left,
            source: util::mir_project_collection(None, "foo", None, None),
            from: mir::Stage::Array(mir::ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: mir::schema::SchemaCache::new()
            })
            .into(),
            local_field: Box::new(util::mir_field_path("foo", vec!["a"])),
            foreign_field: Box::new(util::mir_field_path("bar", vec!["a"])),
            cache: mir::schema::SchemaCache::new(),
        }))
    );
}

mod lateral_join {
    use crate::{map, unchecked_unique_linked_hash_map};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        lateral_join_inner_simple,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Inner,
            left: util::air_project_collection(None, "foo", None),
            right: util::air_project_collection(None, "bar", None),
            let_vars: Some(vec![air::LetVariable {
                name: "vfoo_0".to_string(),
                expr: Box::new(air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: "foo".to_string(),
                }))
            }]),
            condition: None
        })),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::LateralJoin(mir::LateralJoin {
            join_type: mir::JoinType::Inner,
            source: util::mir_project_collection(None, "foo", None, None),
            subquery: util::mir_project_collection(None, "bar", None, None),
            cache: mir::schema::SchemaCache::new(),
        }))
    );

    test_translate_stage!(
        left_lateral_join_let_vars,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Join(air::Join {
                join_type: air::JoinType::Left,
                left: util::air_project_collection(None, "foo", Some("t1")),
                right: air::Stage::Project(air::Project {
                    source: util::air_collection_stage("test_db", "bar"),
                    specifications: unchecked_unique_linked_hash_map!()
                })
                .into(),
                let_vars: Some(vec![air::LetVariable {
                    name: "vt1_0".to_string(),
                    expr: Box::new(air::Expression::FieldRef("t1".to_string().into())),
                }]),
                condition: None
            })),
            specifications: unchecked_unique_linked_hash_map!(
                "foo".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("t1".to_string().into())),
            )
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::MQLIntrinsic(mir::MQLStage::LateralJoin(
                mir::LateralJoin {
                    join_type: mir::JoinType::Left,
                    source: util::mir_project_collection(None, "foo", Some("t1"), None),
                    subquery: mir::Stage::Project(mir::Project {
                        is_add_fields: false,
                        source: util::mir_collection("test_db", "bar"),
                        expression: map! {},
                        cache: mir::schema::SchemaCache::new(),
                    })
                    .into(),
                    cache: mir::schema::SchemaCache::new(),
                }
            ))),
            expression: BindingTuple(map! {
                Key::named("foo", 0u16) => mir::Expression::Reference(("t1".to_string(), 0u16).into()),
            }),
            cache: mir::schema::SchemaCache::new()
        })
    );
}

mod set {
    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::UnionWith(air::UnionWith {
            source: util::air_collection_stage("foo", "a"),
            pipeline: util::air_collection_stage("bar", "b"),
        })),
        input = mir::Stage::Set(mir::Set {
            operation: mir::SetOperation::UnionAll,
            left: util::mir_collection("foo", "a"),
            right: util::mir_collection("bar", "b"),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod derived {
    use crate::{map, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        derived,
        expected = Ok(air::Stage::Project(air::Project {
            source: util::air_collection_stage("test_db", "foo"),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                "bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1)))
            }
        })),
        input = mir::Stage::Derived(mir::Derived {
            source: Box::new(mir::Stage::Project(mir::Project {
                is_add_fields: false,
                source: util::mir_collection("test_db", "foo"),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Reference(("foo", 1u16).into()),
                    Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                }),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        nested_derived_tables_correctness_test,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: util::air_collection_stage("foo", "bar"),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    "d2".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1)))
                }
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("d2.a".to_string().into()))
            }
        })),
        input = mir::Stage::Derived(mir::Derived {
            source: Box::new(mir::Stage::Project(mir::Project {
                is_add_fields: false,
                source: Box::new(mir::Stage::Derived(mir::Derived {
                    source: Box::new(mir::Stage::Project(mir::Project {
                        is_add_fields: false,
                        source: util::mir_collection("foo", "bar"),
                        expression: BindingTuple(map! {
                            Key::bot(1) => mir::Expression::Reference(mir::ReferenceExpr {
                                key: Key::named("bar", 2u16),
                            }),
                            Key::named("d2", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        }),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                        expr: util::mir_field_access("d2", "a", true),
                        target_type: mir::Type::Int32,
                    }),
                }),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod unwind {
    use crate::{air::ExprLanguage, map, unchecked_unique_linked_hash_map};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage! {
        unwind,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: util::air_collection_stage("test_db", "foo"),
            path: util::air_variable_from_root("bar"),
            index: None,
            outer: false,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: util::mir_collection("test_db", "foo"),
            path: util::mir_field_path("foo", vec!["bar"]),
            index: None,
            outer: false,
            cache: mir::schema::SchemaCache::new(),
            is_prefiltered: false,
        })
    }

    test_translate_stage! {
        unwind_outer,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: util::air_collection_stage("test_db", "foo"),
            path: util::air_variable_from_root("bar"),
            index: None,
            outer: true,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: util::mir_collection("test_db", "foo"),
            path: util::mir_field_path("foo", vec!["bar"]),
            index: None,
            outer: true,
            cache: mir::schema::SchemaCache::new(),
            is_prefiltered: false,
        })
    }
    test_translate_stage! {
        unwind_index,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: util::air_collection_stage("test_db", "foo"),
            path: util::air_variable_from_root("bar"),
            index: Some("i".to_string()),
            outer: true,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: util::mir_collection("test_db", "foo"),
            path: util::mir_field_path("foo", vec!["bar"]),
            index: Some("i".into()),
            outer: true,
            cache: mir::schema::SchemaCache::new(),
            is_prefiltered: false,
        })
    }

    test_translate_stage! {
        correctness_test_for_index_option_using_project_and_where,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Match(air::Match::ExprLanguage(ExprLanguage {
                source: Box::new(air::Stage::Unwind(air::Unwind {
                    source: util::air_collection_stage("test_db", "foo"),
                    path: util::air_variable_from_root("arr"),
                    index: Some("idx".to_string()),
                    outer: false,
                })),
                expr: Box::new(air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::Gt,
                    args: vec![
                        util::air_variable_from_root("arr"),
                        air::Expression::Literal(air::LiteralValue::Integer(0)),
                    ],
                })),
            }))),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "arr".to_string() => util::air_variable_from_root("arr"),
                    "idx".to_string() => util::air_variable_from_root("idx")
                })),
            },
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: Box::new(mir::Stage::Filter(mir::Filter {
                source: Box::new(mir::Stage::Unwind(mir::Unwind {
                    source: util::mir_collection("test_db", "foo"),
                    path: util::mir_field_path("foo", vec!["arr"]),
                    index: Some("idx".into()),
                    outer: false,
                    cache: mir::schema::SchemaCache::new(),
                    is_prefiltered: false,
                })),
                condition: mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Gt,
                    args: vec![
                        *util::mir_field_access("foo", "arr", true),
                        mir::Expression::Literal(mir::LiteralValue::Integer(0)),
                    ],
                    is_nullable: true,
                }),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "arr".into() => *util::mir_field_access("foo", "arr", true),
                        "idx".into() => *util::mir_field_access("foo", "idx", true),
                    },
                }),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    }
}

mod mql_intrinsic {
    mod match_filter {
        test_translate_stage!(
            basic,
            expected = Ok(air::Stage::Match(air::Match::MatchLanguage(
                air::MatchLanguage {
                    source: util::air_project_collection(None, "foo", Some("f")),
                    expr: Box::new(air::MatchQuery::Comparison(air::MatchLanguageComparison {
                        function: air::MatchLanguageComparisonOp::Lt,
                        input: Some("f.a".to_string().into()),
                        arg: air::LiteralValue::Integer(1),
                    })),
                }
            ))),
            input = mir::Stage::MQLIntrinsic(mir::MQLStage::MatchFilter(mir::MatchFilter {
                source: util::mir_project_collection(None, "foo", Some("f"), None),
                condition: mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                    function: mir::MatchLanguageComparisonOp::Lt,
                    input: Some(util::mir_field_path("f", vec!["a"])),
                    arg: mir::LiteralValue::Integer(1),
                    cache: mir::schema::SchemaCache::new(),
                }),
                cache: mir::schema::SchemaCache::new(),
            }))
        );
    }
}

mod translate_plan {
    use crate::{map, mir, unchecked_unique_linked_hash_map, util, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_plan!(
        project_with_user_bot_conflict,
        expected = Ok(
            air::Stage::ReplaceWith(air::ReplaceWith {
                source: air::Stage::Project(air::Project {
                    source: util::air_collection_stage("test_db", "foo"),
                    specifications: unchecked_unique_linked_hash_map!{
                        "___bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                        "____bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(4))),
                        "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
                        "_bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    }
                }).into(),
                new_root: air::Expression::UnsetField(air::UnsetField {
                    field: "___bot".to_string(),
                    input: air::Expression::SetField(air::SetField {
                        field: "".to_string(),
                        input: ROOT.clone().into(),
                        value: air::Expression::FieldRef("___bot".to_string().into()).into(),
                    }).into()
                }).into()
            })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: util::mir_collection("test_db", "foo"),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Reference(("foo", 0u16).into()),
                Key::named("__bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                Key::named("_bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                Key::named("____bot", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(4)),
            }),
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    // SELECT * FROM `$foo`, `bar.baz`, `$_foo`, `_foo`, `bar`
    test_translate_plan!(
        restore_original_names_with_dots_and_dollars,
        expected = Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: Box::new(air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: Box::new(air::Stage::Join(air::Join {
                    join_type: air::JoinType::Inner,
                    left: Box::new(air::Stage::Join(air::Join {
                        join_type: air::JoinType::Inner,
                        left: Box::new(air::Stage::Join(air::Join {
                            join_type: air::JoinType::Inner,
                            left: util::air_project_collection(None, "$foo", Some("_foo")),
                            right: util::air_project_collection(None, "bar.baz", Some("bar_baz")),
                            let_vars: None,
                            condition: None,
                        })),
                        right: util::air_project_collection(None, "$_foo", Some("__foo")),
                        let_vars: None,
                        condition: None,
                    })),
                    right: util::air_project_collection(None, "_foo", Some("___foo")),
                    let_vars: None,
                    condition: None,
                })),
                right: util::air_project_collection(None, "bar", None),
                let_vars: None,
                condition: None,
            })),
            new_root: Box::new(air::Expression::UnsetField(air::UnsetField {
                field: "___foo".to_string(),
                input: Box::new(air::Expression::SetField(air::SetField {
                    field: "_foo".to_string(),
                    input: Box::new(air::Expression::UnsetField(air::UnsetField {
                        field: "__foo".to_string(),
                        input: Box::new(air::Expression::SetField(air::SetField {
                            field: "$_foo".to_string(),
                            input: Box::new(air::Expression::UnsetField(air::UnsetField {
                                field: "_foo".to_string(),
                                input: Box::new(air::Expression::SetField(air::SetField {
                                    field: "$foo".to_string(),
                                    input: Box::new(air::Expression::UnsetField(air::UnsetField {
                                        field: "bar_baz".to_string(),
                                        input: Box::new(air::Expression::SetField(air::SetField {
                                            field: "bar.baz".to_string(),
                                            input: Box::new(ROOT.clone()),
                                            value: Box::new(air::Expression::FieldRef(
                                                "bar_baz".to_string().into()
                                            )),
                                        })),
                                    })),
                                    value: Box::new(air::Expression::FieldRef(
                                        "_foo".to_string().into()
                                    )),
                                })),
                            })),
                            value: Box::new(air::Expression::FieldRef("__foo".to_string().into())),
                        })),
                    })),
                    value: Box::new(air::Expression::FieldRef("___foo".to_string().into())),
                })),
            }))
        })),
        input = mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Inner,
            left: Box::new(mir::Stage::Join(mir::Join {
                join_type: mir::JoinType::Inner,
                left: Box::new(mir::Stage::Join(mir::Join {
                    join_type: mir::JoinType::Inner,
                    left: Box::new(mir::Stage::Join(mir::Join {
                        join_type: mir::JoinType::Inner,
                        left: util::mir_project_collection(None, "$foo", None, None),
                        right: util::mir_project_collection(None, "bar.baz", None, None),
                        condition: None,
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    right: util::mir_project_collection(None, "$_foo", None, None),
                    condition: None,
                    cache: mir::schema::SchemaCache::new(),
                })),
                right: util::mir_project_collection(None, "_foo", None, None),
                condition: None,
                cache: mir::schema::SchemaCache::new(),
            })),
            right: util::mir_project_collection(None, "bar", None, None),
            condition: None,
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_plan!(
        single_non_namespaced_results,
        expected = Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: util::air_project_collection(None, "foo", None),
            new_root: Box::new(air::Expression::FieldRef("foo".into())),
        })),
        input = *util::mir_project_collection(None, "foo", None, None),
        options = crate::options::SqlOptions::new(
            crate::options::ExcludeNamespacesOption::ExcludeNamespaces,
            crate::SchemaCheckingMode::default()
        ),
    );

    test_translate_plan!(
        multiple_non_namespaced_results,
        expected = Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: Box::new(air::Stage::Join(air::Join {
                join_type: air::JoinType::Inner,
                left: Box::new(air::Stage::Join(air::Join {
                    join_type: air::JoinType::Inner,
                    left: util::air_project_collection(None, "foo", None),
                    right: util::air_project_collection(None, "bar", None),
                    let_vars: None,
                    condition: None,
                })),
                right: util::air_project_collection(None, "baz", None),
                let_vars: None,
                condition: None,
            })),
            new_root: Box::new(air::Expression::MQLSemanticOperator(
                air::MQLSemanticOperator {
                    op: air::MQLOperator::MergeObjects,
                    args: vec![
                        air::Expression::FieldRef("bar".into()),
                        air::Expression::FieldRef("baz".into()),
                        air::Expression::FieldRef("foo".into()),
                    ]
                }
            ))
        })),
        input = mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Inner,
            left: Box::new(mir::Stage::Join(mir::Join {
                join_type: mir::JoinType::Inner,
                left: util::mir_project_collection(None, "foo", None, None),
                right: util::mir_project_collection(None, "bar", None, None),
                condition: None,
                cache: mir::schema::SchemaCache::new(),
            })),
            right: util::mir_project_collection(None, "baz", None, None),
            condition: None,
            cache: mir::schema::SchemaCache::new(),
        }),
        options = util::sql_options_exclude_namespaces(),
    );

    test_translate_plan!(
        non_namespaced_handles_bot,
        expected = Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: util::air_project_collection(None, "foo", Some("__bot")),
            new_root: Box::new(air::Expression::FieldRef("__bot".into()))
        })),
        input = *util::mir_project_bot_collection("foo"),
        options = util::sql_options_exclude_namespaces(),
    );
}

mod subquery_expr {
    use crate::{
        map, mir::binding_tuple::DatasourceName::Bottom, unchecked_unique_linked_hash_map,
    };

    test_translate_stage!(
        unqualified_correlated_reference,
        expected = Ok(air::Stage::Project(air::Project {
            source: util::air_project_collection(Some("foo"), "schema_coll", Some("q")),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "bar".to_string() => air::Expression::Subquery(air::Subquery {
                        let_bindings: vec![air::LetVariable {
                            name: "vq_0".to_string(),
                            expr: Box::new(air::Expression::FieldRef("q".to_string().into())),
                        },],
                        output_path: vec!["__bot".to_string(), "bar".to_string()],
                        pipeline: Box::new(air::Stage::Limit(air::Limit {
                            source: Box::new(air::Stage::Project(air::Project {
                                source: util::air_project_collection(Some("foo"), "schema_foo", Some("q")),
                                specifications: unchecked_unique_linked_hash_map! {
                                    "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                                        "bar".to_string() => air::Expression::Variable("vq_0.bar".to_string().into()),
                                    })),
                                },
                            })),
                            limit: 1,
                        })),
                    })
                })),
            },
        })),
        input = mir::Stage::Project(mir::Project {
            is_add_fields: false,
            source: util::mir_project_collection(Some("foo"), "schema_coll", Some("q"), None),
            expression: map! {
                (Bottom, 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "bar".to_string() => mir::Expression::Subquery(mir::SubqueryExpr {
                            output_expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference((Bottom, 1u16).into())),
                                field: "bar".to_string(),
                                is_nullable: true,
                            })),
                            subquery: Box::new(mir::Stage::Limit(mir::Limit {
                                source: Box::new(mir::Stage::Project(mir::Project {
            is_add_fields: false,
                                    source: util::mir_project_collection(Some("foo"), "schema_foo", Some("q"), Some(1)),
                                    expression: map! {
                                        (Bottom, 1u16).into() => mir::Expression::Document(mir::DocumentExpr {
                                            document: unchecked_unique_linked_hash_map! {
                                                "bar".to_string() => *util::mir_field_access("q", "bar", true),
                                            },
                                        })
                                    },
                                    cache: mir::schema::SchemaCache::new(),
                                })),
                                limit: 1,
                                cache: mir::schema::SchemaCache::new(),
                            })),
                            is_nullable: true,
                        }),
                    },
                })
            },
            cache: mir::schema::SchemaCache::new(),
        })
    );
}
