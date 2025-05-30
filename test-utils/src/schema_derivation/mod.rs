use agg_ast::definitions::Stage;
use mongodb::bson::doc;
use mongosql::json_schema;
use serde::{Deserialize, Serialize};
use sql_engines_common_test_infra::{
    parse_yaml_test_file, sanitize_description, Error as cti_err, TestGenerator, YamlTestCase,
    YamlTestFile,
};
use std::{fs::File, io::Write, path::PathBuf};

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SchemaDerivationTestExpectations {
    pub result_set_schema: json_schema::Schema,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SchemaDerivationTestOptions {
    pub current_db: Option<String>,
    pub current_collection: Option<String>,
    pub catalog_dbs: Option<Vec<String>>,
}

pub type SchemaDerivationTestCase =
    YamlTestCase<Vec<Stage>, SchemaDerivationTestExpectations, SchemaDerivationTestOptions>;

pub struct SchemaDerivationTestGenerator;

impl TestGenerator for SchemaDerivationTestGenerator {
    fn generate_test_file_header(
        &self,
        generated_test_file: &mut File,
        canonicalized_path: String,
    ) -> sql_engines_common_test_infra::Result<()> {
        write!(
            generated_test_file,
            include_str!("../templates/schema_derivation_test_header_template"),
            path = canonicalized_path,
        )
        .map_err(|e| {
            cti_err::Io(
                format!("failed to write schema derivation test header for '{canonicalized_path}'"),
                e,
            )
        })
    }

    fn generate_test_file_body(
        &self,
        generated_test_file: &mut File,
        original_path: PathBuf,
    ) -> sql_engines_common_test_infra::Result<()> {
        let parsed_test_file: YamlTestFile<SchemaDerivationTestCase> =
            parse_yaml_test_file(original_path)?;

        for (index, test_case) in parsed_test_file.tests.iter().enumerate() {
            let sanitized_test_name = sanitize_description(&test_case.description);
            let res = if let Some(skip_reason) = test_case.skip_reason.as_ref() {
                write!(
                    generated_test_file,
                    include_str!("../templates/ignore_body_template"),
                    feature = "schema_derivation",
                    ignore_reason = skip_reason,
                    name = sanitized_test_name,
                )
            } else {
                write!(
                    generated_test_file,
                    include_str!("../templates/schema_derivation_test_body_template"),
                    name = sanitized_test_name,
                    index = index
                )
            };
            res.map_err(|e| {
                cti_err::Io(
                    format!(
                        "failed to write schema derivation test body for test '{}'",
                        test_case.description
                    ),
                    e,
                )
            })?;
        }

        Ok(())
    }
}
