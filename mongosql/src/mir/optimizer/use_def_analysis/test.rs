use crate::{
    mir::{self, binding_tuple::Key, optimizer::use_def_analysis::FieldPath},
    unchecked_unique_linked_hash_map,
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
                    optimizer::use_def_analysis::Error,
                    schema::SchemaCache,
                    Collection,
                    Expression::{self, *},
                    Filter, Group, LiteralExpr,
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
                    optimizer::use_def_analysis::{Error, FieldPath},
                    schema::SchemaCache,
                    AggregationExpr, AggregationFunction, AggregationFunctionApplication,
                    AliasedAggregation, AliasedExpr, Collection,
                    Expression::{self, *},
                    Filter, Group, LiteralExpr,
                    LiteralValue::*,
                    OptionallyAliasedExpr, ReferenceExpr, ScalarFunction,
                    ScalarFunctionApplication, Sort, SortSpecification, Stage,
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
                    Filter, Group, LiteralExpr,
                    LiteralValue::*,
                    OptionallyAliasedExpr, ReferenceExpr, ScalarFunction,
                    ScalarFunctionApplication, Sort, SortSpecification, Stage,
                },
                set,
            };
            #[allow(unused)]
            use std::collections::{HashMap, HashSet};
            let input = $input;
            let expected = $expected;
            let theta = $theta;
            let actual = input.substitute(theta);
            assert_eq!(expected, actual);
        }
    };
}

fn mir_collection_stage() -> Box<mir::Stage> {
    mir::Stage::Collection(mir::Collection {
        db: "foo".into(),
        collection: "bar".into(),
        cache: mir::schema::SchemaCache::new(),
    })
    .into()
}

fn mir_int_key(alias: &str, i: i32) -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
        alias: alias.into(),
        expr: mir::Expression::Literal(mir::LiteralValue::Integer(i).into()),
    })
}

fn mir_unalaised_key() -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Unaliased(mir::Expression::Reference(mir::ReferenceExpr {
        key: mir::binding_tuple::Key::named("foo", 0),
        cache: mir::schema::SchemaCache::new(),
    }))
}

fn mir_int_expr(i: i32) -> mir::Expression {
    mir::Expression::Literal(mir::LiteralExpr {
        value: mir::LiteralValue::Integer(i),
        cache: mir::schema::SchemaCache::new(),
    })
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

fn mir_field_path(datasource_name: &str, field_names: Vec<&str>) -> mir::FieldPath {
    FieldPath {
        key: if datasource_name == "__bot__" {
            Key::bot(0u16)
        } else {
            Key::named(datasource_name, 0u16)
        },
        fields: field_names.into_iter().map(String::from).collect(),
        cache: mir::schema::SchemaCache::new(),
    }
}

test_method!(
    project_defines,
    method = defines,
    expected = {
        let expected: HashMap<Key, Expression> = map! {
            Key::named("x", 0) => Literal(LiteralExpr {
                value: Integer(0),
                cache: SchemaCache::new(),
            }),
            Key::bot(0) => Literal(LiteralExpr {
                value: Integer(0),
                cache: SchemaCache::new(),
            }),
        };
        expected
    },
    input = Project {
        source: mir_collection_stage(),
        expression: map! {
            Key::named("x", 0) => Literal(LiteralExpr {
                value: Integer(0),
                cache: SchemaCache::new(),
            }),
            Key::bot(0) => Literal(LiteralExpr {
                value: Integer(0),
                cache: SchemaCache::new(),
            })
        },
        cache: SchemaCache::new(),
    },
);
test_method!(
    group_defines,
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
        source: mir_collection_stage(),
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
    group_opaque_field_defines,
    method = opaque_field_defines,
    expected = {
        let expected: HashSet<FieldPath> = set! {
            mir_field_path("__bot__", vec!["agg1"]),
            mir_field_path("__bot__", vec!["agg2"]),
        };
        expected
    },
    input = Group {
        source: mir_collection_stage(),
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
    unwind_opaque_field_defines,
    method = opaque_field_defines,
    expected = {
        let expected: HashSet<FieldPath> = set! {
            mir::FieldPath {
                key: Key::named("foo", 0),
                fields: vec!["bar".to_string(), "arr".to_string()],
                cache: SchemaCache::new(),
            },
            mir_field_path("foo", vec!["idx"]),
        };
        expected
    },
    input = Unwind {
        source: mir_collection_stage(),
        path: mir::FieldPath {
            key: Key::named("foo", 0),
            fields: vec!["bar".to_string(), "arr".to_string()],
            cache: SchemaCache::new(),
        },
        index: Some("idx".to_string()),
        outer: false,
        cache: SchemaCache::new(),
    },
);

test_field_uses!(
    filter_field_uses,
    expected = Ok({
        let expected: HashSet<FieldPath> = set! {
            mir_field_path("foo", vec!["x"]),
            mir_field_path("bar", vec!["x"]),
            mir_field_path("bar", vec!["y"]),
        };
        expected
    }),
    input = Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lt,
            args: vec![
                mir_field_access("foo", "x"),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![mir_field_access("bar", "y"), mir_field_access("bar", "x"),],
                    cache: SchemaCache::new(),
                })
            ],

            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
);
test_field_uses!(
    sort_field_uses,
    expected = Ok({
        let expected: HashSet<FieldPath> = set! {
            mir_field_path("foo", vec!["x"]),
            mir_field_path("foo", vec!["y"]),
        };
        expected
    }),
    input = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("foo", vec!["x"])),
            SortSpecification::Desc(mir_field_path("foo", vec!["y"])),
        ],
        cache: SchemaCache::new(),
    }),
);

