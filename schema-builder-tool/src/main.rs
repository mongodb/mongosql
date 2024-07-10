mod cli;
mod consts;

use anyhow::{Context, Result};
use clap::Parser;
use cli::Cli;
use consts::{DEFAULT_APP_NAME, SCHEMA_COLLECTION_NAME};
use dialoguer::Password;
use indicatif::{ProgressBar, ProgressStyle};
use mongodb::Client;
use schema_builder_library::{
    build_schema,
    client_util::{get_opts, load_password_auth, needs_auth},
    options::BuilderOptions,
    SamplerAction, SamplerNotification, SchemaResult,
};
use std::process;
use tokio::sync::mpsc::unbounded_channel;

#[tokio::main]
async fn main() -> Result<()> {
    let mut cfg = Cli::parse();
    // command line arguments override configuration file
    if let Some(ref config) = cfg.config_file {
        let file_cfg: Cli = confy::load_path(config)
            .with_context(|| format!("Failed to load configuration file: {}", config))?;
        cfg = Cli::merge(cfg, file_cfg);
    }
    run_with_config(cfg).await
}

async fn run_with_config(cfg: Cli) -> Result<()> {
    // Create client
    let mut client_options = get_opts(cfg.uri.unwrap().as_str()).await?;

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

    tokio::spawn(async move {
        // Create the BuilderOptions
        let builder_options = BuilderOptions {
            include_list: cfg.ns_include.unwrap_or_default(),
            exclude_list: cfg.ns_exclude.unwrap_or_default(),
            schema_collection: Some(SCHEMA_COLLECTION_NAME.to_string()),
            dry_run: cfg.dry_run,
            client: mdb_client,
            tx_notifications,
            tx_schemata,
        };
        // Call schema-builder-library
        build_schema(builder_options).await;
    });

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
                if let Some(_schema_res) = schema{

                }
                else{
                    // When the channel is closed, terminate the loop.
                    break;
                }
            }
        }
    }

    Ok(())
}
