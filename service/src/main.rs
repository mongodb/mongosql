use hyper::Server as HyperServer;
use prometheus::{Encoder, Registry};
use service::metrics::{register_metrics, ErrorInterceptor};
use service::translator::translator_service_server::TranslatorServiceServer;
use service::{PanicHandlingTranslateSqlService, TranslateSqlService};
use tonic::service::Interceptor;
use tonic::transport::Server;
pub mod version;
use hyper::Error as HyperError;
use log::{debug, error, info};
use opentelemetry::global;
use opentelemetry::trace::Span;
use prometheus::TextEncoder;
use service::logger::init_logger;
use service::trace::distributed_tracing::{init_tracer_provider, start_span};
use std::env;
use std::error::Error;
use std::net::AddrParseError;
use std::net::SocketAddr;
use std::sync::Arc;
use thiserror::Error;
use tokio::sync::broadcast;
use tokio::task::JoinHandle;
use tonic::transport::Error as TonicError;

#[derive(Debug, Error)]
pub enum MainError {
    #[error("failed to parse address: {0}")]
    AddrParse(AddrParseError),
    #[error("failed to build reflection service: {0}")]
    ReflectionServiceBuild(#[from] tonic_reflection::server::Error),
    #[error("metrics server error: {0}")]
    MetricsServer(#[from] HyperError),
    #[error("gRPC server error: {0}")]
    GrpcServer(TonicError),
    #[error("server shutdown error: {0}")]
    ServerShutdown(String),
}

mod mongosql_proto {
    pub(crate) const FILE_DESCRIPTOR_SET: &[u8] = include_bytes!("translator_descriptor.bin");
}

// Helper function to get environment variable with a default
fn get_env_var(key: &str, default: &str) -> String {
    env::var(key).unwrap_or_else(|_| {
        debug!("{} not set, using default {}", key, default);
        default.to_string()
    })
}

// Helper function to parse addresses
fn parse_addr(
    host_env: &str,
    port_env: &str,
    port_default: &str,
) -> Result<SocketAddr, AddrParseError> {
    let host = get_env_var(host_env, "0.0.0.0");
    let port = get_env_var(port_env, port_default);
    format!("{}:{}", host, port).parse()
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    init_logger();
    info!("Starting the SQL translation server");

    let tracer_provider = init_tracer_provider().expect("Failed to initialize tracer provider");
    global::set_tracer_provider(tracer_provider.clone());

    let mut server_start_span = start_span(
        "main".to_string(),
        opentelemetry::trace::SpanKind::Internal,
        &opentelemetry::Context::current(),
    );
    server_start_span.add_event("Starting SQL Translation Service".to_string(), vec![]);

    let grpc_addr = parse_addr(
        "SQL_TRANSLATION_SERVER_HOST",
        "SQL_TRANSLATION_SERVER_PORT",
        "9001",
    )
    .map_err(MainError::AddrParse)?;
    let metrics_addr = parse_addr("METRICS_SERVER_HOST", "METRICS_SERVER_PORT", "9090")
        .map_err(MainError::AddrParse)?;

    let registry = Arc::new(Registry::new());
    register_metrics(&registry);

    let panic_handling_service = PanicHandlingTranslateSqlService(TranslateSqlService);

    let reflection_service = tonic_reflection::server::Builder::configure()
        .register_encoded_file_descriptor_set(mongosql_proto::FILE_DESCRIPTOR_SET)
        .build()
        .map_err(|e| Box::new(MainError::ReflectionServiceBuild(e)) as Box<dyn Error>)?;

    info!("SQL translation server listening on {}", grpc_addr);
    info!("Metrics server listening on {}", metrics_addr);

    let error_interceptor = ErrorInterceptor::new();
    let server = Server::builder()
        .layer(tonic::service::interceptor(move |req| {
            error_interceptor.clone().call(req)
        }))
        .add_service(TranslatorServiceServer::new(panic_handling_service))
        .add_service(reflection_service);

    // Channel for coordinating shutdown
    let (shutdown_tx, _) = broadcast::channel::<()>(1);
    let shutdown_tx_metrics = shutdown_tx.clone();
    let shutdown_tx_grpc = shutdown_tx.clone();

    // Spawn the metrics server
    let metrics_handle: JoinHandle<Result<(), MainError>> = tokio::spawn(async move {
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
        let mut shutdown_rx = shutdown_tx_metrics.subscribe();

        tokio::select! {
            result = metrics_server => {
                if let Err(e) = result {
                    error!("Metrics server error: {}", e);
                    let _ = shutdown_tx_metrics.send(());
                    Err(MainError::MetricsServer(e))
                } else {
                    Ok(())
                }
            }
            _ = shutdown_rx.recv() => {
                info!("Shutting down metrics server");
                Ok(())
            }
        }
    });
    server_start_span.add_event("SQL Translation Service Started".to_string(), vec![]);
    server_start_span.end();

    let grpc_handle: JoinHandle<Result<(), MainError>> = tokio::spawn(async move {
        let mut shutdown_rx = shutdown_tx_grpc.subscribe();

        tokio::select! {
            result = server.serve(grpc_addr) => {
                if let Err(e) = result {
                    error!("gRPC server error: {}", e);
                    let _ = shutdown_tx_grpc.send(());
                    Err(MainError::GrpcServer(e))
                } else {
                    Ok(())
                }
            }
            _ = shutdown_rx.recv() => {
                info!("Shutting down gRPC server");
                Ok(())
            }
        }
    });

    // Wait for both servers to complete
    let (metrics_result, grpc_result) = tokio::join!(metrics_handle, grpc_handle);

    // Check results and propagate errors
    metrics_result??;
    grpc_result??;
    global::shutdown_tracer_provider();

    info!("All servers shut down successfully");
    Ok(())
}
