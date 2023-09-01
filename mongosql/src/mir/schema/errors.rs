use crate::{
    mir::binding_tuple,
    schema::{Satisfaction, Schema},
    usererror::{
        util::{generate_suggestion, unsat_check},
        UserError, UserErrorDisplay,
    },
};

#[derive(Debug, UserErrorDisplay, PartialEq, Eq, Clone)]
pub enum Error {
    DatasourceNotFoundInSchemaEnv(binding_tuple::Key),
    IncorrectArgumentCount {
        name: &'static str,
        required: usize,
        found: usize,
    },
    SchemaChecking {
        name: &'static str,
        required: Schema,
        found: Schema,
    },
    AggregationArgumentMustBeSelfComparable(String, Schema),
    CountDistinctStarNotSupported,
    InvalidComparison(&'static str, Schema, Schema),
    CannotMergeObjects(Schema, Schema, Satisfaction),
    AccessMissingField(String, Option<Vec<String>>),
    InvalidSubqueryCardinality,
    DuplicateKey(binding_tuple::Key),
    SortKeyNotSelfComparable(usize, Schema),
    GroupKeyNotSelfComparable(usize, Schema),
    UnaliasedFieldAccessWithNoReference(usize),
    UnaliasedNonFieldAccessExpression(usize),
    UnwindIndexNameConflict(String),
    InvalidUnwindPath,
    CollectionNotFound(String, String),
}

impl UserError for Error {
    fn code(&self) -> u32 {
        match self {
            Error::DatasourceNotFoundInSchemaEnv(_) => 1000,
            Error::IncorrectArgumentCount { .. } => 1001,
            Error::SchemaChecking { .. } => 1002,
            Error::AggregationArgumentMustBeSelfComparable(_, _) => 1003,
            Error::CountDistinctStarNotSupported => 1004,
            Error::InvalidComparison(_, _, _) => 1005,
            Error::CannotMergeObjects(_, _, _) => 1006,
            Error::AccessMissingField(_, _) => 1007,
            Error::InvalidSubqueryCardinality => 1008,
            Error::DuplicateKey(_) => 1009,
            Error::SortKeyNotSelfComparable(_, _) => 1010,
            Error::GroupKeyNotSelfComparable(_, _) => 1011,
            Error::UnaliasedFieldAccessWithNoReference(_) => 1012,
            Error::UnaliasedNonFieldAccessExpression(_) => 1013,
            Error::UnwindIndexNameConflict(_) => 1014,
            Error::InvalidUnwindPath => 1015,
            Error::CollectionNotFound(_, _) => 1016,
        }
    }

