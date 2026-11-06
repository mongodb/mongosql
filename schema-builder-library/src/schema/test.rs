//! Tests for the `$sample` fallback taken by [`derive_schema_for_collection`] on large collections.

use bson::doc;
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;

use crate::{
    derive_schema_for_collection,
    partitioning::{SAMPLE_FALLBACK_SIZE, SAMPLE_FALLBACK_THRESHOLD_IN_BYTES},
    test_util::{COLL, DB, MIB, Response, ScriptedDataService, stats},
};

#[tokio::test]
async fn above_threshold_derives_schema_from_sample() {
    let sampled = vec![doc! { "_id": 1, "a": 1 }, doc! { "_id": 2, "b": "x" }];
    let mut service = ScriptedDataService::new(vec![stats(
        2 * SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
        1_000_000,
    )]);
    service.sample = Response::Docs(sampled.clone());

    let result = derive_schema_for_collection(&service, DB, COLL, None).await;

    assert!(
        service.issued_sample(),
        "$sample must run above the threshold"
    );
    assert!(
        !service.issued_sample_rate(),
        "primary partitioning must be skipped above the threshold"
    );
    assert!(
        service
            .pipelines()
            .contains(&vec![doc! { "$sample": { "size": SAMPLE_FALLBACK_SIZE } }]),
        "expected a $sample of SAMPLE_FALLBACK_SIZE, got {:?}",
        service.pipelines()
    );
    let expected = sampled
        .iter()
        .fold(Schema::Unsat, |acc, d| acc.union(&schema_for_document(d)));
    match result {
        Ok(schema) => assert_eq!(schema, expected),
        Err(e) => panic!("expected sampled schema, got error: {e}"),
    }
}

#[tokio::test]
async fn above_threshold_sample_fails_on_open_skips_collection() {
    let mut service = ScriptedDataService::new(vec![stats(
        2 * SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
        1_000_000,
    )]);
    service.sample = Response::FailOnOpen;

    let result = derive_schema_for_collection(&service, DB, COLL, None).await;

    assert!(
        result.is_err(),
        "expected the collection to be skipped, got {result:?}"
    );
    assert!(
        !service.issued_sample_rate(),
        "primary partitioning must be skipped above the threshold"
    );
}

#[tokio::test]
async fn above_threshold_sample_fails_mid_stream_skips_collection() {
    let mut service = ScriptedDataService::new(vec![stats(
        2 * SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
        1_000_000,
    )]);
    service.sample = Response::FailMidStream(vec![doc! { "_id": 1, "a": 1 }]);

    let result = derive_schema_for_collection(&service, DB, COLL, None).await;

    assert!(
        result.is_err(),
        "expected the collection to be skipped, got {result:?}"
    );
}

#[tokio::test]
async fn threshold_uses_total_size_across_shards() {
    // Each shard alone is below the threshold; together they exceed it.
    let mut service =
        ScriptedDataService::new(vec![stats(600 * MIB, 500_000), stats(600 * MIB, 500_000)]);
    service.sample = Response::Docs(vec![doc! { "_id": 1, "a": 1 }]);

    let _ = derive_schema_for_collection(&service, DB, COLL, None).await;

    assert!(
        service.issued_sample(),
        "$sample must run when the summed shard size exceeds the threshold"
    );
    assert!(
        !service.issued_sample_rate(),
        "primary partitioning must be skipped above the threshold"
    );
}

#[tokio::test]
async fn at_threshold_uses_partitioning() {
    let service =
        ScriptedDataService::new(vec![stats(SAMPLE_FALLBACK_THRESHOLD_IN_BYTES, 1_000_000)]);

    let _ = derive_schema_for_collection(&service, DB, COLL, None).await;

    assert!(
        service.issued_sample_rate(),
        "primary partitioning must run at the threshold"
    );
    assert!(
        !service.issued_sample(),
        "$sample must not run at the threshold"
    );
}
