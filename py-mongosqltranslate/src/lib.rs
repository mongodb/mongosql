//! Python bindings for translating `MongoSQL` queries into `MongoDB` aggregation pipelines.
//!
//! The extension module `py_mongosqltranslate` exposes a `MongoSqlTranslator`
//! class whose methods (`translate_query`, `translate_query_verbose`,
//! `translate_query_simple`) wrap [`mongosql::translate_sql`] and return a
//! `Translation` object.
//!
//! Translation always runs with namespaces included in the result set and
//! relaxed schema checking (`SqlOptions::new(IncludeNamespaces, Relaxed)`).

// Rust guideline compliant 2026-02-21

mod py_bson;

use pyo3::prelude::*;

/// A Python module implemented in Rust.
#[pymodule]
mod py_mongosqltranslate {
    use crate::py_bson::{document_to_pydict, py_dict_to_document};
    use agg_ast::Namespace;
    use bson::{Bson, Document};
    use mongosql::catalog::Catalog;
    use mongosql::json_schema;
    use mongosql::options::{ExcludeNamespacesOption, SqlOptions};
    use mongosql::schema::Schema as MongoSqlSchema;
    use mongosql::{SchemaCheckingMode, Translation};
    use pyo3::PyTypeInfo;
    use pyo3::exceptions::{PyUserWarning, PyValueError};
    use pyo3::prelude::*;
    use pyo3::types::{PyDict, PyList};
    use std::collections::BTreeMap;
    use std::collections::btree_map::Entry;
    use std::ffi::CString;

    /// Key under which `SchemaBuilder` (and `$jsonSchema` validators) wrap a
    /// collection's JSON schema. Catalog entries of the form
    /// `{"$jsonSchema": {...}}` are unwrapped so derived schemas can be used
    /// directly as catalog values.
    const JSON_SCHEMA_KEY: &str = "$jsonSchema";

    /// The MQL translation of a SQL query, as returned to Python.
    ///
    /// All fields are read-only and already converted to Python objects:
    /// `pipeline` is a list of stage dicts, `result_set_schema` is a JSON
    /// schema dict, and `select_order` lists the result field paths.
    #[pyclass(name = "Translation", frozen)]
    #[derive(Debug)]
    pub(crate) struct PyTranslation {
        /// Database the pipeline must be run against.
        #[pyo3(get)]
        target_db: String,
        /// Collection the pipeline must be run against; `None` means a
        /// database-level aggregation (e.g. `SELECT 1` or array datasources).
        #[pyo3(get)]
        target_collection: Option<String>,
        /// Aggregation pipeline as a list of stage dicts.
        #[pyo3(get)]
        pipeline: Py<PyList>,
        /// JSON schema describing the documents the pipeline produces.
        #[pyo3(get)]
        result_set_schema: Py<PyDict>,
        /// Field paths of the result set in SELECT-list order.
        #[pyo3(get)]
        select_order: Vec<Vec<String>>,
    }

    impl PyTranslation {
        /// Converts a [`Translation`] into its Python representation.
        fn from_translation(py: Python<'_>, translation: Translation) -> PyResult<Self> {
            let Bson::Array(stages) = translation.pipeline else {
                return Err(PyValueError::new_err(format!(
                    "Expected translated pipeline to be an array, got {}",
                    translation.pipeline
                )));
            };
            let stages = stages
                .iter()
                .map(|stage| match stage {
                    Bson::Document(doc) => document_to_pydict(py, doc),
                    other => Err(PyValueError::new_err(format!(
                        "Expected pipeline stage to be a document, got {other}"
                    ))),
                })
                .collect::<PyResult<Vec<_>>>()?;

            let result_set_schema = match translation.result_set_schema.to_bson() {
                Ok(Bson::Document(doc)) => document_to_pydict(py, &doc)?,
                Ok(other) => {
                    return Err(PyValueError::new_err(format!(
                        "Expected result set schema to be a document, got {other}"
                    )));
                }
                Err(e) => {
                    return Err(PyValueError::new_err(format!(
                        "Failed to convert result set schema to BSON: {e}"
                    )));
                }
            };

            Ok(Self {
                target_db: translation.target_db,
                target_collection: translation.target_collection,
                pipeline: PyList::new(py, stages)?.unbind(),
                result_set_schema: result_set_schema.unbind(),
                select_order: translation.select_order,
            })
        }
    }

