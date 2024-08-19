use service::translator::{
    translator_service_client::TranslatorServiceClient, ExcludeNamespacesOption,
    GetNamespacesRequest, SchemaCheckingMode, TranslateSqlRequest,
};
use std::env;
use std::path::PathBuf;

static TRANSLATE_SQL_OP: &str = "translate_sql";
static GET_NAMESPACES_OP: &str = "get_namespaces";

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();
    let operation = args.get(1).map(|s| s.as_str());
    let hostname = env::var("SERVER_HOSTNAME").unwrap_or("localhost".into());
    let port = env::var("SERVER_PORT").unwrap_or("9001".into());

    let mut client = TranslatorServiceClient::connect(format!("http://{hostname}:{port}")).await?;
    match operation {
        Some(op) if op == TRANSLATE_SQL_OP => run_translate_sql(&mut client).await?,
        Some(op) if op == GET_NAMESPACES_OP => run_get_namespaces(&mut client).await?,
        None => {
            // Run both by default
            run_translate_sql(&mut client).await?;
            run_get_namespaces(&mut client).await?;
        }
        _ => println!(
            "Invalid operation. Use '{}', '{}', or no argument to run both.",
            TRANSLATE_SQL_OP, GET_NAMESPACES_OP
        ),
    }
    Ok(())
}

async fn run_translate_sql(
    client: &mut TranslatorServiceClient<tonic::transport::Channel>,
) -> Result<(), Box<dyn std::error::Error>> {
    println!("Running Translate SQL operation:");
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
        exclude_namespaces: ExcludeNamespacesOption::IncludeNamespaces as i32,
        schema_checking_mode: SchemaCheckingMode::Relaxed as i32,
    };
    let translate_response = client.translate_sql(translate_request).await?;

    // TODO SQL-2218: Implement Logging
    println!("Translate SQL Response:");
    println!("{:#?}", translate_response);
    Ok(())
}

async fn run_get_namespaces(
    client: &mut TranslatorServiceClient<tonic::transport::Channel>,
) -> Result<(), Box<dyn std::error::Error>> {
    println!("Running Get Namespaces operation:");
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
