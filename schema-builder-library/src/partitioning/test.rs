#[cfg(test)]
#[allow(clippy::module_inception)]
mod test {
    use bson::{Bson, Document, doc, oid::ObjectId};
    use mongosql::json_schema;

    use crate::partitioning::{Partition, format_size, get_num_partitions};
    use schema_derivation::schema_for_document;
    use std::sync::LazyLock;

    static DEFAULT_PARTITION_KEY: LazyLock<String> = LazyLock::new(|| "_id".to_string());
    use crate::partitioning::partition::PARTITION_SIZE_IN_BYTES;

    #[test]
    fn generate_partition_match_with_mismatched_bson_types_without_schema_inclusive() {
        let partition = Partition {
            min: Bson::String("my_user_id".to_string()),
            max: Bson::Int64(5000),
            is_max_bound_inclusive: true,
        };

        let ignored_ids = vec![Bson::Int64(100), Bson::Int64(200)];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": vec![Bson::Int64(100), Bson::Int64(200)]}]}},
                            {"$gte": ["$_id", Bson::String("my_user_id".to_string())]},
                            {"$lte": ["$_id", Bson::Int64(5000)]}
                        ]
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_mismatched_bson_types_with_schema_inclusive() {
        let partition = Partition {
            min: Bson::Int64(0),
            max: Bson::String("my_user_id".to_string()),
            is_max_bound_inclusive: true,
        };

        let schema = schema_for_document(&doc! {
            "name": "AzureDiamond",
            "password": "hunter2",
            "iq": 140,
        });

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(
            Some(Document::try_from(schema.clone()).unwrap()),
            &ignored_ids,
            DEFAULT_PARTITION_KEY.as_str(),
        );
        let bson_schema = json_schema::Schema::try_from(schema)
            .unwrap()
            .to_bson()
            .unwrap();
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": vec![ignored_ids[0].clone()]}]}},
                            {"$gte": ["$_id", Bson::Int64(0)]},
                            {"$lte": ["$_id", Bson::String("my_user_id".to_string())]}
                        ]
                    },
                    "$nor": [{
                        "$jsonSchema": bson_schema
                    }]
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_mixed_numeric_bounds_without_schema_exclusive_uses_match_language()
     {
        let partition = Partition {
            min: Bson::Int64(1),
            max: Bson::Double(5000.0),
            is_max_bound_inclusive: false,
        };

        let ignored_ids = vec![Bson::Int64(100)];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "_id": {
                        "$nin": [Bson::Int64(100)],
                        "$gte": Bson::Int64(1),
                        "$lt": Bson::Double(5000.0)
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_wildcard_min_without_schema_inclusive_uses_expr() {
        let partition = Partition {
            min: Bson::MinKey,
            max: Bson::String("my_user_id".to_string()),
            is_max_bound_inclusive: true,
        };

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": vec![ignored_ids[0].clone()]}]}},
                            {"$gte": ["$_id", Bson::MinKey]},
                            {"$lte": ["$_id", Bson::String("my_user_id".to_string())]}
                        ]
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_wildcard_max_without_schema_inclusive_uses_expr() {
        let partition = Partition {
            min: Bson::String("my_user_id".to_string()),
            max: Bson::MaxKey,
            is_max_bound_inclusive: true,
        };

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": vec![ignored_ids[0].clone()]}]}},
                            {"$gte": ["$_id", Bson::String("my_user_id".to_string())]},
                            {"$lte": ["$_id", Bson::MaxKey]}
                        ]
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_mismatched_bson_types_max_bound_exclusive() {
        let partition = Partition {
            min: Bson::String("my_user_id".to_string()),
            max: Bson::Int64(5000),
            is_max_bound_inclusive: false,
        };

        let ignored_ids = vec![Bson::Int64(100)];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": vec![Bson::Int64(100)]}]}},
                            {"$gte": ["$_id", Bson::String("my_user_id".to_string())]},
                            {"$lt": ["$_id", Bson::Int64(5000)]}
                        ]
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_without_schema_doc_max_bound_inclusive() {
        let partition = Partition {
            min: Bson::MinKey,
            max: Bson::MaxKey,
            is_max_bound_inclusive: true,
        };

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "_id": {
                        "$nin": &ignored_ids,
                        "$gte": Bson::MinKey,
                        "$lte": Bson::MaxKey
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_without_schema_doc_max_bound_exclusive() {
        let partition = Partition {
            min: Bson::MinKey,
            max: Bson::MaxKey,
            is_max_bound_inclusive: false,
        };

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage =
            partition.generate_match(None, &ignored_ids, DEFAULT_PARTITION_KEY.as_str());
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "_id": {
                        "$nin": &ignored_ids,
                        "$gte": Bson::MinKey,
                        "$lt": Bson::MaxKey
                    }
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_schema_doc_max_bound_inclusive() {
        let partition = Partition {
            min: Bson::MinKey,
            max: Bson::MaxKey,
            is_max_bound_inclusive: true,
        };

        let schema = schema_for_document(&doc! {
            "name": "AzureDiamond",
            "password": "hunter2",
            "iq": 140,
        });

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(
            Some(Document::try_from(schema.clone()).unwrap()),
            &ignored_ids,
            DEFAULT_PARTITION_KEY.as_str(),
        );
        let bson_schema = json_schema::Schema::try_from(schema)
            .unwrap()
            .to_bson()
            .unwrap();
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "_id": {
                        "$nin": &ignored_ids,
                        "$gte": Bson::MinKey,
                        "$lte": Bson::MaxKey
                    },
                    "$nor": [{
                        "$jsonSchema": bson_schema
                    }]
                }
            }
        );
    }

    #[test]
    fn generate_partition_match_with_schema_doc_max_bound_exclusive() {
        let partition = Partition {
            min: Bson::MinKey,
            max: Bson::MaxKey,
            is_max_bound_inclusive: false,
        };

        let schema = schema_for_document(&doc! {
            "name": "AzureDiamond",
            "password": "hunter2",
            "iq": 140,
        });

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(
            Some(Document::try_from(schema.clone()).unwrap()),
            &ignored_ids,
            DEFAULT_PARTITION_KEY.as_str(),
        );
        let bson_schema = json_schema::Schema::try_from(schema)
            .unwrap()
            .to_bson()
            .unwrap();
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "_id": {
                        "$nin": &ignored_ids,
                        "$gte": Bson::MinKey,
                        "$lt": Bson::MaxKey
                    },
                    "$nor": [{
                        "$jsonSchema": bson_schema
                    }]
                }
            }
        );
    }

    #[test]
    fn generate_schema_with_mismatched_bson_types_with_schema_exclusive() {
        let partition = Partition {
            min: Bson::Int64(0),
            max: Bson::String("my_user_id".to_string()),
            is_max_bound_inclusive: false,
        };

        let schema = schema_for_document(&doc! {
            "name": "AzureDiamond",
            "password": "hunter2",
            "iq": 140,
        });

        let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
        let match_stage = partition.generate_match(
            Some(Document::try_from(schema.clone()).unwrap()),
            &ignored_ids,
            DEFAULT_PARTITION_KEY.as_str(),
        );
        let bson_schema = json_schema::Schema::try_from(schema)
            .unwrap()
            .to_bson()
            .unwrap();
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": vec![ignored_ids[0].clone()]}]}},
                            {"$gte": ["$_id", Bson::Int64(0)]},
                            {"$lt": ["$_id", Bson::String("my_user_id".to_string())]}
                        ]
                    },
                    "$nor": [{
                        "$jsonSchema": bson_schema
                    }]
                }
            }
        );
    }

    #[test]
    fn test_less_than_100_mb() {
        let collection_size = PARTITION_SIZE_IN_BYTES - 1;
        let num_partitions = get_num_partitions(collection_size, PARTITION_SIZE_IN_BYTES);
        assert_eq!(num_partitions, 1);
    }

    #[test]
    fn test_exactly_100_mb() {
        let collection_size = PARTITION_SIZE_IN_BYTES;
        let num_partitions = get_num_partitions(collection_size, PARTITION_SIZE_IN_BYTES);
        assert_eq!(num_partitions, 2);
    }

    #[test]
    fn test_greater_than_100_mb() {
        let collection_size_101_mb = 101 * 1024 * 1024;
        let collection_size_one_gb = 1000 * 1024 * 1024;
        assert_eq!(
            get_num_partitions(collection_size_101_mb, PARTITION_SIZE_IN_BYTES),
            2
        );
        assert_eq!(
            get_num_partitions(collection_size_one_gb, PARTITION_SIZE_IN_BYTES),
            11
        );
    }

    #[test]
    fn format_size_renders_each_unit() {
        assert_eq!(format_size(0), "0 B");
        assert_eq!(format_size(500), "500 B");
        assert_eq!(format_size(1024), "1.0 KiB");
        assert_eq!(format_size(1536), "1.5 KiB");
        assert_eq!(format_size(1024 * 1024), "1.0 MiB");
        assert_eq!(format_size(1024 * 1024 * 1024), "1.0 GiB");
        assert_eq!(format_size(1024 * 1024 * 1024 * 3 / 2), "1.5 GiB");
    }

    /// Tests for the fallbacks taken when partition sampling fails, driven by a scripted service.
    #[cfg(feature = "native-client")]
    mod fallback {
        use bson::Bson;

        use crate::{
            CollectionInfo,
            partitioning::{CollectionSizes, Partition, get_partitions, get_size_counts},
            test_util::{COLL, DB, MIB, Response, ScriptedDataService, stats, stats_without_count},
        };

        /// Above one partition, below the sample fallback threshold.
        const MID_SIZE: i64 = 500 * MIB;

        fn collection_info() -> CollectionInfo {
            CollectionInfo {
                name: COLL.to_string(),
                ..Default::default()
            }
        }

        fn whole_collection_partition() -> Vec<Partition> {
            vec![Partition {
                min: Bson::Int32(0),
                max: Bson::Int32(100),
                is_max_bound_inclusive: true,
            }]
        }

        #[tokio::test]
        async fn sampling_fails_on_open_below_threshold_falls_back_to_single_partition() {
            let mut service = ScriptedDataService::new(vec![stats(MID_SIZE, 1_000_000)]);
            service.sample_rate = Response::FailOnOpen;

            let result = get_partitions(&service, DB, collection_info()).await;

            match result {
                Ok(partitioned) => {
                    assert_eq!(partitioned.partitions, whole_collection_partition());
                }
                Err(e) => panic!("expected single-partition fallback, got error: {e}"),
            }
            assert!(
                !service.issued_sample(),
                "$sample must not run below the threshold"
            );
        }

        #[tokio::test]
        async fn sampling_fails_mid_stream_below_threshold_falls_back_to_single_partition() {
            let mut service = ScriptedDataService::new(vec![stats(MID_SIZE, 1_000_000)]);
            service.sample_rate = Response::FailMidStream(vec![bson::doc! { "_id": 50 }]);

            let result = get_partitions(&service, DB, collection_info()).await;

            match result {
                Ok(partitioned) => {
                    assert_eq!(partitioned.partitions, whole_collection_partition());
                }
                Err(e) => panic!("expected single-partition fallback, got error: {e}"),
            }
        }

        #[tokio::test]
        async fn size_counts_are_summed_across_shards() {
            let service =
                ScriptedDataService::new(vec![stats(600 * MIB, 10), stats(700 * MIB, 20)]);

            let result = get_size_counts(&service, DB, COLL).await;

            match result {
                Ok(sizes) => assert_eq!(
                    sizes,
                    CollectionSizes {
                        size: 1300 * MIB,
                        count: Some(30),
                    }
                ),
                Err(e) => panic!("expected summed sizes, got error: {e}"),
            }
        }

        #[tokio::test]
        async fn size_count_skips_shards_without_count() {
            let service = ScriptedDataService::new(vec![
                stats(600 * MIB, 10),
                stats_without_count(700 * MIB),
            ]);

            let result = get_size_counts(&service, DB, COLL).await;

            match result {
                Ok(sizes) => assert_eq!(
                    sizes,
                    CollectionSizes {
                        size: 1300 * MIB,
                        count: Some(10),
                    }
                ),
                Err(e) => panic!("expected summed sizes, got error: {e}"),
            }
        }

        #[tokio::test]
        async fn size_count_is_none_when_no_shard_reports_count() {
            let service = ScriptedDataService::new(vec![
                stats_without_count(600 * MIB),
                stats_without_count(700 * MIB),
            ]);

            let result = get_size_counts(&service, DB, COLL).await;

            match result {
                Ok(sizes) => assert_eq!(
                    sizes,
                    CollectionSizes {
                        size: 1300 * MIB,
                        count: None,
                    }
                ),
                Err(e) => panic!("expected summed sizes, got error: {e}"),
            }
        }
    }
}