    #[pymethods]
    impl PyTranslation {
        /// Renders field values with Python `repr()` so the output reads as
        /// Python (`None`, `'foo'`, `[['a']]`) rather than Rust `Debug`.
        fn __repr__(slf: &Bound<'_, Self>) -> PyResult<String> {
            let py = slf.py();
            let this = slf.get();
            let target_db = this.target_db.as_str().into_pyobject(py)?.repr()?;
            let target_collection = this
                .target_collection
                .as_deref()
                .into_pyobject(py)?
                .repr()?;
            let select_order = this.select_order.clone().into_pyobject(py)?.repr()?;
            Ok(format!(
                "Translation(target_db={target_db}, target_collection={target_collection}, \
                 pipeline=<{} stages>, select_order={select_order})",
                this.pipeline.bind(py).len(),
            ))
        }
    }

    /// Converts a Python catalog dict into a namespace-to-schema map.
    ///
    /// The expected shape is `{db: {collection: json_schema}}`, the same shape
    /// as the `mongosqltranslate` `schemaCatalog` option. Each schema may be a
    /// bare JSON schema or wrapped as `{"$jsonSchema": {...}}`.
    ///
    /// This intentionally does not reuse
    /// `mongosql::build_catalog_from_catalog_schema`, which panics on invalid
    /// schemas; every failure here is a `ValueError` naming the namespace.
    fn py_dict_to_schemas(
        catalog: &Bound<'_, PyDict>,
    ) -> PyResult<BTreeMap<Namespace, MongoSqlSchema>> {
        let mut entries = BTreeMap::new();
        for (db, collections) in catalog.iter() {
            let db: String = db.extract().map_err(|_| {
                PyValueError::new_err(format!("Catalog database names must be strings, got {db}"))
            })?;
            let collections = collections.cast_into::<PyDict>().map_err(|e| {
                PyValueError::new_err(format!(
                    "Catalog entry for database '{db}' must be a dict of collection schemas: {e}"
                ))
            })?;
            for (collection, schema) in collections.iter() {
                let collection: String = collection.extract().map_err(|_| {
                    PyValueError::new_err(format!(
                        "Catalog collection names in database '{db}' must be strings, got {collection}"
                    ))
                })?;
                let schema = schema.cast_into::<PyDict>().map_err(|e| {
                    PyValueError::new_err(format!(
                        "Catalog schema for '{db}.{collection}' must be a dict: {e}"
                    ))
                })?;
                let schema = py_schema_to_mongosql_schema(&schema).map_err(|e| {
                    // `e.value(py)` renders only the message; `PyErr`'s own
                    // Display would prefix a redundant "ValueError: ".
                    PyValueError::new_err(format!(
                        "Invalid catalog schema for '{db}.{collection}': {}",
                        e.value(catalog.py())
                    ))
                })?;
                entries.insert(
                    Namespace {
                        database: db.clone(),
                        collection,
                    },
                    schema,
                );
            }
        }
        Ok(entries)
    }

    /// Builds the [`Catalog`] for a query, filling in empty schemas for any
    /// referenced namespace the caller's catalog does not cover.
    ///
    /// Collections are discovered with `mongosql::get_namespaces`, which parses
    /// the query (so a syntax error surfaces here). Each uncovered namespace is
    /// assigned `Schema::Any` and reported through a `UserWarning`, mirroring
    /// `mongosql-cli`, which substitutes empty schemas rather than failing.
    ///
    /// The consequence is that a collection missing from the catalog no longer
    /// raises `CollectionNotFound`; it translates with an untyped schema.
    fn resolve_catalog(
        py: Python<'_>,
        current_db: &str,
        sql: &str,
        catalog: Option<&Bound<'_, PyDict>>,
    ) -> PyResult<Catalog> {
        let mut schemas = match catalog {
            Some(catalog) => py_dict_to_schemas(catalog)?,
            None => BTreeMap::new(),
        };

        let mut filled = Vec::new();
        for namespace in mongosql::get_namespaces(current_db, sql)
            .map_err(|e| PyValueError::new_err(e.to_string()))?
        {
            if let Entry::Vacant(entry) = schemas.entry(namespace) {
                filled.push(format!(
                    "{}.{}",
                    entry.key().database,
                    entry.key().collection
                ));
                entry.insert(MongoSqlSchema::Any);
            }
        }
        if !filled.is_empty() {
            warn_empty_schemas(py, &filled)?;
        }

        Ok(Catalog::new(schemas))
    }

