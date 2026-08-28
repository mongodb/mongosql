use crate::ast::{
    self,
    rewrites::{try_exact_args, try_extract_either_args, ArgCount, Error, Pass, Result},
    visitor::Visitor,
    AccessExpr, ArrayCastExpr, BinaryExpr, BinaryOp, CastExpr, ComparisonOp, Expression,
    FilterExpr, FunctionArgument, FunctionArguments, FunctionExpr, FunctionName,
    HigherOrderFunctionExpr, IsExpr, Literal, MapExpr, NamedFunction, ReduceExpr, SubpathExpr,
    TrimExpr, TrimSpec, Type, TypeOrMissing, UnaryExpr, UnaryOp,
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

        let mut func_arg_visitor = FunctionArgumentVisitor { error: None };
        let query = query.walk(&mut func_arg_visitor);

        if let Some(error) = func_arg_visitor.error {
            return Err(error);
        }

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
                expr: Box::new(this()),
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
                expr: Box::new(this()),
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
                    expr: Box::new(this()),
                    target_type: TypeOrMissing::Type(Type::Null),
                })),
            }),
        ))
    }

    /// Rewrite `ARRAY_REMOVE(a, NULL)` into `FILTER(a, this <> NULL)`.
    /// For any other `x`, null elements must be retained, so rewrite
    /// `ARRAY_REMOVE(a, x)` into `FILTER(a, this IS NULL OR this <> x)`.
    fn rewrite_array_remove(args: &[Expression]) -> Result<Expression> {
        let [array, remove_expr] = try_exact_args("ARRAY_REMOVE", args)?;

        let is_filtering_out_nulls = matches!(remove_expr, Expression::Literal(Literal::Null));
        if is_filtering_out_nulls {
            // If we're explicitly checking for null, then this is fine.
            Ok(Self::make_filter(
                array.clone(),
                Self::make_binary(
                    this(),
                    BinaryOp::Comparison(ComparisonOp::Neq),
                    remove_expr.clone(),
                ),
            ))
        } else {
            // Not explicitly removing nulls, so null elements must be retained:
            // FILTER(a, this IS NULL OR this <> remove_expr)
            let include_null_values = Expression::Is(IsExpr {
                expr: Box::new(this()),
                target_type: TypeOrMissing::Type(Type::Null),
            });

            Ok(Self::make_filter(
                array.clone(),
                Self::make_binary(
                    include_null_values,
                    BinaryOp::Or,
                    Self::make_binary(
                        this(),
                        BinaryOp::Comparison(ComparisonOp::Neq),
                        remove_expr.clone(),
                    ),
                ),
            ))
        }
    }

    /// Rewrite `ARRAY_COUNT_IF(a, f)` into `SIZE(FILTER(a, f))`.
    fn rewrite_array_count_if(args: &[Expression]) -> Result<Expression> {
        let [array, f] = try_exact_args("ARRAY_COUNT_IF", args)?;

        Ok(Self::make_size(Self::make_filter(array.clone(), f.clone())))
    }

    /// Rewrite any single-argument reduce-alias function (e.g., `ARRAY_SUM(a)` into
    /// `REDUCE(a, init_value, value op this)`).
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
            Self::make_binary(value(), op, this()),
        ))
    }

    /// Rewrite `ARRAY_AVG(a)` into `REDUCE(a, 0, value + this) / SIZE(a)`.
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
                Self::make_binary(value(), BinaryOp::Concat, this()),
            ))
        } else {
            Ok(Expression::Trim(TrimExpr {
                trim_spec: TrimSpec::Leading,
                trim_chars: Box::new(sep.clone()),
                arg: Box::new(Self::make_reduce(
                    array.clone(),
                    Expression::StringConstructor("".to_string()),
                    Self::make_binary(
                        Self::make_binary(value(), BinaryOp::Concat, sep.clone()),
                        BinaryOp::Concat,
                        this(),
                    ),
                )),
            }))
        }
    }
}

