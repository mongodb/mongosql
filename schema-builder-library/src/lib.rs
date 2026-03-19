use agg_ast::definitions::Namespace;
use futures::future;
use mongodb::{Client, Database, bson::doc};
use mongosql::schema::{Document, Schema};
use result_set::ResultSet;
use std::{
    collections::HashMap,
    fmt::{self, Display, Formatter},
    sync::Arc,
};
use tokio::sync::RwLock;
use tracing::{Level, debug, error, instrument, span, warn};

pub(crate) mod result_set;

pub mod client_util;
mod consts;
use consts::{DISALLOWED_DB_NAMES, VIEW_SAMPLE_SIZE};
mod collection;
use collection::{CollectionDoc, CollectionInfo, query_for_initial_schemas};
mod partitioning;
use partitioning::get_partitions;
mod schema;
use schema::{derive_schema_for_partitions, derive_schema_for_view};
mod errors;
pub use errors::Error;
mod notifications;
pub use notifications::{SamplerAction, SamplerNotification};

/// Re-export of mongosql Schema type for convenience
pub type MongoSqlSchema = Schema;

pub mod options;
use options::BuilderOptions;

use crate::result_set::{Catalog, ShareableResultSet};

/// Re-export for consumers that call [`ResultSet::into_build_output`].
pub use result_set::SchemaBuildOutput;

#[cfg(feature = "integration")]
#[cfg(test)]
mod internal_integration_tests;

pub type Result<T> = std::result::Result<T, Error>;

/// An enum for communicating results of this library to a caller. Results may
/// be namespace-only, meaning they do not include schema information, or they
/// may include schema information.
#[derive(Debug, Clone)]
pub enum SchemaResult {
    /// SchemaResult without schema info. Used in dry_run mode.
    NamespaceOnly(NamespaceInfo),

    /// SchemaResult without schema info. Used when an initial schema
    /// is unstable.
    UnstableInitialSchema(NamespaceInfo),

    /// SchemaResult with initial schema info. It is up to the caller
    /// to decide if they want to do anything special with this schema.
    /// This schema is not guaranteed to be complete or correct.
    InitialSchema(NamespaceInfoWithSchema),

    /// SchemaResult with schema info.
    FullSchema(NamespaceInfoWithSchema),
}

impl PartialEq for SchemaResult {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (SchemaResult::NamespaceOnly(ns1), SchemaResult::NamespaceOnly(ns2)) => ns1 == ns2,
            (
                SchemaResult::UnstableInitialSchema(ns1),
                SchemaResult::UnstableInitialSchema(ns2),
            ) => ns1 == ns2,
            (SchemaResult::FullSchema(ns1), SchemaResult::FullSchema(ns2)) => ns1 == ns2,
            (SchemaResult::InitialSchema(ns1), SchemaResult::InitialSchema(ns2)) => ns1 == ns2,
            _ => false,
        }
    }
}

/// A struct representing namespace information for a view or collection.
#[derive(Debug, PartialEq, Clone)]
pub struct NamespaceInfo {
    /// The name of the database.
    pub db_name: String,

    /// The name of the collection or view which this schema represents.
    pub coll_or_view_name: String,

    /// The type of namespace (collection or view).
    pub namespace_type: NamespaceType,
}

impl From<NamespaceInfo> for Namespace {
    fn from(value: NamespaceInfo) -> Self {
        Namespace {
            database: value.db_name,
            collection: value.coll_or_view_name,
        }
    }
}

/// A struct representing schema information for a specific namespace (a view
/// or collection).
#[derive(Debug, PartialEq, Clone)]
pub struct NamespaceInfoWithSchema {
    pub namespace_info: NamespaceInfo,

    /// The schema for the namespace.
    pub namespace_schema: Arc<Schema>,
}

