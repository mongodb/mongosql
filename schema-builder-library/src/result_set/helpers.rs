use std::collections::{BTreeSet, HashMap};

use agg_ast::Namespace;
use tokio::sync::oneshot;

use crate::{NamespaceInfoWithSchema, Result};

use super::actor::{ResultSetActorHandle, ResultSetActorMessage};

/// Request schemas for a specific database
pub async fn get_schema(
    actor: &ResultSetActorHandle,
    db_name: String,
) -> Result<Option<HashMap<String, NamespaceInfoWithSchema>>> {
    let (send, recv) = oneshot::channel();
    actor
        .send(ResultSetActorMessage::GetSchemaForDB {
            db_name,
            respond_to: send,
        })
        .map_err(|_| crate::Error::ChannelClosed("Failed to send GetSchema message".into()))?;

    recv.await
        .map_err(|_| crate::Error::ChannelClosed("Failed to receive schema response".into()))
}

/// Request schemas for specific namespaces
pub async fn get_schemas_for_namespaces(
    actor: &ResultSetActorHandle,
    namespaces: BTreeSet<Namespace>,
) -> Result<Option<HashMap<String, NamespaceInfoWithSchema>>> {
    let (send, recv) = oneshot::channel();
    actor
        .send(ResultSetActorMessage::GetSchemaForNamespaces {
            namespaces,
            respond_to: send,
        })
        .map_err(|_| {
            crate::Error::ChannelClosed("Failed to send GetSchemaForCollections message".into())
        })?;

    recv.await
        .map_err(|_| crate::Error::ChannelClosed("Failed to receive schema response".into()))
}