test_datasource_uses!(
    filter_datasource_uses,
    expected = {
        let expected: HashSet<Key> = set![Key::named("x", 0), Key::named("y", 0),];
        expected
    },
    input = Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Lt,
            args: vec![
                mir_reference("x"),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Add,
                    args: vec![mir_reference("y"), mir_reference("x"),],
                    cache: SchemaCache::new(),
                })
            ],

            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
);
test_datasource_uses!(
    sort_datasource_uses,
    expected = {
        let expected: HashSet<Key> = set![Key::named("x", 0), Key::named("y", 0),];
        expected
    },
    input = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("x", vec!["a"])),
            SortSpecification::Desc(mir_field_path("y", vec!["b"])),
        ],
        cache: SchemaCache::new(),
    }),
);

test_datasource_uses!(
    group_datasource_uses,
    expected = {
        let expected: HashSet<Key> =
            set![Key::named("x", 0), Key::named("y", 0), Key::named("z", 0)];
        expected
    },
    input = Stage::Group(Group {
        source: mir_collection_stage(),
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

test_substitute!(
    filter_substitute,
    expected = Some(Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![mir_int_expr(42), mir_int_expr(55),],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
    stage = Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![mir_reference("x"), mir_reference("y"),],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir_int_expr(42),
        Key::named("y",0) => mir_int_expr(55),
    },
);

test_substitute!(
    sort_substitute_doc,
    expected = Some(Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("bar", vec!["x2", "a", "b"])),
            SortSpecification::Desc(mir_field_path("bar", vec!["y2"])),
        ],
        cache: SchemaCache::new(),
    })),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
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
                    "x".to_string() => mir::Expression::FieldAccess( mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference( mir::ReferenceExpr {
                            key: Key::named("bar", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                        field: "x2".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                    "y".to_string() => mir::Expression::FieldAccess( mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference( mir::ReferenceExpr {
                            key: Key::named("bar", 0u16),
                            cache: mir::schema::SchemaCache::new(),
                        })),
                        field: "y2".to_string(),
                        cache: mir::schema::SchemaCache::new(),
                    }),
                },
                cache: mir::schema::SchemaCache::new(),
          })
    },
);

test_substitute!(
    sort_substitute_access,
    expected = Some(Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("bar", vec!["a", "x", "a"])),
            SortSpecification::Desc(mir_field_path("bar", vec!["a", "y"])),
        ],
        cache: SchemaCache::new(),
    })),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("foo", vec!["x", "a"])),
            SortSpecification::Desc(mir_field_path("foo", vec!["y"])),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("foo",0) => mir_field_access("bar", "a"),
    },
);

test_substitute!(
    sort_substitute_ref,
    expected = Some(Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("bar", vec!["x"])),
            SortSpecification::Desc(mir_field_path("bar", vec!["y"])),
        ],
        cache: SchemaCache::new(),
    })),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
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
    group_substitute,
    expected = Some(Stage::Group(Group {
        source: mir_collection_stage(),
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
        source: mir_collection_stage(),
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
    sort_attempt_substitute_succeeds_two_levels,
    expected = Some(Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("y", vec!["a"])),
            SortSpecification::Desc(mir_field_path("y", vec!["b"])),
        ],
        cache: SchemaCache::new(),
    })),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("x", vec!["a"])),
            SortSpecification::Desc(mir_field_path("x", vec!["b"])),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => mir_field_access("y", "a"),
            "b".to_string() => mir_field_access("y", "b"),
        }.into()),
    },
);

test_substitute!(
    sort_attempt_substitute_succeeds_three_levels,
    expected = Some(Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("y", vec!["a"])),
            SortSpecification::Desc(mir_field_path("y", vec!["b"])),
        ],
        cache: SchemaCache::new(),
    })),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("x", vec!["a", "c"])),
            SortSpecification::Desc(mir_field_path("x", vec!["b", "d"])),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "c".to_string() => mir_field_access("y", "a"),
            }.into()),
            "b".to_string() => mir::Expression::Document(unchecked_unique_linked_hash_map! {
                "d".to_string() => mir_field_access("y", "b"),
            }.into()),
        }.into()),
    },
);

test_substitute!(
    sort_attempt_substitute_fails_missing_key,
    expected = None,
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_path("x", vec!["a"])),
            SortSpecification::Desc(mir_field_path("x", vec!["b"])),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("z",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => mir_field_access("y", "a"),
            "b".to_string() => mir_field_access("y", "b"),
        }.into()),
    },
);

test_substitute!(
    sort_attempt_substitute_fails_not_field_path,
    expected = None,
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
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
    non_sort_attempt_substitute_trivially_succeeds,
    expected = Some(Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![mir_int_expr(42), mir_int_expr(55),],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    })),
    stage = Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![mir_reference("x"), mir_reference("y"),],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir_int_expr(42),
        Key::named("y",0) => mir_int_expr(55),
    },
);
