use std::sync::Arc;

use bson::{Bson, Document, doc};
use futures::TryStreamExt as _;
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;
use tracing::{info, instrument, warn};

use crate::data_service::{AggregateOptions, LocalDataService};
use crate::partitioning::{SAMPLE_FALLBACK_SIZE, ScanStrategy};
use crate::{Error, data_service::CollectionInfo, partitioning::Partition};
use crate::{PartitionedCollection, get_partitions};

/// The amount of samples to fetch for view schema derivation
const VIEW_SAMPLE_SIZE: i64 = 1000;

/// How many sampled documents to process between progress log lines.
///
/// Frequent enough to show progress through a sample, sparse enough that a
/// full sample only emits a handful of `info!` events.
const SAMPLE_PROGRESS_LOG_INTERVAL: u64 = 100;

#[derive(Debug, PartialEq, Clone)]
pub struct SinglePartition {
    pub partition: Partition,
    pub partition_key: String,
    pub hint: Option<Document>,
    pub partition_ix: usize,
}

pub const PARTITION_DOCS_PER_ITERATION: i64 = 20;

/// Derive a schema for an entire collection
///
/// Note: This breaks up the work into manageable partitions and operates over
/// them in order. If you want to control the execution of these partitions,
/// please refer to [derive_schema_for_partition].
pub async fn derive_schema_for_collection<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
    initial_schema_doc: Option<Arc<Schema>>,
) -> Result<Schema, Error<S::Error>> {
    let collections = service
        .list_collections(db)
        .await
        .map_err(Error::DataServiceError)?;
    let collection_info = collections
        .into_iter()
        .find(|c| c.name == collection)
        .ok_or_else(|| Error::NoCollection(collection.to_string()))?;

    let PartitionedCollection {
        partitions,
        partition_key,
        hint,
    } = get_partitions(service, db, collection_info).await?;

    // Note that we operate over all of the partitions linearly here. We could
    // return a future that polls all partitions at the same time, but we figured
    // that we can leave those sort of optimizations to the caller by exposing
    // the `derive_schema_for_partition` directly.
    let mut result = Schema::Unsat;
    for (index, partition) in partitions.into_iter().enumerate() {
        let schema = derive_schema_for_partition(
            service,
            db,
            collection,
            initial_schema_doc.clone(),
            SinglePartition {
                partition,
                partition_key: partition_key.clone(),
                hint: hint.clone(),
                partition_ix: index,
            },
        )
        .await?;

        result = result.union(&schema);
    }

    Ok(result)
}

