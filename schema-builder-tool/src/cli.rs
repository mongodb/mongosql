use clap::Parser;
use serde::{Deserialize, Serialize};
use tracing_subscriber::filter::LevelFilter;

#[derive(Parser, Debug, Default, Serialize, Deserialize, Clone)]
#[command(author, version, about, long_about = None, arg_required_else_help = true)]
pub struct Cli {
    /// The path to the configuration file (optional). If not provided, you must specify --uri to connect to a MongoDB cluster.
    ///
    /// Command line arguments specified in addition to a configuration file will take precedence over configuration file values.
    #[clap(long = "file", short = 'f')]
    pub config_file: Option<String>,

    /// The Atlas cluster URI (optional). If not provided, you must specify --file to load a configuration file with the URI.
    #[clap(long)]
    pub uri: Option<String>,

    /// Username for authentication (optional). If not provided, you must specify --file to load a configuration file with the username.
    #[clap(long, short)]
    pub username: Option<String>,

    /// Password for authentication (optional). If not provided by command line or in a configuration file, you will be prompted for password (recommended).
    ///
    /// You can also specify username and password, or another authentication method, in your MongoDB URI.
    #[clap(long, short)]
    pub password: Option<String>,

    /// The databases and collections to include (optional). If not provided, all databases and collections are included.
    ///
    /// Glob syntax may be used (i.e. mydb.*).
    #[clap(long = "nsInclude")]
    pub ns_include: Option<Vec<String>>,

    /// The databases and collections to exclude (optional). If not provided, no databases or collections are excluded.
    ///
    /// Glob syntax may be used (i.e. mydb.*). This option takes precedence --include.
    #[clap(long = "nsExclude")]
    pub ns_exclude: Option<Vec<String>>,

    /// Enables quiet mode for less output.
    ///
    /// Default: false
    #[clap(long)]
    pub quiet: bool,

    /// Output path to write a log file. If not provided, logging will be ignored.
    #[clap(long, short = 'o')]
    pub logpath: Option<String>,

    /// The logging level to capture in the log file (optional). Requires logpath to be set.
    #[clap(value_enum, long, short = 'v', default_value = "warn")]
    pub verbosity: Option<Verbosity>,

    /// Specifies the action to take if a schema already exists.
    #[clap(value_enum, long = "action", short = 'a', default_value = "merge")]
    pub schema_action: Option<SchemaAction>,

    /// Perform a dry run without analyzing schema or writing to the database. Useful for testing nsInclude and nsExclude.
    ///
    /// Default: false
    #[clap(long = "dryRun")]
    pub dry_run: bool,

    #[clap(long)]
    /// Resolver option for DNS resolution (optional). Specify a resolver if DNS resolution fails or takes too long.
    pub resolver: Option<Resolver>,
}

#[derive(clap::ValueEnum, Debug, Serialize, Deserialize, Clone, PartialEq)]
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

#[derive(clap::ValueEnum, Debug, Serialize, Deserialize, Clone, PartialEq)]
pub enum Verbosity {
    Trace,
    Debug,
    Info,
    Warn,
    Error,
}

impl From<Verbosity> for LevelFilter {
    fn from(verbosity: Verbosity) -> Self {
        match verbosity {
            Verbosity::Trace => LevelFilter::TRACE,
            Verbosity::Debug => LevelFilter::DEBUG,
            Verbosity::Info => LevelFilter::DEBUG,
            Verbosity::Warn => LevelFilter::DEBUG,
            Verbosity::Error => LevelFilter::ERROR,
        }
    }
}

#[derive(clap::ValueEnum, Debug, Serialize, Deserialize, Clone, PartialEq)]
pub enum SchemaAction {
    /// Ignores and overwrites the existing schema. If no schema exists, this option will create a new schema.
    Overwrite,
    /// Merge the existing schema with the new schema. If no schema exists, this option will create a new schema.
    Merge,
}

impl Cli {
    /// Merge two Cli structs together, preferring left values over right values.
    ///
    /// For optional fields, we want to take the value that is specified. If both are specified,
    /// we want to take the left value.
    ///
    /// For boolean fields, we want to take the truthy value if it is specified.
    /// This is because not specifying a boolean field in the CLI or via the
    /// config file, the value will appear as false. Omission may or may not be
    /// intentional, so we want to take the truthy value if it is specified somewhere.
    pub fn merge(left: Cli, right: Cli) -> Cli {
        Cli {
            config_file: left.config_file.or(right.config_file),
            uri: left.uri.or(right.uri),
            username: left.username.or(right.username),
            password: left.password.or(right.password),
            ns_include: left.ns_include.or(right.ns_include),
            ns_exclude: left.ns_exclude.or(right.ns_exclude),
            quiet: left.quiet || right.quiet,
            logpath: left.logpath.or(right.logpath),
            verbosity: left.verbosity.or(right.verbosity),
            schema_action: left.schema_action.or(right.schema_action),
            dry_run: left.dry_run || right.dry_run,
            resolver: left.resolver.or(right.resolver),
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_merge() {
        let left = Cli {
            config_file: Some("foobar.txt".to_string()),
            uri: Some("mongodb://localhost:27017".to_string()),
            username: Some("admin".to_string()),
            password: Some("password".to_string()),
            ns_include: Some(vec!["testa.*".to_string()]),
            ns_exclude: None,
            quiet: true,
            logpath: None,
            verbosity: Some(Verbosity::Debug),
            schema_action: Some(SchemaAction::Merge),
            dry_run: false,
            resolver: None,
        };
        let right = Cli {
            config_file: None,
            uri: Some("mongodb://localhost:27018".to_string()),
            username: Some("AzureDiamond".to_string()),
            password: Some("hunter2".to_string()),
            ns_include: None,
            ns_exclude: Some(vec!["testb.excluded".to_string()]),
            quiet: false,
            logpath: None,
            verbosity: None,
            schema_action: None,
            dry_run: true,
            resolver: Some(Resolver::Google),
        };
        let cli = Cli::merge(left, right);
        assert_eq!(cli.config_file, Some("foobar.txt".to_string()));
        assert_eq!(cli.uri, Some("mongodb://localhost:27017".to_string()));
        assert_eq!(cli.username, Some("admin".to_string()));
        assert_eq!(cli.password, Some("password".to_string()));
        assert_eq!(cli.ns_include, Some(vec!["testa.*".to_string()]));
        assert_eq!(cli.ns_exclude, Some(vec!["testb.excluded".to_string()]));
        assert!(cli.quiet);
        assert!(cli.logpath.is_none());
        assert_eq!(cli.verbosity, Some(Verbosity::Debug));
        assert_eq!(cli.schema_action, Some(SchemaAction::Merge));
        assert!(cli.dry_run);
        assert_eq!(cli.resolver, Some(Resolver::Google));
    }
}
