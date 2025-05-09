use std::collections::HashMap;

use tokio::sync::oneshot;

use crate::{NamespaceInfoWithSchema, Result};

use super::actor::{ResultSetActorHandle, ResultSetActorMessage};

/// Request a specific schema from the ResultSetActor
pub async fn get_schema(
    actor: &ResultSetActorHandle,
    db_name: String,
) -> Result<Option<HashMap<String, NamespaceInfoWithSchema>>> {
    let (send, recv) = oneshot::channel();
    actor
        .send(ResultSetActorMessage::GetSchema {
            db_name,
            respond_to: send,
        })
        .map_err(|_| crate::Error::ChannelClosed("Failed to send GetSchema message".into()))?;

    recv.await
        .map_err(|_| crate::Error::ChannelClosed("Failed to receive schema response".into()))
}
