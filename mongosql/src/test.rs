#[allow(clippy::redundant_pattern_matching)]
mod test_get_namespaces {
    macro_rules! test_get_namespaces {
        ($func_name:ident, $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? db = $current_db:expr, query = $sql:expr,) => {
            #[test]
            fn $func_name() {
                #[allow(unused_imports)]
                use crate::{get_namespaces, set, Namespace};
                let current_db = $current_db;
                let sql = $sql;
                let actual = get_namespaces(current_db, sql);
                $(assert!(matches!(actual, $expected_pat));)?
                $(assert_eq!($expected, actual);)?
            }
        };
    }

    test_get_namespaces!(
        no_collections,
        expected = Ok(set![]),
        db = "mydb",
        query = "select * from [] as arr",
    );

    test_get_namespaces!(
        implicit,
        expected = Ok(set![Namespace {
            database: "mydb".into(),
            collection: "foo".into()
        }]),
        db = "mydb",
        query = "select * from foo",
    );

    test_get_namespaces!(
        explicit,
        expected = Ok(set![Namespace {
            database: "bar".into(),
            collection: "baz".into()
        }]),
        db = "mydb",
        query = "select * from bar.baz",
    );

    test_get_namespaces!(
        duplicates,
        expected = Ok(set![Namespace {
            database: "mydb".into(),
            collection: "foo".into()
        }]),
        db = "mydb",
        query = "select * from foo a join foo b",
    );

    test_get_namespaces!(
        semantically_invalid,
        expected = Ok(set![
            Namespace {
                database: "mydb".into(),
                collection: "foo".into()
            },
            Namespace {
                database: "mydb".into(),
                collection: "bar".into()
            }
        ]),
        db = "mydb",
        query = "select a from foo join bar",
    );

    test_get_namespaces!(
        syntactically_invalid,
        expected_pat = Err(_),
        db = "mydb",
        query = "not a valid query",
    );
}

mod test_mql_schema_env_to_json_schema {
    use crate::{
        json_schema::{self, BsonType, BsonTypeName},
        map,
        mapping_registry::*,
        mql_schema_env_to_json_schema,
        options::{ExcludeNamespacesOption::*, SqlOptions},
        result::Error::Translator,
        schema::*,
        set,
        translator::Error::{DocumentSchemaTypeNotFound, ReferenceNotFound},
        Schema, SchemaCheckingMode,
    };

    macro_rules! test_mql_schema_env_to_json_schema {
        ($func_name:ident,
         schema_env = $schema_env:expr,
         mapping_registry = $mapping_registry:expr,
         sql_options = $sql_options:expr,
         $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)?) => {
            #[test]
            fn $func_name() {
                let result = mql_schema_env_to_json_schema(
                    $schema_env,
                    &$mapping_registry,
                    $sql_options,
                );

                $(assert!(matches!(result, $expected_pat));)?
                $(assert_eq!($expected, result.unwrap());)?
            }
        };
    }

    test_mql_schema_env_to_json_schema!(
        reference_not_found_in_mapping_registry,
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
                ..Default::default()
                }),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
                ..Default::default()
                }),
        },
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("foobar", 0u16),
                MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("bar", 0u16),
                MqlMappingRegistryValue::new("bar".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
        sql_options = SqlOptions::new(ExcludeNamespaces, SchemaCheckingMode::default()),
        expected_pat = Err(Translator(ReferenceNotFound(_))),
    );

    test_mql_schema_env_to_json_schema!(
        document_schema_type_not_found_in_schema_env,
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Atomic(Atomic::Integer),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
                ..Default::default()
                }),
        },
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("foo", 0u16),
                MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("bar", 0u16),
                MqlMappingRegistryValue::new("bar".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
        sql_options = SqlOptions::new(ExcludeNamespaces, SchemaCheckingMode::default()),
        expected_pat = Err(Translator(DocumentSchemaTypeNotFound(Schema::Atomic(
            Atomic::Integer
        )))),
    );

    test_mql_schema_env_to_json_schema!(
        include_namespaces_in_result_set_schema,
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
                ..Default::default()
                }),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
                ..Default::default()
                }),
        },
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("foo", 0u16),
                MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("bar", 0u16),
                MqlMappingRegistryValue::new("bar".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
        sql_options = SqlOptions::new(IncludeNamespaces, SchemaCheckingMode::default()),
        expected = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
            properties: Some(map! {
                "bar".to_string() => json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                    properties: Some(map!{
                        "b".to_string() =>json_schema::Schema {
                            bson_type: Some(BsonType::Single(BsonTypeName::String)),
                            ..Default::default()
                        }
                    }),
                    required: Some(vec!["b".to_string()]),
                    additional_properties: Some(false),
                    ..Default::default()
                },
                "foo".to_string() => json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                    properties: Some(map!{
                        "a".to_string() => json_schema::Schema {
                            bson_type: Some(BsonType::Single(BsonTypeName::String)),
                            ..Default::default()
                        }
                    }),
                    required: Some(vec!["a".to_string()]),
                    additional_properties: Some(false),
                    ..Default::default()
                }
            }),
            required: Some(vec!["bar".to_string(), "foo".to_string()]),
            additional_properties: Some(false),
            ..Default::default()
        },
    );

    test_mql_schema_env_to_json_schema!(
        exclude_namespaces_in_result_set_schema,
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "a".into() },
                additional_properties: false,
                ..Default::default()
                }),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
                ..Default::default()
                }),
        },
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(
                ("foo", 0u16),
                MqlMappingRegistryValue::new("foo".to_string(), MqlReferenceType::FieldRef),
            );
            mr.insert(
                ("bar", 0u16),
                MqlMappingRegistryValue::new("bar".to_string(), MqlReferenceType::FieldRef),
            );
            mr
        },
        sql_options = SqlOptions::new(ExcludeNamespaces, SchemaCheckingMode::default()),
        expected = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
            properties: Some(map! {
                "a".to_string() => json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                    ..Default::default()
                },
                "b".to_string() => json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                    ..Default::default()
                }
            }),
            required: Some(vec!["a".to_string(), "b".to_string()]),
            additional_properties: Some(false),
            ..Default::default()
        },
    );
}

