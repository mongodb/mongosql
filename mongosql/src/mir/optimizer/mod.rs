use crate::{
    mir::{schema::SchemaInferenceState, Stage},
    SchemaCheckingMode,
};
use tailcall::tailcall;

mod constant_folding;
mod dead_code_elimination;
mod determine_join_semantics;
mod flatten_variadics;
mod lower_joins;
mod match_null_filtering;
mod match_splitting;
mod merge_neighboring_matches;
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
    ) -> (Stage, bool);
}

// Avoiding lifetime hacking by using a fn
// Optimizers must be added to this vec in the order they should be applied
static OPTIMIZERS: fn() -> Vec<Box<dyn Optimizer>> = || {
    vec![
        Box::new(flatten_variadics::FlattenVariadicFunctionsOptimizer {}),
        Box::new(constant_folding::ConstantFoldingOptimizer {}),
        Box::new(match_splitting::MatchSplittingOptimizer {}),
        Box::new(rewrite_to_match_language::MatchLanguageRewriter {}),
        Box::new(match_null_filtering::MatchNullFilteringOptimizer {}),
        Box::new(stage_movement::StageMovementOptimizer {}),
        Box::new(determine_join_semantics::JoinSemanticsOptimizer {}),
        Box::new(lower_joins::LowerJoinsOptimizer {}),
        Box::new(prefilter_unwinds::PrefilterUnwindsOptimizer {}),
        Box::new(dead_code_elimination::DeadCodeEliminator {}),
    ]
};

/// Optimizes the provided MIR stage. Internally, Optimizers determine whether
/// the SchemaCheckingMode is used or not.
#[tailcall]
pub fn optimize_plan(
    st: Stage,
    schema_checking_mode: SchemaCheckingMode,
    schema_state: &SchemaInferenceState,
) -> Stage {
    let (new_st, changed) =
        OPTIMIZERS()
            .into_iter()
            .fold((st, false), |(acc_st, acc_changed), opt| {
                let (new_st, changed) = opt.optimize(acc_st, schema_checking_mode, schema_state);
                (new_st, acc_changed || changed)
            });

    // If the pipeline was updated as a result of optimization, optimize again.
    let plan = if changed {
        optimize_plan(new_st, schema_checking_mode, schema_state)
    } else {
        new_st
    };
    // We perform merge_neighboring_matches last because we do not want to undo this optimization during the fixed
    // point and re-merge the matches every loop of the fixed point. Most of the fixed point optimizations benefit from
    // having matches split, and it is only at the end where we can then merge them back together efficiently.
    let (new_plan, _) = merge_neighboring_matches::MergeNeighboringMatchesOptimizer {}.optimize(
        plan,
        schema_checking_mode,
        schema_state,
    );
    new_plan
}
