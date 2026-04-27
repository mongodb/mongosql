use std::sync::Arc;

use crate::DataService;

/// Handle to a context
pub type ContextHandle<S> = Arc<Context<S>>;

/// Execution Context
///
/// This struct contains all needed handles for executing queries against
/// external services.
#[derive(Debug)]
pub struct Context<S: DataService> {
    /// The data service handle, for making outbound database requests
    service: S,
}

impl<S: DataService> Context<S> {
    pub(crate) fn new(service: S) -> Self {
        Self { service }
    }

    pub(crate) fn service(&self) -> &S {
        &self.service
    }
}
