use std::{
    ffi::{CStr, CString, NulError},
    os::raw,
    string::FromUtf8Error,
};
mod version;

/// Returns the semantic version of this library as a C string.
/// The caller is responsible for freeing the returned value.
#[no_mangle]
pub extern "C" fn version() -> *const raw::c_char {
    to_extern_string(version::VERSION).expect("semver string contained NUL byte")
}

/// Returns an extjson-encoded version of
/// [Translation](/mongosql/struct.Translation.html) for the provided
/// SQL query and database.
#[no_mangle]
pub extern "C" fn translate(
    current_db: *const libc::c_char,
    sql: *const libc::c_char,
) -> *const raw::c_char {
    let current_db = from_extern_string(current_db).expect("current_db not valid UTF-8");
    let sql = from_extern_string(sql).expect("sql not valid UTF-8");
    let res = mongosql::translate_sql_bson_base64(&current_db, &sql);
    to_extern_string(&res).expect("failed to convert base64 string to extern string")
}

/// Creates a String from the provided C string.
fn from_extern_string(s: *const libc::c_char) -> Result<String, FromUtf8Error> {
    let s = unsafe { CStr::from_ptr(s).to_bytes() };
    String::from_utf8(s.to_vec())
}

/// Returns a C string with the same value as the provided &str.
/// The caller is responsible for freeing the returned value.
fn to_extern_string(s: &str) -> Result<*const raw::c_char, NulError> {
    // build a new nul-terminated string
    let c_str = CString::new(s)?;

    // turn it into a pointer
    let c_str_ptr = c_str.as_ptr();

    // tell rust not to free the string when this func ends
    std::mem::forget(c_str);

    Ok(c_str_ptr)
}
