use crate::{
    air::{
        desugarer::{Pass, Result},
        visitor::Visitor,
        AccumulatorExpr, AggregationFunction, Expression,
        Expression::*,
        Group, LiteralValue, MQLOperator, MQLSemanticOperator, Map, Project, ProjectItem, Stage,
        Stage::*,
    },
    make_cond_expr, map,
    util::{REMOVE, ROOT, ROOT_NAME},
};
use linked_hash_map::LinkedHashMap;
use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;

macro_rules! count_single_expr_not_null_or_missing_cond {
    ($arg:expr) => {
        MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::In,
            args: vec![
                MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Type,
                    args: vec![$arg],
                }),
                Array(vec![
                    Literal(LiteralValue::String("missing".to_string())),
                    Literal(LiteralValue::String("null".to_string())),
                ]),
            ],
        })
    };
}

macro_rules! count_doc_arg_not_empty_cond {
    ($arg:expr) => {
        MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Ne,
            args: vec![$arg, Document(UniqueLinkedHashMap::new())],
        })
    };
}

macro_rules! count_doc_arg_not_all_null_values_cond {
    ($arg:expr) => {
        MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::AllElementsTrue,
            args: vec![MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::IfNull,
                args: vec![
                    Map(Map {
                        input: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                            op: MQLOperator::ObjectToArray,
                            args: vec![$arg],
                        })),
                        as_name: None,
                        inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                            op: MQLOperator::Ne,
                            args: vec![Variable("this.v".into()), Literal(LiteralValue::Null)],
                        })),
                    }),
                    Array(vec![Literal(LiteralValue::Boolean(false))]),
                ],
            })],
        })
    };
}

macro_rules! make_single_expr_count_conditional {
    ($arg:expr, $arg_is_possibly_doc:expr, $then:expr, $else:expr) => {{
        let mut cond = count_single_expr_not_null_or_missing_cond!($arg);
        if $arg_is_possibly_doc {
            cond = MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Or,
                args: vec![
                    cond,
                    // TODO: actually, need to negate these! might be better to
                    //       make the then/elses consistent across single- and
                    //       multi-arg so that we don't have to worry about this
                    //       situation.
                    count_doc_arg_not_empty_cond!($arg),
                    count_doc_arg_not_all_null_values_cond!($arg),
                ],
            });
        }
        Box::new(match $arg {
            // No need to create a conditional if the argument is a literal value.
            // We know null will result in 0 and non-null will result in 1.
            Literal(LiteralValue::Null) => Literal(LiteralValue::Integer(0)),
            Literal(_) => Literal(LiteralValue::Integer(1)),
            _ => make_cond_expr!(cond, $then, $else),
        })
    }};
}

macro_rules! make_doc_expr_count_conditional {
    ($arg:expr, $then:expr, $else:expr) => {
        Box::new(make_cond_expr!(
            MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::And,
                args: vec![
                    count_doc_arg_not_empty_cond!($arg),
                    count_doc_arg_not_all_null_values_cond!($arg),
                ],
            }),
            $then,
            $else
        ))
    };
}

/// Desugars any aggregations in Group stages into appropriate, equivalent
/// expressions and/or stages. Specifically, aggregations with distinct: true
/// are replaced in the Group stage with AddToSet and the Group is followed
/// by a Project that performs the actual target aggregation operation.
pub struct AccumulatorsDesugarerPass;

impl Pass for AccumulatorsDesugarerPass {
    fn apply(&self, pipeline: Stage) -> Result<Stage> {
        let mut accumulator_expr_converter = AccumulatorExpressionConverter;
        let new_pipeline = accumulator_expr_converter.visit_stage(pipeline);

        let mut visitor = AccumulatorsDesugarerVisitor;
        Ok(visitor.visit_stage(new_pipeline))
    }
}

#[derive(Default)]
struct AccumulatorsDesugarerVisitor;

