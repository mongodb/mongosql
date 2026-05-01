/**
 * This module contains functionality for processing MongoDB collections
 * and how we operate with them.
 */
use crate::{
    DataService, Error, NamespaceInfo, NamespaceInfoWithSchema, NamespaceType, Result,
    context::ContextHandle,
    data_service::CollectionInfo,
    get_partitions,
    result_set::ShareableResultSet,
    schema::{derive_schema_for_partitions, derive_schema_for_view, initial_schema::InitialSchema},
};

pub(crate) mod patterns;
use agg_ast::Namespace;
use bson::doc;
use futures::{TryStreamExt as _, future};
use mongosql::schema::Schema;
use schema_derivation::{ResultSetState, derive_schema_for_pipeline};
use std::{
    collections::{BTreeMap, HashMap},
    sync::{Arc, LazyLock, OnceLock},
};
use tracing::{debug, info, instrument, warn};

#[cfg(test)]
mod test;

static EXCLUDE_DUNDERSCORE_PATTERN: LazyLock<glob::Pattern> = LazyLock::new(|| {
    #[allow(clippy::expect_used)]
    glob::Pattern::new("__*")
        .expect("Internal error: `__*` could not be converted into a `glob::Pattern`")
});

static INCLUDE_LIST_IN_DB_AND_COLL_PAIRS: OnceLock<Vec<(String, String)>> = OnceLock::new();

/// DatabaseCollections is responsible for extracting the collections and views
/// and preparing them for processing.
#[derive(Debug, Default)]
pub(crate) struct DatabaseCollections {
    pub db: String,
    pub views: Vec<CollectionInfo>,
    pub collections: Vec<CollectionInfo>,
    pub timeseries: Vec<CollectionInfo>,
}

#[instrument(level = "trace", skip_all)]
pub(crate) async fn query_for_initial_schemas<S: DataService>(
    service: &S,
    db: &str,
    schema_collection: &str,
) -> Result<HashMap<String, Schema>, S::Error> {
    let mut initial_collection_schemas = HashMap::new();
    let cursor = service
        .find(db, schema_collection, doc! {})
        .await
        .map_err(Error::DataServiceError)?;

    // Try to parse all of the initial schemas, failing if any of them fail to
    // parse correctly
    let mut cursor = Box::pin(cursor);
    while let Some(doc) = cursor.try_next().await.map_err(Error::DataServiceError)? {
        // Convert the Doc into our InitialSchema struct
        let InitialSchema { collection, schema } = InitialSchema::try_from(doc)
            .map_err(|_| Error::InitialSchemaError(schema_collection.to_string()))?;

        // If conversion is successful, add the schema to the map
        initial_collection_schemas.insert(collection, schema);
    }

    Ok(initial_collection_schemas)
}

