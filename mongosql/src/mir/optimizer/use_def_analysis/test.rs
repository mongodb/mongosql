use crate::{
    mir::{self, optimizer::use_def_analysis::FieldPath},
    unchecked_unique_linked_hash_map,
    util::{mir_collection, mir_field_access, mir_field_access_multi_part, mir_field_path},
};

macro_rules! test_method {
    ($func_name:ident, method = $method:ident, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                map,
                mir::{
                    self,
                    binding_tuple::Key,
                    schema::SchemaCache,
                    Collection,
                    Expression::{self, *},
                    Filter, Group,
                    LiteralValue::*,
                    Project, Stage, Unwind,
                },
                set, unchecked_unique_linked_hash_map,
            };
            #[allow(unused)]
            use std::collections::{HashMap, HashSet};
            let input = $input;
            let expected = $expected;
            let actual = input.$method();
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_method_uses {
    ($func_name:ident, $expected:expr, $input:expr, $method:ident,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                map,
                mir::{
                    self,
                    binding_tuple::Key,
                    optimizer::use_def_analysis::FieldPath,
                    schema::SchemaCache,
                    AggregationExpr, AggregationFunction, AggregationFunctionApplication,
                    AliasedAggregation, AliasedExpr, Collection, ExistsExpr,
                    Expression::{self, *},
                    Filter, Group,
                    LiteralValue::*,
                    MQLStage, MatchFilter, MatchLanguageComparison, MatchLanguageComparisonOp,
                    MatchLanguageLogical, MatchLanguageLogicalOp, MatchQuery,
                    OptionallyAliasedExpr, Project, ReferenceExpr, ScalarFunction,
                    ScalarFunctionApplication, Sort, SortSpecification, Stage, SubqueryComparison,
                    SubqueryComparisonOp, SubqueryExpr, SubqueryModifier,
                },
                set, unchecked_unique_linked_hash_map,
            };
            #[allow(unused)]
            use std::collections::{HashMap, HashSet};
            let input = $input;
            let expected = $expected;
            let actual = input.$method().0;
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_field_uses {
    ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
        test_method_uses! {$func_name, $expected, $input, field_uses,}
    };
}

macro_rules! test_datasource_uses {
    ($func_name:ident, expected = $expected:expr, input = $input:expr,) => {
        test_method_uses! {$func_name, $expected, $input, datasource_uses,}
    };
}

macro_rules! test_substitute {
    ($func_name:ident, expected = $expected:expr, stage = $input:expr, theta = $theta:expr,) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                map,
                mir::{
                    self,
                    binding_tuple::Key,
                    schema::SchemaCache,
                    AggregationExpr, AggregationFunction, AggregationFunctionApplication,
                    AliasedAggregation, AliasedExpr, Collection,
                    Expression::{self, *},
                    Filter, Group,
                    LiteralValue::*,
                    MQLStage, MatchFilter, MatchLanguageComparison, MatchLanguageComparisonOp,
                    MatchQuery, OptionallyAliasedExpr, ReferenceExpr, ScalarFunction,
                    ScalarFunctionApplication, Sort, SortSpecification, Stage,
                },
                set,
            };
            #[allow(unused)]
            use std::collections::{HashMap, HashSet};
            let input = $input;
            let theta = $theta;
            // if None is expected, we convert that to Err(input.clone()
            // it keeps the test cases smaller
            let expected = $expected.ok_or(input.clone());
            let actual = input.substitute(theta);
            assert_eq!(expected, actual);
        }
    };
}

fn mir_int_key(alias: &str, i: i32) -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
        alias: alias.into(),
        expr: mir::Expression::Literal(mir::LiteralValue::Integer(i)),
    })
}

fn mir_unalaised_key() -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Unaliased(mir::Expression::Reference(mir::ReferenceExpr {
        key: mir::binding_tuple::Key::named("foo", 0),
    }))
}

fn mir_int_expr(i: i32) -> mir::Expression {
    mir::Expression::Literal(mir::LiteralValue::Integer(i))
}

fn mir_count_agg(alias: &str) -> mir::AliasedAggregation {
    mir::AliasedAggregation {
        alias: alias.into(),
        agg_expr: mir::AggregationExpr::CountStar(false),
    }
}

fn mir_reference(name: &str) -> mir::Expression {
    mir::Expression::Reference(mir::ReferenceExpr {
        key: mir::binding_tuple::Key::named(name, 0),
    })
}

mod defines {
    use super::*;

