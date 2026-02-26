/**
 * This module contains functionality for processing MongoDB collections
 * and how we operate with them.
 */
use crate::{
    client_util::DatabaseExt, derive_schema_for_partitions, derive_schema_for_view, get_partitions,
    notify, notify_info, notify_warning, result_set::ResultSet, Error, NamespaceInfo,
    NamespaceInfoWithSchema, NamespaceType, Result, SamplerAction, SamplerNotification,
    SchemaResult,
};
mod patterns;
use agg_ast::Namespace;
use futures::TryStreamExt;
use mongodb::{
    bson::{self, doc, Document},
    Database,
};
use schema_derivation::{derive_schema_for_pipeline, ResultSetState};
use serde::{Deserialize, Serialize};
use std::{
    collections::{BTreeMap, HashMap},
    sync::{Arc, LazyLock, OnceLock},
};
use tokio::{sync::mpsc::UnboundedSender, task::JoinHandle};
use tracing::{info, instrument};

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
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq, Default)]
pub(crate) struct CollectionDoc {
    #[serde(rename = "type")]
    pub type_: String,
    pub name: String,
    pub options: ViewOptions,
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq, Default)]
pub(crate) struct ViewOptions {
    #[serde(rename = "viewOn", default)]
    pub view_on: String,
    #[serde(default)]
    pub pipeline: Vec<Document>,
}

