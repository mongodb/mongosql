use crate::catalog;
use crate::metrics::{
    ErrorInterceptor, SERVER_HANDLED_COUNTER, SERVER_HANDLED_HISTOGRAM, SERVER_PANICS_TOTAL,
    SERVER_STARTED_COUNTER,
};
use crate::translator::translator_service_server::TranslatorService;
use crate::translator::{
    self, GetNamespacesRequest, GetNamespacesResponse, Metadata, Namespace, SchemaCheckingMode,
    SelectOrderItem, TranslateSqlRequest, TranslateSqlResponse,
};
use crate::version::VERSION;
use mongosql;
use mongosql::options::SqlOptions;
use serde_json;
use std::collections::BTreeSet;
use tonic::{Request, Response, Status};

use futures::future::FutureExt;
use log::{error, info, warn};
use std::any::Any;
use std::panic::AssertUnwindSafe;
use std::time::Instant;
use thiserror::Error;

use crate::trace::distributed_tracing::{add_event, extract_parent_context, start_span};
use opentelemetry::trace::SpanKind;
use std::panic;

#[derive(Debug, Error)]
pub enum Error {
    #[error("failed to build catalog: {0}")]
    CatalogBuildError(String),
    #[error("invalid exclude_namespaces option")]
    InvalidExcludeNamespacesOption,
    #[error("invalid schema_checking_mode")]
    InvalidSchemaCheckingMode,
    #[error("failed to translate SQL: {0}")]
    SqlTranslationError(String),
    #[error("failed to get namespaces: {0}")]
    GetNamespacesError(String),
}

fn extract_panic_message(panic_info: &Box<dyn Any + Send>) -> String {
    panic_info
        .downcast_ref::<String>()
        .map(|s| s.to_string())
        .or_else(|| panic_info.downcast_ref::<&str>().map(|s| s.to_string()))
        .unwrap_or_else(|| "Unknown panic".to_string())
}

// A wrapper service that handles panics in the `TranslateSqlService`.  It will catch and handle
// panics that occur during the execution of gRPC methods in the TranslateSqlService.
// When a panic is caught, the service increments the SERVER_PANICS_TOTAL metric and returns an
// internal server error status.
pub struct PanicHandlingTranslateSqlService(pub TranslateSqlService);

const INTERNAL_SERVER_ERROR: &str = "Internal server error";
const TRIGGER_PANIC: &str = "__test_panic";

#[tonic::async_trait]
impl TranslatorService for PanicHandlingTranslateSqlService {
    async fn translate_sql(
        &self,
        request: Request<TranslateSqlRequest>,
    ) -> Result<Response<TranslateSqlResponse>, Status> {
        let fut = self.0.translate_sql(request);
        AssertUnwindSafe(fut)
            .catch_unwind()
            .await
            .unwrap_or_else(|panic_info| {
                SERVER_PANICS_TOTAL.inc();
                let panic_message = extract_panic_message(&panic_info);
                error!(
                    "Panic occurred while translating SQL query: {}",
                    panic_message
                );
                Err(Status::internal(format!(
                    "{}: {}",
                    INTERNAL_SERVER_ERROR, panic_message
                )))
            })
    }

    async fn get_namespaces(
        &self,
        request: Request<GetNamespacesRequest>,
    ) -> Result<Response<GetNamespacesResponse>, Status> {
        match panic::catch_unwind(AssertUnwindSafe(|| self.0.get_namespaces(request))) {
            Ok(result) => result.await,
            Err(panic_info) => {
                SERVER_PANICS_TOTAL.inc();
                let panic_message = extract_panic_message(&panic_info);
                error!("Panic occurred while getting namespaces: {}", panic_message);
                Err(Status::internal(format!(
                    "{}: {}",
                    INTERNAL_SERVER_ERROR, panic_message
                )))
            }
        }
    }
}

