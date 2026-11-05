# py-mongo-schema-builder

Python bindings for deriving MongoDB JSON schemas from documents and collections. The
module name is `py_schema_builder_library`.

Schema derivation, MongoDB I/O, and BSON ↔ Python conversion are all pure Rust — there
is **no PyMongo dependency**. Collection-level methods use the embedded Rust driver
(`SynchronousMongoDbDataService` over `mongodb::sync::Client`); document-level methods
need no connection at all.

Every method returns a `dict` of the form `{"$jsonSchema": {...}}` — the same wrapper
`mongosqltranslate` unwraps, so the output drops straight into a translation catalog.

## Installation

```sh
cd py-schema-builder-library
python -m venv .venv && source .venv/bin/activate
pip install maturin
maturin develop
```

Then import it:

```python
from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder
```

## `MongoConnectionProperties`

```python
MongoConnectionProperties(uri=None, username=None, password=None)
```

Connection settings for the embedded Rust driver client. All fields are optional and
read/write.

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `uri` | `str \| None` | `"mongodb://localhost:27017"` | Standard MongoDB connection string. |
| `username` | `str \| None` | `None` | Credential username. |
| `password` | `str \| None` | `None` | Credential password. |

TLS, read preference, replica set, and other advanced options ride in the `uri`
connection string. When `username` or `password` is set, a credential is built from them.

```python
props = MongoConnectionProperties(uri="mongodb://localhost:27017")
props.username = "alice"
props.password = "secret"
```

## `SchemaBuilder`

```python
SchemaBuilder(connection_properties=None)
```

- Pass `MongoConnectionProperties(...)` for collection-level operations.
- Pass nothing (`SchemaBuilder()`) for document-only use. Calling a collection-level
  method on a connection-less builder raises `ValueError`.

```python
from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder

builder = SchemaBuilder()  # document-only
connected = SchemaBuilder(MongoConnectionProperties(uri="mongodb://localhost:27017"))
```

The connection is created lazily: a malformed URI fails fast at construction with
`ValueError`, while a connection failure surfaces as `ValueError` at the first query.

### `build_schema_for_document(document) -> dict`

Builds a schema from a single Python `dict`. No connection required.

```python
builder = SchemaBuilder()
builder.build_schema_for_document({"a": 1, "b": "x"})
# {'$jsonSchema': {'bsonType': 'object',
#                  'properties': {'a': {'bsonType': 'long'}, 'b': {'bsonType': 'string'}},
#                  'required': ['a', 'b'],
#                  'additionalProperties': False}}
```

Python `int` maps to BSON `long` (`int64`). Nested dicts, lists, `bytes`, and
`datetime.datetime` are converted recursively.

### `build_schema_for_documents(documents) -> dict`

Builds a schema from a **list** of Python dicts, widening field types across documents.
No connection required.

```python
builder.build_schema_for_documents([{"a": 1, "b": "x"}, {"a": 2, "b": 3, "c": True}])
# {'$jsonSchema': {'bsonType': 'object',
#                  'properties': {'a': {'bsonType': 'long'},
#                                 'b': {'anyOf': [{'bsonType': 'long'}, {'bsonType': 'string'}]},
#                                 'c': {'bsonType': 'bool'}},
#                  'required': ['a', 'b'],
#                  'additionalProperties': False}}
```

- Heterogeneous types across documents **widen** the schema: `b` is a string in one doc
  and a number in another, so it becomes `anyOf: [long, string]`.
- A field present in only some documents (here `c`) appears in `properties` but **not**
  in `required`.

### `build_schema_for_document_bytes(document_bytes) -> dict`

Builds a schema from an already BSON-encoded document — the raw bytes that `bson`
serialization produces. No connection required.

```python
import bson

raw = bson.encode({"a": 1, "b": "x", "nested": {"c": True}})
builder.build_schema_for_document_bytes(raw)
# {'$jsonSchema': {'bsonType': 'object',
#                  'properties': {'a': {'bsonType': 'int'},
#                                 'b': {'bsonType': 'string'},
#                                 'nested': {'bsonType': 'object',
#                                            'properties': {'c': {'bsonType': 'bool'}},
#                                            'required': ['c'],
#                                            'additionalProperties': False}},
#                  'required': ['a', 'b', 'nested'],
#                  'additionalProperties': False}}
```

Malformed bytes raise `ValueError`.

### `build_schema_for_collection(db_name, collection_name) -> dict`

Builds a schema from **every** document in a collection, connecting with the embedded
Rust driver. Requires a connected builder.

```python
builder = SchemaBuilder(MongoConnectionProperties(uri="mongodb://localhost:27017"))
builder.build_schema_for_collection("sample_mflix", "movies")
# {'$jsonSchema': {'bsonType': 'object',
#                  'properties': {'_id': {...}, 'title': {'bsonType': 'string'},
#                                 'year': {'anyOf': [{'bsonType': 'int'}, {'bsonType': 'string'}]},
#                                 ... 22 properties total ...},
#                  'required': ['_id', 'awards', 'imdb', 'lastupdated',
#                               'num_mflix_comments', 'title', 'type', 'year'],
#                  'additionalProperties': False}}
```

A missing collection raises `ValueError` (e.g. `Failed to derive schema for
sample_mflix.nope: NoCollection nope`), never a panic.

### `sample_schema_for_collection(db_name, collection_name, sample_size) -> dict`

Builds a schema from a **random sample** of documents in a collection. Requires a
connected builder. `sample_size` is an `int`; values above 50,000 are clamped to 50,000.

```python
builder.sample_schema_for_collection("sample_mflix", "movies", sample_size=100)
```

## Return value

Every method returns a `dict` with a single top-level key, `"$jsonSchema"`:

```python
schema = builder.build_schema_for_document({"a": 1})
schema["$jsonSchema"]["bsonType"]        # 'object'
schema["$jsonSchema"]["properties"]["a"] # {'bsonType': 'long'}
schema["$jsonSchema"]["required"]        # ['a']
```

## Errors

All failures surface as Python `ValueError` (never a panic/abort):

| Situation | Where raised |
| --- | --- |
| Unsupported Python value (e.g. `set`) | `build_schema_for_document*` |
| Malformed BSON bytes | `build_schema_for_document_bytes` |
| Collection-level method on a connection-less builder | collection methods |
| Malformed connection string | `SchemaBuilder(...)` construction |
| Connection failure / missing collection | first collection-level call |

```python
import pytest

with pytest.raises(ValueError):
    SchemaBuilder().build_schema_for_document({"a": {1, 2}})  # set is unsupported

with pytest.raises(ValueError):
    SchemaBuilder().build_schema_for_collection("db", "coll")  # no connection

with pytest.raises(ValueError):
    SchemaBuilder(MongoConnectionProperties(uri="not a uri"))  # bad URI, at construction
```

## Notes

- **Conversion degradations** (shared BSON bridge): ObjectId → hex string,
  Decimal128 / exotic BSON → stringified; `datetime` round-trips through epoch millis;
  unsupported Python types raise `ValueError`.
- **`int` is `long` for Python input.** Document-level derivation maps a Python `int`
  to BSON `long`. Schemas derived from a live collection instead report the type the
  server actually stored, so a 32-bit value like `sample_mflix.movies`' `imdb.id` comes
  back as `{"bsonType": "int"}`.