/// Derives the schema for a single partition of a collection.
///
/// A partition is normally scanned in batches between its bounds, skipping documents that
/// already match the schema built so far. When `initial_schema_doc` is supplied, its simplified
/// schema seeds the result.
///
/// Partitions that [`get_partitions`] marks for sampling (collections of 1 GiB or more) instead
/// derive their schema from a single `$sample` of 1000 documents, unioned with
/// `initial_schema_doc`. If that sample fails to open or returns no documents, the partition is
/// scanned over its full bounds instead.
///
/// # Errors
///
/// Returns an error if a range-scan query fails, or if the schema cannot be converted to a
/// `$jsonSchema` document for the scan's filter.
#[instrument(level = "trace", skip_all)]
pub async fn derive_schema_for_partition<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
    initial_schema_doc: Option<Arc<Schema>>,
    single_partition: SinglePartition,
) -> Result<Schema, Error<S::Error>> {
    // ignored_min_id is the id of a document that `schema`'s `$jsonSchema` rendering cannot
    // match. It is remembered so the next query can exclude it by id. At most one such document
    // can survive: each iteration queries `partition_key >= partition.min` sorted ascending, and
    // the batch loop advances `partition.min` to the last document it processes, so the inclusive
    // `>=` bound re-admits only the document sitting exactly at `min`.
    //
    // Reset to `None` whenever a batch widens the schema, since `min` has then advanced past any
    // earlier blocker. Scoped to this partition, whose bounds are independent of all others.
    //
    // See the `ignored_min_id` assignment below for how a blocker is recognized.
    let mut ignored_min_id: Option<Bson> = None;
    let mut partition = single_partition.partition;

    if partition.strategy == ScanStrategy::Sample {
        if let Some(sampled) = sample_schema_for_collection(service, db, collection).await {
            return Ok(match &initial_schema_doc {
                Some(initial) => Schema::simplify(initial).union(&sampled),
                None => sampled,
            });
        }
        // The partition still spans the collection's full bounds, so the range scan below is the
        // single-partition last resort.
        warn!(
            name: "schema.sampling.failed",
            db,
            collection,
            "sampling returned no schema; falling back to scanning the entire collection as a \
             single partition, which may take a long time for large collections"
        );
        partition.strategy = ScanStrategy::Range;
    }

    let partition_key = single_partition.partition_key.as_str();
    let hint = single_partition.hint;
    let partition_ix = single_partition.partition_ix;

    // The initial schema might be empty, in which case we
    // default to `Unsat` if there are also no refinement entries for the DB
    let mut schema = initial_schema_doc
        .map(|s| Schema::simplify(&s))
        .unwrap_or(Schema::Unsat);

    let mut saw_unstable = false;

    loop {
        info!(db, collection, "querying partition: {partition_ix}");

        // This is a somewhat expensive clone, but there isn't a try_from for
        // a schema reference :(
        let doc = (schema != Schema::Unsat)
            .then(|| bson::Document::try_from(schema.clone()))
            .transpose()?;

        let pipeline = vec![
            partition.generate_match(doc, ignored_min_id.as_slice(), partition_key),
            doc! { "$sort": {partition_key: 1}},
            doc! { "$limit": PARTITION_DOCS_PER_ITERATION },
        ];
        let cursor = service
            .aggregate(
                db,
                collection,
                pipeline,
                AggregateOptions {
                    key_hint: hint.clone(),
                },
            )
            .await
            .map_err(Error::DataServiceError)?;

        let mut no_result = true;
        let mut iter_schema = Schema::Unsat;
        let mut cursor = Box::pin(cursor);
        while let Some(doc) = cursor.try_next().await.map_err(Error::DataServiceError)? {
            info!(db, collection, "processing partition {partition_ix}");
            if let Some(id) = doc.get(partition_key) {
                partition.min = id.clone();
                iter_schema = iter_schema.union(&schema_for_document(&doc));
                no_result = false;
            } else {
                warn!(
                    db,
                    collection, "document is missing the {partition_key} field"
                );
                continue;
            };
        }
        if no_result {
            break;
        }

        let new_schema = schema.union(&iter_schema);

        // A batch that widened nothing means `schema` already covers what came back, yet
        // `$nor: [$jsonSchema]` returned it anyway -- so the document at `partition.min`
        // cannot match the `$jsonSchema` rendering of the very schema it contributed to.
        // `required` resolves `"a.b"` as the path `a` -> `b`, and empty keys behave likewise
        // on servers affected by SERVER-92443; see
        // https://github.com/10gen/schema-manager-rs/pull/754. Neither the schema filter nor
        // the inclusive `$gte` bound can drop it, so record its id -- its contribution is
        // already folded into `schema`, so excluding it from here on loses nothing.
        //
        // Otherwise, `min` has advanced past any previous blocker and `$gte` excludes it.
        ignored_min_id = (new_schema == schema).then(|| partition.min.clone());

        schema = new_schema;

        // If the schema for this partition becomes unstable, we should do at most one more
        // iteration to see if we detect any additional properties. After two iterations with an
        // unstable schema, we should stop deriving schema for this partition since it is unlikely
        // new information will be added.

        if schema.is_unstable() {
            if saw_unstable {
                break;
            } else {
                saw_unstable = true;
            }
        }
    }

    Ok(schema)
}

