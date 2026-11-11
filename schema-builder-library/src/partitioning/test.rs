#[cfg(test)]
#[allow(clippy::module_inception)]
mod test {
    use bson::{Bson, Document, doc, oid::ObjectId};
    use mongosql::json_schema;

    use crate::partitioning::{Partition, ScanStrategy, format_size, get_num_partitions};
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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
            strategy: ScanStrategy::Range,
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

    mod get_partitions {
        use std::collections::HashMap;

        use bson::{Bson, Document, doc};

        use crate::{
            Error,
            data_service::{CollectionInfo, test::MockDataService},
            partitioning::{
                Partition, SAMPLE_FALLBACK_THRESHOLD_IN_BYTES, ScanStrategy, get_partitions,
            },
        };

        /// One shard's `$collStats` document, as the server returns it for a sharded collection.
        ///
        /// The mock answers every query on the collection with the same documents, so the bound
        /// queries read their `_id` from the first of these too; real `$collStats` output has no
        /// `_id`.
        fn coll_stats(id: i32, size: i64, count: i64) -> Document {
            doc! {
                "_id": id,
                "ns": "db.coll",
                "shard": format!("shard{id}"),
                "storageStats": { "size": size, "count": count, "storageSize": size },
            }
        }

        /// Three shards whose sizes sum to `total_size`, each reporting `count_per_shard`.
        fn three_shards(total_size: i64, count_per_shard: i64) -> Vec<Document> {
            let quarter = total_size / 4;
            vec![
                coll_stats(1, total_size - 2 * quarter, count_per_shard),
                coll_stats(2, quarter, count_per_shard),
                coll_stats(3, quarter, count_per_shard),
            ]
        }

        fn mock(documents: Vec<Document>) -> MockDataService {
            MockDataService {
                databases: Vec::new(),
                collections: HashMap::new(),
                documents: HashMap::from([("db.coll".to_string(), documents)]),
            }
        }

        fn collection_info() -> CollectionInfo {
            CollectionInfo {
                name: "coll".to_string(),
                ..Default::default()
            }
        }

        #[tokio::test]
        async fn large_collection_returns_single_sample_partition() {
            // No shard reaches the threshold on its own; only their sum does.
            let service = mock(three_shards(SAMPLE_FALLBACK_THRESHOLD_IN_BYTES, 1_000_000));

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("a large collection is partitioned");

            assert_eq!(
                partitioned.partitions,
                vec![Partition {
                    min: Bson::Int32(1),
                    max: Bson::Int32(1),
                    is_max_bound_inclusive: true,
                    strategy: ScanStrategy::Sample,
                }]
            );
            assert_eq!(partitioned.partition_key, "_id");
            assert_eq!(partitioned.hint, Some(doc! { "_id": 1 }));
        }

        #[tokio::test]
        async fn collection_below_threshold_is_not_sampled() {
            let service = mock(three_shards(
                SAMPLE_FALLBACK_THRESHOLD_IN_BYTES - 1,
                1_000_000,
            ));

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("a collection below the threshold is partitioned");

            assert!(
                partitioned
                    .partitions
                    .iter()
                    .all(|p| p.strategy == ScanStrategy::Range)
            );
        }

        #[tokio::test]
        async fn missing_coll_stats_is_an_error() {
            let service = mock(Vec::new());

            let result = get_partitions(&service, "db", collection_info()).await;

            assert!(matches!(result, Err(Error::NoCollectionStats(c)) if c == "coll"));
        }

        #[tokio::test]
        async fn empty_collection_is_an_error() {
            let service = mock(three_shards(SAMPLE_FALLBACK_THRESHOLD_IN_BYTES, 0));

            let result = get_partitions(&service, "db", collection_info()).await;

            assert!(matches!(result, Err(Error::EmptyCollection(c)) if c == "coll"));
        }
    }
}
