use crate::internal_integration_tests::consts::{
    LARGE_PARTITIONS, NONUNIFORM_LARGE_PARTITION_SCHEMAS, SMALL_PARTITIONS,
};
use test_utils::schema_builder_library_integration_test_consts::{
    LARGE_COLL_NAME, NONUNIFORM_DB_NAME, NONUNIFORM_LARGE_SCHEMA, NONUNIFORM_SMALL_SCHEMA,
    NONUNIFORM_VIEW_SCHEMA, SMALL_COLL_NAME, UNIFORM_COLL_SCHEMA, UNIFORM_DB_NAME,
    UNIFORM_VIEW_SCHEMA, VIEW_NAME,
};

mod for_partitions {
    use super::super::create_mdb_client;
    use super::*;
    use crate::internal_integration_tests::consts::{DEFAULT_HINT, DEFAULT_PARTITION_KEY};

    macro_rules! test_derive_schema_for_partitions {
        ($test_name:ident, $db_name:expr, $coll_name:expr, $expected_schema:expr, $partitions:expr$(,)?) => {
            #[tokio::test]
            async fn $test_name() {
                use crate::schema::derive_schema_for_partitions;
                use crate::partitioning::PartitionedCollection;
                use mongodb::bson::Document;

                let client = create_mdb_client().await;
                let db = client.database($db_name);
                let coll = db.collection::<Document>($coll_name);

                let partitions = $partitions.to_vec();

                let partitioned_collection = PartitionedCollection {
                    partitions,
                    partition_key: DEFAULT_PARTITION_KEY.to_string(),
                    hint: DEFAULT_HINT.clone(),
                };

                match derive_schema_for_partitions(
                    $db_name.to_string(),
                    &coll,
                    None,
                    &tokio::runtime::Handle::current(),
                    std::sync::Arc::new(tokio::sync::Semaphore::new(10)),
                    partitioned_collection,
                )
                .await
                {
                    Err(err) => panic!("This test should never receive an error while deriving schema for partitions: {err}"),
                    Ok(None) => panic!("This test should never have received None while deriving schema for partitions"),
                    Ok(Some(schema)) => {
                        assert_eq!(schema, *$expected_schema);
                    }
                }
            }
        };
    }

    test_derive_schema_for_partitions!(
        nonuniform_large,
        NONUNIFORM_DB_NAME,
        LARGE_COLL_NAME,
        *NONUNIFORM_LARGE_SCHEMA,
        LARGE_PARTITIONS,
    );

    test_derive_schema_for_partitions!(
        nonuniform_small,
        NONUNIFORM_DB_NAME,
        SMALL_COLL_NAME,
        *NONUNIFORM_SMALL_SCHEMA,
        SMALL_PARTITIONS,
    );

    test_derive_schema_for_partitions!(
        uniform_large,
        UNIFORM_DB_NAME,
        LARGE_COLL_NAME,
        *UNIFORM_COLL_SCHEMA,
        LARGE_PARTITIONS,
    );
    test_derive_schema_for_partitions!(
        uniform_small,
        UNIFORM_DB_NAME,
        SMALL_COLL_NAME,
        *UNIFORM_COLL_SCHEMA,
        SMALL_PARTITIONS,
    );
}

mod for_partition {
    use super::super::create_mdb_client;
    use super::*;
    use crate::internal_integration_tests::consts::{DEFAULT_HINT, DEFAULT_PARTITION_KEY};

    macro_rules! test_derive_schema_for_partition {
        ($test_name:ident, $db_name:expr, $coll_name:expr, $expected_schema:expr, $partitions:expr$(,)?) => {
            #[tokio::test]
            async fn $test_name() {
                use crate::{partitioning::Partition, schema::{derive_schema_for_partition, SinglePartition}};
                use mongodb::bson::Document;


                let client = create_mdb_client().await;
                let db = client.database($db_name);
                let coll = db.collection::<Document>($coll_name);

                let partitions: Vec<Partition> = $partitions.to_vec();

                for (ix, expected_schema) in $expected_schema.iter().enumerate() {

                    let single_partition = SinglePartition {
                        partition: partitions.get(ix).unwrap().clone(),
                        partition_key: DEFAULT_PARTITION_KEY.to_string(),
                        hint: DEFAULT_HINT.clone(),
                        partition_ix: ix
                    };

                    match derive_schema_for_partition(
                        $db_name,
                        &coll,
                        None,
                        single_partition
                    )
                    .await
                    {
                        Err(err) => panic!("This test should never receive an error while deriving schema for a partition: {err}"),
                        Ok(schema) => {
                            assert_eq!(schema, **expected_schema);
                        }
                    }
                }
            }
        };
    }

    test_derive_schema_for_partition!(
        nonuniform_large,
        NONUNIFORM_DB_NAME,
        LARGE_COLL_NAME,
        NONUNIFORM_LARGE_PARTITION_SCHEMAS,
        LARGE_PARTITIONS,
    );

    test_derive_schema_for_partition!(
        nonuniform_small,
        NONUNIFORM_DB_NAME,
        SMALL_COLL_NAME,
        [NONUNIFORM_SMALL_SCHEMA.clone()],
        SMALL_PARTITIONS,
    );

    test_derive_schema_for_partition!(
        uniform_large,
        UNIFORM_DB_NAME,
        LARGE_COLL_NAME,
        vec![UNIFORM_COLL_SCHEMA.clone(); 4],
        LARGE_PARTITIONS,
    );

    test_derive_schema_for_partition!(
        uniform_small,
        UNIFORM_DB_NAME,
        SMALL_COLL_NAME,
        [UNIFORM_COLL_SCHEMA.clone()],
        SMALL_PARTITIONS,
    );
}

mod for_views {
    use super::super::create_mdb_client;
    use super::*;

    macro_rules! test_derive_schema_for_view {
        ($test_name:ident, $db_name:expr, $view_name:expr, $expected_schema:expr$(,)?) => {
            #[tokio::test]
            async fn $test_name() {
                use crate::{collection::CollectionInfo, schema::derive_schema_for_view};

                let db = create_mdb_client().await.database($db_name);

                let CollectionInfo { views, .. } =
                    CollectionInfo::new(&db, $db_name, vec![], vec![])
                        .await
                        .unwrap_or_else(|e| panic!("Error while creating CollectionInfo: {}", e));

                let view = views
                    .into_iter()
                    .find(|view| view.name == $view_name)
                    .expect("View not found");

                match derive_schema_for_view(&view, &db).await {
                    None => {
                        panic!("This test should never receive None while deriving schema for view")
                    }
                    Some(schema) => {
                        assert_eq!(schema, *$expected_schema);
                    }
                }
            }
        };
    }

    test_derive_schema_for_view!(uniform, UNIFORM_DB_NAME, VIEW_NAME, *UNIFORM_VIEW_SCHEMA,);

    test_derive_schema_for_view!(
        nonuniform,
        NONUNIFORM_DB_NAME,
        VIEW_NAME,
        *NONUNIFORM_VIEW_SCHEMA,
    );
}
