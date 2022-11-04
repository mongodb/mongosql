use mongosql::{
    catalog::{self, Catalog},
    json_schema, schema, SchemaCheckingMode,
};
use std::{
    collections::{BTreeMap, BTreeSet},
    ffi::{CStr, CString, NulError},
    os::raw,
    panic,
    panic::UnwindSafe,
    string::FromUtf8Error,
    sync::mpsc,
};

mod version;

/// Returns the semantic version of this library as a C string.
/// The caller is responsible for freeing the returned value.
#[no_mangle]
pub extern "C" fn version() -> *mut raw::c_char {
    to_raw_c_string(version::VERSION).expect("semver string contained NUL byte")
}

/// Returns a base64-encoded bson representation of
/// [Translation](/mongosql/struct.Translation.html) for the provided
/// SQL query, database, catalog schema, and schema checking mode.
#[no_mangle]
pub extern "C" fn translate(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
    catalog: *const libc::c_char,
    relax_schema_checking: libc::c_int,
) -> *const raw::c_char {
    panic_safe_exec(
        || translate_helper(current_db, sql, catalog, relax_schema_checking),
        Box::new(translation_success_payload),
        Box::new(translation_failure_payload),
    )
}

/// A helper function that encapsulates all the fallible parts of
/// translation whose errors can be returned in the FFI payload.
fn translate_helper(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
    catalog: *const libc::c_char,
    relax_schema_checking: libc::c_int,
) -> Result<mongosql::Translation, String> {
    let current_db =
        from_extern_string(current_db).map_err(|_| "current_db not valid UTF-8".to_string())?;
    let sql =
        from_extern_string(sql).map_err(|_| "sql query string not valid UTF-8".to_string())?;
    let schema_checking_mode = match relax_schema_checking {
        1 => Ok(SchemaCheckingMode::Relaxed),
        0 => Ok(SchemaCheckingMode::Strict),
        n => Err(format!("invalid value {} for relax_schema_checking", n)),
    }?;
    let catalog_str = from_extern_string(catalog)
        .map_err(|_| "catalog schema string not valid UTF-8".to_string())?;
    let catalog = build_catalog(catalog_str.as_str())?;

    // used for testing purpose
    #[cfg(feature = "test")]
    {
        if current_db == "__test_panic" && sql == "__test_panic" {
            panic!("panic thrown")
        }
    }

    mongosql::translate_sql(&current_db, &sql, &catalog, schema_checking_mode)
        .map_err(|e| format!("{}", e))
}

/// Converts the given base64-encoded bson document into a Catalog.
pub fn build_catalog(base_64_doc: &str) -> Result<Catalog, String> {
    let bson_doc_bytes = base64::decode(base_64_doc)
        .map_err(|e| format!("failed to decode base64 string: {}", e))?;
    let json_schemas: BTreeMap<String, BTreeMap<String, json_schema::Schema>> =
        bson::from_reader(&mut bson_doc_bytes.as_slice()).map_err(|e| {
            format!(
                "failed to convert BSON catalog to json_schema::Schema format: {}",
                e
            )
        })?;
    let catalog = json_schemas
        .into_iter()
        .flat_map(|(db, db_schema)| {
            db_schema.into_iter().map(move |(collection, json_schema)| {
                let mongosql_schema = schema::Schema::try_from(json_schema).map_err(|e| {
                    format!(
                        "failed to add JSON schema for collection {}.{} to the catalog: {}",
                        db, collection, e
                    )
                })?;
                Ok((
                    catalog::Namespace {
                        db: db.clone(),
                        collection,
                    },
                    mongosql_schema,
                ))
            })
        })
        .collect::<Result<Catalog, String>>()?;
    Ok(catalog)
}

/// Returns a base64-encoded BSON document representing the payload
/// returned for a successful translation.
fn translation_success_payload(t: mongosql::Translation) -> String {
    let translation = bson::doc! {
        "target_db": t.target_db,
        "target_collection": t.target_collection.unwrap_or_default(),
        "pipeline": t.pipeline,
        "result_set_schema": &bson::to_bson(&t.result_set_schema).expect("failed to convert result_set_schema to bson"),
    };

    base64::encode(bson::to_vec(&translation).expect("serializing bson to bytes failed"))
}

/// ErrorVisibility describes whether an error is "internal" or
/// "external" (i.e. whether it is safe and useful to expose to end
/// users).
enum ErrorVisibility {
    Internal,
    External,
}

/// Returns a base64-encoded BSON document representing the payload
/// returned for an unsuccessful translation with the provided error
/// message.
fn translation_failure_payload(error: String, error_visibility: ErrorVisibility) -> String {
    let internal = match error_visibility {
        ErrorVisibility::Internal => true,
        ErrorVisibility::External => false,
    };
    let translation = bson::doc! {
        "error": error,
        "error_is_internal": internal,
    };

    let mut buf = Vec::new();
    translation
        .to_writer(&mut buf)
        .expect("serializing bson to bytes failed");

    base64::encode(buf)
}

