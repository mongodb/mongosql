mod py_bson;

use pyo3::prelude::*;

/// A Python module implemented in Rust.
#[pymodule]
mod py_schema_builder_library {
    use crate::py_bson::{document_to_pydict, py_dict_to_document};
    use bson::Document;
    use pyo3::prelude::*;
    use pyo3::types::{PyDict, PyList};
    use schema_builder_library::SynchronousMongoDbDataService;
    use std::future::Future;

    const MAX_SAMPLE_SIZE_HARD_LIMIT: u32 = 50_000;

    /// Default MongoDB URI when `MongoConnectionProperties.uri` is not set.
    ///
    /// Matches the default used by `mongosql-cli` (`mongodb://localhost:27017`),
    /// which is the conventional local-single-server default.
    const DEFAULT_MONGODB_URI: &str = "mongodb://localhost:27017";

    /// Connection settings for the embedded Rust MongoDB driver client.
    ///
    /// Passed to `SchemaBuilder` for collection-level operations. TLS,
    /// read preference, replica set, and other advanced options ride in the
    /// `uri` connection string (standard MongoDB connection-string options).
    #[pyclass]
    struct MongoConnectionProperties {
        #[pyo3(get, set)]
        uri: Option<String>,
        #[pyo3(get, set)]
        username: Option<String>,
        #[pyo3(get, set)]
        password: Option<String>,
    }

    #[pymethods]
    impl MongoConnectionProperties {
        #[new]
        #[pyo3(signature = (uri = None, username = None, password = None))]
        fn __new__(
            uri: Option<String>,
            username: Option<String>,
            password: Option<String>,
        ) -> Self {
            Self {
                uri,
                username,
                password,
            }
        }
    }

