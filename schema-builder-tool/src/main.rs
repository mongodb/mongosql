mod cli;
use anyhow::{Context, Result};
use clap::Parser;
use cli::Cli;

#[tokio::main]
async fn main() -> Result<()> {
    let mut cfg = Cli::parse();
    // command line arguments override configuration file
    if let Some(ref config) = cfg.config_file {
        let file_cfg: Cli = confy::load_path(config)
            .with_context(|| format!("Failed to load configuration file: {}", config))?;
        cfg = Cli::merge(cfg, file_cfg);
    }
    run_with_config(cfg)
}

fn run_with_config(cfg: Cli) -> Result<()> {
    // run the application with the configuration
    println!("{:?}", cfg);
    Ok(())
}
