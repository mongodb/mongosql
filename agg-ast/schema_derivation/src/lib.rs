mod match_schema_derivation;
mod negative_normalize;
#[cfg(test)]
mod negative_normalize_tests;
pub mod schema_derivation;
#[allow(unused_imports)]
pub use schema_derivation::*;
#[cfg(test)]
mod schema_derivation_tests;
#[cfg(test)]
mod test;

use bson::{Bson, Document};
use mongosql::schema::{self, Atomic, JaccardIndex, Schema};
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Clone)]
pub enum Error {
    #[error("Cannot derive schema for undefined literals")]
    InvalidLiteralType,
    #[error("Type value for convert invalid: {0}")]
    InvalidConvertTypeValue(String),
    #[error("Invalid type {0} at argument index: {1}")]
    InvalidType(Schema, usize),
    #[error("Invalid expression {0:?} at argument named: {1}")]
    InvalidExpressionForField(String, &'static str),
    #[error("Cannot derive schema for unsupported operator: {0}")]
    InvalidUntaggedOperator(String),
    #[error("Cannot derive schema for unsupported operator: {0:?}")]
    InvalidTaggedOperator(agg_ast::definitions::TaggedOperator),
    #[error("Unknown reference in current context: {0}")]
    UnknownReference(String),
}

/// Gets a mutable reference to a specific field or document path in the schema.
/// This allows us to insert, remove, or modify fields as we derive the schema for
/// operators and stages.
#[allow(dead_code)]
pub(crate) fn get_schema_for_path_mut(
    schema: &mut Schema,
    path: Vec<String>,
) -> Option<&mut Schema> {
    let mut schema = Some(schema);
    for field in path {
        schema = match schema {
            Some(Schema::Document(d)) => d.keys.get_mut(&field),
            _ => {
                return None;
            }
        };
    }
    schema
}

/// remove field is a helper based on get_schema_for_path_mut which removes a field given a field path.
/// this is useful for operators that operate on specific fields, such as $unsetField, or operators
/// involving variables like $$REMOVE.
#[allow(dead_code)]
pub(crate) fn remove_field(schema: &mut Schema, path: Vec<String>) {
    if let Some((field, field_path)) = path.split_last() {
        let input = get_schema_for_path_mut(schema, field_path.into());
        if let Some(Schema::Document(d)) = input {
            d.keys.remove(field);
            d.required.remove(field);
        }
    }
}

pub fn schema_for_bson(b: &Bson) -> Schema {
    use Atomic::*;
    match b {
        Bson::Double(_) => Schema::Atomic(Double),
        Bson::String(_) => Schema::Atomic(String),
        Bson::Array(a) => Schema::Array(Box::new(schema_for_bson_array_elements(a))),
        Bson::Document(d) => schema_for_document(d),
        Bson::Boolean(_) => Schema::Atomic(Boolean),
        Bson::Null => Schema::Atomic(Null),
        Bson::RegularExpression(_) => Schema::Atomic(Regex),
        Bson::JavaScriptCode(_) => Schema::Atomic(Javascript),
        Bson::JavaScriptCodeWithScope(_) => Schema::Atomic(JavascriptWithScope),
        Bson::Int32(_) => Schema::Atomic(Integer),
        Bson::Int64(_) => Schema::Atomic(Long),
        Bson::Timestamp(_) => Schema::Atomic(Timestamp),
        Bson::Binary(_) => Schema::Atomic(BinData),
        Bson::Undefined => Schema::Atomic(Undefined),
        Bson::ObjectId(_) => Schema::Atomic(ObjectId),
        Bson::DateTime(_) => Schema::Atomic(Date),
        Bson::Symbol(_) => Schema::Atomic(Symbol),
        Bson::Decimal128(_) => Schema::Atomic(Decimal),
        Bson::MaxKey => Schema::Atomic(MaxKey),
        Bson::MinKey => Schema::Atomic(MinKey),
        Bson::DbPointer(_) => Schema::Atomic(DbPointer),
    }
}

/// Returns a [Schema] for a given BSON document.
pub fn schema_for_document(doc: &Document) -> Schema {
    Schema::Document(mongosql::schema::Document {
        keys: doc
            .iter()
            .map(|(k, v)| (k.to_string(), schema_for_bson(v)))
            .collect(),
        required: doc.iter().map(|(k, _)| k.to_string()).collect(),
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    })
}

// This may prove costly for very large arrays, and we may want to
// consider a limit on the number of elements to consider.
fn schema_for_bson_array_elements(bs: &[Bson]) -> Schema {
    // if an array is empty, we can't infer anything about it
    // we're safe to mark it as potentially null, as an empty array
    // satisfies jsonSchema search predicate
    if bs.is_empty() {
        return Schema::Atomic(Atomic::Null);
    }
    bs.iter()
        .map(schema_for_bson)
        .reduce(|acc, s| acc.union(&s))
        .unwrap_or(Schema::Any)
}

// schema_for_type_str generates a schema for a type name string used in a $type match operation.
fn schema_for_type_str(type_str: &str) -> Schema {
    match type_str {
        "double" => Schema::Atomic(Atomic::Double),
        "string" => Schema::Atomic(Atomic::String),
        "object" => Schema::Document(schema::Document::any()),
        "array" => Schema::Array(Box::new(Schema::Any)),
        "binData" => Schema::Atomic(Atomic::BinData),
        "undefined" => Schema::Atomic(Atomic::Undefined),
        "objectId" => Schema::Atomic(Atomic::ObjectId),
        "bool" => Schema::Atomic(Atomic::Boolean),
        "date" => Schema::Atomic(Atomic::Date),
        "null" => Schema::Atomic(Atomic::Null),
        "regex" => Schema::Atomic(Atomic::Regex),
        "dbPointer" => Schema::Atomic(Atomic::DbPointer),
        "javascript" => Schema::Atomic(Atomic::Javascript),
        "symbol" => Schema::Atomic(Atomic::Symbol),
        "javascriptWithScope" => Schema::Atomic(Atomic::JavascriptWithScope),
        "int" => Schema::Atomic(Atomic::Integer),
        "timestamp" => Schema::Atomic(Atomic::Timestamp),
        "long" => Schema::Atomic(Atomic::Long),
        "decimal" => Schema::Atomic(Atomic::Decimal),
        "minKey" => Schema::Atomic(Atomic::MinKey),
        "maxKey" => Schema::Atomic(Atomic::MaxKey),
        _ => unreachable!(),
    }
}

fn schema_for_type_numeric(type_as_int: i32) -> Schema {
    match type_as_int {
        1 => Schema::Atomic(Atomic::Double),
        2 => Schema::Atomic(Atomic::String),
        3 => Schema::Document(schema::Document::any()),
        4 => Schema::Array(Box::new(Schema::Any)),
        5 => Schema::Atomic(Atomic::BinData),
        6 => Schema::Atomic(Atomic::Undefined),
        7 => Schema::Atomic(Atomic::ObjectId),
        8 => Schema::Atomic(Atomic::Boolean),
        9 => Schema::Atomic(Atomic::Date),
        10 => Schema::Atomic(Atomic::Null),
        11 => Schema::Atomic(Atomic::Regex),
        12 => Schema::Atomic(Atomic::DbPointer),
        13 => Schema::Atomic(Atomic::Javascript),
        14 => Schema::Atomic(Atomic::Symbol),
        15 => Schema::Atomic(Atomic::JavascriptWithScope),
        16 => Schema::Atomic(Atomic::Integer),
        17 => Schema::Atomic(Atomic::Timestamp),
        18 => Schema::Atomic(Atomic::Long),
        19 => Schema::Atomic(Atomic::Decimal),
        -1 => Schema::Atomic(Atomic::MinKey),
        127 => Schema::Atomic(Atomic::MaxKey),
        _ => unreachable!(),
    }
}