    /// Emits the `UserWarning` that reports auto-filled empty schemas.
    fn warn_empty_schemas(py: Python<'_>, namespaces: &[String]) -> PyResult<()> {
        let message = format!(
            "No schema was found for the following collections: {namespaces:?}. \
             These collections will be assigned empty schemas."
        );
        // `CString::new` only fails on interior NUL bytes, which a namespace
        // cannot contain.
        let message = CString::new(message)
            .map_err(|e| PyValueError::new_err(format!("Failed to build warning message: {e}")))?;
        PyErr::warn(py, &PyUserWarning::type_object(py), &message, 1)
    }

    /// Converts one collection's JSON schema dict into a mongosql schema,
    /// unwrapping a `{"$jsonSchema": {...}}` wrapper if present.
    fn py_schema_to_mongosql_schema(schema: &Bound<'_, PyDict>) -> PyResult<MongoSqlSchema> {
        let document = unwrap_json_schema(py_dict_to_document(schema)?);
        let json_schema = json_schema::Schema::from_document(&document)
            .map_err(|e| PyValueError::new_err(e.to_string()))?;
        MongoSqlSchema::try_from(json_schema).map_err(|e| PyValueError::new_err(e.to_string()))
    }

    /// Returns the inner schema of a document of the exact form
    /// `{"$jsonSchema": {...}}`; any other document is returned unchanged.
    fn unwrap_json_schema(mut document: Document) -> Document {
        if document.len() == 1
            && let Ok(inner) = document.get_document_mut(JSON_SCHEMA_KEY)
        {
            return std::mem::take(inner);
        }
        document
    }

    /// Translates `MongoSQL` queries into `MongoDB` aggregation pipelines.
    ///
    /// Stateless; construct once and reuse. Translation always runs with
    /// namespaces included in the result set and relaxed schema checking
    /// (`SqlOptions::new(IncludeNamespaces, Relaxed)`).
    #[pyclass]
    #[derive(Debug)]
    pub(crate) struct MongoSqlTranslator;

    #[pymethods]
    impl MongoSqlTranslator {
        /// Creates a translator.
        #[new]
        fn new() -> Self {
            Self
        }

