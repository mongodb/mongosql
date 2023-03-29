use crate::air::{
    self,
    desugarer::{Error, Pass, Result},
};

/// Desugars any SQL operators that do not exist in MQL (e.g. Between, Like,
/// etc.) or that do not have the same semantics in MQL (e.g. Cos, Sin, Tan,
/// etc.) into appropriate, equivalent MQL operators.
pub struct UnsupportedOperatorsDesugarerPass;

impl Pass for UnsupportedOperatorsDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        Err(Error::TodoError)
    }
}
