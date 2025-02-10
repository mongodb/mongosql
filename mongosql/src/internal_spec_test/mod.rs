use crate::{
    algebrizer::{Algebrizer, ClauseType},
    ast::{rewrites::rewrite_query, Query},
    catalog::Catalog,
    map, parser,
    schema::{Atomic, Document, Schema},
    set, SchemaCheckingMode,
};
use itertools::Itertools;
use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    io::Read,
    path::PathBuf,
};
use thiserror::Error;

const MAX_NUM_ARGS: usize = 3;
const REWRITE_DIR: &str = "../tests/spec_tests/rewrite_tests";
const TYPE_CONSTRAINT_DIR: &str = "../tests/spec_tests/type_constraint_tests";
const TEST_DB: &str = "test";
lazy_static! {
    static ref TYPE_TO_SCHEMA: BTreeMap<&'static str, Schema> = map! {
        "ARRAY" => Schema::Array(Box::new(Schema::Any)),
        "BINDATA" => Schema::Atomic(Atomic::BinData),
        "BOOL" => Schema::Atomic(Atomic::Boolean),
        "BSON_DATE" => Schema::Atomic(Atomic::Date),
        "BSON_TIMESTAMP" => Schema::Atomic(Atomic::Timestamp),
        "DBPOINTER" => Schema::Atomic(Atomic::DbPointer),
        "DECIMAL" => Schema::Atomic(Atomic::Decimal),
        "DOCUMENT" => Schema::Document(Document::empty()),
        "DOUBLE" => Schema::Atomic(Atomic::Double),
        "INT" => Schema::Atomic(Atomic::Integer),
        "JAVASCRIPT" => Schema::Atomic(Atomic::Javascript),
        "JAVASCRIPTWITHSCOPE" => Schema::Atomic(Atomic::JavascriptWithScope),
        "LONG" => Schema::Atomic(Atomic::Long),
        "MAXKEY" => Schema::Atomic(Atomic::MaxKey),
        "MINKEY" => Schema::Atomic(Atomic::MinKey),
        "MISSING" => Schema::Missing,
        "NULL" => Schema::Atomic(Atomic::Null),
        "OBJECTID" => Schema::Atomic(Atomic::ObjectId),
        "REGEX" => Schema::Atomic(Atomic::Regex),
        "STRING" => Schema::Atomic(Atomic::String),
        "SYMBOL" => Schema::Atomic(Atomic::Symbol),
    };
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("rewrite test failed for test {test}: expected {expected:?}, actual {actual:?}")]
    RewriteTest {
        test: String,
        expected: String,
        actual: String,
    },
    #[error("failed to read directory: {0}")]
    InvalidDirectory(String),
    #[error("failed to load file paths: {0}")]
    InvalidFilePath(String),
    #[error("failed to read file: {0}")]
    InvalidFile(String),
    #[error("unable to read file to string: {0}")]
    CannotReadFileToString(String),
    #[error("unable to deserialize YAML file: {0}")]
    CannotDeserializeYaml(String),
    #[error("{0} is not a valid schema variant")]
    InvalidSchemaVariant(String),
    #[error("unexpected parse error: {0}")]
    ParsingFailed(String),
    #[error("unexpected rewrite error: {0}")]
    RewritesFailed(String),
    #[error("rewrite should have failed with error: {0}")]
    RewriteDidNotFail(String),
    #[error("unexpected algebrize error for test '{0}': {1}")]
    AlgebrizationFailed(String, String),
    #[error("algebrization should have failed, but it didn't for test '{0}'")]
    AlgebrizationDidNotFail(String),
}

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct RewriteYamlTest {
    pub tests: Vec<RewriteTest>,
}

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct RewriteTest {
    pub description: String,
    pub query: String,
    pub result: Option<String>,
    pub error: Option<String>,
    pub skip_reason: Option<String>,
}

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct TypeConstraintYamlTest {
    pub tests: Vec<TypeConstraintTest>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub variables: Option<Variables>,
}

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Variables {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub bool: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub bson_date: Option<Vec<String>>,
    #[serde(
        skip_serializing_if = "Option::is_none",
        rename = "comparisonValidTypes"
    )]
    pub comparison_valid_types: Option<Vec<BTreeMap<String, Vec<String>>>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub document: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub int: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub numerics: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub string: Option<Vec<String>>,
}

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct TypeConstraintTest {
    pub description: String,
    pub query: String,
    pub valid_types: Vec<BTreeMap<String, Vec<String>>>,
    pub skip_reason: Option<String>,
}

