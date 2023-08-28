use mongodb::{
    bson::{doc, Bson},
    sync::Client,
    IndexModel,
};
use mongosql::Translation;
use serde::{Deserialize, Serialize};
use std::{collections::BTreeMap, fs, io::Read, path::PathBuf};

use crate::utils::{build_catalog, load_catalog_data, Error, MONGODB_URI};

#[derive(Debug, Serialize, Deserialize)]
pub(crate) struct IndexUsageYamlTestFile {
    catalog_data: BTreeMap<String, BTreeMap<String, Vec<Bson>>>,
    catalog_schema: BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>,
    indexes: BTreeMap<String, BTreeMap<String, Vec<IndexModel>>>,
    tests: Vec<IndexUsageTest>,
}

#[derive(Debug, Serialize, Deserialize)]
pub(crate) struct IndexUsageTest {
    description: String,
    skip_reason: Option<String>,
    current_db: String,
    query: String,
    expected_utilization: IndexUtilization,
    expected_index_bounds: Option<Vec<Bson>>,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
#[allow(clippy::enum_variant_names)]
pub(crate) enum IndexUtilization {
    #[serde(rename = "COLL_SCAN")]
    CollScan,
    #[serde(rename = "DISTINCT_SCAN")]
    DistinctScan,
    #[serde(rename = "IX_SCAN")]
    IxScan,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub(crate) struct ExplainResult {
    ok: f32,
    query_planner: Option<QueryPlanner>,
    stages: Option<Vec<ExplainStage>>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub(crate) struct QueryPlanner {
    winning_plan: WinningPlan,
    // Omitting unused fields
}
#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct WinningPlan {
    stage: Option<String>,
    input_stage: Option<InputStage>,
    query_plan: Option<QueryPlan>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct QueryPlan {
    stage: String,
    input_stage: Option<InputStage>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
struct InputStage {
    stage: String,
    index_bounds: Option<Bson>,
    input_stage: Option<Box<InputStage>>,
    // If the stage is an OR it will have multiple inputs
    input_stages: Option<Vec<InputStage>>,
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
    query_planner: QueryPlanner,
    // omitting unused fields
}

/// run_index_usage_tests is the main function in this file. This is the
/// function that runs the index usage tests from the YAML files in
/// tests/index_usage_tests.
/// This test is marked with "ignore" so we can continue to run all unit
/// tests via `cargo test` without flags.
#[test]
#[cfg_attr(not(feature = "index-test"), ignore)]
fn run_index_usage_tests() -> Result<(), Error> {
    let test_dir = "../tests/index_usage_tests";
    let test_files = load_index_test_files(PathBuf::from(test_dir))?;

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
            .map_err(Error::Translation)?;

            let explain_result = run_explain_aggregate(&client, translation)?;

            let query_planner = explain_result.get_query_planner()?;

            let input_stage = get_input_stage_of_winning_plan(query_planner.winning_plan);

            let root_input_stages = input_stage.get_root_stages();

            let actual_index_utilizations = root_input_stages
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

            let actual_index_bounds = root_input_stages
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

/// load_index_test_files loads the YAML files in the provided `dir` into a Vec
/// of IndexUsageYamlTestFile structs.
fn load_index_test_files(dir: PathBuf) -> Result<Vec<IndexUsageYamlTestFile>, Error> {
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
    let mut f = fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let yaml: IndexUsageYamlTestFile = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml((format!("in file: {:?}", path), e)))?;
    Ok(yaml)
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
        "verbosity": "queryPlanner"
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
    fn get_query_planner(&self) -> Result<QueryPlanner, Error> {
        match self.query_planner.clone() {
            Some(query_planner) => Ok(query_planner),
            None => match self.stages.clone() {
                Some(stages) => {
                    for stage in stages {
                        if stage.cursor.is_some() {
                            return Ok(stage.cursor.unwrap().query_planner);
                        }
                    }
                    Err(Error::MissingQueryPlanner(self.clone()))
                }
                None => Err(Error::MissingQueryPlanner(self.clone())),
            },
        }
    }
}

/// This function figures out which field of the WinningPlan contains the
/// InputStage to run get_root_stages() on.
fn get_input_stage_of_winning_plan(winning_plan: WinningPlan) -> InputStage {
    match (
        winning_plan.stage,
        winning_plan.input_stage,
        winning_plan.query_plan,
    ) {
        (Some(stage), None, None) => InputStage {
            stage,
            index_bounds: None,
            input_stage: None,
            input_stages: None,
        },
        (_, None, Some(query_plan)) => match (query_plan.stage, query_plan.input_stage) {
            (qp_stage, None) => InputStage {
                stage: qp_stage,
                index_bounds: None,
                input_stage: None,
                input_stages: None,
            },
            (_, Some(qp_input_stage)) => qp_input_stage,
        },
        (_, Some(input_stage), None) => input_stage,
        // The unreachable() scenario applies to (Some,Some,Some), (None,None,None), and (None,Some,Some).
        // This makes sense because we should never have a query_plan and input_stage at the same time,
        // and there should always be at least one Some variant in the tuple.
        _ => unreachable!(),
    }
}

/// Implementation for getting the root stage of an InputStage tree.
impl InputStage {
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
        "DISTINCT_SCAN" => Ok(IndexUtilization::DistinctScan),
        "IXSCAN" => Ok(IndexUtilization::IxScan),
        _ => Err(Error::InvalidRootStage(stage_type)),
    }
}
