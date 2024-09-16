use crate::{
    catalog::{Catalog, Namespace},
    map,
    mir::{
        optimizer::{rewrite_to_match_language::MatchLanguageRewriter, Optimizer},
        schema::{SchemaCache, SchemaInferenceState},
        *,
    },
    schema::{Atomic, Document, Schema, SchemaEnvironment},
    set, unchecked_unique_linked_hash_map,
    util::{mir_collection, mir_field_access, mir_field_path},
    SchemaCheckingMode,
};
use lazy_static::lazy_static;

lazy_static! {
    static ref CATALOG: Catalog = Catalog::new(map! {
        Namespace {db: "db".to_string(), collection: "foo".to_string()} => Schema::Document(Document {
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
    Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
        source: mir_collection("db", "foo"),
        condition,
        cache: SchemaCache::new(),
    }))
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

fn comp_expr() -> Expression {
    Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Lt,
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
            valid_is(),   // rewritable
            valid_like(), // rewritable
            comp_expr(),  // not rewritable - invalid expression
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
            valid_is(),   // rewritable
            valid_like(), // rewritable
            comp_expr(),  // not rewritable - invalid expression
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
