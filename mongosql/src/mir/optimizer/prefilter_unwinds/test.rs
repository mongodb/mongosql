use crate::mir::{self, binding_tuple::Key};

fn mir_reference(name: &str) -> mir::Expression {
    mir::Expression::Reference(mir::ReferenceExpr {
        key: if name == "__bot__" {
            mir::binding_tuple::Key::bot(0)
        } else {
            mir::binding_tuple::Key::named(name, 0)
        },
        cache: mir::schema::SchemaCache::new(),
    })
}

fn mir_field_access(ref_name: &str, field_name: &str) -> mir::Expression {
    mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir_reference(ref_name)),
        field: field_name.to_string(),
        cache: mir::schema::SchemaCache::new(),
    })
}

fn match_path(ref_name: &str, field_name: &str) -> mir::MatchPath {
    mir::MatchPath::MatchFieldAccess(mir::MatchFieldAccess {
        parent: mir::MatchPath::MatchReference(mir::ReferenceExpr {
            key: (Key::named(ref_name, 0)),
            cache: mir::schema::SchemaCache::new(),
        })
        .into(),
        field: field_name.to_string(),
        cache: mir::schema::SchemaCache::new(),
    })
}

macro_rules! test_prefilter {
    ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::mir::{
                self,
                binding_tuple::{BindingTuple, Key},
                optimizer::prefilter_unwinds::PrefilterUnwindsVisitor,
                schema::SchemaCache,
                visitor::Visitor,
                ElemMatch,
                Expression::{self, *},
                Filter, Group, Join, JoinType, Limit, LiteralExpr,
                LiteralValue::*,
                MQLStage, MatchFilter, MatchLanguageComparison, MatchLanguageComparisonOp,
                MatchLanguageLogical, MatchLanguageLogicalOp, MatchQuery, ScalarFunction,
                ScalarFunctionApplication, Stage, Unwind,
            };
            #[allow(unused)]
            let input = $input;
            let expected = $expected;
            let mut visitor = PrefilterUnwindsVisitor {};
            assert_eq!(expected, visitor.visit_stage(input));
        }
    };
}

macro_rules! test_prefilter_no_op {
    ($func_name:ident, $input:expr,) => {
        test_prefilter! { $func_name, expected = $input, input = $input, }
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
                source: Stage::Unwind( Unwind {
                    source: Stage::MQLIntrinsic(MQLStage::MatchFilter( MatchFilter {
                        source: Stage::Sentinel.into(),
                        condition: MatchQuery::ElemMatch(
                            ElemMatch {
                                input: match_path("foo", "bar"),
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
                    })).into(),
                    path: mir_field_access("foo", "bar").into(),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42).into(),
                            )
                        ],
                        cache: SchemaCache::new(),
                    }),
                cache: SchemaCache::new(),
        }),
    input = Stage::Filter(Filter {
                source: Stage::Unwind( Unwind {
                    source: Stage::Sentinel.into(),
                    path: mir_field_access("foo", "bar").into(),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication {
                        function: ScalarFunction::Eq,
                        args: vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42).into(),
                            )
                        ],
                        cache: SchemaCache::new(),
                    }),
                cache: SchemaCache::new(),
        }),
}

test_prefilter_no_op! {
    eq_index_cannot_prefilter,
    Stage::Filter(Filter {
            source: Stage::Unwind( Unwind {
                source: Stage::Sentinel.into(),
                path: mir_field_access("foo", "bar").into(),
                index: Some("idx".to_string()),
                outer: false,
                cache: SchemaCache::new(),
            }).into(),
            condition: Expression::ScalarFunction(
                ScalarFunctionApplication {
                    function: ScalarFunction::Eq,
                    args: vec![
                        mir_field_access("foo", "idx"),
                        Expression::Literal(
                            Integer(42).into(),
                        )
                    ],
                    cache: SchemaCache::new(),
                }),
            cache: SchemaCache::new(),
    }),
}

