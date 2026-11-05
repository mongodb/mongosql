"""Smoke test for py_mongosqltranslate.

Run from this directory with the venv's interpreter:
    .venv/bin/python examples/smoke_test.py

Requires the module to be built first:
    .venv/bin/maturin develop

Translation itself needs no MongoDB. The optional SchemaBuilder interop
section additionally needs `py_schema_builder_library` installed in the same
venv, `mongosh` on PATH, and a running MongoDB (mongodb://localhost:27017);
it is skipped otherwise.
"""

import subprocess
import warnings

from py_mongosqltranslate import MongoSqlTranslator

CATALOG = {
    "test": {
        "foo": {
            "bsonType": "object",
            "properties": {
                "a": {"bsonType": "int"},
                "b": {"bsonType": "string"},
            },
            "required": ["a"],
            "additionalProperties": False,
        }
    }
}


def show(label: str, translation) -> None:
    print(f"--- {label}")
    print("repr:             ", translation)
    print("target:           ", f"{translation.target_db}.{translation.target_collection}")
    print("pipeline:         ", translation.pipeline)
    print("result_set_schema:", translation.result_set_schema)
    print("select_order:     ", translation.select_order)


def translation_tests() -> None:
    translator = MongoSqlTranslator()
    print("translator:", translator)
    t = translator.translate_query("test", "SELECT a, b FROM foo WHERE a > 1 ORDER BY a", CATALOG)
    assert t.target_collection == "foo", t.target_collection
    show("translate_query with catalog", t)

    v = translator.translate_query_verbose("test", "SELECT a, b FROM foo WHERE a > 1 ORDER BY a", CATALOG)
    assert v.pipeline == t.pipeline
    show("translate_query_verbose with catalog", v)

    # catalog defaults to None: queries without collection sources need no schema.
    t = translator.translate_query("test", "SELECT * FROM [{'a': 1}] arr")
    assert t.target_collection is None
    show("translate_query without catalog", t)

    # A collection with no schema is assigned an empty one, as mongosql-cli
    # does, and the caller is warned. The result set is then untyped.
    with warnings.catch_warnings(record=True) as caught:
        warnings.simplefilter("always")
        t = translator.translate_query("test", "SELECT a, b FROM foo")
    for warning in caught:
        print("warning:", warning.message)
    assert t.target_collection == "foo"
    # Namespaces are included, so fields sit under the (empty) namespace key.
    assert t.result_set_schema["properties"][""]["properties"] == {"a": {}, "b": {}}
    show("translate_query with an empty-schema catalog", t)

    # translate_query_simple is the two-argument form of the same thing.
    with warnings.catch_warnings(record=True) as caught:
        warnings.simplefilter("always")
        simple = translator.translate_query_simple("test", "SELECT a, b FROM foo")
    assert simple.pipeline == t.pipeline
    assert len(caught) == 1
    show("translate_query_simple", simple)

    # Error mapping: failures surface as ValueError, never panics.
    for label, args in [
        ("SQL syntax error", ("test", "SELEC a FROM foo", CATALOG)),
        ("malformed catalog", ("test", "SELECT a FROM foo", {"test": {"foo": {"bsonType": "nope"}}})),
    ]:
        try:
            translator.translate_query(*args)
        except ValueError as e:
            print(f"ValueError raised as expected ({label}):", e)
        else:
            raise AssertionError(f"expected ValueError for {label}")


def schema_builder_interop() -> None:
    """Builds the catalog from a real collection with py_schema_builder_library."""
    try:
        from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder
    except ImportError:
        print("Skipping SchemaBuilder interop (py_schema_builder_library not installed)")
        return

    db, coll = "smoke_test_db", "smoke_test_coll"
    script = (
        f'db = db.getSiblingDB("{db}");'
        f"db.{coll}.drop();"
        f'db.{coll}.insertMany([{{"a": 1, "b": "x"}}, {{"a": 2, "b": "y"}}]);'
    )
    try:
        subprocess.run(["mongosh", "--quiet", "--eval", script], check=True, capture_output=True)
    except (FileNotFoundError, subprocess.CalledProcessError) as e:
        print(f"Skipping SchemaBuilder interop (cannot seed via mongosh: {e})")
        return

    try:
        builder = SchemaBuilder(MongoConnectionProperties(uri="mongodb://localhost:27017"))
        # SchemaBuilder returns {"$jsonSchema": {...}}; the catalog unwraps it.
        schema = builder.build_schema_for_collection(db, coll)
        t = MongoSqlTranslator().translate_query(
            db, f"SELECT a, b FROM {coll}", {db: {coll: schema}}
        )
        assert t.target_collection == coll
        show("translate_query with SchemaBuilder catalog", t)
    finally:
        subprocess.run(
            ["mongosh", "--quiet", "--eval", f"db.getSiblingDB('{db}').{coll}.drop()"],
            check=True,
            capture_output=True,
        )


def main() -> None:
    translation_tests()
    schema_builder_interop()


if __name__ == "__main__":
    main()
