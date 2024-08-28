use crate::command::{Command, CommandOptions, CommandType::*};
use mongodb::bson::{self, doc};
use mongosql::{
    json_schema::{self, BsonType, BsonTypeName},
    map,
};
use std::collections::HashMap;

macro_rules! test_deserializing_into_command {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let expected_command = $expected;

            // convert command to bytes
            let mut v: Vec<u8> = Vec::new();
            $input.to_writer(&mut v).unwrap();
            let bytes = v.as_slice();

            let actual_command = Command::new(bytes);

            assert_eq!(actual_command, expected_command)
        }
    };
}

macro_rules! test_command_handler {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let expected_document = $expected;

            let actual_document = $input.run().unwrap();

            assert_eq!(actual_document, expected_document)
        }
    };
}

macro_rules! hashmap(
    { $($key:expr => $value:expr),+ } => {
            {
                let mut m = HashMap::new();
                $(
                  m.insert($key, $value);
                )+
                m
            }
        };
    );

mod deserializing_tests {
    use super::*;

    test_deserializing_into_command!(
        deserialize_translate_command,
        expected = Command {
            command: Translate,
            options: CommandOptions {
                sql: Some("SELECT * FROM db1.coll1".to_string()),
                exclude_namespaces: Some(true),
                relax_schema_checking: Some(true),
                db: Some("db1".to_string()),
                schema_catalog: Some(map!(
                    "db1".to_string() => map!(
                        "coll1".to_string() => json_schema::Schema{
                            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                            properties: Some(hashmap!(
                                "field1".to_string() => json_schema::Schema{
                                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                                    ..json_schema::Schema::default()
                                }
                            )),
                            ..Default::default()
                        }
                    )
                )),
                ..Default::default()
            },
        },
        input = doc! {
            "command": "translate",
            "options": {
                "sql": "SELECT * FROM db1.coll1",
                "db": "db1",
                "excludeNamespaces": true,
                "relaxSchemaChecking": true,
                "schemaCatalog": {
                    "db1": {
                        "coll1": {
                            "bsonType": "object",
                            "properties": {
                                "field1": {
                                    "bsonType": "string"
                                }
                            }
                        }
                    }
                },
            },
        }
    );

    test_deserializing_into_command!(
        deserialize_get_namespaces_command,
        expected = Command {
            command: GetNamespaces,
            options: CommandOptions {
                sql: Some("SELECT * FROM db1.coll1".to_string()),
                db: Some("db1".to_string()),
                ..Default::default()
            },
        },
        input = doc! {
            "command": "getNamespaces",
            "options": {
                "sql": "SELECT * FROM db1.coll1",
                "db": "db1",
            },
        }
    );

    test_deserializing_into_command!(
        deserialize_get_mongosqltranslate_version_command,
        expected = Command {
            command: GetMongosqlTranslateVersion,
            options: CommandOptions {
                ..Default::default()
            },
        },
        input = doc! {
            "command": "getMongosqlTranslateVersion",
            "options": {},
        }
    );

    test_deserializing_into_command!(
        deserialize_check_driver_version_command,
        expected = Command {
            command: CheckDriverVersion,
            options: CommandOptions {
                driver_version: Some("0.0.0".to_string()),
                ..Default::default()
            },
        },
        input = doc! {
            "command": "checkDriverVersion",
            "options": {
                "driverVersion": "0.0.0"
            },
        }
    );

    #[test]
    #[should_panic(
        expected = "Deserializing the provided Bson::Document into `Command` data type failed."
    )]
    fn invalid_command_should_panic() {
        let command = doc! {
            "command": "checkDriver",
            "options": {
                "driverVersion": "0.0.0"
            },
        };

        // convert command to bytes
        let mut v: Vec<u8> = Vec::new();
        command.to_writer(&mut v).unwrap();
        let bytes = v.as_slice();

        let _actual_command = Command::new(bytes);
    }
}

mod translate_tests {
    use super::*;

