/// A small utility module to encapsulate the logic of writing the Transition Readiness Information
/// to a zip file containing separate CSV files for each type of analysis.
use anyhow::Result;
use serde::Serialize;
use std::io::Write;
use std::path::Path;
use zip::write::FileOptions;

use crate::log_parser::LogParseResult;

const DATE_TIME_FORMAT: &str = "%Y-%m-%d %H:%M";

#[derive(Serialize)]
struct ComplexType {
    #[serde(rename = "Type")]
    r#type: String,
    #[serde(rename = "Database")]
    database: String,
    #[serde(rename = "Collection")]
    collection: String,
    #[serde(rename = "Field")]
    field: String,
    #[serde(rename = "Access Count")]
    access_count: u32,
    #[serde(rename = "Last Accessed")]
    last_accessed: String,
}

#[derive(Serialize)]
struct Collection {
    #[serde(rename = "Database")]
    database: String,
    #[serde(rename = "Collection")]
    collection: String,
    #[serde(rename = "Access Count")]
    access_count: u32,
    #[serde(rename = "Last Accessed")]
    last_accessed: String,
}

#[derive(Serialize)]
struct Query {
    #[serde(rename = "Status")]
    status: String,
    #[serde(rename = "Timestamp")]
    timestamp: String,
    #[serde(rename = "Count")]
    count: u32,
    #[serde(rename = "Query")]
    query: String,
    #[serde(rename = "Error")]
    error: String,
}

// get_struct_name is a helper function to obtain the unqualified struct name
fn get_struct_name<T: ?Sized>() -> String {
    let full_type_name = std::any::type_name::<T>();
    let struct_name = full_type_name.split("::").last().unwrap_or(full_type_name);
    struct_name.to_string()
}

// generate_csv writes a zip file to the `file_path` that contains the CSV representation
// of the valid and invalid queries, collections, complex types, and schema analysis
pub fn generate_csv(
    file_path: &Path,
    date: &str,
    log_parse: &LogParseResult,
    file_stem: &str,
) -> Result<()> {
    let zip_file_path = file_path.join(format!("{file_stem}_{date}.zip"));
    let mut zip = zip::ZipWriter::new(std::fs::File::create(zip_file_path)?);

    let options = FileOptions::default()
        .compression_method(zip::CompressionMethod::Deflated)
        .unix_permissions(0o644);

    let csv_files = [
        generate_complex_types_csv(log_parse)?,
        generate_collections_csv(log_parse)?,
        generate_queries_csv(log_parse)?,
        // SQL-1973: Add schema analysis information to CSV output
    ];

    for (file_name, csv_data) in csv_files {
        write_csv_to_zip(&mut zip, date, &file_name, &csv_data, &options)?;
    }
    zip.finish()?;
    Ok(())
}

fn write_csv_to_zip(
    zip: &mut zip::ZipWriter<std::fs::File>,
    date: &str,
    file_name: &str,
    csv_data: &str,
    options: &FileOptions,
) -> Result<()> {
    let file_name = format!("{file_name}_{date}.csv");
    zip.start_file(file_name, *options)?;
    zip.write_all(csv_data.as_bytes())?;
    Ok(())
}

// generate_complex_types_csv generates a CSV of the complex types: Document Field and Array Datasource
pub fn generate_complex_types_csv(log_parse: &LogParseResult) -> Result<(String, String)> {
    let mut writer = csv::Writer::from_writer(Vec::new());
    if let Some(subpath_fields) = &log_parse.subpath_fields {
        for (field, count, last_accessed) in subpath_fields {
            let complex_type = ComplexType {
                r#type: "Document Field".to_string(),
                database: field.db.clone(),
                collection: field.collection.clone(),
                field: field.subpath_field.clone(),
                access_count: *count,
                last_accessed: last_accessed.format(DATE_TIME_FORMAT).to_string(),
            };
            writer.serialize(complex_type)?;
        }
    }
    if let Some(array_datasources) = &log_parse.array_datasources {
        for (db, collection, count, last_accessed) in array_datasources {
            let complex_type = ComplexType {
                r#type: "Array Datasource".to_string(),
                database: db.to_string(),
                collection: collection.to_string(),
                field: "".to_string(),
                access_count: count.to_string().parse()?,
                last_accessed: last_accessed.format(DATE_TIME_FORMAT).to_string(),
            };
            writer.serialize(complex_type)?;
        }
    }
    writer.flush()?;
    let csv_data = String::from_utf8(writer.into_inner()?)?;
    Ok((get_struct_name::<ComplexType>(), csv_data))
}

// This function now uses the Collection struct for serialization
pub fn generate_collections_csv(log_parse: &LogParseResult) -> Result<(String, String)> {
    let mut writer = csv::Writer::from_writer(Vec::new());

    if let Some(collections) = &log_parse.collections {
        for (db, collection, count, last_accessed) in collections {
            let collection_entry = Collection {
                database: db.to_string(),
                collection: collection.to_string(),
                access_count: *count,
                last_accessed: last_accessed.format(DATE_TIME_FORMAT).to_string(),
            };
            writer.serialize(collection_entry)?;
        }
    }
    writer.flush()?;
    let csv_data = String::from_utf8(writer.into_inner()?)?;
    Ok((get_struct_name::<Collection>(), csv_data))
}

// generate_queries_csv Generates a CSV for queries found in the parsed logs
pub fn generate_queries_csv(log_parse: &LogParseResult) -> Result<(String, String)> {
    let mut writer = csv::Writer::from_writer(Vec::new());

    let queries = log_parse
        .valid_queries
        .as_ref()
        .unwrap_or(&vec![])
        .iter()
        .map(|entry| Query {
            status: "Valid".to_string(),
            timestamp: entry.timestamp.format(DATE_TIME_FORMAT).to_string(),
            count: entry.query_count,
            query: entry.query.clone(),
            error: entry.query_representation.to_string(),
        })
        .chain(
            log_parse
                .invalid_queries
                .as_ref()
                .unwrap_or(&vec![])
                .iter()
                .map(|entry| Query {
                    status: "Invalid".to_string(),
                    timestamp: entry.timestamp.format(DATE_TIME_FORMAT).to_string(),
                    count: entry.query_count,
                    query: entry.query.clone(),
                    error: entry.query_representation.to_string(),
                }),
        )
        .collect::<Vec<_>>();

    for query in &queries {
        writer.serialize(query)?;
    }
    writer.flush()?;
    let csv_data = String::from_utf8(writer.into_inner()?)?;
    Ok((get_struct_name::<Query>(), csv_data))
}
