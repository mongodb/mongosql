use bson;
use serde_json;
use service::catalog;
use service::translator::{
    translator_service_client::TranslatorServiceClient, ExcludeNamespacesOption,
    GetNamespacesRequest, GetNamespacesResponse, SchemaCheckingMode, TranslateSqlRequest,
    TranslateSqlResponse,
};
use std::env;
use std::error::Error;
use std::path::PathBuf;
use tonic::transport::Channel;
use tonic::Status;

fn get_catalog_path(file_name: &str) -> Result<String, Box<dyn Error>> {
    let catalog_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("src")
        .join("catalog")
        .join("catalogs")
        .join(file_name);

    catalog_path
        .to_str()
        .ok_or_else(|| format!("Invalid Unicode in path: {:?}", catalog_path).into())
        .map(String::from)
}

pub fn get_schema(file_name: &str) -> Result<Vec<u8>, Box<dyn Error>> {
    let catalog_path = get_catalog_path(file_name)?;

    let catalog_file = catalog::parse_catalog_json_file(catalog_path.clone().into())
        .expect(format!("Failed to parse catalog file {:?}", catalog_path).as_str());

    let catalog_bytes =
        serde_json::to_value(&catalog_file).expect("Failed to serialize catalog to bytes");

    Ok(bson::ser::to_vec(&catalog_bytes).expect("Failed to serialize catalog to BSON"))
}

async fn get_client() -> Result<TranslatorServiceClient<Channel>, Box<dyn Error>> {
    let hostname =
        env::var("SQL_TRANSLATION_SERVER_HOST").unwrap_or_else(|_| "localhost".to_string());
    let port = env::var("SQL_TRANSLATION_SERVER_PORT").unwrap_or_else(|_| "9001".to_string());
    let client = TranslatorServiceClient::connect(format!("http://{hostname}:{port}")).await?;
    Ok(client)
}

pub async fn run_translate_sql_test(
    db: &str,
    query: &str,
    schema_catalog: Vec<u8>,
    exclude_namespaces: ExcludeNamespacesOption,
    schema_checking_mode: SchemaCheckingMode,
) -> Result<TranslateSqlResponse, Status> {
    let mut client = get_client()
        .await
        .map_err(|e| Status::internal(e.to_string()))?;
    let request = TranslateSqlRequest {
        db: db.to_string(),
        query: query.to_string(),
        schema_catalog,
        exclude_namespaces: exclude_namespaces as i32,
        schema_checking_mode: schema_checking_mode as i32,
    };
    client.translate_sql(request).await.map(|r| r.into_inner())
}

pub async fn run_get_namespaces_test(
    db: &str,
    query: &str,
) -> Result<GetNamespacesResponse, Box<dyn Error>> {
    let mut client = get_client().await?;
    let request = GetNamespacesRequest {
        db: db.to_string(),
        query: query.to_string(),
    };
    let response = client.get_namespaces(request).await?.into_inner();
    Ok(response)
}

pub async fn get_metrics_count(metric_name: &str) -> Result<i64, Box<dyn std::error::Error>> {
    let metrics_host = env::var("METRICS_SERVER_HOST").unwrap_or_else(|_| "localhost".to_string());
    let metrics_port = env::var("METRICS_SERVER_PORT").unwrap_or_else(|_| "9090".to_string());

    let metrics_url = format!("http://{}:{}/metrics", metrics_host, metrics_port);
    let response = reqwest::get(&metrics_url).await?;
    let body = response.text().await?;

    for line in body.lines() {
        if line.starts_with(metric_name) {
            let count = line.split_whitespace().last().unwrap_or("0");
            return Ok(count.parse()?);
        }
    }
    Ok(0) // Return 0 if metric not found
}
