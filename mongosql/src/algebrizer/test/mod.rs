#![allow(clippy::redundant_pattern_matching)]
use crate::mir::FieldAccess;
use crate::{
    ast::{self, CollectionSource, Datasource},
    catalog::{Catalog, Namespace},
    map,
    mir::{schema::SchemaCache, Collection, Expression, Project, Stage},
    schema::ANY_DOCUMENT,
};
use lazy_static::lazy_static;
use linked_hash_map::LinkedHashMap;
use mongosql_datastructures::binding_tuple::Key;

#[cfg(test)]
pub mod aggregation;
#[cfg(test)]
pub mod expressions;

#[macro_export]
macro_rules! test_algebrize {
    ($func_name:ident, method = $method:ident, $(in_implicit_type_conversion_context = $in_implicit_type_conversion_context:expr,)? $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? $(expected_error_code = $expected_error_code:literal,)? input = $ast:expr, $(source = $source:expr,)? $(env = $env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)? $(is_add_fields = $is_add_fields:expr, )?) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{
                algebrizer::{Algebrizer, Error, ClauseType},
                catalog::Catalog,
                SchemaCheckingMode,
            };

            #[allow(unused_mut, unused_assignments)]
            let mut catalog = Catalog::default();
            $(catalog = $catalog;)?

            #[allow(unused_mut, unused_assignments)]
            let mut schema_checking_mode = SchemaCheckingMode::Strict;
            $(schema_checking_mode = $schema_checking_mode;)?

            #[allow(unused_mut, unused_assignments)]
            let mut algebrizer = Algebrizer::new("test".into(), &catalog, 0u16, schema_checking_mode, false, ClauseType::Unintialized);
            $(algebrizer = Algebrizer::with_schema_env("test".into(), $env, &catalog, 1u16, schema_checking_mode, false, ClauseType::Unintialized);)?

            let res: Result<_, Error> = algebrizer.$method($ast $(, $source)? $(, $in_implicit_type_conversion_context)? $(, $is_add_fields)?);
            $(assert!(matches!(res, $expected_pat));)?
            $(assert_eq!($expected, res);)?

            #[allow(unused_variables)]
            if let Err(e) = res{
                $(assert_eq!($expected_error_code, e.code()))?
            }
        }
    };
}

#[macro_export]
macro_rules! test_algebrize_expr_and_schema_check {
    ($func_name:ident, method = $method:ident, $(in_implicit_type_conversion_context = $in_implicit_type_conversion_context:expr,)? $(expected = $expected:expr,)? $(expected_error_code = $expected_error_code:literal,)? input = $ast:expr, $(source = $source:expr,)? $(env = $env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                algebrizer::{Algebrizer, Error, ClauseType},
                catalog::Catalog,
                SchemaCheckingMode,
                mir::schema::CachedSchema,
            };

            #[allow(unused_mut, unused_assignments)]
            let mut catalog = Catalog::default();
            $(catalog = $catalog;)?

            #[allow(unused_mut, unused_assignments)]
            let mut schema_checking_mode = SchemaCheckingMode::Strict;
            $(schema_checking_mode = $schema_checking_mode;)?

            #[allow(unused_mut, unused_assignments)]
            let mut algebrizer = Algebrizer::new("test".into(), &catalog, 0u16, schema_checking_mode, false, ClauseType::Unintialized);
            $(algebrizer = Algebrizer::with_schema_env("test".into(), $env, &catalog, 1u16, schema_checking_mode, false, ClauseType::Unintialized);)?

            let res: Result<_, Error> = algebrizer.$method($ast $(, $source)? $(, $in_implicit_type_conversion_context)?);
            let res = res.unwrap().schema(&algebrizer.schema_inference_state()).map_err(|e|Error::SchemaChecking(e));
            $(assert_eq!($expected, res);)?

            #[allow(unused_variables)]
            if let Err(e) = res{
                $(assert_eq!($expected_error_code, e.code()))?
            }
        }
    };
}

#[macro_export]
macro_rules! test_user_error_messages {
    ($func_name:ident, input = $input:expr, expected = $expected:expr) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use crate::{algebrizer::ClauseType, algebrizer::Error, usererror::UserError};

            let user_message = $input.user_message();

            if let Some(message) = user_message {
                assert_eq!($expected, message)
            }
        }
    };
}

fn mir_source_collection_with_project(
    collection_name: &str,
    scope: u16,
    field_names: Vec<&str>,
) -> Stage {
    let database = "test";
    if !field_names.is_empty() {
        let document_fields = field_names
            .iter()
            .map(|field_name| {
                (
                    field_name.to_string(),
                    Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference((collection_name, scope).into())),
                        field: field_name.to_string(),
                        is_nullable: true,
                    }),
                )
            })
            .collect::<LinkedHashMap<_, _>>();

        Stage::Project(Project {
            source: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: database.into(),
                    collection: collection_name.into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (collection_name, scope).into() => Expression::Reference((collection_name, scope).into())
                },
                is_add_fields: false,
                cache: SchemaCache::new(),
            })),
            expression: map! {
                Key::bot(scope) => Expression::Document(document_fields.into())
            },
            is_add_fields: false,
            cache: SchemaCache::new(),
        })
    } else {
        Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Collection(Collection {
                db: database.into(),
                collection: collection_name.into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                (collection_name, scope).into() => Expression::Reference((collection_name, scope).into())
            },
            cache: SchemaCache::new(),
        })
    }
}

