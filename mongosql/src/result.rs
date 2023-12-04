use crate::{air::desugarer, algebrizer, ast, codegen, mir, parser, schema, translator};
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("parse error: {0}")]
    Parse(#[from] parser::Error),
    #[error("rewrite error: {0}")]
    Rewrite(#[from] ast::rewrites::Error),
    #[error("algebrize error: {0}")]
    Algebrize(#[from] algebrizer::Error),
    #[error("schema inference error: {0}")]
    SchemaInference(#[from] mir::schema::Error),
    #[error("result set to json schema conversion error: {0}")]
    JsonSchemaConversion(schema::Error),
    #[error("codegen error: {0}")]
    Codegen(#[from] codegen::Error),
    #[error("translator error: {0}")]
    Translator(#[from] translator::Error),
    #[error("desugarer error: {0}")]
    Desugarer(#[from] desugarer::Error),
    #[error("schema error: {0}")]
    Schema(#[from] schema::Error),
    #[error("catalog error: {0}")]
    Catalog(String),
}
