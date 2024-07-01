use futures::{future, TryStreamExt};
use mongodb::{
    bson::{self, doc, Bson, Document},
    options::{AggregateOptions, ListDatabasesOptions},
    Collection, Cursor, Database,
};
use mongosql::{
    json_schema,
    schema::{Atomic, JaccardIndex, Schema},
    set,
};
use serde::{Deserialize, Serialize};
use tokio::task::JoinHandle;

pub mod client_util;
mod consts;
use consts::{
    DISALLOWED_COLLECTION_NAMES, DISALLOWED_DB_NAMES, MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION,
    PARTITION_SIZE_IN_BYTES, SAMPLE_MIN_DOCS, SAMPLE_RATE, SAMPLE_SIZE,
};
mod errors;
pub use errors::Error;
mod notifications;
pub use notifications::{SamplerAction, SamplerNotification};

pub mod options;
use options::BuilderOptions;

pub type Result<T> = std::result::Result<T, Error>;

/// A struct representing schema information for a specific namespace (a view
/// or collection).
pub struct SchemaResult {
    /// The name of the database.
    pub db_name: String,

    /// The name of the collection or view which this schema represents.
    pub coll_or_view_name: String,

    /// The type of namespace (collection or view)
    pub namespace_type: NamespaceType,

    /// The schema for the namespace.
    pub namespace_schema: Option<Schema>,
}

/// An enum representing the two namespace types for which this library
/// can generate schema: Collection and View.
pub enum NamespaceType {
    Collection,
    View,
}

macro_rules! notify {
    ($channel:expr, $notification:expr) => {
        if let Some(ref notifier) = $channel {
            // notification errors are not critical, so we just ignore them
            let _ = notifier.send($notification);
        }
    };
}

