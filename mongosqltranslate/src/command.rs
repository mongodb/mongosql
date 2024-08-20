use mongodb::bson::{Bson, Deserializer, Document};
use mongosql::json_schema;
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use CommandType::*;

#[derive(Serialize, Deserialize, Debug, PartialEq)]
pub(crate) struct Command {
    pub(crate) command: CommandType,
    pub(crate) options: CommandOptions,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
pub(crate) enum CommandType {
    #[serde(rename = "translate")]
    Translate,
    #[serde(rename = "getNamespaces")]
    GetNamespaces,
    #[serde(rename = "getMongosqlTranslateVersion")]
    GetMongosqlTranslateVersion,
    #[serde(rename = "checkDriverVersion")]
    CheckDriverVersion,
}

#[derive(Serialize, Deserialize, Debug, PartialEq, Default)]
#[serde(rename_all = "camelCase")]
pub(crate) struct CommandOptions {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) sql: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) exclude_namespaces: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) relax_schema_checking: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) db: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) schema_catalog: Option<BTreeMap<String, BTreeMap<String, json_schema::Schema>>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) driver_version: Option<String>,
}

#[allow(dead_code)]
impl Command {
    pub(crate) fn new(bson_bytes_slice: &[u8]) -> Self {
        let reader = std::io::Cursor::new(bson_bytes_slice);
        let as_bson = Bson::from(
            Document::from_reader(reader)
                .expect("Deserializing the provided byte stream into bson::Document failed."),
        );
        let deserializer = Deserializer::new(as_bson);
        let deserializer = serde_stacker::Deserializer::new(deserializer);
        Deserialize::deserialize(deserializer)
            .expect("Deserializing the provided Bson::Document into `Command` data type failed.")
    }

    pub(crate) fn run(&self) -> Document {
        let command: fn(&Self) -> Document = match self.command {
            Translate => Self::translate,
            GetNamespaces => Self::get_namespaces,
            GetMongosqlTranslateVersion => Self::get_mongosqltranslate_version,
            CheckDriverVersion => Self::check_driver_version,
        };
        command(self)
    }

    // Placeholder for CommandType::Translate
    fn translate(&self) -> Document {
        unimplemented!()
    }

    // Placeholder for CommandType::GetNamespaces
    fn get_namespaces(&self) -> Document {
        unimplemented!()
    }

    // Placeholder for CommandType::GetMongosqltranslateVersion
    fn get_mongosqltranslate_version(&self) -> Document {
        unimplemented!()
    }

    // Placeholder for CommandType::CheckDriverVersion
    fn check_driver_version(&self) -> Document {
        unimplemented!()
    }
}

#[cfg(test)]
mod command_tests {
    use super::*;
    use mongodb::bson::doc;
    use mongosql::{
        json_schema::{BsonType, BsonTypeName},
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
                driver_version: None,
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
