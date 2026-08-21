use std::future::Future;

#[derive(Debug, PartialEq, Eq, thiserror::Error)]
pub enum MarkerFetchError {}

/// Token Provider
///
/// This trait signals a way to fetch the entitlement token for a cluster
pub trait MarkerProvider {
    /// Fetch the token from the cluster
    fn fetch_marker(&mut self) -> impl Future<Output = Result<impl AsRef<str>, MarkerFetchError>>;
}