    fn user_message(&self) -> Option<String> {
        match self {
            Error::DatasourceNotFoundInSchemaEnv(_) => None,
            Error::IncorrectArgumentCount { .. } => None,
            Error::SchemaChecking {
                name,
                required,
                found,
            } => {
                let simplified_required = Schema::simplify(required);
                let simplified_found = Schema::simplify(found);

                if let Some(message) =
                    unsat_check(vec![simplified_required.clone(), simplified_found.clone()])
                {
                    Some(message)
                } else {
                    Some(format!("Incorrect argument type for `{name}`. Required: {simplified_required}. Found: {simplified_found}."))
                }
            }
            Error::AggregationArgumentMustBeSelfComparable(agg, schema) => {
                let simplified_schema = Schema::simplify(schema);

                if let Some(message) = unsat_check(vec![simplified_schema.clone()]) {
                    Some(message)
                } else {
                    Some(format!("Cannot perform `{agg}` aggregation over the type `{simplified_schema}` as it is not comparable to itself."))
                }
            }
            Error::CountDistinctStarNotSupported => None,
            Error::InvalidComparison(func, s1, s2) => {
                let simplified_s1 = Schema::simplify(s1);
                let simplified_s2 = Schema::simplify(s2);

                if let Some(message) =
                    unsat_check(vec![simplified_s1.clone(), simplified_s2.clone()])
                {
                    Some(message)
                } else {
                    Some(format ! ("Invalid use of `{func}` due to incomparable types: `{simplified_s1}` cannot be compared to `{simplified_s2}`."))
                }
            }
            Error::AccessMissingField(field, found_fields) => {
                if let Some(possible_fields) = found_fields {
                    let suggestions = generate_suggestion(field, possible_fields);
                    match suggestions {
                        Ok(suggested_fields) => {
                            if suggested_fields.is_empty() {
                                Some(format!("Cannot access field `{field}` because it could not be found."))
                            } else {
                                Some(format!(
                                    "Cannot access field `{field}` because it could not be found. Did you mean: {}",
                                    suggested_fields
                                        .iter()
                                        .map(|x| x.to_string())
                                        .collect::<Vec<String>>()
                                        .join(", ")
                                ))
                            }
                        }
                        Err(e) => Some(format!("Cannot access field `{field}` because it could not be found. Internal error: {e}")),
                    }
                } else {
                    Some(format!(
                        "Cannot access field `{field}` because it could not be found."
                    ))
                }
            }
            Error::CannotMergeObjects(s1, s2, _) => {
                let overlap = s1
                    .keys()
                    .into_iter()
                    .filter(|k| !matches!(s2.contains_field(k), Satisfaction::Not))
                    .collect::<Vec<_>>();

                Some(format!(
                    "Cannot merge objects because they have overlapping key(s): `{}`",
                    overlap.join("`, `")
                ))
            }
            Error::InvalidSubqueryCardinality => None,
            Error::DuplicateKey(_) => None,
            Error::SortKeyNotSelfComparable(_, schema) => {
                let simplified_schema = Schema::simplify(schema);

                if let Some(message) = unsat_check(vec![simplified_schema.clone()]) {
                    Some(message)
                } else {
                    Some(format!(
                        "Cannot sort by key because `{simplified_schema}` can't be compared against itself."
                    ))
                }
            }
            Error::GroupKeyNotSelfComparable(_, schema) => {
                let simplified_schema = Schema::simplify(schema);

                if let Some(message) = unsat_check(vec![simplified_schema.clone()]) {
                    Some(message)
                } else {
                    Some(format!(
                        "Cannot group by key because `{simplified_schema}` can't be compared against itself.",
                    ))
                }
            }
            Error::UnaliasedFieldAccessWithNoReference(_) => None,
            Error::UnaliasedNonFieldAccessExpression(_) => None,
            Error::UnwindIndexNameConflict(_) => None,
            Error::InvalidUnwindPath => None,
            Error::CollectionNotFound(_, _) => None,
        }
    }

    fn technical_message(&self) -> String {
        match self {
            Error::DatasourceNotFoundInSchemaEnv(datasource) => format!("datasource {0:?} not found in schema environment", datasource),
            Error::IncorrectArgumentCount {name, required, found} => format!("incorrect argument count for {name}: required {required}, found {found}"),
            Error::SchemaChecking {name, required, found } => format!("schema checking failed for {name}: required {required:?}, found {found:?}"),
            Error::AggregationArgumentMustBeSelfComparable(aggs, schema) => format!("cannot have {0:?} aggregations over the schema: {1:?} as it is not comparable to itself", aggs, schema),
            Error::CountDistinctStarNotSupported => "COUNT(DISTINCT *) is not supported".to_string(),
            Error::InvalidComparison(func, s1, s2) => format!("invalid comparison for {0}: {1:?} cannot be compared to {2:?}", func, s1, s2),
            Error::CannotMergeObjects(s1, s2, sat) => format!("cannot merge objects {0:?} and {1:?} as they {2:?} have overlapping keys", s1, s2, sat),
            Error::AccessMissingField(field, _) => format!("cannot access field {0} because it does not exist", field),
            Error::InvalidSubqueryCardinality => "cardinality of the subquery's result set may be greater than 1".to_string(),
            Error::DuplicateKey(datasource) => format!("cannot create schema environment with duplicate datasource: {0:?}", datasource),
            Error::SortKeyNotSelfComparable(pos, schema) => format!("sort key at position {0} is not statically comparable to itself because it has the schema {1:?}", pos, schema),
            Error::GroupKeyNotSelfComparable(pos, schema) => format!("group key at position {0} is not statically comparable to itself because it has the schema {1:?}", pos, schema),
            Error::UnaliasedFieldAccessWithNoReference(pos) => format!("group key at position {0} is an unaliased field access with no datasource reference", pos),
            Error::UnaliasedNonFieldAccessExpression(pos) => format!("group key at position {0} is an unaliased non-field access expression", pos),
            Error::UnwindIndexNameConflict(name) => format!("UNWIND INDEX name '{0}' conflicts with existing field name", name),
            Error::InvalidUnwindPath => "UNWIND PATH option must be an identifier".to_string(),
            Error::CollectionNotFound(database, coll) => format!("unknown collection '{1}' in database '{0}'", database, coll),
        }
    }
}
