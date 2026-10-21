//! [`LocalDataService`] implementation backed by a PyMongo client.
//!
//! The service holds a Python object (a synchronous
//! [`pymongo.MongoClient`](https://pymongo.readthedocs.io/en/latest/api/pymongo/mongo_client.html))
//! and calls its methods through pyo3. All Python interaction happens while
//! holding the GIL via [`Python::attach`]; the schema-derivation futures that
//! drive this service are executed on the current thread (they are
//! non-`Send`), with the GIL released around the blocking work.

use bson::{Document, RawDocument};
use futures::Stream;
use pyo3::prelude::*;
use pyo3::types::PyAny;
use schema_builder_library::data_service::AggregateOptions;
use schema_builder_library::{CollectionInfo, LocalDataService};
use thiserror::Error;

/// Error type for [`PyMongoDataService`] operations.
#[derive(Debug, Error)]
pub enum PyMongoDataServiceError {
    /// A BSON serialization/deserialization failure while crossing the FFI boundary.
    #[error("Serialization error: {0}")]
    Serialization(String),
    /// The PyMongo client raised a Python exception, or the call failed.
    #[error("Query error: {0}")]
    Query(String),
    /// A PyMongo result could not be converted into the expected Rust type.
    #[error("Deserialization error: {0}")]
    Deserialization(String),
}

/// [`LocalDataService`] implementation backed by a synchronous PyMongo client.
///
/// The client is held as a [`Py`] handle and may be `None`-free: constructing
/// the service requires a real client object.
#[derive(Debug)]
pub struct PyMongoDataService {
    client: Py<PyAny>,
}

impl PyMongoDataService {
    /// Wraps an existing PyMongo client object.
    pub fn new(client: Py<PyAny>) -> Self {
        Self { client }
    }
}

/// Converts a Python object (typically a dict) to a Rust `bson::Document` by
/// calling PyMongo's `bson.encode`.
///
/// Must be called while holding the GIL.
fn py_object_to_bson_document(
    py: Python<'_>,
    obj: &Bound<'_, PyAny>,
) -> Result<Document, PyMongoDataServiceError> {
    let bytes: Vec<u8> = PyModule::import(py, "bson")
        .and_then(|m| m.getattr("encode"))
        .and_then(|encode| encode.call1((obj,)))
        .and_then(|result| result.extract())
        .map_err(|e| PyMongoDataServiceError::Serialization(e.to_string()))?;

    // NOTE: This re-parses the document, I really think we should try to find a way around this :/
    // For small N, I wouldn't consider this a problem,  but if generating a schema
    // from a collection this may cause problems
    let raw_document = RawDocument::from_bytes(&bytes)
        .map_err(|e| PyMongoDataServiceError::Deserialization(e.to_string()))?;
    Document::try_from(raw_document)
        .map_err(|e| PyMongoDataServiceError::Deserialization(e.to_string()))
}

/// Converts a Rust `bson::Document` to a Python object (a dict) by calling
/// PyMongo's `bson.decode` on its encoded bytes.
///
/// Must be called while holding the GIL.
fn bson_document_to_py_object<'py>(
    py: Python<'py>,
    document: &Document,
) -> Result<Bound<'py, PyAny>, PyMongoDataServiceError> {
    let mut bytes = Vec::new();
    document
        .to_writer(&mut bytes)
        .map_err(|e| PyMongoDataServiceError::Serialization(e.to_string()))?;

    PyModule::import(py, "bson")
        .and_then(|m| m.getattr("decode"))
        .and_then(|decode| decode.call1((bytes,)))
        .map_err(|e| PyMongoDataServiceError::Serialization(e.to_string()))
}

/// Deserializes a Python dict-like object into `T` via the BSON round-trip.
fn py_object_to_serde<T: serde::de::DeserializeOwned>(
    py: Python<'_>,
    obj: &Bound<'_, PyAny>,
) -> Result<T, PyMongoDataServiceError> {
    let document = py_object_to_bson_document(py, obj)?;
    bson::from_document::<T>(document)
        .map_err(|e| PyMongoDataServiceError::Deserialization(e.to_string()))
}

