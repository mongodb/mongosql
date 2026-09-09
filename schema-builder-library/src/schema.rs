use std::sync::Arc;

use bson::{Bson, Document, doc};
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
    // At most one blocker at a time: a document that cannot match a $jsonSchema derived
    // from itself sits at `partition.min`, where the inclusive `$gte` bound cannot exclude
    // it. Once it is skipped, `min` advances past it and the bound takes over.
    let mut ignored_min_id: Option<Bson> = None;
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

        // A document only reaches us if it failed `$nor: [$jsonSchema]`, i.e. the server
        // judged it not to match the schema accumulated so far. If the batch also failed to
        // widen that schema, the two facts contradict each other: our `Schema` already
        // covers the document, yet the `$jsonSchema` rendering of that `Schema` does not
        // match it. The gap is in what the operator can express, not in the data -- its
        // `required` keyword resolves `"a.b"` as the path `a` -> `b` rather than as a
        // literal key, and empty keys behave likewise on servers affected by SERVER-92443.
        // See https://github.com/10gen/schema-manager-rs/pull/754 for more context.
        //
        // No amount of further accumulation will make such a document match, so the schema
        // filter can never exclude it. The `$gte` bound cannot either, being inclusive of
        // `partition.min`. Recording its id is the only thing left, which makes every
        // iteration either forward progress or the permanent removal of one blocker.
        //
        // (Once the schema is unstable it stops absorbing new keys, so this can also fire
        // for a document whose fields were deliberately discarded rather than covered.
        // Ignoring it is consistent with that decision, and the `saw_unstable` break below
        // bounds the loop either way.)
        //
        // Otherwise `partition.min` has advanced past any previous blocker, which the `$gte`
        // bound now excludes on its own, so the id no longer needs to be listed.
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
