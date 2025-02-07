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
                                "c".to_string() => Schema::Atomic(Atomic::BinData),
                                "d".to_string() => Schema::Atomic(Atomic::Boolean)
                            },
                            required: set!("c".to_string()),
                            ..Default::default()
                        })
                    },
                    required: set!("b".to_string()),
                    ..Default::default()
                }),
                "x".to_string() => Schema::Atomic(Atomic::Date),
                "y".to_string() => Schema::Document(Document {
                    keys: map! {
                        "z".to_string() => Schema::Atomic(Atomic::Double)
                    },
                    required: set!("z".to_string()),
                    ..Default::default()
                })
            },
            required: set!("bar".to_string(), "x".to_string(), "y".to_string()),
            ..Default::default()
        })),
        input = r#"{"$densify": {"field": "bar.b.c", "partitionByFields": ["x", "y.z"], "range": { "step": 1, "bounds": "full", "unit": "second" }}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
                "bar".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::String),
                        "b".to_string() => Schema::Document(Document { keys:
                            map! {
                                "c".to_string() => Schema::Atomic(Atomic::BinData),
                                "d".to_string() => Schema::Atomic(Atomic::Boolean)
                            },
                            required: set!("c".to_string(), "d".to_string()),
                            ..Default::default()
                        })
                    },
                    required: set!("a".to_string()),
                    ..Default::default()
                }),
                "x".to_string() => Schema::Atomic(Atomic::Date),
                "y".to_string() => Schema::Document(Document {
                    keys: map! {
                        "z".to_string() => Schema::Atomic(Atomic::Double)
                    },
                    ..Default::default()
                })
            },
            required: set!("foo".to_string()),
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

// mod facet {
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
// }

mod geo_near {
    use super::*;

    test_derive_stage_schema!(
        geo_near_no_options,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Atomic(Atomic::String),
                "location".to_string() => Schema::Document(Document {
                    keys: map! {
                        "type".to_string() => Schema::Atomic(Atomic::String),
                        "coordinates".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                    },
                    required: set!("type".to_string(), "coordinates".to_string()),
                    ..Default::default()
                }),
                "category".to_string() => Schema::Atomic(Atomic::String),
                "f".to_string() => Schema::Atomic(Atomic::Double)
            },
            required: set!(
                "name".to_string(),
                "location".to_string(),
                "category".to_string()
            ),
            ..Default::default()
        })),
        input = r#"{"$geoNear": {
                   "distanceField": "f",
                   "near": {
                       "type": "Point",
                       "coordinates": [-73.856077, 40.848447],
                   }
        }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Atomic(Atomic::String),
                "location".to_string() => Schema::Document(Document {
                    keys: map! {
                        "type".to_string() => Schema::Atomic(Atomic::String),
                        "coordinates".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                    },
                    required: set!("type".to_string(), "coordinates".to_string()),
                    ..Default::default()
                }),
                "category".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!(
                "name".to_string(),
                "location".to_string(),
                "category".to_string()
            ),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        geo_near_include_loc,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Atomic(Atomic::String),
                "location".to_string() => Schema::Document(Document {
                    keys: map! {
                        "type".to_string() => Schema::Atomic(Atomic::String),
                        "coordinates".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                    },
                    required: set!("type".to_string(), "coordinates".to_string()),
                    ..Default::default()
                }),
                "category".to_string() => Schema::Atomic(Atomic::String),
                "dist".to_string() => Schema::Document(Document {
                    keys: map! {
                        "calculated".to_string() => Schema::Atomic(Atomic::Double),
                        "location".to_string() => Schema::Document(Document {
                            keys: map! {
                                "type".to_string() => Schema::Atomic(Atomic::String),
                                "coordinates".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                            },
                            required: set!("type".to_string(), "coordinates".to_string()),
                            ..Default::default()
                        })
                    },
                    required: set!("type".to_string(), "coordinates".to_string()),
                    ..Default::default()
                })
            },
            required: set!(
                "name".to_string(),
                "location".to_string(),
                "category".to_string()
            ),
            ..Default::default()
        })),
        input = r#"{"$geoNear": {
                   "distanceField": "dist.calculated",
                   "near": {
                       "type": "Point",
                       "coordinates": [-73.856077, 40.848447],
                   },
                   "includeLocs": "dist.location"
        }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "name".to_string() => Schema::Atomic(Atomic::String),
                "location".to_string() => Schema::Document(Document {
                    keys: map! {
                        "type".to_string() => Schema::Atomic(Atomic::String),
                        "coordinates".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                    },
                    required: set!("type".to_string(), "coordinates".to_string()),
                    ..Default::default()
                }),
                "category".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!(
                "name".to_string(),
                "location".to_string(),
                "category".to_string()
            ),
            ..Default::default()
        })
    );
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
            required: set!("foo".to_string()),
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
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Null)
                )),
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "i".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("bar".to_string()),
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
