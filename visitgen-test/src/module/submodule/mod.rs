pub mod ast;
// defines pub mod visitor and pub mod walk
include!(concat!(env!("OUT_DIR"), "/module/submodule/visitor.rs"));
include!(concat!(env!("OUT_DIR"), "/module/submodule/walk.rs"));

#[cfg(test)]
mod test;
