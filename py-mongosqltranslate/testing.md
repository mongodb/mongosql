# Testing `py-mongosqltranslate`

How to build the extension module and validate the Python bindings locally.

## Prerequisites

- Rust toolchain (the workspace edition is 2024, so a recent stable is required)
- [maturin](https://www.maturin.rs/): `pip install maturin` (or `uv pip install maturin`)
- [pytest](https://docs.pytest.org/) for the Python tests (`pip install -e '.[test]'` or
  `pip install pytest`)
- **No Python runtime dependencies and no MongoDB.** Translation is pure Rust
  (`mongosql::translate_sql`), and Python ↔ BSON conversion is pure Rust
  (`src/py_bson.rs`, copied from `py-schema-builder-library`).
- A virtualenv for the target Python interpreter:

  ```sh
  python -m venv .venv
  source .venv/bin/activate
  pip install maturin pytest
  ```

## Build and install into the venv

From this directory (`py-mongosqltranslate/`):

```sh
maturin develop            # debug build, fast iteration
maturin develop --release  # optimized build for benchmarks
```

`maturin develop` builds the cdylib with Cargo and installs it into the
**active virtualenv**. Re-run it after every change to `src/`.

## Python tests

```sh
.venv/bin/pytest tests/ -v
```

`tests/test_translate.py` covers translation only (no MongoDB, no e2e):
result shape, `translate_query` ≡ `translate_query_verbose` ≡
`translate_query_simple`, the optional catalog and its empty-schema fallback
(including the `UserWarning`), `$jsonSchema` unwrapping, cross-database queries,
and error mapping.

## Smoke test

```sh
.venv/bin/python examples/smoke_test.py
```

Prints translations (pipeline, result set schema, select order) for a few
queries and exercises the error paths. If `py_schema_builder_library` is
installed in the same venv and `mongosh` can seed a running `mongod`
(default `mongodb://localhost:27017`), it also builds a catalog from a real
collection with `SchemaBuilder` and translates against it; otherwise that
section is skipped.

Or inline:

```sh
python - <<'PY'
from py_mongosqltranslate import MongoSqlTranslator

translator = MongoSqlTranslator()
catalog = {"test": {"foo": {"bsonType": "object",
                            "properties": {"a": {"bsonType": "int"}}}}}
t = translator.translate_query("test", "SELECT a FROM foo", catalog)
print(t)
print(t.pipeline)

# No schema at all: `foo` gets an empty schema and a UserWarning is emitted,
# so the result set schema is untyped.
print(translator.translate_query_simple("test", "SELECT a FROM foo").pipeline)
PY
```

## Rust tests

```sh
cargo test -p py-mongosqltranslate
```

Covers the BSON converters (`py_bson`), catalog conversion and the empty-schema
fill logic, and translation through `MongoSqlTranslator` with an embedded
interpreter (`auto-initialize`).

Lint only this crate (workspace dependencies have pre-existing pedantic warnings):

```sh
cargo clippy -p py-mongosqltranslate --all-targets --no-deps -- -W clippy::pedantic -W clippy::all --deny warnings
```

## Things to validate as the bindings evolve

- **Signatures:** `MongoSqlTranslator()` is a stateless class (constructed with no
  arguments) exposing three instance methods: `translate_query(current_db, sql, catalog=None)`,
  `translate_query_verbose(current_db, sql, catalog=None)` (the same, reserved for future extra
  knobs), and `translate_query_simple(current_db, sql)`. They accept positional and keyword
  arguments; missing `current_db` / `sql` raises `TypeError`. There are no module-level
  `translate_query*` functions.
- **Return type:** all three return a read-only `Translation` with `target_db`
  (str), `target_collection` (str or `None`), `pipeline` (list of dicts),
  `result_set_schema` (dict), and `select_order` (list of lists of str).
  Assigning an attribute raises `AttributeError`.
- **Fixed options:** translation always uses `IncludeNamespaces` and
  `Relaxed` schema checking (matching `mongosql-cli`'s defaults), so
  `select_order` entries are `[namespace, field]` and `result_set_schema` is
  keyed by namespace. For an explicit select list the namespace is the empty
  string (`SELECT a, b FROM foo` → `[['', 'a'], ['', 'b']]`); `SELECT *` uses
  the collection name (`[['foo', 'a'], ['foo', 'b']]`).
- **CLI parity:** because the options now match, `cargo run -p mongosql-cli --
  --db test "<query>"` should print the same pipeline and result-set schema as
  `translate_query` for the same catalog. Useful as an oracle when changing
  translation behavior.
- **Catalog shape:** `{db: {collection: json_schema}}`. Schemas may be bare
  or wrapped as `{"$jsonSchema": {...}}` (what `SchemaBuilder` returns); both
  must produce identical pipelines.
- **Missing schemas are filled, not rejected:** any collection the query reads
  that the catalog does not cover is assigned an empty schema (`Schema::Any`),
  mirroring `mongosql-cli`. This applies whether the catalog is absent
  (`None`), empty (`{}`), or merely partial, so `translate_query_simple(db, sql)`,
  `translate_query(db, sql, None)`, and `translate_query(db, sql, {})` on the
  same translator are all
  equivalent. Consequences: `CollectionNotFound` is **not reachable** from
  Python (a mistyped collection name translates), and the result set schema is
  untyped where no schema was supplied (`properties` values are `{}`).
- **Warnings:** each call that fills in a schema emits a `UserWarning` naming
  the affected namespaces, e.g.
  `No schema was found for the following collections: ["test.foo"]. These
  collections will be assigned empty schemas.` No warning is emitted when the
  catalog covers the query, or when the query reads no collections
  (`SELECT 1`, `FROM [{'a': 1}] arr`). Suppress with the `warnings` module if
  the fallback is intentional.
- **Error mapping:** malformed catalogs (non-str keys, non-dict values,
  unknown `bsonType`) and translation failures (syntax errors, unsupported
  queries) surface as `ValueError`, never as panics/aborts. Catalog
  errors name the offending `db.collection`. A syntax error is reported from
  the namespace scan, before translation runs.
- **Conversion degradations** (inherited from `py_bson`): ObjectId → hex
  string, Decimal128 / exotic BSON → stringified, `datetime` via epoch millis.
- **Rebuild loop:** Python caches the loaded module per process, so start a
  fresh interpreter (or pytest run) after each `maturin develop`.

## Troubleshooting

- `mismatched types ... bson::Bson` at build time: this crate must use the
  workspace `bson` (v2) that `mongosql` uses, not a different major version.
- `ImportError: ... undefined symbol`: the wheel was built for a different
  Python. Make sure the same venv is active for `maturin develop` and `python`.
- `ModuleNotFoundError: py_mongosqltranslate`: run `maturin develop` with
  the venv active (or use `.venv/bin/maturin develop`).
- `error: no matching package named py-mongosqltranslate`: run from the
  `py-mongosqltranslate/` directory (or pass `-m py-mongosqltranslate/Cargo.toml`).