mod test_get_select_order {
    use crate::{ast, get_select_order};

    macro_rules! test_get_select_order {
        ($func_name:ident,
         expected = $expected:expr,
         input = $input:expr
        ) => {
            #[test]
            fn $func_name() {
                let option = get_select_order($input);
                assert_eq!($expected, option);
            }
        };
    }

    test_get_select_order!(
        select_body_standard_is_some,
        expected = Some(ast::SelectBody::Standard(vec![])),
        input = &ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Standard(vec![]),
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            order_by_clause: None,
            having_clause: None,
            limit: None,
            offset: None
        })
    );

    test_get_select_order!(
        select_distinct_is_none,
        expected = None,
        input = &ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::Distinct,
                body: ast::SelectBody::Standard(vec![]),
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            order_by_clause: None,
            having_clause: None,
            limit: None,
            offset: None
        })
    );
}

mod select_list_order {
    use crate::{
        catalog::{Catalog, Namespace},
        map,
        schema::{Atomic, Document, Schema},
    };
    use lazy_static::lazy_static;

    lazy_static! {
        static ref CATALOG: Catalog = Catalog::new(map! {
            Namespace {db: "test".to_string(), collection: "foo".to_string()} => Schema::Document(Document {
                keys: map! {
                    "a".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                    "b".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                    "c".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: map!{},
                additional_properties: false,
                ..Default::default()
                }),
            Namespace {db: "test".to_string(), collection: "bar".to_string()} => Schema::Document(Document {
                keys: map! {
                    "a".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                    "b".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                    "c".to_string() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: map!{},
                additional_properties: false,
                ..Default::default()
                }),
        });
    }

    macro_rules! test_parse_select_list_order {
        ($func_name:ident, sql = $sql:expr, exclude_namespaces = $exclude_namespaces:expr, expected = $expected:expr) => {
            #[test]
            fn $func_name() {
                #[allow(unused_imports)]
                use crate::{
                    translate_sql, ExcludeNamespacesOption, SchemaCheckingMode, SqlOptions,
                };
                let translation = translate_sql(
                    "test",
                    $sql,
                    &*CATALOG,
                    SqlOptions {
                        schema_checking_mode: SchemaCheckingMode::default(),
                        exclude_namespaces: $exclude_namespaces,
                        allow_order_by_missing_columns: false,
                    },
                );
                assert!(translation.is_ok());
                assert_eq!(translation.unwrap().select_order, $expected)
            }
        };
    }

    test_parse_select_list_order!(
        star_sorted,
        sql = "select * from foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["foo".to_string(), "a".to_string()],
            vec!["foo".to_string(), "b".to_string()],
            vec!["foo".to_string(), "c".to_string()]
        ]
    );

    test_parse_select_list_order!(
        star_multiple_collections_sorted,
        sql = "select * from foo, bar",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["bar".to_string(), "a".to_string()],
            vec!["bar".to_string(), "b".to_string()],
            vec!["bar".to_string(), "c".to_string()],
            vec!["foo".to_string(), "a".to_string()],
            vec!["foo".to_string(), "b".to_string()],
            vec!["foo".to_string(), "c".to_string()],
        ]
    );

