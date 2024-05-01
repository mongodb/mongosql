use anyhow::Result;
use futures::TryStreamExt;
use mongodb::{
    bson::{self, doc, Bson, Document},
    options::{AggregateOptions, ClientOptions, ListDatabasesOptions},
    Client, Database,
};
use mongosql::{
    json_schema,
    schema::{Atomic, JaccardIndex, Schema},
    set,
};
use serde::{Deserialize, Serialize};
use std::{
    collections::HashMap,
    convert::From,
    fmt::{self, Display, Formatter},
};
use thiserror::Error;

pub fn needs_auth(options: &ClientOptions) -> bool {
    options.credential.is_none()
}

// allowing dead code since this is a library and we want to keep the code for future use
#[allow(dead_code)]
pub async fn load_password_auth(
    options: &mut ClientOptions,
    username: Option<String>,
    password: Option<String>,
) {
    options.credential = Some(
        mongodb::options::Credential::builder()
            .username(username)
            .password(password)
            .build(),
    );
}

pub async fn get_opts(uri: &str) -> Result<ClientOptions> {
    let mut opts = ClientOptions::parse_async(uri).await?;
    opts.max_pool_size = Some(get_optimal_pool_size());
    opts.max_connecting = Some(5);
    Ok(opts)
}

// allowing dead code since this is a library and we want to keep the code for future use
#[allow(dead_code)]
pub fn get_optimal_pool_size() -> u32 {
    std::thread::available_parallelism().unwrap().get() as u32 * 2 + 1
}

#[derive(Debug, Clone)]
pub enum SamplerAction {
    Querying { partition: u16 },
    Processing { partition: u16 },
    Partitioning { partitions: u16 },
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
        }
    }
}

impl Display for SamplerNotification {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(
            f,
            "{} for collection: {} in database: {}",
            self.action, self.collection, self.db
        )
    }
}

#[derive(Debug, Clone)]
pub struct SamplerNotification {
    pub db: String,
    pub collection: String,
    pub action: SamplerAction,
}

pub type SchemaAnalysis = (String, Vec<HashMap<String, Schema>>);

const DISALLOWED_DB_NAMES: [&str; 4] = ["admin", "config", "local", "system"];
const DISALLOWED_COLLECTION_NAMES: [&str; 6] = [
    "system.buckets",
    "system.namespaces",
    "system.indexes",
    "system.profile",
    "system.js",
    "system.views",
];

pub async fn sample(
    options: ClientOptions,
    tx_notification: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
    tx_schemata: tokio::sync::mpsc::UnboundedSender<Result<SchemaAnalysis>>,
    tx_errors: tokio::sync::mpsc::UnboundedSender<anyhow::Error>,
) {
    let client = Client::with_options(options).unwrap();
    // by MongoDB and not user-created databases
    let databases = client
        .list_database_names(None, Some(ListDatabasesOptions::builder().build()))
        .await;
    if let Err(e) = databases {
        tx_errors.send(e.into()).unwrap();
        drop(tx_errors);
        drop(tx_notification);
        drop(tx_schemata);
        return;
    } else {
        drop(tx_errors);
        for database in databases.unwrap() {
            if DISALLOWED_DB_NAMES.contains(&database.as_str()) {
                continue;
            }
            let mut db_schemata: Vec<HashMap<String, Schema>> = Vec::new();
            let db = client.database(&database);
            for collection in client
                .database(&database)
                .run_command(
                    doc! { "listCollections": 1.0, "authorizedCollections": true, "nameOnly": true},
                    None,
                )
                .await
                .map_err(|e| async {
                    tx_schemata.send(Err(e.into())).unwrap();
                })
                .unwrap_or_default()
                .get_document("cursor")
                .map(|c| {
                    c.get_array("firstBatch")
                        .unwrap()
                        .iter()
                        .filter_map(|d| {
                            let d = d.as_document().unwrap();
                            if d.get_str("type").unwrap() == "view" {
                                None
                            } else {
                                Some(d.get_str("name").unwrap().to_string())
                            }
                        })
                        .collect::<Vec<String>>()
                })
                .unwrap_or_default()
            {
                if DISALLOWED_COLLECTION_NAMES.contains(&collection.as_str()) {
                    continue;
                }
                let notifier = tx_notification.clone();
                let col_parts = gen_partitions(&db, &collection, notifier).await;
                let notifier = tx_notification.clone();
                let schemata = derive_schema_for_partitions(col_parts, &db, notifier).await;
                db_schemata.push(schemata);
            }
            if tx_schemata.send(Ok((database, db_schemata))).is_err() {
                break;
            }
        }
    }
    drop(tx_notification);
    drop(tx_schemata);
}

