use visitgen::gen_visitor;

fn main() {
    lalrpop::process_root().unwrap();

    gen_visitor("src/ast/definitions.rs");
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=src/parser/mongosql.lalrpop");
}
