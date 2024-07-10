use lazy_static::lazy_static;
pub(crate) const TOOL_SHORT_NAME: &str = "sql_schema_tool";
pub(crate) const SCHEMA_COLLECTION_NAME: &str = "__sql_schemas";

lazy_static! {
    pub static ref DEFAULT_APP_NAME: String =
        format!("{}-{}", TOOL_SHORT_NAME, env!("CARGO_PKG_VERSION"));
}
