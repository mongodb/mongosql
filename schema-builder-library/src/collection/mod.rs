/**
 * This module contains functionality for processing MongoDB collections
 * and how we operate with them.
 */
use crate::{
    Error, NamespaceInfo, NamespaceInfoWithSchema, NamespaceType, Result, SamplerAction,
    client_util::DatabaseExt, derive_schema_for_partitions, derive_schema_for_view, get_partitions,
    result_set::ShareableResultSet, schema::initial_schema::InitialSchema,
};

mod patterns;
use agg_ast::Namespace;
use futures::TryStreamExt;
use mongodb::{
    Database,
    bson::{self, Document, doc},
};
use mongosql::schema::Schema;
use schema_derivation::{ResultSetState, derive_schema_for_pipeline};
use serde::{Deserialize, Serialize};
use std::{
    collections::{BTreeMap, HashMap},
    sync::{Arc, LazyLock, OnceLock},
};
use tokio::task::JoinHandle;
use tracing::{debug, info, instrument, warn};

#[cfg(test)]
mod test;

static EXCLUDE_DUNDERSCORE_PATTERN: LazyLock<glob::Pattern> = LazyLock::new(|| {
    #[allow(clippy::expect_used)]
    glob::Pattern::new("__*")
        .expect("Internal error: `__*` could not be converted into a `glob::Pattern`")
});

static INCLUDE_LIST_IN_DB_AND_COLL_PAIRS: OnceLock<Vec<(String, String)>> = OnceLock::new();

/// CollectionInfo is responsible for extracting the collections and views
/// and preparing them for processing.
#[derive(Debug, Default)]
pub(crate) struct CollectionInfo {
    pub views: Vec<CollectionDoc>,
    pub collections: Vec<CollectionDoc>,
    pub timeseries: Vec<CollectionDoc>,
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq, Default)]
pub struct CollectionDoc {
    #[serde(rename = "type")]
    pub type_: String,
    pub name: String,
    pub options: Options,
}

// The "options" field in listCollection output is overloaded to store different values for
// different collection types, so we do the same here. This struct may store view options, or it
// may store timeseries options. Consumers of this type should assert the CollectionDoc.type_ before
// accessing the relevant options values here.
#[derive(Debug, Serialize, Deserialize, Clone, PartialEq, Default)]
pub(crate) struct Options {
    #[serde(flatten, default)]
    pub view_options: Option<ViewOptions>,
    #[serde(rename = "timeseries", default)]
    pub timeseries_options: Option<TimeSeriesOptions>,
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq, Default)]
pub(crate) struct ViewOptions {
    #[serde(rename = "viewOn")]
    pub view_on: String,
    pub pipeline: Vec<Document>,
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq, Default)]
#[serde(rename_all = "camelCase")]
pub(crate) struct TimeSeriesOptions {
    pub time_field: String,
    pub meta_field: Option<String>,
}

#[instrument(level = "trace", skip_all)]
pub(crate) async fn query_for_initial_schemas(
    schema_collection: &str,
    database: &Database,
) -> Result<HashMap<String, (Schema, bool)>> {
    let mut initial_collection_schemas: HashMap<String, (Schema, bool)> = HashMap::new();

    let coll = database.collection::<Document>(schema_collection);
    let mut cursor = coll.find(doc! {}).await?;

    // Try to parse all of the initial schemas, failing if any of them fail to
    // parse correctly

    while let Some(doc) = cursor.try_next().await? {
        // If the unstable flag is not present or is present and is not a boolean, assume it is
        // false (i.e., this schema is stable)
        let unstable = doc
            .get("unstable")
            .unwrap_or(&bson::Bson::Boolean(false))
            .as_bool()
            .unwrap_or(false);

        // Convert the Doc into our InitialSchema struct
        let initial_schema = InitialSchema::try_from(doc)
            .map_err(|_| Error::InitialSchemaError(schema_collection.to_string()));

        // If conversion is successful, add the schema to the map
        match initial_schema {
            Ok(initial_schema) => {
                initial_collection_schemas
                    .insert(initial_schema.collection, (initial_schema.schema, unstable));
            }
            Err(_) => {
                return Err(Error::InitialSchemaError(String::from(schema_collection)));
            }
        }
    }

    Ok(initial_collection_schemas)
}

