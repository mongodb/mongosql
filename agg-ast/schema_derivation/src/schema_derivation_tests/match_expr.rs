use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema, NUMERIC},
    set,
};
use std::collections::BTreeMap;

mod conjunction_ops {
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
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::MinKey),
                    Schema::Atomic(Atomic::Null),
                )),
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null)
                )),
            },
            required: set!("bar".to_string()),
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$lte": ["$foo", "$bar"]}}}"#,
        starting_schema = Schema::Document(Document {
            keys: map! {
                "foo".to_string() => Schema::Any,
                "bar".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Null),
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
                "foo".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Null),
                )),
            },
            ..Default::default()
        })),
        input = r#"{"$match": {"$expr": {"$eq": [{"$abs": "$foo"}, null]}}}"#,
        ref_schema = Schema::Any
    );
}
