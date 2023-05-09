use crate::air;
use thiserror::Error;

mod accumulators;
use crate::air::desugarer::accumulators::AccumulatorsDesugarerPass;
mod join;
use crate::air::desugarer::join::JoinDesugarerPass;
mod match_null_semantics;
use crate::air::desugarer::match_null_semantics::MatchDesugarerPass;
mod root_references;
use crate::air::desugarer::root_references::RootReferenceDesugarerPass;
mod sql_null_semantics_operators;
use crate::air::desugarer::sql_null_semantics_operators::SQLNullSemanticsOperatorsDesugarerPass;
mod subquery;
use crate::air::desugarer::subquery::SubqueryExprDesugarerPass;
mod unsupported_operators;
use crate::air::desugarer::unsupported_operators::UnsupportedOperatorsDesugarerPass;

#[cfg(test)]
mod test;
mod util;

pub type Result<T> = std::result::Result<T, Error>;

/// Errors that can occur during desugarer passes
#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("pattern for $like must be literal")]
    InvalidLikePattern,
}

/// A fallible transformation that can be applied to a pipeline
pub trait Pass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage>;
}

/// Desugar the provided pipeline by applying desugarer passes.
pub fn desugar_pipeline(pipeline: air::Stage) -> Result<air::Stage> {
    // The order of these passes matters. Specifically, SQL null semantic
    // operators must be desugared after any passes that create SQL null
    // semantic operators.
    let passes: Vec<&dyn Pass> = vec![
        &RootReferenceDesugarerPass,
        &JoinDesugarerPass,
        &AccumulatorsDesugarerPass,
        &MatchDesugarerPass,
        &SubqueryExprDesugarerPass,
        &UnsupportedOperatorsDesugarerPass,
        &SQLNullSemanticsOperatorsDesugarerPass,
    ];

    let mut desugared = pipeline;
    for pass in passes {
        desugared = pass.apply(desugared)?
    }
    Ok(desugared)
}
