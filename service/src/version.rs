use std::sync::LazyLock;

pub static VERSION: LazyLock<String> = LazyLock::new(|| {
    format!(
        "{}.{}.{}",
        env!("CARGO_PKG_VERSION_MAJOR"),
        env!("CARGO_PKG_VERSION_MINOR"),
        env!("CARGO_PKG_VERSION_PATCH")
    )
});
