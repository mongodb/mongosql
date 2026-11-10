pub use crate::partitioning::partition::{
    PARTITION_SIZE_IN_BYTES, Partition, SAMPLE_FALLBACK_SIZE, SAMPLE_FALLBACK_THRESHOLD_IN_BYTES,
};
use crate::{
    Error, LocalDataService,
    data_service::{AggregateOptions, CollectionInfo, CollectionType},
};
use bson::{Bson, Document, doc};
use futures::{StreamExt as _, TryStreamExt as _};
use tokio::count;
use tracing::{info, instrument, trace};

mod partition;
mod test;

#[derive(Debug, PartialEq, Clone)]
pub struct PartitionedCollection {
    pub partitions: Vec<Partition>,
    pub partition_key: String,
    pub hint: Option<Document>,
}

/// A utility function for getting a single partition for a collection
/// This is used as a fallback if we are unable to derive a schema from an entire colleciton
/// A single partition gives us some coverage.
#[instrument(level = "trace", skip(service))]
pub async fn get_single_partition<S: LocalDataService>(
    service: &S,
    db: &str,
    collection_info: CollectionInfo,
) -> Result<PartitionedCollection, Error<S::Error>> {
    let (partition_key, hint) = match collection_info.collection_type {
        CollectionType::Timeseries => {
            let timeseries_options = collection_info
                .options
                .timeseries
                .ok_or_else(|| Error::NoTimeFieldSpecified(collection_info.name.clone()))?;

            let partition_key = timeseries_options.time_field;
            let hint = timeseries_options.meta_field.map(|meta_field| {
                doc! { meta_field: 1, partition_key.as_str(): 1 }
            });
            (partition_key, hint)
        }
        CollectionType::Collection | CollectionType::View => {
            ("_id".to_string(), Some(doc! {"_id": 1}))
        }
    };

    let (mut min_bound, max_bound) =
        get_bounds(service, db, &collection_info.name, partition_key.as_str()).await?;
    let partitions = vec![Partition {
        min: min_bound,
        max: max_bound,
        is_max_bound_inclusive: true,
    }];

    Ok(PartitionedCollection {
        partitions,
        partition_key,
        hint,
    })
}

/// A utility function for determining the partitions of a collection.
///
/// If a collection is greater than 100MB and has 101 or more documents, it is partitioned into 100MB chunks.
/// Collections that do not meet this criteria are treated as a single partition.
///
/// The minimum bound within each chunk is found
/// by using a $bucketAuto stage that groups on the _id field.
///
/// Note that the 100MB limit comes from the server, as noted in
/// [$bucketAuto docs](https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucketAuto/#-bucketauto-and-memory-restrictions).
#[instrument(level = "trace", skip(service))]
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

    // If the number of partitions is 1, no need to sample to determine partition boundaries. This
    // usually happens for very small collections.
    if num_partitions == 1 {
        return Ok(PartitionedCollection {
            partitions: vec![Partition {
                min: min_bound,
                max: max_bound,
                is_max_bound_inclusive: true,
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
            tracing::warn!(
                db,
                collection = %collection_info.name,
                "partition sampling failed; falling back to a single partition: {e}"
            )
        })
    else {
        return Ok(PartitionedCollection {
            partitions: vec![Partition {
                min: min_bound,
                max: max_bound,
                is_max_bound_inclusive: true,
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
        });
        min_bound = local_max;
    }

    partitions.push(Partition {
        min: min_bound,
        max: max_bound,
        is_max_bound_inclusive: true,
    });

    trace!("partitions: {:?}", partitions);

    Ok(PartitionedCollection {
        partitions,
        partition_key,
        hint,
    })
}

/// Utility struct for collection size and count information.
#[derive(Debug, PartialEq)]
pub struct CollectionSizes {
    pub size: i64,
    pub count: Option<i64>,
}

/// get_size_counts uses the $collStats aggregation stage to get size and count information for the
/// argued collection.  Sharded collections return one `$collStats` document per shard:
/// https://www.mongodb.com/docs/manual/reference/operator/aggregation/collstats/#mongodb-pipeline-pipe.-collStats
/// To ensure accurate calculation, we sum sizes and counts across every document.
/// A shard that does not report `count` (e.g. timeseries) is skipped rather than discarding the counts reported by other shards; the count is
/// `None` when no shard reports one.
#[instrument(level = "trace", skip(service))]
pub async fn get_size_counts<S: LocalDataService>(
    service: &S,
    db: &str,
    collection: &str,
) -> Result<CollectionSizes, Error<S::Error>> {
    let cursor = service
        .aggregate(
            db,
            collection,
            vec![doc! {"$collStats": {"storageStats": {}}}],
            AggregateOptions::default(),
        )
        .await
        .map_err(Error::DataServiceError)?;
    let mut cursor = Box::pin(cursor);

    let mut size_sum: i64 = 0;
    let mut count_sum: Option<i64> = None;
    let mut has_coll_stats = false;

    while let Some(document) = cursor.try_next().await.map_err(Error::DataServiceError)? {
        has_coll_stats = true;

        let storage_stats = document
            .get_document("storageStats")
            .map_err(|_| Error::BsonFailure)?;

        let shard_size = match storage_stats.get("size") {
            Some(Bson::Int32(size)) => i64::from(*size),
            Some(Bson::Int64(size)) => *size,
            _ => return Err(Error::BsonFailure),
        };

        let shard_count = match storage_stats.get("count") {
            Some(Bson::Int32(count)) => Some(i64::from(*count)),
            Some(Bson::Int64(count)) => Some(*count),
            None => None,
            _ => return Err(Error::BsonFailure),
        };

        size_sum += shard_size;
        // Keep every count we do find; a shard without a count is skipped.
        if let Some(count) = shard_count {
            count_sum = Some(count_sum.unwrap_or(0) + count);
        }
    }

    if !has_coll_stats {
        return Err(Error::NoCollectionStats(collection.to_string()));
    }

    if size_sum == 0 || count_sum.is_some_and(|c| c == 0) {
        Err(Error::EmptyCollection(collection.to_string()))
    } else {
        let sizes = CollectionSizes {
            size: size_sum,
            count: count_sum,
        };
        info!(
            db,
            collection,
            size_in_bytes = sizes.size,
            size_readable = %format_size(sizes.size),
            count = ?sizes.count,
            "collection size and count"
        );
        Ok(sizes)
    }
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
