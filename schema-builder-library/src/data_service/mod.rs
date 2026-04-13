use bson::Document;
use serde::{Deserialize, Serialize};

#[cfg(test)]
mod test;

use crate::Result;

#[cfg(feature = "native-client")]
pub mod mongodb;
#[cfg(feature = "native-client")]
pub use mongodb::MongoDbDataService;

#[cfg(feature = "wasm")]
pub mod wasm;
#[cfg(feature = "wasm")]
pub use wasm::{JsDataService, WasmDataService};

/// The type of a MongoDB collection entry.
#[derive(Debug, Clone, Default, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub enum CollectionType {
    #[default]
    Collection,
    View,
    Timeseries,
}

/// Information about a single collection entry, returned by `list_collections`.
#[derive(Debug, Clone, Default, PartialEq, Serialize, Deserialize)]
pub struct CollectionInfo {
    /// The name of the collection.
    pub name: String,
    /// The type of the collection.
    #[serde(rename = "type")]
    pub collection_type: CollectionType,
    /// Additional options (primarily used for views and timeseries).
    #[serde(default)]
    pub options: CollectionOptions,
}

/// Options for collections. View options and timeseries options are both represented here
/// since the `options` field in `listCollections` output is overloaded for different collection types.
/// Consumers should check `CollectionInfo::collection_type` before accessing the relevant fields.
#[derive(Debug, Clone, Default, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CollectionOptions {
    /// For views, the name of the source collection.
    #[serde(default)]
    pub view_on: String,
    /// For views, the aggregation pipeline.
    #[serde(default)]
    pub pipeline: Vec<Document>,
    /// For timeseries collections, the timeseries options.
    #[serde(default)]
    pub timeseries: Option<TimeSeriesOptions>,
}

/// Options for timeseries collections.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TimeSeriesOptions {
    pub time_field: String,
    pub meta_field: Option<String>,
}

/// Abstraction over database operations used by the schema builder.
///
/// On non-WASM targets `Send + Sync` is required; on WASM it is not.
#[cfg_attr(not(feature = "wasm"), async_trait::async_trait)]
#[cfg_attr(feature = "wasm", async_trait::async_trait(?Send))]
pub trait DataService {
    /// List all database names.
    async fn list_databases(&self) -> Result<Vec<String>>;

    /// List all collections in a database.
    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>>;

    /// Execute an aggregation pipeline on a collection.
    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
    ) -> Result<Vec<Document>>;

    /// Execute a find query on a collection.
    async fn find(&self, db_name: &str, coll_name: &str, filter: Document)
    -> Result<Vec<Document>>;
}
