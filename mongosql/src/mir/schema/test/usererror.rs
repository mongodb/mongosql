macro_rules! test_user_error_messages {
    ($func_name:ident, input = $input:expr, expected = $expected:expr) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{
                mir::schema::{Error, ANY_SCHEMA_ADDENDUM},
                usererror::UserError,
            };

            let user_message = $input.user_message();

            if let Some(message) = user_message {
                assert_eq!($expected, message)
            }
        }
    };
}

mod schema_checking {
    use crate::{
        schema::{
            Atomic, Schema, ANY_DOCUMENT, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, NUMERIC_OR_NULLISH,
            STRING_OR_NULLISH,
        },
        set,
    };

    test_user_error_messages! {
        operation_needs_nullable_numeric_type,
        input = Error::SchemaChecking{
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        },
        expected = "Incorrect argument type for `Add`. Required: nullable numeric type. Found: string."
    }

    test_user_error_messages! {
        operation_needs_nullable_string_type,
        input = Error::SchemaChecking{
            name: "Concat",
            required: STRING_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        },
        expected = "Incorrect argument type for `Concat`. Required: nullable string. Found: int."
    }

    test_user_error_messages! {
        operation_needs_nullable_boolean_type,
        input = Error::SchemaChecking{
            name: "SearchedCase",
            required: BOOLEAN_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        },
        expected = "Incorrect argument type for `SearchedCase`. Required: nullable boolean. Found: string."
    }

    test_user_error_messages! {
        operation_needs_nullable_date_type,
        input = Error::SchemaChecking{
            name: "Second",
            required: DATE_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::Integer),
        },
        expected = "Incorrect argument type for `Second`. Required: nullable date. Found: int."
    }

    test_user_error_messages! {
        operation_uses_any_schema,
        input = Error::SchemaChecking{
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Any,
        },
        expected = "Incorrect argument type for `Add`. Required: nullable numeric type. Found: any type. An `any type` schema may indicate that schema is not set for the relevant collection or field. Please verify that the schema is set as expected."
    }

    test_user_error_messages! {
        array_datasource_has_wrong_type,
        input = Error::SchemaChecking{
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
        },
        expected = "Incorrect argument type for `array datasource items`. Required: object type. Found: int."
    }
}

mod cannot_merge_objects {
    use crate::{
        schema::{Atomic, Schema},
        set,
    };

    test_user_error_messages! {
        overlapping_single_key,
        input = Error::CannotMergeObjects(
            Schema::Document(crate::schema::Document {
                keys: crate::map! {"a".into() => Schema::Atomic(Atomic::Integer) },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
                }),
            Schema::Document(crate::schema::Document {
                keys: crate::map! {"a".into() => Schema::Atomic(Atomic::Double) },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
                }),
            crate::schema::Satisfaction::Must,
        ),
        expected = "Cannot merge objects because they have overlapping key(s): `a`"
    }

    test_user_error_messages! {
        overlapping_multiple_keys,
        input = Error::CannotMergeObjects(
            Schema::Document(crate::schema::Document {
                keys: crate::map! {"a".into() => Schema::Atomic(Atomic::Integer), "b".into() => Schema::Atomic(Atomic::String) },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
                }),
            Schema::Document(crate::schema::Document {
                keys: crate::map! {"a".into() => Schema::Atomic(Atomic::Double), "b".into() => Schema::Any },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
                }),
            crate::schema::Satisfaction::May,
        ),
        expected = "Cannot merge objects because they have overlapping key(s): `a`, `b`"
    }
    test_user_error_messages! {
        overlapping_some_keys,
        input = Error::CannotMergeObjects(
            Schema::Document(crate::schema::Document {
                keys: crate::map! {"a".into() => Schema::Atomic(Atomic::Integer), "b".into() => Schema::Atomic(Atomic::String) },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
                }),
            Schema::Document(crate::schema::Document {
                keys: crate::map! {"a".into() => Schema::Atomic(Atomic::Double), "c".into() => Schema::Any },
                required: set! {"a".into(), "c".into()},
                additional_properties: false,
                ..Default::default()
                }),
            crate::schema::Satisfaction::Must,
        ),
        expected = "Cannot merge objects because they have overlapping key(s): `a`"

    }
}

mod aggregation_argument_must_be_self_comparable {
    use crate::{
        schema::{Atomic, Schema},
        set,
    };

