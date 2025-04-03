use crate::schema_derivation::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Expression;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema},
    set,
};
use std::collections::BTreeMap;

mod misc_ops {
    use super::*;

    test_derive_expression_schema!(
        constant_integral,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ))),
        input = r#"{ "$documentNumber": { } }"#
    );

    // $cond
    test_derive_expression_schema!(
        cond,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        input = r#"{ "$cond": [true, 1, "yes"] }"#
    );
    // $shift
    test_derive_expression_schema!(
        shift_no_default_uses_null,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        })),
        input = r#"{ "$shift": { "output": 1, "by": 1 } }"#
    );
    test_derive_expression_schema!(
        shift_with_default,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        })),
        input = r#"{ "$shift": { "output": 1, "by": 1, "default": "abc" } }"#
    );

    // $switch
    test_derive_expression_schema!(
        switch,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        })),
        input = r#"{ "$switch": {
                        "branches": [
                            { "case": false, "then": 1 },
                            { "case": true, "then": "yes" }
                        ],
                        "default": null
        }}"#
    );
}

mod window_ops {
    use super::*;
    test_derive_expression_schema!(
        window_func_decimal,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$derivative": { "input": "$foo", "unit": "hour" } }"#,
        ref_schema = Schema::Atomic(Atomic::Decimal)
    );

    test_derive_expression_schema!(
        window_func_double_by_default,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$derivative": { "input": 123, "unit": "hour" } }"#
    );

    test_derive_expression_schema!(
        window_func_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$derivative": { "input": null, "unit": "hour" } }"#
    );

    test_derive_expression_schema!(
        window_func_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$derivative": { "input": "$foo", "unit": "hour" } }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Missing, Schema::Atomic(Atomic::Integer)))
    );

    test_derive_expression_schema!(
        locf,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Missing,
        ))),
        input = r#"{ "$locf": "$foo" }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Missing, Schema::Atomic(Atomic::Integer)))
    );

    test_derive_expression_schema!(
        exp_mov_avg,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$expMovingAvg": 
            { "input": "$foo", "N": 3, "alpha": 0.5 }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_expression_schema!(
        integral,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ))),
        input = r#"{"$integral": 
            { "input": "$foo", "unit": "minute" }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
    test_derive_expression_schema!(
        last_n,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))))),
        input = r#"{"$lastN": 
            { "input": "$foo", "n": 3 }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
    test_derive_expression_schema!(
        max_n,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))))),
        input = r#"{"$maxN": 
            { "input": "$foo", "n": 3 }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
    test_derive_expression_schema!(
        median,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$median": {
            "input": "$foo",
            "method": "approximate"
          }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_expression_schema!(
        min_n,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))))),
        input = r#"{"$minN": 
            { "input": "$foo", "n": 3 }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
    test_derive_expression_schema!(
        percentile,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Double),
        ))))),
        input = r#"{"$percentile": 
          {
            "input": "$foo",
            "p": [0.5, 0.6],
            "method": "approximate"
          }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_expression_schema!(
        top,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))))),
        input = r#"{"$top": 
            { 
              "sortBy": {"foo": 1}, 
              "output": ["$foo"]
            }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
    test_derive_expression_schema!(
        top_n,
        expected = Ok(Schema::Array(Box::new(Schema::Array(Box::new(
            Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Decimal),
            ))
        ))))),
        input = r#"{"$topN": 
            {
              "n": 4,
              "sortBy": {"foo": 1}, 
              "output": ["$foo"]
            }
        }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
}