/// build_schema is the entry point for the schema-builder-library. Given a mongodb::Client,
/// channels for communicating results and status notifications, and various options specifying
/// specific behaviors, this function builds schema for the appropriate namespaces of the provided
/// MongoDB instance. Importantly, this function must be called in a tokio::runtime context since
/// it accesses the current tokio runtime handle.
///
/// The function does not return anything. Instead, as it builds schema for each namespace
/// it sends that information over the tx_schemata channel when a namespace's schema is
/// ready. It also sends status notifications over the tx_notification channel as it takes
/// actions such as Querying, Processing, and Partitioning.
///
/// For collections, this function produces a complete schema representing all data. It does
/// this by using an algorithm based on partitioning the data, repeatedly querying each
/// partition (in parallel) for documents that do not match the in-progress schema, and then
/// unifying the schema for each partition.
///
/// For views, this function produces a best-effort schema based on sampling.
///
/// This function parallelizes handling each database, collection, collection partition, and
/// view by using asynchronous tokio tasks which are spawned using the provided runtime::Handle.
pub async fn build_schema<'a>(options: BuilderOptions<'_>) {
    // To start computing the schema for all databases, we need to wait for the
    // list_database_names method to finish.
    let databases = options
        .client
        .list_database_names(None, Some(ListDatabasesOptions::builder().build()))
        .await;

    // If listing the databases fails, there is nothing to do so we report an
    // error, drop all channels, and return.
    if let Err(e) = databases {
        notify!(
            options.tx_notifications,
            SamplerNotification {
                db: "".to_string(),
                collection_or_view: "".to_string(),
                action: SamplerAction::CriticalError {
                    message: e.to_string(),
                },
            }
        );
        drop(options.tx_notifications);
        drop(options.tx_schemata);
        return;
    }

    // Here, we iterate through each database and spawn a new async task to
    // compute each db schema. Importantly, we are not awaiting the spawned
    // tasks. Each async task will start running in the background immediately,
    // but the program will continue executing the iteration since tokio::spawn
    // immediately returns a JoinHandle.
    let db_tasks = databases
        .unwrap()
        .into_iter()
        .filter(|db_name| !DISALLOWED_DB_NAMES.contains(&db_name.as_str()))
        .map(|db_name| {
            // To avoid passing a reference to the mongodb::Client around, we
            // create a mongodb::Database before spawning the db-schema task.
            let db = options.client.database(db_name.as_str());
            let tx_notifications = options.tx_notifications.clone();
            let tx_schemata = options.tx_schemata.clone();

            // Spawn the db task.
            tokio::runtime::Handle::current().spawn(async move {
                // To start computing the schema for all collections and views
                // in a database, we need to wait for list_collections to finish.
                let collection_info = list_collections_and_views(&db).await;

                let (coll_tasks, view_tasks) = match collection_info {
                    Err(e) => {
                        notify!(
                            tx_notifications,
                            SamplerNotification {
                                db: db.name().to_string(),
                                collection_or_view: "".to_string(),
                                action: SamplerAction::Error {
                                    message: format!(
                                        "failed to listing collections with error {e}"
                                    ),
                                },
                            }
                        );
                        return;
                    }
                    // With all collections and views listed, we derive schema for
                    // each of them in parallel. For now, collections and views are
                    // processed differently -- views require sampling to derive
                    // schema. Given that, we do _not_ await the result of
                    // process_collections before moving on to processing views.
                    // In the future, if/when view schemas are calculated based on
                    // their underlying collection and their pipeline, we _should_
                    // await collection derivation before processing views.
                    // Here, we kick off collection processing and view processing
                    // in parallel and await all the collection and view tasks
                    // before concluding this database task.
                    Ok(collection_info) => (
                        collection_info.process_collections(
                            &db,
                            tx_notifications.clone(),
                            tx_schemata.clone(),
                        ),
                        collection_info.process_views(
                            &db,
                            tx_notifications.clone(),
                            tx_schemata.clone(),
                        ),
                    ),
                };

                let mut all_tasks = vec![];
                all_tasks.extend(coll_tasks);
                all_tasks.extend(view_tasks);

                // After spawning async tasks for each collection and view,
                // the final step is to wait for all of those tasks to finish
                // by awaiting the result of join_all for all collection task
                // and view task JoinHandles.
                future::join_all(all_tasks)
                    .await
                    .into_iter()
                    .for_each(|coll_schema_res| {
                        if let Err(e) = coll_schema_res {
                            notify!(
                                tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: "".to_string(),
                                    action: SamplerAction::Error {
                                        message: format!(
                                            "failed to complete schema building with error {e}"
                                        ),
                                    },
                                }
                            );
                        }
                    });

                drop(tx_notifications);
                drop(tx_schemata);
            })
        });

    // After spawning async tasks for each database, the final step is to wait
    // for all of those task to finish by awaiting the result of join_all for
    // all database task JoinHandles.
    future::join_all(db_tasks)
        .await
        .into_iter()
        .for_each(|db_schema_res| {
            if let Err(e) = db_schema_res {
                notify!(
                    options.tx_notifications,
                    SamplerNotification {
                        db: "".to_string(),
                        collection_or_view: "".to_string(),
                        action: SamplerAction::Error {
                            message: format!("failed to complete schema building with error {e}"),
                        },
                    }
                )
            }
        });

    // After waiting for all database tasks to finish, we can safely drop
    // both channels.
    drop(options.tx_notifications);
    drop(options.tx_schemata);
}

/*************************************************************************************************/
/******** Database Utility Functions *************************************************************/
/*************************************************************************************************/

/// A utility struct representing the result of listCollections. It includes the list of collections
/// and the list of views.
#[derive(Debug, Default)]
struct CollectionInfo {
    views: Vec<CollectionDoc>,
    collections: Vec<CollectionDoc>,
}

/// A utility function for listing collections and views.
async fn list_collections_and_views(database: &Database) -> Result<CollectionInfo> {
    let collection_info_cursor = database
        .run_cursor_command(
            doc! { "listCollections": 1.0, "authorizedCollections": true},
            None,
        )
        .await
        .map_err(errors::Error::from)?;

    separate_views_from_collections(collection_info_cursor).await
}