impl DatabaseCollections {
    /// Create a new DatabaseCollections instance. Collections and Views within a database
    /// will be enumerated, checked for inclusion/exclusion, and prepared for
    /// processing. The caller must actually process the collections/views by calling
    /// their JoinHandle.
    #[instrument(
        name = "processing collections for database",
        level = "info",
        skip(service, include_list, exclude_list)
    )]
    pub(crate) async fn new<S: DataService>(
        service: &S,
        db: String,
        include_list: Vec<glob::Pattern>,
        exclude_list: Vec<glob::Pattern>,
    ) -> Result<Self, S::Error> {
        let collection_info = service
            .list_collections(&db)
            .await
            .map_err(Error::DataServiceError)?;

        DatabaseCollections::separate_collection_types(
            db,
            &include_list,
            &exclude_list,
            collection_info,
        )
        .await
        .map_err(Into::into)
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
    pub(crate) async fn process_collections(
        &self,
        ctx: ContextHandle<impl DataService>,
        dry_run: bool,
        result_set: ShareableResultSet,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) {
        let tasks: Vec<_> = self
            .collections
            .iter()
            .chain(self.timeseries.iter())
            .cloned()
            .map(|collection_info| {
                let result_set = result_set.clone();
                let task_semaphore = task_semaphore.clone();

                info!(name: "processing collection", collection = ?collection_info);
                self.process_collection(
                    Arc::clone(&ctx),
                    dry_run,
                    result_set,
                    collection_info,
                    task_semaphore,
                )
            })
            .collect();

        // Wait for all collections to finish
        future::join_all(tasks).await;
    }

    async fn process_collection(
        &self,
        ctx: ContextHandle<impl DataService>,
        dry_run: bool,
        result_set: ShareableResultSet,
        collection_info: CollectionInfo,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) {
        // In dry_run mode, there is no need to derive schema for a view. Instead,
        // we send just the namespace info and return.
        if dry_run {
            debug!(
                "Received namespace-only info for {db}.{collection}",
                db = self.db,
                collection = collection_info.name,
            );
            result_set
                .write()
                .await
                .mark_as_changed(self.db.clone(), collection_info.name.clone());

            return;
        }

        let namespace_info = NamespaceInfo {
            db_name: self.db.clone(),
            coll_or_view_name: collection_info.name.clone(),
            namespace_type: NamespaceType::Collection,
        };

        let initial_schema = result_set
            .read()
            .await
            .get_schema_for_database(&self.db)
            .and_then(|catalog| {
                catalog
                    .get(&collection_info.name)
                    .map(|c| Arc::clone(&c.namespace_schema))
            });
        if let Some(schema) = &initial_schema {
            info!("Using initial schema: {}:{}", self.db, collection_info.name);

            if schema.is_unstable() {
                result_set
                    .write()
                    .await
                    .mark_unstable_initial_schema(self.db.clone(), collection_info.name.clone());
                info!(
                    "Found unstable initial schema for namespace `{db}.{collection}`. Marking as unstable.",
                    db = self.db,
                    collection = collection_info.name,
                );
                return;
            }
        }

        info!(
            db = self.db,
            collection = collection_info.name,
            "Getting partitions"
        );

        // If we successfully retrieve partitions from the collection,
        // derive the schema for each partition.
        let Ok(partitioned_collection) =
            get_partitions(ctx.service(), &self.db, collection_info.clone())
                .await
                .inspect_err(|e| {
                    warn!(
                        db = self.db,
                        collection = collection_info.name,
                        "could not get partitions: {e}"
                    )
                })
        else {
            return;
        };

        // Notify that we've gotten the partitions
        info!(
            db = self.db,
            collection = collection_info.name,
            "found {} partitions",
            partitioned_collection.partitions.len()
        );

        // Derive the schema for each partition, using the initial_schema
        // as the foundation for the derived schema. Then, we do a union on the initial schema
        // with the schema for each document in the partition.
        let schema = derive_schema_for_partitions(
            Arc::clone(&ctx),
            &self.db,
            &collection_info.name,
            initial_schema,
            task_semaphore.clone(),
            partitioned_collection,
        )
        .await;

        let Some(schema) = schema else {
            warn!(
                db = self.db,
                collection = collection_info.name,
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
    pub(crate) async fn process_views_with_catalog(
        &self,
        ctx: ContextHandle<impl DataService>,
        dry_run: bool,
        result_set: ShareableResultSet,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) {
        let tasks: Vec<_> = self
            .views
            .as_slice()
            .iter()
            .cloned()
            .map(|collection_info| {
                let ctx = Arc::clone(&ctx);
                let result_set = Arc::clone(&result_set);
                let task_semaphore = task_semaphore.clone();

                info!(name: "processing view with catalog", view = ?collection_info);
                self.process_view(ctx, dry_run, result_set, collection_info, task_semaphore)
            })
            .collect();

        future::join_all(tasks).await;
    }

    async fn process_view(
        &self,
        ctx: ContextHandle<impl DataService>,
        dry_run: bool,
        result_set: ShareableResultSet,
        collection_info: CollectionInfo,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) -> () {
        // Acquire a permit from the semaphore to limit concurrency
        #[allow(clippy::unwrap_used)]
        let _permit = task_semaphore.acquire().await.unwrap();
        let namespace_info = NamespaceInfo {
            db_name: self.db.clone(),
            coll_or_view_name: collection_info.name.clone(),
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

        info!(
            db = self.db,
            collection = collection_info.name,
            "getting schema for view"
        );

        if let Ok(pipeline) = collection_info
            .options
            .pipeline
            .iter()
            .map(|doc| bson::from_document::<agg_ast::definitions::Stage>(doc.clone()))
            .collect::<core::result::Result<Vec<_>, _>>()
        {
            // Get the catalog for the view the database is in
            // and use it for schema derivation
            let namespaces = schema_derivation::get_namespaces_for_pipeline(
                pipeline.clone(),
                self.db.clone(),
                Some(collection_info.options.view_on.clone()),
            );

            let schemas = {
                let guard = result_set.read().await;
                guard.get_schemas_for_namespaces(&namespaces)
            };
            match schemas {
                None => {
                    // Fall back to the old sampling method if no schema exists
                    fallback_view_task(
                        ctx.service(),
                        &self.db,
                        &collection_info,
                        namespace_info,
                        Arc::clone(&result_set),
                        format!(
                            "No schema found for underlying collection {} backing the view {}.\n Falling back to sampling.",
                            collection_info.options.view_on,
                            collection_info.name
                        )
                    ).await;
                }
                Some(collection_info_with_schema) if collection_info_with_schema.is_empty() => {
                    // Fall back to sampling when the underlying collection schema is empty
                    // (e.g., the collection doesn't exist)
                    fallback_view_task(
                        ctx.service(),
                        &self.db,
                        &collection_info,
                        namespace_info,
                        Arc::clone(&result_set),
                        format!(
                            "Schema for underlying collection {} is empty (collection may not exist). Falling back to sampling for view {}.",
                            collection_info.options.view_on,
                            collection_info.name
                        )
                    ).await;
                }
                Some(collection_info_with_schema) => {
                    // Use the catalog and the pipeline to derive the view schema
                    let catalog = collection_info_with_schema.iter().fold(
                        BTreeMap::new(),
                        |mut acc, (key, value)| {
                            acc.insert(
                                Namespace::new(self.db.clone(), key.to_string()),
                                Schema::clone(value),
                            );
                            acc
                        },
                    );
                    let mut state = ResultSetState::new(&catalog, self.db.clone());

                    match derive_schema_for_pipeline(
                        pipeline,
                        Some(collection_info.options.view_on.clone()),
                        &mut state,
                    ) {
                        Ok(schema) => {
                            info!(
                                db = self.db,
                                collection = collection_info.name,
                                "Successfully derived schema for view"
                            );
                            result_set
                                .write()
                                .await
                                .add_schema(NamespaceInfoWithSchema {
                                    namespace_info,
                                    namespace_schema: Arc::new(schema),
                                });
                        }
                        Err(e) => {
                            fallback_view_task(
                                ctx.service(),
                                &self.db,
                                &collection_info,
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
                ctx.service(),
                &self.db,
                &collection_info,
                namespace_info,
                Arc::clone(&result_set),
                "Unable to parse view pipeline, falling back to sampling".to_string(),
            )
            .await;
        }
    }
}

/// Fallback function for processing views that don't have a schema in the catalog,
/// or for view that we weren't able to derive a schema for.
async fn fallback_view_task(
    service: &impl DataService,
    db: &str,
    collection_info: &CollectionInfo,
    namespace_info: NamespaceInfo,
    result_set: ShareableResultSet,
    reason: String,
) {
    warn!(db, collection = collection_info.name, "{reason}");
    match derive_schema_for_view(service, db, collection_info).await {
        None => warn!(
            db,
            collection = collection_info.name,
            "no schema derived for {db}.{collection}, view may be empty",
            collection = collection_info.name
        ),
        Some(Schema::Document(ref d)) if d.keys.is_empty() && !d.additional_properties => {
            warn!(
                db,
                collection = collection_info.name,
                "sampled view schema is empty for {db}.{collection}; the underlying collection may not exist, no schema will be generated",
                collection = collection_info.name
            );
        }
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
