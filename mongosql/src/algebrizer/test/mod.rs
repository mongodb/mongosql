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

#[macro_export]
macro_rules! test_algebrize {
    ($func_name:ident, method = $method:ident, $(in_implicit_type_conversion_context = $in_implicit_type_conversion_context:expr,)? $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? $(expected_error_code = $expected_error_code:literal,)? input = $ast:expr, $(source = $source:expr,)? $(env = $env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)? $(is_add_fields = $is_add_fields:expr, )?) => {
        #[test]
        fn $func_name() {
            #[allow(unused_imports)]
            use $crate::{
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
            use $crate::{
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
            use $crate::{algebrizer::ClauseType, algebrizer::Error, usererror::UserError};

            let user_message = $input.user_message();

            if let Some(message) = user_message {
                assert_eq!($expected, message)
            }
        }
    };
}

#[cfg(test)]
pub mod aggregation;
#[cfg(test)]
pub mod expressions;
#[cfg(test)]
pub mod filter_clause;
#[cfg(test)]
pub mod from_clause;
#[cfg(test)]
pub mod group_by_clause;
#[cfg(test)]
pub mod limit_or_offset_clause;
#[cfg(test)]
pub mod order_by_clause;
#[cfg(test)]
pub mod schema_checking_mode;
#[cfg(test)]
pub mod select_and_order_by;
#[cfg(test)]
pub mod select_clause;
#[cfg(test)]
pub mod set_query;
#[cfg(test)]
pub mod subquery;
#[cfg(test)]
pub mod user_error_messages;

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
