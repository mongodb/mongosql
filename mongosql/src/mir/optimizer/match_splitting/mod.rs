///
/// Match Splitting
///
/// The Match Splitting pass splits Filter stages with conjunctions into separate
/// consecuitive Filter Stages in order to facilitate more movement in the Stage Movement
/// pass.
///
#[cfg(test)]
mod test;

use super::Optimizer;
use crate::{
    mir::{
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Expression::*,
        Filter,
        ScalarFunction::*,
        Stage,
        Stage::*,
    },
    SchemaCheckingMode,
};

pub(crate) struct MatchSplittingOptimizer {}

impl Optimizer for MatchSplittingOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        MatchSplittingOptimizer::split_matches(st)
    }
}

impl MatchSplittingOptimizer {
    fn split_matches(st: Stage) -> (Stage, bool) {
        let mut v = MatchSplittingVisitor::default();
        let new_stage = v.visit_stage(st);
        (new_stage, v.changed)
    }
}

#[derive(Copy, Clone, Default)]
struct MatchSplittingVisitor {
    changed: bool,
}

impl MatchSplittingVisitor {}

impl Visitor for MatchSplittingVisitor {
    fn visit_filter(&mut self, node: Filter) -> Filter {
        let node = node.walk(self);

        match node.condition.clone() {
            ScalarFunction(sfa) => match sfa.function {
                And => {
                    self.changed = true;
                    let new_filter = sfa.args.into_iter().fold(node.source, |acc, expr| {
                        Box::new(Filter(Filter {
                            source: acc,
                            condition: expr,
                            cache: SchemaCache::new(),
                        }))
                    });
                    match *new_filter {
                        Filter(filter) => filter,
                        _ => unreachable!(),
                    }
                }
                _ => node,
            },
            _ => node,
        }
    }
}
