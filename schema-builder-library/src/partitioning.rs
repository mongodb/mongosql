pub(crate) use crate::partitioning::partition::{
    PARTITION_SIZE_IN_BYTES, Partition, SAMPLE_FALLBACK_SIZE, SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
    ScanStrategy,
};
use crate::{
    Error, LocalDataService,
    data_service::{AggregateOptions, CollectionInfo, CollectionType},
};
use bson::{Bson, Document, doc};
use futures::{StreamExt as _, TryStreamExt as _};
use tracing::{debug, instrument, trace, warn};

mod partition;
mod test;

#[derive(Debug, PartialEq, Clone)]
pub struct PartitionedCollection {
    pub partitions: Vec<Partition>,
    pub partition_key: String,
    pub hint: Option<Document>,
}

/// A utility function for determining the partitions of a collection.
///
/// If a collection is greater than 100MB and has 101 or more documents, it is partitioned into 100MB chunks.
/// Collections that do not meet this criteria are treated as a single partition.
///
/// The partition boundaries are found by sorting on the partition key and keeping a random
/// `$sampleRate` subset of the keys. Each kept key becomes the exclusive upper bound of one
/// partition and the inclusive lower bound of the next.
///
/// Collections of 1 GiB or more skip partitioning entirely and return a single partition marked
/// for sampling: [`crate::derive_schema_for_partition`] derives its schema from a `$sample` of
/// 1000 documents, and only scans the full
/// bounds if that sample fails. The partitioning pipeline sorts the whole collection, which on
/// large sharded clusters can run for hours before the cursor is killed.
///
/// If the partitioning query fails to open, the collection is returned as a single partition
/// covering its full bounds.
///
/// Note that the 100MB partition size matches the server's memory limit for blocking aggregation
/// stages, as noted in the
/// [aggregation pipeline limits](https://www.mongodb.com/docs/manual/core/aggregation-pipeline-limits/#memory-restrictions).
///
/// # Errors
///
/// Returns an error if the collection's size or bounds cannot be read, if a timeseries collection
/// has no `timeField`, or if a non-timeseries collection reports no `count`.
#[instrument(level = "trace", skip(service))]
#[expect(
    clippy::too_many_lines,
    reason = "kept as one linear function; each early return spells out its result explicitly"
)]
pub async fn get_partitions<S: LocalDataService>(
    service: &S,
    db: &str,
    collection_info: CollectionInfo,
) -> Result<PartitionedCollection, Error<S::Error>> {
    let size_info = get_size_counts(service, db, &collection_info.name).await?;
    let num_partitions = get_num_partitions(size_info.size, PARTITION_SIZE_IN_BYTES) as usize;

    // For timeseries collections, there is no `count` field reported, so we sample at a rate of
    // 1 per # of partitions. This is likely a much higher sample rate than for non-timeseries
    // collections, but it is our best effort for now to avoid running a full collection count.
    let (sample_rate, partition_key, hint) = match collection_info.collection_type {
        CollectionType::Timeseries => {
            let sample_rate = 1f64 / num_partitions as f64;

            let timeseries_options = collection_info
                .options
                .timeseries
                .ok_or_else(|| Error::NoTimeFieldSpecified(collection_info.name.clone()))?;

            let partition_key = timeseries_options.time_field;

            let hint = timeseries_options.meta_field.map(|meta_field| {
                doc! { meta_field: 1, partition_key.as_str(): 1 }
            });

            (sample_rate, partition_key, hint)
        }
        CollectionType::Collection | CollectionType::View => {
            let count = size_info.count.ok_or_else(|| {
                Error::MissingCountFieldForCollection(collection_info.name.clone())
            })?;
            (
                num_partitions as f64 / count as f64 * 2.0,
                "_id".to_string(),
                Some(doc! {"_id": 1}),
            )
        }
    };

    let (mut min_bound, max_bound) =
        get_bounds(service, db, &collection_info.name, partition_key.as_str()).await?;

    if size_info.size >= SAMPLE_FALLBACK_THRESHOLD_IN_BYTES {
        warn_sampling_selected(db, &collection_info.name, size_info.size);
        return Ok(PartitionedCollection {
            partitions: vec![Partition {
                min: min_bound,
                max: max_bound,
                is_max_bound_inclusive: true,
                strategy: ScanStrategy::Sample,
            }],
            partition_key,
            hint,
        });
    }

    // If the number of partitions is 1, no need to sample to determine partition boundaries. This
    // usually happens for very small collections.
    if num_partitions == 1 {
        return Ok(PartitionedCollection {
            partitions: vec![Partition {
                min: min_bound,
                max: max_bound,
                is_max_bound_inclusive: true,
                strategy: ScanStrategy::Range,
            }],
            partition_key,
            hint,
        });
    }

    let mut partitions = Vec::with_capacity(num_partitions);

    let sample_pipeline = vec![
        doc! { "$sort": {partition_key.as_str(): 1} },
        doc! { "$project": {partition_key.as_str(): 1} },
        doc! { "$match": { "$sampleRate": sample_rate } },
    ];

    // If the partitioning query fails, check the entire collection. This is safer than missing a
    // namespace.
    let Ok(cursor) = service
        .aggregate(
            db,
            &collection_info.name,
            sample_pipeline,
            AggregateOptions {
                key_hint: hint.clone(),
            },
        )
        .await
        .inspect_err(|e| {
            warn!(
                db,
                collection = %collection_info.name,
                "partition sampling failed; falling back to a single partition: {e}"
            );
        })
    else {
        return Ok(PartitionedCollection {
            partitions: vec![Partition {
                min: min_bound,
                max: max_bound,
                is_max_bound_inclusive: true,
                strategy: ScanStrategy::Range,
            }],
            partition_key,
            hint,
        });
    };

    let mut cursor = Box::pin(cursor);
    while let Some(doc) = cursor.try_next().await.map_err(Error::DataServiceError)? {
        let local_max = doc
            .get(partition_key.as_str())
            .unwrap_or(&Bson::MaxKey)
            .clone();
        partitions.push(Partition {
            min: min_bound.clone(),
            max: local_max.clone(),
            is_max_bound_inclusive: false,
            strategy: ScanStrategy::Range,
        });
        min_bound = local_max;
    }

    partitions.push(Partition {
        min: min_bound,
        max: max_bound,
        is_max_bound_inclusive: true,
        strategy: ScanStrategy::Range,
    });

    trace!("partitions: {:?}", partitions);

    Ok(PartitionedCollection {
        partitions,
        partition_key,
        hint,
    })
}

