mod literal;
pub use literal::{are_literal, is_literal};

mod collections;
pub use collections::get_collection_sources;

#[cfg(test)]
mod test;
