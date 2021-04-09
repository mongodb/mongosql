use std::path::Path;

/// convert_to_snake_case converts a &str assumed to be in camelCase
/// into snake case. It does this by replacing a capital letter, e.g., X,
/// with _x.
pub(crate) fn convert_to_snake_case(s: &str) -> String {
    let mut out = String::with_capacity(s.len() * 2);
    for (i, c) in s.chars().enumerate() {
        if i == 0 {
            out.push_str(&c.to_lowercase().collect::<String>());
            continue;
        }
        if c.is_uppercase() {
            out.push('_');
            out.push_str(&c.to_lowercase().collect::<String>());
        } else {
            out.push(c);
        }
    }
    out
}

/// create_out_directory creates the output directory for generated code, if it
/// does not already exist.
pub(crate) fn create_out_directory(module: &Path) {
    let out_dir = std::env::var_os("OUT_DIR").unwrap();
    let dest_path = Path::new(&out_dir).join(module);
    std::fs::create_dir_all(dest_path).unwrap();
}
