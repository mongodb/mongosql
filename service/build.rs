use std::env;
use std::path::PathBuf;

// This generates code from the .proto files located in the service/proto directory
// using the protoc compiler.
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    let descriptor_path = out_dir.join("translator_descriptor.bin");
    tonic_prost_build::configure()
        .protoc_arg("--experimental_allow_proto3_optional")
        .build_client(true)
        .build_server(true)
        .file_descriptor_set_path(&descriptor_path)
        .out_dir(&out_dir)
        .compile_protos(&["proto/translator/v1/translator.proto"], &["proto/"])?;

    Ok(())
}
