#![allow(dead_code)]

use crate::{
    ir::*,
    schema::{Atomic, ResultSet, Schema, SchemaEnvironment},
};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("datasource {0:?} not found in schema environment")]
    DatasourceNotFoundInSchemaEnv(binding_tuple::Key),
}

pub struct SchemaInferenceState<'a> {
    env: &'a SchemaEnvironment,
}

impl<'a> From<&'a SchemaEnvironment> for SchemaInferenceState<'a> {
    fn from(env: &'a SchemaEnvironment) -> Self {
        Self { env }
    }
}

impl Stage {
    /// Recursively schema checks this stage, its sources, and all
    /// contained expressions. If schema checking succeeds, returns a
    /// [`ResultSet`] describing the schema of the result set returned
    /// by this stage. The provided [`SchemaInferenceState`] should
    /// include schema information for any datasources from outer
    /// scopes. Schema information for the current scope will be
    /// obtained by calling [`Stage::schema`] on source stages.
    pub fn schema(&self, _state: SchemaInferenceState) -> Result<ResultSet, Error> {
        match self {
            Stage::Filter(_) => unimplemented!(),
            Stage::Project(_) => unimplemented!(),
            Stage::Group(_) => unimplemented!(),
            Stage::Limit(_) => unimplemented!(),
            Stage::Offset(_) => unimplemented!(),
            Stage::Sort(_) => unimplemented!(),
            Stage::Collection(_) => Ok(ResultSet::default()),
            Stage::Array(_) => unimplemented!(),
            Stage::Join(_) => unimplemented!(),
            Stage::Set(_) => unimplemented!(),
        }
    }
}

impl Expression {
    /// Recursively schema checks this expression, its arguments, and
    /// all contained expressions/stages. If schema checking succeeds,
    /// returns a [`Schema`] describing this expression's schema. The
    /// provided [`SchemaInferenceState`] should include schema
    /// information for all datasources in scope.
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            Expression::Literal(lit) => lit.schema(state),
            Expression::Reference(key) => state
                .env
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
