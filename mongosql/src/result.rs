use crate::{algebrizer, ast, codegen, parser};
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error)]
pub enum Error {
    #[error("parse error: {0}")]
    Parse(#[from] parser::Error),
    #[error("rewrite error: {0}")]
    Rewrite(#[from] ast::rewrites::Error),
    #[error("algebrize error: {0}")]
    Algebrize(#[from] algebrizer::Error),
    #[error("codegen error: {0}")]
    Codegen(#[from] codegen::Error),
}
