#[cfg(test)]
mod test;

use crate::{
    mir::{optimizer::Optimizer, schema::SchemaInferenceState, visitor::Visitor, Group, Stage},
    SchemaCheckingMode,
};

/// Optimizes Group stages by eliminating preceding Project stages when
/// possible. If a Group stage's source is a Project and all References
/// in the Group can be substituted with their definitions from Project
/// then the Project can be eliminated. This optimization enables index
/// utilization for some queries.
pub(crate) struct DeadCodeEliminator;

impl Optimizer for DeadCodeEliminator {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> Stage {
        let mut v = DeadCodeEliminationVisitor;
        v.visit_stage(st)
    }
}

struct DeadCodeEliminationVisitor;

impl Visitor for DeadCodeEliminationVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);

        match node.clone() {
            Stage::Group(g) => {
                match *g.source {
                    Stage::Project(p) => {
                        let (uses, s) = node.clone().uses();
                        let g = match s {
                            Stage::Group(g) => g,
                            _ => unreachable!(),
                        };
                        let theta = g.source.defines();
                        if uses.iter().all(|used_ref| theta.contains_key(used_ref)) {
                            let subbed = node.substitute(theta);
                            match subbed {
                                // After substituting Reference definitions from the Project
                                // into the Group, remove the Project from the Stage tree.
                                Stage::Group(g) => Stage::Group(Group {
                                    source: p.source,
                                    ..g
                                }),
                                _ => unreachable!(),
                            }
                        } else {
                            // If the Group uses Keys that are not defined by the Project source,
                            // do not perform the substitution or eliminate the Project.
                            node
                        }
                    }
                    // For now, only consider Group stages with Project sources.
                    _ => node,
                }
            }
            // For now, only consider Group stages.
            _ => node,
        }
    }
}
