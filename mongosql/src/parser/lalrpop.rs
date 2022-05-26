use crate::ast;
use lalrpop_util::{lalrpop_mod, lexer::Token};
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

pub struct Parser {
    query_parser: grammar::QueryParser,
    #[cfg(test)]
    _expression_parser: grammar::ExpressionParser,
}

impl Parser {
    pub fn new() -> Self {
        let query_parser = grammar::QueryParser::new();
        let _expression_parser = grammar::ExpressionParser::new();

        Self {
            query_parser,
            #[cfg(test)]
            _expression_parser,
        }
    }

    pub fn parse_query(&self, input: &str) -> Result<ast::Query> {
        let query = self.query_parser.parse(input)?;
        Ok(query)
    }

    #[cfg(test)]
    pub fn parse_expression(&self, input: &str) -> Result<ast::Expression> {
        let expr = self._expression_parser.parse(input)?;
        Ok(expr)
    }
}