mod array_ops {
    use super::*;
    test_derive_expression_schema!(
        zip_no_defaults,
        expected = Ok(Schema::Array(Box::new(Schema::Array(Box::new(
            Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ))
        ))))),
        input = r#"{"$zip": {"inputs": ["$foo", [1,2,3]]}}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))))
    );
    test_derive_expression_schema!(
        zip_defaults,
        expected = Ok(Schema::Array(Box::new(Schema::Array(Box::new(
            Schema::AnyOf(set!(
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::Integer),
            ))
        ))))),
        input = r#"{"$zip": {"inputs": ["$foo", [1,2,3]], "defaults": ["hello"]}}"#,
        ref_schema = Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))
    );
    test_derive_expression_schema!(
        zip_error,
        expected = Err(crate::Error::InvalidExpressionForField(
            "Literal(String(\"hello\"))".to_string(),
            "inputs"
        )),
        input = r#"{"$zip": {"inputs": "hello"}}"#
    );
    test_derive_expression_schema!(
        map_this,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))),
        input = r#"{"$map": {"input": "$foo", "in": "$$this"}}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))
    );
    test_derive_expression_schema!(
        map_as,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))),
        input = r#"{"$map": {"input": "$foo", "as": "bar", "in": "$$bar"}}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))
    );
    test_derive_expression_schema!(
        map_maybe_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Array(Box::new(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ))))
        ))),
        input = r#"{"$map": {"input": "$foo", "as": "bar", "in": "$$bar"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Array(Box::new(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ))))
        ))
    );
    test_derive_expression_schema!(
        map_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$map": {"input": "$foo", "as": "bar", "in": "$$bar"}}"#,
        ref_schema = Schema::Atomic(Atomic::Null)
    );
    test_derive_expression_schema!(
        map_error,
        expected = Err(crate::Error::InvalidExpressionForField(
            "Literal(String(\"hello\"))".to_string(),
            "input"
        )),
        input = r#"{"$map": {"input": "hello", "in": "foo"}}"#
    );
    test_derive_expression_schema!(
        reduce,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))),
        input = r#"{"$reduce": {"input": "$foo", "initialValue": 42, "in": "$$this"}}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))
    );
    test_derive_expression_schema!(
        reduce_maybe_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$reduce": {"input": "$foo", "initialValue": 42, "in": "$$this"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Array(Box::new(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            )))),
            Schema::Missing,
            Schema::Atomic(Atomic::Null)
        ))
    );
    test_derive_expression_schema!(
        reduce_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$reduce": {"input": "$foo", "initialValue": 42, "in": "$$this"}}"#,
        ref_schema = Schema::Atomic(Atomic::Null)
    );
    test_derive_expression_schema!(
        reduce_error,
        expected = Err(crate::Error::InvalidExpressionForField(
            "Literal(String(\"hello\"))".to_string(),
            "input"
        )),
        input = r#"{"$reduce": {"input": "hello", "initialValue": "stuff", "in": "foo"}}"#
    );
    test_derive_expression_schema!(
        replace_one,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$replaceOne": {"input": "$foo", "find": "x", "replacement": "y"}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
    test_derive_expression_schema!(
        replace_one_maybe_null_args,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}
        )),
        input = r#"{ "$replaceOne": { "input": "$i", "find": "$f", "replacement": "$r" } }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "i".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
                "f".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
                "r".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
            },
            required: set! {"i".to_string(), "f".to_string(), "r".to_string()},
            additional_properties: false,
            jaccard_index: None,
        })
    );
    test_derive_expression_schema!(
        replace_one_maybe_missing_args,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}
        )),
        input = r#"{ "$replaceOne": { "input": "$i", "find": "$f", "replacement": "$r" } }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "i".to_string() => Schema::Atomic(Atomic::String),
                "f".to_string() => Schema::Atomic(Atomic::String),
                "r".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: None,
        })
    );
    test_derive_expression_schema!(
        replace_all,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$replaceAll": {"input": "$foo", "find": "x", "replacement": "y"}}"#,
        ref_schema = Schema::Atomic(Atomic::String)
    );
    test_derive_expression_schema!(
        replace_all_maybe_null_args,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}
        )),
        input = r#"{ "$replaceAll": { "input": "$i", "find": "$f", "replacement": "$r" } }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "i".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
                "f".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
                "r".to_string() => Schema::AnyOf(set!{Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}),
            },
            required: set! {"i".to_string(), "f".to_string(), "r".to_string()},
            additional_properties: false,
            jaccard_index: None,
        })
    );
    test_derive_expression_schema!(
        replace_all_maybe_missing_args,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)}
        )),
        input = r#"{ "$replaceAll": { "input": "$i", "find": "$f", "replacement": "$r" } }"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "i".to_string() => Schema::Atomic(Atomic::String),
                "f".to_string() => Schema::Atomic(Atomic::String),
                "r".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: None,
        })
    );
}

