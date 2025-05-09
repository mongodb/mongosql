// use std::ffi::OsStr;
// use std::fs;
// use std::fs::read_dir;
// use std::fs::File;
// use std::fs::ReadDir;
// use std::io::Write;
// use test_utils::TestProcessor;
use sql_engines_common_test_infra::{
    generate_tests, Error, TestGenerator, TestGeneratorFactory, YamlTestCase,
};
use test_utils::{IndexUsageTestGenerator, QueryTestGenerator};

const GENERATED_DIRECTORY: &str = "src/generated";
const GENERATED_MOD: &str = "src/generated/mod.rs";
const YAML_TEST_DIR: &str = "../tests";

fn main() {
    let res = generate_tests(
        GENERATED_DIRECTORY,
        GENERATED_MOD,
        YAML_TEST_DIR,
        &MongoSqlTestGeneratorFactory {},
    );

    if let Err(e) = res {
        panic!("test generation failed: {e:?}");
    }
}

// TODO:
//   - add YamlTestCase type aliases, TestGenerator impls, and a TestGeneratorFactory impl for the
//     various test types supported here (specifically only index, errors, e2e, and spec/query)
//   - update main to simply call out to generate_tests
//   - migrate schema_derivation tests to use legacy code path

// tests we handle
const E2E_TEST: &str = "e2e_tests";
const ERROR_TEST: &str = "errors";
const INDEX_TEST: &str = "index_usage_tests";
const QUERY_TEST: &str = "query_tests";

// tests we don't handle
const REWRITE_TEST: &str = "rewrite_tests";
const SCHEMA_DERIVATION_TESTS: &str = "schema_derivation_tests";
const TYPE_CONSTRAINT_TESTS: &str = "type_constraint_tests";

struct MongoSqlTestGeneratorFactory;

impl TestGeneratorFactory for MongoSqlTestGeneratorFactory {
    fn create_test_generator(
        &self,
        path: String,
    ) -> sql_engines_common_test_infra::Result<impl TestGenerator> {
        let gen: impl TestGenerator = if path.contains(E2E_TEST) {
            QueryTestGenerator {
                feature: "e2e".to_string(),
            }
        } else if path.contains(ERROR_TEST) {
            QueryTestGenerator {
                feature: "error".to_string(),
            }
        } else if path.contains(INDEX_TEST) {
            IndexUsageTestGenerator
        } else if path.contains(QUERY_TEST) {
            QueryTestGenerator {
                feature: "query".to_string(),
            }
        } else if path.contains(REWRITE_TEST) {
            return Err(Error::UnhandledTestType(path));
        } else {
            return Err(Error::UnknownTestType(path));
        };

        Ok(gen)
    }
}

// fn main() {
//     let remove = fs::remove_dir_all(GENERATED_DIRECTORY);
//     let create = fs::create_dir(GENERATED_DIRECTORY);
//     let mut mod_file = fs::OpenOptions::new()
//         .append(true)
//         .create(true)
//         .open(GENERATED_MOD)
//         .unwrap();
//     write!(mod_file, include_str!("templates/mod_header_template")).unwrap();
//
//     match (remove, create) {
//         (Ok(_), Ok(_)) => {}
//         // in this case, it may be the first time run so there is nothing to delete.
//         // No reason to panic here.
//         (Err(_), Ok(_)) => {}
//         (Ok(_), Err(why)) => panic!("failed to create test direcotry: {why:?}"),
//         (Err(delete_err), Err(create_err)) => panic!(
//             "failed to delete and create test directory:\n{:?}\n{:?}",
//             delete_err, create_err
//         ),
//     }
//
//     let test_data_directories = read_dir(YAML_TEST_DIR).unwrap();
//     traverse(test_data_directories, GENERATED_DIRECTORY, mod_file);
// }
//
// // traverse the tests directory, finding all yml files.
// // process each yml file and create a test file for each test case
// fn traverse(path: ReadDir, out_dir: &str, mod_file: File) {
//     for entry in path {
//         let entry = entry.unwrap();
//         if entry.file_type().unwrap().is_dir() {
//             traverse(
//                 read_dir(entry.path()).unwrap(),
//                 out_dir,
//                 mod_file.try_clone().unwrap(),
//             )
//         } else if entry.file_type().unwrap().is_file()
//             && entry.path().extension() == Some(OsStr::new("yml"))
//         {
//             TestProcessor::process(entry, mod_file.try_clone().unwrap(), out_dir);
//         }
//     }
// }
