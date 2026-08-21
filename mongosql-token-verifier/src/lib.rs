use serde::{Deserialize, Serialize};

pub mod clients;
pub mod marker;
pub mod providers;
pub mod validator;

pub const STATUS_DB: &str = "__mdb_internal_sqlinterface";
pub const STATUS_COLLECTION: &str = "__sql_status";
pub const ENTITLEMENT_ID: &str = "entitlement";
pub(crate) const JWKS_WELL_KNOWN_URL: &str =
    "https://cloud.mongodb.com/.well-known/mongosql/jwks.json";

/// Valid issuers for MongoDB SQL Interface markers
#[derive(Debug, Deserialize, Serialize, PartialEq, Eq)]
pub enum Issuer {
    /// Normal issuer
    #[serde(rename = "mongosql-service")]
    Service,

    /// Temporary issuer
    #[serde(rename = "mongosql-emergency")]
    Emergency,
}

#[cfg(test)]
mod test {
    use super::{ENTITLEMENT_ID, JWKS_WELL_KNOWN_URL, STATUS_COLLECTION, STATUS_DB};

    #[test]
    fn marker_namespace_matches_authoritative_producer() {
        // Must match the mms producer; drift here makes every enabled cluster fail closed.
        assert_eq!(STATUS_DB, "__mdb_internal_sqlinterface");
        assert_eq!(STATUS_COLLECTION, "__sql_status");
        assert_eq!(ENTITLEMENT_ID, "entitlement");
    }

    #[test]
    fn jwks_well_known_endpoint_is_correct() {
        assert_eq!(
            JWKS_WELL_KNOWN_URL,
            "https://cloud.mongodb.com/.well-known/mongosql/jwks.json"
        );
    }
}
