use sql_engines_common_test_infra::{generate_tests, Error, TestGenerator, TestGeneratorFactory};
use test_utils::{IndexUsageTestGenerator, QueryTestGenerator, SchemaDerivationTestGenerator};

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

// tests we handle
const E2E_TEST: &str = "e2e_tests";
const ERROR_TEST: &str = "errors";
const INDEX_TEST: &str = "index_usage_tests";
const SCHEMA_DERIVATION_TESTS: &str = "schema_derivation_tests";
const QUERY_TEST: &str = "query_tests";

// tests we don't handle
const REWRITE_TEST: &str = "rewrite_tests";
const TYPE_CONSTRAINT_TESTS: &str = "type_constraint_tests";

struct MongoSqlTestGeneratorFactory;

impl TestGeneratorFactory for MongoSqlTestGeneratorFactory {
    fn create_test_generator(
        &self,
        path: String,
    ) -> sql_engines_common_test_infra::Result<Box<dyn TestGenerator>> {
        if path.contains(E2E_TEST) {
            Ok(Box::new(QueryTestGenerator {
                feature: "e2e".to_string(),
            }))
        } else if path.contains(ERROR_TEST) {
            Ok(Box::new(QueryTestGenerator {
                feature: "error".to_string(),
            }))
        } else if path.contains(INDEX_TEST) {
            Ok(Box::new(IndexUsageTestGenerator))
        } else if path.contains(SCHEMA_DERIVATION_TESTS) {
            Ok(Box::new(SchemaDerivationTestGenerator))
        } else if path.contains(QUERY_TEST) {
            Ok(Box::new(QueryTestGenerator {
                feature: "query".to_string(),
            }))
        } else if path.contains(REWRITE_TEST)
            || path.contains(SCHEMA_DERIVATION_TESTS)
            || path.contains(TYPE_CONSTRAINT_TESTS)
        {
            Err(Error::UnhandledTestType(path))
        } else {
            Err(Error::UnknownTestType(path))
        }
    }
}
