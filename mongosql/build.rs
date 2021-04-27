use visitgen::gen_visitor;

fn main() {
    lalrpop::process_root().unwrap();

    gen_visitor("src/parser/ast.rs");
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=parser/mongosql.lalrpop");
}
