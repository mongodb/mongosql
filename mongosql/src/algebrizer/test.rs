#![allow(clippy::redundant_pattern_matching)]
use crate::{
    ast::{self, CollectionSource, Datasource},
    catalog::{Catalog, Namespace},
    map,
    mir::{schema::SchemaCache, Collection, Expression, Project, Stage},
    schema::ANY_DOCUMENT,
};
use lazy_static::lazy_static;

macro_rules! test_algebrize {
    ($func_name:ident, method = $method:ident, $(in_implicit_type_conversion_context = $in_implicit_type_conversion_context:expr,)? $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? $(expected_error_code = $expected_error_code:literal,)? input = $ast:expr, $(source = $source:expr,)? $(env = $env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::{
                algebrizer::{Algebrizer, Error},
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
            let mut algebrizer = Algebrizer::new("test".into(), &catalog, 0u16, schema_checking_mode);
            $(algebrizer = Algebrizer::with_schema_env("test".into(), $env, &catalog, 1u16, schema_checking_mode);)?

            let res: Result<_, Error> = algebrizer.$method($ast $(, $source)? $(, $in_implicit_type_conversion_context)?);
            $(assert!(matches!(res, $expected_pat));)?
            $(assert_eq!($expected, res);)?

            #[allow(unused_variables)]
            if let Err(e) = res{
                $(assert_eq!($expected_error_code, e.code()))?
            }
        }
    };
}

macro_rules! test_algebrize_expr_and_schema_check {
    ($func_name:ident, method = $method:ident, $(in_implicit_type_conversion_context = $in_implicit_type_conversion_context:expr,)? $(expected = $expected:expr,)? $(expected_error_code = $expected_error_code:literal,)? input = $ast:expr, $(source = $source:expr,)? $(env = $env:expr,)? $(catalog = $catalog:expr,)? $(schema_checking_mode = $schema_checking_mode:expr,)?) => {
        #[test]
        fn $func_name() {
            #[allow(unused)]
            use crate::{
                algebrizer::{Algebrizer, Error},
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
            let mut algebrizer = Algebrizer::new("test".into(), &catalog, 0u16, schema_checking_mode);
            $(algebrizer = Algebrizer::with_schema_env("test".into(), $env, &catalog, 1u16, schema_checking_mode);)?

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

macro_rules! test_user_error_messages {
    ($func_name:ident, input = $input:expr, expected = $expected:expr) => {
        #[test]
        fn $func_name() {
            use crate::{algebrizer::Error, usererror::UserError};

            let user_message = $input.user_message();

            if let Some(message) = user_message {
                assert_eq!($expected, message)
            }
        }
    };
}

fn mir_source_foo() -> Stage {
    Stage::Project(Project {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("foo", 0u16).into() => Expression::Reference(("foo", 0u16).into())
        },
        cache: SchemaCache::new(),
    })
}

fn mir_source_bar() -> Stage {
    Stage::Project(Project {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "bar".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("bar", 0u16).into() => Expression::Reference(("bar", 0u16).into())
        },
        cache: SchemaCache::new(),
    })
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

mod expression {
    use crate::{
        ast, map,
        mir::{self, binding_tuple::Key},
        multimap,
        schema::{
            Atomic, Document, Schema, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, NUMERIC_OR_NULLISH,
            STRING_OR_NULLISH,
        },
        set, unchecked_unique_linked_hash_map,
        usererror::UserError,
    };

    mod literal {
        use super::*;

        test_algebrize!(
            null,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::Null)),
            input = ast::Expression::Literal(ast::Literal::Null),
        );

        test_algebrize!(
            expr_true,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            input = ast::Expression::Literal(ast::Literal::Boolean(true)),
        );

        test_algebrize!(
            expr_false,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::Boolean(false))),
            input = ast::Expression::Literal(ast::Literal::Boolean(false)),
        );

        test_algebrize!(
            string,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
                "hello!".into()
            ))),
            input = ast::Expression::StringConstructor("hello!".into()),
        );

        test_algebrize!(
            int,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            input = ast::Expression::Literal(ast::Literal::Integer(42)),
        );

        test_algebrize!(
            long,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::Long(42))),
            input = ast::Expression::Literal(ast::Literal::Long(42)),
        );

        test_algebrize!(
            double,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Literal(mir::LiteralValue::Double(42f64))),
            input = ast::Expression::Literal(ast::Literal::Double(42f64)),
        );

        test_algebrize!(
            empty_array,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Array(vec![].into())),
            input = ast::Expression::Array(vec![]),
        );

        test_algebrize!(
            nested_array,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Array(
                vec![mir::Expression::Array(
                    vec![
                        mir::Expression::Literal(mir::LiteralValue::Long(42)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ]
                    .into()
                )]
                .into()
            )),
            input = ast::Expression::Array(vec![ast::Expression::Array(vec![
                ast::Expression::Literal(ast::Literal::Long(42)),
                ast::Expression::Literal(ast::Literal::Integer(42)),
            ])]),
        );

        test_algebrize!(
            empty_document,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Document(
                unchecked_unique_linked_hash_map! {}.into()
            )),
            input = ast::Expression::Document(multimap! {}),
        );

        test_algebrize!(
            nested_document,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "foo2".into() => mir::Expression::Document(
                        unchecked_unique_linked_hash_map!{"nested".into() => mir::Expression::Literal(mir::LiteralValue::Integer(52))} .into()
                    ),
                    "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42))
                }.into()
            )),
            input = ast::Expression::Document(multimap! {
                "foo2".into() => ast::Expression::Document(
                    multimap!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                ),
                "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            }),
        );

        test_algebrize!(
            document_with_keys_containing_dots_and_dollars,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a.b".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    "$c".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                }
                .into()
            )),
            input = ast::Expression::Document(multimap! {
                "a.b".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "$c".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            }),
        );

        mod ext_json {
            use super::*;

            test_algebrize!(
                array_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Array(
                    vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        mir::Expression::Literal(mir::LiteralValue::String("hello".to_string()))
                    ]
                    .into()
                )),
                input = ast::Expression::StringConstructor("[1, \"hello\"]".to_string()),
            );

            test_algebrize!(
                bindata_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Binary(
                    bson::Binary {
                        subtype: bson::spec::BinarySubtype::Uuid,
                        bytes: vec![]
                    }
                ))),
                input = ast::Expression::StringConstructor(
                    "{ \"$binary\" : {\"base64\" : \"\", \"subType\" : \"04\"}}".to_string()
                ),
            );

            test_algebrize!(
                boolean_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                input = ast::Expression::StringConstructor("true".to_string()),
            );

            test_algebrize!(
                datetime_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::DateTime(
                    "2019-08-11T17:54:14.692Z"
                        .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                        .unwrap()
                        .into()
                ))),
                input = ast::Expression::StringConstructor(
                    "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
                ),
            );

            test_algebrize!(
                decimal128_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Decimal128(
                    "10.99".parse().unwrap()
                ))),
                input = ast::Expression::StringConstructor(
                    "{\"$numberDecimal\": \"10.99\"}".to_string()
                ),
            );

            test_algebrize!(
                document_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "x".into() => mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                        "y".into() => mir::Expression::Literal(mir::LiteralValue::String("hello".to_string())),
                    }
                .into())),
                input = ast::Expression::StringConstructor("{\"x\": 3, \"y\": \"hello\"}".to_string()),
            );

            test_algebrize!(
                double_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Double(10.5))),
                input = ast::Expression::StringConstructor("10.5".to_string()),
            );

            test_algebrize!(
                int_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Long(3))),
                input = ast::Expression::StringConstructor("{\"$numberLong\": \"3\"}".to_string()),
            );

            test_algebrize!(
                javascript_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::JavaScriptCode(
                    "code here".to_string()
                ))),
                input =
                    ast::Expression::StringConstructor("{\"$code\": \"code here\"}".to_string()),
            );

            test_algebrize!(
                javascript_with_scope_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(
                    mir::LiteralValue::JavaScriptCodeWithScope(bson::JavaScriptCodeWithScope {
                        code: "code here".to_string(),
                        scope: bson::doc! {}
                    })
                )),
                input = ast::Expression::StringConstructor(
                    "{\"$code\": \"code here\", \"$scope\": {}}".to_string()
                ),
            );

            test_algebrize!(
                maxkey_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::MaxKey)),
                input = ast::Expression::StringConstructor("{\"$maxKey\": 1}".to_string()),
            );

            test_algebrize!(
                minkey_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::MinKey)),
                input = ast::Expression::StringConstructor("{\"$minKey\": 1}".to_string()),
            );

            test_algebrize!(
                null_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Null)),
                input = ast::Expression::StringConstructor("null".to_string()),
            );

            test_algebrize!(
                objectid_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::ObjectId(
                    bson::oid::ObjectId::parse_str("5ab9c3da31c2ab715d421285").unwrap()
                ))),
                input = ast::Expression::StringConstructor(
                    "{\"$oid\": \"5ab9c3da31c2ab715d421285\"}".to_string()
                ),
            );

            test_algebrize!(
                regular_expression_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(
                    mir::LiteralValue::RegularExpression(bson::Regex {
                        pattern: "pattern".to_string(),
                        options: "".to_string()
                    })
                )),
                input = ast::Expression::StringConstructor(
                    "{\"$regularExpression\":{\"pattern\": \"pattern\",\"options\": \"\"}}"
                        .to_string()
                ),
            );

            test_algebrize!(
                regular_string_stays_string_with_implicit_casting_true,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
                    "abc".to_string()
                ))),
                input = ast::Expression::StringConstructor("abc".to_string()),
            );

            test_algebrize!(
                json_string_stays_string_with_implicit_casting_true,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
                    "'{this_doc_is_actually_a_string: 1}'".to_string()
                ))),
                input = ast::Expression::StringConstructor(
                    "'{this_doc_is_actually_a_string: 1}'".to_string()
                ),
            );

            test_algebrize!(
                symbol_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Symbol(
                    "symbol".to_string()
                ))),
                input = ast::Expression::StringConstructor("{\"$symbol\": \"symbol\"}".to_string()),
            );

            test_algebrize!(
                timestamp_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Timestamp(
                    bson::Timestamp {
                        time: 1,
                        increment: 2
                    }
                ))),
                input = ast::Expression::StringConstructor(
                    "{\"$timestamp\": {\"t\": 1, \"i\": 2}}".to_string()
                ),
            );

            test_algebrize!(
                undefined_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::Undefined)),
                input = ast::Expression::StringConstructor("{\"$undefined\": true}".to_string()),
            );

            test_algebrize!(
                ext_json_not_converted_when_conversion_false,
                method = algebrize_expression,
                in_implicit_type_conversion_context = false,
                expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberLong\": \"3\"}".to_string()
                ))),
                input = ast::Expression::StringConstructor("{\"$numberLong\": \"3\"}".to_string()),
            );
        }
    }

    mod identifier_and_subpath {
        use super::*;

        test_algebrize!(
            qualified_ref_in_current_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                field: "a".into(),
                is_nullable: true
            })),
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("foo".into())),
                subpath: "a".into(),
            }),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            qualified_ref_in_super_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                is_nullable: true
            })),
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("foo".into())),
                subpath: "a".into(),
            }),
            env = map! {
                ("bar", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "b".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_ref_may_exist_in_current_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                field: "a".into(),
                is_nullable: true
            })),
            input = ast::Expression::Identifier("a".into()),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_ref_must_exist_in_current_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                field: "a".into(),
                is_nullable: false
            })),
            input = ast::Expression::Identifier("a".into()),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_ref_may_exist_only_in_super_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                is_nullable: true
            })),
            input = ast::Expression::Identifier("a".into()),
            env = map! {
                ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_ref_must_exist_in_super_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                is_nullable: false
            })),
            input = ast::Expression::Identifier("a".into()),
            env = map! {
                ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_ref_must_exist_in_super_scope_bot_source,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(Key::bot(0u16).into())),
                field: "a".into(),
                is_nullable: false
            })),
            input = ast::Expression::Identifier("a".into()),
            env = map! {
                ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
                Key::bot(0u16) => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_ref_may_and_must_exist_in_two_sources,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::AmbiguousField("a".into())),
            expected_error_code = 3009,
            input = ast::Expression::Identifier("a".into()),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
                Key::bot(1u16) => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_subpath_in_current_and_super_must_exist_in_current,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("test", 1u16).into())),
                    field: "a".into(),
                    is_nullable: false
                })),
                field: "c".into(),
                is_nullable: true
            })),
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("a".into())),
                subpath: "c".into(),
            }),
            env = map! {
                ("test", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("super_test", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_subpath_in_current_and_super_may_exist_is_ambiguous,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::AmbiguousField("a".into())),
            expected_error_code = 3009,
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("a".into())),
                subpath: "c".into(),
            }),
            env = map! {
                ("test", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("super_test", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );
        test_algebrize!(
            subpath_implicit_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    }
                    .into()
                )),
                field: "a".into(),
                is_nullable: false
            })),
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::StringConstructor("{\"a\": 1}".into())),
                subpath: "a".into(),
            }),
            env = map! {
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                    }),
            },
        );

        test_algebrize!(
            unqualified_subpath_in_super_scope,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("super_test", 0u16).into())),
                    field: "a".into(),
                    is_nullable: false
                })),
                field: "c".into(),
                is_nullable: false
            })),
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("a".into())),
                subpath: "c".into(),
            }),
            env = map! {
                ("test", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "b".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("super_test", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{"c".into()},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            qualified_ref_prefers_super_datasource_to_local_field,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                is_nullable: true
            })),
            // test MongoSQL: SELECT (SELECT foo.a FROM bar) FROM foo => (foo.a)
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("foo".into())),
                subpath: "a".into(),
            }),
            env = map! {
                ("bar", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "foo".into() => Schema::Document( Document {
                            keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic( Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            qualified_ref_to_local_field,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("bar", 1u16).into())),
                    field: "foo".into(),
                    is_nullable: true
                })),
                field: "a".into(),
                is_nullable: true
            })),
            //test MongoSQL: SELECT (SELECT bar.foo.a FROM bar) FROM foo => (bar.foo.a)
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("bar".into())),
                    subpath: "foo".into(),
                })),
                subpath: "a".into(),
            }),
            env = map! {
                ("bar", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "foo".into() => Schema::Document( Document {
                            keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("foo", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic( Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            unqualified_reference_and_may_contain_sub_and_must_contain_outer_is_ambiguous,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::AmbiguousField("a".into())),
            expected_error_code = 3009,
            input = ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("a".into())),
                subpath: "c".into(),
            }),
            env = map! {
                ("test", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
                ("super_test", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Document( Document {
                            keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                    },
                    required: set!{"a".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            ref_does_not_exist,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::FieldNotFound("bar".into(), None)),
            expected_error_code = 3008,
            input = ast::Expression::Identifier("bar".into()),
        );
    }

    mod binary {
        use super::*;

        test_algebrize!(
            add_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Add,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            add_wrong_types,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor("hello".into())),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            sub_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Sub,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                op: ast::BinaryOp::Sub,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            sub_wrong_types,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Sub",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor("hello".into())),
                op: ast::BinaryOp::Sub,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            div_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Div,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Double(42.5)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Double(42.5))),
                op: ast::BinaryOp::Div,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            cast_div_result_of_two_integers_to_integer,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Cast(mir::CastExpr {
                expr: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Div,
                        args: vec![
                            mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(42))
                        ],
                        is_nullable: true
                    }
                )),
                to: mir::Type::Int32,
                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true
            })),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                op: ast::BinaryOp::Div,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            cast_div_result_of_long_and_integer_to_long,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Cast(mir::CastExpr {
                expr: Box::new(mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::Div,
                        args: vec![
                            mir::Expression::Literal(mir::LiteralValue::Long(42)),
                            mir::Expression::Literal(mir::LiteralValue::Integer(42))
                        ],
                        is_nullable: true
                    }
                )),
                to: mir::Type::Int64,
                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true
            })),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Long(42))),
                op: ast::BinaryOp::Div,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );
        test_algebrize!(
            cast_implicit_converts_expr_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Cast(mir::CastExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                to: mir::Type::String,
                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".to_string()
                ))),
                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"2\"}".to_string()
                ))),
                is_nullable: false,
            })),
            input = ast::Expression::Cast(ast::CastExpr {
                expr: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"42\"}".to_string()
                )),
                to: ast::Type::String,
                on_null: Some(Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                ))),
                on_error: Some(Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                ))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            div_wrong_types,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Div",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor("hello".into())),
                op: ast::BinaryOp::Div,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            mul_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Mul,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                op: ast::BinaryOp::Mul,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            mul_wrong_types,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Mul",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor("hello".into())),
                op: ast::BinaryOp::Mul,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            concat_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Concat,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String("42".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("42".into())),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor("42".into())),
                op: ast::BinaryOp::Concat,
                right: Box::new(ast::Expression::StringConstructor("42".into())),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            concat_wrong_types,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Concat",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor("hello".into())),
                op: ast::BinaryOp::Concat,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            eq_bool_and_int,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Eq,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
            }),
        );

        test_algebrize!(
            gt_bool_and_int,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Gt,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(0))),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Gt),
                right: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            }),
        );

        test_algebrize!(
            and_bool_and_int,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::And,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(0))),
                op: ast::BinaryOp::And,
                right: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            }),
        );

        test_algebrize!(
            or_int_and_int,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Or,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(0))),
                op: ast::BinaryOp::Or,
                right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
            }),
        );

        test_algebrize!(
            add_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Add,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            and_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::And,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                op: ast::BinaryOp::And,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            div_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Div,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Double(4.0)),
                        mir::Expression::Literal(mir::LiteralValue::Double(2.0)),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberDouble\": \"4.0\"}".to_string()
                )),
                op: ast::BinaryOp::Div,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberDouble\": \"2.0\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            mul_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Mul,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                )),
                op: ast::BinaryOp::Mul,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"3\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            or_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Or,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                        mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"0\"}".to_string()
                )),
                op: ast::BinaryOp::Or,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            sub_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Sub,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(7)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(6)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"7\"}".to_string()
                )),
                op: ast::BinaryOp::Sub,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"6\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            concat_does_not_convert_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Concat,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(
                            "{\"$numberInt\": \"1\"}".to_string()
                        )),
                        mir::Expression::Literal(mir::LiteralValue::String(
                            "{\"$numberInt\": \"1\"}".to_string()
                        )),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                op: ast::BinaryOp::Concat,
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            comp_with_two_strings_does_not_convert_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Eq,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(
                            "{\"$numberInt\": \"1\"}".to_string()
                        )),
                        mir::Expression::Literal(mir::LiteralValue::String(
                            "{\"$numberInt\": \"1\"}".to_string()
                        )),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
        );

        test_algebrize!(
            comp_with_left_that_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Gt,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                            field: "a".into(),
                            is_nullable: true,
                        }),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Gt),
                right: Box::new(ast::Expression::Identifier("a".to_string())),
            }),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            comp_with_left_that_does_not_convert_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lt,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(
                            "{\"$numberInt\": \"1\"}".to_string()
                        )),
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                            field: "a".into(),
                            is_nullable: true,
                        }),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Lt),
                right: Box::new(ast::Expression::Identifier("a".to_string())),
            }),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            comp_with_right_that_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Gte,
                    args: vec![
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                            field: "a".into(),
                            is_nullable: true,
                        }),
                        mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Identifier("a".to_string())),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Gte),
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );

        test_algebrize!(
            comp_with_right_that_does_not_convert_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lte,
                    args: vec![
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                            field: "a".into(),
                            is_nullable: true,
                        }),
                        mir::Expression::Literal(mir::LiteralValue::String(
                            "{\"$numberInt\": \"1\"}".to_string()
                        )),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Identifier("a".to_string())),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Lte),
                right: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
            }),
            env = map! {
                ("foo", 1u16).into() => Schema::Document( Document {
                    keys: map! {
                        "a".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
        );
    }

    mod unary {
        use super::*;

        test_algebrize!(
            neg_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Neg,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Unary(ast::UnaryExpr {
                op: ast::UnaryOp::Neg,
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            neg_wrong_type,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Neg",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Boolean),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Unary(ast::UnaryExpr {
                op: ast::UnaryOp::Neg,
                expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            }),
        );

        test_algebrize!(
            pos_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Pos,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Unary(ast::UnaryExpr {
                op: ast::UnaryOp::Pos,
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            pos_wrong_type,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Pos",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Boolean),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Unary(ast::UnaryExpr {
                op: ast::UnaryOp::Pos,
                expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
            }),
        );

        test_algebrize!(
            unary_op_implicit_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Neg,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Unary(ast::UnaryExpr {
                op: ast::UnaryOp::Neg,
                expr: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"42\"}".to_string()
                )),
            }),
        );
    }

    mod scalar_function {
        use super::*;

        test_algebrize!(
            standard_scalar_function,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Lower,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                        "hello".into()
                    )),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Lower,
                args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                    "hello".into()
                )]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            replace,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Replace,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Replace,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::StringConstructor(" hello world ".to_string()),
                    ast::Expression::StringConstructor("wo".to_string()),
                    ast::Expression::StringConstructor("wowow".to_string()),
                ]),
                set_quantifier: None,
            }),
        );

        test_algebrize!(
            replace_null_one,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Replace,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Null),
                        mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Replace,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Literal(ast::Literal::Null),
                    ast::Expression::StringConstructor("wo".to_string()),
                    ast::Expression::StringConstructor("wowow".to_string()),
                ]),
                set_quantifier: None,
            }),
        );

        test_algebrize!(
            replace_null_two,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Replace,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                        mir::Expression::Literal(mir::LiteralValue::Null),
                        mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Replace,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::StringConstructor(" hello world ".to_string()),
                    ast::Expression::Literal(ast::Literal::Null),
                    ast::Expression::StringConstructor("wowow".to_string()),
                ]),
                set_quantifier: None,
            }),
        );

        test_algebrize!(
            replace_null_three,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Replace,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                        mir::Expression::Literal(mir::LiteralValue::Null),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Replace,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::StringConstructor(" hello world ".to_string()),
                    ast::Expression::StringConstructor("wo".to_string()),
                    ast::Expression::Literal(ast::Literal::Null),
                ]),
                set_quantifier: None,
            }),
        );

        test_algebrize_expr_and_schema_check!(
            replace_args_must_be_string_or_null,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Replace",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Replace,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Literal(ast::Literal::Integer(42)),
                    ast::Expression::Literal(ast::Literal::Integer(42)),
                    ast::Expression::Literal(ast::Literal::Integer(42)),
                ]),
                set_quantifier: None,
            }),
        );

        test_algebrize!(
            log_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Log,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(100)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Log,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Literal(ast::Literal::Integer(100)),
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                ]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            round_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Round,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Round,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                ]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            cos_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Cos,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Cos,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(10)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            sin_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Sin,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Sin,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(10)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            tan_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Tan,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Tan,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(10)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            radians_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Radians,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Radians,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            sqrt_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Sqrt,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(4)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Sqrt,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(4)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            abs_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Abs,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Abs,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(10)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            ceil_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Ceil,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Ceil,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Double(1.5)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            degrees_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Degrees,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Degrees,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Integer(1)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            floor_unary_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Floor,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Floor,
                args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                    ast::Literal::Double(1.5)
                ),]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            mod_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Mod,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    ],
                    is_nullable: true,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Mod,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                ]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );

        test_algebrize!(
            pow_bin_op,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Pow,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                        mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Pow,
                args: ast::FunctionArguments::Args(vec![
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                    ast::Expression::Literal(ast::Literal::Integer(10)),
                ]),
                set_quantifier: Some(ast::SetQuantifier::All),
            }),
        );
    }

    mod trim {
        use super::*;

        test_algebrize!(
            ltrim,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::LTrim,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String("hello".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("hello world".into()))
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Trim(ast::TrimExpr {
                trim_spec: ast::TrimSpec::Leading,
                trim_chars: Box::new(ast::Expression::StringConstructor("hello".into())),
                arg: Box::new(ast::Expression::StringConstructor("hello world".into())),
            }),
        );

        test_algebrize!(
            rtrim,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::RTrim,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String("world".into())),
                        mir::Expression::Literal(mir::LiteralValue::String("hello world".into()))
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Trim(ast::TrimExpr {
                trim_spec: ast::TrimSpec::Trailing,
                trim_chars: Box::new(ast::Expression::StringConstructor("world".into())),
                arg: Box::new(ast::Expression::StringConstructor("hello world".into())),
            }),
        );

        test_algebrize!(
            btrim,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::BTrim,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::String(" ".into())),
                        mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into()))
                    ],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Trim(ast::TrimExpr {
                trim_spec: ast::TrimSpec::Both,
                trim_chars: Box::new(ast::Expression::StringConstructor(" ".into())),
                arg: Box::new(ast::Expression::StringConstructor(" hello world ".into())),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            trim_arg_must_be_string_or_null,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "BTrim",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Trim(ast::TrimExpr {
                trim_spec: ast::TrimSpec::Both,
                trim_chars: Box::new(ast::Expression::StringConstructor(" ".into())),
                arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            trim_escape_must_be_string,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "BTrim",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Trim(ast::TrimExpr {
                trim_spec: ast::TrimSpec::Both,
                trim_chars: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                arg: Box::new(ast::Expression::StringConstructor(" ".into())),
            }),
        );
    }

    mod extract {
        use super::*;

        test_algebrize!(
            extract_year,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Year,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Year,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_month,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Month,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Month,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_day,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Day,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Day,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_hour,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Hour,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Hour,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_minute,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Minute,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Minute,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_second,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Second,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Second,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_millsecond,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Millisecond,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Millisecond,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_day_of_year,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::DayOfYear,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::DayOfYear,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_iso_week,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::IsoWeek,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::IsoWeek,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_day_of_week,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::DayOfWeek,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::DayOfWeek,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize!(
            extract_iso_weekday,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::IsoWeekday,
                    args: vec![mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }
                    ),],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::IsoWeekday,
                arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            extract_must_be_date,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Second",
                required: DATE_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Second,
                arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            }),
        );

        test_algebrize!(
            extract_implicit_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Year,
                    args: vec![mir::Expression::Literal(mir::LiteralValue::DateTime(
                        "2019-08-11T17:54:14.692Z"
                            .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                            .unwrap()
                            .into()
                    ))],
                    is_nullable: false,
                }
            )),
            input = ast::Expression::Extract(ast::ExtractExpr {
                extract_spec: ast::DatePart::Year,
                arg: Box::new(ast::Expression::StringConstructor(
                    "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
                )),
            }),
        );
    }

    mod date_function {
        use super::*;

        test_algebrize!(
            dateadd,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::DateFunction(
                mir::DateFunctionApplication {
                    function: mir::DateFunction::Add,
                    is_nullable: false,
                    date_part: mir::DatePart::Quarter,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(5),),
                        mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }),
                    ],
                }
            )),
            input = ast::Expression::DateFunction(ast::DateFunctionExpr {
                function: ast::DateFunctionName::Add,
                date_part: ast::DatePart::Quarter,
                args: vec![
                    ast::Expression::Literal(ast::Literal::Integer(5)),
                    ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::CurrentTimestamp,
                        args: ast::FunctionArguments::Args(vec![]),
                        set_quantifier: Some(ast::SetQuantifier::All)
                    })
                ],
            }),
        );

        test_algebrize!(
            datediff,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::DateFunction(
                mir::DateFunctionApplication {
                    function: mir::DateFunction::Diff,
                    is_nullable: false,
                    date_part: mir::DatePart::Week,
                    args: vec![
                        mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }),
                        mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }),
                        mir::Expression::Literal(mir::LiteralValue::String("sunday".to_string()),)
                    ],
                }
            )),
            input = ast::Expression::DateFunction(ast::DateFunctionExpr {
                function: ast::DateFunctionName::Diff,
                date_part: ast::DatePart::Week,
                args: vec![
                    ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::CurrentTimestamp,
                        args: ast::FunctionArguments::Args(vec![]),
                        set_quantifier: Some(ast::SetQuantifier::All)
                    }),
                    ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::CurrentTimestamp,
                        args: ast::FunctionArguments::Args(vec![]),
                        set_quantifier: Some(ast::SetQuantifier::All)
                    }),
                    ast::Expression::StringConstructor("sunday".to_string()),
                ],
            }),
        );

        test_algebrize!(
            datetrunc,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::DateFunction(
                mir::DateFunctionApplication {
                    function: mir::DateFunction::Trunc,
                    is_nullable: false,
                    date_part: mir::DatePart::Year,
                    args: vec![
                        mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::CurrentTimestamp,
                            args: vec![],
                            is_nullable: false,
                        }),
                        mir::Expression::Literal(mir::LiteralValue::String("sunday".to_string()),)
                    ],
                }
            )),
            input = ast::Expression::DateFunction(ast::DateFunctionExpr {
                function: ast::DateFunctionName::Trunc,
                date_part: ast::DatePart::Year,
                args: vec![
                    ast::Expression::Function(ast::FunctionExpr {
                        function: ast::FunctionName::CurrentTimestamp,
                        args: ast::FunctionArguments::Args(vec![]),
                        set_quantifier: Some(ast::SetQuantifier::All)
                    }),
                    ast::Expression::StringConstructor("sunday".to_string()),
                ],
            }),
        );

        test_algebrize!(
            date_function_implicit_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::DateFunction(
                mir::DateFunctionApplication {
                    function: mir::DateFunction::Add,
                    is_nullable: false,
                    date_part: mir::DatePart::Quarter,
                    args: vec![
                        mir::Expression::Literal(mir::LiteralValue::Integer(5),),
                        mir::Expression::Literal(mir::LiteralValue::DateTime(
                            "2019-08-11T17:54:14.692Z"
                                .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                                .unwrap()
                                .into()
                        ))
                    ],
                }
            )),
            input = ast::Expression::DateFunction(ast::DateFunctionExpr {
                function: ast::DateFunctionName::Add,
                date_part: ast::DatePart::Quarter,
                args: vec![
                    ast::Expression::StringConstructor("{\"$numberInt\": \"5\"}".to_string()),
                    ast::Expression::StringConstructor(
                        "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
                    )
                ],
            }),
        );
    }

    mod case {
        use super::*;

        test_algebrize!(
            searched_case,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
                when_branch: vec![mir::WhenBranch {
                    when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                    then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                        "bar".into()
                    ))),
                    is_nullable: false,
                }],
                else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "foo".into()
                ))),
                is_nullable: true,
            })),
            input = ast::Expression::Case(ast::CaseExpr {
                expr: None,
                when_branch: vec![ast::WhenBranch {
                    when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                    then: Box::new(ast::Expression::StringConstructor("bar".into())),
                }],
                else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
            }),
        );

        test_algebrize!(
            searched_case_no_else,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
                when_branch: vec![mir::WhenBranch {
                    when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                    then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                        "bar".into()
                    ))),
                    is_nullable: false,
                }],
                else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true,
            })),
            input = ast::Expression::Case(ast::CaseExpr {
                expr: None,
                when_branch: vec![ast::WhenBranch {
                    when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                    then: Box::new(ast::Expression::StringConstructor("bar".into())),
                }],
                else_branch: None,
            }),
        );

        test_algebrize_expr_and_schema_check!(
            searched_case_when_condition_is_not_bool,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "SearchedCase",
                required: BOOLEAN_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String),
            })),
            expected_error_code = 1002,
            input = ast::Expression::Case(ast::CaseExpr {
                expr: None,
                when_branch: vec![ast::WhenBranch {
                    when: Box::new(ast::Expression::StringConstructor("foo".into())),
                    then: Box::new(ast::Expression::StringConstructor("bar".into())),
                }],
                else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
            }),
        );

        test_algebrize!(
            simple_case,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                when_branch: vec![mir::WhenBranch {
                    when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                    then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                        "bar".into()
                    ))),
                    is_nullable: false,
                }],
                else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "foo".into()
                ))),
                is_nullable: false,
            })),
            input = ast::Expression::Case(ast::CaseExpr {
                expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
                when_branch: vec![ast::WhenBranch {
                    when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                    then: Box::new(ast::Expression::StringConstructor("bar".into())),
                }],
                else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
            }),
        );

        test_algebrize!(
            simple_case_no_else,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                when_branch: vec![mir::WhenBranch {
                    when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                    then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                        "bar".into()
                    ))),
                    is_nullable: false,
                }],
                else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: false,
            })),
            input = ast::Expression::Case(ast::CaseExpr {
                expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
                when_branch: vec![ast::WhenBranch {
                    when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                    then: Box::new(ast::Expression::StringConstructor("bar".into())),
                }],
                else_branch: None,
            }),
        );

        test_algebrize_expr_and_schema_check!(
            simple_case_operand_and_when_operand_not_comparable,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(
                mir::schema::Error::InvalidComparison(
                    "SimpleCase",
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                )
            )),
            expected_error_code = 1005,
            input = ast::Expression::Case(ast::CaseExpr {
                expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
                when_branch: vec![ast::WhenBranch {
                    when: Box::new(ast::Expression::StringConstructor("foo".into())),
                    then: Box::new(ast::Expression::StringConstructor("bar".into())),
                }],
                else_branch: Some(Box::new(ast::Expression::StringConstructor("baz".into()))),
            }),
        );
    }

    mod type_operators {
        use super::*;

        test_algebrize!(
            cast_full,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Cast(mir::CastExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                to: mir::Type::String,
                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "was_null".into()
                ))),
                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "was_error".into()
                ))),
                is_nullable: false,
            })),
            input = ast::Expression::Cast(ast::CastExpr {
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                to: ast::Type::String,
                on_null: Some(Box::new(ast::Expression::StringConstructor(
                    "was_null".into()
                ))),
                on_error: Some(Box::new(ast::Expression::StringConstructor(
                    "was_error".into()
                ))),
            }),
        );

        test_algebrize!(
            cast_simple,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Cast(mir::CastExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                to: mir::Type::String,
                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true,
            })),
            input = ast::Expression::Cast(ast::CastExpr {
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                to: ast::Type::String,
                on_null: None,
                on_error: None,
            }),
        );

        test_algebrize!(
            type_assert_success,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                target_type: mir::Type::Int32,
            })),
            input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                target_type: ast::Type::Int32,
            }),
        );

        test_algebrize_expr_and_schema_check!(
            type_assert_fail,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "::!",
                required: Schema::Atomic(Atomic::String),
                found: Schema::Atomic(Atomic::Integer)
            })),
            expected_error_code = 1002,
            input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                target_type: ast::Type::String,
            }),
        );

        test_algebrize!(
            type_assert_ext_json_string_does_not_convert_if_target_type_is_string,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"42\"}".to_string()
                ))),
                target_type: mir::Type::String,
            })),
            input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
                expr: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"42\"}".to_string()
                )),
                target_type: ast::Type::String,
            }),
        );

        test_algebrize!(
            type_assert_ext_json_string_converts_if_target_type_is_not_string,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                target_type: mir::Type::Int32,
            })),
            input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
                expr: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"42\"}".to_string()
                )),
                target_type: ast::Type::Int32,
            }),
        );

        test_algebrize!(
            is_success,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Is(mir::IsExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
            })),
            input = ast::Expression::Is(ast::IsExpr {
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
            }),
        );

        test_algebrize_expr_and_schema_check!(
            is_recursive_failure,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Add",
                required: NUMERIC_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::String)
            })),
            expected_error_code = 1002,
            input = ast::Expression::Is(ast::IsExpr {
                expr: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                    left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                    op: ast::BinaryOp::Add,
                    right: Box::new(ast::Expression::StringConstructor("a".into())),
                })),
                target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
            }),
        );

        test_algebrize!(
            is_implicit_converts_ext_json,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Is(mir::IsExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
                target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
            })),
            input = ast::Expression::Is(ast::IsExpr {
                expr: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"42\"}".to_string()
                )),
                target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
            }),
        );
    }

    mod like {
        use super::*;

        test_algebrize!(
            like_success_with_pattern,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Like(mir::LikeExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "42".into()
                ))),
                pattern: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "42".into()
                ))),
                escape: Some('f'),
            })),
            input = ast::Expression::Like(ast::LikeExpr {
                expr: Box::new(ast::Expression::StringConstructor("42".into())),
                pattern: Box::new(ast::Expression::StringConstructor("42".into())),
                escape: Some('f'),
            }),
        );

        test_algebrize!(
            like_success_no_pattern,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Like(mir::LikeExpr {
                expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "42".into()
                ))),
                pattern: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "42".into()
                ))),
                escape: None,
            })),
            input = ast::Expression::Like(ast::LikeExpr {
                expr: Box::new(ast::Expression::StringConstructor("42".into())),
                pattern: Box::new(ast::Expression::StringConstructor("42".into())),
                escape: None,
            }),
        );

        test_algebrize_expr_and_schema_check!(
            like_expr_must_be_string,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Like",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer)
            })),
            expected_error_code = 1002,
            input = ast::Expression::Like(ast::LikeExpr {
                expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                pattern: Box::new(ast::Expression::StringConstructor("42".into())),
                escape: None,
            }),
        );

        test_algebrize_expr_and_schema_check!(
            like_pattern_must_be_string,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
                name: "Like",
                required: STRING_OR_NULLISH.clone(),
                found: Schema::Atomic(Atomic::Integer)
            })),
            expected_error_code = 1002,
            input = ast::Expression::Like(ast::LikeExpr {
                expr: Box::new(ast::Expression::StringConstructor("42".into())),
                pattern: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                escape: Some(' '),
            }),
        );
    }
}