    test_method!(
        project,
        method = defines,
        expected = {
            let expected: HashMap<Key, Expression> = map! {
                Key::named("x", 0) => Literal(Integer(0),),
                Key::bot(0) => Literal(Integer(0),),
            };
            expected
        },
        input = Project {
            is_add_fields: false,
            source: mir_collection("foo", "bar"),
            expression: map! {
                Key::named("x", 0) => Literal(Integer(0),),
                Key::bot(0) => Literal(Integer(0),)
            },
            cache: SchemaCache::new(),
        },
    );

    test_method!(
        group,
        method = defines,
        expected = {
            let expected: HashMap<Key, Expression> = map! {
                Key::bot(0) => Document(unchecked_unique_linked_hash_map! {
                    "a".to_string() => mir_int_expr(1),
                    "b".to_string() => mir_int_expr(2),
                }.into()),
            };
            expected
        },
        input = Group {
            source: mir_collection("foo", "bar"),
            keys: vec![
                mir_int_key("a", 1),
                mir_int_key("b", 2),
                mir_unalaised_key()
            ],
            aggregations: vec![mir_count_agg("agg1"), mir_count_agg("agg2")],
            cache: SchemaCache::new(),
            scope: 0,
        },
    );

    test_method!(
        group_opaque_fields,
        method = opaque_field_defines,
        expected = {
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("__bot__", vec!["agg1"]),
                mir_field_path("__bot__", vec!["agg2"]),
            };
            expected
        },
        input = Group {
            source: mir_collection("foo", "bar"),
            keys: vec![
                mir_int_key("a", 1),
                mir_int_key("b", 2),
                mir_unalaised_key()
            ],
            aggregations: vec![mir_count_agg("agg1"), mir_count_agg("agg2")],
            cache: SchemaCache::new(),
            scope: 0,
        },
    );

    test_method!(
        unwind_opaque_fields,
        method = opaque_field_defines,
        expected = {
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["bar", "arr"]),
                mir_field_path("foo", vec!["idx"]),
            };
            expected
        },
        input = Unwind {
            source: mir_collection("foo", "bar"),
            path: mir_field_path("foo", vec!["bar", "arr"]),
            index: Some("idx".to_string()),
            outer: false,
            cache: SchemaCache::new(),
            is_prefiltered: false,
        },
    );
}

mod field_uses {
    use super::*;

