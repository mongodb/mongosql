use std::future::Future;

/// An HTTPS client
pub trait HttpsClient {
    type Error: core::error::Error;

    /// Attempt to fetch JSON from the specified URL
    fn fetch_json(&self, url: &str)
        -> impl Future<Output = Result<serde_json::Value, Self::Error>>;
}

#[cfg(feature = "reqwest")]
impl HttpsClient for reqwest::Client {
    type Error = reqwest::Error;

    async fn fetch_json(&self, url: &str) -> Result<serde_json::Value, Self::Error> {
        let jwks = self
            .get(url)
            .send()
            .await?
            .error_for_status()?
            .json()
            .await?;
        Ok(jwks)
    }
}
