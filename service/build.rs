use std::path::PathBuf;
use std::process::Command;
use std::{env, fs};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // TODO SQL-2214: Standardize the way version info is retrieved
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    let res = Command::new("git").arg("describe").output()?;
    let version = std::str::from_utf8(&res.stdout)?.trim();
    let version_decl = format!(r#"pub const VERSION: &str = "{}";"#, version);
    fs::write(out_dir.join("version.rs"), version_decl)?;

    Ok(())
}
