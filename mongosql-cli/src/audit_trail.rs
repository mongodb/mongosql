//! Audit-trail and translation-stage helpers for the mongosql CLI.
//!
//! This module owns the `TranslationCheckpoint` enum, the intermediate-representation
//! collection struct, and the logic for running each compilation stage, writing
//! `audit_trail.zip`, and dispatching the combined audit-trail flow.

use crate::{run_query_and_display_results, CliError};
use mongosql::catalog::Catalog;

/// Compilation stages at which translation can be halted and inspected.
#[derive(clap::ValueEnum, Debug, Clone, Copy)]
pub(crate) enum TranslationCheckpoint {
    /// Stop after SQL parsing and AST rewrites; print the AST.
    Ast,
    /// Stop after algebrizing to MIR and running optimizer passes; print the MIR tree.
    Mir,
    /// AIR pretty-printing is not yet implemented.
    Air,
    /// Full translation to MQL; print the generated pipeline (default when --stage is omitted).
    Mql,
}

/// Collected intermediate representations for the audit trail.
///
/// Fields are `None` for stages not reached given the requested checkpoint.
struct TranslationStages {
    ast_repr: Option<String>,
    mir_repr: Option<String>,
    air_repr: Option<String>,
    pipeline_json: Option<String>,
    /// Only `Some` when stage is `Mql`; needed for the `--execute` path.
    translation: Option<mongosql::Translation>,
}

/// Runs translation stages sequentially up to `stage`, collecting each
/// intermediate representation.
///
/// The catalog is only built when `stage` is `Mir`, `Air`, or `Mql`.
///
/// # Errors
/// Returns `CliError` if any translation stage or catalog build fails.
fn run_stages_up_to(
    stage: TranslationCheckpoint,
    current_db: &str,
    query: &str,
    catalog: &Catalog,
) -> Result<TranslationStages, CliError> {
    let ast_repr = Some(mongosql::translate_sql_to_ast_repr(query)?);

    if matches!(stage, TranslationCheckpoint::Ast) {
        return Ok(TranslationStages {
            ast_repr,
            mir_repr: None,
            air_repr: None,
            pipeline_json: None,
            translation: None,
        });
    }

    let options = mongosql::options::SqlOptions {
        allow_order_by_missing_columns: true,
        ..Default::default()
    };

    let mir_repr = Some(mongosql::translate_sql_to_mir_repr(
        current_db, query, &catalog, options,
    )?);

    if matches!(stage, TranslationCheckpoint::Mir) {
        return Ok(TranslationStages {
            ast_repr,
            mir_repr,
            air_repr: None,
            pipeline_json: None,
            translation: None,
        });
    }

    let air_repr = Some(mongosql::translate_sql_to_air_repr(
        current_db, query, &catalog, options,
    )?);

    if matches!(stage, TranslationCheckpoint::Air) {
        return Ok(TranslationStages {
            ast_repr,
            mir_repr,
            air_repr,
            pipeline_json: None,
            translation: None,
        });
    }

    // Mql
    let translation = mongosql::translate_sql(current_db, query, &catalog, options)?;
    let pipeline_json =
        serde_json::to_string_pretty(&translation.pipeline).map_err(|e| CliError(e.to_string()))?;

    Ok(TranslationStages {
        ast_repr,
        mir_repr,
        air_repr,
        pipeline_json: Some(pipeline_json),
        translation: Some(translation),
    })
}

/// Writes intermediate translation representations into `audit_trail.zip`
/// in the current working directory, overwriting any existing file.
///
/// `initial_query.sql` is always written. Remaining files are written only
/// for stages that were reached (i.e. their field is `Some`).
///
/// # Errors
/// Returns `CliError` if the zip file cannot be created or written.
fn write_audit_trail(query: &str, stages: &TranslationStages) -> Result<(), CliError> {
    use std::io::Write as _;
    use zip::write::FileOptions;
    use zip::CompressionMethod;

    let file = std::fs::File::create("audit_trail.zip")?;
    let mut zip = zip::write::ZipWriter::new(file);
    let options = FileOptions::<()>::default().compression_method(CompressionMethod::Deflated);

    zip.start_file("audit_trail/initial_query.sql", options)?;
    zip.write_all(query.as_bytes())?;

    if let Some(ast) = &stages.ast_repr {
        zip.start_file("audit_trail/query.ast", options)?;
        zip.write_all(ast.as_bytes())?;
    }
    if let Some(mir) = &stages.mir_repr {
        zip.start_file("audit_trail/query.mir", options)?;
        zip.write_all(mir.as_bytes())?;
    }
    if let Some(air) = &stages.air_repr {
        zip.start_file("audit_trail/query.air", options)?;
        zip.write_all(air.as_bytes())?;
    }
    if let Some(pipeline) = &stages.pipeline_json {
        zip.start_file("audit_trail/pipeline.js", options)?;
        zip.write_all(pipeline.as_bytes())?;
    }

    zip.finish()?;
    Ok(())
}

