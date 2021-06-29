include!(concat!(env!("OUT_DIR"), "/ir/visitor.rs"));
include!(concat!(env!("OUT_DIR"), "/ir/walk.rs"));
pub mod binding_tuple;
mod definitions;
pub use definitions::*;
pub mod constant_folding;
pub mod schema;
#[cfg(test)]
mod test;
