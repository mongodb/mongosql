mod constant_folding {
    macro_rules! test_constant_fold {
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::mir::{optimizer::constant_folding::ConstantFoldingOptimizer, schema::SchemaCheckingMode};
            let input = $input;
            let expected = $expected;

            #[allow(unused_mut, unused_assignments)]
            let mut schema_checking_mode = SchemaCheckingMode::Strict;
            $(schema_checking_mode = $schema_checking_mode;)?

            let actual = ConstantFoldingOptimizer::fold_constants(input, schema_checking_mode);
            assert_eq!(actual, expected);
        }
    };
    }
    use crate::{
        mir::{definitions::*, schema::SchemaCache},
        unchecked_unique_linked_hash_map,
    };

    fn test_source() -> Stage {
        Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })
    }

    test_constant_fold!(
        literal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        or_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        and_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        true_and_nulls_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_and_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        true_and_nulls_and_ref_is_null_and_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        false_or_nulls_or_ref_is_null_or_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_or_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        false_or_nulls_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        or_with_true_literal_is_true,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        or_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Or,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        and_with_false_literal_is_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        and_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(3).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_to_zero_is_zero,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Long(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(-1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(-1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_constant_ref_is_constant_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Double(3.0).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
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
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(-1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Long(8).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_to_one_is_one,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Double(1.0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Double(0.5).into()),
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                ],
                cache: SchemaCache::new(),
            })],
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
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                    Expression::Literal(LiteralValue::Double(0.5).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        arithmetic_null_arg,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Mul,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Long(2).into()),
                            Expression::Literal(LiteralValue::Double(2.0).into()),
                            Expression::Literal(LiteralValue::Null.into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        add_different_num_types,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Long(4).into()),
                    Expression::Literal(LiteralValue::Double(6.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Long(2).into()),
                    Expression::Literal(LiteralValue::Double(3.0).into()),
                    Expression::Literal(LiteralValue::Double(3.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        mul_different_num_types,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(4).into()),
                    Expression::Literal(LiteralValue::Long(9).into()),
                    Expression::Literal(LiteralValue::Double(16.0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Long(3).into()),
                    Expression::Literal(LiteralValue::Long(3).into()),
                    Expression::Literal(LiteralValue::Double(4.0).into()),
                    Expression::Literal(LiteralValue::Double(4.0).into())
                ],
                cache: SchemaCache::new(),
            })],
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
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::Long(0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        sub_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Sub,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Long(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        sub_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Long(-1).into()),
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Long(1).into()),
                            Expression::Literal(LiteralValue::Long(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Sub,
                        args: vec![
                            Expression::Literal(LiteralValue::Double(3.0).into()),
                            Expression::Literal(LiteralValue::Double(1.0).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
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
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Div,
                args: vec![
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::Long(0).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        div_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Div,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Long(2).into())
                ],
                cache: SchemaCache::new(),
            })],
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
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Div,
                args: vec![
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::Long(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        div_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(-1).into()),
                    Expression::Literal(LiteralValue::Double(2.0).into()),
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(LiteralValue::Long(-2).into()),
                            Expression::Literal(LiteralValue::Long(2).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Div,
                        args: vec![
                            Expression::Literal(LiteralValue::Double(2.0).into()),
                            Expression::Literal(LiteralValue::Double(1.0).into())
                        ],
                        cache: SchemaCache::new(),
                    }),
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
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_less_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                    Expression::Literal(LiteralValue::Boolean(false).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_less_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_less_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_greater_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gt,
                args: vec![
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_greater_than,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gt,
                args: vec![
                    Expression::Literal(LiteralValue::Long(0).into()),
                    Expression::Literal(LiteralValue::Long(0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_greater_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gte,
                args: vec![
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_greater_than_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Gte,
                args: vec![
                    Expression::Literal(LiteralValue::Double(0.0).into()),
                    Expression::Literal(LiteralValue::Double(1.0).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(0).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_nequal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(0).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_nequal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Long(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        compare_different_datatypes,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neq,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        compare_null_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lte,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Long(1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_between,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_not_between,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Reference(("foo", 1u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        fold_comparison_nested,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Lt,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Literal(LiteralValue::Boolean(true).into()),
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Integer(3).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::Reference(("foo", 1u16).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(-1).into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                        ],
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        pos_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Pos,
                args: vec![Expression::Literal(LiteralValue::Integer(2).into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        neg_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(-2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Neg,
                args: vec![Expression::Literal(LiteralValue::Integer(2).into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        not_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Not,
                args: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        upper_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("AABBCC".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Upper,
                args: vec![Expression::Literal(
                    LiteralValue::String("aaBBcC".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        lower_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("aabbcc".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lower,
                args: vec![Expression::Literal(
                    LiteralValue::String("aaBBcC".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        lower_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lower,
                args: vec![Expression::Literal(LiteralValue::Null.into()),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        btrim_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("AABBCC".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::String("aAABBCCa".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        ltrim_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("AABBCCa".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::LTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::String("aAABBCCa".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        rtrim_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("aAABBCC".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::RTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::String("aAABBCCa".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        btrim_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BTrim,
                args: vec![
                    Expression::Literal(LiteralValue::String("a".to_string()).into()),
                    Expression::Literal(LiteralValue::Null.into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_nested,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("hello".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into()),
                        ],
                        cache: SchemaCache::new(),
                    })
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_multi_codepoint_char,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("🇷🇺ááá".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("ááá🇷🇺ááá".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_negative_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("world".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                    Expression::Literal(LiteralValue::Integer(-1).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_negative_start_with_smaller_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(-6).into()),
                    Expression::Literal(LiteralValue::Integer(5).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_negative_start_with_larger_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("hello".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(-6).into()),
                    Expression::Literal(LiteralValue::Integer(11).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_start_larger_than_string_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(20).into()),
                    Expression::Literal(LiteralValue::Integer(4).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_end_larger_than_string_length,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("world".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                    Expression::Literal(LiteralValue::Integer(20).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        substring_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Substring,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Literal(LiteralValue::Integer(6).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("hello world".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world".to_string()).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_with_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello world".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::String("hello world2".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                    Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world2".to_string()).into()),
                    Expression::Reference(("a", 0u16).into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("".to_string()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        concat_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::String("hello".to_string()).into()),
                    Expression::Literal(LiteralValue::String("world".to_string()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        char_length_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(11).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::CharLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("hello world".to_string()).into()
                )],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        char_length_multi_codepoint,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(14).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::CharLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("ááá🇷🇺ááá".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        octet_length_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(11).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::OctetLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("hello world".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        octet_length_multi_codepoint,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(26).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::OctetLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("ááá🇷🇺ááá".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        bit_length_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(88).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BitLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("hello world".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        bit_length_multi_codepoint,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(208).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::BitLength,
                args: vec![Expression::Literal(
                    LiteralValue::String("ááá🇷🇺ááá".to_string()).into()
                ),],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        array_size_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Size,
                args: vec![Expression::Array(
                    vec![
                        Expression::Literal(LiteralValue::Integer(0).into()),
                        Expression::Literal(LiteralValue::Integer(0).into())
                    ]
                    .into()
                )],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        array_size_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Size,
                args: vec![Expression::Array(vec![].into())],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        coalesce_empty,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Coalesce,
                args: vec![],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        coalesce_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Coalesce,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Integer(0).into()),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        coalesce_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Coalesce,
                args: vec![
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Null.into()),
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        merge_objects_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                "b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
            .into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        merge_objects_reference,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
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
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![],
                cache: SchemaCache::new(),
            })],
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
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::MergeObjects,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "b".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into()),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"b".into() => Expression::Literal(LiteralValue::Integer(0).into()),
                        "c".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Reference(("a", 0u16).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_if_args_equal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::NullIf,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        null_if_args_unequal,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::NullIf,
                args: vec![
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        computed_field_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Literal(LiteralValue::String("a".into()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        computed_field_missing,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Literal(LiteralValue::String("b".into()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2).into())}
                    .into()),
                    Expression::Literal(LiteralValue::String("b".into()).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![Expression::Literal(LiteralValue::Integer(2).into())].into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Null.into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-2).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_negative_start_within_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![Expression::Literal(LiteralValue::Integer(2).into())].into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-2).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-5).into()),
                    Expression::Literal(LiteralValue::Integer(2).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(-5).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(5).into())
                ],
                cache: SchemaCache::new(),
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
                    Expression::Literal(LiteralValue::Integer(1).into()),
                    Expression::Literal(LiteralValue::Integer(2).into()),
                    Expression::Literal(LiteralValue::Integer(3).into())
                ]
                .into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(5).into())
                ],
                cache: SchemaCache::new(),
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
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::Literal(LiteralValue::Integer(1).into())
                ],
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        slice_with_pos_neg_length_is_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Slice,
                args: vec![
                    Expression::Array(
                        vec![
                            Expression::Literal(LiteralValue::Integer(1).into()),
                            Expression::Literal(LiteralValue::Integer(2).into()),
                            Expression::Literal(LiteralValue::Integer(3).into())
                        ]
                        .into()
                    ),
                    Expression::Literal(LiteralValue::Integer(5).into()),
                    Expression::Literal(LiteralValue::Integer(-1).into())
                ],
                cache: SchemaCache::new(),
            }),],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                to: Type::Boolean,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_mismatched_types,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                to: Type::String,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                to: Type::String,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_array_as_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Array(
                vec![Expression::Literal(LiteralValue::Boolean(true).into())].into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Array(
                    vec![Expression::Literal(LiteralValue::Boolean(true).into())].into()
                )
                .into(),
                to: Type::Array,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_non_array_as_array,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("error".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Integer(0).into()).into(),
                to: Type::Array,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_document_as_document,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
            .into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                .into())
                .into(),
                to: Type::Document,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::Null.into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_non_document_as_document,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("error".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Integer(0).into()).into(),
                to: Type::Document,
                on_null: Expression::Literal(LiteralValue::Null.into()).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        cast_null,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("null".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Cast(CastExpr {
                expr: Expression::Literal(LiteralValue::Null.into()).into(),
                to: Type::Array,
                on_null: Expression::Literal(LiteralValue::String("null".into()).into()).into(),
                on_error: Expression::Literal(LiteralValue::String("error".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::Literal(LiteralValue::Integer(1).into()).into(),
                target_type: TypeOrMissing::Type(Type::Int32),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::Literal(LiteralValue::String("a".into()).into()).into(),
                target_type: TypeOrMissing::Type(Type::Double),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_number,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::Literal(LiteralValue::Double(1.0).into()).into(),
                target_type: TypeOrMissing::Number,
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        is_expr_nested,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Boolean(true).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Is(IsExpr {
                expr: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Concat,
                    args: vec![
                        Expression::Literal(LiteralValue::String("hello ".to_string()).into()),
                        Expression::Literal(LiteralValue::String("world".to_string()).into()),
                    ],
                    cache: SchemaCache::new(),
                })
                .into(),
                target_type: TypeOrMissing::Type(Type::String),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("then 2".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(3).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 2".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_ref_ahead,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 2".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 2".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_prune_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then a".into()).into()).into()
                },],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(3).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then a".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_else,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("else".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Literal(LiteralValue::Integer(2).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(3).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        simple_case_keep_branches,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SimpleCase(SimpleCaseExpr {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("then true".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(false).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then false".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then true".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_ref_ahead,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then true".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(true).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then true".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_prune_false,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![WhenBranch {
                    when: Expression::Reference(("a", 0u16).into()).into(),
                    then: Expression::Literal(LiteralValue::String("then 3".into()).into()).into()
                },],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(false).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then false".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Reference(("a", 0u16).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 3".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::String("false".into()).into())
                            .into(),
                        then: Expression::Literal(
                            LiteralValue::String("then false string".into()).into()
                        )
                        .into()
                    },
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        searched_case_else,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(
                LiteralValue::String("else".into()).into()
            )],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::SearchedCase(SearchedCaseExpr {
                when_branch: vec![
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Boolean(false).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then false".into()).into())
                            .into()
                    },
                    WhenBranch {
                        when: Expression::Literal(LiteralValue::Integer(4).into()).into(),
                        then: Expression::Literal(LiteralValue::String("then 4".into()).into())
                            .into()
                    }
                ],
                else_branch: Expression::Literal(LiteralValue::String("else".into()).into()).into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        field_access_simple,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::Literal(LiteralValue::Integer(0).into())],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into())}
                .into())
                .into(),
                field: "a".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        field_access_missing_field,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into())}
                .into())
                .into(),
                field: "b".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(0).into())}
                .into())
                .into(),
                field: "b".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        field_access_ref,
        expected = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                field: "a".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
        input = Stage::Array(ArraySource {
            alias: "".into(),
            array: vec![Expression::FieldAccess(FieldAccess {
                expr: Expression::Reference(("a", 0u16).into()).into(),
                field: "a".into(),
                cache: SchemaCache::new(),
            })],
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        offset_simple,
        expected = test_source(),
        input = Stage::Offset(Offset {
            source: Box::new(test_source()),
            offset: 0,
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        offset_nonzero,
        expected = Stage::Offset(Offset {
            source: Box::new(test_source()),
            offset: 1,
            cache: SchemaCache::new(),
        }),
        input = Stage::Offset(Offset {
            source: Box::new(test_source()),
            offset: 1,
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        filter_simple,
        expected = test_source(),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::Literal(LiteralValue::Boolean(true).into()),
            cache: SchemaCache::new(),
        }),
    );
    test_constant_fold!(
        filter_non_literal,
        expected = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::Reference(("a", 0u16).into()),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(Filter {
            source: Box::new(test_source()),
            condition: Expression::Reference(("a", 0u16).into()),
            cache: SchemaCache::new(),
        }),
    );
}
