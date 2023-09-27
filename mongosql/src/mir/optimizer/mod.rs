use crate::{
    mir::{schema::SchemaInferenceState, Stage},
    SchemaCheckingMode,
};

mod constant_folding;
mod dead_code_elimination;
mod determine_join_semantics;
mod flatten_variadics;
mod lower_joins;
mod match_null_filtering;
mod match_splitting;
mod prefilter_unwinds;
mod rewrite_to_match_language;
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
        Box::new(rewrite_to_match_language::MatchLanguageRewriter {}),
        Box::new(stage_movement::StageMovementOptimizer {}),
        Box::new(determine_join_semantics::JoinSemanticsOptimizer {}),
        Box::new(lower_joins::LowerJoinsOptimizer {}),
        Box::new(match_null_filtering::MatchNullFilteringOptimizer {}),
        Box::new(prefilter_unwinds::PrefilterUnwindsOptimizer {}),
        Box::new(stage_movement::StageMovementOptimizer {}),
        Box::new(dead_code_elimination::DeadCodeEliminator {}),
    ]
};

/// Optimizes the provided MIR stage. Internally, Optimizers determine whether
/// the SchemaCheckingMode is used or not.
pub fn optimize_plan(
    st: Stage,
    schema_checking_mode: SchemaCheckingMode,
    schema_state: &SchemaInferenceState,
) -> Stage {
    OPTIMIZERS().into_iter().fold(st, |acc, opt| {
        opt.optimize(acc, schema_checking_mode, schema_state)
    })
}
