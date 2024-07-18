use crate::version::VERSION;
use mongosql;
use mongosql::options::SqlOptions;
use serde_json;
use service::translator::translator_service_server::TranslatorService;
use service::translator::{
    self, GetNamespacesRequest, GetNamespacesResponse, Metadata, Namespace, SchemaCheckingMode,
    SelectOrderItem, TranslateSqlRequest, TranslateSqlResponse,
};
use std::collections::BTreeSet;
use tonic::{Request, Response, Status};

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
        let req = request.into_inner();

        // TODO SQL-2218: Implement Logging
        println!("Received a request to db: {}", req.db);

        if req.schema_catalog.is_empty() {
            return Err(Status::invalid_argument("schema_catalog is empty"));
        }

        let catalog = service::catalog::build_catalog_from_bytes(&req.schema_catalog)
            .map_err(|e| Status::internal(format!("Failed to build catalog: {}", e)))?;

        let options = SqlOptions {
            exclude_namespaces: translator::ExcludeNamespacesOption::try_from(
                req.exclude_namespaces,
            )
            .map_err(|_| Status::invalid_argument("Invalid exclude_namespaces option"))
            .map(|option| match option {
                translator::ExcludeNamespacesOption::ExcludeNamespacesUnspecified => {
                    mongosql::options::ExcludeNamespacesOption::ExcludeNamespaces
                }
                translator::ExcludeNamespacesOption::IncludeNamespaces => {
                    mongosql::options::ExcludeNamespacesOption::IncludeNamespaces
                }
            })?,
            schema_checking_mode: SchemaCheckingMode::try_from(req.schema_checking_mode)
                .map_err(|_| Status::invalid_argument("Invalid schema_checking_mode"))
                .map(|mode| match mode {
                    SchemaCheckingMode::StrictUnspecified => mongosql::SchemaCheckingMode::Strict,
                    SchemaCheckingMode::Relaxed => mongosql::SchemaCheckingMode::Relaxed,
                })?,
        };

        let response = match mongosql::translate_sql(&req.db, &req.query, &catalog, options) {
            Ok(translation) => Self::create_translate_sql_response(translation),
            Err(e) => return Err(Status::internal(e.to_string())),
        };

        Ok(Response::new(response))
    }

    async fn get_namespaces(
        &self,
        request: Request<GetNamespacesRequest>,
    ) -> Result<Response<GetNamespacesResponse>, Status> {
        let req = request.into_inner();

        let namespaces = match mongosql::get_namespaces(&req.db, &req.query) {
            Ok(ns) => ns,
            Err(e) => {
                return Err(Status::internal(format!("Failed to get namespaces: {}", e)));
            }
        };

        let response = Self::create_get_namespaces_response(namespaces);

        Ok(Response::new(response))
    }
}
