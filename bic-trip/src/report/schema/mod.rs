#[cfg(test)]
mod test;

use mongosql::{
    json_schema,
    schema::{Document, Schema},
};
use std::collections::HashMap;

pub fn process_schemata(schema: HashMap<String, HashMap<String, Schema>>) -> SchemaAnalysis {
    let mut analysis = SchemaAnalysis::default();
    for (db, v) in schema {
        let mut database_analysis = DatabaseAnalysis {
            database_name: db.clone(),
            ..Default::default()
        };
        for (coll, schema) in v {
            let json_schema: json_schema::Schema = schema.clone().try_into().unwrap();
            database_analysis
                .schema
                .entry(coll.clone())
                .or_insert(serde_json::to_string(&json_schema).unwrap());
            let mut collection_analysis = CollectionAnalysis {
                ..Default::default()
            };
            collection_analysis.collection_name.clone_from(&coll);
            process_schema(coll, schema, &mut collection_analysis, 0);
            collection_analysis
                .arrays
                .retain(|k, _| !collection_analysis.arrays_of_arrays.contains_key(k));
            collection_analysis
                .objects
                .retain(|k, _| !collection_analysis.arrays_of_objects.contains_key(k));
            database_analysis
                .collection_analyses
                .push(collection_analysis);
        }
        analysis
            .database_analyses
            .insert(db.clone(), database_analysis);
    }
    analysis
}

#[derive(Debug, Default, PartialEq, Eq)]
pub struct SchemaAnalysis {
    pub database_analyses: HashMap<String, DatabaseAnalysis>,
}

#[derive(Debug, Default)]
pub struct DatabaseAnalysis {
    pub database_name: String,
    pub schema: HashMap<String, String>,
    pub collection_analyses: Vec<CollectionAnalysis>,
}

impl PartialEq for DatabaseAnalysis {
    fn eq(&self, other: &Self) -> bool {
        let mut self_collection_analyses = self.collection_analyses.clone();
        let mut other_collection_analyses = self.collection_analyses.clone();

        self_collection_analyses.sort_by(|a, b| a.collection_name.cmp(&b.collection_name));
        other_collection_analyses.sort_by(|a, b| a.collection_name.cmp(&b.collection_name));
        self.database_name == other.database_name
            && self_collection_analyses == other_collection_analyses
    }
}

impl Eq for DatabaseAnalysis {}

// type alias to aid in tracking fields. The key is the field name and the value is the depth
type FieldTracker = HashMap<String, u16>;
type FieldTrackerWithTypes = HashMap<String, (u16, Vec<String>)>;

#[derive(Debug, Default, PartialEq, Eq, Clone)]
pub struct CollectionAnalysis {
    pub collection_name: String,
    pub arrays: FieldTracker,
    pub arrays_of_arrays: FieldTracker,
    pub arrays_of_objects: FieldTracker,
    pub objects: FieldTracker,
    pub unstable: FieldTracker,
    pub anyof: FieldTrackerWithTypes,
}

fn append_key(key: &str, k: &str) -> String {
    if key.is_empty() {
        k.to_string()
    } else {
        format!("{}.{}", key, k)
    }
}

fn process_schema(key: String, schema: Schema, analysis: &mut CollectionAnalysis, depth: u16) {
    match schema {
        mongosql::schema::Schema::Unsat | mongosql::schema::Schema::Missing => {
            // we should never see this. If we do, this is an error
            unreachable!();
        }
        mongosql::schema::Schema::Atomic(_) => {
            // do nothing, we don't need to track atomic types
        }
        mongosql::schema::Schema::Array(a) => {
            match *a {
                mongosql::schema::Schema::Array(_) => {
                    analysis
                        .arrays_of_arrays
                        .entry(key.clone())
                        .and_modify(|x| *x = depth)
                        .or_insert(depth);
                }
                mongosql::schema::Schema::Document(_) => {
                    analysis
                        .arrays_of_objects
                        .entry(key.clone())
                        .and_modify(|x| *x = depth)
                        .or_insert(depth);
                }
                _ => {}
            }

            analysis
                .arrays
                .entry(key.clone())
                .and_modify(|x| *x = depth)
                .or_insert(depth);
            process_schema(key.clone(), *a, analysis, depth);
        }
        mongosql::schema::Schema::Document(d) => {
            if d == Document::any() {
                analysis
                    .unstable
                    .entry(key.clone())
                    .and_modify(|x| *x = depth)
                    .or_insert(depth);
            } else {
                // the root of the schema is 0. We don't want to track
                // the root of the schema as a document within the schema
                if depth != 0 {
                    analysis
                        .objects
                        .entry(key.clone())
                        .and_modify(|x| *x = depth)
                        .or_insert(depth);
                }
                for (k, v) in d.keys {
                    process_schema(append_key(&key, &k), v.clone(), analysis, depth + 1);
                }
            }
        }
        mongosql::schema::Schema::AnyOf(a) => {
            let mut types = Vec::new();
            for s in &a {
                types.extend(process_anyof(s))
            }
            analysis
                .anyof
                .entry(key.clone())
                .and_modify(|(x, s)| {
                    *x = depth;
                    s.clone_from(&types);
                })
                .or_insert((depth, types));
            for s in a {
                process_schema(key.clone(), s, analysis, depth);
            }
        }
        mongosql::schema::Schema::Any => {
            analysis
                .unstable
                .entry(key.clone())
                .and_modify(|x| *x = depth)
                .or_insert(depth);
        }
    };
}

/// This function processes an anyOf schema and returns a vector of the atomic types
fn process_anyof(schema: &Schema) -> Vec<String> {
    let mut found_types = vec![];
    match schema {
        Schema::AnyOf(a) => {
            for s in a {
                found_types.extend(process_anyof(s));
            }
        }
        Schema::Array(_) => {
            found_types.push("Array".to_string());
        }
        Schema::Atomic(a) => {
            found_types.push(a.to_string());
        }
        Schema::Unsat => {
            found_types.push("Unsat".to_string());
        }
        Schema::Missing => {
            found_types.push("Missing".to_string());
        }
        Schema::Document(_) => {
            found_types.push("Object".to_string());
        }
        Schema::Any => {
            found_types.push("Any".to_string());
        }
    }
    found_types
}
