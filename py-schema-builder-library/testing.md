# Testing `py-schema-builder-library`

How to build the extension module and validate the Python bindings locally.

## Prerequisites

- Rust toolchain (the workspace edition is 2024, so a recent stable is required)
- [maturin](https://www.maturin.rs/): `pip install maturin` (or `uv pip install maturin`)
- [PyMongo](https://pymongo.readthedocs.io/) (runtime dependency — the Rust
  bindings call `bson.encode`/`bson.decode`): `pip install pymongo`.
  Declared in `pyproject.toml`, so wheels installed via `pip` pull it in
  automatically; `maturin develop` does not install it for you.
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

Run the checked-in script (see `examples/smoke_test.py`):

```sh
.venv/bin/python examples/smoke_test.py
```

Or inline:

```sh
python - <<'EOF'
import pymongo
from py_schema_builder_library import sum_as_string, SchemaBuilder
import bson

print(sum_as_string(2, 40))  # "42"

# Document-level methods work without a client...
b = SchemaBuilder(None)
doc = {"a": 1, "b": "x"}
print(b.build_schema_for_document(doc))

# ...collection-level methods need a PyMongo client and a running mongod.
b = SchemaBuilder(pymongo.MongoClient())
print(b.build_schema_for_collection("smoke_test_db", "smoke_test_coll"))
print(b.sample_schema_for_collection("smoke_test_db", "smoke_test_coll", 100))
EOF
```

Expected: `42` printed, then schema dicts (Python `dict`) for the document
path, and — with a running MongoDB — dicts for the collection paths.
Requires `pip install pymongo` in the active venv. Malformed bytes passed to
`build_schema_for_document_bytes`, a missing collection, or calling a
collection-level method on a client-less builder all raise Python
`ValueError`.

## Things to validate as the bindings evolve

- **Constructor:** `SchemaBuilder(pymongo_client)` for collection-level
  operations, or `SchemaBuilder(None)` for document-only use. Collection
  methods on a client-less builder raise `ValueError`.
- **Signature acceptance:** `build_schema_for_document` must accept a plain
  `dict` positionally; calling it with no arguments should raise
  `TypeError` (missing `document`), not a Rust panic.
- **Error mapping:** failures inside Rust should surface as Python
  `ValueError`s (we use `PyValueError`), never as panics/aborts.
- **Rebuild loop:** after editing `src/lib.rs`, re-run
  `maturin develop` and re-import — Python caches the loaded module per
  process, so start a fresh interpreter between builds.

## Troubleshooting

- `ImportError: ... undefined symbol` — the wheel was built for a different
  Python; make sure the same venv is active for `maturin develop` and `python`.
- `error: no matching package named py-schema-builder-library` — run from the
  `py-schema-builder-library/` directory (or pass `-m py-schema-builder-library`).
- For a Rust-side check without Python: `cargo build -p py-schema-builder-library`.
