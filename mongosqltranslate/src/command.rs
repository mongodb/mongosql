use mongodb::bson::{doc, Bson, Deserializer, Document, Serializer};
use mongosql::{
    build_catalog_from_catalog_schema, json_schema,
    options::{ExcludeNamespacesOption, SqlOptions},
    SchemaCheckingMode,
};
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use CommandType::*;

type Result<T> = std::result::Result<T, Box<dyn std::error::Error>>;

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
    #[serde(rename = "test")]
    Test,
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
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) test: Option<bool>,
}

#[allow(dead_code)]
/// Command represents the command that is sent to the mongosqltranslate service.
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

    pub(crate) fn run(&self) -> Result<Document> {
        let command: fn(&Self) -> Result<Document> = match self.command {
            Translate => Self::translate,
            GetNamespaces => Self::get_namespaces,
            GetMongosqlTranslateVersion => Self::get_mongosqltranslate_version,
            CheckDriverVersion => Self::check_driver_version,
            Test => Self::test,
        };
        command(self)
    }

    /// This is the handler function for the Translate Command. The necessary CommandOptions includes `sql`,
    /// `exclude_namespaces`, `relax_schema_checking`, `db`, and `schema_catalog`. Extra CommandOptions will be ignored.
    /// This function returns a Result<bson::Document> representing a mongosql::Translation.
    fn translate(&self) -> Result<Document> {
        use serde::ser::Serialize;

        let schema_checking_mode = if self
            .options
            .relax_schema_checking
            .expect("`relax_schema_checking` parameter missing for Translate CommandType")
        {
            SchemaCheckingMode::Relaxed
        } else {
            SchemaCheckingMode::Strict
        };

        let exclude_namespaces_mode = if self
            .options
            .exclude_namespaces
            .expect("`exclude_namespaces` parameter missing for Translate CommandType")
        {
            ExcludeNamespacesOption::ExcludeNamespaces
        } else {
            ExcludeNamespacesOption::IncludeNamespaces
        };

        let catalog = build_catalog_from_catalog_schema(
            self.options
                .schema_catalog
                .as_ref()
                .expect("`schema_catalog` parameter missing for Translate CommandType")
                .clone(),
        )
        .map_err(|e| e.to_string())?;

        let translation = mongosql::translate_sql(
            self.options
                .db
                .as_ref()
                .expect("`db` parameter missing for Translate CommandType"),
            self.options
                .sql
                .as_ref()
                .expect("`sql` parameter missing for Translate CommandType"),
            &catalog,
            SqlOptions::new(exclude_namespaces_mode, schema_checking_mode),
        )
        .map_err(|e| e.to_string())?;

        let serializer = Serializer::new();
        let serializer = serde_stacker::Serializer::new(serializer);
        let so = translation
            .select_order
            .serialize(serializer)
            .expect("failed to convert select_order to bson");

        Ok(doc! {
            "target_db": translation.target_db,
            "target_collection": translation.target_collection.unwrap_or_default(),
            "pipeline": translation.pipeline,
            "result_set_schema": &translation.result_set_schema.to_bson().expect("failed to convert result_set_schema to bson"),
            "select_order": &so,
        })
    }

    // Placeholder for CommandType::GetNamespaces
    fn get_namespaces(&self) -> Result<Document> {
        unimplemented!()
    }

    // Placeholder for CommandType::GetMongosqltranslateVersion
    fn get_mongosqltranslate_version(&self) -> Result<Document> {
        unimplemented!()
    }

    // Placeholder for CommandType::CheckDriverVersion
    fn check_driver_version(&self) -> Result<Document> {
        unimplemented!()
    }

    // For testing purposes
    fn test(&self) -> Result<Document> {
        match self.options.test {
            Some(true) => Ok(doc! {"success": true}),
            Some(false) => Err("Test errored".into()),
            None => panic!("Test success value not provided"),
        }
    }
}
