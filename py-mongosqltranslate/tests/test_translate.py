"""Translation tests for py_mongosqltranslate (no MongoDB required).

Run from the crate directory after `maturin develop`:
    .venv/bin/pytest tests/ -v

Namespaces are included in the result set, so `select_order` entries are
`[namespace, field]` and `result_set_schema` is keyed by namespace. For an
explicit select list the namespace is the empty string; `SELECT *` uses the
collection name.
"""

import warnings

import pytest

from py_mongosqltranslate import MongoSqlTranslator, Translation

FOO_SCHEMA = {
    "bsonType": "object",
    "properties": {
        "a": {"bsonType": "int"},
        "b": {"bsonType": "string"},
    },
    "required": ["a"],
    "additionalProperties": False,
}

CATALOG = {"test": {"foo": FOO_SCHEMA}}

TRANSLATOR = MongoSqlTranslator()


def test_construct_and_reuse():
    first = MongoSqlTranslator()
    second = MongoSqlTranslator()

    assert first is not second
    assert first.translate_query("test", "SELECT a FROM foo", CATALOG).pipeline == (
        second.translate_query("test", "SELECT a FROM foo", CATALOG).pipeline
    )


def test_translate_returns_translation():
    t = TRANSLATOR.translate_query("test", "SELECT a, b FROM foo", CATALOG)

    assert isinstance(t, Translation)
    assert t.target_db == "test"
    assert t.target_collection == "foo"
    assert isinstance(t.pipeline, list) and t.pipeline
    assert all(isinstance(stage, dict) for stage in t.pipeline)
    # An explicit select list reports the empty namespace, and both fields
    # appear under it.
    assert t.select_order == [["", "a"], ["", "b"]]
    assert t.result_set_schema["properties"][""]["properties"] == {
        "a": {"bsonType": "int"},
        "b": {"bsonType": "string"},
    }


def test_select_star_uses_the_collection_namespace():
    t = TRANSLATOR.translate_query("test", "SELECT * FROM foo", CATALOG)

    assert t.select_order == [["foo", "a"], ["foo", "b"]]
    assert set(t.result_set_schema["properties"]) == {"foo"}


def test_translate_query_matches_verbose():
    sql = "SELECT a, b FROM foo WHERE a > 1 ORDER BY a"
    simple = TRANSLATOR.translate_query("test", sql, CATALOG)
    verbose = TRANSLATOR.translate_query_verbose("test", sql, CATALOG)

    assert simple.target_db == verbose.target_db
    assert simple.target_collection == verbose.target_collection
    assert simple.pipeline == verbose.pipeline
    assert simple.result_set_schema == verbose.result_set_schema
    assert simple.select_order == verbose.select_order


def test_catalog_keyword_argument():
    t = TRANSLATOR.translate_query(current_db="test", sql="SELECT a FROM foo", catalog=CATALOG)
    assert t.select_order == [["", "a"]]


def test_json_schema_wrapped_catalog_is_unwrapped():
    wrapped = {"test": {"foo": {"$jsonSchema": FOO_SCHEMA}}}
    sql = "SELECT a, b FROM foo"

    assert (
        TRANSLATOR.translate_query("test", sql, wrapped).pipeline
        == TRANSLATOR.translate_query("test", sql, CATALOG).pipeline
    )


@pytest.mark.parametrize("catalog_kwargs", [{}, {"catalog": None}])
def test_catalog_is_optional(catalog_kwargs):
    t = TRANSLATOR.translate_query("test", "SELECT * FROM [{'a': 1}] arr", **catalog_kwargs)

    assert t.target_db == "test"
    assert t.target_collection is None
    assert t.pipeline


def test_no_catalog_translates_collection_query():
    # A collection with no schema is assigned an empty one rather than failing,
    # matching mongosql-cli. The caller is warned.
    with pytest.warns(UserWarning, match=r"test\.foo"):
        t = TRANSLATOR.translate_query("test", "SELECT a FROM foo")

    assert t.target_collection == "foo"
    assert t.pipeline


