use crate::ast::{
    self,
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
    AccessExpr, BinaryExpr, BinaryOp, ComparisonOp, Expression, FilterExpr, FunctionArgument,
    FunctionArguments, FunctionExpr, FunctionName, HigherOrderFunctionExpr, IsExpr, Literal,
    MapExpr, ReduceExpr, SubpathExpr, TrimExpr, TrimSpec, Type, TypeOrMissing, UnaryExpr, UnaryOp,
};

const THIS: &str = "this";
const VALUE: &str = "value";

pub struct HigherOrderFunctionsRewritePass;

impl Pass for HigherOrderFunctionsRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut func_alias_visitor = HigherOrderFunctionsAliasVisitor { error: None };
        let query = query.walk(&mut func_alias_visitor);

        let mut func_arg_visitor = FunctionArgumentVisitor;
        let query = query.walk(&mut func_arg_visitor);

        Ok(query)
    }
}

struct HigherOrderFunctionsAliasVisitor {
    error: Option<Error>,
}

impl Visitor for HigherOrderFunctionsAliasVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        let node = node.walk(self);
        match node {
            Expression::Function(FunctionExpr {
                function,
                args: FunctionArguments::Args(ref args),
                set_quantifier: _,
            }) => {
                let res = match function {
                    FunctionName::ArrayCast => Self::rewrite_array_cast(args),
                    FunctionName::ArrayExtract => Self::rewrite_array_extract(args),
                    FunctionName::ArrayCompact => Self::rewrite_array_compact(args),
                    FunctionName::ArrayRemove => Self::rewrite_array_remove(args),
                    FunctionName::ArrayCountIf => Self::rewrite_array_count_if(args),
                    FunctionName::ArraySum => Self::rewrite_array_sum(args),
                    FunctionName::ArrayProduct => Self::rewrite_array_product(args),
                    FunctionName::ArrayAverage => Self::rewrite_array_average(args),
                    FunctionName::ArrayAll => Self::rewrite_array_all(args),
                    FunctionName::ArrayAny => Self::rewrite_array_any(args),
                    FunctionName::ArrayJoin => Self::rewrite_array_join(args),
                    _ => return node,
                };

                match res {
                    Ok(expr) => expr,
                    Err(err) => {
                        self.error = Some(err);
                        node
                    }
                }
            }
            _ => node,
        }
    }
}

impl HigherOrderFunctionsAliasVisitor {
    fn make_map(array: Expression, f: Expression) -> Expression {
        Expression::HigherOrderFunction(HigherOrderFunctionExpr::Map(MapExpr {
            array: Box::new(array),
            f: Box::new(FunctionArgument::Expr(f)),
        }))
    }

    fn make_filter(array: Expression, f: Expression) -> Expression {
        Expression::HigherOrderFunction(HigherOrderFunctionExpr::Filter(FilterExpr {
            array: Box::new(array),
            f: Box::new(FunctionArgument::Expr(f)),
        }))
    }

    fn make_reduce(array: Expression, init_value: Expression, f: Expression) -> Expression {
        Expression::HigherOrderFunction(HigherOrderFunctionExpr::Reduce(ReduceExpr {
            array: Box::new(array),
            init_value: Box::new(init_value),
            f: Box::new(FunctionArgument::Expr(f)),
        }))
    }

    fn rewrite_array_cast(args: &[Expression]) -> Result<Expression> {
        // TODO: need to handle array_cast specially since it takes a Type as an argument
        todo!()
    }

