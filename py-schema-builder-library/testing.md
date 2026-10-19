# Testing `py-schema-builder-library`

How to build the extension module and validate the Python bindings locally.

## Prerequisites

- Rust toolchain (the workspace edition is 2024, so a recent stable is required)
- [maturin](https://www.maturin.rs/): `pip install maturin` (or `uv pip install maturin`)
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
from py_schema_builder_library import sum_as_string, SchemaBuilder
import bson

print(sum_as_string(2, 40))  # "42"

b = SchemaBuilder()
doc = {"a": 1, "b": "x"}

# Raw-bytes path: encode the document in Python with pymongo's bson...
encoded = bson.encode(doc)
print(b.build_schema_for_document_bytes(encoded))

# ...or let the Rust side do the encoding via bson.encode
print(b.build_schema_for_document(doc))
EOF
```

Expected: `42` printed, then two `Generated schema for: ...` lines (the same
schema from both paths). Requires `pip install pymongo` in the active venv
for the `bson.encode` calls. Malformed bytes passed to
`build_schema_for_document_bytes` raise Python `ValueError`.

## Things to validate as the bindings evolve

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
