/// Removes dead code from the query plan, currently:
///
/// 1) Optimizes Group stages by swapping with preceding Project stages when
///    possible. If a Group stage's source is a Project and all References in
///    the Group can be substituted with their definitions from Project then
///    the Project and Group can be swapped. This optimization enables index
///    utilization for some queries.
///
/// 2) Removes AddFields right before Project stages, setting is_add_fields as
///    appropriate. In the future we could merge adjacent Project stages, but
///    there is currently an issue with partial updates of documents from
///    projects that we would need to account for.
#[cfg(test)]
mod test;

use crate::{
    SchemaCheckingMode,
    mir::{
        Expression, Group, OptionallyAliasedExpr, Project, Stage,
        binding_tuple::{BindingTuple, DatasourceName, Key},
        optimizer::Optimizer,
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
    },
};
use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;

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
    fn group_dead_code_elimination(&mut self, g: Group) -> Stage {
        // For now, only consider Group stages with Project sources.
        if !matches!(*g.source, Stage::Project(_)) {
            return Stage::Group(g);
        }

        // If the Group's source is a Project, we can attempt to swap the Group with the Project if
        // all the Group's Keys can be substituted with definitions from the Project.
        let og = g.clone();
        let p = match *g.source {
            Stage::Project(p) => p,
            _ => unreachable!(), // We already returned if this is the case
        };
        let uses = og.datasource_uses();

        // To swap a Group with its source Project, we must ensure that the Project references all
        // relevant datasources emitted by the Group. Unaliased group keys are emitted under their
        // corresponding datasource. Aliased group keys and aggregations are emitted under the
        // Bottom datasource. This loop tracks which datasources are used by the group keys and
        // builds the new Project expression. In particular, if there are any unaliased group keys,
        // we must be sure to include each of those keys' datasources in the new Project expression.
        let mut new_expr = BindingTuple::new();
        let mut has_aliased_expr = false;
        for k in g.keys.iter() {
            match k {
                OptionallyAliasedExpr::Aliased(_) => {
                    has_aliased_expr = true;
                }
                OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(f)) => {
                    if let Some(r) = f.ref_if_pure()
                        && let Some(v) = p.expression.get(&r.key)
                    {
                        // At this point, we are attempting to swap a Group with a Project
                        // parent. If the Project defines more fields than the set of unaliased
                        // group keys, then the Project would no longer be valid if we directly
                        // swapped the stages. After the swap, the Group stage (now the
                        // parent of the Project) only emits group keys and aggregations;
                        // therefore, any additional fields referenced by the Project would no
                        // longer exist in the Group stage's output.
                        //
                        // The state where a Project could define additional fields is when the
                        // Project expression is a Document literal that contains any fields
                        // besides the unaliased group keys. To avoid a situation where the
                        // now-child Project stage includes such fields, we must be careful with
                        // how we form the new Project expression.
                        //
                        // We must consider the type of expression the Project defines for this
                        // datasource. If it is anything other than a literal Document, we can
                        // simply continue to use the original expression. However, if it is a
                        // literal Document expression, we must only retain the fields that are
                        // referenced by the Group as unaliased keys.
                        let expr = match v {
                            Expression::Document(og_doc) => {
                                let maybe_existing_doc = new_expr.get(&r.key);
                                let mut doc =
                                    if let Some(Expression::Document(d)) = maybe_existing_doc {
                                        // If the new expression already contains this datasource,
                                        // it should be a Document expression. We need to extend the
                                        // Document to include the new unaliased group key field.
                                        d.document.clone()
                                    } else if maybe_existing_doc.is_none() {
                                        // If the new expression does not contain this datasource,
                                        // we need to create a fresh Document expression for it.
                                        UniqueLinkedHashMap::new()
                                    } else {
                                        // If the new expression contains this datasource, but it is
                                        // not a Document expression, we cannot swap the Group with
                                        // the Project. This should never happen, but instead of
                                        // panicking or doing something undefined, we return the
                                        // original Group stage for safety and correctness.
                                        return Stage::Group(og);
                                    };

                                // We must retain the full path that exists in the original doc,
                                // which may or may not include the full path from the unaliased
                                // group key. For example, consider:
                                // ```
                                // SELECT *
                                // FROM FLATTEN(FLATTEN(foo WITH DEPTH => 1) WITH DEPTH => 1)
                                // GROUP BY a_b_c
                                // ```
                                //
                                // This optimization will apply twice (to move the Group past
                                // the two Projects associated with the two Flatten stages).
                                // This query algebrizes into roughly this MIR:
                                // ```
                                // Group {
                                //   keys: [Unaliased('a_b_c')],
                                //   source: Project {
                                //     expression: { 'foo': { 'a_b_c': 'foo.a_b.c' } }
                                //     source: Project {
                                //       expression: { 'foo': { 'a_b': 'foo.a.b' } }
                                //       source: Collection('foo')
                                //     }
                                //   }
                                // }
                                // ```
                                //
                                // The first time this optimization applies, the unaliased group
                                // key is a single component (by definition!) so we trivially
                                // retain the full path 'foo.a_b_c' from the original document.
                                // When the swap occurs, the unaliased group key is effectively
                                // rewritten to 'foo.a_b.c' (assuming constant folding).
                                //
                                // The second time this optimization applies, the unaliased
                                // group key is multi-component, so we need to retain as much of
                                // the full path as exists in the origianl document. In this
                                // example, the parent Project defines the Document
                                // { 'a_b': 'foo.a.b' }, which only includes the `a_b` field,
                                // not the `c` path component.
                                //
                                // You can extrapolate out this example to see how some swaps
                                // would retain more than one component of the full path.
                                //
                                // This is a safe unwrap since we know the FieldAccess is pure.
                                let path_components = f.path_if_pure().unwrap();
                                if let Some(retained_doc) = og_doc.retain_path(&path_components) {
                                    let val =
                                        retained_doc.document.get(&path_components[0]).unwrap();
                                    let _ = doc.insert(path_components[0].clone(), val.clone());
                                }

                                Expression::Document(doc.into())
                            }
                            _ => v.clone(),
                        };

                        new_expr.insert(r.key.clone(), expr);
                    } else {
                        // If we encounter an impure unaliased group key, or if the Project does not
                        // define the unaliased group key's datasource, we cannot swap the Group
                        // with the Project.
                        return Stage::Group(og);
                    }
                }
                OptionallyAliasedExpr::Unaliased(_) => {
                    // If we encounter a non-FieldAccess unaliased group key, we cannot swap the
                    // Group with the Project.
                    return Stage::Group(og);
                }
            }
        }

        // Again, if there is at least one aliased group key or aggregation, we must include the
        // Bottom datasource in the swapped Project. Otherwise, we would lose those values.
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
                // After substituting Reference definitions from the Project into the Group, we can
                // swap the Group with the Project.
                Ok(g) => {
                    self.changed = true;
                    Stage::Project(Project {
                        source: Box::new(Stage::Group(Group {
                            source: p.source,
                            // We obliterate the cache here because it is invalidated by the swap.
                            cache: SchemaCache::new(),
                            ..g
                        })),
                        expression: new_expr,
                        is_add_fields: false,
                        cache: SchemaCache::new(),
                    })
                }
                // It is possible for substitution to fail if the Group clause contains Subqueries.
                // This should be very rare.
                Err(g) => Stage::Group(*g),
            }
        } else {
            // If the Group uses Keys that are not defined by the Project source, do not perform
            // the substitution.
            Stage::Group(og)
        }
    }

    fn project_dead_code_elimination(&mut self, mut p: Project) -> Stage {
        // If this stage is an AddFields, we skip it for this early limited optimization
        if p.is_add_fields {
            return Stage::Project(p);
        }
        // If a Project is preceded by an AddFields, we can merge them.
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
        // just remove an AddFields stage if it defines exactly a subset of what the project after
        // it does with the same values.
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
            Stage::Group(g) => self.group_dead_code_elimination(g),
            Stage::Project(p) => self.project_dead_code_elimination(p),
            _ => node,
        }
    }
}
