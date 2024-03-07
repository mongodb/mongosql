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
};
use serde::{Deserialize, Serialize};
use std::{collections::HashMap, convert::From};
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

pub async fn sample(
    options: ClientOptions,
) -> Result<HashMap<String, Vec<HashMap<String, Schema>>>> {
    let mut full_schemata: HashMap<String, Vec<HashMap<String, Schema>>> = HashMap::new();
    let client = Client::with_options(options).unwrap();
    // we'll filter out the admin, config, local, and system databases as these are used
    // by MongoDB and not user-created databases
    for database in client
        .list_database_names(
            doc! { "name":  {"$nin": ["admin", "config", "local", "system", ""]}},
            Some(
                ListDatabasesOptions::builder()
                    .authorized_databases(true)
                    .build(),
            ),
        )
        .await?
    {
        let db = client.database(&database);
        for collection in client
            .database(&database)
            .run_command(
                doc! { "listCollections": 1.0, "authorizedCollections": true, "nameOnly": true},
                None,
            )
            .await?
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
            let col_parts = gen_partitions(&db, &collection).await;
            let schemata = derive_schema_for_partitions(col_parts, &db).await;
            full_schemata
                .entry(database.clone())
                .or_default()
                .push(schemata);
        }
    }
    Ok(full_schemata)
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
) -> HashMap<String, Schema> {
    let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
    rayon::scope(|s| {
        for (coll, parts) in col_parts {
            for part in parts {
                let coll = coll.clone();
                let database = database.clone();
                let tx = tx.clone();
                let mut rt = tokio::runtime::Runtime::new();
                while rt.is_err() {
                    rt = tokio::runtime::Runtime::new();
                }
                let rt = rt.unwrap();
                s.spawn(move |_| {
                    rt.block_on(async move {
                        let part = part.clone();
                        let schema =
                            derive_schema_for_partition(&database, &coll, part, ITERATIONS, None)
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

pub async fn gen_partitions(database: &Database, coll: &str) -> HashMap<String, Vec<Partition>> {
    let (tx, mut rx) = tokio::sync::mpsc::unbounded_channel();
    rayon::scope(|s| {
        let tx = tx.clone();
        let rt = tokio::runtime::Runtime::new().unwrap();
        let database = database.clone();
        s.spawn(move |_| {
            rt.block_on(async move {
                match get_partitions(&database, coll).await {
                    Ok(partitions) => {
                        tx.send((coll, partitions)).unwrap();
                    }
                    Err(_) => {
                        // nothing, we couldn't get bounds for this collection. It is
                        // highly likely that the collection is empty.
                    }
                };
                drop(tx)
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

    Ok(schema.unwrap_or(Schema::Unsat))
}
