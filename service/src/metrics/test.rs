use crate::metrics::{register_metrics, ErrorInterceptor, SERVER_HANDLED_HISTOGRAM};
use prometheus::Encoder;
use prometheus::{Registry, TextEncoder};
use tonic::service::Interceptor;
use tonic::Code;
use tonic::{Request, Status};

#[test]
fn test_server_handled_histogram() {
    let registry = Registry::new();
    register_metrics(&registry);

    // Record sample data
    let service = "test_service";
    let method = "test_method";
    SERVER_HANDLED_HISTOGRAM
        .with_label_values(&[service, method])
        .observe(0.1);
    SERVER_HANDLED_HISTOGRAM
        .with_label_values(&[service, method])
        .observe(0.5);

    let metric_families = registry.gather();
    let histogram_family = metric_families
        .iter()
        .find(|family| family.get_name() == "grpc_server_handling_seconds")
        .expect("Histogram metric family not found");

    let histogram_metric = histogram_family
        .get_metric()
        .iter()
        .find(|metric| {
            let label_pairs = metric.get_label();
            label_pairs
                .iter()
                .any(|label| label.get_name() == "grpc_service" && label.get_value() == service)
                && label_pairs
                    .iter()
                    .any(|label| label.get_name() == "grpc_method" && label.get_value() == method)
        })
        .expect("Histogram metric not found");

    let histogram_proto = histogram_metric.get_histogram();

    assert_eq!(histogram_proto.get_sample_count(), 2);
    assert_eq!(histogram_proto.get_sample_sum(), 0.6);

    let buckets = histogram_proto.get_bucket();
    // The bucket with upper_bound 0.1 should have 1 item
    assert_eq!(buckets[4].get_cumulative_count(), 1);
    // The bucket with upper_bound 0.5 should have both items
    assert_eq!(buckets[6].get_cumulative_count(), 2);
}

#[test]
fn test_error_interceptor() {
    let registry = Registry::new();
    register_metrics(&registry);
    let interceptor = ErrorInterceptor::new();

    // Record errors
    interceptor.record_error(&Status::new(Code::NotFound, "Not Found"));
    interceptor.record_error(&Status::new(Code::Internal, "Internal Error"));
    interceptor.record_error(&Status::new(Code::NotFound, "Not Found"));

    // Collect the metrics
    let mut buffer = Vec::new();
    let encoder = TextEncoder::new();
    let metric_families = registry.gather();
    encoder.encode(&metric_families, &mut buffer).unwrap();
    let metrics_output = String::from_utf8(buffer).unwrap();

    // Check the error counts
    println!("{:?}", metrics_output);
    assert!(metrics_output
        .contains("grpc_errors_total{code=\"Some requested entity was not found\"} 2"));
    assert!(metrics_output.contains("grpc_errors_total{code=\"Internal error\"} 1"));
}

#[tokio::test]
async fn test_interceptor_in_request() {
    let registry = Registry::new();
    register_metrics(&registry);
    let mut interceptor = ErrorInterceptor::new();
    let request = Request::new(());

    // Call the interceptor
    let result = interceptor.call(request);

    // Check if the interceptor is added to the request extensions
    assert!(result.is_ok());
    let request = result.unwrap();
    let extensions = request.extensions();
    assert!(extensions.get::<ErrorInterceptor>().is_some());
}
