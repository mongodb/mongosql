use crate::schema_derivation::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Expression;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};
use std::collections::BTreeMap;

mod misc_ops {
    use super::*;
    test_derive_schema!(
        constant_integral,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long)
        ))),
        input = r#"{ "$documentNumber": { } }"#
    );
}

mod window_ops {
    use super::*;
    test_derive_schema!(
        window_func_decimal,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$derivative": { "input": "$foo", "unit": "hour" } }"#,
        ref_schema = Schema::Atomic(Atomic::Decimal)
    );

    test_derive_schema!(
        window_func_double_by_default,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$derivative": { "input": 123, "unit": "hour" } }"#
    );

    test_derive_schema!(
        window_func_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$derivative": { "input": null, "unit": "hour" } }"#
    );

    test_derive_schema!(
        window_func_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$derivative": { "input": "$foo", "unit": "hour" } }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Missing, Schema::Atomic(Atomic::Integer)))
    );
}
mod numeric_ops {
    use super::*;

    test_derive_schema!(
        median_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{ "$median": { "input": 123, "method": "approximate" } }"#
    );

    test_derive_schema!(
        median_double_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$median": { "input": "$foo", "method": "approximate" } }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        median_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$median": { "input": "$foo", "method": "approximate" } }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Missing))
    );

    test_derive_schema!(
        median_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$median": { "input": null, "method": "approximate" } }"#
    );
}

mod string_ops {
    use super::*;
    test_derive_schema!(
        regex_find,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Document(Document {
                keys: map! {
                    "match".to_string() => Schema::Atomic(Atomic::String),
                    "idx".to_string() => Schema::Atomic(Atomic::Integer),
                    "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
                },
                required: set!(),
                ..Default::default()
            })
        ))),
        input = r#"{ "$regexFind": { "input": "$category", "regex": "/cafe/" }  }"#
    );

    test_derive_schema!(
        regex_find_all,
        expected = Ok(Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "match".to_string() => Schema::Atomic(Atomic::String),
                "idx".to_string() => Schema::Atomic(Atomic::Integer),
                "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
            },
            required: set!(),
            ..Default::default()
        })))),
        input = r#"{ "$regexFindAll": { "input": "$category", "regex": "/cafe/" }  }"#
    );

    test_derive_schema!(
        trim_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{ "$trim": { "input": "hi friend", "chars": "hi" }  }"#
    );

    test_derive_schema!(
        trim_input_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$trim": { "input": null, "chars": "hi" }  }"#
    );

    test_derive_schema!(
        trim_chars_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$trim": { "input": "hi", "chars": null } }"#
    );

    test_derive_schema!(
        trim_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$trim": { "input": "$foo", "chars": "hi" }  }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        trim_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$trim": { "input": "$foo", "chars": "hi" }  }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Missing))
    );
}

mod date_operators {
    use super::*;
    test_derive_schema!(
        date_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$hour": null }"#
    );

    test_derive_schema!(
        date_integer,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{ "$hour": {"$date": {"$numberLong": "123"}} }"#
    );

