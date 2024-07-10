use crate::{SamplerNotification, SchemaResult};
use std::fmt::Debug;

#[derive(Clone)]
pub struct BuilderOptions {
    /// The namespaces to include
    pub include_list: Vec<String>,
    /// The namespaces to exclude
    pub exclude_list: Vec<String>,
    /// The name of the schema collection
    pub schema_collection: Option<String>,
    /// Whether to perform a dry run, i.e. no analysis and no writing to the database
    pub dry_run: bool,
    /// The MongoDB client
    pub client: mongodb::Client,
    /// The notification channel
    pub tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
    /// The schema channel
    pub tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
}

impl Debug for BuilderOptions {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("BuilderOptions")
            .field("include_list", &self.include_list)
            .field("exclude_list", &self.exclude_list)
            .field("schema_collection", &self.schema_collection)
            .field("dry_run", &self.dry_run)
            .finish()
    }
}
