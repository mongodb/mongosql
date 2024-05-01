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
    // DatasourceNotFoundInSchemaEnv is believed to be an internal error that can't be triggered by users.
    // Error 3007: NoSuchDatasource should cover it. However, since other parts of the code rely
    // on this error to function properly (see ticket SQL-1784 for more details), we've decided not to remove it.
    // Note: The errors.md file does not mention this error because there is no known way to trigger it with a SQL query.
    // Once/if we discover a way to trigger it, we will update the errors.md file accordingly.
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
    InvalidBinaryDataType,
    AggregationArgumentMustBeSelfComparable(String, Schema),
    CountDistinctStarNotSupported,
    InvalidComparison(&'static str, Schema, Schema),
    CannotMergeObjects(Schema, Schema, Satisfaction),
    AccessMissingField(String, Option<Vec<String>>),
    InvalidSubqueryCardinality,
    SortKeyNotSelfComparable(usize, Schema),
    GroupKeyNotSelfComparable(usize, Schema),
    UnwindIndexNameConflict(String),
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
            Error::SortKeyNotSelfComparable(_, _) => 1010,
            Error::GroupKeyNotSelfComparable(_, _) => 1011,
            Error::UnwindIndexNameConflict(_) => 1014,
            Error::CollectionNotFound(_, _) => 1016,
            Error::InvalidBinaryDataType => 1019,
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
            Error::UnwindIndexNameConflict(_) => None,
            Error::CollectionNotFound(_, _) => None,
            Error::InvalidBinaryDataType => None,
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
            Error::SortKeyNotSelfComparable(pos, schema) => format!("sort key at position {0} is not statically comparable to itself because it has the schema {1:?}", pos, schema),
            Error::GroupKeyNotSelfComparable(pos, schema) => format!("group key at position {0} is not statically comparable to itself because it has the schema {1:?}", pos, schema),
            Error::UnwindIndexNameConflict(name) => format!("UNWIND INDEX name '{0}' conflicts with existing field name", name),
            Error::CollectionNotFound(database, coll) => format!("unknown collection '{1}' in database '{0}'", database, coll),
            Error::InvalidBinaryDataType => "Binary data with subtype 3 found in schema".to_string(),
        }
    }
}
