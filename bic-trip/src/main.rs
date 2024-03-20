use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use mongosql::schema::Schema;
use report::{
    csv::generate_csv,
    html::generate_html,
    log_parser::{process_logs, LogParseResult},
    schema::{process_schemata, SchemaAnalysis},
};
use std::{collections::HashMap, env, path::PathBuf, process, time::Instant};
mod cli;
use anyhow::Result;
use cli::Cli;
use indicatif::{HumanDuration, ProgressBar, ProgressStyle};

pub const REPORT_FILE_STEM: &str = "Atlas_SQL_Readiness";
pub const REPORT_NAME: &str = "Atlas SQL Transition Readiness Report";

#[tokio::main]
async fn main() -> Result<()> {
    let Cli {
        input,
        output,
        uri,
        username,
        quiet,
    } = Cli::parse();

    let output_path = process_output_path(output);
    let parse_results = handle_logs(input)?;
    let analysis = handle_schema(uri, username, quiet).await?;
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
        !quiet,
        REPORT_FILE_STEM,
        REPORT_NAME,
    )?;
    generate_csv(
        &output_path,
        &date,
        &parse_results,
        &analysis,
        !quiet,
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
    quiet: bool,
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

            let (tx_notifications, mut rx_notifications) =
                tokio::sync::mpsc::unbounded_channel::<sampler::SamplerNotification>();
            let (tx_schemata, mut rx_schemata) =
                tokio::sync::mpsc::unbounded_channel::<Result<sampler::SchemaAnalysis>>();

            let mut schemata: HashMap<String, Vec<HashMap<String, Schema>>> = HashMap::new();

            let spinner_style =
                ProgressStyle::with_template("{prefix:.bold.dim} {spinner} {wide_msg}")
                    .unwrap()
                    .tick_chars("⠁⠂⠄⡀⢀⠠⠐⠈ ");
            let pb = ProgressBar::new(1024);
            if !quiet {
                pb.set_style(spinner_style);
                pb.println("Sampling Atlas SQL databases...");
            }
            let start = Instant::now();

            tokio::spawn(async move {
                sampler::sample(options, tx_notifications, tx_schemata).await;
            });

            if !quiet {
                while let Some(notification) = rx_notifications.recv().await {
                    pb.set_message(notification.to_string());
                    pb.tick();
                }
            } else {
                while rx_notifications.recv().await.is_some() {}
            }
            while let Some(msg) = rx_schemata.recv().await {
                match msg {
                    Ok((database, schema)) => {
                        if !quiet {
                            pb.set_message(format!("Received schema for database: {}", database));
                        }

                        schemata.insert(database, schema);
                    }
                    Err(e) => {
                        eprintln!("Error: {:?}", e);
                    }
                }
            }
            if !quiet {
                pb.finish_with_message(format!(
                    "Schema analysis complete in {}",
                    HumanDuration(start.elapsed())
                ));
            }

            Ok(Some(process_schemata(schemata)))
        } else {
            Ok(None)
        }
    } else {
        Ok(None)
    }
}
