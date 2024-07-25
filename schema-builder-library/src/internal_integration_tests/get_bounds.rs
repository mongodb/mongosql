macro_rules! test_get_bounds {
    ($test_name:ident, expected_min = $expected_min:expr, expected_max = $expected_max:expr, input_db = $input_db:expr, input_coll = $input_coll:expr) => {
        #[cfg(feature = "integration")]
        #[tokio::test]
        async fn $test_name() {
            use super::create_mdb_client;
            use crate::partitioning::get_bounds;
            use mongodb::bson::{Bson, Document};
            #[allow(unused)]
            use test_utils::schema_builder_library_integration_test_consts::{
                LARGE_COLL_NAME, LARGE_ID_MIN, NONUNIFORM_DB_NAME, NUM_DOCS_IN_SMALL_COLLECTION,
                NUM_DOCS_PER_LARGE_PARTITION, SMALL_COLL_NAME, SMALL_ID_MIN, UNIFORM_DB_NAME,
            };

            let client = create_mdb_client().await;

            let db = client.database($input_db);
            let coll = db.collection::<Document>($input_coll);

            let actual_res = get_bounds(&coll).await;
            match actual_res {
                Err(err) => assert!(false, "unexpected error: {err:?}"),
                Ok((actual_min, actual_max)) => {
                    assert_eq!(
                        actual_min, $expected_min,
                        "actual min does not match expected min"
                    );
                    assert_eq!(
                        actual_max, $expected_max,
                        "actual max does not match expected max"
                    );
                }
            }
        }
    };
}

test_get_bounds!(
    uniform_small,
    expected_min = Bson::Int64(SMALL_ID_MIN),
    expected_max = Bson::Int64(SMALL_ID_MIN + *NUM_DOCS_IN_SMALL_COLLECTION - 1),
    input_db = UNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_bounds!(
    uniform_large,
    expected_min = Bson::Int64(LARGE_ID_MIN),
    expected_max = Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 4) - 1),
    input_db = UNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);

test_get_bounds!(
    nonuniform_small,
    expected_min = Bson::Int64(SMALL_ID_MIN),
    expected_max = Bson::Int64(SMALL_ID_MIN + *NUM_DOCS_IN_SMALL_COLLECTION - 1),
    input_db = NONUNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_bounds!(
    nonuniform_large,
    expected_min = Bson::Int64(LARGE_ID_MIN),
    expected_max = Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 4) - 1),
    input_db = NONUNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);

#[cfg(feature = "integration")]
#[tokio::test]
async fn empty_collection() {
    use super::create_mdb_client;
    use crate::{errors::Error, partitioning::get_bounds};
    use mongodb::bson::Document;
    use test_utils::schema_builder_library_integration_test_consts::UNIFORM_DB_NAME;

    let client = create_mdb_client().await;

    let db = client.database(UNIFORM_DB_NAME);
    let coll = db.collection::<Document>("empty");

    let actual_res = get_bounds(&coll).await;
    match actual_res {
        Err(Error::NoBounds(_)) => {} // expect the NoBounds errors
        Err(err) => assert!(false, "unexpected error: {err:?}"),
        Ok(actual) => assert!(false, "expected error but got: {actual:?}"),
    }
}
