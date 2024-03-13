// This is the CLI struct to handle the command line arguments
// for the report application
use clap::Parser;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
pub struct Cli {
    /// Sets the input file or directory
    #[clap(long, short)]
    pub input: String,

    /// Sets the output directory
    #[clap(long, short)]
    pub output: Option<String>,

    /// The Atlas cluster URI
    #[clap(long)]
    pub uri: Option<String>,

    /// Username for authentication
    #[clap(long, short)]
    pub username: Option<String>,

    /// Enables verbose logging
    #[clap(long)]
    pub verbose: bool,
}
