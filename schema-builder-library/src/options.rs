use std::fmt::Debug;

use crate::DataService;

#[derive(Clone)]
pub struct BuilderOptions<S: DataService> {
    /// The namespaces to include
    pub include_list: Vec<glob::Pattern>,
    /// The namespaces to exclude
    pub exclude_list: Vec<glob::Pattern>,
    /// The name of the schema collection where schemas are persisted
    pub schema_collection: Option<String>,
    /// Whether to perform a dry run, i.e. no analysis and no writing to the database
    pub dry_run: bool,
    /// The backend data service used to connect to MongoDB
    pub service: S,
    /// Task semaphore
    pub task_semaphore: std::sync::Arc<tokio::sync::Semaphore>,
}

impl<S: DataService> Debug for BuilderOptions<S> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("BuilderOptions")
            .field("include_list", &self.include_list)
            .field("exclude_list", &self.exclude_list)
            .field("schema_collection", &self.schema_collection)
            .field("dry_run", &self.dry_run)
            .finish()
    }
}
