use std::{
    path::PathBuf,
    time::{Duration, Instant},
};

use jsonwebtoken::jwk::{Jwk, JwkSet};

mod clock;
mod jwks;

pub use clock::*;
pub use jwks::*;

use crate::{
    clients::{FileClient, HttpsClient},
    JWKS_WELL_KNOWN_URL,
};

/// TTL for how long to cache a JWKS
const JWKS_TTL: Duration = Duration::from_hours(24);

/// An error that can occur during JWKS caching
#[derive(Debug, thiserror::Error)]
pub enum CachedError<H: core::error::Error, F: core::error::Error> {
    #[error("could not fetch JWKS remotely: {0}")]
    HttpsClient(H),

    #[error("could not read JWKS from file")]
    FileClient(F),

    #[error("could not decode JWKS from JSON: {0}")]
    Json(#[from] serde_json::Error),
}

/// An implementation of the JwksProvider that caches subsequent requests.
///
/// This first attempts to fetch the JWKS from the well-known upstream URL,
/// but falls back to reading from a user-provided path if that fails.
/// The cache lives on the instance, so callers must retain the provider
/// across validations for the TTL to have any effect.
pub struct NetworkedCachedJwksProvider<H, F> {
    /// Optional path to a JSON file containing the JWKS to use
    path: Option<PathBuf>,

    /// The cached JWKS
    jwks: Option<(JwkSet, Instant)>,

    /// An implementation of [HttpsClient] for use with fetching a remote JWKS
    https_client: H,

    /// An implementation of [FileClient] for use with reading a JWKS from file
    file_client: F,
}

impl<H: HttpsClient, F: FileClient> NetworkedCachedJwksProvider<H, F> {
    pub fn new(path: Option<PathBuf>, https_client: H, file_client: F) -> Self {
        Self {
            path,
            https_client,
            file_client,

            // Initialize the cache as empty, will lazily resolve as needed
            jwks: None,
        }
    }

    /// Attempt to fetch a JwkSet from the well-known upstream URL
    async fn fetch_from_upstream(&mut self) -> Result<JwkSet, CachedError<H::Error, F::Error>> {
        let jwks = self
            .https_client
            .fetch_json(JWKS_WELL_KNOWN_URL)
            .await
            .map_err(CachedError::HttpsClient)?;

        let result: JwkSet = serde_json::from_value(jwks)?;

        // Update the cache
        self.jwks = Some((result.clone(), Instant::now()));

        Ok(result)
    }

    /// Attempt to read a JwkSet from the supplied path
    async fn read_jwks_from_path(
        &mut self,
        path: PathBuf,
    ) -> Result<JwkSet, CachedError<H::Error, F::Error>> {
        let jwks = self
            .file_client
            .read_json_from(path)
            .await
            .map_err(CachedError::FileClient)?;

        let result: JwkSet = serde_json::from_value(jwks)?;

        // Update the cache
        self.jwks = Some((result.clone(), Instant::now()));

        Ok(result)
    }

    /// Backdate the cached entry so it reads as expired, exercising the refresh path
    #[cfg(test)]
    fn expire_cache(&mut self) {
        if let Some((_, atime)) = &mut self.jwks {
            *atime -= JWKS_TTL;
        }
    }
}

impl<H: HttpsClient, F: FileClient> NetworkedCachedJwksProvider<H, F> {
    /// Refresh the JWKS, ignoring any cached entry: upstream, then file, then
    /// falling back to the last-known-good cache on failure
    async fn refresh(&mut self) -> Result<JwkSet, CachedError<H::Error, F::Error>> {
        // Always attempt to fetch from upstream first
        let networked_jwks = self.fetch_from_upstream().await;
        if let Ok(jwks) = networked_jwks {
            return Ok(jwks);
        }

        // If that failed, then we try the file path
        let refresh = if let Some(path) = &self.path {
            match self.read_jwks_from_path(path.clone()).await {
                Ok(jwks) => return Ok(jwks),
                file_err => file_err,
            }
        } else {
            networked_jwks
        };

        // If neither refresh path worked, serve the last-known-good keys rather
        // than failing closed on a transient outage (keys rotate rarely)
        if let Some((jwks, _)) = &self.jwks {
            return Ok(jwks.clone());
        }

        // If we have never cached anything, bubble up the most recent refresh error
        refresh
    }
}

impl<H: HttpsClient, F: FileClient> JwksProvider for NetworkedCachedJwksProvider<H, F> {
    type Error = CachedError<H::Error, F::Error>;

