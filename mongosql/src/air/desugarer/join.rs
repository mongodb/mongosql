use crate::air::{
    self,
    desugarer::{Pass, Result},
};

/// Desugars any Join stages in the pipeline into sequences of equivalent,
/// existing MQL stages. Specifically, a Join is desugared into a sequence
/// of Lookup, Unwind, ReplaceWith, and Project.
pub struct JoinDesugarerPass;

impl Pass for JoinDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        todo!()
    }
}
