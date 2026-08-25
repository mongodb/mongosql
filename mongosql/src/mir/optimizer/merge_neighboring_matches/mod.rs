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
use crate::mir::optimizer::util::ContainsSubqueryVisitor;
use crate::mir::{MatchFilter, MatchLanguageLogical, MatchLanguageLogicalOp, MatchQuery, MqlStage};
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

impl MergeNeighboringMatchesVisitor {
    fn is_at_pipeline_start(stage: &Stage) -> bool {
        matches!(stage, Stage::Collection(_) | Stage::Array(_))
    }

    fn contains_subquery(expr: &Expression) -> bool {
        let mut visitor = ContainsSubqueryVisitor::default();
        visitor.visit_expression(expr.clone());
        visitor.contains_subquery
    }
}

impl Visitor for MergeNeighboringMatchesVisitor {
    fn visit_match_filter(&mut self, node: MatchFilter) -> MatchFilter {
        // Recurse first so that we've already merged children before we try to merge this node.
        let MatchFilter {
            source: this_source,
            condition: this_condition,
            cache: this_cache,
        } = node.walk(self);
        match *this_source {
            Stage::MqlIntrinsic(MqlStage::MatchFilter(child)) => {
                let MatchFilter {
                    source: other_source,
                    condition: other_condition,
                    ..
                } = *child;

                // Combine the child condition with the parent condition
                let conditions: Vec<MatchQuery> = match other_condition {
                    // the child is already a $and, append this condition to that and
                    // instead of nesting a new $and inside it
                    MatchQuery::Logical(MatchLanguageLogical {
                        op: MatchLanguageLogicalOp::And,
                        mut args,
                        ..
                    }) => {
                        args.push(this_condition);
                        args
                    }
                    // otherwise, create a vector with the child condition, and the condition of this filter
                    combined_conditions => vec![combined_conditions, this_condition],
                };
                MatchFilter {
                    source: other_source,
                    condition: MatchQuery::Logical(MatchLanguageLogical {
                        op: MatchLanguageLogicalOp::And,
                        args: conditions,
                        cache: Default::default(),
                    }),
                    cache: this_cache,
                }
            }
            // nothing to merge; rebuild the filter from its own (unchanged) fields
            other => MatchFilter {
                source: Box::new(other),
                condition: this_condition,
                cache: this_cache,
            },
        }
    }

    fn visit_filter(&mut self, node: Filter) -> Filter {
        let node = node.walk(self);
        match node.source.as_ref() {
            Stage::Filter(f) => {
                if (Self::contains_subquery(&f.condition)
                    || Self::contains_subquery(&node.condition))
                    && Self::is_at_pipeline_start(&f.source)
                {
                    // Avoid merging if it will put a filter with a subquery at the start
                    node
                } else {
                    let (conditions, is_nullable) = match &f.condition {
                        // the parent filter is already a $and, append the condition to that and
                        Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::And,
                            args,
                            is_nullable,
                            ..
                        }) => {
                            let condition_is_nullable =
                                *is_nullable || node.condition.is_nullable();
                            let mut conditions = args.clone();
                            conditions.push(node.condition);
                            (conditions, condition_is_nullable)
                        }
                        // otherwise, create a $and with the two filter conditions
                        _ => {
                            let is_nullable =
                                f.condition.is_nullable() || node.condition.is_nullable();
                            (vec![f.condition.clone(), node.condition], is_nullable)
                        }
                    };
                    Filter {
                        source: f.source.clone(),
                        condition: Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::And,
                            args: conditions,
                            force_mql_semantics: false,
                            is_nullable,
                        }),
                        cache: f.cache.clone(),
                    }
                }
            }
            _ => node,
        }
    }
}
