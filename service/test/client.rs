use service::translator::{
    translator_client::TranslatorClient, ExcludeNamespacesOption, GetNamespacesRequest,
    SchemaCheckingMode, TranslateSqlRequest,
};
use std::env;
use std::path::PathBuf;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let hostname = env::var("SERVER_HOSTNAME").unwrap_or("localhost".into());
    let port = env::var("SERVER_PORT").unwrap_or("9001".into());

    let mut client = TranslatorClient::connect(format!("http://{hostname}:{port}")).await?;
    let file_name = "tpch.json";

    // Test translate_sql
    let catalog_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("src")
        .join("catalog")
        .join("catalogs")
        .join(file_name)
        .to_str()
        .unwrap()
        .to_string();
    let catalog_file = service::catalog::parse_catalog_json_file(catalog_path.into())
        .expect("Failed to parse catalog file");
    let catalog_bytes =
        serde_json::to_value(&catalog_file).expect("Failed to serialize catalog to bytes");

    let translate_request = TranslateSqlRequest {
        db: "tpch".to_string(),
        query: "SELECT * FROM customer".to_string(),
        schema_catalog: bson::ser::to_vec(&catalog_bytes).unwrap(),
        exclude_namespaces: ExcludeNamespacesOption::ExcludeNamespaces.into(),
        schema_checking_mode: SchemaCheckingMode::Strict.into(),
    };
    let translate_response = client.translate_sql(translate_request).await?;

    // TODO SQL-2218: Implement Logging
    println!("Translate SQL Response:");
    println!("{:#?}", translate_response);

    // Test get_namespaces
    let namespaces_request = GetNamespacesRequest {
        db: "tpch".to_string(),
        query: "SELECT * FROM customer JOIN orders ON customer.c_custkey = orders.o_custkey"
            .to_string(),
    };
    let namespaces_response = client.get_namespaces(namespaces_request).await?;

    // TODO SQL-2218: Implement Logging
    println!("Get Namespaces Response:");
    println!("{:#?}", namespaces_response);

    Ok(())
}
