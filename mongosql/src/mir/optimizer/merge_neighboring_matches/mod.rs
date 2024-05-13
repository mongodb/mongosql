///
/// Merge Neighboring Matches
///
/// Over the course of the optimization process, we split and move Filter stages freely.
/// Evaluating these in serial versus in a $and statement is equivalent in the normal pipeline.
/// However, in $lookup stages, evaluating in serial can lead to performance decreases. This
/// optimization merges all neighboring Filters into $and statements, which should improve the performance
/// of $lookups, and achieve the same performance outside of them.
#[cfg(test)]
mod test;

use super::Optimizer;
use crate::{
    mir::{
        schema::SchemaInferenceState, visitor::Visitor, Expression, Filter, ScalarFunction,
        ScalarFunctionApplication, Stage,
    },
    SchemaCheckingMode,
};

pub(crate) struct MergeNeighboringMatchesOptimizer {}

impl Optimizer for MergeNeighboringMatchesOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        MergeNeighboringMatchesOptimizer::merge_neighboring_matches(st)
    }
}

impl MergeNeighboringMatchesOptimizer {
    fn merge_neighboring_matches(st: Stage) -> (Stage, bool) {
        let mut visitor = MergeNeighboringMatchesVisitor;
        let new_stage = visitor.visit_stage(st);
        (new_stage, false)
    }
}

#[derive(Default)]
struct MergeNeighboringMatchesVisitor;

impl Visitor for MergeNeighboringMatchesVisitor {
    fn visit_filter(&mut self, node: Filter) -> Filter {
        let node = node.walk(self);
        match *node.source {
            Stage::Filter(f) => {
                let (conditions, is_nullable) = match f.condition {
                    // the parent filter is already a $and, append the condition to that and
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::And,
                        args,
                        is_nullable,
                        ..
                    }) => {
                        let condition_is_nullable = is_nullable || node.condition.is_nullable();
                        let mut conditions = args;
                        conditions.push(node.condition);
                        (conditions, condition_is_nullable)
                    }
                    // otherwise, create a $and with the two filter conditions
                    _ => {
                        let is_nullable = f.condition.is_nullable() || node.condition.is_nullable();
                        (vec![f.condition, node.condition], is_nullable)
                    }
                };
                Filter {
                    source: f.source,
                    condition: Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::And,
                        args: conditions,
                        is_nullable,
                    }),
                    cache: f.cache,
                }
            }
            _ => node,
        }
    }
}
