use agg_ast::definitions::Namespace;
use bson::{doc, Document};
use mongodb::sync::Client;
use mongosql::{build_catalog_from_catalog_schema, catalog::Catalog, json_schema::Schema};
use serde::{Deserialize, Serialize};
use std::collections::{BTreeMap, BTreeSet};

use crate::CliError;

const SQL_SCHEMAS_COLLECTION: &str = "__sql_schemas";

#[derive(Debug, Serialize, Deserialize)]
struct SchemaFile {
    #[serde(flatten)]
    schemas: BTreeMap<String, BTreeMap<String, Schema>>,
}

pub(crate) fn build_catalog(
    uri: &str,
    current_db: &str,
    namespaces: BTreeSet<Namespace>,
    schema_file: Option<String>,
) -> Result<Catalog, CliError> {
    if let Some(schema_file) = schema_file {
        let contents = std::fs::read_to_string(&schema_file)?;
        let path = std::path::Path::new(&schema_file);
        let extension = path
            .extension()
            .and_then(|ext| ext.to_str())
            .map(str::to_lowercase);

        let catalog: SchemaFile = match extension.as_deref() {
            Some("yaml" | "yml") => serde_yaml::from_str(&contents)?,
            Some("json") => serde_json::from_str(&contents)?,
            _ => {
                return Err(CliError(format!(
                    "Unsupported schema file extension: {extension:?}. Supported formats are .yml, .yaml, .json"
                )))
            }
        };
        Ok(build_catalog_from_catalog_schema(catalog.schemas)?)
    } else {
        get_schema_catalog(uri, current_db, namespaces)
    }
}

#[expect(
    clippy::needless_pass_by_value,
    reason = "changing to &BTreeSet would require updating build_catalog and all callers"
)]
fn get_schema_catalog(
    uri: &str,
    current_db: &str,
    namespaces: BTreeSet<Namespace>,
) -> Result<Catalog, CliError> {
    // If there are no namespaces (e.g. queries with only array datasources), assign
    // an empty schema to `current_db`
    if namespaces.is_empty() {
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
    let client = Client::with_uri_str(uri)?;
    let db = client.database(current_db);
    let schema_collection = db.collection::<Document>(SQL_SCHEMAS_COLLECTION);

    let collection_names = namespaces
        .iter()
        .map(|namespace| namespace.collection.as_str())
        .collect::<Vec<&str>>();

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

    let mut schema_catalog_doc_vec: Vec<Document> = schema_collection
        .aggregate(schema_catalog_aggregation_pipeline)
        .run()?
        .collect::<Result<Vec<Document>, _>>()?;

    if schema_catalog_doc_vec.len() > 1 {
        return Err(CliError("Multiple Schema Documents Returned".to_string()));
    }

    if schema_catalog_doc_vec.is_empty() {
        println!("[WARNING] No schema information was found for the requested collections `{collection_names:?}` in database `{current_db}`. Either the collections don't exist \
                    in `{current_db}` or they don't have a schema. For now, they will be assigned empty schemas. Hint: You either need to generate schemas for your collections \
                    or correct your query.");

        let mut collections_schema_doc = doc! {};

        for collection in collection_names {
            collections_schema_doc.insert(collection, doc! {});
        }

        let schema_catalog_doc = doc! {
          current_db: collections_schema_doc,
        };

        schema_catalog_doc_vec.push(schema_catalog_doc);
    }

    let mut schema_catalog_doc = schema_catalog_doc_vec[0].clone();

    let collections_schema_doc = schema_catalog_doc.get_document_mut(current_db)?;

    if namespaces.len() != collections_schema_doc.len() {
        let missing_collections: Vec<String> = namespaces
            .iter()
            .map(|namespace| namespace.collection.clone())
            .filter(|collection| !collections_schema_doc.contains_key(collection.as_str()))
            .collect();

        println!("[WARNING] No schema was found for the following collections: {missing_collections:?}. These collections will be assigned empty schemas. \
                    Hint: Generate schemas for your collections.");

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
