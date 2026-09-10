use bson::{Bson, Document, doc};
use tracing::instrument;

pub const PARTITION_SIZE_IN_BYTES: i64 = 100 * 1024 * 1024; // 100 MB

/// Returns true when `min` and `max` can be compared using match-language range operators.
/// They can be compared if they are both numeric types, if they are the same type, or if one is MinKey and the other is MaxKey
fn bounds_are_comparable(min: &Bson, max: &Bson) -> bool {
    fn is_numeric(bound: &Bson) -> bool {
        matches!(
            bound,
            Bson::Int32(_) | Bson::Int64(_) | Bson::Double(_) | Bson::Decimal128(_)
        )
    }

    (is_numeric(min) && is_numeric(max))
        || min.element_type() == max.element_type()
        || matches!((min, max), (Bson::MinKey, Bson::MaxKey))
}

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
    // If the min and max bounds are not comparable in match language, the $expr language will be used.
    //
    //  $expr comparisons use BSON total sort order, so a partition key of any type can fall between the bounds, while match language
    // is type-bracketed and only matches keys that have the same type as the bound.
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

        // If the min and max bounds are not comparable in match language, fall back to the
        // $expr language
        if !bounds_are_comparable(&self.min, &self.max) {
            let key_path = format!("${partition_key}");
            let mut expr_body = doc! {
                "$expr": {
                    "$and": [
                        // $literal keeps a $-prefixed ignored value from being read as a
                        // field path.
                        doc! {"$not": {"$in": [&key_path, {"$literal": ignored_ids.to_vec()}]}},
                        doc! {"$gte": [&key_path, self.min.clone()]},
                        doc! {lt_op: [&key_path, self.max.clone()]},
                    ]
                }
            };

            // $jsonSchema has no $expr equivalent, so the schema exclusion stays in match
            // language as a sibling of $expr, which is valid within a single $match.
            if let Some(schema) = doc {
                expr_body.insert("$nor", vec![schema]);
            }

            doc! {
                "$match": expr_body
            }
        } else {
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
}
