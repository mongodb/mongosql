use crate::{
    map,
    mir::{
        schema::{
            errors::HigherOrderFunctionErrorCause, errors::IncorrectArgCountPrecision,
            Error as mir_error, THIS_VARIABLE, VALUE_VARIABLE,
        },
        *,
    },
    schema::{Atomic, Schema, ANY_ARRAY_OR_NULLISH, BOOLEAN_OR_NULLISH, NUMERIC_OR_NULLISH},
    set, test_schema,
};

mod map {
    use super::*;

    // MAP(['a'], 1) => [1]
    test_schema!(
        result_schema_is_array_of_second_arg_schema,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Array(ArrayExpr {
                array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
            })),
            f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
            is_nullable: false,
        })),
    );

    // MAP(1, 1) => error
    test_schema!(
        first_arg_must_be_array,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Map",
            required: ANY_ARRAY_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
            var_cause: None,
        }),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Literal(LiteralValue::Integer(1))),
            f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
            is_nullable: false,
        })),
    );

    // MAP(array_or_null_or_missing_field, 1) => [1, ...] or null
    test_schema!(
        first_arg_may_be_nullish_array,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            Schema::Atomic(Atomic::Null),
        })),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Reference(("foo", 0u16).into())),
            f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
            is_nullable: true,
        })),
        schema_env = map! {("foo", 0u16).into() => ANY_ARRAY_OR_NULLISH.clone()},
    );

    // MAP([1], this + 1) => [2]
    test_schema!(
        valid_when_this_variable_usage_is_satisfied_by_schema_of_items_of_first_arg,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Array(ArrayExpr {
                array: vec![Expression::Literal(LiteralValue::Integer(1))]
            })),
            f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                    Expression::Literal(LiteralValue::Integer(1)),
                ],
                is_nullable: false,
            })),
            is_nullable: false,
        })),
    );

    // MAP(['a'], this + 1) => error
    test_schema!(
        invalid_when_this_variable_usage_is_violated_by_schema_of_items_of_first_arg,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Map",
            cause: HigherOrderFunctionErrorCause::ThisUsage,
            error: Box::new(mir_error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone().into(),
                found: Schema::Atomic(Atomic::String).into(),
                var_cause: Some(THIS_VARIABLE.to_string()),
            })
        }),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Array(ArrayExpr {
                array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
            })),
            f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Add,
                args: vec![
                    Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                    Expression::Literal(LiteralValue::Integer(1)),
                ],
                is_nullable: false,
            })),
            is_nullable: false,
        })),
    );

    // MAP(['a'], [this, string_field + 1]) => error
    test_schema!(
        do_not_incorrectly_blame_variable_usage_when_variable_is_valid_but_function_is_still_valid,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Map",
            cause: HigherOrderFunctionErrorCause::FunctionArgument,
            error: Box::new(mir_error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone().into(),
                found: Schema::Atomic(Atomic::String).into(),
                var_cause: None,
            })
        }),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Array(ArrayExpr {
                array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
            })),
            f: Box::new(Expression::Array(ArrayExpr {
                array: vec![
                    Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::String("a".to_string())),
                            Expression::Literal(LiteralValue::Integer(1)),
                        ],
                        is_nullable: false,
                    }),
                ],
            })),
            is_nullable: false,
        })),
    );

    // This situation may arise as a result of syntactic rewriting. For example, a user may write
    // `SELECT MAP([1], *) FROM foo` which would be rewritten to `SELECT MAP([1], this *) FROM foo`
    // which is invalid because `*` requires two arguments at minimum.
    // MAP([1], this *) => error
    test_schema!(
        invalid_when_function_arg_has_incorrect_arg_count,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Map",
            cause: HigherOrderFunctionErrorCause::FunctionArgument,
            error: Box::new(mir_error::IncorrectArgumentCount {
                name: "Mul",
                required: IncorrectArgCountPrecision::Minimum(2),
                found: 1,
            })
        }),
        input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Map(MapExpr {
            array: Box::new(Expression::Array(ArrayExpr {
                array: vec![Expression::Literal(LiteralValue::Integer(1))]
            })),
            f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Mul,
                args: vec![Expression::Variable(Variable::new(
                    THIS_VARIABLE.to_string()
                )),],
                is_nullable: false,
            })),
            is_nullable: false,
        })),
    );
}

mod filter {
    use super::*;

