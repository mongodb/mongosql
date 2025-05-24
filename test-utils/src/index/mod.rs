#![allow(clippy::result_large_err)]
use super::Error;
use mongodb::{
    bson::{self, doc, Bson, Document},
    sync::Client,
};
use mongosql::Translation;
use serde::{Deserialize, Serialize};
use sql_engines_common_test_infra::{
    parse_yaml_test_file, sanitize_description, Error as cti_err, TestGenerator, YamlTestCase,
    YamlTestFile,
};
use std::{fs::File, io::Write, path::PathBuf};

#[derive(Debug, Serialize, Deserialize)]
pub struct IndexUsageTestExpectations {
    pub expected_utilization: IndexUtilization,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct IndexUsageTestOptions {
    pub current_db: String,
}

pub type IndexUsageTestCase =
    YamlTestCase<String, IndexUsageTestExpectations, IndexUsageTestOptions>;

pub struct IndexUsageTestGenerator;

impl TestGenerator for IndexUsageTestGenerator {
    fn generate_test_file_header(
        &self,
        generated_test_file: &mut File,
        canonicalized_path: String,
    ) -> sql_engines_common_test_infra::Result<()> {
        write!(
            generated_test_file,
            include_str!("../templates/index_usage_test_header_template"),
            path = canonicalized_path,
        )
        .map_err(|e| {
            cti_err::Io(
                format!("failed to write index test header for '{canonicalized_path}'"),
                e,
            )
        })
    }

    fn generate_test_file_body(
        &self,
        generated_test_file: &mut File,
        original_path: PathBuf,
    ) -> sql_engines_common_test_infra::Result<()> {
        let parsed_test_file: YamlTestFile<IndexUsageTestCase> =
            parse_yaml_test_file(original_path)?;

        for (index, test_case) in parsed_test_file.tests.iter().enumerate() {
            let sanitized_test_name = sanitize_description(&test_case.description);
            let res = if let Some(skip_reason) = test_case.skip_reason.as_ref() {
                write!(
                    generated_test_file,
                    include_str!("../templates/ignore_body_template"),
                    feature = "index",
                    ignore_reason = skip_reason,
                    name = sanitized_test_name,
                )
            } else {
                write!(
                    generated_test_file,
                    include_str!("../templates/index_usage_test_body_template"),
                    name = sanitized_test_name,
                    index = index,
                )
            };
            res.map_err(|e| {
                cti_err::Io(
                    format!(
                        "failed to write index test body for test '{}'",
                        test_case.description
                    ),
                    e,
                )
            })?;
        }

        Ok(())
    }
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
    deserialize_explain_result(result)
}

pub(crate) fn deserialize_explain_result(d: Document) -> Result<ExplainResult, Error> {
    let deserializer = bson::Deserializer::new(bson::Bson::Document(d));
    let deserializer = serde_stacker::Deserializer::new(deserializer);
    Deserialize::deserialize(deserializer).map_err(Error::ExplainDeserialization)
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
            input_stage: None,
            input_stages: None,
        },
        (_, None, Some(query_plan)) => match (query_plan.stage, query_plan.input_stage) {
            (qp_stage, None) => InputStage {
                stage: qp_stage,
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

#[cfg(test)]
mod no_stack_overflows {
    #[test]
    fn when_deserializing() {
        use super::*;
        let mut input_stage =
            bson::doc! {"stage": "COLLSCAN", "filter": {"x": {"$eq": 1}, "direction": "forward"}};
        // This uses a smaller limit than the serialization test
        // since the mongodb::bson::Deserializer does not support
        // the disable_recursion_limit method required by serde_stacker
        // to handle greater depths. At 250 nesting depth, the standard
        // bson::from_document encounters a stack overflow, whereas the
        // manual serde_stacker deserialization technique does not. We get_root_stages
        // with 400 here just to be extra sure.
        for _ in 0..400 {
            input_stage = bson::doc! { "stage": "COLLSCAN", "filter": {"x": {"$eq": 1}}, "direction": "forward", "inputStage": input_stage };
        }

        println!("start");

        let _ = deserialize_explain_result(bson::doc! {
            "queryPlanner": {
                "namespace": "test.mycollection",
                "indexFilterSet": false,
                "parsedQuery": { "x": { "$eq": 1 } },
                "queryHash": "DF77A253",
                "planCacheKey": "DF77A253",
                "optimizedPipeline": true,
                "maxIndexedOrSolutionsReached": false,
                "maxIndexedAndSolutionsReached": false,
                "maxScansToExplodeReached": false,
                "winningPlan": input_stage,
            },
        });
    }
}
