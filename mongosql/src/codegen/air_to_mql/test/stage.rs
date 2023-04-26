macro_rules! test_codegen_air_stage {
    (
		$func_name:ident,
		expected = Ok({
			database: $expected_db:expr,
			collection: $expected_collection:expr,
			pipeline: $expected_pipeline:expr,
		}),
		input = $input:expr,
	) => {
        #[test]
        fn $func_name() {
            use crate::codegen::{air_to_mql::MqlTranslation, generate_mql_from_air};

            let input = $input;
            let expected_db = $expected_db;
            let expected_collection = $expected_collection;
            let expected_pipeline = $expected_pipeline;

            let MqlTranslation {
                database: db,
                collection: col,
                pipeline: pipeline,
            } = generate_mql_from_air(input).expect("codegen failed");

            assert_eq!(expected_db, db);
            assert_eq!(expected_collection, col);
            assert_eq!(expected_pipeline, pipeline);
        }
    };

    ($func_name:ident, expected = Err($expected_err:expr), input = $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::generate_mql_from_air;

            let input = $input;
            let expected = Err($expected_err);

            assert_eq!(expected, generate_mql_from_air(input));
        }
    };
}

mod union_with {
    use crate::air::*;
    use bson::doc;

    test_codegen_air_stage!(
        collection_union_with_collection,
        expected = Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![
                doc!{"$unionWith": {"coll": "b", "pipeline": []}}],
        }),
        input = Stage::UnionWith(UnionWith {
            source: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
            }).into(),
            pipeline: Stage::Collection(Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
            }).into(),
        }),
    );
    test_codegen_air_stage!(
        array_union_with_array,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                doc!{"$documents": [{"$literal": 1}]},
                doc!{"$unionWith": {"pipeline": [
                    {"$documents": [{"$literal": 2}]},
                ]}}],
        }),
        input = Stage::UnionWith(UnionWith {
            source: Stage::Documents(Documents {
                array: vec![Expression::Literal(LiteralValue::Integer(1))],
            }).into(),
            pipeline: Stage::Documents(Documents {
                array: vec![Expression::Literal(LiteralValue::Integer(2))],
            }).into(),
        }),
    );
    test_codegen_air_stage!(
        array_union_with_collection,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                doc!{"$documents": [{"$literal": 1}]},
                doc!{"$unionWith": {"coll": "b", "pipeline": []}}],
        }),
        input = Stage::UnionWith(UnionWith {
            source: Stage::Documents(Documents {
                array: vec![Expression::Literal(LiteralValue::Integer(1))],
            }).into(),
            pipeline: Stage::Collection(Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
            }).into(),
        }),
    );
    test_codegen_air_stage!(
        collection_union_with_array,
        expected = Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![
                bson::doc!{"$unionWith": {"pipeline": [
                    {"$documents": [{"$literal": 1}]},
                ]}}],
        }),
        input = Stage::UnionWith(UnionWith {
            source: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
            }).into(),
            pipeline: Stage::Documents(Documents {
                array: vec![Expression::Literal(LiteralValue::Integer(1))],
            }).into(),
        }),
    );
    test_codegen_air_stage!(
        collection_union_with_nested_union_with,
        expected = Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![
                doc!{"$unionWith": {"coll": "b", "pipeline": [
                    {"$unionWith": {"pipeline": [
                        {"$documents": [{"$literal": 1}]},
                    ]}}
                ]}}
            ],
        }),
        input = Stage::UnionWith(UnionWith {
            source: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
            }).into(),
            pipeline: Stage::UnionWith(UnionWith {
                source: Stage::Collection(Collection {
                    db: "bar".to_string(),
                    collection: "b".to_string(),
                }).into(),
                pipeline: Stage::Documents(Documents {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))],
                }).into(),
            }).into(),
        }),
    );
}

