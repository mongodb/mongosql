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
    pub pipeline: Vec<bson::Document>,
}

impl MqlTranslation {
    fn with_additional_stage(mut self, stage: bson::Document) -> Self {
        self.pipeline.push(stage);
        self
    }
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
            Limit(l) => Ok(self
                .codegen_stage(*l.source)?
                .with_additional_stage(bson::doc! {"$limit": l.limit})),
            Offset(o) => Ok(self
                .codegen_stage(*o.source)?
                .with_additional_stage(bson::doc! {"$skip": o.offset})),
            Sort(_) => unimplemented!(),
            Collection(c) => Ok(MqlTranslation {
                database: c.db,
                collection: Some(c.collection.clone()),
                mapping_registry: mappings! { (&c.collection, 0u16).into() => &c.collection },
                pipeline: vec![bson::doc! {"$project": {"_id": 0, &c.collection: "$$ROOT"}}],
            }),
            Array(_) => unimplemented!(),
            Join(_) => unimplemented!(),
            Set(_) => unimplemented!(),
        }
    }
}
