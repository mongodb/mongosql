use crate::parser::ast::*;
use lalrpop_util::lexer::Token;
use lalrpop_util::ParseError;

// process_delimited_ident removes the outer delimiters from an identifier and
// processes any escaped delimiters.
pub fn process_delimited_ident(value: &str) -> String {
    let delimiter = value.chars().next().unwrap_or_default();
    let trimmed_value = value[1..value.len() - 1].to_string();
    if delimiter == '\"' {
        trimmed_value.replace("\"\"", "\"")
    } else if delimiter == '`' {
        trimmed_value.replace("``", "`")
    } else {
        unreachable!("delimiters other than double-quote and backtick are not supported")
    }
}

pub fn parse_between_expr(
    e1: Box<Expression>,
    e2: Expression,
    not: bool,
) -> Result<Box<Expression>, ParseError<usize, Token<'static>, &'static str>> {
    match e2 {
        Expression::Binary(BinaryExpr { left, op, right }) => {
            if op != BinaryOp::And {
                Err(ParseError::from("invalid BinaryOp in BetweenExpr"))
            } else if not {
                Ok(Box::new(Expression::Unary(UnaryExpr {
                    op: UnaryOp::Not,
                    expr: Box::new(Expression::Between(BetweenExpr {
                        expr: e1,
                        min: left,
                        max: right,
                    })),
                })))
            } else {
                Ok(Box::new(Expression::Between(BetweenExpr {
                    expr: e1,
                    min: left,
                    max: right,
                })))
            }
        }
        _ => Err(ParseError::from("failed to parse BetweenExpr")),
    }
}

pub fn parse_position_func(
    e: Expression,
) -> Result<FunctionExpr, ParseError<usize, Token<'static>, &'static str>> {
    match e {
        Expression::Binary(BinaryExpr { left, op, right }) => {
            if op != BinaryOp::In {
                Err(ParseError::from("invalid BinaryOp in call to Position()"))
            } else {
                Ok(FunctionExpr {
                    function: FunctionName("POSITION".to_string()),
                    args: vec![FunctionArg::Expr(*left), FunctionArg::Expr(*right)],
                })
            }
        }
        _ => Err(ParseError::from("failed to parse Position()")),
    }
}