    async fn fetch_jwks(&mut self) -> Result<JwkSet, Self::Error> {
        // If we have it cached and it hasn't expired yet, just short out
        let now = Instant::now();
        if let Some((jwks, atime)) = &self.jwks {
            if now.duration_since(*atime) < JWKS_TTL {
                return Ok(jwks.clone());
            }
        }

        self.refresh().await
    }

    async fn fetch_key(&mut self, kid: &str) -> Result<Option<Jwk>, Self::Error> {
        // Try the cached set first, then re-fetch once on a miss to pick up a
        // key published mid-rotation before reporting it as absent
        if let Some(jwk) = self.fetch_jwks().await?.find(kid).cloned() {
            return Ok(Some(jwk));
        }

        Ok(self.refresh().await?.find(kid).cloned())
    }
}

#[cfg(test)]
mod test {
    use std::cell::Cell;

    use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine as _};
    use ed25519_dalek::SigningKey;
    use jsonwebtoken::{
        jwk::{Jwk, JwkSet},
        Algorithm, DecodingKey,
    };
    use serde_json::json;

    use crate::{
        clients::{FileClient, HttpsClient},
        providers::{CachedError, NetworkedCachedJwksProvider},
        JWKS_WELL_KNOWN_URL,
    };

    use super::JwksProvider;

    #[derive(Debug, PartialEq, Eq, thiserror::Error)]
    enum TestError {
        #[error("Client is intentionally empty")]
        IntentionallyEmpty,
    }
    struct TestFileClient(Option<serde_json::Value>);
    impl FileClient for TestFileClient {
        type Error = TestError;

        async fn read_json_from(
            &self,
            _path: impl AsRef<std::path::Path>,
        ) -> Result<serde_json::Value, Self::Error> {
            self.0.clone().ok_or(TestError::IntentionallyEmpty)
        }
    }

    struct TestHttpsClient(Option<serde_json::Value>);
    impl HttpsClient for TestHttpsClient {
        type Error = TestError;

        async fn fetch_json(&self, url: &str) -> Result<serde_json::Value, Self::Error> {
            if url != JWKS_WELL_KNOWN_URL {
                return Ok(json!({}));
            }

            self.0.clone().ok_or(TestError::IntentionallyEmpty)
        }
    }

    /// Serves the JWKS on the first fetch, then fails, simulating a transient outage
    struct OnceThenFailHttpsClient {
        jwks: serde_json::Value,
        served: Cell<bool>,
    }
    impl HttpsClient for OnceThenFailHttpsClient {
        type Error = TestError;

        async fn fetch_json(&self, url: &str) -> Result<serde_json::Value, Self::Error> {
            if url != JWKS_WELL_KNOWN_URL || self.served.replace(true) {
                return Err(TestError::IntentionallyEmpty);
            }

            Ok(self.jwks.clone())
        }
    }

    /// Serves a different JWKS on each fetch, simulating a key rotation between calls
    struct RotatingHttpsClient {
        responses: Vec<serde_json::Value>,
        call: Cell<usize>,
    }
    impl HttpsClient for RotatingHttpsClient {
        type Error = TestError;

        async fn fetch_json(&self, url: &str) -> Result<serde_json::Value, Self::Error> {
            if url != JWKS_WELL_KNOWN_URL {
                return Err(TestError::IntentionallyEmpty);
            }

            let idx = self.call.replace(self.call.get() + 1);
            Ok(self
                .responses
                .get(idx)
                .or_else(|| self.responses.last())
                .cloned()
                .unwrap_or_else(|| json!({})))
        }
    }

    fn generate_keys(seed: u8, key_id: &str) -> (SigningKey, Jwk) {
        let signing_key = SigningKey::from_bytes(&[seed; 32]);
        let verifying_key = signing_key.verifying_key();
        let public_key = verifying_key.as_bytes();
        let decoding_key = DecodingKey::from_ed_der(public_key);

        let jwk = {
            let mut result = Jwk::from_decoding_key(&decoding_key, Some(Algorithm::EdDSA))
                .expect("test JWK should parse");
            result.common.key_id = Some(key_id.to_string());
            result.common.public_key_use = Some(jsonwebtoken::jwk::PublicKeyUse::Signature);

            result
        };

        (signing_key, jwk)
    }

