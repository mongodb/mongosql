use bson::Document;
use serde::{Deserialize, Serialize};

use crate::Result;

#[cfg(not(target_arch = "wasm32"))]
pub mod mongodb;
#[cfg(not(target_arch = "wasm32"))]
pub use mongodb::MongoDbDataService;

#[cfg(target_arch = "wasm32")]
pub mod wasm;
#[cfg(target_arch = "wasm32")]
pub use wasm::{JsDataService, WasmDataService};

/// Information about a single collection entry, returned by `list_collections`.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CollectionInfo {
    /// The name of the collection.
    pub name: String,
    /// The type of the collection ("collection", "view", or "timeseries").
    #[serde(rename = "type")]
    pub collection_type: String,
    /// Additional options (primarily used for views and timeseries).
    #[serde(default)]
    pub options: CollectionOptions,
}

/// Options for collections, primarily used for views.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct CollectionOptions {
    /// For views, the name of the source collection.
    #[serde(rename = "viewOn", default)]
    pub view_on: String,
    /// For views, the aggregation pipeline.
    #[serde(default)]
    pub pipeline: Vec<Document>,
}

/// Abstraction over database operations used by the schema builder.
///
/// Uses namespace strings in `"database.collection"` format for collection operations.
/// On non-WASM targets `Send + Sync` is required; on WASM it is not.
#[cfg_attr(not(target_arch = "wasm32"), async_trait::async_trait)]
#[cfg_attr(target_arch = "wasm32", async_trait::async_trait(?Send))]
pub trait DataService {
    /// List all database names.
    async fn list_databases(&self) -> Result<Vec<String>>;

    /// List all collections in a database.
    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>>;

    /// Execute an aggregation pipeline on a namespace (`"database.collection"` format).
    async fn aggregate(&self, ns: &str, pipeline: Vec<Document>) -> Result<Vec<Document>>;

    /// Execute a find query on a namespace (`"database.collection"` format).
    async fn find(&self, ns: &str, filter: Document) -> Result<Vec<Document>>;
}

/// Splits a `"database.collection"` namespace string into its two components.
///
/// Splits on the first `.`, so collection names containing dots are handled correctly.
pub fn parse_namespace(ns: &str) -> (&str, &str) {
    let dot_pos = ns
        .find('.')
        .unwrap_or_else(|| panic!("Invalid namespace format: {ns}"));
    (&ns[..dot_pos], &ns[dot_pos + 1..])
}
