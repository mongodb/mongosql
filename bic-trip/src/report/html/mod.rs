// This file contains utilities for writing HTML elements to a file.

use base64::{prelude::BASE64_STANDARD, Engine};
use build_html::{Html, HtmlContainer, HtmlPage};
use chrono::NaiveDateTime;
use std::{collections::HashMap, fs, fs::File, io::Write, path::Path};

use image::{DynamicImage, ImageOutputFormat};

use crate::log_parser::{LogEntry, QueryRepresentation};
use chrono::prelude::*;
use sqlformat::{format, FormatOptions, QueryParams};

/// generate_html takes a file path, a date, and a LogParseResult and writes the HTML report to a file.
pub fn generate_html(
    file_path: &Path,
    date: &str,
    log_parse: &crate::log_parser::LogParseResult,
    file_stem: &str,
    report_name: &str,
) -> std::io::Result<()> {
    let mut report_file = File::create(file_path.join(format!("{file_stem}_{date}.html")))?;
    report_file.write_all(generate_html_elements(log_parse, report_name).as_bytes())
}

/// process_summary_html generates summary page of database and query information
fn process_summary_html(log_parse: &crate::log_parser::LogParseResult) -> String {
    let mut table_html = String::from("<table><tr>")
        + "<th>Database</th>"
        + "<th>Access Count</th>"
        + "<th>Recent access time</th>"
        + "<th>Document fields in queries</th>"
        + "<th>Arrays in queries</th></tr>";

    // Database name -> (Access Count, Most Recent Access, Complex Fields Used, Arrays in Queries)
    let mut summary: HashMap<String, (u32, NaiveDateTime, u32, u32)> = HashMap::new();

    // Process collections
    if let Some(collections) = &log_parse.collections {
        for (db, _, count, last_accessed) in collections {
            let db = db.clone();
            summary
                .entry(db.clone())
                .and_modify(|entry| {
                    entry.0 += count;
                    if entry.1 < *last_accessed {
                        entry.1 = *last_accessed;
                    }
                })
                .or_insert((*count, *last_accessed, 0, 0));
        }
    }

    // Process complex types for subpath_fields
    if let Some(subpath_fields) = &log_parse.subpath_fields {
        for (field, _, _) in subpath_fields {
            let db = field.db.clone();
            summary
                .entry(db)
                .and_modify(|entry| {
                    entry.2 += 1;
                })
                .or_insert((0, NaiveDateTime::from_timestamp_opt(0, 0).unwrap(), 1, 0));
        }
    }

    // Process complex types for array_datasources
    if let Some(array_datasources) = &log_parse.array_datasources {
        for (db, _, _, _) in array_datasources {
            let db = db.clone();
            summary
                .entry(db)
                .and_modify(|entry| {
                    entry.3 += 1;
                })
                .or_insert((0, NaiveDateTime::from_timestamp_opt(0, 0).unwrap(), 0, 1));
        }
    }

    let mut db_names: Vec<&String> = summary.keys().collect();
    // Sorting by database names alphabetically
    db_names.sort_by_key(|name| name.to_lowercase());
    for db_name in db_names {
        let (access_count, most_recent_access, complex_fields_used, arrays_in_queries) =
            summary.get(db_name).unwrap();
        table_html.push_str(&format!(
            "<tr><td>{}</td><td>{}</td><td>{}</td><td>{}</td><td>{}</td></tr>",
            db_name,
            access_count,
            most_recent_access.format("%Y-%m-%d %H:%M"),
            complex_fields_used,
            arrays_in_queries
        ));
    }
    table_html.push_str("</table>");

    let valid_queries = log_parse.valid_queries.as_ref().map_or(0, |v| v.len());
    let invalid_queries = log_parse.invalid_queries.as_ref().map_or(0, |v| v.len());
    let queries_html = format!(
        "<div><h2>Queries</h2></div>\
        <table>\
            <tr><td>Valid:</td><td style='text-align: right;padding-left: 20px;'>{:>8}</td></tr>\
            <tr><td>Invalid:</td><td style='text-align: right;padding-left: 20px;'>{:>8}</td></tr>\
        </table>",
        valid_queries, invalid_queries
    );

    format!("{}{}", table_html, queries_html)
}

