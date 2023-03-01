use crate::air::{
    self,
    desugarer::{Pass, Result},
};

/// Desugars any aggregations in Group stages into appropriate, equivalent
/// expressions and/or stages. Specifically, aggregations with distinct: true
/// are replaced in the Group stage with AddToSet and the Group is followed
/// by a Project that performs the actual target aggregation operation.
pub struct AccumulatorsDesugarerPass;

impl Pass for AccumulatorsDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        todo!()
    }
}
