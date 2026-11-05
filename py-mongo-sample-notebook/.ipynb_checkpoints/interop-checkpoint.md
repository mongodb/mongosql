# Interop: schema builder + MongoSQL translator

`py-mongo-schema-builder` derives a JSON schema from a live collection;
`py-mongosql-translate` turns SQL into an aggregation pipeline, using that schema as its
catalog. Together they give you a typed SQL-over-MongoDB workflow with no manual schema
authoring.

The two pieces fit together with **no glue code**: `SchemaBuilder` returns
`{"$jsonSchema": {...}}`, which is exactly the wrapper the translator unwraps.

```
SchemaBuilder.build_schema_for_collection(db, coll)  ->  {"$jsonSchema": {...}}
                                                              |
                          catalog = {db: {coll: schema}}       |
                                                              v
MongoSqlTranslator.translate_query(db, sql, catalog)  ->  Translation
                                                              |
                          client[target_db][target_collection].aggregate(pipeline)
```

All examples below run against `sample_mflix.movies` and were executed against a live
MongoDB at `mongodb://localhost:27017`. Every block repeats the imports it needs, so you
can read any one on its own.

## Setup

```python
from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder
from py_mongosqltranslate import MongoSqlTranslator
from pymongo import MongoClient

URI = "mongodb://localhost:27017"
DB, COLL = "sample_mflix", "movies"
```

## 1. Derive the catalog from the collection

```python
from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder

URI = "mongodb://localhost:27017"
DB, COLL = "sample_mflix", "movies"

builder = SchemaBuilder(MongoConnectionProperties(uri=URI))
movies_schema = builder.build_schema_for_collection(DB, COLL)

catalog = {DB: {COLL: movies_schema}}
```

`movies_schema` is `{"$jsonSchema": {"bsonType": "object", "properties": {...}, ...}}`.
For `sample_mflix.movies` it describes **22 properties** with these required fields:

```python
# Not runnable code — this is the value of
# catalog["sample_mflix"]["movies"]["$jsonSchema"]["required"].
['_id', 'awards', 'imdb', 'lastupdated', 'num_mflix_comments', 'title', 'type', 'year']
```

Note that `year` widens to `{"anyOf": [{"bsonType": "int"}, {"bsonType": "string"}]}` —
the collection stores some years as strings — and `imdb` is a nested object with its own
`id`, `votes`, and `rating` fields.

## 2. Translate a query against that catalog

```python
from py_mongosqltranslate import MongoSqlTranslator

DB = "sample_mflix"
# `catalog` is the value built in step 1: {DB: {COLL: movies_schema}}.

translator = MongoSqlTranslator()

sql = ("SELECT title, year, imdb.rating FROM movies "
       "WHERE year > 2000 AND imdb.rating > 8.5 ORDER BY year DESC LIMIT 5")

t = translator.translate_query(DB, sql, catalog)

print(t.target_db)          # sample_mflix
print(t.target_collection)  # movies
print(t.select_order)       # [['', 'title'], ['', 'year'], ['', 'rating']]
```

Because the catalog carries types, the translator resolves `imdb.rating` and `year`
without guessing. `t.pipeline` is the ready-to-run aggregation:

```python
# Not runnable code — this is the value of `t.pipeline`.
[{'$match': {'$and': [{'year': {'$gt': 2000}}, {'imdb.rating': {'$gt': 8.5}}]}},
 {'$sort': {'year': -1}},
 {'$limit': 5},
 {'$project': {'movies': '$$ROOT', '_id': 0}},
 {'$project': {'__bot': {'title': '$movies.title',
                         'year': '$movies.year',
                         'rating': '$movies.imdb.rating'},
               '_id': 0}},
 {'$replaceWith': {...}}]
```

## 3. Run the pipeline

`Translation` tells you where to run it: `target_db` and `target_collection`.

```python
from pymongo import MongoClient

URI = "mongodb://localhost:27017"
# `t` is the Translation from step 2.

client = MongoClient(URI)

results = list(client[t.target_db][t.target_collection].aggregate(t.pipeline))
```

## Reading the results

With namespaces included, each result document nests its fields under the namespace key.
For an explicit select list that key is the **empty string**:

```python
# Not runnable code — this is the value of `results`.
[
  {"": {"title": "Homeland (Iraq Year Zero)",                  "year": 2015, "rating": 8.8}},
  {"": {"title": "The Jinx: The Life and Deaths of Robert Durst", "year": 2015, "rating": 8.9}},
  {"": {"title": "Most Likely to Succeed",                     "year": 2015, "rating": 8.9}},
  {"": {"title": "Gayby Baby",                                 "year": 2015, "rating": 8.6}},
  {"": {"title": "A Brave Heart: The Lizzie Velasquez Story",  "year": 2015, "rating": 9.4}},
]
```

To get flat rows, pull the payload out of the namespace key using `select_order`. This
helper is pure Python — it imports nothing:

