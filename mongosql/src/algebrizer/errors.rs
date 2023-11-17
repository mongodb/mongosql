use crate::{
    ast::{self},
    mir::{
        self,
        binding_tuple::{DatasourceName, Key},
    },
    schema::Satisfaction,
    usererror::{util::generate_suggestion, UserError, UserErrorDisplay},
};
use std::collections::HashSet;

#[derive(Debug, UserErrorDisplay, PartialEq)]
pub enum Error {
    NonStarStandardSelectBody,
    ArrayDatasourceMustBeLiteral,
    DistinctSelect,
    DistinctUnion,
    NoSuchDatasource(DatasourceName),
    FieldNotFound(String, Option<Vec<String>>),
    AmbiguousField(String),
    StarInNonCount,
    AggregationInPlaceOfScalar(String),
    ScalarInPlaceOfAggregation(String),
    NonAggregationInPlaceOfAggregation(usize),
    AggregationFunctionMustHaveOneArgument,
    DistinctScalarFunction,
    DerivedDatasourceOverlappingKeys(
        crate::schema::Schema,
        crate::schema::Schema,
        String,
        Satisfaction,
    ),
    SchemaChecking(mir::schema::Error),
    NoOuterJoinCondition,
    DuplicateKey(Key),
    InvalidSubqueryDegree,
    DuplicateDocumentKey(String),
    DuplicateFlattenOption(ast::FlattenOption),
    CannotEnumerateAllFieldPaths(crate::schema::Schema),
    PolymorphicObjectSchema(String),
    DuplicateUnwindOption(ast::UnwindOption),
    NoUnwindPath,
    InvalidUnwindPath,
    InvalidCast(ast::Type),
    InvalidSortKey(mir::Expression),
}

impl From<mir::schema::Error> for Error {
    fn from(value: mir::schema::Error) -> Self {
        Error::SchemaChecking(value)
    }
}

impl From<mir::Error> for Error {
    fn from(value: mir::Error) -> Self {
        match value {
            mir::Error::InvalidType(t) => Error::InvalidCast(t),
        }
    }
}

impl UserError for Error {
    fn code(&self) -> u32 {
        match self {
            Error::NonStarStandardSelectBody => 3002,
            Error::ArrayDatasourceMustBeLiteral => 3004,
            Error::DistinctSelect => 3005,
            Error::DistinctUnion => 3006,
            Error::NoSuchDatasource(_) => 3007,
            Error::FieldNotFound(_, _) => 3008,
            Error::AmbiguousField(_) => 3009,
            Error::StarInNonCount => 3010,
            Error::AggregationInPlaceOfScalar(_) => 3011,
            Error::ScalarInPlaceOfAggregation(_) => 3012,
            Error::NonAggregationInPlaceOfAggregation(_) => 3013,
            Error::AggregationFunctionMustHaveOneArgument => 3014,
            Error::DistinctScalarFunction => 3015,
            Error::DerivedDatasourceOverlappingKeys(_, _, _, _) => 3016,
            Error::SchemaChecking(e) => e.code(),
            Error::NoOuterJoinCondition => 3019,
            Error::DuplicateKey(_) => 3020,
            Error::InvalidSubqueryDegree => 3022,
            Error::DuplicateDocumentKey(_) => 3023,
            Error::DuplicateFlattenOption(_) => 3024,
            Error::CannotEnumerateAllFieldPaths(_) => 3025,
            Error::PolymorphicObjectSchema(_) => 3026,
            Error::DuplicateUnwindOption(_) => 3027,
            Error::NoUnwindPath => 3028,
            Error::InvalidUnwindPath => 3029,
            Error::InvalidCast(_) => 3030,
            Error::InvalidSortKey(_) => 3034,
        }
    }

