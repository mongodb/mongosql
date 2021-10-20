use crate::{ast::rewrites::rewrite_query, parser::Parser};
use serde::{Deserialize, Serialize};
use std::{fs, io::Read, path::PathBuf};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("rewrite test failed for test {test}: expected {expected:?}, actual {actual:?}")]
    RewriteTest {
        test: String,
        expected: String,
        actual: String,
    },
    #[error("failed to read directory: {0}")]
    InvalidDirectory(String),
    #[error("failed to load file paths: {0}")]
    InvalidFilePath(String),
    #[error("failed to read file: {0}")]
    InvalidFile(String),
    #[error("unable to read file to string: {0}")]
    CannotReadFileToString(String),
    #[error("unable to deserialize YAML file: {0}")]
    CannotDeserializeYaml(String),
}

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct YamlTest {
    pub tests: Vec<Test>,
}

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct Test {
    pub description: String,
    pub query: String,
    pub result: Option<String>,
    pub error: Option<String>,
    pub skip_reason: Option<String>,
}

pub fn load_file_paths(dir: PathBuf) -> Result<Vec<String>, Error> {
    let mut paths: Vec<String> = vec![];
    let entries = fs::read_dir(dir).map_err(|e| Error::InvalidDirectory(format!("{:?}", e)))?;
    for entry in entries {
        match entry {
            Ok(de) => {
                let path = de.path();
                if path.extension().unwrap() == "yml" {
                    paths.push(path.to_str().unwrap().to_string());
                }
            }
            Err(e) => return Err(Error::InvalidFilePath(format!("{:?}", e))),
        };
    }
    Ok(paths)
}

pub fn parse_yaml(path: &str) -> Result<YamlTest, Error> {
    let mut f = fs::File::open(path).map_err(|e| Error::InvalidFile(format!("{:?}", e)))?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(|e| Error::CannotReadFileToString(format!("{:?}", e)))?;
    let yaml: YamlTest = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml(format!("{:?}", e)))?;
    Ok(yaml)
}

#[test]
pub fn run_test() -> Result<(), Error> {
    let paths = load_file_paths(PathBuf::from("../tests/spec_tests/rewrite_tests")).unwrap();
    for path in paths {
        let yaml = parse_yaml(&path).unwrap();
        for test in yaml.tests {
            match test.skip_reason {
                Some(_) => continue,
                None => {
                    let parse_res = Parser::new().parse_query(test.query.as_str());
                    let ast = parse_res.expect("expected input to parse, but it failed");
                    let rewrite_res = rewrite_query(ast);
                    match test.error {
                        Some(_) => assert!(rewrite_res.is_err()),
                        None => {
                            let expected = test.result.unwrap();
                            let actual = format!("{}", rewrite_res.unwrap());
                            if expected != actual {
                                return Err(Error::RewriteTest {
                                    test: test.description,
                                    expected,
                                    actual,
                                });
                            }
                        }
                    };
                }
            }
        }
    }
    Ok(())
}