struct FunctionArgumentVisitor {
    error: Option<Error>,
}

impl Visitor for FunctionArgumentVisitor {
    fn visit_higher_order_function_expr(
        &mut self,
        node: HigherOrderFunctionExpr,
    ) -> HigherOrderFunctionExpr {
        let node = node.walk(self);
        match node {
            HigherOrderFunctionExpr::Map(MapExpr { array, f }) => {
                HigherOrderFunctionExpr::Map(MapExpr {
                    array,
                    f: Box::new(self.rewrite_unary_context_arg(*f)),
                })
            }
            HigherOrderFunctionExpr::Filter(FilterExpr { array, f }) => {
                HigherOrderFunctionExpr::Filter(FilterExpr {
                    array,
                    f: Box::new(self.rewrite_unary_context_arg(*f)),
                })
            }
            HigherOrderFunctionExpr::Reduce(ReduceExpr {
                array,
                init_value,
                f,
            }) => HigherOrderFunctionExpr::Reduce(ReduceExpr {
                array,
                init_value,
                f: Box::new(self.rewrite_binary_context_arg(*f)),
            }),
        }
    }
}

impl FunctionArgumentVisitor {
    /// Rewrites a `FunctionArgument` appearing in a "unary context", i.e. the body of a `MAP` or
    /// `FILTER`, where the function is applied to a single argument, `this`.
    ///
    /// The overloaded binary operators `+` and `-` are rewritten to their unary counterparts.
    /// Any other binary operator is an error, since it requires two arguments; in that case the
    /// argument is returned unchanged (this is because error checking happens after the visitor
    /// returns).
    fn rewrite_unary_context_arg(&mut self, f: FunctionArgument) -> FunctionArgument {
        match f {
            FunctionArgument::Expr(_) => f,
            FunctionArgument::NamedFunction(NamedFunction::UnaryOp(op)) => {
                FunctionArgument::Expr(Self::rewrite_named_function_to_unary_expr(op))
            }
            FunctionArgument::NamedFunction(NamedFunction::BinaryOp(BinaryOp::Add)) => {
                FunctionArgument::Expr(Self::rewrite_named_function_to_unary_expr(UnaryOp::Pos))
            }
            FunctionArgument::NamedFunction(NamedFunction::BinaryOp(BinaryOp::Sub)) => {
                FunctionArgument::Expr(Self::rewrite_named_function_to_unary_expr(UnaryOp::Neg))
            }
            // Exhaustively match other BinaryOps so in the future if we add new binary operators
            // with overloaded forms, we will be forced to handle them here.
            FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::And))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::Concat))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::Div))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::In))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::Mul))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::NotIn))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op @ BinaryOp::Or))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(
                op @ BinaryOp::Comparison(ComparisonOp::Eq),
            ))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(
                op @ BinaryOp::Comparison(ComparisonOp::Gt),
            ))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(
                op @ BinaryOp::Comparison(ComparisonOp::Gte),
            ))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(
                op @ BinaryOp::Comparison(ComparisonOp::Lt),
            ))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(
                op @ BinaryOp::Comparison(ComparisonOp::Lte),
            ))
            | FunctionArgument::NamedFunction(NamedFunction::BinaryOp(
                op @ BinaryOp::Comparison(ComparisonOp::Neq),
            )) => {
                self.error = Some(Error::IncorrectArgumentCount {
                    name: op.as_str(),
                    required: ArgCount::Exactly(2),
                    found: 1,
                });
                f
            }
            FunctionArgument::NamedFunction(NamedFunction::Function(op)) => {
                FunctionArgument::Expr(Self::named_function_to_function_call(op, vec![this()]))
            }
        }
    }

    /// Rewrites a `FunctionArgument` appearing in a "binary context", i.e. the body of a `REDUCE`,
    /// where the function is applied to two arguments, `value` and `this`.
    ///
    /// The overloaded unary operators `+` and `-` are rewritten to their binary counterparts.
    /// Any other unary operator is an error, since it requires a single argument; in that case the
    /// argument is returned unchanged (this is because error checking happens after the visitor
    /// returns).
    fn rewrite_binary_context_arg(&mut self, f: FunctionArgument) -> FunctionArgument {
        match f {
            FunctionArgument::Expr(_) => f,
            FunctionArgument::NamedFunction(NamedFunction::UnaryOp(UnaryOp::Pos)) => {
                FunctionArgument::Expr(Self::rewrite_named_function_to_binary_expr(BinaryOp::Add))
            }
            FunctionArgument::NamedFunction(NamedFunction::UnaryOp(UnaryOp::Neg)) => {
                FunctionArgument::Expr(Self::rewrite_named_function_to_binary_expr(BinaryOp::Sub))
            }
            // Exhaustively match other UnaryOps so in the future if we add new unary operators
            // with overloaded forms, we will be forced to handle them here.
            FunctionArgument::NamedFunction(NamedFunction::UnaryOp(op @ UnaryOp::Not)) => {
                self.error = Some(Error::IncorrectArgumentCount {
                    name: op.as_str(),
                    required: ArgCount::Exactly(1),
                    found: 2,
                });
                f
            }
            FunctionArgument::NamedFunction(NamedFunction::BinaryOp(op)) => {
                FunctionArgument::Expr(Self::rewrite_named_function_to_binary_expr(op))
            }
            FunctionArgument::NamedFunction(NamedFunction::Function(op)) => FunctionArgument::Expr(
                Self::named_function_to_function_call(op, vec![value(), this()]),
            ),
        }
    }

    fn rewrite_named_function_to_unary_expr(op: UnaryOp) -> Expression {
        Expression::Unary(UnaryExpr {
            op,
            expr: Box::new(this()),
        })
    }

    fn rewrite_named_function_to_binary_expr(op: BinaryOp) -> Expression {
        Expression::Binary(BinaryExpr {
            left: Box::new(value()),
            op,
            right: Box::new(this()),
        })
    }

    fn named_function_to_function_call(op: FunctionName, args: Vec<Expression>) -> Expression {
        Expression::Function(FunctionExpr {
            function: op,
            args: FunctionArguments::Args(args),
            set_quantifier: None,
        })
    }
}

