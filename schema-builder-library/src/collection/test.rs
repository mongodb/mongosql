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
    ($input:expr, $db:expr, $include_list:expr, $exclude_list:expr $(,)?) => {
        $input
            .iter()
            .filter(|c| CollectionInfo::should_consider($db, c, $include_list, $exclude_list))
            .cloned()
            .collect::<Vec<CollectionDoc>>()
    };
}

#[test]
fn test_inclusion_glob() {
    let include_list = vec!["mydb.*".to_string()];
    let exclude_list = vec!["mydb.excluded".to_string()];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included", "excludeded", "subcollection");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}

#[test]
fn test_exclusion_glob() {
    let include_list = vec!["mydb.excluded".to_string()];
    let exclude_list = vec!["mydb.*".to_string(), "otherdb.*".to_string()];
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
    let exclude_list = vec!["mydb.excluded".to_string()];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included", "excludeded", "subcollection");

    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}

#[test]
fn test_include_contents_exclude_empty() {
    let include_list = vec!["mydb.included".to_string(), "mydb.excluded".to_string()];
    let exclude_list = vec![];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included", "excluded");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_include_and_exclude_contain_overlap() {
    let include_list = vec!["mydb.included".to_string(), "mydb.excluded".to_string()];
    let exclude_list = vec!["mydb.excluded".to_string()];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection");
    let expected = vec_collection_docs!("included");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected
    );
}

#[test]
fn test_inclusion_and_exclusion_rules_are_namespace_bound() {
    let include_list = vec!["mydb.included".to_string()];
    let exclude_list = vec!["otherdb.*".to_string()];
    let input = vec_collection_docs!("included", "excluded", "excludeded", "subcollection",);

    assert!(actual!(input, "otherdb", &include_list, &exclude_list).is_empty());
}

#[test]
fn test_disallowed_collection_names() {
    let include_list = vec!["mydb.*".to_string()];
    let exclude_list: Vec<String> = vec![];
    let input = vec_collection_docs!(
        "system.namespaces",
        "system.indexes",
        "system.profile",
        "system.js",
        "system.views",
        "included",
        "excludeded",
    );
    let expected = vec_collection_docs!("included", "excludeded");
    assert_eq!(
        actual!(input, "mydb", &include_list, &exclude_list),
        expected,
    );
}