const PARTITION_SIZE_IN_BYTES: i64 = 100 * 1024 * 1024; // 100 MB
const SAMPLE_MIN_DOCS: i64 = 101;
const SAMPLE_RATE: f64 = 0.04;
const MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION: u64 = 10;
const ITERATIONS: Option<u32> = None;

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

#[derive(Debug, Error)]
pub enum Error {
    #[error("JsonSchemaFailure")]
    JsonSchemaFailure,
    #[error("BsonFailure")]
    BsonFailure,
    #[error("NoCollectionStats")]
    NoCollectionStats,
    #[error("NoBounds")]
    NoBounds,
    #[error("NoIdInSample")]
    NoIdInSample,
    #[error("Driver Error {0}")]
    DriverError(mongodb::error::Error),
    #[error("NoCollection {0}")]
    NoCollection(String),
}

impl From<mongodb::error::Error> for Error {
    fn from(value: mongodb::error::Error) -> Self {
        Self::DriverError(value)
    }
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an optional Schema. If the Schema is None, the $match will only
// be based on the Partition bounds.
pub fn generate_partition_match(
    partition: &Partition,
    schema: Option<Schema>,
) -> Result<Document, Error> {
    generate_partition_match_with_doc(partition, schema.map(schema_to_schema_doc).transpose()?)
}

// generate_partition_match generates the $match stage for sampling based on the partition
// additional_properties and an input jsonSchema.
pub fn generate_partition_match_with_doc(
    partition: &Partition,
    schema: Option<Document>,
) -> Result<Document, Error> {
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

pub fn schema_to_schema_doc(schema: Schema) -> Result<Document, Error> {
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

pub fn schema_doc_to_schema(schema_doc: Document) -> Result<Schema, Error> {
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

pub async fn get_size_counts(database: &Database, coll: &str) -> Result<CollectionSizes, Error> {
    let collection = database.collection::<Document>(coll);

    let mut cursor = match collection
        .aggregate(vec![doc! {"$collStats": {"storageStats": {}}}], None)
        .await
    {
        Ok(c) => c,
        Err(e) => {
            panic!("{e}")
        }
    };
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
    Err(Error::NoCollectionStats)
}

pub async fn get_bound(database: &Database, coll: &str, direction: i32) -> Result<Bson, Error> {
    let collection = database.collection::<Document>(coll);

    let mut cursor = match collection
        .aggregate(
            vec![
                doc! {"$sort": {"_id": direction}},
                doc! {"$project": {"_id": 1}},
                doc! {"$limit": 1},
            ],
            None,
        )
        .await
    {
        Ok(c) => c,
        Err(e) => {
            panic!("{e}")
        }
    };
    if let Some(doc) = cursor.try_next().await? {
        return doc.get("_id").cloned().ok_or(Error::NoBounds);
    }
    Err(Error::NoBounds)
}

pub async fn get_bounds(database: &Database, coll: &str) -> Result<(Bson, Bson), Error> {
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
    col_parts: HashMap<String, Vec<Partition>>,
    database: &Database,
    tx_notification: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
) -> HashMap<String, Schema> {
    let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
    rayon::scope(|s| {
        for (coll, parts) in col_parts {
            for (ix, part) in parts.into_iter().enumerate() {
                let coll = coll.clone();
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
                            &database, &coll, part, ITERATIONS, None, notifier, ix,
                        )
                        .await
                        .unwrap();
                        tx.send((coll, schema)).unwrap();
                        drop(tx)
                    });
                    drop(rt);
                })
            }
        }
    });
    drop(tx);
    drop(tx_notification);
    let mut schemata = HashMap::new();
    while let Some((coll, schema)) = rx.recv().await {
        if let Some(prev_schema) = schemata.get(&coll) {
            schemata.insert(coll, schema.union(prev_schema));
        } else {
            schemata.insert(coll, schema);
        }
    }
    schemata
}

