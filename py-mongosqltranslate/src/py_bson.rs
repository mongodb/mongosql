//! Pure-Rust conversion between Python objects and `bson::Document`.
//!
//! Copied from `py-schema-builder-library/src/py_bson.rs`. Calling `PyMongo`'s
//! `bson.encode` / `bson.decode` across the FFI is avoided, so the crate has no
//! `PyMongo` runtime dependency.
//!
//! Conversion notes (documented degradations, mirrored by the unit tests):
//! - Python `int` maps to `Bson::Int64`; narrow `Int32` is only produced by
//!   schema derivation and converts back to a plain Python int.
//! - `Bson::ObjectId` converts to/from the hex string form (no `PyMongo`
//!   `ObjectId` class exists here).
//! - Python `datetime.datetime` maps to/from `Bson::DateTime` (naive datetimes
//!   are interpreted as local time, matching `datetime.timestamp()` semantics).
//! - `Bson::Decimal128`, `Bson::JavaScriptCode`, and other exotic BSON types are
//!   stringified rather than rejected, so schemas containing them round-trip.

use std::time::{SystemTime, UNIX_EPOCH};

use bson::{Bson, Document};
use pyo3::PyTypeInfo;
use pyo3::conversion::IntoPyObjectExt;
use pyo3::prelude::*;
use pyo3::types::{PyAny, PyBytes, PyDateTime, PyDict, PyList};

/// Converts a Python dict into a `bson::Document`.
///
/// Nested dicts, lists, tuples, and scalar types are converted recursively;
/// unsupported Python types raise `PyValueError` naming the type.
pub fn py_dict_to_document(dict: &Bound<'_, PyDict>) -> Result<Document, PyErr> {
    let mut document = Document::new();
    for (key, value) in dict.iter() {
        let key: String = key.extract().map_err(|_| {
            PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
                "Document keys must be strings, got {key}"
            ))
        })?;
        document.insert(key, py_to_bson(&value)?);
    }
    Ok(document)
}

/// Converts a `bson::Document` into a fresh Python dict.
pub fn document_to_pydict<'py>(
    py: Python<'py>,
    document: &Document,
) -> Result<Bound<'py, PyDict>, PyErr> {
    let dict = PyDict::new(py);
    for (key, value) in document {
        dict.set_item(key, bson_to_py(py, value)?)?;
    }
    Ok(dict)
}

fn py_to_bson(value: &Bound<'_, PyAny>) -> Result<Bson, PyErr> {
    // bool must be checked before int (bool is a subclass of int in Python).
    if let Ok(v) = value.extract::<bool>() {
        return Ok(Bson::Boolean(v));
    }
    if let Ok(v) = value.extract::<i64>() {
        return Ok(Bson::Int64(v));
    }
    if let Ok(v) = value.extract::<f64>() {
        return Ok(Bson::Double(v));
    }
    if let Ok(v) = value.extract::<String>() {
        return Ok(Bson::String(v));
    }
    if value.is_none() {
        return Ok(Bson::Null);
    }
    if let Ok(dict) = value.cast::<PyDict>() {
        return Ok(Bson::Document(py_dict_to_document(dict)?));
    }
    if let Ok(list) = value.cast::<PyList>() {
        let mut array = Vec::with_capacity(list.len());
        for item in list.iter() {
            array.push(py_to_bson(&item)?);
        }
        return Ok(Bson::Array(array));
    }
    if let Ok(bytes) = value.cast::<PyBytes>() {
        return Ok(Bson::Binary(bson::Binary {
            bytes: bytes.as_bytes().to_vec(),
            subtype: bson::spec::BinarySubtype::Generic,
        }));
    }
    // datetime.datetime (checked via type identity; bool/int/str checks above
    // have already excluded the other datetime components).
    if value.cast::<PyDateTime>().is_ok() {
        // `timestamp()` returns epoch seconds as float for naive (interpreted as
        // local time) and aware datetimes alike.
        let epoch_secs: f64 = value.call_method0("timestamp")?.extract()?;
        #[expect(
            clippy::cast_possible_truncation,
            reason = "epoch millis of any Python datetime fit in i64; `as` saturates on overflow"
        )]
        let millis = (epoch_secs * 1_000.0).round() as i64;
        return Ok(Bson::DateTime(bson::DateTime::from_millis(millis)));
    }

    Err(PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
        "Unsupported document value type '{}'; use str, int, float, bool, None, \
         dict, list, bytes, or datetime.datetime",
        value
            .get_type()
            .qualname()
            .map(|n| n.to_string())
            .unwrap_or_default()
    )))
}

