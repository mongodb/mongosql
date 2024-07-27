use crate::partitioning::Partition;
use lazy_static::lazy_static;
use mongodb::bson::Bson;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};
use test_utils::schema_builder_library_integration_test_consts::{
    DATA_DOC_SIZE_IN_BYTES, LARGE_COLL_SIZE_IN_MB, LARGE_ID_MIN, NUM_DOCS_PER_LARGE_PARTITION,
    SMALL_COLL_SIZE_IN_MB, SMALL_ID_MIN,
};

lazy_static! {

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
        ..Default::default()
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
        ..Default::default()
    });

    pub static ref NONUNIFORM_LARGE_PARTITION_SCHEMAS: Vec<Schema> = vec![
        Schema::Document(Document {
            keys: map!(
                "_id".to_string() => Schema::Atomic(Atomic::Long),
                "padding".to_string() => Schema::Atomic(Atomic::String),
                "var".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                )),
            ),
            required: set! {"_id".to_string(), "padding".to_string(), "var".to_string()},
            ..Default::default()
        }),
        Schema::Document(Document {
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
            required: set! {"_id".to_string(), "padding".to_string(), "second".to_string(), "var".to_string()},
            ..Default::default()
        }),
        Schema::Document(Document {
            keys: map!(
                "_id".to_string() => Schema::Atomic(Atomic::Long),
                "padding".to_string() => Schema::Atomic(Atomic::String),
                "third".to_string() => Schema::Atomic(Atomic::ObjectId),
                "var".to_string() => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                )),
            ),
            required: set! {"_id".to_string(), "padding".to_string(), "third".to_string(), "var".to_string()},
            ..Default::default()
        }),
        Schema::Document(Document {
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
            required: set! {"_id".to_string(), "padding".to_string(), "second".to_string(), "third".to_string(), "var".to_string()},
            ..Default::default()
        }),
    ];

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
                        ..Default::default()
                    }),
                ),
                required: set! {"sub_bool_field".to_string(), "sub_decimal_field".to_string(), "sub_document_field".to_string()},
                ..Default::default()
            }),
            "double_field".to_string() => Schema::Atomic(Atomic::Double),
            "long_field".to_string() => Schema::Atomic(Atomic::Long),
            "oid_field".to_string() => Schema::Atomic(Atomic::ObjectId),
            "string_field".to_string() => Schema::Atomic(Atomic::String),
            "uuid_field".to_string() => Schema::Atomic(Atomic::BinData),
        ),
        required: set! {"_id".to_string(), "array_field".to_string(), "date_field".to_string(), "document_field".to_string(), "double_field".to_string(), "long_field".to_string(), "oid_field".to_string(), "string_field".to_string(), "uuid_field".to_string()},
        ..Default::default()
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
        ..Default::default()
    });

    pub static ref UNIFORM_VIEW_SCHEMA: Schema = Schema::Document(Document {
        keys: map! {
            "_id".to_string() => Schema::Atomic(Atomic::Long),
            "array_field".to_string() => Schema::Atomic(Atomic::Integer),
            "idx".to_string() => Schema::Atomic(Atomic::Long),
        },
        required: set! {"_id".to_string(), "array_field".to_string(), "idx".to_string()},
        ..Default::default()
    });

    pub static ref LARGE_PARTITIONS: Vec<Partition> = vec![
        Partition {
            min: Bson::Int64(LARGE_ID_MIN),
            max: Bson::Int64(LARGE_ID_MIN + *NUM_DOCS_PER_LARGE_PARTITION),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + *NUM_DOCS_PER_LARGE_PARTITION),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 2)),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 2)),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 3)),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 3)),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 4) - 1),
            is_max_bound_inclusive: true,
        },
    ];

    pub static ref SMALL_PARTITIONS: Vec<Partition> = vec![
        Partition {
            min: Bson::Int64(SMALL_ID_MIN),
            max: Bson::Int64(SMALL_ID_MIN + *NUM_DOCS_IN_SMALL_COLLECTION - 1),
            is_max_bound_inclusive: true,
        },
    ];
    // Using the same math above, we know that
    // 90MB / 400B = 235929
    pub static ref NUM_DOCS_IN_SMALL_COLLECTION: i64 =
        (SMALL_COLL_SIZE_IN_MB * 1024 * 1024) /* small collection size in bytes */
        / DATA_DOC_SIZE_IN_BYTES;

    // Partitions are divided equally. They are 100MB at most.
    pub static ref NUM_DOCS_IN_LARGE_COLLECTION: i64 =
        (LARGE_COLL_SIZE_IN_MB * 1024 * 1024) /* large collection size in bytes */
        / DATA_DOC_SIZE_IN_BYTES;

    // these are computed _after_ the number of docs because if the size of the collection in mb
    // is not perfectly divisible by the doc size, we load slightly fewer bytes in practice
    pub static ref SMALL_COLL_SIZE_IN_BYTES: i64 = *NUM_DOCS_IN_SMALL_COLLECTION * DATA_DOC_SIZE_IN_BYTES;
    pub static ref LARGE_COLL_SIZE_IN_BYTES: i64 = *NUM_DOCS_IN_LARGE_COLLECTION * DATA_DOC_SIZE_IN_BYTES;

}
