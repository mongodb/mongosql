///
/// Optimizes Join stages by converting them into LateralJoin stages when
/// possible. This involves moving the Join condition _into_ the right source.
/// This enables stage_movement to move the condition earlier in the right
/// source (or "subquery" in a LateralJoin), which results in better performance
/// when the LateralJoin is translated into a $lookup stage.

#[cfg(test)]
mod test;

use crate::{
    mir::{
        optimizer::Optimizer,
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Filter, Join, LateralJoin, MQLStage, Stage,
    },
    SchemaCheckingMode,
};

pub(crate) struct LowerJoinsOptimizer;

impl Optimizer for LowerJoinsOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> Stage {
        let mut v = LowerJoinsVisitor;
        v.visit_stage(st)
    }
}

struct LowerJoinsVisitor;

impl Visitor for LowerJoinsVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);

        match node {
            Stage::Join(Join {
                join_type,
                left,
                right,
                condition: Some(condition),
                cache,
            }) => Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
                join_type,
                source: left,
                subquery: Box::new(Stage::Filter(Filter {
                    source: right,
                    condition,
                    cache: SchemaCache::new(),
                })),
                cache,
            })),
            _ => node,
        }
    }
}
