use crate::{
    ir::{definitions::*, schema::SchemaInferenceState, visitor::Visitor},
    schema::{Atomic, Satisfaction, Schema, SchemaEnvironment},
};
use lazy_static::lazy_static;

lazy_static! {
    static ref EMPTY_STATE: SchemaInferenceState = SchemaInferenceState {
        scope_level: 0u16,
        env: SchemaEnvironment::default(),
    };
}

#[derive(Default)]
pub(crate) struct ConstantFoldExprVisitor;

impl ConstantFoldExprVisitor {
    // Constant folds boolean functions
    fn fold_logical_functions(&mut self, sf: ScalarFunctionApplication) -> Expression {
        let (nullish, non_nullish): (Vec<Expression>, Vec<Expression>) =
            sf.args.clone().into_iter().partition(|e| {
                e.schema(&EMPTY_STATE)
                    .unwrap_or(Schema::Any)
                    .satisfies(&Schema::AnyOf(vec![
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
