"""Smoke test for py_schema_builder_library.

Run from this directory with the venv's interpreter:
    .venv/bin/python examples/smoke_test.py

Requires the module to be built first:
    .venv/bin/maturin develop

Collection-level tests additionally need a running MongoDB (default
mongodb://localhost:27017) and PyMongo installed.
"""

from py_schema_builder_library import SchemaBuilder, sum_as_string

import bson


def document_tests() -> None:
    # Free function
    print(sum_as_string(2, 40))  # "42"

    # Document-level methods work without a client. Exercises the
    # defaulted constructor (`client` defaults to None).
    b = SchemaBuilder()
    doc = {"a": 1, "b": "x"}

    # Raw-bytes path: encode the document in Python with pymongo's bson...
    encoded = bson.encode(doc)
    schema_bytes = b.build_schema_for_document_bytes(encoded)
    assert isinstance(schema_bytes, dict), type(schema_bytes)
    print("schema from bytes:", schema_bytes)

    # ...or let the Rust side do the encoding via bson.encode
    schema_dict = b.build_schema_for_document(doc)
    assert isinstance(schema_dict, dict), type(schema_dict)
    print("schema from dict:", schema_dict)

    # Multiple documents at once: heterogeneous types across docs should
    # widen the schema (b: string in doc1, int in doc2 -> both listed).
    schema_multi = b.build_schema_for_documents(
        [{"a": 1, "b": "x"}, {"a": 2, "b": 3, "c": True}]
    )
    assert isinstance(schema_multi, dict), type(schema_multi)
    print("schema from list:", schema_multi)

    # Error mapping check: malformed bytes must surface as Python ValueError
    try:
        b.build_schema_for_document_bytes(b"\x00\x01\x02")
    except ValueError as e:
        print("ValueError raised as expected:", e)

    # Client-less guard: collection-level methods must raise ValueError.
    try:
        b.build_schema_for_collection("db", "coll")
    except ValueError as e:
        print("ValueError raised as expected:", e)


def collection_tests(b: SchemaBuilder) -> None:
    import pymongo

    db = "smoke_test_db"
    coll = "smoke_test_coll"

    client = pymongo.MongoClient()
    client[db][coll].delete_many({})
    client[db][coll].insert_many([{"a": 1, "b": "x"}, {"a": 2, "b": "y"}])

    schema = b.build_schema_for_collection(db, coll)
    assert isinstance(schema, dict), type(schema)
    print("collection schema:", schema)

    sampled = b.sample_schema_for_collection(db, coll, sample_size=10)
    assert isinstance(sampled, dict), type(sampled)
    print("sampled schema:", sampled)

    # Missing collection must surface as ValueError, not panic.
    try:
        b.build_schema_for_collection(db, "does_not_exist")
    except ValueError as e:
        print("ValueError raised as expected:", e)

    client[db][coll].drop()


def main() -> None:
    document_tests()

    try:
        import pymongo

        pymongo.MongoClient().admin.command("ping")
    except Exception as e:  # noqa: BLE001 - smoke test: any failure means skip
        print(f"Skipping collection tests (no MongoDB available: {e})")
        return

    collection_tests(SchemaBuilder(pymongo.MongoClient()))


if __name__ == "__main__":
    main()
