use crate::ast;
use thiserror::Error;

mod alias;
pub use alias::AddAliasRewritePass;
mod extended_unwind_rewrite;
pub use extended_unwind_rewrite::ExtendedUnwindRewritePass;
mod select;
pub use select::SelectRewritePass;
pub mod tuples;
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
mod higher_order_functions;
pub use higher_order_functions::HigherOrderFunctionsRewritePass;

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
        required: ArgCount,
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

/// Rewrite the provided query by applying rewrites as specified in the MongoSql spec.
pub fn rewrite_query(query: ast::Query) -> Result<ast::Query> {
    let passes: Vec<&dyn Pass> = vec![
        &ExtendedUnwindRewritePass,
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
        &HigherOrderFunctionsRewritePass,
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

/// Specifies how many arguments a function accepts.
#[derive(PartialEq, Eq, Copy, Clone, Debug)]
pub enum ArgCount {
    /// Exactly this many arguments.
    Exactly(usize),
    /// Either of these many arguments.
    Either(usize, usize),
}

impl std::fmt::Display for ArgCount {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ArgCount::Exactly(n) => write!(f, "{n}"),
            ArgCount::Either(a, b) => write!(f, "{a} or {b}"),
        }
    }
}

/// Validates that `args` contains `N` elements, extracting them into as fixed-size array of length
/// `N` if so. If `args` does not contain exactly `N` elements, an `IncorrectArgumentCount` error is
/// returned.
#[inline]
pub(crate) fn try_exact_args<'a, const N: usize>(
    name: &'static str,
    args: &'a [ast::Expression],
) -> Result<&'a [ast::Expression; N]> {
    let Ok(extracted) = args.try_into() else {
        return Err(Error::IncorrectArgumentCount {
            name,
            required: ArgCount::Exactly(N),
            found: args.len(),
        });
    };
    Ok(extracted)
}

/// Validates that `args` contains either `A` or `B` elements, extracting them into as fixed-size
/// array of length `A` and a slice of the remaining elements if so. If `args` does not contain
/// exactly `A` or exactly `B` elements, an `IncorrectArgumentCount` error is returned.
#[inline]
pub(crate) fn try_extract_either_args<'a, const A: usize, const B: usize>(
    name: &'static str,
    args: &'a [ast::Expression],
) -> Result<(&'a [ast::Expression; A], &'a [ast::Expression])> {
    if args.len() != A && args.len() != B {
        return Err(Error::IncorrectArgumentCount {
            name,
            required: ArgCount::Either(A, B),
            found: args.len(),
        });
    }

    let (min, rest) = args.split_at(A);
    let Ok(min) = min.try_into() else {
        unreachable!();
    };
    Ok((min, rest))
}