/***************************************/
/********** Utility Functions **********/
/***************************************/
/// Returns the `this` identifier expression used within higher order function bodies.
#[inline(always)]
fn this() -> Expression {
    Expression::Identifier(THIS.into())
}

/// Returns the `value` identifier expression used within higher order function bodies.
/// Note that "value" is a reserved keyword in SQL, so we need indicate that it is delimited.
#[inline(always)]
fn value() -> Expression {
    Expression::Identifier((VALUE, true).into())
}

fn prepend_parent_to_field_path_expr(
    parent: &str,
    field_path_expr: &Expression,
) -> Option<Expression> {
    match field_path_expr {
        Expression::Identifier(id) => Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier(parent.into())),
            subpath: id.name.clone(),
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
            expr: Box::new(Expression::Identifier("a".into())),
            subpath: "b".to_string(),
        })),
        input_parent = "a",
        input_field_path_expr = Expression::Identifier("b".into()),
    );

    test_prepend_parent_to_field_path_expr!(
        subpath,
        expected = Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Identifier("a".into())),
                subpath: "b".to_string(),
            })),
            subpath: "c".to_string(),
        })),
        input_parent = "a",
        input_field_path_expr = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier("b".into())),
            subpath: "c".to_string(),
        }),
    );

    test_prepend_parent_to_field_path_expr!(
        deeply_nested_subpath,
        expected = Some(Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Subpath(SubpathExpr {
                    expr: Box::new(Expression::Subpath(SubpathExpr {
                        expr: Box::new(Expression::Identifier("a".into())),
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
                    expr: Box::new(Expression::Identifier("b".into())),
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
