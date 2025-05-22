use std::{
    borrow::Cow,
    fs::{DirEntry, File, OpenOptions},
    io::Write,
    path::Path,
};

// tests we handle
const SCHEMA_DERIVATION_TESTS: &str = "schema_derivation_tests";

// tests we don't handle
const REWRITE_TEST: &str = "rewrite_tests";
const TYPE_CONSTRAINT_TESTS: &str = "type_constraint_tests";

// we also want to filter out the correctness catalogs
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
    // schema derivation tests
    SchemaDerivation,
    // unhandled test types
    Unhandled,
    // unknown test types
    None,
}

impl From<&Cow<'_, str>> for ProcessorType {
    fn from(s: &Cow<'_, str>) -> Self {
        if s.contains(REWRITE_TEST)
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

    fn process_schema_derivation(&mut self) {
        self.write_mod_entry();
        let mut write_file = write_test_file(&self.file_name, &self.out_dir);
        self.write_schema_derivation_header(&write_file);
        let test_file = crate::parse_schema_derivation_yaml_file(self.entry.path()).unwrap();
        for (index, test) in test_file.tests.iter().enumerate() {
            let description = test
                .description
                .clone()
                .expect("missing description for spec query schema derivation test");
            if test.skip_reason.is_some() {
                write!(
                    write_file,
                    include_str!("./templates/ignore_body_template"),
                    name = sanitize_description(description.as_str()),
                    ignore_reason = test.skip_reason.as_ref().unwrap(),
                    feature = "schema_derivation"
                )
                .unwrap();
                continue;
            }
            write!(
                write_file,
                include_str!("./templates/schema_derivation_test_body_template"),
                name = sanitize_description(description.as_str()),
                index = index
            )
            .unwrap();
        }
    }

    #[allow(clippy::format_in_format_args)]
    // we want the debug version of the canonicalized path to play nicely
    // with the template
    fn write_schema_derivation_header(&self, mut file: &File) {
        write!(
            file,
            include_str!("../templates/schema_derivation_test_header_template"),
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
