use std::sync::Arc;

use bson::{Document, doc};
use futures::TryStreamExt as _;
use mongosql::schema::Schema;
use schema_derivation::schema_for_document;
use tracing::{info, instrument, warn};

use crate::data_service::{AggregateOptions, LocalDataService};
use crate::{Error, data_service::CollectionInfo, partitioning::Partition};
use crate::{PartitionedCollection, get_partitions};

/// The amount of samples to fetch for view schema derivation
const VIEW_SAMPLE_SIZE: i64 = 1000;

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

/// A utility function for deriving the schema for a single partition of a collection.
#[instrument(level = "trace", skip_all)]
pub async fn derive_schema_for_partition<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
    initial_schema_doc: Option<Arc<Schema>>,
    single_partition: SinglePartition,
) -> Result<Schema, Error<S::Error>> {
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
        // Carried across documents: the value of `schema.union(&iter_schema)` as of the
        // previous document. Neither operand changes between the end of one document's
        // check and the start of the next, so it can be reused instead of recomputed.
        //
        // Before the first document this is just `schema`: `iter_schema` is `Unsat`, and
        // `union` returns the other operand simplified when either side is `Unsat`, while
        // `schema` is already simplified everywhere it is assigned.
        let mut old_schema = schema.clone();
        let mut cursor = Box::pin(cursor);
        while let Some(doc) = cursor.try_next().await.map_err(Error::DataServiceError)? {
            info!(db, collection, "processing partition {partition_ix}");
            if let Some(id) = doc.get(partition_key) {
                partition.min = id.clone();

                // Some documents cannot match a $jsonSchema derived from themselves, and so are
                // returned by the query above no matter how much schema we have accumulated. The
                // known cases are field names containing a `.`, which $jsonSchema's `required`
                // keyword resolves as a path rather than as a literal name, and empty keys on
                // older servers due to a bug. See SERVER-92443 and
                // https://github.com/10gen/schema-manager-rs/pull/754 for more context. To avoid
                // getting caught in an infinite loop, we push to a list of ignored IDs in the
                // event empty keys or field names containing a `.` exists in the partition.
                //
                // Note that the comparison must be against the accumulated `schema`, not against
                // `iter_schema` alone: `iter_schema` restarts at `Unsat` on every iteration, so
                // comparing against it would never ignore the first document of a batch, and a
                // batch holding exactly one such document would loop forever.
                iter_schema = iter_schema.union(&schema_for_document(&doc));
                let new_schema = schema.union(&iter_schema);

                if old_schema == new_schema {
                    ignored_ids.push(id.clone());
                }
                old_schema = new_schema;
                no_result = false;
            } else {
                warn!(
                    db,
                    collection, "document is missing the {partition_key} field"
                );
                continue;
            };
        }
        // Documents are processed in `partition_key` order and `partition.min` advances to
        // the last one seen, so every ignored id below the minimum is already excluded by the
        // `$gte` bound in the match. Only ids at the minimum itself still need to be listed.
        // Without this, `ignored_ids` -- and the `$nin` array resent on every subsequent
        // query for this partition -- grows without bound.
        ignored_ids.retain(|id| *id == partition.min);

        if no_result {
            break;
        }

        // `old_schema` already holds `schema.union(&iter_schema)` as of the last document,
        // produced by the same call on the same operands, so reuse it rather than
        // recomputing the union.
        schema = old_schema;

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
