#[cfg(test)]
mod test;

use super::Optimizer;
use crate::{
    mir::{visitor::Visitor, Filter, Offset, Sort, Stage},
    SchemaCheckingMode,
};

impl Stage {
    // This is used for moving Offset as high as possible. They can be moved ahead of Any
    // Stage that is not defined as offset invalidating.
    #[allow(dead_code)]
    fn is_offset_invalidating(&self) -> bool {
        match self {
            // A tautological Filter will not invalidate offset, but we don't have a SAT solver.
            Stage::Filter(_) => true,
            Stage::Project(_) => false,
            // It's possible to have a Group that does not modify cardinality and thus invalidate
            // an offset, but we can consider that a very rare occurence.
            Stage::Group(_) => true,
            Stage::Limit(_) => true,
            Stage::Offset(Offset { offset, .. }) => *offset != 0,
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
        }
    }
}

pub(crate) struct StageMovementOptimizer {}

impl Optimizer for StageMovementOptimizer {
    fn optimize(&self, st: Stage, _sm: SchemaCheckingMode) -> Stage {
        StageMovementOptimizer::move_stages(st)
    }
}

impl StageMovementOptimizer {
    fn move_stages(st: Stage) -> Stage {
        let mut v = StageMovementVisitor;
        v.visit_stage(st)
    }
}

#[derive(Copy, Clone)]
struct StageMovementVisitor;

impl StageMovementVisitor {}

impl Visitor for StageMovementVisitor {
    fn visit_offset(&mut self, node: Offset) -> Offset {
        node
    }

    // TODO SQL-1378: When we move a Sort or Filter past a Join or Set,
    // we will check if the Keys in the Sort or filter exist in the SchemaCache
    // for each of the left/right sources of the Join and Set and move the Sort/Filter
    // into whichever one has the key. If some set of the Uses exists on both sides, we will be unable to
    // move. Match splitting should result in this working for Filters where we can move
    // the sides of a Conjunct properly into the side it belongs on. But if there is a Disjunction,
    // it would be incorrect to move it since it needs the results of both sides. Sort copy-and-split
    // may eventually make sense since it *should be* faster to sort x,y if x and y are both
    // already sorted, but I'm not sure if mongodb can actually make use of that info. For now,
    // sorts will just be harder to move than filters: they will not be moved if they require keys
    // that exist in both sides of a Set or Join.
    fn visit_sort(&mut self, node: Sort) -> Sort {
        node
    }

    fn visit_filter(&mut self, node: Filter) -> Filter {
        node
    }
}
