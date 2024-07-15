# MongoSQL Translation Service

## Updating Protobuf Files

When changes are made to `service/proto/translator.proto`, the associated files will need to be regenerated. 
This process will update the following files:

1. `service/src/translator.rs`
2. `service/src/translator_descriptor.bin`

## Code Generation Process

### Prerequisites
`Protobuf 3` is used, and code is generated with `protoc`, the protobuf compiler.  `protoc` should be installed with
an OS specific package manager such as brew, apt, or scoop. See: https://grpc.io/docs/protoc-installation/ for 
more info.

### Code generation
1. Make necessary changes to `service/proto/translator.proto`.
2. Run the `generate_proto` command:
```
cargo run --bin generate_proto
```
3. Verify that both `service/src/translator.rs` and `service/src/translator_descriptor.bin` have been updated.
4. Commit these changes along with the `translator.proto` changes.