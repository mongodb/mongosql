use crate::{Result, SamplerNotification};
pub struct BuilderOptions<'a> {
    /// The namespaces to include
    pub include: &'a Vec<String>,
    /// The namespaces to exclude
    pub exclude: &'a Vec<String>,
    /// The name of the schema collection
    pub schema_collection: &'a Option<String>,
    /// Whether to perform a dry run, i.e. no analysis and no writing to the database
    pub dry_run: bool,
    /// The tokio runtime handle
    pub handle: &'a tokio::runtime::Handle,
    /// The MongoDB client
    pub client: &'a mongodb::Client,
    /// The notification channel
    pub tx_notifications: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
    /// The schema channel
    pub tx_schemata: tokio::sync::mpsc::UnboundedSender<Result<crate::SchemaResult>>,
}
