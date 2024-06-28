pub(crate) const DISALLOWED_DB_NAMES: [&str; 4] = ["admin", "config", "local", "system"];
pub(crate) const DISALLOWED_COLLECTION_NAMES: [&str; 5] = [
    "system.namespaces",
    "system.indexes",
    "system.profile",
    "system.js",
    "system.views",
];

pub(crate) const PARTITION_SIZE_IN_BYTES: i64 = 100 * 1024 * 1024; // 100 MB
pub(crate) const SAMPLE_MIN_DOCS: i64 = 101;
pub(crate) const SAMPLE_RATE: f64 = 0.04;
pub(crate) const MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION: i64 = 20;
pub(crate) const SAMPLE_SIZE: i64 = 1000;
