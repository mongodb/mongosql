use mongodb::{
    bson::{doc, Bson, Document},
    sync::Client,
};
use mongosql::Translation;
use serde::{Deserialize, Serialize};
use std::{
    collections::{BTreeMap, HashSet},
    fs,
    io::Read,
    path::PathBuf,
};

use super::Error;

/// load_query_test_files loads the YAML files in the provided `dir` into a Vec
/// of QueryYamlTestFile structs.
pub fn load_query_test_files(dir: PathBuf) -> Result<Vec<QueryYamlTestFile>, Error> {
    let entries = fs::read_dir(dir).map_err(Error::InvalidDirectory)?;

    entries
        .map(|entry| match entry {
            Ok(de) => parse_query_yaml_file(de.path()),
            Err(e) => Err(Error::InvalidFilePath(e)),
        })
        .collect::<Result<Vec<QueryYamlTestFile>, Error>>()
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueryYamlTestFile {
    pub catalog_data: Option<BTreeMap<String, BTreeMap<String, Vec<Bson>>>>,
    pub catalog_schema: Option<BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>>,
    pub tests: Vec<QueryTest>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueryTest {
    pub description: String,
    pub skip_reason: Option<String>,
    pub current_db: Option<String>,
    pub query: String,
    pub exclude_namespaces: Option<bool>,
    pub should_compile: Option<bool>,
    pub result: Option<Vec<Document>>,
    pub parse_error: Option<String>,
    pub algebrize_error: Option<String>,
    pub catalog_error: Option<String>,
    pub allow_order_by_missing: Option<bool>,
    pub type_compare: Option<bool>,
}

/// parse_query_yaml_file parses a YAML file into a QueryYamlTestFile struct.
pub fn parse_query_yaml_file(path: PathBuf) -> Result<QueryYamlTestFile, Error> {
    let mut f = fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let yaml: QueryYamlTestFile = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml((format!("in file: {:?}: ", path), e)))?;
    Ok(yaml)
}

/*
 * The following functions are used to compare the results of a query test. Why are they necessary?
 * Unfortunately, NaN != NaN, so we need to do some special handling.
 */

/// Compare arrays of Bson values, allowing for NaN == NaN == true. This is not ideal (ballooning O). Because arrays may contain duplicate values,
/// we MUST check every value in the expected array against every value in the actual array. Using the HashSet
/// to mark seen indices, we can ensure that we don't check the same value twice.
/// Since the query tests are small, this shouldn't be much of an impact.
pub fn compare_arrays(expected: &[Bson], actual: &[Bson], type_compare: bool) -> bool {
    if expected.len() != actual.len() {
        return false;
    }
    let mut seen_indices = HashSet::new();
    expected.iter().all(|e| {
        actual.iter().enumerate().any(|(i, a)| {
            if seen_indices.contains(&i) {
                return false;
            }
            if let Bson::Document(d) = e {
                if let Bson::Document(ad) = a {
                    if compare_documents(d, ad, type_compare) {
                        return seen_indices.insert(i);
                    }
                } else {
                    return false;
                }
            }
            if let Bson::Array(ea) = e {
                if let Bson::Array(aa) = a {
                    if compare_arrays(ea, aa, type_compare) {
                        return seen_indices.insert(i);
                    }
                } else {
                    return false;
                }
            }
            if type_compare {
                if e.element_type() == a.element_type() {
                    return seen_indices.insert(i);
                } else {
                    return false;
                }
            }
            if let Bson::Double(d) = e {
                if let Bson::Double(ad) = a {
                    if d.is_nan() && ad.is_nan() || d == ad {
                        return seen_indices.insert(i);
                    }
                } else {
                    return false;
                }
            }
            if e == a {
                return seen_indices.insert(i);
            }
            false
        })
    })
}

/// Compare documents, allowing for NaN == NaN == true, by first checking to make sure they have the same number of keys, then iterating
/// through one document and getting the matching key from the other. Because there can't be duplicate keys within a document (at the same level),
/// this is a much more simple comparison than arrays.
pub fn compare_documents(expected: &Document, actual: &Document, type_compare: bool) -> bool {
    if expected.len() != actual.len() {
        return false;
    }
    expected.iter().all(|(ek, ev)| {
        actual.iter().any(|(ak, av)| {
            if let Bson::Document(d) = ev {
                if let Bson::Document(ad) = av {
                    return compare_documents(d, ad, type_compare);
                } else {
                    return false;
                }
            }

            if let Bson::Array(a) = ev {
                if let Bson::Array(aa) = av {
                    return compare_arrays(a, aa, type_compare);
                } else {
                    return false;
                }
            }
            if type_compare {
                return ek == ak && ev.element_type() == av.element_type();
            }
            if let Bson::Double(d) = ev {
                if let Bson::Double(ad) = av {
                    return d.is_infinite() && ad.is_infinite() && d.signum() == ad.signum()
                        || d.is_nan() && ad.is_nan()
                        || (d - ad).abs() <= f64::EPSILON;
                } else {
                    return false;
                }
            }
            ev == av && ek == ak
        })
    })
}

/// convert_numerics_in_results converts all numeric values in a Vec of Documents
pub fn convert_numerics_in_results(results: Vec<Document>) -> Vec<Document> {
    results.into_iter().map(convert_numerics).collect()
}

/// convert_numerics attempts to convert all i64 values to i32 values, if possible.
/// https://jira.mongodb.org/browse/RUST-1692
pub fn convert_numerics(doc: Document) -> Document {
    doc.into_iter()
        .map(|(k, v)| match v {
            Bson::Document(d) => (k, Bson::Document(convert_numerics(d))),
            Bson::Int64(d) => {
                if d <= i32::MAX as i64 || d >= i32::MIN as i64 {
                    (k, Bson::Int32(d as i32))
                } else {
                    (k, Bson::Int64(d))
                }
            }
            _ => (k, v),
        })
        .collect()
}

/// run_query runs the provided query with the provided client and returns the results.
pub fn run_query(client: &Client, translation: Translation) -> Result<Vec<Document>, Error> {
    let pipeline = translation
        .pipeline
        .as_array()
        .unwrap()
        .iter()
        .map(|d| d.as_document().unwrap().to_owned())
        .collect::<Vec<Document>>();

    let result = if let Some(coll) = translation.target_collection {
        client
            .database(translation.target_db.as_str())
            .collection::<Document>(coll.as_str())
            .aggregate(pipeline)
            .run()
    } else {
        client
            .database(translation.target_db.as_str())
            .aggregate(pipeline)
            .run()
    }
    .map_err(Error::MongoDBErr)?;

    Ok(result.into_iter().map(|d| d.unwrap()).collect())
}