fn process_complex_types_html(
    subpath_fields: &Option<Vec<(crate::log_parser::SubpathField, u32, chrono::NaiveDateTime)>>,
    array_datasources: &Option<Vec<(String, String, u32, chrono::NaiveDateTime)>>,
) -> String {
    let subpath_fields_html = subpath_fields.as_ref().map_or_else(
        || "<h3>Document Fields</h3><p>No document fields found.</p>".to_string(),
        |fields| {
            let items = fields.clone().iter().map(|(field, count, last_accessed)| {
                format!(
                    "<li>Database: {}<br>Collection: {}<br>Document Field: {}<br>Access Count: {}<br>Last Accessed: {}<br><br></li>",
                    field.db,
                    field.collection,
                    field.subpath_field,
                    count,
                    last_accessed.format("%Y-%m-%d %H:%M:%S")
                )
            }).collect::<Vec<String>>().join("");
            format!("<div><h3>Document Fields</h3><ul>{}</ul></div>", items)
        },
    );

    let array_datasources_html = array_datasources.as_ref().map_or_else(
        || "<h3>Array Datasources</h3><p>No array datasources found.</p>".to_string(),
        |datasources| {
            let items = datasources.clone().iter().map(|(db, collection, count, last_accessed)| {
                format!(
                    "<li>Database: {}<br>Collection: {}<br>Access Count: {}<br>Last Accessed: {}<br><br></li>",
                    db,
                    collection,
                    count,
                    last_accessed.format("%Y-%m-%d %H:%M:%S")
                )
            }).collect::<Vec<String>>().join("");
            format!("<div><h3>Array Datasources</h3><ul>{}</ul></div>", items)
        },
    );
    let info = r#"<p><em>Note: The following values are derived from query analyses
                           and should be considered as approximations for guidance.</em></p>"#;
    format!("{}{}{}", info, subpath_fields_html, array_datasources_html)
}

// process_query_html takes a label and a list of LogEntry and returns a String
// with the entries formatted as an HTML list ordered by access count and time
fn process_query_html(label: &str, queries: &Option<Vec<LogEntry>>) -> String {
    if let Some(queries) = queries {
        let queries_html = queries.clone()
            .iter()
            .map(|q| {
                let formatted_sql = format(&q.query, &QueryParams::None, FormatOptions::default());

                match &q.query_representation {
                    QueryRepresentation::Query(_) => format!(
                        "<li>Time: {}<br>Count: {}<br>Query: <pre>{}</pre><br></li>",
                        q.timestamp, q.query_count, formatted_sql
                    ),
                    QueryRepresentation::ParseError(error) => format!(
                        "<li>Time: {}<br>Count: {}<br>Query: <pre>{}</pre>Parse Error: {}<br><br></li>",
                        q.timestamp, q.query_count, formatted_sql, error
                    ),
                }
            })
            .collect::<Vec<String>>()
            .join("");

        format!("<div><h3>{}</h3><ul>{}</ul></div>", label, queries_html)
    } else {
        format!("<h3>{}</h3><p>No Queries found.</p>", label)
    }
}

// process_collections_html takes a label and a list of collections and returns a String
// with the collections formatted as an HTML sorted by access count and time
fn process_collections_html(
    label: &str,
    collections: &Option<Vec<(String, String, u32, NaiveDateTime)>>,
) -> String {
    if let Some(collections) = collections {
        let collections_html = collections
            .iter()
            .map(|(db, collection, count, last_accessed)| {
                format!(
                    "<li>Database: {}<br>Collection: {}<br>Access Count: {}<br>Last Accessed: {}<br><br></li>",
                    db,
                    collection,
                    count,
                    last_accessed.format("%m-%d-%Y %H:%M:%S")
                )
            })
            .collect::<Vec<String>>()
            .join("");
        format!("<h3>{}</h3><ul>{}</ul>", label, collections_html)
    } else {
        format!("<h3>{}</h3><p>No collections found.</p>", label)
    }
}