/// Returns a base64-encoded bson representation of
/// the namespaces referenced by the the provided
/// SQL query, when executed in the provided database.
#[no_mangle]
pub extern "C" fn get_namespaces(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
) -> *const raw::c_char {
    panic_safe_exec(
        || get_namespaces_helper(current_db, sql),
        Box::new(get_namespaces_success_payload),
        Box::new(get_namespaces_failure_payload),
    )
}

/// A helper function that encapsulates all the fallible parts of
/// get_namespaces whose errors can be returned in the FFI payload.
fn get_namespaces_helper(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
) -> Result<BTreeSet<mongosql::Namespace>, String> {
    let current_db =
        from_extern_string(current_db).map_err(|_| "current_db not valid UTF-8".to_string())?;
    let sql =
        from_extern_string(sql).map_err(|_| "sql query string not valid UTF-8".to_string())?;

    mongosql::get_namespaces(&current_db, &sql).map_err(|e| format!("{}", e))
}

/// Returns a base64-encoded BSON document representing the payload
/// returned for a successful get_namespaces call.
fn get_namespaces_success_payload(namespaces: BTreeSet<mongosql::Namespace>) -> String {
    let result = bson::doc! {
        "namespaces": &bson::to_bson(&namespaces).expect("failed to convert namespaces to bson"),
    };

    base64::encode(bson::to_vec(&result).expect("serializing bson to bytes failed"))
}

/// Returns a base64-encoded BSON document representing the payload
/// returned for an unsuccessful get_namespaces call.
fn get_namespaces_failure_payload(error: String, error_visibility: ErrorVisibility) -> String {
    let internal = match error_visibility {
        ErrorVisibility::Internal => true,
        ErrorVisibility::External => false,
    };
    let result = bson::doc! {
        "error": error,
        "error_is_internal": internal,
    };

    let mut buf = Vec::new();
    result
        .to_writer(&mut buf)
        .expect("serializing bson to bytes failed");

    base64::encode(buf)
}

/// Executes function `f` such that any panics do not crash the runtime. The
/// function `f` returns a `Result<T, String>`, and the caller specifies how
/// to handle a success (a `T`) and how to handle a failure (a `String`). If
/// `f` panics during execution, the panic is caught, turned into a String,
/// and argued to the `handle_failure` function with the Internal visibility.
///
/// This function also converts the resulting payload from either success or
/// failure into a base64-encoded string.
fn panic_safe_exec<F: FnOnce() -> Result<T, String> + UnwindSafe, T>(
    f: F,
    handle_success: Box<dyn FnOnce(T) -> String>,
    handle_failure: Box<dyn FnOnce(String, ErrorVisibility) -> String>,
) -> *const raw::c_char {
    let previous_hook = panic::take_hook();
    let (s, r) = mpsc::sync_channel(1);
    panic::set_hook(Box::new(move |i| {
        if let Some(location) = i.location() {
            let info = format!("in file '{}' at line {}", location.file(), location.line());
            let _ = s.send(info);
        }
    }));
    let result = panic::catch_unwind(f);
    panic::set_hook(previous_hook);

    let payload = match result {
        Ok(Ok(success)) => handle_success(success),
        Ok(Err(msg)) => handle_failure(msg, ErrorVisibility::External),
        Err(err) => {
            let msg = if let Some(msg) = err.downcast_ref::<&'static str>() {
                format!("caught panic during translation: {}\n{:?}", msg, r.recv())
            } else {
                format!("caught panic during translation: {:?}\n{:?}", err, r.recv())
            };
            handle_failure(msg, ErrorVisibility::Internal)
        }
    };

    to_raw_c_string(&payload).expect("failed to convert base64 string to extern string")
}

/// # Safety
///
/// Deletes a rust-allocated C string passed as a *mut raw::c_char.
/// The C string MUST have been allocated in rust and obtained using
/// into_raw().
#[no_mangle]
pub unsafe extern "C" fn delete_string(to_delete: *mut raw::c_char) {
    let _ = CString::from_raw(to_delete);
}

/// Creates a String from the provided C string
fn from_extern_string(s: *const libc::c_char) -> Result<String, FromUtf8Error> {
    let s = unsafe { CStr::from_ptr(s).to_bytes() };
    String::from_utf8(s.to_vec())
}

/// Returns a C string with the same value as the provided &str.
/// The returned C string has been forgotten with std::mem::forget, and will not be freed when
/// at the end of scope.
fn to_raw_c_string(s: &str) -> Result<*mut raw::c_char, NulError> {
    // build a new nul-terminated string
    let c_str = CString::new(s)?;
    Ok(c_str.into_raw())
}
