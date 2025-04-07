use std::{
    borrow::Cow,
    fs::{DirEntry, File, OpenOptions},
    io::Write,
    path::Path,
};

use crate::SchemaDerivationYamlTestFile;

// tests we handle
const QUERY_TEST: &str = "query_tests";
const E2E_TEST: &str = "e2e_tests";
const INDEX_TEST: &str = "index_usage_tests";
const ERROR_TEST: &str = "errors";
const SCHEMA_DERIVATION_TESTS: &str = "schema_derivation_tests";

// tests we don't handle
const REWRITE_TEST: &str = "rewrite_tests";
const TYPE_CONSTRAINT_TESTS: &str = "type_constraint_tests";

// we also want to filter out hte correctness catalogs
const CORRECTNESS_CATALOG: &str = "correctness_catalog";

/// sanitize_description sanitizes test names such that they may be used as function names in generated test cases
fn sanitize_description(description: &str) -> String {
    let mut description = description.replace([' ', '-', '(', ')', '\'', ',', '.', ';'], "_");
    description = description.replace("=>", "arrow");
    description = description.replace('$', "dollar_sign");
    description = description.replace('/', "or");
    description = description.replace('?', "question_mark");
    description = description.replace('=', "equals");
    description = description.replace('*', "star");
    description.replace('|', "pipe_")
}

/// normalize_path strips the path of unecessary information and accounts for os
/// specific encoding. This function is used for generating test file names
fn normalize_path(entry: &DirEntry) -> String {
    entry
        .path()
        .into_os_string()
        .into_string()
        .unwrap()
        .replace("../tests/", "")
        .replace("../tests\\", "")
        .replace('/', "_")
        .replace("\\\\", "_")
        .replace('\\', "_")
        .replace(".yml", "")
}

/// ProcessorType is an enum that represents the type of test file being processed
#[derive(Debug, PartialEq)]
pub enum ProcessorType {
    // query tests
    Query,
    // e2e tests
    E2E,
    // error tests
    Error,
    // index usage tests
    Index,
    // schema derivation tests
    SchemaDerivation,
    // unhandled test types
    Unhandled,
    // unknown test types
    None,
}

impl From<&Cow<'_, str>> for ProcessorType {
    fn from(s: &Cow<'_, str>) -> Self {
        if s.contains(INDEX_TEST) {
            Self::Index
        } else if s.contains(QUERY_TEST) {
            Self::Query
        } else if s.contains(E2E_TEST) {
            Self::E2E
        } else if s.contains(ERROR_TEST) {
            Self::Error
        } else if s.contains(REWRITE_TEST)
            || s.contains(TYPE_CONSTRAINT_TESTS)
            || s.contains(CORRECTNESS_CATALOG)
        {
            Self::Unhandled
        } else if s.contains(SCHEMA_DERIVATION_TESTS) {
            Self::SchemaDerivation
        } else {
            Self::None
        }
    }
}

struct Processor {
    processor_type: ProcessorType,
    mod_file: File,
    entry: DirEntry,
    path: String,
    out_dir: String,
    file_name: String,
}

pub struct TestProcessor;

impl TestProcessor {
    pub fn process(entry: DirEntry, mod_file: File, out_dir: &str) {
        let path = normalize_path(&entry);
        Processor {
            processor_type: ProcessorType::from(&entry.path().to_string_lossy()),
            mod_file,
            entry,
            file_name: format!("{}.rs", path),
            path,
            out_dir: out_dir.to_string(),
        }
        .process();
    }
}

impl Processor {
    fn process(&mut self) {
        match self.processor_type {
            ProcessorType::Query | ProcessorType::E2E | ProcessorType::Error => {
                self.process_query();
            }
            ProcessorType::Index => {
                self.process_index();
            }
            ProcessorType::SchemaDerivation => {
                self.process_schema_derivation();
            }
            ProcessorType::Unhandled => {}
            ProcessorType::None => panic!("encountered an unknown test type"),
        }
    }

    fn write_mod_entry(&mut self) {
        writeln!(self.mod_file, "pub mod {};", self.path).unwrap();
    }

    fn process_index(&mut self) {
        self.write_mod_entry();
        let mut write_file = write_test_file(&self.file_name, &self.out_dir);
        self.write_index_header(&write_file);
        let test_file = crate::parse_index_usage_yaml_file(self.entry.path()).unwrap();
        for (index, test) in test_file.tests.iter().enumerate() {
            if test.skip_reason.is_some() {
                write!(
                    write_file,
                    include_str!("./templates/ignore_body_template"),
                    name = sanitize_description(&test.description),
                    ignore_reason = test.skip_reason.as_ref().unwrap(),
                    feature = "index",
                )
                .unwrap();
                continue;
            }
            write!(
                write_file,
                include_str!("./templates/index_usage_test_body_template"),
                name = sanitize_description(&test.description),
                current_db = test.current_db,
                query = test.query,
                description = test.description,
                index = index,
            )
            .unwrap();
        }
    }

