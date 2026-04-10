use futures::TryStreamExt;
use mongodb::{Client, bson::Document, results::CollectionType as DriverCollectionType};
use tracing::warn;

use super::{CollectionInfo, CollectionOptions, CollectionType, DataService, TimeSeriesOptions};
use crate::{Error, Result};

/// [`DataService`] implementation backed by the Rust MongoDB driver.
pub struct MongoDbDataService {
    client: Client,
}

impl MongoDbDataService {
    pub fn new(client: Client) -> Self {
        Self { client }
    }
}

#[async_trait::async_trait]
impl DataService for MongoDbDataService {
    async fn list_databases(&self) -> Result<Vec<String>> {
        self.client.list_database_names().await.map_err(Error::from)
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>> {
        let db = self.client.database(db_name);
        let cursor = db.list_collections().await.map_err(Error::from)?;
        let specs = cursor.try_collect::<Vec<_>>().await.map_err(Error::from)?;

        Ok(specs
            .into_iter()
            .filter_map(|spec| {
                let collection_type = match spec.collection_type {
                    DriverCollectionType::Collection => CollectionType::Collection,
                    DriverCollectionType::View => CollectionType::View,
                    DriverCollectionType::Timeseries => CollectionType::Timeseries,
                    _ => {
                        warn!("Skipping collection '{}' with unknown type", spec.name);
                        return None;
                    }
                };
                let options = CollectionOptions {
                    view_on: spec.options.view_on.unwrap_or_default(),
                    pipeline: spec.options.pipeline.unwrap_or_default(),
                    timeseries: spec.options.timeseries.map(|ts| TimeSeriesOptions {
                        time_field: ts.time_field,
                        meta_field: ts.meta_field,
                    }),
                };
                Some(CollectionInfo {
                    name: spec.name,
                    collection_type,
                    options,
                })
            })
            .collect())
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
    ) -> Result<Vec<Document>> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);
        let cursor = collection.aggregate(pipeline).await.map_err(Error::from)?;
        cursor.try_collect().await.map_err(Error::from)
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<Vec<Document>> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);
        let cursor = collection.find(filter).await.map_err(Error::from)?;
        cursor.try_collect().await.map_err(Error::from)
    }
}
