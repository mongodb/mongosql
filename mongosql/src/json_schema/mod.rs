#[cfg(test)]
mod test;

use bson::{Bson, Document};
use enum_iterator::IntoEnumIterator;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("failed to serialize JSON Schema to BSON")]
    BsonSerializationFailure,
    #[error("failed to deserialize JSON Schema from BSON")]
    BsonDeserializationFailure,
}

#[derive(Serialize, Deserialize, PartialEq, Debug, Clone, Default)]
#[serde(rename_all = "camelCase")]
pub struct Schema {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub bson_type: Option<BsonType>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub properties: Option<HashMap<String, Schema>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub required: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub additional_properties: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub items: Option<Items>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub any_of: Option<Vec<Schema>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub one_of: Option<Vec<Schema>>,
}

#[derive(Serialize, Deserialize, PartialEq, Eq, Clone, Debug)]
#[serde(untagged)]
pub enum BsonType {
    Single(BsonTypeName),
    Multiple(Vec<BsonTypeName>),
}

#[derive(Serialize, Deserialize, PartialEq, Eq, Debug, IntoEnumIterator, Copy, Clone)]
#[serde(rename_all = "camelCase")]
pub enum BsonTypeName {
    Object,
    Array,
    Null,
    String,
    Int,
    Double,
    Long,
    Decimal,
    BinData,
    Undefined,
    ObjectId,
    Bool,
    Date,
    Regex,
    DbPointer,
    Javascript,
    Symbol,
    JavascriptWithScope,
    Timestamp,
    MinKey,
    MaxKey,
}

#[derive(Serialize, Deserialize, PartialEq, Clone, Debug)]
#[serde(untagged)]
pub enum Items {
    Single(Box<Schema>),
    Multiple(Vec<Schema>),
}

impl Schema {
    pub fn to_bson(&self) -> Result<Bson> {
        let serializer = bson::Serializer::new();
        let serializer = serde_stacker::Serializer::new(serializer);
        self.serialize(serializer)
            .map_err(|_| Error::BsonSerializationFailure)
    }

    pub fn from_document(doc: &Document) -> Result<Self> {
        let as_bson = Bson::Document(doc.clone());
        let deserializer = bson::Deserializer::new(as_bson);
        let deserializer = serde_stacker::Deserializer::new(deserializer);
        Deserialize::deserialize(deserializer).map_err(|_| Error::BsonDeserializationFailure)
    }
}
