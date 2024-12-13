use crate::Result;
use mongodb::options::{ClientOptions, ResolverConfig};

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
pub async fn get_opts(uri: &str, resolver: Option<ResolverConfig>) -> Result<ClientOptions> {
    let mut opts = if let Some(resolver) = resolver {
        ClientOptions::parse(uri).resolver_config(resolver).await?
    } else {
        ClientOptions::parse(uri).await?
    };
    opts.max_pool_size = Some(get_optimal_pool_size());
    opts.max_connecting = Some(2);
    Ok(opts)
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