mod sort {
    use crate::{
        air::Expression::*, air::LiteralValue::*, air::SortSpecification::*, air::*,
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_stage!(
        empty,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$sort": {}},
            ],
        }),
        input = Stage::Sort(Sort {
            specs: vec![],
            source: Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
        }),
    );

    test_codegen_air_stage!(
        single_spec,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [bson!({"foo": {"$literal": 1}}), bson!({"foo": {"$literal": 2}})]},
                bson::doc!{"$sort": {"foo": 1}}
            ],
        }),
        input = Stage::Sort(Sort {
            specs: vec![Asc("foo".to_string())],
            source: Box::new(
                Stage::Documents(Documents {
                    array: vec![Document(unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Integer(1))}),
                                Document(unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Integer(2))})],
                })
            )
        }),
    );

    test_codegen_air_stage!(
        multi_spec,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [bson!({"foo": {"$literal": 1}, "bar": {"$literal": 3}}),
                                        bson!({"foo": {"$literal": 2}, "bar": {"$literal": 4}})]},
                bson::doc!{"$sort": {"foo": -1, "bar": 1}}
            ],
        }),
        input = Stage::Sort(Sort {
            specs: vec![Desc("foo".to_string()), Asc("bar".to_string())],
            source: Box::new(
                Stage::Documents(Documents {
                    array: vec![Document(unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Integer(1)), "bar".to_string() => Literal(Integer(3))}),
                                Document(unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Integer(2)), "bar".to_string() => Literal(Integer(4))})],
                })
            )
        }),
    );
}

mod match_stage {
    use crate::air::*;

    use bson::doc;

    test_codegen_air_stage!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![doc!{"$match": {"$expr": { "$eq": [{ "$literal": 1}, { "$literal": 2}]}}}],
        }),
        input = Stage::Match( Match {
            source:Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                    })
            ),
            expr : Box::new(
                Expression::MQLSemanticOperator( MQLSemanticOperator {
                    op: MQLOperator::Eq,
                    args: vec![Expression::Literal(LiteralValue::Integer(1)), Expression::Literal(LiteralValue::Integer(2))]
                })
            )
        }),
    );
}

mod collection {
    use crate::air::*;

    test_codegen_air_stage!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: Vec::<bson::Document>::new(),
        }),
        input = Stage::Collection(Collection {
            db: "mydb".to_string(),
            collection: "col".to_string(),
        }),
    );
}

mod project {
    use crate::{air::*, unchecked_unique_linked_hash_map};
    use bson::doc;

    test_codegen_air_stage!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![doc!{"$project": {"_id": 0, "foo": "$col", "bar": {"$literal": 19}}}],
        }),
        input = Stage::Project(Project {
            source: Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
            specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => Expression::FieldRef(
                    FieldRef { parent: None,  name: "col".to_string() }
                ),
                "bar".to_string() => Expression::Literal(LiteralValue::Integer(19)),
            },
        }),
    );

    test_codegen_air_stage!(
        project_of_id_overwritten,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![doc!{"$project": {"_id": "$col", "bar": {"$literal": 19}}}],
        }),
        input = Stage::Project(Project {
            source: Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
            specifications: unchecked_unique_linked_hash_map! {
                "_id".to_string() => Expression::FieldRef(
                    FieldRef { parent: None,  name: "col".to_string() }
                ),
                "bar".to_string() => Expression::Literal(LiteralValue::Integer(19)),
            },
        }),
    );
}

mod group {
    use crate::air::*;
    use bson::doc;

