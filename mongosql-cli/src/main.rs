use bson::{doc, Document};
use clap::Parser;
use mongodb::sync::{Client, Collection};
use mongosql::{catalog::Catalog, json_schema::Schema, Namespace};
use std::collections::{BTreeMap, BTreeSet};

const SQL_SCHEMAS_COLLECTION: &str = "__sql_schemas";

#[derive(Debug)]
struct CliError(String);

impl std::fmt::Display for CliError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl<T> From<T> for CliError
where
    T: std::error::Error,
{
    fn from(e: T) -> Self {
        CliError(e.to_string())
    }
}

#[derive(Parser, Debug)]
#[command(version, about, long_about=None)]
struct Cli {
    #[arg(
        short,
        long,
        help = "The current database where collections in the query are assumed to live (cross database queries are not supported), default = test"
    )]
    db: Option<String>,
    #[arg(index = 1, help = "The query to translate")]
    query: String,
    #[arg(
        short,
        long,
        help = "translation, automatically true if execute not set"
    )]
    translation: bool,
    #[arg(short, long, help = "Run the query and display the result")]
    execute: bool,
    #[arg(
        short,
        long,
        help = "The mongodb uri, default = mongodb://localhost:27017"
    )]
    uri: Option<String>,
}

fn main() -> Result<(), CliError> {
    let args = Cli::parse();

    let uri = args.uri.unwrap_or("mongodb://localhost:27017".to_string());
    let current_db = args.db.unwrap_or("test".to_string());
    let query = args.query;
    let namespaces = mongosql::get_namespaces(current_db.as_str(), query.as_str())?;
    let catalog = get_schema_catalog(uri.as_str(), current_db.as_str(), namespaces)?;
    let translation = mongosql::translate_sql(
        current_db.as_str(),
        query.as_str(),
        &catalog,
        mongosql::options::SqlOptions::default(),
    )?;
    let print_translation = |pipeline: bson::Bson| -> Result<(), CliError> {
        let schema = serde_json::to_string_pretty(&translation.result_set_schema)
            .map_err(|e| CliError(e.to_string()))?;
        println!(
            "target_db: {},\ntarget_collection: {:?},\nresult set schema:\n{}\npipeline:",
            translation.target_db, translation.target_collection, schema
        );
        let bson::Bson::Array(pipeline) = pipeline else {
            return Err(CliError("pipeline is not an array".to_string()));
        };
        for doc in pipeline {
            println!("    {}", doc);
        }
        Ok(())
    };
    // If the result flag is not set, we always want to print the translation, regardless of the translation flag.
    if !args.execute {
        print_translation(translation.pipeline)?;
        return Ok(());
    }
    // When running the result, we still want to print the translation if it is asked for
    if args.translation {
        print_translation(translation.pipeline.clone())?;
    }
    run_query_and_display_results(uri.as_str(), translation)
}

fn run_query_and_display_results(
    uri: &str,
    translation: mongosql::Translation,
) -> Result<(), CliError> {
    let client = Client::with_uri_str(uri)?;
    let db = client.database(translation.target_db.as_str());
    let bson::Bson::Array(pipeline) = translation.pipeline else {
        return Err(CliError("pipeline is not an array".to_string()));
    };
    let pipeline = pipeline
        .into_iter()
        .map(|doc| doc.as_document().map(|doc| doc.to_owned()))
        .collect::<Option<Vec<Document>>>()
        .ok_or_else(|| CliError("Pipeline contains non-Document!".to_string()))?;
    let results = if let Some(target_collection) = translation.target_collection {
        let collection: Collection<Document> = db.collection(target_collection.as_str());
        let cursor = collection.aggregate(pipeline).run();
        cursor?
    } else {
        let cursor = db.aggregate(pipeline).run();
        cursor?
    };
    println!("result:");
    for result in results {
        let result = result?;
        println!("    {}", result);
    }
    Ok(())
}

