use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema},
    set,
};
use std::collections::BTreeMap;

mod densify {
    use super::*;

    test_derive_stage_schema!(
        densify_fully_specified,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
                "bar".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::String),
                        "b".to_string() => Schema::Document(Document { keys:
                            map! {
                                "x".to_string() => Schema::Atomic(Atomic::Boolean),
                                "y".to_string() => Schema::Atomic(Atomic::Integer)
                            },
                            // y is the field being densified so it should be kept. x is not referenced as a partition by field nor a densified
                            // field, so it is no longer required
                            required: set!("y".to_string()),
                            // additional_properties: true,
                            ..Default::default()
                        }),
                        "partition_one".to_string() => Schema::Atomic(Atomic::Double),
                    },
                    // a should no longer be required, because it is not the densified field, nor one of the partition by fields
                    // b should be kept as required because it is part of the path of the densified field; partition_one should be kept
                    // because it is one of the partition by fields
                    required: set!("b".to_string(), "partition_one".to_string()),
                    // additional_properties: true,
                    ..Default::default()
                }),
                "partition_two".to_string() => Schema::Atomic(Atomic::Double)
            },
            // partition_two does not become required just because it is a partition by field
            required: set!("bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$densify": {"field": "bar.b.y", "partitionByFields": ["bar.partition_one", "partition_two"], "range": { "step": 1, "bounds": "full" }}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
                "bar".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::String),
                        "b".to_string() => Schema::Document(Document { keys:
                            map! {
                                "x".to_string() => Schema::Atomic(Atomic::Boolean),
                                "y".to_string() => Schema::Atomic(Atomic::Integer)
                            },
                            required: set!("x".to_string(), "y".to_string()),
                            ..Default::default()
                        }),
                        "partition_one".to_string() => Schema::Atomic(Atomic::Double),
                    },
                    required: set!("a".to_string(), "b".to_string(), "partition_one".to_string()),
                    ..Default::default()
                }),
                "partition_two".to_string() => Schema::Atomic(Atomic::Double)
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })
    );
}

mod documents {
    use super::*;

    test_derive_stage_schema!(
        empty,
        expected = Ok(Schema::Document(Document::default())),
        input = r#"{"$documents": []}"#
    );

    test_derive_stage_schema!(
        singleton,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set!("a".to_string()),
            ..Default::default()
        })),
        input = r#" {"$documents": [{"a": 1}]}"#
    );

    test_derive_stage_schema!(
        multiple_documents,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                    Schema::Document(Document {
                        keys: map! {
                            "b".to_string() => Schema::Document(Document {
                                keys: map!{
                                    "c".to_string() => Schema::Atomic(Atomic::Boolean)
                                },
                                required: set!("c".to_string()),
                                ..Default::default()
                            })
                        },
                        required: set!("b".to_string()),
                        ..Default::default()
                    })
                )),
                "b".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer)
                ))
            },
            required: set!("a".to_string()),
            ..Default::default()
        })),
        input = r#"{"$documents": [
                     {"a": 1, "b": 2},
                     {"a": "yes", "b": null},
                     {"a": {"b": {"c": true}}}
        ]}"#
    );
}

mod facet {
    // SQL-2369: implement schema derivation for bucketing stages
    //     use super::*;

    //     test_derive_stage_schema!(
    //         empty,
    //         expected = Ok(Schema::Document(Document::default())),
    //         input = r#"stage: {"$facet": {}}"#
    //     );

    //     test_derive_stage_schema!(
    //         single,
    //         expected = Ok(Schema::Document(Document {
    //             keys: map! {
    //                 "outputField1".to_string() => Schema::Array(Box::new(
    //                     Schema::Document(Document {
    //                         keys: map! {
    //                             "x".to_string() => Schema::Atomic(Atomic::Integer)
    //                         },
    //                         required: set!("x".to_string()),
    //                         ..Default::default()
    //                     })
    //                 ))
    //             },
    //             required: set!("outputField1".to_string()),
    //             ..Default::default()
    //         })),
    //         input = r#"{"$facet": { "outputField1": [{"$count": "x"}] }}"#
    //     );

