#[cfg(test)]
mod test;

mod lib;
use super::Optimizer;
use crate::{
    mir::{visitor::Visitor, Stage},
    SchemaCheckingMode,
};
use lib::ConstantFoldExprVisitor;

pub(crate) struct ConstantFoldingOptimizer {}

impl Optimizer for ConstantFoldingOptimizer {
    fn optimize(&self, st: Stage, sm: SchemaCheckingMode) -> Stage {
        ConstantFoldingOptimizer::fold_constants(st, sm)
    }
}

impl ConstantFoldingOptimizer {
    pub(crate) fn fold_constants(st: Stage, schema_checking_mode: SchemaCheckingMode) -> Stage {
        let mut cf = ConstantFoldExprVisitor {
            schema_checking_mode,
        };
        cf.visit_stage(st)
    }
}
