use pyo3::prelude::*;

/// A Python module implemented in Rust.
#[pymodule]
mod py_schema_builder_library {
    use bson::{Document, doc, to_bson};
    use pyo3::prelude::*;
    use pyo3::types::PyDict;

    const MAX_SAMPLE_SIZE_HARD_LIMIT: u32 = 50_000;
    /// Formats the sum of two numbers as string.
    #[pyfunction]
    fn sum_as_string(a: usize, b: usize) -> PyResult<String> {
        Ok((a + b).to_string())
    }

    /// Base class for SchemaBuilder
    #[pyclass]
    struct SchemaBuilder {}

    #[pymethods]
    impl SchemaBuilder {
        #[new]
        fn __new__() -> Self {
            SchemaBuilder {}
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

        /// Builds a schema from an already BSON-encoded document.
        /// Accepts the raw bytes that Python's `bson.encode` produces. Processes the schema in rust
        /// Then decodes the bytes back into a dictionary
        // NOTE: We should consider if we can return a Document, or maybe the raw bytes to decode?
        // Not sure cost of converting back to a PyDict? But we can probably go backwards. Schema -> bytes -> PyMongo.encode()
        #[pyo3(signature = (document_bytes))]
        fn build_schema_for_document_bytes(&self, document_bytes: Vec<u8>) -> PyResult<Py<PyDict>> {
            // Turn the bytes into a rust bson RawDocument
            let raw_document = bson::RawDocument::from_bytes(&document_bytes).map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to create RawDocument: {}",
                    e
                ))
            })?;
            // (Use the RawDocument to generate a schema) (or convert RowDocument to Real Document
            // NOTE: We should probably re-implement the schema stuff against a raw document instead of parsing again
            // Reference: https://patrickfreed.github.io/rust/2022/04/27/unlocking-greater-performance-in-the-mongodb-rust-driver-via-raw-bson-and-zero-copy-deserialization.html
            let document = bson::Document::try_from(raw_document).map_err(|e| {
                PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                    "Failed to convert RawDocument to Document: {}",
                    e
                ))
            })?;
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
            collection_name: &str,
            db_name: &str,
        ) -> PyResult<String> {
            // TBD: Implement logic to build schema for a collection
            Ok(format!(
                "Generated schema for collection: {} in database: {}",
                collection_name, db_name
            ))
        }

        fn sample_schema_for_collection(
            &self,
            collection_name: &str,
            db_name: &str,
            sample_size: u32,
        ) -> PyResult<String> {
            // TBD: Implement logic to build schema for a collection with a sample size
            let sample_size = std::cmp::min(sample_size, MAX_SAMPLE_SIZE_HARD_LIMIT);
            let sampling_pipeline: Vec<Document> = vec![doc! {"$sample": sample_size}];
            // TODO: Add some error handling in case sample fails. Attempt to sample at least 3 times

            Ok(format!(
                "Generated schema for collection: {} in database: {} with sample size: {}",
                collection_name, db_name, sample_size
            ))
        }
    }
}
