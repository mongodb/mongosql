"""Smoke test for py_schema_builder_library.

Run from this directory with the venv's interpreter:
    .venv/bin/python examples/smoke_test.py

Requires the module to be built first:
    .venv/bin/maturin develop
"""

from py_schema_builder_library import SchemaBuilder, sum_as_string

import bson


def main() -> None:
    # Free function
    print(sum_as_string(2, 40))  # "42"

    b = SchemaBuilder()
    doc = {"a": 1, "b": "x"}

    # Raw-bytes path: encode the document in Python with pymongo's bson...
    encoded = bson.encode(doc)
    print(b.build_schema_for_document_bytes(encoded))

    # ...or let the Rust side do the encoding via bson.encode
    print(b.build_schema_for_document(doc))

    # Error mapping check: malformed bytes must surface as Python ValueError
    try:
        b.build_schema_for_document_bytes(b"\x00\x01\x02")
    except ValueError as e:
        print("ValueError raised as expected:", e)


if __name__ == "__main__":
    main()
