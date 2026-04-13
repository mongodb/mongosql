use futures::TryStreamExt;
use mongodb::{Client, bson::Document, results::CollectionType as DriverCollectionType};
use tracing::warn;

use super::{CollectionInfo, CollectionOptions, CollectionType, DataService, TimeSeriesOptions};

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
    type Error = mongodb::error::Error;

    async fn list_databases(&self) -> std::result::Result<Vec<String>, Self::Error> {
        self.client.list_database_names().await
    }

    async fn list_collections(
        &self,
        db_name: &str,
    ) -> std::result::Result<Vec<CollectionInfo>, Self::Error> {
        let db = self.client.database(db_name);
        let cursor = db.list_collections().await?;
        let specs = cursor.try_collect::<Vec<_>>().await?;

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
    ) -> std::result::Result<Vec<Document>, Self::Error> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);
        let cursor = collection.aggregate(pipeline).await?;
        cursor.try_collect().await
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> std::result::Result<Vec<Document>, Self::Error> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);
        let cursor = collection.find(filter).await?;
        cursor.try_collect().await
    }
}
