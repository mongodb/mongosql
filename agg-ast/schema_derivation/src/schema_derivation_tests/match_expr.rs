use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{
        Atomic, Document, Satisfaction, Schema, DATE_COERCIBLE, DATE_COERCIBLE_OR_NULL, NUMERIC,
        NUMERIC_OR_NULL, STRING_OR_NULL,
    },
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
        eq_numeric_recursive,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => NUMERIC.clone(),
            },
            required: set! {"foo".to_string()},
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$abs": "$foo"}, 3]}}}"#,
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

mod date_ops {
    use super::*;

    macro_rules! test_derive_schema_for_date_expression_match {
        ($name_date:ident, $name_timezone:ident, $name_timezone_recursive:ident, $name_date_field_eq_null:ident, $name_timezone_field_eq_null:ident, $name_date_field_must_be_null:ident, $op:expr ) => {
            test_derive_schema_for_match_stage!(
                $name_date,
                expected = Ok(Schema::Document(Document {
                    keys: map! {
                        "foo".to_string() => DATE_COERCIBLE.clone(),
                    },
                    required: set!("foo".to_string()),
                    ..Default::default()
                })),
                input = format!(r#"{{"$match": {{"$expr": {{"{}": {{"date": "$foo", "timezone": "GMT"}}}}}}}}"#, $op).as_str(),
                ref_schema = Schema::Any
            );
            test_derive_schema_for_match_stage!(
                $name_timezone,
                expected = Ok(Schema::Document(Document {
                    keys: map! {
                        "foo".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("foo".to_string()),
                    ..Default::default()
                })),
                input = format!(r#"{{"$match": {{"$expr": {{"{}": {{"date": "2019-01-01", "timezone": "$foo"}}}}}}}}"#, $op).as_str(),
                ref_schema = Schema::Any
            );
            test_derive_schema_for_match_stage!(
                $name_timezone_recursive,
                expected = Ok(Schema::Document(Document {
                    keys: map! {
                        "foo".to_string() => Schema::Atomic(Atomic::String),
                    },
                    required: set!("foo".to_string()),
                    ..Default::default()
                })),
                input = format!(r#"{{"$match": {{"$expr": {{"{}": {{"date": "2019-01-01", "timezone": {{"$concat": ["$foo", "T"]}}}}}}}}}}"#, $op).as_str(),
                ref_schema = Schema::Any
            );
            test_derive_schema_for_match_stage!(
                $name_date_field_eq_null,
                expected = Ok(Schema::Document(Document {
                    keys: map! {
                        // Really, foo MUST be NULL, but only because timezone is NOT null. I think this is
                        // good enough in general because this isn't _wrong_, just not perfectly precise.
                        // If we go futher, we'll have a combinatorial explosion of possibilities,
                        // especially for operators with even more arguments.
                        "foo".to_string() => DATE_COERCIBLE_OR_NULL.clone(),
                    },
                    required: set!(),
                    ..Default::default()
                })),
                input = format!(r#"{{"$match": {{"$expr": {{"$eq": [null, {{"{}": {{"date": "$foo", "timezone": "GMT"}}}}]}}}}}}"#, $op).as_str(),
                ref_schema = Schema::Any
            );
            test_derive_schema_for_match_stage!(
                $name_timezone_field_eq_null,
                expected = Ok(Schema::Document(Document {
                    keys: map! {
                        "foo".to_string() => Schema::AnyOf(set!(
                            Schema::Atomic(Atomic::String),
                            Schema::Atomic(Atomic::Null),
                        )),
                    },
                    required: set!(),
                    ..Default::default()
                })),
                input = format!(r#"{{"$match": {{"$expr": {{"$eq": [null, {{"{}": {{"date": "2019-01-01", "timezone": "$foo"}}}}]}}}}}}"#, $op).as_str(),
                ref_schema = Schema::Any
            );
            test_derive_schema_for_match_stage!(
                $name_date_field_must_be_null,
                expected = Ok(Schema::Document(Document {
                    keys: map! {
                        "foo".to_string() => Schema::Atomic(Atomic::Null),
                    },
                    required: set!(),
                    ..Default::default()
                })),
                input = format!(r#"{{"$match": {{"$expr": {{"$eq": [null, {{"{}": {{"date": "$foo"}}}}]}}}}}}"#, $op).as_str(),
                ref_schema = Schema::Any
            );
        };
    }

    test_derive_schema_for_date_expression_match!(
        year_date_field,
        year_timezone_field,
        year_timezone_field_recursive,
        year_date_field_eq_null,
        year_timezone_field_eq_null,
        year_date_field_must_be_null,
        "$year"
    );
    test_derive_schema_for_date_expression_match!(
        month_date_field,
        month_timezone_field,
        month_timezone_field_recursive,
        month_date_field_eq_null,
        month_timezone_field_eq_null,
        month_date_field_must_be_null,
        "$month"
    );
    test_derive_schema_for_date_expression_match!(
        iso_week_year_date_field,
        iso_week_year_timezone_field,
        iso_week_year_timezone_field_recursive,
        iso_week_year_date_field_eq_null,
        iso_week_year_timezone_field_eq_null,
        iso_week_year_date_field_must_be_null,
        "$isoWeekYear"
    );
    test_derive_schema_for_date_expression_match!(
        week_date_field,
        week_timezone_field,
        week_timezone_field_recursive,
        week_date_field_eq_null,
        week_timezone_field_eq_null,
        week_date_field_must_be_null,
        "$week"
    );
    test_derive_schema_for_date_expression_match!(
        iso_week_date_field,
        iso_week_timezone_field,
        iso_week_timezone_field_recursive,
        iso_week_date_field_eq_null,
        iso_week_timezone_field_eq_null,
        iso_week_date_field_must_be_null,
        "$isoWeek"
    );
    test_derive_schema_for_date_expression_match!(
        day_of_week_date_field,
        day_of_week_timezone_field,
        day_of_week_timezone_field_recursive,
        day_of_week_date_field_eq_null,
        day_of_week_timezone_field_eq_null,
        day_of_week_date_field_must_be_null,
        "$dayOfWeek"
    );
    test_derive_schema_for_date_expression_match!(
        iso_day_of_week_date_field,
        iso_day_of_week_timezone_field,
        iso_day_of_week_timezone_field_recursive,
        iso_day_of_week_date_field_eq_null,
        iso_day_of_week_timezone_field_eq_null,
        iso_day_of_week_date_field_must_be_null,
        "$isoDayOfWeek"
    );
    test_derive_schema_for_date_expression_match!(
        day_of_year_date_field,
        day_of_year_timezone_field,
        day_of_year_timezone_field_recursive,
        day_of_year_date_field_eq_null,
        day_of_year_timezone_field_eq_null,
        day_of_year_date_field_must_be_null,
        "$dayOfYear"
    );
    test_derive_schema_for_date_expression_match!(
        day_of_month_date_field,
        day_of_month_timezone_field,
        day_of_month_timezone_field_recursive,
        day_of_month_date_field_eq_null,
        day_of_month_timezone_field_eq_null,
        day_of_month_date_field_must_be_null,
        "$dayOfMonth"
    );
    test_derive_schema_for_date_expression_match!(
        hour_date_field,
        hour_timezone_field,
        hour_timezone_field_recursive,
        hour_date_field_eq_null,
        hour_timezone_field_eq_null,
        hour_date_field_must_be_null,
        "$hour"
    );
    test_derive_schema_for_date_expression_match!(
        minute_date_field,
        minute_timezone_field,
        minute_timezone_field_recursive,
        minute_date_field_eq_null,
        minute_timezone_field_eq_null,
        minute_date_field_must_be_null,
        "$minute"
    );
    test_derive_schema_for_date_expression_match!(
        second_date_field,
        second_timezone_field,
        second_timezone_field_recursive,
        second_date_field_eq_null,
        second_timezone_field_eq_null,
        second_date_field_must_be_null,
        "$second"
    );
    test_derive_schema_for_date_expression_match!(
        millisecond_date_field,
        millisecond_timezone_field,
        millisecond_timezone_field_recursive,
        millisecond_date_field_eq_null,
        millisecond_timezone_field_eq_null,
        millisecond_date_field_must_be_null,
        "$millisecond"
    );

    test_derive_schema_for_match_stage!(
        date_to_string_date_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => DATE_COERCIBLE.clone(),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input =
            r#"{"$match": {"$expr": {"$dateToString": {"format": "%Y-%m-%d", "date": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_to_string_date_field_eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => DATE_COERCIBLE_OR_NULL.clone(),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {"$dateToString": {"format": "%Y-%m-%d", "date": "$foo"}}]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_to_string_date_field_on_null_is_non_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => DATE_COERCIBLE_OR_NULL.clone(),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateToString": {"format": "%Y-%m-%d", "date": "$foo", "onNull": 42}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_to_string_format_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Atomic(Atomic::Date),
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("d".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateToString": {"format": "$foo", "date": "$d"}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Atomic(Atomic::Date),
                "foo".to_string() => Schema::Any,
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_to_string_format_field_recursive,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Atomic(Atomic::Date),
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("d".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateToString": {"format": {"$concat": ["$foo", "T"]}, "date": "$d"}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Atomic(Atomic::Date),
                "foo".to_string() => Schema::Any,
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_from_string_date_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateFromString": {"dateString": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_from_string_date_field_eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {"$dateFromString": {"dateString": "$foo"}}]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_from_string_date_field_on_null_is_non_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => STRING_OR_NULL.clone(),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateFromString": {"dateString": "$foo", "onNull": "2019-01-01"}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_from_string_format_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Atomic(Atomic::String),
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("d".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input =
            r#"{"$match": {"$expr": {"$dateFromString": {"dateString": "$d", "format": "$foo"}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Any,
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_from_string_format_field_recursive,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Atomic(Atomic::String),
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("d".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateFromString": {"dateString": "$d", "format": {"$concat": ["$foo", "T"]}}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "d".to_string() => Schema::Any,
                "foo".to_string() => Schema::Atomic(Atomic::String),
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_to_parts_date_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => DATE_COERCIBLE.clone(),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateToParts": {"date": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_to_parts_timezone_field,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() =>  Schema::Atomic(Atomic::Date),
                "tz".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string(), "tz".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateToParts": {"date": "$foo", "timezone": "$tz"}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Date),
                "tz".to_string() => Schema::Any,
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_to_parts_timezone_field_recursive,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() =>  Schema::Atomic(Atomic::Date),
                "tz".to_string() => Schema::Atomic(Atomic::String),
            },
            required: set!("foo".to_string(), "tz".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$dateToParts": {"date": "$foo", "timezone": {"$concat": ["$tz", "T"]}}}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Date),
                "tz".to_string() => Schema::Any,
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_to_parts_date_field_eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {"$dateToParts": {"date": "$foo"}}]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_to_parts_timezone_and_date_field_eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => DATE_COERCIBLE_OR_NULL.clone(),
                "tz".to_string() => STRING_OR_NULL.clone(),
            },
            required: set!(),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {"$dateToParts": {"date": "$foo", "timezone": "$tz"}}]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Any,
                "tz".to_string() => Schema::Any,
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_from_parts_all_non_iso_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "year".to_string() => NUMERIC.clone(),
                        "month".to_string() => NUMERIC.clone(),
                        "day".to_string() => NUMERIC.clone(),
                        "hour".to_string() => NUMERIC.clone(),
                        "minute".to_string() => NUMERIC.clone(),
                        "second".to_string() => NUMERIC.clone(),
                        "millisecond".to_string() => NUMERIC.clone(),
                    },
                    required: set!(),
                    additional_properties: true,
                    ..Default::default()
                }),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {
            "$dateFromParts": {
                "year": "$foo.year", 
                "month": "$foo.month", 
                "day": "$foo.day", 
                "hour": "$foo.hour", 
                "minute": "$foo.minute", 
                "second": "$foo.second", 
                "millisecond": "$foo.millisecond"
            }}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        date_from_parts_all_iso_fields,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "isoWeekYear".to_string() => NUMERIC.clone(),
                        "isoWeek".to_string() => NUMERIC.clone(),
                        "isoDayOfWeek".to_string() => NUMERIC.clone(),
                    },
                    required: set!("isoWeekYear".to_string(),
                                  "isoWeek".to_string(),
                                  "isoDayOfWeek".to_string(),
                    ),
                    ..Default::default()
                }),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {
            "$dateFromParts": {
                "isoWeekYear": "$foo.isoWeekYear", 
                "isoWeek": "$foo.isoWeek", 
                "isoDayOfWeek": "$foo.isoDayOfWeek" 
            }}}}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "isoWeekYear".to_string() => Schema::Any,
                "isoWeek".to_string() => Schema::Any,
                "isoDayOfWeek".to_string() => Schema::Any,
            },
            required: set!(),
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_from_parts_all_non_iso_fields_eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "year".to_string() => NUMERIC_OR_NULL.clone(),
                        "month".to_string() => NUMERIC_OR_NULL.clone(),
                        "day".to_string() => NUMERIC_OR_NULL.clone(),
                        "hour".to_string() => NUMERIC_OR_NULL.clone(),
                        "minute".to_string() => NUMERIC_OR_NULL.clone(),
                        "second".to_string() => NUMERIC_OR_NULL.clone(),
                        "millisecond".to_string() => NUMERIC_OR_NULL.clone(),
                    },
                    required: set!(),
                    ..Default::default()
                }),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {
            "$dateFromParts": {
                "year": "$foo.year", 
                "month": "$foo.month", 
                "day": "$foo.day", 
                "hour": "$foo.hour", 
                "minute": "$foo.minute", 
                "second": "$foo.second", 
                "millisecond": "$foo.millisecond"
            }}]}}}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "year".to_string() => Schema::Any,
                "month".to_string() => Schema::Any,
                "day".to_string() => Schema::Any,
                "hour".to_string() => Schema::Any,
                "minute".to_string() => Schema::Any,
                "second".to_string() => Schema::Any,
                "millisecond".to_string() => Schema::Any,
            },
            required: set!("year".to_string(), "month".to_string(), "day".to_string()),
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        date_from_parts_all_iso_fields_eq_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document {
                    keys: map! {
                        "isoWeekYear".to_string() => NUMERIC_OR_NULL.clone(),
                        "isoWeek".to_string() => NUMERIC_OR_NULL.clone(),
                        "isoDayOfWeek".to_string() => NUMERIC_OR_NULL.clone(),
                    },
                    required: set!(),
                    ..Default::default()
                }),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {
            "$dateFromParts": {
                "isoWeekYear": "$foo.isoWeekYear", 
                "isoWeek": "$foo.isoWeek", 
                "isoDayOfWeek": "$foo.isoDayOfWeek" 
            }}]}}}"#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "isoWeekYear".to_string() => Schema::Any,
                "isoWeek".to_string() => Schema::Any,
                "isoDayOfWeek".to_string() => Schema::Any,
            },
            required: set!(),
            ..Default::default()
        })
    );
}

