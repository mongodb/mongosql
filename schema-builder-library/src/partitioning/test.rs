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
    fn test_generate_partition_match_with_mismatched_bson_types() {
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
    fn test_generate_partition_match_with_mismatched_bson_types_and_schema() {
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
    fn test_generate_partition_match_with_mixed_numeric_bounds_uses_match_language() {
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
    fn test_generate_partition_match_with_wildcard_min_uses_match_language() {
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
                    "_id": {
                        "$nin": [ignored_ids[0].clone()],
                        "$gte": Bson::MinKey,
                        "$lte": Bson::String("my_user_id".to_string())
                    }
                }
            }
        );
    }

    #[test]
    fn test_generate_partition_match_with_mismatched_bson_types_max_bound_exclusive() {
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
    fn test_generate_partition_match_without_schema_doc_max_bound_inclusive() {
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
                        "$nin": [ignored_ids[0].clone()],
                        "$gte": Bson::MinKey,
                        "$lte": Bson::MaxKey
                    }
                }
            }
        );
    }

    #[test]
    fn test_generate_partition_match_without_schema_doc_max_bound_exclusive() {
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
                        "$nin": [ignored_ids[0].clone()],
                        "$gte": Bson::MinKey,
                        "$lt": Bson::MaxKey
                    }
                }
            }
        );
    }

    #[test]
    fn test_generate_partition_match_with_schema_doc_max_bound_inclusive() {
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
                        "$nin": [ignored_ids[0].clone()],
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
    fn test_generate_partition_match_with_schema_doc_max_bound_exclusive() {
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
                        "$nin": [ignored_ids[0].clone()],
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
    fn test_generate_partition_match_with_mismatched_id_fields_dataset_bounds() {
        // Bounds taken from the `mismatched_id_fields_small` dataset, whose `_id` values mix
        // int64, string, and ObjectId. int64 and string are not comparable in match language,
        // so this exercises the `$expr` fallback.
        let partition = Partition {
            min: Bson::Int64(10),
            max: Bson::String("S".to_string()),
            is_max_bound_inclusive: true,
        };

        let ignored_ids = vec![Bson::Int64(13), Bson::String("id_string_0001".to_string())];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": ignored_ids.clone()}]}},
                            {"$gte": ["$_id", Bson::Int64(10)]},
                            {"$lte": ["$_id", Bson::String("S".to_string())]}
                        ]
                    }
                }
            }
        );
    }

    #[test]
    fn test_generate_partition_match_with_long_and_object_id_bounds() {
        // `$expr` comparisons use BSON total sort order, so this partition covers every key
        // between the bounds regardless of type -- including the string `_id`s that sort
        // between longs and ObjectIds.
        let partition = Partition {
            min: Bson::Int64(10),
            max: Bson::ObjectId(ObjectId::new()),
            is_max_bound_inclusive: true,
        };

        let ignored_ids = vec![Bson::Int64(13)];
        let match_stage = partition.generate_match(None, &ignored_ids, &DEFAULT_PARTITION_KEY);
        assert_eq!(
            match_stage,
            doc! {
                "$match": {
                    "$expr": {
                        "$and": [
                            {"$not": {"$in": ["$_id", {"$literal": ignored_ids.clone()}]}},
                            {"$gte": ["$_id", Bson::Int64(10)]},
                            {"$lte": ["$_id", partition.max.clone()]}
                        ]
                    }
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
}