    test_field_uses!(
        filter_simple_field_accesses,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["x"]),
                mir_field_path("bar", vec!["x"]),
                mir_field_path("bar", vec!["y"]),
            };
            expected
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    *mir_field_access("foo", "x", true),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            *mir_field_access("bar", "y", true),
                            *mir_field_access("bar", "x", true),
                        ],
                        is_nullable: true,
                    })
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_field_uses!(
        filter_nested_field_accesses,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["x"]),
                mir_field_path("bar", vec!["x", "y", "z"]),
                mir_field_path("bar", vec!["x", "y"]),
                mir_field_path("bar", vec!["x"]),
                mir_field_path("bar", vec!["a", "b"]),
                mir_field_path("bar", vec!["a"]),
            };
            expected
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    *mir_field_access("foo", "x", true),
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::Add,
                        args: vec![
                            *mir_field_access_multi_part("bar", vec!["x", "y", "z"], true),
                            *mir_field_access_multi_part("bar", vec!["a", "b"], true),
                        ],
                        is_nullable: true,
                    })
                ],
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_field_uses!(
        match_filter_simple_field_paths,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["x"]),
                mir_field_path("bar", vec!["x"]),
            };
            expected
        }),
        input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Logical(MatchLanguageLogical {
                op: MatchLanguageLogicalOp::Or,
                args: vec![
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("foo", vec!["x"])),
                        arg: Integer(0),
                        cache: SchemaCache::new(),
                    }),
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("bar", vec!["x"])),
                        arg: Integer(0),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })),
    );

    test_field_uses!(
        match_filter_nested_field_paths,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["x", "y"]),
                mir_field_path("foo", vec!["x"]),
                mir_field_path("bar", vec!["x", "y", "z"]),
                mir_field_path("bar", vec!["x", "y"]),
                mir_field_path("bar", vec!["x"]),
            };
            expected
        }),
        input = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Logical(MatchLanguageLogical {
                op: MatchLanguageLogicalOp::Or,
                args: vec![
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("foo", vec!["x", "y"])),
                        arg: Integer(0),
                        cache: SchemaCache::new(),
                    }),
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(mir_field_path("bar", vec!["x", "y", "z"])),
                        arg: Integer(0),
                        cache: SchemaCache::new(),
                    }),
                ],
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })),
    );

    test_field_uses!(
        sort_simple_field_paths,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["x"]),
                mir_field_path("foo", vec!["y"]),
            };
            expected
        }),
        input = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("foo", vec!["x"])),
                SortSpecification::Desc(mir_field_path("foo", vec!["y"])),
            ],
            cache: SchemaCache::new(),
        }),
    );

    test_field_uses!(
        sort_nested_field_paths,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("foo", vec!["a"]),
                mir_field_path("foo", vec!["x", "y", "z"]),
                mir_field_path("foo", vec!["x", "y"]),
                mir_field_path("foo", vec!["x"]),
            };
            expected
        }),
        input = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("foo", vec!["x", "y", "z"])),
                SortSpecification::Desc(mir_field_path("foo", vec!["a"])),
            ],
            cache: SchemaCache::new(),
        }),
    );

    test_field_uses!(
        field_accesses_and_paths_in_subquery_expr_captured,
        // Note that the output_expr field is omitted since that field is actually being _defined_
        // by the subquery expr, rather than "used".
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("bar", vec!["a", "b", "c"]),
                mir_field_path("bar", vec!["a", "b"]),
                mir_field_path("bar", vec!["a"]),
                mir_field_path("bar", vec!["x"]),
                mir_field_path("foo", vec!["x", "y", "z"]),
                mir_field_path("foo", vec!["x", "y"]),
                mir_field_path("foo", vec!["x"]),
                mir_field_path("foo", vec!["a"]),
            };
            expected
        }),
        // This Filter is semantically nonsensical, but that is not quite the point. The point is
        // it is a minimal SubqueryExpr that contains simple and nested FieldAccess and FieldPath
        // expressions.
        input = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::Subquery(SubqueryExpr {
                output_expr: mir_field_access("sub", "x", true),
                subquery: Box::new(Stage::Sort(Sort {
                    source: Box::new(Stage::Project(Project {
                        source: mir_collection("sub", "col"),
                        expression: map! {
                            Key::named("x", 1) => *mir_field_access_multi_part("bar", vec!["a", "b", "c"], true),
                            Key::named("y", 1) => *mir_field_access_multi_part("bar", vec!["x"], true),
                        },
                        is_add_fields: false,
                        cache: SchemaCache::new(),
                    })),
                    specs: vec![
                        SortSpecification::Asc(mir_field_path("foo", vec!["x", "y", "z"])),
                        SortSpecification::Desc(mir_field_path("foo", vec!["a"])),
                    ],
                    cache: SchemaCache::new(),
                })),
                is_nullable: true
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_field_uses!(
        field_accesses_and_paths_in_subquery_comparison_captured,
        // Again, note that the output_expr field is omitted since that field is actually being
        // _defined_ by the subquery expr, rather than "used".
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("bar", vec!["a", "b", "c"]),
                mir_field_path("bar", vec!["a", "b"]),
                mir_field_path("bar", vec!["a"]),
                mir_field_path("bar", vec!["x"]),
                mir_field_path("foo", vec!["x", "y", "z"]),
                mir_field_path("foo", vec!["x", "y"]),
                mir_field_path("foo", vec!["x"]),
                mir_field_path("foo", vec!["a"]),
            };
            expected
        }),
        // This Filter is also semantically nonsensical, but that is not quite the point. The point
        // is it is a minimal SubqueryComparison that contains simple and nested FieldAccess and
        // FieldPath expressions.
        input = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::SubqueryComparison(SubqueryComparison {
                operator: SubqueryComparisonOp::Eq,
                modifier: SubqueryModifier::Any,
                argument: mir_field_access("foo", "a", true),
                subquery_expr: SubqueryExpr {
                    output_expr: mir_field_access("sub", "x", true),
                    subquery: Box::new(Stage::Sort(Sort {
                        source: Box::new(Stage::Project(Project {
                            source: mir_collection("sub", "col"),
                            expression: map! {
                                Key::named("y", 1) => *mir_field_access_multi_part("bar", vec!["x"], true),
                                Key::named("x", 1) => *mir_field_access_multi_part("bar", vec!["a", "b", "c"], true),
                            },
                            is_add_fields: false,
                            cache: SchemaCache::new(),
                        })),
                        specs: vec![SortSpecification::Asc(mir_field_path(
                            "foo",
                            vec!["x", "y", "z"]
                        )),],
                        cache: SchemaCache::new(),
                    })),
                    is_nullable: true
                },
                is_nullable: true,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_field_uses!(
        field_accesses_and_paths_in_exists_expr_captured,
        expected = Some({
            let expected: HashSet<FieldPath> = set! {
                mir_field_path("bar", vec!["a", "b", "c"]),
                mir_field_path("bar", vec!["a", "b"]),
                mir_field_path("bar", vec!["a"]),
                mir_field_path("bar", vec!["x"]),
                mir_field_path("foo", vec!["x", "y", "z"]),
                mir_field_path("foo", vec!["x", "y"]),
                mir_field_path("foo", vec!["x"]),
                mir_field_path("foo", vec!["a"]),
            };
            expected
        }),
        input = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::Exists(ExistsExpr {
                stage: Box::new(Stage::Sort(Sort {
                    source: Box::new(Stage::Project(Project {
                        source: mir_collection("sub", "col"),
                        expression: map! {
                            Key::named("x", 1) => *mir_field_access_multi_part("bar", vec!["a", "b", "c"], true),
                            Key::named("y", 1) => *mir_field_access_multi_part("bar", vec!["x"], true),
                        },
                        is_add_fields: false,
                        cache: SchemaCache::new(),
                    })),
                    specs: vec![
                        SortSpecification::Asc(mir_field_path("foo", vec!["x", "y", "z"])),
                        SortSpecification::Desc(mir_field_path("foo", vec!["a"])),
                    ],
                    cache: SchemaCache::new(),
                }))
            }),
            cache: SchemaCache::new(),
        }),
    );
}