fn mir_source_foo() -> Stage {
    mir_source_collection_with_project("foo", 0u16, vec![])
}

fn mir_source_bar() -> Stage {
    mir_source_collection_with_project("bar", 0u16, vec![])
}

fn catalog(ns: Vec<(&str, &str)>) -> Catalog {
    ns.into_iter()
        .map(|(db, c)| {
            (
                Namespace {
                    db: db.into(),
                    collection: c.into(),
                },
                ANY_DOCUMENT.clone(),
            )
        })
        .collect::<Catalog>()
}

lazy_static! {
    static ref AST_SOURCE_FOO: Datasource = Datasource::Collection(CollectionSource {
        database: Some("test".into()),
        collection: "foo".into(),
        alias: Some("foo".into()),
    });
    static ref AST_QUERY_FOO: ast::Query = ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
        },
        from_clause: Some(AST_SOURCE_FOO.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    });
    static ref AST_SOURCE_BAR: Datasource = Datasource::Collection(CollectionSource {
        database: Some("test".into()),
        collection: "bar".into(),
        alias: Some("bar".into()),
    });
    static ref AST_QUERY_BAR: ast::Query = ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
        },
        from_clause: Some(AST_SOURCE_BAR.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    });
}

#[cfg(test)]
pub mod select_clause;

#[cfg(test)]
pub mod from_clause;

#[cfg(test)]
pub mod limit_or_offset_clause;

mod set_query {
    use super::{
        catalog, mir_source_bar, mir_source_collection_with_project, mir_source_foo, AST_QUERY_BAR,
        AST_QUERY_FOO, AST_SOURCE_BAR, AST_SOURCE_FOO,
    };
    use crate::schema::{Document, Schema};
    use crate::{ast, map, mir, mir::schema::SchemaCache, multimap, schema, set};
    use mongosql_datastructures::binding_tuple::DatasourceName;

    test_algebrize!(
        union_distinct_star,
        method = algebrize_set_query,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Group(mir::Group {
                source: Box::new(mir::Stage::Set(mir::Set {
                    operation: mir::SetOperation::UnionAll,
                    left: Box::new(mir_source_foo()),
                    right: Box::new(mir_source_bar()),
                    cache: SchemaCache::new()
                })),
                keys: vec![
                    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                        alias: "__groupKey0".to_string(),
                        expr: mir::Expression::Reference(("bar", 0u16).into())
                    }),
                    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                        alias: "__groupKey1".to_string(),
                        expr: mir::Expression::Reference(("foo", 0u16).into())
                    })
                ],
                aggregations: vec![],
                cache: SchemaCache::new(),
                scope: 0
            })),
            expression: map! {
                ("bar", 0u16).into() => *crate::util::mir_field_access("__bot__", "__groupKey0", true),
                ("foo", 0u16).into() => *crate::util::mir_field_access("__bot__", "__groupKey1", true),
            },
            is_add_fields: false,
            cache: SchemaCache::new()
        })),
        input = ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::Union,
            right: Box::new(AST_QUERY_BAR.clone()),
        },
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );

    test_algebrize!(
        union_distinct_values,
        method = algebrize_set_query,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Group(mir::Group {
                source: Box::new(mir::Stage::Set(mir::Set {
                    operation: mir::SetOperation::UnionAll,
                    left: Box::new(mir_source_collection_with_project(
                        "foo",
                        1u16,
                        vec!["a", "b"]
                    )),
                    right: Box::new(mir_source_collection_with_project(
                        "bar",
                        1u16,
                        vec!["a", "b"]
                    )),
                    cache: SchemaCache::new()
                })),
                keys: vec![mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__groupKey0".to_string(),
                    expr: mir::Expression::Reference((DatasourceName::Bottom, 1u16).into())
                })],
                aggregations: vec![],
                cache: SchemaCache::new(),
                scope: 1
            })),
            expression: map! {
                (DatasourceName::Bottom, 1u16).into() => mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "__groupKey0".to_string(),
                    is_nullable: true
                })
            },
            is_add_fields: false,
            cache: SchemaCache::new()
        })),
        input = ast::SetQuery {
            left: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a".into() => ast::Expression::Subpath(ast::SubpathExpr {
                                expr: Box::new(ast::Expression::Identifier("foo".into())),
                                subpath: "a".into(),
                            }),
                            "b".into() => ast::Expression::Subpath(ast::SubpathExpr {
                                expr: Box::new(ast::Expression::Identifier("foo".into())),
                                subpath: "b".into(),
                            })
                        })
                    )])
                },
                from_clause: Some(AST_SOURCE_FOO.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            })),
            op: ast::SetOperator::Union,
            right: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a".into() => ast::Expression::Subpath(ast::SubpathExpr {
                                expr: Box::new(ast::Expression::Identifier("bar".into())),
                                subpath: "a".into(),
                            }),
                            "b".into() => ast::Expression::Subpath(ast::SubpathExpr {
                                expr: Box::new(ast::Expression::Identifier("bar".into())),
                                subpath: "b".into(),
                            })
                        })
                    )])
                },
                from_clause: Some(AST_SOURCE_BAR.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            })),
        },
        env = map! {
            ("foo", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(schema::Atomic::Integer),
                    "b".into() => Schema::Atomic(schema::Atomic::String),
                },
                required: set!{"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
            }),
            ("bar", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(schema::Atomic::Integer),
                    "b".into() => Schema::Atomic(schema::Atomic::String),
                },
                required: set!{"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
            }),
        },
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );

    test_algebrize!(
        basic,
        method = algebrize_set_query,
        expected = Ok(mir::Stage::Set(mir::Set {
            operation: mir::SetOperation::UnionAll,
            left: Box::new(mir_source_foo()),
            right: Box::new(mir_source_bar()),
            cache: SchemaCache::new(),
        })),
        input = ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::UnionAll,
            right: Box::new(AST_QUERY_BAR.clone()),
        },
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
}

