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
    schema::Satisfaction,
    util::{REMOVE, ROOT, ROOT_NAME},
};
use linked_hash_map::LinkedHashMap;
use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;

// Condition used to desugar a $sqlCount when the argument is not a document. This condition asserts
// that the arg is null or missing.
macro_rules! count_arg_is_null_or_missing_cond {
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

// Condition used to desugar a $sqlCount when the argument is a document. This condition asserts
// that the arg is an empty document.
macro_rules! count_doc_arg_is_empty_cond {
    ($arg:expr) => {
        MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![$arg, Document(UniqueLinkedHashMap::new())],
        })
    };
}

// Condition used to desugar a $sqlCount when the argument is a document. This condition asserts
// that the arg contains only null values.
macro_rules! count_doc_arg_has_all_null_values_cond {
    ($arg:expr) => {
        MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::AllElementsTrue,
            args: vec![MQLSemanticOperator(MQLSemanticOperator {
                // $objectToArray and $map return null if the input is null, but $allElementsTrue
                // throws a runtime error if the input is null. Therefore, we wrap this expression
                // in $ifNull.
                op: MQLOperator::IfNull,
                args: vec![
                    Map(Map {
                        input: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                            op: MQLOperator::ObjectToArray,
                            args: vec![$arg],
                        })),
                        as_name: None,
                        inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                            op: MQLOperator::Eq,
                            args: vec![Variable("this.v".into()), Literal(LiteralValue::Null)],
                        })),
                    }),
                    // If the arg is null, this condition should vacuously evaluate to true.
                    Array(vec![]),
                ],
            })],
        })
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
    /// MongoSQL supports the following cases for COUNT:
    ///     - COUNT(DISTINCT? *)
    ///     - COUNT(DISTINCT? <col>)
    ///     - COUNT(DISTINCT? <col1>, <col2>, ...)
    ///
    /// Note: The third case is rewritten into
    ///       COUNT(DISTINCT? {'col1': <col1>, 'col2': <col2>, ...})
    /// which effectively makes it equivalent to the second case.
    ///
    /// For the first case, the * is represented as the "$$ROOT" variable in air. The * indicates we
    /// want to count all documents unconditionally.
    ///
    /// For the other case, the count is conditional. If the argument MUST be a Document, we only
    /// count non-empty documents that contain at least one non-null value. If the argument must NOT
    /// be a Document, we only count values that are not null or missing. If the argument MAY be a
    /// Document, we only count values that are not null or missing and are non-empty and contain at
    /// least one non-null value. Note that this final case depends on the schema-checker to reject
    /// polymorphic fields in COUNT. The argument could be a nullish Document, but not an AnyOf that
    /// contains Document and other non-nullish types.
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
            arg => Self::rewrite_count_non_star(
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

    /// Rewrites COUNT(DISTINCT? *) into the appropriate AccumulatorExpr. AddToSet if it is distinct;
    /// otherwise, Sum.
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
            arg_is_possibly_doc: Satisfaction::Not,
        }
    }

    /// Rewrites COUNT(DISTINCT? <expr>) into the appropriate AccumulatorExpr. AddToSet if it is
    /// distinct; otherwise, Sum. Note that these are both conditional. A value is only counted if
    /// it is not null or missing. Additionally, if the expr may be a document, it is only counted
    /// if it is non-empty and contains at least one non-null value.
    ///
    /// Also note the <expr> may be a single column -- in the case of COUNT(col) -- or may be a
    /// document literal -- in the case of COUNT(col1, col2, ...). Ultimately, these are desugared
    /// similarly.
    fn rewrite_count_non_star(
        distinct: bool,
        alias: String,
        arg: &Expression,
        arg_is_possibly_doc: Satisfaction,
    ) -> AccumulatorExpr {
        let (agg_func, then, r#else) = if distinct {
            // Including "$$REMOVE" in the set effectively doesn't add to the set.
            (AggregationFunction::AddToSet, REMOVE.clone(), arg.clone())
        } else {
            (
                AggregationFunction::Sum,
                Literal(LiteralValue::Integer(0)),
                Literal(LiteralValue::Integer(1)),
            )
        };

        let arg = match arg {
            // No need to create a conditional if the argument is a literal value.
            // We know null will result in the then case and non-null will result in the else.
            Literal(LiteralValue::Null) => then,
            Literal(_) => r#else,
            _ => {
                // Create the conditions for counting based on whether the expr is possibly a
                // Document. If the condition evaluates to true, the expr is NOT counted.
                let cond = match arg_is_possibly_doc {
                    Satisfaction::Not => count_arg_is_null_or_missing_cond!(arg.clone()),
                    Satisfaction::Must => MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Or,
                        args: vec![
                            count_doc_arg_is_empty_cond!(arg.clone()),
                            count_doc_arg_has_all_null_values_cond!(arg.clone()),
                        ],
                    }),
                    Satisfaction::May => MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Or,
                        args: vec![
                            count_arg_is_null_or_missing_cond!(arg.clone()),
                            count_doc_arg_is_empty_cond!(arg.clone()),
                            count_doc_arg_has_all_null_values_cond!(arg.clone()),
                        ],
                    }),
                };

                make_cond_expr!(cond, then, r#else)
            }
        };

        AccumulatorExpr {
            alias,
            function: agg_func,
            distinct: false,
            arg: Box::new(arg),
            arg_is_possibly_doc: Satisfaction::Not,
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
            arg_is_possibly_doc: Satisfaction::Not,
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
                arg_is_possibly_doc: Satisfaction::Not,
            }
        } else {
            node
        }
    }
}
