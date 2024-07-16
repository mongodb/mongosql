# BIC Transition Readiness Informational Profiler (BIC TRIP)

The BIC Transition Readiness Informational Profiler is a tool to help assess the readiness of BI Connector (BIC) workloads for transitioning to Atlas SQL. It analyzes BIC logs and performs schema analysis on the source MongoDB cluster to generate a readiness report consisting of an HTML document and CSV outputs.

## Usage

```sh
mongosql-transition-tool [OPTIONS]

Options:
  -i, --input <INPUT>        Sets the input file or directory to analyze BIC logs (optional). One of `--input` or `--uri` must be provided, or both
  -o, --output <OUTPUT>      Sets the output directory (optional). If not specified, the current directory is used
      --uri <URI>            The Atlas cluster URI to analyze schema (optional). One of `--input` or `--uri` must be provided, or both
  -u, --username <USERNAME>  Username for authentication (optional). This is required if the username and password is not provided in the URI
      --quiet                Enables quiet mode for less output
      --resolver <RESOLVER>  The specified resolver (optional) [possible values: cloudflare, google, quad9]
      --include <INCLUDE>    A list of namespaces to include (optional). If not provided, all namespaces are included. Glob syntax is supported
      --exclude <EXCLUDE>    A list of namespaces to exclude. If not provided, no namespaces are excluded
  -h, --help                 Print help (see more with '--help')
  -V, --version              Print version
```

`mongosql-transition-tool` can be run in the following modes:

- Process BIC logs only: Specify the `--input` option with the path to the BIC log file(s)
- Run schema analysis only: Specify the `--uri` option with the URI of the source MongoDB cluster
- Run both log processing and schema analysis: Specify both `--input` and `--uri` options

### Log Analysis

1. Download the BIC logs for the desired timeframe from the Atlas UI.
2. Decompress the log file(s) into a directory. If you have multiple log files, place them all in the same directory.
3. Run with the `--input` option:

   ```sh
   mongosql-transition-tool --input path/to/bic_logs
   ```

4. The readiness report will be generated in the current directory or the directory specified by `--output`.

### Schema Analysis

1. Obtain the connection URI for your source MongoDB cluster from the Atlas UI.
2. Run BIC TRIP with the `--uri` option and provide the username with `--username`:

   ```sh
   mongosql-transition-tool --uri "mongodb+srv://cluster.example.com" --username myuser
   ```

   You will be prompted to enter the password.

## Testing

### Unit Tests

```sh
cargo test --package bic-trip --lib
```

### Integration Tests

```sh
cargo test --package bic-trip --test log_parse_tests
```
