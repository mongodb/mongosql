#[cfg(test)]
mod test;

use crate::{
    catalog::Catalog,
    mir::{
        binding_tuple::Key, optimizer::constant_folding::ConstantFoldExprVisitor,
        schema::SchemaInferenceState, visitor::Visitor, ExistsExpr, Expression, FieldAccess,
        Filter, Group, Project, ReferenceExpr, Sort, SortSpecification, Stage, SubqueryComparison,
        SubqueryExpr, Unwind,
    },
    schema::SchemaEnvironment,
    set,
    util::unique_linked_hash_map::UniqueLinkedHashMap,
    SchemaCheckingMode,
};
use std::collections::{BTreeMap, BTreeSet};

pub(crate) fn find_field_access_root(field_access: &Expression) -> Option<Key> {
    // We could use the UseVisitor to find the root of the FieldAccess here, but it would
    // require cloning the FieldAccess, which I wish to avoid, because the UseVisitor needs
    // ownership of the Expression.
    let mut root_finder = field_access;
    while let Expression::FieldAccess(FieldAccess { ref expr, .. }) = root_finder {
        root_finder = &**expr;
    }
    if let Expression::Reference(ReferenceExpr { ref key, .. }) = root_finder {
        return Some(key.clone());
    }
    // If the FieldAccess is not rooted in a Reference, we return None.
    None
}

impl Project {
    fn defines(&self) -> BTreeMap<Key, Expression> {
        self.expression
            .iter()
            .map(|(key, e)| (key.clone(), e.clone()))
            .collect()
    }
}

impl Group {
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

    fn opaque_defines(&self) -> BTreeSet<Key> {
        if !self.aggregations.is_empty() {
            set! { Key::bot(self.scope) }
        } else {
            BTreeSet::new()
        }
    }
}

impl Unwind {
    fn opaque_defines(&self) -> BTreeSet<Key> {
        // Since we are checking the root of an Unwind, this unwrap
        // must be safe.
        set! { find_field_access_root(&self.path).unwrap() }
    }
}

impl Stage {
    pub fn defines(&self) -> BTreeMap<Key, Expression> {
        match self {
            Stage::Group(n) => n.defines(),
            Stage::Project(n) => n.defines(),
            _ => BTreeMap::new(),
        }
    }

    pub fn opaque_defines(&self) -> BTreeSet<Key> {
        match self {
            Stage::Group(n) => n.opaque_defines(),
            Stage::Unwind(n) => n.opaque_defines(),
            _ => BTreeSet::new(),
        }
    }
}

#[derive(Clone, Debug, Default)]
struct SingleStageUseVisitor {
    uses: BTreeSet<Key>,
}

