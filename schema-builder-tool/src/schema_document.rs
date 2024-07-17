use mongodb::bson::{Bson, DateTime};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct SchemaDocument {
    #[serde(rename = "_id")]
    pub id: String,
    #[serde(rename = "type")]
    pub namespace_type: String,
    pub schema: Bson,
    #[serde(rename = "lastUpdated")]
    pub last_updated: DateTime,
}