mod aggregation {
    use crate::{
        ast, map, mir, multimap,
        schema::{Atomic, Schema, ANY_DOCUMENT, NUMERIC_OR_NULLISH},
        unchecked_unique_linked_hash_map,
        usererror::UserError,
    };
    test_algebrize!(
        count_star,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::CountStar(false)),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize_expr_and_schema_check!(
        count_distinct_star_is_error,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(
            mir::schema::Error::CountDistinctStarNotSupported
        )),
        expected_error_code = 1004,
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize!(
        count_all_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Count,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        count_distinct_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Count,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize_expr_and_schema_check!(
        count_distinct_expr_argument_not_self_comparable_is_error,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(
            mir::schema::Error::AggregationArgumentMustBeSelfComparable(
                "Count DISTINCT".into(),
                Schema::Any
            )
        )),
        expected_error_code = 1003,
        input = ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Identifier("foo".into())]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
        env = map! {
            ("d", 1u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        sum_star_is_error,
        method = algebrize_aggregation,
        expected = Err(Error::StarInNonCount),
        expected_error_code = 3010,
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Star,
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        sum_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Sum,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        sum_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Sum,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize_expr_and_schema_check!(
        sum_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Sum",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        expected_error_code = 1002,
        input = ast::FunctionExpr {
            function: ast::FunctionName::Sum,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "42".into()
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        avg_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Avg,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        avg_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Avg,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize_expr_and_schema_check!(
        avg_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Avg",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        expected_error_code = 1002,
        input = ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "42".into()
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        stddevpop_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::StddevPop,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        stddevpop_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::StddevPop,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize_expr_and_schema_check!(
        stddevpop_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "StddevPop",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        expected_error_code = 1002,
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevPop,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "42".into()
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        stddevsamp_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::StddevSamp,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        stddevsamp_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::StddevSamp,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );
    test_algebrize_expr_and_schema_check!(
        stddevsamp_argument_must_be_numeric,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "StddevSamp",
            required: NUMERIC_OR_NULLISH.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        expected_error_code = 1002,
        input = ast::FunctionExpr {
            function: ast::FunctionName::StddevSamp,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "42".into()
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        addtoarray_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::AddToArray,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToArray,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        addtoarray_distinct_expr_basic_test,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::AddToArray,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToArray,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        addtoset_expr_is_addtoarray_distinct_in_mir,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::AddToArray,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToSet,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        addtoset_distinct_expr_is_addtoarray_in_mir,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::AddToArray,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::AddToSet,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        first_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::First,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::First,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        first_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::First,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::First,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        last_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Last,
                distinct: false,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Last,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize!(
        last_distinct_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::Last,
                distinct: true,
                arg: mir::Expression::Literal(mir::LiteralValue::Integer(42)).into(),
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::Last,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(42)
            )]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        },
    );

    test_algebrize!(
        mergedocuments_expr,
        method = algebrize_aggregation,
        expected = Ok(mir::AggregationExpr::Function(
            mir::AggregationFunctionApplication {
                function: mir::AggregationFunction::MergeDocuments,
                distinct: false,
                arg: Box::new(mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                        "b".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
                    }
                    .into()
                ))
            }
        )),
        input = ast::FunctionExpr {
            function: ast::FunctionName::MergeDocuments,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            })]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
    test_algebrize_expr_and_schema_check!(
        mergedocuments_argument_must_be_document,
        method = algebrize_aggregation,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "MergeDocuments",
            required: ANY_DOCUMENT.clone(),
            found: Schema::Atomic(Atomic::String),
        })),
        expected_error_code = 1002,
        input = ast::FunctionExpr {
            function: ast::FunctionName::MergeDocuments,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "42".into()
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        },
    );
}