    fn user_message(&self) -> Option<String> {
        match self {
            Error::NonStarStandardSelectBody => None,
            Error::ArrayDatasourceMustBeLiteral => None,
            Error::DistinctSelect => None,
            Error::DistinctUnion => None,
            Error::NoSuchDatasource(_) => None,
            Error::FieldNotFound(field, found_fields) => {
                if let Some(possible_fields) = found_fields {
                    let suggestions = generate_suggestion(field, possible_fields);
                    match suggestions {
                        Ok(suggested_fields) => {
                            if suggested_fields.is_empty() {
                                Some(format!("Field `{}` not found.", field))
                            } else {
                                Some(format!(
                                    "Field `{}` not found. Did you mean: {}",
                                    field,
                                    suggested_fields
                                        .iter()
                                        .map(|x| x.to_string())
                                        .collect::<Vec<String>>()
                                        .join(", ")
                                ))
                            }
                        }
                        Err(e) => Some(e.to_string()),
                    }
                } else {
                    Some(format!("Field `{}` not found.", field))
                }
            }
            Error::AmbiguousField(field) => Some(format!(
                "Field `{}` exists in multiple datasources and is ambiguous. Please qualify.",
                field
            )),
            Error::StarInNonCount => None,
            Error::AggregationInPlaceOfScalar(_) => None,
            Error::ScalarInPlaceOfAggregation(_) => None,
            Error::NonAggregationInPlaceOfAggregation(_) => None,
            Error::AggregationFunctionMustHaveOneArgument => None,
            Error::DistinctScalarFunction => None,
            Error::DerivedDatasourceOverlappingKeys(s1, s2, derived_name, _) => {
                let all_s1_keys: HashSet<String> = s1.keys().into_iter().collect();
                let all_s2_keys: HashSet<String> = s2.keys().into_iter().collect();
                let mut overlapping_keys = all_s1_keys
                    .intersection(&all_s2_keys)
                    .collect::<Vec<&String>>();

                overlapping_keys.sort();

                Some(format!(
                    "Derived datasource `{}` has the following overlapping keys: {}",
                    derived_name,
                    overlapping_keys
                        .iter()
                        .map(|x| x.to_string())
                        .collect::<Vec<String>>()
                        .join(", ")
                ))
            }
            Error::SchemaChecking(e) => e.user_message(),
            Error::NoOuterJoinCondition => None,
            Error::DuplicateKey(_) => None,
            Error::InvalidSubqueryDegree => None,
            Error::DuplicateDocumentKey(_) => None,
            Error::DuplicateFlattenOption(_) => None,
            Error::CannotEnumerateAllFieldPaths(_) => {
                Some("Insufficient schema information.".to_string())
            }
            Error::PolymorphicObjectSchema(_) => None,
            Error::DuplicateUnwindOption(_) => None,
            Error::NoUnwindPath => None,
            Error::InvalidUnwindPath => None,
            Error::InvalidCast(_) => None,
            Error::InvalidSortKey(_) => {
                Some("expressions are not allowed in sort key field paths".to_string())
            }
        }
    }

    fn technical_message(&self) -> String {
        match self{
            Error::NonStarStandardSelectBody => "standard SELECT expressions can only contain *".to_string(),
            Error::ArrayDatasourceMustBeLiteral => "array datasource must be constant".to_string(),
            Error::DistinctSelect => "SELECT DISTINCT not allowed".to_string(),
            Error::DistinctUnion => "UNION DISTINCT not allowed".to_string(),
            Error::NoSuchDatasource(datasource_name) => format!("no such datasource: {0:?}", datasource_name),
            Error::FieldNotFound(field, _) => format!("field `{0}` cannot be resolved to any datasource", field),
            Error::AmbiguousField(field) => format!("ambiguous field `{0}`", field),
            Error::StarInNonCount => "* argument only valid in COUNT function".to_string(),
            Error::AggregationInPlaceOfScalar(func) => format!("aggregation function {0} used in scalar position", func),
            Error::ScalarInPlaceOfAggregation(func) => format!("scalar function {0} used in aggregation position", func),
            Error::NonAggregationInPlaceOfAggregation(pos) => format!("non-aggregation expression found in GROUP BY aggregation function list at position {0}", pos),
            Error::AggregationFunctionMustHaveOneArgument => "aggregation functions must have exactly one argument".to_string(),
            Error::DistinctScalarFunction => "scalar functions don't support DISTINCT".to_string(),
            Error::DerivedDatasourceOverlappingKeys(s1, s2, derived_name, sat) => format!("derived source {derived_name} {sat:?} have overlapping keys between schemata {s1:?} and {s2:?}"),
            Error::SchemaChecking(error) => error.technical_message(),
            Error::NoOuterJoinCondition => "OUTER JOINs must specify a JOIN condition".to_string(),
            Error::DuplicateKey(key) => format!("cannot create schema environment with duplicate key: {0:?}", key),
            Error::InvalidSubqueryDegree => "subquery expressions must have a degree of 1".to_string(),
            Error::DuplicateDocumentKey(key) => format!("found duplicate document key {0:?}", key),
            Error::DuplicateFlattenOption(flatten_opt) => format!("found duplicate FLATTEN option {0:?}", flatten_opt),
            Error::CannotEnumerateAllFieldPaths(schema) => format!("cannot exhaustively enumerate all field paths in schema {0:?}", schema),
            Error::PolymorphicObjectSchema(field) => format!("cannot flatten field {0:?} since it has a polymorphic object schema", field),
            Error::DuplicateUnwindOption(unwind_opt) => format!("found duplicate UNWIND option {0:?}", unwind_opt),
            Error::NoUnwindPath => "UNWIND must specify a PATH option".to_string(),
            Error::InvalidUnwindPath => "UNWIND PATH option must be an identifier".to_string(),
            Error::InvalidCast(ast_type) => format!("invalid CAST target type '{0:?}'", ast_type),
            Error::InvalidSortKey(e) =>
                format!("sort key field path must be a pure field path with no expressions in this context. found {0:?}",
                    e
                ),
        }
    }
}
