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

    #[allow(dead_code)]
    pub fn translate_expression(
        &self,
        ir_expression: ir::Expression,
    ) -> Result<agg_ir::Expression> {
        match ir_expression {
            ir::Expression::Literal(lit) => self.translate_literal(lit.value),
            _ => Err(Error::UnimplementedStruct),
        }
    }

    pub fn translate_literal(&self, lit: ir::LiteralValue) -> Result<agg_ir::Expression> {
        Ok(agg_ir::Expression::Literal(match lit {
            ir::LiteralValue::Null => agg_ir::LiteralValue::Null,
            ir::LiteralValue::Boolean(b) => agg_ir::LiteralValue::Boolean(b),
            ir::LiteralValue::String(s) => agg_ir::LiteralValue::String(s),
            ir::LiteralValue::Integer(i) => agg_ir::LiteralValue::Integer(i),
            ir::LiteralValue::Long(l) => agg_ir::LiteralValue::Long(l),
            ir::LiteralValue::Double(d) => agg_ir::LiteralValue::Double(d),
        }))
    }
}
