use crate::internal_integration_tests::consts::{LARGE_PARTITIONS, SMALL_PARTITIONS};

macro_rules! test_get_partitions {
    ($test_name:ident, expected = $expected:expr, input_db = $input_db:expr, input_coll = $input_coll:expr $(, ignore = $ignore:expr)?) => {

        #[cfg(feature = "integration")]
        #[tokio::test]
        $(#[ignore = $ignore])?
        async fn $test_name() {
            use super::get_mdb_collection;
            use crate::partitioning::{get_partitions, Partition};

            #[allow(unused)]
            use mongodb::bson::Bson;
            #[allow(unused)]
            use test_utils::schema_builder_library_integration_test_consts::{
                DATA_DOC_SIZE_IN_BYTES, LARGE_COLL_NAME, LARGE_ID_MIN, NONUNIFORM_DB_NAME,
                NUM_DOCS_PER_LARGE_PARTITION, SMALL_COLL_NAME, SMALL_COLL_SIZE_IN_MB, SMALL_ID_MIN,
                UNIFORM_DB_NAME, UNITARY_COLL_NAME,
            };

            let coll = get_mdb_collection($input_db, $input_coll).await;

            let expected: Vec<Partition> = $expected.to_vec();

            let actual_res = get_partitions(&coll).await;
            match actual_res {
                Err(err) => assert!(false, "unexpected error: {err:?}"),
                Ok(actual_partitions) => {
                    let a_len = actual_partitions.len();
                    let e_len = expected.len();
                    // The large tests here have potential to fail, be aware. Failures should be
                    // uncommon.
                    assert!(
                        a_len >= e_len,
                        "actual partition #{a_len} is not greater than or equal to expected partition #{e_len}"
                    )
                }
            }
        }
    };
}

test_get_partitions!(
    uniform_small,
    expected = SMALL_PARTITIONS,
    input_db = UNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_partitions!(
    uniform_large,
    expected = LARGE_PARTITIONS,
    input_db = UNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME,
    ignore = "SQL-2338"
);

test_get_partitions!(
    uniform_unit,
    expected = vec![Partition {
        min: Bson::MinKey,
        max: Bson::MaxKey,
        is_max_bound_inclusive: true
    }],
    input_db = UNIFORM_DB_NAME,
    input_coll = UNITARY_COLL_NAME
);

test_get_partitions!(
    nonuniform_small,
    expected = SMALL_PARTITIONS,
    input_db = NONUNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_partitions!(
    nonuniform_large,
    expected = LARGE_PARTITIONS,
    input_db = NONUNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME,
    ignore = "SQL-2338"
);

#[cfg(feature = "integration")]
#[tokio::test]
async fn one_doc_collection() {
    use super::get_mdb_collection;
    use crate::{errors::Error, partitioning::get_partitions};
    use test_utils::schema_builder_library_integration_test_consts::UNIFORM_DB_NAME;

    let coll = get_mdb_collection(UNIFORM_DB_NAME, "empty").await;

    let actual_res = get_partitions(&coll).await;
    match actual_res {
        Err(Error::NoCollectionStats(_)) => {} // expect the NoBounds errors
        Err(err) => panic!("unexpected error: {err:?}"),
        Ok(actual) => panic!("expected error but got: {actual:?}"),
    }
}

#[cfg(feature = "integration")]
#[tokio::test]
async fn empty_collection() {
    use super::get_mdb_collection;
    use crate::{errors::Error, partitioning::get_partitions};
    use test_utils::schema_builder_library_integration_test_consts::UNIFORM_DB_NAME;

    let coll = get_mdb_collection(UNIFORM_DB_NAME, "empty").await;

    let actual_res = get_partitions(&coll).await;
    match actual_res {
        Err(Error::NoCollectionStats(_)) => {} // expect the NoBounds errors
        Err(err) => panic!("unexpected error: {err:?}"),
        Ok(actual) => panic!("expected error but got: {actual:?}"),
    }
}
