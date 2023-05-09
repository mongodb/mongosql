use crate::air::{FieldRef, LiteralValue, MQLOperator, SQLOperator, Stage, Variable};
use std::fmt;

impl Stage {
    pub(crate) fn get_source(&self) -> Box<Stage> {
        match self {
            Stage::Project(p) => p.source.clone(),
            Stage::Group(g) => g.source.clone(),
            Stage::Limit(l) => l.source.clone(),
            Stage::Sort(s) => s.source.clone(),
            Stage::Collection(_) => Box::new(self.clone()),
            Stage::Join(j) => j.left.clone(),
            Stage::Unwind(u) => u.source.clone(),
            Stage::Lookup(l) => l.source.clone(),
            Stage::ReplaceWith(r) => r.source.clone(),
            Stage::Match(m) => m.source.clone(),
            Stage::UnionWith(u) => u.source.clone(),
            Stage::Skip(s) => s.source.clone(),
            Stage::Documents(_) => Box::new(self.clone()),
        }
    }

    pub(crate) fn set_source(&mut self, new_source: Box<Stage>) {
        match self {
            Stage::Project(p) => p.source = new_source,
            Stage::Group(g) => g.source = new_source,
            Stage::Limit(l) => l.source = new_source,
            Stage::Sort(s) => s.source = new_source,
            Stage::Collection(_) => {}
            Stage::Join(j) => j.left = new_source,
            Stage::Unwind(u) => u.source = new_source,
            Stage::Lookup(l) => l.source = new_source,
            Stage::ReplaceWith(r) => r.source = new_source,
            Stage::Match(m) => m.source = new_source,
            Stage::UnionWith(u) => u.source = new_source,
            Stage::Skip(s) => s.source = new_source,
            Stage::Documents(_) => {}
        }
    }
}

impl FieldRef {
    pub(crate) fn root_parent(&self) -> String {
        match &self.parent {
            Some(parent) => parent.root_parent(),
            None => self.name.clone(),
        }
    }
}

impl fmt::Display for FieldRef {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.parent {
            Some(parent) => write!(f, "{}.{}", parent, self.name),
            None => write!(f, "{}", self.name),
        }
    }
}

impl fmt::Display for Variable {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.parent {
            Some(parent) => write!(f, "{}.{}", parent, self.name),
            None => write!(f, "{}", self.name),
        }
    }
}

pub fn match_sql_to_mql_op(sql_op: SQLOperator) -> Option<MQLOperator> {
    match sql_op {
        SQLOperator::And | SQLOperator::Or => None,
        _ => sql_op_to_mql_op(sql_op),
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

impl From<String> for FieldRef {
    fn from(s: String) -> FieldRef {
        let mut split_string = s.split('.');
        let mut fr = FieldRef {
            parent: None,
            name: split_string.next().unwrap().to_string(),
        };
        for name in split_string {
            fr = FieldRef {
                parent: Some(Box::new(fr)),
                name: name.to_string(),
            };
        }
        fr
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

#[cfg(test)]
mod field_ref_from_string_tests {
    use super::*;

    macro_rules! test_field_ref_from_string {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                #[allow(unused_imports)]
                let actual = FieldRef::from($input);
                let expected = $expected;
                assert_eq!(expected, actual);
            }
        };
    }

    test_field_ref_from_string!(
        empty_string,
        expected = FieldRef {
            parent: None,
            name: "".to_string()
        },
        input = "".to_string()
    );

    test_field_ref_from_string!(
        no_nesting,
        expected = FieldRef {
            parent: None,
            name: "a".to_string()
        },
        input = "a".to_string()
    );

    test_field_ref_from_string!(
        nesting,
        expected = FieldRef {
            parent: Some(Box::new(FieldRef {
                parent: Some(Box::new(FieldRef {
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

#[cfg(test)]
mod field_ref_fmt {
    use super::*;

    #[test]
    fn no_parent() {
        let field_ref = FieldRef {
            parent: None,
            name: "field".to_string(),
        };
        assert_eq!(format!("{field_ref}"), "field");
    }

    #[test]
    fn one_parent() {
        let parent = Box::new(FieldRef {
            parent: None,
            name: "parent".to_string(),
        });
        let field_ref = FieldRef {
            parent: Some(parent),
            name: "field".to_string(),
        };
        assert_eq!(format!("{field_ref}"), "parent.field");
    }

    #[test]
    fn two_level_parent() {
        let root = Box::new(FieldRef {
            parent: None,
            name: "root".to_string(),
        });
        let parent = Box::new(FieldRef {
            parent: Some(root),
            name: "parent".to_string(),
        });
        let field_ref = FieldRef {
            parent: Some(parent),
            name: "field".to_string(),
        };
        assert_eq!(format!("{field_ref}"), "root.parent.field");
    }
}

#[cfg(test)]
mod variable_fmt {
    use super::*;

    #[test]
    fn no_parent() {
        let variable = Variable {
            parent: None,
            name: "field".to_string(),
        };
        assert_eq!(format!("{variable}"), "field");
    }

    #[test]
    fn one_parent() {
        let parent = Box::new(Variable {
            parent: None,
            name: "parent".to_string(),
        });
        let variable = Variable {
            parent: Some(parent),
            name: "field".to_string(),
        };
        assert_eq!(format!("{variable}"), "parent.field");
    }

    #[test]
    fn two_level_parent() {
        let root = Box::new(Variable {
            parent: None,
            name: "root".to_string(),
        });
        let parent = Box::new(Variable {
            parent: Some(root),
            name: "parent".to_string(),
        });
        let variable = Variable {
            parent: Some(parent),
            name: "field".to_string(),
        };
        assert_eq!(format!("{variable}"), "root.parent.field");
    }
}
