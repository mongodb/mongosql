macro_rules! test_translate_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{air, mir, options, translator};
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
                source: Box::new(air::Stage::Documents(air::Documents { array: vec![] })),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Integer(42))),
            }
        ))),
        input = mir::Stage::Filter(mir::Filter {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            condition: mir::Expression::Literal(mir::LiteralValue::Integer(42).into()),
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
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                "bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1)))
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
        project_key_contains_dots,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo.bar.baz".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "foo_bar_baz".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
            }
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo.bar.baz".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::named("foo.bar.baz", 0u16) => mir::Expression::Reference(("foo.bar.baz", 0u16).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_key_starts_with_dollar,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "$foo".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "_foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
            }
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "$foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::named("$foo", 0u16) => mir::Expression::Reference(("$foo", 0u16).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_key_is_empty,
        expected = Err(Error::InvalidProjectField),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::named("", 0u16) => mir::Expression::Reference(("", 0u16).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_key_starts_with_special_characters_conflicts_with_other_project_key,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "$foo.bar".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__foo_bar".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                "_foo_bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
            }
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "$foo.bar".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::named("$foo.bar", 0u16) => mir::Expression::Reference(("$foo.bar", 0u16).into()),
                Key::named("_foo_bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );

    test_translate_stage!(
        project_with_user_bot_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "___bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                // reordered because BindingTuple uses BTreeMap
                "____bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(4))),
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
                "_bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
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

    test_translate_stage!(
        select_values_non_literal_document_expr_correctness_test,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "mydb".to_string(),
                    collection: "foo".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "t1".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                ),
            })),
            specifications: unchecked_unique_linked_hash_map!(
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("t1.b".to_string().into())),
            ),
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "mydb".to_string(),
                    collection: "foo".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    Key::named("t1", 0u16) => mir::Expression::Reference(mir::ReferenceExpr {
                        key: Key::named("foo", 0u16),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                },
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                    expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess{
                        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                            key: Key::named("t1", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                        field: "b".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                        is_nullable: false,
                    })),
                    target_type: mir::Type::Document,
                    cache: mir::schema::SchemaCache::new(),
                }),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod group {
    use crate::{map, translator::Error, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::Key;

    test_translate_stage!(
        group_count_star,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".into(),
                    collection: "foo".into()
                })),
                keys: vec![air::NameExprPair {
                    name: "x_key".into(),
                    expr: ROOT.clone()
                },],
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
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                alias: "x_key".into(),
                expr: mir::Expression::Reference(mir::ReferenceExpr {
                    key: Key::named("foo", 0u16),
                    cache: mir::schema::SchemaCache::new(),
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
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".into(),
                    collection: "foo".into()
                })),
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
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            keys: vec![mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                alias: "x_key".into(),
                expr: mir::Expression::Reference(mir::ReferenceExpr {
                    key: Key::named("foo", 0u16),
                    cache: mir::schema::SchemaCache::new(),
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
            scope: 0,
        })
    );

    test_translate_stage!(
        group_key_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: air::Stage::Group(air::Group {
                source: Box::new(air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test_db".into(),
                        collection: "foo".into()
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone())
                    },
                })),
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
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                },
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
                        is_nullable: false,
                    }
                )),
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
                source: Box::new(air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test_db".into(),
                        collection: "foo".into()
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    },
                })),
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
                    arg: air::Expression::Literal(air::LiteralValue::Integer(1)).into(),
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
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                },
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
                        is_nullable: false,
                    }
                )),
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
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".into(),
                collection: "col".into(),
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