/// Warns that a collection is too large to partition and its schema will come from a sample.
///
/// Logged at `warn` so it shows up at the schema manager's default log verbosity: a sampled
/// schema may be missing fields that only appear in documents outside the sample.
fn warn_sampling_selected(db: &str, collection: &str, size_in_bytes: i64) {
    warn!(
        name: "schema.sampling.selected",
        db,
        collection,
        size_in_bytes,
        size_readable = %format_size(size_in_bytes),
        threshold_in_bytes = SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
        threshold_readable = %format_size(SAMPLE_FALLBACK_THRESHOLD_IN_BYTES),
        sample_size = SAMPLE_FALLBACK_SIZE,
        "collection size exceeds the sampling threshold; skipping partitioning and deriving the \
         schema from a $sample. The schema may not reflect every document in the collection."
    );
}

/// Utility struct for collection size and count information.
#[derive(Debug, PartialEq)]
pub(crate) struct CollectionSizes {
    pub size: i64,
    pub count: Option<i64>,
}

/// Reads a collection's size and document count with `$collStats`.
///
/// Sharded collections return one
/// [`$collStats`](https://www.mongodb.com/docs/manual/reference/operator/aggregation/collStats/)
/// document per shard, so the server sums them with `$group` and returns a single document. A
/// shard that does not report `count` (e.g. timeseries) contributes nothing to the sum and is
/// counted in `shardsWithCount` instead, so the count is `None` when no shard reports one.
///
/// # Errors
///
/// Returns an error if the query fails, if the returned document is malformed, if no document is
/// returned, or if the collection is empty.
#[instrument(level = "trace", skip(service))]
pub(crate) async fn get_size_counts<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
) -> Result<CollectionSizes, Error<S::Error>> {
    let pipeline = vec![
        // Run $collStats to get storage stats.
        // On sharded collections, returns document per shard.
        doc! {"$collStats": {"storageStats": {}}},
        // Group the totalSize, totalDocuments, and # of shards
        doc! {"$group": {
            "_id": null,
            "totalSizeBytes": {"$sum": "$storageStats.size"},
            "totalDocumentCount": {"$sum": "$storageStats.count"},
            "shardsWithCount": {"$sum": {
                "$cond": [{"$eq": [{"$type": "$storageStats.count"}, "missing"]}, 0, 1]
            }},
        }},
        // Convert size and doc counts to long, so we can translate to i64
        doc! {"$project": {
            "_id": 0,
            "totalSizeBytes": {"$toLong": "$totalSizeBytes"},
            "totalDocumentCount": {"$toLong": "$totalDocumentCount"},
            "shardsWithCount": {"$toLong": "$shardsWithCount"},
        }},
    ];

    let cursor = service
        .aggregate(db, collection, pipeline, AggregateOptions::default())
        .await
        .map_err(Error::DataServiceError)?;
    let mut cursor = Box::pin(cursor);

    // `$group` yields no document at all when its input is empty, so a missing document means
    // there were no stats to report.
    let totals = cursor
        .try_next()
        .await
        .map_err(Error::DataServiceError)?
        .ok_or_else(|| Error::NoCollectionStats(collection.to_string()))?;

    let size = totals
        .get_i64("totalSizeBytes")
        .map_err(|_| Error::BsonFailure)?;
    let count = totals
        .get_i64("totalDocumentCount")
        .map_err(|_| Error::BsonFailure)?;
    let shards_with_count = totals
        .get_i64("shardsWithCount")
        .map_err(|_| Error::BsonFailure)?;
    let count = (shards_with_count > 0).then_some(count);

    if size == 0 || count.is_some_and(|c| c == 0) {
        return Err(Error::EmptyCollection(collection.to_string()));
    }

    let sizes = CollectionSizes { size, count };
    debug!(
        name: "schema.collection.stats",
        db,
        collection,
        size_in_bytes = sizes.size,
        size_readable = %format_size(sizes.size),
        count = ?sizes.count,
        "collection size and count"
    );
    Ok(sizes)
}

