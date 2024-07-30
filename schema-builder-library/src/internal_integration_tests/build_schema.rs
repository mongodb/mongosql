use test_utils::schema_builder_library_integration_test_consts::{
    LARGE_COLL_NAME, NONUNIFORM_DB_NAME, NONUNIFORM_LARGE_SCHEMA, NONUNIFORM_SMALL_SCHEMA,
    NONUNIFORM_VIEW_SCHEMA, SMALL_COLL_NAME, UNIFORM_COLL_SCHEMA, UNIFORM_DB_NAME,
    UNIFORM_VIEW_SCHEMA, VIEW_NAME,
};

macro_rules! test_build_schema {
    ($test_name:ident, expected = $expected:expr, expected_num_init_schemas_used = $expected_num_init_schemas_used:expr, include = $include:expr, exclude = $exclude:expr, schema_collection = $schema_collection:expr) => {
        #[tokio::test]
        async fn $test_name() {
            use crate::{
                build_schema, SamplerAction, SamplerNotification, SchemaResult,
                NamespaceInfoWithSchema, NamespaceInfo, NamespaceType,
            };
            use mongosql::map;
            use std::collections::HashMap;
            use super::create_mdb_client;

            let client = create_mdb_client().await;

            // Create communication channels.
            let (tx_notifications, mut rx_notifications) =
                tokio::sync::mpsc::unbounded_channel::<SamplerNotification>();
            let (tx_schemata, mut rx_schemata) = tokio::sync::mpsc::unbounded_channel::<SchemaResult>();

            let schema_collection: Option<String> = $schema_collection;

            // Create schema builder options.
            let options = crate::options::BuilderOptions {
                include_list: $include,
                exclude_list: $exclude,
                schema_collection: schema_collection.clone(),
                dry_run: false,
                client,
                tx_notifications,
                tx_schemata,
            };

            // Call build_schema in a separate thread.
            tokio::spawn(build_schema(options));

            let mut expected_schema_results: HashMap<String, NamespaceInfoWithSchema> = $expected;
            let mut actual_num_init_schemas_used = 0;

            // Wait on channels to get results. Assert that received collections are as expected. Fail if
            // we get certain notifications.
            loop {
                tokio::select! {
                    notification = rx_notifications.recv() => {
                        if let Some(notification) = notification {
                            match notification.action {
                                // Ignore these acceptable notifications.
                                SamplerAction::Querying{ .. }
                                | SamplerAction::Processing{ .. }
                                | SamplerAction::Partitioning{ .. }
                                | SamplerAction::SamplingView => {}
                                // Conditionally anticipate this notification.
                                SamplerAction::UsingInitialSchema => {
                                    assert!(schema_collection.is_some(), "unexpected use of initial schema: {notification}");
                                    actual_num_init_schemas_used += 1;
                                }
                                // Fail on warnings and errors. In this test
                                // environment, neither should be encountered.
                                SamplerAction::Warning{ .. }
                                | SamplerAction::Error{ .. } => panic!("unexpected warning or error: {notification}"),
                            }
                        }
                    }
                    schema = rx_schemata.recv() => {
                        match schema {
                            Some(SchemaResult::FullSchema(actual_ns_schema)) => {
                                // When we receive a FullSchema result, we look up the namespace
                                // in the expected values map, assert that it is there, and assert
                                // that the actual SchemaResult matches the expected one.
                                let actual_namespace = format!("{}.{}", actual_ns_schema.namespace_info.db_name, actual_ns_schema.namespace_info.coll_or_view_name);
                                match expected_schema_results.remove(&actual_namespace) {
                                    None => panic!("unexpected namespace found: {actual_namespace}"),
                                    Some(expected_ns_schema) => assert_eq!(
                                        actual_ns_schema,
                                        expected_ns_schema,
                                        "actual schema result does not match expected schema result for {actual_namespace}",
                                    ),
                                }
                            }
                            Some(SchemaResult::NamespaceOnly(schema_res)) => {
                                // These tests do not use dry-run mode, so we should never
                                // receive this type of result.
                                panic!("received namespace-only result for {} {}.{}",
                                    schema_res.namespace_type,
                                    schema_res.db_name,
                                    schema_res.coll_or_view_name,
                                );
                            }
                            None => break
                        }
                    }
                }
            }

            // Ensure we used the correct number of initial schemas.
            assert_eq!(
                actual_num_init_schemas_used,
                $expected_num_init_schemas_used,
                "actual number of initial schemas used does not match expected number of initial schemas used",
            );

            // By this point, if we have not removed all mappings from
            // the expected_schema_result map then the test must fail
            // since the build_schema method did not send all expected
            // namespaces.
            assert!(
                expected_schema_results.is_empty(),
                "failed to build all expected schema; missing: {expected_schema_results:?}"
            );
        }
    };
}

test_build_schema!(
    no_include_or_exclude,
    expected = map! {
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: UNIFORM_VIEW_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_SMALL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_LARGE_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: NONUNIFORM_VIEW_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 0,
    include = vec![],
    exclude = vec![],
    schema_collection = None
);

test_build_schema!(
    include_explicit_namespaces,
    expected = map! {
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: NONUNIFORM_VIEW_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 0,
    include = vec![
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}"),
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}")
    ],
    exclude = vec![],
    schema_collection = None
);

test_build_schema!(
    include_wildcard,
    expected = map! {
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: UNIFORM_VIEW_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 0,
    include = vec![format!("{UNIFORM_DB_NAME}.*")],
    exclude = vec![],
    schema_collection = None
);

test_build_schema!(
    exclude_explicit_namespaces,
    expected = map! {
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: UNIFORM_VIEW_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_SMALL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_LARGE_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 0,
    include = vec![],
    exclude = vec![
        format!("{UNIFORM_DB_NAME}.{LARGE_COLL_NAME}"),
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}")
    ],
    schema_collection = None
);

test_build_schema!(
    exclude_wildcard,
    expected = map! {
        format!("{NONUNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_SMALL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_LARGE_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: NONUNIFORM_VIEW_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 0,
    include = vec![],
    exclude = vec![format!("{UNIFORM_DB_NAME}.*")],
    schema_collection = None
);

test_build_schema!(
    both_include_and_exclude,
    expected = map! {
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_SMALL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_LARGE_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: NONUNIFORM_VIEW_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 0,
    include = vec![
        format!("{UNIFORM_DB_NAME}.*"),
        format!("{NONUNIFORM_DB_NAME}.*")
    ],
    exclude = vec![format!("{UNIFORM_DB_NAME}.{VIEW_NAME}")],
    schema_collection = None
);

test_build_schema!(
    with_initial_schema,
    expected = map! {
        format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: UNIFORM_COLL_SCHEMA.clone(),
        },
        format!("{UNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: UNIFORM_VIEW_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{SMALL_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: SMALL_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_SMALL_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{LARGE_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: LARGE_COLL_NAME.to_string(),
                namespace_type: NamespaceType::Collection,
            },
            namespace_schema: NONUNIFORM_LARGE_SCHEMA.clone(),
        },
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: NONUNIFORM_DB_NAME.to_string(),
                coll_or_view_name: VIEW_NAME.to_string(),
                namespace_type: NamespaceType::View,
            },
            namespace_schema: NONUNIFORM_VIEW_SCHEMA.clone(),
        },
    },
    expected_num_init_schemas_used = 4, // Only used for collections, not views.
    include = vec![],
    exclude = vec![],
    schema_collection = Some("__sql_schemas".to_string())
);
