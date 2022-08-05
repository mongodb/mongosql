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
            Document(document) => Ok(Bson::Document({
                if document.is_empty() {
                    bson::doc! {"$literal": {}}
                } else {
                    document
                        .into_iter()
                        .map(|(k, v)| Ok((k, self.codegen_agg_ir_expression(v)?)))
                        .collect::<Result<bson::Document>>()?
                }
            })),
            _ => Err(Error::UnimplementedAggIR),
        }
    }

    pub fn codegen_agg_ir_stage(&self, stage: agg_ir::Stage) -> Result<MqlTranslation> {
        match stage {
            agg_ir::Stage::Project(_p) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Group(_g) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Limit(_l) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Sort(_s) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Collection(_c) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Join(_j) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Unwind(_u) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Lookup(_l) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::ReplaceRoot(_r) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Match(_m) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::UnionWith(_u) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Skip(_s) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Documents(_d) => Err(Error::UnimplementedAggIR),
        }
    }
}
