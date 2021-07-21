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
    fn fold_logical_functions(&mut self, sf: ScalarFunctionApplication) -> Expression {
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
        let mut non_literals = Vec::<Expression>::new();
        let (int_fold, long_fold, float_fold) = match sf.function {
            ScalarFunction::Add => {
                sf.args
                    .into_iter()
                    .fold((0, 0i64, 0.0), |(i, l, f), expr| match expr {
                        Expression::Literal(Literal::Integer(val)) => (i + val, l, f),
                        Expression::Literal(Literal::Long(val)) => (i, l + val, f),
                        Expression::Literal(Literal::Double(val)) => (i, l, f + val),
                        _ => {
                            non_literals.push(expr);
                            (i, l, f)
                        }
                    })
            }
            ScalarFunction::Mul => {
                sf.args
                    .into_iter()
                    .fold((1, 1i64, 1.0), |(i, l, f), expr| match expr {
                        Expression::Literal(Literal::Integer(val)) => (i * val, l, f),
                        Expression::Literal(Literal::Long(val)) => (i, l * val, f),
                        Expression::Literal(Literal::Double(val)) => (i, l, f * val),
                        _ => {
                            non_literals.push(expr);
                            (i, l, f)
                        }
                    })
            }
            _ => unreachable!("fold associative function is only called on And and Mul"),
        };
        let literals = vec![
            Expression::Literal(Literal::Integer(int_fold)),
            Expression::Literal(Literal::Long(long_fold)),
            Expression::Literal(Literal::Double(float_fold)),
        ];
        let literals = match sf.function {
            ScalarFunction::Add => literals
                .into_iter()
                .filter(|expr| !Self::numeric_eq(expr, 0.0))
                .collect(),
            ScalarFunction::Mul => literals
                .into_iter()
                .filter(|expr| !Self::numeric_eq(expr, 1.0))
                .collect(),
            _ => unreachable!("fold associative function is only called on And and Mul"),
        };
        let args = [literals, non_literals].concat();
        if args.len() == 1 {
            return args[0].clone();
        }
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: sf.function,
            args,
        })
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
                ScalarFunction::And | ScalarFunction::Or => Self::fold_logical_functions(self, f),
                ScalarFunction::Add | ScalarFunction::Mul => {
                    self.fold_associative_arithmetic_function(f)
                }
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
