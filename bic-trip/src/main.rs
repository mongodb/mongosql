use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use report::{
    csv::generate_csv, html::generate_html, log_parser::process_logs, schema::process_schemata,
};
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
    let mut analysis = None;

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

            let schemata = sampler::sample(options).await?;
            analysis = Some(process_schemata(schemata));
        }
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
        &analysis,
        REPORT_FILE_STEM,
        REPORT_NAME,
    )?;
    generate_csv(
        &file_path,
        &date,
        &parse_results,
        &analysis,
        REPORT_FILE_STEM,
    )
}
