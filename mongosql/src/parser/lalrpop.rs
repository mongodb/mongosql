use crate::{parser::ast, result::Result};
use lalrpop_util::{lalrpop_mod, lexer::Token};

lalrpop_mod!(
    #[allow(clippy::all)]
    grammar,
    "/parser/mongosql.rs"
);

pub type ParseError<'t> = lalrpop_util::ParseError<usize, Token<'t>, String>;

pub struct Parser {
    query_parser: grammar::QueryParser,
    #[allow(dead_code)]
    expression_parser: grammar::ExpressionParser,
}

impl Parser {
    pub fn new() -> Self {
        let query_parser = grammar::QueryParser::new();
        let expression_parser = grammar::ExpressionParser::new();

        Self {
            query_parser,
            expression_parser,
        }
    }

    pub fn parse_query(&self, input: &str) -> Result<ast::Query> {
        let query = self.query_parser.parse(input)?;
        Ok(query)
    }

    #[cfg(test)]
    pub fn parse_expression(&self, input: &str) -> Result<ast::Expression> {
        let expr = self.expression_parser.parse(input)?;
        Ok(expr)
    }
}
