extern crate yaml_rust;
use regex::Regex;
use std::env;
use std::fs;
use std::io;
use std::io::*;
use std::path::PathBuf;
use visitgen::gen_visitor;
use yaml_rust::Yaml;
use yaml_rust::YamlLoader;

fn main() {
    lalrpop::process_root().unwrap();
    gen_query_tests();
    gen_rewrite_tests();
    gen_visitor("src/ast/definitions.rs");
    gen_visitor("src/ir/definitions.rs");
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=parser/mongosql.lalrpop");
}

fn load_file_paths(dir: PathBuf) -> io::Result<Vec<String>> {
    let mut paths: Vec<String> = vec![];
    for entry in fs::read_dir(dir)? {
        let entry = entry?.path();
        paths.push(entry.to_str().unwrap().to_string());
    }
    Ok(paths)
}

fn read_test_file(file: String) -> Vec<Yaml> {
    let mut f = fs::File::open(file).expect("failed to open test data file");
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .expect("failed to read test data file");
    YamlLoader::load_from_str(contents.as_str()).unwrap()
}

#[derive(Debug, PartialEq)]
pub struct Test {
    name: String,
    query: String,
}

impl Test {
    fn to_function(&self) -> String {
        format!(
            r##"
#[test]
fn {}() {{
    let res = Parser::new().parse_query(r#"{}"#);
    res.expect("expected input to parse, but it failed");
}}
"##,
            self.name, self.query,
        )
    }
}

fn gen_query_tests() {
    let mut dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    dir.push("../tests/spec_tests/query_tests");
    let paths = load_file_paths(dir).unwrap();
    gen_test("query", paths);
}

fn gen_rewrite_tests() {
    let mut dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    dir.push("../tests/spec_tests/rewrite_tests");
    let paths = load_file_paths(dir).unwrap();
    gen_test("rewrite", paths);
}

/// gen_test generates test functions for all yaml tests of the specified test type
/// (query or rewrite), except for those that have an "error" and/or "skip_reason" field.
fn gen_test(test_type: &str, file_paths: Vec<String>) {
    let out_dir = env::var_os("OUT_DIR").unwrap();
    let mut dest_path = PathBuf::new();
    dest_path.push(&out_dir);
    if test_type == "query" {
        dest_path.push("parser/query_tests.rs");
    } else {
        dest_path.push("parser/rewrite_tests.rs");
    }
    let mut dest_file = fs::File::create(&dest_path).unwrap();
    let mut test_structs: Vec<Test> = vec![];

    for path in file_paths {
        let contents = read_test_file(path.clone());
        let tests = &contents[0]["tests"].as_vec().unwrap();

        for test in tests.iter() {
            let map = test.as_hash().unwrap();
            let description = map.get(&Yaml::from_str("description")).unwrap();
            let query = map.get(&Yaml::from_str("query")).unwrap();
            let parse_err = map.get(&Yaml::from_str("parse_error"));
            let algebrize_err = map.get(&Yaml::from_str("algebrize_error"));
            let skip_reason = map.get(&Yaml::from_str("skip_reason"));

            // Skip tests that have the fields "error" and/or "skip_reason"
            if parse_err.is_some() || algebrize_err.is_some() || skip_reason.is_some() {
                continue;
            }

            // Convert the test's name into snake case by replacing spaces and non-alphanumeric
            // characters with "_", except for ".", which changes to "_DOT_.
            let re = Regex::new(r"[ ;\-,()_]+").unwrap();
            let mut name = Yaml::into_string(description.clone()).unwrap();
            name = re
                .replace_all(name.replace(".", "_DOT_").to_lowercase().as_str(), "_")
                .parse()
                .unwrap();

            let t: Test = Test {
                name,
                query: Yaml::into_string(query.clone()).unwrap(),
            };
            test_structs.push(t);
        }
    }

    let test_output = format!(
        "\
#[cfg(test)]
mod {} {{
use super::*;
{}
}}
",
        test_type.to_owned() + "_tests",
        test_structs
            .iter()
            .map(|t| t.to_function())
            .collect::<String>()
    );
    dest_file.write_all(test_output.as_bytes()).unwrap()
}
