use glob::Pattern;
use std::sync::LazyLock;

pub(crate) const DISALLOWED_DB_NAMES: [&str; 4] = ["admin", "config", "local", "system"];

pub(crate) static DISALLOWED_COLLECTION_NAMES: LazyLock<Vec<Pattern>> = LazyLock::new(|| {
    vec![
        #[allow(clippy::expect_used)]
        Pattern::new("system.*")
            .expect("Internal error: `system.*` could not be converted into a `glob::Pattern`"),
        #[allow(clippy::expect_used)]
        Pattern::new("__sql_schemas").expect(
            "Internal error: `__sql_schemas` could not be converted into a `glob::Pattern`",
        ),
    ]
});

pub(crate) const VIEW_SAMPLE_SIZE: i64 = 1000;
