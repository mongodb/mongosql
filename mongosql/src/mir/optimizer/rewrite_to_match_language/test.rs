use std::str::FromStr;

use crate::{
    catalog::Catalog,
    map,
    mir::{
        optimizer::{rewrite_to_match_language::MatchLanguageRewriter, Optimizer},
        schema::{SchemaCache, SchemaInferenceState},
        *,
    },
    schema::{Atomic, Document, Schema, SchemaEnvironment},
    set, unchecked_unique_linked_hash_map,
    util::{mir_collection, mir_field_access, mir_field_access_multi_part, mir_field_path},
    SchemaCheckingMode,
};
use agg_ast::definitions::Namespace;
use lazy_static::lazy_static;

lazy_static! {
    static ref CATALOG: Catalog = Catalog::new(map! {
        Namespace {database: "db".to_string(), collection: "foo".to_string()} => Schema::Document(Document {
            keys: map! {
                "str".to_string() => Schema::Atomic(Atomic::String),
                "pat".to_string() => Schema::Atomic(Atomic::String),
                "int".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            ..Default::default()
            }),
    });
}

macro_rules! test_rewrite_to_match_language {
    ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let state = SchemaInferenceState::new(
                0u16,
                SchemaEnvironment::default(),
                &*CATALOG,
                SchemaCheckingMode::Relaxed,
            );

            let optimizer = &MatchLanguageRewriter;
            let (actual, actual_changed) =
                optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!($expected_changed, actual_changed);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_rewrite_to_match_language_no_op {
    ($func_name:ident, $input:expr) => {
        test_rewrite_to_match_language! { $func_name, expected = $input, expected_changed = false, input = $input }
    };
}

fn singleton_project(expr: Expression) -> Stage {
    Stage::Project(Project {
        is_add_fields: false,
        source: mir_collection("db", "foo"),
        expression: map! {
            ("foo", 0u16).into() => Expression::Document(DocumentExpr {
                document: unchecked_unique_linked_hash_map! {
                    "expr".to_string() => expr,
                },
            }),
        },
        cache: SchemaCache::new(),
    })
}

fn filter_stage(condition: Expression) -> Stage {
    Stage::Filter(Filter {
        source: mir_collection("db", "foo"),
        condition,
        cache: SchemaCache::new(),
    })
}

fn match_filter_stage(condition: MatchQuery) -> Stage {
    Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
        source: mir_collection("db", "foo"),
        condition,
        cache: SchemaCache::new(),
    })))
}

// The following "helper functions" cannot exist as static variables since
// they all contain SchemaCache which is not thread safe. The amount of
// refactoring required to support these as lazy_static refs outweighs the
// "cost" of having these be functions. Note that since this is test code,
// the cost is inconsequential.
fn valid_is() -> Expression {
    Expression::Is(IsExpr {
        expr: mir_field_access("foo", "str", true),
        target_type: TypeOrMissing::Type(Type::String),
    })
}

fn valid_match_is() -> MatchQuery {
    MatchQuery::Type(MatchLanguageType {
        input: Some(mir_field_path("foo", vec!["str"])),
        target_type: TypeOrMissing::Type(Type::String),
        cache: SchemaCache::new(),
    })
}

fn valid_is_null() -> Expression {
    Expression::Is(IsExpr {
        expr: mir_field_access("foo", "str", true),
        target_type: TypeOrMissing::Type(Type::Null),
    })
}

fn valid_match_is_null() -> MatchQuery {
    MatchQuery::Comparison(MatchLanguageComparison {
        function: MatchLanguageComparisonOp::Eq,
        input: Some(mir_field_path("foo", vec!["str"])),
        arg: LiteralValue::Null,
        cache: SchemaCache::new(),
    })
}

fn invalid_is() -> Expression {
    Expression::Is(IsExpr {
        expr: Box::new(Expression::Literal(LiteralValue::Integer(1))),
        target_type: TypeOrMissing::Type(Type::Int32),
    })
}

fn valid_like() -> Expression {
    Expression::Like(LikeExpr {
        expr: mir_field_access("foo", "str", true),
        pattern: Box::new(Expression::Literal(LiteralValue::String("abc".to_string()))),
        escape: None,
    })
}

fn valid_match_like() -> MatchQuery {
    MatchQuery::Regex(MatchLanguageRegex {
        input: Some(mir_field_path("foo", vec!["str"])),
        regex: "^abc$".to_string(),
        options: "si".to_string(),
        cache: SchemaCache::new(),
    })
}