    //     test_derive_stage_schema!(
    //         multiple,
    //         expected = Ok(Schema::Document(Document {
    //             keys: map! {
    //                 "o1".to_string() => Schema::Array(Box::new(
    //                     Schema::Document(Document {
    //                         keys: map! {
    //                             "x".to_string() => Schema::Atomic(Atomic::String),
    //                         },
    //                         required: set!(),
    //                         ..Default::default()
    //                     })
    //                 )),
    //                 "outputField2".to_string() => Schema::Array(Box::new(
    //                     Schema::Document(Document {
    //                         keys: map! {
    //                             "x".to_string() => Schema::Atomic(Atomic::Integer)
    //                         },
    //                         required: set!(),
    //                         ..Default::default()
    //                     })
    //                 ))
    //             },
    //             required: set!("outputField1".to_string()),
    //             ..Default::default()
    //         })),
    //         input = r#"{"$facet": {
    //             "o1": [{"$limit": 10}, {"$project": {"_id": 0}}],
    //             "outputField2": [{"$count": "x"}],
    //         }}"#,
    //         starting_schema = Schema::Document(Document {
    //             keys: map! {
    //                 "x".to_string() => Schema::Atomic(Atomic::String),
    //                 "_id".to_string() => Schema::Atomic(Atomic::ObjectId)
    //             },
    //             required: set!("_id".to_string()),
    //             ..Default::default()
    //         })
    //     );
}

mod sort_by_count {
    use super::*;

    test_derive_stage_schema!(
        field_ref,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::Symbol),
                "count".to_string() => Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)))
            },
            required: set!("_id".to_string(), "count".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$sortByCount": "$foo" }"#,
        ref_schema = Schema::Atomic(Atomic::Symbol)
    );
}

mod unwind {
    use super::*;

    test_derive_stage_schema!(
        field_ref,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Double)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$unwind": "$foo" }"#,
        ref_schema = Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
    );

    test_derive_stage_schema!(
        field_ref_multiple_different_array_types,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::BinData),
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                ))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$unwind": "$foo" }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::BinData),
            Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
            Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            Schema::Array(Box::new(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::Boolean),
            ))))
        ))
    );

    test_derive_stage_schema!(
        field_ref_nested,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "bar".to_string() => Schema::Atomic(Atomic::Double)
                    },
                    required: set!("bar".to_string()),
                    ..Default::default()
                })
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unwind": {"path": "$foo.bar"}}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
            },
            required: set!("bar".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        document_no_options,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Double)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unwind": {"path": "$foo"}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
    );

    test_derive_stage_schema!(
        document_include_array_index_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Double),
                "i".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set!("foo".to_string(), "i".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unwind": {"path": "$foo", "includeArrayIndex": "i"}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
    );

    test_derive_stage_schema!(
        document_all_options,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Null)
                )),
                "bar".to_string() => Schema::Atomic(Atomic::ObjectId),
                "i".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("bar".to_string(), "i".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unwind": {"path": "$foo", "includeArrayIndex": "i", "preserveNullAndEmptyArrays": true }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                    Schema::Atomic(Atomic::Null)
                )),
                "bar".to_string() => Schema::Atomic(Atomic::ObjectId)
            },
            required: set!("bar".to_string()),
            ..Default::default()
        })
    );
}

mod lookup {
    use super::*;

