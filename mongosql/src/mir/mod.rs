include!(concat!(env!("OUT_DIR"), "/mir/visitor.rs"));
include!(concat!(env!("OUT_DIR"), "/mir/walk.rs"));
pub mod definitions;
pub use definitions::*;
pub mod constant_folding;
pub mod flatten;
pub mod schema;
mod unwind_util;
pub use mongosql_datastructures::binding_tuple;
#[cfg(test)]
mod test;

use thiserror::Error;
#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("{0:?} is not a valid MIR type")]
    InvalidType(crate::ast::Type),
}