fn invalid_like_expr() -> Expression {
    Expression::Like(LikeExpr {
        expr: Box::new(Expression::Literal(LiteralValue::String("abc".to_string()))),
        pattern: Box::new(Expression::Literal(LiteralValue::String("abc".to_string()))),
        escape: None,
    })
}

fn invalid_like_pat() -> Expression {
    Expression::Like(LikeExpr {
        expr: mir_field_access("foo", "str", true),
        pattern: mir_field_access("foo", "pat", true),
        escape: None,
    })
}

// A scalar function that cannot be rewritten to match language (arithmetic, not
// a comparison/logical/is/like/in). Used to verify that logical operations and
// NOT stay in $expr language when they contain a non-rewritable element.
fn invalid_expr() -> Expression {
    Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Add,
        vec![
            *mir_field_access("foo", "int", true),
            Expression::Literal(LiteralValue::Integer(10)),
        ],
    ))
}

test_rewrite_to_match_language_no_op!(only_rewrite_is_in_match, singleton_project(valid_is()));

test_rewrite_to_match_language_no_op!(
    only_rewrite_is_if_expr_is_field_access,
    filter_stage(invalid_is())
);

test_rewrite_to_match_language_no_op!(only_rewrite_like_in_match, singleton_project(valid_like()));

test_rewrite_to_match_language_no_op!(
    only_rewrite_like_if_expr_is_field_access,
    filter_stage(invalid_like_expr())
);

test_rewrite_to_match_language_no_op!(
    only_rewrite_like_if_pattern_is_literal,
    filter_stage(invalid_like_pat())
);

test_rewrite_to_match_language_no_op!(
    cannot_rewrite_conjunction_if_any_element_is_not_rewritable,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::And,
        vec![
            valid_is(),         // rewritable
            invalid_like_pat(), // not rewritable - pattern not constant
        ]
    )))
);

test_rewrite_to_match_language_no_op!(
    cannot_rewrite_conjunction_if_it_contains_invalid_element,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::And,
        args: vec![
            valid_is(),     // rewritable
            valid_like(),   // rewritable
            invalid_expr(), // not rewritable - invalid expression
        ],
        is_nullable: true,
    }))
);

test_rewrite_to_match_language_no_op!(
    cannot_rewrite_disjunction_if_any_element_is_not_rewritable,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Or,
        vec![
            valid_is(),         // rewritable
            invalid_like_pat(), // not rewritable - pattern not constant
        ]
    )))
);

test_rewrite_to_match_language_no_op!(
    cannot_rewrite_disjunction_if_it_contains_invalid_element,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::Or,
        args: vec![
            valid_is(),     // rewritable
            valid_like(),   // rewritable
            invalid_expr(), // not rewritable - invalid expression
        ],
        is_nullable: true,
    }))
);

test_rewrite_to_match_language!(
    rewrite_valid_is,
    expected = match_filter_stage(valid_match_is()),
    expected_changed = true,
    input = filter_stage(valid_is())
);

test_rewrite_to_match_language!(
    rewrite_valid_is_null,
    expected = match_filter_stage(valid_match_is_null()),
    expected_changed = true,
    input = filter_stage(valid_is_null())
);

test_rewrite_to_match_language!(
    rewrite_valid_like_with_no_escape,
    expected = match_filter_stage(valid_match_like()),
    expected_changed = true,
    input = filter_stage(valid_like())
);

test_rewrite_to_match_language!(
    rewrite_valid_like_with_escape,
    expected = match_filter_stage(MatchQuery::Regex(MatchLanguageRegex {
        input: Some(mir_field_path("foo", vec!["str"])),
        regex: "^a_._.*%$".to_string(),
        options: "si".to_string(),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Like(LikeExpr {
        expr: mir_field_access("foo", "str", true),
        pattern: Box::new(Expression::Literal(LiteralValue::String(
            "a|__|_%|%".to_string()
        ),)),
        escape: Some('|'),
    }))
);

test_rewrite_to_match_language!(
    rewrite_valid_conjunction,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::And,
        args: vec![valid_match_like(), valid_match_is()],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::And,
        vec![valid_like(), valid_is()],
    )))
);

test_rewrite_to_match_language!(
    rewrite_valid_disjunction,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::Or,
        args: vec![valid_match_like(), valid_match_is()],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Or,
        vec![valid_like(), valid_is()],
    )))
);

test_rewrite_to_match_language!(
    rewrite_null,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Null))
);

test_rewrite_to_match_language!(
    rewrite_undefined,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Undefined))
);

test_rewrite_to_match_language!(
    rewrite_false_bool,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Boolean(false)))
);

