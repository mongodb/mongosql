use crate::{
    codegen::{Error, Result},
    ir::{self, binding_tuple::Key},
};
use std::collections::HashMap;

pub struct MappingRegistry(HashMap<Key, String>);

impl MappingRegistry {
    pub fn new() -> Self {
        MappingRegistry(HashMap::new())
    }

    pub fn insert<K: Into<Key>, V: Into<String>>(&mut self, k: K, v: V) -> Option<String> {
        self.0.insert(k.into(), v.into())
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
    pub mapping_registry: MappingRegistry,
}

impl MqlCodeGenerator {
    /// Recursively generates a translation for this stage and its
    /// sources. When this function is called, `self.mapping_registry`
    /// should include mappings for any datasources from outer scopes.
    /// Mappings for the current scope will be obtained by calling
    /// `codegen_stage` on source stages.
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

    /// Recursively generates a translation for this expression. When
    /// this function is called, `self.mapping_registry` should
    /// include mappings for all datasources in scope.
    #[allow(dead_code)]
    pub fn codegen_expression(&self, expr: ir::Expression) -> Result<bson::Bson> {
        use bson::Bson;
        use ir::{Expression::*, Literal::*};
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
            Reference(key) => self
                .mapping_registry
                .0
                .get(&key)
                .ok_or(Error::ReferenceNotFound(key))
                .map(|s| bson::Bson::String(format!("${}", s))),
            Array(_) => unimplemented!(),
            Document(_) => unimplemented!(),
            Function(_) => unimplemented!(),
            SubqueryExpression(_) => unimplemented!(),
            SubqueryComparison(_) => unimplemented!(),
            Exists(_) => unimplemented!(),
        }
    }
}
