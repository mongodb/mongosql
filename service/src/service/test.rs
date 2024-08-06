#[cfg(test)]
mod tests {
    use crate::catalog::parse_catalog_json_file;
    use crate::translator::translator_service_server::TranslatorService;
    use crate::translator::{GetNamespacesRequest, TranslateSqlRequest};
    use crate::{PanicHandlingTranslateSqlService, TranslateSqlService};
    use bson::ser::to_vec;
    use std::path::PathBuf;
    use tonic::Request;
    static EMPTY_SCHEMA_CATALOG: Vec<u8> = vec![];

    fn get_catalog_bytes() -> Vec<u8> {
        let catalog_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
            .join("src")
            .join("catalog")
            .join("catalogs")
            .join("tpch.json")
            .to_str()
            .unwrap()
            .to_string();
        let catalog_file =
            parse_catalog_json_file(catalog_path.into()).expect("Failed to parse catalog file");
        let catalog_bytes =
            serde_json::to_value(&catalog_file).expect("Failed to serialize catalog to bytes");
        to_vec(&catalog_bytes).unwrap()
    }

    #[tokio::test]
    async fn test_translate_sql_success() {
        let service = TranslateSqlService;
        let request = Request::new(TranslateSqlRequest {
            db: "tpch".to_string(),
            query: "SELECT * FROM customer".to_string(),
            schema_catalog: get_catalog_bytes(),
            exclude_namespaces: 0,
            schema_checking_mode: 0,
        });

        let response = service.translate_sql(request).await;
        assert!(response.is_ok());
    }

    #[tokio::test]
    async fn test_translate_sql_panic() {
        let service = PanicHandlingTranslateSqlService(TranslateSqlService);
        let request = Request::new(TranslateSqlRequest {
            db: "db".to_string(),
            query: "TRIGGER_PANIC".to_string(),
            schema_catalog: get_catalog_bytes(),
            exclude_namespaces: 0,
            schema_checking_mode: 0,
        });

        let response = service.translate_sql(request).await;
        assert!(response.is_err());
        assert_eq!(response.err().unwrap().code(), tonic::Code::Internal);
    }

    #[tokio::test]
    async fn test_translate_sql_empty_catalog_is_error() {
        let service = TranslateSqlService;
        let request = Request::new(TranslateSqlRequest {
            db: "db".to_string(),
            query: "select 1".to_string(),
            schema_catalog: EMPTY_SCHEMA_CATALOG.clone(),
            exclude_namespaces: 0,
            schema_checking_mode: 0,
        });

        let response = service.translate_sql(request).await;
        assert!(response.is_err());
        assert_eq!(response.err().unwrap().code(), tonic::Code::InvalidArgument);
    }

    #[tokio::test]
    async fn test_get_namespaces_success() {
        let service = TranslateSqlService;
        let request = Request::new(GetNamespacesRequest {
            db: "db".to_string(),
            query: "SELECT * FROM table".to_string(),
        });

        let response = service.get_namespaces(request).await;
        assert!(response.is_ok());
    }
}