test_rewrite_to_match_language_no_op!(
    rewrite_true_bool_is_noop,
    filter_stage(Expression::Literal(LiteralValue::Boolean(true)))
);

test_rewrite_to_match_language!(
    rewrite_0_int,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Integer(0)))
);

test_rewrite_to_match_language_no_op!(
    rewrite_1_int_is_noop,
    filter_stage(Expression::Literal(LiteralValue::Integer(1)))
);

test_rewrite_to_match_language!(
    rewrite_0_long,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Long(0)))
);

test_rewrite_to_match_language_no_op!(
    rewrite_1_long_is_noop,
    filter_stage(Expression::Literal(LiteralValue::Long(1)))
);

test_rewrite_to_match_language!(
    rewrite_0_double,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Double(0.0)))
);

test_rewrite_to_match_language_no_op!(
    rewrite_1_double_is_noop,
    filter_stage(Expression::Literal(LiteralValue::Double(1.0)))
);

test_rewrite_to_match_language!(
    rewrite_0_decimal,
    expected = match_filter_stage(MatchQuery::False(MatchFalse {
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::Literal(LiteralValue::Decimal128(
        bson::Decimal128::from_str("0.0").unwrap()
    )))
);

test_rewrite_to_match_language_no_op!(
    rewrite_1_decimal_is_noop,
    filter_stage(Expression::Literal(LiteralValue::Decimal128(
        bson::Decimal128::from_str("1.0").unwrap()
    )))
);

test_rewrite_to_match_language!(
    rewrite_in_operator_to_match_language,
    expected = match_filter_stage(MatchQuery::In(MatchLanguageIn {
        op: MatchLanguageInOp::In,
        input: mir_field_path("foo", vec!["int"]),
        values: vec![
            LiteralValue::Integer(1),
            LiteralValue::Integer(2),
            LiteralValue::Integer(3),
        ],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::In,
        args: vec![
            *mir_field_access("foo", "int", true),
            Expression::Array(ArrayExpr {
                array: vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(3)),
                ],
            })
        ],
        is_nullable: false,
    }))
);

test_rewrite_to_match_language!(
    rewrite_not_in_operator_to_match_language,
    expected = match_filter_stage(MatchQuery::In(MatchLanguageIn {
        op: MatchLanguageInOp::NotIn,
        input: mir_field_path("foo", vec!["int"]),
        values: vec![
            LiteralValue::Integer(1),
            LiteralValue::Integer(2),
            LiteralValue::Integer(3),
        ],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::NotIn,
        args: vec![
            *mir_field_access("foo", "int", true),
            Expression::Array(ArrayExpr {
                array: vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                    Expression::Literal(LiteralValue::Integer(3)),
                ],
            })
        ],
        is_nullable: false,
    }))
);

test_rewrite_to_match_language_no_op!(
    in_with_non_literal_value_in_tuple_stays_in_expr_language,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::In,
        args: vec![
            *mir_field_access("foo", "int", true),
            Expression::Array(ArrayExpr {
                array: vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    *mir_field_access("foo", "int", true),
                ],
            })
        ],
        is_nullable: false,
    }))
);

test_rewrite_to_match_language_no_op!(
    in_with_non_field_access_lhs_stays_in_expr_language,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication {
        function: ScalarFunction::In,
        args: vec![
            Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Upper,
                args: vec![*mir_field_access("foo", "str", true)],
                is_nullable: true,
            }),
            Expression::Array(ArrayExpr {
                array: vec![
                    Expression::Literal(LiteralValue::Integer(1)),
                    Expression::Literal(LiteralValue::Integer(2)),
                ],
            })
        ],
        is_nullable: false,
    }))
);

test_rewrite_to_match_language!(
    rewrite_boolean_field_access_to_equal_comparison,
    expected = match_filter_stage(MatchQuery::Comparison(MatchLanguageComparison {
        function: MatchLanguageComparisonOp::Eq,
        input: Some(mir_field_path("is_active", vec!["bool"])),
        arg: LiteralValue::Boolean(true),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(*mir_field_access("is_active", "bool", true))
);

// NOT over a rewritable leaf (a bare boolean field access). This also pins the
// expected unary shape of NOT: exactly one arg in, exactly one arg out.
test_rewrite_to_match_language!(
    rewrite_not_over_field_access,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::Not,
        args: vec![MatchQuery::Comparison(MatchLanguageComparison {
            function: MatchLanguageComparisonOp::Eq,
            input: Some(mir_field_path("foo", vec!["bool"])),
            arg: LiteralValue::Boolean(true),
            cache: SchemaCache::new(),
        })],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![*mir_field_access("foo", "bool", true)],
    )))
);

