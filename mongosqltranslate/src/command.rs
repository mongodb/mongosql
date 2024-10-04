use crate::{
    DEV_JDBC_VERSION_SUFFIX, DEV_ODBC_VERSION, MINIMUM_COMPATIBLE_JDBC_VERSION,
    MINIMUM_COMPATIBLE_ODBC_VERSION, MONGOSQLTRANSLATE_VERSION, SNAPSHOT_JDBC_VERSION_SUFFIX,
};
use mongodb::bson::{doc, Bson, Deserializer, Document, Serializer};
use mongosql::{
    build_catalog_from_catalog_schema, json_schema,
    options::{ExcludeNamespacesOption, SqlOptions},
    SchemaCheckingMode,
};
use semver::Version;
use serde::{ser::Serialize, Deserialize};
use std::collections::BTreeMap;
use CommandType::*;

type Result<T> = std::result::Result<T, Box<dyn std::error::Error>>;

#[derive(serde::Serialize, Deserialize, Debug, PartialEq)]
pub(crate) struct Command {
    pub(crate) command: CommandType,
    pub(crate) options: CommandOptions,
}

#[derive(serde::Serialize, Deserialize, Debug, PartialEq)]
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

#[derive(serde::Serialize, Deserialize, Debug, PartialEq, Default)]
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
    pub(crate) odbc_driver: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub(crate) test: Option<bool>,
}

macro_rules! serialize_expr_to_bson {
    ($input:expr) => {{
        let serializer = Serializer::new();
        let serializer = serde_stacker::Serializer::new(serializer);
        let serialized_expr = $input
            .serialize(serializer)
            .expect("Failed to serialize input expression to bson");

        serialized_expr
    }};
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

        let so = serialize_expr_to_bson!(translation.select_order);

        Ok(doc! {
            "target_db": translation.target_db,
            "target_collection": translation.target_collection.unwrap_or_default(),
            "pipeline": translation.pipeline,
            "result_set_schema": &translation.result_set_schema.to_bson().expect("failed to convert result_set_schema to bson"),
            "select_order": &so,
        })
    }

    /// This is the handler function for the GetNamespaces Command. It returns a Result<bson::Document>
    /// containing the namespaces referenced by the the provided SQL query when executed in the provided database.
    /// The necessary CommandOptions includes `sql` and `db`. Extra CommandOptions will be ignored.
    fn get_namespaces(&self) -> Result<Document> {
        let current_db = self
            .options
            .db
            .as_ref()
            .expect("`db` parameter missing for GetNamespaces CommandType");
        let sql = self
            .options
            .sql
            .as_ref()
            .expect("`sql` parameter missing for GetNamespaces CommandType");

        let namespaces = mongosql::get_namespaces(current_db, sql).map_err(|e| e.to_string())?;

        let ns = serialize_expr_to_bson!(namespaces);

        Ok(doc! {
            "namespaces": &ns
        })
    }

    /// This is the handler function for the GetMongosqlTranslateVersion Command. It returns a
    /// Result<bson::Document> containing the version of the mongosqltranslate library.
    /// There are NO necessary CommandOptions. Extra CommandOptions will be ignored.
    fn get_mongosqltranslate_version(&self) -> Result<Document> {
        Ok(doc! {
            "version": &*MONGOSQLTRANSLATE_VERSION,
        })
    }

    /// This is the handler function for the CheckDriverVersion Command. It returns a Result<bson::Document>
    /// containing a boolean value indicating if the driver version is compatible with the mongosqltranslate library.
    /// The only necessary CommandOption is `driver_version`. The `odbc_driver` CommandOption is not necessary; however,
    /// if not specified, it will be assumed that the JDBC driver is being used. Extra CommandOptions will be ignored.
    /// Additionally, the `driver_version` must be a valid SemVer version (https://semver.org/).
    fn check_driver_version(&self) -> Result<Document> {
        let driver_version_command_option = self
            .options
            .driver_version
            .as_ref()
            .expect("`driver_version` parameter missing for CheckDriverVersion CommandType");

        let driver_version = match Version::parse(driver_version_command_option) {
            Ok(version) => version,
            Err(_) => return Err(format!("Invalid `driver_version`: \"{}\". The `driver_version` must be a valid SemVer version (https://semver.org/).", driver_version_command_option).into())
        };

        // *** Dev versions ***
        // For JDBC, versions ending with -SNAPSHOT or -dirty are development/snapshot versions and are valid
        // For ODBC, the development/snapshot version is always `0.0.0`
        if (self.options.odbc_driver == Some(true) && DEV_ODBC_VERSION.matches(&driver_version))
            || (self.options.odbc_driver != Some(true)
                && (driver_version_command_option.ends_with(SNAPSHOT_JDBC_VERSION_SUFFIX)
                    || driver_version_command_option.ends_with(DEV_JDBC_VERSION_SUFFIX)))
        {
            return Ok(doc! {
                "compatible": true
            });
        }

        // *** Release versions ***
        let minimum_version = if let Some(true) = self.options.odbc_driver {
            // The ODBC Driver is being used
            &*MINIMUM_COMPATIBLE_ODBC_VERSION
        } else {
            // The JDBC Driver is being used
            &*MINIMUM_COMPATIBLE_JDBC_VERSION
        };

        dbg!(minimum_version.matches(&driver_version));
        Ok(doc! {
            "compatible": minimum_version.matches(&driver_version)
        })
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

#[test]
pub fn test_versions_compatibility() {
    let minimum_version = semver::VersionReq::parse(">=2.0.0-alpha").unwrap();

    let compatible_min_versions: [&str; 6] = [
        "2.0.0-alpha",
        "2.0.0-beta",
        "2.0.0",
        "2.1.0",
        "2.1.10",
        "3.0.0",
    ];
    let incompatible_min_versions: [&str; 5] =
        ["1.2.0-alpha", "1.0.3-beta", "1.4.5", "0.1.0", "0.0.0"];

    for compatible_version in &compatible_min_versions {
        let driver_version = Version::parse(compatible_version).unwrap();

        assert!(minimum_version.matches(&driver_version));
    }

    for incompatible_version in &incompatible_min_versions {
        let driver_version = match Version::parse(incompatible_version) {
            Ok(version) => version,
            Err(e) => panic!("{}", e),
        };

        assert!(!minimum_version.matches(&driver_version));
    }
}
