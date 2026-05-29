macro_rules! test_get_size_counts {
    ($test_name:ident, expected_coll_sizes= $expected_coll_sizes:expr, input_db = $input_db:expr, input_coll = $input_coll:expr) => {
        #[cfg(feature = "integration")]
        #[tokio::test]
        async fn $test_name() {
            use super::create_mdb_service;
            #[allow(unused)]
            use test_utils::schema_builder_library_integration_test_consts::{
                LARGE_COLL_NAME, LARGE_ID_MIN, NONUNIFORM_DB_NAME, SMALL_COLL_NAME, SMALL_ID_MIN,
                UNIFORM_DB_NAME,
            };
            #[allow(unused)]
            use crate::{
                partitioning::{CollectionSizes, get_size_counts},
                internal_integration_tests::consts::{LARGE_COLL_SIZE_IN_BYTES, SMALL_COLL_SIZE_IN_BYTES, NUM_DOCS_IN_LARGE_COLLECTION, NUM_DOCS_IN_SMALL_COLLECTION}
            };

            let service = create_mdb_service().await;
            let actual_res = get_size_counts(&service, $input_db, $input_coll).await;
            match actual_res {
                Err(err) => assert!(false, "unexpected error: {err:?}"),
                Ok(actual_coll_sizes) => {
                    assert_eq!(
                        actual_coll_sizes, $expected_coll_sizes,
                        "actual collection size information does not match expected collection size information"
                    );
                }
            }
        }
    };
}

test_get_size_counts!(
    small_uniform,
    expected_coll_sizes = CollectionSizes {
        size: *SMALL_COLL_SIZE_IN_BYTES,
        count: Some(*NUM_DOCS_IN_SMALL_COLLECTION)
    },
    input_db = UNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_size_counts!(
    large_uniform,
    expected_coll_sizes = CollectionSizes {
        size: *LARGE_COLL_SIZE_IN_BYTES,
        count: Some(*NUM_DOCS_IN_LARGE_COLLECTION)
    },
    input_db = UNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);

test_get_size_counts!(
    small_nonuniform,
    expected_coll_sizes = CollectionSizes {
        size: *SMALL_COLL_SIZE_IN_BYTES,
        count: Some(*NUM_DOCS_IN_SMALL_COLLECTION)
    },
    input_db = NONUNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_size_counts!(
    large_nonuniform,
    expected_coll_sizes = CollectionSizes {
        size: *LARGE_COLL_SIZE_IN_BYTES,
        count: Some(*NUM_DOCS_IN_LARGE_COLLECTION)
    },
    input_db = NONUNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);

#[cfg(feature = "integration")]
#[tokio::test]
async fn empty_collection() {
    use super::create_mdb_service;
    use crate::{errors::Error, partitioning::get_size_counts};
    use mongodb::error::{CommandError, ErrorKind};
    use test_utils::schema_builder_library_integration_test_consts::UNIFORM_DB_NAME;

    let service = create_mdb_service().await;

    let actual_res = get_size_counts(&service, UNIFORM_DB_NAME, "empty").await;
    match actual_res {
        Err(Error::DataServiceError(e)) => match *e.kind {
            ErrorKind::Command(CommandError { code: 26, .. }) => {}
            ref kind => panic!("unexpected DataService error kind: {kind:?}"),
        },
        Err(err) => panic!("unexpected error: {err:?}"),
        Ok(actual) => panic!("expected error but got: {actual:?}"),
    }
}