    test_derive_stage_schema!(
        eq_lookup,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "baz".to_string() => Schema::Atomic(Atomic::String),
                            "qux".to_string() => Schema::Atomic(Atomic::Integer)
                        },
                        required: set!("baz".to_string(), "qux".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {"from": "bar", "localField": "foo", "foreignField": "baz", "as": "arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        eq_lookup_nested_as,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() =>
                    Schema::Document(Document {
                        keys: map! {
                            "arr".to_string() => Schema::Array(
                                Box::new(Schema::Document(Document {
                                keys: map! {
                                    "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                                    "baz".to_string() => Schema::Atomic(Atomic::String),
                                    "qux".to_string() => Schema::Atomic(Atomic::Integer)
                                },
                                required: set!("baz".to_string(), "qux".to_string(), "_id".to_string()),
                                ..Default::default()
                            }))),
                        },
                        required: set!("arr".to_string()),
                        ..Default::default()
                    }),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {"from": "bar", "localField": "foo", "foreignField": "baz", "as": "arr.arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        eq_lookup_overwrite_as,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "baz".to_string() => Schema::Atomic(Atomic::String),
                            "qux".to_string() => Schema::Atomic(Atomic::Integer)
                        },
                        required: set!("baz".to_string(), "qux".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {"from": "bar", "localField": "foo", "foreignField": "baz", "as": "foo"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        concise_subquery_lookup,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "out".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("out".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {"from": "bar", "localField": "foo", "foreignField": "baz", "let": {"x": "$foo"}, "pipeline": [{"$project": {"out": {"$concat": ["$$x", "$baz"]}}}], "as": "arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        subquery_lookup,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "out".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("out".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {"from": "bar", "let": {"x": "$foo"}, "pipeline": [{"$project": {"out": {"$concat": ["$$x", "$baz"]}}}], "as": "arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        subquery_lookup_overwrite_as,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "out".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("out".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {"from": "bar", "let": {"x": "foo"}, "pipeline": [{"$project": {"out": {"$concat": ["$$x", "$baz"]}}}], "as": "foo"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
}

mod graphlookup {
    use super::*;

    test_derive_stage_schema!(
        graphlookup_simple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "baz".to_string() => Schema::Atomic(Atomic::String),
                            "qux".to_string() => Schema::Atomic(Atomic::Integer)
                        },
                        required: set!("baz".to_string(), "qux".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$graphLookup": {"from": "bar", "startWith": "$foo", "connectFromField": "foo", "connectToField": "baz", "as": "arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        graphlookup_depth_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "baz".to_string() => Schema::Atomic(Atomic::String),
                            "qux".to_string() => Schema::Atomic(Atomic::Integer),
                            "DEPTH".to_string() => Schema::Atomic(Atomic::Long)
                        },
                        required: set!("baz".to_string(), "qux".to_string(), "_id".to_string(), "DEPTH".to_string()),
                        ..Default::default()
                    }))
                ),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$graphLookup": {"from": "bar", "startWith": "$foo", "connectFromField": "foo", "connectToField": "baz", "depthField": "DEPTH", "as": "arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        graphlookup_overwrite_depth_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "arr".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "baz".to_string() => Schema::Atomic(Atomic::Long),
                            "qux".to_string() => Schema::Atomic(Atomic::Integer),
                        },
                        required: set!("baz".to_string(), "qux".to_string(), "_id".to_string()),
                        ..Default::default()
                    }))
                ),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "arr".to_string()),
            ..Default::default()
        })),
        input = r#"{"$graphLookup": {"from": "bar", "startWith": "$foo", "connectFromField": "foo", "connectToField": "baz", "depthField": "baz", "as": "arr"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        graphlookup_overwrite_as,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                            "baz".to_string() => Schema::Atomic(Atomic::String),
                            "qux".to_string() => Schema::Atomic(Atomic::Integer),
                            "DEPTH".to_string() => Schema::Atomic(Atomic::Long)
                        },
                        required: set!("baz".to_string(), "qux".to_string(), "_id".to_string(), "DEPTH".to_string()),
                        ..Default::default()
                    }))
                ),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$graphLookup": {"from": "bar", "startWith": "$foo", "connectFromField": "foo", "connectToField": "baz", "depthField": "DEPTH", "as": "foo"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
}

mod union {
    use super::*;

    test_derive_stage_schema!(
        union_simple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "food".to_string() => Schema::Atomic(Atomic::String),
                "baz".to_string() => Schema::Atomic(Atomic::String),
                "qux".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$unionWith": "bar"}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "food".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        union_pipeline,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "food".to_string() => Schema::Atomic(Atomic::String),
                "out".to_string() => Schema::AnyOf(set!{
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Decimal),
                }),
            },
            required: set!("out".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unionWith": {"collection": "bar", "pipeline": [{"$project": {"out": {"$concat": ["$baz", "$baz"]}}}]}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "food".to_string() => Schema::Atomic(Atomic::String),
                "out".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("foo".to_string(), "out".to_string()),
            ..Default::default()
        })
    );
}

mod replace {
    use super::*;

