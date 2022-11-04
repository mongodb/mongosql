use crate::agg_ir::{self};
use bson::{doc, Bson};
use thiserror::Error;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
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
            Array(array) => Ok(Bson::Array(
                array
                    .into_iter()
                    .map(|e| self.codegen_agg_ir_expression(e))
                    .collect::<Result<Vec<Bson>>>()?,
            )),
            Variable(var) => Ok(Bson::String(format!("$${}", var))),
            FieldRef(fr) => Ok(Bson::String(self.codegen_field_ref(fr))),
            _ => Err(Error::UnimplementedAggIR),
        }
    }

    #[allow(clippy::only_used_in_recursion)] // false positive
    fn codegen_field_ref(&self, field_ref: agg_ir::FieldRefExpr) -> String {
        match field_ref.parent {
            None => format!("${}", field_ref.name),
            Some(parent) => format!("{}.{}", self.codegen_field_ref(*parent), field_ref.name),
        }
    }

    pub fn codegen_agg_ir_stage(&self, stage: agg_ir::Stage) -> Result<MqlTranslation> {
        match stage {
            agg_ir::Stage::Project(_p) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Group(_g) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Limit(_l) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Sort(_s) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Collection(c) => self.codegen_collection(c),
            agg_ir::Stage::Join(_j) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Unwind(_u) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Lookup(_l) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::ReplaceRoot(_r) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Match(_m) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::UnionWith(_u) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Skip(_s) => Err(Error::UnimplementedAggIR),
            agg_ir::Stage::Documents(d) => self.codegen_documents(d),
        }
    }

    fn codegen_documents(&self, agg_ir_docs: agg_ir::Documents) -> Result<MqlTranslation> {
        let docs = agg_ir_docs
            .array
            .into_iter()
            .map(|e| self.codegen_agg_ir_expression(e))
            .collect::<Result<Vec<Bson>>>()?;
        Ok(MqlTranslation {
            database: None,
            collection: None,
            pipeline: vec![doc! {"$documents": Bson::Array(docs)}],
        })
    }

    fn codegen_collection(&self, agg_ir_coll: agg_ir::Collection) -> Result<MqlTranslation> {
        Ok(MqlTranslation {
            database: Some(agg_ir_coll.db),
            collection: Some(agg_ir_coll.collection),
            pipeline: vec![],
        })
    }
}
