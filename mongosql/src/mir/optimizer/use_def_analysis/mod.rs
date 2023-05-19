#[cfg(test)]
mod test;

use crate::{
    mir::{
        binding_tuple::Key, visitor::Visitor, Expression, FieldAccess, Filter, Group, Project,
        ReferenceExpr, Sort, Stage, Unwind,
    },
    set,
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};
use std::collections::{BTreeMap, BTreeSet};

impl Project {
    #[allow(dead_code)]
    fn defines(&self) -> BTreeMap<Key, Expression> {
        self.expression
            .iter()
            .map(|(key, e)| (key.clone(), e.clone()))
            .collect()
    }
}

fn find_field_access_root(field_access: &Expression) -> Key {
    // We could use the UseVisitor to find the root of the FieldAccess here, but it would
    // require cloning the FieldAccess, which I wish to avoid, because the UseVisitor needs
    // ownership of the Expression.
    let mut root_finder = field_access;
    while let Expression::FieldAccess(FieldAccess { ref expr, .. }) = root_finder {
        root_finder = &**expr;
    }
    if let Expression::Reference(ReferenceExpr { ref key, .. }) = root_finder {
        return key.clone();
    }
    // This must only be used on FieldAccesses rooted with ReferenceExpr or ReferenceExpr directly.
    // Since this is currently only used for the the path argument of Unwind, this is safe.
    // If this unreachable is triggered, we know why.
    unreachable!()
}

impl Group {
    #[allow(dead_code)]
    fn defines(&self) -> BTreeMap<Key, Expression> {
        let mut bot_doc = UniqueLinkedHashMap::new();
        let mut out = BTreeMap::new();
        for group_key in self.keys.iter() {
            if let Some(alias) = group_key.get_alias() {
                // duplicate keys will be impossible at this stage, so expect should never trigger.
                bot_doc
                    .insert(alias.to_string(), group_key.get_expr().clone())
                    .expect(
                    "Duplicate Group Key should be impossible during stage_movement optimization",
                )
            }
            // If there is no alias, there is actually no need to register an expression
            // because the group_key does not modify the expression, and thus no substitution
            // will be necessary
        }
        if !bot_doc.is_empty() {
            out.insert(Key::bot(self.scope), Expression::Document(bot_doc.into()));
        }
        out
    }

    #[allow(dead_code)]
    fn opaque_defines(&self) -> BTreeSet<Key> {
        if !self.aggregations.is_empty() {
            set! { Key::bot(self.scope) }
        } else {
            BTreeSet::new()
        }
    }
}

impl Unwind {
    #[allow(dead_code)]
    fn opaque_defines(&self) -> BTreeSet<Key> {
        set! { find_field_access_root(&self.path) }
    }
}

#[derive(Clone, Debug, Default)]
struct UseVisitor {
    uses: BTreeSet<Key>,
}

impl Visitor for UseVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        // We only compute uses for Filter and Sort Stages at this time. We need to make sure
        // we do not recurse down the source field.
        match node {
            Stage::Filter(Filter {
                source,
                condition,
                cache,
            }) => {
                let condition = self.visit_expression(condition);
                Stage::Filter(Filter {
                    source,
                    condition,
                    cache,
                })
            }
            Stage::Sort(Sort {
                source,
                specs,
                cache,
            }) => {
                let specs = specs
                    .into_iter()
                    .map(|s| self.visit_sort_specification(s))
                    .collect();
                Stage::Sort(Sort {
                    source,
                    specs,
                    cache,
                })
            }
            _ => unimplemented!(),
        }
    }

    fn visit_reference_expr(&mut self, node: ReferenceExpr) -> ReferenceExpr {
        self.uses.insert(node.key.clone());
        node
    }
}

#[derive(Clone, Debug, Default)]
struct SubstituteVisitor {
    // It is traditional to call the substitution map in term rewriting the Greek symbol theta.
    theta: BTreeMap<Key, Expression>,
}

impl Visitor for SubstituteVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        match node {
            Expression::Reference(ReferenceExpr { ref key, .. }) => {
                self.theta.get(key).cloned().unwrap_or(node)
            }
            _ => node.walk(self),
        }
    }
}

impl Stage {
    // We compute uses so that we can easily check if any opaque_defines are used by a stage.
    // We do not care about normal defines which can be substituted.
    #[allow(dead_code)]
    fn uses(self) -> (BTreeSet<Key>, Stage) {
        let mut visitor = UseVisitor::default();
        let ret = visitor.visit_stage(self);
        (visitor.uses, ret)
    }

    #[allow(dead_code)]
    fn substitute(self, theta: BTreeMap<Key, Expression>) -> Self {
        let mut visitor = SubstituteVisitor { theta };
        // We only implement substitute for Stages we intend to move for which substitution makes
        // sense: Filter and Sort. Substitution is unneeded for Limit and Offset. Substitution must
        // be very targetted. For instance, if we just visit a stage it would substitute into all
        // the entire pipeline by recursing through the source. This is probably not an issue since
        // the Key just should not exist, but best to be controlled. If nothing else it results
        // in better efficiency. Also, note that substitution cannot invalidate the SchemaCache.
        match self {
            Stage::Filter(Filter {
                condition,
                source,
                cache,
            }) => Stage::Filter(Filter {
                condition: visitor.visit_expression(condition),
                source,
                cache,
            }),
            Stage::Sort(Sort {
                source,
                specs,
                cache,
            }) => Stage::Sort(Sort {
                source,
                // here we need to map over the sort specifications because we do not want to
                // substitute into the source by using visit_sort
                specs: specs
                    .into_iter()
                    .map(|x| visitor.visit_sort_specification(x))
                    .collect(),
                cache,
            }),
            // We could add no-ops for Limit and Offset, but it's better to just not call
            // substitute while we move them!
            _ => unimplemented!(),
        }
    }
}
