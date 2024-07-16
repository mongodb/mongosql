// This is the CLI struct to handle the command line arguments
// for the report application
use clap::Parser;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None, arg_required_else_help = true)]
pub struct Cli {
    /// Sets the input file or directory to analyze BIC logs (optional). One of `--input` or `--uri` must be provided, or both.
    #[clap(long, short)]
    pub input: Option<String>,

    /// Sets the output directory (optional). If not specified, the current directory is used.
    #[clap(long, short)]
    pub output: Option<String>,

    /// The Atlas cluster URI to analyze schema (optional). One of `--input` or `--uri` must be provided, or both.
    #[clap(long)]
    pub uri: Option<String>,

    /// Username for authentication (optional). This is required if the username and password is not
    /// provided in the URI.
    #[clap(long, short)]
    pub username: Option<String>,

    /// Enables quiet mode for less output.
    #[clap(long, default_value = "false")]
    pub quiet: bool,

    /// The specified resolver (optional).
    #[clap(long)]
    pub resolver: Option<Resolver>,

    /// A list of namespaces to include (optional). If not provided, all namespaces are included. Glob syntax is supported.
    ///
    /// Example: `--include "test.*"`
    #[clap(long)]
    pub include: Option<Vec<String>>,

    /// A list of namespaces to exclude. If not provided, no namespaces are excluded.
    ///
    /// Example: `--exclude "customerdb.*" --exclude "senstivedb.*"`
    #[clap(long)]
    pub exclude: Option<Vec<String>>,
}

#[derive(clap::ValueEnum, Debug, Clone, PartialEq)]
pub enum Resolver {
    Cloudflare,
    Google,
    Quad9,
}

impl From<Resolver> for mongodb::options::ResolverConfig {
    fn from(val: Resolver) -> Self {
        match val {
            Resolver::Cloudflare => mongodb::options::ResolverConfig::cloudflare(),
            Resolver::Google => mongodb::options::ResolverConfig::google(),
            Resolver::Quad9 => mongodb::options::ResolverConfig::quad9(),
        }
    }
}
