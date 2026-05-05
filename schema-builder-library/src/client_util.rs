use crate::{DataService, MongoDbDataService};
use async_trait::async_trait;
use mongodb::{
    Cursor, Database,
    bson::Document,
    options::{
        ClientOptions, ReadPreference, ReadPreferenceOptions, ResolverConfig, SelectionCriteria,
        TagSet,
    },
};

#[cfg(not(test))]
use std::sync::OnceLock;

#[cfg(not(test))]
static SELECTION_CRITERIA: OnceLock<SelectionCriteria> = OnceLock::new();

//
// This module contains utility functions for a MongoDB client to make
// working with a CLI easier.
//

/// Returns true if the client options does not have any authentication credentials.
pub fn needs_auth(options: &ClientOptions) -> bool {
    options.credential.is_none()
}

/// Loads the username and password into the client options.
pub async fn load_password_auth(
    options: &mut ClientOptions,
    username: Option<String>,
    password: Option<String>,
) {
    options.credential = Some(
        mongodb::options::Credential::builder()
            .username(username)
            .password(password)
            .build(),
    );
}

/// Returns a client options with the optimal pool size set.
pub async fn get_opts(
    uri: &str,
    resolver: Option<ResolverConfig>,
) -> Result<ClientOptions, <MongoDbDataService as DataService>::Error> {
    let mut opts = if let Some(resolver) = resolver {
        ClientOptions::parse(uri).resolver_config(resolver).await?
    } else {
        ClientOptions::parse(uri).await?
    };
    opts.max_pool_size = Some(get_optimal_pool_size());
    opts.max_connecting = Some(2);

    if opts.selection_criteria.is_none() {
        opts.selection_criteria = Some(get_default_selection_criteria());
    }

    // We have to skip this part for the unit tests below and the internal_integration_tests since SELECTION_CRITERIA is a OnceLock
    // which can only be set once during program execution and testing. Attempting to set it more than once (which will happen if more
    // than one non-integration test that uses `get_opts` is run at the same time) will cause a panic. Due to this, we always use the default
    // `SelectionCriteria` when non-integration testing. See `get_selection_criteria()` below for more details.
    //
    // Note: Since the mongodb-schema-manager has its integration tests in the `tests/` directory,
    // the `test` configuration is false, so this part will be compiled.
    #[cfg(not(test))]
    {
        // The SELECTION_CRITERIA is only set here and should only ever be set here, so we can safely unwrap
        // after the `.set()`, which will only error if the SELECTION_CRITERIA is already set.
        // Likewise, selection_criteria is set above if it is missing, so we can safely unwrap.
        #[allow(clippy::unwrap_used)]
        SELECTION_CRITERIA
            .set(opts.clone().selection_criteria.unwrap())
            .unwrap();
    }

    Ok(opts)
}

#[async_trait]
pub trait DatabaseExt {
    async fn run_command_with_read_preference(
        &self,
        command: Document,
    ) -> Result<Document, <MongoDbDataService as DataService>::Error>;

    async fn run_cursor_command_with_read_preference(
        &self,
        command: Document,
    ) -> Result<Cursor<Document>, <MongoDbDataService as DataService>::Error>;
}

#[async_trait]
impl DatabaseExt for Database {
    async fn run_command_with_read_preference(
        &self,
        command: Document,
    ) -> Result<Document, <MongoDbDataService as DataService>::Error> {
        let selection_criteria = get_selection_criteria();

        self.run_command(command)
            .selection_criteria(selection_criteria)
            .await
    }

    async fn run_cursor_command_with_read_preference(
        &self,
        command: Document,
    ) -> Result<Cursor<Document>, <MongoDbDataService as DataService>::Error> {
        let selection_criteria = get_selection_criteria();

        self.run_cursor_command(command)
            .selection_criteria(selection_criteria)
            .await
    }
}

fn get_selection_criteria() -> SelectionCriteria {
    #[cfg(not(test))]
    {
        // Note: Since the mongodb-schema-manager has its integration tests in the `tests/` directory,
        // the `test` configuration is false, so this part will be compiled.
        #[allow(clippy::unwrap_used)]
        SELECTION_CRITERIA.get().unwrap().clone()
    }
    #[cfg(test)]
    {
        // when non-integration testing, `SELECTION_CRITERIA` doesn't exist (see `get_opts` above
        // for more details), so we always use the default `SelectionCriteria`.
        get_default_selection_criteria()
    }
}

fn get_default_selection_criteria() -> SelectionCriteria {
    let analytics_tag_set: TagSet =
        TagSet::from([("nodeType".to_string(), "ANALYTICS".to_string())]);

    // This ensures that a secondary is chosen even if there isn't an analytics node available.
    let empty_tag_set = TagSet::new();

    SelectionCriteria::ReadPreference(ReadPreference::SecondaryPreferred {
        options: Some(
            ReadPreferenceOptions::builder()
                .tag_sets(Some(vec![analytics_tag_set, empty_tag_set]))
                .build(),
        ),
    })
}

/// Returns the optimal pool size for the client.
/// The pool size is set to the number of available CPUs * 2 + 1, up to 10.
/// If the number of available CPUs cannot be determined,
/// the pool size is set to 1.
fn get_optimal_pool_size() -> u32 {
    match std::thread::available_parallelism() {
        // if try_into fails here, it means the number of available CPUs is greater than u32::MAX
        Ok(parallelism) => usize::from(parallelism).try_into().unwrap_or(10u32) * 2 + 1,
        Err(_) => 1,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_get_opts_no_read_preference_in_uri() {
        let opts = get_opts("mongodb://localhost:27017", None).await.unwrap();

        assert_eq!(
            opts.selection_criteria.unwrap(),
            get_default_selection_criteria()
        );
    }

    #[tokio::test]
    async fn test_get_opts_custom_read_preference_in_uri() {
        let opts = get_opts(
            "mongodb://localhost:27017/?readPreference=primaryPreferred",
            None,
        )
        .await
        .unwrap();

        let expected_selection_criteria =
            SelectionCriteria::ReadPreference(ReadPreference::PrimaryPreferred { options: None });

        assert_eq!(
            opts.selection_criteria.unwrap(),
            expected_selection_criteria
        );
    }
}
