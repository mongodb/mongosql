use test_utils::schema_builder_library_integration_test_consts::{
    LARGE_COLL_NAME, NONUNIFORM_DB_NAME, NONUNIFORM_LARGE_SCHEMA, NONUNIFORM_SMALL_SCHEMA,
    NONUNIFORM_VIEW_SCHEMA, SMALL_COLL_NAME, UNIFORM_COLL_SCHEMA, UNIFORM_DB_NAME,
    UNIFORM_VIEW_SCHEMA, UNITARY_COLL_NAME, VIEW_NAME,
};

macro_rules! test_build_schema {
    ($test_name:ident, expected = $expected:expr, expected_num_init_schemas_used = $expected_num_init_schemas_used:expr, include = $include:expr, exclude = $exclude:expr, schema_collection = $schema_collection:expr) => {
        #[tokio::test]
        async fn $test_name() {
            use crate::{
                build_schema,
                NamespaceInfoWithSchema, NamespaceInfo, NamespaceType,
            };
            use mongosql::map;
            use std::collections::HashMap;
            use super::create_mdb_client;

            let client = create_mdb_client().await;

            let schema_collection: Option<String> = $schema_collection;

            // Create schema builder options.
            let options = crate::options::BuilderOptions {
                include_list: $include,
                exclude_list: $exclude,
                schema_collection: schema_collection.clone(),
                dry_run: false,
                client,
                task_semaphore: std::sync::Arc::new(tokio::sync::Semaphore::new(10)),
            };

            let schemas = build_schema(options).await.expect("failed to build schemas");

            let mut expected_schema_results: HashMap<String, NamespaceInfoWithSchema> = $expected;

            for actual_ns_schema in schemas.into_inner() {
                let actual_namespace = format!("{}.{}", actual_ns_schema.namespace_info.db_name, actual_ns_schema.namespace_info.coll_or_view_name);
                match expected_schema_results.remove(&actual_namespace) {
                    None => panic!("unexpected namespace found: {actual_namespace}"),
                    Some(expected_ns_schema) => assert_eq!(
                        expected_ns_schema,
                        actual_ns_schema,
                        "actual schema result does not match expected schema result for {actual_namespace}",
                    ),
                }
            }



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
        format!("{UNIFORM_DB_NAME}.{UNITARY_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: UNITARY_COLL_NAME.to_string(),
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
    },
    expected_num_init_schemas_used = 0,
    include = [format!("{UNIFORM_DB_NAME}.{SMALL_COLL_NAME}"),]
        .iter()
        .map(|s| glob::Pattern::new(s).unwrap())
        .collect(),
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
        format!("{UNIFORM_DB_NAME}.{UNITARY_COLL_NAME}") => NamespaceInfoWithSchema {
            namespace_info: NamespaceInfo {
                db_name: UNIFORM_DB_NAME.to_string(),
                coll_or_view_name: UNITARY_COLL_NAME.to_string(),
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
    include = [format!("{UNIFORM_DB_NAME}.*")]
        .iter()
        .map(|s| glob::Pattern::new(s).unwrap())
        .collect(),
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
    exclude = [
        format!("{UNIFORM_DB_NAME}.{LARGE_COLL_NAME}"),
        format!("{UNIFORM_DB_NAME}.{UNITARY_COLL_NAME}"),
        format!("{NONUNIFORM_DB_NAME}.{VIEW_NAME}"),
    ]
    .iter()
    .map(|s| glob::Pattern::new(s).unwrap())
    .collect(),
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
    exclude = [format!("{UNIFORM_DB_NAME}.*")]
        .iter()
        .map(|s| glob::Pattern::new(s).unwrap())
        .collect(),
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
    include = [
        format!("{UNIFORM_DB_NAME}.*"),
        format!("{NONUNIFORM_DB_NAME}.*")
    ]
    .iter()
    .map(|s| glob::Pattern::new(s).unwrap())
    .collect(),
    exclude = [
        format!("{UNIFORM_DB_NAME}.{VIEW_NAME}"),
        format!("{UNIFORM_DB_NAME}.{UNITARY_COLL_NAME}")
    ]
    .iter()
    .map(|s| glob::Pattern::new(s).unwrap())
    .collect(),
    schema_collection = None
);
