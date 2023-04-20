use crate::air::{FieldRef, LiteralValue, MQLOperator, SQLOperator, Variable};

impl FieldRef {
    pub(crate) fn root_parent(&self) -> String {
        match &self.parent {
            Some(parent) => parent.root_parent(),
            None => self.name.clone(),
        }
    }
}

pub fn sql_op_to_mql_op(sql_op: SQLOperator) -> Option<MQLOperator> {
    let mql_op = match sql_op {
        SQLOperator::Eq => MQLOperator::Eq,
        SQLOperator::IndexOfCP => MQLOperator::IndexOfCP,
        SQLOperator::Lt => MQLOperator::Lt,
        SQLOperator::Lte => MQLOperator::Lte,
        SQLOperator::Gt => MQLOperator::Gt,
        SQLOperator::Gte => MQLOperator::Gte,
        SQLOperator::Ne => MQLOperator::Ne,
        SQLOperator::Not => MQLOperator::Not,
        SQLOperator::Size => MQLOperator::Size,
        SQLOperator::StrLenBytes => MQLOperator::StrLenBytes,
        SQLOperator::StrLenCP => MQLOperator::StrLenCP,
        SQLOperator::SubstrCP => MQLOperator::SubstrCP,
        SQLOperator::ToLower => MQLOperator::ToLower,
        SQLOperator::ToUpper => MQLOperator::ToUpper,
        SQLOperator::NullIf => MQLOperator::IfNull,
        SQLOperator::And => MQLOperator::And,
        SQLOperator::Or => MQLOperator::Or,
        SQLOperator::Slice => MQLOperator::Slice,
        SQLOperator::Cos => MQLOperator::Cos,
        SQLOperator::Sin => MQLOperator::Sin,
        SQLOperator::Tan => MQLOperator::Tan,
        SQLOperator::Log => MQLOperator::Log,
        SQLOperator::Mod => MQLOperator::Mod,
        SQLOperator::Round => MQLOperator::Round,
        SQLOperator::Sqrt => MQLOperator::Sqrt,
        SQLOperator::Split => MQLOperator::Split,
        SQLOperator::Between
        | SQLOperator::BitLength
        | SQLOperator::Coalesce
        | SQLOperator::ComputedFieldAccess
        | SQLOperator::CurrentTimestamp
        | SQLOperator::Neg
        | SQLOperator::Pos => return None,
    };
    Some(mql_op)
}

impl PartialEq for LiteralValue {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            // for testing air representations, we want to be able to check NaN = NaN
            (LiteralValue::Double(a), LiteralValue::Double(b)) => {
                (a == b) | (a.is_nan() & b.is_nan())
            }
            // other than Double, we use the default implementation of PartialEq
            (LiteralValue::Null, LiteralValue::Null) => true,
            (LiteralValue::Boolean(a), LiteralValue::Boolean(b)) => a == b,
            (LiteralValue::String(a), LiteralValue::String(b)) => a == b,
            (LiteralValue::Integer(a), LiteralValue::Integer(b)) => a == b,
            (LiteralValue::Long(a), LiteralValue::Long(b)) => a == b,
            _ => false,
        }
    }
}

impl From<String> for Variable {
    fn from(s: String) -> Variable {
        let mut split_string = s.split('.');
        let mut v = Variable {
            parent: None,
            name: split_string.next().unwrap().to_string(),
        };
        for name in split_string {
            v = Variable {
                parent: Some(Box::new(v)),
                name: name.to_string(),
            };
        }
        v
    }
}

#[cfg(test)]
mod variable_from_string_tests {
    use super::*;

    macro_rules! test_var_from_string {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                #[allow(unused_imports)]
                let actual = Variable::from($input);
                let expected = $expected;
                assert_eq!(expected, actual);
            }
        };
    }

    test_var_from_string!(
        empty_string,
        expected = Variable {
            parent: None,
            name: "".to_string()
        },
        input = "".to_string()
    );

    test_var_from_string!(
        no_nesting,
        expected = Variable {
            parent: None,
            name: "a".to_string()
        },
        input = "a".to_string()
    );

    test_var_from_string!(
        nesting,
        expected = Variable {
            parent: Some(Box::new(Variable {
                parent: Some(Box::new(Variable {
                    parent: None,
                    name: "a".to_string(),
                })),
                name: "b".to_string()
            })),
            name: "c".to_string()
        },
        input = "a.b.c".to_string()
    );
}
