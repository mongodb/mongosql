use crate::ir::binding_tuple::BindingTuple;
use std::collections::HashMap;

pub type SchemaEnvironment = BindingTuple<Schema>;

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct ResultSet {
    pub schema: SchemaEnvironment,
    pub min_size: Option<u64>,
    pub max_size: Option<u64>,
}

impl Default for ResultSet {
    fn default() -> Self {
        Self {
            schema: SchemaEnvironment::default(),
            min_size: None,
            max_size: None,
        }
    }
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Schema {
    Any,
    Missing,
    Atomic(Atomic),
    OneOf(Vec<Schema>),
    AnyOf(Vec<Schema>),
    Array(Box<Schema>),
    Document(Document),
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Atomic {
    String,
    Int,
    Double,
    Long,
    Decimal,
    BinData,
    ObjectId,
    Boolean,
    Date,
    Null,
    Regex,
    DbPointer,
    Javascript,
    Symbol,
    JavascriptWithScope,
    Timestamp,
    MinKey,
    MaxKey,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Document {
    keys: HashMap<String, Schema>,
    required: Vec<String>,
    additional_properties: bool,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Satisfaction {
    May,
    Must,
    Not,
}

#[allow(dead_code, unused_variables)]
impl Schema {
    fn satisfies(&self, other: &Schema) -> Satisfaction {
        unimplemented!()
    }

    fn contains_field(&self, other: &Schema) -> Satisfaction {
        unimplemented!()
    }
}