/// load_file_paths reads the given directory and returns a list its file path
/// names.
pub fn load_file_paths(dir: PathBuf) -> Result<Vec<String>, Error> {
    let mut paths: Vec<String> = vec![];
    let entries = fs::read_dir(dir).map_err(|e| Error::InvalidDirectory(format!("{e:?}")))?;
    for entry in entries {
        match entry {
            Ok(de) => {
                let path = de.path();
                if path.extension().unwrap() == "yml" {
                    paths.push(path.to_str().unwrap().to_string());
                }
            }
            Err(e) => return Err(Error::InvalidFilePath(format!("{e:?}"))),
        };
    }
    Ok(paths)
}

/// parse_rewrite_yaml deserializes the given YAML file into a RewriteYamlTest
/// struct.
pub fn parse_rewrite_yaml(path: &str) -> Result<RewriteYamlTest, Error> {
    let mut f = fs::File::open(path).map_err(|e| Error::InvalidFile(format!("{e:?}")))?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(|e| Error::CannotReadFileToString(format!("{e:?}")))?;
    let yaml: RewriteYamlTest = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml(format!("{e:?}")))?;
    Ok(yaml)
}

/// parse_type_constraint_yaml deserializes the given YAML file into a
/// TypeConstraintYamlTest struct.
pub fn parse_type_constraint_yaml(path: &str) -> Result<TypeConstraintYamlTest, Error> {
    let mut f = fs::File::open(path).map_err(|e| Error::InvalidFile(format!("{e:?}")))?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(|e| Error::CannotReadFileToString(format!("{e:?}")))?;
    let yaml: TypeConstraintYamlTest = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml(format!("{e:?}")))?;
    Ok(yaml)
}

#[test]
#[ignore]
pub fn run_rewrite_tests() -> Result<(), Error> {
    use crate::ast::pretty_print::PrettyPrint;
    let paths = load_file_paths(PathBuf::from(REWRITE_DIR)).unwrap();
    for path in paths {
        let yaml = parse_rewrite_yaml(&path).unwrap();
        for test in yaml.tests {
            match test.skip_reason {
                Some(_) => continue,
                None => {
                    let parse_res = parser::parse_query(test.query.as_str());
                    let ast = parse_res.map_err(|e| Error::ParsingFailed(format!("{e:?}")))?;
                    let rewrite_res = rewrite_query(ast);
                    match test.error {
                        Some(expected_err) => match rewrite_res {
                            Ok(_) => return Err(Error::RewriteDidNotFail(expected_err)),
                            Err(e) => assert_eq!(expected_err, e.to_string()),
                        },
                        None => {
                            let expected = test.result.unwrap();
                            let actual = rewrite_res.unwrap().pretty_print().unwrap();
                            if expected != actual {
                                return Err(Error::RewriteTest {
                                    test: test.description,
                                    expected,
                                    actual,
                                });
                            }
                        }
                    };
                }
            }
        }
    }
    Ok(())
}

/// create_catalog returns a catalog that holds the given schema
/// information as well as schema information needed to algebrize
/// any query in the type constraint tests.
pub fn create_catalog(schemas: Vec<Schema>) -> Result<Catalog, Error> {
    let keys = schemas
        .into_iter()
        .enumerate()
        .map(|(index, schema)| (format!("arg{}", index + 1), schema))
        .collect::<BTreeMap<String, Schema>>();
    let required = keys.keys().cloned().collect::<BTreeSet<String>>();

    // Every type constraint test uses either one or two datasources:
    // 'foo' and/or 'bar'.
    Ok(map! {
      (TEST_DB.to_string(), "foo".to_string()).into() => Schema::Document( Document {
        keys,
        required,
        additional_properties: false,
        ..Default::default()
        }),
      (TEST_DB.to_string(), "bar".to_string()).into() => Schema::Document( Document {
        keys: map!{},
        required: set!{},
        additional_properties: false,
        ..Default::default()
        }),
    })
}

