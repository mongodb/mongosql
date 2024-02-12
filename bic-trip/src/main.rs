use clap::Parser;
use dialoguer::Password;

use chrono::Local;
use report::{html::generate_report, log_parser::process_logs};
use std::{
    env,
    fs::{self, File},
    io::{self, Error, ErrorKind, Write},
    path::PathBuf,
    process,
};
mod cli;
use cli::Cli;

fn main() -> io::Result<()> {
    let default_base_dir = env::current_dir()?;
    let mut paths = Vec::new();

    let Cli {
        uri,
        input,
        output,
        quiet: _,
        username,
    } = Cli::parse();

    // User-specified output directory, uses current directory if not specified
    let mut file_path = if let Some(ref output_path) = output {
        PathBuf::from(output_path)
    } else {
        default_base_dir
    };

    let mut _password: Option<String> = None;
    if let Some(ref uri) = uri {
        if !uri.is_empty() {
            // URI is specified, so we expect username/password
            if let Some(username) = username {
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
    }

    let metadata = fs::metadata(&input)?;
    if metadata.is_file() {
        paths.push(input.clone());
    } else if metadata.is_dir() {
        for entry in fs::read_dir(input)? {
            let entry = entry?;
            if entry.path().is_file() {
                paths.push(entry.path().to_string_lossy().into_owned());
            }
        }
    }

    let parse_results = process_logs(&paths).map_err(|e| Error::new(ErrorKind::Other, e))?;
    let report = generate_report(parse_results);
    let date = Local::now().format("%m-%d-%Y_%H%M").to_string();
    file_path.push(format!("Atlas_SQL_Readiness_{date}.html"));
    let mut report_file = File::create(file_path)?;
    report_file.write_all(report.as_bytes())?;
    Ok(())
}
