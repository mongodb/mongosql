use crate::ast::{
    self,
    rewrites::{try_exact_args, try_extract_either_args, Error, Pass, Result},
    visitor::Visitor,
    AccessExpr, ArrayCastExpr, BinaryExpr, BinaryOp, CastExpr, ComparisonOp, Expression,
    FilterExpr, FunctionArgument, FunctionArguments, FunctionExpr, FunctionName,
    HigherOrderFunctionExpr, IsExpr, Literal, MapExpr, ReduceExpr, SubpathExpr, TrimExpr, TrimSpec,
    Type, TypeOrMissing, UnaryExpr, UnaryOp,
};

const THIS: &str = "this";
const VALUE: &str = "value";

pub struct HigherOrderFunctionsRewritePass;

impl Pass for HigherOrderFunctionsRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut func_alias_visitor = HigherOrderFunctionsAliasVisitor { error: None };
        let query = query.walk(&mut func_alias_visitor);

        if let Some(error) = func_alias_visitor.error {
            return Err(error);
        }

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
            Expression::ArrayCast(ArrayCastExpr { ref expr, to }) => {
                let res = Self::rewrite_array_cast(expr, to);
                match res {
                    Ok(expr) => expr,
                    Err(err) => {
                        self.error = Some(err);
                        node
                    }
                }
            }
            Expression::Function(FunctionExpr {
                function,
                args: FunctionArguments::Args(ref args),
                set_quantifier: _,
            }) => {
                let res = match function {
                    FunctionName::ArrayExtract => Self::rewrite_array_extract(args),
                    FunctionName::ArrayCompact => Self::rewrite_array_compact(args),
                    FunctionName::ArrayRemove => Self::rewrite_array_remove(args),
                    FunctionName::ArrayCountIf => Self::rewrite_array_count_if(args),
                    FunctionName::ArraySum => Self::rewrite_single_arg_reduce_alias(
                        function.as_str(),
                        args,
                        Literal::Integer(0),
                        BinaryOp::Add,
                    ),
                    FunctionName::ArrayProduct => Self::rewrite_single_arg_reduce_alias(
                        function.as_str(),
                        args,
                        Literal::Integer(1),
                        BinaryOp::Mul,
                    ),
                    FunctionName::ArrayAvg => Self::rewrite_array_avg(args),
                    FunctionName::ArrayAll => Self::rewrite_single_arg_reduce_alias(
                        function.as_str(),
                        args,
                        Literal::Boolean(true),
                        BinaryOp::And,
                    ),
                    FunctionName::ArrayAny => Self::rewrite_single_arg_reduce_alias(
                        function.as_str(),
                        args,
                        Literal::Boolean(false),
                        BinaryOp::Or,
                    ),
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
    #[inline(always)]
    fn make_map(array: Expression, f: Expression) -> Expression {
        Expression::HigherOrderFunction(HigherOrderFunctionExpr::Map(MapExpr {
            array: Box::new(array),
            f: Box::new(FunctionArgument::Expr(f)),
        }))
    }

    #[inline(always)]
    fn make_filter(array: Expression, f: Expression) -> Expression {
        Expression::HigherOrderFunction(HigherOrderFunctionExpr::Filter(FilterExpr {
            array: Box::new(array),
            f: Box::new(FunctionArgument::Expr(f)),
        }))
    }

    #[inline(always)]
    fn make_reduce(array: Expression, init_value: Expression, f: Expression) -> Expression {
        Expression::HigherOrderFunction(HigherOrderFunctionExpr::Reduce(ReduceExpr {
            array: Box::new(array),
            init_value: Box::new(init_value),
            f: Box::new(FunctionArgument::Expr(f)),
        }))
    }

    /// Returns the `this` identifier expression used within higher order function bodies.
    #[inline(always)]
    fn this() -> Expression {
        Expression::Identifier(THIS.to_string())
    }

    /// Returns the `value` identifier expression used within higher order function bodies.
    #[inline(always)]
    fn value() -> Expression {
        Expression::Identifier(VALUE.to_string())
    }

    #[inline(always)]
    fn make_binary(left: Expression, op: BinaryOp, right: Expression) -> Expression {
        Expression::Binary(BinaryExpr {
            left: Box::new(left),
            op,
            right: Box::new(right),
        })
    }

    #[inline(always)]
    fn make_size(array: Expression) -> Expression {
        Expression::Function(FunctionExpr {
            function: FunctionName::Size,
            args: FunctionArguments::Args(vec![array]),
            set_quantifier: None,
        })
    }

    /// Rewrite `ARRAY_CAST(a, to)` into `MAP(a, CAST(this AS to))`.
    fn rewrite_array_cast(array: &Expression, to: Type) -> Result<Expression> {
        Ok(Self::make_map(
            array.clone(),
            Expression::Cast(CastExpr {
                expr: Box::new(Self::this()),
                to,
                on_null: None,
                on_error: None,
            }),
        ))
    }

    /// Rewrite `ARRAY_EXTRACT(a, expr)` into `MAP(a, this.expr)`.
    fn rewrite_array_extract(args: &[Expression]) -> Result<Expression> {
        let [array, extract_expr] = try_exact_args("ARRAY_EXTRACT", args)?;

        // If the second argument is a field-path-like Expression, we should prepend "this" to it.
        // If it is not, we should wrap it in an AccessExpr with "this" as the base.
        let f = prepend_parent_to_field_path_expr(THIS, extract_expr).unwrap_or_else(|| {
            Expression::Access(AccessExpr {
                expr: Box::new(Self::this()),
                subfield: Box::new(extract_expr.clone()),
            })
        });

        Ok(Self::make_map(array.clone(), f))
    }

    /// Rewrite `ARRAY_COMPACT(a)` into `FILTER(a, NOT this IS NULL)`.
    fn rewrite_array_compact(args: &[Expression]) -> Result<Expression> {
        let [array] = try_exact_args("ARRAY_COMPACT", args)?;

        Ok(Self::make_filter(
            array.clone(),
            Expression::Unary(UnaryExpr {
                op: UnaryOp::Not,
                expr: Box::new(Expression::Is(IsExpr {
                    expr: Box::new(Self::this()),
                    target_type: TypeOrMissing::Type(Type::Null),
                })),
            }),
        ))
    }

    /// Rewrite `ARRAY_REMOVE(a, x)` into `FILTER(a, this <> x)`.
    fn rewrite_array_remove(args: &[Expression]) -> Result<Expression> {
        let [array, remove_expr] = try_exact_args("ARRAY_REMOVE", args)?;

        Ok(Self::make_filter(
            array.clone(),
            Self::make_binary(
                Self::this(),
                BinaryOp::Comparison(ComparisonOp::Neq),
                remove_expr.clone(),
            ),
        ))
    }

    /// Rewrite `ARRAY_COUNT_IF(a, f)` into `SIZE(FILTER(a, f))`.
    fn rewrite_array_count_if(args: &[Expression]) -> Result<Expression> {
        let [array, f] = try_exact_args("ARRAY_COUNT_IF", args)?;

        Ok(Self::make_size(Self::make_filter(array.clone(), f.clone())))
    }

    /// Rewrite any single-argument reduce-alias function (e.g., `ARRAY_SUM(a)` into
    /// `REDUCE(a, init_value, value op this)`.
    fn rewrite_single_arg_reduce_alias(
        name: &'static str,
        args: &[Expression],
        init_value: Literal,
        op: BinaryOp,
    ) -> Result<Expression> {
        let [array] = try_exact_args(name, args)?;

        Ok(Self::make_reduce(
            array.clone(),
            Expression::Literal(init_value),
            Self::make_binary(Self::value(), op, Self::this()),
        ))
    }

    /// Rewrite `ARRAY_AVG(a)` into `REDUCE(a, 0, this + value) / SIZE(a)`.
    fn rewrite_array_avg(args: &[Expression]) -> Result<Expression> {
        let rewritten_sum = Self::rewrite_single_arg_reduce_alias(
            "ARRAY_AVG",
            args,
            Literal::Integer(0),
            BinaryOp::Add,
        )?;
        let array = &args[0];

        Ok(Self::make_binary(
            rewritten_sum,
            BinaryOp::Div,
            Self::make_size(array.clone()),
        ))
    }

    /// Rewrite `ARRAY_JOIN(a)` or `ARRAY_JOIN(a, '')` into `REDUCE(a, '', value || this)`,
    /// and rewrite `ARRAY_JOIN(a, sep)` into
    /// `TRIM(LEADING sep FROM REDUCE(a, '', value || sep || this))`.
    fn rewrite_array_join(args: &[Expression]) -> Result<Expression> {
        let ([array], rest_args) = try_extract_either_args::<1, 2>("ARRAY_JOIN", args)?;

        let sep = if let [sep] = rest_args {
            sep.clone()
        } else {
            Expression::StringConstructor("".to_string())
        };

        if sep == Expression::StringConstructor("".to_string()) {
            Ok(Self::make_reduce(
                array.clone(),
                Expression::StringConstructor("".to_string()),
                Self::make_binary(Self::value(), BinaryOp::Concat, Self::this()),
            ))
        } else {
            Ok(Expression::Trim(TrimExpr {
                trim_spec: TrimSpec::Leading,
                trim_chars: Box::new(sep.clone()),
                arg: Box::new(Self::make_reduce(
                    array.clone(),
                    Expression::StringConstructor("".to_string()),
                    Self::make_binary(
                        Self::make_binary(Self::value(), BinaryOp::Concat, sep.clone()),
                        BinaryOp::Concat,
                        Self::this(),
                    ),
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