/// Derive a schema for a collection by sampling a bounded number of documents.
///
/// Issues a single `$sample` of [`SAMPLE_FALLBACK_SIZE`] documents and unions the inferred
/// schemas, mirroring [`derive_schema_for_view`]. This is used for collections too large to
/// partition cheaply.
///
/// Returns `None` if the query fails to open or returns no documents, so the caller can choose a
/// fallback. If the cursor errors mid-stream, the error is logged and the schema built from the
/// documents read so far is returned, which may be incomplete.
#[instrument(level = "trace", skip_all)]
pub(crate) async fn sample_schema_for_collection<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
) -> Option<Schema> {
    let pipeline = vec![doc! { "$sample": { "size": SAMPLE_FALLBACK_SIZE } }];

    let cursor = service
        .aggregate(db, collection, pipeline, AggregateOptions::default())
        .await
        .inspect_err(|e| {
            warn!(
                name: "schema.sampling.open_failed",
                db,
                collection,
                error.message = %e,
                "collection sampling failed to open"
            );
        })
        .ok()?;

    let mut schema = None;
    let mut iterations = 0u64;
    let mut cursor = Box::pin(cursor);
    while let Some(Ok(doc)) = cursor
        .try_next()
        .await
        .inspect_err(|e| {
            warn!(
                name: "schema.sampling.stream_failed",
                db,
                collection,
                iteration = iterations,
                error.message = %e,
                "collection sampling failed mid-stream; the sampled schema may be incomplete"
            );
        })
        .transpose()
    {
        if iterations.is_multiple_of(SAMPLE_PROGRESS_LOG_INTERVAL) {
            info!(
                name: "schema.sampling.progress",
                db,
                collection,
                iteration = iterations,
                "calculating schema for sampled collection"
            );
        }

        schema = schema.map_or(Some(schema_for_document(&doc)), |s: Schema| {
            Some(s.union(&schema_for_document(&doc)))
        });

        iterations += 1;
    }

    info!(
        name: "schema.sampling.complete",
        db,
        collection,
        sampled_documents = iterations,
        "derived schema from sampled documents"
    );

    schema
}

/// derive_schema_for_view takes a CollectionInfo and executes the pipeline
/// against the viewOn collection to generate a schema for the view.
/// It does this by first prepending $sample to the pipeline
#[instrument(level = "trace", skip_all)]
pub async fn derive_schema_for_view<S: LocalDataService>(
    service: &S,
    db: &str,
    view: &CollectionInfo,
) -> Option<Schema> {
    let pipeline = vec![doc! { "$sample": { "size": VIEW_SAMPLE_SIZE } }]
        .into_iter()
        .chain(view.options.pipeline.clone().into_iter())
        .collect::<Vec<Document>>();

    let cursor = service
        .aggregate(
            db,
            &view.options.view_on,
            pipeline,
            AggregateOptions::default(),
        )
        .await
        .inspect_err(|e| {
            warn!(
                db,
                view_name = view.name,
                "view sampling encountered an error: {e}"
            )
        })
        .ok()?;

    let mut schema = None;
    let mut iterations = 0u64;
    let mut cursor = Box::pin(cursor);
    while let Some(Ok(doc)) = cursor
        .try_next()
        .await
        .inspect_err(|e| {
            warn!(
                db,
                view_name = view.name,
                iteration = iterations,
                "view sampling encountered an error: {e}"
            )
        })
        .transpose()
    {
        if iterations.is_multiple_of(SAMPLE_PROGRESS_LOG_INTERVAL) {
            info!(
                db,
                view_name = view.name,
                iteration = iterations,
                "Calculating schema for sampled view"
            );
        }

        schema = schema.map_or(Some(schema_for_document(&doc)), |s: Schema| {
            Some(s.union(&schema_for_document(&doc)))
        });

        iterations += 1;
    }

    schema
}

#[cfg(test)]
mod tests {
    use std::collections::HashMap;

    use bson::{Document, doc};

    use super::*;
    use crate::data_service::test::MockDataService;

    /// Build a [`MockDataService`] whose `aggregate` returns `docs` for the `db.collection`
    /// lookup that `sample_schema_for_collection` performs.
    fn mock_service(db: &str, coll: &str, docs: Vec<Document>) -> MockDataService {
        MockDataService {
            databases: Vec::new(),
            collections: HashMap::new(),
            documents: HashMap::from([(format!("{db}.{coll}"), docs)]),
        }
    }

