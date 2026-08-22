use std::future::Future;

use jsonwebtoken::jwk::{Jwk, JwkSet};

/// JWKS Provider
///
/// This trait signals a way to fetch the upstream JWKS at the specified
/// location.
pub trait JwksProvider {
    type Error: core::error::Error;

    /// Fetch the current JWKS
    fn fetch_jwks(&mut self) -> impl Future<Output = Result<JwkSet, Self::Error>>;

    /// Fetch the key for a `kid`, refreshing on a miss to handle mid-rotation.
    /// The default performs a single lookup; caching providers should override
    /// to re-fetch before reporting the key as absent.
    fn fetch_key(&mut self, kid: &str) -> impl Future<Output = Result<Option<Jwk>, Self::Error>> {
        async move { Ok(self.fetch_jwks().await?.find(kid).cloned()) }
    }
}