mod select_clause {
    use super::catalog;
    use crate::{
        ast, map,
        mir::{self, binding_tuple::Key, schema::SchemaCache},
        multimap,
        schema::ANY_DOCUMENT,
        unchecked_unique_linked_hash_map,
        usererror::UserError,
    };

    fn source() -> mir::Stage {
        mir::Stage::Collection(mir::Collection {
            db: "test".into(),
            collection: "baz".into(),
            cache: SchemaCache::new(),
        })
    }

    test_algebrize!(
        select_distinct_not_allowed,
        method = algebrize_select_clause,
        expected = Err(Error::DistinctSelect),
        expected_error_code = 3005,
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::Distinct,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Identifier("foo".into())
            ),]),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
    );
    test_algebrize!(
        select_duplicate_bot,
        method = algebrize_select_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(source()),
            expression: map! {
                Key::bot(1u16) => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map!{},
                })
            },
            cache: SchemaCache::new(),
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
                ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
            ]),
        },
        source = source(),
        env = map! {},
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_duplicate_doc_key_a,
        method = algebrize_select_clause,
        expected = Err(Error::DuplicateDocumentKey("a".into())),
        expected_error_code = 3023,
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                },)
            ),]),
        },
        source = source(),
        env = map! {},
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_bot_and_double_substar,
        method = algebrize_select_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(source()),
            expression: map! {
                Key::bot(1u16) => mir::Expression::Document(unchecked_unique_linked_hash_map!{}.into()),
                ("bar", 1u16).into() => mir::Expression::Reference(("bar", 1u16).into()),
                ("foo", 1u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Substar("bar".into()),
                ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
                ast::SelectValuesExpression::Substar("foo".into()),
            ]),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
            ("bar", 1u16).into() => ANY_DOCUMENT.clone(),
        },
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_value_expression_must_be_document,
        method = algebrize_select_clause,
        expected = Err(Error::SchemaChecking(
            crate::mir::schema::Error::SchemaChecking {
                name: "project datasource",
                required: ANY_DOCUMENT.clone(),
                found: crate::schema::Schema::Atomic(crate::schema::Atomic::String),
            }
        )),
        expected_error_code = 1002,
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::StringConstructor("foo".into())
            ),]),
        },
        source = source(),
        env = map! {},
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_duplicate_substar,
        method = algebrize_select_clause,
        expected = Err(Error::DuplicateKey(("foo", 1u16).into())),
        expected_error_code = 3020,
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![
                ast::SelectValuesExpression::Substar("foo".into()),
                ast::SelectValuesExpression::Substar("foo".into()),
            ]),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
        catalog = catalog(vec![("test", "baz")]),
    );
    test_algebrize!(
        select_substar_body,
        method = algebrize_select_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(source()),
            expression: map! {
                ("foo", 1u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(
                vec![ast::SelectValuesExpression::Substar("foo".into()),]
            ),
        },
        source = source(),
        env = map! {
            ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        },
        catalog = catalog(vec![("test", "baz")]),
    );
}

