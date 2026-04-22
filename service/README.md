# MongoSQL Translation Service

## Updating Protobuf Files

When changes are made to `service/proto/translator.proto`, the associated files will need to be regenerated.

This is done automatically by the [build.rs](./build.rs) script using a vendored `protoc` binary.

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
