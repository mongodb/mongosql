use crate::ast::{
    self,
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
    *,
};

/// ScalarFunctionsRewritePass rewrites Scalar Functions included for alterate
/// syntax compatibility reasons (e.g., ODBC) to the correct MongoSQL syntax.
pub struct ScalarFunctionsRewritePass;

impl Pass for ScalarFunctionsRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = ScalarFunctionsVisitor { error: None };
        let ret = query.walk(&mut visitor);
        if let Some(error) = visitor.error {
            Err(error)
        } else {
            Ok(ret)
        }
    }
}

/// The visitor that performs the rewrites for the `ScalarFunctionsRewritePass`.
#[derive(Default)]
struct ScalarFunctionsVisitor {
    pub error: Option<Error>,
}

impl ScalarFunctionsVisitor {
    fn gen_log(args: &mut Vec<Expression>, base: f64) -> Expression {
        let args = FunctionArguments::Args(vec![
            args.swap_remove(0),
            Expression::Literal(Literal::Double(base)),
        ]);
        Expression::Function(FunctionExpr {
            function: FunctionName::Log,
            args,
            set_quantifier: None,
        })
    }
}

impl Visitor for ScalarFunctionsVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        let mut node = node.walk(self);
        match node {
            Expression::Function(FunctionExpr {
                function,
                args: FunctionArguments::Args(ref mut args),
                ..
            }) => match function {
                FunctionName::RTrim | FunctionName::LTrim => {
                    let arg = if args.len() == 1 {
                        Box::new(args.swap_remove(0))
                    } else {
                        self.error = Some(Error::IncorrectArgumentCount {
                            name: function.as_str(),
                            required: "1",
                            found: args.len(),
                        });
                        return node;
                    };
                    Expression::Trim(TrimExpr {
                        trim_spec: function.try_into().unwrap(),
                        arg,
                        trim_chars: Box::new(Expression::Literal(Literal::String(" ".to_string()))),
                    })
                }
                FunctionName::Log => match args.len() {
                    1 => ScalarFunctionsVisitor::gen_log(args, std::f64::consts::E),
                    2 => node,
                    _ => {
                        self.error = Some(Error::IncorrectArgumentCount {
                            name: function.as_str(),
                            required: "1 or 2",
                            found: args.len(),
                        });
                        node
                    }
                },
                FunctionName::Log10 => {
                    if args.len() == 1 {
                        ScalarFunctionsVisitor::gen_log(args, 10.0)
                    } else {
                        self.error = Some(Error::IncorrectArgumentCount {
                            name: function.as_str(),
                            required: "1",
                            found: args.len(),
                        });
                        node
                    }
                }
                _ => node,
            },
            _ => node,
        }
    }
}
