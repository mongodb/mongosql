use std::fmt::{self, Display, Formatter};

#[derive(Debug, Clone)]
pub enum SchemaBuilderResult {
    Created,
    Error(String),
    Modified,
    Unchanged,
}

impl Display for SchemaBuilderResult {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        match self {
            SchemaBuilderResult::Created => {
                write!(f, "Schema created for collection or view")
            }
            SchemaBuilderResult::Error(e) => write!(f, "Schema Error: {}", e),
            SchemaBuilderResult::Modified => {
                write!(f, "Schema modified for collection or view")
            }
            SchemaBuilderResult::Unchanged => {
                write!(f, "Schema unchanged for collection or view")
            }
        }
    }
}
