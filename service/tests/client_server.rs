#[cfg(feature = "integration")]
mod grpc {
    use bson::{doc, Document};
    use service::translator::{
        ExcludeNamespacesOption, Namespace, SchemaCheckingMode, SelectOrderItem,
    };
    use std::error::Error;
    mod util;

    use serde_json;
    use service::metrics::{METRIC_ERRORS_TOTAL, METRIC_PANICS_TOTAL};
    use util::{get_metrics_count, run_get_namespaces_test, run_translate_sql_test};

    #[tokio::test]
    async fn test_get_namespaces_and_version() -> Result<(), Box<dyn Error>> {
        let response = run_get_namespaces_test("test", "SELECT 1").await?;

        // Extract the version from the metadata
        let version = response
            .metadata
            .expect("Metadata should be present")
            .version;

        // Check the version format
        let parts: Vec<&str> = version.splitn(2, '-').collect();

        let release = parts[0];
        let release_parts: Vec<&str> = release.split('.').collect();
        assert_eq!(
            release_parts.len(),
            3,
            "expected version {} to have three release parts",
            version
        );

        // Check each release part is non-empty
        for (i, part) in release_parts.iter().enumerate() {
            assert!(!part.is_empty(), "release part {} should not be empty", i);
        }

        if parts.len() == 2 {
            let pre_release = parts[1];
            assert!(
                !pre_release.is_empty(),
                "expected version {} to have non-empty pre-release",
                version
            );
        }

        Ok(())
    }

