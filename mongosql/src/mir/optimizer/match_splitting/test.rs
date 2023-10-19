mod match_splitting_test {
    use crate::mir::{
        self, schema::SchemaCache, Expression::*, LiteralValue::*, ScalarFunction::*,
    };

    macro_rules! test_match_splitting {
        ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
            #[test]
            fn $func_name() {
                use crate::mir::{optimizer::match_splitting::MatchSplittingOptimizer, *};
                let input = $input;
                let expected = $expected;
                let actual = MatchSplittingOptimizer::split_matches(input);
                assert_eq!(expected, actual);
            }
        };
    }

    test_match_splitting!(
        simple_filter_no_split,
        expected = Stage::Filter(mir::Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: Literal(Integer(42).into()),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(mir::Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new()
            })),
            condition: Literal(Integer(42).into()),
            cache: SchemaCache::new(),
        }),
    );

    test_match_splitting!(
        conjunctive_filter,
        expected = Stage::Filter(mir::Filter {
            source: Box::new(Stage::Filter(mir::Filter {
                source: Box::new(Stage::Filter(mir::Filter {
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![],
                        alias: "foo".into(),
                        cache: SchemaCache::new(),
                    })),
                    condition: ScalarFunction(ScalarFunctionApplication {
                        function: Eq,
                        args: vec![Literal(Integer(1).into()), Literal(Integer(1).into())],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                    cache: SchemaCache::new(),
                })),
                condition: ScalarFunction(ScalarFunctionApplication {
                    function: Eq,
                    args: vec![Literal(Integer(2).into()), Literal(Integer(2).into())],
                    cache: SchemaCache::new(),
                    is_nullable: true,
                }),
                cache: SchemaCache::new(),
            })),
            condition: ScalarFunction(ScalarFunctionApplication {
                function: Eq,
                args: vec![Literal(Integer(3).into()), Literal(Integer(3).into())],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
        input = Stage::Filter(mir::Filter {
            source: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "foo".into(),
                cache: SchemaCache::new(),
            })),
            condition: ScalarFunction(ScalarFunctionApplication {
                function: And,
                args: vec![
                    ScalarFunction(ScalarFunctionApplication {
                        function: Eq,
                        args: vec![Literal(Integer(1).into()), Literal(Integer(1).into())],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                    ScalarFunction(ScalarFunctionApplication {
                        function: Eq,
                        args: vec![Literal(Integer(2).into()), Literal(Integer(2).into())],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    }),
                    ScalarFunction(ScalarFunctionApplication {
                        function: Eq,
                        args: vec![Literal(Integer(3).into()), Literal(Integer(3).into())],
                        cache: SchemaCache::new(),
                        is_nullable: true,
                    })
                ],
                cache: SchemaCache::new(),
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );
}
