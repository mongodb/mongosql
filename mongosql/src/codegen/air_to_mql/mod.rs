use crate::air::{self};
use bson::{bson, doc, Bson};
use thiserror::Error;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("air method is not implemented")]
    UnimplementedAIR,
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
    pub fn codegen_air_expression(&self, expr: air::Expression) -> Result<bson::Bson> {
        use air::{Expression::*, LiteralValue::*};
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
                        .map(|(k, v)| Ok((k, self.codegen_air_expression(v)?)))
                        .collect::<Result<bson::Document>>()?
                }
            })),
            Array(array) => Ok(Bson::Array(
                array
                    .into_iter()
                    .map(|e| self.codegen_air_expression(e))
                    .collect::<Result<Vec<Bson>>>()?,
            )),
            Variable(var) => Ok(Bson::String(format!("$${var}"))),
            FieldRef(fr) => Ok(Bson::String(self.codegen_field_ref(fr))),
            GetField(gf) => Ok({
                let input = self.codegen_air_expression(*gf.input)?;
                bson!({
                    "$getField": {
                        "field": gf.field,
                        "input": input,
                    }
                })
            }),
            _ => Err(Error::UnimplementedAIR),
        }
    }

    #[allow(clippy::only_used_in_recursion)] // false positive
    fn codegen_field_ref(&self, field_ref: air::FieldRef) -> String {
        match field_ref.parent {
            None => format!("${}", field_ref.name),
            Some(parent) => format!("{}.{}", self.codegen_field_ref(*parent), field_ref.name),
        }
    }

    pub fn codegen_air_stage(&self, stage: air::Stage) -> Result<MqlTranslation> {
        match stage {
            air::Stage::Project(_p) => Err(Error::UnimplementedAIR),
            air::Stage::Group(_g) => Err(Error::UnimplementedAIR),
            air::Stage::Limit(_l) => Err(Error::UnimplementedAIR),
            air::Stage::Sort(_s) => Err(Error::UnimplementedAIR),
            air::Stage::Collection(c) => self.codegen_collection(c),
            air::Stage::Join(_j) => Err(Error::UnimplementedAIR),
            air::Stage::Unwind(_u) => Err(Error::UnimplementedAIR),
            air::Stage::Lookup(_l) => Err(Error::UnimplementedAIR),
            air::Stage::ReplaceRoot(r) => self.codegen_replace_root(r),
            air::Stage::Match(_m) => Err(Error::UnimplementedAIR),
            air::Stage::UnionWith(_u) => Err(Error::UnimplementedAIR),
            air::Stage::Skip(_s) => Err(Error::UnimplementedAIR),
            air::Stage::Documents(d) => self.codegen_documents(d),
        }
    }

    fn codegen_replace_root(&self, air_replace_root: air::ReplaceRoot) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_replace_root.source)?;
        let mut pipeline = source_translation.pipeline;
        let expr = self.codegen_air_expression(*air_replace_root.new_root)?;

        pipeline.push(doc! {"$replaceRoot": {"newRoot": expr}});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_documents(&self, air_docs: air::Documents) -> Result<MqlTranslation> {
        let docs = air_docs
            .array
            .into_iter()
            .map(|e| self.codegen_air_expression(e))
            .collect::<Result<Vec<Bson>>>()?;
        Ok(MqlTranslation {
            database: None,
            collection: None,
            pipeline: vec![doc! {"$documents": Bson::Array(docs)}],
        })
    }

    fn codegen_collection(&self, air_coll: air::Collection) -> Result<MqlTranslation> {
        Ok(MqlTranslation {
            database: Some(air_coll.db),
            collection: Some(air_coll.collection),
            pipeline: vec![],
        })
    }
}
