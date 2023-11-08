use chrono::{TimeZone, Utc};
use mongosql::{
    catalog::{Catalog, Namespace},
    json_schema,
    options::{ExcludeNamespacesOption, SqlOptions},
    schema::Schema,
    translate_sql, SchemaCheckingMode,
};
use regex::Regex;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs::{create_dir_all, File},
    io::{self, Write},
    path::PathBuf,
    string::ToString,
};
use thiserror::Error;

const CONFIG_FILE: &str = "./benchmark/pipeline_generator/config/config.yml";

#[derive(Debug, Error)]
enum Error {
    #[error("invalid schema for MongoSQL model: {0:?}")]
    InvalidSchema(mongosql::schema::Error),
    #[error("failed to open ConfigFile: {0:?}")]
    ConfigFile(serde_yaml::Error),
    #[error("failed to FileOpen: {0:?}")]
    FileOpen(io::Error),
}

#[derive(Debug, Serialize, Deserialize)]
struct Workload {
    name: String,
    db: String,
    phase_template: String,
    workload_dir: String,
    catalog_schema: BTreeMap<String, BTreeMap<String, json_schema::Schema>>,
    query: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Config {
    workloads: Vec<Workload>,
}

fn build_catalog(
    catalog_schema: BTreeMap<String, BTreeMap<String, json_schema::Schema>>,
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

fn modify_pipeline_value(value: &mut Value) {
    match value {
        Value::Object(obj) => {
            // Workaround added to use $toString where Genny is converting a string to int
            // When SQL-1642 is unblocked, remove this and confirm Q22 passes validation
            if let Some(Value::String(s)) = obj.get("$literal") {
                if s.parse::<i64>().is_ok() {
                    let mut replace_with = serde_json::Map::new();
                    replace_with.insert("$toString".to_string(), obj.remove("$literal").unwrap());
                    *obj = replace_with;
                    return;
                }
            }

            if let Some(Value::Object(literal_obj)) = obj.get_mut("$literal") {
                if let Some(Value::Object(date_obj)) = literal_obj.get_mut("$date") {
                    if let Some(Value::String(timestamp_str)) = date_obj.get("$numberLong") {
                        if let Ok(timestamp) = timestamp_str.parse::<i64>() {
                            let date = Utc
                                .timestamp_millis_opt(timestamp)
                                .unwrap()
                                .format("%Y-%m-%dT%H:%M:%S")
                                .to_string();
                            // if the pipeline is running in genny, we need a special date format.
                            // Otherwise, use ISODate and generate a valid aggregation pipeline.
                            *value = if cfg!(feature = "genny") {
                                Value::String(format!("{{ ^Date: \"{}\" }}", date))
                            } else {
                                Value::String(format!("ISODate(\"{}\")", date))
                            };
                            return;
                        }
                    }
                }
            }
            for (_k, v) in obj.iter_mut() {
                modify_pipeline_value(v);
            }
        }
        Value::Array(arr) => {
            for v in arr.iter_mut() {
                modify_pipeline_value(v);
            }
        }
        _ => {}
    }
}

/// process_workload reads in a phase_template, modifies the pipeline, aggregate, and name.
/// Then writes out the modified phase to specified workload directory.
fn process_workload(config: &Workload, pipeline: &str, collection: &str) {
    let phase_template = File::open(&config.phase_template).unwrap();
    let mut yaml_value: serde_yaml::Value = serde_yaml::from_reader(phase_template).unwrap();

    if let serde_yaml::Value::Mapping(mut map) = yaml_value {
        // update the aggregation phase
        if let Some(serde_yaml::Value::Mapping(aggregation)) =
            map.get_mut("sample_phase_aggregation")
        {
            aggregation.insert(
                serde_yaml::Value::String("aggregate".into()),
                serde_yaml::Value::String(collection.to_string()),
            );
            aggregation.insert(
                serde_yaml::Value::String("pipeline".into()),
                serde_yaml::Value::String(pipeline.to_string()),
            );
        }

        // update the actor:
        let old_value = map.remove("SamplePhase").unwrap();
        map.insert(
            serde_yaml::Value::String(config.name.to_string()),
            old_value,
        );
        yaml_value = serde_yaml::Value::Mapping(map);
    }

    // Genny workload fails with '|-' in the pipeline value, removing it.
    // serde_yaml also does not serialize anchors well, so add them to the final string
    let output = serde_yaml::to_string(&yaml_value)
        .unwrap()
        .replace("pipeline: |-", "pipeline:")
        .replace(
            "sample_phase_aggregation:",
            "sample_phase_aggregation: &sample_phase_aggregation",
        )
        .replace("sample_phase", config.name.as_str())
        .replace("OperationCommand: ", "OperationCommand: *")
        .replace("\"{ ^", "{ ^")
        .replace("\" }\"", "\" }");

    let mut path = PathBuf::from(&config.workload_dir);
    create_dir_all(&path).unwrap();
    path.push(&config.name);
    path.set_extension("yml");
    let mut file = File::create(path).unwrap();
    file.write_all(output.as_bytes()).unwrap();
}

/// replace_final_replacement changes the final stage of the pipeline from an unset field on __bot
/// to a replaceRoot with __bot. This is possible because we know the domain of namespaces and query shapes.
/// This turns the documents from shape {"": {...}} to shape {...} which makes them directly comparable to server
fn replace_final_replacement(value: &mut Value) {
    if let Value::Array(arr) = value {
        arr.pop();
        arr.push(json!({"$replaceRoot": {"newRoot": "$__bot"}}));
    };
}

/// get_pipeline takes in a config workload and returns the pipeline and collection name by
/// calling translate_sql() on the query. It uses the catalog_schema to build the catalog.
fn get_pipeline(config: &Workload) -> (String, String) {
    let catalog = build_catalog(config.catalog_schema.clone()).unwrap();
    let translation = translate_sql(
        &config.db,
        &config.query,
        &catalog,
        SqlOptions::new(
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::Relaxed,
        ),
    );
    let translation = translation.unwrap();
    let pipeline = translation.pipeline;
    let collection = translation.target_collection.unwrap();

    let json_string = serde_json::to_string(&pipeline).unwrap();
    let mut json_value: Value = serde_json::from_str(json_string.as_str()).unwrap();

    modify_pipeline_value(&mut json_value);

    // only alter the pipeline when validating, so result shapes line up with server teams
    if cfg!(feature = "validation") {
        replace_final_replacement(&mut json_value);
    }

    // Removes quotes around ISODate and backslashes before quotes in genny dates
    let pipeline = serde_json::to_string_pretty(&json_value).unwrap();
    let normal_pattern = r#""ISODate\(\\(["\d\-\w:]*)\\"\)""#;
    let normal_re = Regex::new(normal_pattern).unwrap();
    let pipeline = normal_re.replace_all(&pipeline, r#"ISODate($1")"#);

    let genny_pattern = r#"\^Date: \\"([\w\d\-:]*)\\""#;
    let genny_re = Regex::new(genny_pattern).unwrap();
    let pipeline = genny_re.replace_all(&pipeline, r#"^Date: "$1""#);

    (pipeline.to_string(), collection)
}

fn main() {
    let config_file = File::open(CONFIG_FILE).map_err(Error::FileOpen).unwrap();
    let config: Config = serde_yaml::from_reader(config_file)
        .map_err(Error::ConfigFile)
        .unwrap();

    for workload in config.workloads {
        println!("Processing: {}", workload.name);
        let (pipeline, collection) = get_pipeline(&workload);
        process_workload(&workload, &pipeline, collection.as_str());
    }
}