    #[tokio::test]
    async fn sample_unions_sampled_documents() {
        let service = mock_service(
            "db",
            "coll",
            vec![doc! { "a": 1 }, doc! { "a": 2, "b": "x" }],
        );

        let result = sample_schema_for_collection(&service, "db", "coll")
            .await
            .expect("a non-empty sample returns Some");

        let expected = schema_for_document(&doc! { "a": 1 })
            .union(&schema_for_document(&doc! { "a": 2, "b": "x" }));
        assert_eq!(result, expected);
    }

    #[tokio::test]
    async fn sample_returns_none_when_empty() {
        let service = mock_service("db", "coll", Vec::new());

        let result = sample_schema_for_collection(&service, "db", "coll").await;

        assert_eq!(result, None);
    }

    mod sampling_fallback {
        use std::sync::Arc;

        use bson::{Bson, Document, doc};
        use mongosql::schema::Schema;
        use schema_derivation::schema_for_document;

        use crate::{
            CollectionInfo, SinglePartition,
            data_service::test::{PipelineKind, Scripted, ScriptedDataService},
            derive_schema_for_partition, get_partitions,
            partitioning::{Partition, SAMPLE_FALLBACK_THRESHOLD_IN_BYTES, ScanStrategy},
        };

        /// Smallest `_id` the scripted bounds query reports.
        const MIN_ID: i32 = 1;
        /// Largest `_id` the scripted bounds query reports.
        const MAX_ID: i32 = 100;
        /// Below the sampling threshold but large enough for several partitions (200 MiB).
        const SMALL_SIZE: i64 = 200 * 1024 * 1024;

        fn coll_stats(size: i64, count: Option<i64>) -> Document {
            let mut storage_stats = doc! { "size": size };
            if let Some(count) = count {
                storage_stats.insert("count", count);
            }
            doc! { "storageStats": storage_stats }
        }

        fn sampled_docs() -> Vec<Document> {
            vec![doc! { "_id": 5, "a": 1 }, doc! { "_id": 7, "b": "x" }]
        }

        fn range_docs() -> Vec<Document> {
            vec![doc! { "_id": 2, "c": true }, doc! { "_id": 3, "d": 1.5 }]
        }

        fn schema_of(docs: &[Document]) -> Schema {
            docs.iter()
                .map(schema_for_document)
                .reduce(|acc, s| acc.union(&s))
                .unwrap_or(Schema::Unsat)
        }

        /// Script a collection of `size` bytes. `partitioning`, `sample` and `range` decide the
        /// outcome of the partitioning, `$sample`, and first range-scan queries.
        fn service(
            size: i64,
            count: Option<i64>,
            partitioning: Scripted,
            sample: Scripted,
        ) -> ScriptedDataService {
            ScriptedDataService::new(move |kind, prior_calls| match kind {
                PipelineKind::CollStats => Scripted::Docs(vec![coll_stats(size, count)]),
                PipelineKind::Bound { ascending: true } => {
                    Scripted::Docs(vec![doc! { "_id": MIN_ID }])
                }
                PipelineKind::Bound { ascending: false } => {
                    Scripted::Docs(vec![doc! { "_id": MAX_ID }])
                }
                PipelineKind::Partitioning => partitioning.clone(),
                PipelineKind::Sample => sample.clone(),
                // Return the documents once, then report the range as exhausted.
                PipelineKind::RangeScan if prior_calls == 0 => Scripted::Docs(range_docs()),
                PipelineKind::RangeScan => Scripted::Docs(Vec::new()),
            })
        }

        fn collection_info() -> CollectionInfo {
            CollectionInfo {
                name: "coll".to_string(),
                ..Default::default()
            }
        }

        fn full_range(strategy: ScanStrategy) -> Vec<Partition> {
            vec![Partition {
                min: Bson::Int32(MIN_ID),
                max: Bson::Int32(MAX_ID),
                is_max_bound_inclusive: true,
                strategy,
            }]
        }

        async fn derive(
            service: &ScriptedDataService,
            initial_schema_doc: Option<Arc<Schema>>,
            partition: Partition,
        ) -> Schema {
            derive_schema_for_partition(
                service,
                "db",
                "coll",
                initial_schema_doc,
                SinglePartition {
                    partition,
                    partition_key: "_id".to_string(),
                    hint: Some(doc! { "_id": 1 }),
                    partition_ix: 0,
                },
            )
            .await
            .expect("scripted derivation succeeds")
        }

