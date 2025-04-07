use agg_ast::definitions::Stage;
use mongodb::bson::doc;
use mongosql::json_schema;
use serde::{Deserialize, Serialize};
use std::{collections::BTreeMap, fs, io::Read, path::PathBuf};

use super::Error;

#[derive(Debug, Serialize, Deserialize)]
#[serde(untagged)]
pub enum SchemaDerivationYamlTestFile {
    Multiple(SpecQuerySchemaDerivationTestFile),
    Single(SchemaDerivationTest),
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SpecQuerySchemaDerivationTestFile {
    pub catalog_schema: Option<BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>>,
    pub tests: Vec<SchemaDerivationTest>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SchemaDerivationTest {
    pub description: Option<String>,
    pub catalog_schema_file: Option<String>,
    pub current_db: Option<String>,
    pub current_collection: Option<String>,
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
