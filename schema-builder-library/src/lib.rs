use agg_ast::definitions::Namespace;
use futures::future;
use mongosql::schema::Schema;
use result_set::ResultSet;
use std::{
    collections::HashMap,
    fmt::{self, Display, Formatter},
    sync::Arc,
};
use tokio::sync::RwLock;
use tracing::{Level, debug, error, instrument, span, warn};

pub(crate) mod result_set;

// DataService trait and implementations
pub mod data_service;
pub use data_service::{
    CollectionInfo, CollectionOptions, CollectionType, DataService, TimeSeriesOptions,
};

#[cfg(all(feature = "native-client", feature = "wasm"))]
compile_error!("`native-client` and `wasm` features are mutually exclusive");

#[cfg(feature = "native-client")]
pub use data_service::MongoDbDataService;

#[cfg(feature = "wasm")]
pub use data_service::{JsDataService, WasmDataService};

#[cfg(feature = "native-client")]
pub mod client_util;

mod consts;
pub(crate) mod context;
use consts::{DISALLOWED_DB_NAMES, VIEW_SAMPLE_SIZE};
mod collection;
use collection::{DatabaseCollections, query_for_initial_schemas};
mod partitioning;
use partitioning::get_partitions;
mod errors;
mod schema;
pub use errors::Error;

/// Re-export of mongosql Schema type for convenience
pub type MongoSqlSchema = Schema;

pub mod options;
use options::BuilderOptions;

use crate::{context::Context, result_set::Catalog};

/// Re-export for consumers that call [`ResultSet::into_build_output`].
pub use result_set::SchemaBuildOutput;

#[cfg(feature = "integration")]
#[cfg(test)]
mod internal_integration_tests;

pub type Result<T, E> = std::result::Result<T, Error<E>>;

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

/// build_schema is the entry point for the schema-builder-library. Given a [`BuilderOptions`]
/// specifying various behaviors, this function builds schema for the appropriate namespaces of
/// the provided MongoDB instance. Importantly, this function must be called in a tokio::runtime
/// context since it accesses the current tokio runtime handle.
///
/// The function returns a [`ResultSet`] containing the derived schema for all matching namespaces.
///
/// For collections, this function produces a complete schema representing all data. It does
/// this by using an algorithm based on partitioning the data, repeatedly querying each
/// partition (in parallel) for documents that do not match the in-progress schema, and then
/// unifying the schema for each partition.
///
/// For views, this function produces a best-effort schema based on sampling.
///
/// This function parallelizes handling each database, collection, collection partition, and
/// view by using asynchronous tokio tasks.
#[instrument(name = "schema builder", level = "info")]
pub async fn build_schema<S: DataService>(
    options: BuilderOptions<S>,
) -> Result<ResultSet, S::Error> {
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

    // To start computing the schema for all databases, we need to wait for the list of
    // available databases.
    //
    // Note: The built-in mongodb `filter` option doesn't work for Atlas Free and
    // Shared tier clusters, so we have to manually filter out unwanted databases here.
    let valid_databases = options
        .service
        .list_database_names()
        .await
        .inspect_err(|e| tracing::error!("unable to list databases: {e}"))
        .map_err(Error::DataServiceError)?
        .into_iter()
        .filter(|db| !DISALLOWED_DB_NAMES.contains(&db.as_str()))
        .collect();

    Ok(process_databases(options, valid_databases).await)
}

#[instrument(level = "debug", skip_all)]
async fn process_databases<S: DataService>(
    options: BuilderOptions<S>,
    databases: Vec<String>,
) -> ResultSet {
    let span = span!(Level::DEBUG, "initial_schemas", databases = ?databases);

    // If the schema_collection option is set and dry_run is not specified, we query for the initial schemas prior
    // to starting the schema building process. We store these initial schemas in the ResultSet
    // for concurrent access across all tasks. Additionally, if a user specified only a view to be updated,
    // we need the initial schemas for the source collection(s).
    let initial_schemas = if let Some(schema_collection) = options.schema_collection.as_deref()
        && !options.dry_run
    {
        let _enter = span.enter();
        fetch_initial_schemas(&options.service, schema_collection, &databases)
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
    let ctx = Arc::new(Context::new(options.service));

    let db_tasks = databases.into_iter().map(|db| {
        let ctx = ctx.clone();
        let result_set = result_set.clone();
        let task_semaphore = task_semaphore.clone();
        let dry_run = options.dry_run;

        let include_list = options.include_list.clone();
        let exclude_list = options.exclude_list.clone();

        async move {
            let task_semaphore = task_semaphore.clone();

            // To start computing the schema for all collections and views
            // in a database, we need to wait for list_collections to finish.
            let Ok(database_collections) =
                DatabaseCollections::new(ctx.service(), db, include_list, exclude_list)
                    .await
                    .inspect_err(|e| {
                        warn!("failed to list collections in database with error: {e}")
                    })
            else {
                return;
            };

            // Process collections first, then views
            database_collections
                .process_collections(
                    Arc::clone(&ctx),
                    dry_run,
                    Arc::clone(&result_set),
                    Arc::clone(&task_semaphore),
                )
                .await;

            // Now process views using the schema catalog
            database_collections
                .process_views_with_catalog(Arc::clone(&ctx), dry_run, result_set, task_semaphore)
                .await;
        }
    });

    // After spawning async tasks for each database, the final step is to wait
    // for all of those task to finish by awaiting the result of join_all for
    // all database task JoinHandles.
    future::join_all(db_tasks).await;

    let Some(arc) = Arc::into_inner(result_set) else {
        panic!("Unexpected error: References to result_set remain after all tasks completed");
    };
    arc.into_inner()
}

async fn fetch_initial_schema<S: DataService>(
    service: &S,
    db: &str,
    collection: &str,
) -> Result<Catalog, S::Error> {
    query_for_initial_schemas(service, db, collection)
        .await
        .inspect_err(|e| warn!("failed to query for initial schemas in database with error {e}"))
        .map(|initial_collection_schemas| {
            debug!(
                "queried for initial schemas in database {db}, found {}",
                initial_collection_schemas.len()
            );

            initial_collection_schemas
                .into_iter()
                .map(|(coll_name, schema)| {
                    let namespace_info = NamespaceInfo {
                        db_name: db.to_string(),
                        coll_or_view_name: coll_name.clone(),
                        namespace_type: NamespaceType::Collection, // Assume Collection by default
                    };

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

async fn fetch_initial_schemas<S: DataService>(
    service: &S,
    collection: &str,
    databases: &[String],
) -> Result<HashMap<String, Catalog>, S::Error> {
    future::join_all(databases.iter().map(|db| async move {
        fetch_initial_schema(service, db, collection)
            .await
            .map(|catalog| (db.clone(), catalog))
    }))
    .await
    .into_iter()
    .collect()
}
