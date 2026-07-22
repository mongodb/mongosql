use chrono::Utc;
use lazy_static::lazy_static;
use std::str::FromStr;

use crate::{
    mir::{binding_tuple::DatasourceName::Bottom, definitions::*, schema::SchemaCache},
    unchecked_unique_linked_hash_map,
};

macro_rules! test_constant_fold {
    ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr, $(schema_env = $schema_env:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{
                catalog::Catalog,
                map,
                mir::{
                    optimizer::constant_folding::ConstantFoldingOptimizer,
                    schema::{SchemaCheckingMode, SchemaInferenceState},
                },
                schema::SchemaEnvironment,
            };
            let input = $input;
            let expected = $expected;

            #[allow(unused_mut, unused_assignments)]
            let mut schema_env = SchemaEnvironment::default();
            $(schema_env = $schema_env;)?

            let (actual, actual_changed) = ConstantFoldingOptimizer::fold_constants(
                input,
                &SchemaInferenceState::new(
                    0,
                    schema_env,
                    &Catalog::default(),
                    map! {},
                    SchemaCheckingMode::Relaxed,
                ),
            );
            assert_eq!($expected_changed, actual_changed);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_constant_fold_no_op {
    ($func_name:ident, $input:expr $(, schema_env = $schema_env:expr,)?) => {
        test_constant_fold! { $func_name, expected = $input, expected_changed = false, input = $input, $(schema_env = $schema_env,)? }
    };
}

lazy_static! {
    static ref OID: bson::oid::ObjectId =
        bson::oid::ObjectId::parse_str("507f1f77bcf86cd799439011").unwrap();
    static ref DATE_WITH_MS: bson::DateTime = {
        let chrono_dt: chrono::DateTime<Utc> = "2014-01-01T08:15:39.736Z".parse().unwrap();
        chrono_dt.into()
    };
    static ref DATE_WITHOUT_MS: bson::DateTime = {
        let chrono_dt: chrono::DateTime<Utc> = "2014-01-01T08:15:39Z".parse().unwrap();
        chrono_dt.into()
    };
    static ref DATE_WITH_TZ_OFFSET: bson::DateTime = {
        let chrono_dt: chrono::DateTime<Utc> = "2018-03-03T12:00:00+0500".parse().unwrap();
        chrono_dt.into()
    };
    static ref DATE_WITHOUT_TIME: bson::DateTime = {
        let chrono_dt: chrono::DateTime<Utc> = "2014-01-01T00:00:00Z".parse().unwrap();
        chrono_dt.into()
    };
    static ref INT_DEC: bson::Decimal128 = bson::Decimal128::from_str("3").unwrap();
    static ref DOUBLE_DEC: bson::Decimal128 =
        bson::Decimal128::from_str(&format!("{}", std::f64::consts::PI)).unwrap();
    static ref STRING_DEC: bson::Decimal128 = bson::Decimal128::from_str("3.14159265").unwrap();
}

fn test_source() -> Stage {
    Stage::Collection(Collection {
        db: "test".into(),
        collection: "foo".into(),
        cache: SchemaCache::new(),
    })
}

test_constant_fold_no_op!(
    literal,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(1))],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    or_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Boolean(false))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    and_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Boolean(true))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    true_and_nulls_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    null_and_null_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    true_and_nulls_and_ref_is_null_and_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("foo", 1u16).into()),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("foo", 1u16).into())
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    false_or_nulls_or_ref_is_null_or_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("foo", 1u16).into()),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("foo", 1u16).into())
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
#[test]
fn null_or_null_is_null() {
    use crate::{
        catalog::Catalog,
        map,
        mir::{
            optimizer::constant_folding::ConstantFoldingOptimizer,
            schema::{SchemaCheckingMode, SchemaInferenceState},
        },
        schema::SchemaEnvironment,
    };
    let input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    });
    let expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    });
    let (actual, actual_changed) = ConstantFoldingOptimizer::fold_constants(
        input,
        &SchemaInferenceState::new(
            0,
            SchemaEnvironment::default(),
            &Catalog::default(),
            map! {},
            SchemaCheckingMode::Relaxed,
        ),
    );
    assert!(actual_changed);
    assert_eq!(expected, actual);
}
test_constant_fold!(
    false_or_nulls_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Boolean(false)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    or_with_true_literal_is_true,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("foo", 1u16).into())
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    or_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    and_with_false_literal_is_false,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Literal(LiteralValue::Boolean(true)),
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
                Expression::Reference(("foo", 1u16).into())
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    and_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    null_or_ref_reorder_does_not_count_as_change,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Or,
            vec![
                Expression::Reference(("foo", 1u16).into()),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    })
);
test_constant_fold_no_op!(
    null_and_ref_reorder_does_not_count_as_change,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![
                Expression::Reference(("foo", 1u16).into()),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    add_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(3))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    add_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(0))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    add_to_zero_is_zero,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Long(0))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(-1)),
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Long(-1)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    add_constant_ref_is_constant_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Double(3.0)),
                Expression::Reference(("a", 0u16).into()),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Double(1.0)),
                Expression::Literal(LiteralValue::Double(1.0)),
                Expression::Literal(LiteralValue::Double(1.0)),
                Expression::Reference(("a", 0u16).into()),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    add_zero_ref_is_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Reference(("a", 0u16).into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::Literal(LiteralValue::Integer(-1)),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Reference(("a", 0u16).into()),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    mul_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Long(8))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Long(2)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    mul_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(1))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    mul_to_one_is_one,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Double(1.0))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Double(0.5)),
                Expression::Literal(LiteralValue::Double(2.0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    mul_one_ref_is_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Reference(("a", 0u16).into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Double(2.0)),
                Expression::Literal(LiteralValue::Double(0.5)),
                Expression::Reference(("a", 0u16).into()),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    arithmetic_null_arg,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Add,
            args: vec![
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Mul,
                    vec![
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Long(2)),
                        Expression::Literal(LiteralValue::Double(2.0)),
                        Expression::Literal(LiteralValue::Null),
                    ],
                )),
                Expression::Reference(("a", 0u16).into()),
            ],
            is_nullable: true,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    add_different_num_types,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Long(4)),
                Expression::Literal(LiteralValue::Double(6.0))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Add,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Long(2)),
                Expression::Literal(LiteralValue::Double(3.0)),
                Expression::Literal(LiteralValue::Double(3.0))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    mul_different_num_types,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Integer(4)),
                Expression::Literal(LiteralValue::Long(9)),
                Expression::Literal(LiteralValue::Double(16.0))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Mul,
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Long(3)),
                Expression::Literal(LiteralValue::Long(3)),
                Expression::Literal(LiteralValue::Double(4.0)),
                Expression::Literal(LiteralValue::Double(4.0))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    sub_ref_by_zero_is_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Reference(("a", 0u16).into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sub,
            vec![
                Expression::Reference(("a", 0u16).into()),
                Expression::Literal(LiteralValue::Long(0))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    sub_null_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Sub,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Long(2))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    sub_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(0)),
                Expression::Literal(LiteralValue::Long(-1)),
                Expression::Literal(LiteralValue::Double(2.0)),
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Sub,
                    vec![
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(2))
                    ],
                )),
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Sub,
                    vec![
                        Expression::Literal(LiteralValue::Long(1)),
                        Expression::Literal(LiteralValue::Long(2))
                    ],
                )),
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Sub,
                    vec![
                        Expression::Literal(LiteralValue::Double(3.0)),
                        Expression::Literal(LiteralValue::Double(1.0))
                    ],
                )),
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    div_zero_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Div,
            vec![
                Expression::Reference(("a", 0u16).into()),
                Expression::Literal(LiteralValue::Long(0))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    div_null_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Div,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Long(2))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    div_ref_by_one_is_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Reference(("a", 0u16).into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Div,
            vec![
                Expression::Reference(("a", 0u16).into()),
                Expression::Literal(LiteralValue::Long(1))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    div_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Long(-1)),
                Expression::Literal(LiteralValue::Double(2.0)),
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Div,
                    vec![
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(2))
                    ],
                )),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Div,
                    args: vec![
                        Expression::Literal(LiteralValue::Long(-2)),
                        Expression::Literal(LiteralValue::Long(2))
                    ],
                    is_nullable: false,
                }),
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Div,
                    vec![
                        Expression::Literal(LiteralValue::Double(2.0)),
                        Expression::Literal(LiteralValue::Double(1.0))
                    ],
                )),
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_less_than,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lt,
            vec![
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Boolean(true)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_less_than,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lt,
            vec![
                Expression::Literal(LiteralValue::Boolean(false)),
                Expression::Literal(LiteralValue::Boolean(false)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_less_than_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lte,
            vec![
                Expression::Literal(LiteralValue::Integer(0)),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_less_than_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lte,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_greater_than,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Long(0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_greater_than,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gt,
            vec![
                Expression::Literal(LiteralValue::Long(0)),
                Expression::Literal(LiteralValue::Long(0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_greater_than_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gte,
            vec![
                Expression::Literal(LiteralValue::Double(1.0)),
                Expression::Literal(LiteralValue::Double(1.0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_greater_than_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Gte,
            vec![
                Expression::Literal(LiteralValue::Double(0.0)),
                Expression::Literal(LiteralValue::Double(1.0)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Eq,
            vec![
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Eq,
            vec![
                Expression::Literal(LiteralValue::Long(0)),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_nequal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Neq,
            vec![
                Expression::Literal(LiteralValue::Long(0)),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_nequal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Neq,
            vec![
                Expression::Literal(LiteralValue::Long(1)),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    compare_different_datatypes,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Neq,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    compare_null_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lte,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Long(1)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_between,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(0)),
                Expression::Literal(LiteralValue::Integer(2)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_not_between,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Between,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Reference(("foo", 1u16).into())
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    fold_comparison_nested,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Lt,
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                    ],
                )),
                Expression::Literal(LiteralValue::Boolean(true)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    fold_between_args,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Reference(("foo", 1u16).into()),
                Expression::Literal(LiteralValue::Integer(0)),
                Expression::Literal(LiteralValue::Integer(3)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Between,
            args: vec![
                Expression::Reference(("foo", 1u16).into()),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(-1)),
                    ],
                    is_nullable: false,
                }),
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Add,
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                    ],
                )),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    pos_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(2))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Pos,
            vec![Expression::Literal(LiteralValue::Integer(2))],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    neg_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(-2))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Neg,
            vec![Expression::Literal(LiteralValue::Integer(2))],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    not_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![Expression::Literal(LiteralValue::Boolean(true))],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    upper_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "AABBCC".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Upper,
            vec![Expression::Literal(LiteralValue::String(
                "aaBBcC".to_string()
            )),],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    lower_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "aabbcc".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lower,
            vec![Expression::Literal(LiteralValue::String(
                "aaBBcC".to_string()
            )),],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    lower_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Lower,
            vec![Expression::Literal(LiteralValue::Null),],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    btrim_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "AABBCC".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BTrim,
            vec![
                Expression::Literal(LiteralValue::String("a".to_string())),
                Expression::Literal(LiteralValue::String("aAABBCCa".to_string()))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    ltrim_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "AABBCCa".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::LTrim,
            vec![
                Expression::Literal(LiteralValue::String("a".to_string())),
                Expression::Literal(LiteralValue::String("aAABBCCa".to_string()))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    rtrim_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "aAABBCC".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::RTrim,
            vec![
                Expression::Literal(LiteralValue::String("a".to_string())),
                Expression::Literal(LiteralValue::String("aAABBCCa".to_string()))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    btrim_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BTrim,
            vec![
                Expression::Literal(LiteralValue::String("a".to_string())),
                Expression::Literal(LiteralValue::Null)
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_nested,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "hello".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(0)),
                Expression::ScalarFunction(ScalarFunctionApplication::new(
                    ScalarFunction::Add,
                    vec![
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3)),
                    ],
                ))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_multi_codepoint_char,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "🇷🇺ááá".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("ááá🇷🇺ááá".to_string())),
                Expression::Literal(LiteralValue::Integer(6)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_negative_length,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "world".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(6)),
                Expression::Literal(LiteralValue::Integer(-1)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_negative_start_with_smaller_length,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String("".to_string()))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(-6)),
                Expression::Literal(LiteralValue::Integer(5)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_negative_start_with_larger_length,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "hello".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Substring,
            args: vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(-6)),
                Expression::Literal(LiteralValue::Integer(11)),
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_start_larger_than_string_length,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String("".to_string()))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(20)),
                Expression::Literal(LiteralValue::Integer(4)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_end_larger_than_string_length,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "world".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(6)),
                Expression::Literal(LiteralValue::Integer(20)),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    substring_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Substring,
            vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Literal(LiteralValue::Integer(6)),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    concat_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "hello world".to_string()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Literal(LiteralValue::String("hello ".to_string())),
                Expression::Literal(LiteralValue::String("world".to_string())),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    concat_with_ref,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Literal(LiteralValue::String("hello world".to_string())),
                Expression::Reference(("a", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello world2".to_string())),
                Expression::Reference(("a", 0u16).into()),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Literal(LiteralValue::String("hello ".to_string())),
                Expression::Literal(LiteralValue::String("world".to_string())),
                Expression::Reference(("a", 0u16).into()),
                Expression::Literal(LiteralValue::String("hello ".to_string())),
                Expression::Literal(LiteralValue::String("world2".to_string())),
                Expression::Reference(("a", 0u16).into()),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    concat_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String("".to_string()))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    concat_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Concat,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::String("hello".to_string())),
                Expression::Literal(LiteralValue::String("world".to_string()))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    char_length_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(11))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::CharLength,
            vec![Expression::Literal(LiteralValue::String(
                "hello world".to_string()
            ))],
        )),],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    char_length_multi_codepoint,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(14))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::CharLength,
            args: vec![Expression::Literal(LiteralValue::String(
                "ááá🇷🇺ááá".to_string()
            )),],
            is_nullable: false,
        }),],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    octet_length_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(11))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::OctetLength,
            vec![Expression::Literal(LiteralValue::String(
                "hello world".to_string()
            )),],
        )),],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    octet_length_multi_codepoint,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(26))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::OctetLength,
            args: vec![Expression::Literal(LiteralValue::String(
                "ááá🇷🇺ááá".to_string()
            )),],
            is_nullable: false,
        }),],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    bit_length_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(88))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::BitLength,
            vec![Expression::Literal(LiteralValue::String(
                "hello world".to_string()
            )),],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    bit_length_multi_codepoint,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(208))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::BitLength,
            args: vec![Expression::Literal(LiteralValue::String(
                "ááá🇷🇺ááá".to_string()
            )),],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    array_size_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(2))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(0)),
                    Expression::Literal(LiteralValue::Integer(0))
                ]
                .into()
            )],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    array_size_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(0))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Size,
            args: vec![Expression::Array(vec![].into())],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    coalesce_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    coalesce_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(0))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Integer(0)),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Integer(1))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    coalesce_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Coalesce,
            vec![
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Null),
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    merge_objects_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Document(
            unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0)),
            "b".into() => Expression::Literal(LiteralValue::Integer(0)),
            "c".into() => Expression::Literal(LiteralValue::Integer(2))}
        .into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "b".into() => Expression::Literal(LiteralValue::Integer(1))}
                .into()),
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"b".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "c".into() => Expression::Literal(LiteralValue::Integer(2))}
                .into())
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    merge_objects_reference,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "b".into() => Expression::Literal(LiteralValue::Integer(1))}
                .into()),
                Expression::Reference(("a", 0u16).into())
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    merge_objects_empty,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Document(
            unchecked_unique_linked_hash_map! {}.into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::MergeObjects,
            vec![],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    merge_objects_combine_early_docs,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "b".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "c".into() => Expression::Literal(LiteralValue::Integer(2))}
                .into()),
                Expression::Reference(("a", 0u16).into())
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::MergeObjects,
            args: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "b".into() => Expression::Literal(LiteralValue::Integer(1))}
                .into()),
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"b".into() => Expression::Literal(LiteralValue::Integer(0)),
                    "c".into() => Expression::Literal(LiteralValue::Integer(2))}
                .into()),
                Expression::Reference(("a", 0u16).into())
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    null_if_args_equal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(1))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    null_if_args_unequal,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(1))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::NullIf,
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
        ))],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    computed_field_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(2))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2))}
                .into()),
                Expression::Literal(LiteralValue::String("a".into()))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    computed_field_missing,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::ComputedFieldAccess,
            args: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2))}
                .into()),
                Expression::Literal(LiteralValue::String("b".into()))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    slice_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![Expression::Literal(LiteralValue::Integer(2))].into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(1))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Null),
                Expression::Literal(LiteralValue::Integer(1))
            ],
            is_nullable: true,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_negative_length_no_start,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(-2))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_positive_length_no_start,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(2))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_negative_start_within_array,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![Expression::Literal(LiteralValue::Integer(2))].into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(-2)),
                Expression::Literal(LiteralValue::Integer(1))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_negative_start_outside_array,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2))
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(-5)),
                Expression::Literal(LiteralValue::Integer(2))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_negative_len_longer_than_array,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(-5))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_positive_len_longer_than_array,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(5))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_positive_len_longer_than_array_no_start,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(
            vec![
                Expression::Literal(LiteralValue::Integer(1)),
                Expression::Literal(LiteralValue::Integer(2)),
                Expression::Literal(LiteralValue::Integer(3))
            ]
            .into()
        )],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(5))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_start_larger_than_array,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Array(vec![].into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(5)),
                Expression::Literal(LiteralValue::Integer(1))
            ],
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    slice_with_pos_neg_length_is_null,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Null)],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Slice,
            args: vec![
                Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(1)),
                        Expression::Literal(LiteralValue::Integer(2)),
                        Expression::Literal(LiteralValue::Integer(3))
                    ]
                    .into()
                ),
                Expression::Literal(LiteralValue::Integer(5)),
                Expression::Literal(LiteralValue::Integer(-1))
            ],
            is_nullable: false,
        }),],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_expr_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Is(IsExpr {
            expr: Expression::Literal(LiteralValue::Integer(1)).into(),
            target_type: TypeOrMissing::Type(Type::Int32),
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_expr_false,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Is(IsExpr {
            expr: Expression::Literal(LiteralValue::String("a".into())).into(),
            target_type: TypeOrMissing::Type(Type::Double),
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    is_expr_number,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Is(IsExpr {
            expr: Expression::Literal(LiteralValue::Double(1.0)).into(),
            target_type: TypeOrMissing::Number,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    is_expr_null,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Is(IsExpr {
            expr: Expression::Literal(LiteralValue::Double(1.0)).into(),
            target_type: TypeOrMissing::Type(Type::Null),
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    is_expr_nested,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Boolean(true))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Is(IsExpr {
            expr: Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Concat,
                vec![
                    Expression::Literal(LiteralValue::String("hello ".to_string())),
                    Expression::Literal(LiteralValue::String("world".to_string())),
                ],
            ))
            .into(),
            target_type: TypeOrMissing::Type(Type::String),
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    simple_case_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String("then 2".into()))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SimpleCase(SimpleCaseExpr {
            expr: Expression::Literal(LiteralValue::Integer(2)).into(),
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(3)).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(2)).into(),
                    then: Expression::Literal(LiteralValue::String("then 2".into())).into(),
                    is_nullable: false,
                }
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    simple_case_ref_ahead,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SimpleCase(SimpleCaseExpr {
            expr: Expression::Literal(LiteralValue::Integer(2)).into(),
            when_branch: vec![
                WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(2)).into(),
                    then: Expression::Literal(LiteralValue::String("then 2".into())).into(),
                    is_nullable: false,
                }
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    simple_case_prune_false,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SimpleCase(SimpleCaseExpr {
            expr: Expression::Literal(LiteralValue::Integer(2)).into(),
            when_branch: vec![WhenBranch {
                when: Expression::Reference(("a", 0u16).into()).into(),
                then: Expression::Literal(LiteralValue::String("then a".into())).into(),
                is_nullable: false,
            },],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SimpleCase(SimpleCaseExpr {
            expr: Expression::Literal(LiteralValue::Integer(2)).into(),
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(3)).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then a".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(4)).into(),
                    then: Expression::Literal(LiteralValue::String("then 4".into())).into(),
                    is_nullable: false,
                },
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    simple_case_else,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String("else".into()))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SimpleCase(SimpleCaseExpr {
            expr: Expression::Literal(LiteralValue::Integer(2)).into(),
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(3)).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(4)).into(),
                    then: Expression::Literal(LiteralValue::String("then 4".into())).into(),
                    is_nullable: false,
                }
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    simple_case_keep_branches,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SimpleCase(SimpleCaseExpr {
            expr: Expression::Reference(("a", 0u16).into()).into(),
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(4)).into(),
                    then: Expression::Literal(LiteralValue::String("then 4".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(4)).into(),
                    then: Expression::Literal(LiteralValue::String("then 4".into())).into(),
                    is_nullable: false,
                },
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    searched_case_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String(
            "then true".into()
        ))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    then: Expression::Literal(LiteralValue::String("then false".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    then: Expression::Literal(LiteralValue::String("then true".into())).into(),
                    is_nullable: false,
                }
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    searched_case_ref_ahead,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    then: Expression::Literal(LiteralValue::String("then true".into())).into(),
                    is_nullable: true,
                }
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    searched_case_prune_false,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![WhenBranch {
                when: Expression::Reference(("a", 0u16).into()).into(),
                then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                is_nullable: false,
            },],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    then: Expression::Literal(LiteralValue::String("then false".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::String("false".into())).into(),
                    then: Expression::Literal(LiteralValue::String("then false string".into()))
                        .into(),
                    is_nullable: false,
                },
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    searched_case_else,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::String("else".into()))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    then: Expression::Literal(LiteralValue::String("then false".into())).into(),
                    is_nullable: false,
                },
                WhenBranch {
                    when: Expression::Literal(LiteralValue::Integer(4)).into(),
                    then: Expression::Literal(LiteralValue::String("then 4".into())).into(),
                    is_nullable: false,
                }
            ],
            else_branch: Expression::Literal(LiteralValue::String("else".into())).into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    field_access_simple,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Literal(LiteralValue::Integer(0))],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::FieldAccess(FieldAccess {
            expr: Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0))}
            .into())
            .into(),
            field: "a".into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    field_access_reference,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Reference(("b", 0u16).into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::FieldAccess(FieldAccess {
            expr: Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Reference(("b", 0u16).into())}
            .into())
            .into(),
            field: "a".into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    field_access_nested,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::Reference(("c", 0u16).into())],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::FieldAccess(FieldAccess {
            expr: Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::FieldAccess(FieldAccess {
                    expr: Expression::Document(
                        unchecked_unique_linked_hash_map! {"b".into() => Expression::Reference(("c", 0u16).into())}.into()
                    ).into(),
                    field: "b".into(),
                    is_nullable: false,
                })}
            .into())
            .into(),
            field: "a".into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold!(
    field_access_mqlintrinstic_field_exists,
    expected = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::MqlIntrinsicFieldExistence(FieldAccess {
            expr: Expression::Reference(("ITBL", 1u16).into()).into(),
            field: "baz".into(),
            is_nullable: true,
        }),],
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::MqlIntrinsicFieldExistence(FieldAccess {
            expr: Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "ofoo".into() => Expression::FieldAccess(FieldAccess {
                        expr: Expression::Reference(("OTBL", 1u16).into()).into(),
                        field: "foo".into(),
                        is_nullable: false,
                    }),
                    "ifoo".into() => Expression::FieldAccess(FieldAccess {
                        expr: Expression::Reference(("ITBL", 1u16).into()).into(),
                        field: "foo".into(),
                        is_nullable: true,
                    }),
                    "baz".into() => Expression::FieldAccess(FieldAccess {
                        expr: Expression::Reference(("ITBL", 1u16).into()).into(),
                        field: "baz".into(),
                        is_nullable: true,
                    }),
                }
                .into()
            )
            .into(),
            field: "baz".into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    field_access_missing_field,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::FieldAccess(FieldAccess {
            expr: Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0))}
            .into())
            .into(),
            field: "b".into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold_no_op!(
    field_access_ref,
    Stage::Array(ArraySource {
        alias: "".into(),
        array: vec![Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("a", 0u16).into()).into(),
            field: "a".into(),
            is_nullable: false,
        })],
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    offset_simple,
    expected = test_source(),
    expected_changed = true,
    input = Stage::Offset(Offset {
        source: Box::new(test_source()),
        offset: 0,
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    offset_nonzero,
    Stage::Offset(Offset {
        source: Box::new(test_source()),
        offset: 1,
        cache: SchemaCache::new(),
    })
);
test_constant_fold!(
    filter_simple,
    expected = test_source(),
    expected_changed = true,
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::Literal(LiteralValue::Boolean(true)),
        cache: SchemaCache::new(),
    }),
);
test_constant_fold_no_op!(
    filter_non_literal,
    Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::Reference(("a", 0u16).into()),
        cache: SchemaCache::new(),
    })
);

