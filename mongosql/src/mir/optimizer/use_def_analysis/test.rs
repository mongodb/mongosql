use crate::mir::{self, binding_tuple::Key, optimizer::use_def_analysis::FieldPath};

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
                set,
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

macro_rules! test_attempt_substitute {
    ($func_name:ident, expected = $expected:expr, stage = $input:expr, theta = $theta:expr, condition = $condition:expr,) => {
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
                    Filter, Group, LiteralExpr,
                    LiteralValue::*,
                    ReferenceExpr, ScalarFunction, ScalarFunctionApplication, Sort,
                    SortSpecification, Stage,
                },
                set, unchecked_unique_linked_hash_map,
            };
            #[allow(unused)]
            use std::collections::{HashMap, HashSet};
            let input = $input;
            let expected = $expected;
            let theta = $theta;
            let condition = $condition;
            let actual = input.attempt_substitute(theta, condition);
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

fn field_path(datasource: &str, field_name: &str) -> FieldPath {
    FieldPath::Field {
        parent: FieldPath::Ref(if datasource == "__bot__" {
            Key::bot(0)
        } else {
            Key::named(datasource, 0)
        })
        .into(),
        field: field_name.into(),
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
            field_path("__bot__", "agg1"),
            field_path("__bot__", "agg2"),
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
    expected = Ok({
        let expected: HashSet<FieldPath> = set! {
            FieldPath::Field {
                parent: FieldPath::Field {
                    parent: FieldPath::Ref(Key::named("foo", 0)).into(),
                    field: "bar".to_string(),
                }.into(),
                field: "arr".to_string(),
            },
            field_path("foo", "idx"),
        };
        expected
    }),
    input = Unwind {
        source: mir_collection_stage(),
        path: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
            expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                field: "bar".into(),
                cache: SchemaCache::new(),
            })),
            field: "arr".into(),
            cache: SchemaCache::new(),
        })),
        index: Some("idx".to_string()),
        outer: false,
        cache: SchemaCache::new(),
    },
);
test_field_uses!(
    filter_field_uses,
    expected = Ok({
        let expected: HashSet<FieldPath> = set! {
            field_path("foo", "x"),
            field_path("bar", "x"),
            field_path("bar", "y"),
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
            field_path("foo", "x"),
            field_path("foo", "y"),
        };
        expected
    }),
    input = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_access("foo", "x").into()),
            SortSpecification::Desc(mir_field_access("foo", "y").into()),
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
            SortSpecification::Asc(mir_reference("x").into()),
            SortSpecification::Desc(mir_reference("y").into()),
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
    expected = Stage::Filter(Filter {
        source: mir_collection_stage(),
        condition: Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![mir_int_expr(42), mir_int_expr(55),],
            cache: SchemaCache::new(),
        }),
        cache: SchemaCache::new(),
    }),
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
    sort_substitute,
    expected = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_int_expr(42).into()),
            SortSpecification::Desc(mir_int_expr(55).into()),
        ],
        cache: SchemaCache::new(),
    }),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_reference("x").into()),
            SortSpecification::Desc(mir_reference("y").into()),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir_int_expr(42),
        Key::named("y",0) => mir_int_expr(55),
    },
);
test_substitute!(
    group_substitute,
    expected = Stage::Group(Group {
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
    }),
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

test_attempt_substitute!(
    sort_attempt_substitute_succeeds,
    expected = Some(Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_access("y", "a").into()),
            SortSpecification::Desc(mir_field_access("y", "b").into()),
        ],
        cache: SchemaCache::new(),
    })),
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_access("x", "a").into()),
            SortSpecification::Desc(mir_field_access("x", "b").into()),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => mir_field_access("y", "a"),
            "b".to_string() => mir_field_access("y", "b"),
        }.into()),
    },
    condition = |_| true,
);

test_attempt_substitute!(
    sort_attempt_substitute_fails,
    expected = None,
    stage = Stage::Sort(Sort {
        source: mir_collection_stage(),
        specs: vec![
            SortSpecification::Asc(mir_field_access("x", "a").into()),
            SortSpecification::Desc(mir_field_access("x", "b").into()),
        ],
        cache: SchemaCache::new(),
    }),
    theta = map! {
        Key::named("x",0) => mir::Expression::Document(unchecked_unique_linked_hash_map! {
            "a".to_string() => mir_field_access("y", "a"),
            "b".to_string() => mir_field_access("y", "b"),
        }.into()),
    },
    condition = |_| false,
);

test_attempt_substitute!(
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
    condition = |_| false,
);
