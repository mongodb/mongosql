use crate::parser;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug)]
pub enum Error {
    Parse(String),
}

impl From<parser::ParseError<'_>> for Error {
    fn from(e: parser::ParseError<'_>) -> Self {
        Error::Parse(format!("{}", e))
    }
}