test_constant_fold!(
    multi_level_field_access_from_literal_document_is_constant_folded,
    expected = Stage::Project(Project {
        is_add_fields: false,
        source: Box::new(Stage::Array(ArraySource {
            array: vec![Expression::Document(DocumentExpr {
                document: unchecked_unique_linked_hash_map! {},
            })],
            alias: "_dual".to_string(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            (Bottom, 0u16).into() => Expression::Document(DocumentExpr {
                document: unchecked_unique_linked_hash_map! {
                    "c".to_string() =>
                    Expression::Literal(LiteralValue::Integer(1),),
                },
            }),
        },
        cache: SchemaCache::new(),
    }),
    expected_changed = true,
    input = Stage::Project(Project {
        is_add_fields: false,
        source: Box::new(Stage::Array(ArraySource {
            array: vec![Expression::Document(DocumentExpr {
                document: unchecked_unique_linked_hash_map! {},
            })],
            alias: "_dual".to_string(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            (Bottom, 0u16).into() => Expression::Document(DocumentExpr {
                document: unchecked_unique_linked_hash_map! {
                    "c".to_string() => Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Document(DocumentExpr {
                                    document: unchecked_unique_linked_hash_map! {
                                        "a".to_string() => Expression::Document(DocumentExpr {
                                            document: unchecked_unique_linked_hash_map! {
                                                "b".to_string() => Expression::Document(DocumentExpr {
                                                    document: unchecked_unique_linked_hash_map! {
                                                        "c".to_string() => Expression::Literal(
                                                            LiteralValue::Integer(1),
                                                        ),
                                                    },
                                                }),
                                            },
                                        }),
                                    },
                                })),
                                field: "a".to_string(),
                                is_nullable: false,
                            })),
                            field: "b".to_string(),
                            is_nullable: false,
                        })),
                        field: "c".to_string(),
                        is_nullable: false,
                    }),
                },
            }),
        },
        cache: SchemaCache::new(),
    }),
);