mod filter_clause {
    use super::{catalog, mir_source_foo};
    use crate::{ast, mir, mir::schema::SchemaCache};

    fn true_mir() -> mir::Expression {
        mir::Expression::Literal(mir::LiteralValue::Boolean(true))
    }
    const TRUE_AST: ast::Expression = ast::Expression::Literal(ast::Literal::Boolean(true));

    test_algebrize!(
        simple,
        method = algebrize_filter_clause,
        expected = Ok(mir::Stage::Filter(mir::Filter {
            source: Box::new(mir_source_foo()),
            condition: true_mir(),
            cache: SchemaCache::new(),
        })),
        input = Some(TRUE_AST),
        source = mir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        none,
        method = algebrize_filter_clause,
        expected = Ok(mir_source_foo()),
        input = None,
        source = mir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        one_converted_to_bool,
        method = algebrize_filter_clause,
        expected = Ok(mir::Stage::Filter(mir::Filter {
            source: Box::new(mir_source_foo()),
            condition: true_mir(),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Expression::Literal(ast::Literal::Integer(1))),
        source = mir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
}

mod order_by_clause {
    use super::catalog;
    use crate::{
        ast, map, mir,
        mir::schema::SchemaCache,
        schema::{Atomic, Document, Schema},
        set, unchecked_unique_linked_hash_map,
    };

    fn source() -> mir::Stage {
        mir::Stage::Collection(mir::Collection {
            db: "test".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        })
    }

    test_algebrize!(
        asc_and_desc,
        method = algebrize_order_by_clause,
        expected = Ok(mir::Stage::Sort(mir::Sort {
            source: Box::new(source()),
            specs: vec![
                mir::SortSpecification::Asc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["a".to_string()],
                    is_nullable: false,
                }),
                mir::SortSpecification::Desc(mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["b".to_string()],
                    is_nullable: false,
                })
            ],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::OrderByClause {
            sort_specs: vec![
                ast::SortSpec {
                    key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("foo".to_string())),
                        subpath: "a".to_string()
                    })),
                    direction: ast::SortDirection::Asc
                },
                ast::SortSpec {
                    key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("foo".to_string())),
                        subpath: "b".to_string()
                    })),
                    direction: ast::SortDirection::Desc
                }
            ],
        }),
        source = source(),
        env = map! {
            ("foo", 0u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
                }),
        },
        catalog = catalog(vec![("test", "baz")]),
    );

    test_algebrize!(
        sort_key_from_source,
        method = algebrize_order_by_clause,
        expected = Ok(mir::Stage::Sort(mir::Sort {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                    }
                    .into(),
                )],
                alias: "arr".into(),
                cache: SchemaCache::new(),
            })),
            specs: vec![mir::SortSpecification::Asc(mir::FieldPath {
                key: ("arr", 0u16).into(),
                fields: vec!["a".to_string()],
                is_nullable: false,
            }),],
            cache: SchemaCache::new(),
        })),
        input = Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })),
                direction: ast::SortDirection::Asc
            },],
        }),
        source = mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                }
                .into(),
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod group_by_clause {
    use crate::{
        ast, mir, mir::schema::SchemaCache, schema::Satisfaction, unchecked_unique_linked_hash_map,
        usererror::UserError,
    };
    use lazy_static::lazy_static;

    // ARRAY DATASOURCE
    // [{"a" : 1}] AS arr
    fn mir_array_source() -> mir::Stage {
        mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                }
                .into(),
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })
    }
    // GROUP BY KEYS
    // arr.a AS key
    fn mir_field_access() -> mir::OptionallyAliasedExpr {
        mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
            alias: "key".to_string(),
            expr: mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("arr", 0u16).into())),
                field: "a".to_string(),
                is_nullable: false,
            }),
        })
    }
    // 1 AS literal
    fn mir_literal_key() -> mir::OptionallyAliasedExpr {
        mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
            alias: "literal".into(),
            expr: mir::Expression::Literal(mir::LiteralValue::Integer(1)),
        })
    }

    // a + 1 as complex_expr
    fn mir_field_access_complex_expr() -> mir::OptionallyAliasedExpr {
        mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
            alias: "complex_expr".into(),
            expr: mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Add,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("arr", 0u16).into())),
                        field: "a".to_string(),
                        is_nullable: false,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: false,
            }),
        })
    }
    // AVG(DISTINCT arr.a) AS agg1
    fn mir_agg_1_array() -> mir::AliasedAggregation {
        mir::AliasedAggregation {
            alias: "agg1".to_string(),
            agg_expr: mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Avg,
                arg: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("arr", 0u16).into())),
                    field: "a".to_string(),
                    is_nullable: false,
                })),
                distinct: true,
                arg_is_possibly_doc: Satisfaction::Not,
            }),
        }
    }
    // COUNT(*) AS agg2
    fn mir_agg_2() -> mir::AliasedAggregation {
        mir::AliasedAggregation {
            alias: "agg2".to_string(),
            agg_expr: mir::AggregationExpr::CountStar(false),
        }
    }

    lazy_static! {
        // GROUP BY KEYS
        static ref AST_SUBPATH: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
            expr: ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                subpath: "a".to_string()
            }),
            alias: "key".to_string(),
        });

        // 1 AS literal
        static ref AST_LITERAL_KEY: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
            expr: ast::Expression::Literal(ast::Literal::Integer(1)),
            alias: "literal".into(),
        });

        // a + 1 AS complex_expr
        static ref AST_SUBPATH_COMPLEX_EXPR: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
            expr: ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))
            }),
            alias: "complex_expr".into(),
        });

        // AGGREGATION FUNCTIONS

        // AVG(DISTINCT arr.a) AS agg1
        static ref AST_AGG_1_ARRAY: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Avg,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                        subpath: "a".to_string()
                    })
                ]),
                set_quantifier: Some(ast::SetQuantifier::Distinct),
            }),
            alias: "agg1".to_string(),
        };

        // COUNT(*) AS agg2
        static ref AST_AGG_2: ast::AliasedExpr = ast::AliasedExpr {
            expr: ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Count,
                args: ast::FunctionArguments::Star,
                set_quantifier: None
            }),
            alias: "agg2".to_string(),
        };
    }

    // Successful tests.

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE AVG(DISTINCT arr.a) AS agg1, COUNT(*) AS agg2
    test_algebrize!(
        group_by_key_with_aggregation_array_source,
        method = algebrize_group_by_clause,
        expected = Ok(mir::Stage::Group(mir::Group {
            source: Box::new(mir_array_source()),
            keys: vec![mir_field_access()],
            aggregations: vec![mir_agg_1_array(), mir_agg_2()],
            cache: SchemaCache::new(),
            scope: 0,
        })),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![AST_AGG_1_ARRAY.clone(), AST_AGG_2.clone()],
        }),
        source = mir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY 1
    test_algebrize!(
        group_by_key_is_literal,
        method = algebrize_group_by_clause,
        expected = Ok(mir::Stage::Group(mir::Group {
            source: Box::new(mir_array_source()),
            keys: vec![mir_literal_key()],
            aggregations: vec![],
            cache: SchemaCache::new(),
            scope: 0,
        })),
        input = Some(ast::GroupByClause {
            keys: vec![AST_LITERAL_KEY.clone()],
            aggregations: vec![],
        }),
        source = mir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY a + 1
    test_algebrize!(
        group_by_key_is_complex_expression,
        method = algebrize_group_by_clause,
        expected = Ok(mir::Stage::Group(mir::Group {
            source: Box::new(mir_array_source()),
            keys: vec![mir_field_access_complex_expr()],
            aggregations: vec![],
            cache: SchemaCache::new(),
            scope: 0,
        })),
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH_COMPLEX_EXPR.clone()],
            aggregations: vec![],
        }),
        source = mir_array_source(),
    );

    // Error tests.

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE 42 AS agg
    test_algebrize!(
        group_by_key_with_non_function_aggregation_expression,
        method = algebrize_group_by_clause,
        expected = Err(Error::NonAggregationInPlaceOfAggregation(0)),
        expected_error_code = 3013,
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![ast::AliasedExpr {
                expr: ast::Expression::Literal(ast::Literal::Integer(42)),
                alias: "agg".to_string(),
            },],
        }),
        source = mir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key, arr.a AS key
    test_algebrize!(
        group_by_keys_must_have_unique_aliases,
        method = algebrize_group_by_clause,
        expected = Err(Error::DuplicateDocumentKey("key".into())),
        expected_error_code = 3023,
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone(), AST_SUBPATH.clone()],
            aggregations: vec![],
        }),
        source = mir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*) AS a, COUNT(*) AS a
    test_algebrize!(
        group_by_aggregations_must_have_unique_aliases,
        method = algebrize_group_by_clause,
        expected = Err(Error::DuplicateDocumentKey("a".into())),
        expected_error_code = 3023,
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![
                ast::AliasedExpr {
                    expr: ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::Count,
                        args: ast::FunctionArguments::Star,
                        set_quantifier: None
                    }),
                    alias: "a".into(),
                },
                ast::AliasedExpr {
                    expr: ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::Count,
                        args: ast::FunctionArguments::Star,
                        set_quantifier: None
                    }),
                    alias: "a".into(),
                },
            ],
        }),
        source = mir_array_source(),
    );

    // FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*) AS key
    test_algebrize!(
        group_by_aliases_must_be_unique_across_keys_and_aggregates,
        method = algebrize_group_by_clause,
        expected = Err(Error::DuplicateDocumentKey("key".into())),
        expected_error_code = 3023,
        input = Some(ast::GroupByClause {
            keys: vec![AST_SUBPATH.clone()],
            aggregations: vec![ast::AliasedExpr {
                expr: ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::Count,
                    args: ast::FunctionArguments::Star,
                    set_quantifier: None
                }),
                alias: "key".into(),
            },],
        }),
        source = mir_array_source(),
    );
}

