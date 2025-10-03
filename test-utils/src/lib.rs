use itertools::Itertools;
use lazy_static::lazy_static;
use mongodb::{
    bson::{doc, Bson, Document},
    sync::Client,
};
use mongosql::{build_catalog_from_catalog_schema, catalog::Catalog, json_schema, map};
use std::{
    collections::BTreeMap,
    env,
    io::{self},
    string::ToString,
};
use thiserror::Error;

lazy_static! {
    pub static ref MONGODB_URI: String = format!(
        "mongodb://localhost:{}",
        env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
    );
}
pub mod index;
pub use index::*;
pub mod query;
pub use query::*;
pub mod schema_derivation;
pub use schema_derivation::*;
pub mod e2e_db_manager;

#[derive(Debug, Error)]
pub enum Error {
    #[error("failed to read directory: {0:?}")]
    InvalidDirectory(io::Error),
    #[error("failed to load file paths: {0:?}")]
    InvalidFilePath(io::Error),
    #[error("failed to read file: {0:?}")]
    InvalidFile(io::Error),
    #[error("unable to read file to string: {0:?}")]
    CannotReadFileToString(io::Error),
    #[error("unable to deserialize YAML file: {0:?}")]
    CannotDeserializeYaml((String, serde_yaml::Error)),
    #[error("failed to create mongodb client: {0:?}")]
    CannotCreateMongoDBClient(mongodb::error::Error),
    #[error("failed to drop db '{0}': {1:?}")]
    MongoDBDrop(String, mongodb::error::Error),
    #[error("failed to insert into '{0}.{1}': {2:?}")]
    MongoDBInsert(String, String, mongodb::error::Error),
    #[error("failed to create indexes for '{0}.{1}': {2:?}")]
    MongoDBCreateIndexes(String, String, mongodb::error::Error),
    #[error("failed to convert schema to MongoSql model: {0:?}")]
    InvalidSchema(mongosql::schema::Error),
    #[error("{0}")]
    UnsupportedBsonType(mongosql::schema::Error),
    #[error("failed to translate query: {0}")]
    Translation(mongosql::result::Error),
    #[error("failed to run aggregation: {0:?}")]
    MongoDBAggregation(mongodb::error::Error),
    #[error("failed to deserialize ExplainResult: {0:?}")]
    ExplainDeserialization(mongodb::bson::de::Error),
    #[error("invalid root stage: {0}")]
    InvalidRootStage(String),
    #[error("no queryPlanner found: {0:?}")]
    MissingQueryPlanner(Box<ExplainResult>),
    #[error("general mongodb error: {0:?}")]
    MongoDBErr(mongodb::error::Error),
}

impl From<mongosql::schema::Error> for Error {
    fn from(e: mongosql::schema::Error) -> Self {
        match e {
            mongosql::schema::Error::UnsupportedBsonType(_) => Error::UnsupportedBsonType(e),
            _ => Error::InvalidSchema(e),
        }
    }
}

const SQL_SCHEMAS_COLLECTION: &str = "__sql_schemas";

// helper function for get_catalog_for_dbs; split this way because schema derivation
// tests need the underlying schema, not the catalog.
pub fn get_schema_map_for_dbs(
    client: &Client,
    db_names: Vec<String>,
) -> BTreeMap<String, BTreeMap<String, json_schema::Schema>> {
    let mut catalog_schema: BTreeMap<String, BTreeMap<String, json_schema::Schema>> = map! {};

    for db_name in db_names {
        let mut db_catalog_schema: BTreeMap<String, json_schema::Schema> = map! {};
        let db = client.database(&db_name);
        let schema_collection = db.collection::<Document>(SQL_SCHEMAS_COLLECTION);

        let schema_pipeline = vec![doc! {"$project": {
            "coll_name": "$_id",
            "schema": 1,
        }}];

        let schema_docs: Vec<Document> = schema_collection
            .aggregate(schema_pipeline)
            .run()
            .expect("failed to run schema aggregation pipeline")
            .collect::<Result<Vec<Document>, _>>()
            .expect("failed to collect aggregation pipeline results");

        for doc in schema_docs {
            let coll_name = doc
                .get_str("coll_name")
                .expect("expected coll_name but it was missing");
            let schema = doc
                .get_document("schema")
                .expect("expected schema but it was missing");
            let schema_json = json_schema::Schema::from_document(schema)
                .expect("expected schema to deserialize into json_schema::Schema but it failed");

            db_catalog_schema.insert(coll_name.to_string(), schema_json);
        }

        catalog_schema.insert(db_name, db_catalog_schema);
    }
    catalog_schema
}

