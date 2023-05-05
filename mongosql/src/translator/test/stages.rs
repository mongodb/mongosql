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

mod filter {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        basic,
        expected = Ok(air::Stage::Match(air::Match {
            source: air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Documents(air::Documents { array: vec![] })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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

mod project {
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                }
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("foo".to_string().into())),
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
        project_with_user_bot_conflict,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                }
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "___bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("foo".to_string().into())),
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
}

mod group {
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
                    })
                    .into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                    }
                })
                .into(),
                keys: vec![air::NameExprPair {
                    name: "x_key".into(),
                    expr: air::Expression::FieldRef("foo".to_string().into())
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
                    })
                    .into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                    }
                })
                .into(),
                keys: vec![air::NameExprPair {
                    name: "x_key".into(),
                    expr: air::Expression::FieldRef("foo".to_string().into())
                },],
                aggregations: vec![
                    air::AccumulatorExpr {
                        alias: "max_distinct".into(),
                        function: air::AggregationFunction::Max,
                        distinct: true,
                        arg: air::Expression::FieldRef("foo".to_string().into()).into()
                    },
                    air::AccumulatorExpr {
                        alias: "min_nondistinct".into(),
                        function: air::AggregationFunction::Min,
                        distinct: false,
                        arg: air::Expression::FieldRef("foo".to_string().into()).into()
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
                    })
                    .into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                    }
                })
                .into(),
                keys: vec![
                    air::NameExprPair {
                        name: "__unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo".to_string().into())
                    },
                    air::NameExprPair {
                        name: "___unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo.x".to_string().into())
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
                    })
                    .into(),
                    specifications: unchecked_unique_linked_hash_map! {
                        "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                    }
                })
                .into(),
                keys: vec![
                    air::NameExprPair {
                        name: "__unaliasedKey2".into(),
                        expr: air::Expression::FieldRef("foo".to_string().into())
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
            aggregations: vec![mir::AliasedAggregation {
                alias: "_id".into(),
                agg_expr: mir::AggregationExpr::CountStar(false),
            },],
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

mod limit {
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
                    "col".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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

mod offset {
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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

mod sort {
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
                            "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                        }
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map!{
                        "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("foo".to_string().into())),
                        "bar".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                        "baz".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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
                "foo".to_string() => air::ProjectItem::Assignment(translator::ROOT.clone()),
            },
        })),
        input = mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod array {
    use crate::unchecked_unique_linked_hash_map;

    test_translate_stage!(
        non_empty,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Documents(air::Documents {
                array: vec![air::Expression::Literal(air::LiteralValue::Boolean(false))],
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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
                "foo".to_string() => air::ProjectItem::Assignment(translator::ROOT.clone()),
            },
        })),
        input = mir::Stage::Array(mir::ArraySource {
            array: vec![],
            alias: "foo".into(),
            cache: mir::schema::SchemaCache::new(),
        })
    );
}

mod join {
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
                collection_name.to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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
                expr: Box::new(air::Expression::FieldRef("foo".to_string().into())),
            }]),
            condition: Some(air::Expression::SQLSemanticOperator(
                air::SQLSemanticOperator {
                    op: air::SQLOperator::Eq,
                    args: vec![
                        air::Expression::Variable("vfoo_0".to_string().into()),
                        air::Expression::FieldRef("bar".to_string().into()),
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
                    "col".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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
                    "arr".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
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

mod set {
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
                    "a".to_string() => air::ProjectItem::Assignment(translator::ROOT.clone()),
                },
            })),
            pipeline: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "bar".into(),
                    collection: "b".into(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "b".to_string() => air::ProjectItem::Assignment(translator::ROOT.clone()),
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

mod derived {
    use crate::{map, unchecked_unique_linked_hash_map};
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    test_translate_stage!(
        derived,
        expected = Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Project(air::Project {
                source: Box::new(air::Stage::Collection(air::Collection {
                    db: "test_db".to_string(),
                    collection: "foo".to_string()
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into()))
                }
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("foo".to_string().into())),
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
}

mod unwind {
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                }
            })),
            path: Box::new(air::Expression::FieldRef("foo".to_string().into())),
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                }
            })),
            path: Box::new(air::Expression::FieldRef("foo.bar".to_string().into())),
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
                    "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                }
            })),
            path: Box::new(air::Expression::FieldRef("foo.bar".to_string().into())),
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
                            "foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                        }
                    }).into(),
                    specifications: unchecked_unique_linked_hash_map!{
                        "___bot".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("foo".to_string().into())),
                        "____bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(4))),
                        "__bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
                        "_bot".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    }
                }).into(),
                new_root: air::Expression::UnsetField(air::UnsetField {
                    field: "___bot".to_string(),
                    input: air::Expression::SetField(air::SetField {
                        field: "".to_string(),
                        input: air::Expression::Variable("ROOT".to_string().into()).into(),
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
        })
    );
}

mod subquery_expr {
    use crate::{
        map, mir::binding_tuple::DatasourceName::Bottom, unchecked_unique_linked_hash_map,
    };

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
                        "schema_coll".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                    },
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "q".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("schema_coll".to_string().into())),
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
                                    source: Box::new(air::Stage::Project(air::Project {
                                        source: Box::new(air::Stage::Collection(air::Collection {
                                            db: "foo".to_string(),
                                            collection: "schema_foo".to_string(),
                                        })),
                                        specifications: unchecked_unique_linked_hash_map! {
                                            "schema_foo".to_string() => air::ProjectItem::Assignment(air::Expression::Variable("ROOT".to_string().into())),
                                        },
                                    })),
                                    specifications: unchecked_unique_linked_hash_map! {
                                        "q".to_string() => air::ProjectItem::Assignment(air::Expression::FieldRef("schema_foo".to_string().into())),
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
}
