use std::{future::Future, path::Path};

/// A file client
pub trait FileClient {
    type Error: core::error::Error;

    /// Attempt to read JSON from a supplied path
    fn read_json_from(
        &self,
        path: impl AsRef<Path>,
    ) -> impl Future<Output = Result<serde_json::Value, Self::Error>>;
}