#[instrument(level = "trace", skip_all)]
pub(crate) async fn query_for_initial_schemas(
    schema_collection: Option<String>,
    database: &Database,
) -> Result<HashMap<String, (Document, bool)>> {
    let mut initial_collection_schemas = HashMap::new();
    if let Some(schema_coll) = schema_collection {
        let coll = database.collection::<Document>(schema_coll.as_str());
        let mut cursor = coll.find(doc! {}).await?;
        while let Some(doc) = cursor.try_next().await? {
            // If the unstable flag is not present or is present and is not a boolean, assume it is
            // false (i.e., this schema is stable)
            let unstable = doc
                .get("unstable")
                .unwrap_or(&bson::Bson::Boolean(false))
                .as_bool()
                .unwrap_or(false);
            if let (Some(bson::Bson::String(collection_name)), Some(bson::Bson::Document(schema))) =
                (doc.get("_id"), doc.get("schema"))
            {
                initial_collection_schemas
                    .insert(collection_name.to_string(), (schema.clone(), unstable));
            } else {
                return Err(Error::InitialSchemaError(schema_coll));
            }
        }
    };
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

        CollectionInfo::separate_views_from_collections(
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
        tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
        result_set: Arc<ResultSet>,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) -> Vec<JoinHandle<()>> {
        self.collections
            .as_slice()
            .iter()
            .map(|collection_doc| {
                let db = db.clone();
                let collection_doc = collection_doc.clone();
                let tx_notifications = tx_notifications.clone();
                let tx_schemata = tx_schemata.clone();
                let result_set = result_set.clone();

                info!(name: "processing collection", collection = ?collection_doc);

                let task_semaphore = task_semaphore.clone();
                // Spawn the collection task.
                tokio::runtime::Handle::current().spawn(async move {
                    let db_name = db.name();
                    let collection_name = collection_doc.name.clone();

                    let namespace_info = NamespaceInfo {
                        db_name: db_name.to_string(),
                        coll_or_view_name: collection_name.clone(),
                        namespace_type: NamespaceType::Collection,
                    };

                    if dry_run {
                        let _ = tx_schemata.send(SchemaResult::NamespaceOnly(namespace_info));
                        return;
                    }

                    let initial_schema = match result_set.get_schema_for_database(db_name.to_string()).await {
                        Ok(Some(schemas)) => schemas
                            .get(&collection_name)
                            .map(|schema| schema.namespace_schema.clone()),
                        _ => None,
                    };

                    if let Some(ref initial_schema) = initial_schema {
                        if initial_schema.is_unstable() {
                            let _ = tx_schemata.send(SchemaResult::UnstableInitialSchema(namespace_info));
                            return;
                        }
                        notify!(
                            &tx_notifications,
                            SamplerNotification {
                                db: db.name().to_string(),
                                collection_or_view: collection_doc.name.clone(),
                                action: SamplerAction::UsingInitialSchema
                            },
                        );
                    }

                    let collection = db.collection::<Document>(&collection_doc.name.clone());

                    notify!(
                        &tx_notifications,
                        SamplerNotification {
                            db: db.name().to_string(),
                            collection_or_view: collection_doc.name.clone(),
                            action: SamplerAction::Info {
                                message: "Getting partitions".to_string(),
                            },
                        },
                    );

                    match get_partitions(&collection).await {
                        // If we fail to get partitions, log a warning and ignore this collection
                        Err(e) => {
                            notify!(
                                &tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: collection_doc.name.clone(),
                                    action: SamplerAction::Warning {
                                        message: format!("Could not get partitions: {e}"),
                                    },
                                },
                            );
                        }

                        // If we successfully retrieve partitions from the collection,
                        // derive the schema for each partition.
                        Ok(partitions) => {
                            let partition_count = partitions.len();

                            // Notify that we've gotten the partitions
                            notify!(
                                &tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: collection_doc.name.clone(),
                                    action: SamplerAction::Info {
                                        message: format!("Found {partition_count} partitions"),
                                    },
                                },
                            );

                            // Derive the schema for the each partition. We can
                            // use the initial schema if we have one, but
                            // currently we just treat even collections where we
                            // already have a schema as if we're deriving their
                            // schema for the first time.
                            match derive_schema_for_partitions(
                                db.name().to_string(),
                                &collection,
                                partitions,
                                initial_schema,
                                &tokio::runtime::Handle::current(),
                                tx_notifications.clone(),
                                task_semaphore.clone(),
                            )
                                .await
                            {
                                Err(e) => {
                                    notify!(
                                        &tx_notifications,
                                        SamplerNotification {
                                            db: db.name().to_string(),
                                            collection_or_view: collection_doc.name.clone(),
                                            action: SamplerAction::Warning {
                                                message: format!("Could not derive schema: {e}"),
                                            },
                                        },
                                    );
                                }
                                Ok(schema) => {
                                    match schema {
                                        None => {
                                            notify!(
                                                &tx_notifications,
                                                SamplerNotification {
                                                    db: db.name().to_string(),
                                                    collection_or_view: collection_doc.name.clone(),
                                                    action: SamplerAction::Warning {
                                                        message: "no schema derived, collection may be empty"
                                                            .to_string()
                                                    },
                                                },
                                            );
                                        }
                                        Some(namespace_schema) => {
                                            let schema_result = NamespaceInfoWithSchema {
                                                namespace_info,
                                                namespace_schema,
                                            };

                                            // Send to both the result_set (which will store it in the catalog)
                                            // and to tx_schemata (for the caller to use)
                                            let _ = tx_schemata
                                                .send(SchemaResult::FullSchema(schema_result));
                                        }
                                    }
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
        tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
        result_set: Arc<ResultSet>,
        task_semaphore: Arc<tokio::sync::Semaphore>,
    ) -> Vec<JoinHandle<()>> {
        self.views
            .as_slice()
            .iter()
            .map(|view_doc| {
                let db = db.clone();
                let view_doc = view_doc.clone();
                let tx_notifications = tx_notifications.clone();
                let tx_schemata = tx_schemata.clone();
                let result_set = result_set.clone();

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

                    if dry_run {
                        // In dry_run mode, there is no need to derive schema for a view. Instead,
                        // we send just the namespace info and return.
                        let _ = tx_schemata.send(SchemaResult::NamespaceOnly(namespace_info));
                        return;
                    }

                    notify!(
                        &tx_notifications,
                        SamplerNotification {
                            db: db.name().to_string(),
                            collection_or_view: view_doc.name.clone(),
                            action: SamplerAction::Info {
                                message: "Getting schema for view".to_string()
                            },
                        },
                    );


                    if let Ok(pipeline) = view_doc
                        .options
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
                        let namespaces = schema_derivation::get_namespaces_for_pipeline(pipeline.clone(), db.name().to_string(), Some(view_doc.options.view_on.clone()));

                        match result_set.get_schemas_for_namespaces(namespaces).await {
                            Err(e) => {
                                notify_warning!(
                                    tx_notifications,
                                    db.name(),
                                    view_doc.name,
                                    format!(
                                        "Failed to get schema for underlying collection {} backing the view {}: {}.\n Falling
                                        back to sampling.",
                                        view_doc.options.view_on, view_doc.name, e
                                    )
                                );

                                // Fall back to the old sampling method if we can't get the schema
                                fallback_view_task(
                                    &view_doc,
                                    &db,
                                    tx_notifications.clone(),
                                    tx_schemata.clone(),
                                    namespace_info,
                                    format!(
                                        "Failed to get schema for underlying collection {} backing the view {}: {}.\n Falling back to sampling.",
                                        view_doc.options.view_on, view_doc.name, e
                                    ),
                                ).await;
                            }
                            Ok(None) => {

                                // Fall back to the old sampling method if no schema exists
                                fallback_view_task(&view_doc, &db, tx_notifications.clone(), tx_schemata.clone(), namespace_info, format!(
                                    "No schema found for underlying collection {} backing the view {}.\n Falling back to sampling.",
                                    view_doc.options.view_on, view_doc.name
                                )).await;
                            }
                            Ok(Some(collection_info_with_schema)) => {
                                // Use the catalog and the pipeline to derive the view schema
                                let catalog = collection_info_with_schema.iter().fold(
                                    BTreeMap::new(),
                                    |mut acc, (key, value)| {
                                        acc.insert(
                                            Namespace::new(db.name().to_string(), key.clone()),
                                            value.namespace_schema.clone(),
                                        );
                                        acc
                                    },
                                );
                                let mut state = ResultSetState::new(&catalog, db.name().to_string());


                                match derive_schema_for_pipeline(
                                    pipeline,
                                    Some(view_doc.options.view_on.clone()),
                                    &mut state,
                                ) {
                                    Ok(schema) => {
                                        notify_info!(
                                                    &tx_notifications,
                                                    db.name(),
                                                    view_doc.name,
                                                    format!("Successfully derived schema for view")
                                                );

                                        let _ = tx_schemata.send(SchemaResult::FullSchema(
                                            NamespaceInfoWithSchema {
                                                namespace_info,
                                                namespace_schema: schema,
                                            },
                                        ));
                                    }
                                    Err(e) => {
                                        fallback_view_task(&view_doc, &db, tx_notifications.clone(), tx_schemata.clone(), namespace_info, format!(
                                            "Failed to derive schema from pipeline: {e}.\n Falling back to sampling."
                                        )).await;
                                    }
                                }
                            }
                        }
                    } else {
                        fallback_view_task(
                            &view_doc,
                            &db,
                            tx_notifications.clone(),
                            tx_schemata.clone(),
                            namespace_info,
                            "Unable to parse view pipeline, falling back to sampling".to_string(),
                        ).await;
                    }

                    drop(tx_notifications);
                    drop(tx_schemata);
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
    tx_notifications: UnboundedSender<SamplerNotification>,
    tx_schemata: UnboundedSender<SchemaResult>,
    namespace_info: NamespaceInfo,
    reason: String,
) {
    notify_warning!(
        tx_notifications,
        db.name(),
        view_doc.name,
        reason.to_string()
    );
    match derive_schema_for_view(view_doc, db, tx_notifications.clone()).await {
        None => notify_warning!(
            tx_notifications,
            db.name(),
            view_doc.name,
            "No schema derived, view may be empty".to_string()
        ),
        Some(schema) => {
            let _ = tx_schemata.send(SchemaResult::FullSchema(NamespaceInfoWithSchema {
                namespace_info,
                namespace_schema: schema,
            }));
        }
    }
}
