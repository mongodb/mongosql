mod negative_normalize;
#[cfg(test)]
mod negative_normalize_tests;
pub mod schema_derivation;
#[allow(unused_imports)]
pub use schema_derivation::*;
#[cfg(test)]
mod schema_derivation_tests;
#[cfg(test)]
mod test;

use mongosql::schema::Schema;
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Clone)]
pub enum Error {
    #[error("Cannot derive schema for undefined literals")]
    InvalidLiteralType,
    #[error("Cannot derive schema for unsupported operator: {0}")]
    InvalidUntaggedOperator(String),
    #[error("Cannot derive schema for unsupported operator: {0:?}")]
    InvalidTaggedOperator(agg_ast::definitions::TaggedOperator),
    #[error("Unknown reference in current context: {0}")]
    UnknownReference(String),
}

#[allow(dead_code)]
pub(crate) fn get_schema_for_path_mut(
    schema: &mut Schema,
    path: Vec<String>,
) -> Option<&mut Schema> {
    let mut schema = Some(schema);
    for field in path {
        schema = match schema {
            Some(Schema::Document(d)) => d.keys.get_mut(&field),
            _ => {
                return None;
            }
        };
    }
    schema
}
