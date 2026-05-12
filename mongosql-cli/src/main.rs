use agg_ast::definitions::Namespace;
use bson::{doc, Document};
use clap::Parser;
use mongodb::sync::{Client, Collection};
use mongosql::{build_catalog_from_catalog_schema, catalog::Catalog, json_schema::Schema};
use serde::{Deserialize, Serialize};
use std::collections::{BTreeMap, BTreeSet};
use std::path::PathBuf;

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

#[derive(clap::ValueEnum, Debug, Clone, Copy)]
enum TranslationCheckpoint {
    /// Stop after SQL parsing and AST rewrites; print the AST.
    Ast,
    /// Stop after algebrizing to MIR and running optimizer passes; print the MIR tree.
    Mir,
    /// AIR pretty-printing is not yet implemented.
    Air,
    /// Full translation to MQL; print the generated pipeline (default when --stage is omitted).
    Mql,
}

#[derive(Parser, Debug)]
#[command(version, about, long_about = None, arg_required_else_help = true)]
struct Cli {
    #[arg(
        short,
        long,
        help = "The current database where collections in the query are assumed to live (cross database queries are not supported). Required if `execute` is specified, or if no `schema_file` is specified. default = test"
    )]
    db: Option<String>,
    #[arg(index = 1, help = "The query to translate")]
    query: Option<String>,
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
    #[arg(
        short = 'f',
        long,
        help = "A schema file to use instead of querying the database for schema"
    )]
    schema_file: Option<String>,
    #[arg(
        long,
        short = 'q',
        help = "A sql file to use instead of passing the query as an argument. The sql-file argument takes precedence over sql query text."
    )]
    sql_file: Option<PathBuf>,
    #[arg(
        long,
        value_enum,
        help = "Stop at the specified compilation stage and print its intermediate representation. When omitted the CLI behaves as normal."
    )]
    stage: Option<TranslationCheckpoint>,
    #[arg(
        long,
        default_value_t = false,
        help = "Write intermediate representations for each completed translation stage \
                into audit_trail.zip in the current working directory. \
                Extracting the zip produces an audit_trail/ folder containing: \
                initial_query.sql, and whichever of query.ast, query.mir, query.air, \
                pipeline.js were reached. Which stages are included depends on --stage \
                (defaults to all stages). Normal stdout output is unchanged."
    )]
    audit_trail: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SchemaFile {
    #[serde(flatten)]
    pub schemas: BTreeMap<String, BTreeMap<String, Schema>>,
}

/// Collected intermediate representations for the audit trail.
///
/// Fields are `None` for stages not reached given the requested checkpoint.
struct TranslationStages {
    ast_repr: Option<String>,
    mir_repr: Option<String>,
    air_repr: Option<String>,
    pipeline_json: Option<String>,
    /// Only `Some` when stage is `Mql`; needed for the `--execute` path.
    translation: Option<mongosql::Translation>,
}

fn parse_query_from_args(
    query: Option<String>,
    sql_file: Option<PathBuf>,
) -> Result<String, CliError> {
    if let Some(sql_file) = sql_file {
        let parsed_file = std::fs::read_to_string(sql_file)?.trim().to_string();
        if parsed_file.is_empty() {
            Err(CliError(
                "The provided sql file is empty. Please provide a valid sql file or pass the sql query directly to the cli.".to_string(),
            ))
        } else {
            Ok(parsed_file)
        }
    } else if let Some(query) = query {
        Ok(query)
    } else {
        Err(CliError(
            "No query provided. Please provide a query or a sql file using the `sql-file` argument.".to_string(),
        ))
    }
}