    test_derive_schema!(
        date_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
        ))),
        input = r#"{ "$hour": "$foo" }"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Date),
        ))
    );

    test_derive_schema!(
        date_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
        ))),
        input = r#"{ "$hour": "$foo" }"#,
        ref_schema = Schema::AnyOf(set!(Schema::Missing, Schema::Atomic(Atomic::Date),))
    );

    test_derive_schema!(
        date_timezone_specified_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$hour": {"date": {"$date": {"$numberLong": "123"}}, "timezone": null }}"#
    );

    test_derive_schema!(
        date_from_string_simple,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$dateFromString": {"dateString": "hello" }}"#
    );

    test_derive_schema!(
        date_from_string_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$dateFromString": {"dateString": null }}"#
    );

    test_derive_schema!(
        date_from_string_null_with_on_null,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{ "$dateFromString": {"dateString": null, "onNull": 1 }}"#
    );

    test_derive_schema!(
        date_from_string_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$dateFromString": {"dateString": "$foo" }}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_from_string_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{ "$dateFromString": {"dateString": "$foo" }}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Missing))
    );

    test_derive_schema!(
        date_from_parts_simple,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": 1 }}"#
    );

    test_derive_schema!(
        date_from_parts_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ))
    );

    test_derive_schema!(
        date_from_parts_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Missing))
    );

    test_derive_schema!(
        date_from_parts_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{ "$dateFromParts": {"year": 2022, "month": 1, "day": 3, "hour": null, "minute": 2, "second": 4}}"#
    );

    test_derive_schema!(
        date_from_string_with_on_error,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Integer)
        ),)),
        input = r#"{ "$dateFromString": {"dateString": "hello", "onError": 1 }}"#
    );

    test_derive_schema!(
        date_from_string_not_null_ignores_on_null,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{ "$dateFromString": {"dateString": "hello", "onNull": 1 }}"#
    );

    test_derive_schema!(
        date_from_string_nullish_fully_specified,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ))),
        input = r#"{ "$dateFromString": {"dateString": "$foo", "onError": 1, "onNull": "world" }}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_to_parts_standard,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "year".to_string() => Schema::Atomic(Atomic::Integer),
                "month".to_string() => Schema::Atomic(Atomic::Integer),
                "day".to_string() => Schema::Atomic(Atomic::Integer),
                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                "second".to_string() => Schema::Atomic(Atomic::Integer),
                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(
                "year".to_string(),
                "month".to_string(),
                "day".to_string(),
                "hour".to_string(),
                "minute".to_string(),
                "second".to_string(),
                "millisecond".to_string()
            ),
            ..Default::default()
        }),),
        input = r#"{ "$dateToParts": {"date": {"$date": {"$numberLong": "123"}} }}"#
    );

    test_derive_schema!(
        date_to_parts_iso8601,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "isoDayOfWeek".to_string() => Schema::Atomic(Atomic::Integer),
                "isoWeek".to_string() => Schema::Atomic(Atomic::Integer),
                "isoWeekYear".to_string() => Schema::Atomic(Atomic::Integer),
                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                "second".to_string() => Schema::Atomic(Atomic::Integer),
                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set!(
                "isoDayOfWeek".to_string(),
                "isoWeek".to_string(),
                "isoWeekYear".to_string(),
                "hour".to_string(),
                "minute".to_string(),
                "second".to_string(),
                "millisecond".to_string()
            ),
            ..Default::default()
        }),),
        input = r#"{"$dateToParts": {"date":{"$date": {"$numberLong": "123"}}, "iso8601": true }}"#
    );

    test_derive_schema!(
        date_to_parts_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateToParts": {"date": null, "iso8601": true }}"#
    );

    test_derive_schema!(
        date_to_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$dateToString": {"date": "$foo"}}"#,
        ref_schema = Schema::Atomic(Atomic::Date)
    );

    test_derive_schema!(
        date_to_string_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateToString": {"date": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_to_string_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateToString": {"date": "$foo"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_schema!(
        date_to_string_not_null_ignores_on_null,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$dateToString": {"date": "$foo", "onNull": 1}}"#,
        ref_schema = Schema::Atomic(Atomic::Date)
    );

    test_derive_schema!(
        date_to_string_nullish_fully_specified,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Integer),
        ))),
        input = r#"{"$dateToString": {"date": "$foo", "format": "hi", "timezone": "America/New_York", "onNull": 1}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_addition_all_args,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{"$dateAdd": {"startDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#
    );

    test_derive_schema!(
        date_addition_one_arg_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateAdd": {"startDate": null, "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#
    );

    test_derive_schema!(
        date_addition_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateAdd": {"startDate": "$foo", "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_addition_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateAdd": {"startDate": "$foo", "unit": "hour", "amount": "1", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_schema!(
        date_trunc_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateTrunc": {"date": "$foo", "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_trunc_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateTrunc": {"date": "$foo", "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_schema!(
        date_trunc_all_args,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{"$dateTrunc": {"date": {"$date": {"$numberLong": "123"}}, "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );

    test_derive_schema!(
        date_trunc_one_arg_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateTrunc": {"date": null, "unit": "hour", "binSize": "1", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );

    test_derive_schema!(
        date_diff_one_arg_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateDiff": {"startDate": "$foo", "endDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))
    );

    test_derive_schema!(
        date_diff_one_arg_possibly_missing,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Null)
        ))),
        input = r#"{"$dateDiff": {"startDate": "$foo", "endDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#,
        ref_schema = Schema::AnyOf(set!(Schema::Atomic(Atomic::Date), Schema::Missing))
    );

    test_derive_schema!(
        date_diff_all_args,
        expected = Ok(Schema::Atomic(Atomic::Date)),
        input = r#"{"$dateDiff": {"startDate": {"$date": {"$numberLong": "121"}}, "endDate": {"$date": {"$numberLong": "123"}}, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );

    test_derive_schema!(
        date_diff_one_arg_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$dateDiff": {"startDate": {"$date": {"$numberLong": "123"}}, "endDate": null, "unit": "hour", "startOfWeek": "mon", "timezone": "America/New_York"}}"#
    );
}