// NOT wrapping a nested logical operation.
test_rewrite_to_match_language!(
    rewrite_not_over_conjunction,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::Not,
        args: vec![MatchQuery::Logical(MatchLanguageLogical {
            op: MatchLanguageLogicalOp::And,
            args: vec![valid_match_like(), valid_match_is()],
            cache: SchemaCache::new(),
        })],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::And,
            vec![valid_like(), valid_is()],
        ))],
    )))
);

// Double negation exercises the recursion in rewrite_condition.
test_rewrite_to_match_language!(
    rewrite_nested_not,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::Not,
        args: vec![MatchQuery::Logical(MatchLanguageLogical {
            op: MatchLanguageLogicalOp::Not,
            args: vec![valid_match_is()],
            cache: SchemaCache::new(),
        })],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![Expression::ScalarFunction(ScalarFunctionApplication::new(
            ScalarFunction::Not,
            vec![valid_is()],
        ))],
    )))
);

// The inner expression is not rewritable, so the whole NOT stays in expr language.
test_rewrite_to_match_language_no_op!(
    cannot_rewrite_not_over_invalid_expr,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![invalid_expr()],
    )))
);

// A multi-hop boolean field access rewrites to an equality against `true`.
test_rewrite_to_match_language!(
    rewrite_multipart_boolean_field_access,
    expected = match_filter_stage(MatchQuery::Comparison(MatchLanguageComparison {
        function: MatchLanguageComparisonOp::Eq,
        input: Some(mir_field_path("foo", vec!["a", "b"])),
        arg: LiteralValue::Boolean(true),
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(*mir_field_access_multi_part("foo", vec!["a", "b"], true))
);

// A boolean field access composes inside a logical operation.
test_rewrite_to_match_language!(
    rewrite_boolean_field_access_in_conjunction,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::And,
        args: vec![
            MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("foo", vec!["int"])),
                arg: LiteralValue::Boolean(true),
                cache: SchemaCache::new(),
            }),
            valid_match_is(),
        ],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::And,
        vec![*mir_field_access("foo", "int", true), valid_is()],
    )))
);

// NOT is unary; a zero-arg NOT fails the arity guard and stays in expr language.
test_rewrite_to_match_language_no_op!(
    cannot_rewrite_not_with_zero_args,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![],
    )))
);

// NOT is unary; a multi-arg NOT fails the arity guard and stays in expr language
// rather than silently dropping the extra operand.
test_rewrite_to_match_language_no_op!(
    cannot_rewrite_not_with_multiple_args,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![valid_is(), valid_like()],
    )))
);

// -------------------------------------------------------------------------
// Binary comparison operators
// -------------------------------------------------------------------------

// A binary comparison `<field> <op> <literal>`.
fn comparison(function: ScalarFunction, field_nullable: bool) -> Expression {
    Expression::ScalarFunction(ScalarFunctionApplication::new(
        function,
        vec![
            *mir_field_access("foo", "int", field_nullable),
            Expression::Literal(LiteralValue::Integer(10)),
        ],
    ))
}

// A bare native comparison `{foo.int: {<op>: 10}}`.
fn match_comparison(function: MatchLanguageComparisonOp) -> MatchQuery {
    MatchQuery::Comparison(MatchLanguageComparison {
        function,
        input: Some(mir_field_path("foo", vec!["int"])),
        arg: LiteralValue::Integer(10),
        cache: SchemaCache::new(),
    })
}

// The `{foo.int: {$gt: null}}` existence guard added for nullable fields.
fn null_guard() -> MatchQuery {
    MatchQuery::Comparison(MatchLanguageComparison {
        function: MatchLanguageComparisonOp::Gt,
        input: Some(mir_field_path("foo", vec!["int"])),
        arg: LiteralValue::Null,
        cache: SchemaCache::new(),
    })
}

// A nullable comparison is guarded: `{$and: [{field: {$gt: null}}, {field: {op: 10}}]}`.
fn guarded_comparison(function: MatchLanguageComparisonOp) -> MatchQuery {
    MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::And,
        args: vec![null_guard(), match_comparison(function)],
        cache: SchemaCache::new(),
    })
}

// Over a nullable field, every comparison is wrapped in the null guard so that
// null/missing documents are excluded, matching SQL three-valued semantics.
test_rewrite_to_match_language!(
    rewrite_gt_over_nullable_field,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Gt)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Gt, true))
);

