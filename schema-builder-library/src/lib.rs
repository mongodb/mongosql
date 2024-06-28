use futures::TryStreamExt;
use mongodb::{
    bson::{self, doc, Bson, Document},
    options::{AggregateOptions, ListDatabasesOptions},
    Client, Cursor, Database,
};
use mongosql::{
    json_schema,
    schema::{Atomic, JaccardIndex, Schema},
    set,
};
use serde::{Deserialize, Serialize};
use std::{
    collections::HashMap,
    fmt::{self, Display, Formatter},
};
mod errors;
pub use errors::Error;
pub mod client_util;
mod consts;
use consts::{
    DISALLOWED_COLLECTION_NAMES, DISALLOWED_DB_NAMES, MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION,
    PARTITION_SIZE_IN_BYTES, SAMPLE_MIN_DOCS, SAMPLE_RATE, SAMPLE_SIZE,
};

use self::options::BuilderOptions;
pub mod options;

pub type Result<T> = std::result::Result<T, Error>;

// A struct representing schema information for a specific namespace (a view
// or collection).
pub struct SchemaResult {
    // The name of the database.
    pub db_name: String,

    // The name of the collection or view which this schema represents.
    pub coll_or_view_name: String,

    // The type of namespace (collection or view)
    pub namespace_type: NamespaceType,

    // The schema for the namespace.
    pub namespace_schema: Option<Schema>,
}

pub enum NamespaceType {
    Collection,
    View,
}

#[derive(Debug, Clone)]
pub enum SamplerAction {
    Querying { partition: u16 },
    Processing { partition: u16 },
    Partitioning { partitions: u16 },
    Error { message: String },
    SamplingView,
}

macro_rules! notify {
    ($channel:expr, $notification:expr) => {
        if let Some(ref notifier) = $channel {
            // notification errors are not critical, so we just ignore them
            let _ = notifier.send($notification);
        }
    };
}

pub async fn build_schema<'a>(options: options::BuilderOptions<'_>) {
    let databases = options
        .client
        .list_database_names(None, Some(ListDatabasesOptions::builder().build()))
        .await;
    if let Err(e) = databases {
        notify!(
            options.tx_notifications,
            SamplerNotification {
                db: "".to_string(),
                collection_or_view: "".to_string(),
                action: SamplerAction::Error {
                    message: e.to_string(),
                },
            }
        );
        drop(options.tx_notifications);
        drop(options.tx_schemata);
        return;
    } else {
        for database in databases.unwrap() {
            if DISALLOWED_DB_NAMES.contains(&database.as_str()) {
                continue;
            }
            let db = options.client.database(&database);
            let collection_info =
                list_collections_and_views(options.client, &database, options.tx_schemata.clone())
                    .await;
            collection_info.process_collections(&db, &options).await;
            collection_info.process_views(&db, &options).await;
        }
    }
    drop(options.tx_notifications);
    drop(options.tx_schemata);
}

#[derive(Debug, Default)]
struct CollectionInfo {
    views: Vec<CollectionDoc>,
    collections: Vec<CollectionDoc>,
}

