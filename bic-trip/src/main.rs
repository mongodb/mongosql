use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use mongodb::bson::doc;
use mongosql::schema::Schema;
use report::{
    csv::generate_csv,
    html::{generate_html, generate_index},
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
pub const LOG_ANALYSIS_DIR: &str = "log_analysis";
pub const SCHEMA_ANALYSIS_DIR: &str = "schema_analysis";

#[tokio::main]
async fn main() -> Result<()> {
    let date = Local::now().format("%m-%d-%Y_%H%M").to_string();
    let Cli {
        input,
        output,
        uri,
        username,
        quiet,
        resolver,
        include,
        exclude,
    } = Cli::parse();

    if input.is_none() && uri.is_none() {
        eprintln!("No input logs or URI provided for analysis, exiting...");
        process::exit(1);
    }

    let output_path = process_output_path(output)?;

    std::fs::create_dir_all(&output_path).context("Failed to create output directory")?;

    let mut parse_results = None;
    if input.is_some() {
        std::fs::create_dir_all(output_path.join(LOG_ANALYSIS_DIR))
            .context("Failed to create log analysis directory")?;
        parse_results = handle_logs(input, quiet)?;
        generate_html(
            &output_path.join(LOG_ANALYSIS_DIR),
            parse_results.as_ref(),
            None,
            !quiet,
            REPORT_FILE_STEM,
            REPORT_NAME,
        )?;
    }

    let mut analysis = None;

    if uri.is_some() {
        std::fs::create_dir_all(output_path.join(SCHEMA_ANALYSIS_DIR))
            .context("Failed to create schema analysis directory")?;
        let schema_args = SchemaProcessingArgs {
            uri,
            username,
            quiet,
            resolver,
            include: include.unwrap_or_default(),
            exclude: exclude.unwrap_or_default(),
            file_path: output_path.join("schema_analysis"),
            report_name: REPORT_NAME.to_string(),
        };

        analysis = handle_schema(schema_args).await?;
    }

    generate_csv(
        &output_path,
        &date,
        parse_results.as_ref(),
        analysis.as_ref(),
        !quiet,
        REPORT_FILE_STEM,
    )?;
    generate_index(
        &output_path,
        REPORT_NAME,
        LOG_ANALYSIS_DIR,
        SCHEMA_ANALYSIS_DIR,
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

struct SchemaProcessingArgs {
    /// The URI to connect to the MongoDB cluster.
    uri: Option<String>,
    /// The username to use for authentication. Only if --username is set
    username: Option<String>,
    /// Whether to suppress output to the console.
    quiet: bool,
    /// The resolver to use for DNS resolution.
    resolver: Option<cli::Resolver>,
    /// The list of namespaces to include in the schema analysis.
    include: Vec<String>,
    /// The list of namespaces to exclude from the schema analysis.
    exclude: Vec<String>,
    /// The path to the output directory.
    file_path: PathBuf,
    /// The name of the report.
    report_name: String,
}

async fn handle_schema(args: SchemaProcessingArgs) -> Result<Option<SchemaAnalysis>> {
    let SchemaProcessingArgs {
        uri,
        username,
        quiet,
        resolver,
        include,
        exclude,
        file_path,
        report_name,
    } = args;
    if let Some(ref uri) = uri {
        if !uri.is_empty() {
            let mut options =
                schema_builder_library::client_util::get_opts(uri, resolver.map(|r| r.into()))
                    .await?;
            let password: Option<String>;
            if schema_builder_library::client_util::needs_auth(&options) {
                if let Some(ref username) = username {
                    if !username.is_empty() {
                        password = Some(
                            Password::new()
                                .with_prompt("Password")
                                .interact()
                                .expect("Failed to read password"),
                        );
                    } else {
                        eprintln!("No username provided for authentication.");
                        process::exit(1);
                    }
                } else {
                    eprintln!("No username provided for authentication.");
                    process::exit(1);
                }
                schema_builder_library::client_util::load_password_auth(
                    &mut options,
                    username,
                    password,
                )
                .await;
            }

            let (tx_notifications, mut rx_notifications) = tokio::sync::mpsc::unbounded_channel::<
                schema_builder_library::SamplerNotification,
            >();
            let (tx_schemata, mut rx_schemata) =
                tokio::sync::mpsc::unbounded_channel::<schema_builder_library::SchemaResult>();

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
            // Test the connection to the cluster
            client
                .database("admin")
                .run_command(doc! {"hello": 1})
                .await?;

            let include_list: Vec<glob::Pattern> = include
                .iter()
                .map(|g| match glob::Pattern::new(g) {
                    Ok(p) => p,
                    Err(e) => {
                        eprintln!("Error parsing include glob pattern: {}", e);
                        process::exit(1);
                    }
                })
                .collect();
            let exclude_list: Vec<glob::Pattern> = exclude
                .iter()
                .map(|g| match glob::Pattern::new(g) {
                    Ok(p) => p,
                    Err(e) => {
                        eprintln!("Error parsing exclude glob pattern: {}", e);
                        process::exit(1);
                    }
                })
                .collect();

            tokio::spawn(async move {
                let builder_options = schema_builder_library::options::BuilderOptions {
                    include_list,
                    exclude_list,
                    schema_collection: None,
                    dry_run: false,
                    client: client.clone(),
                    tx_notifications,
                    tx_schemata,
                };
                schema_builder_library::build_schema(builder_options).await;
            });

            loop {
                tokio::select! {
                    notification = rx_notifications.recv() => {
                        // The notification channel is not critical to the operation of the program,
                        // so we'll never break out of our loop if the channel is closed.
                        if let Some(notification) = notification {
                            match notification.action {
                                // If we receive an Error notification, we abort the program.
                                schema_builder_library::SamplerAction::Error { message } => anyhow::bail!(message),
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
                            Some(schema_builder_library::SchemaResult::FullSchema(schema_res)) => {
                                if !quiet {
                                    pb.set_message(format!("Schema calculated for namespace: {}.{} ({:?})",
                                        schema_res.namespace_info.db_name,
                                        schema_res.namespace_info.coll_or_view_name,
                                        schema_res.namespace_info.namespace_type,
                                    ));
                                }
                                schemata.entry(schema_res.namespace_info.db_name.clone()).and_modify(|v| {
                                    v.insert(schema_res.namespace_info.coll_or_view_name.clone(), schema_res.namespace_schema.clone());
                                }).or_insert(HashMap::from([(schema_res.namespace_info.coll_or_view_name.clone(), schema_res.namespace_schema.clone())]));

                                let schema_analysis = process_schemata(HashMap::from([(schema_res.namespace_info.db_name.clone(), schemata.get(&schema_res.namespace_info.db_name).unwrap().clone())]));
                                generate_html(&file_path, None, Some(&schema_analysis), !quiet, &schema_res.namespace_info.db_name, &report_name).unwrap();
                            }
                            Some(schema_builder_library::SchemaResult::NamespaceOnly(schema_res)) => {
                                if !quiet {
                                    pb.set_message(format!(
                                        "Namespace acknowledged in dryRun mode: {}.{} ({:?})",
                                        schema_res.db_name,
                                        schema_res.coll_or_view_name,
                                        schema_res.namespace_type,
                                    ))
                                }
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
