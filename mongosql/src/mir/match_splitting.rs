use crate::mir::{
    schema::SchemaCache, visitor::Visitor, Expression::*, Filter, ScalarFunction::*, Stage,
    Stage::*,
};

pub fn split_matches(stage: Stage) -> Stage {
    let mut visitor = MatchSplittingVisitor;
    visitor.visit_stage(stage)
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
