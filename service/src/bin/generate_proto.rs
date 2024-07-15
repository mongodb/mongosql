use std::env;
use std::path::PathBuf;

// This generates code from the .proto files located in the service/proto directory
// using the protoc compiler.
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let proto_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("proto");
    let src_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("src");

    let descriptor_path = src_dir.join("translator_descriptor.bin");
    let proto_path = proto_dir.join("translator.proto");

    tonic_build::configure()
        .protoc_arg("--experimental_allow_proto3_optional")
        .build_client(true)
        .build_server(true)
        .file_descriptor_set_path(&descriptor_path)
        .out_dir(&src_dir)
        .compile(
            &[proto_path.to_str().unwrap()],
            &[proto_dir.to_str().unwrap()],
        )?;

    println!("Generated descriptor set: {:?}", descriptor_path);
    println!("Generated Rust files in: {:?}", src_dir);

    Ok(())
}
