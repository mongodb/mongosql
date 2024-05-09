use report::html::generate_html_elements;
use report::log_parser::handle_logs;
use scraper::{Html, Selector};
mod common;

// log_parse_counts_test parses various logs and verifies that the counts of the queries, complex
// fields, and collections are correct.  Optionally, the user list will be verified.
#[test]
fn log_parse_counts_test() {
    let base_dir = env!("CARGO_MANIFEST_DIR");
    let dir_path = std::path::PathBuf::from(format!("{}/testfiles/tests", base_dir));
    let tests = common::load_transition_report_test_files(dir_path).unwrap();

    for test in tests {
        let result = handle_logs(Some(test.log_dir), true).unwrap().unwrap();
        let valid_queries = &result.valid_queries.unwrap_or_default();
        assert_eq!(&test.valid_query_count, &valid_queries.len());

        if let Some(expected_users_lists) = &test.valid_query_users {
            assert_eq!(valid_queries.len(), expected_users_lists.len());

            for (index, log_entry) in valid_queries.iter().enumerate() {
                let mut actual_users = log_entry.users.clone();
                actual_users.sort();
                let mut expected_users = expected_users_lists[index].clone();
                expected_users.sort();

                assert_eq!(expected_users, actual_users);
            }
        }

        let invalid_queries = &result.invalid_queries.unwrap_or_default();
        assert_eq!(&test.invalid_query_count, &invalid_queries.len());

        let unsupported_queries = &result.unsupported_queries.unwrap_or_default();
        assert_eq!(&test.unsupported_query_count, &unsupported_queries.len());

        let subpath_fields = &result.subpath_fields.unwrap_or_default();
        assert_eq!(&test.subpath_fields_count, &subpath_fields.len());

        let array_datasources = &result.array_datasources.unwrap_or_default();
        assert_eq!(&test.array_datasources_count, &array_datasources.len());
    }
}

// log_parse_html_test will process the logs in the multi_logs/ directory and verifies that the
// expected HTML elements exist by checking that the counts match
#[test]
fn log_parse_html_test() {
    let base_dir = env!("CARGO_MANIFEST_DIR");
    let multi_logs_test =
        std::path::PathBuf::from(format!("{}/testfiles/tests/multi_logs_test.yml", base_dir));
    let test = common::parse_transition_report_yaml_file(multi_logs_test).unwrap();

    let result = handle_logs(Some(test.log_dir), true).unwrap().unwrap();
    let html_content = generate_html_elements(Some(&result), None, "test report");

    let document = Html::parse_document(&html_content);

    // Queries Tab
    let valid_queries_selector = Selector::parse("div#Queries > ul:first-of-type > li").unwrap();
    assert_eq!(
        test.valid_query_count,
        document.select(&valid_queries_selector).count()
    );
    let invalid_queries_selector = Selector::parse("div#invalid-queries + ul li").unwrap();
    assert_eq!(
        test.invalid_query_count,
        document.select(&invalid_queries_selector).count()
    );

    // Collections Tab
    let collections_selector = Selector::parse("div#Collections ul table").unwrap();
    assert_eq!(
        test.collections_count,
        document.select(&collections_selector).count()
    );

    // Complex Types Tab
    let fields_selector = Selector::parse("div#ComplexTypes div:first-of-type table").unwrap();
    let fields_tables = document.select(&fields_selector).collect::<Vec<_>>();
    assert_eq!(test.subpath_fields_count, fields_tables.len());
    let arrays_selector = Selector::parse("div#ComplexTypes div:last-of-type table").unwrap();
    let arrays_tables = document.select(&arrays_selector).collect::<Vec<_>>();
    assert_eq!(test.array_datasources_count, arrays_tables.len());

    // Checks that the valid/invalid query counts in the Summary tab is correct
    let summary_tab_selector = Selector::parse("#Summary").unwrap();
    let summary_tab = document.select(&summary_tab_selector).next().unwrap();

    let div_selector = Selector::parse("div").unwrap();
    let queries_div = summary_tab.select(&div_selector).next().unwrap();
    let first_table = Selector::parse("table.table1").unwrap();
    let second_table = queries_div.select(&first_table).next().unwrap();

    let valid_count_selector =
        Selector::parse("tr:nth-child(1) td:nth-child(2) span.highlight").unwrap();
    let invalid_count_selector =
        Selector::parse("tr:nth-child(2) td:nth-child(2) span.highlight").unwrap();

    if let Some(valid_count_element) = second_table.select(&valid_count_selector).next() {
        let valid_count = valid_count_element.text().collect::<String>();
        let valid_count_num = valid_count
            .parse::<usize>()
            .expect("valid_count does not contain a valid usize value");
        assert_eq!(test.valid_query_count, valid_count_num);
    }

    if let Some(invalid_count_element) = second_table.select(&invalid_count_selector).next() {
        let invalid_count = invalid_count_element.text().collect::<String>();
        let invalid_count_num = invalid_count
            .parse::<usize>()
            .expect("invalid_count does not contain a valid usize value");
        assert_eq!(test.invalid_query_count, invalid_count_num);
    }
}
