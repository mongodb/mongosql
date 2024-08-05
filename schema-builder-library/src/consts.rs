pub(crate) const DISALLOWED_DB_NAMES: [&str; 4] = ["admin", "config", "local", "system"];
pub(crate) const DISALLOWED_COLLECTION_NAMES: [&str; 6] = [
    "system.namespaces",
    "system.indexes",
    "system.profile",
    "system.js",
    "system.views",
    "__sql_schemas",
];

pub(crate) const PARTITION_SIZE_IN_BYTES: i64 = 100 * 1024 * 1024; // 100 MB
pub(crate) const PARTITION_DOCS_PER_ITERATION: i64 = 20;
pub(crate) const VIEW_SAMPLE_SIZE: i64 = 1000;