mod group_ops {
    use super::*;
    test_derive_expression_schema!(
        top_expression,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)}
        )))),
        input = r#"{"$top": {
                     "output": [ "$foo", "hello" ],
                     "sortBy": { "score": -1 }
                   }}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );
    test_derive_expression_schema!(
        top_n_expression,
        expected = Ok(Schema::Array(Box::new(Schema::Array(Box::new(
            Schema::AnyOf(set! {Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)})
        ))))),
        input = r#"{"$topN": {
                     "output": [ "$foo", "hello" ],
                     "sortBy": { "score": -1 },
                     "n": 64
                   }}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );
    test_derive_expression_schema!(
        bottom_expression,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)}
        )))),
        input = r#"{"$bottom": {
                     "output": [ "$foo", "hello" ],
                     "sortBy": { "score": -1 }
                   }}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );
    test_derive_expression_schema!(
        bottom_n_expression,
        expected = Ok(Schema::Array(Box::new(Schema::Array(Box::new(
            Schema::AnyOf(set! {Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)})
        ))))),
        input = r#"{"$bottomN": {
                     "output": [ "$foo", "hello" ],
                     "sortBy": { "score": -1 },
                     "n": 64
                   }}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );
    test_derive_expression_schema!(
        max_n,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::MaxKey),
        })))),
        input = r#"{"$maxN": {"input": "$foo", "n": 64}}"#,
        ref_schema = Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::MaxKey),
        })
    );
    test_derive_expression_schema!(
        min_n,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::MaxKey),
        })))),
        input = r#"{"$minN": {"input": "$foo", "n": 64}}"#,
        ref_schema = Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::MaxKey),
        })
    );
    test_derive_expression_schema!(
        push,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ))))),
        input = r#"{"$push": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_expression_schema!(
        add_to_set,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ))))),
        input = r#"{"$addToSet": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
        ))
    );
}

mod numeric_ops {
    use super::*;

    test_derive_expression_schema!(
        median_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{ "$median": { "input": 123, "method": "approximate" } }"#
    );

    test_derive_expression_schema!(
        median_double_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$median": { "input": "$foo", "method": "approximate" } }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        median_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$median": { "input": "$foo", "method": "approximate" } }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Missing))
    );

    test_derive_expression_schema!(
        median_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$median": { "input": null, "method": "approximate" } }"#
    );
}

mod string_ops {
    use super::*;
    test_derive_expression_schema!(
        regex_find,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Document(Document {
                keys: map! {
                    "match".to_string() => Schema::Atomic(Atomic::String),
                    "idx".to_string() => Schema::Atomic(Atomic::Integer),
                    "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
                },
                required: set!(),
                ..Default::default()
            })
        ))),
        input = r#"{ "$regexFind": { "input": "$category", "regex": "/cafe/" }  }"#
    );

    test_derive_expression_schema!(
        regex_find_all,
        expected = Ok(Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "match".to_string() => Schema::Atomic(Atomic::String),
                "idx".to_string() => Schema::Atomic(Atomic::Integer),
                "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
            },
            required: set!(),
            ..Default::default()
        })))),
        input = r#"{ "$regexFindAll": { "input": "$category", "regex": "/cafe/" }  }"#
    );

    test_derive_expression_schema!(
        trim_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{ "$trim": { "input": "hi friend", "chars": "hi" }  }"#
    );

    test_derive_expression_schema!(
        trim_input_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$trim": { "input": null, "chars": "hi" }  }"#
    );

    test_derive_expression_schema!(
        trim_chars_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$trim": { "input": "hi", "chars": null } }"#
    );

    test_derive_expression_schema!(
        trim_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$trim": { "input": "$foo", "chars": "hi" }  }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        trim_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$trim": { "input": "$foo", "chars": "hi" }  }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Missing))
    );
}

mod date_operators {
    use super::*;
    test_derive_expression_schema!(
        date_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$hour": null }"#
    );

    test_derive_expression_schema!(
        date_integer,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{ "$hour": {"$date": {"$numberLong": "123"}} }"#
    );

