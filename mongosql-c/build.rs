use std::{env, fs, path::Path, process::Command, str};

fn main() {
    let res = Command::new("git").arg("describe").output().unwrap();
    let version = str::from_utf8(&res.stdout).unwrap();
    let version_decl = format!(r#"pub static VERSION: &str = "{}";"#, version.trim());

    let out_dir = env::var_os("OUT_DIR").unwrap();
    let dest_path = Path::new(&out_dir).join("version.rs");
    fs::write(&dest_path, version_decl).unwrap();
}
