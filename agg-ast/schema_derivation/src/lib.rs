mod match_schema_derivation;
pub(crate) use match_schema_derivation::MatchConstrainSchema;
mod negative_normalize;
#[cfg(test)]
mod negative_normalize_tests;
pub mod schema_derivation;

#[allow(unused_imports)]
pub use schema_derivation::*;
use std::collections::BTreeSet;
#[cfg(test)]
mod schema_derivation_tests;
#[cfg(test)]
mod test;

use bson::{Bson, Document};
use mongosql::schema::{self, Atomic, JaccardIndex, Schema, UNFOLDED_ANY};
use tailcall::tailcall;
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Clone)]
pub enum Error {
    #[error("Cannot derive schema for undefined literals")]
    InvalidLiteralType,
    #[error("Invalid JsonSchema")]
    InvalidJsonSchema,
    #[error("Type value for convert invalid: {0}")]
    InvalidConvertTypeValue(String),
    #[error("Cannot derive schema for unsupported group accumulator: {0}")]
    InvalidGroupAccumulator(String),
    #[error("Invalid type {0} at argument index: {1}")]
    InvalidType(Schema, usize),
    #[error("Invalid expression {0:?} at argument named: {1}")]
    InvalidExpressionForField(String, &'static str),
    #[error("Cannot derive schema for unsupported operator: {0}")]
    InvalidUntaggedOperator(String),
    #[error("Cannot derive schema for unsupported operator: {0:?}")]
    InvalidTaggedOperator(agg_ast::definitions::TaggedOperator),
    #[error("Cannot derive schema for unsupported stage: {0:?}")]
    InvalidStage(agg_ast::definitions::Stage),
    #[error("Unknown reference in current context: {0}")]
    UnknownReference(String),
    #[error("Not enough arguments for expression: {0}")]
    NotEnoughArguments(String),
    #[error("Missing from field in lookup")]
    MissingFromField,
}

#[macro_export]
macro_rules! maybe_any_of {
    ($schemas:expr) => {
        if $schemas.len() == 1 {
            $schemas.into_iter().next().unwrap()
        } else {
            Schema::AnyOf($schemas)
        }
    };
}

// promote_missing adds Schema::Missing to any non-required key in a Document schema. This is
// necessary for our operations to work correctly, since they operate on paths, leaving no way
// to check or modify required. We will rely on Schema::simplify to lower Schema::Missing back to
// removing from required, since Schema::Missing cannot be serialized.
fn promote_missing(schema: &Schema) -> Schema {
    // It would be much more efficient to do this in place, but we can't do that because of
    // BTreeSets. At some point we may want to consider moving to Vec, which would have no
    // effect on serialization, but we would have to be more careful to remove duplicates and
    // define an ordering.
    match schema {
        Schema::AnyOf(schemas) => {
            let schemas = schemas.iter().map(promote_missing).collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        Schema::Array(schema) => Schema::Array(Box::new(promote_missing(schema))),
        Schema::Document(doc) => {
            let mut doc = doc.clone();
            for (key, schema) in doc.keys.iter_mut() {
                *schema = promote_missing(schema);
                if !doc.required.contains(key) {
                    *schema = schema.union(&Schema::Missing);
                    doc.required.insert(key.clone());
                }
            }
            Schema::Document(doc)
        }
        _ => schema.clone(),
    }
}

// schema_difference removes a set of Schema from another Schema. This differs from
// Schema::intersection in that it does not use two Schemas as operands. Part of this is that
// schema_difference only ever happens with Atomic Schemas (and Missing, which is rather isomorphic
// to Atomic) and this is expedient. If we ever need to expand this to more complex Schemas, it may
// make sense to make this a real operator on two Schemas in the schema module.
//
// Note that this could also be achieved by complementing the Schema to be removed and intersecting
// it with the Schema to be modified, but this would be quite a bit less efficient.
fn schema_difference(schema: &mut Schema, to_remove: BTreeSet<Schema>) {
    match schema {
        Schema::Any => {
            *schema = UNFOLDED_ANY.clone();
            schema_difference(schema, to_remove);
        }
        Schema::AnyOf(schemas) => {
            let any_of_schemas = schemas
                .difference(&to_remove)
                .cloned()
                .collect::<BTreeSet<_>>();
            *schema = maybe_any_of!(any_of_schemas);
        }
        _ => (),
    }
}

/// Gets a mutable reference to a specific field or document path in the schema.
/// This allows us to insert, remove, or modify fields as we derive the schema for
/// operators and stages.
pub(crate) fn get_schema_for_path_mut(
    schema: &mut Schema,
    path: Vec<String>,
) -> Option<&mut Schema> {
    get_schema_for_path_mut_aux(schema, path, None, 0usize)
}

// This auxiliary function is a tail recursive helper for get_schema_for_path_mut. This allows us
// ways around the borrow checker that are very difficult to do in an iterative version. The
// tailcall package will ensure that this is optimized to a loop.
#[tailcall]
fn get_schema_for_path_mut_aux(
    schema: &mut Schema,
    mut path: Vec<String>,
    current_field: Option<String>,
    field_index: usize,
) -> Option<&mut Schema> {
    // since field_index is 0-indexed, if field_index == path.len(), we have already gone through each segment of the path,
    // so we should return the current schema
    if path.len() == field_index {
        return Some(schema);
    }
    // current_field is Some if the last Schema in the call stack was an Array
    let field = if let Some(current_field) = current_field {
        current_field
    } else {
        // we use index and std::mem::take here rather than remove to avoid reshuffling the Vec.
        std::mem::take(path.get_mut(field_index)?)
    };
    match schema {
        Schema::Document(d) => {
            get_schema_for_path_mut_aux(d.keys.get_mut(&field)?, path, None, field_index + 1)
        }
        Schema::Array(s) => get_schema_for_path_mut_aux(&mut **s, path, Some(field), field_index),
        _ => None,
    }
}

/// Gets or creates a mutable reference to a specific field or document path in the schema. This
/// should only be used in a $match context or in some other context where the MQL operator can
/// actually create fields. Consider a $match: we could think a field has type Any, or
/// AnyOf([String, Document]) before the $match, and the $match stage can only evaluate to true, if
/// that field is specifically a Document. In this case, we can refine that schema to Document, and
/// this can recurse to any depth in the Schema. Note that this still returns an Option because, if
/// the field is known to have a Schema that cannot be a Document, we cannot create a path! This
/// would mean that the aggregation pipeline in question will return no results, in fact, because
/// the $match stage will never evaluate to true.
fn get_or_create_schema_for_path_mut(
    schema: &mut Schema,
    path: Vec<String>,
) -> Option<&mut Schema> {
    get_or_create_schema_for_path_mut_aux(schema, path, None, 0usize)
}

// This auxiliary function is a tail recursive helper for get_schema_for_path_mut. This allows us
// ways around the borrow checker that are very difficult to do in an iterative version. The
// tailcall package will ensure that this is optimized to a loop.
#[tailcall]
fn get_or_create_schema_for_path_mut_aux(
    schema: &mut Schema,
    mut path: Vec<String>,
    current_field: Option<String>,
    field_index: usize,
) -> Option<&mut Schema> {
    // since field_index is 0-indexed, if field_index == path.len(), we have already gone through each segment of the path,
    // so we should return the current schema
    if path.len() == field_index {
        return Some(schema);
    }
    // current_field is Some if the last Schema in the call stack was an Array
    let field = if let Some(current_field) = current_field {
        current_field
    } else {
        // we use index and std::mem::take here rather than remove to avoid reshuffling the Vec.
        std::mem::take(path.get_mut(field_index)?)
    };
    match schema {
        Schema::Document(d) => {
            if !d.keys.contains_key(&field) && !d.additional_properties {
                return None;
            }
            d.keys.entry(field.clone()).or_insert(Schema::Any);
            get_or_create_schema_for_path_mut_aux(
                d.keys.get_mut(&field)?,
                path,
                None,
                field_index + 1,
            )
        }
        Schema::Array(s) => {
            get_or_create_schema_for_path_mut_aux(&mut **s, path, Some(field), field_index)
        }
        Schema::Any => {
            let mut d = schema::Document::any();
            d.keys.insert(field.clone(), Schema::Any);
            // this is a wonky way to do this, putting it in the map and then getting it back
            // out with this match, but it's what the borrow checker forces (we can't keep the
            // reference across the move of ownership into the Schema::Document constructor).
            *schema = Schema::Document(d);
            get_or_create_schema_for_path_mut_aux(
                schema.get_key_mut(&field)?,
                path,
                None,
                field_index + 1,
            )
        }
        Schema::AnyOf(schemas) => {
            // We do not currently support Array in AnyOf. That is an area for future work.
            // It is difficult because we use BTreeSet instead of Vec, which is not in place
            // mutable. We may want to reconsider that at some point.
            //
            // By first checking to see if there is a Document in the AnyOf, we can avoid
            // cloning the Document. In general, I expect that the AnyOf will be smaller than
            // the size of the Document schema, meaning this is more efficient than cloning
            // even ignoring "constant factors". This is especially true given that cloning
            // means memory allocation, which is quite a large "constant factor".
            if !schemas.iter().any(|s| matches!(s, &Schema::Document(_))) {
                // We only investigate Arrays if the AnyOf contains no Document schemas. There
                // could be a situation where the key should be in the Array and not the Document,
                // but this is just a best case attempt here. Worst case is we just don't refine a
                // Schema when we could. We can revisit this in the future, if it comes up more.
                let schemas = std::mem::take(schemas);
                if let Some(a) = schemas.into_iter().find_map(|s| {
                    if let Schema::Array(a) = s {
                        Some(a)
                    } else {
                        None
                    }
                }) {
                    *schema = Schema::Array(a);
                    return get_or_create_schema_for_path_mut_aux(
                        schema,
                        path,
                        Some(field),
                        field_index,
                    );
                }
                return None;
            }
            // This is how we avoid the clone. By doing a std::mem::take here, we can take
            // ownership of the schemas, and thus the Document schema. Sadly, we still have to
            // allocate a BTreeSet::default(), semantically. There is a price to pay for safety
            // sometimes, but, there is a good chance the compiler will be smart enough to know
            // that BTreeSet::default() is never used and optimize it out.
            let schemas = std::mem::take(schemas);
            let mut d = schemas.into_iter().find_map(|s| {
                if let Schema::Document(doc) = s {
                    // we would have to clone here without the std::mem::take above.
                    Some(doc)
                } else {
                    None
                }
            })?;
            // We can only add keys, if additionalProperties is true.
            if d.additional_properties {
                d.keys.entry(field.clone()).or_insert(Schema::Any);
            }
            // this is a wonky way to do this, putting it in the map and then getting it back
            // out with this match, but it's what the borrow checker forces (we can't keep the
            // reference across the move of ownership into the Schema::Document constructor).
            *schema = Schema::Document(d);
            get_or_create_schema_for_path_mut_aux(
                schema.get_key_mut(&field)?,
                path,
                None,
                field_index + 1,
            )
        }
        _ => None,
    }
}

// this helper simply checks for document schemas and inserts a field with a given schema into
// that document, ensuring it is required. It follows the same structure as get_or_create_schema_for_path_mut,
// except that we ignore Schema::Any's.
fn insert_required_key_into_document_helper(
    mut schema: Option<&mut Schema>,
    field_schema: Schema,
    field: String,
    overwrite: bool,
) -> Option<&mut Schema> {
    match schema {
        Some(Schema::Document(d)) => {
            if overwrite {
                d.keys.insert(field.clone(), field_schema);
            } else {
                d.keys.entry(field.clone()).or_insert(field_schema);
            }
            // even if the document already included the key, we want to mark it as required.
            // this enables nested field paths to be marked as required down to the required
            // key being inserted.
            d.required.insert(field.clone());
            d.keys.get_mut(&field)
        }
        Some(Schema::AnyOf(schemas)) => {
            if !schemas.iter().any(|s| matches!(s, &Schema::Document(_))) {
                return None;
            }
            let schemas = std::mem::take(schemas);
            let mut d = schemas.into_iter().find_map(|s| {
                if let Schema::Document(doc) = s {
                    Some(doc)
                } else {
                    None
                }
            })?;
            if overwrite {
                d.keys.insert(field.clone(), field_schema);
            } else {
                d.keys.entry(field.clone()).or_insert(field_schema);
            }
            // see comment above for why we require the field even if it existed in the document
            d.required.insert(field.clone());
            **(schema.as_mut()?) = Schema::Document(d);
            schema?.get_key_mut(&field)
        }
        _ => None,
    }
}

/// This function inserts a field into an existing document schema (or anyof containing a document).
/// It creates default documents as needed to populate the field path to the field, and marks the path
/// to the field as required, so that the field inserted is guaranteed to exist in the schema. This is
/// used to additively insert fields into schemas for stage schema derivation.
pub(crate) fn insert_required_key_into_document(
    schema: &mut Schema,
    field_schema: Schema,
    path: Vec<String>,
    overwrite: bool,
) {
    let mut schema = Some(schema);
    // create a required nested path of document schemas to the field we are trying to insert. For any field
    // that doesn't already exist, just create a default document, since we will add keys to it in the next iteration.
    for field in &path[..path.len() - 1] {
        schema = insert_required_key_into_document_helper(
            schema,
            Schema::Document(schema::Document::default()),
            field.clone(),
            false,
        );
    }
    // with a reference to the nested document that the field exists in, finally, insert the field with its type.
    insert_required_key_into_document_helper(
        schema,
        field_schema,
        path.last().unwrap().clone(),
        overwrite,
    );
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
    // if an array is empty, the only appropriate `items` Schema is Unsat.
    if bs.is_empty() {
        return Schema::Unsat;
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

#[macro_export]
macro_rules! array_element_schema_or_error {
    ($input_schema:expr,$input:expr) => {{
        match $input_schema {
            Schema::Array(a) => *a,
            Schema::AnyOf(ao) => {
                let mut array_type = None;
                for schema in ao.iter() {
                    if let Schema::Array(a) = schema {
                        array_type = Some(*a.clone());
                        break;
                    }
                }
                match array_type {
                    Some(t) => t,
                    None => {
                        return Err(Error::InvalidExpressionForField(
                            format!("{:?}", $input),
                            "input",
                        ))
                    }
                }
            }
            _ => {
                return Err(Error::InvalidExpressionForField(
                    format!("{:?}", $input),
                    "input",
                ))
            }
        }
    }};
}