    test_derive_expression_schema!(
        date_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
        ))),
        input = r#"{ "$hour": "$foo" }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Date),
        ))
    );

    test_derive_expression_schema!(
        date_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
        ))),
        input = r#"{ "$hour": "$foo" }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Missing, Schema::Atomic(Atomic::Date),))
    );

    test_derive_expression_schema!(
        date_timezone_specified_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$hour": {"date": {"$date": {"$numberLong": "123"}}, "timezone": null }}"#
    );

    test_derive_expression_schema!(
        date_from_string_simple,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$dateFromString": {"dateString": "hello" }}"#
    );

    test_derive_expression_schema!(
        date_from_string_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$dateFromString": {"dateString": null }}"#
    );

    test_derive_expression_schema!(
        date_from_string_null_with_on_null,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{ "$dateFromString": {"dateString": null, "onNull": 1 }}"#
    );

    test_derive_expression_schema!(
        date_from_string_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$dateFromString": {"dateString": "$foo" }}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_from_string_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$dateFromString": {"dateString": "$foo" }}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Missing))
    );

    test_derive_expression_schema!(
        date_from_parts_simple,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": 1 }}"#
    );

    test_derive_expression_schema!(
        date_from_parts_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ))
    );

    test_derive_expression_schema!(
        date_from_parts_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Missing))
    );

    test_derive_expression_schema!(
        date_from_parts_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": 3, "hour": null, "minute": 2, "second": 4}}"#
    );

    test_derive_expression_schema!(
        date_from_string_with_on_error,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Integer)
        ),)),
        input = r#"{ "$dateFromString": {"dateString": "hello", "onError": 1 }}"#
    );

    test_derive_expression_schema!(
        date_from_string_not_null_ignores_on_null,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$dateFromString": {"dateString": "hello", "onNull": 1 }}"#
    );

    test_derive_expression_schema!(
        date_from_string_nullish_fully_specified,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        input = r#"{ "$dateFromString": {"dateString": "$foo", "onError": 1, "onNull": "world" }}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_to_parts_standard,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "year".to_string() => Schema::Atomic(Atomic::Integer),
                "month".to_string() => Schema::Atomic(Atomic::Integer),
                "day".to_string() => Schema::Atomic(Atomic::Integer),
                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                "second".to_string() => Schema::Atomic(Atomic::Integer),
                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(
                "year".to_string(),
                "month".to_string(),
                "day".to_string(),
                "hour".to_string(),
                "minute".to_string(),
                "second".to_string(),
                "millisecond".to_string()
            ),
            ..Default::default()
        }),),
        input = r#"{ "$dateToParts": {"date": {"$date": {"$numberLong": "123"}} }}"#
    );

    test_derive_expression_schema!(
        date_to_parts_iso8601,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "isoDayOfWeek".to_string() => Schema::Atomic(Atomic::Integer),
                "isoWeek".to_string() => Schema::Atomic(Atomic::Integer),
                "isoWeekYear".to_string() => Schema::Atomic(Atomic::Integer),
                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                "second".to_string() => Schema::Atomic(Atomic::Integer),
                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(
                "isoDayOfWeek".to_string(),
                "isoWeek".to_string(),
                "isoWeekYear".to_string(),
                "hour".to_string(),
                "minute".to_string(),
                "second".to_string(),
                "millisecond".to_string()
            ),
            ..Default::default()
        }),),
        input = r#"{"$dateToParts": {"date":{"$date": {"$numberLong": "123"}}, "iso8601": true }}"#
    );

    test_derive_expression_schema!(
        date_to_parts_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateToParts": {"date": null, "iso8601": true }}"#
    );

    test_derive_expression_schema!(
        date_to_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$dateToString": {"date": "$foo"}}"#,
        ref_schema = Schema::Atomic(Atomic::Date)
    );

    test_derive_expression_schema!(
        date_to_string_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateToString": {"date": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_to_string_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateToString": {"date": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_expression_schema!(
        date_to_string_not_null_ignores_on_null,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$dateToString": {"date": "$foo", "onNull": 1}}"#,
        ref_schema = Schema::Atomic(Atomic::Date)
    );

    test_derive_expression_schema!(
        date_to_string_nullish_fully_specified,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Integer),
        ))),
        input = r#"{"$dateToString": {"date": "$foo", "format": "hi", "timezone": "America/New_York", "onNull": 1}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_addition_all_args,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{"$dateAdd": {"startDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#
    );

    test_derive_expression_schema!(
        date_addition_one_arg_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateAdd": {"startDate": null, "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#
    );

    test_derive_expression_schema!(
        date_addition_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateAdd": {"startDate": "$foo", "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_addition_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateAdd": {"startDate": "$foo", "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_expression_schema!(
        date_trunc_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateTrunc": {"date": "$foo", "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_trunc_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateTrunc": {"date": "$foo", "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_expression_schema!(
        date_trunc_all_args,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{"$dateTrunc": {"date": {"$date": {"$numberLong": "123"}}, "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );

    test_derive_expression_schema!(
        date_trunc_one_arg_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateTrunc": {"date": null, "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );

    test_derive_expression_schema!(
        date_diff_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateDiff": {"startDate": "$foo", "endDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_expression_schema!(
        date_diff_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateDiff": {"startDate": "$foo", "endDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_expression_schema!(
        date_diff_all_args,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{"$dateDiff": {"startDate": {"$date": {"$numberLong": "121"}}, "endDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );

    test_derive_expression_schema!(
        date_diff_one_arg_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateDiff": {"startDate": {"$date": {"$numberLong": "123"}}, "endDate": null, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );
}

mod field_setter_ops {
    use super::*;

    test_derive_expression_schema!(
        get_field,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{ "$getField": { "input": "$foo", "field": "x" } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        get_field_missing,
        expected = Ok(Schema::Missing),
        input = r#"{ "$getField": { "input": "$foo", "field": "z" } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        set_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::Boolean),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$setField": { "input": "$foo", "field": "x", "value": true } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        set_field_root,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
                "x".to_string() => Schema::Atomic(Atomic::Boolean),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{ "$setField": { "input": "$$ROOT", "field": "x", "value": true } }"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_expression_schema!(
        set_field_new_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
                "z".to_string() => Schema::Atomic(Atomic::Boolean),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{ "$setField": { "input": "$foo", "field": "z", "value": true } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        set_field_remove,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("y".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$setField": { "input": "$foo", "field": "x", "value": "$$REMOVE" } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        set_field_remove_non_existing_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$setField": { "input": "$foo", "field": "z", "value": "$$REMOVE" } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        unset_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("y".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$unsetField": { "input": "$foo", "field": "x" } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        unset_field_non_exising_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })),
        input = r#"{ "$unsetField": { "input": "$foo", "field": "z" } }"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::String),
                "y".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("x".to_string(), "y".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        let_simple,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{ "$let": { "vars": {"x": {"$numberDecimal": "1"}}, "in": {"$multiply": ["$$x", "$foo"]}} }"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_expression_schema!(
        let_accesses_existing_variables,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$let": { "vars": {"x": {"$numberDecimal": "1"}}, "in": {"$multiply": ["$$x", "$$y"]}} }"#,
        ref_schema = Schema::Any,
        variables = map! {
            "y".to_string() => Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Null),
            ))
        }
    );

    test_derive_expression_schema!(
        let_overwrites_existing_variables,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{ "$let": { "vars": {"x": {"$numberDecimal": "1"}}, "in": {"$multiply": ["$$x", "$foo"]}} }"#,
        ref_schema = Schema::Atomic(Atomic::Integer),
        variables = map! {
            "x".to_string() => Schema::Atomic(Atomic::Double)
        }
    );
}

mod convert {
    use super::*;
    use agg_ast::definitions::{Convert, LiteralValue, Ref, TaggedOperator};
    use mongosql::schema::Satisfaction;

    macro_rules! test_convert_op {
        ($func_name:ident, expected = $expected:expr, numeric_rep = $numeric_rep:expr, string_rep = $string_rep:expr) => {
            #[test]
            fn $func_name() {
                let mut state = ResultSetState {
                    catalog: &BTreeMap::new(),
                    variables: BTreeMap::new(),
                    result_set_schema: Schema::Document(Document {
                        keys: map! {"foo".to_string() => Schema::Any },
                        ..Default::default()
                    }),
                    current_db: "test".to_string(),
                    null_behavior: Satisfaction::Not,
                };
                let to_values = vec![
                    Expression::Literal(LiteralValue::String($string_rep.to_string())),
                    Expression::Literal(LiteralValue::Int32($numeric_rep)),
                    Expression::Literal(LiteralValue::Int64($numeric_rep.into())),
                    Expression::Literal(LiteralValue::Double($numeric_rep.into())),
                    Expression::Literal(LiteralValue::Decimal128(
                        $numeric_rep.to_string().parse().unwrap(),
                    )),
                ];
                to_values.into_iter().for_each(|v| {
                    let input = Expression::TaggedOperator(TaggedOperator::Convert(Convert {
                        input: Box::new(Expression::Ref(Ref::FieldRef("foo".to_string()))),
                        to: Box::new(v),
                        format: None,
                        on_error: None,
                        on_null: None,
                    }));
                    let result = input.derive_schema(&mut state);
                    assert_eq!(result, $expected);
                });
            }
        };
    }

    test_convert_op!(
        double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        numeric_rep = 1,
        string_rep = "double"
    );

    test_convert_op!(
        string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        numeric_rep = 2,
        string_rep = "string"
    );

    test_convert_op!(
        object,
        expected = Ok(Schema::Document(Document::any())),
        numeric_rep = 3,
        string_rep = "object"
    );

    test_convert_op!(
        array,
        expected = Ok(Schema::Array(Box::new(Schema::Any))),
        numeric_rep = 4,
        string_rep = "array"
    );

    test_convert_op!(
        bin_data,
        expected = Ok(Schema::Atomic(Atomic::BinData)),
        numeric_rep = 5,
        string_rep = "binData"
    );

    test_convert_op!(
        undefined,
        expected = Ok(Schema::Atomic(Atomic::Undefined)),
        numeric_rep = 6,
        string_rep = "undefined"
    );

    test_convert_op!(
        oid,
        expected = Ok(Schema::Atomic(Atomic::ObjectId)),
        numeric_rep = 7,
        string_rep = "objectId"
    );

    test_convert_op!(
        bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        numeric_rep = 8,
        string_rep = "bool"
    );

    test_convert_op!(
        date,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        numeric_rep = 9,
        string_rep = "date"
    );

    test_convert_op!(
        null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        numeric_rep = 10,
        string_rep = "null"
    );

    test_convert_op!(
        regex,
        expected = Ok(Schema::Atomic(Atomic::Regex)),
        numeric_rep = 11,
        string_rep = "regex"
    );

    test_convert_op!(
        db_pointer,
        expected = Ok(Schema::Atomic(Atomic::DbPointer)),
        numeric_rep = 12,
        string_rep = "dbPointer"
    );

    test_convert_op!(
        javascript,
        expected = Ok(Schema::Atomic(Atomic::Javascript)),
        numeric_rep = 13,
        string_rep = "javascript"
    );

    test_convert_op!(
        javascript_with_scope,
        expected = Ok(Schema::Atomic(Atomic::Symbol)),
        numeric_rep = 14,
        string_rep = "symbol"
    );

    test_convert_op!(
        symbol,
        expected = Ok(Schema::Atomic(Atomic::JavascriptWithScope)),
        numeric_rep = 15,
        string_rep = "javascriptWithScope"
    );

    test_convert_op!(
        integer,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        numeric_rep = 16,
        string_rep = "int"
    );

    test_convert_op!(
        timestamp,
        expected = Ok(Schema::Atomic(Atomic::Timestamp)),
        numeric_rep = 17,
        string_rep = "timestamp"
    );

    test_convert_op!(
        long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        numeric_rep = 18,
        string_rep = "long"
    );

    test_convert_op!(
        decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        numeric_rep = 19,
        string_rep = "decimal"
    );

    test_convert_op!(
        min_key,
        expected = Ok(Schema::Atomic(Atomic::MinKey)),
        numeric_rep = -1,
        string_rep = "minKey"
    );

    test_convert_op!(
        max_key,
        expected = Ok(Schema::Atomic(Atomic::MaxKey)),
        numeric_rep = 127,
        string_rep = "maxKey"
    );

    test_derive_expression_schema!(
        convert_non_literal_to_type,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::BinData),
            Schema::Atomic(Atomic::String),
        ))),
        input = r#"{ "$convert": {"input": "$foo", "to": "$bar"} }"#,
        ref_schema = Schema::Any
    );

    test_derive_expression_schema!(
        convert_with_on_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        input = r#"{ "$convert": {"input": "$foo", "to": "int", "onNull": "hi"} }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
        ))
    );

    test_derive_expression_schema!(
        convert_with_on_error,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        input = r#"{ "$convert": {"input": "$foo", "to": "int", "onError": "hi"} }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Javascript),
        ))
    );

    test_derive_expression_schema!(
        convert_fully_specified,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Boolean),
        ))),
        input =
            r#"{ "$convert": {"input": "$foo", "to": "int", "onError": "hi", "onNull": false} }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Javascript),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
        ))
    );
}