/// An enum representing the two namespace types for which this library
/// can generate schema: Collection and View.
#[derive(Debug, PartialEq, Clone)]
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
pub async fn build_schema(options: BuilderOptions) -> Result<ResultSet> {
    // Ensure that the `include_list` only contains valid patterns.
    for pattern in &options.include_list {
        let pat_as_str = pattern.as_str();
        if !pat_as_str.contains(".") {
            error!(
                "The `include_list` contains the following invalid pattern: `{pat_as_str}`. All patterns must be in `<database_pattern>.<collection_pattern>` format"
            );
            return Err(Error::IncludeOrExcludeListContainsInvalidPatterns(
                "include_list".to_string(),
                pat_as_str.to_string(),
            ));
        }
    }
    // Ensure that the `exclude_list` only contains valid patterns.
    for pattern in &options.exclude_list {
        let pat_as_str = pattern.as_str();
        if !pat_as_str.contains(".") {
            error!(
                "The `exclude_list` contains the following invalid pattern: `{pat_as_str}`. All patterns must be in `<database_pattern>.<collection_pattern>` format"
            );
            return Err(Error::IncludeOrExcludeListContainsInvalidPatterns(
                "exclude_list".to_string(),
                pat_as_str.to_string(),
            ));
        }
    }

    // To start computing the schema for all databases, we need to wait for the
    // list_database_names method to finish.
    match options.client.list_database_names().await {
        Ok(databases) => {
            // The list_database_names() `filter` option doesn't work for Atlas Free and Shared tier clusters,
            // so we have to manually filter out unwanted databases here.
            let valid_databases = databases
                .into_iter()
                .filter(|db| !DISALLOWED_DB_NAMES.contains(&db.as_str()))
                .collect();

            Ok(process_databases(options, valid_databases).await)
        }
        // If listing the databases fails, there is nothing to do so we report an
        // error, drop all channels, and return the error.
        Err(e) => {
            error!("unable to list databases: {e}");
            Err(Error::from(e))
        }
    }
}