fn bson_to_py(py: Python<'_>, bson: &Bson) -> Result<Py<PyAny>, PyErr> {
    let value: Py<PyAny> = match bson {
        Bson::Double(v) => v.into_py_any(py)?,
        Bson::String(v) => v.into_py_any(py)?,
        Bson::Boolean(v) => v.into_py_any(py)?,
        Bson::Null => py.None(),
        Bson::Int32(v) => v.into_py_any(py)?,
        Bson::Int64(v) => v.into_py_any(py)?,
        Bson::Document(d) => document_to_pydict(py, d)?.unbind().into_any(),
        Bson::Array(items) => {
            let mut converted = Vec::with_capacity(items.len());
            for item in items {
                converted.push(bson_to_py(py, item)?);
            }
            PyList::new(py, converted)?.unbind().into_any()
        }
        Bson::Binary(b) => b.bytes.clone().into_py_any(py)?,
        Bson::ObjectId(oid) => {
            // No PyMongo ObjectId class available; use the hex string form.
            oid.to_hex().into_py_any(py)?
        }
        Bson::DateTime(dt) => py_datetime_from_bson(py, *dt)?,
        // Exotic BSON types are stringified so schemas containing them still
        // round-trip to Python callers.
        other => other.to_string().into_py_any(py)?,
    };
    Ok(value)
}

fn py_datetime_from_bson(py: Python<'_>, dt: bson::DateTime) -> Result<Py<PyAny>, PyErr> {
    let millis = system_time_millis(dt.to_system_time())?;
    let py_type = PyDateTime::type_object(py);
    #[expect(
        clippy::cast_precision_loss,
        reason = "Python timestamps are floats; millis below 2^53 (year ~287,000) convert exactly"
    )]
    let epoch_secs = (millis as f64) / 1_000.0;
    Ok(py_type
        .call_method1("fromtimestamp", (epoch_secs,))?
        .unbind())
}

