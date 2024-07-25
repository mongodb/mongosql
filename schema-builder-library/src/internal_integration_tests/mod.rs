#[cfg(feature = "integration")]
#[cfg(test)]
mod dry_run;

#[cfg(feature = "integration")]
#[cfg(test)]
mod get_bounds;

#[cfg(feature = "integration")]
#[cfg(test)]
mod get_partitions;

#[cfg(feature = "integration")]
#[cfg(test)]
async fn create_mdb_client() -> mongodb::Client {
    let mdb_uri = format!(
        "mongodb://localhost:{}/",
        std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or("27017".to_string())
    );
    mongodb::Client::with_uri_str(mdb_uri).await.unwrap()
}
