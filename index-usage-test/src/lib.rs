#![cfg(test)]

use itertools::Itertools;
use lazy_static::lazy_static;
use mongodb::{
    bson::{doc, Bson},
    sync::Client,
    IndexModel,
};
use mongosql::{
    catalog::{Catalog, Namespace},
    schema::Schema,
    Translation,
};
use serde::{Deserialize, Serialize};
use std::{collections::BTreeMap, env, fs, io, io::Read, path::PathBuf, string::ToString};
use thiserror::Error;

#[derive(Debug, Serialize, Deserialize)]
struct IndexUsageYamlTestFile {
    catalog_data: BTreeMap<String, BTreeMap<String, Vec<Bson>>>,
    catalog_schema: BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>,
    indexes: BTreeMap<String, BTreeMap<String, Vec<IndexModel>>>,
    tests: Vec<IndexUsageTest>,
}

#[derive(Debug, Serialize, Deserialize)]
struct IndexUsageTest {
    description: String,
    skip_reason: Option<String>,
    current_db: String,
    query: String,
    expected_utilization: IndexUtilization,
    expected_index_bounds: Option<Vec<Bson>>,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
enum IndexUtilization {
    #[serde(rename = "COLL_SCAN")]
    CollScan,
    #[serde(rename = "IX_SCAN")]
    IxScan,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct ExplainResult {
    ok: f32,
    execution_stats: Option<ExecutionStats>,
    stages: Option<Vec<ExplainStage>>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct ExecutionStats {
    execution_stages: ExecutionStage,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct ExecutionStage {
    stage: String,
    index_bounds: Option<Bson>,
    input_stage: Option<Box<ExecutionStage>>,
    // If the stage is an OR it will have multiple inputs
    input_stages: Option<Vec<ExecutionStage>>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct ExplainStage {
    #[serde(rename = "$cursor")]
    cursor: Option<CursorStage>,
    // omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct CursorStage {
    execution_stats: ExecutionStats,
    // omitting unused fields
}

#[derive(Debug, Error)]
enum Error {
    #[error("failed to read directory: {0:?}")]
    InvalidDirectory(io::Error),
    #[error("failed to load file paths: {0:?}")]
    InvalidFilePath(io::Error),
    #[error("failed to read file: {0:?}")]
    InvalidFile(io::Error),
    #[error("unable to read file to string: {0:?}")]
    CannotReadFileToString(io::Error),
    #[error("unable to deserialize YAML file: {0:?}")]
    CannotDeserializeYaml(serde_yaml::Error),
    #[error("failed to create mongodb client: {0:?}")]
    CannotCreateMongoDBClient(mongodb::error::Error),
    #[error("failed to drop db '{0}': {1:?}")]
    MongoDBDrop(String, mongodb::error::Error),
    #[error("failed to insert into '{0}.{1}': {2:?}")]
    MongoDBInsert(String, String, mongodb::error::Error),
    #[error("failed to create indexes for '{0}.{1}': {2:?}")]
    MongoDBCreateIndexes(String, String, mongodb::error::Error),
    #[error("failed to schema to MongoSQL model: {0:?}")]
    InvalidSchema(mongosql::schema::Error),
    #[error("failed to translate query: {0}")]
    Translation(String),
    #[error("failed to run aggregation: {0:?}")]
    MongoDBAggregation(mongodb::error::Error),
    #[error("failed to deserialize ExplainResult: {0:?}")]
    ExplainDeserialization(mongodb::bson::de::Error),
    #[error("invalid root stage: {0}")]
    InvalidRootStage(String),
    #[error("no executionStats found: {0:?}")]
    MissingExecutionStats(ExplainResult),
}

const TEST_DIR: &str = "../tests/index_usage_tests";

lazy_static! {
    static ref MONGODB_URI: String = format!(
        "mongodb://localhost:{}",
        env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
    );
}

/// run_index_usage_tests is the main function in this file. This is the
/// function that runs the index usage tests from the YAML files in
/// tests/index_usage_tests.
/// This test is marked with "ignore" so we can continue to run all unit
/// tests via `cargo test` without flags.
#[test]
#[ignore]
fn run_index_usage_tests() -> Result<(), Error> {
    let test_files = load_test_files(PathBuf::from(TEST_DIR))?;

    let client =
        Client::with_uri_str(MONGODB_URI.clone()).map_err(Error::CannotCreateMongoDBClient)?;

    for test_file in test_files {
        load_catalog_data(&client, test_file.catalog_data)?;
        create_indexes(&client, test_file.indexes)?;

        let catalog = build_catalog(test_file.catalog_schema)?;

        for test in test_file.tests {
            if test.skip_reason.is_some() {
                continue;
            }

            let translation = mongosql::translate_sql(
                test.current_db.as_str(),
                test.query.as_str(),
                &catalog,
                mongosql::SchemaCheckingMode::Strict,
            )
            .map_err(|e| Error::Translation(format!("{e:?}")))?;

            let explain_result = run_explain_aggregate(&client, translation)?;

            let execution_stats = explain_result.get_execution_stats()?;

            let root_execution_stages = execution_stats.execution_stages.get_root_stages();

            let actual_index_utilizations = root_execution_stages
                .clone()
                .into_iter()
                .map(|root_stage| as_index_utilization(root_stage.stage.clone()))
                .collect::<Result<Vec<IndexUtilization>, Error>>()?;

            for actual_index_utilization in actual_index_utilizations {
                assert_eq!(
                    test.expected_utilization, actual_index_utilization,
                    "{}: unexpected index utilization",
                    test.description
                );
            }

            let actual_index_bounds = root_execution_stages
                .into_iter()
                .map(|root_stage| root_stage.index_bounds.clone())
                .collect::<Option<Vec<Bson>>>();

            assert_eq!(
                test.expected_index_bounds, actual_index_bounds,
                "{}: unexpected index bounds",
                test.description
            );
        }
    }

    Ok(())
}

/// load_test_files loads the YAML files in the provided `dir` into a Vec
/// of IndexUsageYamlTestFile structs.
fn load_test_files(dir: PathBuf) -> Result<Vec<IndexUsageYamlTestFile>, Error> {
    let entries = fs::read_dir(dir).map_err(Error::InvalidDirectory)?;

    entries
        .map(|entry| match entry {
            Ok(de) => parse_index_usage_yaml_file(de.path()),
            Err(e) => Err(Error::InvalidFilePath(e)),
        })
        .collect::<Result<Vec<IndexUsageYamlTestFile>, Error>>()
}

/// parse_index_usage_yaml_file parses a YAML file at the provided `path` into
/// an IndexUsageYamlTestFile struct.
fn parse_index_usage_yaml_file(path: PathBuf) -> Result<IndexUsageYamlTestFile, Error> {
    let mut f = fs::File::open(path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let yaml: IndexUsageYamlTestFile =
        serde_yaml::from_str(&contents).map_err(Error::CannotDeserializeYaml)?;
    Ok(yaml)
}

/// load_catalog_data drops any existing catalog data and then inserts the
/// provided catalog data into the mongodb instance.
fn load_catalog_data(
    client: &Client,
    catalog_data: BTreeMap<String, BTreeMap<String, Vec<Bson>>>,
) -> Result<(), Error> {
    let catalog_dbs = catalog_data.keys().collect_vec();
    drop_catalog_data(client, catalog_dbs)?;

    for (db, coll_data) in catalog_data {
        let client_db = client.database(db.as_str());

        for (coll, documents) in coll_data {
            let client_coll = client_db.collection::<Bson>(coll.as_str());
            client_coll
                .insert_many(documents, None)
                .map_err(|e| Error::MongoDBInsert(db.clone(), coll, e))?;
        }
    }

    Ok(())
}

/// drop_catalog_data drops all dbs in the provided list.
fn drop_catalog_data(client: &Client, catalog_dbs: Vec<&String>) -> Result<(), Error> {
    for db in catalog_dbs {
        client
            .database(db.as_str())
            .drop(None)
            .map_err(|e| Error::MongoDBDrop(db.clone(), e))?;
    }
    Ok(())
}

/// create_indexes creates all provided indexes on the mongodb instance.
fn create_indexes(
    client: &Client,
    indexes: BTreeMap<String, BTreeMap<String, Vec<IndexModel>>>,
) -> Result<(), Error> {
    for (db, coll_indexes) in indexes {
        let client_db = client.database(db.as_str());

        for (coll, indexes) in coll_indexes {
            let client_coll = client_db.collection::<Bson>(coll.as_str());

            client_coll
                .create_indexes(indexes, None)
                .map_err(|e| Error::MongoDBCreateIndexes(db.clone(), coll, e))?;
        }
    }

    Ok(())
}

/// build_catalog converts the json_schema::Schema objects into schema::Schema
/// objects and builds a catalog from those.
fn build_catalog(
    catalog_schema: BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>,
) -> Result<Catalog, Error> {
    catalog_schema
        .into_iter()
        .flat_map(|(db, coll_schemas)| {
            coll_schemas.into_iter().map(move |(coll, schema)| {
                let mongosql_schema = Schema::try_from(schema).map_err(Error::InvalidSchema)?;
                Ok((
                    Namespace {
                        db: db.clone(),
                        collection: coll,
                    },
                    mongosql_schema,
                ))
            })
        })
        .collect()
}

/// run_explain_aggregate runs the provided translation's pipeline against the
/// provided client using an `explain` command that wraps an `aggregate`.
fn run_explain_aggregate(
    client: &Client,
    translation: Translation,
) -> Result<ExplainResult, Error> {
    // Determine if this a db- or collection-level aggregation
    let aggregate = match translation.target_collection {
        None => Bson::Int32(1),
        Some(collection) => Bson::String(collection),
    };

    let cmd = doc! {
        "explain": {
            "aggregate": aggregate,
            "pipeline": translation.pipeline,
            "cursor": {},
        },
        "verbosity": "executionStats"
    };

    // Run the aggregation as an `explain` command
    let result = client
        .database(translation.target_db.as_str())
        .run_command(cmd, None)
        .map_err(Error::MongoDBAggregation)?;

    // Deserialize the `explain` result
    mongodb::bson::from_document(result).map_err(Error::ExplainDeserialization)
}

impl ExplainResult {
    fn get_execution_stats(&self) -> Result<ExecutionStats, Error> {
        match self.execution_stats.clone() {
            Some(execution_stats) => Ok(execution_stats),
            None => match self.stages.clone() {
                Some(stages) => {
                    for stage in stages {
                        if stage.cursor.is_some() {
                            return Ok(stage.cursor.unwrap().execution_stats);
                        }
                    }
                    Err(Error::MissingExecutionStats(self.clone()))
                }
                None => Err(Error::MissingExecutionStats(self.clone())),
            },
        }
    }
}

/// Implementation for getting the root stage of an ExecutionStage tree.
impl ExecutionStage {
    fn get_root_stages(&self) -> Vec<&Self> {
        match &self.input_stage {
            None => match &self.input_stages {
                None => vec![self],
                Some(input_stages) => input_stages
                    .iter()
                    .flat_map(|input_stage| input_stage.get_root_stages())
                    .collect(),
            },
            Some(input_stage) => input_stage.get_root_stages(),
        }
    }
}

/// as_index_utilization converts an ExecutionStage.stage type into an
/// IndexUtilization value. Only COLLSCAN and IXSCAN are valid.
fn as_index_utilization(stage_type: String) -> Result<IndexUtilization, Error> {
    match stage_type.as_str() {
        "COLLSCAN" => Ok(IndexUtilization::CollScan),
        "IXSCAN" => Ok(IndexUtilization::IxScan),
        _ => Err(Error::InvalidRootStage(stage_type)),
    }
}
