use crate::{
    ast::{pretty_print::PrettyPrint, *},
    parser::lalrpop::LalrpopError,
};
use std::str::FromStr;

/// process_delimited_ident removes the outer delimiters from an identifier and
/// processes any escaped delimiters.
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

pub fn parse_position_func(e: Expression) -> Result<FunctionExpr, LalrpopError<'static>> {
    match e {
        Expression::Binary(BinaryExpr { left, op, right }) => {
            if op != BinaryOp::In {
                Err(LalrpopError::from(
                    "invalid BinaryOp in call to Position()".to_string(),
                ))
            } else {
                Ok(FunctionExpr {
                    function: FunctionName::Position,
                    args: FunctionArguments::Args(vec![*left, *right]),
                    set_quantifier: None,
                })
            }
        }
        _ => Err(LalrpopError::from("failed to parse Position()".to_string())),
    }
}

pub fn parse_sort_key(e: Expression) -> Result<SortKey, LalrpopError<'static>> {
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

pub fn parse_unwind_path(e: Expression) -> Result<UnwindOption, LalrpopError<'static>> {
    match e {
        Expression::Identifier(_) => Ok(UnwindOption::Path(e)),
        Expression::Subpath(_) => Ok(UnwindOption::Path(e)),
        _ => Err(LalrpopError::from(
            "UNWIND PATH option must be an identifier".to_string(),
        )),
    }
}

pub fn parse_simple_datasource(
    ae: OptionallyAliasedExpr,
) -> Result<Datasource, LalrpopError<'static>> {
    let (expr, alias) = ae.take_fields();
    match expr {
        Expression::Identifier(collection) => Ok(Datasource::Collection(CollectionSource {
            database: None,
            collection,
            alias,
        })),
        Expression::Array(array) => alias.map_or(
            Err(LalrpopError::from(
                "array datasources must have aliases".to_string(),
            )),
            |alias| Ok(Datasource::Array(ArraySource { array, alias })),
        ),
        Expression::Subquery(query) => alias.map_or(
            Err(LalrpopError::from(
                "derived query datasources must have aliases".to_string(),
            )),
            |alias| Ok(Datasource::Derived(DerivedSource { query, alias })),
        ),
        Expression::Subpath(SubpathExpr {
            expr: possible_db,
            subpath: collection,
        }) if (*possible_db).is_identifier() => Ok(Datasource::Collection(CollectionSource {
            database: Some(possible_db.take_identifier_name().unwrap()),
            collection,
            alias,
        })),
        Expression::Subpath(_) => Err(LalrpopError::from(format!(
            "collection datasources can only have database qualification, found: {}",
            expr.pretty_print().unwrap(),
        ))),
        _ => Err(LalrpopError::from(format!(
            "found unsupported expression used as datasource: {}",
            expr.pretty_print().unwrap(),
        ))),
    }
}

impl Expression {
    pub fn is_identifier(&self) -> bool {
        matches!(self, Expression::Identifier(_))
    }

    pub fn take_identifier_name(self) -> Option<String> {
        match self {
            Expression::Identifier(s) => Some(s),
            _ => None,
        }
    }
}

pub fn parse_like_expr(
    expr: Box<Expression>,
    pattern: Box<Expression>,
    escape: Option<String>,
) -> Result<Box<Expression>, LalrpopError<'static>> {
    if escape
        .clone()
        // String::len is in bytes, not characters!
        .map_or(1, |s| s.chars().count())
        == 1
    {
        Ok(Box::new(Expression::Like(LikeExpr {
            expr,
            pattern,
            escape,
        })))
    } else {
        Err(LalrpopError::from(
            "Escape character must be a string of length 1.".to_string(),
        ))
    }
}
