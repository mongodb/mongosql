use crate::air::{
    self,
    desugarer::{Error, Pass, Result},
    visitor::Visitor,
    Expression, LiteralValue, Type,
};
use bson::{oid::ObjectId, Decimal128};
use chrono::prelude::*;
use std::str::FromStr;

/// Desugars any Convert with a constant input into a LiteralValue with the correct target type.
#[derive(Default)]
pub struct FoldConvertsDesugarerPass;

impl Pass for FoldConvertsDesugarerPass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage> {
        let mut visitor = FoldConvertsDesugarerVisitor::default();
        let out = pipeline.walk(&mut visitor);
        if let Some(err) = visitor.error {
            Err(err)
        } else {
            Ok(out)
        }
    }
}

#[derive(Default)]
struct FoldConvertsDesugarerVisitor {
    error: Option<Error>,
}

impl FoldConvertsDesugarerVisitor {
    // None means no conversion could occur
    // Some(Err(_)) suggests there was an opportunity for conversion but the conversion statically
    // fails... since it will also fail at runtime, we want to report an error now.
    // Some(Ok(_)) means we successfully converted a literal.
    fn convert_literal(l: &LiteralValue, to: Type) -> Option<Result<Expression>> {
        match l {
            LiteralValue::String(s) => Self::convert_string_literal(s, to),
            LiteralValue::Integer(i) => Self::convert_numerical_literal(*i, to),
            LiteralValue::Long(l) => Self::convert_numerical_literal(*l, to),
            LiteralValue::Double(d) => Self::convert_numerical_literal(*d, to),
            _ => None,
        }
    }

    fn convert_string_literal(s: &str, to: Type) -> Option<Result<Expression>> {
        match to {
            // We'll handle the no-op convert too. But we are not handling every possible case.
            // This is mostly used to easily test that we are properly recursing, but could offer
            // some benefit.
            Type::String => Some(Ok(Expression::Literal(LiteralValue::String(s.to_string())))),
            Type::Datetime => {
                let chrono_dt: Option<chrono::DateTime<Utc>> = s.parse().ok();
                match chrono_dt {
                    Some(chrono_dt) => Some(Ok(Expression::Literal(LiteralValue::DateTime(
                        chrono_dt.into(),
                    )))),
                    None => Some(Err(Error::InvalidConstantConvert(Type::Datetime))),
                }
            }
            Type::Decimal128 => {
                let dec = Decimal128::from_str(s).ok();
                match dec {
                    Some(dec) => Some(Ok(Expression::Literal(LiteralValue::Decimal128(dec)))),
                    None => Some(Err(Error::InvalidConstantConvert(Type::Decimal128))),
                }
            }
            Type::ObjectId => {
                let oid = ObjectId::parse_str(s).ok();
                match oid {
                    Some(oid) => Some(Ok(Expression::Literal(LiteralValue::ObjectId(oid)))),
                    None => Some(Err(Error::InvalidConstantConvert(Type::ObjectId))),
                }
            }
            _ => None,
        }
    }

    fn convert_numerical_literal<T: std::fmt::Display>(
        i: T,
        to: Type,
    ) -> Option<Result<Expression>> {
        match to {
            Type::Decimal128 => {
                let dec = Decimal128::from_str(&format!("{}", i)).ok();
                match dec {
                    Some(dec) => Some(Ok(Expression::Literal(LiteralValue::Decimal128(dec)))),
                    None => Some(Err(Error::InvalidConstantConvert(Type::Decimal128))),
                }
            }
            _ => None,
        }
    }
}

impl Visitor for FoldConvertsDesugarerVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        let node = node.walk(self);
        if let Expression::Convert(ref c) = node {
            if let Expression::Literal(ref l) = *c.input {
                let folded = Self::convert_literal(l, c.to);
                return match folded {
                    None => node,
                    Some(Err(e)) => {
                        self.error = Some(e);
                        node
                    }
                    Some(Ok(new_node)) => new_node,
                };
            }
        }
        node
    }
}
