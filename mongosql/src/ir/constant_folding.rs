use crate::{
    ir::{definitions::*, schema::SchemaInferenceState, visitor::Visitor},
    schema::{Atomic, Satisfaction, Schema, SchemaEnvironment},
    set,
};
use lazy_static::lazy_static;

#[derive(Default)]
struct ConstantFoldExprVisitor;

lazy_static! {
    static ref EMPTY_STATE: SchemaInferenceState = SchemaInferenceState {
        scope_level: 0u16,
        env: SchemaEnvironment::default(),
    };
}

impl ConstantFoldExprVisitor {
    // Constant folds boolean functions
    fn fold_logical_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        let (nullish, non_nullish): (Vec<Expression>, Vec<Expression>) =
            sf.args.clone().into_iter().partition(|e| {
                e.schema(&EMPTY_STATE)
                    .unwrap_or(Schema::Any)
                    .satisfies(&Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::Atomic(Atomic::Null),
                    ]))
                    == Satisfaction::Must
            });
        let has_null = !nullish.is_empty();
        let (fold_init, op): (bool, Box<dyn Fn(bool, bool) -> bool>) = match sf.function {
            ScalarFunction::And => (true, Box::new(|acc, x| x && acc)),
            ScalarFunction::Or => (false, Box::new(|acc, x| x || acc)),
            _ => unreachable!("fold logical functions is only called on And and Or"),
        };
        let mut non_literals = Vec::<Expression>::new();
        let folded_constant = non_nullish
            .into_iter()
            .fold(fold_init, |acc, expr| match expr {
                Expression::Literal(Literal::Boolean(val)) => op(acc, val),
                expr => {
                    non_literals.push(expr);
                    acc
                }
            });
        let folded_expr = Expression::Literal(Literal::Boolean(folded_constant));
        if non_literals.is_empty() && !has_null {
            return folded_expr;
        }
        match sf.function {
            ScalarFunction::And => {
                if !folded_constant {
                    return Expression::Literal(Literal::Boolean(false));
                }
            }
            ScalarFunction::Or => {
                if folded_constant {
                    return Expression::Literal(Literal::Boolean(true));
                }
            }
            _ => unreachable!("fold logical functions is only called on And and Or"),
        };
        let args = if has_null {
            [vec![Expression::Literal(Literal::Null)], non_literals].concat()
        } else {
            [vec![folded_expr], non_literals].concat()
        };
        if args.len() == 1 {
            return args[0].clone();
        }
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: sf.function,
            args,
        })
    }

    // This is not a general purpose function and is not capable of checking equality of very
    // large longs. It is used to check special arithmetic edge cases like 0 and 1.
    fn numeric_eq(expr: &Expression, num: f64) -> bool {
        match expr {
            Expression::Literal(Literal::Integer(val)) => *val == num as i32,
            Expression::Literal(Literal::Long(val)) => *val == num as i64,
            Expression::Literal(Literal::Double(val)) => *val == num,
            _ => false,
        }
    }

    // Constant folds constants of the same type within an associative arithmetic function
    fn fold_associative_arithmetic_function(
        &mut self,
        sf: ScalarFunctionApplication,
    ) -> Expression {
        for expr in &sf.args {
            match expr.schema(&EMPTY_STATE) {
                Err(_) => break,
                Ok(sch) => {
                    if sch.satisfies(&Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::Atomic(Atomic::Null),
                    ])) == Satisfaction::Must
                    {
                        return Expression::Literal(Literal::Null);
                    }
                }
            }
        }
        if sf.args.is_empty() {
            match sf.function {
                ScalarFunction::Add => return Expression::Literal(Literal::Integer(0)),
                ScalarFunction::Mul => return Expression::Literal(Literal::Integer(1)),
                _ => unreachable!("fold associative function only called on Add and Mul"),
            }
        }
        let mut non_literals = Vec::<Expression>::new();
        let (int_fold, long_fold, float_fold) = match sf.function {
            ScalarFunction::Add => {
                sf.args
                    .into_iter()
                    .fold((None, None, None), |(i, l, f), expr| match expr {
                        Expression::Literal(Literal::Integer(val)) => match i {
                            None => (Some(val), l, f),
                            Some(num) => (Some(num + val), l, f),
                        },
                        Expression::Literal(Literal::Long(val)) => match l {
                            None => (i, Some(val), f),
                            Some(num) => (i, Some(num + val), f),
                        },
                        Expression::Literal(Literal::Double(val)) => match f {
                            None => (i, l, Some(val)),
                            Some(num) => (i, l, Some(num + val)),
                        },
                        _ => {
                            non_literals.push(expr);
                            (i, l, f)
                        }
                    })
            }
            ScalarFunction::Mul => {
                sf.args
                    .into_iter()
                    .fold((None, None, None), |(i, l, f), expr| match expr {
                        Expression::Literal(Literal::Integer(val)) => match i {
                            None => (Some(val), l, f),
                            Some(num) => (Some(num * val), l, f),
                        },
                        Expression::Literal(Literal::Long(val)) => match l {
                            None => (i, Some(val), f),
                            Some(num) => (i, Some(num * val), f),
                        },
                        Expression::Literal(Literal::Double(val)) => match f {
                            None => (i, l, Some(val)),
                            Some(num) => (i, l, Some(num * val)),
                        },
                        _ => {
                            non_literals.push(expr);
                            (i, l, f)
                        }
                    })
            }
            _ => unreachable!("fold associative function is only called on And and Mul"),
        };
        let literals: Vec<Expression> = vec![
            int_fold.map(|val| Expression::Literal(Literal::Integer(val))),
            long_fold.map(|val| Expression::Literal(Literal::Long(val))),
            float_fold.map(|val| Expression::Literal(Literal::Double(val))),
        ]
        .into_iter()
        .flatten()
        .collect();
        let filtered_literals: Vec<Expression> = match sf.function {
            ScalarFunction::Add => literals
                .clone()
                .into_iter()
                .filter(|expr| !Self::numeric_eq(expr, 0.0))
                .collect(),
            ScalarFunction::Mul => literals
                .clone()
                .into_iter()
                .filter(|expr| !Self::numeric_eq(expr, 1.0))
                .collect(),
            _ => unreachable!("fold associative function is only called on And and Mul"),
        };
        let args = [filtered_literals, non_literals].concat();
        if args.is_empty() {
            return literals.last().unwrap().clone();
        }
        if args.len() == 1 {
            return args[0].clone();
        }
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: sf.function,
            args,
        })
    }

    // constant folds binary comparison functions
    fn fold_comparison_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use std::cmp::Ordering;
        assert!(
            sf.args.len() == 2,
            "binary comparison scalar functions must contain 2 args"
        );
        for expr in &sf.args {
            match expr.schema(&EMPTY_STATE) {
                Err(_) => break,
                Ok(sch) => {
                    if sch.satisfies(&Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::Atomic(Atomic::Null),
                    ])) == Satisfaction::Must
                    {
                        return Expression::Literal(Literal::Null);
                    }
                }
            }
        }
        let ord = match (&sf.args[0], &sf.args[1]) {
            (
                Expression::Literal(Literal::Boolean(l)),
                Expression::Literal(Literal::Boolean(r)),
            ) => l.partial_cmp(&r),
            (
                Expression::Literal(Literal::Integer(l)),
                Expression::Literal(Literal::Integer(r)),
            ) => l.partial_cmp(&r),
            (Expression::Literal(Literal::Long(l)), Expression::Literal(Literal::Long(r))) => {
                l.partial_cmp(&r)
            }
            (Expression::Literal(Literal::Double(l)), Expression::Literal(Literal::Double(r))) => {
                l.partial_cmp(&r)
            }
            (Expression::Literal(Literal::String(l)), Expression::Literal(Literal::String(r))) => {
                l.partial_cmp(&r)
            }
            _ => None,
        };
        if ord.is_none() {
            return Expression::ScalarFunction(sf);
        }
        let ord = ord.unwrap();
        let val = match sf.function {
            ScalarFunction::Eq => ord == Ordering::Equal,
            ScalarFunction::Gt => ord == Ordering::Greater,
            ScalarFunction::Gte => ord != Ordering::Less,
            ScalarFunction::Lt => ord == Ordering::Less,
            ScalarFunction::Lte => ord != Ordering::Greater,
            ScalarFunction::Neq => ord != Ordering::Equal,
            _ => unreachable!("non-comparison function cannot be called"),
        };
        Expression::Literal(Literal::Boolean(val))
    }

    // folds the between function
    fn fold_between(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert!(
            sf.args.len() == 3,
            "between scalar function must contain 3 args"
        );
        let (arg, bottom, top) = (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone());
        let new_sf = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Lte,
                    args: vec![arg.clone(), top],
                }),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gte,
                    args: vec![arg, bottom],
                }),
            ],
        });
        let folded_expr = self.visit_expression(new_sf);
        if let Expression::Literal(_) = folded_expr {
            folded_expr
        } else {
            Expression::ScalarFunction(sf)
        }
    }
}
impl Visitor for ConstantFoldExprVisitor {
    fn visit_expression(&mut self, e: Expression) -> Expression {
        let e = e.walk(self);
        match e {
            Expression::Array(_) => e,
            Expression::Cast(_) => e,
            Expression::Document(_) => e,
            Expression::Exists(_) => e,
            Expression::FieldAccess(_) => e,
            Expression::Is(_) => e,
            Expression::Like(_) => e,
            Expression::Literal(_) => e,
            Expression::Reference(_) => e,
            Expression::ScalarFunction(f) => match f.function {
                ScalarFunction::And | ScalarFunction::Or => self.fold_logical_function(f),
                ScalarFunction::Add | ScalarFunction::Mul => {
                    self.fold_associative_arithmetic_function(f)
                }
                ScalarFunction::Eq
                | ScalarFunction::Gt
                | ScalarFunction::Gte
                | ScalarFunction::Lt
                | ScalarFunction::Lte
                | ScalarFunction::Neq => self.fold_comparison_function(f),
                ScalarFunction::Between => self.fold_between(f),
                _ => Expression::ScalarFunction(f),
            },
            Expression::SearchedCase(_) => e,
            Expression::SimpleCase(_) => e,
            Expression::SubqueryComparison(_) => e,
            Expression::SubqueryExpression(_) => e,
            Expression::TypeAssertion(_) => e,
        }
    }

    fn visit_stage(&mut self, st: Stage) -> Stage {
        let st = st.walk(self);
        match st {
            Stage::Array(_) => st,
            Stage::Collection(_) => st,
            Stage::Filter(_) => st,
            Stage::Group(_) => st,
            Stage::Join(_) => st,
            Stage::Limit(_) => st,
            Stage::Offset(_) => st,
            Stage::Project(_) => st,
            Stage::Set(_) => st,
            Stage::Sort(_) => st,
        }
    }
}

pub fn fold_constants(st: Stage) -> Stage {
    let mut cf = ConstantFoldExprVisitor::default();
    cf.visit_stage(st)
}
