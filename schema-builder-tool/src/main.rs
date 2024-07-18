mod builder_result;
mod cli;
mod consts;
mod schema_document;

use anyhow::{Context, Result};
use builder_result::SchemaBuilderResult::{self, *};
use clap::Parser;
use cli::{Cli, SchemaAction::*};
use consts::{DEFAULT_APP_NAME, SCHEMA_COLLECTION_NAME};
use dialoguer::Password;
use indicatif::{ProgressBar, ProgressStyle};
use mongodb::bson::{self, datetime, doc};
use mongodb::Client;
use mongosql::{
    json_schema,
    schema::definitions::Error::{BsonFailure, JsonSchemaFailure},
};
use schema_builder_library::{
    build_schema,
    client_util::{get_opts, load_password_auth, needs_auth},
    options::BuilderOptions,
    SamplerAction, SamplerNotification, SchemaResult,
};
use schema_document::SchemaDocument;
use std::collections::HashMap;
use std::process;
use tokio::sync::mpsc::unbounded_channel;
use tracing::{event, instrument, Level};

#[tokio::main]
async fn main() -> Result<()> {
    let mut cfg = Cli::parse();
    // command line arguments override configuration file
    if let Some(ref config) = cfg.config_file {
        let file_cfg: Cli = confy::load_path(config)
            .with_context(|| format!("Failed to load configuration file: {}", config))?;
        cfg = Cli::merge(cfg, file_cfg);
    }

    if let (Some(path), Some(verbosity)) = (cfg.logpath.clone(), cfg.verbosity.clone()) {
        let file_appender = tracing_appender::rolling::hourly(path, "mongo-schema-builder.log");
        let (non_blocking, _guard) = tracing_appender::non_blocking(file_appender);
        tracing_subscriber::fmt()
            .with_ansi(false)
            .with_max_level(verbosity)
            .with_writer(non_blocking)
            .init();
        run_with_config(cfg).await
    } else {
        run_with_config(cfg).await
    }
}

