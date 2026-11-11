use mongosql::schema::Schema;

// DataService trait and implementations
pub mod data_service;
pub use data_service::{
    CollectionInfo, CollectionOptions, CollectionType, DataService, LocalDataService,
    TimeSeriesOptions,
};

#[cfg(all(feature = "native-client", feature = "wasm"))]
compile_error!("`native-client` and `wasm` features are mutually exclusive");

#[cfg(feature = "native-client")]
pub use data_service::MongoDbDataService;

#[cfg(feature = "wasm")]
pub use data_service::{JsDataService, WasmDataService};

#[cfg(feature = "native-client")]
pub mod client_util;

mod partitioning;
pub use partitioning::{PartitionedCollection, get_partitions};
mod errors;
mod schema;
pub use errors::Error;
pub use schema::{
    SinglePartition, derive_schema_for_collection, derive_schema_for_partition,
    derive_schema_for_view,
};

/// Re-export of mongosql Schema type for convenience
pub type MongoSqlSchema = Schema;

#[cfg(feature = "integration")]
#[cfg(test)]
mod internal_integration_tests;
