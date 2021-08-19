# MongoDB MongoSQL Compiler

This document contains instructions for building and testing the MongoSQL compiler

## Build Requirements

### Minimum Rust version
1.52.0

### Minimum go version
1.15

## Building

`cargo build` from the main directory

For release mode, use `cargo build --release`, this will remove debugging support and will optimize
the code.

## Rust testing

`cargo test` from the main directory

## Go testing

To test the go interface, the rust code must be built with a special feature

### Building for go testing

`cargo build --features "mongosql-c/test"` from the main directory. This compiles in code necessary
for the go tests to pass

### Running go testing

`$ cd go/mongosql`
`$ export GOPRIVATE=github.com/10gen/*`
`$ export LIBRARY_PATH=$(cd ../.. && pwd)/target/debug`
`$ export LD_LIBRARY_PATH=$(cd ../.. && pwd)/target/debug`
`$ go test`

Replace `debug` with `release` in paths above to test release builds

## Dependencies

All are managed by go modules for go, and cargo for rust