fn get_schema_catalog(
    uri: &str,
    current_db: &str,
    namespaces: BTreeSet<Namespace>,
) -> Result<Catalog, CliError> {
    let client = Client::with_uri_str(uri)?;
    let db = client.database(current_db);
    let collection_names = db.list_collection_names().run()?;

    let sql_schemas_collection_exists =
        collection_names.contains(&SQL_SCHEMAS_COLLECTION.to_string());

    if !sql_schemas_collection_exists {
        return Err(CliError(format!("There is no schema information in database `{0}`, so all the collections will be assigned empty schemas. \
                Therefore, SQL capabilities will be very limited. Hint: Please make sure to generate schemas before using the driver", current_db)));
    }

    // If there are no namespaces (most importantly for the `SELECT 1` query) or no schema information,
    // assign an empty schema to `current_db`.
    if namespaces.is_empty() || !sql_schemas_collection_exists {
        let schema_catalog_doc = doc! {
                current_db: doc! {},
        };

        return Ok(mongosql::build_catalog_from_catalog_schema(
            serde_json::from_str::<BTreeMap<String, BTreeMap<String, Schema>>>(
                &schema_catalog_doc.to_string(),
            )?,
        )?);
    }

    // Otherwise, fetch the schema information for the specified collections.
    let schema_collection = db.collection::<Document>(SQL_SCHEMAS_COLLECTION);

    let collection_names = namespaces
        .iter()
        .map(|namespace| namespace.collection.as_str())
        .collect::<Vec<&str>>();

    // Create an aggregation pipeline to fetch the schema information for the specified collections.
    // The pipeline uses $in to query all the specified collections and projects them into the desired format:
    // "dbName": { "collection1" : "Schema1", "collection2" : "Schema2", ... }
    let schema_catalog_aggregation_pipeline = vec![
        doc! {"$match": {
            "_id": {
                "$in": &collection_names
                }
            }
        },
        doc! {"$project":{
            "_id": 1,
            "schema": 1
            }
        },
        doc! {"$group": {
            "_id": null,
            "collections": {
                "$push": {
                    "collectionName": "$_id",
                    "schema": "$schema"
                    }
                }
            }
        },
        doc! {"$project": {
            "_id": 0,
            current_db: {
                "$arrayToObject": [{
                    "$map": {
                        "input": "$collections",
                        "as": "coll",
                        "in": {
                            "k": "$$coll.collectionName",
                            "v": "$$coll.schema"
                            }
                        }
                    }]
                }
            }
        },
    ];

    // create the schema_catalog document
    let mut schema_catalog_doc_vec: Vec<Document> = schema_collection
        .aggregate(schema_catalog_aggregation_pipeline)
        .run()?
        .collect::<Result<Vec<Document>, _>>()?;

    if schema_catalog_doc_vec.len() > 1 {
        return Err(CliError("Multiple Schema Documents Returned".to_string()));
    }

    if schema_catalog_doc_vec.is_empty() {
        println!("No schema information was found for the requested collections `{:?}` in database `{1}`. Either the collections don't exists \
                    in `{1}` or they don't have a schema. For now, they will be assigned empty schemas. Hint: You either need to generate schemas for your collections \
                    or correct your query.", collection_names, current_db);

        let mut collections_schema_doc = doc! {};

        for collection in collection_names {
            collections_schema_doc.insert(collection, doc! {});
        }

        let schema_catalog_doc = doc! {
          current_db: collections_schema_doc,
        };

        schema_catalog_doc_vec.push(schema_catalog_doc);
    }

    let mut schema_catalog_doc = schema_catalog_doc_vec[0].to_owned();

    let collections_schema_doc = schema_catalog_doc.get_document_mut(current_db)?;

    // If there are collections with no schema available, assign them empty schemas.
    if namespaces.len() != collections_schema_doc.len() {
        let missing_collections: Vec<String> = namespaces
            .iter()
            .map(|namespace| namespace.collection.clone())
            .filter(|collection| !collections_schema_doc.contains_key(collection.as_str()))
            .collect();

        println!("No schema was found for the following collections: {:?}. These collections will be assigned empty schemas.\
                    Hint: Generate schemas for your collections.", missing_collections);

        for collection in missing_collections {
            collections_schema_doc.insert(collection, doc! {});
        }
    }

    Ok(mongosql::build_catalog_from_catalog_schema(
        serde_json::from_str::<BTreeMap<String, BTreeMap<String, Schema>>>(
            &schema_catalog_doc.to_string(),
        )?,
    )?)
}
