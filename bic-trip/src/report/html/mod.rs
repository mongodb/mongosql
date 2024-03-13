// This file contains utilities for writing HTML elements to a file.
mod log_html;
mod schema_html;

use base64::{prelude::BASE64_STANDARD, Engine};
use build_html::{Html, HtmlContainer, HtmlPage};
use std::{fs, fs::File, io::Write, path::Path};

use image::{DynamicImage, ImageOutputFormat};

use anyhow::Result;
use chrono::prelude::*;

use log_html::{
    process_collections_html, process_complex_types_html, process_query_html, process_summary_html,
};

use self::schema_html::add_schema_analysis_html;

/// generate_html takes a file path, a date, and a LogParseResult and writes the HTML report to a file.
pub fn generate_html(
    file_path: &Path,
    date: &str,
    log_parse: &crate::log_parser::LogParseResult,
    schema_analysis: &Option<crate::schema::SchemaAnalysis>,
    file_stem: &str,
    report_name: &str,
) -> Result<()> {
    let mut report_file = File::create(file_path.join(format!("{file_stem}_{date}.html")))?;
    Ok(report_file
        .write_all(generate_html_elements(log_parse, schema_analysis, report_name).as_bytes())?)
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

fn get_mongodb_icon_html(base_dir: &str, report_name: &str) -> String {
    let logo_path = format!("{base_dir}/resources/html/MongoDB_ForestGreen.png");
    let icon_base64 = encode_image_path_to_base64(&logo_path).expect("Incorrect path");
    format!(
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
    )
}

fn generate_html_elements(
    log_parse: &crate::log_parser::LogParseResult,
    schema_analysis: &Option<crate::schema::SchemaAnalysis>,
    report_name: &str,
) -> String {
    let base_dir = env!("CARGO_MANIFEST_DIR");
    let tab_js_path = format!("{base_dir}/resources/html/tab_logic.js");
    let datetime_str = Local::now().format("%m-%d-%Y %H:%M").to_string();

    let mut page = HtmlPage::new().with_title(report_name);

    // Add MongoDB logo
    page.add_raw(get_mongodb_icon_html(base_dir, report_name));

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
        process_query_html("Valid", &log_parse.valid_queries),
        process_query_html("Invalid", &log_parse.invalid_queries)
    ));

    page.add_raw(&format!(
        r#"<div id="Collections" class="tabcontent"><h2>Collections</h2>{}</div>"#,
        process_collections_html("Found Collections", &log_parse.collections)
    ));

    page.add_raw(&format!(
        r#"<div id="ComplexTypes" class="tabcontent"><h2>Complex Types</h2>{}</div>"#,
        process_complex_types_html(&log_parse.subpath_fields, &log_parse.array_datasources)
    ));

    if schema_analysis.is_some() {
        page.add_raw(format!(
            r#"<div id="Schema" class="tabcontent"><h2>Schema</h2>{}</div>"#,
            add_schema_analysis_html(schema_analysis).to_html_string()
        ));
    }

    page.add_raw(&format!(
        r#"<div id="Summary" class="tabcontent"><h2>Summary</h2>{}</div>"#,
        process_summary_html(log_parse)
    ));

    page.add_raw(&format!(
        r#"<script>{}</script>"#,
        fs::read_to_string(&tab_js_path)
            .unwrap_or_else(|_| panic!("Failed to read file {}", tab_js_path))
    ));
    page.to_html_string()
}
