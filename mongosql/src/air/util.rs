use crate::air::{FieldRef, LiteralValue, Match, MqlOperator, SqlOperator, Stage, Variable};
use std::fmt;

impl Stage {
    pub(crate) fn get_source(&self) -> Box<Stage> {
        match self {
            Stage::AddFields(a) => a.source.clone(),
            Stage::Project(p) => p.source.clone(),
            Stage::Group(g) => g.source.clone(),
            Stage::Limit(l) => l.source.clone(),
            Stage::Sort(s) => s.source.clone(),
            Stage::Collection(_) => Box::new(self.clone()),
            Stage::Join(j) => j.left.clone(),
            Stage::Unwind(u) => u.source.clone(),
            Stage::Lookup(l) => l.source.clone(),
            Stage::ReplaceWith(r) => r.source.clone(),
            Stage::Match(m) => m.get_source(),
            Stage::UnionWith(u) => u.source.clone(),
            Stage::Skip(s) => s.source.clone(),
            Stage::Documents(_) => Box::new(self.clone()),
            Stage::EquiJoin(j) => j.source.clone(),
            Stage::EquiLookup(l) => l.source.clone(),
            Stage::Sentinel => Box::new(self.clone()),
        }
    }

    pub(crate) fn set_source(&mut self, new_source: Box<Stage>) {
        match self {
            Stage::AddFields(a) => a.source = new_source,
            Stage::Project(p) => p.source = new_source,
            Stage::Group(g) => g.source = new_source,
            Stage::Limit(l) => l.source = new_source,
            Stage::Sort(s) => s.source = new_source,
            Stage::Collection(_) => {}
            Stage::Join(j) => j.left = new_source,
            Stage::Unwind(u) => u.source = new_source,
            Stage::Lookup(l) => l.source = new_source,
            Stage::ReplaceWith(r) => r.source = new_source,
            Stage::Match(m) => m.set_source(new_source),
            Stage::UnionWith(u) => u.source = new_source,
            Stage::Skip(s) => s.source = new_source,
            Stage::Documents(_) => {}
            Stage::EquiJoin(j) => j.source = new_source,
            Stage::EquiLookup(l) => l.source = new_source,
            Stage::Sentinel => {}
        }
    }
}

impl Match {
    fn get_source(&self) -> Box<Stage> {
        match self {
            Match::ExprLanguage(e) => e.source.clone(),
            Match::MatchLanguage(m) => m.source.clone(),
        }
    }

