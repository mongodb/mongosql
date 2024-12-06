///
/// Stage Movement
///
/// The Stage Movement passes moves stages up in the MIR query plan as high as possible
/// in order to improve the usage of indexes and reduce the number of generated documents.
///
/// Filter and MatchFilter stages are moved both to improve the usage of indexes and reduce the
/// number of computations on documents that are eventually thrown away.
///
/// Sort is moved as high as possible in order to improve index usage.
///
/// Limit and Offset are moved as highly as possible to reduce the number of computations on
/// documents that are eventually thrown away.
///
/// Note that this optimization uses the `field_uses` function which is an exponential algorithm.
/// There is an extremely miniscule chance that this could cause long-running query compilation
/// times. In case we ever see a customer query stuck in the optimization phase for a long time,
/// this is likely a prime candidate to investigate first. The telltale sign for this would be
/// one or more extremely deeply nested field paths in a query (specifically in WHERE, HAVING, or
/// ORDER BY clauses). For example, `... WHERE a.b.c.d....x.y.z.even.further....extreme.nesting`.
///
#[cfg(test)]
mod test;

use super::Optimizer;
use crate::mir::optimizer::util::ContainsSubqueryVisitor;
use crate::{
    mir::{
        binding_tuple::Key, schema::SchemaInferenceState, visitor::Visitor, Derived, EquiJoin,
        Expression, Filter, Group, Join, JoinType, LateralJoin, Limit, MQLStage, MatchFilter,
        Offset, Project, ScalarFunction, ScalarFunctionApplication, Set, Sort, Stage, Unwind,
    },
    schema::ResultSet,
    SchemaCheckingMode,
};
use std::collections::HashSet;

impl Stage {
    // This is used for moving Offset as high as possible. They can be moved ahead of Any
    // Stage that is not defined as offset invalidating.
    fn is_offset_invalidating(&self) -> bool {
        match self {
            // A tautological Filter will not invalidate offset, but we don't have a SAT solver.
            Stage::Filter(_) => true,
            Stage::Project(_) => false,
            // It's possible to have a Group that does not modify cardinality and thus invalidate
            // an offset, but we can consider that a very rare occurrence.
            Stage::Group(_) => true,
            Stage::Limit(_) => true,
            // ordering of two Offsets does not matter. Offset 5, Offset 3 = Offset 3, Offset 5.
            Stage::Offset(_) => false,
            Stage::Sort(_) => true,
            Stage::Collection(_) => true,
            Stage::Array(_) => true,
            Stage::Join(_) => true,
            Stage::Set(_) => true,
            // We can push the offset into the derived query.
            Stage::Derived(_) => false,
            // It's possible to have an Unwind that does not modify cardinality and thus invalidate
            // an offset, but we can consider that a very rare occurrence.
            Stage::Unwind(_) => true,
            Stage::MQLIntrinsic(_) => true,
            Stage::Sentinel => unreachable!(),
        }
    }