impl CollectionInfo {
    async fn process_collections(&self, db: &Database, options: &BuilderOptions<'_>) {
        for collection in self.collections.as_slice() {
            if DISALLOWED_COLLECTION_NAMES.contains(&collection.name.as_str()) {
                continue;
            }
            let col_parts =
                gen_partitions(db, &collection.name, options.tx_notifications.clone()).await;
            let schemata = derive_schema_for_partitions(
                &collection.name,
                col_parts,
                db,
                options.tx_notifications.clone(),
            )
            .await;
            let tx_schemata = options.tx_schemata.clone();
            if tx_schemata
                .send(Ok(SchemaResult {
                    db_name: db.name().to_string(),
                    coll_or_view_name: collection.name.clone(),
                    namespace_type: NamespaceType::Collection,
                    namespace_schema: schemata,
                }))
                .is_err()
            {
                drop(tx_schemata);
                return;
            }
            drop(tx_schemata);
        }
    }
    async fn process_views(&self, db: &Database, options: &BuilderOptions<'_>) {
        for view in self.views.as_slice() {
            let schema = derive_schema_for_view(view, db, options.tx_notifications.clone()).await;
            let tx_schemata = options.tx_schemata.clone();
            if tx_schemata
                .send(Ok(SchemaResult {
                    db_name: db.name().to_string(),
                    coll_or_view_name: view.name.clone(),
                    namespace_type: NamespaceType::View,
                    namespace_schema: schema,
                }))
                .is_err()
            {
                drop(tx_schemata);
                return;
            }
            drop(tx_schemata);
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct CollectionDoc {
    #[serde(rename = "type")]
    type_: String,
    name: String,
    options: ViewOptions,
}

#[derive(Debug, Serialize, Deserialize)]
struct ViewOptions {
    #[serde(rename = "viewOn", default)]
    view_on: String,
    #[serde(default)]
    pipeline: Vec<Document>,
}

impl Default for ViewOptions {
    fn default() -> Self {
        Self {
            view_on: "".to_string(),
            pipeline: vec![],
        }
    }
}

async fn list_collections_and_views(
    client: &Client,
    database: &str,
    tx_schemata: tokio::sync::mpsc::UnboundedSender<Result<SchemaResult>>,
) -> CollectionInfo {
    let collection_info_cursor = client
        .database(database)
        .run_cursor_command(
            doc! { "listCollections": 1.0, "authorizedCollections": true},
            None,
        )
        .await
        .map_err(|e| async {
            tx_schemata.send(Err(e.into())).unwrap();
        });

    let collection_info = match collection_info_cursor {
        Ok(collection_info) => separate_views_from_collections(collection_info)
            .await
            .unwrap(),
        Err(_) => CollectionInfo::default(),
    };

    drop(tx_schemata);
    collection_info
}

async fn separate_views_from_collections(
    mut collection_doc: Cursor<Document>,
) -> Result<CollectionInfo> {
    let mut collection_info = CollectionInfo::default();
    while let Some(collection_doc) = collection_doc.try_next().await.unwrap() {
        let collection_doc: CollectionDoc =
            bson::from_bson(bson::Bson::Document(collection_doc)).unwrap();
        if collection_doc.type_ == "view" {
            collection_info.views.push(collection_doc);
        } else {
            collection_info.collections.push(collection_doc);
        }
    }

    Ok(collection_info)
}

/// derive_schema_for_view takes a CollectionDoc and executes the pipeline
/// against the viewOn collection to generate a schema for the view.
/// It does this by first prepending $sample to the pipeline
async fn derive_schema_for_view(
    view: &CollectionDoc,
    database: &Database,
    tx_notification: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
) -> Option<Schema> {
    let pipeline = vec![doc! { "$sample": { "size": SAMPLE_SIZE } }]
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
                // we want to notify every 100 iterations so it isn't too spammy
                if iterations % 100 == 0 {
                    notify!(
                        tx_notification,
                        SamplerNotification {
                            db: database.name().to_string(),
                            collection_or_view: view.name.clone(),
                            action: SamplerAction::SamplingView,
                        }
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
                tx_notification,
                SamplerNotification {
                    db: database.name().to_string(),
                    collection_or_view: view.name.clone(),
                    action: SamplerAction::Error {
                        message: e.to_string(),
                    },
                }
            );
        }
    }

    drop(tx_notification);
    schema
}

#[derive(Debug, Clone)]
pub struct SamplerNotification {
    pub db: String,
    pub collection_or_view: String,
    pub action: SamplerAction,
}

impl Display for SamplerAction {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        match self {
            SamplerAction::Querying { partition } => write!(f, "Querying partition {}", partition),
            SamplerAction::Processing { partition } => {
                write!(f, "Processing partition {}", partition)
            }
            SamplerAction::Partitioning { partitions } => {
                write!(f, "Partitioning into {} parts", partitions)
            }
            SamplerAction::Error { message } => write!(f, "Error: {}", message),
            SamplerAction::SamplingView => write!(f, "Sampling"),
        }
    }
}

impl Display for SamplerNotification {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(
            f,
            "{} for collection/view: {} in database: {}",
            self.action, self.collection_or_view, self.db
        )
    }
}