mod misc_ops {
    use super::*;
    test_derive_schema_for_match_stage!(
        add_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$add": ["$foo", 1, 2, 3]}}}"#,
        ref_schema = Schema::Atomic(Atomic::Integer)
    );
    test_derive_schema_for_match_stage!(
        add_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Date),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$add": ["$foo", 1, 2, 3]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        add_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Date),
                    Schema::Atomic(Atomic::Null),
                )),
            },
            // noted in the mod.rs file for the macro -- foo can be missing here, represented
            // by its absence from the required set.
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$ne": [1, {"$add": ["$foo", 1, 2, 3]}]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        subtract_first_arg_allows_date,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Date),
                )),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$subtract": ["$foo", 3]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        subtract_second_arg_cannot_be_date,
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
        input = r#"{"$match": {"$expr": {"$subtract": [3, "$foo"]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        subtract_both_args_maybe_date,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Date),
                )),
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Date),
                )),
            },
            required: set!("bar".to_string(), "foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$subtract": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::Any,
                "foo".to_string() => Schema::Any,
            },
            ..Default::default()
        })
    );
    test_derive_schema_for_match_stage!(
        object_to_array_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Document(Document::default())
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$objectToArray": "$foo"}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        object_to_array_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null)
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$not": {"$objectToArray": "$foo"}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        object_to_array_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Document(Document::default()),
                    Schema::Atomic(Atomic::Null)
                ))
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$ne": [{}, {"$objectToArray": "$foo"}]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        max_atomic,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Double),
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$max": ["$foo", 1, 2, 3]}}}"#,
        ref_schema = Schema::Atomic(Atomic::Double)
    );
    test_derive_schema_for_match_stage!(
        max_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$max": ["$foo", 1, 2, 3]}}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Missing
        ))
    );
    test_derive_schema_for_match_stage!(
        max_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Null),
                    Schema::Array(Box::new(Schema::Any))
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [null, {"$max": ["$foo", null]}]}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        let_not_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Integer)
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$let": {"vars": {"x": "$foo"}, "in": {"$ne": ["$$x", null]}}}}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Missing
        ))
    );
    test_derive_schema_for_match_stage!(
        let_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Null)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$let": {"vars": {"x": "$foo"}, "in": {"$eq": ["$$x", null]}}}}}"#,
        ref_schema = Schema::Any
    );
    test_derive_schema_for_match_stage!(
        let_maybe_null,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null)
                ))
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$let": {"vars": {"x": "$foo"}, "in": {"$lt": ["$$x", 2]}}}}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Missing
        ))
    );
    test_derive_schema_for_match_stage!(
        let_nested,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null)
                ))
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$let": {"vars": {"x": "$foo"}, "in": {"$let": {"vars": {"y": "$$x"}, "in": {"$lt": ["$$y", 2]}}}}}}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Missing
        ))
    );
    test_derive_schema_for_match_stage!(
        switch,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Atomic(Atomic::Boolean)
            },
            required: set!("foo".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$switch": {"branches": [{"case": "$foo", "then": 1}], "default": 2}}}}"#,
        ref_schema = Schema::Any
    );
}