    test_derive_stage_schema!(
        replace_root_nested_doc,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("b".to_string()),
            ..Default::default()
        })),
        input = r#"{"$replaceRoot": {"newRoot": "$name"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::Integer),
                        "b".to_string() => Schema::Atomic(Atomic::String)
                    },
                    required: set!("b".to_string()),
                    ..Default::default()
                }),
                "foo".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("foo".to_string(), "name".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        replace_root_expression,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
                "first".to_string() => Schema::Atomic(Atomic::String),
                "last".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("b".to_string(), "first".to_string(), "last".to_string()),
            ..Default::default()
        })),
        input = r#"{"$replaceRoot": {"newRoot": {"$mergeObjects": [{"first": "", "last": "" }, "$name"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::Integer),
                        "b".to_string() => Schema::Atomic(Atomic::String)
                    },
                    required: set!("b".to_string()),
                    ..Default::default()
                }),
                "foo".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("foo".to_string(), "name".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        replace_with_nested_doc,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("b".to_string()),
            ..Default::default()
        })),
        input = r#"{"$replaceWith": "$name"}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::Integer),
                        "b".to_string() => Schema::Atomic(Atomic::String)
                    },
                    required: set!("b".to_string()),
                    ..Default::default()
                }),
                "foo".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("foo".to_string(), "name".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        replace_with_expression,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
                "first".to_string() => Schema::Atomic(Atomic::String),
                "last".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("b".to_string(), "first".to_string(), "last".to_string()),
            ..Default::default()
        })),
        input = r#"{"$replaceWith": {"$mergeObjects": [{"first": "", "last": "" }, "$name"]}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::Integer),
                        "b".to_string() => Schema::Atomic(Atomic::String)
                    },
                    required: set!("b".to_string()),
                    ..Default::default()
                }),
                "foo".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("foo".to_string(), "name".to_string()),
            ..Default::default()
        })
    );
}

mod add_fields {
    use super::*;

    test_derive_stage_schema!(
        add_fields_multiple_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "student".to_string() => Schema::Atomic(Atomic::String),
                "homework".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                "quiz".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                "avg_homework".to_string() => Schema::Atomic(Atomic::Double),
                "avg_quiz".to_string() => Schema::Atomic(Atomic::Double),
            },
            required: set!(
                "student".to_string(),
                "homework".to_string(),
                "quiz".to_string(),
                "avg_homework".to_string(),
                "avg_quiz".to_string(),
            ),
            ..Default::default()
        })),
        input = r#"{ "$addFields": { "avg_homework": { "$avg": "$homework" }, "avg_quiz": { "$avg": "$quiz" } }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "student".to_string() => Schema::Atomic(Atomic::String),
                "homework".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                "quiz".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            },
            required: set!(
                "student".to_string(),
                "homework".to_string(),
                "quiz".to_string()
            ),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        add_fields_embedded_document,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "specs".to_string() => Schema::Document(Document {
                    keys: map! {
                        "doors".to_string() => Schema::Atomic(Atomic::Integer),
                        "wheels".to_string() => Schema::Atomic(Atomic::Integer),
                        "fuel_type".to_string() => Schema::Atomic(Atomic::String)
                    },
                    required: set!("doors".to_string(), "wheels".to_string(), "fuel_type".to_string()),
                    ..Default::default()
                }),
                "type".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("specs".to_string(), "type".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$addFields": { "specs.fuel_type": "unleaded" }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "specs".to_string() => Schema::Document(Document {
                    keys: map! {
                        "doors".to_string() => Schema::Atomic(Atomic::Integer),
                        "wheels".to_string() => Schema::Atomic(Atomic::Integer)
                    },
                    required: set!("doors".to_string(), "wheels".to_string()),
                    ..Default::default()
                }),
                "type".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("specs".to_string(), "type".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        add_fields_overwrite_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "cats".to_string() => Schema::Atomic(Atomic::String),
                "dogs".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("dogs".to_string(), "cats".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$addFields": { "cats": "none" }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "dogs".to_string() => Schema::Atomic(Atomic::Integer),
                "cats".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("dogs".to_string(), "cats".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        add_fields_add_element_to_array,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "student".to_string() => Schema::Atomic(Atomic::String),
                "homework".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            },
            required: set!("student".to_string(), "homework".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$addFields": { "homework": { "$concatArrays": [ "$homework", [ 7 ] ] } } }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "student".to_string() => Schema::Atomic(Atomic::String),
                "homework".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            },
            required: set!("student".to_string(), "homework".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        add_fields_remove_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "temperature".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("temperature".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$addFields": { "date": "$$REMOVE" } }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "date".to_string() => Schema::Atomic(Atomic::Date),
                "temperature".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("date".to_string(), "temperature".to_string()),
            ..Default::default()
        })
    );
}
