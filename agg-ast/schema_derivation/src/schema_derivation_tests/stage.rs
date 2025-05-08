use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema},
    set,
};
use std::collections::BTreeMap;

mod add_fields {
    use super::*;

    test_derive_stage_schema!(
        add_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Double),
                "bar".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$addFields": {"bar": "baz"}}"#,
        ref_schema = Schema::Atomic(Atomic::Double)
    );
    test_derive_stage_schema!(
        add_fields_multiple_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "student".to_string() => Schema::Atomic(Atomic::String),
                "homework".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                "quiz".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                "avg_homework".to_string() => Schema::Atomic(Atomic::Double),
                "avg_quiz".to_string() => Schema::Atomic(Atomic::Double),
                "total_quiz".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal)
                ))
            },
            required: set!(
                "student".to_string(),
                "homework".to_string(),
                "quiz".to_string(),
                "avg_homework".to_string(),
                "avg_quiz".to_string(),
                "total_quiz".to_string()
            ),
            ..Default::default()
        })),
        input = r#"{ "$addFields": { "avg_homework": { "$avg": "$homework" }, "avg_quiz": { "$avg": "$quiz" }, "total_quiz": { "$sum": "$quiz"} }}"#,
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

mod bucket {
    use super::*;

    test_derive_stage_schema!(
        bucket_no_output,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::String),
                "count".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long)
                ))
            },
            required: set!("_id".to_string(), "count".to_string()),
            ..Default::default()
        })),
        input = r#"{"$bucket": {"groupBy": "$foo", "boundaries": ["hello", "world", "zod"]}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
    test_derive_stage_schema!(
        bucket_no_output_and_default,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::AnyOf(
                    set!{
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Integer),
                    }),
                "count".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long)
                ))
            },
            required: set!("_id".to_string(), "count".to_string()),
            ..Default::default()
        })),
        input = r#"{"$bucket": {"groupBy": "$foo", "default": 1, "boundaries": ["hello", "world", "zod"]}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
    test_derive_stage_schema!(
        bucket_with_output,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                )),
                "values".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::String)))
            },
            required: set!("_id".to_string(), "c".to_string(), "values".to_string()),
            ..Default::default()
        })),
        input = r#"{"$bucket": {"groupBy": "$foo", "boundaries": ["hello", "world", "zod"], "output": {"c": {"$sum": 1}, "values": {"$push": "$foo"}}}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );

    test_derive_stage_schema!(
        bucket_auto_no_output,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Document(Document {
                    keys: map! {
                        "min".to_string() => Schema::Atomic(Atomic::String),
                        "max".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("min".to_string(), "max".to_string()),
                    ..Default::default()
                }),
                "count".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long)
                ))
            },
            required: set!("_id".to_string(), "count".to_string()),
            ..Default::default()
        })),
        input = r#"{"$bucketAuto": {"groupBy": "$foo", "buckets": 5}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
    test_derive_stage_schema!(
        bucket_auto_with_output,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Document(Document {
                    keys: map! {
                        "min".to_string() => Schema::Atomic(Atomic::String),
                        "max".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("min".to_string(), "max".to_string()),
                    ..Default::default()
                }),
                "c".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                )),
                "values".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::String)))
            },
            required: set!("_id".to_string(), "c".to_string(), "values".to_string()),
            ..Default::default()
        })),
        input = r#"{"$bucketAuto": {"groupBy": "$foo", "buckets": 5, "output": {"c": {"$sum": 1}, "values": {"$push": "$foo"}}}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
}

mod collection {
    use super::*;

    test_derive_stage_schema!(
        collection,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "baz".to_string() => Schema::Atomic(Atomic::String),
                "qux".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"baz".to_string(), "qux".to_string(), "_id".to_string()},
            ..Default::default()
        }),),
        input = r#"{"$collection": {"db": "test", "collection": "bar"}}"#
    );
}

mod count {
    use super::*;

