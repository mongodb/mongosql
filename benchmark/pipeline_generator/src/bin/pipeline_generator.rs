use chrono::{TimeZone, Utc};
use mongosql::{
    catalog::{Catalog, Namespace},
    json_schema,
    schema::Schema,
    translate_sql, SchemaCheckingMode,
};
use serde::{Deserialize, Serialize};
use serde_json::{map::Map, Value};
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

fn modify_date_literal(parent: &mut Map<String, Value>, key: String) {
    match parent.get_mut(&key) {
        // for each object, look for a $date to convert
        Some(Value::Object(obj)) => {
            if let Some(Value::Object(date_obj)) = obj.get_mut("$date") {
                if let Some(Value::String(timestamp_str)) = date_obj.get("$numberLong") {
                    if let Ok(timestamp) = timestamp_str.parse::<i64>() {
                        let date = Utc
                            .timestamp_millis_opt(timestamp)
                            .unwrap()
                            .format("%Y-%m-%dT%H:%M:%S")
                            .to_string();
                        parent.insert(key, Value::String(format!("ISODate(\"{}\")", date)));
                        return;
                    }
                }
            }
            let keys: Vec<String> = obj.keys().cloned().collect();
            for k in keys {
                modify_date_literal(obj, k);
            }
        }
        // for array values, recurse on any objects contained within the array
        Some(Value::Array(a)) => {
            for item in a {
                if let Value::Object(obj) = item {
                    let keys: Vec<String> = obj.keys().cloned().collect();
                    for k in keys {
                        modify_date_literal(obj, k);
                    }
                }
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
        if let Some(serde_yaml::Value::Mapping(sample_phase)) = map.get_mut("SamplePhase") {
            if let Some(serde_yaml::Value::Sequence(operations)) =
                sample_phase.get_mut("Operations")
            {
                for operation in operations {
                    if let serde_yaml::Value::Mapping(operation) = operation {
                        if let Some(serde_yaml::Value::Mapping(command)) =
                            operation.get_mut("OperationCommand")
                        {
                            command.insert(
                                serde_yaml::Value::String("aggregate".into()),
                                serde_yaml::Value::String(collection.to_string()),
                            );
                            command.insert(
                                serde_yaml::Value::String("pipeline".into()),
                                serde_yaml::Value::String(pipeline.to_string()),
                            );
                        }
                    }
                }
            }
        }
        let old_value = map.remove("SamplePhase").unwrap();
        map.insert(
            serde_yaml::Value::String(config.name.to_string()),
            old_value,
        );
        yaml_value = serde_yaml::Value::Mapping(map);
    }

    // Genny workload fails with '|-' in the pipeline value, removing it.
    let output = serde_yaml::to_string(&yaml_value)
        .unwrap()
        .replace("pipeline: |-", "pipeline:");
    let mut path = PathBuf::from(&config.workload_dir);
    create_dir_all(&path).unwrap();
    path.push(&config.name);
    path.set_extension("yml");
    let mut file = File::create(path).unwrap();
    file.write_all(output.as_bytes()).unwrap();
}

/// get_pipeline takes in a config workload and returns the pipeline and collection name by
/// calling translate_sql() on the query. It uses the catalog_schema to build the catalog.
fn get_pipeline(config: &Workload) -> (String, String) {
    let catalog = build_catalog(config.catalog_schema.clone()).unwrap();
    let translation = translate_sql(
        &config.db,
        &config.query,
        &catalog,
        SchemaCheckingMode::default(),
    );
    let translation = translation.unwrap();
    let pipeline = translation.pipeline;
    let collection = translation.target_collection.unwrap();

    let json_string = serde_json::to_string(&pipeline).unwrap();
    let mut json_value: Value = serde_json::from_str(json_string.as_str()).unwrap();

    // Modifies the date literals to be ISODate
    if let Value::Array(arr) = &mut json_value {
        for item in arr {
            if let Value::Object(obj) = item {
                let keys: Vec<String> = obj.keys().cloned().collect();
                for k in keys {
                    modify_date_literal(obj, k);
                }
            }
        }
    }

    // Removes quotes around ISODate and backslashes before quotes
    let pipeline = serde_json::to_string_pretty(&json_value)
        .unwrap()
        .replace("\\\"", "\"")
        .replace("\"ISODate(", "ISODate(")
        .replace(")\"", ")");

    (pipeline, collection)
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
