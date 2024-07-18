use mongodb::bson::{self, doc, oid::ObjectId, Bson};
use mongosql::json_schema;

use crate::{
    generate_partition_match,
    partitioning::{get_num_partitions, Partition},
    schema::schema_for_document,
    PARTITION_SIZE_IN_BYTES,
};

#[test]
fn test_generate_partition_match_without_schema_doc() {
    let partition = Partition {
        min: Bson::MinKey,
        max: Bson::MaxKey,
    };

    let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
    let match_stage = generate_partition_match(&partition, None, &ignored_ids).unwrap();
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
fn test_generate_partition_match_with_schema_doc() {
    let partition = Partition {
        min: Bson::MinKey,
        max: Bson::MaxKey,
    };
    let schema = schema_for_document(&doc! {
        "name": "AzureDiamond",
        "password": "hunter2",
        "iq": 140,
    });

    let ignored_ids = vec![Bson::ObjectId(ObjectId::new())];
    let match_stage =
        generate_partition_match(&partition, Some(schema.clone()), &ignored_ids).unwrap();
    let bson_schema = bson::to_bson(&json_schema::Schema::try_from(schema).unwrap()).unwrap();
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
