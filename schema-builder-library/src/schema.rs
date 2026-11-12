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
/// If the partition has ScanStrategy::Sample, the schema is generated via sampling.
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
            "Sampling returned no schema; falling back to scanning the entire collection as a \
             single partition."
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
/// Returns `None` if the query fails to open or returns no documents, so the caller can choose a
/// fallback.
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
}
