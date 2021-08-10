mod mql;
pub(crate) use mql::MqlTranslation;
use mql::{MappingRegistry, MqlCodeGenerator};

use crate::ir;
use thiserror::Error;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("binding tuple key {0:?} not found in mapping registry")]
    ReferenceNotFound(ir::binding_tuple::Key),
    #[error("document keys may not contain dots or start with dollars")]
    DotsOrDollarsInDocumentKey,
    #[error("cannot generate MQL for {0:?} function")]
    UnsupportedFunction(ir::ScalarFunction),
    #[error("GROUP BY keys must be valid field references")]
    InvalidGroupKey,
    #[error("sort key must be a field reference")]
    InvalidSortKey,
    #[error("field paths can only be generated for Reference and FieldAccess exprs")]
    NoFieldPathForExpr,
}

pub fn generate_mql(plan: ir::Stage) -> Result<MqlTranslation> {
    let cg = MqlCodeGenerator {
        mapping_registry: MappingRegistry::new(),
        scope_level: 0u16,
    };
    cg.codegen_stage(plan)
}
