# Testing `py-schema-builder-library`

How to build the extension module and validate the Python bindings locally.

## Prerequisites

- Rust toolchain (the workspace edition is 2024, so a recent stable is required)
- [maturin](https://www.maturin.rs/): `pip install maturin` (or `uv pip install maturin`)
- **No Python MongoDB dependencies.** All Mongo I/O goes through the embedded
  Rust driver (`SynchronousMongoDbDataService` over `mongodb::sync::Client`),
  and all BSON ↔ Python conversion is pure Rust (`src/py_bson.rs`).
- A virtualenv for the target Python interpreter:

  ```sh
  python -m venv .venv
  source .venv/bin/activate
  ```

## Build and install into the venv

From this directory (`py-schema-builder-library/`):

```sh
maturin develop            # debug build, fast iteration
maturin develop --release  # optimized build for benchmarks
```

`maturin develop` builds the cdylib with Cargo and installs it into the
**active virtualenv**. Re-run it after every `cargo build`-worthy change.

If you don't use a virtualenv, use `maturin build` and `pip install` the
produced wheel manually.

## Smoke test

Run the checked-in script (see `examples/smoke_test.py`; seeding uses
`mongosh`, which is optional — collection tests are skipped without it):

```sh
.venv/bin/python examples/smoke_test.py
```

Or inline:

```sh
python - <<'EOF'
from py_schema_builder_library import SchemaBuilder, MongoConnectionProperties

print(SchemaBuilder().build_schema_for_document({"a": 1, "b": "x"}))

# Collection-level operations connect with the embedded Rust driver.
b = SchemaBuilder(MongoConnectionProperties(uri="mongodb://localhost:27017"))
print(b.build_schema_for_collection("smoke_test_db", "smoke_test_coll"))
print(b.sample_schema_for_collection("smoke_test_db", "smoke_test_coll", 100))
EOF
```

Expected: schema dicts (Python `dict`) for every path. Requires a running
`mongod` (default `mongodb://localhost:27017`) for the collection paths.

## Things to validate as the bindings evolve

- **Constructor:** `SchemaBuilder(MongoConnectionProperties(uri=..., username=...,
  password=...))` for collection-level operations, or `SchemaBuilder()` for
  document-only use. Collection methods on a connection-less builder raise
  `ValueError`. A malformed URI fails fast at construction with `ValueError`;
  connection failures surface as `ValueError` at the first query (the driver
  connects lazily).
- **Signature acceptance:** `build_schema_for_document` must accept a plain
  `dict` positionally; calling it with no arguments should raise
  `TypeError` (missing `document`), not a Rust panic.
- **Multi-document path:** `build_schema_for_documents` must accept a plain
  Python `list` of dicts. Heterogeneous types across documents should widen
  the schema (`{"b": "x"}, {"b": 3}` → `b` becomes
  `anyOf: [long, string]`); a field present in only some docs appears in
  `properties` but not `required`.
- **BSON type coverage:** a collection containing every BSON type
  (string, int, long, double, decimal, bool, date, timestamp, objectId,
  binData, null, regex, javascript, array, nested doc) must derive a schema
  with the correct `bsonType` for each (verified against `bson_types_db.all_types`).
- **Conversion degradations** (pure-Rust conversion, no PyMongo types):
  ObjectId → hex string, Decimal128/exotic BSON → stringified. Python
  `datetime` round-trips through epoch millis; unsupported Python types
  (e.g. `set`) raise `ValueError`.
- **Error mapping:** failures inside Rust should surface as Python
  `ValueError`s (we use `PyValueError`), never as panics/aborts.
- **Rebuild loop:** after editing `src/lib.rs`, re-run `maturin develop` and
  re-import — Python caches the loaded module per process, so start a fresh
  interpreter between builds.

## Troubleshooting

- `ImportError: ... undefined symbol` — the wheel was built for a different
  Python; make sure the same venv is active for `maturin develop` and `python`.
- `error: no matching package named py-schema-builder-library` — run from the
  `py-schema-builder-library/` directory (or pass `-m py-schema-builder-library`).
- `panicked: Cannot start a runtime from within a runtime` — an async
  (tokio `block_on`) harness snuck in; schema futures here must be driven with
  `futures::executor::block_on` because the `mongodb::sync` driver owns its own
  internal runtime.
- For a Rust-side check without Python: `cargo build -p py-schema-builder-library`.
- Rust-side unit tests for the converters: `cargo test -p py-schema-builder-library`.