    test_codegen_air_stage!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$group": {"_id": {"foo": "$foo", "bar": {"$add": ["$bar", {"$literal": 1}]}},
                                "x": {"$min": "$x"},
                                "y": {"$max": {"$mod": ["$x", {"$literal": 1}]}}
                               }
                }
            ],
        }),
        input = Stage::Group(Group {
            source: Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
            keys: vec![
                NameExprPair {
                    name: "foo".into(),
                    expr: Expression::FieldRef(FieldRef {
                        parent: None,
                        name: "foo".into(),
                    })
                },
                NameExprPair {
                    name: "bar".into(),
                    expr: Expression::MQLSemanticOperator( MQLSemanticOperator {
                        op: MQLOperator::Add,
                        args: vec![
                            Expression::FieldRef(FieldRef {
                                parent: None,
                                name: "bar".into(),
                            }),
                            Expression::Literal(LiteralValue::Integer(1))
                        ],
                    })
                },
            ],
            aggregations: vec![
                AccumulatorExpr {
                    alias: "x".into(),
                    function: AggregationFunction::Min,
                    distinct: false,
                    arg: Expression::FieldRef(FieldRef {
                        parent: None,
                        name: "x".into(),
                    }).into(),
                },
                AccumulatorExpr {
                    alias: "y".into(),
                    function: AggregationFunction::Max,
                    distinct: false,
                    arg: Expression::MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Mod,
                        args: vec![
                            Expression::FieldRef(FieldRef {
                                parent: None,
                                name: "x".into(),
                            }),
                            Expression::Literal(LiteralValue::Integer(1i32))
                        ],
                    }).into(),
                },
            ],
        }),
    );

    test_codegen_air_stage!(
        distinct_ops_are_sql_ops,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$group": {"_id": {"foo": "$foo"},
                                "x": {"$sqlMin": {"var": "$x", "distinct": true}},
                               }
                }
            ],
        }),
        input = Stage::Group(Group {
            source: Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
            keys: vec![
                NameExprPair {
                    name: "foo".into(),
                    expr: Expression::FieldRef(FieldRef {
                        parent: None,
                        name: "foo".into(),
                    })
                },
            ],
            aggregations: vec![
                AccumulatorExpr {
                    alias: "x".into(),
                    function: AggregationFunction::Min,
                    distinct: true,
                    arg: Expression::FieldRef(FieldRef {
                        parent: None,
                        name: "x".into(),
                    }).into(),
                },
            ],
        }),
    );

    test_codegen_air_stage!(
        count_is_always_a_sql_op,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$group": {"_id": {"foo": "$foo"},
                                "x": {"$sqlCount": {"var": "$x", "distinct": false}},
                               }
                }
            ],
        }),
        input = Stage::Group(Group {
            source: Box::new(
                Stage::Collection( Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
            keys: vec![
                NameExprPair {
                    name: "foo".into(),
                    expr: Expression::FieldRef(FieldRef {
                        parent: None,
                        name: "foo".into(),
                    })
                },
            ],
            aggregations: vec![
                AccumulatorExpr {
                    alias: "x".into(),
                    function: AggregationFunction::Count,
                    distinct: false,
                    arg: Expression::FieldRef(FieldRef {
                        parent: None,
                        name: "x".into(),
                    }).into(),
                },
            ],
        }),
    );
}

mod unwind {
    use crate::{air::*, unchecked_unique_linked_hash_map};
    use bson::doc;

    test_codegen_air_stage!(
        unwind_with_only_path,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$unwind": {"path": "$array" }}
            ],
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            path: Expression::FieldRef(FieldRef {
                parent: None,
                name: "array".into(),
            }).into(),
            index: None,
            outer: false
        }),
    );

    test_codegen_air_stage!(
        unwind_with_index_string,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$unwind": {"path": "$array", "includeArrayIndex": "i" }}
            ],
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            path: Expression::FieldRef(FieldRef {
                parent: None,
                name: "array".into(),
            }).into(),
            index: Some("i".into()),
            outer: false
        }),
    );

    test_codegen_air_stage!(
        unwind_with_preserve_null_and_empty_arrays,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$unwind": {"path": "$array", "preserveNullAndEmptyArrays": true }}
            ],
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            path: Expression::FieldRef(FieldRef {
                parent: None,
                name: "array".into(),
            }).into(),
            index: None,
            outer: true
        }),
    );
    test_codegen_air_stage!(
        unwind_with_all_args,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$unwind": {"path": "$array", "includeArrayIndex": "i", "preserveNullAndEmptyArrays": true }}
            ],
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            path: Expression::FieldRef(FieldRef {
                parent: None,
                name: "array".into(),
            }).into(),
            index: Some("i".into()),
            outer: true
        }),
    );
    test_codegen_air_stage!(
        unwind_proper_field_paths,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                doc!{"$project": {"_id": 0, "foo": "$col"}},
                doc!{"$unwind": {"path": "$foo.a.b", "includeArrayIndex": "foo.i", "preserveNullAndEmptyArrays": true }}
            ],
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                "foo".to_string() => Expression::FieldRef(
                    FieldRef { parent: None,  name: "col".to_string() }
                )}
            })),
            path: Expression::FieldRef(FieldRef {
                parent:Some(FieldRef {
                    parent: Some(FieldRef {
                        parent: None,
                        name: "foo".into(),
                    }.into()),
                    name: "a".into(),
                }.into()),
                name: "b".into(),
            }).into(),
            index: Some("i".into()),
            outer: true,
        }),
    );
}

