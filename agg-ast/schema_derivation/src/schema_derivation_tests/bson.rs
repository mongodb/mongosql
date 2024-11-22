use crate::schema_for_bson;
use bson::{bson, doc, Bson};
use mongosql::{
    map,
    schema::{Atomic, Document, JaccardIndex, Schema},
    set,
};

macro_rules! test_schema_for_bson {
    ($name:ident, input=$input:expr, expected=$expected:expr) => {
        #[test]
        fn $name() {
            let schema = schema_for_bson(&$input);
            assert_eq!(schema, $expected);
        }
    };
}

test_schema_for_bson!(
    simple_document,
    input = bson!({
        "a": 1,
        "b": 1.0,
        "c": "foo"
    }),
    expected = Schema::Document(Document {
        keys: map!(
            "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Double),
            "c".into() => Schema::Atomic(Atomic::String),
        ),
        required: set!("a".into(), "b".into(), "c".into()),
        additional_properties: false,
        jaccard_index: JaccardIndex::default().into(),
    })
);

test_schema_for_bson!(
    empty_doc,
    input = bson!({}),
    expected = Schema::Document(Document {
        keys: map!(),
        required: set!(),
        additional_properties: false,
        jaccard_index: JaccardIndex::default().into(),
    })
);

test_schema_for_bson!(
    nested_document,
    input = bson!({
        "middle": doc! {
            "inner": doc! {
                "foo": "bar"
            }
        }
    }),
    expected = Schema::Document(Document {
        keys: map!(
            "middle".into() => Schema::Document(Document {
                keys: map!(
                    "inner".into() => Schema::Document(Document {
                        keys: map!(
                            "foo".into() => Schema::Atomic(Atomic::String),
                        ),
                        required: set!("foo".into()),
                        additional_properties: false,
                        jaccard_index: JaccardIndex::default().into(),
                    }),
                ),
                required: set!("inner".into()),
                additional_properties: false,
                jaccard_index: JaccardIndex::default().into(),
            })
        ),
        required: set!("middle".into()),
        additional_properties: false,
        jaccard_index: JaccardIndex::default().into(),
    })
);

test_schema_for_bson!(
    empty_array,
    input = bson!([]),
    expected = Schema::Array(Schema::Atomic(Atomic::Null).into())
);

test_schema_for_bson!(
    array_single,
    input = bson!([1.0]),
    expected = Schema::Array(Schema::Atomic(Atomic::Double).into(),)
);

test_schema_for_bson!(
    array_multi_same,
    input = bson!([1.0, 42.0]),
    expected = Schema::Array(Schema::Atomic(Atomic::Double).into(),)
);

test_schema_for_bson!(
    array_multi_differ,
    input = bson!([1.0, null]),
    expected = Schema::Array(
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ])
        .into(),
    )
);

test_schema_for_bson!(
    nested_bson_array,
    input = bson!([[1.0, null], ["foo"]]),
    expected = Schema::Array(
        Schema::Array(
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Null),
                Schema::Atomic(Atomic::String),
            ])
            .into(),
        )
        .into(),
    )
);

test_schema_for_bson!(
    all_atomic_types,
    input = bson!({
        "double": 1.0,
        "string": "foo",
        "boolean": false,
        "null": Bson::Null,
        "undefined": Bson::Undefined,
        "regex": Bson::RegularExpression(bson::Regex {
            pattern: "pattern".to_string(),
            options: "options".to_string()
        }),
        "javascript": Bson::JavaScriptCode("js".to_string()),
        "javascript_with_scope": Bson::JavaScriptCodeWithScope(bson::JavaScriptCodeWithScope {
            code: "js".to_string(),
            scope: doc! {}
        }),
        "int32": Bson::Int32(1),
        "int64": Bson::Int64(1),
        "timestamp": Bson::Timestamp(bson::Timestamp { time: 1, increment: 2 }),
        "binary": Bson::Binary(bson::Binary {
            subtype: bson::spec::BinarySubtype::Uuid,
            bytes: vec![]
        }),
        "objectid": Bson::ObjectId(bson::oid::ObjectId::parse_str("000000000000000000000000").unwrap()),
        "datetime": Bson::DateTime(bson::DateTime::MAX),
        "symbol": Bson::Symbol("test".to_string()),
        "decimal128": Bson::Decimal128(
            bson::Decimal128::from_bytes([0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0])
        ),
        "maxkey": Bson::MaxKey,
        "minkey": Bson::MinKey,
    }),
    expected = Schema::Document(Document {
        keys: map!(
            "binary".into() => Schema::Atomic(Atomic::BinData),
            "boolean".into() => Schema::Atomic(Atomic::Boolean),
            "datetime".into() => Schema::Atomic(Atomic::Date),
            "decimal128".into() => Schema::Atomic(Atomic::Decimal),
            "double".into() => Schema::Atomic(Atomic::Double),
            "int32".into() => Schema::Atomic(Atomic::Integer),
            "int64".into() => Schema::Atomic(Atomic::Long),
            "javascript".into() => Schema::Atomic(Atomic::Javascript),
            "javascript_with_scope".into() => Schema::Atomic(Atomic::JavascriptWithScope),
            "maxkey".into() => Schema::Atomic(Atomic::MaxKey),
            "minkey".into() => Schema::Atomic(Atomic::MinKey),
            "null".into() => Schema::Atomic(Atomic::Null),
            "objectid".into() => Schema::Atomic(Atomic::ObjectId),
            "regex".into() => Schema::Atomic(Atomic::Regex),
            "string".into() => Schema::Atomic(Atomic::String),
            "symbol".into() => Schema::Atomic(Atomic::Symbol),
            "timestamp".into() => Schema::Atomic(Atomic::Timestamp),
            "undefined".into() => Schema::Atomic(Atomic::Undefined),
        ),
        required: set!(
            "binary".into(),
            "boolean".into(),
            "datetime".into(),
            "decimal128".into(),
            "double".into(),
            "int32".into(),
            "int64".into(),
            "javascript".into(),
            "javascript_with_scope".into(),
            "maxkey".into(),
            "minkey".into(),
            "null".into(),
            "objectid".into(),
            "regex".into(),
            "string".into(),
            "symbol".into(),
            "timestamp".into(),
            "undefined".into(),
        ),
        additional_properties: false,
        jaccard_index: JaccardIndex::default().into(),
    })
);
