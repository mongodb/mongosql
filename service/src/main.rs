pub mod server;
pub mod translator;
use tonic::transport::Server;

use server::TranslateSqlService;
use translator::translator_server::TranslatorServer;
pub mod version;

mod mongosql_proto {
    pub(crate) const FILE_DESCRIPTOR_SET: &[u8] = include_bytes!("translator_descriptor.bin");
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let host = std::env::var("SQL_TRANSLATION_SERVER_HOST").unwrap_or("0.0.0.0".to_string());
    let port = std::env::var("SQL_TRANSLATION_SERVER_PORT").unwrap_or("9001".to_string());
    let addr = format!("{}:{}", host, port).parse()?;

    let inventory = TranslateSqlService;

    let reflection_service = tonic_reflection::server::Builder::configure()
        .register_encoded_file_descriptor_set(mongosql_proto::FILE_DESCRIPTOR_SET)
        .build()
        .unwrap();

    // TODO SQL-2218: Implement Logging
    println!("Starting SQL translation server");
    Server::builder()
        .add_service(TranslatorServer::new(inventory))
        .add_service(reflection_service)
        .serve(addr)
        .await?;
    Ok(())
}
