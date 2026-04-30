use futures::TryStreamExt;
use mongodb::{
    Client, bson::Document, options::Hint, results::CollectionType as DriverCollectionType,
};
use tracing::warn;

use super::{CollectionInfo, CollectionOptions, CollectionType, DataService, TimeSeriesOptions};

impl TryFrom<DriverCollectionType> for CollectionType {
    type Error = DriverCollectionType;

    fn try_from(value: DriverCollectionType) -> std::result::Result<Self, Self::Error> {
        match value {
            DriverCollectionType::Collection => Ok(CollectionType::Collection),
            DriverCollectionType::View => Ok(CollectionType::View),
            DriverCollectionType::Timeseries => Ok(CollectionType::Timeseries),
            other => Err(other),
        }
    }
}

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
    type Cursor = mongodb::Cursor<Document>;
    type Error = mongodb::error::Error;

    async fn list_databases(&self) -> Result<Vec<String>, Self::Error> {
        self.client.list_database_names().await
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        let db = self.client.database(db_name);
        let cursor = db.list_collections().await?;
        let specs = cursor.try_collect::<Vec<_>>().await?;

        Ok(specs
            .into_iter()
            .filter_map(|spec| {
                let collection_type = spec
                    .collection_type
                    .try_into()
                    .inspect_err(|_| warn!("Skipping collection '{}' with unknown type", spec.name))
                    .ok()?;
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
        key_hint: Option<Document>,
    ) -> Result<Self::Cursor, Self::Error> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);

        // Create a native cursor over the specified aggregate, adding a key hint
        // if supplied.
        let mut cursor = collection.aggregate(pipeline);
        if let Some(hint) = key_hint {
            cursor = cursor.hint(Hint::Keys(hint))
        };

        cursor.await
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<Self::Cursor, Self::Error> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);

        collection.find(filter).await
    }
}
