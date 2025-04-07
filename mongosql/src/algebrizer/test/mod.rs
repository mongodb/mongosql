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

#[cfg(test)]
pub mod set_query;

#[cfg(test)]
pub mod filter_clause;

#[cfg(test)]
pub mod order_by_clause;

#[cfg(test)]
pub mod group_by_clause;

#[cfg(test)]
pub mod subquery;

#[cfg(test)]
pub mod schema_checking_mode;

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