mod subquery {
    use super::catalog;
    use crate::{
        ast, map,
        mir::{binding_tuple::DatasourceName, schema::SchemaCache, *},
        multimap,
        schema::{Atomic, Document, Schema},
        set, unchecked_unique_linked_hash_map,
        usererror::UserError,
    };
    use lazy_static::lazy_static;

    fn mir_array(scope: u16) -> Stage {
        Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(1))
                    }
                    .into(),
                )],
                alias: "arr".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("arr", scope).into() => Expression::Reference(("arr", scope).into()),
            },
            cache: SchemaCache::new(),
        })
    }
    lazy_static! {
        static ref AST_ARRAY: ast::Datasource = ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
            },)],
            alias: "arr".into(),
        });
    }
    test_algebrize!(
        uncorrelated_exists,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
                            is_add_fields: false,
            source: Box::new(mir_array(1u16)),
            expression: map! {
                (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                    "a".into() => Expression::Literal(LiteralValue::Integer(1))
                }.into())
            },
            cache: SchemaCache::new(),
        })).into())),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        correlated_exists,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
                            is_add_fields: false,
            source: Box::new(mir_array(2u16)),
            expression: map! {
                (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                    "b_0".into() => Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: false,
                    })
                }.into())
            },
            cache: SchemaCache::new(),
        })).into())),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"b".to_string()},
                additional_properties: false,
                ..Default::default()
                }),
        },
    );
    test_algebrize!(
        exists_cardinality_gt_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
                            is_add_fields: false,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1))}
                    .into()),
                    Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2))}
                    .into())
                ],
                alias: "arr".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
            },
            cache: SchemaCache::new(),
        })).into())),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    },),
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                    },)
                ],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        exists_degree_gt_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Exists(
            Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".to_string() => Expression::Literal(LiteralValue::Integer(1)),
                            "b".to_string() => Expression::Literal(LiteralValue::Integer(2))
                        }
                        .into(),
                    )],
                    alias: "arr".to_string(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
                },
                cache: SchemaCache::new(),
            }))
            .into(),
        )),
        input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                },),],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        uncorrelated_subquery_expr,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(1u16)),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        correlated_subquery_expr,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
                field: "b_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(2u16)),
                expression: map! {
                    (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "b_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                            field: "b".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"b".to_string()},
                additional_properties: false,
                ..Default::default()
                })
        },
    );
    test_algebrize!(
        degree_zero_unsat_output,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::InvalidSubqueryDegree),
        expected_error_code = 3022,
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        substar_degree_eq_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                field: "a".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(1u16)),
                expression: map! {
                    ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                    ast::SubstarExpr {
                        datasource: "arr".into(),
                    }
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        select_values_degree_gt_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::InvalidSubqueryDegree),
        expected_error_code = 3022,
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into()),
                        "b_0".into() => ast::Expression::Identifier("b".into())
                    })
                ),])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                    },),
                    ast::Expression::Document(multimap! {
                        "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                    },)
                ],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        star_degree_eq_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                field: "a".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(mir_array(1u16)),
            is_nullable: false,
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        select_star_degree_gt_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::InvalidSubqueryDegree),
        expected_error_code = 3022,
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                })],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        substar_degree_gt_1,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::InvalidSubqueryDegree),
        expected_error_code = 3022,
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                    ast::SubstarExpr {
                        datasource: "arr".into(),
                    }
                )])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                })],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))),
    );
    test_algebrize!(
        uncorrelated_subquery_comparison_all,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "a_0".to_string(),
                    is_nullable: false,
                })),
                subquery: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: Box::new(mir_array(1u16)),
                    expression: map! {
                    (DatasourceName::Bottom,1u16).into() =>
                        Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() =>
                                Expression::FieldAccess(FieldAccess{
                                    expr:Box::new(Expression::Reference(("arr",1u16).into())),
                                    field:"a".into(),
                                    is_nullable:false,
                                })
                        }.into(),
                    )},
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: false,
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(5))),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::All,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        uncorrelated_subquery_comparison_any,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::Any,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "a_0".to_string(),
                    is_nullable: false,
                })),
                subquery: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: Box::new(mir_array(1u16)),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                                field: "a".into(),
                                is_nullable: false,
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: false,
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(5))),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::Any,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        subquery_comparison_ext_json_arg_converted_if_subquery_is_not_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::Any,
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "a_0".to_string(),
                    is_nullable: false,
                })),
                subquery: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: Box::new(mir_array(1u16)),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                                field: "a".into(),
                                is_nullable: false,
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: false,
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"5\"}".to_string()
            )),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::Any,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        subquery_comparison_ext_json_arg_not_converted_if_subquery_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::Any,
            argument: Box::new(Expression::Literal(LiteralValue::String("{\"$numberInt\": \"5\"}".to_string()))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                    field: "a_0".to_string(),
                    is_nullable: false,
                })),
                subquery: Box::new(Stage::Project(Project {
                            is_add_fields: false,
                    source: Box::new(Stage::Project(Project {
                            is_add_fields: false,
                        source: Box::new(Stage::Array(ArraySource {
                            array: vec![Expression::Document(
                                unchecked_unique_linked_hash_map! {
                                    "a".into() => Expression::Literal(LiteralValue::String("abc".to_string()))
                                }
                                .into(),
                            )],
                            alias: "arr".into(),
                            cache: SchemaCache::new(),
                        })),
                        expression: map! {
                            ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
                        },
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                                field: "a".into(),
                                is_nullable: false,
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: false,
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::StringConstructor("{\"$numberInt\": \"5\"}".to_string())),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::Any,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::StringConstructor("abc".to_string()),
                    })],
                    alias: "arr".into(),
                })),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        argument_from_super_scope,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::All,
            argument: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "b".to_string(),
                is_nullable: false,
            })),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
                    field: "a_0".to_string(),
                    is_nullable: false,
                })),
                subquery: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: Box::new(mir_array(2u16)),
                    expression: map! {
                        (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "a_0".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("arr", 2u16).into())),
                                field: "a".into(),
                                is_nullable: false,
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                is_nullable: false,
            },
            is_nullable: false,
        })),
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Identifier("b".into())),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::All,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{"b".to_string()},
                additional_properties: false,
                ..Default::default()
                })
        },
    );
    test_algebrize!(
        argument_only_evaluated_in_super_scope,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::FieldNotFound(
            "a".into(),
            None,
            ClauseType::Unintialized,
            0u16
        )),
        expected_error_code = 3008,
        input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
            expr: Box::new(ast::Expression::Identifier("a".into())),
            op: ast::ComparisonOp::Eq,
            quantifier: ast::SubqueryQuantifier::All,
            subquery: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                        ast::Expression::Document(multimap! {
                            "a_0".into() => ast::Expression::Identifier("a".into())
                        })
                    )])
                },
                from_clause: Some(AST_ARRAY.clone()),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            },))
        }),
    );
    test_algebrize!(
        potentially_missing_column,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "x".to_string(),
                is_nullable: true,
            })),
            subquery: Box::new(Stage::Limit(Limit {
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: Box::new(Stage::Project(Project {
                        is_add_fields: false,
                        source: Box::new(Stage::Collection(Collection {
                            db: "test".to_string(),
                            collection: "bar".to_string(),
                            cache: SchemaCache::new(),
                        })),
                        expression: map! {
                            (DatasourceName::Named("bar".to_string()), 1u16).into() => Expression::Reference(("bar".to_string(), 1u16).into())
                        },
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                            "x".into() => Expression::FieldAccess(FieldAccess {
                                expr: Box::new(Expression::Reference(("bar", 1u16).into())),
                                field: "x".into(),
                                is_nullable: true,
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                limit: 1,
                cache: SchemaCache::new(),
            })),
            is_nullable: true,
        })),
        input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "x".into() => ast::Expression::Subpath(ast::SubpathExpr {
                            expr: Box::new(ast::Expression::Identifier("bar".into())),
                            subpath: "x".to_string()
                        })
                    })
                )])
            },
            from_clause: Some(ast::Datasource::Collection(ast::CollectionSource {
                database: None,
                collection: "bar".to_string(),
                alias: Some("bar".to_string()),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(1),
            offset: None,
        }))),
        catalog = catalog(vec![("test", "bar")]),
    );
}