fn system_time_millis(time: SystemTime) -> Result<i64, PyErr> {
    let duration = time.duration_since(UNIX_EPOCH).map_err(|e| {
        PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
            "BSON date predates the Unix epoch and has no Python representation: {e}"
        ))
    })?;
    i64::try_from(duration.as_millis()).map_err(|e| {
        PyErr::new::<pyo3::exceptions::PyValueError, _>(format!(
            "BSON date out of range for Python datetime: {e}"
        ))
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_py(py: Python<'_>) -> Bound<'_, PyDict> {
        let dict = PyDict::new(py);
        dict.set_item("s", "hello").unwrap();
        dict.set_item("i", 42_i64).unwrap();
        dict.set_item("f", 1.5_f64).unwrap();
        dict.set_item("b", true).unwrap();
        dict.set_item("n", Option::<i64>::None).unwrap();

        let nested = PyDict::new(py);
        nested.set_item("k", "v").unwrap();
        dict.set_item("d", nested).unwrap();

        let list = PyList::new(py, [1_i64, 2_i64, 3_i64]).unwrap();
        dict.set_item("a", list).unwrap();

        let bytes = PyBytes::new(py, b"\x01\x02");
        dict.set_item("bin", bytes).unwrap();

        let dt = PyDateTime::new(py, 2026, 10, 20, 14, 30, 0, 0, None).unwrap();
        dict.set_item("date", dt).unwrap();

        dict
    }

    #[test]
    fn round_trip() {
        Python::initialize();
        Python::attach(|py| {
            let original = test_py(py);
            let document = py_dict_to_document(&original).unwrap();

            assert_eq!(document.get_str("s"), Ok("hello"));
            assert_eq!(document.get_i64("i"), Ok(42));
            assert_eq!(document.get_f64("f"), Ok(1.5));
            assert_eq!(document.get_bool("b"), Ok(true));
            assert_eq!(document.get("n"), Some(&Bson::Null));
            assert_eq!(
                document.get_document("d").and_then(|d| d.get_str("k")),
                Ok("v")
            );
            assert_eq!(
                document
                    .get_array("a")
                    .unwrap()
                    .iter()
                    .map(Bson::as_i64)
                    .collect::<Vec<_>>(),
                vec![Some(1), Some(2), Some(3)]
            );
            assert!(matches!(document.get("bin"), Some(Bson::Binary(_))));
            assert!(matches!(document.get("date"), Some(Bson::DateTime(_))));

            let back = document_to_pydict(py, &document).unwrap();
            let round_tripped = py_dict_to_document(&back).unwrap();
            assert_eq!(document, round_tripped);
        });
    }

    #[test]
    fn unsupported_type_raises() {
        Python::initialize();
        Python::attach(|py| {
            let dict = PyDict::new(py);
            let set = pyo3::types::PySet::empty(py).unwrap();
            dict.set_item("bad", set).unwrap();
            let err = py_dict_to_document(&dict).unwrap_err();
            assert!(err.to_string().contains("Unsupported document value type"));
        });
    }

    #[test]
    fn object_id_maps_to_hex_string() {
        Python::initialize();
        Python::attach(|py| {
            let oid = bson::oid::ObjectId::from_bytes([7u8; 12]);
            let mut document = Document::new();
            document.insert("id", Bson::ObjectId(oid));
            let dict = document_to_pydict(py, &document).unwrap();
            let hex: String = dict.get_item("id").unwrap().unwrap().extract().unwrap();
            assert_eq!(hex, oid.to_hex());
        });
    }

    #[test]
    fn datetime_values_survive_round_trip() {
        Python::initialize();
        Python::attach(|py| {
            // BSON dates are UTC millis; compare epoch millis through
            // `timestamp()` rather than wall-clock repr.
            let bson_dt = bson::DateTime::from_millis(1_792_000_000_000);
            let mut document = Document::new();
            document.insert("t", Bson::DateTime(bson_dt));
            let dict = document_to_pydict(py, &document).unwrap();
            let py_dt = dict.get_item("t").unwrap();
            let epoch_secs: f64 = py_dt
                .unwrap()
                .call_method0("timestamp")
                .unwrap()
                .extract()
                .unwrap();
            #[expect(
                clippy::cast_possible_truncation,
                reason = "test timestamp is well within i64 range"
            )]
            let millis = (epoch_secs * 1_000.0).round() as i64;
            assert_eq!(millis, 1_792_000_000_000);
        });
    }

    #[test]
    fn non_string_key_raises() {
        Python::initialize();
        Python::attach(|py| {
            let dict = PyDict::new(py);
            dict.set_item(1_i64, "v").unwrap();
            assert!(py_dict_to_document(&dict).is_err());
        });
    }

    #[test]
    fn exotic_bson_stringified() {
        Python::initialize();
        Python::attach(|py| {
            let mut document = Document::new();
            document.insert(
                "dec",
                Bson::Decimal128(bson::Decimal128::from_bytes([0u8; 16])),
            );
            let dict = document_to_pydict(py, &document).unwrap();
            let repr: String = dict.get_item("dec").unwrap().unwrap().extract().unwrap();
            assert!(!repr.is_empty());
        });
    }
}
