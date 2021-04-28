use crate::{codegen, parser};
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error)]
pub enum Error {
    #[error("parse error: {0}")]
    Parse(#[from] parser::Error),
    #[error("codegen error: {0}")]
    Codegen(#[from] codegen::Error),
    #[error("failed to serialize bson to base64: {0}")]
    SerializeBsonBase64(String),
}
