pub mod definitions;
pub use definitions::*;
pub mod schema;

pub use mongosql_datastructures::binding_tuple;
pub mod optimizer;

use thiserror::Error;
#[derive(Debug, Error, PartialEq, Clone)]
pub enum Error {
    #[error("{0:?} is not a valid MIR type")]
    InvalidType(crate::ast::Type),
}
