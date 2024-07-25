// This file contains utilities for writing HTML elements to a file.
mod log_html;
mod schema_html;

use base64::{prelude::BASE64_STANDARD, Engine};
use build_html::{Container, Html, HtmlContainer, HtmlPage};
use std::{
    fs::File,
    io::Write,
    path::{Path, PathBuf},
};

use image::{DynamicImage, ImageFormat};

use anyhow::Result;
use chrono::prelude::*;

use log_html::{
    process_collections_html, process_complex_types_html, process_query_html, process_summary_html,
};

use self::schema_html::add_schema_analysis_html;

/// generate_html takes a file path, a date, and a LogParseResult and writes the HTML report to a file.
pub fn generate_html(
    file_path: &Path,
    log_parse: Option<&crate::log_parser::LogParseResult>,
    schema_analysis: Option<&crate::schema::SchemaAnalysis>,
    verbose: bool,
    file_stem: &str,
    report_name: &str,
) -> Result<()> {
    let mut report_file = File::create(file_path.join(format!("{file_stem}.html")))?;
    if verbose {
        println!("Writing HTML report to {}", file_path.display());
    }
    Ok(report_file
        .write_all(generate_html_elements(log_parse, schema_analysis, report_name).as_bytes())?)
}

pub fn generate_index(
    file_path: &Path,
    report_name: &str,
    log_analysis_dir_name: &str,
    schema_analysis_dir_name: &str,
) -> Result<()> {
    let mut report_file = File::create(file_path.join("index.html"))?;
    Ok(report_file.write_all(
        generate_index_elements(
            file_path,
            report_name,
            log_analysis_dir_name,
            schema_analysis_dir_name,
        )
        .as_bytes(),
    )?)
}

fn generate_index_elements(
    file_path: &Path,
    report_name: &str,
    log_analysis_dir_name: &str,
    schema_analysis_dir_name: &str,
) -> String {
    let mut page = HtmlPage::new().with_title("Index").with_style(get_css());
    let mut body = Container::default().with_attributes([("class", "container")]);

    // Add MongoDB logo and title
    body.add_raw(get_mongodb_icon_html(report_name));

    body.add_raw(&format!(
        r#"<h2>Schema Reports</h2><ul>{}</ul>"#,
        find_html_files(&file_path.join(schema_analysis_dir_name))
            .unwrap_or_default()
            .iter()
            .fold(String::new(), |acc, path| {
                let relative_path = get_relative_path(file_path, path);
                format!(
                    r#"{}
                    <li><a href="{path}">{path}</a></li>"#,
                    acc,
                    path = relative_path.display()
                )
            })
    ));
    body.add_raw(&format!(
        r#"<h2>Log Reports</h2><ul>{}</ul>"#,
        find_html_files(&file_path.join(log_analysis_dir_name))
            .unwrap_or_default()
            .iter()
            .fold(String::new(), |acc, path| {
                let relative_path = get_relative_path(file_path, path);
                format!(
                    r#"{}
                    <li><a href="{path}">{path}</a></li>"#,
                    acc,
                    path = relative_path.display()
                )
            })
    ));

    page.add_container(body);
    page.to_html_string()
}

fn find_html_files(dir: &Path) -> std::io::Result<Vec<PathBuf>> {
    let mut html_files = Vec::new();

    for entry in std::fs::read_dir(dir)? {
        let entry = entry?;
        let path = entry.path();
        if path.is_file() {
            if let Some(ext) = path.extension() {
                if ext == "html" {
                    html_files.push(path);
                }
            }
        }
    }

    Ok(html_files)
}

fn get_relative_path(base: &Path, path: &Path) -> PathBuf {
    path.strip_prefix(base)
        .ok()
        .map(|p| p.to_path_buf())
        .unwrap()
}

/// encode_image_to_base64 encodes image into a base64 string to be added directly to HTML
fn encode_image_to_base64(image_data: &[u8]) -> String {
    let img = image::load_from_memory(image_data)
        .expect("Failed to load image from bytes")
        .into_rgba8();

    let mut encoded_data: Vec<u8> = Vec::new();
    DynamicImage::ImageRgba8(img)
        .write_to(
            &mut std::io::Cursor::new(&mut encoded_data),
            ImageFormat::Png,
        )
        .unwrap();

    BASE64_STANDARD.encode(&encoded_data)
}

fn get_icon_data() -> &'static [u8] {
    #[cfg(target_os = "windows")]
    {
        include_bytes!(r"..\..\..\resources\html\MongoDB_ForestGreen.png")
    }

    #[cfg(not(target_os = "windows"))]
    {
        include_bytes!("../../../resources/html/MongoDB_ForestGreen.png")
    }
}

fn get_css() -> &'static str {
    #[cfg(target_os = "windows")]
    {
        include_str!(r"..\..\..\resources\html\mongodb.css")
    }

    #[cfg(not(target_os = "windows"))]
    {
        include_str!("../../../resources/html/mongodb.css")
    }
}