def test_empty_schema_fallback_is_untyped():
    with pytest.warns(UserWarning):
        t = TRANSLATOR.translate_query("test", "SELECT a FROM foo")

    # No schema means no types: the field schema is empty.
    assert t.result_set_schema["properties"][""]["properties"] == {"a": {}}


def test_partial_catalog_fills_missing_namespace():
    catalog = {"test": {"bar": {"bsonType": "object", "properties": {"c": {"bsonType": "int"}}}}}

    with pytest.warns(UserWarning, match=r"test\.foo"):
        t = TRANSLATOR.translate_query(
            "test", "SELECT foo.a, bar.c FROM foo JOIN bar ON foo.a = bar.c", catalog
        )

    assert t.pipeline
    # The provided schema is still used for the covered collection, so the
    # result set is typed where possible and untyped only for `foo`.
    fields = t.result_set_schema["properties"][""]["properties"]
    assert fields["c"]["bsonType"] == "int"
    assert fields["a"] == {}


def test_no_warning_when_catalog_covers_query():
    with warnings.catch_warnings(record=True) as caught:
        warnings.simplefilter("always")
        TRANSLATOR.translate_query("test", "SELECT a, b FROM foo", CATALOG)

    assert caught == []


def test_no_warning_without_collection_sources():
    with warnings.catch_warnings(record=True) as caught:
        warnings.simplefilter("always")
        TRANSLATOR.translate_query("test", "SELECT * FROM [{'a': 1}] arr")

    assert caught == []


def test_translate_query_simple():
    with pytest.warns(UserWarning, match=r"test\.foo"):
        simple = TRANSLATOR.translate_query_simple("test", "SELECT a FROM foo")
    with pytest.warns(UserWarning):
        explicit = TRANSLATOR.translate_query("test", "SELECT a FROM foo", None)

    assert simple.target_collection == explicit.target_collection
    assert simple.pipeline == explicit.pipeline


def test_translate_query_simple_requires_sql():
    with pytest.raises(TypeError):
        TRANSLATOR.translate_query_simple("test")  # type: ignore[call-arg]


def test_sql_syntax_error_raises():
    with pytest.raises(ValueError):
        TRANSLATOR.translate_query("test", "SELEC a FROM foo", CATALOG)


@pytest.mark.parametrize(
    "catalog",
    [
        pytest.param({"test": "not a dict"}, id="collections-not-dict"),
        pytest.param({"test": {"foo": "not a dict"}}, id="schema-not-dict"),
        pytest.param({"test": {"foo": {"bsonType": "notAType"}}}, id="unknown-bson-type"),
        pytest.param({1: {"foo": FOO_SCHEMA}}, id="non-str-db"),
    ],
)
def test_malformed_catalog_raises(catalog):
    with pytest.raises(ValueError):
        TRANSLATOR.translate_query("test", "SELECT a FROM foo", catalog)


def test_missing_required_args_raise_type_error():
    with pytest.raises(TypeError):
        TRANSLATOR.translate_query("test")  # type: ignore[call-arg]
    with pytest.raises(TypeError):
        TRANSLATOR.translate_query_verbose()  # type: ignore[call-arg]


def test_cross_database_query_sets_target_db():
    catalog = {
        **CATALOG,
        "other": {"bar": {"bsonType": "object", "properties": {"x": {"bsonType": "long"}}}},
    }
    t = TRANSLATOR.translate_query("test", "SELECT x FROM other.bar", catalog)

    assert t.target_db == "other"
    assert t.target_collection == "bar"
    assert t.select_order == [["", "x"]]


def test_translation_is_read_only():
    t = TRANSLATOR.translate_query("test", "SELECT a FROM foo", CATALOG)
    with pytest.raises(AttributeError):
        t.target_db = "changed"  # type: ignore[misc]


def test_repr():
    t = TRANSLATOR.translate_query("test", "SELECT a FROM foo", CATALOG)
    assert repr(t) == (
        "Translation(target_db='test', target_collection='foo', "
        f"pipeline=<{len(t.pipeline)} stages>, select_order=[['', 'a']])"
    )
