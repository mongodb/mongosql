use lazy_static::lazy_static;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};

pub const SMALL_COLL_SIZE_IN_MB: i64 = 90;
pub const SMALL_ID_MIN: i64 = 12345;
pub const LARGE_COLL_SIZE_IN_MB: i64 = 350;
pub const LARGE_ID_MIN: i64 = 0;

pub const DATA_DOC_SIZE_IN_BYTES: i64 = 400;

const PARTITION_SIZE_IN_MB: i64 = 100;

lazy_static! {
    // Partitions are divided equally. They are 100MB at most.
    pub static ref NUM_DOCS_PER_LARGE_PARTITION: i64 =
        (LARGE_COLL_SIZE_IN_MB * 1024 * 1024) /* large collection size in bytes */
        / (LARGE_COLL_SIZE_IN_MB as f64 / PARTITION_SIZE_IN_MB as f64).ceil() as i64 /* number of partitions in large collection */
        / DATA_DOC_SIZE_IN_BYTES;

    pub static ref NONUNIFORM_LARGE_SCHEMA: Schema = Schema::Document(Document {
        keys: map!(
            "_id".to_string() => Schema::Atomic(Atomic::Long),
            "padding".to_string() => Schema::Atomic(Atomic::String),
            "second".to_string() => Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
            )),
            "third".to_string() => Schema::Atomic(Atomic::ObjectId),
            "var".to_string() => Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Null),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            )),
        ),
        required: set! {"_id".to_string(), "padding".to_string(), "var".to_string()},
        additional_properties: false,
        jaccard_index: None,
    });

    pub static ref NONUNIFORM_SMALL_SCHEMA:Schema = Schema::Document(Document {
        keys: map!(
            "_id".to_string() => Schema::Atomic(Atomic::Long),
            "padding".to_string() => Schema::Atomic(Atomic::String),
            "second".to_string() => Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
            )),
            "var".to_string() => Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Null),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            )),
        ),
        required: set! {"_id".to_string(), "padding".to_string(), "var".to_string()},
        additional_properties: false,
        jaccard_index: None,
    });

    pub static ref NONUNIFORM_VIEW_SCHEMA: Schema = Schema::Document(Document {
        keys: map! {
            "_id".to_string() => Schema::Atomic(Atomic::Long),
            "var".to_string() => Schema::AnyOf(set! {
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            }),
            "second".to_string() => Schema::Atomic(Atomic::Long),
            "third".to_string() => Schema::AnyOf(set! {
                Schema::Atomic(Atomic::ObjectId),
                Schema::Atomic(Atomic::String),
            }),
            "padding_len".to_string() => Schema::Atomic(Atomic::Integer),
        },
        required: set! {"_id".to_string(), "var".to_string(), "second".to_string(), "third".to_string(), "padding_len".to_string()},
        additional_properties: false,
        jaccard_index: None,
    });

    // both large and small collections have the same schema in the uniform db
    pub static ref UNIFORM_COLL_SCHEMA: Schema = Schema::Document(Document {
        keys: map!(
            "_id".to_string() => Schema::Atomic(Atomic::Long),
            "array_field".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            "date_field".to_string() => Schema::Atomic(Atomic::Date),
            "document_field".to_string() => Schema::Document(Document {
                keys: map!(
                    "sub_bool_field".to_string() => Schema::Atomic(Atomic::Boolean),
                    "sub_decimal_field".to_string() => Schema::Atomic(Atomic::Decimal),
                    "sub_document_field".to_string() => Schema::Document(Document {
                        keys: map!(
                            "sub_sub_int_field".to_string() => Schema::Atomic(Atomic::Integer),
                        ),
                        required: set! {"sub_sub_int_field".to_string()},
                        additional_properties: false,
                        jaccard_index: None,
                    }),
                ),
                required: set! {"sub_bool_field".to_string(), "sub_decimal_field".to_string(), "sub_document_field".to_string()},
                additional_properties: false,
                jaccard_index: None,
            }),
            "double_field".to_string() => Schema::Atomic(Atomic::Double),
            "long_field".to_string() => Schema::Atomic(Atomic::Long),
            "oid_field".to_string() => Schema::Atomic(Atomic::ObjectId),
            "string_field".to_string() => Schema::Atomic(Atomic::String),
            "uuid_field".to_string() => Schema::Atomic(Atomic::BinData),
        ),
        required: set! {"_id".to_string(), "array_field".to_string(), "date_field".to_string(), "document_field".to_string(), "double_field".to_string(), "long_field".to_string(), "oid_field".to_string(), "string_field".to_string(), "uuid_field".to_string()},
        additional_properties: false,
        jaccard_index: None,
    });

    pub static ref UNIFORM_VIEW_SCHEMA: Schema = Schema::Document(Document {
        keys: map! {
            "_id".to_string() => Schema::Atomic(Atomic::Long),
            "array_field".to_string() => Schema::Atomic(Atomic::Integer),
            "idx".to_string() => Schema::Atomic(Atomic::Long),
        },
        required: set! {"_id".to_string(), "array_field".to_string(), "idx".to_string()},
        additional_properties: false,
        jaccard_index: None,
    });
}

pub const UNIFORM_DB_NAME: &str = "uniform";
pub const NONUNIFORM_DB_NAME: &str = "nonuniform";
pub const SMALL_COLL_NAME: &str = "small";
pub const LARGE_COLL_NAME: &str = "large";
pub const UNITARY_COLL_NAME: &str = "unit";
pub const VIEW_NAME: &str = "test_view";
