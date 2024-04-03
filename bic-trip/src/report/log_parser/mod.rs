// The log_parser module is responsible for parsing bic log files and returning the results
// It contains the structures necessary to hold parsing results for display to user,
// and is used in both the html and csv modules.
use std::{
    collections::HashMap,
    fmt,
    fs::File,
    io::{BufRead, BufReader},
};

use anyhow::{anyhow, Context, Result};
use chrono::prelude::*;
use mongosql::usererror::UserError;

// LogParseResult is a struct that holds the results of parsing a log file
pub struct LogParseResult {
    pub valid_queries: Option<Vec<LogEntry>>,
    pub invalid_queries: Option<Vec<LogEntry>>,
    pub unsupported_queries: Option<Vec<LogEntry>>,
    pub collections: Option<Vec<(String, String, u32, chrono::NaiveDateTime)>>,
    pub subpath_fields: Option<Vec<(SubpathField, u32, chrono::NaiveDateTime)>>,
    pub array_datasources: Option<Vec<(String, String, u32, chrono::NaiveDateTime)>>,
}

// SubpathField is a struct that holds a subpath field and its datasource
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct SubpathField {
    pub db: String,
    pub collection: String,
    pub subpath_field: String,
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
#[derive(Debug, PartialEq, Clone)]
pub struct LogEntry {
    pub timestamp: chrono::NaiveDateTime,
    pub query: String,
    pub query_type: QueryType,
    pub query_count: u32,
    pub query_representation: QueryRepresentation,
    pub users: Vec<String>,
}

// QueryRepresentation holds the query AST if parsing was successful or
// the parse error if query parsing failed
#[derive(Debug, PartialEq, Clone)]
pub enum QueryRepresentation {
    Query(mongosql::ast::Query),
    ParseError(String),
}

impl fmt::Display for QueryRepresentation {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            QueryRepresentation::Query(_) => {
                write!(f, "")
            }
            QueryRepresentation::ParseError(parse_error) => {
                write!(f, "ParseError: {}", parse_error)
            }
        }
    }
}

// sort_by_access_and_time will sort the data vector by access_count, if those are equal
// it will then sort with time
fn sort_by_access_and_time<T, F1, F2>(data: &mut [T], access_count_fn: F1, time_fn: F2)
where
    F1: Fn(&T) -> u32,
    F2: Fn(&T) -> chrono::NaiveDateTime,
{
    data.sort_by(|a, b| {
        let access_count_cmp = access_count_fn(b).cmp(&access_count_fn(a));
        let time_cmp = time_fn(b).cmp(&time_fn(a));
        access_count_cmp.then_with(|| time_cmp)
    });
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
        Ok(query_ast) => QueryRepresentation::Query(query_ast),
        Err(err) => QueryRepresentation::ParseError(err.user_message().unwrap_or_default()),
    };

    Some(LogEntry {
        timestamp,
        query,
        query_type,
        query_count: 1,
        query_representation,
        users: Vec::new(),
    })
}

pub fn handle_logs(input: Option<String>, quiet: bool) -> Result<Option<LogParseResult>> {
    if let Some(input) = input {
        let metadata = std::fs::metadata(&input)
            .with_context(|| format!("Failed to read  input file or directory: {:?}", &input))?;
        let mut paths = Vec::new();

        if metadata.is_file() {
            paths.push(input.clone());
        } else if metadata.is_dir() {
            for entry in std::fs::read_dir(input)? {
                let entry = entry?;
                if entry.path().is_file() {
                    paths.push(entry.path().to_string_lossy().into_owned());
                }
            }
        }
        Ok(Some(process_logs(&paths, quiet)?))
    } else {
        Ok(None)
    }
}

