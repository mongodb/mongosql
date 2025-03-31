use crate::{
    schema_derivation::{DeriveSchema, ResultSetState},
    Error,
};
use agg_ast::definitions::Expression;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema},
    set,
};
use std::collections::BTreeMap;

mod literal {
    use super::*;

    test_derive_expression_schema!(
        literal_binary,
        expected = Ok(Schema::Atomic(Atomic::BinData)),
        input = r#"{"$binary" : {"base64" : "", "subType" : "04"}}"#
    );

    test_derive_expression_schema!(
        literal_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = r#"true"#
    );

    test_derive_expression_schema!(
        literal_date,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$date": { "$numberLong": "1655956513000" } }"#
    );

    test_derive_expression_schema!(
        literal_dbpointer,
        expected = Ok(Schema::Atomic(Atomic::DbPointer)),
        input =
            r#"{ "$dbPointer": { "$ref": "foo", "$id": { "$oid": "57e193d7a9cc81b4027498b5" } } }"#
    );

    test_derive_expression_schema!(
        literal_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{ "$numberDecimal": "3.0" }"#
    );

    test_derive_expression_schema!(
        literal_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{ "$numberDouble": "3.0" }"#
    );

    test_derive_expression_schema!(
        literal_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{ "$numberInt": "3" }"#
    );

    test_derive_expression_schema!(
        literal_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = r#"{ "$numberLong": "3" }"#
    );

    test_derive_expression_schema!(
        literal_javascript,
        expected = Ok(Schema::Atomic(Atomic::Javascript)),
        input = r#"{ "$code": "function() {}" }"#
    );

    test_derive_expression_schema!(
        literal_javascript_with_scope,
        expected = Ok(Schema::Atomic(Atomic::JavascriptWithScope)),
        input = r#"{ "$code": "function() {}", "$scope": { } }"#
    );

    test_derive_expression_schema!(
        literal_maxkey,
        expected = Ok(Schema::Atomic(Atomic::MaxKey)),
        input = r#"{ "$maxKey": 1 }"#
    );

    test_derive_expression_schema!(
        literal_minkey,
        expected = Ok(Schema::Atomic(Atomic::MinKey)),
        input = r#"{ "$minKey": 1 }"#
    );

    test_derive_expression_schema!(
        literal_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"null"#
    );

    test_derive_expression_schema!(
        literal_oid,
        expected = Ok(Schema::Atomic(Atomic::ObjectId)),
        input = r#"{"$oid": "5d505646cf6d4fe581014ab2"}"#
    );

    test_derive_expression_schema!(
        literal_regex,
        expected = Ok(Schema::Atomic(Atomic::Regex)),
        input = r#" { "$regularExpression": { "pattern": "abc*", "options": "ix" } }"#
    );

    test_derive_expression_schema!(
        literal_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#""foo bar""#
    );

    test_derive_expression_schema!(
        literal_symbol,
        expected = Ok(Schema::Atomic(Atomic::Symbol)),
        input = r#"{ "$symbol": "sym2" }"#
    );

    test_derive_expression_schema!(
        literal_timestamp,
        expected = Ok(Schema::Atomic(Atomic::Timestamp)),
        input = r#"{ "$timestamp": { "t": 42, "i": 1 } }"#
    );
}

mod field_ref {
    use super::*;

    test_derive_expression_schema!(
        ref_present,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$$foo""#,
        ref_schema = Schema::Any,
        variables = map! {
            "foo".to_string() => Schema::Atomic(Atomic::Double)
        }
    );

    test_derive_expression_schema!(
        variable_ref_nested,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$$a.b.c""#,
        ref_schema = Schema::Any,
        variables = map! {
            "a".to_string() => Schema::Document(Document {
                keys: map! {
                    "b".to_string() => Schema::Document(Document {
                        keys: map! {
                            "c".to_string() => Schema::Atomic(Atomic::Double)
                        },
                        ..Default::default()
                    })
                },
                ..Default::default()
            })
        }
    );