```python
# `t` is the Translation and `results` the rows from the steps above.

def flatten(translation, rows):
    """Collapse the namespace-keyed output into flat dicts of the selected fields."""
    return [
        {field: row.get(ns, {}).get(field) for ns, field in translation.select_order}
        for row in rows
    ]

for row in flatten(t, results):
    print(row)
# {'title': 'Homeland (Iraq Year Zero)', 'year': 2015, 'rating': 8.8}
# {'title': 'The Jinx: The Life and Deaths of Robert Durst', 'year': 2015, 'rating': 8.9}
# {'title': 'Most Likely to Succeed', 'year': 2015, 'rating': 8.9}
# {'title': 'Gayby Baby', 'year': 2015, 'rating': 8.6}
# {'title': 'A Brave Heart: The Lizzie Velasquez Story', 'year': 2015, 'rating': 9.4}
```

For `SELECT *`, the namespace key is the collection name instead of `''`:

```python
from py_mongosqltranslate import MongoSqlTranslator

DB = "sample_mflix"
# `catalog` is the value built in step 1.

translator = MongoSqlTranslator()

t_star = translator.translate_query(DB, "SELECT * FROM movies LIMIT 1", catalog)
t_star.select_order[:3]   # [['movies', '_id'], ['movies', 'awards'], ['movies', 'cast']]
```

## Sampling instead of full-collection derivation

For large collections, sample a subset of documents. The resulting schema is a valid
catalog entry just the same:

```python
from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder
from py_mongosqltranslate import MongoSqlTranslator
from pymongo import MongoClient

URI = "mongodb://localhost:27017"
DB, COLL = "sample_mflix", "movies"

builder = SchemaBuilder(MongoConnectionProperties(uri=URI))
translator = MongoSqlTranslator()
client = MongoClient(URI)

sampled = builder.sample_schema_for_collection(DB, COLL, sample_size=500)
t = translator.translate_query(
    DB, "SELECT title FROM movies WHERE year = 1994 LIMIT 2", {DB: {COLL: sampled}}
)
rows = list(client[t.target_db][t.target_collection].aggregate(t.pipeline))
# [{'': {'title': 'Anna ot 6 do 18'}}, {'': {'title': 'Backbeat'}}]
```

## Skipping the schema entirely

The catalog is optional. With no schema, every collection gets an empty one and the
result set is untyped — useful for a quick probe, but you lose type resolution and get a
`UserWarning`:

```python
import warnings

from py_mongosqltranslate import MongoSqlTranslator
from pymongo import MongoClient

URI = "mongodb://localhost:27017"
DB = "sample_mflix"

translator = MongoSqlTranslator()
client = MongoClient(URI)

with warnings.catch_warnings(record=True) as caught:
    warnings.simplefilter("always")
    t = translator.translate_query(DB, "SELECT title FROM movies LIMIT 2", None)

print([str(w.message) for w in caught])
# ['No schema was found for the following collections: ["sample_mflix.movies"].
#   These collections will be assigned empty schemas.']

rows = list(client[t.target_db][t.target_collection].aggregate(t.pipeline))
# [{'': {'title': 'The Great Train Robbery'}}, {'': {'title': 'A Corner in Wheat'}}]
```

## Wrapping it in a reusable helper

```python
import warnings

from py_mongosqltranslate import MongoSqlTranslator
from pymongo import MongoClient

URI = "mongodb://localhost:27017"
DB = "sample_mflix"

def run_sql(sql, db, catalog=None, uri=URI):
    """Translate `sql` and run it, returning (translation, rows)."""
    translator = MongoSqlTranslator()
    with warnings.catch_warnings():
        warnings.simplefilter("ignore")  # empty-schema fallback is expected here
        t = translator.translate_query(db, sql, catalog)
    rows = list(MongoClient(uri)[t.target_db][t.target_collection].aggregate(t.pipeline))
    return t, rows

# `catalog` is the value built in step 1.
t, rows = run_sql("SELECT type, COUNT(*) AS n FROM movies GROUP BY type", DB, catalog)
for row in rows:
    print(row)
# {'': {'type': 'movie', 'n': 21117}}
# {'': {'type': 'series', 'n': 232}}
```

## Notes and gotchas

- **One schema, two uses.** `SchemaBuilder` output is a bare `{"$jsonSchema": {...}}`
  document; the translator accepts both that wrapper and an unwrapped schema, so no
  conversion step is needed.
- **The catalog is `{db: {collection: schema}}`.** Wrap the derived schema in the
  database and collection names as shown above.
- **Partial catalogs are fine.** A query that joins a catalogued collection with an
  uncatalogued one translates: the missing collection is filled with an empty schema and
  a `UserWarning` names it.
- **Namespaced output.** Result documents are keyed by namespace (`''` for an explicit
  select list, the collection name for `SELECT *`). Use `Translation.select_order` to
  flatten.
- **Mixed-type fields are real.** `sample_mflix.movies.year` is `anyOf: [int, string]`,
  so `ORDER BY year` and comparisons follow MongoDB's cross-type ordering — expect string
  years like `"2014"` alongside numeric ones.
- **Rebuild after editing Rust.** Python caches the loaded extension per process; start a
  fresh interpreter after each `maturin develop`.