    fn process_query(&mut self) {
        self.write_mod_entry();
        let mut write_file = write_test_file(&self.file_name, &self.out_dir);
        self.write_query_header(&write_file);
        let test_file = crate::parse_query_yaml_file(self.entry.path()).unwrap();
        for (index, test) in test_file.tests.iter().enumerate() {
            if test.skip_reason.is_some() {
                write!(
                    write_file,
                    include_str!("./templates/ignore_body_template"),
                    name = sanitize_description(&test.description),
                    ignore_reason = test.skip_reason.as_ref().unwrap(),
                    feature = if self.processor_type == ProcessorType::Query {
                        "query"
                    } else if self.processor_type == ProcessorType::E2E {
                        "e2e"
                    } else {
                        "error"
                    },
                )
                .unwrap();
                continue;
            }

            // if this is an error test that has a catalog error, the test is expect
            // to panic, so we use a different template
            if let Some(catalog_error) = test.catalog_error.as_ref() {
                let catalog_error = format!("{:?}", catalog_error);
                write!(
                    write_file,
                    include_str!("./templates/error_test_with_panic_body_template"),
                    name = sanitize_description(&test.description),
                    expected_panic = catalog_error,
                )
                .unwrap();
            } else {
                write!(
                    write_file,
                    include_str!("./templates/query_test_body_template"),
                    name = sanitize_description(&test.description),
                    index = index,
                    feature = if self.processor_type == ProcessorType::Query {
                        "query"
                    } else if self.processor_type == ProcessorType::E2E {
                        "e2e"
                    } else {
                        "error"
                    },
                    allow_order_by_missing = test.allow_order_by_missing.unwrap_or(true),
                )
                .unwrap();
            }
        }
    }

    fn process_schema_derivation(&mut self) {
        self.write_mod_entry();
        let mut write_file = write_test_file(&self.file_name, &self.out_dir);
        self.write_schema_derivation_header(&write_file);
        let test_file = crate::parse_schema_derivation_yaml_file(self.entry.path()).unwrap();
        match test_file {
            SchemaDerivationYamlTestFile::Single(_) => {
                write!(
                    write_file,
                    include_str!("./templates/schema_derivation_test_body_template"),
                    name = &self.file_name.replace(".rs", ""),
                    index = 0,
                )
                .unwrap();
            }
            SchemaDerivationYamlTestFile::Multiple(spec_query_test) => {
                for (index, test) in spec_query_test.tests.iter().enumerate() {
                    let description = test
                        .description
                        .clone()
                        .expect("missing description for spec query schema derivation test");
                    write!(
                        write_file,
                        include_str!("./templates/schema_derivation_test_body_template"),
                        name = sanitize_description(description.as_str()),
                        index = index
                    )
                    .unwrap();
                }
            }
        }
    }

    #[allow(clippy::format_in_format_args)]
    // we want the debug version of the canonicalized path to play nicely
    // with the template
    fn write_query_header(&self, mut file: &File) {
        write!(
            file,
            include_str!("./templates/query_test_header_template"),
            path = format!(
                "{:?}",
                self.entry.path().canonicalize().unwrap().to_string_lossy()
            ),
        )
        .unwrap();
    }

    #[allow(clippy::format_in_format_args)]
    // we want the debug version of the canonicalized path to play nicely
    // with the template
    fn write_schema_derivation_header(&self, mut file: &File) {
        write!(
            file,
            include_str!("./templates/schema_derivation_test_header_template"),
            path = format!(
                "{:?}",
                self.entry.path().canonicalize().unwrap().to_string_lossy()
            ),
        )
        .unwrap();
    }

    #[allow(clippy::format_in_format_args)]
    // we want the debug version of the canonicalized path to play nicely
    // with the template
    fn write_index_header(&self, mut file: &File) {
        write!(
            file,
            include_str!("./templates/index_usage_test_header_template"),
            path = format!(
                "{:?}",
                self.entry.path().canonicalize().unwrap().to_string_lossy()
            ),
        )
        .unwrap();
    }
}

fn write_test_file(file_name: &str, out_dir: &str) -> File {
    let file_path = Path::new(out_dir).join(file_name);
    match OpenOptions::new().create(true).append(true).open(file_path) {
        Ok(file) => file,
        Err(e) => panic!("{e}: {file_name} {out_dir}"),
    }
}
