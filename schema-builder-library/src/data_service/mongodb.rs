use bson::doc;
use futures::{Stream, TryStreamExt};
use mongodb::{
    Client, bson::Document, options::Hint, results::CollectionType as DriverCollectionType,
};
use tracing::warn;

use crate::{client_util::DatabaseExt as _, data_service::AggregateOptions};

use super::{CollectionInfo, CollectionType, DataService};

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
    type Error = mongodb::error::Error;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        self.client.list_database_names().await
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        let db = self.client.database(db_name);

        // Note: Here we manually run the `listCollection` method rather than use the
        // driver's `list_collections` method since the latter does not filter authorized
        // collections only. See [1] for more context and [2] for when this was originally
        // changed.
        //
        // [1]: https://github.com/mongodb/mongosql/pull/143#discussion_r3164925954
        // [2]: https://github.com/mongodb/mongosql/pull/136#discussion_r3066410249
        let cursor = db
            .run_cursor_command_with_read_preference(
                doc! { "listCollections": 1.0, "authorizedCollections": true },
            )
            .await?;

        let specs = cursor.try_collect::<Vec<_>>().await?;
        Ok(specs
            .into_iter()
            .filter_map(|doc| {
                // Note: We try to extract the collection name here in case it fails
                // for better error messages.
                let name = doc.get_str("name").unwrap_or("<unknown>").to_string();
                bson::from_document(doc)
                    .map_err(|e| warn!("Skipping malformed listCollections entry '{name}': {e}"))
                    .ok()
            })
            .collect())
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
        options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);

        // Create a native cursor over the specified aggregate, adding a key hint
        // if supplied.
        let mut cursor = collection.aggregate(pipeline);
        if let Some(hint) = options.key_hint {
            cursor = cursor.hint(Hint::Keys(hint))
        };

        cursor.await
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let collection = self
            .client
            .database(db_name)
            .collection::<Document>(coll_name);

        collection.find(filter).await
    }
}
