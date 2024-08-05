use assert_cmd::prelude::*;
use predicates::prelude::*;
use std::{cell::LazyCell, process::Command};
use test_utils::e2e_db_manager::TestDatabaseManager;

const TEST_DBS: LazyCell<Vec<String>> = LazyCell::new(|| {
    vec![
        "trr_cli_test_db1".to_string(),
        "trr_cli_test_db2".to_string(),
    ]
});

const TEST_COLLS: LazyCell<Vec<String>> =
    LazyCell::new(|| vec!["acoll".to_string(), "bcoll".to_string()]);

const TEST_VIEW: LazyCell<Option<String>> = LazyCell::new(|| Some("cview".to_string()));

#[test]
fn no_args_prints_help() -> Result<(), Box<dyn std::error::Error>> {
    let cmd = Command::cargo_bin("mongosql-transition-tool")?.output()?;

    cmd.assert().failure().stderr(predicate::str::contains(
        "Usage: mongosql-transition-tool [OPTIONS]",
    ));

    Ok(())
}

#[test]
fn invalid_input_directory_is_error() -> Result<(), Box<dyn std::error::Error>> {
    let cmd = Command::cargo_bin("mongosql-transition-tool")?
        .arg("--input")
        .arg("/a/spoonful/of/sugar")
        .output()?;

    cmd.assert()
        .failure()
        .stderr(predicate::str::contains("No such file or directory"));

    Ok(())
}

#[test]
fn uri_without_auth_and_no_username_arg_is_error() -> Result<(), Box<dyn std::error::Error>> {
    let cmd = Command::cargo_bin("mongosql-transition-tool")?
        .arg("--uri")
        .arg("mongodb://localhost:27017")
        .output()?;

    cmd.assert().failure().stderr(predicate::str::contains(
        "No username provided for authentication.",
    ));

    Ok(())
}

#[test]
fn incorrect_credentials_is_error() -> Result<(), Box<dyn std::error::Error>> {
    let uri = format!(
        "mongodb://AzureDiamond:hunter2@localhost:{}",
        std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
    );
    let cmd = Command::cargo_bin("mongosql-transition-tool")?
        .arg("--uri")
        .arg(uri)
        .output()?;

    cmd.assert()
        .failure()
        .stderr(predicate::str::contains("Authentication failed."));

    Ok(())
}

