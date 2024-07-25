use super::Error;
use mongodb::{
    bson::{doc, Bson},
    sync::Client,
    IndexModel,
};
use mongosql::Translation;
use serde::{Deserialize, Serialize};
use std::{collections::BTreeMap, fs, io::Read, path::PathBuf};

#[derive(Debug, Serialize, Deserialize)]
pub struct IndexUsageYamlTestFile {
    pub catalog_data: BTreeMap<String, BTreeMap<String, Vec<Bson>>>,
    pub catalog_schema: BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>,
    pub indexes: BTreeMap<String, BTreeMap<String, Vec<IndexModel>>>,
    pub tests: Vec<IndexUsageTest>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct IndexUsageTest {
    pub description: String,
    pub skip_reason: Option<String>,
    pub current_db: String,
    pub query: String,
    pub expected_utilization: IndexUtilization,
    pub expected_index_bounds: Option<Vec<Bson>>,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
#[allow(clippy::enum_variant_names)]
pub enum IndexUtilization {
    #[serde(rename = "COLL_SCAN")]
    CollScan,
    #[serde(rename = "DISTINCT_SCAN")]
    DistinctScan,
    #[serde(rename = "IX_SCAN")]
    IxScan,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct ExplainResult {
    pub query_planner: Option<QueryPlanner>,
    pub stages: Option<Vec<ExplainStage>>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct QueryPlanner {
    pub winning_plan: WinningPlan,
    // Omitting unused fields
}
#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct WinningPlan {
    pub stage: Option<String>,
    pub input_stage: Option<InputStage>,
    pub query_plan: Option<QueryPlan>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct QueryPlan {
    pub stage: String,
    pub input_stage: Option<InputStage>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct InputStage {
    pub stage: String,
    pub index_bounds: Option<Bson>,
    pub input_stage: Option<Box<InputStage>>,
    // If the stage is an OR it will have multiple inputs
    pub input_stages: Option<Vec<InputStage>>,
    // Omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ExplainStage {
    #[serde(rename = "$cursor")]
    pub cursor: Option<CursorStage>,
    // omitting unused fields
}

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct CursorStage {
    pub query_planner: QueryPlanner,
    // omitting unused fields
}

/// load_index_test_files loads the YAML files in the provided `dir` into a Vec
/// of IndexUsageYamlTestFile structs.
pub fn load_index_test_files(dir: PathBuf) -> Result<Vec<IndexUsageYamlTestFile>, Error> {
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
pub fn parse_index_usage_yaml_file(path: PathBuf) -> Result<IndexUsageYamlTestFile, Error> {
    let mut f = fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let yaml: IndexUsageYamlTestFile = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml((format!("in file: {:?}", path), e)))?;
    Ok(yaml)
}

/// create_indexes creates all provided indexes on the mongodb instance.
pub fn create_indexes(
    client: &Client,
    indexes: BTreeMap<String, BTreeMap<String, Vec<IndexModel>>>,
) -> Result<(), Error> {
    for (db, coll_indexes) in indexes {
        let client_db = client.database(db.as_str());

        for (coll, indexes) in coll_indexes {
            let client_coll = client_db.collection::<Bson>(coll.as_str());

            client_coll
                .create_indexes(indexes)
                .run()
                .map_err(|e| Error::MongoDBCreateIndexes(db.clone(), coll, e))?;
        }
    }

    Ok(())
}

/// run_explain_aggregate runs the provided translation's pipeline against the
/// provided client using an `explain` command that wraps an `aggregate`.
pub fn run_explain_aggregate(
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
        .run_command(cmd)
        .run()
        .map_err(Error::MongoDBAggregation)?;

    // Deserialize the `explain` result
    mongodb::bson::from_document(result).map_err(Error::ExplainDeserialization)
}

impl ExplainResult {
    pub fn get_query_planner(&self) -> Result<QueryPlanner, Error> {
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
pub fn get_input_stage_of_winning_plan(winning_plan: WinningPlan) -> InputStage {
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
    pub fn get_root_stages(&self) -> Vec<&Self> {
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
pub fn as_index_utilization(stage_type: String) -> Result<IndexUtilization, Error> {
    match stage_type.as_str() {
        "COLLSCAN" => Ok(IndexUtilization::CollScan),
        "DISTINCT_SCAN" => Ok(IndexUtilization::DistinctScan),
        "IXSCAN" => Ok(IndexUtilization::IxScan),
        _ => Err(Error::InvalidRootStage(stage_type)),
    }
}