        #[tokio::test]
        async fn large_collection_skips_partitioning_and_samples() {
            // The partitioning pipeline would fail; it must never be issued.
            let service = service(
                SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
                Some(1_000_000),
                Scripted::FailOpen,
                Scripted::Docs(sampled_docs()),
            );

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("large collection yields a sample partition");
            assert_eq!(partitioned.partitions, full_range(ScanStrategy::Sample));

            let schema = derive(&service, None, partitioned.partitions[0].clone()).await;

            assert_eq!(schema, schema_of(&sampled_docs()));
            let calls = service.calls();
            assert!(calls.contains(&PipelineKind::Sample));
            assert!(!calls.contains(&PipelineKind::Partitioning));
            assert!(!calls.contains(&PipelineKind::RangeScan));
        }

        #[tokio::test]
        async fn small_collection_partitioning_failure_uses_single_partition() {
            let service = service(
                SMALL_SIZE,
                Some(1_000),
                Scripted::FailOpen,
                Scripted::Docs(sampled_docs()),
            );

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("partitioning failure falls back to a single partition");
            assert_eq!(partitioned.partitions, full_range(ScanStrategy::Range));

            let schema = derive(&service, None, partitioned.partitions[0].clone()).await;

            assert_eq!(schema, schema_of(&range_docs()));
            assert!(!service.calls().contains(&PipelineKind::Sample));
        }

        #[tokio::test]
        async fn failed_sample_falls_back_to_full_range_scan() {
            let service = service(
                SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
                Some(1_000_000),
                Scripted::FailOpen,
                Scripted::FailOpen,
            );

            let partitioned = get_partitions(&service, "db", collection_info())
                .await
                .expect("large collection yields a sample partition");
            let schema = derive(&service, None, partitioned.partitions[0].clone()).await;

            assert_eq!(schema, schema_of(&range_docs()));
            let calls = service.calls();
            assert!(calls.contains(&PipelineKind::Sample));
            assert!(calls.contains(&PipelineKind::RangeScan));
        }

        #[tokio::test]
        async fn empty_sample_falls_back_to_full_range_scan() {
            let service = service(
                SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
                Some(1_000_000),
                Scripted::FailOpen,
                Scripted::Docs(Vec::new()),
            );

            let schema = derive(&service, None, full_range(ScanStrategy::Sample).remove(0)).await;

            assert_eq!(schema, schema_of(&range_docs()));
        }

        #[tokio::test]
        async fn sample_partition_unions_initial_schema_with_sampled_documents() {
            let service = service(
                SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
                Some(1_000_000),
                Scripted::FailOpen,
                Scripted::Docs(sampled_docs()),
            );
            let initial = Arc::new(schema_for_document(&doc! { "_id": 1, "seed": true }));

            let schema = derive(
                &service,
                Some(Arc::clone(&initial)),
                full_range(ScanStrategy::Sample).remove(0),
            )
            .await;

            assert_eq!(
                schema,
                Schema::simplify(&initial).union(&schema_of(&sampled_docs()))
            );
            assert!(!service.calls().contains(&PipelineKind::RangeScan));
        }

        #[tokio::test]
        async fn empty_sample_with_initial_schema_falls_back_to_full_range_scan() {
            // A stored schema must not mask an empty sample; the full scan still runs.
            let service = service(
                SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
                Some(1_000_000),
                Scripted::FailOpen,
                Scripted::Docs(Vec::new()),
            );
            let initial = Arc::new(schema_for_document(&doc! { "_id": 1, "seed": true }));

            let schema = derive(
                &service,
                Some(Arc::clone(&initial)),
                full_range(ScanStrategy::Sample).remove(0),
            )
            .await;

            assert_eq!(
                schema,
                Schema::simplify(&initial).union(&schema_of(&range_docs()))
            );
            assert_eq!(
                service
                    .calls()
                    .iter()
                    .filter(|k| **k != PipelineKind::RangeScan)
                    .last(),
                Some(&PipelineKind::Sample)
            );
            assert!(service.calls().contains(&PipelineKind::RangeScan));
        }
    }
}