/// encode_image_path_to_base64 encodes image from path into a base64 string to be
/// added directly to HTML
fn encode_image_path_to_base64(path: &str) -> std::io::Result<String> {
    let img = image::open(path)
        .map_err(|e| std::io::Error::new(std::io::ErrorKind::Other, e))?
        .into_rgba8();

    let mut image_data: Vec<u8> = Vec::new();

    DynamicImage::ImageRgba8(img)
        .write_to(
            &mut std::io::Cursor::new(&mut image_data),
            ImageOutputFormat::Png,
        )
        .unwrap();

    Ok(BASE64_STANDARD.encode(&image_data))
}

fn generate_html_elements(
    log_parse: &crate::log_parser::LogParseResult,
    report_name: &str,
) -> String {
    let base_dir = env!("CARGO_MANIFEST_DIR");
    let logo_path = format!("{base_dir}/resources/html/MongoDB_ForestGreen.png");
    let tab_js_path = format!("{base_dir}/resources/html/tab_logic.js");

    let invalid_queries_html = process_query_html("Invalid", &log_parse.invalid_queries);
    let valid_queries_html = process_query_html("Valid", &log_parse.valid_queries);
    let collections_html = process_collections_html("Found Collections", &log_parse.collections);
    let complex_types_html =
        process_complex_types_html(&log_parse.subpath_fields, &log_parse.array_datasources);
    let summary_html = process_summary_html(log_parse);

    let datetime_str = Local::now().format("%m-%d-%Y %H:%M").to_string();

    let mut page = HtmlPage::new().with_title(report_name);

    // Add MongoDB logo
    let icon_base64 = encode_image_path_to_base64(&logo_path).expect("Incorrect path");
    page.add_raw(&format!(
        r#"<div class="header-container">
        <img src="data:image/png;base64,{}" alt="MongoDB Icon" class="mongodb-icon">
        <h1>{}</h1>
        </div>
        <style>
        .header-container {{
            display: flex;
            flex-direction: column;
            align-items: left;
            text-align: left;
        }}
        .mongodb-icon {{
            width: 150px;
            height: auto;
            margin-bottom: 0px;
        }}
        </style>"#,
        icon_base64, report_name
    ));

    // Add timestamp
    page.add_raw(&format!("<p>Report was generated at: {}</p>", datetime_str));

    page.add_raw(
        r#"<div class="tab">
                <button class="tablinks" onclick="openTab(event, 'Summary')">Summary</button>
                <button class="tablinks" onclick="openTab(event, 'Queries')">Queries</button>
                <button class="tablinks" onclick="openTab(event, 'Collections')">Collections</button>
                <button class="tablinks" onclick="openTab(event, 'ComplexTypes')">Complex Types</button>
                <button class="tablinks" onclick="openTab(event, 'Schema')">Schema</button>
            </div>"#
            );

    // Add generated values to the tabs
    page.add_raw(&format!(
        r#"<div id="Queries" class="tabcontent"><h2>Queries</h2>{}{}</div>"#,
        valid_queries_html, invalid_queries_html
    ));
    page.add_raw(&format!(
        r#"<div id="Collections" class="tabcontent"><h2>Collections</h2>{}</div>"#,
        collections_html
    ));
    page.add_raw(&format!(
        r#"<div id="ComplexTypes" class="tabcontent"><h2>Complex Types</h2>{}</div>"#,
        complex_types_html
    ));
    page.add_raw(&format!(
        r#"<div id="Summary" class="tabcontent"><h2>Summary</h2>{}</div>"#,
        summary_html
    ));

    // SQL-1956: Add schema analysis information to HTML
    page.add_raw(&format!(
        r#"<div id="Schema" class="tabcontent"><h2>Schema Analysis</h2>{}</div>"#,
        ""
    ));

    page.add_raw(&format!(
        r#"<script>{}</script>"#,
        fs::read_to_string(&tab_js_path)
            .unwrap_or_else(|_| panic!("Failed to read file {}", tab_js_path))
    ));
    page.to_html_string()
}