    test_derive_expression_schema!(
        variable_ref_missing,
        expected = Err(Error::UnknownReference("foo".to_string())),
        input = r#""$$foo""#,
        ref_schema = Schema::Any,
        variables = map!()
    );

    test_derive_expression_schema!(
        field_ref,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$foo""#,
        ref_schema = Schema::Atomic(Atomic::Double),
        variables = map!()
    );

    test_derive_expression_schema!(
        nested_ref_present,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Missing, Schema::Atomic(Atomic::Double)}
        )),
        input = r#""$foo.bar""#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::Atomic(Atomic::Double)
            },
            ..Default::default()
        }),
        variables = map!()
    );

    test_derive_expression_schema!(
        ref_missing,
        expected = Ok(Schema::Missing),
        input = r#""$foo""#
    );

    test_derive_expression_schema!(
        nested_ref_set_current,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$foo.bar""#,
        variables = map! {
            "CURRENT".to_string() => Schema::Document(Document {
                keys: map! {
                    "foo".to_string() => Schema::Document(Document {
                        keys: map! {
                            "bar".to_string() => Schema::Atomic(Atomic::Double)
                        },
                        required: set!["bar".to_string()],
                        ..Default::default()
                    })
                },
                required: set!["foo".to_string()],
                ..Default::default()
            })
        }
    );
}

mod variable_ref {
    use super::*;

    test_derive_expression_schema!(
        ref_present,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$$foo""#,
        ref_schema = Schema::Any,
        variables = map! {
            "foo".to_string() => Schema::Atomic(Atomic::Double)
        }
    );

    test_derive_expression_schema!(
        ref_nested_present,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$$foo.bar""#,
        ref_schema = Schema::Any,
        variables = map! {
            "foo".to_string() => Schema::Document(Document {
                keys: map! {
                    "bar".to_string() => Schema::Atomic(Atomic::Double)
                },
                ..Default::default()
            })
        }
    );

    test_derive_expression_schema!(
        ref_nested_top_level_present,
        expected = Ok(Schema::Document(Document::empty())),
        input = r#""$$foo.bar""#,
        ref_schema = Schema::Any,
        variables = map! {
            // foo can be anything -- if the top level key is present in the
            // var map, and the nested path is not, we will get an exmpty document.
            "foo".to_string() => Schema::Atomic(Atomic::Integer)
        }
    );

    test_derive_expression_schema!(
        ref_missing,
        expected = Err(Error::UnknownReference("foo".to_string())),
        input = r#""$$foo""#,
        ref_schema = Schema::Any,
        variables = map!()
    );

    test_derive_expression_schema!(
        now,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#""$$NOW""#
    );

    test_derive_expression_schema!(
        current_time,
        expected = Ok(Schema::Atomic(Atomic::Timestamp)),
        input = r#""$$CURRENT_TIME""#
    );

    test_derive_expression_schema!(
        root,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Date)
            },
            required: set!["foo".to_string()],
            ..Default::default()
        })),
        input = r#""$$ROOT""#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Date)
            },
            required: set!["foo".to_string()],
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        current,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Date)
            },
            required: set!["foo".to_string()],
            ..Default::default()
        })),
        input = r#""$$CURRENT""#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Date)
            },
            required: set!["foo".to_string()],
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        user_roles,
        expected = Ok(Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::String),
                "role".to_string() => Schema::Atomic(Atomic::String),
                "db".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("_id".to_string(), "role".to_string(), "db".to_string()),
            ..Default::default()
        })))),
        input = r#""$$USER_ROLES""#
    );

    test_derive_expression_schema!(
        search_meta,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "count".to_string() => Schema::Document(Document {
                    keys: map! {
                        "total".to_string() => Schema::Atomic(Atomic::Long),
                        "lowerBound".to_string() => Schema::Atomic(Atomic::Long),
                    },
                    required: set![],
                    ..Default::default()
                }),
            },
            required: set!["count".to_string(),],
            ..Default::default()
        })),
        input = r#""$$SEARCH_META""#
    );
}