mod from_clause {
    use super::{catalog, mir_source_bar, mir_source_foo, AST_SOURCE_BAR, AST_SOURCE_FOO};
    use crate::{
        ast::{self, JoinSource},
        catalog::{Catalog, Namespace},
        map,
        mir::{self, binding_tuple::Key, schema::SchemaCache, JoinType},
        multimap,
        schema::{Atomic, Document, Schema, ANY_DOCUMENT},
        set, unchecked_unique_linked_hash_map,
        usererror::UserError,
    };

    fn mir_array_source() -> mir::Stage {
        mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {"a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map!{
                            "b".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(5),)
                        },
                    })},
                })],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("arr", 0u16).into() => mir::Expression::Reference(("arr", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })
    }

    fn ast_array_source() -> ast::Datasource {
        ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Document(
                            multimap!{"b".into() => ast::Expression::Literal(ast::Literal::Integer(5))},
                        ),
            })],
            alias: "arr".to_string(),
        })
    }

    test_algebrize!(
        basic_collection,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    mir::Expression::Reference(("foo", 0u16).into())
            },
            cache: SchemaCache::new(),
        },),),
        input = Some(ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: Some("bar".into()),
        })),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        qualified_collection,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test2".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    mir::Expression::Reference(("foo", 0u16).into())
            },
            cache: SchemaCache::new(),
        }),),
        input = Some(ast::Datasource::Collection(ast::CollectionSource {
            database: Some("test2".into()),
            collection: "foo".into(),
            alias: Some("bar".into()),
        })),
        catalog = catalog(vec![("test2", "foo")]),
    );
    test_algebrize!(
        collection_and_alias_contains_dot,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test2".into(),
                collection: "foo.bar".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("foo.bar", 0u16).into() =>
                    mir::Expression::Reference(("foo.bar", 0u16).into())
            },
            cache: SchemaCache::new(),
        }),),
        input = Some(ast::Datasource::Collection(ast::CollectionSource {
            database: Some("test2".into()),
            collection: "foo.bar".into(),
            alias: Some("foo.bar".into()),
        })),
        catalog = catalog(vec![("test2", "foo.bar")]),
    );
    test_algebrize!(
        collection_and_alias_starts_with_dollar,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Collection(mir::Collection {
                db: "test2".into(),
                collection: "$foo".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("$foo", 0u16).into() =>
                    mir::Expression::Reference(("$foo", 0u16).into())
            },
            cache: SchemaCache::new(),
        }),),
        input = Some(ast::Datasource::Collection(ast::CollectionSource {
            database: Some("test2".into()),
            collection: "$foo".into(),
            alias: Some("$foo".into()),
        })),
        catalog = catalog(vec![("test2", "$foo")]),
    );
    test_algebrize!(
        empty_array,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        dual,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {}.into()
                )],
                alias: "_dual".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("_dual", 0u16).into() => mir::Expression::Reference(("_dual", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {},)],
            alias: "_dual".into(),
        })),
    );
    test_algebrize!(
        int_array,
        method = algebrize_from_clause,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Integer)]),
        })),
        expected_error_code = 1002,
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Integer(42))],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        null_array,
        method = algebrize_from_clause,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "array datasource items",
            required: ANY_DOCUMENT.clone(),
            found: Schema::AnyOf(set![Schema::Atomic(Atomic::Null)]),
        })),
        expected_error_code = 1002,
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Literal(ast::Literal::Null)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        array_datasource_must_be_literal,
        method = algebrize_from_clause,
        expected = Err(Error::ArrayDatasourceMustBeLiteral),
        expected_error_code = 3004,
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "foo".into() => ast::Expression::Identifier("foo".into()),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        single_document_array,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                    }
                    .into()
                )],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
            },)],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        two_document_array,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![
                    mir::Expression::Document(unchecked_unique_linked_hash_map! {
                        "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                    }.into()),
                    mir::Expression::Document(unchecked_unique_linked_hash_map! {
                        "foo2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(41)),
                        "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42))
                    }.into())
                ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                }),
                ast::Expression::Document(multimap! {
                    "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(41)),
                    "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                },)
            ],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        two_document_with_nested_document_array,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Array(mir::ArraySource {
                array: vec![
                    mir::Expression::Document(unchecked_unique_linked_hash_map! {
                        "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                        "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                    }.into()),
                    mir::Expression::Document(unchecked_unique_linked_hash_map! {
                        "foo2".into() => mir::Expression::Document(
                            unchecked_unique_linked_hash_map!{"nested".into() => mir::Expression::Literal(mir::LiteralValue::Integer(52))}
                        .into()),
                        "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42))
                    }.into())
                ],
                alias: "bar".into(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
            },
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                }),
                ast::Expression::Document(multimap! {
                    "foo2".into() => ast::Expression::Document(
                        multimap!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                    ),
                    "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                },)
            ],
            alias: "bar".into(),
        })),
    );
    test_algebrize!(
        left_join,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Join(mir::Join {
            join_type: JoinType::Left,
            left: Box::new(mir_source_foo()),
            right: Box::new(mir_source_bar()),
            condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Left,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: Some(ast::Expression::Literal(ast::Literal::Boolean(true)))
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        right_join,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Join(mir::Join {
            join_type: JoinType::Left,
            left: Box::new(mir_source_bar()),
            right: Box::new(mir_source_foo()),
            condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Right,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: Some(ast::Expression::Literal(ast::Literal::Boolean(true)))
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        left_outer_join_without_condition,
        method = algebrize_from_clause,
        expected = Err(Error::NoOuterJoinCondition),
        expected_error_code = 3019,
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Left,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        right_outer_join_without_condition,
        method = algebrize_from_clause,
        expected = Err(Error::NoOuterJoinCondition),
        expected_error_code = 3019,
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Right,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        inner_join,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Join(mir::Join {
            join_type: JoinType::Inner,
            left: Box::new(mir_source_foo()),
            right: Box::new(mir_source_bar()),
            condition: None,
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Inner,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        cross_join,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Join(mir::Join {
            join_type: JoinType::Inner,
            left: Box::new(mir_source_foo()),
            right: Box::new(mir_source_bar()),
            condition: None,
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: None
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        join_on_one_true,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Join(mir::Join {
            join_type: JoinType::Inner,
            left: Box::new(mir_source_foo()),
            right: Box::new(mir_source_bar()),
            condition: Some(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(AST_SOURCE_FOO.clone()),
            right: Box::new(AST_SOURCE_BAR.clone()),
            condition: Some(ast::Expression::Literal(ast::Literal::Integer(1)))
        })),
        catalog = catalog(vec![("test", "foo"), ("test", "bar")]),
    );
    test_algebrize!(
        invalid_join_condition,
        method = algebrize_from_clause,
        expected = Err(Error::FieldNotFound(
            "x".into(),
            Some(vec!["bar".into(), "foo".into()])
        )),
        expected_error_code = 3008,
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                })],
                alias: "foo".into()
            })),
            right: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                })],
                alias: "bar".into()
            })),
            condition: Some(ast::Expression::Identifier("x".into())),
        })),
    );
    test_algebrize!(
        join_condition_must_have_boolean_schema,
        method = algebrize_from_clause,
        expected_pat = Err(Error::SchemaChecking(_)),
        expected_error_code = 1002,
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Cross,
            left: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                })],
                alias: "foo".into()
            })),
            right: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                })],
                alias: "bar".into()
            })),
            condition: Some(ast::Expression::Literal(ast::Literal::Integer(42))),
        })),
    );
    test_algebrize!(
        derived_single_datasource,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Derived(mir::Derived {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Array(mir::ArraySource {
                        array: vec![mir::Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                                "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                            }
                        .into())],
                        alias: "bar".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        ("bar", 1u16).into() => mir::Expression::Reference(("bar", 1u16).into()),
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("d", 0u16).into() => mir::Expression::Reference(("bar", 1u16).into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star]),
                },
                from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    },)],
                    alias: "bar".into(),
                })),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        derived_multiple_datasources,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Derived(mir::Derived {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Project(mir::Project {
                    source: Box::new(mir::Stage::Project(mir::Project {
                        source: Box::new(mir::Stage::Array(mir::ArraySource {
                            array: vec![mir::Expression::Document(
                                unchecked_unique_linked_hash_map! {
                                    "foo".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                                    "bar".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                                }
                            .into())],
                            alias: "bar".into(),
                            cache: SchemaCache::new(),
                        })),
                        expression: map! {
                            ("bar", 1u16).into() => mir::Expression::Reference(("bar", 1u16).into()),
                        },
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        Key::bot(1u16) => mir::Expression::Document(
                            unchecked_unique_linked_hash_map!{"baz".into() => mir::Expression::Literal(mir::LiteralValue::String("hello".into()))}
                        .into()),
                        ("bar", 1u16).into() =>
                            mir::Expression::Reference(("bar", 1u16).into())
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! { ("d", 0u16).into() =>
                    mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::MergeObjects,
                            args:
                                vec![
                                    mir::Expression::Reference(Key::bot(1u16).into()),
                                    mir::Expression::Reference(("bar", 1u16).into()),
                                ],
                            is_nullable: false,
                        }
                    )
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Values(vec![
                        ast::SelectValuesExpression::Substar("bar".into()),
                        ast::SelectValuesExpression::Expression(ast::Expression::Document(
                            multimap! {
                                "baz".into() => ast::Expression::StringConstructor("hello".into())
                            }
                        )),
                    ]),
                },
                from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                    array: vec![ast::Expression::Document(multimap! {
                        "foo".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    },)],
                    alias: "bar".into(),
                })),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        derived_join_datasources_distinct_keys_succeeds,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Derived(mir::Derived {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: mir::Stage::Join(mir::Join {
                    join_type: mir::JoinType::Inner,
                    left: mir::Stage::Project(mir::Project {
                        source: Box::new(mir::Stage::Array(mir::ArraySource {
                            array: vec![mir::Expression::Document(
                                unchecked_unique_linked_hash_map! {
                                "foo1".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                                "bar1".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                                                            }
                            .into())],
                            alias: "bar1".into(),
                            cache: SchemaCache::new(),
                        })),
                        expression: map! {
                            ("bar1", 1u16).into() => mir::Expression::Reference(("bar1", 1u16).into()),
                        },
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    right: mir::Stage::Project(mir::Project {
                        source: Box::new(mir::Stage::Array(mir::ArraySource {
                            array: vec![mir::Expression::Document(
                                unchecked_unique_linked_hash_map! {
                                "foo2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                                "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                                                            }
                            .into())],
                            alias: "bar2".into(),
                            cache: SchemaCache::new(),
                        })),
                        expression: map! {
                            ("bar2", 1u16).into() => mir::Expression::Reference(("bar2", 1u16).into()),
                        },
                        cache: SchemaCache::new(),
                    })
                    .into(),
                    condition: None,
                    cache: SchemaCache::new(),
                })
                .into(),
                expression: map! {("d", 0u16).into() =>
                    mir::Expression::ScalarFunction(
                        mir::ScalarFunctionApplication {
                            function: mir::ScalarFunction::MergeObjects,
                            args: vec![mir::Expression::Reference(("bar1", 1u16).into()),
                                       mir::Expression::Reference(("bar2", 1u16).into())],
                            is_nullable: false,
                        }
                    )
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star,]),
                },
                from_clause: Some(ast::Datasource::Join(JoinSource {
                    join_type: ast::JoinType::Inner,
                    left: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(multimap! {
                            "foo1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                            "bar1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        },)],
                        alias: "bar1".into(),
                    })
                    .into(),
                    right: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(multimap! {
                            "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                            "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        },)],
                        alias: "bar2".into(),
                    })
                    .into(),
                    condition: None,
                })),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        join_condition_referencing_non_correlated_fields,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Join(mir::Join {
            join_type: mir::JoinType::Left,
            left: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Array(mir::ArraySource {
                    array: vec![mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(1))}
                    .into())],
                    alias: "foo".to_string(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 0u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            right: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Array(mir::ArraySource {
                    array: vec![mir::Expression::Document(
                        unchecked_unique_linked_hash_map! {"b".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(4))}
                    .into())],
                    alias: "bar".to_string(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("bar", 0u16).into() => mir::Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            condition: Some(mir::Expression::ScalarFunction(
                mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::Eq,
                    args: vec![
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
                            field: "a".to_string(),
                            is_nullable: false,
                        }),
                        mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(("bar", 0u16).into())),
                            field: "b".to_string(),
                            is_nullable: false,
                        })
                    ],
                    is_nullable: false,
                }
            )),
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Join(JoinSource {
            join_type: ast::JoinType::Left,
            left: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                },)],
                alias: "foo".into(),
            })),
            right: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(4)),
                },)],
                alias: "bar".into(),
            })),
            condition: Some(ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Identifier("a".to_string())),
                op: ast::BinaryOp::Comparison(ast::ComparisonOp::Eq),
                right: Box::new(ast::Expression::Identifier("b".to_string())),
            }))
        })),
    );
    test_algebrize!(
        derived_join_datasources_overlapped_keys_fails,
        method = algebrize_from_clause,
        expected = Err(Error::DerivedDatasourceOverlappingKeys(
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "foo1".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "foo1".into()
                },
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "foo2".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "foo2".into()
                },
                additional_properties: false,
                ..Default::default()
            }),
            "d".into(),
            crate::schema::Satisfaction::Must,
        )),
        expected_error_code = 3016,
        input = Some(ast::Datasource::Derived(ast::DerivedSource {
            query: Box::new(ast::Query::Select(ast::SelectQuery {
                select_clause: ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star,]),
                },
                from_clause: Some(ast::Datasource::Join(JoinSource {
                    join_type: ast::JoinType::Inner,
                    left: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(multimap! {
                            "foo1".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                            "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        },)],
                        alias: "bar1".into(),
                    })
                    .into(),
                    right: ast::Datasource::Array(ast::ArraySource {
                        array: vec![ast::Expression::Document(multimap! {
                            "foo2".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                            "bar".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                        },)],
                        alias: "bar2".into(),
                    })
                    .into(),
                    condition: None,
                })),
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            })),
            alias: "d".into(),
        })),
    );
    test_algebrize!(
        flatten_simple,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir_array_source()),
            expression: map! {
                ("arr", 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!{
                    "a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("arr", 0u16).into(),
                            })),
                            field: "a".to_string(),
                            is_nullable: false,
                        })),
                        field: "b".to_string(),
                        is_nullable: false,
                    })},
            })},
            cache: SchemaCache::new()
        })),
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(ast_array_source()),
            options: vec![]
        })),
    );
    test_algebrize!(
        flatten_array_source_multiple_docs,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Array(mir::ArraySource {
                    array: vec![mir::Expression::Document(mir::DocumentExpr {
                        document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                document: unchecked_unique_linked_hash_map!{
                                    "b".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(5),)
                                },
                            }),
                            "x".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                document: unchecked_unique_linked_hash_map! {
                                    "y".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(8),)
                                },
                            })
                        },
                    })],
                    alias: "arr".to_string(),
                    cache: SchemaCache::new()
                })),
                expression: map! {
                    ("arr", 0u16).into() => mir::Expression::Reference(("arr", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("arr", 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {
                        "a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                    key: ("arr", 0u16).into(),
                                })),
                                field: "a".to_string(),
                                is_nullable: false,
                            })),
                            field: "b".to_string(),
                            is_nullable: false,

                        }),
                        "x_y".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                    key: ("arr", 0u16).into(),
                                })),
                                field: "x".to_string(),
                                is_nullable: false,
                            })),
                            field: "y".to_string(),
                            is_nullable: false,
                        })
                    },
                })
            },
            cache: SchemaCache::new()
        })),
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                            "a".into() => ast::Expression::Document(
                                multimap!{"b".into() => ast::Expression::Literal(ast::Literal::Integer(5))},
                            ),
                    "x".into() => ast::Expression::Document(
                                multimap!{"y".into() => ast::Expression::Literal(ast::Literal::Integer(8))},
                            ),
                })],
                alias: "arr".to_string()
            })),
            options: vec![]
        })),
    );
    test_algebrize!(
        flatten_duplicate_options,
        method = algebrize_from_clause,
        expected = Err(Error::DuplicateFlattenOption(ast::FlattenOption::Depth(2))),
        expected_error_code = 3024,
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(ast_array_source()),
            options: vec![ast::FlattenOption::Depth(1), ast::FlattenOption::Depth(2)]
        })),
    );
    test_algebrize!(
        flatten_polymorphic_non_document_schema_array_source,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir::Stage::Project(mir::Project {
                source: Box::new(mir::Stage::Array(mir::ArraySource {
                    array: vec![
                        mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                document: unchecked_unique_linked_hash_map!{
                                    "b".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                        document: unchecked_unique_linked_hash_map!{
                                            "c".to_string() => mir::Expression::Literal(mir::LiteralValue::Integer(5),)},
                                    })},
                            })},
                        }),
                        mir::Expression::Document(mir::DocumentExpr {
                            document: unchecked_unique_linked_hash_map! {
                            "a".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                document: unchecked_unique_linked_hash_map!{
                                    "b".to_string() => mir::Expression::Document(mir::DocumentExpr {
                                        document: unchecked_unique_linked_hash_map!{
                                            "c".to_string() => mir::Expression::Literal(mir::LiteralValue::String("hello".to_string()),)},
                                    })},
                            })},
                        })
                    ],
                    alias: "arr".to_string(),
                    cache: SchemaCache::new()
                })),
                expression: map! {
                    ("arr", 0u16).into() => mir::Expression::Reference(("arr", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            expression: map! {
            ("arr", 0u16).into() => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!{
                    "a_b_c".to_string() => mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                    key: ("arr", 0u16).into(),
                                })),
                                field: "a".to_string(),
                                is_nullable: false,
                            })),
                            field: "b".to_string(),
                            is_nullable: false,
                        })),
                        field: "c".to_string(),
                        is_nullable: false,
                    })},
            })},
            cache: SchemaCache::new()
        })),
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Document(
                        multimap!{"b".into() => ast::Expression::Document(
                        multimap!{"c".into() => ast::Expression::Literal(ast::Literal::Integer(5))},
                    )},
                    )}),
                    ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Document(
                        multimap!{"b".into() => ast::Expression::Document(
                        multimap!{"c".into() => ast::Expression::StringConstructor("hello".to_string())},
                    )},
                    )}),
                ],
                alias: "arr".to_string()
            })),
            options: vec![]
        })),
    );
    test_algebrize!(
        flatten_polymorphic_object_schema_array_source,
        method = algebrize_from_clause,
        expected = Err(Error::PolymorphicObjectSchema("a".to_string())),
        expected_error_code = 3026,
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(ast::Datasource::Array(ast::ArraySource {
                array: vec![
                    ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(5))}),
                    ast::Expression::Document(multimap! {
                        "a".into() => ast::Expression::Document(
                            multimap!{"b".into() => ast::Expression::Literal(ast::Literal::Integer(6))},
                        )
                    }),
                ],
                alias: "arr".to_string()
            })),
            options: vec![]
        })),
    );
    test_algebrize!(
        flattening_polymorphic_objects_other_than_just_null_or_missing_polymorphism_causes_error,
        method = algebrize_from_clause,
        expected = Err(Error::PolymorphicObjectSchema("a".to_string())),
        expected_error_code = 3026,
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(AST_SOURCE_FOO.clone()),
            options: vec![]
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::AnyOf(set!{
                        Schema::Document(Document {
                            keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Integer)
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
        }),
    );

    test_algebrize!(
        flattening_polymorphic_objects_with_just_null_polymorphism_works,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir_source_foo()),
            expression: map! {("foo", 0u16).into() => mir::Expression::Document(
                mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {"a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess{
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("foo", 0u16).into(),
                            })),
                            field: "a".to_string(),
                            is_nullable: true,
                        })),

                        field: "b".to_string(),
                        is_nullable: true,
                    }
                )}},
            )},
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(AST_SOURCE_FOO.clone()),
            options: vec![]
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::AnyOf(set!{
                        Schema::Document(Document {
                            keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Null),
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
        }),
    );

    test_algebrize!(
        flattening_polymorphic_objects_with_just_missing_polymorphism_works,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir_source_foo()),
            expression: map! {("foo", 0u16).into() => mir::Expression::Document(
                mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {"a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess{
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("foo", 0u16).into(),
                            })),
                            field: "a".to_string(),
                            is_nullable: true,
                        })),

                        field: "b".to_string(),
                        is_nullable: true,
                    }
                )}},
            )},
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(AST_SOURCE_FOO.clone()),
            options: vec![]
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::AnyOf(set!{
                        Schema::Document(Document {
                            keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                        Schema::Missing,
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
        }),
    );

    test_algebrize!(
        flattening_polymorphic_objects_with_just_null_and_missing_polymorphism_works,
        method = algebrize_from_clause,
        expected = Ok(mir::Stage::Project(mir::Project {
            source: Box::new(mir_source_foo()),
            expression: map! {("foo", 0u16).into() => mir::Expression::Document(
                mir::DocumentExpr {
                    document: unchecked_unique_linked_hash_map! {"a_b".to_string() => mir::Expression::FieldAccess(mir::FieldAccess{
                        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                            expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
                                key: ("foo", 0u16).into(),
                            })),
                            field: "a".to_string(),
                            is_nullable: true,
                        })),

                        field: "b".to_string(),
                        is_nullable: true,
                    }
                )}},
            )},
            cache: SchemaCache::new(),
        })),
        input = Some(ast::Datasource::Flatten(ast::FlattenSource {
            datasource: Box::new(AST_SOURCE_FOO.clone()),
            options: vec![]
        })),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::AnyOf(set!{
                        Schema::Document(Document {
                            keys: map! {"b".into() => Schema::Atomic(Atomic::Integer)},
                            required: set!{},
                            additional_properties: false,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    }),
                },
                required: set!{"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
        }),
    );
    mod unwind {
        use super::*;

        /// Most tests use the same collection source and need to specify the
        /// collection schema for the test to work. This helper allows easy
        /// definition of that collection schema.
        fn make_catalog(s: Schema) -> Catalog {
            Catalog::new(map! {
                Namespace {db: "test".into(), collection: "foo".into()} => s,
            })
        }

        test_algebrize!(
            simple,
            method = algebrize_from_clause,
            expected = Ok(mir::Stage::Unwind(mir::Unwind {
                source: Box::new(mir_source_foo()),
                path: mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["arr".to_string()],
                    is_nullable: false,
                },
                index: None,
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            })),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Identifier(
                    "arr".into()
                ))]
            })),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set! {"arr".into()},
                additional_properties: false,
                ..Default::default()
            })),
        );
        test_algebrize!(
            all_opts,
            method = algebrize_from_clause,
            expected = Ok(mir::Stage::Unwind(mir::Unwind {
                source: Box::new(mir_source_foo()),
                path: mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["arr".to_string()],
                    is_nullable: false,
                },
                index: Some("i".into()),
                outer: true,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            })),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![
                    ast::UnwindOption::Path(ast::Expression::Identifier("arr".into())),
                    ast::UnwindOption::Index("i".into()),
                    ast::UnwindOption::Outer(true),
                ]
            })),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set! {"arr".into()},
                additional_properties: false,
                ..Default::default()
            })),
        );
        test_algebrize!(
            compound_path,
            method = algebrize_from_clause,
            expected = Ok(mir::Stage::Unwind(mir::Unwind {
                source: Box::new(mir_source_foo()),
                path: mir::FieldPath {
                    key: ("foo", 0u16).into(),
                    fields: vec!["doc".to_string(), "arr".to_string()],
                    is_nullable: false,
                },
                index: None,
                outer: false,
                cache: SchemaCache::new(),
                is_prefiltered: false,
            })),
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
                    ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("doc".into())),
                        subpath: "arr".into()
                    }
                ))]
            })),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "doc".into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                        },
                        required: set!{"arr".into()},
                        additional_properties: false,
                        ..Default::default()
                        }),
                },
                required: set! {"doc".into()},
                additional_properties: false,
                ..Default::default()
            })),
        );
        test_algebrize!(
            duplicate_opts,
            method = algebrize_from_clause,
            expected = Err(Error::DuplicateUnwindOption(ast::UnwindOption::Path(
                ast::Expression::Identifier("dup".into())
            ))),
            expected_error_code = 3027,
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![
                    ast::UnwindOption::Path(ast::Expression::Identifier("arr".into())),
                    ast::UnwindOption::Path(ast::Expression::Identifier("dup".into())),
                ]
            })),
            catalog = make_catalog(ANY_DOCUMENT.clone()),
        );
        test_algebrize!(
            missing_path,
            method = algebrize_from_clause,
            expected = Err(Error::NoUnwindPath),
            expected_error_code = 3028,
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![]
            })),
            catalog = make_catalog(ANY_DOCUMENT.clone()),
        );
        test_algebrize!(
            invalid_path,
            method = algebrize_from_clause,
            expected = Err(Error::InvalidUnwindPath),
            expected_error_code = 3029,
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
                    ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Document(vec![ast::DocumentPair {
                            key: "arr".into(),
                            value: ast::Expression::Array(vec![
                                ast::Expression::Literal(ast::Literal::Integer(1)),
                                ast::Expression::Literal(ast::Literal::Integer(2)),
                                ast::Expression::Literal(ast::Literal::Integer(3))
                            ])
                        }])),
                        subpath: "arr".into()
                    }
                )),]
            })),
            catalog = make_catalog(ANY_DOCUMENT.clone()),
        );
        test_algebrize!(
            correlated_path_disallowed,
            method = algebrize_from_clause,
            expected = Err(Error::FieldNotFound("bar".into(), Some(vec!["arr".into()]))),
            expected_error_code = 3008,
            input = Some(ast::Datasource::Unwind(ast::UnwindSource {
                datasource: Box::new(AST_SOURCE_FOO.clone()),
                options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
                    ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("bar".into())),
                        subpath: "arr".into()
                    }
                )),]
            })),
            env = map! {
                ("bar", 0u16).into() => Schema::Document( Document {
                    keys: map! {
                        "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                    },
                    required: set!{ "arr".into() },
                    additional_properties: false,
                    ..Default::default()
                    }),
            },
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set! {"arr".into()},
                additional_properties: false,
                ..Default::default()
            })),
        );
    }
}

