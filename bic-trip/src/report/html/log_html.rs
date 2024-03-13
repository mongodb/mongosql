use chrono::NaiveDateTime;
use std::collections::HashMap;

use crate::log_parser::{LogEntry, QueryRepresentation};
use sqlformat::{format, FormatOptions, QueryParams};

/// process_summary_html generates summary page of database and query information
pub fn process_summary_html(log_parse: &crate::log_parser::LogParseResult) -> String {
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

pub fn process_complex_types_html(
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
pub fn process_query_html(label: &str, queries: &Option<Vec<LogEntry>>) -> String {
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
pub fn process_collections_html(
    label: &str,
    collections: &Option<Vec<(String, String, u32, NaiveDateTime)>>,
) -> String {
    let contents_html = if let Some(collections) = collections {
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
    };
    contents_html
}
