use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use report::{
    csv::generate_csv,
    html::generate_html,
    log_parser::{process_logs, LogParseResult},
    schema::{process_schemata, SchemaAnalysis},
};
use std::{env, path::PathBuf, process};
mod cli;
use anyhow::Result;
use cli::Cli;

pub const REPORT_FILE_STEM: &str = "Atlas_SQL_Readiness";
pub const REPORT_NAME: &str = "Atlas SQL Transition Readiness Report";

#[tokio::main]
async fn main() -> Result<()> {
    let Cli {
        input,
        output,
        uri,
        username,
        verbose,
    } = Cli::parse();

    let output_path = process_output_path(output);
    let parse_results = handle_logs(input)?;
    let analysis = handle_schema(uri, username, verbose).await?;
    let date = Local::now().format("%m-%d-%Y_%H%M").to_string();

    if parse_results.is_none() && analysis.is_none() {
        eprintln!("No input logs or URI provided for analysis, exiting...");
        process::exit(1);
    }

    generate_html(
        &output_path,
        &date,
        &parse_results,
        &analysis,
        verbose,
        REPORT_FILE_STEM,
        REPORT_NAME,
    )?;
    generate_csv(
        &output_path,
        &date,
        &parse_results,
        &analysis,
        verbose,
        REPORT_FILE_STEM,
    )
}

fn process_output_path(output: Option<String>) -> PathBuf {
    // User-specified output directory, uses current directory if not specified
    if let Some(ref output_path) = output {
        PathBuf::from(output_path)
    } else {
        env::current_dir().unwrap()
    }
}

fn handle_logs(input: Option<String>) -> Result<Option<LogParseResult>> {
    if let Some(input) = input {
        let metadata = std::fs::metadata(&input)
            .map_err(|_| {
                eprintln!("Input file or directory does not exist, {:?}", &input);
                process::exit(1);
            })
            .unwrap();
        let mut paths = Vec::new();

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
        Ok(Some(process_logs(&paths)?))
    } else {
        Ok(None)
    }
}

async fn handle_schema(
    uri: Option<String>,
    username: Option<String>,
    verbose: bool,
) -> Result<Option<SchemaAnalysis>> {
    if let Some(ref uri) = uri {
        if !uri.is_empty() {
            let mut options = sampler::get_opts(uri).await?;
            let password: Option<String>;
            if sampler::needs_auth(&options) {
                if let Some(ref username) = username {
                    if !username.is_empty() {
                        password = Some(
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
                sampler::load_password_auth(&mut options, username, password).await;
            }

            let schemata = sampler::sample(options, verbose).await?;
            Ok(Some(process_schemata(schemata)))
        } else {
            Ok(None)
        }
    } else {
        Ok(None)
    }
}
