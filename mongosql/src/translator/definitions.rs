use crate::{
    agg_ir, ir, map,
    mapping_registry::{Key, MqlMappingRegistry},
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};
use lazy_static::lazy_static;
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

lazy_static! {
    pub static ref ROOT: agg_ir::Expression = agg_ir::Expression::Variable("ROOT".into());
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("Struct is not implemented")]
    UnimplementedStruct,
    #[error("invalid document key '{0}': document keys may not be empty, contain dots, or start with dollars")]
    InvalidDocumentKey(String),
    #[error("binding tuple key {0:?} not found in mapping registry")]
    ReferenceNotFound(Key),
}

pub struct MqlTranslator {
    pub mapping_registry: MqlMappingRegistry,
}

impl MqlTranslator {
    pub fn new() -> Self {
        Self {
            mapping_registry: Default::default(),
        }
    }

    pub fn translate_stage(&self, ir_stage: ir::Stage) -> Result<agg_ir::Stage> {
        match ir_stage {
            ir::Stage::Filter(_f) => Err(Error::UnimplementedStruct),
            ir::Stage::Project(_p) => Err(Error::UnimplementedStruct),
            ir::Stage::Group(_g) => Err(Error::UnimplementedStruct),
            ir::Stage::Limit(_l) => Err(Error::UnimplementedStruct),
            ir::Stage::Offset(_o) => Err(Error::UnimplementedStruct),
            ir::Stage::Sort(_s) => Err(Error::UnimplementedStruct),
            ir::Stage::Collection(c) => self.translate_collection(c),
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
                ir_arr.alias => ROOT.clone(),
            },
        }))
    }

    fn translate_collection(&self, ir_collection: ir::Collection) -> Result<agg_ir::Stage> {
        let coll_stage = agg_ir::Stage::Collection(agg_ir::Collection {
            db: ir_collection.db,
            collection: ir_collection.collection.clone(),
        });

        Ok(agg_ir::Stage::Project(agg_ir::Project {
            source: Box::new(coll_stage),
            specifications: map! {
                ir_collection.collection => ROOT.clone(),
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
            ir::Expression::Reference(reference) => self.translate_reference(reference.key),
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

    fn translate_reference(&self, key: Key) -> Result<agg_ir::Expression> {
        self.mapping_registry
            .get(&key)
            .ok_or(Error::ReferenceNotFound(key))
            .map(|s| {
                agg_ir::Expression::FieldRef(agg_ir::FieldRefExpr {
                    parent: None,
                    name: s.clone(),
                })
            })
    }
}
