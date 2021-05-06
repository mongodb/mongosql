#![allow(dead_code)]

use crate::{
    ir::*,
    schema::{Atomic, ResultSet, Schema},
};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("datasource {0:?} not found in schema environment")]
    DatasourceNotFoundInSchemaEnv(binding_tuple::Key),
}

pub struct SchemaInferenceState<'a> {
    env: &'a ResultSet,
}

impl<'a> From<&'a ResultSet> for SchemaInferenceState<'a> {
    fn from(env: &'a ResultSet) -> Self {
        Self { env }
    }
}

impl Expression {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            Expression::Literal(lit) => lit.schema(state),
            Expression::Reference(key) => state
                .env
                .schema
                .get(key)
                .cloned()
                .ok_or_else(|| Error::DatasourceNotFoundInSchemaEnv(key.clone())),
            Expression::Array(_) => unimplemented!(),
            Expression::Document(_) => unimplemented!(),
            Expression::Function(_) => unimplemented!(),
            Expression::SubqueryExpression(_) => unimplemented!(),
            Expression::SubqueryComparison(_) => unimplemented!(),
            Expression::Exists(_) => unimplemented!(),
        }
    }
}

impl Literal {
    pub fn schema(&self, _state: &SchemaInferenceState) -> Result<Schema, Error> {
        use Literal::*;
        Ok(match self {
            Null => Schema::Atomic(Atomic::Null),
            Boolean(_) => Schema::Atomic(Atomic::Boolean),
            String(_) => Schema::Atomic(Atomic::String),
            Integer(_) => Schema::Atomic(Atomic::Int),
            Long(_) => Schema::Atomic(Atomic::Long),
            Double(_) => Schema::Atomic(Atomic::Double),
        })
    }
}
