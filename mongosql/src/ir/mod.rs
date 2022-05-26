include!(concat!(env!("OUT_DIR"), "/ir/visitor.rs"));
include!(concat!(env!("OUT_DIR"), "/ir/walk.rs"));
pub mod definitions;
pub use definitions::*;
pub mod constant_folding;
pub mod flatten;
pub mod namespace;
pub mod schema;
mod unwind_util;
pub use mongosql_datastructures::binding_tuple;
#[cfg(test)]
mod test;
