use crate::ast;
use thiserror::Error;

mod alias;
pub use alias::AddAliasRewritePass;
mod select;
pub use select::SelectRewritePass;
pub mod tuples;
pub use tuples::InTupleRewritePass;
pub use tuples::SingleTupleRewritePass;
mod from;
pub use from::ImplicitFromRewritePass;
mod order_by;
pub use order_by::PositionalSortKeyRewritePass;
mod aggregate;
pub use aggregate::AggregateRewritePass;
mod table_subquery;
use table_subquery::TableSubqueryRewritePass;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

/// Errors that can occur during rewrite passes
#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("positional sort keys are not allowed with SELECT VALUE")]
    PositionalSortKeyWithSelectValue,
    #[error("positional sort key {0} out of range")]
    PositionalSortKeyOutOfRange(usize),
    #[error("positional sort key {0} references a select expression with no alias")]
    NoAliasForSortKeyAtPosition(usize),
    #[error("aggregation functions may not be used as GROUP BY keys")]
    AggregationFunctionInGroupByKeyList,
    #[error("cannot specify aggregation functions in GROUP BY AGGREGATE clause and elsewhere")]
    AggregationFunctionInGroupByAggListAndElsewhere,
    #[error("all select expressions must be given aliases before the SelectRewritePass")]
    NoAliasForSelectExpression,
    #[error("the top-level SELECT in a subquery expression must be a standard SELECT")]
    SubqueryWithSelectValue,
}

/// A fallible transformation that can be applied to a query
pub trait Pass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query>;
}

/// Rewrite the provided query by applying rewrites as specified in the MongoSQL spec.
pub fn rewrite_query(query: ast::Query) -> Result<ast::Query> {
    let passes: Vec<&dyn Pass> = vec![
        &AddAliasRewritePass,
        &PositionalSortKeyRewritePass,
        &SelectRewritePass,
        &ImplicitFromRewritePass,
        &AggregateRewritePass,
        &InTupleRewritePass,
        &SingleTupleRewritePass,
        &TableSubqueryRewritePass,
    ];

    let mut rewritten = query;
    for pass in passes {
        rewritten = pass.apply(rewritten)?;
    }
    Ok(rewritten)
}
