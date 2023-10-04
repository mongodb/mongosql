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
            }),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
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
            }),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
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
                    properties: Some(map! {
                        "b".to_string() => json_schema::Schema {
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
                    properties: Some(map! {
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
            }),
            ("bar", 0u16).into() => Schema::Document( Document{
                keys: map! {
                    "b".into() => Schema::Atomic(Atomic::String)
                },
                required: set! { "b".into() },
                additional_properties: false,
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
                "b".to_string() => json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                    ..Default::default()
                },
                "a".to_string() => json_schema::Schema {
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
