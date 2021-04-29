// defines pub mod visitor and pub mod walk
include!(concat!(env!("OUT_DIR"), "/ast/visitor.rs"));
include!(concat!(env!("OUT_DIR"), "/ast/walk.rs"));
pub(crate) mod definitions;
pub(crate) mod pretty_print;
pub(crate) use definitions::*;

#[cfg(test)]
mod pretty_print_test;
