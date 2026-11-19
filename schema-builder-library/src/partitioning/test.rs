#[cfg(test)]
#[allow(clippy::module_inception)]
mod test {
    use bson::{Bson, Document, doc, oid::ObjectId};
    use mongosql::json_schema;

    use crate::partitioning::{Partition, get_num_partitions};
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

    mod get_partitions {
        use std::collections::HashMap;

        use bson::{Bson, Document, doc};

        use crate::{
            Error,
            data_service::{CollectionInfo, test::MockDataService},
            partitioning::{Partition, get_partitions},
        };

        /// One shard's `$collStats` document, as the server returns it for a sharded collection.
        fn coll_stats(id: i32, docs_per_shard: i64) -> Document {
            doc! {
                "_id": id,
                "ns": "db.coll",
                "shard": format!("shard{id}"),
                "storageStats": { "size": 100, "count": docs_per_shard, "storageSize": 100 },
            }
        }

        /// Three shards whose sizes sum to `total_size`, each reporting `count_per_shard`.
        fn coll_stats_documents(documents_per_shard: i64) -> Vec<Document> {
            vec![coll_stats(1, 10), coll_stats(2, 10), coll_stats(3, 10)]
        }

        fn collection_info() -> CollectionInfo {
            CollectionInfo {
                name: "coll".to_string(),
                ..Default::default()
            }
        }

        #[tokio::test]
        async fn sharded_collection_sampled() {
            let service = MockDataService {
                databases: Vec::new(),
                collections: HashMap::new(),
                documents: HashMap::from([("db.coll".to_string(), coll_stats_documents(100))]),
            };

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("a sharded collection is partitioned");

            // A sharded collection should return 1 partition, indicating that we should use sampling.
            assert_eq!(partitioned.partitions.len(), 1);
            assert!(partitioned.is_sharded_collection);
        }

        #[tokio::test]
        async fn unsharded_collection_is_not_sampled() {
            // An unsharded collection returns a single `$collStats` document.
            let service = MockDataService {
                databases: Vec::new(),
                collections: HashMap::new(),
                documents: HashMap::from([("db.coll".to_string(), vec![coll_stats(1, 100)])]),
            };

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("an unsharded collection is partitioned");

            assert!(!partitioned.is_sharded_collection);
            assert_eq!(partitioned.partitions.len(), 1);
        }

        #[tokio::test]
        async fn missing_coll_stats_is_an_error() {
            let service = MockDataService {
                databases: Vec::new(),
                collections: HashMap::new(),
                documents: HashMap::from([("db.coll".to_string(), Vec::new())]),
            };

            let result = get_partitions(&service, "db", collection_info()).await;

            assert!(matches!(result, Err(Error::NoCollectionStats(c)) if c == "coll"));
        }

        #[tokio::test]
        async fn empty_collection_is_an_error() {
            let service = MockDataService {
                databases: Vec::new(),
                collections: HashMap::new(),
                documents: HashMap::from([("db.coll".to_string(), coll_stats_documents(100))]),
            };

            let result = get_partitions(&service, "db", collection_info()).await;

            assert!(matches!(result, Err(Error::EmptyCollection(c)) if c == "coll"));
        }
    }
}
