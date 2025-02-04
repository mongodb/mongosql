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

There are several types of tests for the Rust code: unit tests, fuzz tests, index usage tests, e2e
tests, and spec tests. Each of these test types can be run in isolation. Fuzz tests should always
exist in (sub)modules named `fuzz_test`, so this common name is used as a filter for running the
tests. The index usage, e2e, and spec tests are ignored by default via the `#[ignore]` directive,
but can be run by specifying this directive after `cargo test`.

The [Rust handbook](https://doc.rust-lang.org/cargo/commands/cargo-test.html) has full guidelines
on how to use `cargo test`. Below are suggested ways of running the different sets of tests.

### Unit testing

All unit tests in the repository.

`cargo test -- --skip fuzz_test` from the main directory

### Fuzz testing

Fuzz tests for pretty-printing.

`cargo test fuzz_test` from the main directory

### Index Usage testing

Index-usage assertion tests. These tests specify queries and expected index utilization.
Requires a running mongod.

`cargo test --features=index,generated --package=e2e-tests -- --test-threads=1` from the main directory

### e2e testing

End-to-end query tests that do not exist in the spec. These tests specify queries and expected
result sets and execute against an actual database. Requires a running mongod.

`cargo test --features=e2e,generated --package=e2e-tests -- --test-threads=1` from the main directory

### errors testing

error tests are e2e tests for our errors. These tests specify a SQL query and have an expected
error that they should cause. Requires a running mongod.

`cargo test --features=error,generated --package=e2e-tests -- --test-threads=1` from the main directory

### Spec query testing

The query spec tests that specify language behavior. Requires a running mongod.

`cargo test --features=query,generated --package=e2e-tests -- --test-threads=1` from the main directory

### Evergreen testing

Our suite of integration tests is automatically created by the `build.rs` script in
`evergreen/create-tasks/build.rs`. If you modify this file to add a new version of
MongoDB or to test against different topologies and do not have your editor/IDE
configured to automatically run a check command that builds on save, ensure you
run `cargo build --package create-tasks`.

Ensure you commit the changes to `evergreen/suite-tasks.yml`.

### Spec testing

The syntactic rewriter and type-constraint spec tests.

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

### Environment Variables
If testing on a Mac using ARM architecture, such as the M1 chip, the following Go environment variables may be required.
```
export GOOS=darwin
export GOARCH=arm64
export CGO_ENABLED=1
```

### Integration testing

```
$ cd go/mongosql
$ export GOPRIVATE=github.com/10gen/*
$ export LIBRARY_PATH=$(cd ../.. && pwd)/target/debug
$ export LD_LIBRARY_PATH=$(cd ../.. && pwd)/target/debug
$ go test
```

Replace `debug` with `release` in paths above to test release builds

### Spec testing

`go test -tags spectests`

You will need a running `mongod` on the default port (27017) in order
to run result tests.

#### Specifying tests

To target specific tests, specify the appropriate file and test name:

`go test -v -tags spectests -run TestSpecResultSets/filename.yml/Test_name_snake_case`

For example, the following test is in `mongosql-rs/tests/spec_tests/query_tests/identifier.yml`:

```yml
  - description: Unaliased use of field reference expression with $ and .
    query: "SELECT * FROM bar WHERE `$a.b` = 4"
    current_db: foo
    result:
      - {'bar': {'_id': 4, '$a.b': 4}}
```

* To run this test only: `go test -v -tags spectests -run TestSpecResultSets/identifier.yml/Unaliased_use_of_field_reference_expression_with_$_and_.`
* To run this file only: `go test -v -tags spectests -run TestSpecResultSets/identifier.yml`

## Debugging visitgen

Since our visitors are created by a proc_macro that utilizes many macros, it is difficult to figure out
what the generated walk and visit rust files look like just by reading the code. To see what the generated
rust files look like, run:

`cargo build --features debug-visitor-output`

This command will generate the visit and walk rust files and put them in subdirectories of the `target`
directory called `<some_enum_or_struct>_visit` and `<some_enum_or_struct>_walk` respectively. The
`<some_enum_or_struct>` value will be one of the structs or enums that is input to `visitgen::generate_visitors! {...}`
and will be the same for corresponding visit and walk files.

## Dependencies

All are managed by Go modules for Go, and Cargo for Rust

## License

This project is licensed under the [Apache License 2.0](/LICENSE).
This product includes software developed by the OpenSSL Project for use in the OpenSSL Toolkit (<http://www.openssl.org/>).
