use crate::catalog;
use crate::metrics::{
    SERVER_HANDLED_COUNTER, SERVER_HANDLED_HISTOGRAM, SERVER_PANICS_TOTAL, SERVER_STARTED_COUNTER,
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
use std::panic::AssertUnwindSafe;
use std::time::Instant;
use thiserror::Error;

use std::panic;

#[cfg(test)]
fn trigger_panic() {
    panic!("This is a test panic");
}

// A wrapper service that handles panics in the `TranslateSqlService`.  It will catch and handle
// panics that occur during the execution of gRPC methods in the TranslateSqlService.
// When a panic is caught, the service increments the SERVER_PANICS_TOTAL metric and returns an
// internal server error status.
pub struct PanicHandlingTranslateSqlService(pub TranslateSqlService);

const INTERNAL_SERVER_ERROR: &str = "Internal server error";

#[tonic::async_trait]
impl TranslatorService for PanicHandlingTranslateSqlService {
    async fn translate_sql(
        &self,
        request: Request<TranslateSqlRequest>,
    ) -> Result<Response<TranslateSqlResponse>, Status> {
        let fut = self.0.translate_sql(request);

        panic::AssertUnwindSafe(fut)
            .catch_unwind()
            .await
            .unwrap_or_else(|_| {
                SERVER_PANICS_TOTAL.inc();
                Err(Status::internal(INTERNAL_SERVER_ERROR))
            })
    }

    async fn get_namespaces(
        &self,
        request: Request<GetNamespacesRequest>,
    ) -> Result<Response<GetNamespacesResponse>, Status> {
        match panic::catch_unwind(AssertUnwindSafe(|| self.0.get_namespaces(request))) {
            Ok(result) => result.await,
            Err(_) => {
                SERVER_PANICS_TOTAL.inc();
                Err(Status::internal(INTERNAL_SERVER_ERROR))
            }
        }
    }
}

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
        let start = Instant::now();
        SERVER_STARTED_COUNTER
            .with_label_values(&["TranslatorService", "TranslateSql"])
            .inc();

        let req = request.into_inner();

        // TODO SQL-2218: Implement Logging
        println!("Received a request to db: {}", req.db);

        if req.schema_catalog.is_empty() {
            SERVER_HANDLED_COUNTER
                .with_label_values(&["TranslatorService", "TranslateSql", "INVALID_ARGUMENT"])
                .inc();
            return Err(Status::invalid_argument("schema_catalog is empty"));
        }

        // Trigger a panic for testing purposes
        #[cfg(test)]
        if req.query == "TRIGGER_PANIC" {
            trigger_panic();
        }

        let catalog = catalog::build_catalog_from_bytes(&req.schema_catalog)
            .map_err(|e| Error::CatalogBuildError(e.to_string()))?;

        let options = SqlOptions {
            exclude_namespaces: translator::ExcludeNamespacesOption::try_from(
                req.exclude_namespaces,
            )
            .map_err(|_| Error::InvalidExcludeNamespacesOption)
            .map(|option| match option {
                translator::ExcludeNamespacesOption::ExcludeNamespacesUnspecified => {
                    mongosql::options::ExcludeNamespacesOption::ExcludeNamespaces
                }
                translator::ExcludeNamespacesOption::IncludeNamespaces => {
                    mongosql::options::ExcludeNamespacesOption::IncludeNamespaces
                }
            })?,
            schema_checking_mode: SchemaCheckingMode::try_from(req.schema_checking_mode)
                .map_err(|_| Error::InvalidSchemaCheckingMode)
                .map(|mode| match mode {
                    SchemaCheckingMode::StrictUnspecified => mongosql::SchemaCheckingMode::Strict,
                    SchemaCheckingMode::Relaxed => mongosql::SchemaCheckingMode::Relaxed,
                })?,
        };

        let response = mongosql::translate_sql(&req.db, &req.query, &catalog, options)
            .map(Self::create_translate_sql_response)
            .map_err(|e| Error::SqlTranslationError(e.to_string()))?;

        let duration = start.elapsed();
        SERVER_HANDLED_HISTOGRAM
            .with_label_values(&["TranslatorService", "TranslateSql"])
            .observe(duration.as_secs_f64());

        SERVER_HANDLED_COUNTER
            .with_label_values(&["TranslatorService", "TranslateSql", "OK"])
            .inc();

        Ok(Response::new(response))
    }

    async fn get_namespaces(
        &self,
        request: Request<GetNamespacesRequest>,
    ) -> Result<Response<GetNamespacesResponse>, Status> {
        let start = Instant::now();
        SERVER_STARTED_COUNTER
            .with_label_values(&["TranslatorService", "GetNamespaces"])
            .inc();

        let req = request.into_inner();

        let namespaces = mongosql::get_namespaces(&req.db, &req.query)
            .map_err(|e| Error::GetNamespacesError(e.to_string()))?;

        let response = Self::create_get_namespaces_response(namespaces);

        let duration = start.elapsed();
        SERVER_HANDLED_HISTOGRAM
            .with_label_values(&["TranslatorService", "GetNamespaces"])
            .observe(duration.as_secs_f64());

        SERVER_HANDLED_COUNTER
            .with_label_values(&["TranslatorService", "GetNamespaces", "OK"])
            .inc();

        Ok(Response::new(response))
    }
}
