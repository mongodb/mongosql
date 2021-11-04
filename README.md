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

There are three types of tests for the Rust code: unit tests, fuzz tests, and spec tests. Each of
these test types can be run in isolation. Fuzz tests should always exist in (sub)modules named
`fuzz_test`, so this common name is used as a filter for running the tests. The spec tests are
ignored by default via the `#[ignore]` directive, but can be run by specifying this directive
after `cargo test`.

The [Rust handbook](https://doc.rust-lang.org/cargo/commands/cargo-test.html) has full guidelines
on how to use `cargo test`. Below are suggested ways of running the different sets of tests.

### Unit testing 

`cargo test -- --skip fuzz_test` from the main directory

### Fuzz testing

`cargo test fuzz_test` from the main directory

### Spec testing

`cargo test -- --ignored` from the main directory

### All testing

`cargo test -- --include-ignored` from the main directory

## Go testing

There are two types of tests for the Go code: integration tests and spec tests.
The spec tests have been separated out since they are not part of the MongoSQL Go
API.

### Building for go testing

In order to run the integration tests, the Rust code must be built with a
special feature.

`cargo build --features "mongosql-c/test"` from the main directory. This compiles
in code necessary for the Go tests to pass.

### Integration testing

`$ cd go/mongosql`
`$ export GOPRIVATE=github.com/10gen/*`
`$ export LIBRARY_PATH=$(cd ../.. && pwd)/target/debug`
`$ export LD_LIBRARY_PATH=$(cd ../.. && pwd)/target/debug`
`$ go test`

Replace `debug` with `release` in paths above to test release builds

### Spec testing

`go test -tags spectests`

## Dependencies

All are managed by Go modules for Go, and Cargo for Rust
