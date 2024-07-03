/**
 * This module contains functionality for processing MongoDB collections
 * and how we operate with them.
 */
use crate::{
    consts::DISALLOWED_COLLECTION_NAMES, derive_schema_for_partitions, derive_schema_for_view,
    get_partitions, notify, NamespaceType, Result, SamplerAction, SamplerNotification,
    SchemaResult,
};
use futures::TryStreamExt;
use mongodb::{
    bson::{self, doc, Document},
    error, Cursor, Database,
};
use serde::{Deserialize, Serialize};
use tokio::task::JoinHandle;
use tracing::{info, instrument};
#[cfg(test)]
mod test;

/// CollectionInfo is responsible for extracting the collections and views
/// and preparing them for processing.
#[derive(Debug, Default)]
pub(crate) struct CollectionInfo {
    views: Vec<CollectionDoc>,
    collections: Vec<CollectionDoc>,
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
        include_list: Vec<String>,
        exclude_list: Vec<String>,
    ) -> Result<Self> {
        let collection_info_cursor = db
            .run_cursor_command(
                doc! { "listCollections": 1.0, "authorizedCollections": true},
                None,
            )
            .await
            .map_err(error::Error::from)?;

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
        tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
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

                info!(name: "processing collection", collection = ?coll_doc);

                tokio::runtime::Handle::current().spawn(async move {
                    // To start computing the schema for a collection, we need
                    // to determine the partitions of this collection.
                    let partitions = get_partitions(&coll).await;

                    match partitions {
                        Err(e) => {
                            // If partitioning the collection fails, there is nothing to
                            // do so we report and error and return.
                            notify!(
                                tx_notifications.as_ref(),
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: coll.name().to_string(),
                                    action: SamplerAction::Warning {
                                        message: format!("failed to partition with error {e}"),
                                    },
                                },
                            );
                        }
                        Ok(partitions) => {
                            // If partitioning succeeds, we send a notification
                            // to indicate partitioning is happening, then we
                            // derive schema for the partitions.
                            notify!(
                                tx_notifications.as_ref(),
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: coll.name().to_string(),
                                    action: SamplerAction::Partitioning {
                                        partitions: partitions.len() as u16,
                                    },
                                },
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
                                    // report an error.
                                    notify!(
                                        tx_notifications.as_ref(),
                                        SamplerNotification {
                                            db: db.name().to_string(),
                                            collection_or_view: coll.name().to_string(),
                                            action: SamplerAction::Warning {
                                                message: format!(
                                                    "failed to derive schema with error {e}"
                                                ),
                                            },
                                        },
                                    );
                                }
                                Ok(None) => {
                                    // If no schema is derived, then there is
                                    // nothing to do so we report a warning.
                                    notify!(
                                        tx_notifications,
                                        SamplerNotification {
                                            db: db.name().to_string(),
                                            collection_or_view: coll.name().to_string(),
                                            action: SamplerAction::Warning {
                                                message:
                                                    "no schema derived, collection may be empty"
                                                        .to_string()
                                            },
                                        }
                                    );
                                }
                                Ok(Some(coll_schema)) => {
                                    // If deriving schema succeeds, we send
                                    // the schema over the schemata channel.
                                    let _ = tx_schemata.send(SchemaResult {
                                        db_name: db.name().to_string(),
                                        coll_or_view_name: coll.name().to_string(),
                                        namespace_type: NamespaceType::Collection,
                                        namespace_schema: coll_schema,
                                    });
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
    #[instrument(skip_all)]
    pub(crate) fn process_views(
        &self,
        db: &Database,
        tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
    ) -> Vec<JoinHandle<()>> {
        self.views
            .as_slice()
            .iter()
            .map(|view_doc| {
                let db = db.clone();
                let view_doc = view_doc.clone();
                let tx_notifications = tx_notifications.clone();
                let tx_schemata = tx_schemata.clone();

                info!(name: "processing view", view = ?view_doc);

                tokio::runtime::Handle::current().spawn(async move {
                    let view_doc = view_doc.clone();
                    // Since view schemas depend on sampling, this is a
                    // straightforward task: simply await the result of schema
                    // derivation and send it when it's done.
                    match derive_schema_for_view(&view_doc, &db, tx_notifications.clone()).await {
                        None => notify!(
                            tx_notifications,
                            SamplerNotification {
                                db: db.name().to_string(),
                                collection_or_view: view_doc.name.clone(),
                                action: SamplerAction::Warning {
                                    message: "no schema derived, view may be empty".to_string()
                                }
                            }
                        ),
                        Some(schema) => {
                            let _ = tx_schemata.send(SchemaResult {
                                db_name: db.name().to_string(),
                                coll_or_view_name: view_doc.name.clone(),
                                namespace_type: NamespaceType::View,
                                namespace_schema: schema,
                            });
                        }
                    }
                    drop(tx_notifications);
                    drop(tx_schemata);
                })
            })
            .collect()
    }

    /// process_inclusion filters an input collectiondoc by the include_list and
    /// exclude_list.
    /// First, it filters the input collection_list by the include_list, retaining
    /// items that are in the include_list.
    /// Second, it filters the collection_list by the exclude_list, removing items
    /// that are in the exclude_list.
    /// Lastly, it filters out any collections that are in the disallowed list.
    ///
    /// Glob syntax is supported, i.e. mydb.* will match all collections in mydb.
    #[instrument(level = "trace")]
    fn should_consider(
        database: &str,
        collection_or_view: &CollectionDoc,
        include_list: &[String],
        exclude_list: &[String],
    ) -> bool {
        (include_list.is_empty()
            || include_list.iter().any(|i| {
                glob::Pattern::new(i)
                    .unwrap()
                    .matches(&format!("{database}.{}", collection_or_view.name.as_str()))
            }))
            && !exclude_list.iter().any(|e| {
                glob::Pattern::new(e)
                    .unwrap()
                    .matches(&format!("{database}.{}", collection_or_view.name.as_str()))
            })
            && !DISALLOWED_COLLECTION_NAMES.contains(&collection_or_view.name.as_str())
    }

    #[instrument(level = "trace")]
    async fn separate_views_from_collections(
        database: &str,
        include_list: &[String],
        exclude_list: &[String],
        mut collection_doc: Cursor<Document>,
    ) -> Result<CollectionInfo> {
        let mut collection_info = CollectionInfo::default();
        while let Some(collection_doc) = collection_doc.try_next().await.unwrap() {
            let collection_doc: CollectionDoc =
                bson::from_bson(bson::Bson::Document(collection_doc)).unwrap();
            if CollectionInfo::should_consider(
                database,
                &collection_doc,
                include_list,
                exclude_list,
            ) {
                if collection_doc.type_ == "view" {
                    collection_info.views.push(collection_doc);
                } else {
                    collection_info.collections.push(collection_doc);
                }
            }
        }

        Ok(collection_info)
    }
}
