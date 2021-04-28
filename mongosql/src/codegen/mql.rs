use crate::{
    codegen::Result,
    ir::{self, binding_tuple::Key},
};
use std::collections::HashMap;

pub struct MappingRegistry(HashMap<Key, String>);

impl MappingRegistry {
    pub fn new() -> Self {
        MappingRegistry(HashMap::new())
    }
}

impl Default for MappingRegistry {
    fn default() -> Self {
        Self::new()
    }
}

macro_rules! mappings {
	($($key:expr => $ref:expr),* $(,)?) => {
		MappingRegistry(std::iter::Iterator::collect(std::array::IntoIter::new([
			$({
				let key: Key = $key;
				let name: String = $ref.to_string();
				(key, name)
			},)*
		])))
	}
}

pub struct MqlTranslation {
    pub database: String,
    pub collection: Option<String>,
    pub mapping_registry: MappingRegistry,
    pub bson: bson::Bson,
}

pub struct MqlCodeGenerator {
    pub current_database: String,
    pub correlated_mapping_registry: MappingRegistry,
}

impl MqlCodeGenerator {
    pub fn codegen_stage(&self, stage: ir::Stage) -> Result<MqlTranslation> {
        use ir::Stage::*;
        match stage {
            Filter(_) => unimplemented!(),
            Project(_) => unimplemented!(),
            Group(_) => unimplemented!(),
            Limit(_) => unimplemented!(),
            Offset(_) => unimplemented!(),
            Sort(_) => unimplemented!(),
            Collection(c) => Ok(MqlTranslation {
                database: c.db,
                collection: Some(c.collection.clone()),
                mapping_registry: mappings! { (&c.collection, 0u16).into() => &c.collection },
                bson: bson::bson!([{"$project": {"_id": 0, &c.collection: "$$ROOT"}}]),
            }),
            Array(_) => unimplemented!(),
            Join(_) => unimplemented!(),
            Set(_) => unimplemented!(),
        }
    }
}