mod schema_checking_mode {
    use super::catalog;
    use crate::{
        ast,
        mir::{schema::SchemaCache, *},
        usererror::UserError,
        Schema,
    };

    test_algebrize!(
        comparison_fails_in_strict_mode,
        method = algebrize_order_by_clause,
        expected = Err(Error::SchemaChecking(
            schema::Error::SortKeyNotSelfComparable(0, Schema::Any.into())
        )),
        expected_error_code = 1010,
        input = Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Identifier("a".into())),
                direction: ast::SortDirection::Asc
            }]
        }),
        source = Stage::Collection(Collection {
            db: "".into(),
            collection: "test".into(),
            cache: SchemaCache::new(),
        }),
        catalog = catalog(vec![("", "test")]),
    );

    test_algebrize!(
        comparison_passes_in_relaxed_mode,
        method = algebrize_order_by_clause,
        expected_pat = Ok(_),
        input = Some(ast::OrderByClause {
            sort_specs: vec![ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Identifier("a".into())),
                direction: ast::SortDirection::Asc
            }]
        }),
        source = Stage::Collection(Collection {
            db: "".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        }),
        catalog = catalog(vec![("", "foo")]),
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );
}

mod user_error_messages {

    mod field_not_found {
        test_user_error_messages! {
            no_found_fields,
            input = Error::FieldNotFound("x".into(), None, ClauseType::Select, 1u16),
            expected = "Field `x` of the `SELECT` clause at the 1 scope level not found.".to_string()
        }