    test_user_error_messages! {
        max,
        input = Error::AggregationArgumentMustBeSelfComparable(
            "Max".into(),
            Schema::AnyOf(set! {
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            }),
        ),
        expected = "Cannot perform `Max` aggregation over the type `polymorphic type` as it is not comparable to itself."
    }

    test_user_error_messages! {
        max_distinct,
        input = Error::AggregationArgumentMustBeSelfComparable(
            "Max DISTINCT".into(),
            Schema::AnyOf(set! {
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            }),
        ),
        expected = "Cannot perform `Max DISTINCT` aggregation over the type `polymorphic type` as it is not comparable to itself."
    }

    test_user_error_messages! {
        aggregation_uses_any_schema,
        input = Error::AggregationArgumentMustBeSelfComparable(
            "Max".into(),
            Schema::Any
        ),
        expected = "Cannot perform `Max` aggregation over the type `any type` as it is not comparable to itself. An `any type` schema may indicate that schema is not set for the relevant collection or field. Please verify that the schema is set as expected."
    }
}

mod invalid_comparison {
    use crate::schema::{Atomic, Schema};

    test_user_error_messages! {
        invalid_comparison,
        input = Error::InvalidComparison(
            "Lte",
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        ),
        expected = "Invalid use of `Lte` due to incomparable types: `int` cannot be compared to `string`."
    }

    test_user_error_messages! {
        invalid_comparison_one_any,
        input = Error::InvalidComparison(
            "Lte",
            Schema::Atomic(Atomic::Integer),
            Schema::Any,
        ),
        expected = format!("Invalid use of `Lte` due to incomparable types: `int` cannot be compared to `any type`. {ANY_SCHEMA_ADDENDUM}")
    }

    test_user_error_messages! {
        invalid_comparison_both_any,
        input = Error::InvalidComparison(
            "Lte",
            Schema::Any,
            Schema::Any,
        ),
        expected = format!("Invalid use of `Lte` due to incomparable types: `any type` cannot be compared to `any type`. {ANY_SCHEMA_ADDENDUM}")
    }
}

mod sort_key_comparable {
    use crate::{
        schema::{Atomic, Schema},
        set,
    };

    test_user_error_messages! {
        sort_key_not_comparable_any,
        input = Error::SortKeyNotSelfComparable(0, Schema::Any),
        expected = format!("Cannot sort by key because `any type` can't be compared against itself. {ANY_SCHEMA_ADDENDUM}")
    }

    test_user_error_messages! {
        sort_key_not_comparable_other_types,
        input = Error::SortKeyNotSelfComparable(0,
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String)
            ])
        ),
        expected = "Cannot sort by key because `polymorphic type` can't be compared against itself."
    }
}

mod group_key_comparable {
    use crate::{
        schema::{Atomic, Schema},
        set,
    };

    test_user_error_messages! {
        group_key_not_comparable_any,
        input = Error::GroupKeyNotSelfComparable(0, Schema::Any),
        expected = format!("Cannot group by key because `any type` can't be compared against itself. {ANY_SCHEMA_ADDENDUM}")
    }

    test_user_error_messages! {
        group_key_not_comparable_other_types,
        input = Error::GroupKeyNotSelfComparable(
            1,
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String)
            ]),
        ),
        expected = "Cannot group by key because `polymorphic type` can't be compared against itself."

    }
}

mod access_missing_field {
    test_user_error_messages! {
        access_missing_field_no_keys,
        input = Error::AccessMissingField("foo".to_string(), None),
        expected = "Cannot access field `foo` because it could not be found."
    }

    test_user_error_messages! {
        access_missing_field_no_close_keys,
        input = Error::AccessMissingField("foo".to_string(), Some(vec!["bar".to_string(), "baz".to_string()])),
        expected = "Cannot access field `foo` because it could not be found."
    }

    test_user_error_messages! {
        access_missing_field_some_close_keys,
        input = Error::AccessMissingField("foo".to_string(), Some(vec!["bar".to_string(), "baz".to_string(), "foof".to_string(), "fo".to_string()])),
        expected = "Cannot access field `foo` because it could not be found. Did you mean: foof, fo"
    }

    test_user_error_messages! {
        access_missing_field_all_close_keys,
        input = Error::AccessMissingField("foo".to_string(), Some(vec!["foo".to_string(), "foof".to_string(), "fo".to_string()])),
        expected = "Cannot access field `foo` because it could not be found. Internal error: Unexpected edit distance of 0 found with input: foo and expected: [\"foo\", \"foof\", \"fo\"]"
    }
}
