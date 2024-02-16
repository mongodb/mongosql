/// A small utility module to encapsulate the logic of writing a Vec<LogEntry> to a CSV file.
use std::{
    io::{self, Error, ErrorKind},
    path::Path,
};

use crate::log_parser::LogEntry;

/// generate_csv takes a file path, a date, and a LogParseResult and writes elements of the LogParseResult
/// to csv files.
pub fn generate_csv(
    file_path: &Path,
    date: &str,
    log_parse: &crate::log_parser::LogParseResult,
    file_stem: &str,
) -> io::Result<()> {
    write_csv(
        &log_parse.invalid_queries,
        "InvalidQueries",
        file_path,
        date,
        file_stem,
    )
    .map_err(|e| Error::new(ErrorKind::Other, e))?;
    write_csv(
        &log_parse.invalid_queries,
        "ValidQueries",
        file_path,
        date,
        file_stem,
    )
    .map_err(|e| Error::new(ErrorKind::Other, e))?;
    Ok(())
}

fn write_csv(
    data: &Option<Vec<LogEntry>>,
    label: &str,
    file_path: &Path,
    date: &str,
    file_stem: &str,
) -> io::Result<()> {
    if data.is_none() {
        Ok(())
    } else {
        let data = data.as_ref().unwrap();
        let mut writer =
            csv::Writer::from_path(file_path.join(format!("{file_stem}_{label}_{date}.csv")))?;
        for entry in data {
            writer.serialize(entry)?;
        }
        writer.flush()
    }
}