fn build_catalog(
    uri: &str,
    current_db: &str,
    namespaces: std::collections::BTreeSet<Namespace>,
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

/// Runs translation stages sequentially up to `stage`, collecting each
/// intermediate representation.
///
/// The catalog is only built when `stage` is `Mir`, `Air`, or `Mql`.
///
/// # Errors
/// Returns `CliError` if any translation stage or catalog build fails.
fn run_stages_up_to(
    stage: TranslationCheckpoint,
    current_db: &str,
    query: &str,
    uri: &str,
    schema_file: Option<String>,
) -> Result<TranslationStages, CliError> {
    let ast_repr = Some(mongosql::translate_sql_to_ast_repr(query)?);

    if matches!(stage, TranslationCheckpoint::Ast) {
        return Ok(TranslationStages {
            ast_repr,
            mir_repr: None,
            air_repr: None,
            pipeline_json: None,
            translation: None,
        });
    }

    let namespaces = mongosql::get_namespaces(current_db, query)?;
    let catalog = build_catalog(uri, current_db, namespaces, schema_file)?;
    let options = mongosql::options::SqlOptions {
        allow_order_by_missing_columns: true,
        ..Default::default()
    };

    let mir_repr = Some(mongosql::translate_sql_to_mir_repr(
        current_db, query, &catalog, options,
    )?);

    if matches!(stage, TranslationCheckpoint::Mir) {
        return Ok(TranslationStages {
            ast_repr,
            mir_repr,
            air_repr: None,
            pipeline_json: None,
            translation: None,
        });
    }

    let air_repr = Some(mongosql::translate_sql_to_air_repr(
        current_db, query, &catalog, options,
    )?);

    if matches!(stage, TranslationCheckpoint::Air) {
        return Ok(TranslationStages {
            ast_repr,
            mir_repr,
            air_repr,
            pipeline_json: None,
            translation: None,
        });
    }

    // Mql
    let translation = mongosql::translate_sql(current_db, query, &catalog, options)?;
    let pipeline_json =
        serde_json::to_string_pretty(&translation.pipeline).map_err(|e| CliError(e.to_string()))?;

    Ok(TranslationStages {
        ast_repr,
        mir_repr,
        air_repr,
        pipeline_json: Some(pipeline_json),
        translation: Some(translation),
    })
}

/// Writes intermediate translation representations into `audit_trail.zip`
/// in the current working directory, overwriting any existing file.
///
/// `initial_query.sql` is always written. Remaining files are written only
/// for stages that were reached (i.e. their field is `Some`).
///
/// # Errors
/// Returns `CliError` if the zip file cannot be created or written.
fn write_audit_trail(query: &str, stages: &TranslationStages) -> Result<(), CliError> {
    use std::io::Write as _;
    use zip::write::FileOptions;
    use zip::CompressionMethod;

    let file = std::fs::File::create("audit_trail.zip")?;
    let mut zip = zip::write::ZipWriter::new(file);
    let options = FileOptions::<()>::default().compression_method(CompressionMethod::Deflated);

    zip.start_file("audit_trail/initial_query.sql", options)?;
    zip.write_all(query.as_bytes())?;

    if let Some(ast) = &stages.ast_repr {
        zip.start_file("audit_trail/query.ast", options)?;
        zip.write_all(ast.as_bytes())?;
    }
    if let Some(mir) = &stages.mir_repr {
        zip.start_file("audit_trail/query.mir", options)?;
        zip.write_all(mir.as_bytes())?;
    }
    if let Some(air) = &stages.air_repr {
        zip.start_file("audit_trail/query.air", options)?;
        zip.write_all(air.as_bytes())?;
    }
    if let Some(pipeline) = &stages.pipeline_json {
        zip.start_file("audit_trail/pipeline.js", options)?;
        zip.write_all(pipeline.as_bytes())?;
    }

    zip.finish()?;
    Ok(())
}

/// Runs all translation stages up to `stage`, writes `audit_trail.zip`, prints
/// the stage output to stdout, and optionally executes the query.
///
/// Stdout output is identical to running `--stage` alone.
///
/// # Errors
/// Returns `CliError` if any translation stage, zip write, or query execution fails.
fn handle_audit_trail(
    stage: TranslationCheckpoint,
    current_db: &str,
    query: &str,
    uri: &str,
    schema_file: Option<String>,
    execute: bool,
) -> Result<(), CliError> {
    let stages = run_stages_up_to(stage, current_db, query, uri, schema_file)?;
    write_audit_trail(query, &stages)?;
    eprintln!("audit_trail.zip written to current directory.");

    match stage {
        TranslationCheckpoint::Ast => {
            println!("{}", stages.ast_repr.expect("ast computed for Ast stage"));
        }
        TranslationCheckpoint::Mir => {
            println!("{}", stages.mir_repr.expect("mir computed for Mir stage"));
        }
        TranslationCheckpoint::Air => {
            println!("{}", stages.air_repr.expect("air computed for Air stage"));
        }
        TranslationCheckpoint::Mql => {
            let translation = stages
                .translation
                .as_ref()
                .expect("translation computed for Mql stage");
            let schema = serde_json::to_string_pretty(&translation.result_set_schema)
                .map_err(|e| CliError(e.to_string()))?;
            println!(
                "target_db: {},\ntarget_collection: {:?},\nresult set schema:\n{}\npipeline:\n[",
                translation.target_db, translation.target_collection, schema
            );
            let bson::Bson::Array(ref pipeline) = translation.pipeline else {
                return Err(CliError("pipeline is not an array".to_string()));
            };
            for doc in pipeline {
                println!("    {doc},");
            }
            println!("]");
        }
    }

    if execute {
        run_query_and_display_results(
            uri,
            stages
                .translation
                .expect("translation computed for Mql stage"),
        )?;
    }
    Ok(())
}

#[expect(
    clippy::too_many_lines,
    reason = "each TranslationCheckpoint arm repeats catalog/options setup; extracting further would obscure the control flow"
)]
fn main() -> Result<(), CliError> {
    let args = Cli::parse();

    let uri = args.uri.unwrap_or("mongodb://localhost:27017".to_string());
    let current_db = args.db.unwrap_or("test".to_string());
    let query = parse_query_from_args(args.query, args.sql_file)?;
    let stage = args.stage.unwrap_or(TranslationCheckpoint::Mql);

    if args.execute && !matches!(stage, TranslationCheckpoint::Mql) {
        return Err(CliError(
            "--execute is only valid with --stage mql or without --stage".to_string(),
        ));
    }

    if args.audit_trail {
        return handle_audit_trail(
            stage,
            current_db.as_str(),
            query.as_str(),
            uri.as_str(),
            args.schema_file,
            args.execute,
        );
    }

    match stage {
        TranslationCheckpoint::Ast => {
            let output = mongosql::translate_sql_to_ast_repr(query.as_str())?;
            println!("{output}");
            return Ok(());
        }
        TranslationCheckpoint::Mir => {
            let namespaces = mongosql::get_namespaces(current_db.as_str(), query.as_str())?;
            let catalog = build_catalog(
                uri.as_str(),
                current_db.as_str(),
                namespaces,
                args.schema_file,
            )?;
            let options = mongosql::options::SqlOptions {
                allow_order_by_missing_columns: true,
                ..Default::default()
            };
            let output = mongosql::translate_sql_to_mir_repr(
                current_db.as_str(),
                query.as_str(),
                &catalog,
                options,
            )?;
            println!("{output}");
            return Ok(());
        }
        TranslationCheckpoint::Air => {
            let namespaces = mongosql::get_namespaces(current_db.as_str(), query.as_str())?;
            let catalog = build_catalog(
                uri.as_str(),
                current_db.as_str(),
                namespaces,
                args.schema_file,
            )?;
            let options = mongosql::options::SqlOptions {
                allow_order_by_missing_columns: true,
                ..Default::default()
            };
            let output = mongosql::translate_sql_to_air_repr(
                current_db.as_str(),
                query.as_str(),
                &catalog,
                options,
            )?;
            println!("{output}");
            return Ok(());
        }
        TranslationCheckpoint::Mql => {}
    }

    let namespaces = mongosql::get_namespaces(current_db.as_str(), query.as_str())?;
    let catalog = build_catalog(
        uri.as_str(),
        current_db.as_str(),
        namespaces,
        args.schema_file,
    )?;
    let options = mongosql::options::SqlOptions {
        allow_order_by_missing_columns: true,
        ..Default::default()
    };
    let translation =
        mongosql::translate_sql(current_db.as_str(), query.as_str(), &catalog, options)?;
    let print_translation = |pipeline: bson::Bson| -> Result<(), CliError> {
        let schema = serde_json::to_string_pretty(&translation.result_set_schema)
            .map_err(|e| CliError(e.to_string()))?;
        println!(
            "target_db: {},\ntarget_collection: {:?},\nresult set schema:\n{}\npipeline:\n[",
            translation.target_db, translation.target_collection, schema
        );
        let bson::Bson::Array(pipeline) = pipeline else {
            return Err(CliError("pipeline is not an array".to_string()));
        };
        for doc in pipeline {
            println!("    {doc},");
        }
        println!("]");
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
        .map(|doc| doc.as_document().map(std::borrow::ToOwned::to_owned))
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
        println!("    {result}");
    }
    Ok(())
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

    // If there are collections with no schema available, assign them empty schemas.
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

#[cfg(test)]
mod tests {
    use super::*;

    // Serializes tests that mutate the process-wide current directory.
    static CWD_LOCK: std::sync::Mutex<()> = std::sync::Mutex::new(());
    #[test]
    fn it_parses_query_correctly() {
        let query = "SELECT * FROM users".to_string();
        let sql_file = None;

        let parse_result = parse_query_from_args(Some(query), sql_file);

        assert!(parse_result.is_ok());
        assert_eq!(parse_result.unwrap(), "SELECT * FROM users".to_string());
    }

    #[test]
    fn it_parses_sql_file_correctly() {
        let query: Option<String> = None;
        let sql_file = Some(PathBuf::from("./test/sample_query.sql"));

        let parse_result = parse_query_from_args(query, sql_file);

        assert!(parse_result.is_ok());
        assert_eq!(parse_result.unwrap(), "SELECT customerAge, COUNT(*) FROM sample_supplies.sales GROUP BY customer.age AS customerAge limit 10;".trim().to_string());
    }

    #[test]
    fn sql_file_takes_precedence_over_query() {
        let query = "SELECT * FROM users".to_string();
        let sql_file = Some(PathBuf::from("./test/sample_query.sql"));

        let parse_result = parse_query_from_args(Some(query), sql_file);

        assert!(parse_result.is_ok());
        assert_eq!(parse_result.unwrap(), "SELECT customerAge, COUNT(*) FROM sample_supplies.sales GROUP BY customer.age AS customerAge limit 10;".trim().to_string());
    }

    #[test]
    fn no_query_provided_returns_error() {
        let query: Option<String> = None;
        let sql_file: Option<PathBuf> = None;

        let parse_result = parse_query_from_args(query, sql_file);

        assert!(parse_result.is_err());
    }

    #[test]
    fn empty_sql_file_returns_error() {
        let query: Option<String> = None;
        let sql_file = Some(PathBuf::from("./test/empty_query.sql"));

        let parse_result = parse_query_from_args(query, sql_file);
        assert!(parse_result.is_err());
    }

    fn make_test_stages_mql() -> TranslationStages {
        TranslationStages {
            ast_repr: Some("ast output".to_string()),
            mir_repr: Some("mir output".to_string()),
            air_repr: Some("air output".to_string()),
            pipeline_json: Some(r#"[{"$match":{}}]"#.to_string()),
            translation: None,
        }
    }

    #[test]
    fn write_audit_trail_creates_correct_entries_for_mql_stage() {
        // Arrange
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let stages = make_test_stages_mql();

        // Act
        write_audit_trail("SELECT 1", &stages).unwrap();
        std::env::set_current_dir(original).unwrap();

        // Assert
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let names: Vec<String> = (0..archive.len())
            .map(|i| archive.by_index(i).unwrap().name().to_string())
            .collect();

        assert!(names.contains(&"audit_trail/initial_query.sql".to_string()));
        assert!(names.contains(&"audit_trail/query.ast".to_string()));
        assert!(names.contains(&"audit_trail/query.mir".to_string()));
        assert!(names.contains(&"audit_trail/query.air".to_string()));
        assert!(names.contains(&"audit_trail/pipeline.js".to_string()));
        assert_eq!(names.len(), 5);
    }

    #[test]
    fn write_audit_trail_creates_correct_entries_for_ast_stage() {
        // Arrange
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let stages = TranslationStages {
            ast_repr: Some("ast only".to_string()),
            mir_repr: None,
            air_repr: None,
            pipeline_json: None,
            translation: None,
        };

        // Act
        write_audit_trail("SELECT 1", &stages).unwrap();
        std::env::set_current_dir(original).unwrap();

        // Assert
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let names: Vec<String> = (0..archive.len())
            .map(|i| archive.by_index(i).unwrap().name().to_string())
            .collect();

        assert_eq!(names.len(), 2);
        assert!(names.contains(&"audit_trail/initial_query.sql".to_string()));
        assert!(names.contains(&"audit_trail/query.ast".to_string()));
    }

    #[test]
    fn write_audit_trail_sql_content_is_verbatim() {
        // Arrange
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let query = "SELECT\n  á\tFROM t";
        let stages = make_test_stages_mql();

        // Act
        write_audit_trail(query, &stages).unwrap();
        std::env::set_current_dir(original).unwrap();

        // Assert
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let mut entry = archive.by_name("audit_trail/initial_query.sql").unwrap();
        let mut content = Vec::new();
        std::io::Read::read_to_end(&mut entry, &mut content).unwrap();

        assert_eq!(content, query.as_bytes());
    }

    #[test]
    fn write_audit_trail_pipeline_json_is_valid_json() {
        // Arrange
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let stages = TranslationStages {
            ast_repr: Some("ast".to_string()),
            mir_repr: Some("mir".to_string()),
            air_repr: Some("air".to_string()),
            pipeline_json: Some(r#"[{"$match": {"x": 1}}]"#.to_string()),
            translation: None,
        };

        // Act
        write_audit_trail("SELECT 1", &stages).unwrap();
        std::env::set_current_dir(original).unwrap();

        // Assert
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let mut entry = archive.by_name("audit_trail/pipeline.js").unwrap();
        let mut content = String::new();
        std::io::Read::read_to_string(&mut entry, &mut content).unwrap();

        assert!(serde_json::from_str::<serde_json::Value>(&content).is_ok());
    }

    #[test]
    fn write_audit_trail_overwrites_existing_zip() {
        // Arrange
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        // Pre-create a dummy zip
        std::fs::write("audit_trail.zip", b"not a real zip").unwrap();
        let stages = make_test_stages_mql();

        // Act
        let result = write_audit_trail("SELECT 1", &stages);
        std::env::set_current_dir(original).unwrap();

        // Assert
        assert!(result.is_ok());
        let file = std::fs::File::open(&zip_path).unwrap();
        let archive_result = zip::ZipArchive::new(file);
        assert!(archive_result.is_ok());
        let mut archive = archive_result.unwrap();
        assert!(archive.by_name("audit_trail/initial_query.sql").is_ok());
    }
}