mod offset {
    test_translate_stage!(
        simple,
        expected = Ok(air::Stage::Skip(air::Skip {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string()
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

mod sort {
    use crate::{map, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        sort_stage_multi_key_reference,
        expected = Ok(air::Stage::Sort(air::Sort {
                source: air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test_db".to_string(),
                        collection: "foo".to_string(),
                    })),
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
                    mir::FieldPath {
                        key: ("bar", 0u16).into(),
                        fields: vec!["a".to_string()],
                        cache: mir::schema::SchemaCache::new(),
                        is_nullable: false,
                    }
                ),
                mir::SortSpecification::Asc(
                    mir::FieldPath {
                        key: ("baz", 0u16).into(),
                        fields: vec!["a".to_string()],
                        cache: mir::schema::SchemaCache::new(),
                        is_nullable: false,
                    }
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
                    "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                }
            })
            .into(),

            specs: vec![
                air::SortSpecification::Desc("foo.bar".to_string()),
                air::SortSpecification::Asc("foo.baz".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: mir::Stage::Project(mir::Project {
                source: mir::Stage::Collection(mir::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })
                .into(),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            specs: vec![
                mir::SortSpecification::Desc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["bar".to_string()],
                    cache: mir::schema::SchemaCache::new(),
                    is_nullable: false,
                }),
                mir::SortSpecification::Asc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["baz".to_string()],
                    cache: mir::schema::SchemaCache::new(),
                    is_nullable: false,
                })
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
                    "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                }
            })
            .into(),
            specs: vec![
                air::SortSpecification::Desc("foo.bar.quz".to_string()),
                air::SortSpecification::Asc("foo.baz.fizzle.bazzle".to_string())
            ],
        })),
        input = mir::Stage::Sort(mir::Sort {
            source: mir::Stage::Project(mir::Project {
                source: mir::Stage::Collection(mir::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })
                .into(),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            })
            .into(),
            specs: vec![
                mir::SortSpecification::Desc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["bar".to_string(), "quz".to_string()],
                    cache: mir::schema::SchemaCache::new(),
                    is_nullable: false,
                }),
                mir::SortSpecification::Asc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec![
                        "baz".to_string(),
                        "fizzle".to_string(),
                        "bazzle".to_string()
                    ],
                    cache: mir::schema::SchemaCache::new(),
                    is_nullable: false,
                })
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
            array: vec![mir::Expression::Literal(
                mir::LiteralValue::Boolean(false).into()
            )],
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
                        "a".to_string() => mir::Expression::Literal(mir::LiteralExpr {
                            value: mir::LiteralValue::Integer(1),
                            cache: mir::schema::SchemaCache::new(),
                        }),
                    },
                    cache: mir::schema::SchemaCache::new(),
                }),
                mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Literal(mir::LiteralExpr {
                            value: mir::LiteralValue::Integer(2),
                            cache: mir::schema::SchemaCache::new(),
                        }),
                    },
                    cache: mir::schema::SchemaCache::new(),
                }),
                mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a".to_string() => mir::Expression::Literal(mir::LiteralExpr {
                            value: mir::LiteralValue::Integer(3),
                            cache: mir::schema::SchemaCache::new(),
                        }),
                    },
                    cache: mir::schema::SchemaCache::new(),
                }),
            ],
            alias: "arr".to_string(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod join {
    use crate::{air, map, mir, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    fn input_collection(collection_name: &str) -> Box<mir::Stage> {
        Box::new(mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: collection_name.into(),
            cache: mir::schema::SchemaCache::new(),
        }))
    }

    fn transformed_collection(collection_name: &str) -> Box<air::Stage> {
        Box::new(air::Stage::Collection(air::Collection {
            db: "test_db".into(),
            collection: collection_name.into(),
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
        left_join_with_condition,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Left,
            left: transformed_collection("foo"),
            right: transformed_collection("bar"),
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
                    is_nullable: true,
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
                left: Box::new(air::Stage::Project(air::Project {
                    source: transformed_collection("Foo"),
                    specifications: unchecked_unique_linked_hash_map! {
                        "Foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    },
                })),
                right: Box::new(air::Stage::Project(air::Project {
                    source: transformed_collection("foo"),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    },
                })),
                condition: None,
                let_vars: None,
            })),
            right: transformed_collection("bar"),
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
            condition: Some(mir::Expression::Literal(
                mir::LiteralValue::Boolean(true).into()
            )),
            left: mir::Stage::Join(mir::Join {
                condition: None,
                left: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".to_string(),
                        collection: "Foo".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {
                        ("Foo", 0u16).into() => mir::Expression::Reference(("Foo", 0u16).into()),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })),
                right: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".to_string(),
                        collection: "foo".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {
                        ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })),
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
            left: Box::new(air::Stage::Collection(air::Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            right: Box::new(air::Stage::Documents(air::Documents {
                array: vec![
                    air::Expression::Literal(air::LiteralValue::Integer(1)),
                    air::Expression::Literal(air::LiteralValue::Integer(1))
                ]
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

    test_translate_stage!(
            joins_retain_aliases_for_left_and_right,
            expected = Ok(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Join(air::Join {
                    left: air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Collection(
                            air::Collection{
                                db:"test_db".to_string(), collection:"foo".to_string()
                            })),
                            specifications: unchecked_unique_linked_hash_map!(
                                "t1".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                            )
                        }).into(),
                    right: air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Collection(
                            air::Collection{
                                db:"test_db".to_string(), collection:"bar".to_string()
                            })),
                            specifications: unchecked_unique_linked_hash_map!(
                                "t2".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                            )
                        }).into(),
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
                source: Box::new(mir::Stage::Join(mir::Join {
                    join_type: mir::JoinType::Inner,
                    left: mir::Stage::Project(mir::Project {
                        source: input_collection("foo"),
                        expression: map! {
                            Key::named("t1", 0u16) => mir::Expression::Reference(("foo".to_string(), 0u16).into()),
                        },
                        cache: mir::schema::SchemaCache::new(),
                    }).into(),
                    right: mir::Stage::Project(mir::Project {
                        source: input_collection("bar"),
                        expression: map! {
                            Key::named("t2", 0u16) => mir::Expression::Reference(("bar".to_string(), 0u16).into()),
                        },
                        cache: mir::schema::SchemaCache::new(),
                    }).into(),
                    condition: Some(mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Eq,
                            args: vec![
                                mir::Expression::FieldAccess(mir::FieldAccess{
                                    expr: mir::Expression::Reference(("t1".to_string(), 0u16).into()).into(),
                                    field: "a".to_string(),
                                    cache: mir::schema::SchemaCache::new(),
                                    is_nullable: false,
                                }),
                                mir::Expression::FieldAccess(mir::FieldAccess{
                                    expr: mir::Expression::Reference(("t2".to_string(), 0u16).into()).into(),
                                    field: "b".to_string(),
                                    cache: mir::schema::SchemaCache::new(),
                                    is_nullable: false,
                                }),
                            ],
                            cache: mir::schema::SchemaCache::new(),
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
                left: Box::new(air::Stage::Project(air::Project {
                    source: transformed_collection("foo"),
                    specifications: unchecked_unique_linked_hash_map! {
                        "Foo".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    },
                })),
                right: Box::new(air::Stage::Project(air::Project {
                    source: transformed_collection("bar"),
                    specifications: unchecked_unique_linked_hash_map! {
                        "_bar".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                    },
                })),
                condition: None,
                let_vars: None,
            })),
            right: transformed_collection("bar"),
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
            condition: Some(mir::Expression::Literal(
                mir::LiteralValue::Boolean(true).into()
            )),
            left: mir::Stage::Join(mir::Join {
                condition: None,
                left: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".to_string(),
                        collection: "foo".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {
                        ("Foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })),
                right: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".to_string(),
                        collection: "bar".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {
                        ("_bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })),
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
}

mod equijoin {
    use crate::{
        map,
        translator::Error,
        util::{air_collection, air_pipeline_collection, mir_collection, mir_field_path},
    };
    use mongosql_datastructures::binding_tuple::Key;

    test_translate_stage!(
        inner_join,
        expected = Ok(air::Stage::EquiJoin(air::EquiJoin {
            join_type: air::JoinType::Inner,
            source: air_pipeline_collection("foo"),
            from: air_collection("bar"),
            local_field: "foo.a".into(),
            foreign_field: "a".into(),
            as_name: "bar".to_string(),
        })),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::EquiJoin(mir::EquiJoin {
            join_type: mir::JoinType::Inner,
            source: mir_collection("foo"),
            from: mir_collection("bar"),
            local_field: Box::new(mir_field_path("foo", vec!["a"])),
            foreign_field: Box::new(mir_field_path("bar", vec!["a"])),
            cache: mir::schema::SchemaCache::new(),
        }))
    );

    test_translate_stage!(
        left_join,
        expected = Ok(air::Stage::EquiJoin(air::EquiJoin {
            join_type: air::JoinType::Left,
            source: air_pipeline_collection("foo"),
            from: air_collection("bar"),
            local_field: "foo.a".into(),
            foreign_field: "a".into(),
            as_name: "x".to_string(),
        })),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::EquiJoin(mir::EquiJoin {
            join_type: mir::JoinType::Left,
            source: mir_collection("foo"),
            from: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test_db".to_string(),
                    collection: "bar".to_string(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: map! {
                    Key::named("x", 0u16) => mir::Expression::Reference(("bar".to_string(), 0u16).into()),
                },
                cache: mir::schema::SchemaCache::new(),
            })),
            local_field: Box::new(mir_field_path("foo", vec!["a"])),
            foreign_field: Box::new(mir_field_path("x", vec!["a"])),
            cache: mir::schema::SchemaCache::new(),
        }))
    );

    test_translate_stage!(
        from_must_be_collection,
        expected = Err(Error::ExpectedCollection),
        input = mir::Stage::MQLIntrinsic(mir::MQLStage::EquiJoin(mir::EquiJoin {
            join_type: mir::JoinType::Left,
            source: mir_collection("foo"),
            from: mir::Stage::Array(mir::ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: mir::schema::SchemaCache::new()
            })
            .into(),
            local_field: Box::new(mir_field_path("foo", vec!["a"])),
            foreign_field: Box::new(mir_field_path("bar", vec!["a"])),
            cache: mir::schema::SchemaCache::new(),
        }))
    );
}

mod lateral_join {
    use crate::{
        map, unchecked_unique_linked_hash_map,
        util::{air_pipeline_collection, mir_collection, ROOT},
    };
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        lateral_join_inner_simple,
        expected = Ok(air::Stage::Join(air::Join {
            join_type: air::JoinType::Inner,
            left: air_pipeline_collection("foo"),
            right: air_pipeline_collection("bar"),
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
            source: mir_collection("foo"),
            subquery: mir_collection("bar"),
            cache: mir::schema::SchemaCache::new(),
        }))
    );

    test_translate_stage!(
        left_lateral_join_let_vars,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Join(air::Join {
                join_type: air::JoinType::Left,
                left: air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(
                        air::Collection{
                            db:"test_db".to_string(), collection:"foo".to_string()
                        })),
                        specifications: unchecked_unique_linked_hash_map!(
                            "t1".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                        )
                    }).into(),
                right: air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(
                        air::Collection{
                            db:"test_db".to_string(), collection:"bar".to_string()
                        })),
                        specifications: unchecked_unique_linked_hash_map!()
                    }).into(),
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
            source: Box::new(mir::Stage::MQLIntrinsic(mir::MQLStage::LateralJoin(mir::LateralJoin {
                join_type: mir::JoinType::Left,
                source: mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".into(),
                        collection: "foo".into(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {
                        Key::named("t1", 0u16) => mir::Expression::Reference(("foo".to_string(), 0u16).into()),
                    },
                    cache: mir::schema::SchemaCache::new(),
                }).into(),
                subquery: mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".into(),
                        collection: "bar".into(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: map! {},
                    cache: mir::schema::SchemaCache::new(),
                }).into(),
                cache: mir::schema::SchemaCache::new(),
            }))),
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
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "foo".into(),
                collection: "a".into(),
            })),
            pipeline: Box::new(air::Stage::Collection(air::Collection {
                db: "bar".into(),
                collection: "b".into(),
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

mod derived {
    use crate::{map, unchecked_unique_linked_hash_map, util::ROOT};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        derived,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                "bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1)))
            }
        })),
        input = mir::Stage::Derived(mir::Derived {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Collection(mir::Collection {
                    db: "test_db".into(),
                    collection: "foo".into(),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::Reference(("foo", 1u16).into()),
                    Key::named("bar", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
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
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "foo".to_string(),
                    collection: "bar".to_string()
                })),
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
                source: Box::new(mir::Stage::Derived(mir::Derived {
                    source: Box::new(mir::Stage::Project(mir::Project {
                        source: Box::new(mir::Stage::Collection(mir::Collection {
                            db: "foo".to_string(),
                            collection: "bar".to_string(),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                        expression: BindingTuple(map! {
                            Key::bot(1) => mir::Expression::Reference(mir::ReferenceExpr {
                                key: Key::named("bar", 2u16),
                                cache: mir::schema::SchemaCache::new(),
                            }),
                            Key::named("d2", 0u16) => mir::Expression::Literal(mir::LiteralValue::Integer(1).into()),
                        }),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    cache: mir::schema::SchemaCache::new(),
                })),
                expression: BindingTuple(map! {
                    Key::bot(0) => mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: Key::named("d2", 0u16),
                                cache: mir::schema::SchemaCache::new(),
                            })),
                            field: "a".to_string(),
                            cache: mir::schema::SchemaCache::new(),
                            is_nullable: false,
                        })),
                        target_type: mir::Type::Int32,
                        cache: mir::schema::SchemaCache::new(),
                    }),
                }),
                cache: mir::schema::SchemaCache::new(),
            })),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod unwind {
    use crate::{
        air::ExprLanguage, map, unchecked_unique_linked_hash_map, util::air_variable_from_root,
    };
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage! {
        unwind,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string()
            })),
            path: air::Expression::Variable("ROOT.bar".to_string().into()),
            index: None,
            outer: false,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(mir::Stage::Collection(mir::Collection{
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            path: mir::FieldPath {
                key: ("foo",0u16).into(),
                fields: vec!["bar".to_string()],
                cache: mir::schema::SchemaCache::new(),
                is_nullable: false,
            },
            index: None,
            outer: false,
            cache: mir::schema::SchemaCache::new(),
        })
    }

    test_translate_stage! {
        unwind_outer,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string()
            })),
            path: air::Expression::Variable("ROOT.bar".to_string().into()),
            index: None,
            outer: true,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(mir::Stage::Collection(mir::Collection{
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            path: mir::FieldPath {
                key: ("foo",0u16).into(),
                fields: vec!["bar".to_string()],
                cache: mir::schema::SchemaCache::new(),
                is_nullable: false,
            },
            index: None,
            outer: true,
            cache: mir::schema::SchemaCache::new(),
        })
    }
    test_translate_stage! {
        unwind_index,
        expected = Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test_db".to_string(),
                collection: "foo".to_string(),
            })),
            path: air::Expression::Variable("ROOT.bar".to_string().into()),
            index: Some("i".to_string()),
            outer: true,
        })),
        input = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(mir::Stage::Collection(mir::Collection{
                db: "test_db".into(),
                collection: "foo".into(),
                cache: mir::schema::SchemaCache::new(),
            })),
            path: mir::FieldPath {
                key: ("foo",0u16).into(),
                fields: vec!["bar".to_string()],
                cache: mir::schema::SchemaCache::new(),
                is_nullable: false,
            },
            index: Some("i".into()),
            outer: true,
            cache: mir::schema::SchemaCache::new(),
        })
    }

    test_translate_stage! {
        correctness_test_for_index_option_using_project_and_where,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Match(air::Match::ExprLanguage(ExprLanguage {
                source: Box::new(air::Stage::Unwind(air::Unwind {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test".to_string(),
                        collection: "foo".to_string(),
                    })),
                    path: air_variable_from_root("arr"),
                    index: Some("idx".to_string()),
                    outer: false,
                })),
                expr: Box::new(air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::Gt,
                    args: vec![
                        air_variable_from_root("arr"),
                        air::Expression::Literal(air::LiteralValue::Integer(0)),
                    ],
                })),
            }))),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Document(unchecked_unique_linked_hash_map! {
                    "arr".to_string() => air_variable_from_root("arr"),
                    "idx".to_string() => air_variable_from_root("idx")
                })),
            },
        })),
        input = mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Filter(mir::Filter {
                source: Box::new(mir::Stage::Unwind(mir::Unwind {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test".into(),
                        collection: "foo".into(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    path: mir::FieldPath {
                        key: ("foo", 0u16).into(),
                        fields: vec!["arr".to_string()],
                        cache: mir::schema::SchemaCache::new(),
                        is_nullable: true,
                    },
                    index: Some("idx".into()),
                    outer: false,
                    cache: mir::schema::SchemaCache::new(),
                })),
                condition: mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Gt,
                    args: vec![
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                            field: "arr".into(),
                            cache: mir::schema::SchemaCache::new(),
                            is_nullable: true,
                        }),
                        mir::Expression::Literal(mir::LiteralValue::Integer(0).into()),
                    ],
                    cache: mir::schema::SchemaCache::new(),
                    is_nullable: true,
                }),
                cache: mir::schema::SchemaCache::new(),
            })),
            expression: BindingTuple(map! {
                Key::bot(0) => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "arr".into() => mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                            field: "arr".into(),
                            cache: mir::schema::SchemaCache::new(),
                            is_nullable: false,
                        }),
                        "idx".into() => mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                            field: "idx".into(),
                            cache: mir::schema::SchemaCache::new(),
                            is_nullable: false,
                        })
                    },
                    cache: mir::schema::SchemaCache::new(),
                }),
            }),
            cache: mir::schema::SchemaCache::new(),
        })
    }
}

