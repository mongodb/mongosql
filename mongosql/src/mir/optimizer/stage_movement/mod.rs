#[cfg(test)]
mod test;

use super::Optimizer;
use crate::{
    mir::{
        binding_tuple::Key, optimizer::use_def_analysis::find_field_access_root,
        schema::SchemaInferenceState, visitor::Visitor, Derived, Expression, Filter, Group, Join,
        Limit, Offset, Project, Set, Sort, Stage, Unwind,
    },
    schema::ResultSet,
    SchemaCheckingMode,
};
use std::collections::{BTreeMap, BTreeSet};

impl Stage {
    // This is used for moving Offset as high as possible. They can be moved ahead of Any
    // Stage that is not defined as offset invalidating.
    fn is_offset_invalidating(&self) -> bool {
        match self {
            // A tautological Filter will not invalidate offset, but we don't have a SAT solver.
            Stage::Filter(_) => true,
            Stage::Project(_) => false,
            // It's possible to have a Group that does not modify cardinality and thus invalidate
            // an offset, but we can consider that a very rare occurence.
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
            // an offset, but we can consider that a very rare occurence.
            Stage::Unwind(_) => true,
            Stage::Sentinel => unreachable!(),
        }
    }

    fn take_sources(self) -> (Vec<Stage>, Self) {
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
        }
    }

    fn change_sources(self, mut sources: Vec<Stage>) -> Stage {
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
    ) -> Stage {
        StageMovementOptimizer::move_stages(st, schema_state)
    }
}

impl StageMovementOptimizer {
    fn move_stages(st: Stage, schema_state: &SchemaInferenceState) -> Stage {
        let mut v = StageMovementVisitor {
            schema_state: schema_state.clone(),
        };
        v.visit_stage(st)
    }
}

#[derive(Debug, Copy, Clone)]
enum BubbleUpSide {
    Left,
    Right,
    Both,
}

#[derive(Debug, Clone, PartialEq)]
enum SubstitutionPossibility {
    Always,
    Maybe,
    Never,
}

#[derive(Clone)]
struct StageMovementVisitor<'a> {
    schema_state: SchemaInferenceState<'a>,
}

