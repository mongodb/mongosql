use crate::{ast::*, parser::lalrpop::LalrpopError};
use std::str::FromStr;

/// process_delimited_ident removes the outer delimiters from an identifier and
/// processes any escaped delimiters.
pub(crate) fn process_delimited_ident(value: &str) -> String {
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

pub(crate) fn parse_between_expr(
    e1: Box<Expression>,
    e2: Expression,
    not: bool,
) -> Result<Box<Expression>, LalrpopError<'static>> {
    match e2 {
        Expression::Binary(BinaryExpr { left, op, right }) => {
            if op != BinaryOp::And {
                Err(LalrpopError::from(
                    "invalid BinaryOp in BetweenExpr".to_string(),
                ))
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
        _ => Err(LalrpopError::from(
            "failed to parse BetweenExpr".to_string(),
        )),
    }
}

pub(crate) fn parse_position_func(e: Expression) -> Result<FunctionExpr, LalrpopError<'static>> {
    match e {
        Expression::Binary(BinaryExpr { left, op, right }) => {
            if op != BinaryOp::In {
                Err(LalrpopError::from(
                    "invalid BinaryOp in call to Position()".to_string(),
                ))
            } else {
                Ok(FunctionExpr {
                    function: FunctionName::Position,
                    args: vec![FunctionArg::Expr(*left), FunctionArg::Expr(*right)],
                    set_quantifier: None,
                })
            }
        }
        _ => Err(LalrpopError::from("failed to parse Position()".to_string())),
    }
}

pub(crate) fn parse_sort_key(e: Expression) -> Result<SortKey, LalrpopError<'static>> {
    match e {
        Expression::Identifier(_) => Ok(SortKey::Simple(e)),
        Expression::Subpath(SubpathExpr {
            expr: _,
            subpath: _,
        }) => Ok(SortKey::Simple(e)),
        Expression::Literal(Literal::Integer(i)) => {
            let u: Result<u32, LalrpopError> = u32::from_str(i.to_string().as_str())
                .map_err(|_| LalrpopError::from("failed to convert number to u32".to_string()));
            match u {
                Ok(x) => Ok(SortKey::Positional(x)),
                Err(x) => Err(x),
            }
        }
        _ => Err(LalrpopError::from(
            "failed to parse ORDER BY sort key".to_string(),
        )),
    }
}

pub(crate) fn parse_simple_datasource(
    ae: AliasedExpr,
) -> Result<Datasource, LalrpopError<'static>> {
    match ae {
        AliasedExpr {
            expr: Expression::Identifier(collection),
            alias,
        } => Ok(Datasource::Collection(CollectionSource {
            database: None,
            collection,
            alias,
        })),
        AliasedExpr {
            expr: Expression::Array(array),
            alias: Some(a),
        } => Ok(Datasource::Array(ArraySource { array, alias: a })),
        AliasedExpr {
            expr: Expression::Array(_),
            alias: None,
        } => Err(LalrpopError::from(
            "array datasources must have aliases".to_string(),
        )),
        AliasedExpr {
            expr: Expression::Subquery(query),
            alias: Some(a),
        } => Ok(Datasource::Derived(DerivedSource { query, alias: a })),
        AliasedExpr {
            expr: Expression::Subquery(_),
            alias: None,
        } => Err(LalrpopError::from(
            "derived query datasources must have aliases".to_string(),
        )),
        AliasedExpr {
            expr:
                Expression::Subpath(SubpathExpr {
                    expr: possible_db,
                    subpath: collection,
                }),
            alias,
        } if (*possible_db).is_identifier() => Ok(Datasource::Collection(CollectionSource {
            database: Some(possible_db.take_identifier_name().unwrap()),
            collection,
            alias,
        })),
        AliasedExpr {
            expr: Expression::Subpath(_),
            alias: _,
        } => Err(LalrpopError::from(format!(
            "collection data sources can only have database qualification, found: {}",
            ae.expr,
        ))),
        _ => Err(LalrpopError::from(format!(
            "found unsupported expression used as datasource: {}",
            ae.expr,
        ))),
    }
}

impl Expression {
    pub(crate) fn is_identifier(&self) -> bool {
        matches!(self, Expression::Identifier(_))
    }

    pub(crate) fn take_identifier_name(self) -> Option<String> {
        match self {
            Expression::Identifier(s) => Some(s),
            _ => None,
        }
    }
}