    /// Rewrite `ARRAY_EXTRACT(a, expr)` into `MAP(a, this.expr)`.
    fn rewrite_array_extract(args: &[Expression]) -> Result<Expression> {
        if args.len() != 2 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_EXTRACT",
                required: "2",
                found: args.len(),
            });
        }
        let array = &args[0];
        let extract_expr = &args[1];

        // If the second argument is a field-path-like Expression, we should prepend "this" to it.
        // If it is not, we should wrap it in an AccessExpr with "this" as the base.
        let f = prepend_parent_to_field_path_expr(THIS, extract_expr).unwrap_or_else(|| {
            Expression::Access(AccessExpr {
                expr: Box::new(Expression::Identifier(THIS.to_string())),
                subfield: Box::new(extract_expr.clone()),
            })
        });

        Ok(Self::make_map(array.clone(), f))
    }

    /// Rewrite `ARRAY_COMPACT(a)` into `FILTER(a, NOT this IS NULL)`.
    fn rewrite_array_compact(args: &[Expression]) -> Result<Expression> {
        if args.len() != 1 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_COMPACT",
                required: "1",
                found: args.len(),
            });
        }

        let array = &args[0];
        Ok(Self::make_filter(
            array.clone(),
            Expression::Unary(UnaryExpr {
                op: UnaryOp::Not,
                expr: Box::new(Expression::Is(IsExpr {
                    expr: Box::new(Expression::Identifier(THIS.to_string())),
                    target_type: TypeOrMissing::Type(Type::Null),
                })),
            }),
        ))
    }

    /// Rewrite `ARRAY_REMOVE(a, x)` into `FILTER(a, this <> x)`.
    fn rewrite_array_remove(args: &[Expression]) -> Result<Expression> {
        if args.len() != 2 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_REMOVE",
                required: "2",
                found: args.len(),
            });
        }

        let array = &args[0];
        let x = &args[1];
        Ok(Self::make_filter(
            array.clone(),
            Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier(THIS.to_string())),
                op: BinaryOp::Comparison(ComparisonOp::Neq),
                right: Box::new(x.clone()),
            }),
        ))
    }

    /// Rewrite `ARRAY_COUNT_IF(a, f)` into `SIZE(FILTER(a, f))`.
    fn rewrite_array_count_if(args: &[Expression]) -> Result<Expression> {
        if args.len() != 2 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_COUNT_IF",
                required: "2",
                found: args.len(),
            });
        }

        let array = &args[0];
        let f = &args[1];
        Ok(Expression::Function(FunctionExpr {
            function: FunctionName::Size,
            args: FunctionArguments::Args(vec![Self::make_filter(array.clone(), f.clone())]),
            set_quantifier: None,
        }))
    }

    /// Rewrite `ARRAY_SUM(a)` into `REDUCE(a, 0, this + value)`.
    fn rewrite_array_sum(args: &[Expression]) -> Result<Expression> {
        if args.len() != 1 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_SUM",
                required: "1",
                found: args.len(),
            });
        }

        let array = &args[0];
        Ok(Self::make_reduce(
            array.clone(),
            Expression::Literal(Literal::Integer(0)),
            Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier(THIS.to_string())),
                op: BinaryOp::Add,
                right: Box::new(Expression::Identifier(VALUE.to_string())),
            }),
        ))
    }

    /// Rewrite `ARRAY_PRODUCT(a)` into `REDUCE(a, 1, this * value)`.
    fn rewrite_array_product(args: &[Expression]) -> Result<Expression> {
        if args.len() != 1 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_PRODUCT",
                required: "1",
                found: args.len(),
            });
        }

        let array = &args[0];
        Ok(Self::make_reduce(
            array.clone(),
            Expression::Literal(Literal::Integer(1)),
            Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier(THIS.to_string())),
                op: BinaryOp::Mul,
                right: Box::new(Expression::Identifier(VALUE.to_string())),
            }),
        ))
    }

    /// Rewrite `ARRAY_AVERAGE(a)` into `REDUCE(a, 0, this + value) / SIZE(a)`.
    fn rewrite_array_average(args: &[Expression]) -> Result<Expression> {
        if args.len() != 1 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_AVERAGE",
                required: "1",
                found: args.len(),
            });
        }

        let rewritten_sum = Self::rewrite_array_sum(args)?;
        let array = &args[0];
        Ok(Expression::Binary(BinaryExpr {
            left: Box::new(rewritten_sum),
            op: BinaryOp::Div,
            right: Box::new(Expression::Function(FunctionExpr {
                function: FunctionName::Size,
                args: FunctionArguments::Args(vec![array.clone()]),
                set_quantifier: None,
            })),
        }))
    }

    /// Rewrite `ARRAY_ALL(a)` into `REDUCE(a, true, value AND this)`.
    fn rewrite_array_all(args: &[Expression]) -> Result<Expression> {
        if args.len() != 1 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_ALL",
                required: "1",
                found: args.len(),
            });
        }

        let array = &args[0];

        Ok(Self::make_reduce(
            array.clone(),
            Expression::Literal(Literal::Boolean(true)),
            Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier(VALUE.to_string())),
                op: BinaryOp::And,
                right: Box::new(Expression::Identifier(THIS.to_string())),
            }),
        ))
    }

    /// Rewrite `ARRAY_ANY(a)` into `REDUCE(a, false, value OR this)`.
    fn rewrite_array_any(args: &[Expression]) -> Result<Expression> {
        if args.len() != 1 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_ANY",
                required: "1",
                found: args.len(),
            });
        }

        let array = &args[0];

        Ok(Self::make_reduce(
            array.clone(),
            Expression::Literal(Literal::Boolean(false)),
            Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier(VALUE.to_string())),
                op: BinaryOp::Or,
                right: Box::new(Expression::Identifier(THIS.to_string())),
            }),
        ))
    }

    /// Rewrite `ARRAY_JOIN(a)` or `ARRAY_JOIN(a, '')` into `REDUCE(a, '', value || this)`,
    /// and rewrite `ARRAY_JOIN(a, sep)` into
    /// `TRIM(LEADING sep FROM REDUCE(a, '', value || sep || this))`.
    fn rewrite_array_join(args: &[Expression]) -> Result<Expression> {
        if args.len() < 1 || args.len() > 2 {
            return Err(Error::IncorrectArgumentCount {
                name: "ARRAY_JOIN",
                required: "1 or 2",
                found: args.len(),
            });
        }

        let array = &args[0];
        let mut sep = Expression::StringConstructor("".to_string());
        if args.len() == 2 {
            sep = args[1].clone();
        }

        if sep == Expression::StringConstructor("".to_string()) {
            Ok(Self::make_reduce(
                array.clone(),
                Expression::StringConstructor("".to_string()),
                Expression::Binary(BinaryExpr {
                    left: Box::new(Expression::Identifier(VALUE.to_string())),
                    op: BinaryOp::Concat,
                    right: Box::new(Expression::Identifier(THIS.to_string())),
                }),
            ))
        } else {
            Ok(Expression::Trim(TrimExpr {
                trim_spec: TrimSpec::Leading,
                trim_chars: Box::new(sep.clone()),
                arg: Box::new(Self::make_reduce(
                    array.clone(),
                    Expression::StringConstructor("".to_string()),
                    Expression::Binary(BinaryExpr {
                        left: Box::new(Expression::Binary(BinaryExpr {
                            left: Box::new(Expression::Identifier(VALUE.to_string())),
                            op: BinaryOp::Concat,
                            right: Box::new(sep.clone()),
                        })),
                        op: BinaryOp::Concat,
                        right: Box::new(Expression::Identifier(THIS.to_string())),
                    }),
                )),
            }))
        }
    }
}

