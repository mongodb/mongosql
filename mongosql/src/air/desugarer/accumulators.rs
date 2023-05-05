use crate::{
    air::{
        desugarer::{Pass, Result},
        visitor::Visitor,
        AccumulatorExpr, AggregationFunction, Expression,
        Expression::*,
        Group, LiteralValue, MQLOperator, MQLSemanticOperator, Project, ProjectItem, Reduce, Stage,
        Stage::*,
    },
    make_cond_expr, map,
};
use linked_hash_map::LinkedHashMap;

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
    fn make_count_cond(arg: Expression) -> Expression {
        let arg_is_missing_or_null_check = MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::In,
            args: vec![
                MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Type,
                    args: vec![arg],
                }),
                Array(vec![
                    Literal(LiteralValue::String("missing".to_string())),
                    Literal(LiteralValue::String("null".to_string())),
                ]),
            ],
        });
        make_cond_expr!(
            arg_is_missing_or_null_check,
            Literal(LiteralValue::Integer(0)),
            Literal(LiteralValue::Integer(1))
        )
    }

    fn create_distinct_count_project_expression(alias: String) -> Expression {
        Reduce(Reduce {
            input: Box::new(FieldRef(alias.into())),
            init_value: Box::new(Literal(LiteralValue::Integer(0))),
            inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Add,
                args: vec![
                    Variable("value".to_string().into()),
                    Self::make_count_cond(Variable("this".to_string().into())),
                ],
            })),
        })
    }

    fn create_distinct_non_count_project_expression(agg: AccumulatorExpr) -> Expression {
        MQLSemanticOperator(MQLSemanticOperator {
            op: match agg.function {
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
            args: vec![FieldRef(agg.alias.into())],
        })
    }

    fn create_non_distinct_count_aggregation(alias: String, arg: Expression) -> AccumulatorExpr {
        AccumulatorExpr {
            alias,
            function: AggregationFunction::Sum,
            distinct: false,
            arg: Box::new(Self::make_count_cond(arg)),
        }
    }
}

impl Visitor for AccumulatorsDesugarerVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Group(group) => {
                let mut new_aggregations: Vec<AccumulatorExpr> = Vec::new();
                let mut project_specs: LinkedHashMap<String, ProjectItem> = map! {
                    "_id".into() => ProjectItem::Inclusion
                };
                let mut needs_project = false;
                for agg in group.aggregations.clone().into_iter() {
                    let (new_aggregation, project_item) = if agg.distinct {
                        needs_project = true;
                        let new_aggregation = AccumulatorExpr {
                            alias: agg.alias.clone(),
                            function: AggregationFunction::AddToSet,
                            distinct: false,
                            arg: Box::new(*agg.arg.clone()),
                        };

                        let project_item = ProjectItem::Assignment(
                            if agg.function == AggregationFunction::Count {
                                // distinct count
                                Self::create_distinct_count_project_expression(agg.alias.clone())
                            } else {
                                // distinct non-count
                                Self::create_distinct_non_count_project_expression(agg.clone())
                            },
                        );
                        (new_aggregation, project_item)
                    } else {
                        let new_aggregation = if agg.function == AggregationFunction::Count {
                            // non-distinct count
                            Self::create_non_distinct_count_aggregation(agg.alias.clone(), *agg.arg)
                        } else {
                            // non-distinct non-count
                            agg.clone()
                        };
                        let project_item = ProjectItem::Inclusion;
                        (new_aggregation, project_item)
                    };
                    new_aggregations.push(new_aggregation);
                    project_specs.insert(agg.alias.clone(), project_item);
                }

                let new_group = Group(Group {
                    source: group.source,
                    keys: group.keys,
                    aggregations: new_aggregations,
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
                arg: _
            }
        ) {
            AccumulatorExpr {
                alias: node.alias,
                function: AggregationFunction::AddToSet,
                distinct: false,
                arg: node.arg,
            }
        } else {
            node
        }
    }
}