// process_logs takes a list of file paths and returns a LogParseResult
pub fn process_logs(paths: &[String], quiet: bool) -> Result<LogParseResult> {
    let mut all_valid_queries = vec![];
    let mut all_invalid_queries = vec![];
    let mut all_unsupported_queries = vec![];
    let mut all_queries: HashMap<String, LogEntry> = HashMap::new();
    let mut connections: HashMap<String, String> = HashMap::new();

    for path in paths {
        let file = File::open(path)?;
        let reader = BufReader::new(file);

        let mut lines = reader.lines().peekable();

        let line = lines.peek();
        if line.is_none() {
            continue;
        }
        if line.unwrap().is_err() {
            if !quiet {
                eprintln!(
                    "Ignoring file: {:?} because:\n  {:?}\ncontinuing...",
                    path,
                    line.unwrap().as_ref().err().unwrap()
                );
            }
            continue;
        }

        for line in lines {
            let line = line?;
            if line.contains("configuring client authentication for principal") {
                let conn_start = line
                    .find('[')
                    .ok_or_else(|| anyhow!("Missing '[' in line"))?;
                let conn_end = line
                    .find(']')
                    .ok_or_else(|| anyhow!("Missing ']' in line"))?;
                let connection_id = &line[conn_start + 1..conn_end];
                let user_start = line
                    .find("principal")
                    .ok_or_else(|| anyhow!("Missing 'principal' in line"))?
                    + "principal".len()
                    + 1;
                let user = &line[user_start..].trim();

                connections.insert(connection_id.to_string(), user.to_string());
            } else if line.contains("EVALUATOR") {
                if let Some(query) = parse_line(&line) {
                    let conn_start = line
                        .find('[')
                        .ok_or_else(|| anyhow!("Missing '[' in line"))?;
                    let conn_end = line
                        .find(']')
                        .ok_or_else(|| anyhow!("Missing ']' in line"))?;
                    let connection_id = &line[conn_start + 1..conn_end];
                    let user = connections.get(connection_id).cloned().unwrap_or_default();

                    let entry = all_queries.entry(query.query.clone()).or_insert_with(|| {
                        let mut new_query = query.clone();
                        new_query.users.push(user.clone());
                        new_query
                    });

                    if !entry.users.contains(&user) {
                        entry.users.push(user);
                    }
                    entry.query_count += 1;
                }
            }
        }
    }

    let mut all_queries = all_queries.values().collect::<Vec<&LogEntry>>();
    sort_by_access_and_time(
        &mut all_queries,
        |entry| entry.query_count,
        |entry| entry.timestamp,
    );
    for query in all_queries {
        let is_unsupported = query.query_type == QueryType::Set
            || query.query_type == QueryType::Show
            // Marking queries with `INFORMATION_SCHEMA` as invalid since they are specific to BIC
            || query.query.to_lowercase().contains("information_schema");
        if is_unsupported {
            all_unsupported_queries.push(query.clone());
        } else {
            match &query.query_representation {
                QueryRepresentation::Query(_) => {
                    all_valid_queries.push(query.clone());
                }
                QueryRepresentation::ParseError(_) => {
                    all_invalid_queries.push(query.clone());
                }
            }
        }
    }

    let mut collections = HashMap::<(String, String), (u32, NaiveDateTime)>::new();
    let mut subpath_fields_map: HashMap<SubpathField, (u32, NaiveDateTime)> = HashMap::new();
    let mut array_datasources = vec![];

    for query in all_valid_queries.iter() {
        if let QueryRepresentation::Query(repr) = &query.query_representation {
            let subpath_fields = mongosql::ast::visitors::get_subpath_fields(repr.clone());
            let collection_sources = mongosql::ast::visitors::get_collection_sources(repr.clone());

            for subpath_vec in subpath_fields {
                if let Some((identifier, _)) = subpath_vec.split_first() {
                    let field_name = subpath_vec.join(".");

                    // Check if the initial value in the subpath matches the collection name or alias
                    // If not, see if there is only one collection in the CollectionSource that it
                    // can be associated with.  If neither of these are true, then we cannot find
                    // the CollectionSource for the subpath and don't add it.
                    let matching_collections: Vec<_> = collection_sources
                        .iter()
                        .filter(|cs| {
                            cs.collection == *identifier || cs.alias.as_deref() == Some(identifier)
                        })
                        .collect();
                    let subpath_field_opt = match matching_collections.len() {
                        1 => {
                            let collection = matching_collections[0];
                            Some(SubpathField {
                                db: collection.database.clone().unwrap_or_default(),
                                collection: collection.collection.clone(),
                                subpath_field: field_name,
                            })
                        }
                        _ => {
                            if collection_sources.len() == 1 {
                                let collection = &collection_sources[0];
                                Some(SubpathField {
                                    db: collection.database.clone().unwrap_or_default(),
                                    collection: collection.collection.clone(),
                                    subpath_field: field_name,
                                })
                            } else {
                                None
                            }
                        }
                    };
                    if let Some(subpath_field) = subpath_field_opt {
                        subpath_fields_map
                            .entry(subpath_field)
                            .and_modify(|(count, last_accessed)| {
                                *count += 1;
                                *last_accessed = std::cmp::max(*last_accessed, query.timestamp);
                            })
                            .or_insert((1, query.timestamp));
                    }
                }
            }

            for collection in collection_sources {
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

    let mut collections_vec = collections
        .into_iter()
        .map(|((db, collection), (count, last_accessed))| (db, collection, count, last_accessed))
        .collect::<Vec<(String, String, u32, chrono::NaiveDateTime)>>();

    sort_by_access_and_time(&mut collections_vec, |entry| entry.2, |entry| entry.3);

    // Add collections that have an underscore '_' to our array_datasources
    for (db, collection, count, last_accessed) in collections_vec.iter() {
        if collection.contains('_') {
            array_datasources.push((db.clone(), collection.clone(), *count, *last_accessed));
        }
    }

    let mut subpath_fields_vec = subpath_fields_map
        .into_iter()
        .map(|(subpath_field, (count, last_accessed))| (subpath_field, count, last_accessed))
        .collect::<Vec<(SubpathField, u32, NaiveDateTime)>>();
    sort_by_access_and_time(&mut subpath_fields_vec, |entry| entry.1, |entry| entry.2);

    Ok(LogParseResult {
        valid_queries: (!all_valid_queries.is_empty()).then_some(all_valid_queries),
        invalid_queries: (!all_invalid_queries.is_empty()).then_some(all_invalid_queries),
        unsupported_queries: (!all_unsupported_queries.is_empty())
            .then_some(all_unsupported_queries),
        collections: (Some(collections_vec)),
        subpath_fields: (Some(subpath_fields_vec)),
        array_datasources: (Some(array_datasources)),
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
