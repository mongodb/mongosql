use std::ffi::OsStr;
use std::fs;
use std::fs::read_dir;
use std::fs::File;
use std::fs::ReadDir;
use std::io::Write;
use test_utils::TestProcessor;

const GENERATED_DIRECTORY: &str = "src/generated";
const GENERATED_MOD: &str = "src/generated/mod.rs";
const YAML_TEST_DIR: &str = "../tests";

fn main() {
    let remove = fs::remove_dir_all(GENERATED_DIRECTORY);
    let create = fs::create_dir(GENERATED_DIRECTORY);
    let mut mod_file = fs::OpenOptions::new()
        .append(true)
        .create(true)
        .open(GENERATED_MOD)
        .unwrap();
    write!(mod_file, include_str!("templates/mod_header_template")).unwrap();

    match (remove, create) {
        (Ok(_), Ok(_)) => {}
        // in this case, it may be the first time run so there is nothing to delete.
        // No reason to panic here.
        (Err(_), Ok(_)) => {}
        (Ok(_), Err(why)) => panic!("failed to create test direcotry: {why:?}"),
        (Err(delete_err), Err(create_err)) => panic!(
            "failed to delete and create test directory:\n{:?}\n{:?}",
            delete_err, create_err
        ),
    }

    let test_data_directories = read_dir(YAML_TEST_DIR).unwrap();
    traverse(test_data_directories, GENERATED_DIRECTORY, mod_file);
}

// traverse the tests directory, finding all yml files.
// process each yml file and create a test file for each test case
fn traverse(path: ReadDir, out_dir: &str, mod_file: File) {
    for entry in path {
        let entry = entry.unwrap();
        if entry.file_type().unwrap().is_dir() {
            traverse(
                read_dir(entry.path()).unwrap(),
                out_dir,
                mod_file.try_clone().unwrap(),
            )
        } else if entry.file_type().unwrap().is_file()
            && entry.path().extension() == Some(OsStr::new("yml"))
        {
            TestProcessor::process(entry, mod_file.try_clone().unwrap(), out_dir);
        }
    }
}