impl From<Error> for Status {
    fn from(err: Error) -> Self {
        match err {
            Error::CatalogBuildError(e) => Status::internal(e),
            Error::InvalidExcludeNamespacesOption => {
                Status::invalid_argument("Invalid exclude_namespaces option")
            }
            Error::InvalidSchemaCheckingMode => {
                Status::invalid_argument("Invalid schema_checking_mode")
            }
            Error::SqlTranslationError(e) => Status::internal(e),
            Error::GetNamespacesError(e) => Status::internal(e),
        }
    }
}

#[derive(Debug)]
pub struct TranslateSqlService;

impl TranslateSqlService {
    fn create_translate_sql_response(translation: mongosql::Translation) -> TranslateSqlResponse {
        TranslateSqlResponse {
            metadata: Some(Metadata {
                version: VERSION.to_string(),
            }),
            db: translation.target_db,
            target_collection: translation.target_collection.unwrap_or_default(),
            pipeline: translation.pipeline.to_string(),
            result_set_schema: serde_json::to_string(&translation.result_set_schema)
                .unwrap_or_default(),
            select_order: translation
                .select_order
                .into_iter()
                .map(|item| SelectOrderItem {
                    namespace: Some(item.first().cloned().unwrap_or_default()),
                    field_name: item.get(1).cloned().unwrap_or_default(),
                })
                .collect(),
        }
    }

    fn create_get_namespaces_response(
        namespaces: BTreeSet<mongosql::Namespace>,
    ) -> GetNamespacesResponse {
        GetNamespacesResponse {
            metadata: Some(Metadata {
                version: VERSION.to_string(),
            }),
            namespaces: namespaces
                .into_iter()
                .map(|ns| Namespace {
                    db: ns.database,
                    collection: ns.collection,
                })
                .collect(),
        }
    }
}

#[tonic::async_trait]
impl TranslatorService for TranslateSqlService {
    async fn translate_sql(
        &self,
        request: Request<TranslateSqlRequest>,
    ) -> Result<Response<TranslateSqlResponse>, Status> {
        let interceptor = request
            .extensions()
            .get::<ErrorInterceptor>()
            .cloned()
            .unwrap_or_else(ErrorInterceptor::new);

        let metadata = request.metadata().clone();
        let req = request.into_inner();

        let parent_cx = extract_parent_context(&metadata);
        let mut span = start_span("translate_sql".to_string(), SpanKind::Server, &parent_cx);
        let start = Instant::now();
        SERVER_STARTED_COUNTER
            .with_label_values(&["TranslatorService", "TranslateSql"])
            .inc();

        info!("Received a translate_sql() request to db: {}", req.db);
        add_event(&mut span, &format!("Received request for db: {}", req.db));

        if req.schema_catalog.is_empty() {
            let error_message = "schema_catalog is empty";
            add_event(&mut span, &format!("Error: {}", error_message));
            let status = Status::invalid_argument(error_message);
            interceptor.record_error(&status);
            warn!("Invalid argument: {}", status.message());

            SERVER_HANDLED_COUNTER
                .with_label_values(&["TranslatorService", "TranslateSql", "INVALID_ARGUMENT"])
                .inc();
            return Err(status);
        }

        // Trigger a panic for testing purposes
        if req.query == TRIGGER_PANIC && req.db == TRIGGER_PANIC {
            panic!("This is a test panic");
        }

        let catalog = match catalog::build_catalog_from_bytes(&req.schema_catalog) {
            Ok(cat) => cat,
            Err(e) => {
                let error_message = format!("Error building catalog: {}", e);
                add_event(&mut span, &error_message);
                let status = Status::internal(error_message);
                interceptor.record_error(&status);
                return Err(status);
            }
        };

        let options = match self.build_sql_options(&req) {
            Ok(opts) => opts,
            Err(e) => {
                interceptor.record_error(&e);
                add_event(&mut span, &format!("Error building SQL options: {}", e));
                return Err(e);
            }
        };

        let response = match mongosql::translate_sql(&req.db, &req.query, &catalog, options) {
            Ok(translation) => Self::create_translate_sql_response(translation),
            Err(e) => {
                let error_message = format!("Error translating SQL: {}", e);
                add_event(&mut span, &error_message);
                let status = Status::invalid_argument(error_message);
                interceptor.record_error(&status);

                return Err(Error::SqlTranslationError(e.to_string()).into());
            }
        };

        let duration = start.elapsed();
        SERVER_HANDLED_HISTOGRAM
            .with_label_values(&["TranslatorService", "TranslateSql"])
            .observe(duration.as_secs_f64());
        SERVER_HANDLED_COUNTER
            .with_label_values(&["TranslatorService", "TranslateSql", "OK"])
            .inc();
        add_event(&mut span, "SQL translation completed");
        Ok(Response::new(response))
    }

