use futures::{future, TryStreamExt};
use mongodb::{
    bson::{doc, Bson, Document},
    options::AggregateOptions,
    Collection, Database,
};
use mongosql::schema::{Atomic, JaccardIndex, Schema};
use tracing::instrument;

use crate::{
    generate_partition_match, generate_partition_match_with_doc, notify, CollectionDoc, Error,
    Partition, Result, SamplerAction, SamplerNotification, PARTITION_DOCS_PER_ITERATION,
    VIEW_SAMPLE_SIZE,
};

#[cfg(test)]
pub mod test;

/// A utility function for deriving the schema for a collection based on its partitions.
///
/// For each partition:
///  1. If there is an initial schema
///     The initial schema is used as the "seed" for the partition.
///
///  2. If there is no initial schema
///     An aggregation operation is issued that sorts based on the _id and limits the result set to
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
    col_parts: Vec<Partition>,
    initial_schema_doc: Option<Document>,
    rt_handle: &tokio::runtime::Handle,
    tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
) -> Result<Option<Schema>> {
    let partition_tasks = col_parts.into_iter().enumerate().map(|(ix, partition)| {
        let db_name = db_name.clone();
        let tx_notifications = tx_notifications.clone();
        let collection = collection.clone();
        let initial_schema_doc = initial_schema_doc.clone();
        rt_handle.spawn(async move {
            let schema_res = derive_schema_for_partition(
                db_name.clone(),
                &collection,
                partition,
                initial_schema_doc,
                tx_notifications.clone(),
                ix,
            )
            .await;

            let schema = match schema_res {
                Err(e) => {
                    notify!(
                        &tx_notifications,
                        SamplerNotification {
                            db: db_name.clone(),
                            collection_or_view: collection.name().to_string(),
                            action: SamplerAction::Warning {
                                message: format!(
                                    "failed derive schema for partition {ix} with error {e}",
                                ),
                            },
                        },
                    );
                    // If we encounter an error when processing a partition,
                    // we effectively ignore it by saying the schema is Unsat.
                    // Note that for any schema s, Unsat.union(s) == s. Users
                    // can check there error notifications to see that schema
                    // building for this collection may be incomplete and can
                    // consider rerunning the builder if they wish.
                    Schema::Unsat
                }
                Ok(schema) => schema,
            };

            drop(tx_notifications);
            schema
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
async fn derive_schema_for_partition(
    db_name: String,
    collection: &Collection<Document>,
    mut partition: Partition,
    initial_schema_doc: Option<Document>,
    tx_notifications: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
    partition_ix: usize,
) -> Result<Schema> {
    let mut schema: Option<Schema> = initial_schema_doc
        .clone()
        .map(Schema::try_from)
        .transpose()?;
    let mut ignored_ids = vec![];
    let mut first_stage = Some(generate_partition_match_with_doc(
        &partition,
        initial_schema_doc,
        &ignored_ids,
    )?);

    loop {
        if first_stage.is_none() {
            first_stage = Some(generate_partition_match(
                &partition,
                schema.clone(),
                &ignored_ids,
            )?);
        };
        notify!(
            &tx_notifications,
            SamplerNotification {
                db: db_name.clone(),
                collection_or_view: collection.name().to_string(),
                action: SamplerAction::Querying {
                    partition: partition_ix as u16,
                },
            },
        );
        let mut cursor = collection
            .aggregate(
                vec![
                    // first_stage must be Some here.
                    first_stage.unwrap(),
                    doc! { "$sort": {"_id": 1}},
                    doc! { "$limit": PARTITION_DOCS_PER_ITERATION },
                ],
                AggregateOptions::builder()
                    .hint(Some(mongodb::options::Hint::Keys(doc! {"_id": 1})))
                    .build(),
            )
            .await?;
        first_stage = None;

        let mut no_result = true;
        while let Some(doc) = cursor.try_next().await.unwrap() {
            notify!(
                &tx_notifications,
                SamplerNotification {
                    db: db_name.clone(),
                    collection_or_view: collection.name().to_string(),
                    action: SamplerAction::Processing {
                        partition: partition_ix as u16,
                    },
                },
            );
            let id = doc.get("_id").unwrap().clone();
            partition.min = id.clone();
            let old_schema = schema.clone();
            schema = Some(schema_for_document(&doc).union(&schema.unwrap_or(Schema::Unsat)));
            if old_schema == schema {
                ignored_ids.push(id);
            }
            no_result = false;
        }
        if no_result {
            break;
        }
    }
    drop(tx_notifications);
    Ok(schema.unwrap_or(Schema::Unsat))
}

/// derive_schema_for_view takes a CollectionDoc and executes the pipeline
/// against the viewOn collection to generate a schema for the view.
/// It does this by first prepending $sample to the pipeline
#[instrument(level = "trace", skip_all)]
pub(crate) async fn derive_schema_for_view(
    view: &CollectionDoc,
    database: &Database,
    tx_notification: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
) -> Option<Schema> {
    let pipeline = vec![doc! { "$sample": { "size": VIEW_SAMPLE_SIZE } }]
        .into_iter()
        .chain(view.options.pipeline.clone().into_iter())
        .collect::<Vec<Document>>();

    let mut schema = None;

    match database
        .collection::<Document>(&view.options.view_on)
        .aggregate(pipeline, None)
        .await
    {
        Ok(mut cursor) => {
            let mut iterations = 0;
            while let Some(doc) = cursor.try_next().await.unwrap() {
                // Notify every 100 iterations, so it isn't too spammy
                if iterations % 100 == 0 {
                    notify!(
                        &tx_notification,
                        SamplerNotification {
                            db: database.name().to_string(),
                            collection_or_view: view.name.clone(),
                            action: SamplerAction::SamplingView,
                        },
                    );
                }
                if schema.is_none() {
                    schema = Some(schema_for_document(&doc));
                } else {
                    schema = Some(schema.unwrap().union(&schema_for_document(&doc)));
                }
                iterations += 1;
            }
        }
        Err(e) => {
            notify!(
                &tx_notification,
                SamplerNotification {
                    db: database.name().to_string(),
                    collection_or_view: view.name.clone(),
                    action: SamplerAction::Warning {
                        message: format!("View sampling encountered an error: {e}"),
                    },
                },
            );
        }
    }

    drop(tx_notification);
    schema
}

/// Returns a [Schema] for a given BSON document.
#[instrument(level = "trace", skip_all)]
fn schema_for_document(doc: &Document) -> Schema {
    Schema::Document(mongosql::schema::Document {
        keys: doc
            .iter()
            .map(|(k, v)| (k.to_string(), schema_for_bson(v)))
            .collect(),
        required: doc.iter().map(|(k, _)| k.to_string()).collect(),
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    })
}

#[instrument(level = "trace", skip_all)]
fn schema_for_bson(b: &Bson) -> Schema {
    use Atomic::*;
    match b {
        Bson::Double(_) => Schema::Atomic(Double),
        Bson::String(_) => Schema::Atomic(String),
        Bson::Array(a) => Schema::Array(Box::new(schema_for_bson_array_elements(a))),
        Bson::Document(d) => schema_for_document(d),
        Bson::Boolean(_) => Schema::Atomic(Boolean),
        Bson::Null => Schema::Atomic(Null),
        Bson::RegularExpression(_) => Schema::Atomic(Regex),
        Bson::JavaScriptCode(_) => Schema::Atomic(Javascript),
        Bson::JavaScriptCodeWithScope(_) => Schema::Atomic(JavascriptWithScope),
        Bson::Int32(_) => Schema::Atomic(Integer),
        Bson::Int64(_) => Schema::Atomic(Long),
        Bson::Timestamp(_) => Schema::Atomic(Timestamp),
        Bson::Binary(_) => Schema::Atomic(BinData),
        Bson::ObjectId(_) => Schema::Atomic(ObjectId),
        Bson::DateTime(_) => Schema::Atomic(Date),
        Bson::Symbol(_) => Schema::Atomic(Symbol),
        Bson::Decimal128(_) => Schema::Atomic(Decimal),
        Bson::Undefined => Schema::Atomic(Null),
        Bson::MaxKey => Schema::Atomic(MaxKey),
        Bson::MinKey => Schema::Atomic(MinKey),
        Bson::DbPointer(_) => Schema::Atomic(DbPointer),
    }
}

// This may prove costly for very large arrays, and we may want to
// consider a limit on the number of elements to consider.
#[instrument(level = "trace", skip_all)]
fn schema_for_bson_array_elements(bs: &[Bson]) -> Schema {
    // if an array is empty, we can't infer anything about it
    // we're safe to mark it as potentially null, as an empty array
    // satisfies jsonSchema search predicate
    if bs.is_empty() {
        return Schema::Atomic(Atomic::Null);
    }
    bs.iter()
        .map(schema_for_bson)
        .reduce(|acc, s| acc.union(&s))
        .unwrap_or(Schema::Any)
}
