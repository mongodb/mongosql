#[cfg(test)]
mod json_schema_test;

use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[allow(dead_code)]
#[derive(Serialize, Deserialize, PartialEq, Debug, Default)]
#[serde(rename_all = "camelCase", default)]
pub struct Schema {
    #[serde(skip_serializing_if = "is_default")]
    pub bson_type: BsonType,
    #[serde(skip_serializing_if = "is_default")]
    pub properties: HashMap<String, Schema>,
    #[serde(skip_serializing_if = "is_default")]
    pub required: Vec<String>,
    #[serde(skip_serializing_if = "is_default")]
    pub additional_properties: bool,
    #[serde(skip_serializing_if = "is_default")]
    pub items: Vec<Schema>,
    #[serde(skip_serializing_if = "is_default")]
    pub any_of: Vec<Schema>,
    #[serde(skip_serializing_if = "is_default")]
    pub one_of: Vec<Schema>,
}

#[allow(dead_code)]
#[derive(Serialize, Deserialize, PartialEq, Debug)]
#[serde(untagged)]
pub enum BsonType {
    Single(String),
    Multiple(Vec<String>),
}

impl Default for BsonType {
    fn default() -> Self {
        BsonType::Multiple(vec![])
    }
}

fn is_default<T: Default + PartialEq>(t: &T) -> bool {
    t == &T::default()
}
