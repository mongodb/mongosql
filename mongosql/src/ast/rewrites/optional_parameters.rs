use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};

/// This pass combines a variety of individual rewrite visitors that all share
/// the common theme of rewriting ASTs into a canonical form when there are
/// equivalent implicit and explicit representations of the same query fragment.
/// See the docstrings for each individual visitor below for more details.
pub struct OptionalParameterRewritePass;

impl Pass for OptionalParameterRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let query = query.walk(&mut FlattenOptionVisitor);
        let query = query.walk(&mut UnwindOptionVisitor);
        let query = query.walk(&mut CaseElseVisitor);
        let query = query.walk(&mut FunctionVisitor);
        Ok(query)
    }
}

/// This visitor rewrites explicit SEPARATOR => '_' options in flatten
/// datasources into their implicit form by removing them, provided that the
/// FlattenSource has exactly one SEPARATOR option specified. Duplicate
/// SEPARATOR options are left as-is so that the algebrizer can continue to
/// error on them.
struct FlattenOptionVisitor;
impl Visitor for FlattenOptionVisitor {
    fn visit_flatten_source(&mut self, node: ast::FlattenSource) -> ast::FlattenSource {
        let num_separator_opts = node.options.iter().fold(0, |acc, opt| match opt {
            ast::FlattenOption::Separator(_) => acc + 1,
            _ => acc,
        });

        // short circuit if we have duplicate separator opts
        if num_separator_opts > 1 {
            return node;
        }

        let options = node
            .options
            .into_iter()
            .filter(|opt| !matches!(opt, ast::FlattenOption::Separator(sep) if sep == "_"))
            .collect();

        ast::FlattenSource { options, ..node }
    }
}

/// This visitor rewrites explicit OUTER => false options in unwind datasources
/// into their implicit form by removing them, provided that the UnwindSource
/// has exactly one OUTER option specified. Duplicate OUTER options are left
/// as-is so that the algebrizer can continue to error on them.
struct UnwindOptionVisitor;
impl Visitor for UnwindOptionVisitor {
    fn visit_unwind_source(&mut self, node: ast::UnwindSource) -> ast::UnwindSource {
        let num_outer_opts = node.options.iter().fold(0, |acc, opt| match opt {
            ast::UnwindOption::Outer(_) => acc + 1,
            _ => acc,
        });

        // short circuit if we have duplicate outer opts
        if num_outer_opts > 1 {
            return node;
        }

        let options = node
            .options
            .into_iter()
            .filter(|opt| !matches!(opt, ast::UnwindOption::Outer(false)))
            .collect();

        ast::UnwindSource { options, ..node }
    }
}

/// This visitor rewrites implicit ELSE NULLs in case exprs into explicit ELSE
/// NULLs.
struct CaseElseVisitor;
impl Visitor for CaseElseVisitor {
    fn visit_case_expr(&mut self, node: ast::CaseExpr) -> ast::CaseExpr {
        let else_branch = Some(
            node.else_branch
                .unwrap_or_else(|| Box::new(ast::Expression::Literal(ast::Literal::Null))),
        );
        ast::CaseExpr {
            else_branch,
            ..node
        }
    }
}

/// This visitor rewrites explicit default values for functions when not specified.
/// For SUBSTRING, a default third argument of -1 is added when only two arguments are present.
/// For CURRENT_TIMESTAMP, the default precision value of 6 is set when no argument is present.
struct FunctionVisitor;
impl Visitor for FunctionVisitor {
    fn visit_function_expr(&mut self, node: ast::FunctionExpr) -> ast::FunctionExpr {
        match node.function {
            ast::FunctionName::Substring => match node.args {
                ast::FunctionArguments::Star => unreachable!(),
                ast::FunctionArguments::Args(mut ve) => {
                    let arguments = match ve.len() {
                        2 => {
                            ve.push(ast::Expression::Literal(ast::Literal::Integer(-1)));
                            ve
                        }
                        _ => ve,
                    };
                    ast::FunctionExpr {
                        args: ast::FunctionArguments::Args(arguments),
                        ..node
                    }
                }
            },
            ast::FunctionName::CurrentTimestamp => match node.args {
                ast::FunctionArguments::Star => unreachable!(),
                ast::FunctionArguments::Args(ve) => {
                    let precision = if ve.is_empty() {
                        vec![ast::Expression::Literal(ast::Literal::Integer(6))]
                    } else {
                        ve
                    };
                    ast::FunctionExpr {
                        args: ast::FunctionArguments::Args(precision),
                        ..node
                    }
                }
            },
            _ => node,
        }
    }
}
