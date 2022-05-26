use crate::{algebrizer, ast, codegen, ir, parser, schema};
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
    SchemaInference(#[from] ir::schema::Error),
    #[error("result set to json schema conversion error: {0}")]
    JsonSchemaConversion(schema::Error),
    #[error("codegen error: {0}")]
    Codegen(#[from] codegen::Error),
}
