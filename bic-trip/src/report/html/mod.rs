// This file contains utilities for writing HTML elements to a file.

use std::{fs::File, io::Write, path::Path};

use build_html::{Container, ContainerType, Html, HtmlContainer, HtmlPage};

use crate::log_parser::{LogEntry, QueryRepresentation};
use chrono::prelude::*;

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

// process_query_html takes a label and a list of LogEntry and returns a Container
// with the entries formatted as an HTML unordered list
fn process_query_html(label: &str, queries: &Option<Vec<LogEntry>>) -> Container {
    if let Some(mut queries) = queries.clone() {
        queries.sort_by(|a, b| b.query_count.cmp(&a.query_count));
        let queries_html = queries
            .iter()
            .map(|q| match &q.query_representation {
                QueryRepresentation::Query(_) => format!(
                    "<li>Time: {}</br>Count: {}</br>Query: {}</br></br></li>",
                    q.timestamp, q.query_count, q.query
                ),
                QueryRepresentation::ParseError(error) => format!(
                    "<li>Time: {}</br>Count: {}</br>Query: {}</br>Parse Error: {}</br></br></li>",
                    q.timestamp, q.query_count, q.query, error
                ),
            })
            .collect::<Vec<String>>()
            .concat();
        Container::default()
            .with_header(3, label)
            .with_raw(format!("<ul>{}</ul>", queries_html))
    } else {
        Container::default()
    }
}

// process_collections_html takes a label and a list of collections and returns a Container
// with the collections formatted as an HTML unordered list
fn process_collections_html(
    label: &str,
    collections: &Option<Vec<(String, String, u32, NaiveDateTime)>>,
) -> Container {
    if let Some(mut collections) = collections.clone() {
        collections.sort_by(|a, b| b.2.cmp(&a.2));
        let collections_html = collections
            .iter()
            .map(|(db, collection, count, last_accessed)| {
                format!(
                    "<li>Database: {}</br>Collection: {}</br>Access Count: {}</br>Last Accessed: {}</br></br></li>",
                    db,
                    collection,
                    count,
                    last_accessed.format("%m-%d-%Y %H:%M:%S")
                )
            })
            .collect::<Vec<String>>()
            .concat();
        Container::default()
            .with_header(3, label)
            .with_raw(format!("<ul>{}</ul>", collections_html))
    } else {
        Container::default()
    }
}

// generate_html_elements takes a LogParseResult and returns an HTML string
fn generate_html_elements(
    log_parse: &crate::log_parser::LogParseResult,
    report_name: &str,
) -> String {
    let invalid_queries_html = process_query_html("Invalid", &log_parse.invalid_queries);
    let valid_queries_html = process_query_html("Valid", &log_parse.valid_queries);
    let collections_html = process_collections_html("Found Collections", &log_parse.collections);
    let datetime_str = Local::now().format("%m-%d-%Y %H:%M").to_string();
    let timestamp_html = format!("<p>Report was generated at: {}</p>", datetime_str);

    HtmlPage::new()
        .with_title(report_name)
        .with_header(1, report_name)
        .with_paragraph(timestamp_html)
        .with_container(
            Container::new(ContainerType::Div)
                .with_header_attr(2, "Collections", [("id", "collections")])
                .with_container(collections_html),
        )
        .with_container(
            Container::new(ContainerType::Div)
                .with_header_attr(2, "Queries", [("id", "queries")])
                .with_container(valid_queries_html)
                .with_container(invalid_queries_html),
        )
        .to_html_string()
}
