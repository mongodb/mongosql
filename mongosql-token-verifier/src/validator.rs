use jsonwebtoken::{Algorithm, DecodingKey, Validation};

use crate::{
    marker::Claims,
    providers::{ClockProvider, JwksProvider, MarkerFetchError, MarkerProvider},
    Issuer,
};

/// An error that can occur durin marker validation
#[derive(Debug, PartialEq, Eq, thiserror::Error)]
pub enum ValidatorError<J>
where
    J: core::error::Error,
{
    #[error("marker is for cluster '{0}', but found cluster {1}")]
    ClusterMismatch(String, String),

    #[error("failed to fetch JWKS: {0}")]
    JwksFetch(J),

    #[error("marker has expired")]
    MarkerExpired,

    #[error("marker is disabled")]
    MarkerDisabled,

    #[error("temporary marker is missing an expiration time")]
    MissingExp,

    #[error("failed to fetch marker: {0}")]
    MarkerFetch(#[from] MarkerFetchError),

    #[error("JWKS is missing specified key ID: {0}")]
    MissingJwk(String),

    #[error("marker is missing the required key ID for decoding")]
    MissingKeyID,

    #[error("failed to validate: {0}")]
    Validation(#[from] jsonwebtoken::errors::Error),
}

/// A MongoDB entitlement marker validator
pub struct Validator {
    cluster: String,
}

impl Validator {
    /// Construct a validator that enforces entitlement for a specified cluster
    pub fn for_cluster(cluster: String) -> Self {
        Self { cluster }
    }

    /// Validate a marker
    pub async fn validate<M, J, C>(
        &self,
        marker: M,
        mut jwks: J,
        clock: C,
    ) -> Result<(), ValidatorError<J::Error>>
    where
        M: MarkerProvider,
        J: JwksProvider,
        C: ClockProvider,
    {
        let jwks = jwks.fetch_jwks().await.map_err(ValidatorError::JwksFetch)?;
        let token = marker.fetch_marker().await?;

        let header = jsonwebtoken::decode_header(token.as_ref())?;
        let key_id = header.kid.ok_or(ValidatorError::MissingKeyID)?;
        let key = jwks
            .find(&key_id)
            .ok_or(ValidatorError::MissingJwk(key_id))?;

        let key = DecodingKey::from_jwk(key)?;
        let validation = {
            let mut result = Validation::new(Algorithm::EdDSA);
            result.set_required_spec_claims(&["iss", "sub", "enabled"]);
            result.validate_exp = false;

            result
        };

        let decoded = jsonwebtoken::decode::<Claims>(token.as_ref(), &key, &validation)?;
        if !decoded.claims.enabled {
            return Err(ValidatorError::MarkerDisabled);
        }

        // Validate that the marker is for the specified cluster
        if decoded.claims.sub != self.cluster {
            return Err(ValidatorError::ClusterMismatch(
                decoded.claims.sub,
                self.cluster.clone(),
            ));
        }

        // Also ensure that the marker is not expired
        // Note: It would be nice if the library could be dynamically do this for us,
        // but we are kind of stuck in a catch-22 because we don't know the claims
        // until we have decoded them but we can't decode them without knowing
        // if `exp` should be there or not :(
        if decoded.claims.iss == Issuer::Emergency && decoded.claims.exp.is_none() {
            return Err(ValidatorError::MissingExp);
        }
        if let Some(exp) = decoded.claims.exp {
            let now = clock.now_unix();
            if now > exp {
                return Err(ValidatorError::MarkerExpired);
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod test {
    use std::sync::LazyLock;

    use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine};
    use ed25519_dalek::{Signer, SigningKey};
    use jsonwebtoken::{
        errors::ErrorKind,
        jwk::{Jwk, JwkSet},
        Algorithm, DecodingKey,
    };
    use serde_json::json;

    use crate::{
        providers::{ClockProvider, JwksProvider, MarkerFetchError, MarkerProvider},
        validator::{Validator, ValidatorError},
        Issuer,
    };

    const JWS_TYP: &str = "JWT";
    const JWS_ALG: &str = "EdDSA";
    const JWS_CRV: &str = "Ed25519";

    const KID: &str = "sql-interface-2026-01";
    const CLUSTER: &str = "Cluster0";
    const NOW: i64 = 1_700_000_000;
    const IAT: i64 = 1_700_000_000;

    static ISSUER_SERVICE: LazyLock<String> = LazyLock::new(|| {
        json!(Issuer::Service)
            .as_str()
            .expect("issuer to be a string")
            .to_string()
    });
    static ISSUER_EMERGENCY: LazyLock<String> = LazyLock::new(|| {
        json!(Issuer::Emergency)
            .as_str()
            .expect("issuer to be a string")
            .to_string()
    });

    struct TestJwksProvider(SigningKey);
    impl JwksProvider for TestJwksProvider {
        type Error = std::convert::Infallible;

        async fn fetch_jwks(&mut self) -> Result<JwkSet, Self::Error> {
            let verifying_key = self.0.verifying_key();
            let public_key = verifying_key.as_bytes();
            let decoding_key = DecodingKey::from_ed_der(public_key);

            let jwk = {
                let mut result = Jwk::from_decoding_key(&decoding_key, Some(Algorithm::EdDSA))
                    .expect("test JWK should parse");
                result.common.key_id = Some(KID.to_string());

                result
            };

            Ok(JwkSet { keys: vec![jwk] })
        }
    }

    struct TestClockProvider(i64);
    impl ClockProvider for TestClockProvider {
        fn now_unix(&self) -> i64 {
            self.0
        }
    }

    struct TestMarkerProvider(String);
    impl MarkerProvider for TestMarkerProvider {
        async fn fetch_marker(&self) -> Result<impl AsRef<str>, MarkerFetchError> {
            Ok(&self.0)
        }
    }

    async fn validate_marker<M, J, C>(
        cluster: &str,
        token: M,
        jwks: J,
        clock: C,
    ) -> Result<(), ValidatorError<J::Error>>
    where
        M: MarkerProvider,
        J: JwksProvider,
        C: ClockProvider,
    {
        Validator::for_cluster(cluster.to_string())
            .validate(token, jwks, clock)
            .await
    }

    fn signing_key(seed: u8) -> SigningKey {
        SigningKey::from_bytes(&[seed; 32])
    }

    fn base64_encode(value: &serde_json::Value) -> String {
        URL_SAFE_NO_PAD.encode(serde_json::to_vec(value).expect("serialize json segment"))
    }

    fn header(alg: &str, crv: Option<&str>, typ: Option<&str>, kid: &str) -> serde_json::Value {
        json!({
            "alg": alg,
            "crv": crv,
            "typ": typ,
            "kid": kid,
        })
    }

    fn claims(
        iss: &str,
        sub: &str,
        iat: i64,
        enabled: bool,
        exp: Option<i64>,
    ) -> serde_json::Value {
        json!({
            "iss": iss,
            "sub": sub,
            "iat": iat,
            "enabled": enabled,
            "exp": exp,
        })
    }

    fn sign_token(
        sk: &SigningKey,
        header: &serde_json::Value,
        claims: &serde_json::Value,
    ) -> String {
        let signing_input = format!("{}.{}", base64_encode(header), base64_encode(claims));
        let signature = sk.sign(signing_input.as_bytes());
        format!(
            "{signing_input}.{}",
            URL_SAFE_NO_PAD.encode(signature.to_bytes())
        )
    }

    fn normal_token(sk: &SigningKey) -> String {
        sign_token(
            sk,
            &header(JWS_ALG, Some(JWS_CRV), Some("jwt"), KID),
            &claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None),
        )
    }

    #[tokio::test]
    async fn valid_normal_token_without_exp_is_accepted() {
        let sk = signing_key(1);
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(normal_token(&sk)),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Ok(())
        );
    }

    #[tokio::test]
    async fn valid_emergency_token_with_future_exp_is_accepted() {
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), KID),
            &claims(&ISSUER_EMERGENCY, CLUSTER, IAT, true, Some(NOW + 3600)),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW),
            )
            .await,
            Ok(())
        );
    }

    #[tokio::test]
    async fn disabled_marker_is_rejected() {
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), KID),
            &claims(&ISSUER_SERVICE, CLUSTER, IAT, false, None),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::MarkerDisabled)
        );
    }

    #[tokio::test]
    async fn expired_emergency_token_is_rejected() {
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), KID),
            &claims(&ISSUER_EMERGENCY, CLUSTER, IAT, true, Some(NOW - 1)),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::MarkerExpired)
        );
    }

    #[tokio::test]
    async fn disallowed_issuer_is_rejected() {
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), KID),
            &claims("mongosql", CLUSTER, IAT, true, None),
        );

        validate_marker(
            CLUSTER,
            TestMarkerProvider(token),
            TestJwksProvider(sk),
            TestClockProvider(NOW),
        )
        .await
        .expect_err("validating an invalid issuer should have failed");
    }

    #[tokio::test]
    async fn wrong_algorithm_is_rejected() {
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header("RS256", None, Some(JWS_TYP), KID),
            &claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::Validation(jsonwebtoken::errors::new_error(
                ErrorKind::InvalidAlgorithm
            )))
        );
    }

    #[tokio::test]
    async fn alg_none_with_empty_signature_is_rejected() {
        let header = header("none", None, Some(JWS_TYP), KID);
        let claims = claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None);
        let token = format!("{}.{}.", base64_encode(&header), base64_encode(&claims));
        let sk = signing_key(1);

        validate_marker(
            CLUSTER,
            TestMarkerProvider(token),
            TestJwksProvider(sk),
            TestClockProvider(NOW),
        )
        .await
        .expect_err("token with no algorithm in header should not validate");
    }

    #[tokio::test]
    async fn symmetric_alg_with_valid_signature_is_rejected() {
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header("HS256", None, Some(JWS_TYP), KID),
            &claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::Validation(jsonwebtoken::errors::new_error(
                ErrorKind::InvalidAlgorithm
            )))
        );
    }

    #[tokio::test]
    async fn missing_enabled_claim_is_rejected() {
        let sk = signing_key(1);
        let mut claims = claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None);
        claims
            .as_object_mut()
            .expect("claims object")
            .remove("enabled");
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), KID),
            &claims,
        );

        validate_marker(
            CLUSTER,
            TestMarkerProvider(token),
            TestJwksProvider(sk),
            TestClockProvider(NOW),
        )
        .await
        .expect_err("token with no enabled should not validate");
    }

    #[tokio::test]
    async fn missing_kid_is_rejected() {
        let sk = signing_key(1);

        let mut header = header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), "");
        header.as_object_mut().expect("header object").remove("kid");
        let token = sign_token(
            &sk,
            &header,
            &claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::MissingKeyID)
        );
    }

    #[tokio::test]
    async fn unknown_kid_is_rejected() {
        let other_kid = "other-kid";
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), other_kid),
            &claims(&ISSUER_SERVICE, CLUSTER, IAT, true, None),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::MissingJwk(other_kid.to_string()))
        );
    }

    #[tokio::test]
    async fn signature_signed_by_wrong_key_is_rejected() {
        let signer = signing_key(2);
        let expected = signing_key(1);
        let token = normal_token(&signer);
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(expected),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::Validation(jsonwebtoken::errors::new_error(
                ErrorKind::InvalidSignature
            )))
        );
    }

    #[tokio::test]
    async fn wrong_segment_count_is_rejected() {
        let sk = signing_key(1);
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider("aaaa.bbbb".to_string()),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::Validation(jsonwebtoken::errors::new_error(
                ErrorKind::InvalidToken
            )))
        );
    }

    #[tokio::test]
    async fn invalid_base64url_header_is_rejected() {
        let sk = signing_key(1);
        validate_marker(
            CLUSTER,
            TestMarkerProvider("!!!.payload.sig".to_string()),
            TestJwksProvider(sk),
            TestClockProvider(NOW),
        )
        .await
        .expect_err("token with invalid base64 should not validate");
    }

    #[tokio::test]
    async fn subject_mismatch_is_rejected() {
        let other_cluster = "OtherCluster";
        let sk = signing_key(1);
        let token = sign_token(
            &sk,
            &header(JWS_ALG, Some(JWS_CRV), Some(JWS_TYP), KID),
            &claims(&ISSUER_SERVICE, other_cluster, IAT, true, None),
        );
        assert_eq!(
            validate_marker(
                CLUSTER,
                TestMarkerProvider(token),
                TestJwksProvider(sk),
                TestClockProvider(NOW)
            )
            .await,
            Err(ValidatorError::ClusterMismatch(
                other_cluster.to_string(),
                CLUSTER.to_string()
            ))
        );
    }

    #[tokio::test]
    async fn error_messages_never_leak_token_material() {
        let sk = signing_key(2);
        let expected = signing_key(1);
        let token = normal_token(&sk);
        let err = validate_marker(
            CLUSTER,
            TestMarkerProvider(token.clone()),
            TestJwksProvider(expected),
            TestClockProvider(NOW),
        )
        .await
        .expect_err("signature should be rejected");

        let message = err.to_string();
        assert!(!message.contains(&token));
        assert!(!message.contains(&token[..token.find('.').expect("token has a dot")]));
    }
}
