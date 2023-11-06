use bson::Bson;
use serde_json::Value;

use crate::mir::{schema::Error, Expression, LiteralValue};

pub fn ext_json_check(expr: &Expression) -> Result<&Expression, Error> {
    let literal_value = match expr {
        Expression::Literal(LiteralValue::String(s)) => s,
        _ => return Ok(expr),
    };

    // This is protection against implicit conversions. The string must convert to a Value::Object type
    // for it to have been actual extended json we care to warn the user about.
    let value = match serde_json::from_str(literal_value) {
        Ok(v) => match v {
            Value::Object(o) => o,
            _ => return Ok(expr),
        },
        Err(_) => return Ok(expr),
    };
    if let Ok(bson_value) = Bson::try_from(value) {
        let msg = "MongoSQL does not support direct comparisons with extended JSON";

        // Only types that can be constructed with extended json objects are checked.
        // All other types are impossible to arrive at, so they will be ignored.

        match match bson_value {
            Bson::Double(d) => Some(format!("Try `= {d}`")),
            Bson::String(_) => return Ok(expr),
            Bson::Array(_) => None,
            Bson::Document(_) => None,
            Bson::Boolean(_) => return Ok(expr),
            Bson::Null => return Ok(expr),
            Bson::RegularExpression(_) => None,
            Bson::JavaScriptCode(_) => None,
            Bson::JavaScriptCodeWithScope(_) => None,
            Bson::Int32(i) => Some(format!("Try `= {i}`")),
            Bson::Int64(i) => Some(format!("Try `= {i}`")),
            Bson::Timestamp(_) => None,
            Bson::Binary(_) => None,
            Bson::ObjectId(o) => Some(format!(r#"Try `= CAST("{o}" as objectId)`"#)),
            Bson::DateTime(_) => None,
            Bson::Symbol(_) => None,
            Bson::Decimal128(d) => Some(format!(r#"Try `= {d}`"#)),
            Bson::Undefined => None,
            Bson::MaxKey => None,
            Bson::MinKey => None,
            Bson::DbPointer(_) => None,
        } {
            Some(m) => Err(Error::ExtJsonComparison(format!("{msg}. {m}."))),
            None => Err(Error::ExtJsonComparison(format!("{msg}."))),
        }
    } else {
        Ok(expr)
    }
}

#[cfg(test)]
mod test {
    use crate::mir::{Expression, LiteralValue};

    use super::ext_json_check;

    #[test]
    fn oid() {
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            "5ca4bbcea2dd94ee58162a6a".to_string()
        ),))
        .is_ok());
        assert!(
            ext_json_check(&Expression::Literal(crate::mir::LiteralValue::String(
                r#"{"$oid":"5ca4bbcea2dd94ee58162a6a"}"#.to_string()
            ),))
            .is_err()
        );
    }

    #[test]
    fn bool() {
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            "true".to_string()
        ),))
        .is_ok())
    }

    #[test]
    fn int() {
        assert!(
            ext_json_check(&Expression::Literal(LiteralValue::String("55".to_string()),)).is_ok()
        );
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            r#"{"$numberInt": "55"}"#.to_string()
        ),))
        .is_err());
    }

    #[test]
    fn double() {
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            "55.55".to_string()
        ),))
        .is_ok());
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            r#"{"$numberDouble": "55.55"}"#.to_string()
        ),))
        .is_err());
    }

    #[test]
    fn decimal() {
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            "55.55".to_string()
        ),))
        .is_ok());
        assert!(ext_json_check(&Expression::Literal(LiteralValue::String(
            r#"{"$numberDecimal": "55.55"}"#.to_string()
        ),))
        .is_err());
    }
}