    pub(crate) fn take_sources(self) -> (Vec<Stage>, Self) {
        match self {
            Stage::Filter(n) => (
                vec![*n.source],
                Stage::Filter(Filter {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Project(n) => (
                vec![*n.source],
                Stage::Project(Project {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Group(n) => (
                vec![*n.source],
                Stage::Group(Group {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Limit(n) => (
                vec![*n.source],
                Stage::Limit(Limit {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Offset(n) => (
                vec![*n.source],
                Stage::Offset(Offset {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Sort(n) => (
                vec![*n.source],
                Stage::Sort(Sort {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Collection(_) => (vec![], self),
            Stage::Array(_) => (vec![], self),
            Stage::Sentinel => (vec![], self),
            Stage::Join(n) => (
                vec![*n.left, *n.right],
                Stage::Join(Join {
                    left: Box::new(Stage::Sentinel),
                    right: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Set(n) => (
                vec![*n.left, *n.right],
                Stage::Set(Set {
                    left: Box::new(Stage::Sentinel),
                    right: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Derived(n) => (
                vec![*n.source],
                Stage::Derived(Derived {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            Stage::Unwind(n) => (
                vec![*n.source],
                Stage::Unwind(Unwind {
                    source: Box::new(Stage::Sentinel),
                    ..n
                }),
            ),
            // Only consider the LHS for EquiJoins since the RHS must
            // remain as just a collection source.
            Stage::MQLIntrinsic(MQLStage::EquiJoin(n)) => (
                vec![*n.source],
                Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
                    source: Box::new(Stage::Sentinel),
                    ..n
                })),
            ),
            Stage::MQLIntrinsic(MQLStage::LateralJoin(n)) => (
                vec![*n.source, *n.subquery],
                Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
                    source: Box::new(Stage::Sentinel),
                    subquery: Box::new(Stage::Sentinel),
                    ..n
                })),
            ),
            Stage::MQLIntrinsic(MQLStage::MatchFilter(n)) => (
                vec![*n.source],
                Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                    source: Box::new(Stage::Sentinel),
                    ..n
                })),
            ),
        }
    }

    pub(crate) fn change_sources(self, mut sources: Vec<Stage>) -> Stage {
        match self {
            Stage::Filter(s) => Stage::Filter(Filter {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Project(s) => Stage::Project(Project {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Group(s) => Stage::Group(Group {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Limit(s) => Stage::Limit(Limit {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Offset(s) => Stage::Offset(Offset {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Sort(s) => Stage::Sort(Sort {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Join(s) => Stage::Join(Join {
                left: sources.swap_remove(0).into(),
                right: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Set(s) => Stage::Set(Set {
                left: sources.swap_remove(0).into(),
                right: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Derived(s) => Stage::Derived(Derived {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Unwind(s) => Stage::Unwind(Unwind {
                source: sources.swap_remove(0).into(),
                ..s
            }),
            Stage::Collection(_) | Stage::Array(_) => self,
            // Only consider the LHS for EquiJoins since the RHS must
            // remain as just a collection source.
            Stage::MQLIntrinsic(MQLStage::EquiJoin(s)) => {
                Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
                    source: sources.swap_remove(0).into(),
                    ..s
                }))
            }
            Stage::MQLIntrinsic(MQLStage::LateralJoin(s)) => {
                Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
                    source: sources.swap_remove(0).into(),
                    subquery: sources.swap_remove(0).into(),
                    ..s
                }))
            }
            Stage::MQLIntrinsic(MQLStage::MatchFilter(s)) => {
                Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                    source: sources.swap_remove(0).into(),
                    ..s
                }))
            }
            Stage::Sentinel => unreachable!(),
        }
    }
}

pub(crate) struct StageMovementOptimizer {}

impl Optimizer for StageMovementOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        StageMovementOptimizer::move_stages(st, schema_state)
    }
}

impl StageMovementOptimizer {
    fn move_stages(st: Stage, schema_state: &SchemaInferenceState) -> (Stage, bool) {
        let mut v = StageMovementVisitor {
            schema_state,
            changed: false,

            // We start by assuming the only change is Filter-Filter swap.
            // This should be invalidated iff we observe a different swap.
            is_filter_filter_only_change: true,
        };
        let new_stage = v.visit_stage(st);
        (new_stage, v.changed && !v.is_filter_filter_only_change)
    }
}

#[derive(PartialEq, Debug, Copy, Clone)]
enum BubbleUpSide {
    Left,
    Right,
    Both,
}

#[derive(Clone)]
struct StageMovementVisitor<'a> {
    schema_state: &'a SchemaInferenceState<'a>,
    changed: bool,

    // This optimization allows Filter stages to move above other Filter
    // stages. This is useful when later Filter stages can move higher
    // than earlier ones or when only the later ones can be pushed into
    // Join conditions, etc.
    //
    // Every time a stage actually moves, this optimization should report
    // that it "changed" the pipeline. However, this Filter-Filter reorder
    // is not really a change, so we need to ensure the optimization does
    // not report it as one. In fact, if it were reported as a change,
    // then the optimizer would run indefinitely by doing the Filter-Filter
    // swap over and over again. As in, run 1 moves Filter1 above Filter2,
    // then run 2 moves Filter2 above Filter1, then run 3 moves Filter1
    // above Filter2, and so on.
    //
    // This flag should be set to true when the only change made by the
    // optimization is moving a Filter stage above another Filter stage.
    // The optimization will report that it "changed" the pipeline iff
    // `changed` is true and `is_filter_filter_only_change` is false.
    is_filter_filter_only_change: bool,
}

impl StageMovementVisitor<'_> {
    // bubble_up should only be used after we are _sure_ we want to move a Stage up. The handle_x
    // methods should be determining if the swap is necessary.
    fn bubble_up(
        &mut self,
        f: impl Fn(&mut Self, Stage) -> (Stage, bool),
        stage: Stage,
        side: BubbleUpSide,
    ) -> Stage {
        let (mut sources, stage) = stage.take_sources();
        // We do not bubble up Join, Set, Collection, or Array, so sources.len() must be 1, or it's
        // unimplemented (for now).
        if sources.len() != 1 {
            unimplemented!();
        }
        let out = sources.swap_remove(0);

        // Check if this is a Filter-Filter swap. If it is any other kind of
        // stage movement, set is_filter_filter_only_change to false.
        match (&stage, &out) {
            (Stage::Filter(_), Stage::Filter(_))
            | (Stage::Filter(_), Stage::MQLIntrinsic(MQLStage::MatchFilter(_)))
            | (Stage::MQLIntrinsic(MQLStage::MatchFilter(_)), Stage::Filter(_))
            | (
                Stage::MQLIntrinsic(MQLStage::MatchFilter(_)),
                Stage::MQLIntrinsic(MQLStage::MatchFilter(_)),
            ) => (),
            _ => self.is_filter_filter_only_change = false,
        };

        let (mut new_sources, out) = out.take_sources();
        // We should never try to bubble_up with a source that has 0 sources (i.e. past an Array or
        // Collection)
        match new_sources.len() {
            // stage       -- the Stage we are moving
            // out         -- its source
            // new_sources -- out's source(s)
            // Here, we change stage's source to its source's source. Then we attempt to continue
            // bubbling stage up (by calling the function f, which will differ depending on where
            // bubble_up was called from) given its new source. When that is done, we assign out's
            // source to be the result of this bubbling up.
            1 => out.change_sources(vec![f(self, stage.change_sources(new_sources)).0]),
            2 => {
                let (left, right) = match side {
                    BubbleUpSide::Both => (
                        f(
                            self,
                            stage
                                .clone()
                                .change_sources(vec![new_sources.swap_remove(0)]),
                        )
                        .0,
                        f(self, stage.change_sources(vec![new_sources.swap_remove(0)])).0,
                    ),
                    BubbleUpSide::Left => (
                        f(self, stage.change_sources(vec![new_sources.swap_remove(0)])).0,
                        new_sources.swap_remove(0),
                    ),
                    BubbleUpSide::Right => (
                        new_sources.swap_remove(0),
                        f(self, stage.change_sources(vec![new_sources.swap_remove(0)])).0,
                    ),
                };
                out.change_sources(vec![left, right])
            }
            _ => unimplemented!(),
        }
    }

    // This merges the condition from a filter into a join on attribute, producing
    // a logically equivalent condition in one expression.
    fn create_new_join_condition(
        join_condition: Option<Expression>,
        filter_condition: Expression,
    ) -> Option<Expression> {
        match join_condition {
            None => Some(filter_condition),
            Some(Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                is_nullable,
                mut args,
            })) => {
                let filter_is_nullable = filter_condition.is_nullable();
                args.push(filter_condition);
                Some(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    is_nullable: is_nullable || filter_is_nullable,
                    args,
                }))
            }
            Some(condition) => Some(Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                is_nullable: condition.is_nullable() || filter_condition.is_nullable(),
                args: vec![condition, filter_condition],
            })),
        }
    }

    fn handle_offset(&mut self, node: Stage) -> (Stage, bool) {
        if let Stage::Offset(ref o) = node {
            return if o.source.is_offset_invalidating() {
                (node, false)
            } else {
                // We actually cannot bubble_up Offset past any Stage that has two sources,
                // but we use BubbleUpSide::Both as a placeholder.
                (
                    self.bubble_up(Self::handle_offset, node, BubbleUpSide::Both),
                    true,
                )
            };
        }
        // handle_offset should only be called with Offset Stages
        unreachable!()
    }

    fn has_collection_or_array_source(&self, stage: &Stage) -> Option<bool> {
        match Self::get_source(stage) {
            (Some(source), None) => {
                let result = matches!(source, Stage::Collection(_) | Stage::Array(_));
                Some(result)
            }
            _ => None,
        }
    }

    fn contains_subquery(expr: &Expression) -> bool {
        let mut visitor = ContainsSubqueryVisitor::default();
        visitor.visit_expression(expr.clone());
        visitor.contains_subquery
    }

    // Returns the source stage for stages that have a single source, None otherwise
    fn get_source(stage: &Stage) -> (Option<&Stage>, Option<&Stage>) {
        match stage {
            Stage::Sort(s) => (Some(s.source.as_ref()), None),
            Stage::Filter(f) => (Some(f.source.as_ref()), None),
            Stage::Project(p) => (Some(p.source.as_ref()), None),
            Stage::Group(g) => (Some(g.source.as_ref()), None),
            Stage::Limit(l) => (Some(l.source.as_ref()), None),
            Stage::Offset(o) => (Some(o.source.as_ref()), None),
            Stage::Unwind(u) => (Some(u.source.as_ref()), None),
            Stage::Derived(d) => (Some(d.source.as_ref()), None),
            Stage::MQLIntrinsic(MQLStage::MatchFilter(m)) => (Some(m.source.as_ref()), None),
            Stage::MQLIntrinsic(MQLStage::EquiJoin(e)) => {
                (Some(e.source.as_ref()), Some(e.from.as_ref()))
            }
            Stage::MQLIntrinsic(MQLStage::LateralJoin(l)) => {
                (Some(l.source.as_ref()), Some(l.subquery.as_ref()))
            }
            Stage::Join(j) => (Some(j.left.as_ref()), Some(j.right.as_ref())),
            Stage::Set(s) => (Some(s.left.as_ref()), Some(s.right.as_ref())),
            Stage::Collection(_) => (None, None),
            Stage::Array(_) => (None, None),
            Stage::Sentinel => (None, None),
        }
    }

    fn handle_def_user(&mut self, node: Stage) -> (Stage, bool) {
        use crate::mir::schema::CachedSchema;
        // We cannot move above a terminal node because they do not have a source, we also cannot
        // move a Sort above another Sort because that would actually change the Sort ordering.
        if Self::source_prevents_reorder(&node) {
            return (node, false);
        }

        // Check if the current node is a Filter stage with a subquery in its condition
        if let Stage::Filter(f) = &node {
            if Self::contains_subquery(&f.condition) {
                match self.has_collection_or_array_source(&f.source) {
                    Some(true) | None => {
                        // Source is at start of pipeline or None, don't move it. This
                        // is because a subquery expression appearing at the start of
                        // a pipeline results in a $lookup as the first stage, preventing any
                        // chance for index utilization.
                        return (node, false);
                    }
                    _ => (),
                }
            }
        }

        // unfortunately, due to the borrow checker, we compute uses we may not need.
        let (field_uses, node) = node.field_uses();
        let (datasource_uses, node) = node.datasource_uses();
        let source = match node {
            Stage::Sort(ref n) => &n.source,
            Stage::Filter(ref n) => &n.source,
            Stage::MQLIntrinsic(MQLStage::MatchFilter(ref n)) => &n.source,
            _ => unreachable!(),
        };
        match source.as_ref() {
            // What is tricky about dual source stages is we need to know which side actually defines the keys.
            // In a single source stage we can just assume all the keys are defined, or we would have failed schema checking.
            // Specifically the *defines* method is only for what is defined by a specific stage not what is defined in the
            // entire pipeline, so we cannot just use *defines* on the left and right source here.
            Stage::Join(ref n) => {
                let right_schema = n.right.schema(self.schema_state).unwrap();
                // If this is a filter, we cannot move it if the Join's JoinType is Left and any use
                // is in the RHS. Merging a WHERE or HAVING into a LEFT JOIN ON or into the right
                // datasource is semantically incorrect when the filter depends on the RHS values
                // due to how $unwind with preserveNullAndEmptyArrays works.
                // If this is a sort, we cannot move it if any use is in the RHS. Merging an ORDER
                // BY into the right datasource is semantically incorrect when the sort depends on
                // the RHS values because it will only sort locally per LHS value, not overall.
                if (node.is_filter() && n.join_type == JoinType::Left) || node.is_sort() {
                    for u in datasource_uses.iter() {
                        if right_schema.has_datasource(u) {
                            return (node, false);
                        }
                    }
                }
                // We have to compute the schema outside the dual_source call
                // because passing references to the left, right sources to dual_sources
                // upsets the borrow checker since we also pass *node* by value.
                let left_schema = n.left.schema(self.schema_state).unwrap();
                let (stage, changed) =
                    self.dual_source(node, datasource_uses, left_schema, right_schema, false);
                if changed {
                    (stage, true)
                } else if let Stage::Filter(f) = stage {
                    if let Stage::Join(j) = *f.source {
                        // Moving a Filter condition into a Join condition is a
                        // non-Filter-Filter swap, so we must set this to false.
                        self.is_filter_filter_only_change = false;
                        let condition = Self::create_new_join_condition(j.condition, f.condition);
                        (Stage::Join(Join { condition, ..j }), true)
                    } else {
                        unreachable!()
                    }
                } else {
                    // we cannot move a Sort or a MatchFilter into an On clause
                    (stage, false)
                }
            }
            Stage::Set(ref n) => {
                // We have to compute the schema outside of the dual_source call
                // because passing references to the left, right sources to dual_sources
                // upsets the borrow checker since we also pass *node* by value.
                let left_schema = n.left.schema(self.schema_state).unwrap();
                let right_schema = n.right.schema(self.schema_state).unwrap();
                // Set does not have an On field, so we ignore the changed bool second return
                // value.
                self.dual_source(node, datasource_uses, left_schema, right_schema, false)
            }
            // For EquiJoin, we can only move stages up the LHS (source) since the
            // RHS (from) must always remain a simple collection source. The dual_source
            // method ensures that the stage, node, is only able to move up the Left
            // source if possible.
            Stage::MQLIntrinsic(MQLStage::EquiJoin(ref n)) => {
                let left_schema = n.source.schema(self.schema_state).unwrap();
                let right_schema = n.from.schema(self.schema_state).unwrap();
                self.dual_source(node, datasource_uses, left_schema, right_schema, true)
            }
            // For LateralJoin, we can move stages up either side depending
            // on which datasources are used. If any datasources from the RHS
            // (subquery) are used, then the stage is moved up that side.
            // Recall that the LHS (source) datasources are considered in-scope
            // inside the RHS (subquery) so this is safe.
            Stage::MQLIntrinsic(MQLStage::LateralJoin(ref n)) => {
                let right_schema = n.subquery.schema(self.schema_state).unwrap();
                // If this is a filter, we cannot move it, if the Join's JoinType is Left and any use is in the RHS.
                // It is not semantically correct to merge WHERE conditions into lateral JOIN RHS clauses.
                if (node.is_filter() && n.join_type == JoinType::Left) || node.is_sort() {
                    for u in datasource_uses.iter() {
                        if right_schema.has_datasource(u) {
                            return (node, false);
                        }
                    }
                }

                let side = if datasource_uses
                    .iter()
                    .any(|u| right_schema.has_datasource(u))
                {
                    BubbleUpSide::Right
                } else {
                    BubbleUpSide::Left
                };
                (self.bubble_up(Self::handle_def_user, node, side), true)
            }
            source => {
                let opaque_field_defines = source.opaque_field_defines();
                let field_uses = if let Some(field_uses) = field_uses {
                    field_uses
                } else {
                    // if there is a computed FieldAccess, we just don't do anything to be safe.
                    return (node, false);
                };
                // We can check that the intersection is non-empty without collecting by checking
                // if next() exists.
                if field_uses
                    .intersection(&opaque_field_defines)
                    .next()
                    .is_some()
                {
                    (node, false)
                } else {
                    let theta = source.defines();
                    // unfortunately, since substitution can fail, we need to clone the node.
                    match node.substitute(theta) {
                        Ok(subbed) =>
                        // The source here is not a Set or a Join so the BubbleUpSide does not actually
                        // matter, we use Both as a placeholder.
                        {
                            (
                                self.bubble_up(Self::handle_def_user, subbed, BubbleUpSide::Both),
                                true,
                            )
                        }
                        Err(original) => (original, false),
                    }
                }
            }
        }
    }

    // A source prevents a reorder if it is a terminal source: Collection, Array, or if
    // it is a Sort and the current node is also a Sort or Group because reordering Sorts amongst themselves
    // actually changes is_nullable and reordering Groups will change sort order.
    fn source_prevents_reorder(node: &Stage) -> bool {
        match *node {
            Stage::Sort(ref n) => matches!(
                &*n.source,
                Stage::Collection(_)
                    | Stage::Array(_)
                    | Stage::Sort(_)
                    | Stage::Group(_)
                    | Stage::MQLIntrinsic(MQLStage::LateralJoin(_))
            ),
            Stage::Filter(ref n) => matches!(&*n.source, Stage::Collection(_) | Stage::Array(_)),
            Stage::MQLIntrinsic(MQLStage::MatchFilter(ref n)) => {
                matches!(&*n.source, Stage::Collection(_) | Stage::Array(_))
            }
            _ => unreachable!(),
        }
    }

    // Handle movement for dual_source stages (Join, Set, EquiJoin).
    // The bool return value is true when a change is made. We can use a false
    // value to move a filter into the On field for a Join.
    // It is invalid to move a stage up the RHS of an EquiJoin, so that is
    // prevented here.
    fn dual_source(
        &mut self,
        node: Stage,
        datasource_uses: HashSet<Key>,
        left_schema: ResultSet,
        right_schema: ResultSet,
        left_only: bool,
    ) -> (Stage, bool) {
        let mut side = BubbleUpSide::Both;
        // An interesting side effect of how this is architected is that a Filter
        // with no Key usages will be bubbled up both sides, which is actually good and correct,
        // though generally trivial.
        for u in datasource_uses.into_iter() {
            match side {
                BubbleUpSide::Left => {
                    // If we are attempting to move this stage up the Left source,
                    // but encounter a datasource used by the stage that is defined
                    // by the Right source, then we cannot bubble this stage up either
                    // side and return the node without moving.
                    // the right side of the || here should be impossible, but best to be safe
                    if right_schema.has_datasource(&u) || !left_schema.has_datasource(&u) {
                        return (node, false);
                    }
                }
                BubbleUpSide::Right => {
                    // the right side of the || here should be impossible, but best to be safe
                    if left_schema.has_datasource(&u) || !right_schema.has_datasource(&u) {
                        return (node, false);
                    }
                }
                BubbleUpSide::Both => {
                    match (
                        left_schema.has_datasource(&u),
                        right_schema.has_datasource(&u),
                    ) {
                        (true, true) => return (node, false),
                        (true, _) => {
                            side = BubbleUpSide::Left;
                        }
                        (_, true) => {
                            side = BubbleUpSide::Right;
                        }
                        // This case should be impossible, but best to be safe
                        (false, false) => return (node, false),
                    }
                }
            }
        }

        // If left_only is true (if the stage we are attempting to move above
        // is an EquiJoin), we must ensure the BubbleUpSide is not Right. It is
        // only valid to bubble up on the Left side. Both is also acceptable (in
        // the case the stage being moved uses no datasources) since EquiJoin's
        // 'from' field is not considered a source in this optimization.
        if left_only && side == BubbleUpSide::Right {
            (node, false)
        } else {
            (self.bubble_up(Self::handle_def_user, node, side), true)
        }
    }
}

impl Visitor for StageMovementVisitor<'_> {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Stage::Offset(_) => {
                let (new_node, changed) = self.handle_offset(node);
                self.changed |= changed;
                new_node
            }
            Stage::Sort(_) | Stage::MQLIntrinsic(MQLStage::MatchFilter(_)) | Stage::Filter(_) => {
                let (new_node, changed) = self.handle_def_user(node);
                self.changed |= changed;
                new_node
            }
            _ => node,
        }
    }
}