mod limit_or_offset_clause {
    use super::{catalog, mir_source_foo, AST_SOURCE_FOO};
    use crate::{ast, mir, mir::schema::SchemaCache};

    test_algebrize!(
        limit_set,
        method = algebrize_limit_clause,
        expected = Ok(mir::Stage::Limit(mir::Limit {
            source: Box::new(mir_source_foo()),
            limit: 42_u64,
            cache: SchemaCache::new(),
        })),
        input = Some(42_u32),
        source = mir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        limit_unset,
        method = algebrize_limit_clause,
        expected = Ok(mir_source_foo()),
        input = None,
        source = mir_source_foo(),
    );
    test_algebrize!(
        offset_set,
        method = algebrize_offset_clause,
        expected = Ok(mir::Stage::Offset(mir::Offset {
            source: Box::new(mir_source_foo()),
            offset: 3,
            cache: SchemaCache::new(),
        })),
        input = Some(3_u32),
        source = mir_source_foo(),
        catalog = catalog(vec![("test", "foo")]),
    );
    test_algebrize!(
        offset_unset,
        method = algebrize_offset_clause,
        expected = Ok(mir_source_foo()),
        input = None,
        source = mir_source_foo(),
    );
    test_algebrize!(
        limit_and_offset,
        method = algebrize_select_query,
        expected = Ok(mir::Stage::Limit(mir::Limit {
            source: Box::new(mir::Stage::Offset(mir::Offset {
                source: Box::new(mir_source_foo()),
                offset: 3,
                cache: SchemaCache::new(),
            })),
            limit: 10,
            cache: SchemaCache::new(),
        })),
        input = ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
            },
            from_clause: Some(AST_SOURCE_FOO.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(10_u32),
            offset: Some(3_u32)
        },
        catalog = catalog(vec![("test", "foo")]),
    );
}

