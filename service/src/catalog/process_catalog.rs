use mongosql::{build_catalog_from_catalog_schema, catalog::Catalog};
use serde::{Deserialize, Serialize};
use std::{
    collections::BTreeMap,
    io::{self, Read},
};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("failed to read file: {0:?}")]
    InvalidFile(io::Error),
    #[error("unable to read file to string: {0:?}")]
    CannotReadFileToString(io::Error),
    #[error("unable to deserialize BSON: {0:?}")]
    CannotDeserializeBson(bson::de::Error),
    #[error("unable to deserialize JSON file: {0}")]
    CannotDeserializeJson(String, #[source] serde_json::Error),
    #[error("catalog build error: {0}")]
    CatalogBuildFail(String),
}

pub fn build_catalog_from_bytes(bytes: &[u8]) -> Result<Catalog, Error> {
    let json: CatalogJsonFile = bson::from_slice(bytes).map_err(Error::CannotDeserializeBson)?;
    build_catalog_from_catalog_schema(json.catalog_schema)
        .map_err(|e| Error::CatalogBuildFail(e.to_string()))
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CatalogJsonFile {
    catalog_schema: BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>,
}

/// parse_catalog_json_file parses a JSON file into a QueryJsonTestFile struct.
pub fn parse_catalog_json_file(path: std::path::PathBuf) -> Result<CatalogJsonFile, Error> {
    let mut f = std::fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let json: CatalogJsonFile = serde_json::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeJson(format!("in file: {:?}", path), e))?;
    Ok(json)
}