test_prefilter! {
    between_path,
    expected = Stage::Filter(Filter {
                source: Stage::Unwind( Unwind {
                    source: Stage::MQLIntrinsic(MQLStage::MatchFilter( MatchFilter {
                        source: Stage::Sentinel.into(),
                        condition: MatchQuery::ElemMatch(
                            ElemMatch {
                                input: match_path("foo", "bar"),
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
                    })).into(),
                    path: mir_field_access("foo", "bar").into(),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication {
                        function: ScalarFunction::Between,
                        args: vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42).into(),
                            ),
                            Expression::Literal(
                                Integer(46).into(),
                            )
                        ],
                        cache: SchemaCache::new(),
                    }),
                cache: SchemaCache::new(),
        }),
    input = Stage::Filter(Filter {
                source: Stage::Unwind( Unwind {
                    source: Stage::Sentinel.into(),
                    path: mir_field_access("foo", "bar").into(),
                    index: Some("idx".to_string()),
                    outer: false,
                    cache: SchemaCache::new(),
                }).into(),
                condition: Expression::ScalarFunction(
                    ScalarFunctionApplication {
                        function: ScalarFunction::Between,
                        args: vec![
                            mir_field_access("foo", "bar"),
                            Expression::Literal(
                                Integer(42).into(),
                            ),
                            Expression::Literal(
                                Integer(46).into(),
                            )
                        ],
                        cache: SchemaCache::new(),
                    }),
                cache: SchemaCache::new(),
        }),
}

test_prefilter_no_op! {
    between_index_cannot_prefilter,
    Stage::Filter(Filter {
            source: Stage::Unwind( Unwind {
                source: Stage::Sentinel.into(),
                path: mir_field_access("foo", "bar").into(),
                index: Some("idx".to_string()),
                outer: false,
                cache: SchemaCache::new(),
            }).into(),
            condition: Expression::ScalarFunction(
                ScalarFunctionApplication {
                    function: ScalarFunction::Between,
                    args: vec![
                        mir_field_access("foo", "idx"),
                        Expression::Literal(
                            Integer(42).into(),
                        ),
                        Expression::Literal(
                            Integer(46).into(),
                        )
                    ],
                    cache: SchemaCache::new(),
                }),
            cache: SchemaCache::new(),
    }),
}

test_prefilter_no_op! {
    // this actually should not happen if stage_movement is run before this because the Filter
    // would already be moved before the Unwind, but we want to make sure this pass works in
    // isolation
    eq_wrong_generates_no_prefilter,
    Stage::Filter(Filter {
            source: Stage::Unwind( Unwind {
                source: Stage::Sentinel.into(),
                path: mir_field_access("foo", "bar").into(),
                index: Some("idx".to_string()),
                outer: false,
                cache: SchemaCache::new(),
            }).into(),
            condition: Expression::ScalarFunction(
                ScalarFunctionApplication {
                    function: ScalarFunction::Eq,
                    args: vec![
                        mir_field_access("foo", "i"),
                        Expression::Literal(
                            Integer(42).into(),
                        )
                    ],
                    cache: SchemaCache::new(),
                }),
            cache: SchemaCache::new(),
    }),
}

test_prefilter_no_op! {
    single_field_use_in_non_simple_expr,
    Stage::Filter(Filter {
        source: Stage::Unwind( Unwind {
            source: Stage::Sentinel.into(),
            path: mir_field_access("foo", "bar").into(),
            index: Some("idx".to_string()),
            outer: false,
            cache: SchemaCache::new(),
        }).into(),
        condition: Expression::ScalarFunction(
            ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    Expression::ScalarFunction( ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![ mir_field_access("foo", "bar"), Expression::Literal(Integer(1).into()) ],
                        cache: SchemaCache::new(),
                    } ),
                    Expression::Literal( Integer(42).into(), )
                ],
                cache: SchemaCache::new(),
            }),
        cache: SchemaCache::new(),
    }),
}

test_prefilter_no_op! {
    between_single_field_use_in_non_simple_expr,
    Stage::Filter(Filter {
        source: Stage::Unwind( Unwind {
            source: Stage::Sentinel.into(),
            path: mir_field_access("foo", "bar").into(),
            index: Some("idx".to_string()),
            outer: false,
            cache: SchemaCache::new(),
        }).into(),
        condition: Expression::ScalarFunction(
            ScalarFunctionApplication {
                function: ScalarFunction::Between,
                args: vec![
                    Expression::ScalarFunction( ScalarFunctionApplication{
                        function: ScalarFunction::Add,
                        args: vec![mir_field_access("foo", "idx")],
                        cache: SchemaCache::new(),
                    }),
                    Expression::Literal(
                        Integer(42).into(),
                    ),
                    Expression::Literal(
                        Integer(46).into(),
                    )
                ],
                cache: SchemaCache::new(),
            }),
        cache: SchemaCache::new(),
    }),
}