    test_derive_stage_schema!(
        count,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "count".to_string() => Schema::AnyOf(set!(
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Long),
                ))
            },
            required: set!("count".to_string()),
            ..Default::default()
        })),
        input = r#"{"$count": "count"}"#
    );
}

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

    test_derive_stage_schema!(
        densify_nested_anyofs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Integer),
                            Schema::Document(Document { keys:
                                map! {
                                    "c".to_string() => Schema::Atomic(Atomic::Integer),
                                    "d".to_string() => Schema::Atomic(Atomic::Integer),
                                },
                                ..Default::default()
                            })
                        )),
                    },
                    ..Default::default()
                }),
            },
            required: set!("a".to_string()),
            ..Default::default()
        })),
        input = r#"{"$densify": {"field": "a.b.c", "partitionByFields": ["a.b.d"], "range": { "step": 1, "bounds": "full" }}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Integer),
                            Schema::Document(Document { keys:
                                map! {
                                    "c".to_string() => Schema::Atomic(Atomic::Integer),
                                    "d".to_string() => Schema::Atomic(Atomic::Integer),
                                },
                                ..Default::default()
                            })
                        )),
                    },
                    ..Default::default()
                }),
            },
            required: set!("a".to_string()),
            ..Default::default()
        })
    );
}

mod documents {
    use super::*;

    test_derive_stage_schema!(
        empty,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Missing,
            Schema::Document(Document::any())
        ))),
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

    test_derive_stage_schema!(
        expression_input,
        expected = Ok(Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set!("a".to_string()),
            ..Default::default()
        })))),
        input = r#" {"$documents": {"$filter": {"input": [{"a": 1}, {"a": "hello"}, {"a": false}], "as": "bar", "cond": {"$isNumber": "$$bar.a"}}}}"#
    );
}

mod facet {
    use super::*;

    test_derive_stage_schema!(
        empty,
        expected = Ok(Schema::Document(Document::default())),
        input = r#"{"$facet": {}}"#
    );

    test_derive_stage_schema!(
        single,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "outputField1".to_string() => Schema::Array(Box::new(
                    Schema::Document(Document {
                        keys: map! {
                            "x".to_string() => Schema::AnyOf(set!{
                                Schema::Atomic(Atomic::Integer),
                                Schema::Atomic(Atomic::Long)
                            })
                        },
                        required: set!("x".to_string()),
                        ..Default::default()
                    })
                ))
            },
            required: set!("outputField1".to_string()),
            ..Default::default()
        })),
        input = r#"{"$facet": { "outputField1": [{"$count": "x"}] }}"#
    );

    test_derive_stage_schema!(
        multiple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "o1".to_string() => Schema::Array(Box::new(
                    Schema::Document(Document {
                        keys: map! {
                            "x".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!(),
                        ..Default::default()
                    })
                )),
                "outputField2".to_string() => Schema::Array(Box::new(
                    Schema::Document(Document {
                        keys: map! {
                            "x".to_string() => Schema::AnyOf(set!{
                                Schema::Atomic(Atomic::Integer),
                                Schema::Atomic(Atomic::Long)
                            })
                        },
                        required: set!("x".to_string()),
                        ..Default::default()
                    })
                ))
            },
            required: set!("o1".to_string(), "outputField2".to_string()),
            ..Default::default()
        })),
        input = r#"{"$facet": {
            "o1": [{"$limit": 10}, {"$project": {"_id": 0}}],
            "outputField2": [{"$count": "x"}]
        }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId)
            },
            required: set!("_id".to_string()),
            ..Default::default()
        })
    );
}

mod fill {
    use super::*;

    test_derive_stage_schema!(
        fill,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)}),
                "bar".to_string() => Schema::Atomic(Atomic::String),
                "baz".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)}),
            },
            required: set!("foo".to_string(), "bar".to_string(), "baz".to_string()),
            ..Default::default()
        })),
        input = r#"{"$fill": {"output": {"foo": {"value": {"$add": [3, 4]}}, "bar": {"method": "linear"}, "baz": {"value": null}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)}),
                "bar".to_string() => Schema::Atomic(Atomic::String),
                "baz".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Null)}),
            },
            required: set!(),
            ..Default::default()
        })
    );
}