mod set_query {
    use super::{catalog, mir_source_bar, mir_source_foo, AST_QUERY_BAR, AST_QUERY_FOO};
    use crate::{ast, mir, mir::schema::SchemaCache, usererror::UserError};

    test_algebrize!(
        union_distinct_not_allowed,
        method = algebrize_set_query,
        expected = Err(Error::DistinctUnion),
        expected_error_code = 3006,
        input = ast::SetQuery {
            left: Box::new(AST_QUERY_FOO.clone()),
            op: ast::SetOperator::Union,
            right: Box::new(AST_QUERY_BAR.clone()),
        },
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
                    .into()
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
                .into()
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod group_by_clause {
    use crate::{
        ast, mir, mir::schema::SchemaCache, unchecked_unique_linked_hash_map, usererror::UserError,
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
            alias: "arr".into()
        });
    }
    test_algebrize!(
        uncorrelated_exists,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
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
                alias: "arr".into()
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
                source: Box::new(Stage::Array(ArraySource {
                    array: vec![Expression::Document(
                        unchecked_unique_linked_hash_map! {
                            "a".to_string() => Expression::Literal(LiteralValue::Integer(1)),
                            "b".to_string() => Expression::Literal(LiteralValue::Integer(2))
                        }
                        .into()
                    )],
                    alias: "arr".to_string(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
                },
                cache: SchemaCache::new(),
            }))
            .into()
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
                alias: "arr".into()
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
                alias: "arr".into()
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
                        datasource: "arr".into()
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
                alias: "arr".into()
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
                alias: "arr".into()
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
                        datasource: "arr".into()
                    }
                )])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                })],
                alias: "arr".into()
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
                        }.into()
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
                    source: Box::new(Stage::Project(Project {
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
                    alias: "arr".into()
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
        expected = Err(Error::FieldNotFound("a".into(), None)),
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
                    source: Box::new(Stage::Project(Project {
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
            schema::Error::SortKeyNotSelfComparable(0, Schema::Any)
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
            input = Error::FieldNotFound("x".into(), None),
            expected = "Field `x` not found.".to_string()
        }

        test_user_error_messages! {
            suggestions,
            input = Error::FieldNotFound("foo".into(), Some(vec!["feo".to_string(), "fooo".to_string(), "aaa".to_string(), "bbb".to_string()])),
            expected =  "Field `foo` not found. Did you mean: feo, fooo".to_string()
        }

        test_user_error_messages! {
            no_suggestions,
            input = Error::FieldNotFound("foo".into(), Some(vec!["aaa".to_string(), "bbb".to_string(), "ccc".to_string()])),
            expected = "Field `foo` not found.".to_string()
        }

        test_user_error_messages! {
            exact_match_found,
            input = Error::FieldNotFound("foo".into(), Some(vec!["foo".to_string()])),
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
                    "foo1".into()
                },
                additional_properties: false,
                ..Default::default()
                }),
            Schema::Document(Document {
                keys: map! {
                    "bar".into() => Schema::Atomic(Atomic::Integer),
                    "baz".into() => Schema::Atomic(Atomic::Integer),
                    "foo2".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {
                    "bar".into(),
                    "baz".into(),
                    "foo2".into()
                },
            additional_properties: false,
            ..Default::default()
            }),
            "foo".into(),
            crate::schema::Satisfaction::Must,
        ),
        expected = "Derived datasource `foo` has the following overlapping keys: bar, baz"
        }
    }

    mod ambiguous_field {
        test_user_error_messages! {
            ambiguous_field,
            input = Error::AmbiguousField("foo".into()),
            expected = "Field `foo` exists in multiple datasources and is ambiguous. Please qualify."

        }
    }

    mod cannot_enumerate_all_field_paths {
        test_user_error_messages! {
            cannot_enumerate_all_field_paths,
            input = Error::CannotEnumerateAllFieldPaths(crate::schema::Schema::Any),
            expected = "Insufficient schema information."
        }
    }
}
