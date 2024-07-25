macro_rules! test_get_partitions {
    ($test_name:ident, expected = $expected:expr, input_db = $input_db:expr, input_coll = $input_coll:expr) => {
        #[cfg(feature = "integration")]
        #[tokio::test]
        async fn $test_name() {
            use super::create_mdb_client;
            use crate::partitioning::{get_partitions, Partition};
            use mongodb::bson::{Bson, Document};
            #[allow(unused)]
            use test_utils::schema_builder_library_integration_test_consts::{
                DATA_DOC_SIZE_IN_BYTES, LARGE_COLL_NAME, LARGE_ID_MIN,
                NONUNIFORM_DB_NAME, NUM_DOCS_PER_LARGE_PARTITION, NUM_DOCS_IN_SMALL_COLLECTION,
                SMALL_COLL_NAME, SMALL_COLL_SIZE_IN_MB, SMALL_ID_MIN, UNIFORM_DB_NAME,
            };

            let client = create_mdb_client().await;

            let db = client.database($input_db);
            let coll = db.collection::<Document>($input_coll);

            let expected: Vec<Partition> = $expected;

            let actual_res = get_partitions(&coll).await;
            match actual_res {
                Err(err) => assert!(false, "unexpected error: {err:?}"),
                Ok(actual_partitions) => {
                    assert_eq!(
                        actual_partitions.len(), expected.len(),
                        "# of actual partitions does not match # of expected partitions: {actual_partitions:?}"
                    );
                    actual_partitions
                        .into_iter()
                        .zip(expected)
                        .enumerate()
                        .for_each(|(part_idx, (actual_part, expected_part))| {
                            assert_eq!(actual_part, expected_part, "actual partition #{part_idx} does not match expected partition #{part_idx}")
                        });
                }
            }
        }
    };
}

test_get_partitions!(
    uniform_small,
    expected = vec![Partition {
        min: Bson::Int64(SMALL_ID_MIN),
        max: Bson::Int64(SMALL_ID_MIN + *NUM_DOCS_IN_SMALL_COLLECTION - 1),
        is_max_bound_inclusive: true,
    }],
    input_db = UNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_partitions!(
    uniform_large,
    expected = vec![
        Partition {
            min: Bson::Int64(LARGE_ID_MIN),
            max: Bson::Int64(LARGE_ID_MIN + *NUM_DOCS_PER_LARGE_PARTITION),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + *NUM_DOCS_PER_LARGE_PARTITION),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 2)),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 2)),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 3)),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 3)),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 4) - 1),
            is_max_bound_inclusive: true,
        },
    ],
    input_db = UNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);

test_get_partitions!(
    nonuniform_small,
    expected = vec![Partition {
        min: Bson::Int64(SMALL_ID_MIN),
        max: Bson::Int64(SMALL_ID_MIN + *NUM_DOCS_IN_SMALL_COLLECTION - 1),
        is_max_bound_inclusive: true,
    }],
    input_db = NONUNIFORM_DB_NAME,
    input_coll = SMALL_COLL_NAME
);

test_get_partitions!(
    nonuniform_large,
    expected = vec![
        Partition {
            min: Bson::Int64(LARGE_ID_MIN),
            max: Bson::Int64(LARGE_ID_MIN + *NUM_DOCS_PER_LARGE_PARTITION),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + *NUM_DOCS_PER_LARGE_PARTITION),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 2)),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 2)),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 3)),
            is_max_bound_inclusive: false,
        },
        Partition {
            min: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 3)),
            max: Bson::Int64(LARGE_ID_MIN + (*NUM_DOCS_PER_LARGE_PARTITION * 4) - 1),
            is_max_bound_inclusive: true,
        },
    ],
    input_db = NONUNIFORM_DB_NAME,
    input_coll = LARGE_COLL_NAME
);
