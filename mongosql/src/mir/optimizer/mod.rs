use crate::{
    mir::{schema::SchemaInferenceState, Stage},
    SchemaCheckingMode,
};

mod constant_folding;
mod flatten_variadics;
mod match_null_filtering;
mod match_splitting;
mod stage_movement;
mod use_def_analysis;

pub(crate) trait Optimizer {
    fn optimize(
        &self,
        st: Stage,
        sm: SchemaCheckingMode,
        schema_state: &SchemaInferenceState,
    ) -> Stage;
}

// Avoiding lifetime hacking by using a fn
// Optimizers must be added to this vec in the order they should be applied
static OPTIMIZERS: fn() -> Vec<Box<dyn Optimizer>> = || {
    vec![
        Box::new(flatten_variadics::FlattenVariadicFunctionsOptimizer {}),
        Box::new(constant_folding::ConstantFoldingOptimizer {}),
        Box::new(match_splitting::MatchSplittingOptimizer {}),
        Box::new(stage_movement::StageMovementOptimizer {}),
        Box::new(match_null_filtering::MatchNullFilteringOptimizer {}),
    ]
};

/// Optimizes the provided MIR stage. Internally, Optimizers determine whether
/// the SchemaCheckingMode is used or not.
pub fn optimize_plan(
    st: Stage,
    schema_checking_mode: SchemaCheckingMode,
    schema_state: &SchemaInferenceState,
) -> crate::mir::Stage {
    OPTIMIZERS().into_iter().fold(st, |acc, opt| {
        opt.optimize(acc, schema_checking_mode, schema_state)
    })
}
