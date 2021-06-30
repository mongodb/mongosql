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
    #[error("field names may not contain dots or dollars")]
    DotsOrDollarsInFieldName,
    #[error("sort keys must be field references")]
    InvalidSortKey,
    #[error("cannot generate MQL for {0:?} function")]
    UnsupportedFunction(ir::ScalarFunction),
}

pub fn generate_mql(plan: ir::Stage) -> Result<MqlTranslation> {
    let cg = MqlCodeGenerator {
        mapping_registry: MappingRegistry::new(),
    };
    cg.codegen_stage(plan)
}