/// Runs all translation stages up to `stage`, writes `audit_trail.zip`, prints
/// the stage output to stdout, and optionally executes the query.
///
/// Stdout output is identical to running `--stage` alone.
///
/// # Errors
/// Returns `CliError` if any translation stage, zip write, or query execution fails.
pub(crate) fn handle_audit_trail(
    stage: TranslationCheckpoint,
    current_db: &str,
    query: &str,
    uri: &str,
    execute: bool,
    catalog: &Catalog,
) -> Result<(), CliError> {
    if execute && !matches!(stage, TranslationCheckpoint::Mql) {
        return Err(CliError(
            "--execute is only valid with --stage mql or without --stage".to_string(),
        ));
    }
    let stages = run_stages_up_to(stage, current_db, query, catalog)?;
    write_audit_trail(query, &stages)?;
    eprintln!("[INFO] audit_trail.zip written to current working directory.");

    match stage {
        TranslationCheckpoint::Ast => {
            println!("{}", stages.ast_repr.expect("ast computed for Ast stage"));
        }
        TranslationCheckpoint::Mir => {
            println!("{}", stages.mir_repr.expect("mir computed for Mir stage"));
        }
        TranslationCheckpoint::Air => {
            println!("{}", stages.air_repr.expect("air computed for Air stage"));
        }
        TranslationCheckpoint::Mql => {
            let translation = stages
                .translation
                .as_ref()
                .expect("translation computed for Mql stage");
            let schema = serde_json::to_string_pretty(&translation.result_set_schema)
                .map_err(|e| CliError(e.to_string()))?;
            println!(
                "target_db: {},\ntarget_collection: {:?},\nresult set schema:\n{}\npipeline:\n[",
                translation.target_db, translation.target_collection, schema
            );
            let bson::Bson::Array(ref pipeline) = translation.pipeline else {
                return Err(CliError("pipeline is not an array".to_string()));
            };
            for doc in pipeline {
                println!("    {doc},");
            }
            println!("]");
        }
    }

    if execute {
        run_query_and_display_results(
            uri,
            stages
                .translation
                .expect("translation computed for Mql stage"),
        )?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    static CWD_LOCK: std::sync::Mutex<()> = std::sync::Mutex::new(());

    fn make_test_stages_mql() -> TranslationStages {
        TranslationStages {
            ast_repr: Some("ast output".to_string()),
            mir_repr: Some("mir output".to_string()),
            air_repr: Some("air output".to_string()),
            pipeline_json: Some(r#"[{"$match":{}}]"#.to_string()),
            translation: None,
        }
    }

    #[test]
    fn write_audit_trail_creates_correct_entries_for_mql_stage() {
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let stages = make_test_stages_mql();
        write_audit_trail("SELECT 1", &stages).unwrap();
        std::env::set_current_dir(original).unwrap();
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let names: Vec<String> = (0..archive.len())
            .map(|i| archive.by_index(i).unwrap().name().to_string())
            .collect();
        assert!(names.contains(&"audit_trail/initial_query.sql".to_string()));
        assert!(names.contains(&"audit_trail/query.ast".to_string()));
        assert!(names.contains(&"audit_trail/query.mir".to_string()));
        assert!(names.contains(&"audit_trail/query.air".to_string()));
        assert!(names.contains(&"audit_trail/pipeline.js".to_string()));
        assert_eq!(names.len(), 5);
    }

    #[test]
    fn write_audit_trail_creates_correct_entries_for_ast_stage() {
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let stages = TranslationStages {
            ast_repr: Some("ast only".to_string()),
            mir_repr: None,
            air_repr: None,
            pipeline_json: None,
            translation: None,
        };
        write_audit_trail("SELECT 1", &stages).unwrap();
        std::env::set_current_dir(original).unwrap();
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let names: Vec<String> = (0..archive.len())
            .map(|i| archive.by_index(i).unwrap().name().to_string())
            .collect();
        assert_eq!(names.len(), 2);
        assert!(names.contains(&"audit_trail/initial_query.sql".to_string()));
        assert!(names.contains(&"audit_trail/query.ast".to_string()));
    }

    #[test]
    fn write_audit_trail_sql_content_is_verbatim() {
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let query = "SELECT\n  á\tFROM t";
        let stages = make_test_stages_mql();
        write_audit_trail(query, &stages).unwrap();
        std::env::set_current_dir(original).unwrap();
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let mut entry = archive.by_name("audit_trail/initial_query.sql").unwrap();
        let mut content = Vec::new();
        std::io::Read::read_to_end(&mut entry, &mut content).unwrap();
        assert_eq!(content, query.as_bytes());
    }

    #[test]
    fn write_audit_trail_pipeline_json_is_valid_json() {
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        let stages = TranslationStages {
            ast_repr: Some("ast".to_string()),
            mir_repr: Some("mir".to_string()),
            air_repr: Some("air".to_string()),
            pipeline_json: Some(r#"[{"$match": {"x": 1}}]"#.to_string()),
            translation: None,
        };
        write_audit_trail("SELECT 1", &stages).unwrap();
        std::env::set_current_dir(original).unwrap();
        let file = std::fs::File::open(&zip_path).unwrap();
        let mut archive = zip::ZipArchive::new(file).unwrap();
        let mut entry = archive.by_name("audit_trail/pipeline.js").unwrap();
        let mut content = String::new();
        std::io::Read::read_to_string(&mut entry, &mut content).unwrap();
        assert!(serde_json::from_str::<serde_json::Value>(&content).is_ok());
    }

    #[test]
    fn write_audit_trail_overwrites_existing_zip() {
        let _guard = CWD_LOCK.lock().unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        let zip_path = dir.path().join("audit_trail.zip");
        let original = std::env::current_dir().unwrap();
        std::env::set_current_dir(dir.path()).unwrap();
        std::fs::write("audit_trail.zip", b"not a real zip").unwrap();
        let stages = make_test_stages_mql();
        let result = write_audit_trail("SELECT 1", &stages);
        std::env::set_current_dir(original).unwrap();
        assert!(result.is_ok());
        let file = std::fs::File::open(&zip_path).unwrap();
        let archive_result = zip::ZipArchive::new(file);
        assert!(archive_result.is_ok());
        let mut archive = archive_result.unwrap();
        assert!(archive.by_name("audit_trail/initial_query.sql").is_ok());
    }
}
