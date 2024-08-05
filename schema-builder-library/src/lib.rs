use futures::future;
use mongodb::{
    bson::{doc, Document},
    options::ListDatabasesOptions,
};
use mongosql::schema::Schema;
use std::{
    collections::HashMap,
    fmt::{self, Display, Formatter},
    sync::Arc,
};
use tokio::sync::RwLock;
use tracing::{debug, instrument, span, Level};

pub mod client_util;
mod consts;
use consts::{
    DISALLOWED_DB_NAMES, PARTITION_DOCS_PER_ITERATION, PARTITION_SIZE_IN_BYTES, VIEW_SAMPLE_SIZE,
};
mod collection;
use collection::{query_for_initial_schemas, CollectionDoc, CollectionInfo};
mod partitioning;
use partitioning::{
    generate_partition_match, generate_partition_match_with_doc, get_partitions, Partition,
};
mod schema;
use schema::{derive_schema_for_partitions, derive_schema_for_view};
mod errors;
pub use errors::Error;
mod notifications;
pub use notifications::{SamplerAction, SamplerNotification};

pub mod options;
use options::BuilderOptions;

#[cfg(feature = "integration")]
#[cfg(test)]
mod internal_integration_tests;

pub type Result<T> = std::result::Result<T, Error>;

/// An enum for communicating results of this library to a caller. Results may
/// be namespace-only, meaning they do not include schema information, or they
/// may include schema information.
#[derive(Debug, PartialEq)]
pub enum SchemaResult {
    /// SchemaResult without schema info. Used in dry_run mode.
    NamespaceOnly(NamespaceInfo),

    /// SchemaResult with schema info.
    FullSchema(NamespaceInfoWithSchema),
}

/// A struct representing namespace information for a view or collection.
#[derive(Debug, PartialEq)]
pub struct NamespaceInfo {
    /// The name of the database.
    pub db_name: String,

    /// The name of the collection or view which this schema represents.
    pub coll_or_view_name: String,

    /// The type of namespace (collection or view).
    pub namespace_type: NamespaceType,
}

/// A struct representing schema information for a specific namespace (a view
/// or collection).
#[derive(Debug, PartialEq)]
pub struct NamespaceInfoWithSchema {
    pub namespace_info: NamespaceInfo,

    /// The schema for the namespace.
    pub namespace_schema: Schema,
}

/// An enum representing the two namespace types for which this library
/// can generate schema: Collection and View.
#[derive(Debug, PartialEq)]
pub enum NamespaceType {
    Collection,
    View,
}

impl Display for NamespaceType {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        match self {
            NamespaceType::Collection => write!(f, "Collection"),
            NamespaceType::View => write!(f, "View"),
        }
    }
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
#[instrument(name = "schema builder", level = "info")]
pub async fn build_schema(options: BuilderOptions) {
    // To start computing the schema for all databases, we need to wait for the
    // list_database_names method to finish.
    match options
        .client
        .list_database_names()
        .with_options(Some(
            ListDatabasesOptions::builder()
                .filter(doc! {"name": { "$nin": DISALLOWED_DB_NAMES.to_vec()} })
                .build(),
        ))
        .await
    {
        Ok(databases) => process_databases(options, databases).await,
        // If listing the databases fails, there is nothing to do so we report an
        // error, drop all channels, and return.
        Err(e) => {
            notify!(
                &options.tx_notifications,
                SamplerNotification {
                    db: "".to_string(),
                    collection_or_view: "".to_string(),
                    action: SamplerAction::Error {
                        message: format!("Unable to list databases: {e}"),
                    },
                },
            );
            drop(options.tx_notifications);
            drop(options.tx_schemata);
            return;
        }
    }
}

