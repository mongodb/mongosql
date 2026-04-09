use futures::TryStreamExt;
use mongodb::{Client, bson, bson::Document, bson::doc};

use super::{CollectionInfo, DataService};
use crate::{Error, Result, client_util::DatabaseExt};

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
        let cursor = db
            .run_cursor_command_with_read_preference(
                doc! { "listCollections": 1.0, "authorizedCollections": true },
            )
            .await?;

        let docs: Vec<Document> = cursor.try_collect().await.map_err(Error::from)?;
        docs.into_iter()
            .map(|doc| bson::from_document(doc).map_err(|_| Error::BsonFailure))
            .collect()
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
