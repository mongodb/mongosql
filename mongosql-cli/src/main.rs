mod audit_trail;
mod catalog;

use audit_trail::{handle_audit_trail, TranslationCheckpoint};
use bson::Document;
use catalog::build_catalog;
use clap::Parser;
use mongodb::sync::{Client, Collection};
use std::path::PathBuf;

#[derive(Debug)]
pub(crate) struct CliError(String);

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

    let namespaces = mongosql::get_namespaces(current_db.as_str(), query.as_str())?;
    let catalog = build_catalog(
        uri.as_str(),
        current_db.as_str(),
        namespaces,
        args.schema_file,
    )?;

    if args.audit_trail {
        return handle_audit_trail(
            stage,
            current_db.as_str(),
            query.as_str(),
            uri.as_str(),
            args.execute,
            &catalog,
        );
    }

    match stage {
        TranslationCheckpoint::Ast => {
            let output = mongosql::translate_sql_to_ast_repr(query.as_str())?;
            println!("{output}");
            return Ok(());
        }
        TranslationCheckpoint::Mir => {
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

pub(crate) fn run_query_and_display_results(
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

#[cfg(test)]
mod tests {
    use super::*;

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
}
