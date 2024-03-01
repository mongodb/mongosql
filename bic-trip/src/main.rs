use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use report::{csv::generate_csv, html::generate_html, log_parser::process_logs};
use std::{env, path::PathBuf, process};
mod cli;
use anyhow::Result;
use cli::Cli;

pub const REPORT_FILE_STEM: &str = "Atlas_SQL_Readiness";
pub const REPORT_NAME: &str = "Atlas SQL Transition Readiness Report";

#[tokio::main]
async fn main() -> Result<()> {
    let default_base_dir = env::current_dir()?;
    let mut paths = Vec::new();

    let Cli {
        input,
        output,
        uri,
        username,
    } = Cli::parse();

    // User-specified output directory, uses current directory if not specified
    let file_path = if let Some(ref output_path) = output {
        PathBuf::from(output_path)
    } else {
        default_base_dir
    };

    let mut _password: Option<String> = None;
    if let Some(ref uri) = uri {
        if !uri.is_empty() {
            // URI is specified, so we expect username/password
            if let Some(ref username) = username {
                if !username.is_empty() {
                    _password = Some(
                        Password::new()
                            .with_prompt("Password")
                            .interact()
                            .expect("Failed to read password"),
                    );
                } else {
                    println!("Username is required for authentication with URI");
                    process::exit(1);
                }
            } else {
                println!("No username provided for authentication with URI");
                process::exit(1);
            }
        }

        // TODO: SQL-1964
        // sample(uri, username, password, "sample_mflix", "movies").await?
    }

    let metadata = std::fs::metadata(&input)?;
    if metadata.is_file() {
        paths.push(input.clone());
    } else if metadata.is_dir() {
        for entry in std::fs::read_dir(input)? {
            let entry = entry?;
            if entry.path().is_file() {
                paths.push(entry.path().to_string_lossy().into_owned());
            }
        }
    }
    let parse_results = process_logs(&paths)?;
    let date = Local::now().format("%m-%d-%Y_%H%M").to_string();
    generate_html(
        &file_path,
        &date,
        &parse_results,
        REPORT_FILE_STEM,
        REPORT_NAME,
    )?;
    generate_csv(&file_path, &date, &parse_results, REPORT_FILE_STEM)
}

// TODO: SQL-1964
// async fn sample(
//     uri: &str,
//     username: Option<String>,
//     password: Option<String>,
//     db: &str,
//     coll: &str,
// ) -> Result<()> {
//     let client = Client::with_options(sampler::get_opts(uri, username, password).await?).unwrap();
//     let database = client.database(db);

//     let col_parts = sampler::gen_partitions(&database, coll).await;
//     let schemata = sampler::derive_schema_for_partitions(col_parts, &database).await;
//     traverse_schemata(&schemata);
//     Ok(())
// }

// fn traverse_schemata(schemata: &HashMap<String, mongosql::schema::Schema>) {
//     for (k, v) in schemata {
//         match v {
//             mongosql::schema::Schema::Unsat => todo!(),
//             mongosql::schema::Schema::Missing => todo!(),
//             mongosql::schema::Schema::Atomic(a) => {}
//             mongosql::schema::Schema::Array(a) => {}
//             mongosql::schema::Schema::Document(d) => {}
//             mongosql::schema::Schema::AnyOf(a) => todo!(),
//             mongosql::schema::Schema::Any => todo!(),
//         }
//     }
// }

// #[derive(Debug)]
// struct ArrayCount {
//     keys: Vec<String>,
//     count: usize,
// }

// fn count_arrays(d: &mongosql::schema::Document) -> ArrayCount {
//     d.keys.iter().fold(
//         ArrayCount {
//             keys: vec![],
//             count: 0,
//         },
//         |acc, (k, v)| match v {
//             mongosql::schema::Schema::Array(_) => {
//                 let mut keys = acc.keys.clone();
//                 keys.push(k.clone());
//                 ArrayCount {
//                     keys,
//                     count: acc.count + 1,
//                 }
//             }
//             _ => acc,
//         },
//     )
// }
