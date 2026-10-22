//! [`LocalDataService`] implementation backed by the synchronous Rust MongoDB driver.
//!
//! Unlike [`super::MongoDbDataService`] (which implements the `Send` `DataService`
//! variant), this service implements the non-`Send` `LocalDataService` directly and
//! performs all I/O with the blocking `mongodb::sync` driver. This is intended for
//! hosts that cannot provide an async runtime (e.g. a Python extension running on a
//! foreign thread); an async host should prefer [`super::MongoDbDataService`].

use futures::{Stream, stream};
use mongodb::{bson::Document, sync::Client};

use super::{AggregateOptions, CollectionInfo, LocalDataService};

/// [`LocalDataService`] implementation backed by the synchronous Rust MongoDB driver.
#[derive(Debug, Clone)]
pub struct SynchronousMongoDbDataService {
    client: Client,
}

impl SynchronousMongoDbDataService {
    pub fn new(client: Client) -> Self {
        Self { client }
    }
}

impl LocalDataService for SynchronousMongoDbDataService {
    type Error = mongodb::error::Error;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        self.client.list_database_names().run()
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        // Note: Here we manually run the `listCollections` command rather than use the
        // driver's `list_collections` method since the latter does not filter authorized
        // collections only. This mirrors `MongoDbDataService::list_collections`.
        let cursor = self
            .client
            .database(db_name)
            .run_cursor_command(
                bson::doc! { "listCollections": 1.0, "authorizedCollections": true },
            )
            .run()?;
        let specs: Vec<Document> = cursor.collect::<Result<Vec<_>, _>>()?;

        Ok(specs
            .into_iter()
            .filter_map(|doc| {
                // Note: We try to extract the collection name here in case it fails
                // for better error messages.
                let name = doc.get_str("name").unwrap_or("<unknown>").to_string();
                bson::from_document::<CollectionInfo>(doc)
                    .map_err(|e| {
                        tracing::warn!("Skipping malformed listCollections entry '{name}': {e}")
                    })
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
        let mut builder = collection.aggregate(pipeline);
        if let Some(hint) = options.key_hint {
            builder = builder.hint(mongodb::options::Hint::Keys(hint));
        }

        let cursor = builder.run()?;
        Ok(sync_cursor_stream(cursor))
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

        let cursor = collection.find(filter).run()?;
        Ok(sync_cursor_stream(cursor))
    }
}

/// Adapts a blocking `sync::Cursor` into a (non-`Send`) `Stream` of documents.
///
/// Each `poll_next` performs one blocking `Iterator::next()` call on the cursor,
/// which is safe here because `LocalDataService` futures are driven to completion
/// on the calling thread (see the host's `run_current_thread` harness).
fn sync_cursor_stream(
    cursor: mongodb::sync::Cursor<Document>,
) -> impl Stream<Item = Result<Document, mongodb::error::Error>> {
    stream::try_unfold(cursor, |mut cursor| async move {
        match cursor.next() {
            Some(Ok(doc)) => Ok(Some((doc, cursor))),
            // The cursor is dropped on error, which terminates the stream.
            Some(Err(e)) => Err(e),
            None => Ok(None),
        }
    })
}
