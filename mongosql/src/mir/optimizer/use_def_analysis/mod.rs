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
    util::unique_linked_hash_map::UniqueLinkedHashMap,
    SchemaCheckingMode,
};
use std::collections::{HashMap, HashSet};

use thiserror::Error;
#[derive(Debug, Error, PartialEq, Clone)]
pub enum Error {
    #[error("{0:?} is not a viable FieldPath")]
    InvalidFieldPath(Expression),
}

#[derive(Eq, PartialEq, Debug, Hash, Clone)]
pub enum FieldPath {
    Ref(Key),
    Field {
        parent: Box<FieldPath>,
        field: String,
    },
}

impl FieldPath {
    pub fn get_datasource(&self) -> Key {
        match self {
            Self::Ref(k) => k.clone(),
            Self::Field { parent: p, .. } => p.get_datasource(),
        }
    }
}

impl TryFrom<&Expression> for FieldPath {
    type Error = Error;

    fn try_from(value: &Expression) -> Result<Self, Self::Error> {
        match value {
            Expression::FieldAccess(ref f) => f.try_into(),
            Expression::Reference(ref r) => r.try_into(),
            _ => Err(Error::InvalidFieldPath(value.clone())),
        }
    }
}

impl TryFrom<&FieldAccess> for FieldPath {
    type Error = Error;

    fn try_from(f: &FieldAccess) -> Result<Self, Self::Error> {
        Ok(FieldPath::Field {
            parent: Box::new(f.expr.as_ref().try_into()?),
            field: f.field.clone(),
        })
    }
}

impl TryFrom<&ReferenceExpr> for FieldPath {
    type Error = Error;

    fn try_from(r: &ReferenceExpr) -> Result<Self, Self::Error> {
        Ok(FieldPath::Ref(r.key.clone()))
    }
}

pub fn find_field_access_root(field_access: &Expression) -> Option<Key> {
    // We could use the FieldUseVisitor to find the root of the FieldAccess here, but it would
    // require cloning the FieldAccess, which I wish to avoid, because the FieldUseVisitor needs
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
    pub(crate) fn defines(&self) -> HashMap<Key, Expression> {
        self.expression
            .iter()
            .map(|(key, e)| (key.clone(), e.clone()))
            .collect()
    }
}

impl Group {
    pub fn defines(&self) -> HashMap<Key, Expression> {
        let mut bot_doc = UniqueLinkedHashMap::new();
        let mut out = HashMap::new();
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

    pub fn opaque_field_defines(&self) -> HashSet<FieldPath> {
        let mut ret = HashSet::new();
        for agg in self.aggregations.iter() {
            ret.insert(FieldPath::Field {
                parent: Box::new(FieldPath::Ref(Key::bot(self.scope))),
                field: agg.alias.clone(),
            });
        }
        ret
    }
}

impl Unwind {
    pub fn opaque_field_defines(&self) -> Result<HashSet<FieldPath>, Error> {
        let mut ret = HashSet::new();
        ret.insert(self.path.as_ref().try_into()?);
        if let Some(ref index) = self.index {
            let _ = ret.insert(FieldPath::Field {
                parent: Box::new(FieldPath::Ref(
                    find_field_access_root(self.path.as_ref())
                        .ok_or_else(|| Error::InvalidFieldPath(self.path.as_ref().clone()))?,
                )),
                field: index.clone(),
            });
        }
        Ok(ret)
    }
}

impl Stage {
    pub fn defines(&self) -> HashMap<Key, Expression> {
        match self {
            Stage::Group(n) => n.defines(),
            Stage::Project(n) => n.defines(),
            _ => HashMap::new(),
        }
    }

    pub fn opaque_field_defines(&self) -> Result<HashSet<FieldPath>, Error> {
        match self {
            Stage::Group(n) => Ok(n.opaque_field_defines()),
            Stage::Unwind(n) => n.opaque_field_defines(),
            _ => Ok(HashSet::new()),
        }
    }
}

#[derive(Clone, Debug)]
struct SingleStageFieldUseVisitor {
    field_uses: Result<HashSet<FieldPath>, Error>,
}

impl Default for SingleStageFieldUseVisitor {
    fn default() -> Self {
        Self {
            field_uses: Ok(HashSet::new()),
        }
    }
}

impl Visitor for SingleStageFieldUseVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        if self.field_uses.is_err() {
            return node;
        }
        // We only compute field_uses for Filter and Sort Stages at this time. We need to make sure
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

    fn visit_field_access(&mut self, node: FieldAccess) -> FieldAccess {
        if let Ok(ref mut field_uses) = self.field_uses {
            let f: Result<FieldPath, _> = (&node).try_into();
            match f {
                Ok(fp) => {
                    let _ = field_uses.insert(fp);
                }
                Err(e) => self.field_uses = Err(e),
            }
        }
        node
    }

    fn visit_subquery_expr(&mut self, node: SubqueryExpr) -> SubqueryExpr {
        // When we visit a SubqueryExpr in a Filter, we need to create a new Visitor that
        // collects ALL field_uses from the SubqueryExpr.
        if let Ok(ref mut field_uses) = self.field_uses {
            let mut all_use_visitor = AllFieldUseVisitor::default();
            let node = node.walk(&mut all_use_visitor);
            match all_use_visitor.field_uses {
                Ok(u) => field_uses.extend(u),
                Err(e) => self.field_uses = Err(e),
            }
            node
        } else {
            node
        }
    }