    /// Converts a derived `Schema` into a Python dict via the `$jsonSchema`
    /// document (pure-Rust BSON conversion).
    fn schema_to_py_dict(
        py: Python<'_>,
        schema: schema_builder_library::MongoSqlSchema,
    ) -> PyResult<Py<PyDict>> {
        // Schema -> $jsonSchema BSON document (TryFrom impl in mongosql).
        let json_schema = bson::Document::try_from(schema).map_err(|e| {
            PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                "Failed to convert Schema to jsonSchema: {e}"
            ))
        })?;
        Ok(document_to_pydict(py, &json_schema)?.unbind())
    }

    /// Runs a non-`Send` schema-derivation future to completion on the current
    /// thread.
    ///
    /// The futures produced by `schema-builder-library` (generic over the
    /// `LocalDataService` local variant) are not `Send`, so they must never run
    /// on `tokio::spawn` or a multi-threaded runtime. With the synchronous data
    /// service no await point ever actually suspends (all I/O is blocking), so
    /// `futures::executor::block_on` is sufficient — and mandatory: a tokio
    /// `block_on` would put this thread in a runtime context, and the
    /// `mongodb::sync` driver's internal runtime would then panic with
    /// "Cannot start a runtime from within a runtime".
    ///
    /// The GIL is intentionally NOT released around `block_on`: the future is
    /// `!Send`, which fails `Python::detach`'s `Ungil` bounds. No I/O re-enters
    /// Python, so this is a plain blocking call for the Python thread.
    fn run_current_thread<F: Future>(fut: F) -> PyResult<F::Output> {
        Ok(futures::executor::block_on(fut))
    }

    /// Base class for SchemaBuilder
    #[pyclass]
    struct SchemaBuilder {
        // Embedded synchronous data service over the Rust driver client; `None`
        // restricts the builder to document-level methods (collection-level
        // calls raise ValueError). Created once at construction and shared by
        // every collection-level call. The service (and `mongodb::sync::Client`
        // beneath it) is `Send + Sync` and cheap to clone (Arc-backed internals).
        data_service: Option<SynchronousMongoDbDataService>,
    }

    impl SchemaBuilder {
        /// Builds the driver client from connection properties (fully synchronous:
        /// `ClientOptions::parse` runs on a throwaway current-thread runtime
        /// because it performs DNS resolution).
        fn build_client(properties: &MongoConnectionProperties) -> PyResult<mongodb::sync::Client> {
            let uri = properties
                .uri
                .clone()
                .unwrap_or_else(|| DEFAULT_MONGODB_URI.to_string());
            // NOTE: `ClientOptions::parse` returns an action type that completes
            // via `.await` (IntoFuture, not a bare Future) — the async block
            // bridges it into `run_current_thread`.
            let mut options = run_current_thread(async {
                mongodb::options::ClientOptions::parse(uri.as_str()).await
            })?
            .map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Invalid MongoDB connection properties (uri '{uri}'): {e}"
                ))
            })?;

            if properties.username.is_some() || properties.password.is_some() {
                options.credential = Some(
                    mongodb::options::Credential::builder()
                        .username(properties.username.clone())
                        .password(properties.password.clone())
                        .build(),
                );
            }

            mongodb::sync::Client::with_options(options).map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to create MongoDB client: {e}"
                ))
            })
        }
    }

    #[pymethods]
    impl SchemaBuilder {
        // `connection_properties` defaults to None so `SchemaBuilder()` works for
        // document-level use; collection-level methods raise ValueError on a
        // client-less builder.
        #[new]
        #[pyo3(signature = (connection_properties = None))]
        fn __new__(
            connection_properties: Option<Bound<'_, MongoConnectionProperties>>,
        ) -> PyResult<Self> {
            let data_service = match connection_properties {
                Some(props) => {
                    let borrowed = props.try_borrow()?;
                    Some(SynchronousMongoDbDataService::new(Self::build_client(
                        &borrowed,
                    )?))
                }
                None => None,
            };
            Ok(SchemaBuilder { data_service })
        }

        /// Builds a schema from a Python dictionary.
        #[pyo3(signature = (document))]
        fn build_schema_for_document(&self, document: Bound<'_, PyDict>) -> PyResult<Py<PyDict>> {
            let py = document.py();
            // Python dict -> bson::Document, entirely in Rust.
            let document = py_dict_to_document(&document)?;
            let derived_schema = schema_derivation::schema_for_document(&document);
            schema_to_py_dict(py, derived_schema)
        }

        /// Builds a schema from a list of Python dictionaries, widening field
        /// types across documents.
        #[pyo3(signature = (documents))]
        fn build_schema_for_documents(&self, documents: Bound<'_, PyList>) -> PyResult<Py<PyDict>> {
            let py = documents.py();

            // 1. Map the documents to Rust bson Documents (in Rust; no FFI
            //    round-trips). Malformed documents raise ValueError.
            let documents_as_documents: Vec<Document> = documents
                .iter()
                .map(|doc| -> Result<Document, PyErr> {
                    let dict = doc.cast_into::<PyDict>().map_err(|cast_err| {
                        PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                            "Expected a dict, got {cast_err}"
                        ))
                    })?;
                    py_dict_to_document(&dict)
                })
                .collect::<Result<Vec<_>, _>>()?;

            // 2. call schema::derivation on the list of documents
            let derived_schema = schema_derivation::schema_for_documents(&documents_as_documents);

            // 3. Convert back to a Python Dict
            schema_to_py_dict(py, derived_schema)
        }

        /// Builds a schema from an already BSON-encoded document.
        /// Accepts the raw bytes that `bson` serialization produces. Processes the schema in rust
        /// Then decodes the bytes back into a dictionary
        #[pyo3(signature = (document_bytes))]
        fn build_schema_for_document_bytes(&self, document_bytes: Vec<u8>) -> PyResult<Py<PyDict>> {
            Python::attach(|py| {
                let document = bytes_to_document(&document_bytes)?;
                let derived_schema = schema_derivation::schema_for_document(&document);
                schema_to_py_dict(py, derived_schema)
            })
        }

        /// Builds a schema for every document in a collection, connecting to
        /// MongoDB with the embedded Rust driver.
        fn build_schema_for_collection(
            &self,
            py: Python<'_>,
            db_name: &str,
            collection_name: &str,
        ) -> PyResult<Py<PyDict>> {
            let service = self.data_service.as_ref().ok_or_else(|| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(
                    "SchemaBuilder was constructed without connection properties; \
                     a MongoDB connection is required for collection-level operations. \
                     Pass one when constructing: SchemaBuilder(MongoConnectionProperties(...))",
                )
            })?;

            let schema = run_current_thread(schema_builder_library::derive_schema_for_collection(
                service,
                db_name,
                collection_name,
                None,
            ))?
            .map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to derive schema for {db_name}.{collection_name}: {e}"
                ))
            })?;

            schema_to_py_dict(py, schema)
        }

        /// Builds a schema from a random sample of documents in a collection,
        /// connecting to MongoDB with the embedded Rust driver.
        fn sample_schema_for_collection(
            &self,
            py: Python<'_>,
            db_name: &str,
            collection_name: &str,
            sample_size: u32,
        ) -> PyResult<Py<PyDict>> {
            let service = self.data_service.as_ref().ok_or_else(|| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(
                    "SchemaBuilder was constructed without connection properties; \
                     a MongoDB connection is required for collection-level operations. \
                     Pass one when constructing: SchemaBuilder(MongoConnectionProperties(...))",
                )
            })?;
            let sample_size = std::cmp::min(sample_size, MAX_SAMPLE_SIZE_HARD_LIMIT)
                .try_into()
                .map_err(|e| {
                    PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                        "Invalid sample size {sample_size}: {e}"
                    ))
                })?;

            let schema = run_current_thread(schema_builder_library::sample_schema_for_collection(
                service,
                db_name,
                collection_name,
                sample_size,
            ))?
            .map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to sample schema for {db_name}.{collection_name}: {e}"
                ))
            })?;

            schema_to_py_dict(py, schema)
        }
    }

    /// Converts BSON-encoded bytes into a Rust `bson::Document`.
    ///
    /// Internal helper; not exposed to Python. Lives outside the
    /// `#[pymethods]` block because pyo3 requires every method's parameters
    /// and return type to be convertible across the Python boundary, and
    /// `bson::Document` has no Python representation.
    fn bytes_to_document(document_bytes: &[u8]) -> Result<Document, PyErr> {
        // Turn the bytes into a rust bson RawDocument
        let raw_document = bson::RawDocument::from_bytes(document_bytes).map_err(|e| {
            PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                "Failed to create RawDocument: {e}"
            ))
        })?;
        // (Use the RawDocument to generate a schema) (or convert RowDocument to Real Document
        // NOTE: We should probably re-implement the schema stuff against a raw document instead of parsing again
        // Reference: https://patrickfreed.github.io/rust/2022/04/27/unlocking-greater-performance-in-the-mongodb-rust-driver-via-raw-bson-and-zero-copy-deserialization.html
        let document = bson::Document::try_from(raw_document).map_err(|e| {
            PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                "Failed to convert RawDocument to Document: {e}"
            ))
        })?;
        Ok(document)
    }
}
