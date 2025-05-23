use std::collections::{BTreeSet, HashMap};

use crate::{NamespaceInfoWithSchema, Result, SchemaResult};
use actor::{ResultSetActor, ResultSetActorHandle, SchemaResultSender};
use agg_ast::Namespace;
use tokio::sync::mpsc::{unbounded_channel, UnboundedReceiver, UnboundedSender};
use tracing::error;
pub mod actor;
pub mod helpers;

/// ResultSet provides access to the catalog of schemas managed by a ResultSetActor
#[derive(Debug)]
pub struct ResultSet {
    actor_handle: ResultSetActorHandle,
}

impl ResultSet {
    /// Creates a new ResultSet that interfaces with a ResultSetActor
    pub fn new(
        forwards_channels: Vec<UnboundedSender<SchemaResult>>,
    ) -> (Self, UnboundedSender<SchemaResult>) {
        let (actor, actor_handle) = ResultSetActor::new(forwards_channels);
        let (tx_schemas, rx_schemas) = unbounded_channel();

        // Spawn the actor task
        tokio::spawn(actor.run());

        // Spawn a task to forward messages from tx_schemas to the actor
        tokio::spawn(ResultSet::forward_schemas(rx_schemas, actor_handle.clone()));

        (Self { actor_handle }, tx_schemas)
    }

    /// Forward schema results from the tx_schemas channel to the actor
    async fn forward_schemas(mut rx: UnboundedReceiver<SchemaResult>, actor: ResultSetActorHandle) {
        while let Some(schema) = rx.recv().await {
            if let Err(e) = actor.send_schema(schema) {
                error!("Failed to forward schema result to actor: {}", e);
            }
        }
    }

    /// Get a specific schema by database and collection/view name
    pub async fn get_schema_for_database(
        &self,
        db_name: String,
    ) -> Result<Option<HashMap<String, NamespaceInfoWithSchema>>> {
        helpers::get_schema(&self.actor_handle, db_name).await
    }

    /// Get schemas for specific collections in a database
    pub async fn get_schemas_for_namespaces(
        &self,
        namespaces: BTreeSet<Namespace>,
    ) -> Result<Option<HashMap<String, NamespaceInfoWithSchema>>> {
        helpers::get_schemas_for_namespaces(&self.actor_handle, namespaces).await
    }
}