mod filter {
    use super::*;

    test_derive_expression_schema!(
        filter_simple,
        expected = Ok(Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "item_id".to_string() => Schema::Atomic(Atomic::Integer),
                "quantity".to_string() => Schema::Atomic(Atomic::Integer),
                "price".to_string() => Schema::Atomic(Atomic::Integer),
                "name".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!(
                "item_id".to_string(),
                "quantity".to_string(),
                "price".to_string(),
                "name".to_string()
            ),
            ..Default::default()
        })))),
        input = r#"{ "$filter": { "input": "$items", "as": "item", "cond": { "$gte": [ "$$item.price", 100 ] }}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::Integer),
                "items".to_string() => Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "item_id".to_string() => Schema::Atomic(Atomic::Integer),
                        "quantity".to_string() => Schema::Atomic(Atomic::Integer),
                        "price".to_string() => Schema::Atomic(Atomic::Integer),
                        "name".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("item_id".to_string(), "quantity".to_string(), "price".to_string(), "name".to_string()),
                    ..Default::default()
                })))
            },
            required: set!("_id".to_string(), "items".to_string()),
            ..Default::default()
        })
    );

    test_derive_expression_schema!(
        filter_maybe_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Array(Box::new(Schema::Document(Document {
                keys: map! {
                    "item_id".to_string() => Schema::Atomic(Atomic::Integer),
                    "quantity".to_string() => Schema::Atomic(Atomic::Integer),
                    "price".to_string() => Schema::Atomic(Atomic::Integer),
                    "name".to_string() => Schema::Atomic(Atomic::String),
                },
                required: set!(
                    "item_id".to_string(),
                    "quantity".to_string(),
                    "price".to_string(),
                    "name".to_string()
                ),
                ..Default::default()
            })))
        ))),
        input = r#"{ "$filter": { "input": "$items", "as": "item", "cond": { "$gte": [ "$$item.price", 100 ] }}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::Integer),
                "items".to_string() => Schema::AnyOf(set!(
                    Schema::Array(Box::new(Schema::Document(Document {
                        keys: map! {
                            "item_id".to_string() => Schema::Atomic(Atomic::Integer),
                            "quantity".to_string() => Schema::Atomic(Atomic::Integer),
                            "price".to_string() => Schema::Atomic(Atomic::Integer),
                            "name".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("item_id".to_string(), "quantity".to_string(), "price".to_string(), "name".to_string()),
                        ..Default::default()
                    }))),
                    Schema::Atomic(Atomic::Null)
                ))
            },
            required: set!("_id".to_string()),
            ..Default::default()
        })
    );
    test_derive_expression_schema!(
        filter_constrains_type,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Array(Box::new(Schema::Document(Document {
                keys: map! {
                    "item_id".to_string() => Schema::Atomic(Atomic::Integer),
                    "quantity".to_string() => Schema::Atomic(Atomic::Integer),
                    "price".to_string() => Schema::Atomic(Atomic::Integer),
                    "name".to_string() => Schema::Atomic(Atomic::String),
                },
                required: set!(
                    "item_id".to_string(),
                    "quantity".to_string(),
                    "price".to_string(),
                    "name".to_string()
                ),
                ..Default::default()
            })))
        ))),
        input = r#"{ "$filter": { "input": "$items", "as": "item", "cond": { "$isNumber": "$$item.item_id" }}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "_id".to_string() => Schema::Atomic(Atomic::Integer),
                "items".to_string() => Schema::AnyOf(set!(
                    Schema::Array(Box::new(Schema::Document(Document {
                        keys: map! {
                            "item_id".to_string() => Schema::AnyOf(set!(
                                Schema::Atomic(Atomic::Integer),
                                Schema::Atomic(Atomic::String),
                            )),
                            "quantity".to_string() => Schema::Atomic(Atomic::Integer),
                            "price".to_string() => Schema::Atomic(Atomic::Integer),
                            "name".to_string() => Schema::Atomic(Atomic::String),
                        },
                        required: set!("item_id".to_string(), "quantity".to_string(), "price".to_string(), "name".to_string()),
                        ..Default::default()
                    }))),
                    Schema::Atomic(Atomic::Null)
                ))
            },
            required: set!("_id".to_string()),
            ..Default::default()
        })
    );
    test_derive_expression_schema!(
        filter_error,
        expected = Err(crate::Error::InvalidExpressionForField(
            "Literal(String(\"foo\"))".to_string(),
            "input"
        )),
        input = r#"{ "$filter": { "input": "foo", "as": "item", "cond": { "$gte": [ "$$item.price", 100 ] }}}"#
    );
}
