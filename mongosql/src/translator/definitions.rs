use crate::{
    air,
    mapping_registry::{Key, MqlMappingRegistry},
    mir,
};
use lazy_static::lazy_static;
use mongosql_datastructures::{
    unique_linked_hash_map,
    unique_linked_hash_map::{DuplicateKeyError, UniqueLinkedHashMap, UniqueLinkedHashMapEntry},
};
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

lazy_static! {
    pub static ref ROOT: air::Expression = air::Expression::Variable("ROOT".into());
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("Struct is not implemented")]
    UnimplementedStruct,
    #[error("invalid document key '{0}': document keys may not be empty, contain dots, or start with dollars")]
    InvalidDocumentKey(String),
    #[error("binding tuple key {0:?} not found in mapping registry")]
    ReferenceNotFound(Key),
    #[error("duplicate key found: {0}")]
    DuplicateKey(#[from] DuplicateKeyError),
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

    pub fn translate_stage(&self, mir_stage: mir::Stage) -> Result<air::Stage> {
        match mir_stage {
            mir::Stage::Filter(_f) => Err(Error::UnimplementedStruct),
            mir::Stage::Project(_p) => Err(Error::UnimplementedStruct),
            mir::Stage::Group(_g) => Err(Error::UnimplementedStruct),
            mir::Stage::Limit(_l) => Err(Error::UnimplementedStruct),
            mir::Stage::Offset(_o) => Err(Error::UnimplementedStruct),
            mir::Stage::Sort(_s) => Err(Error::UnimplementedStruct),
            mir::Stage::Collection(c) => self.translate_collection(c),
            mir::Stage::Array(arr) => self.translate_array_stage(arr),
            mir::Stage::Join(_j) => Err(Error::UnimplementedStruct),
            mir::Stage::Set(_s) => Err(Error::UnimplementedStruct),
            mir::Stage::Derived(_d) => Err(Error::UnimplementedStruct),
            mir::Stage::Unwind(_u) => Err(Error::UnimplementedStruct),
        }
    }

    fn translate_array_stage(&self, mir_arr: mir::ArraySource) -> Result<air::Stage> {
        let doc_stage = air::Stage::Documents(air::Documents {
            array: mir_arr
                .array
                .iter()
                .map(|mir_expr| self.translate_expression(mir_expr.clone()))
                .collect::<Result<Vec<air::Expression>>>()?,
        });

        Ok(air::Stage::Project(air::Project {
            source: Box::new(doc_stage),
            specifications: unique_linked_hash_map! {
                mir_arr.alias => ROOT.clone(),
            },
        }))
    }

    fn translate_collection(&self, mir_collection: mir::Collection) -> Result<air::Stage> {
        let coll_stage = air::Stage::Collection(air::Collection {
            db: mir_collection.db,
            collection: mir_collection.collection.clone(),
        });

        Ok(air::Stage::Project(air::Project {
            source: Box::new(coll_stage),
            specifications: unique_linked_hash_map! {
                mir_collection.collection => ROOT.clone(),
            },
        }))
    }

    #[allow(dead_code)]
    pub fn translate_expression(&self, mir_expression: mir::Expression) -> Result<air::Expression> {
        match mir_expression {
            mir::Expression::Literal(lit) => self.translate_literal(lit.value),
            mir::Expression::Document(doc) => self.translate_document(doc.document),
            mir::Expression::Array(expr) => self.translate_array(expr.array),
            mir::Expression::Reference(reference) => self.translate_reference(reference.key),
            _ => Err(Error::UnimplementedStruct),
        }
    }

    fn translate_literal(&self, lit: mir::LiteralValue) -> Result<air::Expression> {
        Ok(air::Expression::Literal(match lit {
            mir::LiteralValue::Null => air::LiteralValue::Null,
            mir::LiteralValue::Boolean(b) => air::LiteralValue::Boolean(b),
            mir::LiteralValue::String(s) => air::LiteralValue::String(s),
            mir::LiteralValue::Integer(i) => air::LiteralValue::Integer(i),
            mir::LiteralValue::Long(l) => air::LiteralValue::Long(l),
            mir::LiteralValue::Double(d) => air::LiteralValue::Double(d),
        }))
    }

    fn translate_document(
        &self,
        mir_document: UniqueLinkedHashMap<String, mir::Expression>,
    ) -> Result<air::Expression> {
        Ok(air::Expression::Document(
            mir_document
                .into_iter()
                .map(|(k, v)| {
                    if k.starts_with('$') || k.contains('.') || k.is_empty() {
                        Err(Error::InvalidDocumentKey(k))
                    } else {
                        Ok(UniqueLinkedHashMapEntry::new(
                            k,
                            self.translate_expression(v)?,
                        ))
                    }
                })
                .collect::<Result<
                    std::result::Result<
                        UniqueLinkedHashMap<String, air::Expression>,
                        DuplicateKeyError,
                    >,
                >>()??,
        ))
    }

    fn translate_array(&self, array: Vec<mir::Expression>) -> Result<air::Expression> {
        Ok(air::Expression::Array(
            array
                .into_iter()
                .map(|x| self.translate_expression(x))
                .collect::<Result<Vec<air::Expression>>>()?,
        ))
    }

    fn translate_reference(&self, key: Key) -> Result<air::Expression> {
        self.mapping_registry
            .get(&key)
            .ok_or(Error::ReferenceNotFound(key))
            .map(|s| {
                air::Expression::FieldRef(air::FieldRefExpr {
                    parent: None,
                    name: s.clone(),
                })
            })
    }
}
