use crate::{
    ast,
    usererror::{UserError, UserErrorDisplay},
};
use lalrpop_util::{lalrpop_mod, lexer::Token};
use lazy_static::lazy_static;
use std::collections::HashMap;

lalrpop_mod!(
    #[allow(clippy::all)]
    grammar,
    "/parser/mongosql.rs"
);

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, UserErrorDisplay, PartialEq, Eq)]
pub enum Error {
    Lalrpop(String),
    UnexpectedToken(String, Vec<String>),
}

impl UserError for Error {
    fn code(&self) -> u32 {
        match self {
            Error::Lalrpop(_) => 2000,
            Error::UnexpectedToken(_, _) => 2001,
        }
    }

    fn user_message(&self) -> Option<String> {
        match self {
            Error::Lalrpop(_) => None,
            Error::UnexpectedToken(_, _) => None,
        }
    }

    fn technical_message(&self) -> String {
        match self {
            Error::Lalrpop(string) => string.clone(),
            Error::UnexpectedToken(_, _) => todo!(),
        }
    }
}

pub type LalrpopError<'t> = lalrpop_util::ParseError<usize, Token<'t>, String>;

impl From<LalrpopError<'_>> for Error {
    fn from(e: LalrpopError<'_>) -> Self {
        Error::Lalrpop(format!("{e}"))
    }
}

lazy_static! {
    static ref QUERY_PARSER: grammar::QueryParser = grammar::QueryParser::new();
    static ref EXPRESSION_PARSER: grammar::ExpressionParser = grammar::ExpressionParser::new();
    static ref TOKEN_MAP: HashMap<&'static str, &'static str> = HashMap::from([
        ("ADD", "+"),
        ("ARROW", "=>"),
        ("COMMA", ","),
        ("COLON", ":"),
        ("CONCAT", "||"),
        ("DELIMITED_IDENT_BACKTICK", "`"),
        ("DELIMITED_IDENT_QUOTE", "\""),
        ("DIV", "/"),
        ("DOT", "."),
        ("DOT_STAR", ".*"),
        ("DOUBLE_COLON", "::"),
        ("EQ", "="),
        ("FETCH_FIRST", "FETCH FIRST"),
        ("GT", ">"),
        ("GTE", ">="),
        ("LEFT_BRACKET", "["),
        ("LEFT_PAREN", "("),
        ("LEFT_CURLY_BRACE", "{"),
        ("LT", "<"),
        ("LTE", "<="),
        ("NEQ", "<>"),
        ("NOT_IN", "NOT IN"),
        ("NOT_LIKE", "NOT LIKE"),
        ("RIGHT_BRACKET", "]"),
        ("RIGHT_CURLY_BRACE", "}"),
        ("RIGHT_PAREN", ")"),
        ("ROWS_ONLY", "ROWS ONLY"),
        ("STAR", "*"),
        ("SUB", "-"),
        ("TYPE_ASSERTION", "::!"),
    ]);
}

// remove `#[allow(dead_code)]` during SQL-1605
#[allow(dead_code)]
pub fn get_token(input: &str) -> String {
    match TOKEN_MAP.get(input) {
        Some(token) => (*token).to_string(),
        None => input.to_string(),
    }
}

pub fn parse_query(input: &str) -> Result<ast::Query> {
    Ok(QUERY_PARSER.parse(input)?)
}

#[cfg(test)]
pub fn parse_expression(input: &str) -> Result<ast::Expression> {
    let expr = EXPRESSION_PARSER.parse(input)?;
    Ok(expr)
}
