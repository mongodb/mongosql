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
