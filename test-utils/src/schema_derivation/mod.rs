#![allow(clippy::result_large_err)]
use agg_ast::definitions::Stage;
use mongodb::bson::doc;
use mongosql::json_schema;
use serde::{Deserialize, Serialize};
use std::{fs, io::Read, path::PathBuf};

use super::Error;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SchemaDerivationYamlTestFile {
    pub tests: Vec<SchemaDerivationTest>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SchemaDerivationTest {
    pub description: Option<String>,
    pub skip_reason: Option<String>,
    pub current_db: Option<String>,
    pub current_collection: Option<String>,
    pub catalog_dbs: Option<Vec<String>>,
    pub pipeline: Vec<Stage>,
    pub result_set_schema: json_schema::Schema,
}

/// parse_schema_derivation_yaml_file parses a YAML file into a SchemaDerivationYamlTestFile struct.
pub fn parse_schema_derivation_yaml_file(
    path: PathBuf,
) -> Result<SchemaDerivationYamlTestFile, Error> {
    let mut f = fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let yaml: SchemaDerivationYamlTestFile = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml((format!("in file: {:?}: ", path), e)))?;
    Ok(yaml)
}
