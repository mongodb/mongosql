#[cfg(test)]
mod test;

use crate::{
    mir::{
        binding_tuple::Key, schema::SchemaCache, visitor::Visitor, Error, ExistsExpr, Expression,
        FieldAccess, FieldPath, Filter, Group, MQLStage, MatchFilter, Project, ReferenceExpr, Sort,
        Stage, SubqueryComparison, SubqueryExpr, Unwind,
    },
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};
use std::collections::{HashMap, HashSet};

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
            ret.insert(FieldPath {
                key: Key::bot(self.scope),
                fields: vec![agg.alias.clone()],
                cache: SchemaCache::new(),
            });
        }
        ret
    }
}

impl Unwind {
    pub fn opaque_field_defines(&self) -> HashSet<FieldPath> {
        let mut ret = HashSet::new();
        ret.insert(self.path.clone());
        if let Some(ref index) = self.index {
            let _ = ret.insert(FieldPath {
                key: self.path.key.clone(),
                fields: vec![index.clone()],
                cache: SchemaCache::new(),
            });
        }
        ret
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

    pub fn opaque_field_defines(&self) -> HashSet<FieldPath> {
        match self {
            Stage::Group(n) => n.opaque_field_defines(),
            Stage::Unwind(n) => n.opaque_field_defines(),
            _ => HashSet::new(),
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
            Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                source,
                condition,
                cache,
            })) => {
                let condition = self.visit_match_query(condition);
                Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                    source,
                    condition,
                    cache,
                }))
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

    fn visit_field_path(&mut self, node: FieldPath) -> FieldPath {
        if let Ok(ref mut field_uses) = self.field_uses {
            field_uses.insert(node.clone());
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
    // At some point we may prefer to change field_uses to HashSet<&FieldPath>
    // to avoid some cloning, but this would likely be a difficult change.
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
            Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                source,
                condition,
                cache,
            })) => {
                let condition = self.visit_match_query(condition);
                Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                    source,
                    condition,
                    cache,
                }))
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

    fn visit_field_path(&mut self, node: FieldPath) -> FieldPath {
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
    // Substitution can fail when trying to substitute a FieldPath
    failed: bool,
}

impl Visitor for SubstituteVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        if self.failed {
            return node;
        }
        match node {
            Expression::Reference(ReferenceExpr { ref key, .. }) => {
                self.theta.get(key).cloned().unwrap_or(node)
            }
            _ => node.walk(self),
        }
    }

    fn visit_field_path(&mut self, mut node: FieldPath) -> FieldPath {
        if let Some(rep) = self.theta.get(&node.key) {
            if self.failed {
                return node;
            }
            let mut cur = rep;
            // traverse through the fields to get to the proper place in the Document
            // that is replacing the Key in this substitution
            let mut field_idx = 0;
            loop {
                if node.fields.is_empty() {
                    break;
                }
                let field = node.fields.get(field_idx);
                match cur {
                    // Each level of the Key substitution must be a Document, FieldAccess, or
                    // Reference.
                    Expression::Document(d) => {
                        if let Some(next) = d.document.get(field.unwrap()) {
                            cur = next;
                            // since this was a Document, we need to remove addvance the field_idx
                            // so that we can get the next field and so that this current field
                            // will be omitted from the output, assuming this substitution
                            // ultimately succeeds.
                            field_idx += 1;
                        } else {
                            self.failed = true;
                            return node;
                        }
                    }
                    // If we hit a FieldAccess or Reference we end iteration and keep the remaining fields.
                    // The remaining node.fields will be concatenated with the fields from this
                    // FieldAccess, assuming it can be converted to a FieldPath, or the node.key will
                    // just be replaced, if this is a Reference.
                    Expression::FieldAccess(_) | Expression::Reference(_) => {
                        break;
                    }
                    _ => {
                        self.failed = true;
                        return node;
                    }
                }
            }
            // If we are subbing in a Reference, we simply replace the FieldPath key with the Reference key.
            if let Expression::Reference(r) = cur {
                node.key = r.key.clone();
                node.fields = node.fields.into_iter().skip(field_idx).collect();
                return node;
            }
            // If we are subbing in a FieldAccess, it must be convertable to a FieldPath
            // or this substitution has failed.
            if let Expression::FieldAccess(fa) = cur {
                let fp: Result<FieldPath, _> = fa.try_into();
                if let Ok(fp) = fp {
                    node.key = fp.key;
                    node.fields = fp
                        .fields
                        .into_iter()
                        .chain(node.fields.into_iter().skip(field_idx))
                        .collect();
                    return node;
                }
            }
        }
        self.failed = true;
        node
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

    pub fn substitute(self, theta: HashMap<Key, Expression>) -> Result<Self, Self> {
        let mut visitor = SubstituteVisitor {
            theta,
            failed: false,
        };
        // We only implement substitute for Stages we intend to move for which substitution makes
        // sense: Filter, Group, and Sort. Substitution is unneeded for Limit and Offset.
        // Substitution must be very targeted. For instance, if we just visit a stage it would
        // substitute into all the entire pipeline by recursing through the source. This is probably
        // not an issue since the Key just should not exist, but best to be controlled. If nothing
        // else it results in better efficiency. Also, note that substitution cannot invalidate the
        // SchemaCache.
        match self {
            Stage::Filter(mut f) => {
                let subbed = visitor.visit_expression(f.condition.clone());
                if visitor.failed {
                    return Err(Stage::Filter(f));
                }
                f.condition = subbed;
                Ok(Stage::Filter(f))
            }
            Stage::MQLIntrinsic(MQLStage::MatchFilter(mut f)) => {
                let subbed = visitor.visit_match_query(f.condition.clone());
                if visitor.failed {
                    return Err(Stage::MQLIntrinsic(MQLStage::MatchFilter(f)));
                }
                f.condition = subbed;
                Ok(Stage::MQLIntrinsic(MQLStage::MatchFilter(f)))
            }
            Stage::Sort(mut s) => {
                let cloned_specs = s.specs.clone();
                let mut subbed_specs = Vec::new();
                for spec in cloned_specs.into_iter() {
                    let subbed = visitor.visit_sort_specification(spec);
                    if visitor.failed {
                        return Err(Stage::Sort(s));
                    }
                    subbed_specs.push(subbed);
                }
                s.specs = subbed_specs;
                Ok(Stage::Sort(s))
            }
            Stage::Group(mut g) => {
                let cloned_keys = g.keys.clone();
                let mut subbed_keys = Vec::new();
                for key in cloned_keys.into_iter() {
                    let subbed = visitor.visit_optionally_aliased_expr(key);
                    if visitor.failed {
                        return Err(Stage::Group(g));
                    }
                    subbed_keys.push(subbed);
                }
                let cloned_aggregations = g.aggregations.clone();
                let mut subbed_aggregations = Vec::new();
                for aggregation in cloned_aggregations.into_iter() {
                    let subbed = visitor.visit_aliased_aggregation(aggregation);
                    if visitor.failed {
                        return Err(Stage::Group(g));
                    }
                    subbed_aggregations.push(subbed);
                }
                g.keys = subbed_keys;
                g.aggregations = subbed_aggregations;
                Ok(Stage::Group(g))
            }
            // We could add no-ops for Limit and Offset, but it's better to just not call
            // substitute while we move them!
            _ => unimplemented!(),
        }
    }
}