    #[tokio::test]
    async fn parses_published_well_known_jwks_shape() {
        let service_key_id = "mongosql-2025-01";
        let (service_key, service_jwk) = generate_keys(1, service_key_id);

        let emergency_key_id = "mongosql-emergency-2025-01";
        let (emergency_key, emergency_jwk) = generate_keys(2, emergency_key_id);

        let jwks = json!({
            "keys": [
                {
                    "kty": "OKP", "alg": "EdDSA", "crv": "Ed25519", "kid": service_key_id, "use": "sig",
                    "x": URL_SAFE_NO_PAD.encode(service_key.verifying_key().to_bytes())
                },
                {
                    "kty": "OKP", "alg": "EdDSA", "crv": "Ed25519", "kid": emergency_key_id, "use": "sig",
                    "x": URL_SAFE_NO_PAD.encode(emergency_key.verifying_key().to_bytes())
                }
            ]
        });

        let https_client = TestHttpsClient(Some(jwks));
        let file_client = TestFileClient(None);
        let jwks = NetworkedCachedJwksProvider::new(None, https_client, file_client)
            .fetch_jwks()
            .await
            .expect("parse json JWKS");

        assert_eq!(
            *jwks.find(service_key_id).expect("service key"),
            service_jwk
        );
        assert_eq!(
            *jwks.find(emergency_key_id).expect("emergency key"),
            emergency_jwk
        );
    }

    #[tokio::test]
    async fn falls_back_to_local_file_when_remote_unreachable() {
        let key_id = "example";
        let (_, jwk) = generate_keys(1, key_id);

        let https_client = TestHttpsClient(None);
        let file_client = TestFileClient(Some(json!(JwkSet {
            keys: vec![jwk.clone()]
        })));
        let jwks =
            NetworkedCachedJwksProvider::new(Some("nonsense".into()), https_client, file_client)
                .fetch_jwks()
                .await
                .expect("parse json JWKS");

        assert_eq!(*jwks.find(key_id).expect("service key"), jwk);
    }

    #[tokio::test]
    async fn errors_when_neither_remote_nor_file_yields_key() {
        let https_client = TestHttpsClient(None);
        let file_client = TestFileClient(None);
        let result =
            NetworkedCachedJwksProvider::new(Some("nonsense".into()), https_client, file_client)
                .fetch_jwks()
                .await;

        assert!(
            matches!(
                result,
                Err(CachedError::FileClient(TestError::IntentionallyEmpty)),
            ),
            "empty remote / file should have errored: {result:?}"
        );
    }

    #[tokio::test]
    async fn serves_stale_cache_when_refresh_fails() {
        let key_id = "example";
        let (_, jwk) = generate_keys(1, key_id);
        let https_client = OnceThenFailHttpsClient {
            jwks: json!(JwkSet {
                keys: vec![jwk.clone()]
            }),
            served: Cell::new(false),
        };
        let mut provider =
            NetworkedCachedJwksProvider::new(None, https_client, TestFileClient(None));

        // Prime the cache from upstream, then expire it and force a failing refresh
        provider.fetch_jwks().await.expect("initial fetch");
        provider.expire_cache();
        let jwks = provider
            .fetch_jwks()
            .await
            .expect("stale cache should be served when refresh fails");

        assert_eq!(*jwks.find(key_id).expect("cached key"), jwk);
    }

    #[tokio::test]
    async fn refetches_on_kid_miss_to_pick_up_rotated_key() {
        let old_kid = "old";
        let new_kid = "new";
        let (_, old_jwk) = generate_keys(1, old_kid);
        let (_, new_jwk) = generate_keys(2, new_kid);
        let https_client = RotatingHttpsClient {
            responses: vec![
                json!(JwkSet {
                    keys: vec![old_jwk.clone()]
                }),
                json!(JwkSet {
                    keys: vec![old_jwk.clone(), new_jwk.clone()]
                }),
            ],
            call: Cell::new(0),
        };
        let mut provider =
            NetworkedCachedJwksProvider::new(None, https_client, TestFileClient(None));

        // The fresh cache lacks the rotated key; fetch_key must re-fetch to find it
        let key = provider
            .fetch_key(new_kid)
            .await
            .expect("fetch_key")
            .expect("rotated key should be found after a re-fetch");

        assert_eq!(key, new_jwk);
    }
}
