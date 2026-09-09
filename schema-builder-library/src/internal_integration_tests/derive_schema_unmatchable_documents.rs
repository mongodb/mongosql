//! Regression tests for documents that cannot match a `$jsonSchema` derived from the
//! schema they contributed to. Such a document is handed back by
//! `derive_schema_for_partition` on every iteration, and once it is the one at the
//! inclusive `$gte` lower bound nothing excludes it, so the loop spins (SQL-3436). The
//! bug is a hang, not a wrong answer, so each test races derivation against `TIMEOUT`:
//! the timeout is the real assertion and the schema equality check is the secondary
//! one, guarding a fix that terminates by discarding data.
//!
//! Two known causes: a field name containing a `.`, which `required` resolves as a path
//! (`"a.b"` as `a` -> `b`) rather than a literal key; and an empty field name on servers
//! affected by SERVER-92443 (see https://github.com/10gen/schema-manager-rs/pull/754).
//! In both cases the key must appear in *every* document: `required` is the intersection
//! of the unioned documents' required sets, so a key only some documents carry drops out,
//! and the document then matches on `properties`, which compares keys literally. These
//! tests use dots because that is the version-independent case.

use crate::{derive_schema_for_collection, internal_integration_tests::create_mdb_client};
use bson::{Document, doc};
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;
use std::time::Duration;

const DB_NAME: &str = "unmatchable_document_regression";

/// A liveness bound, not a performance one: generous enough that a slow machine cannot
/// trip it, while an actual infinite loop overruns it by any margin.
const TIMEOUT: Duration = Duration::from_secs(30);

/// Asserts that deriving a schema for a collection holding `docs` terminates, and
/// that the derived schema is the union of the schemas of every document -- i.e.
/// that nothing was dropped -- an unmatchable document is recorded in `ignored_min_id`
/// only *after* its contribution has been folded in, so a fix that bailed out of the
/// loop, or skipped such documents up front, would terminate but lose information.
#[allow(clippy::unwrap_used)]
async fn assert_derivation_terminates(coll_name: &str, docs: Vec<Document>) {
    let expected = docs.iter().fold(Schema::Unsat, |acc, doc| {
        acc.union(&schema_for_document(doc))
    });

    let client = create_mdb_client().await;
    let db = client.database(DB_NAME);
    let coll = db.collection::<Document>(coll_name);
    coll.drop().await.unwrap();
    coll.insert_many(docs).await.unwrap();

    let service = crate::data_service::MongoDbDataService::new(client);
    let derived = tokio::time::timeout(
        TIMEOUT,
        derive_schema_for_collection(&service, DB_NAME, coll_name, None),
    )
    .await;

    coll.drop().await.unwrap();

    match derived {
        Err(_) => panic!(
            "derive_schema_for_collection did not terminate within {TIMEOUT:?} for `{DB_NAME}.{coll_name}`"
        ),
        Ok(Err(err)) => panic!("unexpected error: {err:?}"),
        Ok(Ok(actual)) => assert_eq!(
            expected, actual,
            "derived schema for `{DB_NAME}.{coll_name}` does not match the union of its documents"
        ),
    }
}

macro_rules! test_derivation_terminates {
    ($test_name:ident, docs = $docs:expr) => {
        #[cfg(feature = "integration")]
        #[tokio::test]
        async fn $test_name() {
            super::derive_schema_unmatchable_documents::assert_derivation_terminates(
                stringify!($test_name),
                $docs,
            )
            .await
        }
    };
}

// A single unmatchable document is the guaranteed instance of the bug: it is alone in
// its batch and sits at the partition minimum, so neither the schema filter nor the
// inclusive `$gte` bound can exclude it.
test_derivation_terminates!(
    single_dotted_document,
    docs = vec![doc! {"_id": 0, "a.b": 1}]
);

test_derivation_terminates!(
    single_nested_dotted_document,
    docs = vec![doc! {"_id": 0, "a": {"b.c": 1}}]
);

test_derivation_terminates!(
    single_dotted_document_in_array,
    docs = vec![doc! {"_id": 0, "a": [{"b.c": 1}]}]
);

// More generally the loop hangs once a batch holds a single unmatchable document: it is that
// batch's first, which the pre-fix comparison against `iter_schema` (reset to `Unsat` every
// batch) could never recognize, and it sits at the inclusive `$gte` bound.
//
// A document is unmatchable only while its dotted key survives in `required`, which is the
// *intersection* of the unioned required sets -- so every document must carry the same dotted
// key. Differing dotted keys all drop out and then match via `properties`, which compares
// keys literally. The counts below are specific to these documents sharing one shape:
// duplicates within a batch are recognized and skipped, so a batch of one arises exactly at a
// count congruent to 1 modulo PARTITION_DOCS_PER_ITERATION.
test_derivation_terminates!(
    dotted_documents_one_over_a_full_batch,
    docs = (0..21).map(|i| doc! {"_id": i, "a.b": i}).collect()
);

test_derivation_terminates!(
    dotted_documents_two_over_a_full_batch,
    docs = (0..41).map(|i| doc! {"_id": i, "a.b": i}).collect()
);

// The count is not what matters. These two share `"a.b"`, so it survives the `required`
// intersection and both stay unmatchable, but they differ elsewhere -- so neither is ever
// recognized as a duplicate, `min` lands on the second, and the next batch holds it alone.
// A count of two, nowhere near 1 modulo PARTITION_DOCS_PER_ITERATION.
test_derivation_terminates!(
    shared_dotted_key_with_differing_shapes,
    docs = vec![
        doc! {"_id": 0, "a.b": 0, "u0": 0},
        doc! {"_id": 1, "a.b": 1, "u1": 1},
    ]
);

// Controls: no trailing batch of one, and no unmatchable document at all. A failure here
// means the fix broke ordinary derivation rather than that the bug is unfixed.
test_derivation_terminates!(
    dotted_documents_filling_whole_batches,
    docs = (0..40).map(|i| doc! {"_id": i, "a.b": i}).collect()
);

test_derivation_terminates!(
    single_document_without_dotted_fields,
    docs = vec![doc! {"_id": 0, "a": 1}]
);
