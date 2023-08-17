use crate::air;
pub use crate::mapping_registry::MqlMappingRegistry;
use thiserror::Error;

#[cfg(test)]
mod test;

mod expressions;
mod functions;
mod match_query;
mod stages;
mod utils;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("cannot generate MQL for {0:?} operator")]
    UnsupportedOperator(air::SQLOperator),
    #[error("cannot $convert to document")]
    ConvertToDocument,
    #[error("cannot $convert to array")]
    ConvertToArray,
}

#[derive(PartialEq, Debug)]
pub struct MqlTranslation {
    pub database: Option<String>,
    pub collection: Option<String>,
    pub pipeline: Vec<bson::Document>,
}

#[derive(Clone, Debug)]
pub struct MqlCodeGenerator {}

pub fn generate_mql(plan: air::Stage) -> Result<MqlTranslation> {
    let cg = MqlCodeGenerator {};

    cg.codegen_stage(plan)
}
