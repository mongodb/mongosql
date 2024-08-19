# MongoSQL Translation Service

## Updating Protobuf Files

When changes are made to `service/proto/translator.proto`, the associated files will need to be regenerated. 
This process will update the following files:

1. `service/src/translator.rs`
2. `service/src/translator_descriptor.bin`

## Code Generation Process

### Prerequisites
`Protobuf 3` is used, and code is generated with `protoc`, the protobuf compiler.  `protoc` should be installed with
an OS specific package manager such as brew, apt, or scoop. See: https://grpc.io/docs/protoc-installation/ for 
more info.

### Code generation
1. Make necessary changes to `service/proto/translator.proto`.
2. Run the `generate_proto` command:
```
cargo run --bin generate_proto
```
3. Verify that both `service/src/translator.rs` and `service/src/translator_descriptor.bin` have been updated.
4. Commit these changes along with the `translator.proto` changes.

## OpenTelemetry Collector Setup

To test distributed tracing with the OpenTelemetry Collector, 
you need to download a config used in the docker command. You can download the config
[here](https://raw.githubusercontent.com/open-telemetry/opentelemetry-rust/main/opentelemetry-otlp/examples/basic-otlp-http/otel-collector-config.yaml).  

To run the OpenTelemetry collector for tracing:
```
docker run --rm -it -p 4317:4317 -v $(pwd):/cfg otel/opentelemetry-collector:latest --config=/cfg/otel-collector-config.yaml
```

Set endpoint for your tracing exporter:
```
export COLLECTOR_ENDPOINT="http://localhost:4317"
```
If `COLLECTOR_ENDPOINT` is not set, the trace spans will be written to stdout.
