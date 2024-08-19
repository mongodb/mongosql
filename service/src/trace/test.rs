use crate::trace::distributed_tracing::{
    add_event, init_tracer_provider, start_span, COLLECTOR_ENDPOINT, SQL_SERVICE_NAME,
};
use opentelemetry::global;
use opentelemetry::trace::Span;
use opentelemetry_sdk::testing::trace::InMemorySpanExporter;
use opentelemetry_sdk::trace::TracerProvider;
use std::env;
use std::sync::LazyLock;
use tokio::sync::Mutex;

static TRACE_TEST_MUTEX: LazyLock<Mutex<()>> = LazyLock::new(|| Mutex::new(()));

// This test verifies the creation of a span with events and checks the exported span details.
// It ensures that span creation, event addition, and service name are correctly implemented.
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
async fn test_start_span_and_events() {
    let _guard = TRACE_TEST_MUTEX.lock().await;

    let span_exporter = InMemorySpanExporter::default();
    let tracer_provider = TracerProvider::builder()
        .with_simple_exporter(span_exporter.clone())
        .build();
    global::set_tracer_provider(tracer_provider);
    let parent_cx = opentelemetry::Context::current();

    let span_name = "test_span";
    let mut span = start_span(
        span_name.to_string(),
        opentelemetry::trace::SpanKind::Internal,
        &parent_cx,
    );

    add_event(&mut span, "Event 1");
    add_event(&mut span, "Event 2");

    span.end();

    // Retrieve and verify the exported spans from InMemorySpanExporter
    let spans = span_exporter
        .get_finished_spans()
        .expect("Failed to get finished spans");
    assert_eq!(spans.len(), 1, "Expected one exported span");

    let test_span = &spans[0];
    let expected_span_name = format!("{}.{}", SQL_SERVICE_NAME, span_name);
    assert_eq!(test_span.name, expected_span_name, "Span name should match");

    // Check that events were added
    let events: Vec<_> = test_span.events.iter().map(|e| e.name.as_ref()).collect();
    assert_eq!(
        events,
        vec!["Event 1", "Event 2"],
        "Events should match what was added"
    );

    // Confirm Name is "SQLTranslationService"
    assert_eq!(
        test_span.instrumentation_lib.name.as_ref(),
        SQL_SERVICE_NAME.to_string()
    );
    global::shutdown_tracer_provider();
}

// Checks that the tracer provider initialized is "SimpleSpanProcessor" when COLLECTOR_ENDPOINT is not set.
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
async fn test_init_tracer_provider_without_endpoint_var() {
    let _guard = TRACE_TEST_MUTEX.lock().await;
    env::remove_var(COLLECTOR_ENDPOINT);
    let provider = init_tracer_provider().expect("Failed to initialize tracer provider");

    assert!(format!("{:?}", provider).contains("SimpleSpanProcessor"));
    assert!(format!("{:?}", provider).contains(SQL_SERVICE_NAME));
    global::shutdown_tracer_provider();
}

// Checks that the tracer provider initialized is "BatchSpanProcessor" when COLLECTOR_ENDPOINT is specified.
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
async fn test_init_tracer_provider_with_endpoint_var() {
    let _guard = TRACE_TEST_MUTEX.lock().await;
    env::set_var(COLLECTOR_ENDPOINT, "http://non-existent-endpoint:4317");
    let provider = init_tracer_provider().expect("Failed to initialize tracer provider");

    assert!(format!("{:?}", provider).contains("BatchSpanProcessor"));
    assert!(format!("{:?}", provider).contains(SQL_SERVICE_NAME));
    global::shutdown_tracer_provider();
}
