use crate::{analysis::EnumOrStruct, util::convert_to_snake_case};

/// gen_visitor_trait generates the Visitor trait for a rust file.
pub fn gen_visitor_trait(module_path: &[&str], types: &[EnumOrStruct]) {
    use std::env;
    use std::fs;
    use std::path::PathBuf;
    let out_dir = env::var_os("OUT_DIR").unwrap();
    let mut dest_path = PathBuf::new();
    dest_path.push(&out_dir);
    for dir in module_path[..module_path.len() - 1].iter() {
        dest_path.push(dir);
    }
    dest_path.push("visitor.rs");

    let target_module_path = module_path.join("::");
    let target_module_name = module_path[module_path.len() - 1];
    let mut out = format!(
        "pub mod visitor {{
     use crate::{target_module_path};

     pub trait Visitor : Sized {{\n"
    );
    for t in types.iter() {
        let type_name = t.get_name();
        let visit_name = format!("visit_{}", convert_to_snake_case(&type_name));
        let full_type_name = format!("{target_module_name}::{type_name}");
        out.push_str("        fn ");
        out.push_str(&visit_name);
        out.push_str("(&mut self, node: ");
        out.push_str(&full_type_name);
        out.push_str(") -> ");
        out.push_str(&full_type_name);
        out.push_str(" { node.walk(self) }\n");
    }
    out.push_str("    }\n}");
    fs::write(&dest_path, &out).expect("failed to write visitor trait");
}
