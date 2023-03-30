use crate::air::SQLOperator;
use thiserror::Error;

#[cfg(test)]
mod test;

mod expressions;
mod functions;
mod stages;
mod utils;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("air method is not implemented")]
    UnimplementedAIR,
    #[error("cannot generate MQL for {0:?} operator")]
    UnsupportedOperator(SQLOperator),
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