mod documents {
    use crate::air::*;

    test_codegen_air_stage!(
        empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
            ],
        }),
        input = Stage::Documents(Documents {
            array: vec![],
        }),
    );
    test_codegen_air_stage!(
        non_empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": false}]},
            ],
        }),
        input = Stage::Documents(Documents {
            array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        }),
    );
}

mod replace_with {
    use crate::{air::*, unchecked_unique_linked_hash_map};

    test_codegen_air_stage!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$replaceWith": {"$literal": "$name"}},
            ],
        }),
        input = Stage::ReplaceWith(ReplaceWith {
            source: Box::new(
                Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
            new_root: Box::new(Expression::Literal(LiteralValue::String("$name".to_string()))),
        }),
    );
    test_codegen_air_stage!(
        document,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {
                    "$replaceWith": {
                        "$mergeDocuments": [
                            {"$literal": "$name"},
                            {"_id": {"$literal": "$_id"}}
                        ]
                    }
                },
            ],
        }),
        input = Stage::ReplaceWith(ReplaceWith {
            source: Box::new(
                Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
            new_root: Box::new(
                Expression::Document(unchecked_unique_linked_hash_map! {
                    "$mergeDocuments".to_string() => Expression::Array(vec![
                        Expression::Literal(LiteralValue::String("$name".to_string())),
                        Expression::Document(unchecked_unique_linked_hash_map! {
                            "_id".to_string() => Expression::Literal(
                                LiteralValue::String("$_id".to_string())
                            )
                        })
                    ])
                })
            ),
        }),
    );
}

mod lookup {
    use crate::air::*;

    macro_rules! test_input {
        ($let_vars:expr) => {
            Stage::Lookup(Lookup {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                let_vars: $let_vars,
                pipeline: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                as_var: "as_var".to_string(),
            })
        };
    }

    test_codegen_air_stage!(
        with_no_from,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$lookup": {"pipeline": [{"$documents": []}], "as": "as_var"}},
            ],
        }),
        input = Stage::Lookup(Lookup {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            let_vars: None,
            pipeline: Box::new(Stage::Documents(Documents {
                array: vec![],
            })),
            as_var: "as_var".to_string()
        }),
    );

    test_codegen_air_stage!(
        with_from_same_database,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$lookup": {"from": "col", "pipeline": [], "as": "as_var"}},
            ],
        }),
        input = test_input!(None),
    );

    test_codegen_air_stage!(
        with_from_clause_different_database,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$lookup": {"from": {"db": "mydb2", "coll": "col2"}, "pipeline": [], "as": "as_var"}},
            ],
        }),
        input = Stage::Lookup(Lookup {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            let_vars: None,
            pipeline: Box::new(Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string()
            })),
            as_var: "as_var".to_string()
        }),
    );

    test_codegen_air_stage!(
        with_single_let_var,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$lookup": {
                    "from": "col",
                    "let": {"x": {"$literal": 9}},
                    "pipeline": [],
                    "as": "as_var"
                }},
            ],
        }),
        input = test_input!(
            Some(vec![LetVariable{name: "x".to_string(), expr: Box::new(Expression::Literal(LiteralValue::Integer(9)))}])
        ),
    );

    test_codegen_air_stage!(
        with_multiple_let_vars,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$lookup": {
                    "from": "col",
                    "let": {
                        "x": {"$literal": 9},
                        "y": "$a"
                    },
                    "pipeline": [],
                    "as": "as_var"
                }},
            ],
        }),
        input = test_input!(
            Some(vec![
                LetVariable{name: "x".to_string(), expr: Box::new(Expression::Literal(LiteralValue::Integer(9)))},
                LetVariable{name: "y".to_string(), expr: Box::new(Expression::FieldRef(FieldRef {
                    parent: None,
                    name: "a".into(),
                }))},
            ])
        ),
    );
}

