# Sampler in Rust!!

The goals of this project are:
- Partition a collection based on _id, much like mongosync
- On each partition, run the following algorithm
  1. Sample 100 documents
  2. Calculate their schema
  3. Find documents that _do not match_ the schema
  4. Calculate the schema for any returned documents
  5. Merge the schemas
  6. Repeat 3-5 until no new documents are returned
- Merge each partition schema into a single schema

## Integration testing for schema-builder-library

The `schema-builder-library` integration tests cover library method correctness against a live
database. These tests require a running enterprise mongod server (version at least 6.0). They also
require that data be loaded in before running the tests. The data is
stored [here](https://mongosql-noexpire.s3.us-east-2.amazonaws.com/schema_manager/schema-builder-library.tar.gz).
After decompressing the data, to load it into the database,
use the [sql-engines-common-test-infra](https://github.com/mongodb/sql-engines-common-test-infra)
`data-loader` tool. See `cargo run --bin data-loader -- --help` in that repo for more details.

With the data loaded, the `schema-builder-library` integration tests can be run from the root of
the mongosql repository via

```
cargo test --features=integration --package=schema-builder-library --lib internal_integration_tests -- --test-threads=1
```