mod array {
    use super::*;

    test_derive_expression_schema!(
        empty_array,
        expected = Ok(Schema::Array(Box::new(Schema::Unsat))),
        input = r#"[]"#
    );

    test_derive_expression_schema!(
        array_single_type,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input = r#"[1, 2, 3]"#
    );

    test_derive_expression_schema!(
        array_multiple_types,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))))),
        input = r#"[1, 2, 3, "foo", "bar"]"#
    );

    test_derive_expression_schema!(
        nested_arrays_not_merged,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Array(Box::new(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String)
            )))),
            Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
        ))))),
        input = r#"[[1, 2, "hi"], ["foo", "bar"]]"#
    );
}

mod document {
    use super::*;

    test_derive_expression_schema!(
        empty_document,
        expected = Ok(Schema::Document(Document::default())),
        input = r#"{}"#
    );

    test_derive_expression_schema!(
        document_simple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("a".to_string(), "b".to_string()),
            ..Default::default()
        })),
        input = r#"{"a": 1, "b": "foo"}"#
    );

    test_derive_expression_schema!(
        document_nested,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Document(Document { keys: map! {
                    "c".to_string() => Schema::Atomic(Atomic::String),
                    "d".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Boolean))),
                }, required: set!("c".to_string(), "d".to_string()), ..Default::default() }),
            },
            required: set!("a".to_string(), "b".to_string()),
            ..Default::default()
        })),
        input = r#"{"a": 1, "b": {"c": "hi", "d": [false]}}"#
    );
}

mod group_accumulator {
    use super::*;

    test_derive_expression_schema!(
        first,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))),
        input = r#"{"$first": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        last,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))),
        input = r#"{"$last": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        max,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$max": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        min,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$min": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        avg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$avg": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))
    );
    test_derive_expression_schema!(
        avg_decimal,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ))),
        input = r#"{"$avg": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal)
        ))
    );
    test_derive_expression_schema!(
        avg_or_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$avg": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        avg_decimal_or_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$avg": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Decimal)
        ))
    );
    test_derive_expression_schema!(
        sum,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal)
        ))),
        input = r#"{"$sum": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sum_double,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal)
        ))),
        input = r#"{"$sum": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        push,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))))),
        input = r#"{"$push": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        add_to_set,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))))),
        input = r#"{"$addToSet": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        std_dev_pop,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$stdDevPop": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        std_dev_samp,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$stdDevSamp": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        std_dev_pop_or_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$stdDevPop": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
            Schema::Missing
        ))
    );
    test_derive_expression_schema!(
        std_dev_samp_or_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$stdDevSamp": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
            Schema::Missing
        ))
    );
    test_derive_expression_schema!(
        merge_objects,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("a".to_string(), "b".to_string()),
            ..Default::default()
        })),
        input = r#"{"$mergeObjects": "$foo"}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("a".to_string(), "b".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        sql_first,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))),
        input = r#"{"$sqlFirst": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_last,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))),
        input = r#"{"$sqlLast": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_max,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$sqlMax": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_min,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$sqlMin": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_avg,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$sqlAvg": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer),))
    );
    test_derive_expression_schema!(
        sql_avg_decimal,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$sqlAvg": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Decimal)
        ))
    );
    test_derive_expression_schema!(
        sql_sum,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal)
        ))),
        input = r#"{"$sqlSum": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_count,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ))),
        input = r#"{"$sqlCount": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_std_dev_pop,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null)
        ))),
        input =
            r#"{"$sqlStdDevPop": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_std_dev_samp,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$sqlStdDevSamp": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))
    );
    test_derive_expression_schema!(
        sql_merge_objects,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("a".to_string(), "b".to_string()),
            ..Default::default()
        })),
        input = r#"{"$sqlMergeObjects": {"distinct": false, "var": "$foo", "arg_is_possibly_doc": null}}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("a".to_string(), "b".to_string()),
            ..Default::default()
        })
    );
}