    #[tokio::test]
    async fn test_translate() -> Result<(), Box<dyn Error>> {
        let schema = util::get_schema("translate_test.json")?;

        let response = run_translate_sql_test(
            "bar",
            "SELECT * FROM foo",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        assert_eq!(response.db, "bar", "Expected db to be 'bar'");
        assert_eq!(
            response.target_collection, "foo",
            "Expected target_collection to be 'foo'"
        );

        let actual_pipeline: Vec<Document> = serde_json::from_str(&response.pipeline)?;
        assert_eq!(
            actual_pipeline.len(),
            1,
            "Expected pipeline to have one stage"
        );

        let expected_pipeline = doc! {
            "$project": {
                "foo": "$$ROOT",
                "_id": 0
            }
        };
        assert_eq!(
            actual_pipeline[0], expected_pipeline,
            "Pipeline stage does not match expected"
        );

        let expected_result_set_schema = doc! {
            "bsonType": "object",
            "properties": {
                "foo": {
                    "bsonType": "object",
                    "properties": {},
                    "additionalProperties": true
                }
            },
            "required": ["foo"],
            "additionalProperties": false
        };

        let actual_result_set_schema: Document = serde_json::from_str(&response.result_set_schema)?;
        assert_eq!(
            actual_result_set_schema, expected_result_set_schema,
            "Result set schema does not match expected"
        );

        assert!(
            response.select_order.is_empty(),
            "Expected select_order to be empty"
        );

        Ok(())
    }

    #[tokio::test]
    async fn translate_error() -> Result<(), Box<dyn Error>> {
        let schema = util::get_schema("translate_test.json")?;
        let initial_error_count = get_metrics_count(METRIC_ERRORS_TOTAL).await?;

        let result = run_translate_sql_test(
            "bar",
            "notavalidquery",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await;

        let final_error_count = get_metrics_count(METRIC_ERRORS_TOTAL).await?;
        assert_eq!(
            final_error_count,
            initial_error_count + 1,
            "Error count should have increased by 1"
        );

        assert!(
            result.is_err(),
            "Expected an error, but the function succeeded."
        );

        if let Err(status) = result {
            let expected_message =
                "parse error: Error 2001: Unrecognized token `notavalidquery`, expected: `SELECT`";
            assert!(
                status.message().contains(expected_message),
                "Error message does not match expected message."
            );
        }

        Ok(())
    }

    #[tokio::test]
    async fn translate_panic() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("translate_test.json")?;
        let initial_panic_count = get_metrics_count(METRIC_PANICS_TOTAL).await?;

        let response = run_translate_sql_test(
            "__test_panic",
            "__test_panic",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await;

        let final_panic_count = get_metrics_count(METRIC_PANICS_TOTAL).await?;
        assert_eq!(
            final_panic_count,
            initial_panic_count + 1,
            "Panic count should have increased by 1"
        );

        assert!(
            response.is_err(),
            "Expected an error, but the function succeeded."
        );
        if let Err(error) = response {
            let expected_message = "Internal server error";
            assert!(
                error.message().contains(expected_message),
                "Error message does not match expected message."
            );
        }

        Ok(())
    }

    #[tokio::test]
    async fn catalog_schema() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("catalog_schema_test.json")?;

        let response = run_translate_sql_test(
            "bar",
            "select * from foo",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        let expected_result_set_schema = doc! {
            "bsonType": "object",
            "properties": {
                "foo": {
                    "bsonType": "object",
                    "properties": {
                        "a": {
                            "bsonType": "double"
                        },
                    },
                    "additionalProperties": true,
                },
            },
            "required": ["foo"],
            "additionalProperties": false,
        };

        let actual_result_set_schema: Document = serde_json::from_str(&response.result_set_schema)?;
        assert_eq!(actual_result_set_schema, expected_result_set_schema);

        let expected_select_order: Vec<SelectOrderItem> = vec![SelectOrderItem {
            namespace: Some("foo".to_string()),
            field_name: "a".to_string(),
        }];

        assert_eq!(response.select_order, expected_select_order);
        Ok(())
    }

    #[tokio::test]
    async fn catalog_schema_multiple_namespaces() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("catalog_schema_multiple_namespaces_test.json")?;

        let response_bar_foo = run_translate_sql_test(
            "bar",
            "select * from foo",
            schema.clone(),
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        let expected_result_set_schema_bar_foo = doc! {
            "bsonType": "object",
            "properties": {
                "foo": {
                    "bsonType": "object",
                    "properties": {
                        "a": {
                            "bsonType": "double"
                        }
                    },
                    "additionalProperties": true
                }
            },
            "required": ["foo"],
            "additionalProperties": false,
        };

        let actual_result_set_schema_bar_foo: Document =
            serde_json::from_str(&response_bar_foo.result_set_schema)?;
        assert_eq!(
            actual_result_set_schema_bar_foo, expected_result_set_schema_bar_foo,
            "Schema for bar.foo does not match expected"
        );

        let response_bar_baz = run_translate_sql_test(
            "bar",
            "select * from baz",
            schema.clone(),
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        let expected_result_set_schema_bar_baz = doc! {
            "bsonType": "object",
            "properties": {
                "baz": {
                    "bsonType": "object",
                    "properties": {
                        "b": {
                            "bsonType": "string"
                        }
                    },
                    "additionalProperties": true
                }
            },
            "required": ["baz"],
            "additionalProperties": false,
        };

        let actual_result_set_schema_bar_baz: Document =
            serde_json::from_str(&response_bar_baz.result_set_schema)?;
        assert_eq!(
            actual_result_set_schema_bar_baz, expected_result_set_schema_bar_baz,
            "Schema for bar.baz does not match expected"
        );

        let response_qux_foo = run_translate_sql_test(
            "qux",
            "select * from foo",
            schema.clone(),
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        let expected_result_set_schema_qux_foo = doc! {
            "bsonType": "object",
            "properties": {
                "foo": {
                    "bsonType": "object",
                    "properties": {
                        "c": {
                            "bsonType": "int"
                        }
                    },
                    "additionalProperties": true
                }
            },
            "required": ["foo"],
            "additionalProperties": false,
        };

        let actual_result_set_schema_qux_foo: Document =
            serde_json::from_str(&response_qux_foo.result_set_schema)?;
        assert_eq!(
            actual_result_set_schema_qux_foo, expected_result_set_schema_qux_foo,
            "Schema for qux.foo does not match expected"
        );

        Ok(())
    }

    #[tokio::test]
    async fn catalog_schema_empty() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("empty.json")?;

        let result = run_translate_sql_test(
            "bar",
            "select * from foo",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await;

        assert!(
            result.is_err(),
            "Expected an error, but the function succeeded."
        );

        if let Err(status) = result {
            let error_message = status.message();
            assert!(
                error_message.contains(
                    "algebrize error: Error 1016: unknown collection 'foo' in database 'bar'"
                ),
                "Error message did not contain expected text: {}",
                error_message
            );
        }

        Ok(())
    }

    #[tokio::test]
    async fn select_order_multiple_elements() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("catalog_schema_select_order_test.json")?;

        let response = run_translate_sql_test(
            "bar",
            "select c, a, b from foo",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        let expected_select_order = vec![
            SelectOrderItem {
                namespace: Some("".to_string()),
                field_name: "c".to_string(),
            },
            SelectOrderItem {
                namespace: Some("".to_string()),
                field_name: "a".to_string(),
            },
            SelectOrderItem {
                namespace: Some("".to_string()),
                field_name: "b".to_string(),
            },
        ];

        assert_eq!(
            response.select_order, expected_select_order,
            "Select order does not match expected"
        );

        Ok(())
    }

    #[tokio::test]
    async fn get_namespaces() -> Result<(), Box<dyn Error>> {
        struct NamespaceTestCase {
            name: &'static str,
            db: &'static str,
            sql: &'static str,
            expected_error: bool,
            expected_namespaces: Vec<Namespace>,
        }

        let tests = vec![
            NamespaceTestCase {
                name: "no namespaces (implicit array datasource)",
                db: "test",
                sql: "SELECT 1 + 2",
                expected_error: false,
                expected_namespaces: vec![],
            },
            NamespaceTestCase {
                name: "no namespaces (explicit array datasource)",
                db: "test",
                sql: "SELECT * FROM [{'a': 1}] arr",
                expected_error: false,
                expected_namespaces: vec![],
            },
            NamespaceTestCase {
                name: "one top-level unqualified namespace",
                db: "test",
                sql: "SELECT * FROM foo",
                expected_error: false,
                expected_namespaces: vec![Namespace {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                }],
            },
            NamespaceTestCase {
                name: "one top-level qualified namespace (same as current db)",
                db: "test",
                sql: "SELECT * FROM test.foo",
                expected_error: false,
                expected_namespaces: vec![Namespace {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                }],
            },
            NamespaceTestCase {
                name: "one top-level qualified namespace (different than current db)",
                db: "test",
                sql: "SELECT * FROM db2.foo",
                expected_error: false,
                expected_namespaces: vec![Namespace {
                    db: "db2".to_string(),
                    collection: "foo".to_string(),
                }],
            },
            NamespaceTestCase {
                name: "multiple top-level qualified and unqualified namespaces",
                db: "test",
                sql: "SELECT * FROM foo, test.bar, db2.baz, xyz",
                expected_error: false,
                expected_namespaces: vec![
                    Namespace {
                        db: "db2".to_string(),
                        collection: "baz".to_string(),
                    },
                    Namespace {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    },
                    Namespace {
                        db: "test".to_string(),
                        collection: "foo".to_string(),
                    },
                    Namespace {
                        db: "test".to_string(),
                        collection: "xyz".to_string(),
                    },
                ],
            },
            NamespaceTestCase {
                name: "namespaces nested in derived table",
                db: "test",
                sql: "SELECT * FROM (SELECT * FROM foo) d",
                expected_error: false,
                expected_namespaces: vec![Namespace {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                }],
            },
            NamespaceTestCase {
                name: "namespaces nested in subquery expression",
                db: "test",
                sql: "SELECT (SELECT bar.x IS INT FROM bar LIMIT 1) FROM foo",
                expected_error: false,
                expected_namespaces: vec![
                    Namespace {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    },
                    Namespace {
                        db: "test".to_string(),
                        collection: "foo".to_string(),
                    },
                ],
            },
            NamespaceTestCase {
                name: "namespaces set query",
                db: "test",
                sql: "SELECT * FROM foo UNION ALL SELECT * FROM bar",
                expected_error: false,
                expected_namespaces: vec![
                    Namespace {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    },
                    Namespace {
                        db: "test".to_string(),
                        collection: "foo".to_string(),
                    },
                ],
            },
            NamespaceTestCase {
                name: "relaxed schema checking",
                db: "test",
                sql: "SELECT * FROM foo ORDER BY _id",
                expected_error: false,
                expected_namespaces: vec![Namespace {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                }],
            },
            NamespaceTestCase {
                name: "syntactically invalid query errors",
                db: "test",
                sql: "SELECT * FROM [{'a': 1}]", // no alias is syntactically invalid
                expected_error: true,
                expected_namespaces: vec![],
            },
            NamespaceTestCase {
                name: "semantically invalid query does not error",
                db: "test",
                sql: "SELECT x FROM foo JOIN bar", // x is ambiguous
                expected_error: false,
                expected_namespaces: vec![
                    Namespace {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    },
                    Namespace {
                        db: "test".to_string(),
                        collection: "foo".to_string(),
                    },
                ],
            },
        ];

        for test in tests {
            let result = run_get_namespaces_test(&test.db, &test.sql).await;

            match result {
                Ok(response) => {
                    if test.expected_error {
                        panic!("Expected error for test '{}', but got success", test.name);
                    }
                    assert_eq!(
                        response.namespaces, test.expected_namespaces,
                        "Test '{}' failed: namespaces don't match",
                        test.name
                    );
                }
                Err(e) => {
                    if !test.expected_error {
                        panic!("Unexpected error for test '{}': {:?}", test.name, e);
                    }
                }
            }
        }

        Ok(())
    }

    #[tokio::test]
    async fn relaxed_schema_checking() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("schema_checking_mode_test.json")?;

        let response = run_translate_sql_test(
            "test",
            "select studentid from grades where score > 80",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::Relaxed,
        )
        .await?;

        assert_eq!(response.db, "test", "Expected targetDB to be 'test'");
        assert_eq!(
            response.target_collection, "grades",
            "Expected targetCollection to be 'grades'"
        );

        let pipeline: Vec<bson::Document> = serde_json::from_str(&response.pipeline)?;

        assert_eq!(pipeline.len(), 4, "Expected pipeline to have four stages");

        Ok(())
    }

    #[tokio::test]
    async fn strict_schema_checking() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("schema_checking_mode_test.json")?;

        let response = run_translate_sql_test(
            "test",
            "select studentid from grades where score > 80",
            schema,
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await;

        assert!(
            response.is_err(),
            "Expected an error, but the function succeeded."
        );
        if let Err(error) = response {
            println!("{:?}", error.message());
            let expected_message =
                "algebrize error: Error 1005: Invalid use of `Gt` due to incomparable types: \
                `any type` cannot be compared to `int`. An `any type` schema may indicate that \
                schema is not set for the relevant collection or field. \
                Please verify that the schema is set as expected.\n\t\
                Caused by:\n\t\
                invalid comparison for Gt: \
                Any cannot be compared to Atomic(Integer)";

            assert!(
                error.message().contains(expected_message),
                "Error message does not match expected message."
            );
        }

        Ok(())
    }

    #[tokio::test]
    async fn exclude_namespaces() -> Result<(), Box<dyn std::error::Error>> {
        let schema = util::get_schema("exclude_namespaces_test.json")?;

        let response = run_translate_sql_test(
            "bar",
            "select * from foo",
            schema,
            ExcludeNamespacesOption::ExcludeNamespacesUnspecified,
            SchemaCheckingMode::StrictUnspecified,
        )
        .await?;

        assert_eq!(response.db, "bar", "Expected targetDB to be 'bar'");
        assert_eq!(
            response.target_collection, "foo",
            "Expected targetCollection to be 'foo'"
        );

        let pipeline: Vec<serde_json::Value> = serde_json::from_str(&response.pipeline)?;

        assert_eq!(pipeline.len(), 2, "Expected pipeline to have two stages");

        let expected_first_stage = doc! {
            "$project": {
                "foo": "$$ROOT",
                "_id": 0
            }
        };
        let actual_pipeline: Vec<Document> = serde_json::from_str(&response.pipeline)?;

        assert_eq!(
            actual_pipeline[0], expected_first_stage,
            "First pipeline stage does not match expected"
        );

        let expected_second_stage = doc! {
            "$replaceWith": "$foo"
        };
        assert_eq!(
            actual_pipeline[1], expected_second_stage,
            "Second pipeline stage does not match expected $replaceWith"
        );

        let expected_result_set_schema = doc! {
            "bsonType": "object",
            "properties": {
                "a": {
                    "bsonType": "double"
                }
            },
            "required": ["a"],
            "additionalProperties": false
        };

        let actual_result_set_schema: bson::Document =
            serde_json::from_str(&response.result_set_schema)?;
        assert_eq!(
            actual_result_set_schema, expected_result_set_schema,
            "Result set schema does not match expected"
        );

        Ok(())
    }
}
