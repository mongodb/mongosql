use crate::{
    mir::*,
    schema::{Document, Satisfaction, Schema, SchemaEnvironment, NULLISH},
};

impl FieldAccess {
    /// If this FieldAccess is "pure", then return the scope of its root
    /// Reference. A "pure" FieldAccess consists only of other FieldAccess
    /// expressions up the expr tree, ending at a Reference.
    pub(crate) fn scope_if_pure(&self) -> Option<u16> {
        match &*self.expr {
            Expression::Reference(r) => Some(r.key.scope),
            Expression::FieldAccess(fa) => fa.scope_if_pure(),
            _ => None,
        }
    }

    /// If this FieldAccess is "pure", then return a string representation
    /// of it. A "pure" FieldAccess consists only of other FieldAccess
    /// expressions up the expr tree, ending at a Reference.
    pub(crate) fn to_string_if_pure(&self) -> Option<String> {
        match &*self.expr {
            Expression::Reference(r) => Some(format!(
                "{:?}_{}.{}",
                r.key.datasource, r.key.scope, self.field
            )),
            Expression::FieldAccess(fa) => fa
                .to_string_if_pure()
                .map(|s| format!("{s}.{}", self.field)),
            _ => None,
        }
    }
}

/// Lift inner schema from any instance of Array schema in the input schema.
/// This can either be at the top level, e.g. s = Array(Integer), or nested
/// in an AnyOf, e.g. s = AnyOf(Array(Integer), Array(String)). If the input
/// schema is an AnyOf, the other non-Array schemas are retained, unless it
/// is indicated that null and missing should be excluded.
///
/// This method is specifically intended for use with Unwind schema checking
/// and is therefore not available publicly. Also, given this use case, it is
/// not comprehensively defined for non-Array/non-AnyOf schemas since those
/// are irrelevant here.
pub fn lift_array_schemas(s: Schema, retain_null_and_missing: bool) -> Schema {
    match s {
        Schema::Array(a) => *a,
        Schema::AnyOf(ao) => Schema::AnyOf(
            ao.iter()
                .filter(|s1| {
                    // If we want to retain null and missing, trivially keep everything
                    retain_null_and_missing
                        // otherwise, only keep things that are not null or missing
                        || s1.satisfies(&NULLISH) != Satisfaction::Must
                })
                .map(|s1| lift_array_schemas(s1.clone(), retain_null_and_missing))
                .collect::<_>(),
        ),
        _ => s,
    }
}

/// set_field_schema sets the schema of a field in the argument schema `s`
/// for any Document schemas described by `s`.
///
/// The `field_path` is specified as a stack of Strings. The idea of a
/// "field path" isn't generally well defined for the Schema type. For
/// the purposes of this method it means a sequence of "keys" in nested
/// Schema::Documents.
///
/// When `s` describes a Document schema, the method attempts to set the
/// schema for the `field_path` according to the following rules:
///   - If there is only one element left,
///     - Add the element to the Document's keys, using the `field_schema`
///       as the schema for the field.
///     - Update the Document's required set according to the
///       `field_required` argument
///   - If there are multiple elements left, pop the `field_path` stack.
///     - If the Document does not contain this element in its keys map,
///       stop because there is nothing left to do.
///     - If the Document does contain this element in its keys map,
///       attempt to set the schema of the remaining `field_path` in the
///       element's corresponding schema.
pub fn set_field_schema(
    s: Schema,
    field_path: &mut Vec<String>,
    field_schema: Schema,
    field_required: bool,
) -> Schema {
    match s {
        Schema::Any | Schema::Missing | Schema::Unsat | Schema::Array(_) | Schema::Atomic(_) => {
            s.clone()
        }
        Schema::AnyOf(ao) => Schema::AnyOf(
            ao.iter()
                .map(|s| {
                    set_field_schema(
                        s.clone(),
                        &mut field_path.clone(),
                        field_schema.clone(),
                        field_required,
                    )
                })
                .collect(),
        ),
        Schema::Document(ref d) => {
            let mut keys = d.clone().keys;
            let mut required = d.clone().required;

            let field_name = field_path.pop().unwrap();
            if field_path.is_empty() {
                // At this point, the field_path is exhausted, so
                // this is where the field_schema should be set.
                keys.insert(field_name.clone(), field_schema);

                if field_required {
                    required.insert(field_name);
                } else {
                    required.remove(field_name.as_str());
                }
            } else {
                // If the Document contains the next part of the field path,
                // attempt to set the schema at the next level. Otherwise,
                // do nothing since the field described by the path is not
                // described by this Document.
                let new_sub_schema = match d.keys.get(field_name.as_str()) {
                    None => s.clone(),
                    Some(sub_schema) => set_field_schema(
                        sub_schema.clone(),
                        field_path,
                        field_schema,
                        field_required,
                    ),
                };
                keys.insert(field_name.clone(), new_sub_schema);
            }

            Schema::Document(Document {
                keys,
                required,
                additional_properties: d.additional_properties,
                ..Default::default()
            })
        }
    }
}

// When constructing the schema environment for Group and AddFields, calls to union() and
// union_schema_for_datasource() combine schemas with the same Key under an AnyOf. However, we want
// group keys and aggregations under bottom to be combined under a single document, and we want
// AddFields keys under bottom to be added to the incoming bottom schema. To accomplish this, we
// pull out the schema for bottom, and if it's an AnyOf set of documents, use Document::merge() to
// combine the documents.
pub(crate) fn merge_bot_any_of_document_schemas(
    scope_level: u16,
    mut schema_env: SchemaEnvironment,
) -> SchemaEnvironment {
    if let Some(bot_schema) = schema_env.remove(&binding_tuple::Key::bot(scope_level)) {
        // Call simplify() to flatten the AnyOfs into a single set.
        if let Schema::AnyOf(schemas) = Schema::simplify(&bot_schema) {
            let merged_document = schemas.into_iter().fold(Document::empty(), |acc, schema| {
                acc.merge(match schema {
                    Schema::Document(doc) => doc,
                    // Bottom is always a document according to the spec, and so
                    // its schema can only be Document or AnyOf(documents). After
                    // simplification, we know that we're inside the only AnyOf,
                    // so there should only be documents.
                    _ => unreachable!("Bottom should always be Document"),
                })
            });
            schema_env.union_schema_for_datasource(
                binding_tuple::Key::bot(scope_level),
                Schema::Document(merged_document),
            )
        } else {
            // Put the value back unchanged if the top-level bottom schema wasn't AnyOf.
            schema_env.union_schema_for_datasource(binding_tuple::Key::bot(scope_level), bot_schema)
        }
    } else {
        schema_env
    }
}
