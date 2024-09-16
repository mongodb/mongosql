use crate::mir::schema::SchemaCheckingMode;

/// Options passed in for translation, used throughout the various translation components
#[derive(Debug, Copy, Clone, Default)]
pub struct SqlOptions {
    pub exclude_namespaces: ExcludeNamespacesOption,
    pub schema_checking_mode: SchemaCheckingMode,
    pub allow_order_by_missing_columns: bool,
}

impl SqlOptions {
    #[allow(dead_code)]
    pub fn new(
        exclude_namespaces: ExcludeNamespacesOption,
        schema_checking_mode: SchemaCheckingMode,
    ) -> Self {
        SqlOptions {
            exclude_namespaces,
            schema_checking_mode,
            allow_order_by_missing_columns: false,
        }
    }
}

/// Specifies whether or not to exclude namespaces in the result set
#[allow(dead_code)]
#[derive(Debug, Copy, Clone, Default, PartialEq)]
pub enum ExcludeNamespacesOption {
    ExcludeNamespaces,
    #[default]
    IncludeNamespaces,
}
