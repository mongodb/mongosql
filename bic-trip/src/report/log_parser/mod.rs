// The log_parser module is responsible for parsing bic log files and returning the results
// It contains the structures necessary to hold parsing results for display to user,
// and is used in both the html and csv modules.
use std::{
    collections::HashMap,
    fs::File,
    io::{BufRead, BufReader},
};

use chrono::prelude::*;
use serde::Serialize;

// LogParseResult is a struct that holds the results of parsing a log file
pub struct LogParseResult {
    pub valid_queries: Option<Vec<LogEntry>>,
    pub invalid_queries: Option<Vec<LogEntry>>,
    pub collections: Option<Vec<(String, String, u32, chrono::NaiveDateTime)>>,
}

// QueryType is an enum that represents the type of query parsed from the log file
#[derive(Debug, PartialEq, Clone)]
pub enum QueryType {
    Select,
    Set,
    Show,
}

impl QueryType {
    fn from_query(query: &str) -> Option<QueryType> {
        let query = clean_query(query);
        if query.starts_with("SELECT") | query.starts_with("select") {
            Some(QueryType::Select)
        } else if query.starts_with("SET") | query.starts_with("set") {
            Some(QueryType::Set)
        } else if query.starts_with("SHOW") | query.starts_with("show") {
            Some(QueryType::Show)
        } else {
            None
        }
    }
}

impl std::fmt::Display for QueryType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            QueryType::Select => write!(f, "SELECT"),
            QueryType::Set => write!(f, "SET"),
            QueryType::Show => write!(f, "SHOW"),
        }
    }
}

// LogEntry represents a single log entry from the log file
// and contains the information we know about the query at that time.
// Currently, serialization is skipped for the query representation, but
// is a TODO for future work to show the user the reason for a parse failure.
#[derive(Debug, PartialEq, Clone, Serialize)]
pub struct LogEntry {
    pub timestamp: chrono::NaiveDateTime,
    pub query: String,
    #[serde(skip)]
    pub query_type: QueryType,
    pub query_count: u32,
    #[serde(skip)]
    pub query_representation: Option<mongosql::ast::Query>,
}

// clean_query takes a query string and removes any escape characters and trailing semicolons
fn clean_query(query: &str) -> String {
    let mut cleaned_query = query
        .replace("\\n", " ")
        .replace("\\r", " ")
        .replace("\\t", " ");
    if cleaned_query.ends_with(';') {
        cleaned_query.pop();
    }
    cleaned_query
}

/// parse_line parses a line from the log file and returns an `Option<LogEntry>`
fn parse_line(line: &str) -> Option<LogEntry> {
    let timestamp_end = line.find('+')?;
    let timestamp = &line[..timestamp_end];
    let timestamp = NaiveDateTime::parse_from_str(timestamp, "%Y-%m-%dT%H:%M:%S%.f").ok()?;

    let query_start = line.find("parsing \"")?;
    let query = clean_query(&line[query_start + "parsing \"".len()..line.len() - 1]).to_string();

    let query_type = QueryType::from_query(&query)?;

    let query_representation = match mongosql::parse_query(&query) {
        Ok(query) => Some(query),
        Err(_) => None,
    };

    Some(LogEntry {
        timestamp,
        query,
        query_type,
        query_representation,
        query_count: 1,
    })
}

