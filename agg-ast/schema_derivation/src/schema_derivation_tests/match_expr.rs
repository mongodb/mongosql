use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema, NUMERIC},
    set,
};
use std::collections::BTreeMap;

mod logical_ops {
    use super::*;

    test_derive_schema_for_match_stage!(
        or_1_is_always_true,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Any,
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$or": ["$foo", 1]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        or_constrains_schema,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => NUMERIC.clone()
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$or": [{"$abs": "$foo"}]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        or_joins_multiple_constraints,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$or": [{"$abs": "$foo"}, {"$eq": ["$foo", "hello world"]}]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        and_simple,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$and": [{"$abs": "$foo"}]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        and_joins_multiple_constraints_to_fixed_point,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::BinData),
                    Schema::Atomic(Atomic::ObjectId),
                )),
                "b".to_string() => Schema::Atomic(Atomic::ObjectId),
                "c".to_string() => Schema::Atomic(Atomic::ObjectId),
            },
            required: set! {"a".to_string(), "b".to_string(), "c".to_string()},
            ..Default::default()
        })),
        input =
            r#"{"$match": {"$expr": {"$and": [{"$lt": ["$a", "$b"]}, {"$lt": ["$b", "$c"]}]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::BinData),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::Boolean),
                )),
                "b".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::Boolean),
                )),
                "c".to_string() => Schema::Atomic(Atomic::ObjectId),
            },
            required: set! {"a".to_string(), "b".to_string(), "c".to_string()},
            ..Default::default()
        })
    );
}

mod comparison_ops {
    use super::*;

    test_derive_schema_for_match_stage!(
        eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": ["$foo", null]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        eq_numeric,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => NUMERIC.clone(),
            },
            required: set! {"foo".to_string()},
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": ["$foo", 1]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        eq_string,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set! {"foo".to_string()},
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": ["$foo", "hello world"]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        eq_two_refs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null)
                )),
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Any,
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            required: set! {"foo".to_string(), "bar".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        ne_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$ne": ["$foo", null]}}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema_for_match_stage!(
        ne_non_unitary_no_op,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Any,
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$ne": ["$foo", 1]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        ne_two_refs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
                "bar".to_string() => Schema::Atomic(Atomic::MaxKey),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$ne": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::MaxKey),
                )),
                "bar".to_string() => Schema::Atomic(Atomic::MaxKey),
            },
            required: set! {"foo".to_string(), "bar".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        ne_two_refs_unsat,
        expected = Ok(Schema::Document(Document::default())),
        input = r#"{"$match": {"$expr": {"$ne": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::Atomic(Atomic::Null),
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            required: set! {"foo".to_string(), "bar".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        lte_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::MinKey),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lte": ["$foo", null]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        lte_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Symbol),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lte": ["$foo", "hello world"]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        lte_numeric,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lte": ["$foo", 2.0]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        lte_numeric_atomic_reference,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lte": ["$foo", 2.0]}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        lte_two_refs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lte": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Timestamp),
                )),
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            required: set! {"bar".to_string(), "foo".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        lt_two_refs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
                "bar".to_string() => Schema::Atomic(Atomic::Long),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lt": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Timestamp),
                )),
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            required: set! {"bar".to_string(), "foo".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        gte_two_refs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$gte": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Timestamp),
                )),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            required: set! {"bar".to_string(), "foo".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        gt_two_refs,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
                "foo".to_string() => Schema::Atomic(Atomic::Long),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$gt": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Timestamp),
                )),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            required: set! {"bar".to_string(), "foo".to_string()},
            ..Default::default()
        })
    );

    test_derive_schema_for_match_stage!(
        gt_two_refs_one_missing,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
                "foo".to_string() => Schema::Atomic(Atomic::Long),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$gt": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Timestamp),
                )),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::MinKey),
                )),
            },
            required: set! {"bar".to_string()},
            ..Default::default()
        })
    );
}

mod numeric_ops {
    use super::*;

    test_derive_schema_for_match_stage!(
        abs_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => NUMERIC.clone(),
            },
            required: set! {"foo".to_string()},
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$abs": "$foo"}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        abs_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$abs": "$foo"}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        abs_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$not": {"$abs": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        abs_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$abs": "$foo"}, null]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        bit_and_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$bitAnd": "$foo"}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        bit_and_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$bitAnd": "$foo"}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        bit_and_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$not": {"$bitAnd": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        bit_and_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$bitAnd": "$foo"}, null]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        range_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$range": [0, "$foo"]}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        range_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$range": [0, "$foo"]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        is_number_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$isNumber": "$foo"}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        is_number_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$isNumber": "$foo"}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        is_number_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Array(Box::new(Schema::Any)),
                    Schema::Document(Document {additional_properties: true, ..Default::default()}),
                    Schema::Atomic(Atomic::BinData),
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Date),
                    Schema::Atomic(Atomic::DbPointer),
                    Schema::Atomic(Atomic::Javascript),
                    Schema::Atomic(Atomic::JavascriptWithScope),
                    Schema::Atomic(Atomic::MinKey),
                    Schema::Atomic(Atomic::MaxKey),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Regex),
                    Schema::Atomic(Atomic::ObjectId),
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Symbol),
                    Schema::Atomic(Atomic::Timestamp),
                    Schema::Atomic(Atomic::Undefined),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$isNumber": "$foo"}, false]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        round_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$round": "$foo"}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        round_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$round": "$foo"}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        round_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$not": {"$round": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        round_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$round": "$foo"}, null]}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        to_int_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$toInt": "$foo"}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );

    test_derive_schema_for_match_stage!(
        to_int_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Boolean),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$toInt": "$foo"}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        to_int_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Boolean),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$not": {"$toInt": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );

    test_derive_schema_for_match_stage!(
        to_int_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$toInt": "$foo"}, null]}}}"#,
        ref_schema = Schema::Any
    );
}