    async fn get_namespaces(
        &self,
        request: Request<GetNamespacesRequest>,
    ) -> Result<Response<GetNamespacesResponse>, Status> {
        let interceptor = request
            .extensions()
            .get::<ErrorInterceptor>()
            .cloned()
            .unwrap_or_else(ErrorInterceptor::new);
        let metadata = request.metadata().clone();
        let req = request.into_inner();

        let parent_cx = extract_parent_context(&metadata);
        let mut span = start_span("get_namespaces".to_string(), SpanKind::Server, &parent_cx);

        let start = Instant::now();
        SERVER_STARTED_COUNTER
            .with_label_values(&["TranslatorService", "GetNamespaces"])
            .inc();

        info!("Received a get_namespaces() request to db: {}", req.db);

        let namespaces = match mongosql::get_namespaces(&req.db, &req.query) {
            Ok(ns) => ns,
            Err(e) => {
                let error_msg = e.to_string();
                add_event(
                    &mut span,
                    &format!("Error getting namespaces: {}", error_msg),
                );
                let error = Error::GetNamespacesError(error_msg.clone());
                let status = Status::internal(format!("Failed to get namespaces: {}", error_msg));
                interceptor.record_error(&status);
                return Err(error.into());
            }
        };

        let response = Self::create_get_namespaces_response(namespaces);

        let duration = start.elapsed();
        SERVER_HANDLED_HISTOGRAM
            .with_label_values(&["TranslatorService", "GetNamespaces"])
            .observe(duration.as_secs_f64());

        SERVER_HANDLED_COUNTER
            .with_label_values(&["TranslatorService", "GetNamespaces", "OK"])
            .inc();

        add_event(&mut span, "Get namespaces completed");

        Ok(Response::new(response))
    }
}

impl TranslateSqlService {
    fn build_sql_options(&self, req: &TranslateSqlRequest) -> Result<SqlOptions, Status> {
        let exclude_namespaces =
            translator::ExcludeNamespacesOption::try_from(req.exclude_namespaces)
                .map_err(|_| Status::invalid_argument("Invalid exclude_namespaces option"))?;

        let schema_checking_mode = SchemaCheckingMode::try_from(req.schema_checking_mode)
            .map_err(|_| Status::invalid_argument("Invalid schema_checking_mode"))?;

        Ok(SqlOptions {
            exclude_namespaces: match exclude_namespaces {
                translator::ExcludeNamespacesOption::ExcludeNamespacesUnspecified => {
                    mongosql::options::ExcludeNamespacesOption::ExcludeNamespaces
                }
                translator::ExcludeNamespacesOption::IncludeNamespaces => {
                    mongosql::options::ExcludeNamespacesOption::IncludeNamespaces
                }
            },
            schema_checking_mode: match schema_checking_mode {
                SchemaCheckingMode::StrictUnspecified => mongosql::SchemaCheckingMode::Strict,
                SchemaCheckingMode::Relaxed => mongosql::SchemaCheckingMode::Relaxed,
            },
            allow_order_by_missing_columns: false,
        })
    }
}
