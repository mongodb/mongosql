macro_rules! test_get_bounds {
    ($test_name:ident, expected_min = $expected_min:expr, input_db = $input_db:expr, input_coll = $input_coll:expr) => {
        #[cfg(feature = "integration")]
        #[tokio::test]
        async fn $test_name() {
            use super::create_mdb_client;
            use crate::partitioning::get_bounds;
            use mongodb::bson::{Bson, Document};
            #[allow(unused)]
            use test_utils::schema_builder_library_integration_test_consts::{
                LARGE_COLL_NAME, LARGE_ID_MIN, NONUNIFORM_DB_NAME, SMALL_COLL_NAME, SMALL_ID_MIN,
                UNIFORM_DB_NAME,
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
                        actual_max,
                        Bson::MaxKey,
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
    input_db = UNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_bounds!(
    uniform_large,
    expected_min = Bson::Int64(LARGE_ID_MIN),
    input_db = UNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);

test_get_bounds!(
    nonuniform_small,
    expected_min = Bson::Int64(SMALL_ID_MIN),
    input_db = NONUNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_bounds!(
    nonuniform_large,
    expected_min = Bson::Int64(LARGE_ID_MIN),
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
