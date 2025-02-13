/**
 * This module contains functionality for processing MongoDB collections
 * and how we operate with them.
 */
use crate::{
    consts::DISALLOWED_COLLECTION_NAMES, derive_schema_for_partitions, derive_schema_for_view,
    get_partitions, notify, Error, NamespaceInfo, NamespaceInfoWithSchema, NamespaceType, Result,
    SamplerAction, SamplerNotification, SchemaResult,
};
use futures::TryStreamExt;
use mongodb::{
    bson::{self, doc, Document},
    error, Cursor, Database,
};
use serde::{Deserialize, Serialize};
use std::{
    collections::HashMap,
    sync::{Arc, LazyLock, OnceLock},
    time::Duration,
};
use tokio::{sync::RwLock, task::JoinHandle};
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
) -> Result<HashMap<String, Document>> {
    let mut initial_collection_schemas = HashMap::new();
    if let Some(schema_coll) = schema_collection {
        let coll = database.collection::<Document>(schema_coll.as_str());
        let mut cursor = coll.find(doc! {}).await?;
        while let Some(doc) = cursor.try_next().await? {
            if let (Some(bson::Bson::String(collection_name)), Some(bson::Bson::Document(schema))) =
                (doc.get("_id"), doc.get("schema"))
            {
                initial_collection_schemas.insert(collection_name.to_string(), schema.clone());
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
            .run_cursor_command(doc! { "listCollections": 1.0, "authorizedCollections": true})
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
        dry_run: bool,
        initial_schemas: Arc<RwLock<HashMap<String, HashMap<String, Document>>>>,
        tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
        tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
    ) -> Vec<JoinHandle<()>> {
        self.collections
            .as_slice()
            .iter()
            .map(|coll_doc| {
                let db = db.clone();
                let coll = db.collection::<Document>(coll_doc.name.as_str());
                let tx_notifications = tx_notifications.clone();
                let tx_schemata = tx_schemata.clone();
                let initial_schemas = initial_schemas.clone();

                info!(name: "processing collection", collection = ?coll_doc);

                tokio::runtime::Handle::current().spawn(async move {
                    let namespace_info = NamespaceInfo {
                        db_name: db.name().to_string(),
                        coll_or_view_name: coll.name().to_string(),
                        namespace_type: NamespaceType::Collection,
                    };

                    if dry_run {
                        // In dry_run mode, there is no need to partition and derive schema for a
                        // collection. Instead, we send just the namespace info and return.
                        let _ = tx_schemata.send(SchemaResult::NamespaceOnly(namespace_info));
                        return;
                    }

                    // start a heartbeat to notify the user that we are getting partitions
                    let (heartbeat_tx, heartbeat_rx) = tokio::sync::watch::channel(false);
                    let tx_hb_notifications = tx_notifications.clone();
                    let db_name = db.name().to_string();
                    let coll_name = coll.name().to_string();
                    tokio::spawn(async move {
                        let mut interval = tokio::time::interval(Duration::from_secs(60));
                        let mut h_rx = heartbeat_rx.clone();
                        let mut counter = 0;
                        loop {
                            tokio::select! {
                                // notify the user every 60 seconds that we are getting partitions
                                _ = interval.tick() => {
                                    counter += 1;
                                    notify!(&tx_hb_notifications,
                                        SamplerNotification {
                                            db: db_name.clone(),
                                            collection_or_view: coll_name.clone(),
                                            action: SamplerAction::Info {
                                                message: format!("Update {counter}: Getting partitions for collection:"),
                                            },
                                        },
                                    );
                                }
                                _ = h_rx.changed() => {
                                    break;
                                }
                            }
                        }
                    });

                    // To start computing the schema for a collection, we need
                    // to determine the partitions of this collection.
                    let partitions = get_partitions(&coll).await;
                    // Once we have the partitions, we can stop the heartbeat
                    heartbeat_tx.send(true).unwrap_or_default();

                    match partitions {
                        Err(Error::EmptyCollection(name)) => {
                            // If the collection is empty, there is nothing to do so we report a warning.
                            notify!(
                                &tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: name.clone(),
                                    action: SamplerAction::Warning {
                                        message: format!(
                                            "Collection {name} appears to be empty, skipping..."
                                        )
                                    },
                                },
                            );
                        }
                        Err(e) => {
                            // If partitioning the collection fails, there is nothing to
                            // do so we report and error and return.
                            notify!(
                                &tx_notifications,
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
                                &tx_notifications,
                                SamplerNotification {
                                    db: db.name().to_string(),
                                    collection_or_view: coll.name().to_string(),
                                    action: SamplerAction::Partitioning {
                                        partitions: partitions.len() as u16,
                                    },
                                },
                            );

                            // If there was a schema already set in the database for this
                            // collection, use that to seed the derivation of schemas
                            // for the partitions
                            let initial_schema = {
                                let read_guard = initial_schemas.read().await;
                                read_guard
                                    .get(db.name())
                                    .and_then(|colls| colls.get(&coll.name().to_string()).cloned())
                            };

                            if initial_schema.is_some() {
                                notify!(
                                    &tx_notifications,
                                    SamplerNotification {
                                        db: db.name().to_string(),
                                        collection_or_view: coll.name().to_string(),
                                        action: SamplerAction::UsingInitialSchema,
                                    },
                                );
                            }

                            // The derive_schema_for_partitions function
                            // parallelizes schema derivation per partition.
                            // So here, we await its result and then send it
                            // over the schemata channel as the final step in
                            // the collection task.
                            let coll_schema = derive_schema_for_partitions(
                                db.name().to_string(),
                                &coll,
                                partitions,
                                initial_schema,
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
                                        &tx_notifications,
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
                                    // Check for empty keys in the schema and warn the user
                                    if coll_schema.keys().iter().any(|key| key.is_empty()) {
                                        notify!(
                                            &tx_notifications,
                                            SamplerNotification {
                                                db: db.name().to_string(),
                                                collection_or_view: coll.name().to_string(),
                                                action: SamplerAction::Warning {
                                                    message: "Found empty keys".to_string()
                                                },
                                            },
                                        );
                                    }
                                    // If deriving schema succeeds, we send
                                    // the schema over the schemata channel.
                                    let _ = tx_schemata.send(SchemaResult::FullSchema(
                                        NamespaceInfoWithSchema {
                                            namespace_info,
                                            namespace_schema: coll_schema,
                                        },
                                    ));
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
        dry_run: bool,
        tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
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
                            let _ = tx_schemata.send(SchemaResult::FullSchema(
                                NamespaceInfoWithSchema {
                                    namespace_info,
                                    namespace_schema: schema,
                                },
                            ));
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
        include_list: &[glob::Pattern],
        exclude_list: &[glob::Pattern],
    ) -> Result<bool> {
        let allow_dunderscore_namespace =
            Self::should_allow_dunderscore_namespace(database, collection_or_view, include_list)?;

        Ok((include_list.is_empty()
            || include_list.iter().any(|pattern| {
                pattern.matches(&format!("{database}.{}", collection_or_view.name.as_str()))
            }))
            && (!exclude_list.iter().any(|pattern| {
                pattern.matches(&format!("{database}.{}", collection_or_view.name.as_str()))
            }))
            && (!(EXCLUDE_DUNDERSCORE_PATTERN.matches(database)
                || EXCLUDE_DUNDERSCORE_PATTERN.matches(collection_or_view.name.as_str()))
                || allow_dunderscore_namespace)
            && (!DISALLOWED_COLLECTION_NAMES.contains(&collection_or_view.name.as_str())))
    }

    /// Since we automatically exclude DBs and collections that start with dunderscores,
    /// this function is used to determine if a dunderscore namespace should be included.
    /// If a non-dunderscore namespace is passed to this function, nothing is done, and
    /// `false` is returned.
    #[instrument(level = "trace")]
    fn should_allow_dunderscore_namespace(
        database: &str,
        collection_or_view: &CollectionDoc,
        include_list: &[glob::Pattern],
    ) -> Result<bool> {
        let db_starts_with_dunderscore = database.starts_with("__");
        let coll_starts_with_dunderscore = collection_or_view.name.as_str().starts_with("__");

        // If the DB or collection starts with dunderscores, check if we should include it.
        if db_starts_with_dunderscore || coll_starts_with_dunderscore {
            // Since our test cases have different `include_lists`, we can't use the static OnceLock for testing.
            let include_list_in_db_and_coll_pairs = if cfg!(test) {
                &include_list
                    .iter()
                    .map(|pattern| {
                        let pattern_as_str = pattern.as_str();
                        let (db, collection) = pattern_as_str.split_once(".")
                            .unwrap_or_else(|| unreachable!("Internal Error: The pattern `{pattern_as_str}` is not in the format `<database_pattern>.<collection_pattern>`. However, this should have been caught earlier."));
                        (db.to_string(), collection.to_string())
                    })
                    .collect::<Vec<(String, String)>>()
            } else {
                INCLUDE_LIST_IN_DB_AND_COLL_PAIRS.get_or_init(|| {
                    include_list
                        .iter()
                        .map(|pattern| {
                            let pattern_as_str = pattern.as_str();
                            let (db, collection) = pattern_as_str.split_once(".")
                                .unwrap_or_else(|| unreachable!("Internal Error: The pattern `{pattern_as_str}` is not in the format `<database_pattern>.<collection_pattern>`. However, this should have been caught earlier."));
                            (db.to_string(), collection.to_string())
                        })
                        .collect::<Vec<(String, String)>>()
                })
            };

            let mut allow_dunderscore_namespace = false;

            for (db_pat, coll_pat) in include_list_in_db_and_coll_pairs {
                let mut allow_db = false;
                let mut allow_coll = false;

                if db_starts_with_dunderscore {
                    allow_db = Self::pattern_allows_dunderscore_name(db_pat, database)?;
                }
                if coll_starts_with_dunderscore {
                    allow_coll = Self::pattern_allows_dunderscore_name(
                        coll_pat,
                        collection_or_view.name.as_str(),
                    )?;
                }

                // The below boolean logic works by checking our three possible cases: (1) only the collection starts with dunderscores,
                // (2) only the DB starts with dunderscores, or (3) both start with dunderscores. Additionally, we only check the conditions
                // from each case that are necessary (e.g., if only the DB starts with dunderscores, only the DB must be checked that it matches a
                // dunderscore inclusion pattern).
                allow_dunderscore_namespace = allow_dunderscore_namespace
                    || ((allow_db && db_starts_with_dunderscore && !coll_starts_with_dunderscore)
                        || (allow_coll
                            && coll_starts_with_dunderscore
                            && !db_starts_with_dunderscore)
                        || (allow_coll
                            && coll_starts_with_dunderscore
                            && allow_db
                            && db_starts_with_dunderscore));
            }

            Ok(allow_dunderscore_namespace)
        } else {
            // If neither the DB or collection are prefixed with a dunderscore, there's no reason to check if we need
            // to include this namespace because it won't be automatically excluded, so we return `false`.
            Ok(false)
        }
    }

    /// This is a helper function for `should_allow_dunderscore_namespace`. This function checks
    /// if the passed `pattern_as_str` allows for the inclusion of the passed dunderscore-prefixed `name`.
    ///
    /// It works by first checking if the `pattern_as_str` fits this format:
    /// `<[<range_including_underscore>] or '_'><[<range_including_underscore>] or '_'><name or wildcard>`
    /// and then checks if it matches the provided `name`. In other words, first this functions checks
    /// if the provided pattern allows for _any_ dunderscore-prefixed name, and then it checks if
    /// it allows for the provided `name`.
    #[instrument(level = "trace")]
    fn pattern_allows_dunderscore_name(pattern_as_str: &str, name: &str) -> Result<bool> {
        // Proof: The only way in glob syntax to explicitly include characters is by using the character itself
        // or by using brackets with the character included in the brackets. Therefore, to disallow implicit inclusion
        // of dunderscore-prefixed names and allow for explicit inclusion of them, we want to ignore every
        // possible case except four:
        //
        //  (1) pattern starts with `__`, (2) pattern starts with `_[...]`,
        //  (3) pattern starts with `[...]_`, or (4) pattern starts with `[...][...]`.
        //
        // Furthermore, once we have identified one of these cases, we only know that it's possible
        // that our pattern matches our dunderscore `name`, so we then have to check if our pattern actually
        // matches our dunderscore `name`. Therefore, if our pattern fits one of the above cases
        // and matches our `name`, the pattern allows for `name`.

        // Checks cases (1) and (2)
        if pattern_as_str.starts_with("__")
            || (pattern_as_str.starts_with("_[") && !pattern_as_str.starts_with("_[!"))
        {
            let pattern = glob::Pattern::new(pattern_as_str).map_err(Error::GlobPatternError)?;
            return Ok(pattern.matches(name));
        }
        // Checks cases (3) and (4)
        else if pattern_as_str.starts_with("[") && !pattern_as_str.starts_with("[!") {
            // According to glob syntax, to include `]` as a possible character, you must put it right after the initial `[`,
            // so if we encounter `[]`, this means the user has chosen to include the `]` in the brackets as a possible character.
            // If this is the case, we need to find the next `]` because that is the one the concludes the list of characters to include.
            let close_bracket_index =
                if let Some(stripped_slice) = pattern_as_str.strip_prefix("[]") {
                    // Omit first `[]` from slice and then use the `find()` function.
                    stripped_slice.find("]").ok_or(
                        Error::InclusionBracketPatternIsMissingClosingBracket(
                            pattern_as_str.to_string(),
                        ),
                    )? + 2 // Add two to the result because we subtracted two from the length by stripping the first two characters.
                } else {
                    pattern_as_str.find("]").ok_or(
                        Error::InclusionBracketPatternIsMissingClosingBracket(
                            pattern_as_str.to_string(),
                        ),
                    )?
                };

            let slice_excluding_first_inclusion_brackets =
                &pattern_as_str[(close_bracket_index + 1)..];

            // Make sure we are only including cases `[...]_` and `[...][...]`.
            if !(slice_excluding_first_inclusion_brackets.starts_with("*")
                || slice_excluding_first_inclusion_brackets.starts_with("?")
                || slice_excluding_first_inclusion_brackets.starts_with("[!"))
            {
                let pattern =
                    glob::Pattern::new(pattern_as_str).map_err(Error::GlobPatternError)?;
                return Ok(pattern.matches(name));
            }
        }
        // If none of the above logic returns `true`, then `pattern_as_str` does not allow for the inclusion of `name`,
        // so we return `false`.
        Ok(false)
    }

    #[instrument(level = "trace")]
    async fn separate_views_from_collections(
        database: &str,
        include_list: &[glob::Pattern],
        exclude_list: &[glob::Pattern],
        mut collection_doc: Cursor<Document>,
    ) -> Result<CollectionInfo> {
        let mut collection_info = CollectionInfo::default();
        while let Some(collection_doc) = collection_doc.try_next().await? {
            if let Ok(collection_doc) = bson::from_bson(bson::Bson::Document(collection_doc)) {
                if CollectionInfo::should_consider(
                    database,
                    &collection_doc,
                    include_list,
                    exclude_list,
                )? {
                    if collection_doc.type_ == "view" {
                        collection_info.views.push(collection_doc);
                    } else {
                        collection_info.collections.push(collection_doc);
                    }
                }
            }
        }

        Ok(collection_info)
    }
}