/// get_catalog_for_dbs builds a Catalog from the schemas stored in mongod for
/// the provided list of databases.
pub fn get_catalog_for_dbs(client: &Client, db_names: Vec<String>) -> Catalog {
    let catalog_schema = get_schema_map_for_dbs(client, db_names);
    build_catalog_from_catalog_schema(catalog_schema).expect("failed to build catalog")
}

/// load_catalog_data drops any existing catalog data and then inserts the
/// provided catalog data into the mongodb instance.
pub fn load_catalog_data(
    client: &Client,
    catalog_data: BTreeMap<String, BTreeMap<String, Vec<Bson>>>,
) -> Result<(), Error> {
    let catalog_dbs = catalog_data.keys().collect_vec();
    drop_catalog_data(client, catalog_dbs)?;

    for (db, coll_data) in catalog_data {
        let client_db = client.database(db.as_str());

        for (coll, documents) in coll_data {
            let client_coll = client_db.collection::<Bson>(coll.as_str());
            client_coll
                .insert_many(documents)
                .run()
                .map_err(|e| Error::MongoDBInsert(db.clone(), coll, e))?;
        }
    }

    Ok(())
}

/// drop_catalog_data drops all dbs in the provided list.
pub fn drop_catalog_data<T: Into<String>>(
    client: &Client,
    catalog_dbs: Vec<T>,
) -> Result<(), Error> {
    for db in catalog_dbs {
        let db = db.into();
        client
            .database(&db)
            .drop()
            .run()
            .map_err(|e| Error::MongoDBDrop(db.clone(), e))?;
    }
    Ok(())
}

/// check_server_version_for_test verifies that the server version is valid if the test specifies min or max server versions.
fn check_server_version_for_test(
    min_server_version: &Option<String>,
    max_server_version: &Option<String>,
    server_version: &Option<String>,
) -> bool {
    // max server version implies there is some version above which the test should not run; hence
    // if the server version is "latest" we skip the test.
    if server_version == &Some("latest".to_string()) {
        return !max_server_version.is_some();
    }
    // rapid can be a moving target, so for now, we will skip tests that have version constraints on rapid
    if server_version == &Some("rapid".to_string()) {
        return max_server_version.is_none() && min_server_version.is_none();
    }

    // convert versions from Option<String> to Option<f32> for comparison
    let min_server_version = min_server_version.as_ref().map(|v| {
        v.parse::<f32>()
            .expect("expected min server version to be numeric")
    });
    let max_server_version = max_server_version.as_ref().map(|v| {
        v.parse::<f32>()
            .expect("expected max server version to be numeric")
    });
    let server_version = server_version.as_ref().map(|v| {
        v.parse::<f32>()
            .expect("expected server version to be numeric")
    });

    // if the min server version is specified, check that the server version is >= min
    if let Some(min_version) = min_server_version.as_ref() {
        if let Some(server_version) = server_version.as_ref() {
            if server_version < min_version {
                return false;
            }
        }
    }
    // if the max server version is specified, check that the server version is <= max
    if let Some(max_version) = max_server_version.as_ref() {
        if let Some(server_version) = server_version.as_ref() {
            if server_version > max_version {
                return false;
            }
        }
    }
    true
}