    test_parse_select_list_order!(
        substar_simple,
        sql = "select foo.* from foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["foo".to_string(), "a".to_string()],
            vec!["foo".to_string(), "b".to_string()],
            vec!["foo".to_string(), "c".to_string()]
        ]
    );

    test_parse_select_list_order!(
        one_collection_non_alphabetical,
        sql = "select c, a, b from foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "c".to_string()],
            vec!["".to_string(), "a".to_string()],
            vec!["".to_string(), "b".to_string()]
        ]
    );

    test_parse_select_list_order!(
        fields_from_two_collections,
        sql = " select foo.a, bar.b from bar, foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "a".to_string()],
            vec!["".to_string(), "b".to_string()]
        ]
    );

    test_parse_select_list_order!(
        substar_between_fields,
        sql = " select foo.a, bar.*, foo.b from bar, foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "a".to_string()],
            vec!["bar".to_string(), "a".to_string()],
            vec!["bar".to_string(), "b".to_string()],
            vec!["bar".to_string(), "c".to_string()],
            vec!["".to_string(), "b".to_string()]
        ]
    );

    test_parse_select_list_order!(
        multiple_substars,
        sql = " select foo.*, bar.* from bar, foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["foo".to_string(), "a".to_string()],
            vec!["foo".to_string(), "b".to_string()],
            vec!["foo".to_string(), "c".to_string()],
            vec!["bar".to_string(), "a".to_string()],
            vec!["bar".to_string(), "b".to_string()],
            vec!["bar".to_string(), "c".to_string()],
        ]
    );

    test_parse_select_list_order!(
        fields_aliased_datasources,
        sql = " select f.a, b.b, f.c from bar as b, foo as f",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "a".to_string()],
            vec!["".to_string(), "b".to_string()],
            vec!["".to_string(), "c".to_string()],
        ]
    );

    test_parse_select_list_order!(
        substar_with_aliased_datasources,
        sql = " select f.a, b.*, f.b from bar as b, foo as f",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "a".to_string()],
            vec!["b".to_string(), "a".to_string()],
            vec!["b".to_string(), "b".to_string()],
            vec!["b".to_string(), "c".to_string()],
            vec!["".to_string(), "b".to_string()],
        ]
    );

    test_parse_select_list_order!(
        aliased_fields,
        sql = " select foo.a as f_a, bar.a as b_a from foo, bar",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "f_a".to_string()],
            vec!["".to_string(), "b_a".to_string()],
        ]
    );

    test_parse_select_list_order!(
        aggregations_and_fields,
        sql = "select foo.a, sum(bar.a), foo.b,  count(bar.b) from foo, bar group by foo.a, foo.b",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "a".to_string()],
            vec!["".to_string(), "_2".to_string()],
            vec!["".to_string(), "b".to_string()],
            vec!["".to_string(), "_4".to_string()],
        ]
    );

    test_parse_select_list_order!(
        values_simple,
        sql = "select values {'b': foo.b, 'a': 1} from foo",
        exclude_namespaces = ExcludeNamespacesOption::IncludeNamespaces,
        expected = vec![
            vec!["".to_string(), "b".to_string()],
            vec!["".to_string(), "a".to_string()],
        ]
    );
    test_parse_select_list_order!(
        star_no_reordering_exclude_namespaces,
        sql = "select * from foo",
        exclude_namespaces = ExcludeNamespacesOption::ExcludeNamespaces,
        expected = vec![
            vec!["a".to_string()],
            vec!["b".to_string()],
            vec!["c".to_string()]
        ]
    );

    test_parse_select_list_order!(
        one_collection_non_alphabetical_exclude_namespaces,
        sql = "select c, a, b from foo",
        exclude_namespaces = ExcludeNamespacesOption::ExcludeNamespaces,
        expected = vec![
            vec!["c".to_string()],
            vec!["a".to_string()],
            vec!["b".to_string()]
        ]
    );

    test_parse_select_list_order!(
        fields_from_two_collections_exclude_namespaces,
        sql = " select foo.a, bar.b from bar, foo",
        exclude_namespaces = ExcludeNamespacesOption::ExcludeNamespaces,
        expected = vec![vec!["a".to_string()], vec!["b".to_string()]]
    );
}
