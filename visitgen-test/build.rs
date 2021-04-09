use visitgen::gen_visitor;

fn main() {
    gen_visitor("src/module/submodule/ast.rs");

    println!("cargo:rerun-if-changed=build.rs");
}