        test_user_error_messages! {
            suggestions,
            input = Error::FieldNotFound("foo".into(), Some(vec!["feo".to_string(), "fooo".to_string(), "aaa".to_string(), "bbb".to_string()]), ClauseType::Where, 1u16),
            expected =  "Field `foo` not found in the `WHERE` clause at the 1 scope level. Did you mean: feo, fooo".to_string()
        }

        test_user_error_messages! {
            no_suggestions,
            input = Error::FieldNotFound("foo".into(), Some(vec!["aaa".to_string(), "bbb".to_string(), "ccc".to_string()]), ClauseType::Having, 16u16),
            expected = "Field `foo` in the `HAVING` clause at the 16 scope level not found.".to_string()
        }

        test_user_error_messages! {
            exact_match_found,
            input = Error::FieldNotFound("foo".into(), Some(vec!["foo".to_string()]), ClauseType::GroupBy, 0u16),
            expected = "Unexpected edit distance of 0 found with input: foo and expected: [\"foo\"]"
        }
    }

    mod derived_datasource_overlapping_keys {
        use crate::{
            map,
            schema::{Atomic, Document, Schema},
            set,
        };

        test_user_error_messages! {
        derived_datasource_overlapping_keys,
        input = Error::DerivedDatasourceOverlappingKeys(
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "baz".into() => Schema::Atomic(Atomic::Integer),
                    "foo1".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "baz".into(),
                    "foo1".into(),
                },
                additional_properties: false,
                ..Default::default()
                }).into(),
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "baz".into() => Schema::Atomic(Atomic::Integer),
                    "foo2".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "baz".into(),
                    "foo2".into(),
                },
            additional_properties: false,
            ..Default::default()
            }).into(),
            "foo".into(),
            crate::schema::Satisfaction::Must,
        ),
        expected = "Derived datasource `foo` has the following overlapping keys: bar, baz"
        }
    }

    mod ambiguous_field {
        test_user_error_messages! {
            ambiguous_field,
            input = Error::AmbiguousField("foo".into(), ClauseType::Select, 0u16),
            expected = "Field `foo` in the `SELECT` clause at the 0 scope level exists in multiple datasources and is ambiguous. Please qualify."
        }
    }

    mod cannot_enumerate_all_field_paths {
        test_user_error_messages! {
            cannot_enumerate_all_field_paths,
            input = Error::CannotEnumerateAllFieldPaths(crate::schema::Schema::Any.into()),
            expected = "Insufficient schema information."
        }
    }
}

