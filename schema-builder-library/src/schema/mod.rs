use std::sync::Arc;

use bson::{Document, doc};
use futures::{TryStreamExt as _, future};
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;
use tracing::{info, instrument, warn};

pub(crate) mod initial_schema;

use crate::context::ContextHandle;
use crate::data_service::AggregateOptions;
use crate::{DataService, partitioning::Partition};
use crate::{
    Error, Result, VIEW_SAMPLE_SIZE, data_service::CollectionInfo,
    partitioning::PartitionedCollection,
};

#[derive(Debug, PartialEq, Clone)]
pub struct SinglePartition {
    pub partition: Partition,
    pub partition_key: String,
    pub hint: Option<Document>,
    pub partition_ix: usize,
}

pub const PARTITION_DOCS_PER_ITERATION: i64 = 20;

/// A utility function for deriving the schema for a collection based on its partitions.
///
/// For each partition:
///  1. If there is an initial schema
///     The initial schema is used as the "seed" for the partition.
///
///  2. If there is no initial schema
///     An aggregation operation is issued that sorts based on the partition key and limits the result set to
///     20 documents. The schema for these documents is computed and "seeds" the schema for the
///     partition.
///
/// The builder then enters a loop where it repeatedly issues query operations for documents that do
/// not match the computed schema. The schema for each resulting document is calculated and merged
/// with the existing schema, resulting in a new schema that meets all encountered documents so far.
/// This operation repeats until there are no more results within the partition.
///
/// The results of each partition are unioned together to produce the schema of the entire
/// collection.
#[instrument(level = "trace", skip_all)]
pub(crate) async fn derive_schema_for_partitions<S: DataService>(
    ctx: ContextHandle<S>,
    db: &str,
    collection: &str,
    initial_schema_doc: Option<Arc<Schema>>,
    task_semaphore: std::sync::Arc<tokio::sync::Semaphore>,
    partitioned_collection: PartitionedCollection,
) -> Option<Schema> {
    let partition_tasks = partitioned_collection
        .partitions
        .into_iter()
        .enumerate()
        .map(|(ix, partition)| {
            let ctx = Arc::clone(&ctx);
            let initial_schema_doc = initial_schema_doc.clone();
            let task_semaphore = task_semaphore.clone();

            let single_partition = SinglePartition {
                partition,
                partition_key: partitioned_collection.partition_key.clone(),
                hint: partitioned_collection.hint.clone(),
                partition_ix: ix,
            };

            async move {
                // Acquire a permit from the semaphore to limit concurrency
                #[allow(clippy::unwrap_used)]
                let _permit = task_semaphore.acquire().await.unwrap();

                // If we encounter an error when processing a partition,
                // we effectively ignore it by saying the schema is Unsat.
                // Note that for any schema s, Unsat.union(s) == s. Users
                // can check their error notifications to see that schema
                // building for this collection may be incomplete and can
                // consider rerunning the builder if they wish.
                derive_schema_for_partition(
                    ctx.service(),
                    db,
                    collection,
                    initial_schema_doc,
                    single_partition,
                )
                .await
                .inspect_err(|e| {
                    tracing::warn!(
                        db,
                        collection,
                        "failed to derive schema for partition {ix} with error: {e}"
                    )
                })
                .unwrap_or(Schema::Unsat)
            }
        });

    // Here, we await all partitions and then union them together to create the
    // full collection schema. Note that we could union the partition schemas as
    // they are produced for a marginal speedup, however that would require
    // a mutable Schema value that is guarded by a lock. The overhead of locking
    // and unioning per thread is likely equivalent to the more straightforward
    // wait-and-then-union solution we have here.
    future::join_all(partition_tasks)
        .await
        .into_iter()
        .reduce(|full_coll_schema, part_schema| full_coll_schema.union(&part_schema))
}

/// A utility function for deriving the schema for a single partition of a collection.
#[instrument(level = "trace", skip_all)]
pub(crate) async fn derive_schema_for_partition<S: DataService>(
    service: &S,
    db: &str,
    collection: &str,
    initial_schema_doc: Option<Arc<Schema>>,
    single_partition: SinglePartition,
) -> Result<Schema, S::Error> {
    let mut ignored_ids = Vec::new();
    let mut partition = single_partition.partition;
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
            partition.generate_match(doc, &ignored_ids, partition_key),
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
                let old_schema = iter_schema.clone();
                iter_schema = iter_schema.union(&schema_for_document(&doc));

                // There is a bug in Server where $jsonSchema operator don't work with empty keys.
                // To avoid getting caught in an infinite loop, we push to a list of ignored IDs in the event
                // empty keys exists in the partition.
                // See SERVER-92443 and https://github.com/10gen/schema-manager-rs/pull/754 for more context.

                if old_schema == iter_schema {
                    ignored_ids.push(id.clone());
                }
                no_result = false;
            } else {
                warn!(db, collection, "document {partition_key} field");
                continue;
            };
        }
        if no_result {
            break;
        }

        schema = schema.union(&iter_schema);

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

/// derive_schema_for_view takes a CollectionInfo and executes the pipeline
/// against the viewOn collection to generate a schema for the view.
/// It does this by first prepending $sample to the pipeline
#[instrument(level = "trace", skip_all)]
pub(crate) async fn derive_schema_for_view<S: DataService>(
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
        // Notify every 100 iterations, so it isn't too spammy
        if iterations.is_multiple_of(100) {
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