struct FunctionArgumentVisitor;

// SQL-3296: Implement FunctionArgumentVisitor
impl Visitor for FunctionArgumentVisitor {}

fn prepend_parent_to_field_path_expr(
    parent: &str,
    field_path_expr: &Expression,
) -> Option<Expression> {
    match field_path_expr {
        Expression::Identifier(id) => Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier(parent.to_string())),
            subpath: id.clone(),
        })),
        Expression::Subpath(expr) => Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(prepend_parent_to_field_path_expr(
                parent,
                expr.expr.as_ref(),
            )?),
            subpath: expr.subpath.clone(),
        })),
        _ => None,
    }
}

#[cfg(test)]
mod prepend_parent_to_field_path_expr_tests {
    use super::*;

    macro_rules! test_prepend_parent_to_field_path_expr {
        ($func_name:ident, expected = $expected:expr, input_parent = $input_parent:expr, input_field_path_expr = $input_field_path_expr:expr,) => {
            #[test]
            fn $func_name() {
                let expected = $expected;
                let input_parent = $input_parent;
                let input_field_path_expr = $input_field_path_expr;

                let actual =
                    prepend_parent_to_field_path_expr(input_parent, &input_field_path_expr);

                assert_eq!(expected, actual);
            }
        };
    }

    test_prepend_parent_to_field_path_expr!(
        identifier,
        expected = Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier("a".to_string())),
            subpath: "b".to_string(),
        })),
        input_parent = "a",
        input_field_path_expr = Expression::Identifier("b".to_string()),
    );

    test_prepend_parent_to_field_path_expr!(
        subpath,
        expected = Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Identifier("a".to_string())),
                subpath: "b".to_string(),
            })),
            subpath: "c".to_string(),
        })),
        input_parent = "a",
        input_field_path_expr = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier("b".to_string())),
            subpath: "c".to_string(),
        }),
    );

    test_prepend_parent_to_field_path_expr!(
        deeply_nested_subpath,
        expected = Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Subpath(SubpathExpr {
                    expr: Box::new(Expression::Subpath(SubpathExpr {
                        expr: Box::new(Expression::Identifier("a".to_string())),
                        subpath: "b".to_string(),
                    })),
                    subpath: "c".to_string(),
                })),
                subpath: "d".to_string(),
            })),
            subpath: "e".to_string(),
        })),
        input_parent = "a",
        input_field_path_expr = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Subpath(SubpathExpr {
                    expr: Box::new(Expression::Identifier("b".to_string())),
                    subpath: "c".to_string(),
                })),
                subpath: "d".to_string(),
            })),
            subpath: "e".to_string(),
        }),
    );

    test_prepend_parent_to_field_path_expr!(
        other,
        expected = None,
        input_parent = "a",
        input_field_path_expr = Expression::Literal(Literal::Integer(1)),
    );
}