    fn visit_subquery_comparison(&mut self, node: SubqueryComparison) -> SubqueryComparison {
        // When we visit a SubqueryComparison in a Filter, we need to create a new Visitor that
        // collects ALL field_uses from the SubqueryComparison.
        if let Ok(ref mut field_uses) = self.field_uses {
            let mut all_use_visitor = AllFieldUseVisitor::default();
            let node = node.walk(&mut all_use_visitor);
            match all_use_visitor.field_uses {
                Ok(u) => field_uses.extend(u),
                Err(e) => self.field_uses = Err(e),
            }
            node
        } else {
            node
        }
    }

    fn visit_exists_expr(&mut self, node: ExistsExpr) -> ExistsExpr {
        // When we visit an ExistsExpr in a Filter, we need to create a new Visitor that
        // collects ALL field_uses from the SubqueryComparison.
        if let Ok(ref mut field_uses) = self.field_uses {
            let mut all_use_visitor = AllFieldUseVisitor::default();
            let node = node.walk(&mut all_use_visitor);
            match all_use_visitor.field_uses {
                Ok(u) => field_uses.extend(u),
                Err(e) => self.field_uses = Err(e),
            }
            node
        } else {
            node
        }
    }
}

#[derive(Clone, Debug)]
struct AllFieldUseVisitor {
    field_uses: Result<HashSet<FieldPath>, Error>,
}

impl Default for AllFieldUseVisitor {
    fn default() -> Self {
        Self {
            field_uses: Ok(HashSet::new()),
        }
    }
}

impl Visitor for AllFieldUseVisitor {
    fn visit_field_access(&mut self, node: FieldAccess) -> FieldAccess {
        if let Ok(ref mut field_uses) = self.field_uses {
            let f: Result<FieldPath, _> = (&node).try_into();
            match f {
                Ok(fp) => {
                    let _ = field_uses.insert(fp);
                }
                Err(e) => self.field_uses = Err(e),
            }
        }
        node
    }
}

#[derive(Clone, Debug, Default)]
struct SingleStageDatasourceUseVisitor {
    datasource_uses: HashSet<Key>,
}

impl Visitor for SingleStageDatasourceUseVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        // We only compute datasource_uses for Filter and Sort Stages at this time. We need to make sure
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
        self.datasource_uses.insert(node.key.clone());
        node
    }

    fn visit_subquery_expr(&mut self, node: SubqueryExpr) -> SubqueryExpr {
        // When we visit a SubqueryExpr in a Filter, we need to create a new Visitor that
        // collects ALL datasource_uses from the SubqueryExpr.
        let mut all_use_visitor = AllDatasourceUseVisitor::default();
        let node = node.walk(&mut all_use_visitor);
        self.datasource_uses.extend(all_use_visitor.datasource_uses);
        node
    }

    fn visit_subquery_comparison(&mut self, node: SubqueryComparison) -> SubqueryComparison {
        // When we visit a SubqueryComparison in a Filter, we need to create a new Visitor that
        // collects ALL datasource_uses from the SubqueryComparison.
        let mut all_use_visitor = AllDatasourceUseVisitor::default();
        let node = node.walk(&mut all_use_visitor);
        self.datasource_uses.extend(all_use_visitor.datasource_uses);
        node
    }

    fn visit_exists_expr(&mut self, node: ExistsExpr) -> ExistsExpr {
        // When we visit an ExistsExpr in a Filter, we need to create a new Visitor that
        // collects ALL datasource_uses from the SubqueryComparison.
        let mut all_use_visitor = AllDatasourceUseVisitor::default();
        let node = node.walk(&mut all_use_visitor);
        self.datasource_uses.extend(all_use_visitor.datasource_uses);
        node
    }
}

#[derive(Clone, Debug, Default)]
struct AllDatasourceUseVisitor {
    datasource_uses: HashSet<Key>,
}

impl Visitor for AllDatasourceUseVisitor {
    fn visit_reference_expr(&mut self, node: ReferenceExpr) -> ReferenceExpr {
        self.datasource_uses.insert(node.key.clone());
        node
    }
}

#[derive(Clone, Debug, Default)]
struct SubstituteVisitor {
    // It is traditional to call the substitution map in term rewriting the Greek symbol theta.
    theta: HashMap<Key, Expression>,
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

impl Expression {
    // We compute field_uses so that we can easily check if any opaque_field_defines are used by a stage.
    // We do not care about normal defines which can be substituted.
    pub fn field_uses(self) -> (Result<HashSet<FieldPath>, Error>, Self) {
        let mut visitor = SingleStageFieldUseVisitor::default();
        let ret = visitor.visit_expression(self);
        (visitor.field_uses, ret)
    }
}

impl Stage {
    // We compute field_uses so that we can easily check if any opaque_field_defines are used by a stage.
    // We do not care about normal defines which can be substituted.
    pub fn field_uses(self) -> (Result<HashSet<FieldPath>, Error>, Stage) {
        let mut visitor = SingleStageFieldUseVisitor::default();
        let ret = visitor.visit_stage(self);
        (visitor.field_uses, ret)
    }

    // We compute field_uses so that we can easily check if any opaque_field_defines are used by a stage.
    // We do not care about normal defines which can be substituted.
    pub fn datasource_uses(self) -> (HashSet<Key>, Stage) {
        let mut visitor = SingleStageDatasourceUseVisitor::default();
        let ret = visitor.visit_stage(self);
        (visitor.datasource_uses, ret)
    }

    pub fn substitute(self, theta: HashMap<Key, Expression>) -> Self {
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
        theta: HashMap<Key, Expression>,
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
                // Create a constant folding visitor that field_uses empty SchemaInferenceState.
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