mod group {
    use super::*;

    test_derive_stage_schema!(
        simple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::String),
                "count".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                ))
            },
            required: set!("_id".to_string(), "count".to_string()),
            ..Default::default()
        })),
        input = r#"{"$group": {"_id": "$foo", "count": {"$sum": 1}}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );

    test_derive_stage_schema!(
        multiple_keys,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Document(Document {
                    keys: map! {
                        "foo".to_string() => Schema::Atomic(Atomic::String),
                        "bar".to_string() => Schema::Atomic(Atomic::Integer)
                    },
                    required: set!("foo".to_string(), "bar".to_string()),
                    ..Default::default()
                }),
                "count".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                )),
                "sum".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                ))
            },
            required: set!("_id".to_string(), "count".to_string(), "sum".to_string()),
            ..Default::default()
        })),
        input = r#"{"$group": {"_id": {"foo": "$foo", "bar": "$bar"}, "count": {"$sum": 1}, "sum": {"$sum": "$bar"}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set!("foo".to_string(), "bar".to_string()),
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
        field_ref_nested_array,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "bar".to_string() => Schema::Atomic(Atomic::Double)
                    },
                    required: set!("bar".to_string()),
                    ..Default::default()
                })))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$unwind": "$foo.bar" }"#,
        ref_schema = Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
            },
            required: set!("bar".to_string()),
            ..Default::default()
        })))
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
        field_ref_nested_anyofs_in_path,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Atomic(Atomic::Double)
                            },
                            required: set!("c".to_string()),
                            ..Default::default()
                        })
                    },
                    required: set!("b".to_string()),
                    ..Default::default()
                })
            },
            required: set!("a".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unwind": {"path": "$a.b.c"}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Null),
                            Schema::Atomic(Atomic::Integer),
                            Schema::Document(Document {
                                keys: map! {
                                    "c".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                                },
                                ..Default::default()
                            })
                        ))
                    },
                    ..Default::default()
                })
            },
            required: set!("a".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        field_ref_nested_anyofs_in_path_preserve_null_and_empty_arrays,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Null),
                            Schema::Atomic(Atomic::Integer),
                            Schema::Document(Document {
                                keys: map! {
                                    "c".to_string() => Schema::Atomic(Atomic::Double)
                                },
                                ..Default::default()
                            })
                        )),
                    },
                    ..Default::default()
                })
            },
            required: set!("a".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unwind": {"path": "$a.b.c", "preserveNullAndEmptyArrays": true}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::Null),
                            Schema::Atomic(Atomic::Integer),
                            Schema::Document(Document {
                                keys: map! {
                                    "c".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Double)))
                                },
                                ..Default::default()
                            })
                        ))
                    },
                    ..Default::default()
                })
            },
            required: set!("a".to_string()),
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
                "i".to_string() => Schema::Atomic(Atomic::Long)
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
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("i".to_string()),
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
    test_derive_stage_schema!(
        subquery_lookup_documents_instead_of_from,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
                "bar".to_string() => Schema::Array(
                    Box::new(Schema::Document(Document {
                        keys: map! {
                            "out".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("out".to_string()),
                        ..Default::default()
                    }))
                ),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$lookup": {
            "let": {"x": "foo"},
            "pipeline": [
                {"$documents": [{"baz": "string"}]},
                {"$project": {"out": {"$concat": ["$$x", "$baz"]}}}
            ],
            "as": "bar"
        }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer)
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

mod set_window_fields {
    use super::*;