mod mql_intrinsic {
    mod match_filter {
        use crate::{map, unchecked_unique_linked_hash_map, util::ROOT};
        use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

        test_translate_stage!(
            basic,
            expected = Ok(air::Stage::Match(air::Match::MatchLanguage(
                air::MatchLanguage {
                    source: Box::new(air::Stage::Project(air::Project {
                        source: Box::new(air::Stage::Collection(air::Collection {
                            db: "test_db".to_string(),
                            collection: "foo".to_string()
                        })),
                        specifications: unchecked_unique_linked_hash_map! {
                            "f".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                        },
                    })),
                    expr: Box::new(air::MatchQuery::Comparison(air::MatchLanguageComparison {
                        function: air::MatchLanguageComparisonOp::Lt,
                        input: Some("f.a".to_string().into()),
                        arg: air::LiteralValue::Integer(1),
                    })),
                }
            ))),
            input = mir::Stage::MQLIntrinsic(mir::MQLStage::MatchFilter(mir::MatchFilter {
                source: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Collection(mir::Collection {
                        db: "test_db".into(),
                        collection: "foo".into(),
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    expression: BindingTuple(map! {
                        Key::named("f", 0u16) => mir::Expression::Reference(("foo", 0u16).into()),
                    }),
                    cache: mir::schema::SchemaCache::new(),
                })),
                condition: mir::MatchQuery::Comparison(mir::MatchLanguageComparison {
                    function: mir::MatchLanguageComparisonOp::Lt,
                    input: Some(mir::FieldPath {
                        key: ("f", 0u16).into(),
                        fields: vec!["a".to_string()],
                        cache: mir::schema::SchemaCache::new(),
                        is_nullable: false,
                    }),
                    arg: mir::LiteralValue::Integer(1),
                    cache: mir::schema::SchemaCache::new(),
                }),
                cache: mir::schema::SchemaCache::new(),
            }))
        );
    }
}

