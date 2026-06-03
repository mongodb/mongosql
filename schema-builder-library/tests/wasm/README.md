# WASM Tests

This module tests that the generated WASM version of the schema builder library
works as expected against a sample dataset used in other tests.

## Testing Overview

These tests use vitest as a runner with snapshots for verifying that the output
is as expected. Snapshots are stored in the `snaps/` folder, while tests are
defined in files matching the pattern `*.test.ts` in the `tests/` folder.

## Set Up

To run the tests manually, make sure that you have a local instance of MongoDB
running pre-loaded with the schema-builder-library testing dataset, as used in
the evergreen scripts.

You will also need to drop a compiled wasm pack of the `schema-builder-library`
in the `dist/` folder. You can generate one by running the commands below:

```sh
$ wasm-pack build --package schema-builder-library --release --no-default-features --features wasm
$ mv ../../pkg ./schema-builder-library/tests/wasm/dist
```

## Running tests

```sh
$ npm test
```
