mod mql;
pub(crate) use mql::MqlTranslation;
use mql::{MappingRegistry, MqlCodeGenerator};

use crate::ir;
use thiserror::Error;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error)]
pub enum Error {}

pub fn generate_mql(current_database: String, plan: ir::Stage) -> Result<MqlTranslation> {
    let cg = MqlCodeGenerator {
        current_database,
        correlated_mapping_registry: MappingRegistry::new(),
    };
    cg.codegen_stage(plan)
}
