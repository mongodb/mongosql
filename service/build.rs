use std::env;
use std::path::PathBuf;

use protoc_bin_vendored::protoc_bin_path;

// This generates code from the .proto files located in the service/proto directory
// using the protoc compiler.
fn main() -> Result<(), Box<dyn std::error::Error>> {
    // We bundle protoc during builds to ensure that the generated output
    // is always the same and consistent across all build targets
    let vendored_config = {
        let mut config = tonic_prost_build::Config::new();
        config.protoc_executable(protoc_bin_path().expect("valid vendored protoc"));

        config
    };

    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    let descriptor_path = out_dir.join("translator_descriptor.bin");
    tonic_prost_build::configure()
        .protoc_arg("--experimental_allow_proto3_optional")
        .build_client(true)
        .build_server(true)
        .file_descriptor_set_path(&descriptor_path)
        .out_dir(&out_dir)
        .compile_with_config(
            vendored_config,
            &["proto/translator/v1/translator.proto"],
            &["proto/"],
        )?;

    Ok(())
}