// process_logs takes a list of file paths and returns a LogParseResult
pub fn process_logs(paths: &[String]) -> Result<LogParseResult, String> {
    let mut all_valid_queries = vec![];
    let mut all_invalid_queries = vec![];

    let mut all_queries: HashMap<String, LogEntry> = HashMap::new();

    for path in paths {
        let file = File::open(path).map_err(|e| e.to_string())?;
        let reader = BufReader::new(file);

        for line in reader.lines() {
            let line = line.map_err(|e| e.to_string())?;
            if line.contains("EVALUATOR") {
                if let Some(query) = parse_line(&line) {
                    all_queries
                        .entry(query.query.clone())
                        .and_modify(|entry: &mut LogEntry| entry.query_count += 1u32)
                        .or_insert(query);
                }
            }
        }
    }
    for query in all_queries.values().collect::<Vec<&LogEntry>>() {
        match (&query.query_type, &query.query_representation) {
            (QueryType::Select, Some(_)) => {
                all_valid_queries.push(query.clone());
            }
            (_, _) => all_invalid_queries.push(query.clone()),
        }
    }

    all_valid_queries.sort_by(|a, b| b.timestamp.cmp(&a.timestamp));
    all_invalid_queries.sort_by(|a, b| b.timestamp.cmp(&a.timestamp));

    let mut collections = HashMap::<(String, String), (u32, NaiveDateTime)>::new();

    for query in all_valid_queries.iter() {
        if let Some(repr) = &query.query_representation {
            for collection in mongosql::ast::visitors::get_collection_sources(repr.clone()) {
                let key = (
                    collection.database.unwrap_or_default(),
                    collection.collection.clone(),
                );
                collections
                    .entry(key)
                    .and_modify(|e| {
                        e.0 += 1;
                        e.1 = e.1.max(query.timestamp);
                    })
                    .or_insert((1, query.timestamp));
            }
        }
    }

    let collections_vec = collections
        .into_iter()
        .map(|((db, collection), (count, last_accessed))| (db, collection, count, last_accessed))
        .collect::<Vec<(String, String, u32, chrono::NaiveDateTime)>>();

    Ok(LogParseResult {
        valid_queries: (!all_valid_queries.is_empty()).then_some(all_valid_queries),
        invalid_queries: (!all_invalid_queries.is_empty()).then_some(all_invalid_queries),
        collections: (Some(collections_vec)),
    })
}

#[test]
fn test_parse_line() {
    let line1 = r#"2021-11-05T17:25:18.842+1100 I EVALUATOR  [conn3] parsing "SET character_set_results = NULL""#;
    let line2 = r#"2021-11-05T17:25:18.842+1100 I EVALUATOR  [conn7] parsing "SELECT * FROM `test`.`test` WHERE `test`.`test`.`_id` = 1 LIMIT 1""#;
    let line3 = r#"2021-11-05T17:25:26.713+1100 I EVALUATOR  [conn10] parsing "SHOW KEYS FROM `business_events`.`auth0-anz-smartchoice_data_details_response_body_identities`""#;
    let line4 =
        r#"2021-11-05T17:25:26.713+1100 I EVALUATOR  [conn10] some operation we don't know about"#;

    let res1 = parse_line(line1).unwrap();
    let res2 = parse_line(line2).unwrap();
    let res3 = parse_line(line3).unwrap();

    assert_eq!(
        res1.timestamp,
        chrono::NaiveDate::from_ymd_opt(2021, 11, 5)
            .unwrap()
            .and_hms_milli_opt(17, 25, 18, 842)
            .unwrap()
    );
    assert_eq!(res1.query, "SET character_set_results = NULL");
    assert_eq!(res1.query_type, QueryType::Set);

    assert_eq!(
        res2.timestamp,
        chrono::NaiveDate::from_ymd_opt(2021, 11, 5)
            .unwrap()
            .and_hms_milli_opt(17, 25, 18, 842)
            .unwrap()
    );
    assert_eq!(
        res2.query,
        "SELECT * FROM `test`.`test` WHERE `test`.`test`.`_id` = 1 LIMIT 1"
    );
    assert_eq!(res2.query_type, QueryType::Select);

    assert_eq!(
        res3.timestamp,
        chrono::NaiveDate::from_ymd_opt(2021, 11, 5)
            .unwrap()
            .and_hms_milli_opt(17, 25, 26, 713)
            .unwrap()
    );
    assert_eq!(
        res3.query,
        "SHOW KEYS FROM `business_events`.`auth0-anz-smartchoice_data_details_response_body_identities`"
    );
    assert_eq!(res3.query_type, QueryType::Show);
    assert!(parse_line(line4).is_none());
}