    test_derive_stage_schema!(
        set_windows_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "documents".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                )),
                "no_window".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Null)
                )),
                "range_and_unit".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long)
                )),
                "set".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "push".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "avg".to_string() => Schema::Atomic(Atomic::Null),
                "bottom".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "bottomN".to_string() => Schema::Array(Box::new(Schema::Array(Box::new(Schema::Atomic(Atomic::String))))),
                "count".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long)
                ))
            },
            required: set!(
                "foo".to_string(),
                "documents".to_string(),
                "no_window".to_string(),
                "range_and_unit".to_string(),
                "set".to_string(),
                "push".to_string(),
                "avg".to_string(),
                "bottom".to_string(),
                "bottomN".to_string(),
                "count".to_string()
            ),
            ..Default::default()
        })),
        input = r#"{"$setWindowFields": {
                        "output": {
                            "documents": {
                                "$sum": 1,
                                "window": {
                                    "documents": [-1, 1]
                                }
                            },
                            "no_window": {
                                "$derivative": {
                                    "input": 1,
                                    "unit": "seconds"
                                }
                            },
                            "range_and_unit": {
                                "$denseRank": {},
                                "window": {
                                    "range": [-10, 10],
                                    "unit": "seconds"
                                }
                            },
                            "set": {
                                "$addToSet": "$foo"
                            },
                            "push": {
                                "$push": "$foo"
                            },
                            "avg": {
                                "$avg": "$foo"
                            },
                            "bottom": {
                                "$bottom":
                                {
                                    "output": [ "$foo" ],
                                    "sortBy": { "score": -1 }
                                }
                            },
                            "bottomN": {
                                "$bottomN":
                                {
                                    "n": 2,
                                    "output": [ "$foo" ],
                                    "sortBy": { "score": -1 }
                                }
                            },
                            "count": {
                                "$count": {}
                            }
                        }
                }}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
}

mod project {
    use super::*;

