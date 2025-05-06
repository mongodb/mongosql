mod get_schema_for_path {
    use crate::{get_or_create_schema_for_path_mut, get_schema_for_path_mut};
    use mongosql::{
        map,
        schema::{Atomic, Document, Schema},
        set,
    };
    use std::collections::BTreeSet;

    macro_rules! test_schema_for_path {
        ($func_name:ident, expected = $expected:expr, input = $input:expr, path = $path:expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_schema_for_path_mut(input_cloned, $path);
                assert_eq!($expected, result);
            }
        };
    }

    macro_rules! test_get_or_create_schema_for_path {
        ($func_name:ident, expected = $expected:expr, output = $output:expr, input = $input:expr, path = $path: expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_or_create_schema_for_path_mut(input_cloned, $path);
                assert_eq!($expected, result);
                assert_eq!($output, *input_cloned);
            }
        };
        ($func_name:ident, expected = $expected:expr, input = $input:expr, path = $path:expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_or_create_schema_for_path_mut(input_cloned, $path);
                assert_eq!($expected, result);
                assert_eq!($input, *input_cloned);
            }
        };
    }

    test_schema_for_path!(
        simple,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string()]
    );

    test_schema_for_path!(
        nested,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        })
                    },
                    ..Default::default()
                })
            },
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_schema_for_path!(
        nested_in_array,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Array(Box::new(Schema::Array(Box::new(Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        }))))),
                    },
                    ..Default::default()
                })
            },
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_schema_for_path!(
        missing,
        expected = None,
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["d".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_get_two_levels,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    additional_properties: false,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_two_levels,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Any,
                    },
                    additional_properties: true,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Any,
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_two_levels_with_array,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Array(Box::new(Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Any,
                    },
                    additional_properties: true,
                    ..Default::default()
                }))))),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Array(Box::new(Schema::Array(Box::new(Schema::Any)))),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_get_three_levels,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            additional_properties: false,
                            ..Default::default()
                        })
                    },
                    additional_properties: false,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_three_levels,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Any,
                            },
                            additional_properties: true,
                            ..Default::default()
                        })
                    },
                    additional_properties: true,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Any,
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_refine_two_levels_any_of,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    additional_properties: false,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                        Schema::Document(Document {
                            keys: map! {
                                "b".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            additional_properties: false,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Integer)
                    )),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_and_refine_two_levels_any_of,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Any,
                    },
                    additional_properties: true,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                        Schema::Document(Document {
                            keys: map! {
                                "b".to_string() => Schema::Any,
                            },
                            additional_properties: true,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Integer)
                    )),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_and_refine_two_levels_any_of_array,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Array(Box::new(Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Any,
                    },
                    additional_properties: true,
                    ..Default::default()
                })))))
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                        Schema::Array(Box::new(Schema::Array(Box::new(Schema::Document(Document {
                            keys: map! {
                                "b".to_string() => Schema::Any,
                            },
                            additional_properties: true,
                            ..Default::default()
                        }))))),
                        Schema::Atomic(Atomic::Integer)
                    )),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );
}

mod get_namespaces_for_pipeline {
    use crate::get_namespaces_for_pipeline;
    use agg_ast::definitions::{Namespace, Stage};
    use mongosql::set;
    use std::collections::BTreeSet;