impl AccumulatorsDesugarerVisitor {
    /// Rewrites $sqlCount into the appropriate new accumulator expression and project item.
    /// MongoSQL supports the following cases for COUNT, and each is handled differently.
    ///     - COUNT(DISTINCT? <col>)
    ///     - COUNT(DISTINCT? <col1>, <col2>, ...)
    ///     - COUNT(DISTINCT? *)
    ///
    /// For the first case, when counting non-documents, we want to omit missing and null values;
    /// when counting documents, we want to omit empty documents and documents where all values are
    /// null.
    ///
    /// For the second case, we know with certainty that we are counting documents (a column list is
    /// represented as a document with those columns as fields). In this case, we want to omit empty
    /// documents and documents where all values are null.
    ///
    /// For the third case, we again know with certainty that we are counting documents (* is
    /// represented as the "$$ROOT" variable). The * indicates we want to count all documents
    /// unconditionally.
    ///
    /// For all cases, if the count is marked as distinct, we rewrite the accumulator to a $addToSet
    /// and follow it with a projection assignment that does the counting (via $size). If it is
    /// non-distinct, we rewrite the accumulator to use $sum. The conditions described above for
    /// each case are applied to the $addToSet or $sum, depending on the value of distinct.
    fn rewrite_count(count_expr: &AccumulatorExpr) -> (AccumulatorExpr, ProjectItem) {
        let new_acc_expr = match count_expr.arg.as_ref() {
            Variable(v) if v.parent.is_none() && v.name == *ROOT_NAME => {
                Self::rewrite_count_star(count_expr.distinct, count_expr.alias.clone())
            }
            Document(_) => Self::rewrite_count_multi_col(
                count_expr.distinct,
                count_expr.alias.clone(),
                count_expr.arg.as_ref(),
            ),
            arg => Self::rewrite_count_single_expr(
                count_expr.distinct,
                count_expr.alias.clone(),
                arg,
                count_expr.arg_is_possibly_doc,
            ),
        };

        let project_item = if count_expr.distinct {
            ProjectItem::Assignment(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Size,
                args: vec![FieldRef(count_expr.alias.clone().into())],
            }))
        } else {
            ProjectItem::Inclusion
        };

        (new_acc_expr, project_item)
    }

    fn rewrite_count_star(distinct: bool, alias: String) -> AccumulatorExpr {
        let (function, arg) = if distinct {
            (AggregationFunction::AddToSet, Box::new(ROOT.clone()))
        } else {
            (
                AggregationFunction::Sum,
                Box::new(Literal(LiteralValue::Integer(1))),
            )
        };
        AccumulatorExpr {
            alias,
            function,
            distinct: false,
            arg,
            arg_is_possibly_doc: false,
        }
    }

    fn rewrite_count_multi_col(distinct: bool, alias: String, arg: &Expression) -> AccumulatorExpr {
        let (agg_func, then, r#else) = if distinct {
            (
                // When distinct, we rewrite to AddToSet.
                AggregationFunction::AddToSet,
                // In the case the condition evaluates to true, we include the argument in the set.
                arg.clone(),
                // In the case the condition evaluates to false, we include "$$REMOVE" in the set,
                // which effectively doesn't add to the set.
                REMOVE.clone(),
            )
        } else {
            (
                // When not distinct, we rewrite to Sum.
                AggregationFunction::Sum,
                // In the case the condition evaluates to true, we count this argument.
                Literal(LiteralValue::Integer(1)),
                // In the case the condition evaluates to false, we do not count this argument.
                Literal(LiteralValue::Integer(0)),
            )
        };

        AccumulatorExpr {
            alias,
            function: agg_func,
            distinct: false,
            arg: make_doc_expr_count_conditional!(arg.clone(), then, r#else),
            arg_is_possibly_doc: false,
        }
    }

    fn rewrite_count_single_expr(
        distinct: bool,
        alias: String,
        arg: &Expression,
        arg_is_possibly_doc: bool,
    ) -> AccumulatorExpr {
        // Note that the then and else values are opposite to the multi-col version. This is a
        // consequence of the condition expressions we decided to use. We could make these match
        // each other by negating one or the other, but the runtime overhead of that is, although
        // minimal, is probably not worth it. Careful testing and documentation here ensures correct
        // intended behavior.
        let (agg_func, then, r#else) = if distinct {
            (
                // When distinct, we rewrite to AddToSet.
                AggregationFunction::AddToSet,
                // In the case the condition evaluates to true, we include "$$REMOVE" in the set,
                // which effectively doesn't add to the set.
                REMOVE.clone(),
                // In the case the condition evaluates to false, we include the argument in the set.
                arg.clone(),
            )
        } else {
            (
                // When not distinct, we rewrite to Sum.
                AggregationFunction::Sum,
                // In the case the condition evaluates to true, we do not count this argument.
                Literal(LiteralValue::Integer(0)),
                // In the case the condition evaluates to false, we count this argument.
                Literal(LiteralValue::Integer(1)),
            )
        };

        AccumulatorExpr {
            alias,
            function: agg_func,
            distinct: false,
            arg: make_single_expr_count_conditional!(
                arg.clone(),
                arg_is_possibly_doc,
                then,
                r#else
            ),
            arg_is_possibly_doc: false,
        }
    }

    /// Rewrites distinct non-count accumulators into the appropriate new accumulator expression and
    /// project item. These accumulators are rewritten into AddToSet accumulators which are followed
    /// by $project assignment expressions that perform the actual accumulator operation on the set.
    fn rewrite_distinct_non_count(acc_expr: &AccumulatorExpr) -> (AccumulatorExpr, ProjectItem) {
        let new_acc_expr = AccumulatorExpr {
            alias: acc_expr.alias.clone(),
            function: AggregationFunction::AddToSet,
            distinct: false,
            arg: acc_expr.arg.clone(),
            arg_is_possibly_doc: false,
        };

        let project_item = ProjectItem::Assignment(MQLSemanticOperator(MQLSemanticOperator {
            op: match acc_expr.function {
                AggregationFunction::AddToArray => unreachable!(),
                AggregationFunction::AddToSet => unreachable!(),
                AggregationFunction::Avg => MQLOperator::Avg,
                AggregationFunction::Count => unreachable!(),
                AggregationFunction::First => MQLOperator::First,
                AggregationFunction::Last => MQLOperator::Last,
                AggregationFunction::Max => MQLOperator::Max,
                AggregationFunction::MergeDocuments => MQLOperator::MergeObjects,
                AggregationFunction::Min => MQLOperator::Min,
                AggregationFunction::StddevPop => MQLOperator::StddevPop,
                AggregationFunction::StddevSamp => MQLOperator::StddevSamp,
                AggregationFunction::Sum => MQLOperator::Sum,
            },
            args: vec![FieldRef(acc_expr.alias.clone().into())],
        }));

        (new_acc_expr, project_item)
    }
}

impl Visitor for AccumulatorsDesugarerVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Group(group) => {
                let mut new_accumulators: Vec<AccumulatorExpr> = Vec::new();
                let mut project_specs: LinkedHashMap<String, ProjectItem> = map! {
                    "_id".into() => ProjectItem::Inclusion
                };
                let mut needs_project = false;
                for acc_expr in group.aggregations.iter() {
                    needs_project = needs_project || acc_expr.distinct;
                    let (new_acc_expr, project_item) =
                        if acc_expr.function == AggregationFunction::Count {
                            Self::rewrite_count(acc_expr)
                        } else if acc_expr.distinct {
                            Self::rewrite_distinct_non_count(acc_expr)
                        } else {
                            (acc_expr.clone(), ProjectItem::Inclusion)
                        };

                    new_accumulators.push(new_acc_expr);
                    project_specs.insert(acc_expr.alias.clone(), project_item);
                }

                let new_group = Group(Group {
                    source: group.source,
                    keys: group.keys,
                    aggregations: new_accumulators,
                });

                if needs_project {
                    Project(Project {
                        source: Box::new(new_group),
                        specifications: project_specs.into(),
                    })
                } else {
                    new_group
                }
            }
            _ => node,
        }
    }
}

/// This visitor rewrites DISTINCT ADD_TO_ARRAY to ADD_TO_SET. This can be done
/// directly in the $group stage as opposed to how the other distinct operators
/// are handled (by replacing the accumulator with $addToSet and deferring the
/// original operation to a follow-up $project stage).
struct AccumulatorExpressionConverter;

impl Visitor for AccumulatorExpressionConverter {
    fn visit_accumulator_expr(&mut self, node: AccumulatorExpr) -> AccumulatorExpr {
        let node = node.walk(self);

        if matches!(
            node,
            AccumulatorExpr {
                alias: _,
                function: AggregationFunction::AddToArray,
                distinct: true,
                arg: _,
                arg_is_possibly_doc: _,
            }
        ) {
            AccumulatorExpr {
                alias: node.alias,
                function: AggregationFunction::AddToSet,
                distinct: false,
                arg: node.arg,
                arg_is_possibly_doc: false,
            }
        } else {
            node
        }
    }
}