    test_derive_stage_schema!(
        project_empty,
        expected = Err(crate::Error::InvalidProjectStage),
        input = r#"{"$project": {}}"#
    );
    test_derive_stage_schema!(
        project_simple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("_id".to_string(), "foo".to_string(), "bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo": 1, "bar": {"$concat": ["$foo", "hello"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::Integer),
                "baz".to_string() => Schema::Atomic(Atomic::Boolean),
            },
            required: set!(
                "_id".to_string(),
                "foo".to_string(),
                "bar".to_string(),
                "baz".to_string()
            ),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        project_non_required_key,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            ..Default::default()
        })),
        input = r#"{"$project": {"foo": 1}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        project_additional_properties_true,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Any
                    },
                    ..Default::default()
                })
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 1}}"#,
        ref_schema = Schema::Document(Document::any())
    );
    test_derive_stage_schema!(
        project_any,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Document(Document {
                        keys: map! {
                            "a".to_string() => Schema::Any
                        },
                        ..Default::default()
                    }),
                    Schema::Array(Box::new(Schema::Document(Document {
                        keys: map! {
                            "a".to_string() => Schema::Any
                        },
                        ..Default::default()
                    })))
                ))
            },
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 1}}"#,
        ref_schema = Schema::Any
    );
    test_derive_stage_schema!(
        project_multiple_fields_array_of_docs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::String),
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    ..Default::default()
                })))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 1, "foo.b": 1}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
            },
            ..Default::default()
        })))
    );
    test_derive_stage_schema!(
        project_multiple_fields_but_not_all_array_fields_include,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::String),
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    ..Default::default()
                })))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 1, "foo.b": 1}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Integer),
                "d".to_string() => Schema::Atomic(Atomic::Integer),
            },
            ..Default::default()
        })))
    );
    test_derive_stage_schema!(
        project_include_multiple_neighboring_fields_with_assignment,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Array(Box::new(Schema::Document(Document {
                        keys: map! {
                            "a".to_string() => Schema::Atomic(Atomic::String),
                            "b".to_string() => Schema::Atomic(Atomic::Integer),
                            "c".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("a".to_string(), "c".to_string()),
                        ..Default::default()
                    }))),
                    Schema::Document(Document {
                        keys: map! {
                            "a".to_string() => Schema::Atomic(Atomic::String),
                            "b".to_string() => Schema::Atomic(Atomic::Integer),
                            "c".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("b".to_string(), "c".to_string()),
                        ..Default::default()
                    })
                ))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 1, "foo.b": 1, "foo.c": "hello world"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Array(Box::new(Schema::Document(Document {
                keys: map! {
                    "a".to_string() => Schema::Atomic(Atomic::String),
                    "b".to_string() => Schema::Atomic(Atomic::Integer),
                },
                required: set!("a".to_string()),
                ..Default::default()
            }))),
            Schema::Document(Document {
                keys: map! {
                    "a".to_string() => Schema::Atomic(Atomic::String),
                    "b".to_string() => Schema::Atomic(Atomic::Integer),
                },
                required: set!("b".to_string()),
                ..Default::default()
            })
        ))
    );
    // this test aims to cover all possible combinations of 3 level fields paths where we
    // include multiple fields. That is, a stage that looks like the following:
    // db.aggregate([
    //     {$documents: [
    //         {_id: 1, foo: [{bar: {a: 1, b: 2}}]},
    //         {_id: 2, foo: [{bar: [{a: 1, b: 2}]}]},
    //         {_id: 3, foo: {bar: [{a: 1, b: 2}]}},
    //         {_id: 4, foo: {bar: {a: 1, b: 1}}},
    //     ]},
    //     {$project: {"foo.bar.a": 1, "foo.bar.b": 1}}
    // ])
    test_derive_stage_schema!(
        project_multiple_fields_mixed_docs_arrays,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Array(Box::new(Schema::Document(Document {
                        keys: map! {
                            "bar".to_string() => Schema::AnyOf(set!(
                                Schema::Document(Document {
                                    keys: map! {
                                        "a".to_string() => Schema::Atomic(Atomic::String),
                                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                                    },
                                    ..Default::default()
                                }),
                                Schema::Array(Box::new(Schema::Document(Document {
                                    keys: map! {
                                        "a".to_string() => Schema::Atomic(Atomic::String),
                                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                                    },
                                    ..Default::default()
                                })))
                            ))
                        },
                        ..Default::default()
                    }))),
                    Schema::Document(Document {
                        keys: map! {
                            "bar".to_string() => Schema::AnyOf(set!(
                                Schema::Document(Document {
                                    keys: map! {
                                        "a".to_string() => Schema::Atomic(Atomic::String),
                                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                                    },
                                    ..Default::default()
                                }),
                                Schema::Array(Box::new(Schema::Document(Document {
                                    keys: map! {
                                        "a".to_string() => Schema::Atomic(Atomic::String),
                                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                                    },
                                    ..Default::default()
                                })))
                            ))
                        },
                        ..Default::default()
                    })
                ))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.bar.a": 1, "foo.bar.b": 1}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Array(Box::new(Schema::Document(Document {
                keys: map! {
                    "bar".to_string() => Schema::AnyOf(set!(
                        Schema::Document(Document {
                            keys: map! {
                                "a".to_string() => Schema::Atomic(Atomic::String),
                                "b".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        }),
                        Schema::Array(Box::new(Schema::Document(Document {
                            keys: map! {
                                "a".to_string() => Schema::Atomic(Atomic::String),
                                "b".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        })))
                    ))
                },
                ..Default::default()
            }))),
            Schema::Document(Document {
                keys: map! {
                    "bar".to_string() => Schema::AnyOf(set!(
                        Schema::Document(Document {
                            keys: map! {
                                "a".to_string() => Schema::Atomic(Atomic::String),
                                "b".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        }),
                        Schema::Array(Box::new(Schema::Document(Document {
                            keys: map! {
                                "a".to_string() => Schema::Atomic(Atomic::String),
                                "b".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        })))
                    ))
                },
                ..Default::default()
            })
        ))
    );
    test_derive_stage_schema!(
        project_remove_id,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"_id": 0, "foo": 1, "bar": {"$concat": ["$foo", "hello"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::Integer),
                "baz".to_string() => Schema::Atomic(Atomic::Boolean),
            },
            required: set!(
                "_id".to_string(),
                "foo".to_string(),
                "bar".to_string(),
                "baz".to_string()
            ),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        project_exclude,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "baz".to_string() => Schema::Atomic(Atomic::Boolean),
            },
            required: set!("_id".to_string(), "baz".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo": 0, "bar": 0}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::Integer),
                "baz".to_string() => Schema::Atomic(Atomic::Boolean),
            },
            required: set!(
                "_id".to_string(),
                "foo".to_string(),
                "bar".to_string(),
                "baz".to_string()
            ),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        project_exclude_nested_anyof,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("_id".to_string(), "foo".to_string(),),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.bar": 0 }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::String),
                    Schema::Document(Document {
                        keys: map! {
                            "bar".to_string() => Schema::Atomic(Atomic::Integer)
                        },
                        ..Default::default()
                    })
                )),
            },
            required: set!("_id".to_string(), "foo".to_string(),),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        project_include_nested_anyof,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "bar".to_string() => Schema::Atomic(Atomic::Integer)
                    },
                    ..Default::default()
                })
            },
            required: set!("_id".to_string(), "foo".to_string(),),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.bar": 1 }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::ObjectId),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::String),
                    Schema::Document(Document {
                        keys: map! {
                            "bar".to_string() => Schema::Atomic(Atomic::Integer)
                        },
                        ..Default::default()
                    })
                )),
            },
            required: set!("_id".to_string(), "foo".to_string(),),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        project_include_exclude,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    ..Default::default()
                })))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 1, "foo.a": 0}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
            },
            ..Default::default()
        })))
    );
    test_derive_stage_schema!(
        project_exclude_include,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Atomic(Atomic::String),
                    },
                    ..Default::default()
                })))
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a": 0, "foo.a": 1}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
            },
            ..Default::default()
        })))
    );
    test_derive_stage_schema!(
        project_nested_fields_not_required,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "a".to_string() => Schema::Document(Document {
                            keys: map! {
                                "b".to_string() => Schema::Atomic(Atomic::Double)
                            },
                            ..Default::default()
                        }),
                    },
                    ..Default::default()
                })
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$project": {"foo.a.b": 1}}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Double)
                    },
                    ..Default::default()
                }),
            },
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
        input = r#"{"$unionWith": {"coll": "bar", "pipeline": [{"$project": {"out": {"$concat": ["$baz", "$baz"]}}}]}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "food".to_string() => Schema::Atomic(Atomic::String),
                "out".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("foo".to_string(), "out".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        union_pipeline_no_collection,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::Integer),
                "a".to_string() => Schema::Atomic(Atomic::String),
                "food".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Integer)
                )),
                "out".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("food".to_string()),
            ..Default::default()
        })),
        input = r#"{"$unionWith": {"pipeline": [{"$documents": [{"_id": 1, "a": "hello world", "food": 1}]}]}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "food".to_string() => Schema::Atomic(Atomic::String),
                "out".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("food".to_string(), "out".to_string()),
            ..Default::default()
        })
    );
    test_derive_stage_schema!(
        union_not_enough_args,
        expected = Err(crate::Error::NotEnoughArguments("$unionWith".to_string())),
        input = r#"{"$unionWith": {}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "food".to_string() => Schema::Atomic(Atomic::String),
                "out".to_string() => Schema::Atomic(Atomic::Decimal),
            },
            required: set!("food".to_string(), "out".to_string()),
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