mod cast {
    use super::*;

    use crate::schema::{Atomic, Schema, NULLISH};

    // Regardless of target type, casting from null literal folds to the on_null branch
    test_constant_fold!(
        from_null_literal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::String("null".into()))],
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Null).into(),
                to: Type::Int32,
                on_null: Expression::Literal(LiteralValue::String("null".into())).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into())).into(),
                is_nullable: false,
            })],
            cache: SchemaCache::new(),
        }),
    );

    // Regardless of target type, casting from an exactly nullish expr folds to the on_null
    // branch
    test_constant_fold!(
        from_exactly_nullish_expr,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::String("null".into()))],
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Reference(("bar", 0u16).into()).into(),
                to: Type::Int32,
                on_null: Expression::Literal(LiteralValue::String("null".into())).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into())).into(),
                is_nullable: false,
            })],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("bar", 0u16).into() => NULLISH.clone(),
        },
    );

    test_constant_fold!(
        valid_nested_casts,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Decimal128(*STRING_DEC))],
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("3.14159265".into())).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })
                .into(),
                to: Type::Decimal128,
                on_null: Expression::Literal(LiteralValue::Null).into(),
                on_error: Expression::Literal(LiteralValue::Null).into(),
                is_nullable: true,
            })],
            cache: SchemaCache::new(),
        }),
    );

    test_constant_fold_no_op!(
        mismatched_types,
        Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Reference(("bar", 0u16).into()).into(),
                to: Type::String,
                on_null: Expression::Literal(LiteralValue::Null).into(),
                on_error: Expression::Literal(LiteralValue::Null).into(),
                is_nullable: true,
            })],
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean),
        },
    );

    // Note that starting in server version 8.3, $convert to "array" is supported. However, the
    // MongoSQL spec does not support that conversion, so the tests here and implementation of this
    // optimization, as well as general translation of CAST "AS ARRAY", are limited.
    mod to_array {
        use super::*;

        test_constant_fold!(
            from_array,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Array(
                    vec![Expression::Literal(LiteralValue::Boolean(true))].into()
                )],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Array(
                        vec![Expression::Literal(LiteralValue::Boolean(true))].into()
                    )
                    .into(),
                    to: Type::Array,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_non_array,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String("error".into()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(0)).into(),
                    to: Type::Array,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".into())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Note that starting in server version 8.3, $convert to "object" is supported. However, the
    // MongoSQL spec does not support that conversion, so the tests here and implementation of this
    // optimization, as well as general translation of CAST "AS DOCUMENT", are limited.
    mod to_document {
        use super::*;

        test_constant_fold!(
            from_document,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(1))
                    }
                    .into()
                )],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::Literal(LiteralValue::Integer(1))
                        }
                        .into()
                    )
                    .into(),
                    to: Type::Document,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_non_document,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String("error".into()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(0)).into(),
                    to: Type::Document,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".into())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-boolean
    mod to_boolean {
        use super::*;

        test_constant_fold!(
            from_boolean_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Boolean),
            },
        );

        test_constant_fold!(
            from_array_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Array(ArrayExpr { array: vec![] }).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_bindata_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Binary(bson::Binary {
                        bytes: vec![],
                        subtype: bson::spec::BinarySubtype::Generic,
                    }))
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_true_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_false_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(false))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_date_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::DateTime(bson::DateTime::now())).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(false))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("0.0").unwrap()
                    ))
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_non_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("1.0").unwrap()
                    ))
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_document_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Document(DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {},
                    })
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(false))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(0.0)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_non_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(1.0)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_int_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(false))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(0)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_int_non_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(1)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_javascript_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::JavaScriptCode("".to_string())).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_javascript_with_scope_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::JavaScriptCodeWithScope(
                        bson::JavaScriptCodeWithScope {
                            code: "".to_string(),
                            scope: bson::Document::new(),
                        }
                    ))
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(false))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(0)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_non_zero_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(1)).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_max_key_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::MaxKey).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_min_key_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::MinKey).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_object_id_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::ObjectId(bson::oid::ObjectId::new()))
                        .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_regex_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::RegularExpression(bson::Regex {
                        pattern: "".to_string(),
                        options: "".to_string(),
                    }))
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("".to_string())).into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_timestamp_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Boolean(true))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Timestamp(bson::Timestamp {
                        time: 0,
                        increment: 0,
                    }))
                    .into(),
                    to: Type::Boolean,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-integer
    mod to_int {
        use super::*;

        test_constant_fold!(
            from_integer_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Integer),
            },
        );

        test_constant_fold!(
            from_boolean_false_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_true_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(1))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // In general, for Decimal and Double, $convert truncates. The truncated value must be
        // within the range of a 32-bit signed integer, otherwise the onError value is returned.
        test_constant_fold!(
            from_decimal_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(1))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("1.2").unwrap()
                    ))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}.9", i32::MAX).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}.9", i32::MIN).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_greater_than_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}", i32::MAX as i64 + 1i64).as_str())
                            .unwrap()
                    ))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_less_than_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}", i32::MIN as i64 - 1i64).as_str())
                            .unwrap()
                    ))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(1))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(1.2)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(i32::MAX as f64 + 0.9)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(i32::MIN as f64 - 0.9)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_greater_than_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(i32::MAX as f64 + 1.0)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_less_than_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(i32::MIN as f64 - 1.0)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(5))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(5)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(i32::MAX as i64)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(i32::MIN as i64)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_greater_than_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(i32::MAX as i64 + 1)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_less_than_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(i32::MIN as i64 - 1)).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
        test_constant_fold!(
            from_string_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(5))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("5".to_string())).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(i32::MAX.to_string())).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Integer(i32::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(i32::MIN.to_string())).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_greater_than_max_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(format!(
                        "{}",
                        i32::MAX as i64 + 1i64
                    )))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_less_than_min_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(format!(
                        "{}",
                        i32::MIN as i64 - 1i64
                    )))
                    .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_not_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("5.5".to_string())).into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_not_a_number_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("not a number".to_string()))
                        .into(),
                    to: Type::Int32,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-decimal
    mod to_decimal {
        use super::*;

        test_constant_fold!(
            from_decimal_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Decimal),
            },
        );

        test_constant_fold!(
            from_boolean_false_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(
                    bson::Decimal128::from_str("0").unwrap()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_true_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(
                    bson::Decimal128::from_str("1").unwrap()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(*DOUBLE_DEC))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(std::f64::consts::PI)).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(*INT_DEC))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(3)).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(*INT_DEC))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(3)).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_valid_string_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(*STRING_DEC))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("3.14159265".into())).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // A string that cannot be parsed as a Decimal128 statically fails, so the cast
        // folds to its on_error branch.
        test_constant_fold!(
            from_invalid_string_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String("error".into()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("hello".into())).into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".into())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_date_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Decimal128(
                    bson::Decimal128::from_str("1522039108044").unwrap()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::DateTime(bson::DateTime::from_millis(
                        1522039108044
                    )))
                    .into(),
                    to: Type::Decimal128,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-double
    mod to_double {
        use super::*;

        test_constant_fold!(
            from_double_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Double),
            },
        );

        test_constant_fold!(
            from_boolean_false_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(0.0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_true_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(1.0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(1.2))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("1.2").unwrap()
                    ))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_max_f64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(f64::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}", f64::MAX).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_min_f64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(f64::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}", f64::MIN).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_greater_than_max_f64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("1{}", f64::MAX).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_less_than_min_f64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(
                            format!("{}", f64::MIN).replacen('1', "2", 1).as_str()
                        )
                        .unwrap()
                    ))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(5.0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(5)).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(10000.0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(10000)).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // MongoDB loses precision when converting from sufficiently large long values to double.
        test_constant_fold!(
            from_long_max_i64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(i64::MAX as f64))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(i64::MAX)).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // MongoDB loses precision when converting from sufficiently large long values to double.
        test_constant_fold!(
            from_long_min_i64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(i64::MIN as f64))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(i64::MIN)).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(-5.5))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("-5.5".to_string())).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_greater_than_max_f64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(format!("1{}", f64::MAX)))
                        .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_less_than_min_f64_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(
                        format!("{}", f64::MIN).replacen('1', "2", 1)
                    ))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // f64::NAN does not equal f64::NAN, so we can't use test_constant_fold! conveniently.
        // Instead, we manually implement this test and assert that the value is folded to a
        // LiteralValue::Double(f64::NAN).
        #[test]
        fn from_string_nan_literal() {
            use crate::{
                catalog::Catalog,
                map,
                mir::{
                    optimizer::constant_folding::ConstantFoldingOptimizer,
                    schema::{SchemaCheckingMode, SchemaInferenceState},
                },
                schema::SchemaEnvironment,
            };
            let input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("NaN".to_string())).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            });

            let (actual, actual_changed) = ConstantFoldingOptimizer::fold_constants(
                input,
                &SchemaInferenceState::new(
                    0,
                    SchemaEnvironment::default(),
                    &Catalog::default(),
                    map! {},
                    SchemaCheckingMode::Relaxed,
                ),
            );
            // We should have constant-folded
            assert!(actual_changed);

            let Stage::Array(ArraySource {
                alias: _,
                array,
                cache: _,
            }) = actual
            else {
                panic!("Expected ArraySource, got {:?}", actual);
            };
            let [Expression::Literal(LiteralValue::Double(folded_val))] = array.as_slice() else {
                panic!("Expected single Literal in array, got {:?}", array);
            };
            assert!(folded_val.is_nan());
        }

        test_constant_fold!(
            from_string_inf_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(f64::INFINITY))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("inf".to_string())).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_neg_inf_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(f64::NEG_INFINITY))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("-inf".to_string())).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_scientific_notation_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(50000000000.0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("5e10".to_string())).into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_invalid_string_not_a_number_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("not a number".to_string()))
                        .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_date_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Double(1522039108044.0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::DateTime(bson::DateTime::from_millis(
                        1522039108044
                    )))
                    .into(),
                    to: Type::Double,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-long
    mod to_long {
        use super::*;

        test_constant_fold!(
            from_long_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Long),
            },
        );

        test_constant_fold!(
            from_boolean_false_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(0))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_true_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(1))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // In general, for Decimal and Double, $convert truncates. The truncated value must be
        // within the range of a 64-bit signed integer, otherwise the onError value is returned.
        test_constant_fold!(
            from_decimal_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(1))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("1.2").unwrap()
                    ))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_max_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(i64::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}.9", i64::MAX).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_min_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(i64::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}.9", i64::MIN).as_str()).unwrap()
                    ))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_greater_than_max_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}", i64::MAX as f64 + 1.0).as_str())
                            .unwrap()
                    ))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_less_than_min_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str(format!("{}", i64::MIN as f64 - 1.0).as_str())
                            .unwrap()
                    ))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(1))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(1.2)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_max_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(i64::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(i64::MAX as f64 + 0.9)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_min_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(i64::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(i64::MIN as f64 - 0.9)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_greater_than_max_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(f64::MAX)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_less_than_min_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(f64::MIN)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(5))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(5)).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(5))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("5".to_string())).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_max_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(i64::MAX))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(i64::MAX.to_string())).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_min_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(i64::MIN))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(i64::MIN.to_string())).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_greater_than_max_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(format!(
                        "{}",
                        i64::MAX as f64 + 1.0
                    )))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_less_than_min_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(format!(
                        "{}",
                        i64::MIN as f64 - 1.0
                    )))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_not_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("5.5".to_string())).into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_not_a_number_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("not a number".to_string()))
                        .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_date_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::Long(1522039108044))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::DateTime(bson::DateTime::from_millis(
                        1522039108044
                    )))
                    .into(),
                    to: Type::Int64,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-date
    mod to_datetime {
        use super::*;

        test_constant_fold!(
            from_datetime_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::Date),
            },
        );

        test_constant_fold!(
            from_double_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::DateTime(
                    bson::DateTime::from_millis(120000000000)
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(120000000000.5)).into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::DateTime(
                    bson::DateTime::from_millis(1253372036000)
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("1253372036000.50").unwrap()
                    ))
                    .into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::DateTime(
                    bson::DateTime::from_millis(1100000000000)
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(1100000000000)).into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_with_date_and_time_with_milliseconds,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITH_MS)),
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITH_MS)),
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITH_MS)),
                ],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2014-01-01T08:15:39.736Z".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2014-01-01 08:15:39.736Z".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2014-01-01 08:15:39.736".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                ],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_with_date_and_time_without_milliseconds,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITHOUT_MS)),
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITHOUT_MS)),
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITHOUT_MS)),
                ],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2014-01-01T08:15:39Z".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2014-01-01 08:15:39Z".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2014-01-01 08:15:39".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                ],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_with_date_with_timezone_offset,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITH_TZ_OFFSET)),
                    Expression::Literal(LiteralValue::DateTime(*DATE_WITH_TZ_OFFSET)),
                ],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2018-03-03T12:00:00+0500".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                    Expression::Cast(CastExpr {
                        expr: Expression::Literal(LiteralValue::String(
                            "2018-03-03 12:00:00+0500".into()
                        ))
                        .into(),
                        to: Type::Datetime,
                        on_null: Expression::Literal(LiteralValue::Null).into(),
                        on_error: Expression::Literal(LiteralValue::Null).into(),
                        is_nullable: true,
                    }),
                ],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_string_with_date_without_time,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::DateTime(
                    *DATE_WITHOUT_TIME
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String("2014-01-01".into())).into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                }),],
                cache: SchemaCache::new(),
            }),
        );

        // An unparseable date string cannot be folded and is left as a Cast, to be
        // evaluated at runtime by MongoDB.
        test_constant_fold_no_op!(
            invalid_string_to_date,
            Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(
                        "2014-01-56T08:15:39.736Z".into()
                    ))
                    .into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            })
        );

        test_constant_fold!(
            from_object_id_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::DateTime(OID.timestamp()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::ObjectId(*OID)).into(),
                    to: Type::Datetime,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-objectid
    mod to_object_id {
        use super::*;

        test_constant_fold!(
            from_object_id_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::ObjectId,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::ObjectId),
            },
        );

        test_constant_fold!(
            from_valid_string,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::ObjectId(*OID))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(
                        "507f1f77bcf86cd799439011".into()
                    ))
                    .into(),
                    to: Type::ObjectId,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        // A string that cannot be parsed as an ObjectId statically fails, so the cast
        // folds to its on_error branch.
        test_constant_fold!(
            from_invalid_string,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "error".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::String(
                        "57f1f77bcf86cd799439011".into()
                    ))
                    .into(),
                    to: Type::ObjectId,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::String("error".to_string())).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }

    // Tests adapted from:
    // https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/#convert-to-string
    mod to_string {
        use super::*;

        test_constant_fold!(
            from_string_non_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Reference(("bar", 0u16).into())],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Reference(("bar", 0u16).into()).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
            schema_env = map! {
                ("bar", 0u16).into() => Schema::Atomic(Atomic::String),
            },
        );

        test_constant_fold!(
            from_boolean_false_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "false".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(false)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_boolean_true_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "true".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Boolean(true)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_date_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    DATE_WITH_MS.to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::DateTime(*DATE_WITH_MS)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_decimal_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String("1.2".to_string()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Decimal128(
                        bson::Decimal128::from_str("1.2").unwrap()
                    ))
                    .into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_double_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String("2.5".to_string()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Double(2.5)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_int_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String("2".to_string()))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Integer(2)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_long_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "1000".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::Long(1000)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );

        test_constant_fold!(
            from_objectid_literal,
            expected = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Literal(LiteralValue::String(
                    "507f1f77bcf86cd799439011".to_string()
                ))],
                cache: SchemaCache::new(),
            }),
            expected_changed = true,
            input = Stage::Array(ArraySource {
                alias: "".into(),
                array: vec![Expression::Cast(CastExpr {
                    expr: Expression::Literal(LiteralValue::ObjectId(*OID)).into(),
                    to: Type::String,
                    on_null: Expression::Literal(LiteralValue::Null).into(),
                    on_error: Expression::Literal(LiteralValue::Null).into(),
                    is_nullable: true,
                })],
                cache: SchemaCache::new(),
            }),
        );
    }
}