#[test]
fn valid_input_and_no_uri_produces_log_report() -> Result<(), Box<dyn std::error::Error>> {
    let cwd = std::env::current_dir()?;
    let out_dir_name = "./flibbertygibbet";

    // clear the directory, just in case it exists
    let _ = std::fs::remove_dir_all(cwd.join(out_dir_name));

    let cmd = Command::cargo_bin("mongosql-transition-tool")?
        .arg("--input")
        .arg("./testfiles/logfiles/logs_with_invalid/")
        .arg("--output")
        .arg(out_dir_name)
        .output()?;

    cmd.assert().success();

    let log_analysis_file_count =
        std::fs::read_dir(cwd.join(out_dir_name).join("log_analysis"))?.count();
    assert_eq!(
        log_analysis_file_count, 1,
        "Expected 1 log analysis file, found {log_analysis_file_count}",
    );
    // check file contents
    let log_file_path = std::fs::read_dir(cwd.join(out_dir_name).join("log_analysis"))?
        .next()
        .unwrap()
        .expect("No analysis file")
        .path();
    let log_file = std::fs::read_to_string(log_file_path)?;
    let tab_links = log_file
        .matches(r#"<button class="tablinks btn-green""#)
        .count();
    assert_eq!(tab_links, 4, "Expected 4 tab links, found {}", tab_links);

    // do not remove this directive. rustfmt will remove the indentation and cause the test to fail
    #[rustfmt::skip]
    let table_count = log_file
        .matches( r#"        <tr>
                <td style='padding-right: 1em;'>Queries Found with Valid Syntax for AtlasSQL:</td>
                <td style='text-align: right;'><span class='highlight'>36</span></td>
             </tr><tr>
                <td style='padding-right: 1em;'>Queries Found with Invalid Syntax for AtlasSQL:</td>
                <td style='text-align: right;'><span class='highlight'>15</span></td>
             </tr>"#
            )
        .count();

    assert_eq!(
        table_count, 1,
        "Expected 1 table with 2 rows for 'Queries Found with Valid Sytax` in log_analysis file, found {table_count}",
    );

    assert!(cwd.join(out_dir_name).join("index.html").exists());

    let index_file = std::fs::read_to_string(cwd.join(out_dir_name).join("index.html"))?;

    assert_eq!(
        index_file.matches("Log Reports").count(),
        1,
        "Expected 1 Log Reports section"
    );
    assert_eq!(
        index_file
            .matches(r#"<a href="log_analysis/Atlas_SQL_Readiness.html">log_analysis/Atlas_SQL_Readiness.html</a>"#).count(),
        1,
        "Expected 1 link to Atlas SQL Readiness log analysis"
    );
    assert!(std::fs::read_dir(cwd.join(out_dir_name))?.any(|entry| {
        entry
            .as_ref()
            .map(|entry| entry.file_name().to_string_lossy().ends_with(".zip"))
            .unwrap_or(false)
    }));

    let _ = std::fs::remove_dir_all(cwd.join(out_dir_name));
    Ok(())
}

// This test allows some flexibility in the number of output files so as
// not to be overly brittle when run locally. There is a chance it could
// be seeing too many databases, but that is mitigated by the next test.
#[tokio::test]
async fn valid_uri_and_no_input_is_ok() -> Result<(), Box<dyn std::error::Error>> {
    let test_db_manager =
        TestDatabaseManager::new(TEST_DBS.clone(), TEST_COLLS.clone(), TEST_VIEW.clone()).await;
    let uri = format!(
        "mongodb://test:test@localhost:{}",
        std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
    );

    let cwd = std::env::current_dir()?;
    let out_dir_name = "./ignored";

    // clear the test directory, just in case it exists
    let _ = std::fs::remove_dir_all(cwd.join(out_dir_name));

    let cmd = Command::cargo_bin("mongosql-transition-tool")?
        .arg("--uri")
        .arg(uri)
        .arg("--output")
        .arg(out_dir_name)
        .output()?;

    cmd.assert().success();

    let schema_analysis_file_count =
        std::fs::read_dir(cwd.join(out_dir_name).join("schema_analysis"))?.count();

    assert!(
        schema_analysis_file_count >= 2,
        "Expected at least 2 schema analysis files, found {schema_analysis_file_count}",
    );

    for entry in std::fs::read_dir(cwd.join(out_dir_name).join("schema_analysis"))? {
        let entry = entry?;
        let file_name = entry.file_name().into_string().unwrap();
        let file_contents = std::fs::read_to_string(entry.path())?;
        let database_count = file_contents.matches("Database").count();
        let collection_count = file_contents.matches("Collection").count();
        assert_eq!(database_count, 1, "Expected 1 database");

        if file_name.contains("trr_cli_test_db1") {
            assert_eq!(
                file_contents.matches("Database: trr_cli_test_db1").count(),
                1,
                "Expected 1 database named trr_cli_test_db1",
            );
            assert_eq!(collection_count, 3, "Expected 3 collections");
        } else if file_name.contains("trr_cli_test_db2") {
            assert_eq!(
                file_contents.matches("Database: trr_cli_test_db2").count(),
                1,
                "Expected 1 database named trr_cli_test_db2",
            );
            assert_eq!(collection_count, 3, "Expected 3 collections");
        }

        if file_name.starts_with("trr_cli") {
            assert_eq!(collection_count, 3, "Expected 3 collections");
            assert_eq!(
                file_contents.matches("Collection: acoll").count(),
                1,
                "Expected 1 collection named zcoll",
            );
            assert_eq!(
                file_contents.matches("Collection: bcoll").count(),
                1,
                "Expected 1 collection named bcoll",
            );
            assert_eq!(
                file_contents.matches("Collection: cview").count(),
                1,
                "Expected 1 collection named cview",
            );
        }
    }

    assert!(cwd.join(out_dir_name).join("index.html").exists());

    let index_file = std::fs::read_to_string(cwd.join(out_dir_name).join("index.html"))?;
    assert_eq!(
        index_file
            .matches(r#"<a href="schema_analysis/trr_cli_test_db1.html">schema_analysis/trr_cli_test_db1.html</a>"#)
            .count(),
        1,
        "Expected 1 link to trr_cli_test_db1 schema analysis"
    );
    assert_eq!(
        index_file

            .matches(r#"<a href="schema_analysis/trr_cli_test_db2.html">schema_analysis/trr_cli_test_db2.html</a>"#)
            .count(),
        1,
        "Expected 1 link to trr_cli_test_db2 schema analysis"
    );

    assert!(std::fs::read_dir(cwd.join(out_dir_name))?.any(|entry| {
        entry
            .as_ref()
            .map(|entry| entry.file_name().to_string_lossy().ends_with(".zip"))
            .unwrap_or(false)
    }));

    let _ = std::fs::remove_dir_all(cwd.join(out_dir_name));
    test_db_manager.cleanup().await;
    Ok(())
}

// This test also uses the --include and --exclude flags. If the CLI is not
// passing these correctly and the library is not properly excluding based on the inputs
// AND the global ignores, we'll see a failure here.
#[tokio::test]
async fn include_and_exclude_produces_correct_output() -> Result<(), Box<dyn std::error::Error>> {
    let test_db_manager =
        TestDatabaseManager::new(TEST_DBS.clone(), TEST_COLLS.clone(), TEST_VIEW.clone()).await;
    let cwd = std::env::current_dir()?;
    let out_dir_name = "./excluded";
    // clear the test directory, just in case it exists
    let _ = std::fs::remove_dir_all(cwd.join(out_dir_name));

    let uri = format!(
        "mongodb://test:test@localhost:{}",
        std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
    );
    let cmd = Command::cargo_bin("mongosql-transition-tool")?
        .arg("--uri")
        .arg(uri)
        .arg("--output")
        .arg(out_dir_name)
        .arg("--include")
        .arg("trr_cli_test_db1.*")
        .arg("--include")
        .arg("trr_cli_test_db2.*") // to ensure it makes it past the inclusion filter
        .arg("--exclude")
        .arg("trr_cli_test_db2.*")
        .output()?;

    cmd.assert().success();

    let schema_analysis_file_count =
        std::fs::read_dir(cwd.join(out_dir_name).join("schema_analysis"))?.count();

    assert_eq!(
        schema_analysis_file_count, 1,
        "Expected 1 schema analysis files, found {schema_analysis_file_count}",
    );

    for entry in std::fs::read_dir(cwd.join(out_dir_name).join("schema_analysis"))? {
        let entry = entry?;
        let file_contents = std::fs::read_to_string(entry.path())?;
        let database_count = file_contents.matches("Database").count();
        let collection_count = file_contents.matches("Collection").count();
        assert_eq!(database_count, 1, "Expected 1 database");
        assert_eq!(collection_count, 3, "Expected 3 collections");

        assert_eq!(
            file_contents.matches("Database: trr_cli_test_db1").count(),
            1,
            "Expected 1 trr_cli_test_db_1 database"
        );

        assert_eq!(
            file_contents.matches("Collection: acoll").count(),
            1,
            "Expected 1 acoll collection"
        );
        assert_eq!(
            file_contents.matches("Collection: bcoll").count(),
            1,
            "Expected 1 bcoll collection"
        );
        assert_eq!(
            file_contents.matches("Collection: cview").count(),
            1,
            "Expected 1 cview collection"
        );
    }

    assert!(cwd.join(out_dir_name).join("index.html").exists());

    let index_file = std::fs::read_to_string(cwd.join(out_dir_name).join("index.html"))?;
    assert_eq!(
        index_file
            .matches(r#"<a href="schema_analysis/trr_cli_test_db1.html">schema_analysis/trr_cli_test_db1.html</a>"#)
            .count(),
        1,
        "Expected 1 link to trr_cli_test_db1 schema analysis"
    );
    assert_ne!(
        index_file
            .matches(
                r#"<a href="schema_analysis/trr_cli_test_db2.html">schema_analysis/trr_cli_test_db2.html</a>"#
            )
            .count(),
        1,
        "Did not expect a link to trr_cli_test_db2 schema analysis"
    );

    let _ = std::fs::remove_dir_all(cwd.join(out_dir_name));
    test_db_manager.cleanup().await;
    Ok(())
}
