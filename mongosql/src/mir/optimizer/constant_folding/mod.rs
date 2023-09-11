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
    fn optimize(&self, st: Stage, _: SchemaCheckingMode, state: &SchemaInferenceState) -> Stage {
        ConstantFoldingOptimizer::fold_constants(st, state)
    }
}

impl ConstantFoldingOptimizer {
    pub(crate) fn fold_constants(st: Stage, state: &SchemaInferenceState) -> Stage {
        let mut cf = ConstantFoldExprVisitor {
            state: state.clone(),
        };
        cf.visit_stage(st)
    }
}
