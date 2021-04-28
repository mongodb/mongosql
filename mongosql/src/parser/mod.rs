include!(concat!(env!("OUT_DIR"), "/parser/query_tests.rs"));
include!(concat!(env!("OUT_DIR"), "/parser/rewrite_tests.rs"));

mod lalrpop;
mod util;

#[cfg(test)]
mod test;

pub(crate) use lalrpop::{Error, Parser};