impl Visitor for SingleStageUseVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        // We only compute uses for Filter, Group, and Sort Stages at this time.
        // We need to make sure we do not recurse down the source field.
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
            Stage::Group(Group {
                source,
                keys,
                aggregations,
                scope,
                cache,
            }) => {
                let keys = keys
                    .into_iter()
                    .map(|k| self.visit_optionally_aliased_expr(k))
                    .collect();
                let aggregations = aggregations
                    .into_iter()
                    .map(|a| self.visit_aliased_aggregation(a))
                    .collect();
                Stage::Group(Group {
                    source,
                    keys,
                    aggregations,
                    scope,
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

    fn visit_subquery_expr(&mut self, node: SubqueryExpr) -> SubqueryExpr {
        // When we visit a SubqueryExpr in a Filter, we need to create a new Visitor that
        // collects ALL uses from the SubqueryExpr.
        let mut all_use_visitor = AllUseVisitor::default();
        let node = node.walk(&mut all_use_visitor);
        self.uses.extend(all_use_visitor.uses);
        node
    }

    fn visit_subquery_comparison(&mut self, node: SubqueryComparison) -> SubqueryComparison {
        // When we visit a SubqueryComparison in a Filter, we need to create a new Visitor that
        // collects ALL uses from the SubqueryComparison.
        let mut all_use_visitor = AllUseVisitor::default();
        let node = node.walk(&mut all_use_visitor);
        self.uses.extend(all_use_visitor.uses);
        node
    }

    fn visit_exists_expr(&mut self, node: ExistsExpr) -> ExistsExpr {
        // When we visit an ExistsExpr in a Filter, we need to create a new Visitor that
        // collects ALL uses from the SubqueryComparison.
        let mut all_use_visitor = AllUseVisitor::default();
        let node = node.walk(&mut all_use_visitor);
        self.uses.extend(all_use_visitor.uses);
        node
    }
}

#[derive(Clone, Debug, Default)]
struct AllUseVisitor {
    uses: BTreeSet<Key>,
}

impl Visitor for AllUseVisitor {
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
    pub fn uses(self) -> (BTreeSet<Key>, Stage) {
        let mut visitor = SingleStageUseVisitor::default();
        let ret = visitor.visit_stage(self);
        (visitor.uses, ret)
    }

    pub fn substitute(self, theta: BTreeMap<Key, Expression>) -> Self {
        let mut visitor = SubstituteVisitor { theta };
        // We only implement substitute for Stages we intend to move for which substitution makes
        // sense: Filter, Group, and Sort. Substitution is unneeded for Limit and Offset.
        // Substitution must be very targeted. For instance, if we just visit a stage it would
        // substitute into all the entire pipeline by recursing through the source. This is probably
        // not an issue since the Key just should not exist, but best to be controlled. If nothing
        // else it results in better efficiency. Also, note that substitution cannot invalidate the
        // SchemaCache.
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
            Stage::Group(Group {
                keys,
                aggregations,
                source,
                scope,
                cache,
            }) => Stage::Group(Group {
                source,
                keys: keys
                    .into_iter()
                    .map(|x| visitor.visit_optionally_aliased_expr(x))
                    .collect(),
                aggregations: aggregations
                    .into_iter()
                    .map(|x| visitor.visit_aliased_aggregation(x))
                    .collect(),
                scope,
                cache,
            }),
            // We could add no-ops for Limit and Offset, but it's better to just not call
            // substitute while we move them!
            _ => unimplemented!(),
        }
    }

    /// attempt_substitute attempts to substitute theta into self, but will only
    /// succeed if all substituted expressions pass the provided condition. None
    /// is returned otherwise.
    pub fn attempt_substitute<F>(
        self,
        theta: BTreeMap<Key, Expression>,
        condition: F,
    ) -> Option<Self>
    where
        F: Fn(Box<Expression>) -> bool,
    {
        // Substitute theta
        let substituted = self.substitute(theta);

        match substituted {
            Stage::Sort(Sort {
                source,
                specs,
                cache,
            }) => {
                // Create a constant folding visitor that uses empty SchemaInferenceState.
                // The state is irrelevant since we are only constant-folding so that
                // FieldAccesses that are rooted at Document expressions are simplified.
                let c = &Catalog::default();
                let empty_state = SchemaInferenceState::new(
                    0,
                    SchemaEnvironment::default(),
                    c,
                    SchemaCheckingMode::Relaxed,
                );
                let mut visitor = ConstantFoldExprVisitor { state: empty_state };

                let mut folded_specs = vec![];
                for spec in specs {
                    // Attempt to constant fold each sort spec
                    let folded_spec = visitor.visit_sort_specification(spec);

                    // If the result meets the provided condition, retain it.
                    if condition(folded_spec.clone().get_expr()) {
                        folded_specs.push(folded_spec)
                    } else {
                        // Otherwise, substitution is not possible so return None.
                        return None;
                    }
                }

                Some(Stage::Sort(Sort {
                    source,
                    specs: folded_specs,
                    cache,
                }))
            }
            // Anything other than Sort is safe
            _ => Some(substituted),
        }
    }
}

impl SortSpecification {
    fn get_expr(self) -> Box<Expression> {
        match self {
            SortSpecification::Asc(e) => e,
            SortSpecification::Desc(e) => e,
        }
    }
}
