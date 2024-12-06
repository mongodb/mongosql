use opentelemetry::global;
use opentelemetry::propagation::Extractor;
use opentelemetry::trace::TraceError;
use opentelemetry::{trace::Tracer, KeyValue};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::trace::Config;
use opentelemetry_sdk::{runtime, trace as sdktrace, Resource};
use tonic::metadata::MetadataMap;

use std::env;
use std::sync::LazyLock;

pub static SQL_SERVICE_NAME: &str = "SQLTranslationService";
pub static COLLECTOR_ENDPOINT: &str = "COLLECTOR_ENDPOINT";

static RESOURCE: LazyLock<Resource> =
    LazyLock::new(|| Resource::new(vec![KeyValue::new("service.name", SQL_SERVICE_NAME)]));

// Initializes the tracer provider based on the COLLECTOR_ENDPOINT environment variable.
// If COLLECTOR_ENDPOINT is set, it uses an OTLP exporter to send data to a collector at that endpoint.
// Otherwise, it falls back to a stdout exporter for local tracing.
pub fn init_tracer_provider() -> Result<sdktrace::TracerProvider, TraceError> {
    let collector_endpoint = env::var(COLLECTOR_ENDPOINT).ok();

    match collector_endpoint {
        Some(endpoint) => {
            // Use OTLP exporter
            opentelemetry_otlp::new_pipeline()
                .tracing()
                .with_exporter(
                    opentelemetry_otlp::new_exporter()
                        .tonic()
                        .with_endpoint(endpoint),
                )
                .with_trace_config(Config::default().with_resource(RESOURCE.clone()))
                .install_batch(runtime::Tokio)
        }
        None => {
            // Use stdout exporter
            let exporter = opentelemetry_stdout::SpanExporter::default();
            let provider = sdktrace::TracerProvider::builder()
                .with_simple_exporter(exporter)
                .with_config(Config::default().with_resource(RESOURCE.clone()))
                .build();
            global::set_tracer_provider(provider.clone());
            Ok(provider)
        }
    }
}

pub struct MetadataMapExtractor<'a>(pub &'a MetadataMap);

impl Extractor for MetadataMapExtractor<'_> {
    fn get(&self, key: &str) -> Option<&str> {
        self.0.get(key).and_then(|metadata| metadata.to_str().ok())
    }

    fn keys(&self) -> Vec<&str> {
        self.0
            .keys()
            .map(|key| match key {
                tonic::metadata::KeyRef::Ascii(v) => v.as_str(),
                tonic::metadata::KeyRef::Binary(v) => v.as_str(),
            })
            .collect::<Vec<_>>()
    }
}

use opentelemetry::trace::{Span, SpanKind};

pub fn get_tracer() -> impl Tracer {
    global::tracer(SQL_SERVICE_NAME)
}

pub fn start_span(name: String, kind: SpanKind, parent_cx: &opentelemetry::Context) -> impl Span {
    let tracer = get_tracer();
    tracer
        .span_builder(format!("{}.{}", SQL_SERVICE_NAME, name))
        .with_kind(kind)
        .start_with_context(&tracer, parent_cx)
}

pub fn add_event<S: Span>(span: &mut S, event: &str) {
    span.add_event(event.to_string(), vec![]);
}

pub fn extract_parent_context(metadata: &MetadataMap) -> opentelemetry::Context {
    global::get_text_map_propagator(|prop| prop.extract(&MetadataMapExtractor(metadata)))
}
