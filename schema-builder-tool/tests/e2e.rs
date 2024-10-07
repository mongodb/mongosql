#![cfg(test)]

#[cfg(feature = "integration")]
mod enterprise {
    use assert_cmd::prelude::*;
    use mongodb::{
        bson::{doc, Bson, Document},
        Collection,
    };
    use predicates::prelude::*;
    use std::{cell::LazyCell, process::Command};
    use test_utils::e2e_db_manager::{TestDatabaseManager, ALT_DOC, MERGED_ALT_SCHEMA, SCHEMA};

    #[allow(clippy::declare_interior_mutable_const)]
    const SQL_SCHEMAS_COLL: &str = "__sql_schemas";

    #[allow(clippy::declare_interior_mutable_const)]
    const VALID_MDB_URI: LazyCell<String> = LazyCell::new(|| {
        format!(
            "mongodb://test:test@localhost:{}",
            std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
        )
    });

    #[cfg(test)]
    macro_rules! assert_no_mdb_error {
        ($assertion:expr, $desc:expr) => {
            if let Err(err) = $assertion {
                panic!("failed to execute {}: {err:?}", $desc);
            }
        };
    }

    #[test]
    fn no_args_prints_help() -> Result<(), Box<dyn std::error::Error>> {
        let cmd = Command::cargo_bin("schema-builder-tool")?.output()?;

        cmd.assert().failure().stderr(predicate::str::contains(
            "Usage: schema-builder-tool [OPTIONS]",
        ));

        Ok(())
    }

    #[test]
    fn invalid_config_file_path_is_error() -> Result<(), Box<dyn std::error::Error>> {
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--file")
            .arg("./testfiles/invalid_config.yml")
            .output()?;

        cmd.assert().failure().stderr(predicate::str::contains(
            "Failed to load configuration file",
        ));

        Ok(())
    }

    #[test]
    fn missing_uri_is_error() -> Result<(), Box<dyn std::error::Error>> {
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--username")
            .arg("user")
            .output()?;

        cmd.assert().failure().stderr(predicate::str::contains(
            "No URI provided for MongoDB connection",
        ));

        Ok(())
    }

    #[test]
    fn uri_without_auth_and_no_username_arg_is_error() -> Result<(), Box<dyn std::error::Error>> {
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg("mongodb://localhost:27017")
            .output()?;

        cmd.assert().failure().stderr(predicate::str::contains(
            "No username provided for authentication",
        ));

        Ok(())
    }

    #[test]
    fn incorrect_credentials_is_error() -> Result<(), Box<dyn std::error::Error>> {
        let uri = format!(
            "mongodb://invalid_user:invalid_pass@localhost:{}",
            std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
        );
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(uri)
            .output()?;

        cmd.assert()
            .failure()
            .stderr(predicate::str::contains("Authentication failed."));

        Ok(())
    }

    #[tokio::test]
    async fn invalid_include_is_error() -> Result<(), Box<dyn std::error::Error>> {
        let _ = TestDatabaseManager::new(vec![], vec![], None).await;

        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--ns-include")
            .arg("db.***")
            .output()?;

        cmd.assert().failure().stderr(predicate::str::contains(
            "Error parsing include glob pattern",
        ));

        Ok(())
    }

    #[tokio::test]
    async fn invalid_exclude_is_error() -> Result<(), Box<dyn std::error::Error>> {
        let _ = TestDatabaseManager::new(vec![], vec![], None).await;

        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--ns-exclude")
            .arg("db.***")
            .output()?;

        cmd.assert().failure().stderr(predicate::str::contains(
            "Error parsing exclude glob pattern",
        ));

        Ok(())
    }

    #[tokio::test]
    async fn multiple_dbs_and_collections() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let db_1_name = "test_db1";
        let db_2_name = "test_db2";
        let dbs = vec![db_1_name.to_string(), db_2_name.to_string()];
        let collections = vec!["test_coll1".to_string(), "test_coll2".to_string()];

        let test_db_manager =
            TestDatabaseManager::new(dbs.clone(), collections.clone(), None).await;

