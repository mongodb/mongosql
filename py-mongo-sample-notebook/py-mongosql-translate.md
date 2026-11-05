# py-mongosql-translate

Python bindings for translating [MongoSQL](https://www.mongodb.com/docs/manual/reference/sql-aggregation-language/)
queries into MongoDB aggregation pipelines. The module name is `py_mongosqltranslate`.

Translation is pure Rust (`mongosql::translate_sql`): **no MongoDB connection and no
Python runtime dependencies are required**. Give the translator a SQL query and an
optional schema catalog, and it returns a `Translation` you can run against MongoDB.

## Installation

Build the extension into your virtualenv with [maturin](https://www.maturin.rs/):

```sh
cd py-mongosqltranslate
python -m venv .venv && source .venv/bin/activate
pip install maturin
maturin develop
```

Then import it:

```python
from py_mongosqltranslate import MongoSqlTranslator, Translation
```

## `MongoSqlTranslator`

```python
MongoSqlTranslator()
```

A stateless translator. Construct it once and reuse it; there is no configuration and
no connection. All three methods live on the instance — there are **no module-level
`translate_query*` functions**.

Every method always translates with namespaces included in the result set and relaxed
schema checking (the `IncludeNamespaces` / `Relaxed` options, matching `mongosql-cli`).
These options are fixed and not configurable.

### `translate_query(current_db, sql, catalog=None) -> Translation`

Translates `sql` into a MongoDB aggregation pipeline.

- `current_db` (`str`) — the database that unqualified collection names resolve against.
- `sql` (`str`) — the MongoSQL query text.
- `catalog` (`dict | None`) — maps `{db: {collection: json_schema}}` for the collections
  the query reads. Optional; defaults to `None`.

```python
from py_mongosqltranslate import MongoSqlTranslator

catalog = {
    "sample_mflix": {
        "movies": {
            "bsonType": "object",
            "properties": {
                "title": {"bsonType": "string"},
                "year": {"bsonType": "int"},
            },
            "required": ["title"],
            "additionalProperties": False,
        }
    }
}

translator = MongoSqlTranslator()
t = translator.translate_query(
    "sample_mflix",
    "SELECT title, year FROM movies WHERE year > 2010 ORDER BY year DESC LIMIT 2",
    catalog,
)

print(t.target_db)          # sample_mflix
print(t.target_collection)  # movies
print(t.pipeline)           # [{'$match': {'year': {'$gt': 2010}}}, ...]
print(t.select_order)       # [['', 'title'], ['', 'year']]
```

### `translate_query_verbose(current_db, sql, catalog=None) -> Translation`

Currently identical to `translate_query`; it is the same method, reserved for future
extra knobs. Use whichever reads better.

```python
t = translator.translate_query_verbose("sample_mflix", "SELECT title FROM movies", catalog)
```

### `translate_query_simple(current_db, sql) -> Translation`

The two-argument form: equivalent to `translate_query(current_db, sql, None)`. Every
collection the query reads is assigned an empty schema, so the translation is untyped.

```python
import warnings

with warnings.catch_warnings():
    warnings.simplefilter("ignore")  # see "Warnings" below
    t = translator.translate_query_simple("sample_mflix", "SELECT title FROM movies")
```

### The `catalog` argument

The catalog is `{db: {collection: json_schema}}` — the same shape as the
`mongosqltranslate` `schemaCatalog` option. Each schema may be either a **bare JSON
schema** or **wrapped as `{"$jsonSchema": {...}}`** (the format `SchemaBuilder` returns);
both produce identical pipelines.

```python
bare    = {"sample_mflix": {"movies": {"bsonType": "object",
                                       "properties": {"title": {"bsonType": "string"}}}}}
wrapped = {"sample_mflix": {"movies": {"$jsonSchema": {
               "bsonType": "object", "properties": {"title": {"bsonType": "string"}}}}}}

# Same result either way.
translator.translate_query("sample_mflix", "SELECT title FROM movies", bare).pipeline == \
    translator.translate_query("sample_mflix", "SELECT title FROM movies", wrapped).pipeline
# True
```

**The catalog is optional, and partial catalogs are allowed.** Any collection the query
reads that the catalog does not cover is assigned an empty schema rather than raising,
mirroring `mongosql-cli`. This has two consequences:

- `CollectionNotFound` is **not reachable** from Python — a mistyped collection name
  translates instead of failing.
- Where no schema was supplied, the result set schema is untyped: its `properties`
  values are `{}`.

```python
with warnings.catch_warnings():
    warnings.simplefilter("ignore")
    t = translator.translate_query("sample_mflix", "SELECT title FROM movies")
print(t.result_set_schema["properties"][""]["properties"])
# {'title': {}}   <- untyped: no schema was provided
```

## `Translation`

All three methods return a read-only `Translation` (`frozen`; assigning any attribute
raises `AttributeError`).

| Attribute | Type | Meaning |
| --- | --- | --- |
| `target_db` | `str` | Database the pipeline must run against. |
| `target_collection` | `str \| None` | Collection to run against; `None` for a database-level aggregation (e.g. `SELECT 1` or array datasources). |
| `pipeline` | `list[dict]` | The aggregation pipeline as a list of stage dicts. |
| `result_set_schema` | `dict` | JSON schema describing the documents the pipeline produces. |
| `select_order` | `list[list[str]]` | Field paths of the result set, in SELECT-list order. |

Because namespaces are included, each `select_order` entry is `[namespace, field]`:

- An **explicit select list** uses the empty namespace:
  `SELECT title, year FROM movies` → `[['', 'title'], ['', 'year']]`.
- **`SELECT *`** uses the collection name, listing every field:
  `SELECT * FROM movies` → `[['movies', '_id'], ['movies', 'awards'], ['movies', 'cast'], ...]`.

`result_set_schema` is keyed by the same namespace.

`repr()` gives a compact summary:

```python
>>> t
Translation(target_db='sample_mflix', target_collection='movies', pipeline=<6 stages>, select_order=[['', 'title'], ['', 'year']])
```

### Running the pipeline

`target_collection` tells you where to run the pipeline. When it is `None`, the pipeline
is database-level (e.g. it starts with `$documents`).

```python
# With PyMongo:
db = client[t.target_db]
results = list(db[t.target_collection].aggregate(t.pipeline))
```

## Warnings

Each call that fills in an empty schema emits a `UserWarning` naming the affected
namespaces:

```
No schema was found for the following collections: ["sample_mflix.movies"].
These collections will be assigned empty schemas.
```

No warning is emitted when the catalog covers the query, or when the query reads no
collections (`SELECT 1`, `SELECT * FROM [{'a': 1}] arr`). Suppress it with the
`warnings` module when the fallback is intentional.

## Errors

Failures raise `ValueError` (never a panic/abort):

- **Malformed catalog** — non-string keys, non-dict values, or schemas that are not valid
  JSON schemas. The message names the offending `db.collection`.
- **Translation failure** — the query fails to parse or otherwise fails to translate.

```python
import pytest

with pytest.raises(ValueError):
    translator.translate_query("sample_mflix", "SELEC title FROM movies", catalog)

with pytest.raises(ValueError):
    translator.translate_query(
        "sample_mflix", "SELECT title FROM movies",
        {"sample_mflix": {"movies": {"bsonType": "notAType"}}},
    )
```

A SQL syntax error is reported from the namespace scan, before translation runs.

## Notes

- **Stateless and reusable.** Two `MongoSqlTranslator()` instances are independent but
  produce identical pipelines for identical input.
- **Conversion degradations** (inherited from the shared BSON bridge): ObjectId → hex
  string, Decimal128 / exotic BSON → stringified.
- **Cross-database queries** set `target_db`/`target_collection` to the resolved target,
  e.g. `SELECT x FROM other.bar` yields `target_db == "other"`.