#[instrument(level = "debug", skip_all)]
async fn process_databases(options: BuilderOptions, databases: Vec<String>) -> ResultSet {
    // we've checked the error condition above, so this should be safe

    let span = span!(Level::DEBUG, "initial_schemas", databases = ?databases);

    // If the schema_collection option is set and dry_run is not specified, we query for the initial schemas prior
    // to starting the schema building process. We store these initial schemas in the ResultSet
    // for concurrent access across all tasks. Additionally, if a user specified only a view to be updated,
    // we need the initial schemas for the source collection(s).
    let initial_schemas = if let Some(schema_collection) = options.schema_collection.as_deref()
        && !options.dry_run
    {
        let _enter = span.enter();
        fetch_initial_schemas(&options.client, schema_collection, &databases)
            .await
            .inspect_err(|e| {
                warn!("failed to query for initial schemas in database with error {e}")
            })
            .unwrap_or_default()
    } else {
        Default::default()
    };

    // Create a ResultSet to track all schemas
    let result_set = Arc::new(RwLock::new(ResultSet::new(initial_schemas)));

    // Here, we iterate through each database and spawn a new async task to
    // compute each db schema. Importantly, we are not awaiting the spawned
    // tasks. Each async task will start running in the background immediately,
    // but the program will continue executing the iteration since tokio::spawn
    // immediately returns a JoinHandle.
    let span = span!(Level::DEBUG, "database_tasks", databases = ?databases);
    let _enter = span.enter();
    let task_semaphore = options.task_semaphore.clone();
    let db_tasks = databases.into_iter().map(|db_name| {
        // To avoid passing a reference to the mongodb::Client around, we
        // create a mongodb::Database before spawning the db-schema task.
        let db = options.client.database(&db_name);
        let result_set = result_set.clone();

        let include_list = options.include_list.clone();
        let exclude_list = options.exclude_list.clone();

        let task_semaphore = task_semaphore.clone();
        // Spawn the db task.
        tokio::runtime::Handle::current().spawn(async move {
            let task_semaphore = task_semaphore.clone();

            // To start computing the schema for all collections and views
            // in a database, we need to wait for list_collections to finish.
            let Ok(collection_info) =
                CollectionInfo::new(&db, &db_name, include_list, exclude_list)
                    .await
                    .inspect_err(|e| {
                        warn!("failed to list collections in database with error: {e}")
                    })
            else {
                return;
            };

            // Process collections first, then views
            process_collection_tasks(
                &collection_info,
                &db,
                options.dry_run,
                Arc::clone(&result_set),
                task_semaphore.clone(),
            )
            .await;

            // Now process views using the schema catalog
            process_view_tasks(
                &collection_info,
                &db,
                options.dry_run,
                Arc::clone(&result_set),
                task_semaphore.clone(),
            )
            .await;
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
                warn!("failed to complete schema building with error: {e}");
            }
        });

    let Some(arc) = Arc::into_inner(result_set) else {
        panic!("Unexpected error: References to result_set remain after all tasks completed");
    };
    arc.into_inner()
}

async fn fetch_initial_schema(db: Database, collection: &str) -> Result<Catalog> {
    query_for_initial_schemas(collection, &db)
        .await
        .inspect_err(|e| warn!("failed to query for initial schemas in database with error {e}"))
        .map(|initial_collection_schemas| {
            debug!(
                "queried for initial schemas in database {}, found {}",
                db.name(),
                initial_collection_schemas.len()
            );

            initial_collection_schemas
                .into_iter()
                .map(|(coll_name, (mut schema, is_unstable))| {
                    let namespace_info = NamespaceInfo {
                        db_name: db.name().to_string(),
                        coll_or_view_name: coll_name.clone(),
                        namespace_type: NamespaceType::Collection, // Assume Collection by default
                    };

                    if is_unstable {
                        schema = match schema {
                            Schema::Document(d) => Schema::Document(Document {
                                unstable: true,
                                ..d
                            }),
                            _ => schema,
                        }
                    }

                    (
                        namespace_info.coll_or_view_name.clone(),
                        NamespaceInfoWithSchema {
                            namespace_info,
                            namespace_schema: Arc::new(schema),
                        },
                    )
                })
                .collect()
        })
}

async fn fetch_initial_schemas(
    client: &Client,
    schema_collection: &str,
    databases: &[String],
) -> Result<HashMap<String, Catalog>> {
    future::join_all(databases.iter().map(|db_name| {
        let db = client.database(db_name);
        let db_name = db.name().to_string();
        async move {
            fetch_initial_schema(db, schema_collection)
                .await
                .map(|catalog| (db_name, catalog))
        }
    }))
    .await
    .into_iter()
    .collect()
}

async fn process_collection_tasks(
    collection_info: &CollectionInfo,
    db: &mongodb::Database,
    dry_run: bool,
    result_set: ShareableResultSet,
    task_semaphore: Arc<tokio::sync::Semaphore>,
) {
    // Process collections and wait for them to complete
    let coll_tasks = collection_info.process_collections(
        db,
        dry_run,
        Arc::clone(&result_set),
        task_semaphore.clone(),
    );

    // Wait for all collections to finish before processing views
    future::join_all(coll_tasks)
        .await
        .into_iter()
        .for_each(|coll_schema_res| {
            if let Err(e) = coll_schema_res {
                warn!("failed to complete schema building with error: {e}");
            }
        });
}

async fn process_view_tasks(
    collection_info: &CollectionInfo,
    db: &mongodb::Database,
    dry_run: bool,
    result_set: ShareableResultSet,
    task_semaphore: Arc<tokio::sync::Semaphore>,
) {
    // Process views using the schema catalog
    let view_tasks = collection_info.process_views_with_catalog(
        db,
        dry_run,
        Arc::clone(&result_set),
        task_semaphore.clone(),
    );

    future::join_all(view_tasks)
        .await
        .into_iter()
        .for_each(|view_schema_res| {
            if let Err(e) = view_schema_res {
                warn!("failed to complete schema building with error: {e}");
            }
        });
}
