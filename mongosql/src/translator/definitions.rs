use crate::{agg_ir, ir, map, util::unique_linked_hash_map::UniqueLinkedHashMap};
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("Struct is not implemented")]
    UnimplementedStruct,
    #[error("invalid document key '{0}': document keys may not be empty, contain dots, or start with dollars")]
    InvalidDocumentKey(String),
}

pub struct MqlTranslator {}

impl MqlTranslator {
    pub fn new() -> Self {
        Self {}
    }

    pub fn translate_stage(&self, ir_stage: ir::Stage) -> Result<agg_ir::Stage> {
        match ir_stage {
            ir::Stage::Filter(_f) => Err(Error::UnimplementedStruct),
            ir::Stage::Project(_p) => Err(Error::UnimplementedStruct),
            ir::Stage::Group(_g) => Err(Error::UnimplementedStruct),
            ir::Stage::Limit(_l) => Err(Error::UnimplementedStruct),
            ir::Stage::Offset(_o) => Err(Error::UnimplementedStruct),
            ir::Stage::Sort(_s) => Err(Error::UnimplementedStruct),
            ir::Stage::Collection(_c) => Err(Error::UnimplementedStruct),
            ir::Stage::Array(arr) => self.translate_array_stage(arr),
            ir::Stage::Join(_j) => Err(Error::UnimplementedStruct),
            ir::Stage::Set(_s) => Err(Error::UnimplementedStruct),
            ir::Stage::Derived(_d) => Err(Error::UnimplementedStruct),
            ir::Stage::Unwind(_u) => Err(Error::UnimplementedStruct),
        }
    }

    fn translate_array_stage(&self, ir_arr: ir::ArraySource) -> Result<agg_ir::Stage> {
        let doc_stage = agg_ir::Stage::Documents(agg_ir::Documents {
            array: ir_arr
                .array
                .iter()
                .map(|ir_expr| self.translate_expression(ir_expr.clone()))
                .collect::<Result<Vec<agg_ir::Expression>>>()?,
        });

        Ok(agg_ir::Stage::Project(agg_ir::Project {
            source: Box::new(doc_stage),
            specifications: map! {
                ir_arr.alias => agg_ir::Expression::Variable("ROOT".into()),
            },
        }))
    }

    #[allow(dead_code)]
    pub fn translate_expression(
        &self,
        ir_expression: ir::Expression,
    ) -> Result<agg_ir::Expression> {
        match ir_expression {
            ir::Expression::Literal(lit) => self.translate_literal(lit.value),
            ir::Expression::Document(doc) => self.translate_document(doc.document),
            ir::Expression::Array(expr) => self.translate_array(expr.array),
            _ => Err(Error::UnimplementedStruct),
        }
    }

    fn translate_literal(&self, lit: ir::LiteralValue) -> Result<agg_ir::Expression> {
        Ok(agg_ir::Expression::Literal(match lit {
            ir::LiteralValue::Null => agg_ir::LiteralValue::Null,
            ir::LiteralValue::Boolean(b) => agg_ir::LiteralValue::Boolean(b),
            ir::LiteralValue::String(s) => agg_ir::LiteralValue::String(s),
            ir::LiteralValue::Integer(i) => agg_ir::LiteralValue::Integer(i),
            ir::LiteralValue::Long(l) => agg_ir::LiteralValue::Long(l),
            ir::LiteralValue::Double(d) => agg_ir::LiteralValue::Double(d),
        }))
    }

    fn translate_document(
        &self,
        ir_document: UniqueLinkedHashMap<String, ir::Expression>,
    ) -> Result<agg_ir::Expression> {
        Ok(agg_ir::Expression::Document(
            ir_document
                .into_iter()
                .map(|(k, v)| {
                    if k.starts_with('$') || k.contains('.') || k.is_empty() {
                        Err(Error::InvalidDocumentKey(k))
                    } else {
                        Ok((k, self.translate_expression(v)?))
                    }
                })
                .collect::<Result<UniqueLinkedHashMap<String, agg_ir::Expression>>>()?,
        ))
    }

    fn translate_array(&self, array: Vec<ir::Expression>) -> Result<agg_ir::Expression> {
        Ok(agg_ir::Expression::Array(
            array
                .into_iter()
                .map(|x| self.translate_expression(x))
                .collect::<Result<Vec<agg_ir::Expression>>>()?,
        ))
    }
}
