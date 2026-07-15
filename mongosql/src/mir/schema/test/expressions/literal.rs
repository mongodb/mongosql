use crate::{
    map,
    mir::{schema::Error as mir_error, *},
    schema::{Atomic, Document, Schema},
    set, test_schema, unchecked_unique_linked_hash_map,
};

mod scalar {
    use super::*;

    test_schema!(
        literal_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::Literal(LiteralValue::Null),
    );

    test_schema!(
        literal_bool,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::Literal(LiteralValue::Boolean(true)),
    );

    test_schema!(
        literal_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = Expression::Literal(LiteralValue::String("foobar".to_string())),
    );

    test_schema!(
        literal_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = Expression::Literal(LiteralValue::Integer(5)),
    );

    test_schema!(
        literal_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = Expression::Literal(LiteralValue::Long(6)),
    );

    test_schema!(
        literal_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = Expression::Literal(LiteralValue::Double(7.0)),
    );

    test_schema!(
        reference_does_not_exist_in_schema_env,
        expected_error_code = 1000,
        expected = Err(mir_error::DatasourceNotFoundInSchemaEnv(("a", 0u16).into())),
        input = Expression::Reference(("a", 0u16).into()),
    );

    test_schema!(
        reference_exists_in_schema_env,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = Expression::Reference(("a", 0u16).into()),
        schema_env = map! {("a", 0u16).into() => Schema::Atomic(Atomic::Null),},
    );

    test_schema!(
        invalid_binary_data_type,
        expected = Err(mir_error::InvalidBinaryDataType),
        input = Expression::Literal(LiteralValue::Binary(bson::Binary {
            subtype: bson::spec::BinarySubtype::UuidOld,
            bytes: vec![]
        })),
    );
}

mod array {
    use super::*;

    test_schema!(
        empty,
        expected = Ok(Schema::Array(Box::new(Schema::Unsat))),
        input = Expression::Array(vec![].into()),
    );

    test_schema!(
        with_single_atomic_item_results_in_array_of_atomic,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
        input = Expression::Array(vec![Expression::Literal(LiteralValue::Null)].into()),
    );

    test_schema!(
        with_multiple_atomic_items_results_in_array_of_atomic,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
        input = Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null)
            ]
            .into()
        ),
    );

    test_schema!(
        missing_items_upconverted_to_null,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
        input = Expression::Array(vec![Expression::Reference(("a", 0u16).into()),].into()),
        schema_env = map! {("a", 0u16).into() => Schema::Missing,},
    );

    test_schema!(
        document_item_with_possibly_missing_field_preserved,
        expected = Ok(Schema::Array(Box::new(Schema::Document(Document {
            keys: map! {
                "bar".into() => Schema::Atomic(Atomic::String),
            },
            required: set! {},
            additional_properties: false,
            ..Default::default()
        })))),
        input = Expression::Array(vec![Expression::Reference(("doc", 0u16).into())].into()),
        schema_env = map! {
            ("doc", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::String),
                },
                required: set! {},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_schema!(
        with_unsimplified_nested_any_of_items_results_in_array_of_simplified_schema,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ])))),
        input = Expression::Array(vec![Expression::Reference(("a", 0u16).into())].into()),
        schema_env = map! {
            ("a", 0u16).into() => Schema::AnyOf(set![
                Schema::AnyOf(set![Schema::Missing, Schema::Atomic(Atomic::Integer)]),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        nested_arrays,
        expected = Ok(Schema::Array(Box::new(Schema::Array(Box::new(
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Null),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Double),
            ])
        ))))),
        input = Expression::Array(
            vec![
                Expression::Array(vec![Expression::Reference(("a", 0u16).into()),].into()),
                Expression::Array(vec![Expression::Reference(("a", 0u16).into()),].into()),
            ]
            .into()
        ),
        schema_env = map! {
            ("a", 0u16).into() => Schema::AnyOf(set![
                Schema::AnyOf(set![Schema::Missing, Schema::Atomic(Atomic::Integer)]),
                Schema::Atomic(Atomic::Double),
            ]),
        },
    );

    test_schema!(
        with_polymorphic_items,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::String),
        ])))),
        input = Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::String("hello".to_string())),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::String("world".to_string())),
            ]
            .into()
        ),
    );
}

mod document {
    use super::*;

    test_schema!(
        document_literal_empty,
        expected = Ok(Schema::Document(Document {
            keys: map! {},
            required: set! {},
            additional_properties: false,
            ..Default::default()
        })),
        input = Expression::Document(unchecked_unique_linked_hash_map! {}.into()),
    );

    test_schema!(
        document_literal_all_required,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "b".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::Atomic(Atomic::Null),
                "d".to_string() => Schema::Atomic(Atomic::Long),
            },
            required: set! {
                "a".to_string(),
                "b".to_string(),
                "c".to_string(),
                "d".to_string(),
            },
            additional_properties: false,
            ..Default::default()
        })),
        input = Expression::Document(
            unchecked_unique_linked_hash_map! {
                "a".to_string() => Expression::Literal(LiteralValue::String("Hello".to_string())),
                "b".to_string() => Expression::Literal(LiteralValue::String("World".to_string())),
                "c".to_string() => Expression::Literal(LiteralValue::Null),
                "d".to_string() => Expression::Literal(LiteralValue::Long(42)),
            }
            .into()
        ),
    );

    test_schema!(
        document_literal_some_keys_may_or_must_satisfy_missing,
        expected = Ok(Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::String),
                "c".to_string() => Schema::Atomic(Atomic::Null),
                "d".to_string() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]),
            },
            required: set! {
                "a".to_string(),
                "c".to_string(),
            },
            additional_properties: false,
            ..Default::default()
        })),
        input = Expression::Document(
            unchecked_unique_linked_hash_map! {
                "a".to_string() => Expression::Literal(LiteralValue::String("Hello".to_string())),
                "b".to_string() => Expression::Reference(("b", 0u16).into()),
                "c".to_string() => Expression::Literal(LiteralValue::Null),
                "d".to_string() => Expression::Reference(("a", 0u16).into()),
            }
            .into()
        ),
        schema_env = map! {
            ("a", 0u16).into() => Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing]),
            ("b", 0u16).into() => Schema::Missing,
        },
    );
}
