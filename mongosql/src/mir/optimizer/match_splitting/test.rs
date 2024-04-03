mod match_splitting_test {
    use crate::mir::{schema::SchemaCache, Expression::*, LiteralValue::*, ScalarFunction::*};

    macro_rules! test_match_splitting {
        ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr,) => {
            #[test]
            fn $func_name() {
                use crate::mir::{optimizer::match_splitting::MatchSplittingOptimizer, *};
                let input = $input;
                let expected = $expected;
                let (actual, _) = MatchSplittingOptimizer::split_matches(input);
                assert_eq!(expected, actual);
            }
        };
    }

    test_match_splitting!(
        simple_filter_no_split,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: Literal(Integer(42)),
            cache: SchemaCache::new(),
        }),
        expected_changed = false,
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: Literal(Integer(42)),
            cache: SchemaCache::new(),
        }),
    );

    test_match_splitting!(
        conjunctive_filter,
        expected = Stage::Filter(Filter {
            source: Box::new(Stage::Filter(Filter {
                source: Box::new(Stage::Filter(Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "foo".into(),
                        cache: SchemaCache::new(),
                    })),
                    condition: ScalarFunction(ScalarFunctionApplication::new(
                        Eq,
                        vec![Literal(Integer(1)), Literal(Integer(1))],
                    )),
                    cache: SchemaCache::new(),
                })),
                condition: ScalarFunction(ScalarFunctionApplication::new(
                    Eq,
                    vec![Literal(Integer(2)), Literal(Integer(2))],
                )),
                cache: SchemaCache::new(),
            })),
            condition: ScalarFunction(ScalarFunctionApplication::new(
                Eq,
                vec![Literal(Integer(3)), Literal(Integer(3))],
            )),
            cache: SchemaCache::new(),
        }),
        expected_changed = true,
        input = Stage::Filter(Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            condition: ScalarFunction(ScalarFunctionApplication {
                function: And,
                args: vec![
                    ScalarFunction(ScalarFunctionApplication::new(
                        Eq,
                        vec![Literal(Integer(1)), Literal(Integer(1))],
                    )),
                    ScalarFunction(ScalarFunctionApplication::new(
                        Eq,
                        vec![Literal(Integer(2)), Literal(Integer(2))],
                    )),
                    ScalarFunction(ScalarFunctionApplication::new(
                        Eq,
                        vec![Literal(Integer(3)), Literal(Integer(3))],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );
}