mod select_and_order_by {
    use crate::mir;

    #[test]
    fn select_and_order_by_column_not_in_select() {
        use crate::{
            algebrizer::{Algebrizer, ClauseType},
            ast,
            catalog::{Catalog, Namespace},
            map,
            mir::{
                binding_tuple::Key, schema::SchemaCache, Collection, Expression, FieldAccess,
                Project, Sort, Stage,
            },
            schema, set, unchecked_unique_linked_hash_map, SchemaCheckingMode,
        };
        let select = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(vec![
                    ast::DocumentPair {
                        key: "_id".into(),
                        value: ast::Expression::Identifier("_id".into()),
                    },
                    ast::DocumentPair {
                        key: "a".into(),
                        value: ast::Expression::Identifier("a".into()),
                    },
                    ast::DocumentPair {
                        key: "c".into(),
                        value: ast::Expression::Binary(ast::BinaryExpr {
                            left: ast::Expression::Identifier("b".into()).into(),
                            op: ast::BinaryOp::Add,
                            right: ast::Expression::Literal(ast::Literal::Integer(42)).into(),
                        }),
                    },
                ]),
            )]),
        };

        let order_by = Some(ast::OrderByClause {
            sort_specs: vec![
                ast::SortSpec {
                    key: ast::SortKey::Simple(ast::Expression::Identifier("b".into())),
                    direction: ast::SortDirection::Asc,
                },
                ast::SortSpec {
                    key: ast::SortKey::Simple(ast::Expression::Identifier("c".into())),
                    direction: ast::SortDirection::Asc,
                },
            ],
        });

        let source = Stage::Project(Project {
            source: Stage::Collection(Collection {
                db: "testquerydb".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })
            .into(),
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into()),
            },
            is_add_fields: false,
            cache: SchemaCache::new(),
        });

        let expected =
            Ok(Stage::Project(Project {
                source: Stage::Sort(Sort {
                    source: Stage::Project(Project {
                        source: Stage::Project(Project {
                            source: Stage::Collection(Collection {
                                db: "testquerydb".into(),
                                collection: "foo".into(),
                                cache: SchemaCache::new(),
                            }).into(),
                            expression: map! {
                                ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into()),
                            },
                            is_add_fields: false,
                            cache: SchemaCache::new(),
                        }).into(),
                        expression: map! {
                            Key::bot(0) => Expression::Document(unchecked_unique_linked_hash_map! {
                                "_id".into() => Expression::FieldAccess(FieldAccess {
                                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                                    field: "_id".into(),
                                    is_nullable: false,
                                }),
                                "a".into() => Expression::FieldAccess(FieldAccess {
                                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                                    field: "a".into(),
                                    is_nullable: false,
                                }),
                                "c".into() => Expression::ScalarFunction(mir::ScalarFunctionApplication {
                                    function: mir::ScalarFunction::Add,
                                    args: vec![
                                        Expression::FieldAccess(FieldAccess {
                                            expr: Expression::Reference(("foo", 0u16).into()).into(),
                                            field: "b".into(),
                                            is_nullable: false,
                                        }),
                                        Expression::Literal(mir::LiteralValue::Integer(42)),
                                    ],
                                    is_nullable: false,
                                }),
                            }.into())
                        },
                        is_add_fields: true,
                        cache: SchemaCache::new(),
                    }).into(),
                    specs: vec![
                        mir::SortSpecification::Asc(mir::FieldPath {
                            key: ("foo", 0u16).into(),
                            fields: vec!["b".into()],
                            is_nullable: false,
                        }),
                        mir::SortSpecification::Asc(mir::FieldPath {
                            key: Key::bot(0u16),
                            fields: vec!["c".into()],
                            is_nullable: false,
                        }),
                    ],
                    cache: SchemaCache::new(),
                }).into(),
                expression: map! {
                    Key::bot(0) => Expression::Document(unchecked_unique_linked_hash_map! {
                        "_id".into() => Expression::FieldAccess(FieldAccess {
                            expr: Expression::Reference(("foo", 0u16).into()).into(),
                            field: "_id".into(),
                            is_nullable: false,
                        }),
                        "a".into() => Expression::FieldAccess(FieldAccess {
                            expr: Expression::Reference(("foo", 0u16).into()).into(),
                            field: "a".into(),
                            is_nullable: false,
                        }),
                        "c".into() => Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::Add,
                            args: vec![
                                Expression::FieldAccess(FieldAccess {
                                    expr: Expression::Reference(("foo", 0u16).into()).into(),
                                    field: "b".into(),
                                    is_nullable: false,
                                }),
                                Expression::Literal(mir::LiteralValue::Integer(42)),
                            ],
                            is_nullable: false,
                        }),
                    }.into())
                },
                is_add_fields: false,
                cache: SchemaCache::new(),
            }));

        let catalog = vec![("testquerydb", "foo")]
            .into_iter()
            .map(|(db, c)| {
                (
                    Namespace {
                        db: db.into(),
                        collection: c.into(),
                    },
                    schema::Schema::Document(schema::Document {
                        keys: map! {
                            "_id".into() => schema::Schema::Atomic(schema::Atomic::ObjectId),
                            "a".into() => schema::Schema::Atomic(schema::Atomic::Integer),
                            "b".into() => schema::Schema::Atomic(schema::Atomic::Integer),
                        },
                        required: set! {"_id".into(), "a".into(), "b".into()},
                        additional_properties: false,
                        jaccard_index: None,
                    }),
                )
            })
            .collect::<Catalog>();

        let algebrizer = Algebrizer::new(
            "test",
            &catalog,
            0u16,
            SchemaCheckingMode::Strict,
            false,
            ClauseType::Unintialized,
        );
        assert_eq!(
            expected,
            algebrizer.algebrize_select_and_order_by_clause(select, order_by, source,),
        );
    }
}