/// Returns a [Schema] for a given BSON document.
pub fn schema_for_document(doc: &bson::Document) -> Schema {
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

fn schema_for_bson(b: &Bson) -> Schema {
    use Atomic::*;
    match b {
        Bson::Double(_) => Schema::Atomic(Double),
        Bson::String(_) => Schema::Atomic(String),
        Bson::Array(a) => Schema::Array(Box::new(schema_for_bson_array(a))),
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
fn schema_for_bson_array(bs: &[Bson]) -> Schema {
    // if an array is empty, we can't infer anything about it
    // we're safe to mark it as potentially null, as an empty array
    // satisfies jsonSchema search predicate
    if bs.is_empty() {
        return Schema::AnyOf(set!(Schema::Atomic(Atomic::Null)));
    }
    bs.iter()
        .map(schema_for_bson)
        .reduce(|acc, s| acc.union(&s))
        .unwrap_or(Schema::Any)
}

#[derive(Debug, PartialEq, Clone)]
pub struct Partition {
    pub min: Bson,
    pub max: Bson,
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an optional Schema. If the Schema is None, the $match will only
// be based on the Partition bounds.
pub fn generate_partition_match(partition: &Partition, schema: Option<Schema>) -> Result<Document> {
    generate_partition_match_with_doc(partition, schema.map(schema_to_schema_doc).transpose()?)
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an input jsonSchema.
pub fn generate_partition_match_with_doc(
    partition: &Partition,
    schema: Option<Document>,
) -> Result<Document> {
    let mut match_body = doc! {
        "_id": {
            "$gte": partition.min.clone(),
            "$lt": partition.max.clone(),
        }
    };
    if let Some(schema) = schema {
        match_body.insert("$nor", vec![schema]);
    }
    Ok(doc! {
        "$match": match_body
    })
}

pub fn schema_to_schema_doc(schema: Schema) -> Result<Document> {
    let json_schema: json_schema::Schema = schema
        .clone()
        .try_into()
        .map_err(|_| Error::JsonSchemaFailure)?;
    let bson_schema = bson::to_bson(&json_schema).map_err(|_| Error::BsonFailure)?;
    let ret = doc! {
        "$jsonSchema": bson_schema
    };
    Ok(ret)
}

pub fn schema_doc_to_schema(schema_doc: Document) -> Result<Schema> {
    let json_schema: json_schema::Schema =
        bson::from_document(schema_doc.get_document("$jsonSchema").unwrap().clone())
            .map_err(|_| Error::BsonFailure)?;
    let sampler_schema: Schema = json_schema
        .try_into()
        .map_err(|_| Error::JsonSchemaFailure)?;
    Ok(sampler_schema)
}

pub struct CollectionSizes {
    pub size: i64,
    pub count: i64,
}

pub fn get_num_partitions(coll_size: i64, partition_size: i64) -> i64 {
    let num_parts = (coll_size as f64) / (partition_size as f64);
    num_parts as i64 + 1
}

pub async fn get_size_counts(database: &Database, coll: &str) -> Result<CollectionSizes> {
    let collection = database.collection::<Document>(coll);

    let mut cursor = collection
        .aggregate(vec![doc! {"$collStats": {"storageStats": {}}}], None)
        .await
        .map_err(|_| Error::NoCollectionStats(coll.to_string()))?;
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
        return Ok(CollectionSizes { size, count });
    }
    Err(Error::NoCollectionStats(coll.to_string()))
}

pub async fn get_bound(database: &Database, coll: &str, direction: i32) -> Result<Bson> {
    let collection = database.collection::<Document>(coll);

    let mut cursor = collection
        .aggregate(
            vec![
                doc! {"$sort": {"_id": direction}},
                doc! {"$limit": 1},
                doc! {"$project": {"_id": 1}},
            ],
            None,
        )
        .await
        .map_err(|e| Error::NoBounds(format!("{coll}: {e}",)))?;
    if let Some(doc) = cursor.try_next().await? {
        return doc
            .get("_id")
            .cloned()
            .ok_or(Error::NoBounds(coll.to_string()));
    }
    Err(Error::NoBounds(coll.to_string()))
}

pub async fn get_bounds(database: &Database, coll: &str) -> Result<(Bson, Bson)> {
    Ok((
        get_bound(database, coll, 1).await?,
        // we actually will just always use MaxKey as our upper bound since we
        // match $lt max bound
        Bson::MaxKey,
    ))
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Schemata {
    #[serde(rename = "_id")]
    pub id: SchemataId,
    pub schema: Document,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct SchemataId {
    pub db: String,
    pub collection: String,
}

pub async fn derive_schema_for_partitions(
    collection: &str,
    col_parts: Vec<Partition>,
    database: &Database,
    tx_notification: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
) -> Option<Schema> {
    let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
    rayon::scope(|s| {
        for (ix, part) in col_parts.into_iter().enumerate() {
            let database = database.clone();
            let tx = tx.clone();
            let notifier = tx_notification.clone();
            let mut rt = tokio::runtime::Runtime::new();
            while rt.is_err() {
                rt = tokio::runtime::Runtime::new();
            }
            let rt = rt.unwrap();
            s.spawn(move |_| {
                rt.block_on(async move {
                    let part = part.clone();
                    let schema = derive_schema_for_partition(
                        &database, collection, part, None, notifier, ix,
                    )
                    .await
                    .unwrap();
                    tx.send(schema).unwrap();
                    drop(tx)
                });
                drop(rt);
            })
        }
    });
    drop(tx);
    drop(tx_notification);
    let mut schemata = None;
    while let Some(schema) = rx.recv().await {
        if let Some(prev_schema) = schemata {
            schemata = Some(schema.union(&prev_schema))
        } else {
            schemata = Some(schema)
        }
    }
    schemata
}

pub async fn gen_partitions(
    database: &Database,
    coll: &str,
    tx_notification: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
) -> Vec<Partition> {
    let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
    rayon::scope(|s| {
        let tx = tx.clone();
        let notifier = tx_notification.clone();
        let rt = tokio::runtime::Runtime::new().unwrap();
        let database = database.clone();
        s.spawn(move |_| {
            rt.block_on(async move {
                match get_partitions(&database, coll).await {
                    Ok(partitions) => {
                        notify!(
                            notifier,
                            SamplerNotification {
                                db: database.name().to_string(),
                                collection_or_view: coll.to_string(),
                                action: SamplerAction::Partitioning {
                                    partitions: partitions.len() as u16,
                                },
                            }
                        );
                        tx.send(partitions).unwrap();
                    }
                    Err(e) => {
                        notify!(
                            notifier,
                            SamplerNotification {
                                db: database.name().to_string(),
                                collection_or_view: coll.to_string(),
                                action: SamplerAction::Error {
                                    message: e.to_string(),
                                },
                            }
                        )
                    }
                };
                drop(tx);
                drop(notifier)
            });
            drop(rt);
        })
    });
    drop(tx);
    rx.recv().await.unwrap_or_default()
}

pub async fn gen_partitions_with_initial_schema(
    collections_and_schemata: Vec<(String, Document)>,
    database: &Database,
) -> HashMap<String, (Document, Vec<Partition>)> {
    let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
    rayon::scope(|s| {
        for (coll, sch) in collections_and_schemata {
            let coll = coll.clone();
            let tx = tx.clone();
            let rt = tokio::runtime::Runtime::new().unwrap();
            let database = database.clone();
            s.spawn(move |_| {
                rt.block_on(async move {
                    let partitions = get_partitions(&database, &coll).await.unwrap();
                    tx.send(((coll, sch), partitions)).unwrap();
                    drop(tx)
                });
                drop(rt);
            })
        }
    });
    drop(tx);
    let mut col_parts = HashMap::new();
    while let Some(((coll, sch), partitions)) = rx.recv().await {
        col_parts.insert(coll, (sch, partitions));
    }
    col_parts
}

pub async fn get_partitions(database: &Database, coll: &str) -> Result<Vec<Partition>> {
    let size_info = get_size_counts(database, coll).await?;
    let num_partitions = get_num_partitions(size_info.size, PARTITION_SIZE_IN_BYTES) as usize;
    let (mut min_bound, max_bound) = get_bounds(database, coll).await?;
    let mut partitions = Vec::with_capacity(num_partitions);

    let collection = database.collection::<Document>(coll);

    let first_stage = if size_info.count >= SAMPLE_MIN_DOCS && num_partitions > 1 {
        let num_docs_to_sample = std::cmp::min(
            (SAMPLE_RATE * size_info.count as f64) as u64,
            MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION as u64 * num_partitions as u64,
        );
        doc! { "$sample": {"size": num_docs_to_sample as i64 } }
    } else {
        doc! { "$match": {} }
    };

    let mut cursor = collection
        .aggregate(
            vec![
                first_stage,
                doc! { "$project": {"_id": 1} },
                doc! { "$bucketAuto": {
                    "groupBy": "$_id",
                    "buckets": num_partitions as i32
                }},
            ],
            None,
        )
        .await?;

    while let Some(doc) = cursor.try_next().await.unwrap() {
        let doc = doc.get("_id").unwrap().as_document().unwrap();
        let max = doc.get("max").cloned().unwrap_or(Bson::MaxKey);
        partitions.push(Partition {
            min: min_bound.clone(),
            max: max.clone(),
        });
        if max != max_bound {
            min_bound = max;
        }
    }
    partitions.push(Partition {
        min: min_bound,
        max: max_bound,
    });

    Ok(partitions)
}

pub async fn derive_schema_for_partition(
    database: &Database,
    coll: &str,
    mut partition: Partition,
    initial_schema_doc: Option<Document>,
    tx_notification: Option<tokio::sync::mpsc::UnboundedSender<SamplerNotification>>,
    partition_ix: usize,
) -> Result<Schema> {
    let collection = database.collection::<Document>(coll);
    let mut schema: Option<Schema> = initial_schema_doc
        .clone()
        .map(schema_doc_to_schema)
        .transpose()?;
    let mut first_stage = Some(generate_partition_match_with_doc(
        &partition,
        initial_schema_doc,
    )?);

    loop {
        if partition.min == partition.max {
            break;
        }

        if first_stage.is_none() {
            first_stage = Some(generate_partition_match(&partition, schema.clone())?);
        };
        notify!(
            tx_notification,
            SamplerNotification {
                db: database.name().to_string(),
                collection_or_view: coll.to_string(),
                action: SamplerAction::Querying {
                    partition: partition_ix as u16,
                },
            }
        );

        let mut cursor = collection
            .aggregate(
                vec![
                    // first_stage must be Some here.
                    first_stage.unwrap(),
                    doc! { "$sort": {"_id": 1}},
                    doc! { "$limit": MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION },
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
                tx_notification,
                SamplerNotification {
                    db: database.name().to_string(),
                    collection_or_view: coll.to_string(),
                    action: SamplerAction::Processing {
                        partition: partition_ix as u16,
                    },
                }
            );

            let id = doc.get("_id").unwrap().clone();
            partition.min = id;
            schema = Some(schema_for_document(&doc).union(&schema.unwrap_or(Schema::Unsat)));
            no_result = false;
        }
        if no_result {
            break;
        }
    }
    drop(tx_notification);
    Ok(schema.unwrap_or(Schema::Unsat))
}
