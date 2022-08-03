use crate::ast;
use lalrpop_util::{lalrpop_mod, lexer::Token};
use lazy_static::lazy_static;
use thiserror::Error;

lalrpop_mod!(
    #[allow(clippy::all)]
    grammar,
    "/parser/mongosql.rs"
);

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("{0}")]
    Lalrpop(String),
}

pub type LalrpopError<'t> = lalrpop_util::ParseError<usize, Token<'t>, String>;

impl From<LalrpopError<'_>> for Error {
    fn from(e: LalrpopError<'_>) -> Self {
        Error::Lalrpop(format!("{}", e))
    }
}

lazy_static! {
    static ref QUERY_PARSER: grammar::QueryParser = grammar::QueryParser::new();
    static ref EXPRESSION_PARSER: grammar::ExpressionParser = grammar::ExpressionParser::new();
}

pub fn parse_query(input: &str) -> Result<ast::Query> {
    let query = QUERY_PARSER.parse(input)?;
    Ok(query)
}

#[cfg(test)]
pub fn parse_expression(input: &str) -> Result<ast::Expression> {
    let expr = EXPRESSION_PARSER.parse(input)?;
    Ok(expr)
}
