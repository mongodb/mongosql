use crate::air::{
    self,
    desugarer::{Pass, Result},
};

/// Desugars any SQL operators that require SQL null semantics into their
/// corresponding MQL operators wrapped in operations to null-check the
/// arguments.
pub struct SQLNullSemanticsOperatorsDesugarerPass;

impl Pass for SQLNullSemanticsOperatorsDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        todo!()
    }
}