mod translate_plan {
    use crate::{
        map, mir, unchecked_unique_linked_hash_map,
        util::{
            air_project_bot_collection, air_project_collection,
            air_project_collection_with_expected_rename, mir_collection,
            mir_project_bot_collection, mir_project_collection, sql_options_exclude_namespaces,
            ROOT,
        },
    };
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_plan!(
        project_with_user_bot_conflict,
        expected = Ok(
            air::Stage::ReplaceWith(air::ReplaceWith {
                source: air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test_db".to_string(),
                        collection: "foo".to_string(),
                    })),
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
                            left: air_project_collection_with_expected_rename("$foo", "_foo"),
                            right: air_project_collection_with_expected_rename(
                                "bar.baz", "bar_baz"
                            ),
                            let_vars: None,
                            condition: None,
                        })),
                        right: air_project_collection_with_expected_rename("$_foo", "__foo"),
                        let_vars: None,
                        condition: None,
                    })),
                    right: air_project_collection_with_expected_rename("_foo", "___foo"),
                    let_vars: None,
                    condition: None,
                })),
                right: air_project_collection("bar"),
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
                        left: mir_project_collection("$foo"),
                        right: mir_project_collection("bar.baz"),
                        condition: None,
                        cache: mir::schema::SchemaCache::new(),
                    })),
                    right: mir_project_collection("$_foo"),
                    condition: None,
                    cache: mir::schema::SchemaCache::new(),
                })),
                right: mir_project_collection("_foo"),
                condition: None,
                cache: mir::schema::SchemaCache::new(),
            })),
            right: mir_project_collection("bar"),
            condition: None,
            cache: mir::schema::SchemaCache::new(),
        }),
    );

    test_translate_plan!(
        single_non_namespaced_results,
        expected = Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: air_project_collection("foo"),
            new_root: Box::new(air::Expression::FieldRef("foo".into())),
        })),
        input = *mir_collection("foo"),
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
                    left: air_project_collection("foo"),
                    right: air_project_collection("bar"),
                    let_vars: None,
                    condition: None,
                })),
                right: air_project_collection("baz"),
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
                left: mir_project_collection("foo"),
                right: mir_project_collection("bar"),
                condition: None,
                cache: mir::schema::SchemaCache::new(),
            })),
            right: mir_project_collection("baz"),
            condition: None,
            cache: mir::schema::SchemaCache::new(),
        }),
        options = sql_options_exclude_namespaces(),
    );

    test_translate_plan!(
        non_namespaced_handles_bot,
        expected = Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: air_project_bot_collection("foo"),
            new_root: Box::new(air::Expression::FieldRef("__bot".into()))
        })),
        input = *mir_project_bot_collection("foo"),
        options = sql_options_exclude_namespaces(),
    );
}

mod subquery_expr {
    use crate::{
        map, mir::binding_tuple::DatasourceName::Bottom, unchecked_unique_linked_hash_map,
        util::ROOT,
    };

    test_translate_stage!(
        unqualified_correlated_reference,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "foo".to_string(),
                    collection: "schema_coll".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "q".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                },
            })),
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
                                source: Box::new(air::Stage::Project(air::Project {
                                    source: Box::new(air::Stage::Collection(air::Collection {
                                        db: "foo".to_string(),
                                        collection: "schema_foo".to_string(),
                                    })),
                                    specifications: unchecked_unique_linked_hash_map! {
                                        "q".to_string() => air::ProjectItem::Assignment(ROOT.clone()),
                                    },
                                })),
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
                                is_nullable: false,
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
                                                    is_nullable: false,
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
                            is_nullable: false,
                        }),
                    },
                    cache: mir::schema::SchemaCache::new(),
                })
            },
            cache: mir::schema::SchemaCache::new(),
        })
    );
}