    test_command_handler!(
        valid_translate_command_should_succeed,
        expected = doc! {
            "target_db": "db1".to_string(),
            "target_collection": "coll1".to_string(),
            "pipeline": &bson::to_bson(&vec![
                    doc! {
                        "$project": doc!{
                            "coll1": "$$ROOT".to_string(),
                            "_id": 0,
                        }
                    },
                    doc! {
                        "$replaceWith": "$coll1".to_string()
                    },
                ]).expect("failed to convert expected_pipeline to bson"),
            "result_set_schema": &json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                    properties: Some(hashmap!(
                        "field1".to_string() => json_schema::Schema{
                            bson_type: Some(BsonType::Single(BsonTypeName::String)),
                            ..json_schema::Schema::default()
                        }
                    )),
                    additional_properties: Some(false),
                    required: Some(vec!["field1".to_string()]),
                    ..Default::default()
                }.to_bson().expect("failed to convert expected_result_set_schema to bson"),
            "select_order": &bson::to_bson(&vec![vec!["field1".to_string()]]).expect("failed to convert expected_select_order to bson"),
        },
        input = Command {
            command: Translate,
            options: CommandOptions {
                sql: Some("SELECT * FROM db1.coll1".to_string()),
                exclude_namespaces: Some(true),
                relax_schema_checking: Some(true),
                db: Some("db1".to_string()),
                schema_catalog: Some(map!(
                    "db1".to_string() => map!(
                        "coll1".to_string() => json_schema::Schema{
                            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                            properties: Some(hashmap!(
                                "field1".to_string() => json_schema::Schema{
                                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                                    ..json_schema::Schema::default()
                                }
                            )),
                            additional_properties: Some(false),
                            ..Default::default()
                        }
                    )
                )),
                ..Default::default()
            },
        }
    );

    test_command_handler!(
        valid_translate_command_with_extra_parameter_should_succeed,
        expected = doc! {
            "target_db": "db1".to_string(),
            "target_collection": "coll1".to_string(),
            "pipeline": &bson::to_bson(&vec![
                    doc! {
                        "$project": doc!{
                            "coll1": "$$ROOT".to_string(),
                            "_id": 0,
                        }
                    },
                    doc! {
                        "$replaceWith": "$coll1".to_string()
                    },
                ]).expect("failed to convert expected_pipeline to bson"),
            "result_set_schema": &json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                    properties: Some(hashmap!(
                        "field1".to_string() => json_schema::Schema{
                            bson_type: Some(BsonType::Single(BsonTypeName::String)),
                            ..json_schema::Schema::default()
                        }
                    )),
                    additional_properties: Some(false),
                    required: Some(vec!["field1".to_string()]),
                    ..Default::default()
                }.to_bson().expect("failed to convert expected_result_set_schema to bson"),
            "select_order": &bson::to_bson(&vec![vec!["field1".to_string()]]).expect("failed to convert expected_select_order to bson"),
        },
        input = Command {
            command: Translate,
            options: CommandOptions {
                sql: Some("SELECT * FROM db1.coll1".to_string()),
                exclude_namespaces: Some(true),
                relax_schema_checking: Some(true),
                db: Some("db1".to_string()),
                schema_catalog: Some(map!(
                    "db1".to_string() => map!(
                        "coll1".to_string() => json_schema::Schema{
                            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                            properties: Some(hashmap!(
                                "field1".to_string() => json_schema::Schema{
                                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                                    ..json_schema::Schema::default()
                                }
                            )),
                            additional_properties: Some(false),
                            ..Default::default()
                        }
                    )
                )),
                driver_version: Some("extra_parameter".to_string()),
                ..Default::default()
            },
        }
    );

    #[test]
    #[should_panic(expected = "`db` parameter missing for Translate CommandType")]
    fn translate_command_with_missing_parameter_should_panic() {
        let command = Command {
            command: Translate,
            options: CommandOptions {
                sql: Some("SELECT * FROM db1.coll1".to_string()),
                exclude_namespaces: Some(true),
                relax_schema_checking: Some(true),
                schema_catalog: Some(map!(
                    "db1".to_string() => map!(
                        "coll1".to_string() => json_schema::Schema{
                            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
                            properties: Some(hashmap!(
                                "field1".to_string() => json_schema::Schema{
                                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                                    ..json_schema::Schema::default()
                                }
                            )),
                            additional_properties: Some(false),
                            ..Default::default()
                        }
                    )
                )),
                ..Default::default()
            },
        };

        let _actual = command.run();
    }
}