impl CollectionInfo {
    /// Create a new CollectionInfo instance. Collections and Views within a database
    /// will be enumerated, checked for inclusion/exclusion, and prepared for
    /// processing. The caller must actually process the collections/views by calling
    /// their JoinHandle.
    #[instrument(
        name = "processing collections for database",
        level = "info",
        skip(db, include_list, exclude_list)
    )]
    pub(crate) async fn new(
        db: &Database,
        db_name: &str,
        include_list: Vec<glob::Pattern>,
        exclude_list: Vec<glob::Pattern>,
    ) -> Result<Self> {
        let collection_info_cursor = db
            .run_cursor_command_with_read_preference(
                doc! { "listCollections": 1.0, "authorizedCollections": true},
            )
            .await?;

        CollectionInfo::separate_collection_types(
            db_name,
            &include_list,
            &exclude_list,
            collection_info_cursor,
        )
        .await
    }

    /// process_collections creates parallel, async tasks for deriving the
    /// schema for each collection in the CollectionInfo. It iterates through
    /// each collection and spawns a new async task to compute its schema.
    /// Importantly, like database tasks, we do not await the spawned tasks.
    /// Each async task will start running in the background immediately,
    /// but the program will continue executing the iteration through all
    /// collections since tokio::spawn immediately returns a JoinHandle.
    /// This method returns the list of JoinHandles for the caller to await
    /// as needed.
    #[instrument(skip_all)]
    pub(crate) fn process_collections(
        &self,
        db: &Database,
        dry_run: bool,
        result_set: ShareableResultSet,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) -> Vec<JoinHandle<()>> {
        self.collections
            .as_slice()
            .iter()
            .chain(self.timeseries.as_slice())
            .map(|collection_doc| {
                let db = db.clone();
                let collection_doc = collection_doc.clone();
                let result_set = result_set.clone();

                info!(name: "processing collection", collection = ?collection_doc);
                let collection_name = collection_doc.name.clone();

                let task_semaphore = task_semaphore.clone();
                // Spawn the collection task.
                tokio::runtime::Handle::current().spawn(async move {
                    let db_name = db.name();

                    // In dry_run mode, there is no need to derive schema for a view. Instead,
                    // we send just the namespace info and return.
                    if dry_run {
                        debug!("Received namespace-only info for {db_name}.{collection_name}");
                        result_set
                            .write()
                            .await
                            .mark_as_changed(db_name.to_string(), collection_name);

                        return;
                    }

                    let namespace_info = NamespaceInfo {
                        db_name: db_name.to_string(),
                        coll_or_view_name: collection_name.clone(),
                        namespace_type: NamespaceType::Collection,
                    };

                    let initial_schema = result_set
                        .read()
                        .await
                        .get_schema_for_database(db_name)
                        .and_then(|catalog| {
                            catalog
                                .get(&collection_name)
                                .map(|c| Arc::clone(&c.namespace_schema))
                        });
                    if let Some(schema) = &initial_schema {
                        info!(
                            "{}: {}:{}",
                            SamplerAction::UsingInitialSchema,
                            db.name(),
                            collection_name
                        );

                        if schema.is_unstable() {
                            result_set
                                .write()
                                .await
                                .mark_unstable_initial_schema(db_name.to_string(), collection_name.clone());
                            info!(
                                "Found unstable initial schema for namespace `{}.{}`. Marking as unstable.",
                                db_name, collection_name
                            );
                            return;
                        }
                    }

                    let collection = db.collection::<Document>(&collection_name);
                    info!(
                        db = db.name(),
                        collection = collection_name,
                        "Getting partitions"
                    );

                    // If we successfully retrieve partitions from the collection,
                    // derive the schema for each partition.
                    let Ok(partitioned_collection) = get_partitions(&collection, collection_doc.clone()).await.inspect_err(|e| {
                        warn!(
                            db = db.name(),
                            collection = collection_name,
                            "could not get partitions: {e}"
                        )
                    }) else {
                        return;
                    };

                    // Notify that we've gotten the partitions
                    info!(
                        db = db.name(),
                        collection = collection_name,
                        "found {} partitions",
                        partitioned_collection.partitions.len()
                    );

                    // Derive the schema for each partition, using the initial_schema
                    // as the foundation for the derived schema. Then, we do a union on the initial schema
                    // with the schema for each document in the partition.
                    let Ok(schema) = derive_schema_for_partitions(
                        db.name().to_string(),
                        &collection,
                        initial_schema,
                        &tokio::runtime::Handle::current(),
                        task_semaphore.clone(),
                        partitioned_collection,
                    )
                    .await
                    .inspect_err(|e| {
                        warn!(
                            db = db.name(),
                            collection = collection_name,
                            "could not derive schema: {e}"
                        )
                    }) else {
                        return;
                    };

                    let Some(schema) = schema else {
                        warn!(
                            db = db.name(),
                            collection = collection_name,
                            "no schema derived, collection may be empty"
                        );

                        return;
                    };

                    // Add the final schema to our results
                    result_set
                        .write()
                        .await
                        .add_schema(NamespaceInfoWithSchema {
                            namespace_info,
                            namespace_schema: Arc::new(schema),
                        });
                })
            })
            .collect()
    }

    /// process_views_with_catalog creates parallel, async tasks for deriving the schema
    /// for each view in the CollectionInfo using the schemas stored in the catalog.
    /// This method first gets the schemas from the catalog, then uses
    /// derive_schema_for_pipeline to generate view schemas.
    /// There are many fallback points to the old sampling method:
    /// - if the backing collection schemas are not present in the catalog
    /// - if we get a None schema for the underlying collection
    /// - if we fail to derive the schema from the pipeline
    #[instrument(skip_all)]
    pub(crate) fn process_views_with_catalog(
        &self,
        db: &Database,
        dry_run: bool,
        result_set: ShareableResultSet,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) -> Vec<JoinHandle<()>> {
        self.views
            .as_slice()
            .iter()
            .map(|view_doc| {
                let db = db.clone();
                let view_doc = view_doc.clone();
                let result_set = Arc::clone(&result_set);

                info!(name: "processing view with catalog", view = ?view_doc);

                let task_semaphore = task_semaphore.clone();
                tokio::runtime::Handle::current().spawn(async move {
                    // Acquire a permit from the semaphore to limit concurrency
                    #[allow(clippy::unwrap_used)]
                    let _permit = task_semaphore.acquire().await.unwrap();
                    let namespace_info = NamespaceInfo {
                        db_name: db.name().to_string(),
                        coll_or_view_name: view_doc.name.clone(),
                        namespace_type: NamespaceType::View,
                    };

                    // In dry_run mode, there is no need to derive schema for a view. Instead,
                    // we send just the namespace info and return.
                    if dry_run {
                        debug!(
                            "Received namespace-only info for {}.{}",
                            namespace_info.db_name, namespace_info.coll_or_view_name
                        );
                        result_set
                            .write()
                            .await
                            .mark_as_changed(namespace_info.db_name, namespace_info.coll_or_view_name);

                        return;
                    }

                    info!(db = db.name(), collection = view_doc.name, "getting schema for view");

                    let Some(ref view_options) = view_doc.options.view_options else {
                        unreachable!("process_views_with_catalog was a passed a CollectionDoc without view options.")
                    };

                    if let Ok(pipeline) = view_options
                        .pipeline
                        .iter()
                        .map(|doc| bson::from_document(doc.clone()))
                        .try_fold(Vec::new(), |mut acc, doc| match doc {
                            Ok(stage) => {
                                acc.push(stage);
                                Ok(acc)
                            }
                            Err(e) => Err(e),
                        }) {

                        // Get the catalog for the view the database is in
                        // and use it for schema derivation
                        let namespaces = schema_derivation::get_namespaces_for_pipeline(pipeline.clone(), db.name().to_string(), Some(view_options.view_on.clone()));

                        let schemas = {
                            let guard = result_set.read().await;
                            guard.get_schemas_for_namespaces(&namespaces)
                        };
                        match schemas {
                            None => {
                                // Fall back to the old sampling method if no schema exists
                                fallback_view_task(
                                    &view_doc,
                                    &db,
                                    namespace_info,
                                    Arc::clone(&result_set),
                                    format!(
                                        "No schema found for underlying collection {} backing the view {}.\n Falling back to sampling.",
                                       view_options.view_on,
                                        view_doc.name
                                    )
                                ).await;
                            }
                            Some(collection_info_with_schema) => {
                                // Use the catalog and the pipeline to derive the view schema
                                let catalog = collection_info_with_schema.iter().fold(
                                    BTreeMap::new(),
                                    |mut acc, (key, value)| {
                                        acc.insert(
                                            Namespace::new(db.name().to_string(), key.to_string()),
                                            Schema::clone(value),
                                        );
                                        acc
                                    },
                                );
                                let mut state = ResultSetState::new(&catalog, db.name().to_string());

                                match derive_schema_for_pipeline(
                                    pipeline,
                                    Some(view_options.view_on.clone()),
                                    &mut state,
                                ) {
                                    Ok(schema) => {
                                        info!(db = db.name(), collection = view_doc.name, "Successfully derived schema for view");
                                        result_set.write().await.add_schema(NamespaceInfoWithSchema {
                                            namespace_info,
                                            namespace_schema: Arc::new(schema),
                                        });
                                    }
                                    Err(e) => {
                                        fallback_view_task(
                                            &view_doc,
                                            &db,
                                            namespace_info,
                                            Arc::clone(&result_set),
                                            format!(
                                                "Failed to derive schema from pipeline: {e}.\n Falling back to sampling."
                                            )
                                        ).await;
                                    }
                                }
                            }
                        }
                    } else {
                        fallback_view_task(
                            &view_doc,
                            &db,
                            namespace_info,
                            Arc::clone(&result_set),
                            "Unable to parse view pipeline, falling back to sampling".to_string(),
                        ).await;
                    }
                })
            })
            .collect()
    }
}

/// Fallback function for processing views that don't have a schema in the catalog,
/// or for view that we weren't able to derive a schema for.
async fn fallback_view_task(
    view_doc: &CollectionDoc,
    db: &Database,
    namespace_info: NamespaceInfo,
    result_set: ShareableResultSet,
    reason: String,
) {
    warn!(db = db.name(), collection = view_doc.name, "{reason}");
    match derive_schema_for_view(view_doc, db).await {
        None => warn!(
            db = db.name(),
            collection = view_doc.name,
            "no schema derived, view may be empty"
        ),
        Some(schema) => {
            result_set
                .write()
                .await
                .add_schema(NamespaceInfoWithSchema {
                    namespace_info,
                    namespace_schema: Arc::new(schema),
                });
        }
    };
}
