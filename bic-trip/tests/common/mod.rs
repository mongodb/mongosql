use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use std::{fs, io::Read};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("failed to read directory: {0:?}")]
    InvalidDirectory(String),
    #[error("failed to load file paths: {0:?}")]
    InvalidFilePath(String),
    #[error("failed to read file: {0:?}")]
    InvalidFile(String),
    #[error("unable to read file to string: {0:?}")]
    CannotReadFileToString(String),
    #[error("unable to deserialize YAML file: {0:?}")]
    CannotDeserializeYaml((String, serde_yaml::Error)),
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TransitionReportTest {
    pub description: String,
    pub skip_reason: Option<String>,
    pub valid_query_count: usize,
    pub invalid_query_count: usize,
    pub unsupported_query_count: usize,
    pub collections_count: usize,
    pub subpath_fields_count: usize,
    pub array_datasources_count: usize,
    pub valid_query_users: Option<Vec<Vec<String>>>,
    pub log_dir: String,
}

pub fn load_transition_report_test_files(dir: PathBuf) -> Result<Vec<TransitionReportTest>, Error> {
    let entries = fs::read_dir(dir).map_err(|e| Error::InvalidDirectory(format!("{e:?}")))?;

    let mut tests = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|e| Error::InvalidFilePath(format!("{e:?}")))?;
        let path = entry.path();
        if path.is_file()
            && (path.extension().unwrap_or_default() == "yaml"
                || path.extension().unwrap_or_default() == "yml")
        {
            let test = parse_transition_report_yaml_file(entry.path())?;
            tests.push(test);
        }
    }
    Ok(tests)
}

pub fn parse_transition_report_yaml_file(path: PathBuf) -> Result<TransitionReportTest, Error> {
    let mut f = fs::File::open(&path).map_err(|e| Error::InvalidFile(format!("{e:?}")))?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(|e| Error::CannotReadFileToString(format!("{e:?}")))?;
    let yaml: TransitionReportTest = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml((format!("in file: {:?}", path), e)))?;
    Ok(yaml)
}
