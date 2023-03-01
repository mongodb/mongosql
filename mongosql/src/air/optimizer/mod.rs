use crate::air;
use thiserror::Error;

mod simplify_lookups;
use crate::air::optimizer::simplify_lookups::get_simplify_lookups;

pub type Result<T> = std::result::Result<T, Error>;

/// Errors that can occur during optimization passes
#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[allow(dead_code)]
    #[error("TODO replace error when passes are implemented")]
    TodoError,
}

/// A fallible transformation that can be applied to a pipeline
pub trait Pass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage>;
}

/// Optimize the provided pipeline by applying optimization passes.
#[allow(dead_code)]
pub fn optimize_pipeline(pipeline: air::Stage, current_db: &str) -> Result<air::Stage> {
    let passes: Vec<Box<dyn Pass>> = vec![get_simplify_lookups(current_db)];

    let mut optimized = pipeline;
    for pass in passes {
        optimized = pass.apply(optimized)?
    }
    Ok(optimized)
}
