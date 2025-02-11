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
