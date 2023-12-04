use mongodb::{
    bson::{doc, Bson, Document},
    sync::Client,
};
use mongosql::{
    build_catalog_from_catalog_schema,
    catalog::Catalog,
    options::{ExcludeNamespacesOption, SqlOptions},
    Translation,
};
use serde::{Deserialize, Serialize};
use std::{
    collections::{BTreeMap, HashSet},
    fs,
    io::Read,
    path::PathBuf,
    string::ToString,
};

use crate::utils::{drop_catalog_data, load_catalog_data, Error, MONGODB_URI};

#[derive(Debug, Serialize, Deserialize)]
struct QueryYamlTestFile {
    catalog_data: Option<BTreeMap<String, BTreeMap<String, Vec<Bson>>>>,
    catalog_schema: Option<BTreeMap<String, BTreeMap<String, mongosql::json_schema::Schema>>>,
    tests: Vec<QueryTest>,
}

#[derive(Debug, Serialize, Deserialize)]
struct QueryTest {
    description: String,
    skip_reason: Option<String>,
    current_db: Option<String>,
    query: String,
    exclude_namespaces: Option<bool>,
    should_compile: Option<bool>,
    result: Option<Vec<Document>>,
    parse_error: Option<String>,
    algebrize_error: Option<String>,
    catalog_error: Option<String>,
}

#[test]
#[cfg_attr(not(feature = "query-test"), ignore)]
fn run_query_tests() -> Result<(), Error> {
    // specifying --features query-test,e2e-test will flip this to
    // the e2e tests. Both features are required.
    let test_dir = if cfg!(feature = "e2e-test") {
        "../tests/e2e_tests"
    } else if cfg!(feature = "errors") {
        "../tests/errors"
    } else {
        "../tests/spec_tests/query_tests"
    };
    let test_files = load_query_test_files(PathBuf::from(test_dir))?;
    let client =
        Client::with_uri_str(MONGODB_URI.clone()).map_err(Error::CannotCreateMongoDBClient)?;
    for test_file in test_files {
        let no_catalog = test_file.catalog_data.is_none() && test_file.catalog_schema.is_none();
        // we need to drop the catalog data before loading the new one, so we collect the db names
        // here to use in a drop after the inner loop
        let catalog_dbs = test_file
            .catalog_data
            .clone()
            .map_or(vec![], |cat| cat.keys().map(|k| k.to_string()).collect());
        let catalog = if !no_catalog {
            load_catalog_data(&client, test_file.catalog_data.unwrap())?;

            build_catalog_from_catalog_schema(test_file.catalog_schema.unwrap()).unwrap()
        } else {
            // some query tests don't have a catalog, so we generate a default one
            Catalog::default()
        };

        for test in test_file.tests {
            if test.skip_reason.is_some() {
                continue;
            }

            let db = test.current_db.unwrap_or_else(|| "test".to_string());

            let exclude_namespaces_option = if let Some(true) = test.exclude_namespaces {
                ExcludeNamespacesOption::ExcludeNamespaces
            } else {
                ExcludeNamespacesOption::IncludeNamespaces
            };

            let translation = mongosql::translate_sql(
                db.as_str(),
                test.query.as_str(),
                &catalog,
                SqlOptions::new(
                    exclude_namespaces_option,
                    mongosql::SchemaCheckingMode::Strict,
                ),
            )
            .map_err(Error::Translation);

            if let Some(should_compile) = test.should_compile {
                assert_eq!(
                    should_compile,
                    translation.is_ok(),
                    "{}: unexpected compilation result",
                    test.description
                );
            }

            if let Some(parse_error) = test.parse_error {
                assert!(
                    translation
                        .as_ref()
                        .err()
                        .unwrap()
                        .to_string()
                        .contains(&parse_error),
                    "{}: unexpected parse result.\nexpected: {}\nactual: {}",
                    test.description,
                    parse_error,
                    translation.unwrap_err()
                );
                continue;
            }

            if let Some(algebrize_error) = test.algebrize_error {
                assert!(
                    translation
                        .as_ref()
                        .err()
                        .unwrap()
                        .to_string()
                        .contains(&algebrize_error),
                    "{}: unexpected algebrize result.\nexpected: {}\nactual: {}",
                    test.description,
                    algebrize_error,
                    translation.unwrap_err()
                );
                continue;
            }

            let result = run_query(&client, translation.unwrap())?;

            if let Some(expected_results) = test.result {
                // fun times here. If there isn't a catalog where the type is specified, then the
                // type is elided. Unfortunately, when coming in from serde_yaml, all positive numbers
                // are just assumed to be u64. When the driver gets this, it just puts it in an i64, not
                // bothering to see if it will fit in an i32.
                // https://jira.mongodb.org/browse/RUST-1692
                let expected_results = if no_catalog {
                    convert_numerics_in_results(expected_results)
                } else {
                    expected_results
                };
                assert_eq!(
                    expected_results.len(),
                    result.len(),
                    "{}: unexpected number of query results\nexpected results: {:?}\nactual results: {:?}",
                    test.description,
                    expected_results,
                    result
                );

                for actual in result.iter() {
                    assert!(
                        expected_results.iter().any(|expected| {
                            // because NaN != NaN, we have to use custom comparison functions
                            compare_documents(expected, actual)
                        }),
                        "unexpected query result for {}, \nexpected results: {:?}\nactual results: {:?}",
                        test.description,
                        expected_results,
                        actual
                    );
                }
            }
        }
        // drop the catalog data for this particular test file
        if !no_catalog {
            drop_catalog_data(&client, catalog_dbs.clone())?;
        }
    }
    Ok(())
}

