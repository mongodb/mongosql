use mongodb::bson::{self, doc, oid::ObjectId, Bson};
use mongosql::json_schema;

use crate::{generate_partition_match, partitioning::Partition, schema::schema_for_document};

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
