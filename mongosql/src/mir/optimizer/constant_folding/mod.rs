///
/// Constant Folding
///
/// This optimization replaces constant expressions with the values they will evaluate to at
/// query time. The goal of this is to reduce the amount of work done during query execution.
///
#[cfg(test)]
mod test;

mod lib;
use super::Optimizer;
use crate::{
    mir::{schema::SchemaInferenceState, visitor::Visitor, Stage},
    SchemaCheckingMode,
};
pub(crate) use lib::ConstantFoldExprVisitor;

pub(crate) struct ConstantFoldingOptimizer {}

impl Optimizer for ConstantFoldingOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        ConstantFoldingOptimizer::fold_constants(st, schema_state)
    }
}

impl ConstantFoldingOptimizer {
    pub(crate) fn fold_constants(st: Stage, state: &SchemaInferenceState) -> (Stage, bool) {
        let mut cf = ConstantFoldExprVisitor {
            state,
            changed: false,
        };
        let new_stage = cf.visit_stage(st);
        (new_stage, cf.changed)
    }
}
