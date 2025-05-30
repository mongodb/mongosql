use crate::{mir, util::mir_field_path};

fn mir_reference(name: &str) -> mir::Expression {
    mir::Expression::Reference(mir::ReferenceExpr {
        key: if name == "__bot__" {
            mir::binding_tuple::Key::bot(0)
        } else {
            mir::binding_tuple::Key::named(name, 0)
        },
    })
}

fn mir_field_access(ref_name: &str, field_name: &str) -> mir::Expression {
    mir::Expression::FieldAccess(mir::FieldAccess::new(
        Box::new(mir_reference(ref_name)),
        field_name.to_string(),
    ))
}

macro_rules! test_prefilter {
    ($func_name:ident, expected = $expected:expr, expected_changed = $expected_changed:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::mir::{
                self,
                binding_tuple::{BindingTuple, Key},
                optimizer::prefilter_unwinds::PrefilterUnwindsOptimizer,
                schema::SchemaCache,
                visitor::Visitor,
                ElemMatch,
                Expression::{self, *},
                Filter, Group, Join, JoinType, Limit,
                LiteralValue::*,
                MatchFilter, MatchLanguageComparison, MatchLanguageComparisonOp,
                MatchLanguageLogical, MatchLanguageLogicalOp, MatchQuery, MqlStage, ScalarFunction,
                ScalarFunctionApplication, Stage, Unwind,
            };
            #[allow(unused)]
            let input = $input;
            let expected = $expected;
            let (actual, actual_changed) = PrefilterUnwindsOptimizer::prefilter_unwinds(input);
            assert_eq!($expected_changed, actual_changed);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_prefilter_no_op {
    ($func_name:ident, $input:expr,) => {
        test_prefilter! { $func_name, expected = $input, expected_changed = false, input = $input, }
    };
}

#[test]
fn match_comparison_op() {
    assert_eq!(
        mir::MatchLanguageComparisonOp::Lt,
        mir::ScalarFunction::Lt.try_into().unwrap(),
    );
    assert_eq!(
        mir::MatchLanguageComparisonOp::Lte,
        mir::ScalarFunction::Lte.try_into().unwrap(),
    );
    assert_eq!(
        mir::MatchLanguageComparisonOp::Gt,
        mir::ScalarFunction::Gt.try_into().unwrap(),
    );
    assert_eq!(
        mir::MatchLanguageComparisonOp::Gte,
        mir::ScalarFunction::Gte.try_into().unwrap(),
    );
    assert_eq!(
        mir::MatchLanguageComparisonOp::Ne,
        mir::ScalarFunction::Neq.try_into().unwrap(),
    );
    assert_eq!(
        mir::MatchLanguageComparisonOp::Eq,
        mir::ScalarFunction::Eq.try_into().unwrap(),
    );
}

test_prefilter! {
    eq_path,
    expected = Stage::Filter(Filter {
                source: Stage::Unwind(Unwind {
                    source: Stage::MqlIntrinsic(MqlStage::MatchFilter( Box::new(MatchFilter {
                        source: Stage::Sentinel.into(),
                        condition: MatchQuery::ElemMatch(
                            ElemMatch {
                                input: mir_field_path("foo", vec!["bar"]),
                                condition: MatchQuery::Comparison(
                                    MatchLanguageComparison {
                                        function: MatchLanguageComparisonOp::Eq,
                                        input: None,
                                        arg: Integer(42),
                                        cache: SchemaCache::new(),
                                    }).into(),
                                cache: SchemaCache::new(),
                            }),
                            cache: SchemaCache::new(),
                    }))).into(),
                    path: mir_field_path("foo", vec!["bar"]),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                    is_prefiltered: true,
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication::new(ScalarFunction::Eq,vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42),
                            )
                        ],)),
                cache: SchemaCache::new(),
        }),
    expected_changed = true,
    input = Stage::Filter(Filter {
                source: Stage::Unwind(Unwind {
                    source: Stage::Sentinel.into(),
                    path: mir_field_path("foo", vec!["bar"]),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                    is_prefiltered: false,
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication::new(ScalarFunction::Eq,vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42),
                            )
                        ],)),
                cache: SchemaCache::new(),
        }),
}

test_prefilter_no_op! {
    eq_index_cannot_prefilter,
    Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Sentinel.into(),
                path: mir_field_path("foo", vec!["bar"]),
                index: Some("idx".to_string()),
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            }).into(),
            condition: Expression::ScalarFunction(
                ScalarFunctionApplication::new(ScalarFunction::Eq,vec![
                        mir_field_access("foo", "idx"),
                        Expression::Literal(
                            Integer(42),
                        )
                    ],)),
            cache: SchemaCache::new(),
    }),
}

