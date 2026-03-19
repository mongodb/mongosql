use super::{
    CollectionDoc, CollectionInfo, EXCLUDE_DUNDERSCORE_PATTERN, INCLUDE_LIST_IN_DB_AND_COLL_PAIRS,
};
use crate::{Error, Result, consts::DISALLOWED_COLLECTION_NAMES};
use futures::TryStreamExt;
use mongodb::{
    Cursor,
    bson::{self, Document},
};
use tracing::instrument;

impl CollectionInfo {
    /// process_inclusion filters an input CollectionDoc by the include_list and
    /// exclude_list.
    /// First, it filters the input collection_list by the include_list, retaining
    /// items that are in the include_list.
    /// Second, it filters the collection_list by the exclude_list, removing items
    /// that are in the exclude_list.
    /// Lastly, it filters out any collections that are in the disallowed list.
    ///
    /// Glob syntax is supported, i.e. mydb.* will match all collections in mydb.
    #[instrument(level = "trace")]
    pub(crate) fn should_consider(
        database: &str,
        collection_or_view: &CollectionDoc,
        include_list: &[glob::Pattern],
        exclude_list: &[glob::Pattern],
    ) -> Result<bool> {
        let allow_dunderscore_namespace =
            Self::should_allow_dunderscore_namespace(database, collection_or_view, include_list)?;

        Ok((include_list.is_empty()
            || include_list.iter().any(|pattern| {
                pattern.matches(&format!("{database}.{}", collection_or_view.name.as_str()))
            }))
            && (!exclude_list.iter().any(|pattern| {
                pattern.matches(&format!("{database}.{}", collection_or_view.name.as_str()))
            }))
            && (!(EXCLUDE_DUNDERSCORE_PATTERN.matches(database)
                || EXCLUDE_DUNDERSCORE_PATTERN.matches(collection_or_view.name.as_str()))
                || allow_dunderscore_namespace)
            && (!DISALLOWED_COLLECTION_NAMES
                .iter()
                .any(|pattern| pattern.matches(collection_or_view.name.as_str()))))
    }

    /// Since we automatically exclude DBs and collections that start with dunderscores,
    /// this function is used to determine if a dunderscore namespace should be included.
    /// If a non-dunderscore namespace is passed to this function, nothing is done, and
    /// `false` is returned.
    #[instrument(level = "trace")]
    pub(crate) fn should_allow_dunderscore_namespace(
        database: &str,
        collection_or_view: &CollectionDoc,
        include_list: &[glob::Pattern],
    ) -> Result<bool> {
        let db_starts_with_dunderscore = database.starts_with("__");
        let coll_starts_with_dunderscore = collection_or_view.name.as_str().starts_with("__");

        // If the DB or collection starts with dunderscores, check if we should include it.
        if db_starts_with_dunderscore || coll_starts_with_dunderscore {
            // Since our test cases have different `include_lists`, we can't use the static OnceLock for testing.
            let include_list_in_db_and_coll_pairs = if cfg!(test) {
                &include_list
                        .iter()
                        .map(|pattern| {
                            let pattern_as_str = pattern.as_str();
                            let (db, collection) = pattern_as_str.split_once(".")
                                .unwrap_or_else(|| unreachable!("Internal Error: The pattern `{pattern_as_str}` is not in the format `<database_pattern>.<collection_pattern>`. However, this should have been caught earlier."));
                            (db.to_string(), collection.to_string())
                        })
                        .collect::<Vec<(String, String)>>()
            } else {
                INCLUDE_LIST_IN_DB_AND_COLL_PAIRS.get_or_init(|| {
                        include_list
                            .iter()
                            .map(|pattern| {
                                let pattern_as_str = pattern.as_str();
                                let (db, collection) = pattern_as_str.split_once(".")
                                    .unwrap_or_else(|| unreachable!("Internal Error: The pattern `{pattern_as_str}` is not in the format `<database_pattern>.<collection_pattern>`. However, this should have been caught earlier."));
                                (db.to_string(), collection.to_string())
                            })
                            .collect::<Vec<(String, String)>>()
                    })
            };

            let mut allow_dunderscore_namespace = false;

            for (db_pat, coll_pat) in include_list_in_db_and_coll_pairs {
                let mut allow_db = false;
                let mut allow_coll = false;

                if db_starts_with_dunderscore {
                    allow_db = Self::pattern_allows_dunderscore_name(db_pat, database)?;
                }
                if coll_starts_with_dunderscore {
                    allow_coll = Self::pattern_allows_dunderscore_name(
                        coll_pat,
                        collection_or_view.name.as_str(),
                    )?;
                }

                // The below boolean logic works by checking our three possible cases: (1) only the collection starts with dunderscores,
                // (2) only the DB starts with dunderscores, or (3) both start with dunderscores. Additionally, we only check the conditions
                // from each case that are necessary (e.g., if only the DB starts with dunderscores, only the DB must be checked that it matches a
                // dunderscore inclusion pattern).
                allow_dunderscore_namespace = allow_dunderscore_namespace
                    || ((allow_db && db_starts_with_dunderscore && !coll_starts_with_dunderscore)
                        || (allow_coll
                            && coll_starts_with_dunderscore
                            && !db_starts_with_dunderscore)
                        || (allow_coll
                            && coll_starts_with_dunderscore
                            && allow_db
                            && db_starts_with_dunderscore));
            }

            Ok(allow_dunderscore_namespace)
        } else {
            // If neither the DB or collection are prefixed with a dunderscore, there's no reason to check if we need
            // to include this namespace because it won't be automatically excluded, so we return `false`.
            Ok(false)
        }
    }

