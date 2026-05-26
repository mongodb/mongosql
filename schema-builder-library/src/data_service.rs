use bson::Document;
use futures::Stream;
use serde::{Deserialize, Serialize};

#[cfg(test)]
mod test;

#[cfg(feature = "native-client")]
pub mod mongodb;
#[cfg(feature = "native-client")]
pub use mongodb::MongoDbDataService;

#[cfg(feature = "wasm")]
pub mod wasm;
#[cfg(feature = "wasm")]
pub use wasm::{JsDataService, WasmDataService};

/// Options specific to a database collection
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CollectionInfo {
    /// The name of the collection.
    pub name: String,
}

/// Information for a database view
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ViewInfo {
    /// The name of the view.
    pub name: String,

    /// The options specific to this view
    pub options: ViewOptions,
}

/// Information for a database timeseries collection
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TimeseriesInfo {
    /// The name of the timeseries collection.
    pub name: String,

    /// The options for a time series collection
    pub options: TimeseriesOptions,
}

/// Options for a database timeseries collection
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TimeseriesOptions {
    /// The options set for this timeseries collection
    pub timeseries: TimeSeriesOptions,
}

/// Options specific to a database view
#[derive(Debug, Clone, Default, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase", default)]
pub struct ViewOptions {
    /// The name of the source collection.
    pub view_on: String,

    /// The aggregation pipeline.
    pub pipeline: Vec<Document>,
}

/// Information about a single collection entry, returned by `list_collections`.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase", tag = "type")]
pub enum Collection {
    Collection(CollectionInfo),
    View(ViewInfo),
    Timeseries(TimeseriesInfo),
}

impl Collection {
    pub fn name(&self) -> &str {
        match self {
            Collection::Collection(CollectionInfo { name })
            | Collection::View(ViewInfo { name, .. })
            | Collection::Timeseries(TimeseriesInfo { name, .. }) => name,
        }
    }
}

/// Options for timeseries collections.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TimeSeriesOptions {
    pub time_field: String,
    pub meta_field: Option<String>,
}

/// Options for aggregate queries
#[derive(Debug, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
#[non_exhaustive]
pub struct AggregateOptions {
    /// A key hint for indexing
    pub key_hint: Option<Document>,
}

/// Abstraction over database operations used by the schema builder.
#[trait_variant::make(DataService: Send)]
pub trait LocalDataService {
    /// The error type returned by this service's operations.
    type Error: core::error::Error;

    /// List all database names.
    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error>;

    /// List all collections in a database.
    async fn list_collections(&self, db_name: &str) -> Result<Vec<Collection>, Self::Error>;

    /// Execute an aggregation pipeline on a collection.
    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
        options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error>;

    /// Execute a find query on a collection.
    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error>;
}