    macro_rules! test_get_namespaces_for_pipeline {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let pipeline: Vec<Stage> = serde_json::from_str($input).unwrap();
                let expected: BTreeSet<Namespace> = $expected;
                let actual = get_namespaces_for_pipeline(
                    pipeline,
                    "test".to_string(),
                    Some("foo".to_string()),
                );
                assert_eq!(actual, expected);
            }
        };
    }

    test_get_namespaces_for_pipeline!(
        simple,
        expected = set!(Namespace::new("test".to_string(), "foo".to_string()),),
        input = r#"[{"$match": {"foo": 1}}]"#
    );

    test_get_namespaces_for_pipeline!(
        collection,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("a".to_string(), "b".to_string()),
        ),
        input = r#"[{"$collection": {"db": "a", "collection": "b"}}]"#
    );

    test_get_namespaces_for_pipeline!(
        equi_join_no_db,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "bar".to_string()),
        ),
        input = r#"[{ "$equiJoin": { "collection": "bar", "joinType": "left", "localField": "a", "foreignField" : "b", "as": "bar" } }]"#
    );

    test_get_namespaces_for_pipeline!(
        equi_join_with_db,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test_2".to_string(), "bar".to_string()),
        ),
        input = r#"[{ "$equiJoin": { "collection": "bar", "database": "test_2", "joinType": "left", "localField": "a", "foreignField" : "b", "as": "bar" } }]"#
    );

    test_get_namespaces_for_pipeline!(
        join_no_db,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "bar".to_string()),
        ),
        input = r#"[{ "$join": { "collection": "bar", "joinType": "left", "pipeline": [] } }]"#
    );

    test_get_namespaces_for_pipeline!(
        union_with_collection,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "bar".to_string()),
        ),
        input = r#"[{ "$unionWith": "bar" }]"#
    );

    test_get_namespaces_for_pipeline!(
        union_with_empty_pipeline,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "bar".to_string()),
        ),
        input = r#"[{ "$unionWith": {"coll": "bar", "pipeline": [] } }]"#
    );

    test_get_namespaces_for_pipeline!(
        union_with_singleton_pipeline,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "single".to_string()),
        ),
        input = r#"[{ "$unionWith": {"coll": "single", "pipeline": [{"$limit": 10}] } }]"#
    );

    test_get_namespaces_for_pipeline!(
        union_with_multiple_element_pipeline,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "multiple".to_string()),
        ),
        input = r#"[{ "$unionWith": {"coll": "multiple", "pipeline": [{"$skip": 5}, {"$limit": 10}] } }]"#
    );

    test_get_namespaces_for_pipeline!(
        graph_lookup,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "start".to_string()),
        ),
        input = r#"[{"$graphLookup": {
                "from": "start",
                "startWith": "$start",
                "connectFromField": "start",
                "connectToField": "end",
                "as": "path",
                "maxDepth": 5,
                "depthField": "depth",
                "restrictSearchWithMatch": { "sql": true }
            }}]"#
    );

    test_get_namespaces_for_pipeline!(
        subquery_no_optional_fields,
        expected = set!(Namespace::new("test".to_string(), "foo".to_string()),),
        input = r#"[{"$lookup": {"pipeline": [], "as": "as_var"}}]"#
    );

    test_get_namespaces_for_pipeline!(
        subquery_lookup_from_collection,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "from_coll".to_string()),
        ),
        input = r#"[{"$lookup": {"from": "from_coll", "pipeline": [], "as": "as_var"}}]"#
    );

    test_get_namespaces_for_pipeline!(
        subquery_lookup_from_namespace,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("from_db".to_string(), "from_coll".to_string()),
        ),
        input = r#"[{"$lookup": {"from": {"db": "from_db", "coll": "from_coll"}, "pipeline": [], "as": "as_var"}}]"#
    );

    test_get_namespaces_for_pipeline!(
        subquery_lookup_with_single_let_var,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("from_db".to_string(), "from_coll".to_string()),
        ),
        input = r#"[{"$lookup": {
            "from": {"db": "from_db", "coll": "from_coll"},
            "let": {"x": 9},
            "pipeline": [],
            "as": "as_var"
        }}]"#
    );

    test_get_namespaces_for_pipeline!(
        subquery_lookup_with_multiple_let_vars,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("from_db".to_string(), "from_coll".to_string()),
        ),
        input = r#"[{"$lookup": {
            "from": {"db": "from_db", "coll": "from_coll"},
            "let": {
                "x": 9,
                "y": "$z"
            },
            "pipeline": [],
            "as": "as_var"
        }}]"#
    );

    test_get_namespaces_for_pipeline!(
        subquery_lookup_with_pipeline,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("db".to_string(), "bar".to_string()),
        ),
        input = r#"[{
            "$lookup":
              {
                "from": { "db": "db", "coll": "bar" },
                "let": { "foo_b_0": "$b" },
                "pipeline":
                  [
                    { "$match": { "$expr": { "$eq": ["$$foo_b_0", "$b"] } } },
                    { "$project": { "_id": 0, "a": 1 } }
                  ],
                "as": "__subquery_result_0"
          }}]"#
    );

    test_get_namespaces_for_pipeline!(
        concise_subquery_lookup_with_no_optional_fields,
        expected = set!(Namespace::new("test".to_string(), "foo".to_string()),),
        input = r#"[{"$lookup": {"pipeline": [], "as": "as_var", "localField": "foo", "foreignField": "bar"}}]"#
    );

    test_get_namespaces_for_pipeline!(
        concise_subquery_lookup_fully_specified,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "coll".to_string()),
        ),
        input = r#"[{
            "$lookup":
              {
                "from": "coll",
                "let": { "foo_b_0": "$b" },
                "pipeline":
                  [
                    { "$match": { "$expr": { "$eq": ["$$foo_b_0", "$b"] } } },
                    { "$project": { "_id": 0, "a": 1 } }
                  ],
                "as": "__subquery_result_0",
                "localField": "foo",
                "foreignField": "bar"
          }}]"#
    );

    test_get_namespaces_for_pipeline!(
        equality_lookup,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "coll".to_string()),
        ),
        input = r#"[{ "$lookup": { "from": "coll", "as": "__subquery_result_0", "localField": "foo", "foreignField": "bar" }}]"#
    );

    test_get_namespaces_for_pipeline!(
        multiple_nested_stages,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "union_with_coll".to_string()),
            Namespace::new("test".to_string(), "union_with_pipeline_lookup".to_string()),
            Namespace::new("lookup_db".to_string(), "lookup_coll".to_string()),
            Namespace::new(
                "lookup_db".to_string(),
                "lookup_pipeline_graph_lookup".to_string()
            ),
        ),
        input = r#"[
            { "$unionWith": {"coll": "union_with_coll", "pipeline": [
                { "$lookup": { "from": "union_with_pipeline_lookup", "as": "__subquery_result_0", "localField": "foo", "foreignField": "bar" }}
            ] }},
            { "$lookup": {
                "from": { "db": "lookup_db", "coll": "lookup_coll" },
                "let": { "foo_b_0": "$b" },
                "pipeline":
                [
                    {"$graphLookup": {
                        "from": "lookup_pipeline_graph_lookup",
                        "startWith": "$start",
                        "connectFromField": "start",
                        "connectToField": "end",
                        "as": "path",
                        "maxDepth": 5,
                        "depthField": "depth",
                        "restrictSearchWithMatch": { "sql": true }
                    }}
                ],
                "as": "__subquery_result_0"
            }}
        ]"#
    );


    test_get_namespaces_for_pipeline!(
        facet_one_namespace,
        expected = set!(Namespace::new("test".to_string(), "foo".to_string()),),
        input = r#"[{
            "$facet": {
                "facet1": [
                    { "$match": { "foo": 1 } }
                ]
            }
        }]"#
    );

    test_get_namespaces_for_pipeline!(
        facet_with_lookups,
        expected = set!(
            Namespace::new("test".to_string(), "foo".to_string()),
            Namespace::new("test".to_string(), "bar".to_string()),
            Namespace::new("test".to_string(), "baz".to_string()),
            Namespace::new("test".to_string(), "biz".to_string()),
            Namespace::new("test".to_string(), "quz".to_string()),
        ),
        input = r#"[{
            "$facet": {
                "facet1": [
                    { "$lookup": { "from": "bar", "pipeline": [], "as": "as_var" } },
                    { "$lookup": { "from": "baz", "pipeline": [], "as": "as_var" } }
                ],
                "facet2": [
                    { "$lookup": { "from": "biz", "pipeline": [], "as": "as_var" } },
                    { "$lookup": { "from": "quz", "pipeline": [], "as": "as_var" } }
                ]
            }
        }]"#
    );

