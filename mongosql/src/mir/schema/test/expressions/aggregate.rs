use crate::{
    map,
    mir::{schema::Error as mir_error, *},
    schema::{Atomic, Document, Schema, ANY_DOCUMENT, NUMERIC_OR_NULLISH},
    set, test_schema,
};

mod add_to_array {
    use super::*;

    test_schema!(
        add_to_array_has_array_schema,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])))),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::AddToArray,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        add_to_set_has_array_schema,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])))),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::AddToArray,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
}

mod avg {
    use super::*;

    test_schema!(
        arg_to_avg_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Avg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );

    test_schema!(
        avg_of_int_and_long_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        avg_of_int_and_decimal_is_double_and_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        avg_of_decimal_and_missing_is_decimal_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Missing,
        ])},
    );

    test_schema!(
        avg_of_long_and_null_is_double_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null),
        ])},
    );

    test_schema!(
        avg_of_decimal_long_and_null_is_decimal_double_and_null,
        expected = Ok(Schema::AnyOf(set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Decimal),
                Schema::Atomic(Atomic::Double)
            ],),
            Schema::Atomic(Atomic::Null),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Null),
        ])},
    );

    test_schema!(
        avg_of_integer_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
        ])},
    );

    test_schema!(
        avg_of_decimal_is_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Avg,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
}

mod count {
    use super::*;

    test_schema!(
        distinct_count_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Count DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Count,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        count_star_is_int,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ])),
        input = AggregationExpr::CountStar(false),
    );

    test_schema!(
        count_expr_is_int,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Count,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );
}

mod first {
    use super::*;

    test_schema!(
        distinct_first_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "First DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::First,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        first_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::First,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
}

mod last {
    use super::*;

    test_schema!(
        distinct_last_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Last DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Last,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        last_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Last,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
}

mod max {
    use super::*;

    test_schema!(
        max_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Max".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Max,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        distinct_max_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Max DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Max,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        max_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Max,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
}

mod merge_documents {
    use super::*;

    test_schema!(
        arg_to_mergedocuments_must_be_document,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "MergeDocuments",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::MergeDocuments,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );

    test_schema!(
        mergedocuments_returns_field_schema,
        expected = Ok(Schema::Document(Document {
            keys: map! {"foo".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: true,
        })),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::MergeDocuments,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::Document(Document {
            keys: map!{"foo".into() => Schema::Atomic(Atomic::Integer)},
            required: set!{},
            additional_properties: true,
        })},
    );
}

mod min {
    use super::*;

    test_schema!(
        min_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Min".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Min,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        distinct_min_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Min DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Min,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        min_has_field_schema,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Min,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );
}

mod stddev_pop {
    use super::*;

    test_schema!(
        distinct_stddevpop_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "StddevPop DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        arg_to_stddev_pop_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "StddevPop",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );

    test_schema!(
        stddev_pop_of_integer_and_long_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double),),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        stddev_pop_of_integer_and_decimal_is_double_and_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        stddev_pop_of_decimal_is_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevPop,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
}

mod stddev_samp {
    use super::*;

    test_schema!(
        distinct_stddevsamp_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "StddevSamp DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        arg_to_stddev_samp_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "StddevSamp",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );

    test_schema!(
        stddev_samp_of_integer_and_long_is_double,
        expected = Ok(Schema::Atomic(Atomic::Double),),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        stddev_samp_of_integer_and_decimal_is_double_and_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );

    test_schema!(
        stddev_samp_of_decimal_is_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::StddevSamp,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
}

mod sum {
    use super::*;

    test_schema!(
        distinct_sum_args_must_be_comparable,
        expected_error_code = 1003,
        expected = Err(mir_error::AggregationArgumentMustBeSelfComparable(
            "Sum DISTINCT".into(),
            ANY_DOCUMENT.clone()
        )),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: true,
        }),
        schema_env = map! {("bar", 0u16).into() => ANY_DOCUMENT.clone()},
    );

    test_schema!(
        arg_to_sum_must_be_numeric,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Sum",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            ]),
        }),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ])},
    );

    test_schema!(
        sum_of_int_and_long_is_int_long,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ])},
    );

    test_schema!(
        sum_of_int_and_double_is_int_double,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])},
    );

    test_schema!(
        sum_of_int_and_decimal_is_int_decimal,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = AggregationExpr::Function(AggregationFunctionApplication {
            function: AggregationFunction::Sum,
            arg: Box::new(Expression::Reference(("bar", 0u16).into())),
            distinct: false,
        }),
        schema_env = map! {("bar", 0u16).into() => Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ])},
    );
}