/// validate_algebrization creates an Algebrizer struct using the given type
/// information, algebrizes the given AST, and then validates the result of
/// algebrization. [`is_valid`] is true if the arguments in the query can have the
/// given types, and false if not.
///
/// This function assumes that [`types[i]`] describes the type of the i`th
/// argument in the query.
fn validate_algebrization(
    types: Vec<String>,
    ast: Query,
    is_valid: bool,
    test_name: String,
) -> Result<(), Error> {
    let schemas = types
        .clone()
        .into_iter()
        .map(|arg_type| match TYPE_TO_SCHEMA.get(arg_type.as_str()) {
            None => Err(Error::InvalidSchemaVariant(arg_type)),
            Some(schema) => Ok(schema.clone()),
        })
        .collect::<Result<Vec<Schema>, Error>>()?;
    let catalog = create_catalog(schemas)?;
    let algebrizer = Algebrizer::new(
        TEST_DB,
        &catalog,
        0u16,
        SchemaCheckingMode::Strict,
        false,
        ClauseType::Unintialized,
    );
    let plan = algebrizer
        .algebrize_query(ast.clone())
        .map_err(|e| Error::AlgebrizationFailed(test_name.clone(), format!("{e:?}")));
    match plan {
        Ok(_) => {
            if !is_valid {
                // these make it easier to figure out what is failing when we have a failing test
                // here
                dbg!(&types);
                dbg!(&ast);
                Err(Error::AlgebrizationDidNotFail(test_name))
            } else {
                Ok(())
            }
        }
        Err(e) => {
            if is_valid {
                Err(e)
            } else {
                Ok(())
            }
        }
    }
}

#[test]
#[ignore]
pub fn run_type_constraint_tests() -> Result<(), Error> {
    let paths = load_file_paths(PathBuf::from(TYPE_CONSTRAINT_DIR)).unwrap();
    for path in paths {
        let yaml = parse_type_constraint_yaml(&path).unwrap();
        // Calculate P(num_types, n) for 1 <= n <= 3, where num_types is the number
        // of types specified by the MIR.
        let all_type_permutations = (1..MAX_NUM_ARGS + 1)
            .map(|num_args| {
                (
                    num_args,
                    TYPE_TO_SCHEMA
                        .keys()
                        .map(|s| s.to_string())
                        .permutations(num_args)
                        .collect::<BTreeSet<Vec<String>>>(),
                )
            })
            .collect::<BTreeMap<usize, BTreeSet<Vec<String>>>>();
        for test in yaml.tests {
            match test.skip_reason {
                Some(_) => continue,
                None => {
                    let parse_res = parser::parse_query(test.query.as_str());
                    let ast = parse_res.map_err(|e| Error::ParsingFailed(format!("{e:?}")))?;
                    let rewrite_res = rewrite_query(ast);
                    let ast = rewrite_res.map_err(|e| Error::RewritesFailed(format!("{e:?}")))?;
                    let mut all_valid_permutations: BTreeSet<Vec<String>> = BTreeSet::new();
                    let num_args = test.valid_types.first().unwrap().len();
                    // Ensure that algebrization succeeds for all valid type
                    // combinations.
                    for valid_types in test.valid_types {
                        // Find all valid type combinations by computing a cross
                        // product of the valid types.
                        let cross_product = valid_types
                            .values()
                            .cloned()
                            .multi_cartesian_product()
                            .collect::<BTreeSet<Vec<String>>>();
                        all_valid_permutations = all_valid_permutations
                            .union(&cross_product.clone())
                            .cloned()
                            .collect::<BTreeSet<Vec<String>>>();
                        cross_product.into_iter().try_for_each(|types| {
                            validate_algebrization(
                                types,
                                ast.clone(),
                                true,
                                test.description.clone(),
                            )
                        })?;
                    }
                    // Ensure that algebrization fails for all invalid type combinations
                    all_type_permutations
                        .get(&num_args)
                        .unwrap()
                        .difference(&all_valid_permutations)
                        .try_for_each(|types| {
                            validate_algebrization(
                                types.clone(),
                                ast.clone(),
                                false,
                                test.description.clone(),
                            )
                        })?;
                }
            }
        }
    }
    Ok(())
}
