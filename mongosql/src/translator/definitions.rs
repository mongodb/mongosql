use crate::{agg_ir, ir};
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("Struct is not implemented")]
    UnimplementedStruct,
}

pub struct MqlTranslator {}

impl MqlTranslator {
    pub fn new() -> Self {
        Self {}
    }

    pub fn translate_to_agg(self, _ir_stage: ir::Stage) -> Result<agg_ir::Stage> {
        Err(Error::UnimplementedStruct)
    }
}
