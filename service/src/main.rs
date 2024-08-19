use hyper::Server as HyperServer;
use prometheus::{Encoder, Registry};
use service::metrics::{register_metrics, ErrorInterceptor};
use service::translator::translator_service_server::TranslatorServiceServer;
use service::{PanicHandlingTranslateSqlService, TranslateSqlService};
use tonic::service::Interceptor;
use tonic::transport::Server;
pub mod version;
use hyper::Error as HyperError;
use opentelemetry::global;
use opentelemetry::trace::Span;
use prometheus::TextEncoder;
use service::trace::distributed_tracing::{init_tracer_provider, start_span};
use std::env;
use std::error::Error;
use std::net::AddrParseError;
use std::net::SocketAddr;
use std::sync::Arc;
use thiserror::Error;
use tonic::transport::Error as TonicError;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;
use tracing_subscriber::EnvFilter;

#[derive(Debug, Error)]
pub enum MainError {
    #[error("failed to parse gRPC address: {0}")]
    GrpcAddrParse(AddrParseError),
    #[error("failed to parse metrics address: {0}")]
    MetricsAddrParse(AddrParseError),
    #[error("failed to build reflection service: {0}")]
    ReflectionServiceBuild(#[from] tonic_reflection::server::Error),
    #[error("metrics server error: {0}")]
    MetricsServer(#[from] HyperError),
    #[error("gRPC server error: {0}")]
    GrpcServer(TonicError),
}

mod mongosql_proto {
    pub(crate) const FILE_DESCRIPTOR_SET: &[u8] = include_bytes!("translator_descriptor.bin");
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let tracer_provider = init_tracer_provider().expect("Failed to initialize tracer provider");
    global::set_tracer_provider(tracer_provider.clone());

    let log_level = env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string());

    tracing_subscriber::registry()
        .with(EnvFilter::new(log_level))
        .with(tracing_opentelemetry::layer())
        .init();

    let mut server_start_span = start_span(
        "main".to_string(),
        opentelemetry::trace::SpanKind::Internal,
        &opentelemetry::Context::current(),
    );
    server_start_span.add_event("Starting SQL Translation Service".to_string(), vec![]);

    let grpc_host = env::var("SQL_TRANSLATION_SERVER_HOST").unwrap_or_else(|_| {
        eprintln!("SQL_TRANSLATION_SERVER_HOST not set, using default 0.0.0.0");
        "0.0.0.0".to_string()
    });
    let grpc_port = env::var("SQL_TRANSLATION_SERVER_PORT").unwrap_or_else(|_| {
        eprintln!("SQL_TRANSLATION_SERVER_PORT not set, using default 9001");
        "9001".to_string()
    });
    let grpc_addr: SocketAddr = format!("{}:{}", grpc_host, grpc_port)
        .parse()
        .map_err(|e| Box::new(MainError::GrpcAddrParse(e)) as Box<dyn Error>)?;

    let metrics_host = env::var("METRICS_SERVER_HOST").unwrap_or_else(|_| {
        eprintln!("METRICS_SERVER_HOST not set, using default 0.0.0.0");
        "0.0.0.0".to_string()
    });
    let metrics_port = env::var("METRICS_SERVER_PORT").unwrap_or_else(|_| {
        eprintln!("METRICS_SERVER_PORT not set, using default 9090");
        "9090".to_string()
    });
    let metrics_addr: SocketAddr = format!("{}:{}", metrics_host, metrics_port)
        .parse()
        .map_err(|e| Box::new(MainError::MetricsAddrParse(e)) as Box<dyn Error>)?;

    let registry = Arc::new(Registry::new());
    register_metrics(&registry);

    let translate_sql_service = TranslateSqlService;
    let panic_handling_service = PanicHandlingTranslateSqlService(translate_sql_service);

    let reflection_service = tonic_reflection::server::Builder::configure()
        .register_encoded_file_descriptor_set(mongosql_proto::FILE_DESCRIPTOR_SET)
        .build()
        .map_err(|e| Box::new(MainError::ReflectionServiceBuild(e)) as Box<dyn Error>)?;

    // TODO SQL-2218: Implement Logging
    println!("SQL translation server listening on {}", grpc_addr);
    println!("Metrics server listening on {}", metrics_addr);

    let error_interceptor = ErrorInterceptor::new();
    let server = Server::builder()
        .layer(tonic::service::interceptor(move |req| {
            error_interceptor.clone().call(req)
        }))
        .add_service(TranslatorServiceServer::new(panic_handling_service))
        .add_service(reflection_service);

    // Spawn the metrics server
    tokio::spawn(async move {
        let metrics_service = hyper::service::make_service_fn(move |_conn| {
            let registry = Arc::clone(&registry);
            async move {
                Ok::<_, hyper::Error>(hyper::service::service_fn(move |_req| {
                    let registry = Arc::clone(&registry);
                    async move {
                        let metric_families = registry.gather();
                        let mut buffer = vec![];
                        let encoder = TextEncoder::new();
                        encoder.encode(&metric_families, &mut buffer).unwrap();
                        let response = hyper::Response::builder()
                            .header(hyper::header::CONTENT_TYPE, encoder.format_type())
                            .body(hyper::Body::from(buffer))
                            .unwrap();
                        Ok::<_, hyper::Error>(response)
                    }
                }))
            }
        });

        let metrics_server = HyperServer::bind(&metrics_addr).serve(metrics_service);
        if let Err(e) = metrics_server.await {
            eprintln!("Metrics server error: {}", e);
        }
    });
    server_start_span.add_event("SQL Translation Service Started".to_string(), vec![]);
    server_start_span.end();

    server
        .serve(grpc_addr)
        .await
        .map_err(|e| Box::new(MainError::GrpcServer(e)) as Box<dyn Error>)?;

    global::shutdown_tracer_provider();

    Ok(())
}