/// Formats a byte count into a human-readable string, e.g. `1.3 GiB`.
///
/// Uses integer arithmetic (one decimal place) so the result is stable across platforms and does
/// not lose precision to floating-point conversion.
pub(crate) fn format_size(bytes: i64) -> String {
    const KIB: i64 = 1024;
    const MIB: i64 = KIB * 1024;
    const GIB: i64 = MIB * 1024;

    let (divisor, unit) = if bytes >= GIB {
        (GIB, "GiB")
    } else if bytes >= MIB {
        (MIB, "MiB")
    } else if bytes >= KIB {
        (KIB, "KiB")
    } else {
        (1, "B")
    };

    if divisor == 1 {
        format!("{bytes} {unit}")
    } else {
        let whole = bytes / divisor;
        let tenths = (bytes % divisor) * 10 / divisor;
        format!("{whole}.{tenths} {unit}")
    }
}

/// get_num_partitions determines the number of partitions by dividing the collection size by the
/// partition size (and adding 1).
#[instrument(level = "trace", skip_all)]
fn get_num_partitions(coll_size: i64, partition_size: i64) -> i64 {
    let num_parts = (coll_size as f64) / (partition_size as f64);
    num_parts as i64 + 1
}

/// get_bounds determines the minimum and maximum values for the _id field in the argued collection.
#[instrument(level = "trace", skip_all)]
pub(crate) async fn get_bounds<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
    partition_key: &str,
) -> Result<(Bson, Bson), Error<S::Error>> {
    Ok((
        get_bound(service, db, collection, partition_key, 1).await?,
        get_bound(service, db, collection, partition_key, -1).await?,
    ))
}

/// get_bound determines the minimum or maximum bound (depending on the direction) for the _id field
/// in the provided collection.
#[instrument(level = "trace", skip_all)]
async fn get_bound<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
    partition_key: &str,
    direction: i32,
) -> Result<Bson, Error<S::Error>> {
    let pipeline = vec![
        doc! {"$sort": {partition_key: direction}},
        doc! {"$limit": 1},
        doc! {"$project": {partition_key: 1}},
    ];
    let cursor = service
        .aggregate(db, collection, pipeline, AggregateOptions::default())
        .await
        .map_err(Error::DataServiceError)?;
    let mut cursor = Box::pin(cursor);

    let doc = cursor
        .next()
        .await
        .transpose()
        // The original code here returned `NoBounds` instead of a data service error
        .map_err(Error::DataServiceError)?
        .ok_or_else(|| Error::NoBounds(collection.to_string()))?;

    doc.get(partition_key)
        .cloned()
        .ok_or_else(|| Error::NoBounds(collection.to_string()))
}
