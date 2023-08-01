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
    fn optimize(&self, st: Stage, _sm: SchemaCheckingMode, _: &SchemaInferenceState) -> Stage {
        MatchSplittingOptimizer::split_matches(st)
    }
}

impl MatchSplittingOptimizer {
    fn split_matches(st: Stage) -> Stage {
        let mut v = MatchSplittingVisitor;
        v.visit_stage(st)
    }
}

#[derive(Copy, Clone)]
struct MatchSplittingVisitor;

impl MatchSplittingVisitor {}

impl Visitor for MatchSplittingVisitor {
    fn visit_filter(&mut self, node: Filter) -> Filter {
        let node = node.walk(self);

        match node.condition.clone() {
            ScalarFunction(sfa) => match sfa.function {
                And => {
                    // SQL-1550: remove logic for ignoring match splitting in the presence of subqueries
                    let mut has_subquery = false;
                    let new_filter = sfa.args.into_iter().fold(node.clone().source, |acc, expr| {
                        match expr {
                            Subquery(_) | SubqueryComparison(_) => has_subquery = true,
                            _ => {}
                        }
                        Box::new(Filter(Filter {
                            source: acc,
                            condition: expr,
                            cache: SchemaCache::new(),
                        }))
                    });
                    match has_subquery {
                        true => node,
                        false => match *new_filter {
                            Filter(filter) => filter,
                            _ => unreachable!(),
                        },
                    }
                }
                _ => node,
            },
            _ => node,
        }
    }
}
