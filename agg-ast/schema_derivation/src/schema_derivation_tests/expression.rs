use crate::{
    schema_derivation::{DeriveSchema, ResultSetState},
    Error,
};
use agg_ast::definitions::Expression;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};
use std::collections::BTreeMap;

mod literal {
    use super::*;

    test_derive_schema!(
        literal_binary,
        expected = Ok(Schema::Atomic(Atomic::BinData)),
        input = r#"{"$binary" : {"base64" : "", "subType" : "04"}}"#
    );

    test_derive_schema!(
        literal_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = r#"true"#
    );

    test_derive_schema!(
        literal_date,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$date": { "$numberLong": "1655956513000" } }"#
    );

    test_derive_schema!(
        literal_dbpointer,
        expected = Ok(Schema::Atomic(Atomic::DbPointer)),
        input =
            r#"{ "$dbPointer": { "$ref": "foo", "$id": { "$oid": "57e193d7a9cc81b4027498b5" } } }"#
    );

    test_derive_schema!(
        literal_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{ "$numberDecimal": "3.0" }"#
    );

    test_derive_schema!(
        literal_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{ "$numberDouble": "3.0" }"#
    );

    test_derive_schema!(
        literal_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{ "$numberInt": "3" }"#
    );

    test_derive_schema!(
        literal_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = r#"{ "$numberLong": "3" }"#
    );

    test_derive_schema!(
        literal_javascript,
        expected = Ok(Schema::Atomic(Atomic::Javascript)),
        input = r#"{ "$code": "function() {}" }"#
    );

    test_derive_schema!(
        literal_javascript_with_scope,
        expected = Ok(Schema::Atomic(Atomic::JavascriptWithScope)),
        input = r#"{ "$code": "function() {}", "$scope": { } }"#
    );

    test_derive_schema!(
        literal_maxkey,
        expected = Ok(Schema::Atomic(Atomic::MaxKey)),
        input = r#"{ "$maxKey": 1 }"#
    );

    test_derive_schema!(
        literal_minkey,
        expected = Ok(Schema::Atomic(Atomic::MinKey)),
        input = r#"{ "$minKey": 1 }"#
    );

    test_derive_schema!(
        literal_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"null"#
    );

    test_derive_schema!(
        literal_oid,
        expected = Ok(Schema::Atomic(Atomic::ObjectId)),
        input = r#"{"$oid": "5d505646cf6d4fe581014ab2"}"#
    );

    test_derive_schema!(
        literal_regex,
        expected = Ok(Schema::Atomic(Atomic::Regex)),
        input = r#" { "$regularExpression": { "pattern": "abc*", "options": "ix" } }"#
    );

    test_derive_schema!(
        literal_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#""foo bar""#
    );

    test_derive_schema!(
        literal_symbol,
        expected = Ok(Schema::Atomic(Atomic::Symbol)),
        input = r#"{ "$symbol": "sym2" }"#
    );

    test_derive_schema!(
        literal_timestamp,
        expected = Ok(Schema::Atomic(Atomic::Timestamp)),
        input = r#"{ "$timestamp": { "t": 42, "i": 1 } }"#
    );
}
mod field_ref {
    use super::*;

    test_derive_schema!(
        variable_ref_present,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$$foo""#,
        ref_schema = Schema::Any,
        variables = map! {
            "foo".to_string() => Schema::Atomic(Atomic::Double)
        }
    );

    test_derive_schema!(
        variable_ref_missing,
        expected = Err(Error::UnknownReference("foo".to_string())),
        input = r#""$$foo""#,
        ref_schema = Schema::Any,
        variables = map!()
    );

    test_derive_schema!(
        field_ref,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$foo""#,
        ref_schema = Schema::Atomic(Atomic::Double),
        variables = map!()
    );

    test_derive_schema!(
        nested_field_ref,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#""$foo.bar""#,
        ref_schema = Schema::Document(Document {
            keys: map! {
                "bar".to_string() => Schema::Atomic(Atomic::Double)
            },
            ..Default::default()
        }),
        variables = map!()
    );

    test_derive_schema!(
        field_ref_missing,
        expected = Ok(Schema::Missing),
        input = r#""$foo""#
    );
}

mod array {
    use super::*;

    test_derive_schema!(
        empty_array,
        expected = Ok(Schema::Array(Box::new(Schema::Unsat))),
        input = r#"[]"#
    );

    test_derive_schema!(
        array_single_type,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input = r#"[1, 2, 3]"#
    );

    test_derive_schema!(
        array_multiple_types,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String)
        ))))),
        input = r#"[1, 2, 3, "foo", "bar"]"#
    );

    test_derive_schema!(
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

    test_derive_schema!(
        empty_document,
        expected = Ok(Schema::Document(Document::default())),
        input = r#"{}"#
    );

    test_derive_schema!(
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

    test_derive_schema!(
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
