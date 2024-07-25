use lazy_static::lazy_static;

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

    // Using the same math above, we know that
    // 90MB / 400B = 235
    pub static ref NUM_DOCS_IN_SMALL_COLLECTION: i64 =
        (SMALL_COLL_SIZE_IN_MB * 1024 * 1024) /* small collection size in bytes */
        / DATA_DOC_SIZE_IN_BYTES;
}

pub const UNIFORM_DB_NAME: &str = "uniform";
pub const NONUNIFORM_DB_NAME: &str = "nonuniform";
pub const SMALL_COLL_NAME: &str = "small";
pub const LARGE_COLL_NAME: &str = "large";
pub const VIEW_NAME: &str = "test_view";
