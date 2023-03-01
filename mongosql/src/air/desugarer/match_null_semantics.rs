use crate::air::{
    self,
    desugarer::{Pass, Result},
};

/// Desugars Match stages with SQL operators such that they can partially
/// utilize indexes. Specifically, if a Match contains a SQL operator, a
/// separate Match stage is created with a constraint for the arguments
/// that ensures the values exist and are not null. This new Match is placed
/// before the original Match and the original is updated to use the MQL
/// version of the operator.
pub struct MatchDesugarerPass;

impl Pass for MatchDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        todo!()
    }
}
