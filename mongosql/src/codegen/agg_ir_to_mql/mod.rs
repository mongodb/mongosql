use crate::agg_ir::{self};
use bson::Bson;
use thiserror::Error;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("agg_ir method is not implemented")]
    UnimplementedAggIR,
}

#[derive(PartialEq, Debug)]
pub struct MqlTranslation {
    pub database: Option<String>,
    pub collection: Option<String>,
    pub pipeline: Vec<bson::Document>,
}

impl MqlTranslation {}

#[derive(Clone)]
pub struct MqlCodeGenerator {
    pub scope_level: u16,
}

impl MqlCodeGenerator {
    #[allow(dead_code)]
    pub fn codegen_agg_ir_expression(&self, expr: agg_ir::Expression) -> Result<bson::Bson> {
        use agg_ir::{Expression::*, LiteralValue::*};
        match expr {
            Literal(lit) => Ok(bson::bson!({
                "$literal": match lit {
                    Null => Bson::Null,
                    Boolean(b) => Bson::Boolean(b),
                    String(s) => Bson::String(s),
                    Integer(i) => Bson::Int32(i),
                    Long(l) => Bson::Int64(l),
                    Double(d) => Bson::Double(d),
                },
            })),
            _ => Err(Error::UnimplementedAggIR),
        }
    }
}
