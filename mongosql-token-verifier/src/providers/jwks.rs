use std::future::Future;

use jsonwebtoken::jwk::JwkSet;

/// JWKS Provider
///
/// This trait signals a way to fetch the upstream JWKS at the specified
/// location.
pub trait JwksProvider {
    type Error: core::error::Error;

    /// Fetch the current JWKS
    fn fetch_jwks(&mut self) -> impl Future<Output = Result<JwkSet, Self::Error>>;
}
