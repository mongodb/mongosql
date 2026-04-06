#[cfg(feature = "integration")]
#[tokio::test]
async fn enabled() {
    // Create a mongodb client
    let uri = "mongodb://localhost:27017/";
    let client = mongodb::Client::with_uri_str(uri).await.unwrap();

    // Create schema builder options with dry_run set to true.
    let options = crate::options::BuilderOptions {
        include_list: vec![],
        exclude_list: vec![],
        schema_collection: None,
        dry_run: true,
        client,
        task_semaphore: std::sync::Arc::new(tokio::sync::Semaphore::new(10)),
    };

    // Call build_schema with the options. This should not build any schemas because it's a dry-run, but should still return successfully.
    let schemas = crate::build_schema(options)
        .await
        .expect("failed to build schemas")
        .into_inner();

    // Assert that the built schemas are empty.
    assert!(
        schemas.is_empty(),
        "full schemas were built in dry run mode: {} built",
        schemas.len()
    );
}
