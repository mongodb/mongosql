use crate::air::{
    self,
    desugarer::{Error, Pass, Result},
};

/// Desugars Lookup stages by removing the from_db name from any Lookup when the
/// from_db matches the current database name. This is because the information is
/// redundant and some implementations of mongodb aggregation language,
/// e.g. mongod itself, do not support the from_db field.
pub struct LookupDesugarerPass;

impl Pass for LookupDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        Err(Error::TodoError)
    }
}
