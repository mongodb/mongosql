use crate::{parser::ast, result::Result};
use lalrpop_util::{lalrpop_mod, lexer::Token};

lalrpop_mod!(
    #[allow(clippy::all)]
    grammar,
    "/parser/mongosql.rs"
);

pub type ParseError<'t> = lalrpop_util::ParseError<usize, Token<'t>, &'static str>;

pub fn parse(input: &str) -> Result<ast::Query> {
    let query = grammar::QueryParser::new().parse(input)?;
    Ok(query)
}