mod datasource_uses {
    use super::*;

    test_datasource_uses!(
        filter,
        expected = {
            let expected: HashSet<Key> = set![Key::named("x", 0), Key::named("y", 0),];
            expected
        },
        input = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    mir_reference("x"),
                    Expression::ScalarFunction(ScalarFunctionApplication::new(
                        ScalarFunction::Add,
                        vec![mir_reference("y"), mir_reference("x"),],
                    ))
                ],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
    );

    test_datasource_uses!(
        sort,
        expected = {
            let expected: HashSet<Key> = set![Key::named("x", 0), Key::named("y", 0),];
            expected
        },
        input = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("x", vec!["a"])),
                SortSpecification::Desc(mir_field_path("y", vec!["b"])),
            ],
            cache: SchemaCache::new(),
        }),
    );

    test_datasource_uses!(
        group,
        expected = {
            let expected: HashSet<Key> =
                set![Key::named("x", 0), Key::named("y", 0), Key::named("z", 0)];
            expected
        },
        input = Stage::Group(Group {
            source: mir_collection("foo", "bar"),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "a".to_string(),
                    expr: mir_reference("x"),
                }),
                OptionallyAliasedExpr::Unaliased(mir_reference("y")),
            ],
            aggregations: vec![AliasedAggregation {
                alias: "agg".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: mir_reference("z").into(),
                }),
            }],
            scope: 0u16,
            cache: SchemaCache::new(),
        }),
    );
}

mod substitute {
    use super::*;

