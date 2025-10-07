/// Removes deadcode from the query plan, currently:
///
/// 1) Optimizes Group stages by swapping with preceding Project stages when
///    possible. If a Group stage's source is a Project and all References
///    in the Group can be substituted with their definitions from Project
///    then the Project and Group can be swapped. This optimization enables index
///    utilization for some queries.
///
/// 2) Removes AddFields right before Project stages, setting is_add_fields as appropriate. In the future we could
///    merge adjacent Project stages, but there is currently an issue with partial updates of
///    documents from projects that we would need to account for.
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

impl DeadCodeEliminationVisitor {
    fn group_deadcode_elimination(&mut self, g: Group) -> Stage {
        // For now, only consider Group stages with Project sources.
        if !matches!(*g.source, Stage::Project(_)) {
            return Stage::Group(g);
        }
        // If the Group's source is a Project, we can attempt to swap the Group with the Project
        // if all of the Group's Keys can be substituted with definitions from the Project.
        // This allows us to eliminate the Project stage and potentially utilize indexes.
        let og = g.clone();
        let p = match *g.source {
            Stage::Project(p) => p,
            _ => unreachable!(), // We already returned if this is the case
        };
        let (uses, og) = og.datasource_uses();
        // in order to swap a Group with its source Project, we must ensure that
        // the Project references the Groups aggregations. Otherwise, they will be
        // lost when the project gets translated.
        let mut new_expr = BindingTuple::new();
        let mut has_aliased_expr = false;
        for k in g.keys.iter() {
            match k {
                OptionallyAliasedExpr::Unaliased(u) => {
                    if let Expression::FieldAccess(f) = u {
                        if let Expression::Reference(ref r) = *f.expr {
                            if let Some(v) = p.expression.get(&r.key) {
                                new_expr.insert(r.key.clone(), v.clone());
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
        let theta = p.defines();
        if uses.iter().all(|used_ref| theta.contains_key(used_ref)) {
            let subbed = og.substitute(theta);
            match subbed {
                // After substituting Reference definitions from the Project
                // into the Group, remove the Project from the Stage tree.
                Ok(g) => {
                    self.changed = true;
                    Stage::Project(Project {
                        source: Box::new(Stage::Group(Group {
                            source: p.source,
                            ..g
                        })),
                        expression: new_expr,
                        is_add_fields: false,
                        cache: SchemaCache::new(),
                    })
                }
                // It is possible for substitution to fail if the Group clause
                // contains Subqueries. This will be very rare.
                Err(g) => Stage::Group(*g),
            }
        } else {
            // If the Group uses Keys that are not defined by the Project source,
            // do not perform the substitution or eliminate the Project.
            Stage::Group(og)
        }
    }

    fn project_deadcode_elimination(&mut self, mut p: Project) -> Stage {
        // If this stage is an AddFields, we skip it for this early limited optimization
        if p.is_add_fields {
            return Stage::Project(p);
        }
        // If a Project is preceeded by an AddFields, we can merge them.
        if !matches!(
            *p.source,
            Stage::Project(Project {
                is_add_fields: true,
                ..
            })
        ) {
            return Stage::Project(p);
        }
        let Stage::Project(source_p) = *std::mem::replace(&mut p.source, Box::new(Stage::Sentinel))
        else {
            unreachable!(); // We already checked this
        };
        let mut failed = false;
        // Substitution does wonky things with respect to document field accesses, so for now we
        // just remove an AddFields stage if it defines exactly a subset of what the project after it does with the same values.
        for (k, v) in source_p.expression.iter() {
            if let Some(expr) = p.expression.get(k) {
                if expr != v {
                    failed = true;
                    break;
                }
            } else {
                failed = true;
                break;
            }
        }
        if failed {
            return Stage::Project(Project {
                // we failed, make sure to put the source back
                source: Box::new(Stage::Project(source_p)),
                ..p
            });
        }
        // We suceeded, so we drop the source by setting the source to the source's source
        self.changed = true;
        Stage::Project(Project {
            source: source_p.source,
            ..p
        })
    }
}

impl Visitor for DeadCodeEliminationVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Stage::Group(g) => self.group_deadcode_elimination(g),
            Stage::Project(p) => self.project_deadcode_elimination(p),
            _ => node,
        }
    }
}