mod search {
    use super::*;

    test_derive_stage_schema!(
        search,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$search": {
            "near": {
                "path": "released",
                "origin": "2011-09-01T00:00:00.000+00:00",
                "pivot": 7776000000
            }
        }}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        vector_search,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String)
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$vectorSearch": {
                "exact": true,
                "filter": {},
                "index": "x",
                "limit": 23,
                "numCandidates": 42,
                "path": "baz",
                "queryVector": [1,2,3,41]
            }
        }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        search_meta,
        expected = Ok(crate::schema_derivation::SEARCH_META.clone()),
        input = r#"{
            "$searchMeta": {
                "range": {
                    "path": "year",
                    "gte": 1998,
                    "lt": 1999
                },
                "count": {
                    "type": "total"
                }
            }
        }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
                "bar".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string(), "bar".to_string()),
            ..Default::default()
        })
    );
}

mod unset_fields {
    use super::*;

    test_derive_stage_schema!(
        unset_single_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::Document(Document {
                    keys: map! {
                        "first".to_string() => Schema::Atomic(Atomic::String),
                        "last".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("first".to_string(), "last".to_string()),
                    ..Default::default()
                }),
            },
            required: set!("title".to_string(), "author".to_string(),),
            ..Default::default()
        })),
        input = r#"{ "$unset": "isbn" }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "isbn".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::Document(Document {
                    keys: map! {
                        "first".to_string() => Schema::Atomic(Atomic::String),
                        "last".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("first".to_string(), "last".to_string()),
                    ..Default::default()
                }),
            },
            required: set!(
                "title".to_string(),
                "author".to_string(),
                "isbn".to_string(),
            ),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        unset_multiple_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("title".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$unset": ["isbn", "author"] }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "isbn".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::Document(Document {
                    keys: map! {
                        "first".to_string() => Schema::Atomic(Atomic::String),
                        "last".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("first".to_string(), "last".to_string()),
                    ..Default::default()
                }),
            },
            required: set!(
                "title".to_string(),
                "author".to_string(),
                "isbn".to_string(),
            ),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        unset_nested_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "isbn".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::Document(Document {
                    keys: map! {
                        "last".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("last".to_string()),
                    ..Default::default()
                }),
            },
            required: set!(
                "title".to_string(),
                "author".to_string(),
                "isbn".to_string(),
            ),
            ..Default::default()
        })),
        input = r#"{ "$unset": "author.first"}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "isbn".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::Document(Document {
                    keys: map! {
                        "first".to_string() => Schema::Atomic(Atomic::String),
                        "last".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("first".to_string(), "last".to_string()),
                    ..Default::default()
                }),
            },
            required: set!(
                "title".to_string(),
                "author".to_string(),
                "isbn".to_string(),
            ),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        unset_nested_anyof_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "isbn".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::AnyOf(set!(
                    Schema::Document(Document {
                        keys: map! {
                            "last".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("last".to_string()),
                        ..Default::default()
                    }),
                    Schema::Atomic(Atomic::String)
                )),
            },
            required: set!(
                "title".to_string(),
                "author".to_string(),
                "isbn".to_string(),
            ),
            ..Default::default()
        })),
        input = r#"{ "$unset": "author.first"}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "title".to_string() => Schema::Atomic(Atomic::String),
                "isbn".to_string() => Schema::Atomic(Atomic::String),
                "author".to_string() => Schema::AnyOf(set!(
                    Schema::Document(Document {
                        keys: map! {
                            "first".to_string() => Schema::Atomic(Atomic::String),
                            "last".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("first".to_string(), "last".to_string()),
                        ..Default::default()
                    }),
                    Schema::Atomic(Atomic::String),
                )),
            },
            required: set!(
                "title".to_string(),
                "author".to_string(),
                "isbn".to_string(),
            ),
            ..Default::default()
        })
    );

    test_derive_stage_schema!(
        unset_non_existing_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("x".to_string(), "y".to_string(),),
            ..Default::default()
        })),
        input = r#"{ "$unset": "z"}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("x".to_string(), "y".to_string(),),
            ..Default::default()
        })
    );
}
