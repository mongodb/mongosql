use crate::air;
use thiserror::Error;

mod accumulators;
use crate::air::desugarer::accumulators::AccumulatorsDesugarerPass;
mod join;
use crate::air::desugarer::join::JoinDesugarerPass;
mod root_references;
use crate::air::desugarer::root_references::RootReferenceDesugarerPass;
mod sql_null_semantics_operators;
use crate::air::desugarer::sql_null_semantics_operators::SqlNullSemanticsOperatorsDesugarerPass;
mod fold_converts;
use crate::air::desugarer::fold_converts::FoldConvertsDesugarerPass;
mod subquery;
use crate::air::desugarer::subquery::SubqueryExprDesugarerPass;
mod unsupported_operators;
use crate::air::desugarer::unsupported_operators::UnsupportedOperatorsDesugarerPass;
mod remove_id;
use crate::air::desugarer::remove_id::RemoveIdDesugarerPass;

#[cfg(test)]
mod test;
mod util;

pub type Result<T> = std::result::Result<T, Error>;

/// Errors that can occur during desugarer passes
#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("pattern for $like must be literal")]
    InvalidLikePattern,
    #[error("could not statically evaluate constant $convert to Type {0:?}, due to improper constant input value")]
    InvalidConstantConvert(air::Type),
}

/// A fallible transformation that can be applied to a pipeline
pub trait Pass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage>;
}

/// Desugar the provided pipeline by applying desugarer passes.
pub fn desugar_pipeline(pipeline: air::Stage) -> Result<air::Stage> {
    // The order of these passes matters. Specifically, Sql null semantic
    // operators must be desugared after any passes that create Sql null
    // semantic operators.
    let passes: Vec<&dyn Pass> = vec![
        &RootReferenceDesugarerPass,
        &JoinDesugarerPass,
        &AccumulatorsDesugarerPass,
        &SubqueryExprDesugarerPass,
        &UnsupportedOperatorsDesugarerPass,
        &SqlNullSemanticsOperatorsDesugarerPass,
        &FoldConvertsDesugarerPass,
        &RemoveIdDesugarerPass,
    ];

    let mut desugared = pipeline;
    for pass in passes {
        desugared = pass.apply(desugared)?
    }
    Ok(desugared)
}
