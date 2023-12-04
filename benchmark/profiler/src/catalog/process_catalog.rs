use mongosql::{build_catalog_from_catalog_schema, catalog::Catalog};
use serde::{Deserialize, Serialize};
use std::{collections::BTreeMap, fs, io, io::Read, path::PathBuf};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("failed to read file: {0:?}")]
    InvalidFile(io::Error),
    #[error("unable to read file to string: {0:?}")]
    CannotReadFileToString(io::Error),
    #[error("unable to deserialize JSON file: {0:?}")]
    CannotDeserializeJson((String, serde_json::Error)),
    #[error("failed to convert schema to MongoSQL model: {0:?}")]
    InvalidSchema(mongosql::result::Error),
}

#[derive(Debug, Serialize, Deserialize)]
struct CatalogJsonFile {
    catalog_schema: BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>,
}

pub fn load_catalog(catalog: &str) -> Result<Catalog, Error> {
    match parse_catalog_json_file(catalog.into()) {
        Ok(file) => build_catalog_from_catalog_schema(file.catalog_schema.to_owned())
            .map_err(Error::InvalidSchema),
        Err(e) => Err(e),
    }
}

/// parse_catalog_json_file parses a JSON file into a QueryJsonTestFile struct.
fn parse_catalog_json_file(path: PathBuf) -> Result<CatalogJsonFile, Error> {
    let mut f = fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let json: CatalogJsonFile = serde_json::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeJson((format!("in file: {:?}: ", path), e)))?;
    Ok(json)
}
