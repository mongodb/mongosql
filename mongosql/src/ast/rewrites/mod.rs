use crate::ast;
use thiserror::Error;

mod tuples;
pub use tuples::InTupleRewritePass;
mod from;
pub use from::ImplicitFromRewritePass;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

/// Errors that can occur during rewrite passes
#[derive(Debug, Error)]
pub enum Error {}

/// A fallible transformation that can be applied to a query
pub trait Pass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query>;
}

/// Rewrite the provided query by applying rewrites as specified in the MongoSQL spec.
pub fn rewrite_query(query: ast::Query) -> Result<ast::Query> {
    let passes: Vec<&dyn Pass> = vec![&ImplicitFromRewritePass, &InTupleRewritePass];

    let mut rewritten = query;
    for pass in passes {
        rewritten = pass.apply(rewritten)?;
    }
    Ok(rewritten)
}