async fn separate_views_from_collections(
    mut collection_doc: Cursor<Document>,
) -> Result<CollectionInfo> {
    let mut collection_info = CollectionInfo::default();
    while let Some(collection_doc) = collection_doc.try_next().await.unwrap() {
        let collection_doc: CollectionDoc =
            bson::from_bson(bson::Bson::Document(collection_doc)).unwrap();
        if collection_doc.type_ == "view" {
            collection_info.views.push(collection_doc);
        } else {
            collection_info.collections.push(collection_doc);
        }
    }

    Ok(collection_info)
}

impl CollectionInfo {
    /// process_collections creates parallel, async tasks for deriving the
    /// schema for each collection in the CollectionInfo. It iterates through
    /// each collection and spawns a new async task to compute its schema.
    /// Importantly, like database tasks, we do not await the spawned tasks.
    /// Each async task will start running in the background immediately,
    /// but the program will continue executing the iteration through all
    /// collections since tokio::spawn immediately returns a JoinHandle.
    /// This method returns the list of JoinHandles for the caller to await
    /// as needed.
    fn process_collections(
        &self,
        db: &Database,
        tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<Result<SchemaResult>>,
    ) -> Vec<JoinHandle<()>> {
        self.collections
            .as_slice()
            .iter()
            .filter(|coll_doc| !DISALLOWED_COLLECTION_NAMES.contains(&coll_doc.name.as_str()))
            .map(|coll_doc| {
                let db = db.clone();
                let coll = db.collection::<Document>(coll_doc.name.as_str());
                let tx_notifications = tx_notifications.clone();
                let tx_schemata = tx_schemata.clone();

                tokio::runtime::Handle::current().spawn(async move {
                    // To start computing the schema for a collection, we need
                    // to determine the partitions of this collection.
                    let partitions = get_partitions(&coll).await;

                    match partitions {
                        Err(e) => {
                            // If partitioning the collection fails, there is nothing to
                            // do so we report and error and return.
                            notify!(
                                tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: coll.name().to_string(),
                                    action: SamplerAction::Error {
                                        message: format!("failed to partition with error {e}"),
                                    },
                                }
                            )
                        }
                        Ok(partitions) => {
                            // If partitioning succeeds, we send a notification
                            // to indicate partitioning is happening, then we
                            // derive schema for the partitions.
                            notify!(
                                tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: coll.name().to_string(),
                                    action: SamplerAction::Partitioning {
                                        partitions: partitions.len() as u16,
                                    },
                                }
                            );

                            // The derive_schema_for_partitions function
                            // parallelizes schema derivation per partition.
                            // So here, we await its result and then send it
                            // over the schemata channel as the final step in
                            // the collection task.
                            let coll_schema = derive_schema_for_partitions(
                                db.name().to_string(),
                                &coll,
                                partitions,
                                &tokio::runtime::Handle::current(),
                                tx_notifications.clone(),
                            )
                            .await;
                            match coll_schema {
                                Err(e) => {
                                    // If deriving schema for the collection
                                    // fails, there is nothing to do so we
                                    // report and error and return.
                                    notify!(
                                        tx_notifications,
                                        SamplerNotification {
                                            db: db.name().to_string(),
                                            collection_or_view: coll.name().to_string(),
                                            action: SamplerAction::Error {
                                                message: format!(
                                                    "failed to derive schema with error {e}"
                                                ),
                                            },
                                        }
                                    )
                                }
                                Ok(coll_schema) => {
                                    // If deriving schema succeeds, we send
                                    // the schema over the schemata channel.
                                    let _ = tx_schemata.send(Ok(SchemaResult {
                                        db_name: db.name().to_string(),
                                        coll_or_view_name: coll.name().to_string(),
                                        namespace_type: NamespaceType::Collection,
                                        namespace_schema: coll_schema,
                                    }));
                                }
                            }
                            drop(tx_notifications);
                            drop(tx_schemata);
                        }
                    }
                })
            })
            .collect()
    }

    /// process_views creates parallel, async tasks for deriving the schema
    /// for each view in the CollectionInfo. It iterates through each view
    /// and spawns a new async task to compute its schema. Importantly, again,
    /// we do not await the spawned tasks. Each async task will start running
    /// in the background immediately, but the program will continue executing
    /// the iteration through all views since tokio::spawn immediately returns
    /// a JoinHandle. This method return the list of JoinHandles for the caller
    /// to await as needed.
    fn process_views(
        &self,
        db: &Database,
        tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<Result<SchemaResult>>,
    ) -> Vec<JoinHandle<()>> {
        self.views
            .as_slice()
            .iter()
            .map(|view_doc| {
                let db = db.clone();
                let view_doc = view_doc.clone();
                let tx_notifications = tx_notifications.clone();
                let tx_schemata = tx_schemata.clone();

                tokio::runtime::Handle::current().spawn(async move {
                    let view_doc = view_doc.clone();
                    // Since view schemas depend on sampling, this is a
                    // straightforward task: simply await the result of schema
                    // derivation and send it when it's done.
                    let schema =
                        derive_schema_for_view(&view_doc, &db, tx_notifications.clone()).await;
                    let _ = tx_schemata.send(Ok(SchemaResult {
                        db_name: db.name().to_string(),
                        coll_or_view_name: view_doc.name.clone(),
                        namespace_type: NamespaceType::View,
                        namespace_schema: schema,
                    }));
                    drop(tx_notifications);
                    drop(tx_schemata);
                })
            })
            .collect()
    }
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct CollectionDoc {
    #[serde(rename = "type")]
    type_: String,
    name: String,
    options: ViewOptions,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct ViewOptions {
    #[serde(rename = "viewOn", default)]
    view_on: String,
    #[serde(default)]
    pipeline: Vec<Document>,
}

impl Default for ViewOptions {
    fn default() -> Self {
        Self {
            view_on: "".to_string(),
            pipeline: vec![],
        }
    }
}

/*************************************************************************************************/
/******** Partitioning Utility Functions *********************************************************/
/*************************************************************************************************/

/// A utility function for determining the partitions of a collection.
///
/// Collections are partitioned into 100MB chunks. The minimum bound within each chunk is found
/// by using a $bucketAuto stage that groups on the _id field. If collections are <= 100MB this
/// operation is skipped and instead the entire collection is considered one partition. Note that
/// the 100MB limit comes from the server, as noted in $bucketAuto docs [here|https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucketAuto/#-bucketauto-and-memory-restrictions].
async fn get_partitions(collection: &Collection<Document>) -> Result<Vec<Partition>> {
    let size_info = get_size_counts(collection).await?;
    let num_partitions = get_num_partitions(size_info.size, PARTITION_SIZE_IN_BYTES) as usize;
    let (mut min_bound, max_bound) = get_bounds(collection).await?;
    let mut partitions = Vec::with_capacity(num_partitions);

    let first_stage = if size_info.count >= SAMPLE_MIN_DOCS && num_partitions > 1 {
        let num_docs_to_sample = std::cmp::min(
            (SAMPLE_RATE * size_info.count as f64) as u64,
            MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION as u64 * num_partitions as u64,
        );
        doc! { "$sample": {"size": num_docs_to_sample as i64 } }
    } else {
        doc! { "$match": {} }
    };

    let mut cursor = collection
        .aggregate(
            vec![
                first_stage,
                doc! { "$project": {"_id": 1} },
                doc! { "$bucketAuto": {
                    "groupBy": "$_id",
                    "buckets": num_partitions as i32
                }},
            ],
            None,
        )
        .await?;

    while let Some(doc) = cursor.try_next().await.unwrap() {
        let doc = doc.get("_id").unwrap().as_document().unwrap();
        let max = doc.get("max").cloned().unwrap_or(Bson::MaxKey);
        partitions.push(Partition {
            min: min_bound.clone(),
            max: max.clone(),
        });
        if max != max_bound {
            min_bound = max;
        }
    }
    partitions.push(Partition {
        min: min_bound,
        max: max_bound,
    });

    Ok(partitions)
}

/// Utility struct for collection size and count information.
struct CollectionSizes {
    pub size: i64,
    pub count: i64,
}

/// get_size_counts uses the $collStats aggregation stage to get size and count information for the
/// argued collection.
async fn get_size_counts(collection: &Collection<Document>) -> Result<CollectionSizes> {
    let mut cursor = collection
        .aggregate(vec![doc! {"$collStats": {"storageStats": {}}}], None)
        .await
        .map_err(|_| Error::NoCollectionStats(collection.name().to_string()))?;
    if let Some(stats) = cursor.try_next().await.unwrap() {
        let stats = stats
            .get_document("storageStats")
            .map_err(|_| Error::BsonFailure)?;
        let size = stats
            .get("size")
            .cloned()
            .ok_or_else(|| Error::BsonFailure)?;
        let count = stats
            .get("count")
            .cloned()
            .ok_or_else(|| Error::BsonFailure)?;
        let (size, count) = match (size, count) {
            (Bson::Int32(size), Bson::Int32(count)) => (size as i64, count as i64),
            (Bson::Int32(size), Bson::Int64(count)) => (size as i64, count),
            (Bson::Int64(size), Bson::Int32(count)) => (size, count as i64),
            (Bson::Int64(size), Bson::Int64(count)) => (size, count),
            _ => {
                return Err(Error::BsonFailure);
            }
        };
        return Ok(CollectionSizes { size, count });
    }
    Err(Error::NoCollectionStats(collection.name().to_string()))
}

/// get_num_partitions determines the number of partitions by dividing the collection size by the
/// partition size (and adding 1).
fn get_num_partitions(coll_size: i64, partition_size: i64) -> i64 {
    let num_parts = (coll_size as f64) / (partition_size as f64);
    num_parts as i64 + 1
}

/// get_bounds determines the minimum and maximum values for the _id field in the argued collection.
async fn get_bounds(collection: &Collection<Document>) -> Result<(Bson, Bson)> {
    Ok((
        get_bound(collection, 1).await?,
        // we actually will just always use MaxKey as our upper bound since we
        // match $lt max bound
        Bson::MaxKey,
    ))
}

/// get_bound determines the minimum or maximum bound (depending on the direction) for the _id field
/// in the provided collection.
async fn get_bound(collection: &Collection<Document>, direction: i32) -> Result<Bson> {
    let mut cursor = collection
        .aggregate(
            vec![
                doc! {"$sort": {"_id": direction}},
                doc! {"$limit": 1},
                doc! {"$project": {"_id": 1}},
            ],
            None,
        )
        .await
        .map_err(|e| Error::NoBounds(format!("{}: {e}", collection.name())))?;
    if let Some(doc) = cursor.try_next().await? {
        return doc
            .get("_id")
            .cloned()
            .ok_or(Error::NoBounds(collection.name().to_string()));
    }
    Err(Error::NoBounds(collection.name().to_string()))
}

/*************************************************************************************************/
/******** Schema Derivation Utility Functions ****************************************************/
/*************************************************************************************************/

/// A utility function for deriving the schema for a collection based on its partitions.
///
/// For each partition:
///  1. If there is an initial schema
///     The initial schema is used as the "seed" for the partition.
///
///  2. If there is no initial schema
///     An aggregation operation is issued that sorts based on the _id and limits the result set to
///     20 documents. The schema for these documents is computed and "seeds" the schema for the
///     partition.
///
/// The builder then enters a loop where it repeatedly issues query operations for documents that do
/// not match the computed schema. The schema for each resulting document is calculated and merged
/// with the existing schema, resulting in a new schema that meets all encountered documents so far.
/// This operation repeats until there are no more results within the partition.
///
/// The results of each partition are unioned together to produce the schema of the entire
/// collection.
async fn derive_schema_for_partitions(
    db_name: String,
    collection: &Collection<Document>,
    col_parts: Vec<Partition>,
    rt_handle: &tokio::runtime::Handle,
    tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
) -> Result<Option<Schema>> {
    let partition_tasks = col_parts.into_iter().enumerate().map(|(ix, partition)| {
        let db_name = db_name.clone();
        let tx_notifications = tx_notifications.clone();
        let collection = collection.clone();
        rt_handle.spawn(async move {
            let schema_res = derive_schema_for_partition(
                db_name.clone(),
                &collection,
                partition,
                None,
                tx_notifications.clone(),
                ix,
            )
            .await;

            let schema = match schema_res {
                Err(e) => {
                    notify!(
                        tx_notifications,
                        SamplerNotification {
                            db: db_name,
                            collection_or_view: collection.name().to_string(),
                            action: SamplerAction::Error {
                                message: format!(
                                    "failed derive schema for partition {ix} with error {e}",
                                ),
                            },
                        }
                    );
                    // If we encounter an error when processing a partition,
                    // we effectively ignore it by saying the schema is Unsat.
                    // Note that for any schema s, Unsat.union(s) == s. Users
                    // can check there error notifications to see that schema
                    // building for this collection may be incomplete and can
                    // consider rerunning the builder if they wish.
                    Schema::Unsat
                }
                Ok(schema) => schema,
            };

            drop(tx_notifications);
            schema
        })
    });

    // Here, we await all partitions and then union them together to create the
    // full collection schema. Note that we could union the partition schemas as
    // they are produced for a marginal speedup, however that would require
    // a mutable Schema value that is guarded by a lock. The overhead of locking
    // and unioning per thread is likely equivalent to the more straightforward
    // wait-and-then-union solution we have here.
    Ok(future::try_join_all(partition_tasks)
        .await
        .map_err(Error::TokioError)?
        .into_iter()
        .reduce(|full_coll_schema, part_schema| full_coll_schema.union(&part_schema)))
}

/// A utility function for deriving the schema for a single partition of a collection.
async fn derive_schema_for_partition(
    db_name: String,
    collection: &Collection<Document>,
    mut partition: Partition,
    initial_schema_doc: Option<Document>,
    tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
    partition_ix: usize,
) -> Result<Schema> {
    let mut schema: Option<Schema> = initial_schema_doc
        .clone()
        .map(schema_doc_to_schema)
        .transpose()?;
    let mut first_stage = Some(generate_partition_match_with_doc(
        &partition,
        initial_schema_doc,
    )?);

    loop {
        if partition.min == partition.max {
            break;
        }

        if first_stage.is_none() {
            first_stage = Some(generate_partition_match(&partition, schema.clone())?);
        };
        notify!(
            tx_notifications,
            SamplerNotification {
                db: db_name.clone(),
                collection_or_view: collection.name().to_string(),
                action: SamplerAction::Querying {
                    partition: partition_ix as u16,
                },
            }
        );
        let mut cursor = collection
            .aggregate(
                vec![
                    // first_stage must be Some here.
                    first_stage.unwrap(),
                    doc! { "$sort": {"_id": 1}},
                    doc! { "$limit": MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION },
                ],
                AggregateOptions::builder()
                    .hint(Some(mongodb::options::Hint::Keys(doc! {"_id": 1})))
                    .build(),
            )
            .await?;
        first_stage = None;

        let mut no_result = true;
        while let Some(doc) = cursor.try_next().await.unwrap() {
            notify!(
                tx_notifications,
                SamplerNotification {
                    db: db_name.clone(),
                    collection_or_view: collection.name().to_string(),
                    action: SamplerAction::Processing {
                        partition: partition_ix as u16,
                    },
                }
            );
            let id = doc.get("_id").unwrap().clone();
            partition.min = id;
            schema = Some(schema_for_document(&doc).union(&schema.unwrap_or(Schema::Unsat)));
            no_result = false;
        }
        if no_result {
            break;
        }
    }
    drop(tx_notifications);
    Ok(schema.unwrap_or(Schema::Unsat))
}

/// derive_schema_for_view takes a CollectionDoc and executes the pipeline
/// against the viewOn collection to generate a schema for the view.
/// It does this by first prepending $sample to the pipeline
async fn derive_schema_for_view(
    view: &CollectionDoc,
    database: &Database,
    tx_notification: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
) -> Option<Schema> {
    let pipeline = vec![doc! { "$sample": { "size": SAMPLE_SIZE } }]
        .into_iter()
        .chain(view.options.pipeline.clone().into_iter())
        .collect::<Vec<Document>>();

    let mut schema = None;

    match database
        .collection::<Document>(&view.options.view_on)
        .aggregate(pipeline, None)
        .await
    {
        Ok(mut cursor) => {
            let mut iterations = 0;
            while let Some(doc) = cursor.try_next().await.unwrap() {
                // we want to notify every 100 iterations so it isn't too spammy
                if iterations % 100 == 0 {
                    notify!(
                        tx_notification,
                        SamplerNotification {
                            db: database.name().to_string(),
                            collection_or_view: view.name.clone(),
                            action: SamplerAction::SamplingView,
                        }
                    );
                }
                if schema.is_none() {
                    schema = Some(schema_for_document(&doc));
                } else {
                    schema = Some(schema.unwrap().union(&schema_for_document(&doc)));
                }
                iterations += 1;
            }
        }
        Err(e) => {
            notify!(
                tx_notification,
                SamplerNotification {
                    db: database.name().to_string(),
                    collection_or_view: view.name.clone(),
                    action: SamplerAction::Error {
                        message: e.to_string(),
                    },
                }
            );
        }
    }

    drop(tx_notification);
    schema
}

/// Returns a [Schema] for a given BSON document.
pub fn schema_for_document(doc: &Document) -> Schema {
    Schema::Document(mongosql::schema::Document {
        keys: doc
            .iter()
            .map(|(k, v)| (k.to_string(), schema_for_bson(v)))
            .collect(),
        required: doc.iter().map(|(k, _)| k.to_string()).collect(),
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    })
}

fn schema_for_bson(b: &Bson) -> Schema {
    use Atomic::*;
    match b {
        Bson::Double(_) => Schema::Atomic(Double),
        Bson::String(_) => Schema::Atomic(String),
        Bson::Array(a) => Schema::Array(Box::new(schema_for_bson_array(a))),
        Bson::Document(d) => schema_for_document(d),
        Bson::Boolean(_) => Schema::Atomic(Boolean),
        Bson::Null => Schema::Atomic(Null),
        Bson::RegularExpression(_) => Schema::Atomic(Regex),
        Bson::JavaScriptCode(_) => Schema::Atomic(Javascript),
        Bson::JavaScriptCodeWithScope(_) => Schema::Atomic(JavascriptWithScope),
        Bson::Int32(_) => Schema::Atomic(Integer),
        Bson::Int64(_) => Schema::Atomic(Long),
        Bson::Timestamp(_) => Schema::Atomic(Timestamp),
        Bson::Binary(_) => Schema::Atomic(BinData),
        Bson::ObjectId(_) => Schema::Atomic(ObjectId),
        Bson::DateTime(_) => Schema::Atomic(Date),
        Bson::Symbol(_) => Schema::Atomic(Symbol),
        Bson::Decimal128(_) => Schema::Atomic(Decimal),
        Bson::Undefined => Schema::Atomic(Null),
        Bson::MaxKey => Schema::Atomic(MaxKey),
        Bson::MinKey => Schema::Atomic(MinKey),
        Bson::DbPointer(_) => Schema::Atomic(DbPointer),
    }
}

// This may prove costly for very large arrays, and we may want to
// consider a limit on the number of elements to consider.
fn schema_for_bson_array(bs: &[Bson]) -> Schema {
    // if an array is empty, we can't infer anything about it
    // we're safe to mark it as potentially null, as an empty array
    // satisfies jsonSchema search predicate
    if bs.is_empty() {
        return Schema::AnyOf(set!(Schema::Atomic(Atomic::Null)));
    }
    bs.iter()
        .map(schema_for_bson)
        .reduce(|acc, s| acc.union(&s))
        .unwrap_or(Schema::Any)
}

#[derive(Debug, PartialEq, Clone)]
pub struct Partition {
    pub min: Bson,
    pub max: Bson,
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an optional Schema. If the Schema is None, the $match will only
// be based on the Partition bounds.
pub fn generate_partition_match(partition: &Partition, schema: Option<Schema>) -> Result<Document> {
    generate_partition_match_with_doc(partition, schema.map(schema_to_schema_doc).transpose()?)
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an input jsonSchema.
pub fn generate_partition_match_with_doc(
    partition: &Partition,
    schema: Option<Document>,
) -> Result<Document> {
    let mut match_body = doc! {
        "_id": {
            "$gte": partition.min.clone(),
            "$lt": partition.max.clone(),
        }
    };
    if let Some(schema) = schema {
        match_body.insert("$nor", vec![schema]);
    }
    Ok(doc! {
        "$match": match_body
    })
}

pub fn schema_to_schema_doc(schema: Schema) -> Result<Document> {
    let json_schema: json_schema::Schema = schema
        .clone()
        .try_into()
        .map_err(|_| Error::JsonSchemaFailure)?;
    let bson_schema = bson::to_bson(&json_schema).map_err(|_| Error::BsonFailure)?;
    let ret = doc! {
        "$jsonSchema": bson_schema
    };
    Ok(ret)
}

pub fn schema_doc_to_schema(schema_doc: Document) -> Result<Schema> {
    let json_schema: json_schema::Schema =
        bson::from_document(schema_doc.get_document("$jsonSchema").unwrap().clone())
            .map_err(|_| Error::BsonFailure)?;
    let sampler_schema: Schema = json_schema
        .try_into()
        .map_err(|_| Error::JsonSchemaFailure)?;
    Ok(sampler_schema)
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Schemata {
    #[serde(rename = "_id")]
    pub id: SchemataId,
    pub schema: Document,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct SchemataId {
    pub db: String,
    pub collection: String,
}

// TODO: SQL-2156: Query for initial schema
// pub async fn gen_partitions_with_initial_schema(
//     collections_and_schemata: Vec<(String, Document)>,
//     database: &Database,
// ) -> HashMap<String, (Document, Vec<Partition>)> {
//     println!("getting partitions with schemata");
//     let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
//     rayon::scope(|s| {
//         for (coll, sch) in collections_and_schemata {
//             let coll = coll.clone();
//             let tx = tx.clone();
//             let rt = tokio::runtime::Runtime::new().unwrap();
//             let database = database.clone();
//             s.spawn(move |_| {
//                 rt.block_on(async move {
//                     let partitions = get_partitions(&database, &coll).await.unwrap();
//                     tx.send(((coll, sch), partitions)).unwrap();
//                     drop(tx)
//                 });
//                 drop(rt);
//             })
//         }
//     });
//     drop(tx);
//     let mut col_parts = HashMap::new();
//     while let Some(((coll, sch), partitions)) = rx.recv().await {
//         col_parts.insert(coll, (sch, partitions));
//     }
//     col_parts
// }
