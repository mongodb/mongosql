use crate::air::{
    self,
    desugarer::{Error, Pass, Result},
};

/// Desugars any top-level subquery expressions (Subquery, SubqueryComparison,
/// and SubqueryExists). A subquery expression is desugared into three parts:
///   1. A Lookup stage that is placed before the containing stage. This
///      Lookup stage performs the subquery when the pipeline is executed.
///   2. An expression that replaces the subquery expression in the containing
///      stage. This replacement is dependent on the type of subquery expr.
///   3. A Project stage that is placed after the containing stage. This
///      Project removes the field introduced by the Lookup; if there are
///      multiple subquery expressions in a stage, there will be multiple
///      Lookup stages but only one Project stage at the end.
///      If the subquery is in a Group stage, we don't want to exclude the
///      _id because it is the group key
pub struct SubqueryExprDesugarerPass;

impl Pass for SubqueryExprDesugarerPass {
    fn apply(&self, _pipeline: air::Stage) -> Result<air::Stage> {
        Err(Error::TodoError)
    }
}
