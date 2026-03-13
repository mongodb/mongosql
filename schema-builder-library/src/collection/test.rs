use crate::{CollectionDoc, CollectionInfo};
impl From<&str> for CollectionDoc {
    fn from(name: &str) -> Self {
        CollectionDoc {
            name: name.to_string(),
            ..Default::default()
        }
    }
}

macro_rules! vec_collection_docs {
        ($($x:expr),* $(,)?) => {
            vec![$(CollectionDoc::from($x)),*]
        };
    }

macro_rules! actual {
    ($input:expr, $db:expr, $include_list:expr, $exclude_list:expr $(,)?) => {{
        $input
            .iter()
            .filter(|c| {
                CollectionInfo::should_consider($db, c, $include_list, $exclude_list).unwrap()
            })
            .cloned()
            .collect::<Vec<CollectionDoc>>()
    }};
}

fn glob_it(patterns: Vec<&str>) -> Vec<glob::Pattern> {
    patterns
        .into_iter()
        .map(glob::Pattern::new)
        .map(Result::unwrap)
        .collect()
}

#[test]
fn test_inclusion_glob() {
    let include_list = glob_it(vec!["mydb.*"]);
    let exclude_list = glob_it(vec!["mydb.excluded"]);
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included", "excludeded", "subcollection");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}

#[test]
fn test_exclusion_glob() {
    let include_list = glob_it(vec!["mydb.excluded"]);
    let exclude_list = glob_it(vec!["mydb.*", "otherdb.*"]);
    let input = vec_collection_docs!("included", "excluded", "excludeded", "system.views");

    let expected: Vec<CollectionDoc> = vec![];
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}

#[test]
fn test_include_empty_exclude_empty() {
    let include_list = vec![];
    let exclude_list = vec![];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = input.clone();
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}

#[test]
fn test_include_empty_exclude_contents() {
    let include_list = vec![];
    let exclude_list = glob_it(vec!["mydb.excluded"]);
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included", "excludeded", "subcollection");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}

#[test]
fn test_include_contents_exclude_empty() {
    let include_list = glob_it(vec!["mydb.included", "mydb.excluded"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included", "excluded");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_dunderscore_dbs_and_colls_are_excluded_by_default() {
    let include_list = vec![];
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = vec_collection_docs!("coll");
    let dunderscore_mydb_expected = vec_collection_docs!();
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );
    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        dunderscore_mydb_expected
    );
}

#[test]
fn test_include_list_pattern_equals_name_but_does_not_match_it() {
    let include_list = glob_it(vec!["__mydb_[abc].__coll_[abc]"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__coll_[abc]");

    assert_eq!(
        actual!(input, "__mydb_[abc]", &include_list, &exclude_list),
        vec_collection_docs!()
    );
}

#[test]
fn test_include_list_pattern_does_not_equal_name_but_does_match_it() {
    let include_list = glob_it(vec!["__mydb_[abc].__coll_[abc]"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__coll_a");

    assert_eq!(
        actual!(input, "__mydb_a", &include_list, &exclude_list),
        vec_collection_docs!("__coll_a")
    );
}

#[test]
fn test_include_dunderscore_colls_in_non_dunderscore_dbs() {
    let include_list = glob_it(vec!["*.__*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = vec_collection_docs!("__coll");
    let dunderscore_mydb_expected = vec_collection_docs!();

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        dunderscore_mydb_expected
    );
}

#[test]
fn test_include_non_dunderscore_colls_in_dunderscore_dbs() {
    let include_list = glob_it(vec!["__*.*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = vec_collection_docs!();
    let dunderscore_mydb_expected = vec_collection_docs!("coll");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        dunderscore_mydb_expected
    );
}

#[test]
fn test_include_dunderscore_colls_in_regular_dbs_and_regular_colls_in_dunderscore_dbs() {
    let include_list = glob_it(vec!["__*.*", "*.__*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll", "foo", "__foo");

    let mydb_expected = vec_collection_docs!("__coll", "__foo");
    let dunderscore_mydb_expected = vec_collection_docs!("coll", "foo");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        dunderscore_mydb_expected
    );
}

#[test]
fn test_include_colls_from_a_specific_db_and_all_dunderscore_colls_in_regular_dbs() {
    let include_list = glob_it(vec!["mydb.*", "*.__*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = input.clone();
    let foo_expected = vec_collection_docs!("__coll");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    assert_eq!(
        actual!(input, "foo", &include_list, &exclude_list),
        foo_expected
    );
}

#[test]
fn test_include_all_dunderscore_dbs_and_a_specific_db() {
    let include_list = glob_it(vec!["mydb.*", "__*.*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let expected = vec_collection_docs!("coll");
    let foo_expected = vec_collection_docs!();

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "__foo", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "foo", &include_list, &exclude_list),
        foo_expected
    );
}

#[test]
fn test_include_specific_dunderscore_dbs_only() {
    let include_list = glob_it(vec!["__db1.*", "__db2.*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let expected = vec_collection_docs!("coll");
    let db3_db4_expected = vec_collection_docs!();

    assert_eq!(
        actual!(input, "__db1", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "__db2", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "__db3", &include_list, &exclude_list),
        db3_db4_expected
    );

    assert_eq!(
        actual!(input, "db4", &include_list, &exclude_list),
        db3_db4_expected
    );
}

#[test]
fn test_include_specific_dunderscore_dbs_and_dunderscore_colls() {
    let include_list = glob_it(vec!["__db1.__coll", "__db2.coll", "__db3.*", "db4.__coll"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let db1_db4_expected = vec_collection_docs!("__coll");
    let db2_db3_expected = vec_collection_docs!("coll");
    let foo_expected = vec_collection_docs!();

    assert_eq!(
        actual!(input, "__db1", &include_list, &exclude_list),
        db1_db4_expected
    );

    assert_eq!(
        actual!(input, "__db2", &include_list, &exclude_list),
        db2_db3_expected
    );

    assert_eq!(
        actual!(input, "__db3", &include_list, &exclude_list),
        db2_db3_expected
    );

    assert_eq!(
        actual!(input, "db4", &include_list, &exclude_list),
        db1_db4_expected
    );

    assert_eq!(
        actual!(input, "foo", &include_list, &exclude_list),
        foo_expected
    );
}

#[test]
fn test_automatically_exclude_dunderscore_colls_in_included_db_if_not_explicitly_included() {
    let include_list = glob_it(vec!["mydb.*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = vec_collection_docs!("coll");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );
}

#[test]
fn test_include_is_additive_for_dunderscores() {
    // Include is additive for dunderscore collections
    let include_list = glob_it(vec!["mydb.*", "mydb.__*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = input.clone();

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    // Include is additive for dunderscore DBs.
    let include_list = glob_it(vec!["__*.*", "*.*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll");
    let mydb_expected = vec_collection_docs!("coll");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        mydb_expected
    );
}

#[test]
fn test_exclude_takes_precedence_over_include_for_dunderscores() {
    // Exclude takes precedence over include for dunderscore collections
    let include_list = glob_it(vec!["mydb.*", "mydb.__*"]);
    let exclude_list = glob_it(vec!["mydb.__foo"]);
    let input = vec_collection_docs!("coll", "__coll", "__foo");
    let mydb_expected = vec_collection_docs!("coll", "__coll");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        mydb_expected
    );

    // Exclude takes precedence over include for dunderscore DBs
    let include_list = glob_it(vec!["mydb.*", "__*.*"]);
    let exclude_list = glob_it(vec!["__foo.*"]);
    let input = vec_collection_docs!("coll", "__coll",);
    let expected = vec_collection_docs!("coll");
    let foo_expected = vec_collection_docs!();

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "__foo", &include_list, &exclude_list),
        foo_expected
    );
}

#[test]
fn test_include_coll_for_all_dbs() {
    let include_list = glob_it(vec!["*.coll", "__*.coll"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("coll", "__coll", "__foo");
    let expected = vec_collection_docs!("coll");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_brackets_include_underscore() {
    let include_list = glob_it(vec![
        "[1234_][32142546_]mydb.[1234_][32142546_]coll",
        "[1234_]_mydb.[1234_]_foo",
        "_[1234_]mydb._[1234_]bar",
        "[]1234_][]1234_]mydb.[]1234_][]1234_]baz", // This case tests for when `]` is included as a possible character.
    ]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__coll", "__foo", "__bar", "__baz", "__coll2");
    let expected = vec_collection_docs!("__coll", "__foo", "__bar", "__baz");

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_star_wildcards_for_dunderscores() {
    let include_list = glob_it(vec!["__*abc.*", "*.__*abc"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__abc", "__foo", "__abcabc", "coll");

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        vec_collection_docs!()
    );

    assert_eq!(
        actual!(input, "__mydbabc", &include_list, &exclude_list),
        vec_collection_docs!("coll")
    );

    assert_eq!(
        actual!(input, "__abc", &include_list, &exclude_list),
        vec_collection_docs!("coll")
    );

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        vec_collection_docs!("__abc", "__abcabc")
    );
}

#[test]
fn test_do_not_allow_question_mark_wildcard_to_act_as_dunderscore() {
    let include_list = glob_it(vec![
        "_?mydb._?coll",
        "??mydb.??foo",
        "?_mydb.?_bar",
        "__mydb.__baz",
    ]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__coll", "__foo", "__bar", "__baz");
    let expected = vec_collection_docs!("__baz");

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_do_not_allow_star_wildcard_to_act_as_dunderscore() {
    let include_list = glob_it(vec![
        "_*mydb._*coll",
        "*mydb.*foo",
        "*_mydb.*_bar",
        "__mydb.__baz",
    ]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__coll", "__foo", "__bar", "__baz");
    let expected = vec_collection_docs!("__baz");

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_do_not_allow_negation_brackets_that_allow_underscores_to_act_as_dunderscore() {
    let include_list = glob_it(vec![
        "[!1234][!32142546]mydb.[!1234][!32142546]coll",
        "[!1234]_mydb.[!1234]_foo",
        "_[!1234]mydb._[!1234]bar",
        "__mydb.__baz",
    ]);
    let exclude_list = vec![];
    let input = vec_collection_docs!("__coll", "__foo", "__bar", "__baz");
    let expected = vec_collection_docs!("__baz");

    assert_eq!(
        actual!(input, "__mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_include_and_exclude_contain_overlap() {
    let include_list = glob_it(vec!["mydb.included", "mydb.excluded"]);
    let exclude_list = glob_it(vec!["mydb.excluded"]);
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_inclusion_and_exclusion_rules_are_namespace_bound() {
    let include_list = glob_it(vec!["mydb.included"]);
    let exclude_list = glob_it(vec!["otherdb.*"]);
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection",);

    assert!(actual!(input, "otherdb", &include_list, &exclude_list).is_empty());
}

#[test]
fn test_disallowed_collection_names() {
    let include_list = glob_it(vec!["mydb.*"]);
    let exclude_list = vec![];
    let input = vec_collection_docs!(
        "system.namespaces",
        "system.indexes",
        "system.profile",
        "system.js",
        "system.views",
        "system.buckets.timeSeriesCollection",
        "system.buckets.otherCollection",
        "included",
        "excludeded",
    );
    let expected = vec_collection_docs!("included", "excludeded");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}
