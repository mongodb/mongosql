use futures::TryStreamExt;
use mongodb::{
    bson::{doc, Bson, Document},
    Collection,
};
use mongosql::schema::Schema;
use tracing::{instrument, trace};

#[cfg(test)]
mod test;

use crate::{Error, Result, PARTITION_MIN_DOC_COUNT, PARTITION_SIZE_IN_BYTES};

#[derive(Debug, PartialEq, Clone)]
pub struct Partition {
    pub min: Bson,
    pub max: Bson,
    pub is_max_bound_inclusive: bool,
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
#[instrument(level = "trace", skip(collection))]
pub(crate) async fn get_partitions(collection: &Collection<Document>) -> Result<Vec<Partition>> {
    let size_info = get_size_counts(collection).await?;
    let num_partitions = get_num_partitions(size_info.size, PARTITION_SIZE_IN_BYTES) as usize;
    let (mut min_bound, max_bound) = get_bounds(collection).await?;
    let mut partitions = Vec::with_capacity(num_partitions);

    let buckets = if size_info.count >= PARTITION_MIN_DOC_COUNT && num_partitions > 1 {
        num_partitions
    } else {
        1
    };

    let mut cursor = collection
        .aggregate(vec![
            doc! { "$project": {"_id": 1} },
            doc! { "$bucketAuto": {
                "groupBy": "$_id",
                "buckets": buckets as i32
            }},
        ])
        .await?;

    while let Some(doc) = cursor.try_next().await.unwrap() {
        let doc = doc.get("_id").unwrap().as_document().unwrap();
        let max = doc.get("max").cloned().unwrap_or(Bson::MaxKey);
        let is_bucket_max_equal_to_collection_max = max == max_bound;
        partitions.push(Partition {
            min: min_bound.clone(),
            max: max.clone(),
            // The max bound of Partition should be considered inclusive only
            // if it is the final partition.
            is_max_bound_inclusive: is_bucket_max_equal_to_collection_max,
        });
        if max != max_bound {
            min_bound = max;
        }
    }

    trace!("partitions: {:?}", partitions);

    Ok(partitions)
}

/// Utility struct for collection size and count information.
struct CollectionSizes {
    pub size: i64,
    pub count: i64,
}

/// get_size_counts uses the $collStats aggregation stage to get size and count information for the
/// argued collection.
#[instrument(level = "trace", skip(collection))]
async fn get_size_counts(collection: &Collection<Document>) -> Result<CollectionSizes> {
    let mut cursor = collection
        .aggregate(vec![doc! {"$collStats": {"storageStats": {}}}])
        .await
        .map_err(|_| Error::NoCollectionStats(collection.name().to_string()))?;
    if let Some(stats) = cursor.try_next().await.unwrap() {
        let stats = stats
            .get_document("storageStats")
            .map_err(|_| Error::BsonFailure)?;
        let size = stats
            .get("size")
            .cloned()
            .ok_or_else(|| Error::BsonFailure)?;
        let count = stats
            .get("count")
            .cloned()
            .ok_or_else(|| Error::BsonFailure)?;
        let (size, count) = match (size, count) {
            (Bson::Int32(size), Bson::Int32(count)) => (size as i64, count as i64),
            (Bson::Int32(size), Bson::Int64(count)) => (size as i64, count),
            (Bson::Int64(size), Bson::Int32(count)) => (size, count as i64),
            (Bson::Int64(size), Bson::Int64(count)) => (size, count),
            _ => {
                return Err(Error::BsonFailure);
            }
        };
        if size == 0 || count == 0 {
            return Err(Error::EmptyCollection(collection.name().to_string()));
        } else {
            return Ok(CollectionSizes { size, count });
        }
    }
    Err(Error::NoCollectionStats(collection.name().to_string()))
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
pub(crate) async fn get_bounds(collection: &Collection<Document>) -> Result<(Bson, Bson)> {
    Ok((
        get_bound(collection, 1).await?,
        get_bound(collection, -1).await?,
    ))
}

/// get_bound determines the minimum or maximum bound (depending on the direction) for the _id field
/// in the provided collection.
#[instrument(level = "trace", skip_all)]
async fn get_bound(collection: &Collection<Document>, direction: i32) -> Result<Bson> {
    let mut cursor = collection
        .aggregate(vec![
            doc! {"$sort": {"_id": direction}},
            doc! {"$limit": 1},
            doc! {"$project": {"_id": 1}},
        ])
        .await
        .map_err(|e| Error::NoBounds(format!("{}: {e}", collection.name())))?;
    if let Some(doc) = cursor.try_next().await? {
        return doc
            .get("_id")
            .cloned()
            .ok_or(Error::NoBounds(collection.name().to_string()));
    }
    Err(Error::NoBounds(collection.name().to_string()))
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an optional Schema. If the Schema is None, the $match will only
// be based on the Partition bounds.
// This function also accepts a list of ignored_ids which will be used to exclude certain documents.
#[instrument(level = "trace", skip_all)]
pub(crate) fn generate_partition_match(
    partition: &Partition,
    schema: Option<Schema>,
    ignored_ids: &[Bson],
) -> Result<Document> {
    generate_partition_match_with_doc(
        partition,
        schema.map(Document::try_from).transpose()?,
        ignored_ids,
    )
}

// generate_partition_match_with_doc generates the $match stage for sampling based on the partition
// additional_properties and an optional input jsonSchema. If the jsonSchema doc is None, the $match
// will only be based on the Partition bounds.
// This function also accepts a list of ignored_ids which will be used to exclude certain documents.
#[instrument(level = "trace", skip_all)]
pub(crate) fn generate_partition_match_with_doc(
    partition: &Partition,
    schema: Option<Document>,
    ignored_ids: &[Bson],
) -> Result<Document> {
    let lt_op = if partition.is_max_bound_inclusive {
        "$lte"
    } else {
        "$lt"
    };

    let mut match_body = doc! {
        "_id": {
            "$nin": ignored_ids,
            "$gte": partition.min.clone(),
            lt_op: partition.max.clone(),
        }
    };
    if let Some(schema) = schema {
        match_body.insert("$nor", vec![schema]);
    }
    Ok(doc! {
        "$match": match_body
    })
}