mod skip {
    use crate::air::*;
    use bson::Bson;

    test_codegen_air_stage!(
        skip,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$skip": Bson::Int64(10)},
            ],
        }),
        input = Stage::Skip(Skip {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            skip: 10,
        }),
    );
}

mod limit {
    use crate::air::*;
    use bson::Bson;

    test_codegen_air_stage!(
        limit,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$limit": Bson::Int64(123)},
            ],
        }),
        input = Stage::Limit(Limit {
            source: Box::new(Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })),
            limit: 123,
        }),
    );
}

mod join {
    use crate::{air::*, unchecked_unique_linked_hash_map};

    test_codegen_air_stage!(
        simple_inner_join,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}},
            bson::doc!{"$join": {"collection": "col2", "joinType": "inner", "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}]}}],
        }),
        input = Stage::Join(Join {
            condition: None,
            left: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            right: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col2".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col2".to_string() =>Expression::Variable("ROOT".to_string().into())
                )
            })),
            let_vars: None,
            join_type: JoinType::Inner,
        }),
    );

    test_codegen_air_stage!(
        left_join_different_databases,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}},
            bson::doc!{"$join": {"database": "mydb2", "collection": "col2", "joinType": "left", "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}]}}],
        }),
        input = Stage::Join(Join {
            condition: None,
            left: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            right: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb2".to_string(),
                    collection: "col2".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col2".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            let_vars: None,
            join_type: JoinType::Left,
        }),
    );

    test_codegen_air_stage!(
        left_join_different_databases_with_condition,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"collection": "col2",
                "joinType":"left",
                "database": "mydb2",
                "let": {"vcol_0": "$col"},
                "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}],
                "condition": {"$match": {"$expr": {"$literal": true}}}
            }}],
        }),
        input = Stage::Join(Join {
            condition: Some(Expression::Literal(LiteralValue::Boolean(true))),
            left: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            right: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb2".to_string(),
                    collection: "col2".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col2".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            let_vars: Some(vec![LetVariable{name: "vcol_0".to_string(), expr: Box::new(Expression::FieldRef(FieldRef{parent: None, name: "col".to_string()}))}]),
            join_type: JoinType::Left,
        }),
    );

    test_codegen_air_stage!(
        join_references_left_and_right,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"collection": "col2",
                "joinType":"left",
                "database": "mydb2",
                "let": {"vcol_0": "$col"},
                "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}],
                "condition": {"$match": {"$expr": {"$sqlEq": ["$$vcol_0", "$col2"]}}}
            }}],
        }),
        input = Stage::Join(Join {
            condition: Some(Expression::SQLSemanticOperator(SQLSemanticOperator {
                op: SQLOperator::Eq,
                args: vec![
                    Expression::Variable("vcol_0".to_string().into()),
                    Expression::FieldRef(FieldRef{parent: None, name: "col2".to_string()}),
                ]
            })),
            left: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            right: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb2".to_string(),
                    collection: "col2".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col2".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            let_vars: Some(vec![LetVariable{name: "vcol_0".to_string(), expr: Box::new(Expression::FieldRef(FieldRef{parent: None, name: "col".to_string()}))}]),
            join_type: JoinType::Left,
        }),
    );

    test_codegen_air_stage!(
        join_with_array, // array sources require no collection or database in the $join
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"joinType":"left",
                "pipeline": [
                    {"$documents": [{"$literal": 1}, {"$literal": 1}]},
                    bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                ],
            }}],
        }),
        input = Stage::Join(Join {
            condition: None,
            left: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "col".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            right: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Documents(Documents {
                    array: vec![Expression::Literal(LiteralValue::Integer(1)), Expression::Literal(LiteralValue::Integer(1))]
                })),
                specifications: unchecked_unique_linked_hash_map!(
                    "arr".to_string() => Expression::Variable("ROOT".to_string().into())
                )
            })),
            let_vars: None,
            join_type: JoinType::Left,
        }),
    );
}
