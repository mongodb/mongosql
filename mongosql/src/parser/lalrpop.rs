use crate::{
    ast,
    usererror::{util::generate_suggestion, UserError, UserErrorDisplay},
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
            Error::UnexpectedToken(input, expected) => {
                match generate_suggestion(
                    input,
                    &expected.iter().map(get_token).collect::<Vec<_>>(),
                ) {
                    Ok(translated_expected) => {
                        if translated_expected.len() > 3 || translated_expected.is_empty() {
                            return Some(format!(
                                "Unrecognized token `{input}`, expected: `{}`",
                                expected
                                    .iter()
                                    .map(get_token)
                                    .collect::<Vec<_>>()
                                    .join("`, `")
                            ));
                        }
                        Some(format!(
                            "Unrecognized token `{input}`, did you mean `{}`?",
                            translated_expected.join("`, `")
                        ))
                    }
                    Err(e) => Some(format!("Unrecognized token `{input}`. {e}")),
                }
            }
        }
    }

    fn technical_message(&self) -> String {
        match self {
            Error::Lalrpop(string) => string.clone(),
            Error::UnexpectedToken(t, e) => format!("Unrecognized token: `{t}`, expected: {e:?}"),
        }
    }
}

pub type LalrpopError<'t> = lalrpop_util::ParseError<usize, Token<'t>, String>;

impl From<LalrpopError<'_>> for Error {
    fn from(value: LalrpopError<'_>) -> Self {
        match value {
            lalrpop_util::ParseError::UnrecognizedToken { token, expected } => {
                Self::UnexpectedToken(token.1.to_string(), expected.clone())
            }
            lalrpop_util::ParseError::InvalidToken { location } => {
                Self::Lalrpop(format!("InvalidToken at {}", location))
            }
            lalrpop_util::ParseError::UnrecognizedEOF { location, expected } => {
                Self::Lalrpop(format!(
                    "UnrecognizedEOF at {} with expected {:?}",
                    location, expected
                ))
            }
            lalrpop_util::ParseError::ExtraToken { token } => Self::Lalrpop(format!(
                "ExtraToken at {} with token {:?}",
                token.0, token.1
            )),
            lalrpop_util::ParseError::User { error } => Self::Lalrpop(error),
        }
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

#[allow(dead_code)]
pub fn get_token<T: Into<String>>(input: T) -> String {
    let input = input.into();
    match TOKEN_MAP.get(input.as_str()) {
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