test_prefilter! {
    between_path,
    expected = Stage::Filter(Filter {
                source: Stage::Unwind(Unwind {
                    source: Stage::MqlIntrinsic(MqlStage::MatchFilter( Box::new(MatchFilter {
                        source: Stage::Sentinel.into(),
                        condition: MatchQuery::ElemMatch(
                            ElemMatch {
                                input: mir_field_path("foo", vec!["bar"]),
                                condition:
                                    MatchQuery::Logical( MatchLanguageLogical {
                                        op: MatchLanguageLogicalOp::And,
                                        args: vec![
                                            MatchQuery::Comparison(
                                            MatchLanguageComparison {
                                                function: MatchLanguageComparisonOp::Gte,
                                                input: None,
                                                arg: Integer(42),
                                                cache: SchemaCache::new(),
                                            }),
                                            MatchQuery::Comparison(
                                            MatchLanguageComparison {
                                                function: MatchLanguageComparisonOp::Lte,
                                                input: None,
                                                arg: Integer(46),
                                                cache: SchemaCache::new(),
                                            }),

                                        ],
                                        cache: SchemaCache::new(),
                                }).into(),
                                cache: SchemaCache::new(),
                            }),
                            cache: SchemaCache::new(),
                    }))).into(),
                    path: mir_field_path("foo", vec!["bar"]),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                    is_prefiltered: true,
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication::new(ScalarFunction::Between,vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42),
                            ),
                            Expression::Literal(
                                Integer(46),
                            )
                        ],)),
                cache: SchemaCache::new(),
        }),
    expected_changed = true,
    input = Stage::Filter(Filter {
                source: Stage::Unwind(Unwind {
                    source: Stage::Sentinel.into(),
                    path: mir_field_path("foo", vec!["bar"]),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                    is_prefiltered: false,
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication::new(ScalarFunction::Between,vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42),
                            ),
                            Expression::Literal(
                                Integer(46),
                            )
                        ],)),
                cache: SchemaCache::new(),
        }),
}

test_prefilter_no_op! {
    between_index_cannot_prefilter,
    Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Sentinel.into(),
                path: mir_field_path("foo", vec!["bar"]),
                index: Some("idx".to_string()),
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            }).into(),
            condition: Expression::ScalarFunction(
                ScalarFunctionApplication::new(ScalarFunction::Between,vec![
                        mir_field_access("foo", "idx"),
                        Expression::Literal(
                            Integer(42),
                        ),
                        Expression::Literal(
                            Integer(46),
                        )
                    ],)),
            cache: SchemaCache::new(),
    }),
}

test_prefilter_no_op! {
    // this actually should not happen if stage_movement is run before this because the Filter
    // would already be moved before the Unwind, but we want to make sure this pass works in
    // isolation
    eq_wrong_generates_no_prefilter,
    Stage::Filter(Filter {
            source: Stage::Unwind(Unwind {
                source: Stage::Sentinel.into(),
                path: mir_field_path("foo", vec!["bar"]),
                index: Some("idx".to_string()),
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            }).into(),
            condition: Expression::ScalarFunction(
                ScalarFunctionApplication::new(ScalarFunction::Eq,vec![
                        mir_field_access("foo", "i"),
                        Expression::Literal(
                            Integer(42),
                        )
                    ],)),
            cache: SchemaCache::new(),
    }),
}

test_prefilter_no_op! {
    single_field_use_in_non_simple_expr,
    Stage::Filter(Filter {
        source: Stage::Unwind(Unwind {
            source: Stage::Sentinel.into(),
            path: mir_field_path("foo", vec!["bar"]),
            index: Some("idx".to_string()),
            outer: false,
            cache: SchemaCache::new(),
            is_prefiltered: false,
        }).into(),
        condition: Expression::ScalarFunction(
            ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::ScalarFunction( ScalarFunctionApplication::new(ScalarFunction::Add,vec![ mir_field_access("foo", "bar"), Expression::Literal(Integer(1)) ],) ),
                    Expression::Literal( Integer(42), )
                ],
                is_nullable: false,
            }),
        cache: SchemaCache::new(),
    }),
}

test_prefilter_no_op! {
    between_single_field_use_in_non_simple_expr,
    Stage::Filter(Filter {
        source: Stage::Unwind(Unwind {
            source: Stage::Sentinel.into(),
            path: mir_field_path("foo", vec!["bar"]),
            index: Some("idx".to_string()),
            outer: false,
            cache: SchemaCache::new(),
            is_prefiltered: false,
        }).into(),
        condition: Expression::ScalarFunction(
            ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::ScalarFunction( ScalarFunctionApplication{
                        function: ScalarFunction::Add,
                        args: vec![mir_field_access("foo", "idx")],
                        is_nullable: false,
                    }),
                    Expression::Literal(
                        Integer(42),
                    ),
                    Expression::Literal(
                        Integer(46),
                    )
                ],
                is_nullable: false,
            }),
        cache: SchemaCache::new(),
    }),
}