test_rewrite_to_match_language!(
    rewrite_gte_over_nullable_field,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Gte)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Gte, true))
);

test_rewrite_to_match_language!(
    rewrite_eq_over_nullable_field,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Eq)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Eq, true))
);

test_rewrite_to_match_language!(
    rewrite_lt_over_nullable_field,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Lt)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Lt, true))
);

test_rewrite_to_match_language!(
    rewrite_lte_over_nullable_field,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Lte)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Lte, true))
);

test_rewrite_to_match_language!(
    rewrite_neq_over_nullable_field,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Ne)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Neq, true))
);

// Over a non-nullable field, schema guarantees the value can never be
// null/missing, so no guard is emitted.
test_rewrite_to_match_language!(
    rewrite_gt_over_non_nullable_field_is_unguarded,
    expected = match_filter_stage(match_comparison(MatchLanguageComparisonOp::Gt)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Gt, false))
);

test_rewrite_to_match_language!(
    rewrite_lt_over_non_nullable_field_is_unguarded,
    expected = match_filter_stage(match_comparison(MatchLanguageComparisonOp::Lt)),
    expected_changed = true,
    input = filter_stage(comparison(ScalarFunction::Lt, false))
);

// Literal on the left commutes the operator so the field ends up on the input
// side: `10 < foo.int` becomes `foo.int > 10` (guarded because the field is nullable).
test_rewrite_to_match_language!(
    rewrite_comparison_commutes_literal_on_left,
    expected = match_filter_stage(guarded_comparison(MatchLanguageComparisonOp::Gt)),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Lt,
        vec![
            Expression::Literal(LiteralValue::Integer(10)),
            *mir_field_access("foo", "int", true),
        ],
    )))
);

// Conjunction of two comparisons (`x > 10 AND y > 20`), each guarded per-branch.
test_rewrite_to_match_language!(
    rewrite_conjunction_of_comparisons,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::And,
        args: vec![
            guarded_comparison(MatchLanguageComparisonOp::Gt),
            guarded_comparison(MatchLanguageComparisonOp::Lt),
        ],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::And,
        vec![
            comparison(ScalarFunction::Gt, true),
            comparison(ScalarFunction::Lt, true),
        ],
    )))
);

// Disjunction of two comparisons (`x > 10 OR y < 10`).
test_rewrite_to_match_language!(
    rewrite_disjunction_of_comparisons,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::Or,
        args: vec![
            guarded_comparison(MatchLanguageComparisonOp::Gt),
            guarded_comparison(MatchLanguageComparisonOp::Lt),
        ],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Or,
        vec![
            comparison(ScalarFunction::Gt, true),
            comparison(ScalarFunction::Lt, true),
        ],
    )))
);

// NOT over a comparison now rewrites (comparisons are rewritable leaves).
test_rewrite_to_match_language!(
    rewrite_not_over_comparison,
    expected = match_filter_stage(MatchQuery::Logical(MatchLanguageLogical {
        op: MatchLanguageLogicalOp::Not,
        args: vec![guarded_comparison(MatchLanguageComparisonOp::Gte)],
        cache: SchemaCache::new(),
    })),
    expected_changed = true,
    input = filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Not,
        vec![comparison(ScalarFunction::Gte, true)],
    )))
);

// A comparison whose non-literal side is a computed expression (not a plain
// field access) cannot map to native match language and stays in $expr.
test_rewrite_to_match_language_no_op!(
    cannot_rewrite_comparison_with_computed_side,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Gt,
        vec![
            Expression::ScalarFunction(ScalarFunctionApplication::new(
                ScalarFunction::Add,
                vec![
                    *mir_field_access("foo", "int", true),
                    Expression::Literal(LiteralValue::Integer(1)),
                ],
            )),
            Expression::Literal(LiteralValue::Integer(10)),
        ],
    )))
);

// A comparison between two field accesses cannot use native match language.
test_rewrite_to_match_language_no_op!(
    cannot_rewrite_comparison_between_two_fields,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Gt,
        vec![
            *mir_field_access("foo", "int", true),
            *mir_field_access("foo", "str", true),
        ],
    )))
);

// `Between` is not one of the six rewritable comparison operators.
test_rewrite_to_match_language_no_op!(
    cannot_rewrite_between,
    filter_stage(Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Between,
        vec![
            *mir_field_access("foo", "int", true),
            Expression::Literal(LiteralValue::Integer(1)),
            Expression::Literal(LiteralValue::Integer(10)),
        ],
    )))
);
