use crate::{
    air,
    mapping_registry::{Key, MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
    mir,
    translator::utils::ROOT,
};
use mongosql_datastructures::{
    binding_tuple::DatasourceName, unique_linked_hash_map::DuplicateKeyError,
};
use thiserror::Error;

#[cfg(test)]
mod test;

mod expressions;
mod stages;
mod utils;

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("invalid document key '{0}': document keys may not be empty, contain dots, or start with dollars")]
    InvalidDocumentKey(String),
    #[error("binding tuple key {0:?} not found in mapping registry")]
    ReferenceNotFound(Key),
    #[error("duplicate key found: {0}")]
    DuplicateKey(#[from] DuplicateKeyError),
    #[error("project fields may not be empty, contain dots, or start with dollars")]
    InvalidProjectField,
    #[error("invalid group key, unaliased key must be field ref")]
    InvalidGroupKey,
    #[error("invalid sqlConvert target type: {0:?}")]
    InvalidSqlConvertToType(air::Type),
    #[error("unexpected expr type for sort: not a FieldAccess or Reference")]
    ExprNotReferenceOrFieldAccess,
    #[error("LIMIT ({0}) cannot be converted to i64")]
    LimitOutOfI64Range(u64),
    #[error("expected FieldRef for subquery output path")]
    SubqueryOutputPathNotFieldRef,
}

#[derive(Clone)]
pub struct MqlTranslator {
    pub mapping_registry: MqlMappingRegistry,
    pub scope_level: u16,
}

impl MqlTranslator {
    pub fn new() -> Self {
        Self {
            mapping_registry: Default::default(),
            scope_level: 0u16,
        }
    }

    /// translate_plan is the entry point, it mostly just calls translate_stage, but also
    /// sets up a ReplaceWith to replace __bot with the empty key: ''.
    pub fn translate_plan(&mut self, mir_stage: mir::Stage) -> Result<air::Stage> {
        let source = self.translate_stage(mir_stage)?;
        self.append_replace_bot_stage(source)
    }

    /// append_replace_bot_stage will be called at the end of translate_plan to add a stage
    /// that replaces the unique bottom name with an empty string "" as the last stage.
    fn append_replace_bot_stage(&mut self, source: air::Stage) -> Result<air::Stage> {
        let key = Key {
            datasource: DatasourceName::Bottom,
            scope: 0u16,
        };
        let mongo_bot_name = self.mapping_registry.remove(&key);
        Ok(match mongo_bot_name {
            Some(registry_value) => {
                self.mapping_registry.insert(
                    key,
                    MqlMappingRegistryValue::new("".to_string(), MqlReferenceType::FieldRef),
                );
                air::Stage::ReplaceWith(air::ReplaceWith {
                    source: Box::new(source),
                    new_root: Box::new(air::Expression::UnsetField(air::UnsetField {
                        field: registry_value.name.clone(),
                        input: Box::new(air::Expression::SetField(air::SetField {
                            field: "".to_string(),
                            input: Box::new(ROOT.clone()),
                            value: Box::new(air::Expression::FieldRef(registry_value.name.into())),
                        })),
                    })),
                })
            }
            None => source,
        })
    }
}