    fn set_source(&mut self, new_source: Box<Stage>) {
        match self {
            Match::ExprLanguage(e) => e.source = new_source,
            Match::MatchLanguage(m) => m.source = new_source,
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

pub fn sql_op_to_mql_op(sql_op: SqlOperator) -> Option<MqlOperator> {
    let mql_op = match sql_op {
        SqlOperator::Eq => MqlOperator::Eq,
        SqlOperator::IndexOfCP => MqlOperator::IndexOfCP,
        SqlOperator::Lt => MqlOperator::Lt,
        SqlOperator::Lte => MqlOperator::Lte,
        SqlOperator::Gt => MqlOperator::Gt,
        SqlOperator::Gte => MqlOperator::Gte,
        SqlOperator::Ne => MqlOperator::Ne,
        SqlOperator::Not => MqlOperator::Not,
        SqlOperator::Size => MqlOperator::Size,
        SqlOperator::StrLenBytes => MqlOperator::StrLenBytes,
        SqlOperator::StrLenCP => MqlOperator::StrLenCP,
        SqlOperator::SubstrCP => MqlOperator::SubstrCP,
        SqlOperator::ToLower => MqlOperator::ToLower,
        SqlOperator::ToUpper => MqlOperator::ToUpper,
        SqlOperator::NullIf => MqlOperator::IfNull,
        SqlOperator::And => MqlOperator::And,
        SqlOperator::Or => MqlOperator::Or,
        SqlOperator::Slice => MqlOperator::Slice,
        SqlOperator::Cos => MqlOperator::Cos,
        SqlOperator::Sin => MqlOperator::Sin,
        SqlOperator::Tan => MqlOperator::Tan,
        SqlOperator::Log => MqlOperator::Log,
        SqlOperator::Mod => MqlOperator::Mod,
        SqlOperator::Round => MqlOperator::Round,
        SqlOperator::Sqrt => MqlOperator::Sqrt,
        SqlOperator::Split => MqlOperator::Split,
        SqlOperator::Between
        | SqlOperator::BitLength
        | SqlOperator::Coalesce
        | SqlOperator::ComputedFieldAccess
        | SqlOperator::CurrentTimestamp
        | SqlOperator::Neg
        | SqlOperator::Pos => return None,
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
            (LiteralValue::Undefined, LiteralValue::Undefined) => true,
            (LiteralValue::Boolean(a), LiteralValue::Boolean(b)) => a == b,
            (LiteralValue::String(a), LiteralValue::String(b)) => a == b,
            (LiteralValue::Integer(a), LiteralValue::Integer(b)) => a == b,
            (LiteralValue::Long(a), LiteralValue::Long(b)) => a == b,
            (LiteralValue::ObjectId(a), LiteralValue::ObjectId(b)) => a == b,
            (LiteralValue::DateTime(a), LiteralValue::DateTime(b)) => a == b,
            (LiteralValue::Decimal128(a), LiteralValue::Decimal128(b)) => a == b,
            (LiteralValue::DbPointer(a), LiteralValue::DbPointer(b)) => a == b,
            (LiteralValue::Timestamp(a), LiteralValue::Timestamp(b)) => a == b,
            (LiteralValue::RegularExpression(a), LiteralValue::RegularExpression(b)) => a == b,
            (LiteralValue::Symbol(a), LiteralValue::Symbol(b)) => a == b,
            (LiteralValue::Binary(a), LiteralValue::Binary(b)) => a == b,
            (LiteralValue::JavaScriptCode(a), LiteralValue::JavaScriptCode(b)) => a == b,
            (
                LiteralValue::JavaScriptCodeWithScope(a),
                LiteralValue::JavaScriptCodeWithScope(b),
            ) => a == b,
            (LiteralValue::MinKey, LiteralValue::MinKey) => true,
            (LiteralValue::MaxKey, LiteralValue::MaxKey) => true,
            (LiteralValue::Null, _) => false,
            (LiteralValue::Boolean(_), _) => false,
            (LiteralValue::String(_), _) => false,
            (LiteralValue::Integer(_), _) => false,
            (LiteralValue::Long(_), _) => false,
            (LiteralValue::Double(_), _) => false,
            (LiteralValue::ObjectId(_), _) => false,
            (LiteralValue::DateTime(_), _) => false,
            (LiteralValue::Decimal128(_), _) => false,
            (LiteralValue::DbPointer(_), _) => false,
            (LiteralValue::JavaScriptCode(_), _) => false,
            (LiteralValue::MinKey, _) => false,
            (LiteralValue::MaxKey, _) => false,
            (LiteralValue::JavaScriptCodeWithScope(_), _) => false,
            (LiteralValue::Binary(_), _) => false,
            (LiteralValue::Symbol(_), _) => false,
            (LiteralValue::RegularExpression(_), _) => false,
            (LiteralValue::Timestamp(_), _) => false,
            (LiteralValue::Undefined, _) => false,
        }
    }
}

macro_rules! generate_path_from {
    ($t: ty, $o: ident) => {
        impl From<$t> for $o {
            fn from(s: $t) -> $o {
                let mut split_string = s.split('.');
                let mut v = $o {
                    parent: None,
                    name: split_string.next().unwrap().to_string(),
                };
                for name in split_string {
                    v = $o {
                        parent: Some(Box::new(v)),
                        name: name.to_string(),
                    };
                }
                v
            }
        }
    };
}

generate_path_from!(&str, Variable);
generate_path_from!(String, Variable);
generate_path_from!(&str, FieldRef);
generate_path_from!(String, FieldRef);

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
