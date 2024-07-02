use crate::{SamplerNotification, SchemaResult};
#[derive(Debug, Clone)]
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
    pub tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
    /// The schema channel
    pub tx_schemata: tokio::sync::mpsc::UnboundedSender<SchemaResult>,
}