#[instrument(skip_all)]
async fn run_with_config(cfg: Cli) -> Result<()> {
    event!(Level::INFO, "{:?}", cfg);

    // Create MongoDB client
    let mut client_options =
        get_opts(cfg.uri.unwrap().as_str(), cfg.resolver.map(|r| r.into())).await?;

    client_options.app_name = Some(DEFAULT_APP_NAME.clone());

    // Create the client credential if the username and password was not in the URI.
    if needs_auth(&client_options) {
        let password: Option<String>;
        match (&cfg.username, &cfg.password) {
            (None, _) => {
                println!("No username provided for authentication with URI");
                process::exit(1);
            }
            (Some(username), None) => {
                if !username.is_empty() {
                    password = Some(
                        Password::new()
                            .with_prompt("Enter password:")
                            .interact()
                            .expect("Failed to read password"),
                    );
                } else {
                    println!("Username is required for authentication with URI");
                    process::exit(1);
                }
            }
            (Some(_), Some(_)) => {
                password = cfg.password;
            }
        }
        load_password_auth(&mut client_options, cfg.username, password).await;
    }

    let mdb_client =
        Client::with_options(client_options).with_context(|| "Failed to create MongoDB client.")?;

    // Create necessary channels for communication
    let (tx_notifications, mut rx_notifications) = unbounded_channel::<SamplerNotification>();

    let (tx_schemata, mut rx_schemata) = unbounded_channel::<SchemaResult>();

    let pb = ProgressBar::new(1024);
    // spinner style errors are caught at compile time so are safe to unwrap on
    let spinner_style = ProgressStyle::with_template("{prefix:.bold.dim} {spinner} {wide_msg}")
        .unwrap()
        .tick_chars("⠁⠂⠄⡀⢀⠠⠐⠈ ");
    pb.set_style(spinner_style);

    // Keeps track of namespaces with modified or created schemas.
    let mut modified_namespaces: HashMap<String, Vec<String>> = HashMap::new();

    // The `Cli.schema_action` is handled in the schema_builder_library based on the `BuilderOptions.schema_collection` field.
    // When the value is Some(<name>), the schema_builder_library does the Merge action, and when it is None, the library
    // does the Overwrite action.
    // Since Merge is the default action, if `Cli.schema_action` is None, Merge will be assumed.
    let schema_collection = if cfg.schema_action.is_some_and(|action| action == Overwrite) {
        None
    } else {
        Some(SCHEMA_COLLECTION_NAME.to_string())
    };

    let builder_options_client = mdb_client.clone();
    tokio::spawn(async move {
        // Create the BuilderOptions
        let builder_options = BuilderOptions {
            include_list: cfg.ns_include.unwrap_or_default(),
            exclude_list: cfg.ns_exclude.unwrap_or_default(),
            schema_collection,
            dry_run: cfg.dry_run,
            client: builder_options_client,
            tx_notifications,
            tx_schemata,
        };
        // Call schema-builder-library
        build_schema(builder_options).await;
    });

    // Keep checking the channels repeatedly until the schema channel closes.
    loop {
        tokio::select! {
                notification = rx_notifications.recv() => {
                    if let Some(notification) = notification {
                        match notification.action {
                            // If we receive an Error notification, we abort the program.
                            SamplerAction::Error { message } => anyhow::bail!(message),
                            // All other notification types are simply logged, depending on the
                            // value of quiet.
                            _ => {
                                if !cfg.quiet {
                                    pb.set_message(notification.to_string());
                                }
                            }
                        }
                    }
                }
                schema = rx_schemata.recv() => {
                    match schema {
                        Some(SchemaResult::FullSchema(schema_res)) =>{
                        let collection = mdb_client.database(&schema_res.namespace_info.db_name).collection::<SchemaDocument>(SCHEMA_COLLECTION_NAME);

                        let filter_doc = doc! {
                            "_id": schema_res.namespace_info.coll_or_view_name.clone(),
                        };

                        let namespace_json_schema: json_schema::Schema = schema_res.namespace_schema.try_into().map_err(|_| JsonSchemaFailure)?;
                        let namespace_bson_schema = bson::to_bson(&namespace_json_schema).map_err(|_| BsonFailure)?;

                        let update_doc = doc! {
                            "_id": schema_res.namespace_info.coll_or_view_name.clone(),
                            "type": schema_res.namespace_info.namespace_type.to_string(),
                            "schema": namespace_bson_schema,
                            "lastUpdated": datetime::DateTime::now(),
                        };

                        let update_options = mongodb::options::UpdateOptions::builder().upsert(true).build();

                        let update_result = collection.update_one(filter_doc, update_doc).with_options(update_options).await;

                        let schema_builder_result: SchemaBuilderResult =
                        match update_result {
                            Ok(result) => match (result.matched_count, result.modified_count, result.upserted_id){
                                (1,0,None) => Unchanged,
                                (1,1,None) => Modified,
                                (0,0, Some(_)) => Created,
                                (0,0,None) => Error("No write occurred".to_string()),
                                _ => unreachable!()
                            }
                            Err(e) => Error(e.to_string()),
                        };

                        // Keep track of namespaces with created or modified schemas if quiet is false
                        if !cfg.quiet && matches!(schema_builder_result, Created | Modified) {
                            modified_namespaces.entry(schema_res.namespace_info.db_name.clone()).and_modify(|coll_vec| {
                                coll_vec.push(schema_res.namespace_info.coll_or_view_name.clone());
                            }).or_insert(vec![schema_res.namespace_info.coll_or_view_name.clone()]);
                        }

                        if !cfg.quiet {
                            // log the schema_builder_result
                            pb.set_message(format!("Schema builder result for namespace `{}.{}`: {}", schema_res.namespace_info.db_name, schema_res.namespace_info.coll_or_view_name, schema_builder_result));
                        }

                    },
                    Some(SchemaResult::NamespaceOnly(schema_res)) =>{
                        // For dry_run mode
                        if !cfg.quiet && cfg.dry_run {
                            pb.set_message(format!(
                                "Namespace acknowledged in dryRun mode: {}.{} ({:?})",
                                schema_res.db_name,
                                schema_res.coll_or_view_name,
                                schema_res.namespace_type,
                            ))
                        }
                    },
                    None => {
                            // When the channel is closed, terminate the loop.
                            break;
                    }
                }
            }
        }
    }

    if !cfg.quiet {
        pb.finish_with_message("Schema creation has completed successfully.\nNow printing all namespaces with created or modified schemas:");

        for (db, namespace) in modified_namespaces.iter() {
            println!("Database: {}. Namespaces: {:?}.", db, namespace);
        }
    }

    Ok(())
}