    test_substitute!(
        filter,
        expected = Some(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![mir_int_expr(42), mir_int_expr(55),],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        })),
        stage = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![mir_reference("x"), mir_reference("y"),],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("x",0) => mir_int_expr(42),
            Key::named("y",0) => mir_int_expr(55),
        },
    );

    test_substitute!(
        sort_doc,
        expected = Some(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("bar", vec!["x2", "a", "b"])),
                SortSpecification::Desc(mir_field_path("bar", vec!["y2"])),
            ],
            cache: SchemaCache::new(),
        })),
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("foo", vec!["x", "a", "b"])),
                SortSpecification::Desc(mir_field_path("foo", vec!["y"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("foo",0) =>
                mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "x".to_string() => *mir_field_access("bar", "x2", true),
                        "y".to_string() => *mir_field_access("bar", "y2", true),
                    },
              })
        },
    );

    test_substitute!(
        sort_access,
        expected = Some(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("bar", vec!["a", "x", "a"])),
                SortSpecification::Desc(mir_field_path("bar", vec!["a", "y"])),
            ],
            cache: SchemaCache::new(),
        })),
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("foo", vec!["x", "a"])),
                SortSpecification::Desc(mir_field_path("foo", vec!["y"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("foo",0) => *mir_field_access("bar", "a", true),
        },
    );

    test_substitute!(
        sort_ref,
        expected = Some(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("bar", vec!["x"])),
                SortSpecification::Desc(mir_field_path("bar", vec!["y"])),
            ],
            cache: SchemaCache::new(),
        })),
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("foo", vec!["x"])),
                SortSpecification::Desc(mir_field_path("foo", vec!["y"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("foo",0) => mir_reference("bar"),
        },
    );

    test_substitute!(
        group,
        expected = Some(Stage::Group(Group {
            source: mir_collection("foo", "bar"),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "a".to_string(),
                    expr: mir_int_expr(42),
                }),
                OptionallyAliasedExpr::Unaliased(mir_int_expr(55)),
            ],
            aggregations: vec![AliasedAggregation {
                alias: "agg".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: mir_int_expr(0).into(),
                }),
            }],
            scope: 0,
            cache: SchemaCache::new(),
        })),
        stage = Stage::Group(Group {
            source: mir_collection("foo", "bar"),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "a".to_string(),
                    expr: mir_reference("x"),
                }),
                OptionallyAliasedExpr::Unaliased(mir_reference("y")),
            ],
            aggregations: vec![AliasedAggregation {
                alias: "agg".to_string(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: mir_reference("z").into(),
                }),
            }],
            scope: 0,
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("x",0) => mir_int_expr(42),
            Key::named("y",0) => mir_int_expr(55),
            Key::named("z",0) => mir_int_expr(0),
        },
    );

    test_substitute!(
        sort_succeeds_two_levels,
        expected = Some(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("y", vec!["a"])),
                SortSpecification::Desc(mir_field_path("y", vec!["b"])),
            ],
            cache: SchemaCache::new(),
        })),
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("x", vec!["a"])),
                SortSpecification::Desc(mir_field_path("x", vec!["b"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => *mir_field_access("y", "a", true),
                "b".to_string() => *mir_field_access("y", "b", true),
            }.into()),
        },
    );

    test_substitute!(
        match_filter_succeeds_two_levels,
        expected = Some(Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("y", vec!["a"])),
                arg: Integer(42),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        }))),
        stage = Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source: mir_collection("foo", "bar"),
            condition: MatchQuery::Comparison(MatchLanguageComparison {
                function: MatchLanguageComparisonOp::Eq,
                input: Some(mir_field_path("x", vec!["a"])),
                arg: Integer(42),
                cache: SchemaCache::new(),
            }),
            cache: SchemaCache::new(),
        })),
        theta = map! {
            Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => *mir_field_access("y", "a", true),
                "b".to_string() => *mir_field_access("y", "b", true),
            }.into()),
        },
    );

    test_substitute!(
        sort_succeeds_three_levels,
        expected = Some(Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("y", vec!["a"])),
                SortSpecification::Desc(mir_field_path("y", vec!["b"])),
            ],
            cache: SchemaCache::new(),
        })),
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("x", vec!["a", "c"])),
                SortSpecification::Desc(mir_field_path("x", vec!["b", "d"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "c".to_string() => *mir_field_access("y", "a", true),
                }.into()),
                "b".to_string() => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                    "d".to_string() => *mir_field_access("y", "b", true),
                }.into()),
            }.into()),
        },
    );

    test_substitute!(
        sort_fails_missing_key,
        expected = None,
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("x", vec!["a"])),
                SortSpecification::Desc(mir_field_path("x", vec!["b"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("z",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => *mir_field_access("y", "a", true),
                "b".to_string() => *mir_field_access("y", "b", true),
            }.into()),
        },
    );

    test_substitute!(
        sort_fails_not_field_path,
        expected = None,
        stage = Stage::Sort(Sort {
            source: mir_collection("foo", "bar"),
            specs: vec![
                SortSpecification::Asc(mir_field_path("x", vec!["a"])),
                SortSpecification::Desc(mir_field_path("x", vec!["b"])),
            ],
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "a".to_string() => mir_int_expr(42),
                "b".to_string() => mir_int_expr(43),
            }.into()),
        },
    );

    test_substitute!(
        non_sort_trivially_succeeds,
        expected = Some(Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![mir_int_expr(42), mir_int_expr(55),],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        })),
        stage = Stage::Filter(Filter {
            source: mir_collection("foo", "bar"),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![mir_reference("x"), mir_reference("y"),],
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        }),
        theta = map! {
            Key::named("x",0) => mir_int_expr(42),
            Key::named("y",0) => mir_int_expr(55),
        },
    );
}
