use crate::{
    mir::{binding_tuple, schema::errors::Error, visitor::Visitor, *},
    schema::{Document, Satisfaction, Schema, NULLISH},
};

impl FieldAccess {
    /// Get the datasource Key at the root of a FieldAccess
    pub(crate) fn get_root_datasource(self) -> Result<binding_tuple::Key, Error> {
        match *self.expr {
            Expression::Reference(r) => Ok(r.key),
            Expression::FieldAccess(f) => f.get_root_datasource(),
            _ => Err(Error::InvalidUnwindPath),
        }
    }

    /// Get the field path of a FieldAccess. In the returned path, the
    /// last element is the beginning of the path and the first is the
    /// end of the path.  This method is specifically intended for use
    /// with set_field_schema (later in the file) and is therefore not
    /// available publicly.
    pub(in crate::mir) fn get_field_path(self) -> Result<Vec<String>, Error> {
        fn get_field_path_aux(
            e: Expression,
            path: &mut Vec<String>,
        ) -> Result<&mut Vec<String>, Error> {
            match e {
                Expression::Reference(_) => Ok(path),
                Expression::FieldAccess(f) => {
                    path.push(f.field.clone());
                    get_field_path_aux(*f.expr, path)
                }
                _ => Err(Error::InvalidUnwindPath),
            }
        }

        let mut v = vec![];
        let path = get_field_path_aux(Expression::FieldAccess(self), &mut v)?;
        Ok(path.clone())
    }

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
            })
        }
    }
}

/// get_optimized_field_accesses gets all optimized field accesses
/// in the provided Filter stage's condition.
pub fn get_optimized_field_accesses(filter: &Filter) -> Vec<FieldAccess> {
    let mut visitor = OptimizedMatchExistsFieldAccessGatherer {
        optimized_field_accesses: vec![],
    };

    visitor.visit_expression(filter.condition.clone());

    visitor.optimized_field_accesses
}

struct OptimizedMatchExistsFieldAccessGatherer {
    optimized_field_accesses: Vec<FieldAccess>,
}

impl Visitor for OptimizedMatchExistsFieldAccessGatherer {
    // Do not walk nested stages.
    fn visit_stage(&mut self, node: Stage) -> Stage {
        node
    }

    fn visit_optimized_match_exists(&mut self, node: OptimizedMatchExists) -> OptimizedMatchExists {
        self.optimized_field_accesses
            .push(node.field_access.clone());
        node
    }
}