/// Runs a method on the held PyMongo client, converting Python errors.
///
/// Must be called while holding the GIL.
fn call_client_method<'py>(
    py: Python<'py>,
    client: &Py<PyAny>,
    method: &str,
    args: impl IntoIterator<Item = String>,
) -> Result<Bound<'py, PyAny>, PyMongoDataServiceError> {
    let bound = client.bind(py);
    let args: Vec<Bound<'py, PyAny>> = args
        .into_iter()
        .map(|name| pyo3::types::PyString::new(py, name.as_str()).into_any())
        .collect();
    let mut obj = bound.clone();
    for name in args {
        // MongoClient[db][coll] — subscript access avoids clashing with
        // client methods named like databases.
        obj = obj
            .get_item(&name)
            .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;
    }
    obj.call_method0(method)
        .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))
}

/// Wraps a Python iterator object into a `Stream` of documents.
///
/// Each `next()` polls the Python iterator while holding the GIL and converts
/// the returned dict into a `bson::Document`. Takes the iterator as an owned
/// [`Py`] handle so the returned stream captures no lifetimes.
fn py_cursor_to_stream(
    iterator: Py<PyAny>,
) -> Result<impl Stream<Item = Result<Document, PyMongoDataServiceError>>, PyMongoDataServiceError>
{
    Ok(futures::stream::try_unfold(
        iterator,
        |iterator| async move {
            Python::attach(|py| {
                match iterator.bind(py).call_method0("__next__") {
                    Ok(item) => {
                        let document = py_object_to_bson_document(py, &item)?;
                        Ok(Some((document, iterator)))
                    }
                    // StopIteration signals the end of the Python cursor.
                    Err(ref e) if e.is_instance_of::<pyo3::exceptions::PyStopIteration>(py) => {
                        Ok(None)
                    }
                    Err(e) => Err(PyMongoDataServiceError::Query(e.to_string())),
                }
            })
        },
    ))
}

/// Converts a Python cursor-like object into an owned iterator handle.
fn py_cursor_to_iterator(cursor: Bound<'_, PyAny>) -> Result<Py<PyAny>, PyMongoDataServiceError> {
    cursor
        .try_iter()
        .map(|it| it.into_any().unbind())
        .map_err(|e| PyMongoDataServiceError::Deserialization(e.to_string()))
}

impl LocalDataService for PyMongoDataService {
    type Error = PyMongoDataServiceError;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        Python::attach(|py| {
            let names = call_client_method(py, &self.client, "list_database_names", [])?;
            names
                .extract::<Vec<String>>()
                .map_err(|e| PyMongoDataServiceError::Deserialization(e.to_string()))
        })
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        Python::attach(|py| {
            let cursor =
                call_client_method(py, &self.client, "list_collections", [db_name.to_string()])?;
            let infos: Vec<CollectionInfo> = cursor
                .try_iter()
                .map_err(|e| PyMongoDataServiceError::Deserialization(e.to_string()))?
                .map(|row| {
                    let row = row.map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;
                    py_object_to_serde::<CollectionInfo>(py, &row)
                })
                .collect::<Result<Vec<_>, _>>()?;
            Ok(infos)
        })
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
        _options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        Python::attach(|py| {
            // Pipeline is Vec<bson::Document>; hand PyMongo a list of Python dicts.
            let pipeline_py: Vec<Bound<'_, PyAny>> = pipeline
                .iter()
                .map(|doc| bson_document_to_py_object(py, doc).map(|o| o.into_any()))
                .collect::<Result<_, _>>()?;

            let db = self
                .client
                .bind(py)
                .get_item(db_name)
                .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;
            let coll = db
                .get_item(coll_name)
                .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;
            let cursor = coll
                .call_method1("aggregate", (pipeline_py,))
                .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;

            py_cursor_to_stream(py_cursor_to_iterator(cursor)?)
        })
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        Python::attach(|py| {
            let filter_py = bson_document_to_py_object(py, &filter)?;

            let db = self
                .client
                .bind(py)
                .get_item(db_name)
                .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;
            let coll = db
                .get_item(coll_name)
                .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;
            let cursor = coll
                .call_method1("find", (filter_py,))
                .map_err(|e| PyMongoDataServiceError::Query(e.to_string()))?;

            py_cursor_to_stream(py_cursor_to_iterator(cursor)?)
        })
    }
}
