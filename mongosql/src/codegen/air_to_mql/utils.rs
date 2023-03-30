use super::{Error, MqlCodeGenerator, Result};
use crate::air::{self};
use bson::{bson, Bson};

impl MqlCodeGenerator {
    /// Wraps a string value, s, in $literal if the condition, f, is true for the string.
    pub(crate) fn wrap_in_literal_if<F>(s: String, f: F) -> Bson
    where
        F: Fn(String) -> bool,
    {
        if f(s.clone()) {
            bson!({ "$literal": s })
        } else {
            Bson::String(s)
        }
    }

    pub(crate) fn convert_mql_type(ty: air::Type) -> Result<&'static str> {
        use air::Type::*;
        Ok(match ty {
            Array => return Err(Error::ConvertToArray),
            Document => return Err(Error::ConvertToDocument),
            _ => ty.to_str(),
        })
    }

    #[allow(clippy::only_used_in_recursion)] // false positive
    pub(crate) fn codegen_field_ref(&self, field_ref: air::FieldRef) -> String {
        match field_ref.parent {
            None => format!("${}", field_ref.name),
            Some(parent) => format!("{}.{}", self.codegen_field_ref(*parent), field_ref.name),
        }
    }
}