    /// This is a helper function for `should_allow_dunderscore_namespace`. This function checks
    /// if the passed `pattern_as_str` allows for the inclusion of the passed dunderscore-prefixed `name`.
    ///
    /// It works by first checking if the `pattern_as_str` fits this format:
    /// `<[<range_including_underscore>] or '_'><[<range_including_underscore>] or '_'><name or wildcard>`
    /// and then checks if it matches the provided `name`. In other words, first this functions checks
    /// if the provided pattern allows for _any_ dunderscore-prefixed name, and then it checks if
    /// it allows for the provided `name`.
    #[instrument(level = "trace")]
    pub(crate) fn pattern_allows_dunderscore_name(
        pattern_as_str: &str,
        name: &str,
    ) -> Result<bool> {
        // Proof: The only way in glob syntax to explicitly include characters is by using the character itself
        // or by using brackets with the character included in the brackets. Therefore, to disallow implicit inclusion
        // of dunderscore-prefixed names and allow for explicit inclusion of them, we want to ignore every
        // possible case except four:
        //
        //  (1) pattern starts with `__`, (2) pattern starts with `_[...]`,
        //  (3) pattern starts with `[...]_`, or (4) pattern starts with `[...][...]`.
        //
        // Furthermore, once we have identified one of these cases, we only know that it's possible
        // that our pattern matches our dunderscore `name`, so we then have to check if our pattern actually
        // matches our dunderscore `name`. Therefore, if our pattern fits one of the above cases
        // and matches our `name`, the pattern allows for `name`.

        // Checks cases (1) and (2)
        if pattern_as_str.starts_with("__")
            || (pattern_as_str.starts_with("_[") && !pattern_as_str.starts_with("_[!"))
        {
            let pattern = glob::Pattern::new(pattern_as_str).map_err(Error::GlobPatternError)?;
            return Ok(pattern.matches(name));
        }
        // Checks cases (3) and (4)
        else if pattern_as_str.starts_with("[") && !pattern_as_str.starts_with("[!") {
            // According to glob syntax, to include `]` as a possible character, you must put it right after the initial `[`,
            // so if we encounter `[]`, this means the user has chosen to include the `]` in the brackets as a possible character.
            // If this is the case, we need to find the next `]` because that is the one the concludes the list of characters to include.
            let close_bracket_index =
                if let Some(stripped_slice) = pattern_as_str.strip_prefix("[]") {
                    // Omit first `[]` from slice and then use the `find()` function.
                    stripped_slice.find("]").ok_or(
                        Error::InclusionBracketPatternIsMissingClosingBracket(
                            pattern_as_str.to_string(),
                        ),
                    )? + 2 // Add two to the result because we subtracted two from the length by stripping the first two characters.
                } else {
                    pattern_as_str.find("]").ok_or(
                        Error::InclusionBracketPatternIsMissingClosingBracket(
                            pattern_as_str.to_string(),
                        ),
                    )?
                };

            let slice_excluding_first_inclusion_brackets =
                &pattern_as_str[(close_bracket_index + 1)..];

            // Make sure we are only including cases `[...]_` and `[...][...]`.
            if !(slice_excluding_first_inclusion_brackets.starts_with("*")
                || slice_excluding_first_inclusion_brackets.starts_with("?")
                || slice_excluding_first_inclusion_brackets.starts_with("[!"))
            {
                let pattern =
                    glob::Pattern::new(pattern_as_str).map_err(Error::GlobPatternError)?;
                return Ok(pattern.matches(name));
            }
        }
        // If none of the above logic returns `true`, then `pattern_as_str` does not allow for the inclusion of `name`,
        // so we return `false`.
        Ok(false)
    }

    #[instrument(level = "trace")]
    pub async fn separate_collection_types(
        database: &str,
        include_list: &[glob::Pattern],
        exclude_list: &[glob::Pattern],
        mut collection_doc: Cursor<Document>,
    ) -> Result<CollectionInfo> {
        let mut collection_info = CollectionInfo::default();
        while let Some(collection_doc) = collection_doc.try_next().await? {
            let Ok(collection_doc) = bson::from_bson(bson::Bson::Document(collection_doc)) else {
                continue;
            };
            if CollectionInfo::should_consider(
                database,
                &collection_doc,
                include_list,
                exclude_list,
            )? {
                if collection_doc.type_ == "view" {
                    collection_info.views.push(collection_doc);
                } else if collection_doc.type_ == "timeseries" {
                    collection_info.timeseries.push(collection_doc);
                } else {
                    collection_info.collections.push(collection_doc);
                }
            }
        }

        Ok(collection_info)
    }
}
