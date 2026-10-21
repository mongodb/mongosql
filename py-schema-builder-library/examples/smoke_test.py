"""Smoke test for py_schema_builder_library.

Run from this directory with the venv's interpreter:
    .venv/bin/python examples/smoke_test.py

Requires the module to be built first:
    .venv/bin/maturin develop

Collection-level tests additionally need a running MongoDB (default
mongodb://localhost:27017). No Python MongoDB dependencies are required:
all I/O and BSON conversion happen in Rust.
"""

from py_schema_builder_library import MongoConnectionProperties, SchemaBuilder


def document_tests() -> None:
    # Document-level methods work without a connection.
    b = SchemaBuilder()
    doc = {"a": 1, "b": "x"}

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


def collection_tests() -> None:
    props = MongoConnectionProperties(uri="mongodb://localhost:27017")
    b = SchemaBuilder(props)

    db = "smoke_test_db"
    coll = "smoke_test_coll"

    # Seed via mongosh (collection seeding is outside the library's scope;
    # the all-BSON-types collection used for deeper testing is seeded the
    # same way). _id is server-generated.
    import subprocess

    seed = ['{"a": 1, "b": "x"}', '{"a": 2, "b": "y"}']
    script = (
        'db = db.getSiblingDB("smoke_test_db");'
        "db.smoke_test_coll.drop();"
        "db.smoke_test_coll.insertMany([" + ",".join(seed) + "]);"
    )
    try:
        subprocess.run(["mongosh", "--quiet", "--eval", script], check=True, capture_output=True)
    except (FileNotFoundError, subprocess.CalledProcessError) as e:
        print(f"Skipping collection tests (cannot seed via mongosh: {e})")
        return

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

    subprocess.run(
        ["mongosh", "--quiet", "--eval", f"db.getSiblingDB('{db}').{coll}.drop()"],
        check=True,
        capture_output=True,
    )


def main() -> None:
    document_tests()
    collection_tests()


if __name__ == "__main__":
    main()