#[instrument(level = "debug", skip_all)]
async fn process_databases(options: BuilderOptions, databases: Vec<String>) {
    // we've checked the error condition above, so this should be safe

    let span = span!(Level::DEBUG, "initial_schemas", databases = ?databases);

    // If the schema_collection option is set and dry_run is not specified, we query for the initial schemas prior
    // to starting the schema building process. We store these initial schemas in a
    // Arc<RwLock<_>> to allow for concurrent access to the initial schemas and free clones
    // across all tasks.
    let initial_schemas: Arc<RwLock<HashMap<String, HashMap<String, Document>>>> =
        Arc::new(RwLock::new(HashMap::new()));
    if options.schema_collection.is_some() && !options.dry_run {
        let _enter = span.enter();
        future::join_all(databases.as_slice().iter().map(|db_name| {
            let db = options.client.database(db_name);
            let schema_collection = options.schema_collection.clone();
            let initial_schemas = initial_schemas.clone();
            let tx_notifications = options.tx_notifications.clone();
            async move {
                match query_for_initial_schemas(schema_collection, &db).await {
                    Ok(initial_collection_schemas) => {
                        debug!(
                            "queried for intial schemas in database {}, found {}",
                            db_name,
                            initial_collection_schemas.len()
                        );
                        initial_schemas
                            .write()
                            .await
                            .insert(db_name.to_string(), initial_collection_schemas);
                    }
                    Err(e) => {
                        notify!(
                            tx_notifications,
                            SamplerNotification {
                                db: db_name.to_string(),
                                collection_or_view: "".to_string(),
                                action: SamplerAction::Warning {
                                    message: format!(
                                    "failed to query for initial schemas in database with error {e}"
                                ),
                                },
                            },
                        );
                    }
                }
            }
        }))
        .await;
    }
    // Here, we iterate through each database and spawn a new async task to
    // compute each db schema. Importantly, we are not awaiting the spawned
    // tasks. Each async task will start running in the background immediately,
    // but the program will continue executing the iteration since tokio::spawn
    // immediately returns a JoinHandle.
    let span = span!(Level::DEBUG, "database_tasks", databases = ?databases);
    let _enter = span.enter();
    let db_tasks = databases.into_iter().map(|db_name| {
        // To avoid passing a reference to the mongodb::Client around, we
        // create a mongodb::Database before spawning the db-schema task.
        let db = options.client.database(&db_name);
        let initial_schemas = initial_schemas.clone();
        let tx_notifications = options.tx_notifications.clone();
        let tx_schemata = options.tx_schemata.clone();

        let include_list = options.include_list.clone();
        let exclude_list = options.exclude_list.clone();

        // Spawn the db task.
        tokio::runtime::Handle::current().spawn(async move {
            // To start computing the schema for all collections and views
            // in a database, we need to wait for list_collections to finish.
            let collection_info =
                CollectionInfo::new(&db, &db_name, include_list, exclude_list).await;

            let (coll_tasks, view_tasks) = match collection_info {
                Err(e) => {
                    notify!(
                        &tx_notifications,
                        SamplerNotification {
                            db: db.name().to_string(),
                            collection_or_view: "".to_string(),
                            action: SamplerAction::Warning {
                                message: format!(
                                    "failed to listing collections in database with error {e}"
                                ),
                            },
                        },
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
                        options.dry_run,
                        initial_schemas.clone(),
                        tx_notifications.clone(),
                        tx_schemata.clone(),
                    ),
                    collection_info.process_views(
                        &db,
                        options.dry_run,
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
                            &tx_notifications,
                            SamplerNotification {
                                db: db.name().to_string(),
                                collection_or_view: "".to_string(),
                                action: SamplerAction::Warning {
                                    message: format!(
                                        "failed to complete schema building with error {e}"
                                    ),
                                },
                            },
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
                    &options.tx_notifications,
                    SamplerNotification {
                        db: "".to_string(),
                        collection_or_view: "".to_string(),
                        action: SamplerAction::Warning {
                            message: format!("failed to complete schema building with error {e}"),
                        },
                    },
                );
            }
        });

    // After waiting for all database tasks to finish, we can safely drop
    // both channels.
    drop(options.tx_notifications);
    drop(options.tx_schemata);
}
