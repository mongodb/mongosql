# Usage

The general usage of mongodb-schema-manager (given the name for your platform):

mongodb-schema-manager [OPTIONS]

The options are:

Options:
*  -f, --file <CONFIG_FILE>       The path to the configuration file (optional). If not provided, you must specify --uri to connect to a MongoDB cluster
    - the config file must be specified in yaml and specify some or all of the following options, with at least uri
      present. An example of a valid config file can be seen [here](./valid_config.yml)
*  --uri <URI>                The cluster URI (optional). If not provided, you must specify --file to load a configuration file with the URI
*  -u, --username <USERNAME>      Username for authentication (optional). Not all authentication mechanisms require username.
*  -p, --password <PASSWORD>      Password for authentication (optional). Not all authentication mechanisms require password, and it is generally safer to specify in the config file so that it does not appear on task lists
* --ns-include <NS_INCLUDE>  The databases and collections to include (optional). If not provided, all databases and collections are included
* --ns-exclude <NS_EXCLUDE>  The databases and collections to exclude (optional). If not provided, no databases or collections are excluded
* --quiet                    Enables quiet mode for less output
* -o, --logpath <LOGPATH>        Log directory path where to write log files. If not provided, logging will be ignored
*  -v, --verbosity <VERBOSITY>    The logging level to capture in the log file (optional). Requires logpath to be set \[default: warn\] \[possible values: trace, debug, info, warn, error]\
*  -a, --action <SCHEMA_ACTION>   Specifies the action to take if a schema already exists \[default: merge\] \[possible values: overwrite, merge\]
*  --dry-run                  Perform a dry run without analyzing schema or writing to the database. Useful for testing ns-include and ns-exclude
* --resolver <RESOLVER>      Resolver option for DNS resolution (optional). Specify a resolver if DNS resolution fails or takes too long \[possible values: cloudflare, google, quad9\]
*  -h, --help                     Print help (see more with '--help')
*  -V, --version                  Print version
