mod analysis;
mod gen_visitor_trait;
mod gen_walk_implementations;
mod util;

#[cfg(test)]
mod test;

/// gen_visitor is the entry point for the library: it generates
/// both the Visitor trait and the walk implementations for all types
/// in the argument file_path. The walk and visitor modules are placed
/// in the same module hierarchy position as the argument file. E.g.,
/// if you pass 'src/module/submodule/ast.rs' the generated module paths
/// will be module::submodule::{visitor,walk}. The 'src' is automatically
/// ignored since this is designed to be called from build.rs.
pub fn gen_visitor(file_path: &str) {
    use std::path::{Path, PathBuf};

    let types = analysis::collect_types(file_path);
    let mut module_path = file_path.split('/');
    // remove src
    let _ = module_path.next();
    let mut module_path = module_path.collect::<Vec<_>>();
    let l = module_path.len();
    let target_module_name = Path::new(module_path[l - 1])
        .file_stem()
        .expect("Failed to find file stem")
        .to_str()
        .expect("Failed to convert OSStr to str");
    module_path[l - 1] = target_module_name;
    let mut dest_path = PathBuf::new();
    for dir in module_path[..l - 1].iter() {
        dest_path.push(dir);
    }
    util::create_out_directory(&dest_path);
    gen_visitor_trait::gen_visitor_trait(&module_path, &types);
    gen_walk_implementations::gen_walk_implementations(&module_path, &types);
    println!("cargo:rerun-if-changed={file_path}");
}