impl<'a> StageMovementVisitor<'a> {
    // bubble_up should only be used after we are _sure_ we want to move a Stage up. The handle_x
    // methods should be determining if the swap is necessary.
    fn bubble_up(
        &mut self,
        f: impl Fn(&mut Self, Stage) -> Stage,
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
            1 => out.change_sources(vec![f(self, stage.change_sources(new_sources))]),
            2 => {
                let (left, right) = match side {
                    BubbleUpSide::Both => (
                        f(
                            self,
                            stage
                                .clone()
                                .change_sources(vec![new_sources.swap_remove(0)]),
                        ),
                        f(self, stage.change_sources(vec![new_sources.swap_remove(0)])),
                    ),
                    BubbleUpSide::Left => (
                        f(self, stage.change_sources(vec![new_sources.swap_remove(0)])),
                        new_sources.swap_remove(0),
                    ),
                    BubbleUpSide::Right => (
                        new_sources.swap_remove(0),
                        f(self, stage.change_sources(vec![new_sources.swap_remove(0)])),
                    ),
                };
                out.change_sources(vec![left, right])
            }
            _ => unimplemented!(),
        }
    }

    fn handle_offset(&mut self, node: Stage) -> Stage {
        if let Stage::Offset(ref o) = node {
            return if o.source.is_offset_invalidating() {
                node
            } else {
                // We actually cannot bubble_up Offset past any Stage that has two sources,
                // but we use BubbleUpSide::Both as a place holder.
                self.bubble_up(Self::handle_offset, node, BubbleUpSide::Both)
            };
        }
        // handle_offset should only be called with Offset Stages
        unreachable!()
    }

    fn handle_def_user(&mut self, node: Stage) -> Stage {
        use crate::mir::schema::CachedSchema;
        // We decided to preserve the predefined order of stages that are the def_users
        // (Sort/Filter). Without this check, since the traversal is bottom up, we would always
        // swap def_users.
        if Self::source_is_terminal_or_def_user(&node) {
            return node;
        }
        let (uses, node) = node.uses();
        let (source, is_sort) = match node {
            Stage::Sort(ref n) => (&n.source, true),
            Stage::Filter(ref n) => (&n.source, false),
            _ => unreachable!(),
        };
        match source.as_ref() {
            // What is tricky about dual source stages is we need to know which side actually defines the keys.
            // In a single source stage we can just assume all the keys are defined, or we would have failed schema checking.
            // Specifically the *defines* method is only for what is defined by a specific stage not what is defined in the
            // entire pipeline, so we cannot just use *defines* on the left and right source here.
            Stage::Join(ref n) => {
                // We have to compute the schema outside of the dual_source call
                // because passing references to the left, right sources to dual_sources
                // upsets the borrow checker since we also pass *node* by value.
                let left_schema = n.left.schema(&self.schema_state).unwrap();
                let right_schema = n.right.schema(&self.schema_state).unwrap();
                self.dual_source(node, uses, left_schema, right_schema)
            }
            Stage::Set(ref n) => {
                // We have to compute the schema outside of the dual_source call
                // because passing references to the left, right sources to dual_sources
                // upsets the borrow checker since we also pass *node* by value.
                let left_schema = n.left.schema(&self.schema_state).unwrap();
                let right_schema = n.right.schema(&self.schema_state).unwrap();
                self.dual_source(node, uses, left_schema, right_schema)
            }
            source => {
                // We can check that the intersection is non-empty without collecting by checking
                // if next() exists.
                if source.opaque_defines().intersection(&uses).next().is_some() {
                    node
                } else {
                    let theta = source.defines();
                    let subbed = if is_sort {
                        match Self::is_sort_substitution_possible(&theta, &uses) {
                            SubstitutionPossibility::Always => node.substitute(theta),
                            SubstitutionPossibility::Never => return node,
                            SubstitutionPossibility::Maybe => {
                                match node.clone().attempt_substitute(theta, |e| {
                                    find_field_access_root(&e).is_some()
                                }) {
                                    None => return node,
                                    Some(subbed) => subbed,
                                }
                            }
                        }
                    } else {
                        node.substitute(theta)
                    };
                    // The source here is not a Set or a Join so the BubbleUpSide does not actually
                    // matter, we use Both as a place holder.
                    self.bubble_up(Self::handle_def_user, subbed, BubbleUpSide::Both)
                }
            }
        }
    }

    fn source_is_terminal_or_def_user(node: &Stage) -> bool {
        matches!(
            match *node {
                Stage::Sort(ref n) => &*n.source,
                Stage::Filter(ref n) => &*n.source,
                _ => unreachable!(),
            },
            Stage::Collection(_) | Stage::Array(_) | Stage::Filter(_) | Stage::Sort(_)
        )
    }

    fn dual_source(
        &mut self,
        node: Stage,
        uses: BTreeSet<Key>,
        left_schema: ResultSet,
        right_schema: ResultSet,
    ) -> Stage {
        let mut side = BubbleUpSide::Both;
        // An interesting side affect of how this is architected is that a Filter
        // with no Key usages will be bubbled up both sides, which is actually good and correct,
        // though generally trivial.
        for u in uses.into_iter() {
            match side {
                BubbleUpSide::Left => {
                    // If we are attempting to move this stage up the Left source,
                    // but encounter a datasource used by the stage that is defined
                    // by the Right source, then we cannot bubble this stage up either
                    // side and return the node without moving.
                    // the right side of the || here should be impossible, but best to be safe
                    if right_schema.has_datasource(&u) || !left_schema.has_datasource(&u) {
                        return node;
                    }
                }
                BubbleUpSide::Right => {
                    // the right side of the || here should be impossible, but best to be safe
                    if left_schema.has_datasource(&u) || !right_schema.has_datasource(&u) {
                        return node;
                    }
                }
                BubbleUpSide::Both => {
                    match (
                        left_schema.has_datasource(&u),
                        right_schema.has_datasource(&u),
                    ) {
                        (true, true) => return node,
                        (true, _) => {
                            side = BubbleUpSide::Left;
                        }
                        (_, true) => {
                            side = BubbleUpSide::Right;
                        }
                        // This case should be impossible, but best to be safe
                        (false, false) => return node,
                    }
                }
            }
        }
        self.bubble_up(Self::handle_def_user, node, side)
    }

    // is_sort_substitution_possible determines if substitutions are possible
    // in a Sort stage. References are always substitutable. Documents and
    // FieldAccesses are possibly substitutable. All other Expressions are
    // disallowed in Sorts.
    fn is_sort_substitution_possible(
        theta: &BTreeMap<Key, Expression>,
        uses: &BTreeSet<Key>,
    ) -> SubstitutionPossibility {
        for u in uses {
            if let Some(e) = theta.get(u) {
                match e {
                    Expression::Reference(_) => (),
                    Expression::Document(_) | Expression::FieldAccess(_) => {
                        return SubstitutionPossibility::Maybe
                    }
                    _ => return SubstitutionPossibility::Never,
                }
            }
        }
        SubstitutionPossibility::Always
    }
}

impl<'a> Visitor for StageMovementVisitor<'a> {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Stage::Offset(_) => self.handle_offset(node),
            Stage::Sort(_) | Stage::Filter(_) => self.handle_def_user(node),
            _ => node,
        }
    }
}
