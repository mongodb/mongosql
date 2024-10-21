use mongodb::{
    bson::{doc, Bson, Decimal128, Document},
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
    pub ordered: Option<bool>,
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
            if is_numeric(e) && is_numeric(a) {
                let d = numeric_to_double(e);
                let ad = numeric_to_double(a);
                if compare_doubles_for_test(d, ad) {
                    return seen_indices.insert(i);
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

// According to the IEEE 754 standard, a 64-bit floating-point number (double precision) has a
// significand (mantissa) precision of 53 bits. This translates to a maximum base 10 precision of
// approximately 15 to 17 significant decimal digits. To be more specific: The 53-bit significand
// represents a fraction with a maximum value of 2^52 (approximately 4.9 × 10^15). The exponent is
// 11 bits wide, allowing for a range of values from 2^(-1022) to 2^1024. With the combination of
// the significand and exponent, the 64-bit IEEE double precision floating-point format can
// represent numbers with a precision of approximately 15 to 17 significant decimal digits.
//
// Because of this, we just discard digits past the 15th decimal place. Since we only check
// approximate equality, this should be fine.
//
fn double_from_decimal128(d: &Decimal128) -> f64 {
    // Note that rust f64 parse supports:
    // 1e6, 1E6,  Infinity, -Infinity, NaN, -NaN, nan, -nan. The rust bson library only produces
    // 1E6, Infinity, -Infinity, NaN, -NaN.
    // but this code will also support 1e6 should that ever change.
    //
    // It actually seems like the yaml library may remove e/E exponents from the string representation
    // of Decimal128, so this code may not be necessary, but better safe than sorry.
    //
    // https://play.rust-lang.org/?version=stable&mode=debug&edition=2021&gist=8c516ac0c3d901512979027fd1ae9f34
    // for a test of this code.
    let truncate_decimal_string = |s: &str, precision: usize| -> String {
        let mut parts = s.split(".");
        let whole = parts.next().unwrap();
        let decimal = parts.next().unwrap();
        let mut new_decimal = String::from(decimal);
        let new_precision: i64 = precision as i64 - whole.len() as i64;
        let new_precision = if new_precision < 0 { 0 } else { new_precision };
        new_decimal.truncate(new_precision as usize);

        format!("{}.{}", whole, new_decimal)
    };

    let s = d.to_string();
    let splitter = if s.contains("E") {
        Some("E")
    } else if s.contains("e") {
        Some("e")
    } else {
        None
    };
    let s = if s.contains(".") {
        if let Some(splitter) = splitter {
            let mut parts = s.split(splitter);
            let mantissa = parts.next().unwrap();
            let exponent = parts.next().unwrap();
            let new_mantissa = truncate_decimal_string(mantissa, 15);
            let new_double_str = format!("{}{}{}", new_mantissa, splitter, exponent);
            new_double_str
        } else {
            truncate_decimal_string(&s, 15)
        }
    // the else case here is for when the number is a whole number with no decimal parts or Inf, or NaN
    } else {
        s
    };
    s.parse::<f64>().unwrap()
}

fn numeric_to_double(b: &Bson) -> f64 {
    match b {
        Bson::Int32(i) => *i as f64,
        Bson::Int64(i) => *i as f64,
        Bson::Double(d) => *d,
        // Decimal128 supports more precision than double, but is hard to work with for comparisons
        // because many important features for comparison are missing from Decimal128 in the bson
        // crate. So we convert to double here, truncating extra precision.
        Bson::Decimal128(d) => double_from_decimal128(d),
        _ => panic!("Expected numeric value, got {:?}", b),
    }
}

fn is_numeric(b: &Bson) -> bool {
    matches!(
        b,
        Bson::Int32(_) | Bson::Int64(_) | Bson::Double(_) | Bson::Decimal128(_)
    )
}

fn compare_doubles_for_test(d: f64, ad: f64) -> bool {
    d.is_infinite() && ad.is_infinite() && d.signum() == ad.signum()
        || d.is_nan() && ad.is_nan()
        || (d - ad).abs() <= f64::EPSILON
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
            if is_numeric(ev) && is_numeric(av) {
                let d = numeric_to_double(ev);
                let ad = numeric_to_double(av);
                return compare_doubles_for_test(d, ad);
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