pub async fn gen_partitions(
    database: &Database,
    coll: &str,
    tx_notification: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
) -> HashMap<String, Vec<Partition>> {
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
                        if notifier
                            .send(SamplerNotification {
                                db: database.name().to_string(),
                                collection: coll.to_string(),
                                action: SamplerAction::Partitioning {
                                    partitions: partitions.len() as u16,
                                },
                            })
                            .is_err()
                        {
                            return;
                        }
                        tx.send((coll, partitions)).unwrap();
                    }
                    Err(_) => {
                        // nothing, we couldn't get bounds for this collection. It is
                        // highly likely that the collection is empty.
                    }
                };
                drop(tx);
                drop(notifier)
            });
            drop(rt);
        })
    });
    drop(tx);
    let mut col_parts = HashMap::new();
    while let Some((coll, partitions)) = rx.recv().await {
        col_parts.insert(coll.to_string(), partitions);
    }
    col_parts
}

pub async fn gen_partitions_with_initial_schema(
    collections_and_schemata: Vec<(String, Document)>,
    database: &Database,
) -> HashMap<String, (Document, Vec<Partition>)> {
    println!("getting partitions with schemata");
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

pub async fn get_partitions(database: &Database, coll: &str) -> Result<Vec<Partition>, Error> {
    let size_info = get_size_counts(database, coll).await?;
    let num_partitions = get_num_partitions(size_info.size, PARTITION_SIZE_IN_BYTES) as usize;
    let (mut min_bound, max_bound) = get_bounds(database, coll).await?;
    let mut partitions = Vec::with_capacity(num_partitions);

    let collection = database.collection::<Document>(coll);

    let first_stage = if size_info.count >= SAMPLE_MIN_DOCS && num_partitions > 1 {
        let num_docs_to_sample = std::cmp::min(
            (SAMPLE_RATE * size_info.count as f64) as u64,
            MAX_NUM_DOCS_TO_SAMPLE_PER_PARTITION * num_partitions as u64,
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
    iterations: Option<u32>,
    initial_schema_doc: Option<Document>,
    tx_notification: tokio::sync::mpsc::UnboundedSender<SamplerNotification>,
    partition_ix: usize,
) -> Result<Schema, Error> {
    let collection = database.collection::<Document>(coll);
    let mut schema: Option<Schema> = initial_schema_doc
        .clone()
        .map(schema_doc_to_schema)
        .transpose()?;
    let mut iteration = 0;
    let mut first_stage = Some(generate_partition_match_with_doc(
        &partition,
        initial_schema_doc,
    )?);

    loop {
        if iterations.is_some() && iteration >= iterations.unwrap() {
            break;
        }
        if partition.min == partition.max {
            break;
        }

        if first_stage.is_none() {
            first_stage = Some(generate_partition_match(&partition, schema.clone())?);
        };
        if tx_notification
            .send(SamplerNotification {
                db: database.name().to_string(),
                collection: coll.to_string(),
                action: SamplerAction::Querying {
                    partition: partition_ix as u16,
                },
            })
            .is_err()
        {
            break;
        }
        let mut cursor = collection
            .aggregate(
                vec![
                    // first_stage must be Some here.
                    first_stage.unwrap(),
                    doc! { "$sort": {"_id": 1}},
                    doc! { "$limit": 1 },
                ],
                AggregateOptions::builder()
                    .hint(Some(mongodb::options::Hint::Keys(doc! {"_id": 1})))
                    .build(),
            )
            .await?;
        first_stage = None;

        let mut no_result = true;
        while let Some(doc) = cursor.try_next().await.unwrap() {
            if tx_notification
                .send(SamplerNotification {
                    db: database.name().to_string(),
                    collection: coll.to_string(),
                    action: SamplerAction::Processing {
                        partition: partition_ix as u16,
                    },
                })
                .is_err()
            {
                break;
            }
            let id = doc.get("_id").unwrap().clone();
            partition.min = id;
            schema = Some(schema_for_document(&doc).union(&schema.unwrap_or(Schema::Unsat)));
            no_result = false;
        }
        if no_result {
            break;
        }
        iteration += 1;
    }
    drop(tx_notification);
    Ok(schema.unwrap_or(Schema::Unsat))
}
