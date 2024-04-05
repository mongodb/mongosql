///
/// Optimizes Group stages by swapping with preceding Project stages when
/// possible. If a Group stage's source is a Project and all References
/// in the Group can be substituted with their definitions from Project
/// then the Project and Group can be swapped. This optimization enables index
/// utilization for some queries.
///
#[cfg(test)]
mod test;

use crate::{
    mir::{
        binding_tuple::{BindingTuple, DatasourceName, Key},
        optimizer::Optimizer,
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Expression, Group, OptionallyAliasedExpr, Project, Stage,
    },
    SchemaCheckingMode,
};

pub(crate) struct DeadCodeEliminator;

impl Optimizer for DeadCodeEliminator {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        let mut v = DeadCodeEliminationVisitor::default();
        let new_stage = v.visit_stage(st);
        (new_stage, v.changed)
    }
}

#[derive(Default)]
struct DeadCodeEliminationVisitor {
    changed: bool,
}

impl Visitor for DeadCodeEliminationVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node.clone() {
            Stage::Group(g) => {
                match *g.source {
                    Stage::Project(p) => {
                        let (uses, s) = node.clone().datasource_uses();
                        let g = match s {
                            Stage::Group(g) => g,
                            _ => unreachable!(),
                        };
                        // in order to swap a Group with its source Project, we must ensure that
                        // the Project references the Groups aggregations. Otherwise, they will be
                        // lost when the project gets translated.
                        let mut new_expr = BindingTuple::new();
                        let mut has_aliased_expr = false;
                        for k in g.keys.into_iter() {
                            match k {
                                OptionallyAliasedExpr::Unaliased(u) => {
                                    if let Expression::FieldAccess(f) = u {
                                        if let Expression::Reference(r) = *f.expr {
                                            if let Some(v) = p.expression.get(&r.key) {
                                                new_expr.insert(r.key, v.clone());
                                            }
                                        }
                                    }
                                }
                                OptionallyAliasedExpr::Aliased(_) => {
                                    has_aliased_expr = true;
                                }
                            }
                        }
                        // in the case that we don't have any Aliased expressions of aggregations,
                        // we don't want the project to reference Bottom.
                        if has_aliased_expr || !g.aggregations.is_empty() {
                            new_expr.insert(
                                Key::bot(g.scope),
                                Expression::Reference((DatasourceName::Bottom, g.scope).into()),
                            );
                        }
                        let theta = g.source.defines();
                        if uses.iter().all(|used_ref| theta.contains_key(used_ref)) {
                            let subbed = node.substitute(theta);
                            match subbed {
                                // After substituting Reference definitions from the Project
                                // into the Group, remove the Project from the Stage tree.
                                Ok(Stage::Group(g)) => {
                                    self.changed = true;
                                    Stage::Project(Project {
                                        source: Box::new(Stage::Group(Group {
                                            source: p.source,
                                            ..g
                                        })),
                                        expression: new_expr,
                                        cache: SchemaCache::new(),
                                    })
                                }
                                // It is possible for substitution to fail if the Group clause
                                // contains Subqueries. This will be very rare.
                                Err(n) => n,
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
