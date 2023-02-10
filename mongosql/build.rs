use visitgen::gen_visitor;

fn main() {
    lalrpop::process_root().unwrap();
    gen_visitor("src/ast/definitions.rs");
    gen_visitor("src/mir/definitions.rs");
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=parser/mongosql.lalrpop");
}
