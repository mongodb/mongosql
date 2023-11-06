use crate::{map, mir::*, schema::*, set, unchecked_unique_linked_hash_map};
use lazy_static::lazy_static;

mod expressions;
mod stages;
mod usererror;

pub(crate) fn test_document_a() -> Expression {
    Expression::Document(
        unchecked_unique_linked_hash_map! {
            "a".into() => Expression::Literal(LiteralValue::Integer(1))
        }
        .into(),
    )
}

pub(crate) fn test_document_b() -> Expression {
    Expression::Document(
        unchecked_unique_linked_hash_map! {
            "b".into() => Expression::Literal(LiteralValue::Integer(1))
        }
        .into(),
    )
}

pub(crate) fn test_document_c() -> Expression {
    Expression::Document(
        unchecked_unique_linked_hash_map! {
            "c".into() => Expression::Literal(LiteralValue::Integer(1))
        }
        .into(),
    )
}

lazy_static! {
    pub static ref TEST_DOCUMENT_SCHEMA_A: Schema = Schema::Document(Document {
        keys: map! {
            "a".into() => Schema::Atomic(Atomic::Integer),
        },
        required: set! {"a".into()},
        additional_properties: false,
    });
    pub static ref TEST_DOCUMENT_SCHEMA_B: Schema = Schema::Document(Document {
        keys: map! {
            "b".into() => Schema::Atomic(Atomic::Integer),
        },
        required: set! {"b".into()},
        additional_properties: false,
    });
    pub static ref TEST_DOCUMENT_SCHEMA_C: Schema = Schema::Document(Document {
        keys: map! {
            "c".into() => Schema::Atomic(Atomic::Integer),
        },
        required: set! {"c".into()},
        additional_properties: false,
    });
    pub static ref TEST_DOCUMENT_SCHEMA_S: Schema = Schema::Document(Document {
        keys: map! {
            "s".into() => Schema::Atomic(Atomic::String),
        },
        required: set! {"s".into()},
        additional_properties: false,
    });
}

#[macro_export]
macro_rules! test_schema {
    ($func_name:ident, $(expected_error_code = $expected_error_code:literal,)? $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? input = $input:expr, $(schema_env = $schema_env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports, clippy::redundant_pattern_matching)]
            use $crate::{mir::{FieldPath, schema::SchemaInferenceState}, SchemaEnvironment, SchemaCheckingMode, CachedSchema, catalog::Catalog, usererror::UserError};

            let input = $input;

            #[allow(unused_mut, unused_assignments)]
            let mut schema_env = SchemaEnvironment::default();
            $(schema_env = $schema_env;)?

            #[allow(unused_mut, unused_assignments)]
            let mut catalog = Catalog::default();
            $(catalog = $catalog;)?

            #[allow(unused_mut, unused_assignments)]
            let mut schema_checking_mode = SchemaCheckingMode::Strict;
            $(schema_checking_mode = $schema_checking_mode;)?

            let state = SchemaInferenceState::new(0u16, schema_env, &catalog, schema_checking_mode);
            let actual = input.schema(&state);

            $(assert!(matches!(actual, $expected_pat));)?
            $(assert_eq!($expected, actual);)?

            #[allow(unused_variables)]
             if let Err(e) = actual {
                 $(assert_eq!($expected_error_code, e.code()))?
             }
        }
    };
}

macro_rules! test_retain {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::mir::schema::retain;
            let expected = $expected;
            let actual = retain(&$input);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_max_numeric {
    ($func_name:ident, expected = $expected:expr, input1 = $input1:expr, input2 = $input2:expr) => {
        #[test]
        fn $func_name() {
            use crate::mir::schema::max_numeric;
            let expected = $expected;
            let actual = max_numeric(&$input1, &$input2);
            assert_eq!(expected, actual);
        }
    };
}

mod arithmetic_retain {
    use crate::{schema::*, set};

    test_retain!(
        atomics_retain_dominant,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Integer),
        ]
    );

    test_retain!(
        atomic_anyof_retains_any_gte_atomic,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::Atomic(Atomic::Double),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ])
        ]
    );

    test_retain!(
        anyof_atomic_retains_any_gte_atomic,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ]),
            Schema::Atomic(Atomic::Long),
        ]
    );

    test_retain!(
        anyof_anyof_retains_dominant_result_of_all_possible_pairs,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ]),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        ]
    );

    test_retain!(
        anyof_atomic_anyof_retains_dominant_result_of_all_possible_pairs,
        expected = Ok(Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
        ])),
        input = set![
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Decimal),
            ]),
            Schema::Atomic(Atomic::Double),
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
            ]),
        ]
    );
}

mod max_numeric {
    use crate::schema::*;
    test_max_numeric!(
        integer_returned,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input1 = Schema::Atomic(Atomic::Integer),
        input2 = Schema::Atomic(Atomic::Integer)
    );

    test_max_numeric!(
        long_takes_priority_over_integer,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input1 = Schema::Atomic(Atomic::Long),
        input2 = Schema::Atomic(Atomic::Integer)
    );

    test_max_numeric!(
        double_takes_priority_over_long,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input1 = Schema::Atomic(Atomic::Long),
        input2 = Schema::Atomic(Atomic::Double)
    );

    test_max_numeric!(
        double_takes_priority_over_integer,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input1 = Schema::Atomic(Atomic::Integer),
        input2 = Schema::Atomic(Atomic::Double)
    );

    test_max_numeric!(
        decimal_takes_priority_over_double,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input1 = Schema::Atomic(Atomic::Decimal),
        input2 = Schema::Atomic(Atomic::Double)
    );

    test_max_numeric!(
        decimal_takes_priority_over_long,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input1 = Schema::Atomic(Atomic::Decimal),
        input2 = Schema::Atomic(Atomic::Long)
    );

    test_max_numeric!(
        decimal_takes_priority_over_integer,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input1 = Schema::Atomic(Atomic::Decimal),
        input2 = Schema::Atomic(Atomic::Integer)
    );
}
