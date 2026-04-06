use std::sync::Arc;

use futures::{TryStreamExt, future};
use mongodb::{
    Collection, Database,
    bson::{self, Document, doc},
    options::AggregateOptions,
};
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;
use tracing::{info, instrument, warn};

pub(crate) mod initial_schema;

use crate::partitioning::Partition;
use crate::{CollectionDoc, Error, Result, VIEW_SAMPLE_SIZE, partitioning::PartitionedCollection};

#[derive(Debug, PartialEq, Clone)]
pub struct SinglePartition {
    pub partition: Partition,
    pub partition_key: String,
    pub hint: Option<mongodb::options::Hint>,
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
pub(crate) async fn derive_schema_for_partitions(
    db_name: String,
    collection: &Collection<Document>,
    initial_schema_doc: Option<Arc<Schema>>,
    rt_handle: &tokio::runtime::Handle,
    task_semaphore: std::sync::Arc<tokio::sync::Semaphore>,
    partitioned_collection: PartitionedCollection,
) -> Result<Option<Schema>> {
    let partition_tasks = partitioned_collection
        .partitions
        .into_iter()
        .enumerate()
        .map(|(ix, partition)| {
            let db_name = db_name.clone();
            let collection = collection.clone();
            let initial_schema_doc = initial_schema_doc.clone();
            let task_semaphore = task_semaphore.clone();

            let single_partition = SinglePartition {
                partition,
                partition_key: partitioned_collection.partition_key.clone(),
                hint: partitioned_collection.hint.clone(),
                partition_ix: ix,
            };

            rt_handle.spawn(async move {
                // Acquire a permit from the semaphore to limit concurrency
                #[allow(clippy::unwrap_used)]
                let _permit = task_semaphore.acquire().await.unwrap();
                let schema_res = derive_schema_for_partition(
                    &db_name,
                    &collection,
                    initial_schema_doc,
                    single_partition,
                )
                .await;

                // If we encounter an error when processing a partition,
                // we effectively ignore it by saying the schema is Unsat.
                // Note that for any schema s, Unsat.union(s) == s. Users
                // can check their error notifications to see that schema
                // building for this collection may be incomplete and can
                // consider rerunning the builder if they wish.
                schema_res
                    .inspect_err(|e| {
                        warn!(
                            db = db_name,
                            collection = collection.name(),
                            "failed to derive schema for partition {ix} with error: {e}"
                        )
                    })
                    .unwrap_or(Schema::Unsat)
            })
        });

    // Here, we await all partitions and then union them together to create the
    // full collection schema. Note that we could union the partition schemas as
    // they are produced for a marginal speedup, however that would require
    // a mutable Schema value that is guarded by a lock. The overhead of locking
    // and unioning per thread is likely equivalent to the more straightforward
    // wait-and-then-union solution we have here.
    Ok(future::try_join_all(partition_tasks)
        .await
        .map_err(Error::TokioError)?
        .into_iter()
        .reduce(|full_coll_schema, part_schema| full_coll_schema.union(&part_schema)))
}

/// A utility function for deriving the schema for a single partition of a collection.
#[instrument(level = "trace", skip_all)]
pub(crate) async fn derive_schema_for_partition(
    db_name: &str,
    collection: &Collection<Document>,
    initial_schema_doc: Option<Arc<Schema>>,
    single_partition: SinglePartition,
) -> Result<Schema> {
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
        info!(
            db = db_name,
            collection = collection.name(),
            "querying partition: {partition_ix}"
        );

        // This is a somewhat expensive clone, but there isn't a try_from for
        // a schema reference :(
        let doc = (schema != Schema::Unsat)
            .then(|| bson::Document::try_from(schema.clone()))
            .transpose()?;

        let mut cursor = collection
            .aggregate(vec![
                partition.generate_match(doc, &ignored_ids, partition_key),
                doc! { "$sort": {partition_key: 1}},
                doc! { "$limit": PARTITION_DOCS_PER_ITERATION },
            ])
            .with_options(AggregateOptions::builder().hint(hint.clone()).build())
            .await?;

        let mut no_result = true;
        let mut iter_schema = Schema::Unsat;
        while let Some(doc) = cursor.try_next().await? {
            info!(
                db = db_name,
                collection = collection.name(),
                "processing partition {partition_ix}"
            );
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
                warn!(
                    db = db_name,
                    collection = collection.name(),
                    "document {partition_key} field"
                );
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

/// derive_schema_for_view takes a CollectionDoc and executes the pipeline
/// against the viewOn collection to generate a schema for the view.
/// It does this by first prepending $sample to the pipeline
#[instrument(level = "trace", skip_all)]
pub(crate) async fn derive_schema_for_view(
    view: &CollectionDoc,
    database: &Database,
) -> Option<Schema> {
    let Some(ref view_options) = view.options.view_options else {
        unreachable!("derive_schema_for_view was a passed a CollectionDoc without view options.")
    };

    let pipeline = vec![doc! { "$sample": { "size": VIEW_SAMPLE_SIZE } }]
        .into_iter()
        .chain(view_options.pipeline.clone().into_iter())
        .collect::<Vec<Document>>();

    let Ok(mut cursor) = database
        .collection::<Document>(&view_options.view_on)
        .aggregate(pipeline)
        .await
        .inspect_err(|e| {
            warn!(
                db = database.name(),
                view_name = view.name,
                "view sampling encountered an error: {e}"
            )
        })
    else {
        return None;
    };

    let mut schema = None;
    let mut iterations = 0u64;
    while let Some(Ok(doc)) = cursor.try_next().await.transpose() {
        // Notify every 100 iterations, so it isn't too spammy
        if iterations.is_multiple_of(100) {
            info!(
                db = database.name(),
                view_name = view.name,
                iteration = iterations,
                "Sampling view"
            );
        }

        schema = schema.map_or(Some(schema_for_document(&doc)), |s: Schema| {
            Some(s.union(&schema_for_document(&doc)))
        });
        iterations += 1;
    }

    schema
}
