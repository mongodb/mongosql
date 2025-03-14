use crate::ast;
use thiserror::Error;

mod alias;
pub use alias::AddAliasRewritePass;
mod extended_unwind_rewrite;
pub use extended_unwind_rewrite::ExtendedUnwindRewritePass;
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
mod group_by_select_alias;
use group_by_select_alias::GroupBySelectAliasRewritePass;
mod not;
use not::NotComparisonRewritePass;
mod optional_parameters;
use optional_parameters::OptionalParameterRewritePass;
mod scalar_functions;
use scalar_functions::ScalarFunctionsRewritePass;
mod with_query;
pub use with_query::WithQueryRewritePass;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

/// Errors that can occur during rewrite passes
#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("positional sort keys are not allowed with SELECT VALUE")]
    PositionalSortKeyWithSelectValue,
    #[error("positional sort keys are not allowed with SELECT *")]
    PositionalSortKeyWithSelectStar,
    #[error("positional sort key {0} out of range")]
    PositionalSortKeyOutOfRange(usize),
    #[error("positional sort key {0} references a select expression with no alias")]
    NoAliasForSortKeyAtPosition(usize),
    #[error("aggregation functions may not be used as GROUP BY keys")]
    AggregationFunctionInGroupByKeyList,
    #[error("cannot specify aggregation functions in GROUP BY AGGREGATE clause and elsewhere")]
    AggregationFunctionInGroupByAggListAndElsewhere,
    #[error("all SELECT expressions must be given aliases before the SelectRewritePass")]
    NoAliasForSelectExpression,
    #[error("the top-level SELECT in a subquery expression must be a standard SELECT")]
    SubqueryWithSelectValue,
    #[error("incorrect argument count for {name}: required {required}, found {found}")]
    IncorrectArgumentCount {
        name: &'static str,
        required: &'static str,
        found: usize,
    },
    #[error("invalid date part: {0}")]
    InvalidDatePart(&'static str),
    #[error("UNWIND datasource must have a PATH")]
    UnwindSourceWithoutPath,
    #[error("duplicate option in UNWIND: {0}")]
    DuplicateOptionInUnwind(&'static str),
}

/// A fallible transformation that can be applied to a query
pub trait Pass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query>;
}

/// Rewrite the provided query by applying rewrites as specified in the MongoSQL spec.
pub fn rewrite_query(query: ast::Query) -> Result<ast::Query> {
    let passes: Vec<&dyn Pass> = vec![
        &ExtendedUnwindRewritePass,
        &InTupleRewritePass,
        &SingleTupleRewritePass,
        &GroupBySelectAliasRewritePass,
        &AddAliasRewritePass,
        &PositionalSortKeyRewritePass,
        &AggregateRewritePass,
        &SelectRewritePass,
        &ImplicitFromRewritePass,
        &TableSubqueryRewritePass,
        &OptionalParameterRewritePass,
        &NotComparisonRewritePass,
        &ScalarFunctionsRewritePass,
        // WithQueryRewritePass can introduce duplicated queries, so it should be the last pass so
        // any rewrites that apply in the WithQuery queries are applied only once.
        &WithQueryRewritePass,
    ];

    let mut rewritten = query;
    for pass in passes {
        rewritten = pass.apply(rewritten)?;
    }
    Ok(rewritten)
}
