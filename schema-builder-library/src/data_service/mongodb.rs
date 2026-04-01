//! MongoDB implementation of the DataService trait.
//!
//! This module provides `MongoDbDataService`, which implements the `DataService` trait
//! using the Rust MongoDB driver. This implementation is intended for use by
//! schema-manager and other native Rust consumers.

use futures::TryStreamExt;
use mongodb::{Client, bson::Document, bson::doc};

use crate::{Error, Result};
use super::{CollectionInfo, CollectionOptions, DataService, parse_namespace};

/// MongoDB implementation of the DataService trait.
///
/// Wraps a `mongodb::Client` and delegates all database operations to it.
pub struct MongoDbDataService {
    client: Client,
}

impl MongoDbDataService {
    /// Create a new `MongoDbDataService` with the given MongoDB client.
    pub fn new(client: Client) -> Self {
        Self { client }
    }
}

#[async_trait::async_trait]
impl DataService for MongoDbDataService {
    async fn list_databases(&self) -> Result<Vec<String>> {
        self.client
            .list_database_names()
            .await
            .map_err(Error::from)
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>> {
        let db = self.client.database(db_name);
        let cursor = db
            .run_cursor_command(doc! { "listCollections": 1.0, "authorizedCollections": true })
            .await
            .map_err(Error::from)?;

        let docs: Vec<Document> = cursor.try_collect().await.map_err(Error::from)?;

        let collections = docs
            .into_iter()
            .filter_map(|doc| {
                let name = doc.get_str("name").ok()?.to_string();
                let collection_type = doc.get_str("type").unwrap_or("collection").to_string();
                let options = doc
                    .get_document("options")
                    .ok()
                    .map(|opts| {
                        let view_on = opts.get_str("viewOn").unwrap_or("").to_string();
                        let pipeline = opts
                            .get_array("pipeline")
                            .ok()
                            .map(|arr| {
                                arr.iter()
                                    .filter_map(|v| v.as_document().cloned())
                                    .collect()
                            })
                            .unwrap_or_default();
                        CollectionOptions { view_on, pipeline }
                    })
                    .unwrap_or_default();

                Some(CollectionInfo {
                    name,
                    collection_type,
                    options,
                })
            })
            .collect();

        Ok(collections)
    }

    async fn aggregate(&self, ns: &str, pipeline: Vec<Document>) -> Result<Vec<Document>> {
        let (db_name, coll_name) = parse_namespace(ns);
        let collection = self.client.database(db_name).collection::<Document>(coll_name);

        let cursor = collection.aggregate(pipeline).await.map_err(Error::from)?;
        cursor.try_collect().await.map_err(Error::from)
    }

    async fn find(&self, ns: &str, filter: Document) -> Result<Vec<Document>> {
        let (db_name, coll_name) = parse_namespace(ns);
        let collection = self.client.database(db_name).collection::<Document>(coll_name);

        let cursor = collection.find(filter).await.map_err(Error::from)?;
        cursor.try_collect().await.map_err(Error::from)
    }
}
