mod pymongo_data_service;

use pyo3::prelude::*;

/// A Python module implemented in Rust.
#[pymodule]
mod py_schema_builder_library {
    use crate::pymongo_data_service::PyMongoDataService;
    use bson::Document;
    use pyo3::prelude::*;
    use pyo3::types::{PyDict, PyList};
    use std::future::Future;

    const MAX_SAMPLE_SIZE_HARD_LIMIT: u32 = 50_000;

    /// Converts BSON-encoded bytes (as produced by Python's `bson.encode`)
    /// into a Rust `bson::Document`.
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

    /// Formats the sum of two numbers as string.
    #[pyfunction]
    fn sum_as_string(a: usize, b: usize) -> PyResult<String> {
        Ok((a + b).to_string())
    }

    /// Runs a non-`Send` schema-derivation future to completion on the current
    /// thread.
    ///
    /// The futures produced by `schema-builder-library` (generic over the
    /// `LocalDataService` local variant) are not `Send`, so they must never run
    /// on `tokio::spawn` or a multi-threaded runtime — a fresh current-thread
    /// runtime per call is the simplest correct harness for a POC.
    ///
    /// The GIL is intentionally NOT released around `block_on`: the future's
    /// output crosses back into Python and the service re-enters Python on
    /// every I/O call, so `Python::detach`'s `Ungil` bounds don't hold here.
    // TODO: Release the GIL around network I/O once the service isolates its
    // Python calls from the Rust state (would need a Send-able output type).
    fn run_current_thread<F: Future>(fut: F) -> PyResult<F::Output> {
        let runtime = tokio::runtime::Builder::new_current_thread()
            .enable_all()
            .build()
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(e.to_string()))?;
        Ok(runtime.block_on(fut))
    }

    /// Converts a derived `Schema` into a Python dict via the `$jsonSchema`
    /// document and PyMongo's `bson.decode`.
    fn schema_to_py_dict(
        py: Python<'_>,
        schema: schema_builder_library::MongoSqlSchema,
    ) -> PyResult<Py<PyDict>> {
        // Schema -> $jsonSchema BSON document (TryFrom impl in mongosql).
        let json_schema = bson::Document::try_from(schema).map_err(|e| {
            PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                "Failed to convert Schema to jsonSchema: {}",
                e
            ))
        })?;

        let mut schema_bytes = Vec::new();
        json_schema
            .to_writer(&mut schema_bytes)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(e.to_string()))?;

        let bson_module = PyModule::import(py, "bson")?;
        let decoded = bson_module
            .getattr("decode")?
            .call1((schema_bytes,))?
            .extract::<Bound<'_, PyDict>>()?;
        Ok(decoded.unbind())
    }

    /// Base class for SchemaBuilder
    #[pyclass]
    struct SchemaBuilder {
        client: Option<Py<PyAny>>,
    }

    #[pymethods]
    impl SchemaBuilder {
        // `client` defaults to None so `SchemaBuilder()` works for
        // document-level use; collection-level methods raise ValueError
        // on a client-less builder.
        #[new]
        #[pyo3(signature = (client = None))]
        fn __new__(client: Option<Bound<'_, PyAny>>) -> Self {
            SchemaBuilder {
                client: client.map(|c| c.unbind().into_any()),
            }
        }

        /// Example method for SchemaBuilder
        /// The Python Code accepts a dictionary. Per the pyo3 documentation, we use a PyDict to represent that:
        /// https://pyo3.rs/v0.30.0/types#concrete-python-types
        #[pyo3(signature = (document))]
        fn build_schema_for_document(&self, document: Bound<'_, PyDict>) -> PyResult<Py<PyDict>> {
            // We are inside a pyo3 call, so the GIL is already held and the
            // interpreter is initialized — use the token from the argument
            // directly rather than Python::attach.
            let py = document.py();
            // Call PyMongo's bson.encode(dict) -> bytes from Rust.
            // TBD: I'm not sure how we verify that whereever this code is run, that it has PyMongo available.
            // Is there a way to do a requirements.txt or something to say a library is needed? Not sure - will need to talk with python team or read packaging docs.
            let bson_module = PyModule::import(py, "bson")?;
            let bytes: Vec<u8> = bson_module
                .getattr("encode")?
                .call1((document,))?
                .extract()?;
            self.build_schema_for_document_bytes(bytes)
        }

        #[pyo3(signature = (documents))]
        fn build_schema_for_documents(&self, documents: Bound<'_, PyList>) -> PyResult<Py<PyDict>> {
            let py = documents.py();
            let bson_module = PyModule::import(py, "bson")?;

            // 1. Map the documents to bson bytes
            // a. Get an iterator for the PyList
            let mut documents_as_bytes: Vec<bson::Document> = vec![];
            let mut documents_iterator = documents.iter();

            // b. Iterator through the documents, converting to bytes and append to the list
            while let Some(document) = documents_iterator.next() {
                let document_as_bytes: Vec<u8> = bson_module
                    .getattr("encode")?
                    .call1((document,))?
                    .extract()?;

                let bytes_as_rust_bson_document = bytes_to_document(&document_as_bytes);
                if let Ok(converted_document) = bytes_as_rust_bson_document {
                    documents_as_bytes.push(converted_document);
                } else {
                    // TODO: Do something with error. Throw saying one of the documents was malformed
                    // And could not interface with rust.
                }
            }

            // 2. call schema::derivation on the list of documents
            let derived_schema = schema_derivation::schema_for_documents(&documents_as_bytes);

            // 3. Convert back to a Python Dict
            // Schema -> $jsonSchema BSON document (TryFrom impl in mongosql).
            // bson::Document is Serialize, so serde_json can print it.
            // NOTE: This is basically copy-oaste of the other method. Once we get this working
            // We can consolidate a lot of the shared logic between these into helper conversion
            // methods
            let json_schema = bson::Document::try_from(derived_schema).map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to convert Schema to jsonSchema: {}",
                    e
                ))
            })?;

            let mut derived_schema_byte_array: Vec<u8> = vec![];
            json_schema
                .to_writer(&mut derived_schema_byte_array)
                .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(e.to_string()))?;

            // TODO: Revisit the error handling. This is very rough
            let schema_as_python_dictionary: PyResult<Py<PyDict>> = Python::attach(|py| {
                let bson_module =
                    PyModule::import(py, "bson").expect("Unable to import bson module");
                let decoded_to_python_dict: Bound<'_, PyDict> = bson_module
                    .getattr("decode")
                    .expect("Unable to find method decode")
                    .call1((derived_schema_byte_array,))
                    .expect("Could not derive schema from bytes")
                    .extract()
                    .expect("This should work");

                // Bound<T> implements Display via Python's str(), so this prints the dict.
                println!("Schema as Python dictionary: {}", decoded_to_python_dict);
                Ok(decoded_to_python_dict.unbind())
            });

            schema_as_python_dictionary
        }

        /// Builds a schema from an already BSON-encoded document.
        /// Accepts the raw bytes that Python's `bson.encode` produces. Processes the schema in rust
        /// Then decodes the bytes back into a dictionary
        // NOTE: We should consider if we can return a Document, or maybe the raw bytes to decode?
        // Not sure cost of converting back to a PyDict? But we can probably go backwards. Schema -> bytes -> PyMongo.encode()
        #[pyo3(signature = (document_bytes))]
        fn build_schema_for_document_bytes(&self, document_bytes: Vec<u8>) -> PyResult<Py<PyDict>> {
            let document = bytes_to_document(&document_bytes)?;
            let derived_schema = schema_derivation::schema_for_document(&document);
            // Schema -> $jsonSchema BSON document (TryFrom impl in mongosql).
            // bson::Document is Serialize, so serde_json can print it.
            let json_schema = bson::Document::try_from(derived_schema).map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to convert Schema to jsonSchema: {}",
                    e
                ))
            })?;

            let mut derived_schema_byte_array: Vec<u8> = vec![];
            json_schema
                .to_writer(&mut derived_schema_byte_array)
                .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(e.to_string()))?;

            // TODO: Revisit the error handling. This is very rough
            let schema_as_python_dictionary: PyResult<Py<PyDict>> = Python::attach(|py| {
                let bson_module =
                    PyModule::import(py, "bson").expect("Unable to import bson module");
                let decoded_to_python_dict: Bound<'_, PyDict> = bson_module
                    .getattr("decode")
                    .expect("Unable to find method decode")
                    .call1((derived_schema_byte_array,))
                    .expect("Could not derive schema from bytes")
                    .extract()
                    .expect("This should work");

                // Bound<T> implements Display via Python's str(), so this prints the dict.
                println!("Schema as Python dictionary: {}", decoded_to_python_dict);
                Ok(decoded_to_python_dict.unbind())
            });

            schema_as_python_dictionary
        }

        fn build_schema_for_collection(
            &self,
            py: Python<'_>,
            db_name: &str,
            collection_name: &str,
        ) -> PyResult<Py<PyDict>> {
            let client = self.client.as_ref().ok_or_else(|| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(
                    "SchemaBuilder was constructed without a PyMongo client; \
                     a client is required for collection-level operations. \
                     Pass one when constructing: SchemaBuilder(pymongo.MongoClient(...))",
                )
            })?;

            let service = PyMongoDataService::new(client.clone_ref(py));
            let schema = run_current_thread(schema_builder_library::derive_schema_for_collection(
                &service,
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

        fn sample_schema_for_collection(
            &self,
            py: Python<'_>,
            db_name: &str,
            collection_name: &str,
            sample_size: u32,
        ) -> PyResult<Py<PyDict>> {
            let client = self.client.as_ref().ok_or_else(|| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(
                    "SchemaBuilder was constructed without a PyMongo client; \
                     a client is required for collection-level operations. \
                     Pass one when constructing: SchemaBuilder(pymongo.MongoClient(...))",
                )
            })?;
            let sample_size = std::cmp::min(sample_size, MAX_SAMPLE_SIZE_HARD_LIMIT)
                .try_into()
                .map_err(|e| {
                    PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                        "Invalid sample size {sample_size}: {e}"
                    ))
                })?;

            let service = PyMongoDataService::new(client.clone_ref(py));
            let schema = run_current_thread(schema_builder_library::sample_schema_for_collection(
                &service,
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
}
