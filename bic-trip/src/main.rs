use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use mongosql::schema::Schema;
use report::{
    csv::generate_csv,
    html::generate_html,
    log_parser::handle_logs,
    schema::{process_schemata, SchemaAnalysis},
};
use std::{collections::HashMap, env, path::PathBuf, process, time::Instant};
mod cli;
use anyhow::{Context, Result};
use cli::Cli;
use indicatif::{HumanDuration, ProgressBar, ProgressStyle};
use schema_builder_library as dcsb; // dcsb == Direct Cluster Schema Builder

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

    let output_path = process_output_path(output)?;
    let parse_results = handle_logs(input, quiet)?;
    let analysis = handle_schema(uri, username, quiet).await?;
    let date = Local::now().format("%m-%d-%Y_%H%M").to_string();

    if parse_results.is_none() && analysis.is_none() {
        eprintln!("No input logs or URI provided for analysis, exiting...");
        process::exit(1);
    }

    generate_html(
        &output_path,
        &date,
        parse_results.as_ref(),
        analysis.as_ref(),
        !quiet,
        REPORT_FILE_STEM,
        REPORT_NAME,
    )?;
    generate_csv(
        &output_path,
        &date,
        parse_results.as_ref(),
        analysis.as_ref(),
        !quiet,
        REPORT_FILE_STEM,
    )
}

fn process_output_path(output: Option<String>) -> Result<PathBuf> {
    // User-specified output directory, uses current directory if not specified
    if let Some(ref output_path) = output {
        Ok(PathBuf::from(output_path))
    } else {
        env::current_dir().context("Failed to get current directory. Do you have permissions?")
    }
}

async fn handle_schema(
    uri: Option<String>,
    username: Option<String>,
    quiet: bool,
) -> Result<Option<SchemaAnalysis>> {
    if let Some(ref uri) = uri {
        if !uri.is_empty() {
            let mut options = dcsb::client_util::get_opts(uri).await?;
            let password: Option<String>;
            if dcsb::client_util::needs_auth(&options) {
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
                dcsb::client_util::load_password_auth(&mut options, username, password).await;
            }

            let (tx_notifications, mut rx_notifications) =
                tokio::sync::mpsc::unbounded_channel::<dcsb::SamplerNotification>();
            let (tx_schemata, mut rx_schemata) =
                tokio::sync::mpsc::unbounded_channel::<dcsb::SchemaResult>();

            let mut schemata: HashMap<String, HashMap<String, Schema>> = HashMap::new();

            // spinner style errors are caught at compile time so are safe to unwrap on
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
            let client = mongodb::Client::with_options(options.clone())
                .with_context(|| "Failed to create MongoDB client.")?;

            tokio::spawn(async move {
                let builder_options = dcsb::options::BuilderOptions {
                    include_list: vec![],
                    exclude_list: vec![],
                    schema_collection: None,
                    dry_run: false,
                    client: client.clone(),
                    tx_notifications: Some(tx_notifications),
                    tx_schemata,
                };
                dcsb::build_schema(builder_options).await;
            });

            loop {
                tokio::select! {
                    notification = rx_notifications.recv() => {
                        // The notification channel is not critical to the operation of the program,
                        // so we'll never break out of our loop if the channel is closed.
                        if let Some(notification) = notification {
                            match notification.action {
                                // If we receive an Error notification, we abort the program.
                                dcsb::SamplerAction::Error { message } => anyhow::bail!(message),
                                // All other notification types are simply logged, depending on the
                                // value of quiet.
                                _ => {
                                    if !quiet {
                                        pb.set_message(notification.to_string());
                                    }
                                }
                            }
                        }
                    }
                    schema = rx_schemata.recv() => {
                        match schema {
                            Some(schema_res) => {
                                if !quiet {
                                    pb.set_message(format!("Schema calculated for database: {}", schema_res.db_name));
                                }
                                schemata.entry(schema_res.db_name).and_modify(|v| {
                                    v.insert(schema_res.coll_or_view_name.clone(), schema_res.namespace_schema.clone());
                                }).or_insert(HashMap::from([(schema_res.coll_or_view_name, schema_res.namespace_schema)]));
                            }
                            None => {
                                // When the channel is closed, terminate the loop.
                                break;
                            }
                        }
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