/*
 * The following functions are used to compare the results of a query test. Why are they necessary?
 * Unfortunately, NaN != NaN, so we need to do some special handling.
 */

/// Compare arrays of Bson values, allowing for NaN == NaN == true. This is not ideal (ballooning O). Because arrays may contain duplicate values,
/// we MUST check every value in the expected array against every value in the actual array. Using the HashSet
/// to mark seen indices, we can ensure that we don't check the same value twice.
/// Since the query tests are small, this shouldn't be much of an impact.
fn compare_arrays(expected: &Vec<Bson>, actual: &Vec<Bson>) -> bool {
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
                    if compare_documents(d, ad) {
                        return seen_indices.insert(i);
                    }
                } else {
                    return false;
                }
            }
            if let Bson::Array(ea) = e {
                if let Bson::Array(aa) = a {
                    if compare_arrays(ea, aa) {
                        return seen_indices.insert(i);
                    }
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
fn compare_documents(expected: &Document, actual: &Document) -> bool {
    if expected.len() != actual.len() {
        return false;
    }
    expected.iter().all(|(ek, ev)| {
        actual.iter().any(|(ak, av)| {
            if let Bson::Document(d) = ev {
                if let Bson::Document(ad) = av {
                    return compare_documents(d, ad);
                } else {
                    return false;
                }
            }
            if let Bson::Array(a) = ev {
                if let Bson::Array(aa) = av {
                    return compare_arrays(a, aa);
                } else {
                    return false;
                }
            }
            if let Bson::Double(d) = ev {
                if let Bson::Double(ad) = av {
                    return d.is_nan() && ad.is_nan() || d == ad;
                } else {
                    return false;
                }
            }
            ev == av && ek == ak
        })
    })
}

/// load_query_test_files loads the YAML files in the provided `dir` into a Vec
/// of QueryYamlTestFile structs.
fn load_query_test_files(dir: PathBuf) -> Result<Vec<QueryYamlTestFile>, Error> {
    let entries = fs::read_dir(dir).map_err(Error::InvalidDirectory)?;

    entries
        .map(|entry| match entry {
            Ok(de) => parse_query_yaml_file(de.path()),
            Err(e) => Err(Error::InvalidFilePath(e)),
        })
        .collect::<Result<Vec<QueryYamlTestFile>, Error>>()
}

/// parse_query_yaml_file parses a YAML file into a QueryYamlTestFile struct.
fn parse_query_yaml_file(path: PathBuf) -> Result<QueryYamlTestFile, Error> {
    let mut f = fs::File::open(&path).map_err(Error::InvalidFile)?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(Error::CannotReadFileToString)?;
    let yaml: QueryYamlTestFile = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml((format!("in file: {:?}: ", path), e)))?;
    Ok(yaml)
}

/// convert_numerics_in_results converts all numeric values in a Vec of Documents
fn convert_numerics_in_results(results: Vec<Document>) -> Vec<Document> {
    results.into_iter().map(convert_numerics).collect()
}

/// convert_numerics attempts to convert all i64 values to i32 values, if possible.
/// https://jira.mongodb.org/browse/RUST-1692
fn convert_numerics(doc: Document) -> Document {
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
fn run_query(client: &Client, translation: Translation) -> Result<Vec<Document>, Error> {
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
            .aggregate(pipeline, None)
    } else {
        client
            .database(translation.target_db.as_str())
            .aggregate(pipeline, None)
    }
    .map_err(Error::MongoDBError)?;

    Ok(result.into_iter().map(|d| d.unwrap()).collect())
}
