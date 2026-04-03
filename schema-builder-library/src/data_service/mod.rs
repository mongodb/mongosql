//! DataService trait for abstracting database operations.
//!
//! This module defines the `DataService` trait which provides a flat, namespace-based
//! interface for database operations. This design mirrors Compass's DataService pattern
//! and enables the schema builder to work with different backends (Rust MongoDB driver,
//! WASM with JavaScript callbacks, etc.)

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

/// A trait for abstracting database operations.
///
/// This trait provides a flat, namespace-based interface that mirrors Compass's
/// DataService pattern. It uses namespace strings (`"database.collection"`) for
/// collection operations instead of database/collection handles.
///
/// # Design Decisions
///
/// - **Flat interface**: Methods are flat on the trait (no `db.collection()` chaining)
/// - **Namespace strings**: Uses `ns: &str` in `"database.collection"` format for collection operations
/// - **Vec returns**: Returns `Vec<Document>` instead of cursors (simpler, loads all into memory)
///
/// # Example
///
/// ```ignore
/// let databases = data_service.list_databases().await?;
/// let collections = data_service.list_collections("mydb").await?;
/// let docs = data_service.aggregate("mydb.mycoll", vec![doc! {"$limit": 10}]).await?;
/// ```
///
/// # Send requirement
///
/// On non-WASM targets, functions that spawn tasks (e.g. via `tokio::spawn`) require
/// `D: DataService + Send + Sync`. On WASM targets, `Send + Sync` is not required
/// since JavaScript is single-threaded.
#[cfg_attr(not(target_arch = "wasm32"), async_trait::async_trait)]
#[cfg_attr(target_arch = "wasm32", async_trait::async_trait(?Send))]
pub trait DataService {
    /// List all database names.
    async fn list_databases(&self) -> Result<Vec<String>>;

    /// List all collections in a database.
    ///
    /// Returns collection information including name, type, and options (for views).
    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>>;

    /// Execute an aggregation pipeline on a namespace.
    ///
    /// # Arguments
    ///
    /// * `ns` - The namespace in `"database.collection"` format
    /// * `pipeline` - The aggregation pipeline stages
    async fn aggregate(&self, ns: &str, pipeline: Vec<Document>) -> Result<Vec<Document>>;

    /// Execute a find query on a namespace.
    ///
    /// # Arguments
    ///
    /// * `ns` - The namespace in `"database.collection"` format
    /// * `filter` - The query filter document
    async fn find(&self, ns: &str, filter: Document) -> Result<Vec<Document>>;
}

/// Parse a namespace string into database and collection name components.
///
/// Splits on the first `.`, so collection names containing dots are handled correctly.
///
/// # Arguments
///
/// * `ns` - The namespace in `"database.collection"` format
///
/// # Panics
///
/// Panics if the namespace does not contain a `.` separator.
pub fn parse_namespace(ns: &str) -> (&str, &str) {
    let dot_pos = ns
        .find('.')
        .unwrap_or_else(|| panic!("Invalid namespace format: {ns}"));
    (&ns[..dot_pos], &ns[dot_pos + 1..])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_namespace() {
        let (db, coll) = parse_namespace("mydb.mycoll");
        assert_eq!(db, "mydb");
        assert_eq!(coll, "mycoll");
    }

    #[test]
    fn test_parse_namespace_with_dots_in_collection() {
        let (db, coll) = parse_namespace("mydb.my.coll.name");
        assert_eq!(db, "mydb");
        assert_eq!(coll, "my.coll.name");
    }
}
