fn main() {
    lalrpop::process_root().unwrap();
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=parser/mongosql.lalrpop");
}
