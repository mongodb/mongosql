pub mod definitions;
pub use definitions::*;
pub mod schema;
mod schema_util;
pub use mongosql_datastructures::binding_tuple;
pub mod optimizer;

#[cfg(test)]
mod test;

use thiserror::Error;
#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("{0:?} is not a valid MIR type")]
    InvalidType(crate::ast::Type),
}