        // Run the command.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .output()?;

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {db_1_name}. Namespaces: {collections:?}.\nDatabase: {db_2_name}. Namespaces: {collections:?}.\n",
        )));

        // Assert that the databases have schemas written for each namespace.
        for db_name in &dbs {
            let db = test_db_manager.client.database(db_name);
            let sql_schemas_coll = db.collection::<Document>(SQL_SCHEMAS_COLL);
            for name in &collections {
                assert_no_mdb_error!(
                    assert_single_schema(
                        &sql_schemas_coll,
                        name.clone(),
                        "Collection".to_string(),
                        SCHEMA.clone(),
                    )
                    .await,
                    format!("`find` on `{db_name}.{SQL_SCHEMAS_COLL}` for Collection `{name}`")
                );
            }
        }

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    #[tokio::test]
    async fn include_works_as_expected() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let included_db_name = "test_db1";
        let other_db_name = "test_db2";
        let collections = vec!["test_coll1".to_string(), "test_coll2".to_string()];
        let view_name = "test_view";

        let test_db_manager = TestDatabaseManager::new(
            vec![included_db_name.to_string(), other_db_name.to_string()],
            collections.clone(),
            Some(view_name.to_string()),
        )
        .await;

        // Run the command.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--ns-include")
            .arg(format!("{included_db_name}.*"))
            .output()?;

        // Sort the list of expected namespaces.
        let mut expected_namespaces = collections
            .clone()
            .into_iter()
            .chain(vec![view_name.to_string()])
            .collect::<Vec<_>>();
        expected_namespaces.sort();

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {included_db_name}. Namespaces: {expected_namespaces:?}.\n",
        )));

        // Assert that the included database has schemas written for each namespace.
        let included_db = test_db_manager.client.database(included_db_name);
        let included_sql_schemas_coll = included_db.collection::<Document>(SQL_SCHEMAS_COLL);

        let mut collections = collections
            .into_iter()
            .map(|name| (name, "Collection".to_string()))
            .collect::<Vec<(String, String)>>();
        collections.push((view_name.to_string(), "View".to_string()));

        for (name, typ) in collections {
            assert_no_mdb_error!(
                assert_single_schema(
                    &included_sql_schemas_coll,
                    name.clone(),
                    typ.clone(),
                    SCHEMA.clone()
                )
                .await,
                format!("`find` on `{included_db_name}.{SQL_SCHEMAS_COLL}` for {typ} `{name}`")
            );
        }

        // Assert that the other database does not have any schemas written.
        let other_db = test_db_manager.client.database(other_db_name);
        let other_sql_schemas_coll = other_db.collection::<Document>(SQL_SCHEMAS_COLL);
        assert_no_mdb_error!(
            assert_empty(&other_sql_schemas_coll).await,
            format!("`find` on `{other_db_name}.{SQL_SCHEMAS_COLL}`")
        );

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    #[tokio::test]
    async fn exclude_works_as_expected() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let excluded_db_name = "test_db1";
        let other_db_name = "test_db2";
        let collections = vec!["test_coll1".to_string(), "test_coll2".to_string()];

        let test_db_manager = TestDatabaseManager::new(
            vec![excluded_db_name.to_string(), other_db_name.to_string()],
            collections.clone(),
            None,
        )
        .await;

        // Run the command.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--ns-exclude")
            .arg(format!("{excluded_db_name}.*"))
            .output()?;

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {other_db_name}. Namespaces: {collections:?}.\n",
        )));

        // Assert that the excluded database does not have any schemas written.
        let excluded_db = test_db_manager.client.database(excluded_db_name);
        let excluded_sql_schemas_coll = excluded_db.collection::<Document>(SQL_SCHEMAS_COLL);
        assert_no_mdb_error!(
            assert_empty(&excluded_sql_schemas_coll).await,
            format!("`find` on `{excluded_db_name}.{SQL_SCHEMAS_COLL}`")
        );

        // Assert that the other database has schemas written for each namespace.
        let other_db = test_db_manager.client.database(other_db_name);
        let other_sql_schemas_coll = other_db.collection::<Document>(SQL_SCHEMAS_COLL);
        for name in collections {
            assert_no_mdb_error!(
                assert_single_schema(
                    &other_sql_schemas_coll,
                    name.clone(),
                    "Collection".to_string(),
                    SCHEMA.clone(),
                )
                .await,
                format!(
                    "`find` on `{excluded_db_name}.{SQL_SCHEMAS_COLL}` for Collection `{name}`"
                )
            );
        }

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    #[tokio::test]
    async fn dry_run_works_as_expected() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let db_name = "test_db";
        let coll_name = "test_coll";

        let test_db_manager =
            TestDatabaseManager::new(vec![db_name.to_string()], vec![coll_name.to_string()], None)
                .await;

        // Run the command.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--dry-run")
            .output()?;

        // Assert the command succeeded.
        cmd.assert().success().stdout("");

        // Assert that the database does not have any schemas written.
        let db = test_db_manager.client.database(db_name);
        let sql_schemas_coll = db.collection::<Document>(SQL_SCHEMAS_COLL);
        assert_no_mdb_error!(
            assert_empty(&sql_schemas_coll).await,
            format!("`find` on `{db_name}.{SQL_SCHEMAS_COLL}`")
        );

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    #[tokio::test]
    async fn init_schema_works_as_expected() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let db_name = "test_db";
        let coll_name = "test_coll";

        let test_db_manager =
            TestDatabaseManager::new(vec![db_name.to_string()], vec![coll_name.to_string()], None)
                .await;

        // Run the command once to create __sql_schemas.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .output()?;

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {db_name}. Namespaces: {:?}.\n",
            vec![coll_name]
        )));

        // Assert that the database has the schema written.
        let db = test_db_manager.client.database(db_name);
        let sql_schemas_coll = db.collection::<Document>(SQL_SCHEMAS_COLL);
        assert_no_mdb_error!(
            assert_single_schema(
                &sql_schemas_coll,
                coll_name.to_string(),
                "Collection".to_string(),
                SCHEMA.clone(),
            )
            .await,
            format!("`find` on `{db_name}.{SQL_SCHEMAS_COLL}` for Collection `{coll_name}`")
        );

        // Drop the collection, then recreate with a different doc.
        let coll = db.collection::<Document>(coll_name);
        assert_no_mdb_error!(
            coll.drop().await,
            format!("drop on `{db_name}.{coll_name}`")
        );
        assert_no_mdb_error!(
            coll.insert_one(ALT_DOC.clone()).await,
            format!("insertOne on `{db_name}.{coll_name}`")
        );

        // Rerun the command with explicit "merge" action to ensure we use initial schema.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--action")
            .arg("merge")
            .output()?;

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {db_name}. Namespaces: {:?}.\n",
            vec![coll_name]
        )));

        // Assert that the database has the schema written, and it still
        // includes the old schema info even though that's been dropped.
        assert_no_mdb_error!(
            assert_single_schema(
                &sql_schemas_coll,
                coll_name.to_string(),
                "Collection".to_string(),
                MERGED_ALT_SCHEMA.clone(),
            )
            .await,
            format!("`find` on `{db_name}.{SQL_SCHEMAS_COLL}` for Collection `{coll_name}`")
        );

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    #[tokio::test]
    async fn config_file() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let config_db_name = "config_db";
        let collections = vec!["foo".to_string(), "bar".to_string()];

        let test_db_manager =
            TestDatabaseManager::new(vec![config_db_name.to_string()], collections.clone(), None)
                .await;

        // Run the command.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--file")
            .arg("./testfiles/valid_config.yml")
            .output()?;

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {config_db_name}. Namespaces: {:?}.\n",
            vec!["bar"] // foo is excluded!
        )));

        // Assert that the database has schemas written for each included namespace.
        let config_db = test_db_manager.client.database(config_db_name);
        let config_sql_schemas_coll = config_db.collection::<Document>(SQL_SCHEMAS_COLL);

        for name in collections {
            if name == "foo".to_string() {
                // "foo" was excluded, so we assert it is not present.
                assert_no_mdb_error!(
                    assert_not_found(&config_sql_schemas_coll, name.clone()).await,
                    format!(
                        "`find` on `{config_db_name}.{SQL_SCHEMAS_COLL}` for Collection `{name}`"
                    )
                )
            } else {
                assert_no_mdb_error!(
                    assert_single_schema(
                        &config_sql_schemas_coll,
                        name.clone(),
                        "Collection".to_string(),
                        SCHEMA.clone()
                    )
                    .await,
                    format!(
                        "`find` on `{config_db_name}.{SQL_SCHEMAS_COLL}` for Collection `{name}`"
                    )
                );
            }
        }

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    #[tokio::test]
    async fn cli_opts_take_precedence_over_config_file() -> Result<(), Box<dyn std::error::Error>> {
        // Set up the database.
        let config_included_db_name = "config_db";
        let cli_included_db_name = "cli_db";
        let collections = vec!["foo".to_string(), "bar".to_string()];

        let test_db_manager = TestDatabaseManager::new(
            vec![
                config_included_db_name.to_string(),
                cli_included_db_name.to_string(),
            ],
            collections.clone(),
            None,
        )
        .await;

        // Run the command.
        let cmd = Command::cargo_bin("schema-builder-tool")?
            .arg("--uri")
            .arg(VALID_MDB_URI.clone())
            .arg("--file")
            .arg("./testfiles/valid_config.yml")
            .arg("--ns-include")
            .arg(format!("{cli_included_db_name}.*"))
            .arg("--ns-exclude")
            .arg(format!("{cli_included_db_name}.bar"))
            .output()?;

        // Assert the command succeeded and the stdout output is correct.
        cmd.assert().success().stdout(predicate::eq(format!(
            "Database: {cli_included_db_name}. Namespaces: {:?}.\n",
            vec!["foo"] // bar is excluded!
        )));

        // Assert that the cli-included database has schemas written for each namespace.
        let cli_included_db = test_db_manager.client.database(cli_included_db_name);
        let cli_included_sql_schemas_coll =
            cli_included_db.collection::<Document>(SQL_SCHEMAS_COLL);

        for name in collections {
            if name == "bar".to_string() {
                // "bar" was excluded via the cli argument, so we assert it
                // is not present.
                assert_no_mdb_error!(
                    assert_not_found(&cli_included_sql_schemas_coll, name.clone()).await,
                    format!(
                        "`find` on `{cli_included_db_name}.{SQL_SCHEMAS_COLL}` for Collection `{name}`"
                    )
                )
            } else {
                assert_no_mdb_error!(
                    assert_single_schema(
                        &cli_included_sql_schemas_coll,
                        name.clone(),
                        "Collection".to_string(),
                        SCHEMA.clone()
                    )
                    .await,
                    format!(
                        "`find` on `{cli_included_db_name}.{SQL_SCHEMAS_COLL}` for Collection `{name}`"
                    )
                );
            }
        }

        // Assert that the config-included database does not have any schemas written.
        let config_included_db = test_db_manager.client.database(config_included_db_name);
        let config_included_sql_schemas_coll =
            config_included_db.collection::<Document>(SQL_SCHEMAS_COLL);
        assert_no_mdb_error!(
            assert_empty(&config_included_sql_schemas_coll).await,
            format!("`find` on `{config_included_db_name}.{SQL_SCHEMAS_COLL}`")
        );

        // Clean up the database.
        test_db_manager.cleanup().await;

        Ok(())
    }

    async fn assert_single_schema(
        coll: &Collection<Document>,
        id: String,
        typ: String,
        schema: Document,
    ) -> Result<(), mongodb::error::Error> {
        // Execute a find with a filter using the _id value.
        let mut cursor = coll.find(doc! { "_id": id.clone() }).await?;

        // Assert that there is a result.
        assert!(
            cursor.advance().await?,
            "expected one result but found none"
        );

        // Assert the values of the result.
        let res = cursor.deserialize_current()?;
        assert_eq!(
            Some(Bson::String(id)),
            res.get("_id").cloned(),
            "actual `_id` does not match expected"
        );
        assert_eq!(
            Some(Bson::String(typ)),
            res.get("type").cloned(),
            "actual `type` does not match expected"
        );
        assert_eq!(
            Some(Bson::Document(schema)),
            res.get("schema").cloned(),
            "actual `schema` does not match expected"
        );
        assert!(
            matches!(res.get("lastUpdated").cloned(), Some(Bson::DateTime(_))),
            "actual `lastUpdated` does not match expected pattern"
        );

        // Assert that there is not a second result.
        assert!(
            !cursor.advance().await?,
            "expected one result but found second: {:?}",
            cursor.deserialize_current()
        );

        Ok(())
    }

    async fn assert_not_found(
        coll: &Collection<Document>,
        id: String,
    ) -> Result<(), mongodb::error::Error> {
        // Execute a find with a filter using the _id value.
        let mut cursor = coll.find(doc! { "_id": id.clone() }).await?;

        // Assert that there is no result.
        assert!(
            !cursor.advance().await?,
            "expected no result but found one: {:?}",
            cursor.deserialize_current(),
        );

        Ok(())
    }

    async fn assert_empty(coll: &Collection<Document>) -> Result<(), mongodb::error::Error> {
        // Execute a find with no filter.
        let mut cursor = coll.find(doc! {}).await?;

        // Assert there are no results.
        assert!(
            !cursor.advance().await?,
            "expected no results but found one: {:?}",
            cursor.deserialize_current()
        );

        Ok(())
    }
}