fn get_mongodb_icon_html(report_name: &str) -> String {
    let icon_data = get_icon_data();
    let icon_base64 = encode_image_to_base64(icon_data);

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

fn get_tab_string(log_parse: bool, schema_analysis: bool) -> String {
    format!(
        r#"<div class="tab">
    {}
    {}
    </div>"#,
        if log_parse {
            r#"
                <button class="tablinks btn-green" onclick="openTab(event, 'Summary')">Summary</button>
                <button class="tablinks btn-green" onclick="openTab(event, 'Queries')">Queries</button>
                <button class="tablinks btn-green" onclick="openTab(event, 'Collections')">Collections</button>
                <button class="tablinks btn-green" onclick="openTab(event, 'ComplexTypes')">Complex Types</button>
        "#
        } else {
            ""
        },
        if schema_analysis {
            r#"<button class="tablinks btn-green" onclick="openTab(event, 'SchemaAnalysis')">Schema Analysis</button>"#
        } else {
            ""
        }
    )
}

#[test]
fn test_load_js() {
    assert!(get_tab_js().contains("function openTab(evt, tabName)"));
}

fn get_tab_js() -> &'static str {
    #[cfg(target_os = "windows")]
    {
        include_str!(r"..\..\..\resources\html\tab_logic.js")
    }

    #[cfg(not(target_os = "windows"))]
    {
        include_str!("../../../resources/html/tab_logic.js")
    }
}

fn get_version(datetime_str: String) -> String {
    let default = format!("Report was generated at: {datetime_str}");
    option_env!("CARGO_PKG_VERSION").map_or_else(
        || default.clone(),
        |version| format!("Version: {version}. {default}"),
    )
}

pub fn generate_html_elements(
    log_parse: Option<&crate::log_parser::LogParseResult>,
    schema_analysis: Option<&crate::schema::SchemaAnalysis>,
    report_name: &str,
) -> String {
    let datetime_str = Local::now().format("%m-%d-%Y %H:%M").to_string();
    let mut page = HtmlPage::new()
        .with_title(report_name)
        .with_style(get_css());
    let mut body = Container::default().with_attributes([("class", "container")]);

    // Add MongoDB logo
    body.add_raw(get_mongodb_icon_html(report_name));

    // Add version and timestamp
    body.add_raw(format!("<p>{}</p>", get_version(datetime_str)));

    body.add_raw(get_tab_string(
        log_parse.is_some(),
        schema_analysis.is_some(),
    ));

    if let Some(log_parse) = log_parse {
        // Add generated values to the tabs
        let unsupported_queries_html =
            process_query_html("Unsupported", &log_parse.unsupported_queries);
        let collapsible_unsupported_queries_html = if !unsupported_queries_html.is_empty() {
            format!(
                r#"<button class="collapsible btn-green">Show Unsupported Queries</button>
            <div class="content" style="display:none;">
                {}
            </div>
            <script>
            var coll = document.getElementsByClassName("collapsible");
            for (var i = 0; i < coll.length; i++) {{
                coll[i].addEventListener("click", function() {{
                    this.classList.toggle("active");
                    var content = this.nextElementSibling;
                    if (content.style.display === "block") {{
                        content.style.display = "none";
                    }} else {{
                        content.style.display = "block";
                    }}
                }});
            }}
            </script>
            "#,
                unsupported_queries_html
            )
        } else {
            "".to_string()
        };

        body.add_raw(&format!(
            "<div id=\"Queries\" class=\"tabcontent\"><h2>Queries</h2><a href=\"#invalid-queries\"> \
            Jump to Invalid Queries</a><br><br>{}{}{}</div>",
            process_query_html("Valid", &log_parse.valid_queries),
            process_query_html("Invalid", &log_parse.invalid_queries),
            collapsible_unsupported_queries_html
        ));

        body.add_raw(&format!(
            r#"<div id="Collections" class="tabcontent"><h2>Collections</h2>{}</div>"#,
            process_collections_html("Found Collections", &log_parse.collections)
        ));

        body.add_raw(&format!(
            r#"<div id="ComplexTypes" class="tabcontent"><h2>Complex Types</h2>{}</div>"#,
            process_complex_types_html(&log_parse.subpath_fields, &log_parse.array_datasources)
        ));
        body.add_raw(&format!(
            r#"<div id="Summary" class="tabcontent"><h2>Summary</h2>{}</div>"#,
            process_summary_html(log_parse)
        ));
    }

    if schema_analysis.is_some() {
        body.add_raw(format!(
            r#"<div id="SchemaAnalysis" class="tabcontent"><h2>Schema Analysis</h2>{}</div>"#,
            add_schema_analysis_html(schema_analysis).to_html_string()
        ));
    }
    page.add_container(body);
    page.add_raw(&format!(r#"<script>{}</script>"#, get_tab_js()));

    page.to_html_string()
}