    // FILTER([1], true) => [1]
    test_schema!(
        result_schema_is_same_as_input_schema,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                f: Box::new(Expression::Literal(LiteralValue::Boolean(true))),
                is_nullable: false,
            })),
    );

    // FILTER(1, true) => error
    test_schema!(
        first_arg_must_be_array,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Filter",
            required: ANY_ARRAY_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
            var_cause: None,
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                f: Box::new(Expression::Literal(LiteralValue::Boolean(true))),
                is_nullable: false,
            })),
    );

    // FILTER(array_or_null_or_missing_field, true) => [...] or null
    test_schema!(
        first_arg_may_be_nullish_array,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Array(Box::new(Schema::Any)),
            Schema::Atomic(Atomic::Null),
        })),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Reference(("foo", 0u16).into())),
                f: Box::new(Expression::Literal(LiteralValue::Boolean(true))),
                is_nullable: true,
            })),
        schema_env = map! {("foo", 0u16).into() => ANY_ARRAY_OR_NULLISH.clone()},
    );

    // FILTER([1], 1) => error
    test_schema!(
        second_arg_must_be_boolean,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Filter",
            required: BOOLEAN_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
            var_cause: None,
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                is_nullable: false,
            })),
    );

    // FILTER(['a'], boolean_or_null_or_missing_field) => [...]
    test_schema!(
        second_arg_may_be_nullish_boolean,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::String)))),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
                })),
                f: Box::new(Expression::Reference(("foo", 0u16).into())),
                is_nullable: false,
            })),
        schema_env = map! {("foo", 0u16).into() => BOOLEAN_OR_NULLISH.clone()},
    );

    // FILTER([1], this > 1) => [...]
    test_schema!(
        valid_when_this_variable_usage_is_satisfied_by_schema_of_items_of_first_arg,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![
                        Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );

    // FILTER(['a'], this > 1) => error
    test_schema!(
        invalid_when_this_variable_usage_is_violated_by_schema_of_items_of_first_arg,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Filter",
            cause: HigherOrderFunctionErrorCause::ThisUsage,
            error: Box::new(mir_error::InvalidComparison {
                name: "Gt",
                left: Schema::Atomic(Atomic::String).into(),
                right: Schema::Atomic(Atomic::Integer).into(),
                var_cause: Some(THIS_VARIABLE.to_string()),
            }),
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
                })),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![
                        Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );

    // This situation may arise as a result of syntactic rewriting. For example, a user may write
    // `SELECT FILTER([1], >) FROM foo` which would be rewritten to
    // `SELECT FILTER([1], this >) FROM foo`
    // which is invalid because `>` requires two arguments.
    // FILTER([1], this >) => error
    test_schema!(
        invalid_when_function_arg_has_incorrect_arg_count,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Filter",
            cause: HigherOrderFunctionErrorCause::FunctionArgument,
            error: Box::new(mir_error::IncorrectArgumentCount {
                name: "Gt",
                required: IncorrectArgCountPrecision::Exact(2),
                found: 1,
            })
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Filter(FilterExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gt,
                    args: vec![Expression::Variable(Variable::new(
                        THIS_VARIABLE.to_string()
                    )),],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );
}

mod reduce {
    use super::*;

