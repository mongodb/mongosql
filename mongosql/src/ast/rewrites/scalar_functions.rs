use crate::ast::{
    self,
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
    *,
};

/// ScalarFunctionsRewritePass rewrites Scalar Functions included for alterate
/// syntax compatibility reasons (e.g., ODBC) to the correct MongoSQL syntax.
pub struct ScalarFunctionsRewritePass;

const DATE_PART_ERROR: Error = Error::InvalidDatePart(
    "the first argument to a date function must be a date part, which must be one of: \
SQL_TSI_FRAC_SECOND, \
SQL_TSI_SECOND, \
SQL_TSI_MINUTE, \
SQL_TSI_HOUR, \
SQL_TSI_DAY, \
SQL_TSI_DAYOFYEAR, \
SQL_TSI_WEEK, \
SQL_TSI_MONTH, \
SQL_TSI_QUARTER, \
SQL_TSI_YEAR, \
MILLISECOND, \
SECOND, \
MINUTE, \
HOUR, \
DAY, \
DAY_OF_YEAR, \
DAYOFYEAR, \
ISO_WEEKDAY, \
WEEK, \
ISO_WEEK, \
MONTH, \
QUARTER, \
YEAR",
);

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
                FunctionName::DateAdd => {
                    if args.len() == 3 {
                        let date_part = match args.swap_remove(0).into_date_part() {
                            Some(date_part) => date_part,
                            None => {
                                self.error = Some(DATE_PART_ERROR);
                                return node;
                            }
                        };
                        Expression::DateFunction(DateFunctionExpr {
                            function: DateFunctionName::Add,
                            date_part,
                            args: vec![args.swap_remove(1), args.swap_remove(0)],
                        })
                    } else {
                        self.error = Some(Error::IncorrectArgumentCount {
                            name: function.as_str(),
                            required: "3",
                            found: args.len(),
                        });
                        node
                    }
                }
                FunctionName::DateDiff => {
                    if args.len() == 4 {
                        let date_part = match args.swap_remove(0).into_date_part() {
                            Some(date_part) => date_part,
                            None => {
                                self.error = Some(DATE_PART_ERROR);
                                return node;
                            }
                        };
                        Expression::DateFunction(DateFunctionExpr {
                            function: DateFunctionName::Diff,
                            date_part,
                            args: vec![
                                args.swap_remove(1), // input arg 1
                                args.swap_remove(1), // input arg 2
                                args.swap_remove(0), // input arg 3
                            ],
                        })
                    } else if args.len() == 3 {
                        let date_part = match args.swap_remove(0).into_date_part() {
                            Some(date_part) => date_part,
                            None => {
                                self.error = Some(DATE_PART_ERROR);
                                return node;
                            }
                        };
                        Expression::DateFunction(DateFunctionExpr {
                            function: DateFunctionName::Diff,
                            date_part,
                            args: vec![
                                args.swap_remove(1),
                                args.swap_remove(0),
                                Expression::Literal(Literal::String("sunday".to_string())),
                            ],
                        })
                    } else {
                        self.error = Some(Error::IncorrectArgumentCount {
                            name: function.as_str(),
                            required: "3 or 4",
                            found: args.len(),
                        });
                        node
                    }
                }
                FunctionName::DateTrunc => {
                    if args.len() == 3 {
                        let date_part = match args.swap_remove(0).into_date_part() {
                            Some(date_part) => date_part,
                            None => {
                                self.error = Some(DATE_PART_ERROR);
                                return node;
                            }
                        };
                        Expression::DateFunction(DateFunctionExpr {
                            function: DateFunctionName::Trunc,
                            date_part,
                            args: vec![args.swap_remove(1), args.swap_remove(0)],
                        })
                    } else if args.len() == 2 {
                        let date_part = match args.swap_remove(0).into_date_part() {
                            Some(date_part) => date_part,
                            None => {
                                self.error = Some(DATE_PART_ERROR);
                                return node;
                            }
                        };
                        Expression::DateFunction(DateFunctionExpr {
                            function: DateFunctionName::Trunc,
                            date_part,
                            args: vec![
                                args.swap_remove(0),
                                Expression::Literal(Literal::String("sunday".to_string())),
                            ],
                        })
                    } else {
                        self.error = Some(Error::IncorrectArgumentCount {
                            name: function.as_str(),
                            required: "2 or 3",
                            found: args.len(),
                        });
                        node
                    }
                }
                FunctionName::Year
                | FunctionName::Month
                | FunctionName::Week
                | FunctionName::DayOfWeek
                | FunctionName::DayOfMonth
                | FunctionName::DayOfYear
                | FunctionName::Hour
                | FunctionName::Minute
                | FunctionName::Second
                | FunctionName::Millisecond => {
                    if args.len() == 1 {
                        Expression::Extract(ExtractExpr {
                            extract_spec: function.try_into().unwrap(),
                            arg: Box::new(args.swap_remove(0)),
                        })
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
