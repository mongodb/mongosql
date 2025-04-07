use crate::{
    ast, map, mir, multimap,
    schema::{Atomic, Satisfaction, Schema, ANY_DOCUMENT, NUMERIC_OR_NULLISH},
    test_algebrize, unchecked_unique_linked_hash_map,
    usererror::UserError,
};
test_algebrize!(
    count_star,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::CountStar(false)),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Count,
        args: ast::FunctionArguments::Star,
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    count_distinct_star,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::CountStar(true)),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Count,
        args: ast::FunctionArguments::Star,
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);
test_algebrize!(
    count_all_expr_basic_test,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Count,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Count,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    count_distinct_expr_basic_test,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Count,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Count,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);
test_algebrize_expr_and_schema_check!(
    count_distinct_expr_argument_not_self_comparable_is_error,
    method = algebrize_aggregation,
    expected = Err(Error::SchemaChecking(
        mir::schema::Error::AggregationArgumentMustBeSelfComparable(
            "Count DISTINCT".into(),
            Schema::Any.into(),
        )
    )),
    expected_error_code = 1003,
    input = ast::FunctionExpr {
        function: ast::FunctionName::Count,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Identifier("foo".into())]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
    env = map! {
        ("d", 1u16).into() => ANY_DOCUMENT.clone(),
    },
);
test_algebrize!(
    sum_star_is_error,
    method = algebrize_aggregation,
    expected = Err(Error::StarInNonCount),
    expected_error_code = 3010,
    input = ast::FunctionExpr {
        function: ast::FunctionName::Sum,
        args: ast::FunctionArguments::Star,
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    sum_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Sum,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Sum,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    sum_distinct_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Sum,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Sum,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);
test_algebrize_expr_and_schema_check!(
    sum_argument_must_be_numeric,
    method = algebrize_aggregation,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Sum",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::FunctionExpr {
        function: ast::FunctionName::Sum,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor("42".into(),)]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    avg_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Avg,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Avg,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    avg_distinct_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Avg,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Avg,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize_expr_and_schema_check!(
    avg_argument_must_be_numeric,
    method = algebrize_aggregation,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "Avg",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::FunctionExpr {
        function: ast::FunctionName::Avg,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor("42".into(),)]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    stddevpop_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::StddevPop,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::StddevPop,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    stddevpop_distinct_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::StddevPop,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::StddevPop,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);
test_algebrize_expr_and_schema_check!(
    stddevpop_argument_must_be_numeric,
    method = algebrize_aggregation,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "StddevPop",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::FunctionExpr {
        function: ast::FunctionName::StddevPop,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor("42".into(),)]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    stddevsamp_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::StddevSamp,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::StddevSamp,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    stddevsamp_distinct_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::StddevSamp,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::StddevSamp,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);
test_algebrize_expr_and_schema_check!(
    stddevsamp_argument_must_be_numeric,
    method = algebrize_aggregation,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "StddevSamp",
        required: NUMERIC_OR_NULLISH.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::FunctionExpr {
        function: ast::FunctionName::StddevSamp,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor("42".into(),)]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    addtoarray_expr_basic_test,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::AddToArray,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::AddToArray,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    addtoarray_distinct_expr_basic_test,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::AddToArray,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::AddToArray,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    addtoset_expr_is_addtoarray_distinct_in_mir,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::AddToArray,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::AddToSet,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    addtoset_distinct_expr_is_addtoarray_in_mir,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::AddToArray,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::AddToSet,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    first_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::First,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::First,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    first_distinct_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::First,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::First,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    last_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Last,
            distinct: false,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Last,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize!(
    last_distinct_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Last,
            distinct: true,
            arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            arg_is_possibly_doc: Satisfaction::Not,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::Last,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(ast::Literal::Integer(
            42
        ))]),
        set_quantifier: Some(ast::SetQuantifier::Distinct),
    },
);

test_algebrize!(
    mergedocuments_expr,
    method = algebrize_aggregation,
    expected = Ok(mir::AggregationExpr::Function(
        mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::MergeDocuments,
            distinct: false,
            arg: Box::new(mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    "b".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                }
                .into(),
            )),
            arg_is_possibly_doc: Satisfaction::Must,
        }
    )),
    input = ast::FunctionExpr {
        function: ast::FunctionName::MergeDocuments,
        args: ast::FunctionArguments::Args(vec![ast::Expression::Document(multimap! {
            "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            "b".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
        })]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
test_algebrize_expr_and_schema_check!(
    mergedocuments_argument_must_be_document,
    method = algebrize_aggregation,
    expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
        name: "MergeDocuments",
        required: ANY_DOCUMENT.clone().into(),
        found: Schema::Atomic(Atomic::String).into(),
    })),
    expected_error_code = 1002,
    input = ast::FunctionExpr {
        function: ast::FunctionName::MergeDocuments,
        args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor("42".into(),)]),
        set_quantifier: Some(ast::SetQuantifier::All),
    },
);
