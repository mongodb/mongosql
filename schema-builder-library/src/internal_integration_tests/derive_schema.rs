use super::create_mdb_client;
use crate::internal_integration_tests::consts::{
    LARGE_PARTITIONS, NONUNIFORM_LARGE_PARTITION_SCHEMAS, SMALL_PARTITIONS,
};
use test_utils::schema_builder_library_integration_test_consts::{
    LARGE_COLL_NAME, NONUNIFORM_DB_NAME, NONUNIFORM_LARGE_SCHEMA, NONUNIFORM_SMALL_SCHEMA,
    NONUNIFORM_VIEW_SCHEMA, SMALL_COLL_NAME, UNIFORM_COLL_SCHEMA, UNIFORM_DB_NAME,
    UNIFORM_VIEW_SCHEMA, VIEW_NAME,
};

mod for_partitions {
    use super::*;

    macro_rules! test_derive_schema_for_partitions {
        ($test_name:ident, $db_name:expr, $coll_name:expr, $expected_schema:expr, $partitions:expr$(,)?) => {
            #[tokio::test]
            async fn $test_name() {
                use super::create_mdb_client;
                use crate::schema::derive_schema_for_partitions;
                use mongodb::bson::Document;

                let client = create_mdb_client().await;
                let db = client.database($db_name);
                let coll = db.collection::<Document>($coll_name);

                let (tx_notifications, _rx_notifications) =
                    tokio::sync::mpsc::unbounded_channel::<crate::SamplerNotification>();

                let partitions = $partitions.to_vec();

                match derive_schema_for_partitions(
                    $db_name.to_string(),
                    &coll,
                    partitions,
                    None,
                    &tokio::runtime::Handle::current(),
                    tx_notifications.clone(),
                )
                .await
                {
                    Err(err) => panic!("This test should never receive an error while deriving schema for partitions: {err}"),
                    Ok(None) => panic!("This test should never have received None while deriving schema for partitions"),
                    Ok(Some(schema)) => {
                        assert_eq!(schema, $expected_schema);
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
    use super::*;

    macro_rules! test_derive_schema_for_partition {
        ($test_name:ident, $db_name:expr, $coll_name:expr, $expected_schema:expr, $partitions:expr$(,)?) => {
            #[tokio::test]
            async fn $test_name() {
                use super::create_mdb_client;
                use crate::{partitioning::Partition, schema::derive_schema_for_partition};
                use mongodb::bson::Document;

                let client = create_mdb_client().await;
                let db = client.database($db_name);
                let coll = db.collection::<Document>($coll_name);

                let (tx_notifications, _rx_notifications) =
                    tokio::sync::mpsc::unbounded_channel::<crate::SamplerNotification>();

                let partitions: Vec<Partition> = $partitions.to_vec();

                for (ix, expected_schema) in $expected_schema.iter().enumerate() {
                    match derive_schema_for_partition(
                        $db_name.to_string(),
                        &coll,
                        partitions.get(ix).unwrap().clone(),
                        None,
                        tx_notifications.clone(),
                        ix,
                    )
                    .await
                    {
                        Err(err) => panic!("This test should never receive an error while deriving schema for a partition: {err}"),
                        Ok(schema) => {
                            assert_eq!(&schema, expected_schema);
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
    use super::*;

    macro_rules! test_derive_schema_for_view {
        ($test_name:ident, $db_name:expr, $view_name:expr, $expected_schema:expr$(,)?) => {
            #[tokio::test]
            async fn $test_name() {
                use super::create_mdb_client;
                use crate::{collection::CollectionInfo, schema::derive_schema_for_view};

                let db = create_mdb_client().await.database($db_name);

                let CollectionInfo { views, .. } =
                    match CollectionInfo::new(&db, $db_name, vec![], vec![]).await {
                        Ok(collection_info) => collection_info,
                        Err(e) => panic!("Error while creating CollectionInfo: {}", e),
                    };

                let view = views
                    .into_iter()
                    .find(|view| view.name == $view_name)
                    .expect("View not found");

                let (tx_notifications, _rx_notifications) =
                    tokio::sync::mpsc::unbounded_channel::<crate::SamplerNotification>();

                match derive_schema_for_view(&view, &db, tx_notifications.clone()).await {
                    None => {
                        panic!("This test should never receive None while deriving schema for view")
                    }
                    Some(schema) => {
                        assert_eq!(schema, $expected_schema);
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
