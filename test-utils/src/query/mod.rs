#![allow(clippy::result_large_err)]
use super::{check_server_version_for_test, Error};
use mongodb::{
    bson::{doc, Bson, Decimal128, Document},
    sync::Client,
};
use mongosql::Translation;
use serde::{Deserialize, Serialize};
use sql_engines_common_test_infra::{
    parse_yaml_test_file, sanitize_description, Error as cti_err, TestGenerator, YamlTestCase,
    YamlTestFile,
};
use std::{env, fs::File, io::Write, path::PathBuf};

#[derive(Debug, Serialize, Deserialize)]
pub struct QueryTestExpectations {
    pub result: Option<Vec<Document>>,
    pub parse_error: Option<String>,
    pub algebrize_error: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueryTestOptions {
    pub current_db: Option<String>,
    pub catalog_dbs: Option<Vec<String>>,
    pub exclude_namespaces: Option<bool>,
    pub should_compile: Option<bool>,
    pub allow_order_by_missing: Option<bool>,
    pub ordered: Option<bool>,
    pub type_compare: Option<bool>,
    pub min_server_version: Option<String>,
    pub max_server_version: Option<String>,
}

pub type QueryTestCase = YamlTestCase<String, QueryTestExpectations, QueryTestOptions>;

pub struct QueryTestGenerator {
    pub feature: String,
}

impl TestGenerator for QueryTestGenerator {
    fn generate_test_file_header(
        &self,
        generated_test_file: &mut File,
        canonicalized_path: String,
    ) -> sql_engines_common_test_infra::Result<()> {
        write!(
            generated_test_file,
            include_str!("../templates/query_test_header_template"),
            path = canonicalized_path,
        )
        .map_err(|e| {
            cti_err::Io(
                format!(
                    "failed to write {} test header for '{canonicalized_path}'",
                    self.feature
                ),
                e,
            )
        })
    }

