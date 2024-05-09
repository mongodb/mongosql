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
                tokio::sync::mpsc::unbounded_channel::<sampler::Result<sampler::SchemaAnalysis>>();
            let (tx_errors, rx_errors) = tokio::sync::oneshot::channel::<sampler::Result<()>>();

            let mut schemata: HashMap<String, Vec<HashMap<String, Schema>>> = HashMap::new();

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

            tokio::spawn(async move {
                sampler::sample(options, Some(tx_notifications), tx_schemata, tx_errors).await;
            });

            if let Ok(Err(e)) = rx_errors.await {
                anyhow::bail!(e);
            }

            loop {
                tokio::select! {
                    notification = rx_notifications.recv() => {
                        // The notification channel is not critical to the operation of the program,
                        // so we'll never break out of our loop if the channel is closed.
                        if let Some(ref notification) = notification {
                            if !quiet {
                                pb.set_message(notification.to_string());
                            }
                        }
                    }
                    schema = rx_schemata.recv() => {
                        match schema {
                            Some(Ok((database, schema))) => {
                                if !quiet {
                                    pb.set_message(format!("Schema calculated for database: {database}"));
                                }
                                schemata.insert(database, schema);
                            }
                            Some(Err(e)) => {
                                if !quiet {
                                    println!("Error: {e}");
                                    pb.set_message(format!("Error: {}", e));
                                }
                            }
                            None => {
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
