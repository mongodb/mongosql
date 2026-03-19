use bson::{Bson, Document, doc};
use tracing::instrument;

pub const PARTITION_SIZE_IN_BYTES: i64 = 100 * 1024 * 1024; // 100 MB

#[derive(Debug, PartialEq, Clone)]
pub struct Partition {
    pub min: Bson,
    pub max: Bson,
    pub is_max_bound_inclusive: bool,
}

impl Partition {
    // generate_partition_match generates the $match stage for sampling based on the partition, an
    // optional schema, and a list of _id values to ignore. If the Schema is None, the $match will
    // only be based on the Partition bounds and the ignored_ids list.
    #[instrument(level = "trace", skip_all)]
    pub fn generate_match(
        &self,
        doc: Option<Document>,
        ignored_ids: &[Bson],
        partition_key: &str,
    ) -> Document {
        let lt_op = if self.is_max_bound_inclusive {
            "$lte"
        } else {
            "$lt"
        };

        let mut match_body = doc! {
            partition_key: {
                "$nin": ignored_ids,
                "$gte": self.min.clone(),
                lt_op: self.max.clone(),
            }
        };
        if let Some(schema) = doc {
            match_body.insert("$nor", vec![schema]);
        }
        doc! {
            "$match": match_body
        }
    }
}
