use chrono::NaiveDateTime;
use std::collections::HashMap;

use crate::log_parser::{LogEntry, QueryRepresentation};
use sqlformat::{format, FormatOptions, QueryParams};

const DATE_TIME_FORMAT: &str = "%m-%d-%Y %H:%M";

macro_rules! generate_table_row {
    ($label:expr, $value:expr) => {
        format!(
            "<tr>
                <td style='padding-right: 1em;'>{label}:</td>
                <td style='text-align: right;'><span class='highlight'>{value}</span></td>
             </tr>",
            label = $label,
            value = $value
        )
    };
}

/// process_summary_html generates summary page of database and query information
pub fn process_summary_html(log_parse: &crate::log_parser::LogParseResult) -> String {
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

    let summary_string = format!(
        "<table class='table1'>
        <tbody>
        {}
        </tbody>
     </table>",
        db_names
            .iter()
            .map(|db_name| {
                let (access_count, most_recent_access, complex_fields_used, arrays_in_queries) =
                    summary.get(*db_name).unwrap();
                let db_name_or_default = if db_name.is_empty() {
                    "Not Specified"
                } else {
                    db_name
                };
                format!(
                    "<tr>
                    <td colspan='2'>
                        <table class='table1'>
                            {}
                            {}
                            {}
                            {}
                            {}
                        </table>
                    </td>
                </tr>",
                    generate_table_row!("Database", db_name_or_default),
                    generate_table_row!("Access Count", access_count),
                    generate_table_row!(
                        "Last Accessed",
                        most_recent_access.format(DATE_TIME_FORMAT)
                    ),
                    generate_table_row!("Complex Fields Found", complex_fields_used),
                    generate_table_row!("Arrays Found", arrays_in_queries),
                )
            })
            .collect::<Vec<String>>()
            .join("")
    );

    let valid_queries = log_parse.valid_queries.as_ref().map_or(0, |v| v.len());
    let invalid_queries = log_parse.invalid_queries.as_ref().map_or(0, |v| v.len());
    let queries_html = format!(
        r#"<div>
        <h2>Queries</h2>
        <table class='table1'>
        {}{}
        </table>
      </div>"#,
        generate_table_row!(
            "Queries Found with Valid Syntax for AtlasSQL",
            valid_queries
        ),
        generate_table_row!(
            "Queries Found with Invalid Syntax for AtlasSQL",
            invalid_queries
        )
    );

    let logs_processed_html = if let Some(log_files) = &log_parse.log_files {
        let header = "<h2>Logs Processed</h2>";
        let table_header =
            "<table class='table1'><tr><th>Filename</th><th>Timestamp Range</th></tr>";
        let log_files_html = log_files
            .iter()
            .map(|log_file| {
                format!(
                    "<tr>
                    <td>{}</td>
                    <td>{} - {}</td>
                 </tr>",
                    log_file.filename,
                    log_file.oldest_timestamp.format(DATE_TIME_FORMAT),
                    log_file.newest_timestamp.format(DATE_TIME_FORMAT)
                )
            })
            .collect::<Vec<String>>()
            .join("");
        let table_footer = "</table>";
        format!(
            "{}{}{}{}",
            header, table_header, log_files_html, table_footer
        )
    } else {
        "<h2>Logs Processed</h2><p>No log files processed.</p>".to_string()
    };

    let info = r#"Provides an overview of the databases and the number of queries that reference fields within documents and arrays.<hr><br>"#;
    format!(
        "{}{}{}{}",
        info, summary_string, queries_html, logs_processed_html
    )
}

pub fn process_complex_types_html(
    subpath_fields: &Option<Vec<(crate::log_parser::SubpathField, u32, chrono::NaiveDateTime)>>,
    array_datasources: &Option<Vec<(String, String, u32, chrono::NaiveDateTime)>>,
) -> String {
    let subpath_fields_html = subpath_fields.as_ref().map_or_else(
        || "<h3>Fields in Documents</h3><p>No document fields found.</p>".to_string(),
        |fields| {
            let items = fields
                .iter()
                .map(|(field, count, last_accessed)| {
                    let db_html = if !field.db.is_empty() {
                        generate_table_row!("Database", &field.db)
                    } else {
                        String::new()
                    };
                    format!(
                        "<table class='table1'>
                        {}
                        {}
                        {}
                        {}
                        {}
                     </table>",
                        db_html,
                        generate_table_row!("Collection", &field.collection),
                        generate_table_row!("Field", &field.subpath_field),
                        generate_table_row!("Query Count", count),
                        generate_table_row!(
                            "Last Accessed",
                            &last_accessed.format(DATE_TIME_FORMAT)
                        ),
                    )
                })
                .collect::<Vec<String>>()
                .join("<br>");
            format!("<div><h3>Fields in Documents</h3>{}</div>", items)
        },
    );

    let array_datasources_html = array_datasources.as_ref().map_or_else(
        || "<h3>Array Datasources</h3><p>No array datasources found.</p>".to_string(),
        |datasources| {
            let items = datasources
                .iter()
                .map(|(db, collection, count, last_accessed)| {
                    let db_html = if !db.is_empty() {
                        generate_table_row!("Database", db)
                    } else {
                        String::new()
                    };
                    format!(
                        "<table class='table1'>
                        {}
                        {}
                        {}
                        {}
                     </table>",
                        db_html,
                        generate_table_row!("Collection", collection),
                        generate_table_row!("Query Count", count),
                        generate_table_row!(
                            "Last Accessed",
                            &last_accessed.format(DATE_TIME_FORMAT)
                        ),
                    )
                })
                .collect::<Vec<String>>()
                .join("<br>");
            format!("<div><h3>Array Datasources</h3>{}</div>", items)
        },
    );

    let fields_in_documents_info = r#"<p style='margin-top: 0; margin-bottom: 0;'><strong>Fields in Documents</strong> represents fields within a document that are used in the logged queries. These queries may require modifications to <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#flatten" target="_blank">FLATTEN</a> the complex data structure in order to access the same data in Atlas SQL.</p>"#;
    let array_datasources_info = r#"<p style='margin-top: 0; margin-bottom: 0;'><strong>Array Datasources</strong> represents collections that contain array fields which the BI Connector has flattened into separate tables for simplified querying.</p>"#;
    let info = r#"<p style='margin-top: 10px; margin-bottom: 10px;'><em>Note: The following values are derived from query analyses and should be considered as approximations for guidance.</em></p>"#;
    format!(
        "{}{}{}<hr><br>{}{}",
        fields_in_documents_info,
        array_datasources_info,
        info,
        subpath_fields_html,
        array_datasources_html
    )
}

