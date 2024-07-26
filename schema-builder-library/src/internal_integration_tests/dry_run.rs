#[cfg(feature = "integration")]
#[tokio::test]
async fn enabled() {
    // Create a mongodb client
    let uri = "mongodb://localhost:27017/";
    let client = mongodb::Client::with_uri_str(uri).await.unwrap();

    // Create communication channels.
    let (tx_notifications, mut rx_notifications) =
        tokio::sync::mpsc::unbounded_channel::<crate::SamplerNotification>();
    let (tx_schemata, mut rx_schemata) =
        tokio::sync::mpsc::unbounded_channel::<crate::SchemaResult>();

    // Create schema builder options with dry_run set to true.
    let options = crate::options::BuilderOptions {
        include_list: vec![],
        exclude_list: vec![],
        schema_collection: None,
        dry_run: true,
        client,
        tx_notifications,
        tx_schemata,
    };

    // Call build_schema in a separate thread.
    tokio::spawn(crate::build_schema(options));

    // Wait on channels to get results. Assert that received collections are as expected. Fail if
    // we get certain notifications.
    loop {
        tokio::select! {
            notification = rx_notifications.recv() => {
                if let Some(notification) = notification {
                    // If we receive any notification, then dry run functionality is not implemented
                    // correctly. During a dry run, we should not partition, query, sample, or
                    // process any namespaces. Receiving any of those notification types causes this
                    // test to fail. Also, if receiving any errors or warnings causes this test to
                    // fail.
                    panic!("received a notification during a dry run: {notification:?}");
                }
            }
            schema = rx_schemata.recv() => {
                match schema {
                    Some(crate::SchemaResult::FullSchema(schema_res)) => {
                        // If we receive a FullSchema in dry-run mode, then dry run functionality is
                        // not implemented correctly so the test fails.
                        panic!(
                            "full schema built for {:?} {}.{}",
                            schema_res.namespace_info.namespace_type,
                            schema_res.namespace_info.db_name,
                            schema_res.namespace_info.coll_or_view_name,
                        );
                    }
                    Some(crate::SchemaResult::NamespaceOnly(_)) => {}
                    None => break
                }
            }
        }
    }
}
