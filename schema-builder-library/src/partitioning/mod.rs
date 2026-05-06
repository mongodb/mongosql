pub(crate) use crate::partitioning::partition::{PARTITION_SIZE_IN_BYTES, Partition};
use crate::{
    DataService, Error, Result,
    data_service::{AggregateOptions, CollectionInfo, CollectionType},
};
use bson::{Bson, Document, doc};
use futures::{StreamExt as _, TryStreamExt as _};
use tracing::{instrument, trace};

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
/// The minimum bound within each chunk is found
/// by using a $bucketAuto stage that groups on the _id field.
///
/// Note that the 100MB limit comes from the server, as noted in
/// [$bucketAuto docs](https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucketAuto/#-bucketauto-and-memory-restrictions).
#[instrument(level = "trace", skip(service))]
pub(crate) async fn get_partitions<S: DataService>(
    service: &S,
    db: &str,
    collection_info: CollectionInfo,
) -> Result<PartitionedCollection, S::Error> {
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
pub(crate) struct CollectionSizes {
    pub size: i64,
    pub count: Option<i64>,
}

/// get_size_counts uses the $collStats aggregation stage to get size and count information for the
/// argued collection.
#[instrument(level = "trace", skip(service))]
pub(crate) async fn get_size_counts<S: DataService>(
    service: &S,
    db: &str,
    collection: &str,
) -> Result<CollectionSizes, S::Error> {
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

    let stats = cursor
        .next()
        .await
        .transpose()
        .map_err(Error::DataServiceError)?
        .ok_or_else(|| Error::NoCollectionStats(collection.to_string()))?;

    let stats = stats
        .get_document("storageStats")
        .map_err(|_| Error::BsonFailure)?;
    let size = stats.get("size").cloned().ok_or(Error::BsonFailure)?;
    // For timeseries collections, there is no `count` field reported, so we ignore it.
    let count = stats.get("count").cloned();

    let (size, count) = match (size, count) {
        (Bson::Int32(size), Some(Bson::Int32(count))) => (size as i64, Some(count as i64)),
        (Bson::Int32(size), Some(Bson::Int64(count))) => (size as i64, Some(count)),
        (Bson::Int64(size), Some(Bson::Int32(count))) => (size, Some(count as i64)),
        (Bson::Int64(size), Some(Bson::Int64(count))) => (size, Some(count)),
        (Bson::Int32(size), None) => (size as i64, None),
        (Bson::Int64(size), None) => (size, None),
        _ => {
            return Err(Error::BsonFailure);
        }
    };

    if size == 0 || count.is_some_and(|c| c == 0) {
        Err(Error::EmptyCollection(collection.to_string()))
    } else {
        Ok(CollectionSizes { size, count })
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
pub(crate) async fn get_bounds<S: DataService>(
    service: &S,
    db: &str,
    collection: &str,
    partition_key: &str,
) -> Result<(Bson, Bson), S::Error> {
    Ok((
        get_bound(service, db, collection, partition_key, 1).await?,
        get_bound(service, db, collection, partition_key, -1).await?,
    ))
}

/// get_bound determines the minimum or maximum bound (depending on the direction) for the _id field
/// in the provided collection.
#[instrument(level = "trace", skip_all)]
async fn get_bound<S: DataService>(
    service: &S,
    db: &str,
    collection: &str,
    partition_key: &str,
    direction: i32,
) -> Result<Bson, S::Error> {
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