    fn generate_test_file_body(
        &self,
        generated_test_file: &mut File,
        original_path: PathBuf,
    ) -> sql_engines_common_test_infra::Result<()> {
        let parsed_test_file: YamlTestFile<QueryTestCase> = parse_yaml_test_file(original_path)?;
        let server_version = env::var("MONGODB_VERSION").ok();

        for (index, test_case) in parsed_test_file.tests.iter().enumerate() {
            let sanitized_test_name = sanitize_description(&test_case.description);
            let res = if let Some(skip_reason) = test_case.skip_reason.as_ref() {
                write!(
                    generated_test_file,
                    include_str!("../templates/ignore_body_template"),
                    feature = self.feature,
                    ignore_reason = skip_reason,
                    name = sanitized_test_name,
                )
            } else if !check_server_version_for_test(
                &test_case.options.min_server_version,
                &test_case.options.max_server_version,
                &server_version,
            ) {
                write!(
                    generated_test_file,
                    include_str!("../templates/ignore_body_template"),
                    feature = self.feature,
                    ignore_reason = "server version does not meet test requirements",
                    name = sanitized_test_name,
                )
            } else {
                write!(
                    generated_test_file,
                    include_str!("../templates/query_test_body_template"),
                    feature = self.feature,
                    name = sanitized_test_name,
                    index = index,
                )
            };
            res.map_err(|e| {
                cti_err::Io(
                    format!(
                        "failed to write {} test body for test '{}'",
                        self.feature, test_case.description
                    ),
                    e,
                )
            })?;
        }

        Ok(())
    }
}

/*
 * The following functions are used to compare the results of a query test. Why are they necessary?
 * Unfortunately, NaN != NaN, so we need to do some special handling.
 */

/// Compare arrays of Bson values, allowing for NaN == NaN == true.
///
/// Note that arrays are considered ordered in MongoDB and in MongoSQL. This
/// function should never be modified to make array comparisons unordered as
/// that would break the contract of what arrays mean in MongoDB and MongoSQL.
pub fn compare_arrays(expected: &[Bson], actual: &[Bson], type_compare: bool) -> bool {
    if expected.len() != actual.len() {
        return false;
    }

    expected
        .iter()
        .zip(actual.iter())
        .all(|(expected_value, actual_value)| {
            compare_bson_values(expected_value, actual_value, type_compare)
        })
}

fn compare_bson_values(expected: &Bson, actual: &Bson, type_compare: bool) -> bool {
    match (expected, actual) {
        (Bson::Document(expected_document), Bson::Document(actual_document)) => {
            compare_documents(expected_document, actual_document, type_compare)
        }
        (Bson::Array(expected_array), Bson::Array(actual_array)) => {
            compare_arrays(expected_array, actual_array, type_compare)
        }
        _ if type_compare => expected.element_type() == actual.element_type(),
        _ if is_numeric(expected) && is_numeric(actual) => {
            compare_doubles_for_test(numeric_to_double(expected), numeric_to_double(actual))
        }
        _ => expected == actual,
    }
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

        format!("{whole}.{new_decimal}")
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
            let new_double_str = format!("{new_mantissa}{splitter}{exponent}");
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
        _ => panic!("Expected numeric value, got {b:?}"),
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

/// Compare documents, allowing for NaN == NaN == true.
///
/// First, check to make sure they have the same number of keys, then iterate through one document and getting the matching key from the other.
/// Because there can't be duplicate keys within a document (at the same level), this is a much more simple comparison than arrays.
pub fn compare_documents(expected: &Document, actual: &Document, type_compare: bool) -> bool {
    if expected.len() != actual.len() {
        return false;
    }
    expected
        .iter()
        .all(|(ek, expected_value)| match actual.get(ek) {
            Some(actual_value) => compare_bson_values(expected_value, actual_value, type_compare),
            None => false,
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

/// assert_result_sets_equal compares two result sets for equality. It ensures
/// that all expected results appear in the actual result set, and that no
/// extra results appear in the actual result set.
pub fn assert_result_sets_equal(
    mut expected: Vec<Document>,
    actual: Vec<Document>,
    type_compare: bool,
    ordered: bool,
    desc: String,
) {
    assert_eq!(
        expected.len(),
        actual.len(),
        "{}: unexpected number of query results\nexpected results: {:?}\nactual results: {:?}",
        desc,
        expected,
        actual,
    );

    if ordered {
        for (index, (e, a)) in expected.iter().zip(actual.iter()).enumerate() {
            assert!(
                // because NaN != NaN, we have to use custom comparison functions
                compare_documents(e, a, type_compare),
                "unexpected query result for {desc} at index {index}, \nexpected: {e:?}\nactual: {a:?}",
            );
        }
    } else {
        let og_expected = expected.clone();
        for a in actual.iter() {
            match expected
                .iter()
                .position(|e| compare_documents(e, a, type_compare))
            {
                None => panic!(
                    "unexpected query result for {}\nexpected results: {:?}\nactual results: {:?}",
                    desc, og_expected, actual
                ),
                Some(idx) => {
                    expected.remove(idx);
                }
            }
        }
    }
}

#[cfg(test)]
mod test {
    use super::assert_result_sets_equal;
    use mongodb::bson::doc;

    #[test]
    fn assert_result_sets_equal_unordered_empty() {
        let expected = vec![];
        let actual = vec![];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_empty() {
        let expected = vec![];
        let actual = vec![];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_singleton() {
        let expected = vec![doc! {"a": 1}];
        let actual = vec![doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_singleton() {
        let expected = vec![doc! {"a": 1}];
        let actual = vec![doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_singleton() {
        let expected = vec![doc! {"a": 1}];
        let actual = vec![doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_singleton() {
        let expected = vec![doc! {"a": 1}];
        let actual = vec![doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_multiple() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        let actual = vec![doc! {"b": 1}, doc! {"c": 1}, doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_multiple() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_multiple() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        let actual = vec![doc! {"c": 1}, doc! {"x": 1}, doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_multiple() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"c": 1}, doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_different_lengths() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_different_lengths() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_duplicates() {
        let expected = vec![doc! {"a": 1}, doc! {"a": 1}, doc! {"b": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"b": 1}, doc! {"a": 1}, doc! {"b": 1}, doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_duplicates() {
        let expected = vec![doc! {"a": 1}, doc! {"a": 1}, doc! {"b": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"a": 1}, doc! {"b": 1}, doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_duplicates() {
        let expected = vec![doc! {"a": 1}, doc! {"a": 1}, doc! {"c": 1}, doc! {"c": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"c": 1}, doc! {"a": 1}, doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_duplicates() {
        let expected = vec![doc! {"a": 1}, doc! {"a": 1}, doc! {"b": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"a": 1}, doc! {"b": 1}, doc! {"c": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_duplicate_actuals() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_duplicate_expecteds() {
        let expected = vec![doc! {"a": 1}, doc! {"a": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_duplicate_actuals() {
        let expected = vec![doc! {"a": 1}, doc! {"b": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"a": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_duplicate_expecteds() {
        let expected = vec![doc! {"a": 1}, doc! {"a": 1}];
        let actual = vec![doc! {"a": 1}, doc! {"b": 1}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_with_arrays() {
        let expected = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        let actual = vec![doc! {"a": [4, 5, 6]}, doc! {"a": [1, 2, 3]}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_with_arrays() {
        let expected = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        let actual = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_with_arrays_different_values() {
        let expected = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        let actual = vec![doc! {"a": [4, 5, 6]}, doc! {"a": [1, 20, 30]}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_with_arrays_different_values() {
        let expected = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        let actual = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 50, 6]}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_with_arrays_same_values_different_order() {
        let expected = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        let actual = vec![doc! {"a": [4, 5, 6]}, doc! {"a": [2, 1, 3]}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_with_arrays_same_values_different_order() {
        let expected = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 5, 6]}];
        let actual = vec![doc! {"a": [1, 2, 3]}, doc! {"a": [4, 6, 5]}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_with_nested_arrays() {
        let expected = vec![doc! {"a": [[1], [2, 3]]}, doc! {"a": [[4, 5, 6]]}];
        let actual = vec![doc! {"a": [[4, 5, 6]]}, doc! {"a": [[1], [2, 3]]}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_with_nested_arrays() {
        let expected = vec![doc! {"a": [[1], [2, 3]]}, doc! {"a": [[4, 5, 6]]}];
        let actual = vec![doc! {"a": [[1], [2, 3]]}, doc! {"a": [[4, 5, 6]]}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_with_nested_arrays() {
        let expected = vec![doc! {"a": [[1], [2], [3]]}, doc! {"a": [[4], [5, 6]]}];
        let actual = vec![doc! {"a": [[4], [5, 6]]}, doc! {"a": [[1], [3], [2]]}];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_with_nested_arrays() {
        let expected = vec![doc! {"a": [[1], [2, 3]]}, doc! {"a": [[4, 5, 6]]}];
        let actual = vec![doc! {"a": [[1], [2, 3]]}, doc! {"a": [[6]]}];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_with_nested_documents() {
        let expected = vec![
            doc! {"a": {"b": 1}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 3}},
        ];
        let actual = vec![
            doc! {"a": {"b": 3}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 1}},
        ];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_with_nested_documents() {
        let expected = vec![
            doc! {"a": {"b": 1}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 3}},
        ];
        let actual = vec![
            doc! {"a": {"b": 1}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 3}},
        ];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_unordered_with_nested_documents_different_values() {
        let expected = vec![
            doc! {"a": {"b": 1}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 3}},
        ];
        let actual = vec![
            doc! {"a": {"b": 3}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 4}},
        ];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    #[should_panic]
    fn assert_result_sets_not_equal_ordered_with_nested_documents_different_values() {
        let expected = vec![
            doc! {"a": {"b": 1}},
            doc! {"a": {"b": 2}},
            doc! {"a": {"b": 3}},
        ];
        let actual = vec![
            doc! {"a": {"b": 1}},
            doc! {"a": {"b": 4}},
            doc! {"a": {"b": 3}},
        ];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_unordered_with_nested_documents_same_values_different_order() {
        let expected = vec![
            doc! {"a": {"b": 1, "c": 1}},
            doc! {"a": {"b": 2, "c": 2}},
            doc! {"a": {"b": 3, "c": 3}},
        ];
        let actual = vec![
            doc! {"a": {"c": 3, "b": 3}},
            doc! {"a": {"c": 1, "b": 1}},
            doc! {"a": {"b": 2, "c": 2}},
        ];
        assert_result_sets_equal(expected, actual, false, false, "test".into());
    }

    #[test]
    fn assert_result_sets_equal_ordered_with_nested_documents_same_values_different_order() {
        let expected = vec![
            doc! {"a": {"b": 1, "c": 1}},
            doc! {"a": {"b": 2, "c": 2}},
            doc! {"a": {"b": 3, "c": 3}},
        ];
        let actual = vec![
            doc! {"a": {"c": 1, "b": 1}},
            doc! {"a": {"b": 2, "c": 2}},
            doc! {"a": {"c": 3, "b": 3}},
        ];
        assert_result_sets_equal(expected, actual, false, true, "test".into());
    }
}
