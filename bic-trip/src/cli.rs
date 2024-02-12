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

    /// Option to omit queries in report
    #[clap(long, short)]
    pub quiet: Option<bool>,

    /// The Atlas cluster URI
    #[clap(long, requires_all=["username"])]
    pub uri: Option<String>,

    /// Username for authentication
    #[clap(long, short)]
    pub username: Option<String>,
}
