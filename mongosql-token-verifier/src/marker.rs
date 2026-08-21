use serde::{Deserialize, Serialize};

use crate::Issuer;

/// Required claims on MongoDB Markers
#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    /// Optional expiration time. Enforced for emergency markers
    pub exp: Option<i64>,

    /// The issuer
    pub iss: Issuer,

    /// The name of the cluster entitled to this marker
    pub sub: String,

    /// Whether or not this entitlement is active
    pub enabled: bool,
}