    // REDUCE(1, 1, 1) => error
    test_schema!(
        first_arg_must_be_array,
        expected_error_code = 1002,
        expected = Err(mir_error::SchemaChecking {
            name: "Reduce",
            required: ANY_ARRAY_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
            var_cause: None,
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                init_value: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                is_nullable: false,
            })),
    );

    // REDUCE(array_or_null_or_missing_field, 1, 1) => 1 or null
    test_schema!(
        first_arg_may_be_nullish_array,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        })),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Reference(("foo", 0u16).into())),
                init_value: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                is_nullable: true,
            })),
        schema_env = map! {("foo", 0u16).into() => ANY_ARRAY_OR_NULLISH.clone()},
    );

    // REDUCE([1], 'a' + 1, 1) => error
    test_schema!(
        second_arg_must_be_valid,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::InitialValue,
            error: Box::new(mir_error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone().into(),
                found: Schema::Atomic(Atomic::String).into(),
                var_cause: None,
            })
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                init_value: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![
                        Expression::Literal(LiteralValue::String("a".to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                is_nullable: false,
            })),
    );

    // In general, the result schema of a Reduce is the union of the second and third arguments'
    // schemas. The following cases may all be valid:
    //  - second and third args have same schema
    //  - second arg schema is a subset of third arg schema
    //  - second arg schema is a superset of third arg schema
    //  - second arg schema is distinct from third arg schema
    // This module captures these cases and asserts the expected result set schema for each.
    mod result_schema {
        use super::*;

        // Same:
        // REDUCE(string_array, 1, 1)
        // - array: Schema::Array(Atomic::String)
        // - init_value: Schema::Atomic(Atomic::Integer)
        // - f: Schema::Atomic(Atomic::Integer)
        // => result: Schema::Atomic(Atomic::Integer)
        test_schema!(
            same_as_second_and_third_arg_when_second_and_third_args_have_same_schema,
            expected = Ok(Schema::Atomic(Atomic::Integer)),
            input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(
                ReduceExpr {
                    array: Box::new(Expression::Array(ArrayExpr {
                        array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
                    })),
                    init_value: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                    f: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                    is_nullable: false,
                }
            )),
        );

        // Subset:
        // REDUCE(int_array, 0, int_or_string_field)
        // - array: Schema::Array(Schema::Atomic(Atomic::Integer))
        // - init_value: Schema::Atomic(Atomic::Integer)
        // - f: Schema::AnyOf(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String))
        // => result: Schema::AnyOf(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String))
        // Note: this example is clearly a toy example; in practice, a user may use a polymorphic
        // expression for the function, such as CASE, but that is a bit convoluted for a unit test
        // and the ultimate purpose is to demonstrate that when the third argument is a superset of
        // the second argument and is valid, the result schema is the same as the third argument.
        test_schema!(
            same_as_third_arg_when_second_arg_is_subset_of_third_arg,
            expected = Ok(Schema::AnyOf(set! {
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            })),
            input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(
                ReduceExpr {
                    array: Box::new(Expression::Reference(("int_array", 0u16).into())),
                    init_value: Box::new(Expression::Literal(LiteralValue::Integer(0))),
                    f: Box::new(Expression::Reference(("int_or_string_field", 0u16).into())),
                    is_nullable: false,
                }
            )),
            schema_env = map! {
                ("int_array", 0u16).into() => Schema::Array(
                    Box::new(Schema::Atomic(Atomic::Integer))
                ),
                ("int_or_string_field", 0u16).into() => Schema::AnyOf(set! {
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                }),
            },
        );

        // Superset:
        // REDUCE(int_array, int_or_string_field, this + value::INT)
        // - array: Schema::Array(Schema::Atomic(Atomic::Integer))
        // - init_value:
        //     Schema::AnyOf(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String))
        // - f: Schema::Atomic(Atomic::Integer)
        // => result: Schema::AnyOf(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String))
        test_schema!(
            same_as_second_arg_when_second_arg_is_superset_of_third_arg,
            expected = Ok(Schema::AnyOf(set! {
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            })),
            input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(
                ReduceExpr {
                    array: Box::new(Expression::Reference(("int_array", 0u16).into())),
                    init_value: Box::new(Expression::Reference(
                        ("int_or_string_field", 0u16).into()
                    )),
                    f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                            Expression::Cast(CastExpr {
                                expr: Box::new(Expression::Variable(Variable::new(
                                    VALUE_VARIABLE.to_string()
                                ))),
                                to: Type::Int32,
                                on_null: Box::new(Expression::Literal(LiteralValue::Integer(0))),
                                on_error: Box::new(Expression::Literal(LiteralValue::Integer(0))),
                                is_nullable: false
                            }),
                        ],
                        is_nullable: false,
                    })),
                    is_nullable: false,
                }
            )),
            schema_env = map! {
                ("int_array", 0u16).into() => Schema::Array(
                    Box::new(Schema::Atomic(Atomic::Integer))
                ),
                ("int_or_string_field", 0u16).into() => Schema::AnyOf(set! {
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                }),
            },
        );

        // Distinct:
        // REDUCE(int_array, "1", this + value::INT)
        // - array: Schema::Array(Schema::Atomic(Atomic::Integer))
        // - init_value: Schema::Atomic(Atomic::String)
        // - f: Schema::Atomic(Atomic::Integer)
        // => result: Schema::AnyOf(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String))
        test_schema!(
            union_of_second_and_third_arg_when_second_arg_is_distinct_from_third_arg,
            expected = Ok(Schema::AnyOf(set! {
                Schema::Atomic(Atomic::Null),
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::String),
            })),
            input = Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(
                ReduceExpr {
                    array: Box::new(Expression::Reference(("int_array", 0u16).into())),
                    init_value: Box::new(Expression::Literal(LiteralValue::String(
                        "1".to_string()
                    ))),
                    f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                            Expression::Cast(CastExpr {
                                expr: Box::new(Expression::Variable(Variable::new(
                                    VALUE_VARIABLE.to_string()
                                ))),
                                to: Type::Int32,
                                on_null: Box::new(Expression::Literal(LiteralValue::Null)),
                                on_error: Box::new(Expression::Literal(LiteralValue::Null)),
                                is_nullable: false,
                            }),
                        ],
                        is_nullable: false,
                    })),
                    is_nullable: false,
                }
            )),
            schema_env = map! {
                ("int_array", 0u16).into() => Schema::Array(
                    Box::new(Schema::Atomic(Atomic::Integer))
                ),
            },
        );
    }

    // A Reduce is valid when:
    // - usages of the `this` variable are satisfied by the schema of the items of the first
    // argument, and
    // - usages of the `value` variable are satisfied by the schema of the second argument and by
    // the schema of the third argument.
    // REDUCE([1], 1, this + 1) => 2
    test_schema!(
        valid_when_this_variable_usage_is_satisfied_by_schema_of_items_of_first_arg,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                init_value: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![
                        Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );

    // REDUCE([1], 1, value + 1) => 2
    test_schema!(
        valid_when_value_variable_usage_is_satisfied_by_schema_of_second_arg_and_schema_of_third_arg,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                init_value: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![
                        Expression::Variable(Variable::new(VALUE_VARIABLE.to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );

    // REDUCE(['a'], 1, this + 1) => error
    test_schema!(
        invalid_when_this_variable_usage_is_violated_by_schema_of_items_of_first_arg,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::ThisUsage,
            error: Box::new(mir_error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone().into(),
                found: Schema::Atomic(Atomic::String).into(),
                var_cause: Some(THIS_VARIABLE.to_string()),
            })
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::String("a".to_string()))]
                })),
                init_value: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![
                        Expression::Variable(Variable::new(THIS_VARIABLE.to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );

    // REDUCE([1], 'a', value + 1) => error
    test_schema!(
        invalid_when_value_variable_usage_is_violated_by_schema_of_second_arg,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::InitialValueUsage,
            error: Box::new(mir_error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone().into(),
                found: Schema::Atomic(Atomic::String).into(),
                var_cause: Some(VALUE_VARIABLE.to_string()),
            })
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                init_value: Box::new(Expression::Literal(LiteralValue::String("a".to_string()))),
                f: Box::new(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![
                        Expression::Variable(Variable::new(VALUE_VARIABLE.to_string())),
                        Expression::Literal(LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );

    // REDUCE([1], 'a', CASE value WHEN 'a' THEN 1 ELSE 1 END) => error
    test_schema!(
        invalid_when_value_variable_usage_is_violated_by_schema_of_third_arg,
        expected_error_code = 1020,
        expected = Err(mir_error::HigherOrderFunctionWrapper {
            name: "Reduce",
            cause: HigherOrderFunctionErrorCause::AccumulatedValueUsage,
            error: Box::new(mir_error::InvalidComparison {
                name: "SimpleCase",
                left: Schema::Atomic(Atomic::Integer).into(),
                right: Schema::Atomic(Atomic::String).into(),
                var_cause: Some(VALUE_VARIABLE.to_string()),
            }),
        }),
        input =
            Expression::HigherOrderFunction(HigherOrderFunctionApplication::Reduce(ReduceExpr {
                array: Box::new(Expression::Array(ArrayExpr {
                    array: vec![Expression::Literal(LiteralValue::Integer(1))]
                })),
                init_value: Box::new(Expression::Literal(LiteralValue::String("a".to_string()))),
                f: Box::new(Expression::SimpleCase(SimpleCaseExpr {
                    expr: Box::new(Expression::Variable(Variable::new(
                        VALUE_VARIABLE.to_string()
                    ))),
                    when_branch: vec![WhenBranch {
                        when: Box::new(Expression::Literal(LiteralValue::String("a".to_string()))),
                        then: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                        is_nullable: false,
                    }],
                    else_branch: Box::new(Expression::Literal(LiteralValue::Integer(1))),
                    is_nullable: false,
                })),
                is_nullable: false,
            })),
    );
}