        /// Translates a SQL query into a MongoDB aggregation pipeline.
        ///
        /// `current_db` is the database unqualified collection names resolve
        /// against. `catalog` maps `{db: {collection: json_schema}}` for the
        /// collections the query reads; schemas may be bare JSON schemas or wrapped
        /// as `{"$jsonSchema": {...}}` (the `SchemaBuilder` output format).
        ///
        /// Any collection the query reads that `catalog` does not cover is assigned
        /// an empty schema, so the catalog is optional: pass `None` (or omit it) to
        /// translate without schema information. Empty schemas make the result
        /// untyped, so the result set schema's properties are `{}` rather than
        /// typed, and a mistyped collection name translates instead of failing.
        ///
        /// # Warnings
        ///
        /// Emits a `UserWarning` naming every collection assigned an empty schema.
        /// No warning is emitted when the catalog covers the query, or when the
        /// query reads no collections (`SELECT 1`, `SELECT * FROM [{'a': 1}] arr`).
        ///
        /// # Errors
        ///
        /// Raises `ValueError` if the catalog is malformed (non-string keys,
        /// non-dict values, or schemas that are not valid JSON schemas), or if the
        /// query fails to parse or otherwise fails to translate.
        #[expect(
            clippy::unused_self,
            reason = "MongoSqlTranslator is stateless; the Python API is instance-based to match SchemaBuilder"
        )]
        #[pyo3(signature = (current_db, sql, catalog = None))]
        fn translate_query_verbose(
            &self,
            py: Python<'_>,
            current_db: &str,
            sql: &str,
            catalog: Option<&Bound<'_, PyDict>>,
        ) -> PyResult<PyTranslation> {
            let catalog = resolve_catalog(py, current_db, sql, catalog)?;
            let sql_options = SqlOptions::new(
                ExcludeNamespacesOption::IncludeNamespaces,
                SchemaCheckingMode::Relaxed,
            );

            // Translation is pure CPU work over owned Rust data, so release the GIL.
            let translation = py
                .detach(|| mongosql::translate_sql(current_db, sql, &catalog, sql_options))
                .map_err(|e| PyValueError::new_err(e.to_string()))?;

            PyTranslation::from_translation(py, translation)
        }

        /// Translates a SQL query into a MongoDB aggregation pipeline.
        ///
        /// Currently equivalent to `translate_query_verbose`; see it for the
        /// meaning of `current_db`, `sql`, and the optional `catalog`.
        ///
        /// # Warnings
        ///
        /// Emits a `UserWarning` naming every collection assigned an empty schema.
        ///
        /// # Errors
        ///
        /// Raises `ValueError` under the same conditions as
        /// `translate_query_verbose`.
        #[pyo3(signature = (current_db, sql, catalog = None))]
        fn translate_query(
            &self,
            py: Python<'_>,
            current_db: &str,
            sql: &str,
            catalog: Option<&Bound<'_, PyDict>>,
        ) -> PyResult<PyTranslation> {
            self.translate_query_verbose(py, current_db, sql, catalog)
        }

        /// Translates a SQL query into a MongoDB aggregation pipeline without a catalog.
        ///
        /// Equivalent to `translate_query(current_db, sql, None)`: every collection
        /// the query reads is assigned an empty schema, so the translation is
        /// untyped. See `translate_query_verbose` for the meaning of `current_db`
        /// and `sql`.
        ///
        /// # Warnings
        ///
        /// Emits a `UserWarning` naming every collection assigned an empty schema.
        ///
        /// # Errors
        ///
        /// Raises `ValueError` under the same conditions as
        /// `translate_query_verbose`.
        #[pyo3(signature = (current_db, sql))]
        fn translate_query_simple(
            &self,
            py: Python<'_>,
            current_db: &str,
            sql: &str,
        ) -> PyResult<PyTranslation> {
            self.translate_query_verbose(py, current_db, sql, None)
        }
    }

    #[cfg(test)]
    mod tests {
        use super::*;

        /// Builds `{"test": {"foo": <schema>}}` with `foo` having `a: int`.
        fn catalog_dict(py: Python<'_>, wrap: bool) -> Bound<'_, PyDict> {
            let props = PyDict::new(py);
            let a = PyDict::new(py);
            a.set_item("bsonType", "int").unwrap();
            props.set_item("a", a).unwrap();

            let schema = PyDict::new(py);
            schema.set_item("bsonType", "object").unwrap();
            schema.set_item("properties", props).unwrap();
            schema.set_item("required", vec!["a"]).unwrap();
            schema.set_item("additionalProperties", false).unwrap();
            let schema = if wrap {
                let wrapped = PyDict::new(py);
                wrapped.set_item(JSON_SCHEMA_KEY, schema).unwrap();
                wrapped
            } else {
                schema
            };

            let collections = PyDict::new(py);
            collections.set_item("foo", schema).unwrap();
            let catalog = PyDict::new(py);
            catalog.set_item("test", collections).unwrap();
            catalog
        }

        fn foo_namespace() -> Namespace {
            Namespace {
                database: "test".into(),
                collection: "foo".into(),
            }
        }

        #[test]
        fn catalog_from_bare_schema() {
            Python::initialize();
            Python::attach(|py| {
                let catalog = py_dict_to_schemas(&catalog_dict(py, false)).unwrap();
                assert!(catalog.contains_key(&foo_namespace()));
            });
        }

        #[test]
        fn catalog_unwraps_json_schema_key() {
            Python::initialize();
            Python::attach(|py| {
                let bare = py_dict_to_schemas(&catalog_dict(py, false)).unwrap();
                let wrapped = py_dict_to_schemas(&catalog_dict(py, true)).unwrap();
                assert_eq!(bare, wrapped);
            });
        }

        #[test]
        fn catalog_rejects_non_dict_collections() {
            Python::initialize();
            Python::attach(|py| {
                let catalog = PyDict::new(py);
                catalog.set_item("test", "not a dict").unwrap();
                let err = py_dict_to_schemas(&catalog).unwrap_err();
                assert!(err.to_string().contains("'test'"), "{err}");
            });
        }

        #[test]
        fn catalog_rejects_invalid_bson_type() {
            Python::initialize();
            Python::attach(|py| {
                let schema = PyDict::new(py);
                schema.set_item("bsonType", "notAType").unwrap();
                let collections = PyDict::new(py);
                collections.set_item("foo", schema).unwrap();
                let catalog = PyDict::new(py);
                catalog.set_item("test", collections).unwrap();
                let err = py_dict_to_schemas(&catalog).unwrap_err();
                assert!(err.to_string().contains("test.foo"), "{err}");
            });
        }

        #[test]
        fn resolve_catalog_fills_every_namespace_when_absent() {
            Python::initialize();
            Python::attach(|py| {
                let catalog = resolve_catalog(
                    py,
                    "test",
                    "SELECT foo.a FROM foo JOIN bar ON foo.a = bar.c",
                    None,
                )
                .unwrap();
                let schema = catalog.get_schema_for_namespace(&foo_namespace()).unwrap();
                assert_eq!(schema, &MongoSqlSchema::Any);
                assert_eq!(
                    catalog.get_schema_for_namespace(&Namespace {
                        database: "test".into(),
                        collection: "bar".into(),
                    }),
                    Some(&MongoSqlSchema::Any)
                );
            });
        }

        #[test]
        fn resolve_catalog_fills_only_missing_namespaces() {
            Python::initialize();
            Python::attach(|py| {
                let provided = catalog_dict(py, false);
                let catalog = resolve_catalog(
                    py,
                    "test",
                    "SELECT foo.a FROM foo JOIN bar ON foo.a = bar.c",
                    Some(&provided),
                )
                .unwrap();
                // The provided schema is kept...
                assert_ne!(
                    catalog.get_schema_for_namespace(&foo_namespace()),
                    Some(&MongoSqlSchema::Any)
                );
                // ...and only the uncovered namespace is filled.
                assert_eq!(
                    catalog.get_schema_for_namespace(&Namespace {
                        database: "test".into(),
                        collection: "bar".into(),
                    }),
                    Some(&MongoSqlSchema::Any)
                );
            });
        }

        #[test]
        fn resolve_catalog_fills_nothing_without_collections() {
            Python::initialize();
            Python::attach(|py| {
                let catalog =
                    resolve_catalog(py, "test", "SELECT * FROM [{'a': 1}] arr", None).unwrap();
                assert!(catalog.get_schema_for_namespace(&foo_namespace()).is_none());
            });
        }

        #[test]
        fn translate_without_catalog() {
            Python::initialize();
            Python::attach(|py| {
                let t = MongoSqlTranslator::new()
                    .translate_query_verbose(py, "test", "SELECT * FROM [{'a': 1}] arr", None)
                    .unwrap();
                assert_eq!(t.target_db, "test");
                assert!(!t.pipeline.bind(py).is_empty());
            });
        }

        #[test]
        fn translate_with_catalog() {
            Python::initialize();
            Python::attach(|py| {
                let catalog = catalog_dict(py, true);
                let t = MongoSqlTranslator::new()
                    .translate_query(py, "test", "SELECT a FROM foo", Some(&catalog))
                    .unwrap();
                assert_eq!(t.target_collection.as_deref(), Some("foo"));
                // Namespaces are included, so each entry is [namespace, field];
                // an explicit select list carries the empty namespace.
                assert_eq!(t.select_order, vec![vec![String::new(), "a".to_string()]]);
            });
        }

        /// A collection missing from the catalog is filled with an empty schema
        /// rather than raising, matching `mongosql-cli`.
        #[test]
        fn translate_fills_missing_collection() {
            Python::initialize();
            Python::attach(|py| {
                let t = MongoSqlTranslator::new()
                    .translate_query(py, "test", "SELECT a FROM foo", None)
                    .unwrap();
                assert_eq!(t.target_collection.as_deref(), Some("foo"));
                assert!(!t.pipeline.bind(py).is_empty());
            });
        }

        #[test]
        fn translate_simple_matches_verbose_without_catalog() {
            Python::initialize();
            Python::attach(|py| {
                let translator = MongoSqlTranslator::new();
                let simple = translator
                    .translate_query_simple(py, "test", "SELECT a FROM foo")
                    .unwrap();
                let verbose = translator
                    .translate_query_verbose(py, "test", "SELECT a FROM foo", None)
                    .unwrap();
                assert_eq!(simple.target_collection, verbose.target_collection);
                assert_eq!(
                    simple.pipeline.bind(py).to_string(),
                    verbose.pipeline.bind(py).to_string()
                );
            });
        }
    }
}