// process_query_html takes a label and a list of LogEntry and returns a String
// with the entries formatted as an HTML list ordered by access count and time
pub fn process_query_html(label: &str, queries: &Option<Vec<LogEntry>>) -> String {
    if let Some(queries) = queries {
        let info = r#"This tab aids in identifying and troubleshooting queries that fail
            to execute due to syntax errors or unsupported commands.  Unsupported Queries have functionality specific to the BI Connector."#;
        let note = r#"<p><em>Note: Queries have not been executed against the database.</em></p>"#;
        let queries_header_html = if label == "Invalid" {
            format!("<div id='invalid-queries'><h3>{}</h3></div>", label)
        } else if label == "Valid" {
            format!("<div>{}{}<hr><br><h3>{}</h3></div>", info, note, label)
        } else {
            format!("<div><h3>{}</h3></div>", label)
        };
        let queries_html = queries.clone()
            .iter()
            .map(|q| {
                let formatted_sql = format(&q.query, &QueryParams::None, FormatOptions::default());
                let non_empty_users: Vec<&String> = q.users.iter().filter(|user| !user.is_empty()).collect();
                let user_display = match non_empty_users.len() {
                    0 => String::new(),
                    1 => format!("User: {}", q.users[0]),
                    _ => format!("Users: [{}]", q.users.join(", ")),
                };
                let user_html = if !user_display.is_empty() {
                    format!("{}<br>", user_display)
                } else {
                    String::new()
                };

                let error_message = match &q.query_representation {
                    QueryRepresentation::ParseError(error) if !error.is_empty() => format!("Parse Error: {}", error),
                    QueryRepresentation::ParseError(_) => "Error: Query marked invalid, uses BIC-specific functions or functionality not supported in the new compiler.".to_string(),
                    _ => String::new(),
                };

                let error_html = if !error_message.is_empty() {
                    format!("{}<br>", error_message)
                } else {
                    String::new()
                };

                format!(
                    "<li>Time: {}<br>Count: {}<br>{}{}Query: <pre>{}</pre><br></li>",
                    q.timestamp, q.query_count, user_html, error_html, formatted_sql
                )
            })
            .collect::<Vec<String>>()
            .join("");

        format!("{}<ul>{}</ul>", queries_header_html, queries_html)
    } else {
        format!("<h3>{}</h3><p>No Queries found.</p>", label)
    }
}

// process_collections_html takes a label and a list of collections and returns a String
// with the collections formatted as an HTML sorted by access count and time
pub fn process_collections_html(
    label: &str,
    collections: &Option<Vec<(String, String, u32, NaiveDateTime)>>,
) -> String {
    let contents_html = if let Some(collections) = collections {
        let collections_html = collections
            .iter()
            .map(|(db, collection, count, last_accessed)| {
                let db_html = if !db.is_empty() {
                    generate_table_row!("Database", db)
                } else {
                    String::new()
                };
                format!(
                    "<table class='table1'>
                        {}
                        {}
                        {}
                        {}
                     </table>",
                    db_html,
                    generate_table_row!("Collection", collection),
                    generate_table_row!("Access Count", count),
                    generate_table_row!("Last Accessed", last_accessed.format(DATE_TIME_FORMAT)),
                )
            })
            .collect::<Vec<String>>()
            .join("<br>");
        let info = r#"This tab offers insights into the databases and collections accessed by
            the logged queries, making it convenient for identifying the most frequently and
            recently used collections.."#;
        format!(
            "{}<hr><br><h3>{}</h3><ul>{}</ul>",
            info, label, collections_html
        )
    } else {
        format!("<h3>{}</h3><p>No collections found.</p>", label)
    };
    contents_html
}
