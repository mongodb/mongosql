use std::{
    ffi::{CStr, CString, NulError},
    os::raw,
    panic,
    string::FromUtf8Error,
};
mod version;

/// Returns the semantic version of this library as a C string.
/// The caller is responsible for freeing the returned value.
#[no_mangle]
pub extern "C" fn version() -> *mut raw::c_char {
    to_raw_c_string(version::VERSION).expect("semver string contained NUL byte")
}

/// Returns an extjson-encoded version of
/// [Translation](/mongosql/struct.Translation.html) for the provided
/// SQL query and database.
#[no_mangle]
pub extern "C" fn translate(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
) -> *const raw::c_char {
    let previous_hook = panic::take_hook();
    panic::set_hook(Box::new(|_| {}));
    let result = panic::catch_unwind(|| translate_sql_bson_base64(current_db, sql));
    panic::set_hook(previous_hook);

    let payload = match result {
        Ok(result) => match result {
            Ok(translation) => translation_success_payload(translation),
            Err(msg) => translation_failure_payload(msg),
        },
        Err(err) => {
            if let Some(msg) = err.downcast_ref::<&'static str>() {
                translation_failure_payload(format!("caught panic during translation: {}", msg))
            } else {
                translation_failure_payload(format!("caught panic during translation: {:?}", err))
            }
        }
    };

    to_raw_c_string(&payload).expect("failed to convert base64 string to extern string")
}

/// A helper function that encapsulates all the fallible parts of
/// translation whose errors can be returned in the FFI payload.
fn translate_sql_bson_base64(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
) -> Result<mongosql::Translation, String> {
    let current_db =
        from_extern_string(current_db).map_err(|_| "current_db not valid UTF-8".to_string())?;
    let sql =
        from_extern_string(sql).map_err(|_| "sql query string not valid UTF-8".to_string())?;

    // used for testing purpose
    #[cfg(feature = "test")]
    {
        if current_db == "__test_panic" && sql == "__test_panic" {
            panic!("panic thrown")
        }
    }

    mongosql::translate_sql(&current_db, &sql).map_err(|e| format!("{}", e))
}

/// Returns a base64-encoded BSON document representing the payload
/// returned for a successful translation.
fn translation_success_payload(t: mongosql::Translation) -> String {
    let translation = bson::doc! {
        "target_db": t.target_db.unwrap_or_else(|| "".to_string()),
        "target_collection": t.target_collection.unwrap_or_else(|| "".to_string()),
        "pipeline": t.pipeline,
    };

    let mut buf = Vec::new();
    translation
        .to_writer(&mut buf)
        .expect("serializing bson to bytes failed");

    base64::encode(buf)
}

/// Returns a base64-encoded BSON document representing the payload
/// returned for an unsuccessful translation with the provided error
/// message.
fn translation_failure_payload(error: String) -> String {
    let translation = bson::doc! {
        "error": error,
    };

    let mut buf = Vec::new();
    translation
        .to_writer(&mut buf)
        .expect("serializing bson to bytes failed");

    base64::encode(buf)
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
