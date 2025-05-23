use crate::{NamespaceInfoWithSchema, Result, SchemaResult};
use agg_ast::Namespace;
use std::collections::{BTreeSet, HashMap};
use tokio::sync::{
    mpsc::{unbounded_channel, UnboundedReceiver, UnboundedSender},
    oneshot,
};
use tracing::{debug, instrument};

/// Messages that can be sent to the ResultSetActor
#[derive(Debug)]
pub enum ResultSetActorMessage {
    /// Add a schema result to the catalog
    AddSchema(SchemaResult),

    /// Get the schema for a specific database
    GetSchemaForDB {
        db_name: String,
        respond_to: oneshot::Sender<Option<HashMap<String, NamespaceInfoWithSchema>>>,
    },

    /// Get the schema for specific collections in a database
    GetSchemaForNamespaces {
        namespaces: BTreeSet<Namespace>,
        respond_to: oneshot::Sender<Option<HashMap<String, NamespaceInfoWithSchema>>>,
    },
}

/// ResultSetActor manages the ResultSet catalog
/// It receives messages via a channel, processes them,
/// and can respond to catalog requests.
#[derive(Debug)]
pub struct ResultSetActor {
    rx: UnboundedReceiver<ResultSetActorMessage>,
    catalog: HashMap<String, HashMap<String, NamespaceInfoWithSchema>>,
    forward_channels: Vec<UnboundedSender<SchemaResult>>,
}

// Add a type alias for the sender
pub type ResultSetActorHandle = UnboundedSender<ResultSetActorMessage>;

impl ResultSetActor {
    /// Create a new ResultSetActor and return its handle
    pub fn new(
        forward_channels: Vec<UnboundedSender<SchemaResult>>,
    ) -> (Self, ResultSetActorHandle) {
        let (sender, receiver) = unbounded_channel();
        (
            Self {
                rx: receiver,
                catalog: HashMap::new(),
                forward_channels,
            },
            sender,
        )
    }

    /// Get a schema for a database
    fn get_schema_for_db(
        &self,
        db_name: &str,
    ) -> Option<&HashMap<String, NamespaceInfoWithSchema>> {
        self.catalog.get(db_name)
    }

    /// Get schema for a database with a filter for specific collections/views
    fn get_schema_for_namespaces(
        &self,
        namespaces: BTreeSet<Namespace>,
    ) -> Option<HashMap<String, NamespaceInfoWithSchema>> {
        namespaces
            .iter()
            .filter_map(|namespace| {
                self.catalog
                    .get(&namespace.database)
                    .and_then(|db_catalog| db_catalog.get(&namespace.collection))
                    .map(|schema| (namespace.collection.clone(), schema.clone()))
            })
            .collect::<HashMap<_, _>>()
            .into()
    }

    /// Process a schema result
    fn process_schema_result(&mut self, result: SchemaResult) {
        match result {
            SchemaResult::NamespaceOnly(namespace_info) => {
                // In dry-run mode, we only receive namespace info without schemas
                debug!(
                    "Received namespace-only info for {}.{}",
                    namespace_info.db_name, namespace_info.coll_or_view_name
                );
            }
            SchemaResult::FullSchema(namespace_info_with_schema) => {
                let db_name = namespace_info_with_schema.namespace_info.db_name.clone();
                let coll_or_view_name = namespace_info_with_schema
                    .namespace_info
                    .coll_or_view_name
                    .clone();

                debug!("Adding schema for {}.{}", db_name, coll_or_view_name);
                let db_catalog = self.catalog.entry(db_name.clone()).or_default();
                db_catalog.insert(coll_or_view_name.clone(), namespace_info_with_schema);
            }
            // This is a special case for initial schema. It is up to the caller to decide if
            // they are interested in this schema or not.
            SchemaResult::InitialSchema(namespace_info_with_schema) => {
                let db_name = namespace_info_with_schema.namespace_info.db_name.clone();
                let coll_or_view_name = namespace_info_with_schema
                    .namespace_info
                    .coll_or_view_name
                    .clone();

                debug!(
                    "Adding initial schema for {}.{}",
                    db_name, coll_or_view_name
                );
                let db_catalog = self.catalog.entry(db_name.clone()).or_default();
                db_catalog.insert(coll_or_view_name.clone(), namespace_info_with_schema);
            }
        }
    }

    /// Start the actor's processing loop
    #[instrument(name = "resultset_actor", level = "debug", skip(self))]
    pub async fn run(mut self) {
        debug!("ResultSetActor started");

        while let Some(msg) = self.rx.recv().await {
            match msg {
                ResultSetActorMessage::AddSchema(schema_result) => {
                    // If this is a dry-run mode, we only receive namespace info without schemas
                    if let SchemaResult::NamespaceOnly(namespace_info) = &schema_result {
                        for channel in &self.forward_channels {
                            let _ =
                                channel.send(SchemaResult::NamespaceOnly(namespace_info.clone()));
                        }
                    }
                    // Process the schema normally
                    if let SchemaResult::FullSchema(ref schema_info) = schema_result {
                        // Forward to all registered channels before processing
                        for channel in &self.forward_channels {
                            let _ = channel.send(SchemaResult::FullSchema(schema_info.clone()));
                        }
                    }
                    self.process_schema_result(schema_result);
                }
                ResultSetActorMessage::GetSchemaForDB {
                    db_name,
                    respond_to,
                } => {
                    let schema = self.get_schema_for_db(&db_name).cloned();
                    let _ = respond_to.send(schema);
                }
                ResultSetActorMessage::GetSchemaForNamespaces {
                    namespaces,
                    respond_to,
                } => {
                    let schema = self.get_schema_for_namespaces(namespaces);
                    let _ = respond_to.send(schema);
                }
            }
        }

        debug!("ResultSetActor stopped");
    }
}

/// Extension trait for SchemaResult channels
pub trait SchemaResultSender {
    /// Send a schema result to the ResultSetActor
    fn send_schema(&self, result: SchemaResult) -> Result<()>;
}

impl SchemaResultSender for UnboundedSender<ResultSetActorMessage> {
    fn send_schema(&self, result: SchemaResult) -> Result<()> {
        self.send(ResultSetActorMessage::AddSchema(result))
            .map_err(|_| {
                crate::Error::ChannelClosed("Failed to send schema to ResultSetActor".into())
            })
    }
}
