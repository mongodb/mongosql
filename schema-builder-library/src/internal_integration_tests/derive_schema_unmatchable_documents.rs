//! Regression tests for documents that cannot match a `$jsonSchema` derived from
//! themselves.
//!
//! `derive_schema_for_partition` makes progress by re-querying a partition with
//! `$nor: [$jsonSchema]`, so that documents already described by the accumulated
//! schema stop coming back. That only works if a document matches the schema it
//! contributed to. A document that cannot match its own `$jsonSchema` is returned
//! forever, no matter how much schema has been accumulated, and once it is the
//! document sitting at the inclusive `$gte` lower bound of the partition, nothing
//! excludes it and the derivation loop spins. `ignored_min_id` exists solely to
//! break that cycle.
//!
//! There are two known ways for a document to be unmatchable by its own schema:
//!
//!   1. A field name containing a `.`. The `required` keyword resolves `"a.b"` as
//!      the path `a` -> `b`, which is absent, so the constraint can never be
//!      satisfied. This is how `$jsonSchema` is specified to behave, so it holds on
//!      every server version.
//!   2. An empty field name, on servers affected by SERVER-92443. This is the case
//!      the original `ignored_ids` workaround was written for. See
//!      https://github.com/10gen/schema-manager-rs/pull/754 for context.
//!
//! What these tests actually assert is that schema derivation **terminates**. The
//! failure mode of the bug (SQL-3436) is not a wrong answer but an infinite loop:
//! every input to the query is unchanged from one iteration to the next -- `schema`
//! stops widening, `partition.min` stops advancing, and nothing is added to the
//! exclusion list -- so the loop reissues an identical aggregate forever and the call
//! never returns. In production that presents as a schema-building job that hangs
//! rather than one that fails.
//!
//! A hang cannot be asserted directly, so each test races the derivation against
//! `TIMEOUT` and treats expiry as the failure. That makes these tests unusual for
//! this suite: a *timeout* is the real assertion, and the schema equality check below
//! is the secondary one that guards against a fix that terminates by discarding data.
//!
//! These tests use dotted field names because they are the version-independent case:
//! case 2 reproduces only against affected server versions, so it cannot be asserted
//! from a test suite that runs against whatever `mongod` is at hand. The bug and the
//! fix are not specific to dots -- any future third cause would land in the same code
//! path -- so read the dotted documents below as a stand-in for the whole class.

use crate::{derive_schema_for_collection, internal_integration_tests::create_mdb_client};
use bson::{Document, doc};
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;
use std::time::Duration;

const DB_NAME: &str = "unmatchable_document_regression";

/// Deriving a schema should never take anywhere near this long for a handful of tiny
/// documents; exceeding it means the derivation loop in `derive_schema_for_partition`
/// is not terminating. Generous on purpose -- it is a liveness bound, not a
/// performance one, so a slow machine or a cold `mongod` must not trip it, while an
/// actual infinite loop overruns it by any margin.
const TIMEOUT: Duration = Duration::from_secs(30);

/// Asserts that deriving a schema for a collection holding `docs` terminates, and
/// that the derived schema is the union of the schemas of every document -- i.e.
/// that nothing was dropped. The latter matters because an unmatchable document is
/// only ever recorded in `ignored_min_id` *after* its contribution has been folded
/// into the accumulated schema; a fix that simply bailed out of the loop, or that
/// skipped such documents before reading them, would terminate but lose information.
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

// A collection holding a single unmatchable document is the guaranteed instance of the
// bug: it is the only member of its batch and it sits at the partition minimum, so
// neither the `$nor` schema filter nor the inclusive `$gte` bound can exclude it. The
// pre-fix code compared each document against `iter_schema`, which restarts at `Unsat`
// every batch, so the first document of a batch could never be recognized as already
// covered -- and a batch of one has nothing but a first document.
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

// More generally, the loop fails to terminate whenever the trailing batch of a
// partition holds exactly one unmatchable document, i.e. when the number of such
// documents is congruent to 1 modulo PARTITION_DOCS_PER_ITERATION. Uniform shapes are
// used here so that the arithmetic is exact; with differing shapes the loop reaches the
// same stuck state, just an iteration or two later.
test_derivation_terminates!(
    dotted_documents_one_over_a_full_batch,
    docs = (0..21).map(|i| doc! {"_id": i, "a.b": i}).collect()
);

test_derivation_terminates!(
    dotted_documents_two_over_a_full_batch,
    docs = (0..41).map(|i| doc! {"_id": i, "a.b": i}).collect()
);

// Controls. The first has no trailing batch of one, and the second has no unmatchable
// document at all; neither should ever have been affected, so a failure here means the
// fix broke ordinary derivation rather than that the bug is unfixed.
test_derivation_terminates!(
    dotted_documents_filling_whole_batches,
    docs = (0..40).map(|i| doc! {"_id": i, "a.b": i}).collect()
);

test_derivation_terminates!(
    single_document_without_dotted_fields,
    docs = vec![doc! {"_id": 0, "a": 1}]
);
