use std::{panic, panic::UnwindSafe, sync::mpsc};

use mongodb::bson::{self, doc, Document};

/// Executes function `f` such that any panics do not crash the runtime. The
/// function `f` returns a `Result<Document, Box<dyn std::error::Error>>`.
///
/// If `f` panics during execution, the panic is caught, turned into a String,
/// and marked as an internal_error.
///
/// If `f` returns an error, the error message is returned as a String and
/// marked as an external error.
///
/// If `f` executes normally, the resulting Document is used as-is.
///
/// This function also converts the resulting bson payload from either success or
/// failure into a Vec<u8> for ease of return.
// TODO: SQL-2280: Remove this dead_code
#[allow(dead_code)]
pub(crate) fn panic_safe_exec<
    F: FnOnce() -> Result<Document, Box<dyn std::error::Error>> + UnwindSafe,
>(
    f: F,
) -> Vec<u8> {
    let previous_hook = panic::take_hook();
    let (s, r) = mpsc::channel();
    panic::set_hook(Box::new(move |i| {
        if let Some(location) = i.location() {
            let info = format!("in file '{}' at line {}", location.file(), location.line());
            let _ = s.send(info);
        }
    }));
    let result = panic::catch_unwind(f);
    panic::set_hook(previous_hook);

    let payload = match result {
        Ok(Ok(success)) => success,
        Ok(Err(msg)) => doc! {
            "error": msg.to_string(),
            "error_is_internal": false,
        },
        Err(err) => {
            let msg = if let Some(msg) = err.downcast_ref::<&'static str>() {
                format!(
                    "Internal Error: report this to MongoDB: {}\n{:?}",
                    msg,
                    r.recv()
                )
            } else if let Some(msg) = err.downcast_ref::<String>() {
                format!(
                    "Internal Error: report this to MongoDB: {}\n{:?}",
                    msg,
                    r.recv()
                )
            } else {
                format!(
                    "Internal Error: report this to MongoDB: {:?}\n{:?}",
                    err,
                    r.recv()
                )
            };
            doc! {
                "error": msg,
                "error_is_internal": true,
            }
        }
    };
    bson::to_vec(&payload).expect("serializing bson to bytes failed")
}

#[cfg(test)]
mod test {

    use super::*;
    use crate::command::Command;

    #[test]
    fn test_panic_safe_exec_ok() {
        let command = bson::to_vec(&doc! {
            "command": "test",
            "options": {
                "test": true,
            },
        })
        .expect("failed to serialize bson");
        let result = panic_safe_exec(|| Command::new(&command).run());
        let expected = doc! {
            "success": true,
        };
        assert_eq!(
            bson::from_slice::<Document>(&result).expect("failed to deserialize in test"),
            expected
        );
    }

    #[test]
    fn test_panic_safe_exec_err() {
        let command = bson::to_vec(&doc! {
            "command": "test",
            "options": {
                "test": false,
            },
        })
        .expect("failed to serialize bson");
        let result = panic_safe_exec(|| Command::new(&command).run());
        let expected = doc! {
            "error": "Test errored",
            "error_is_internal": false,
        };
        assert_eq!(
            bson::from_slice::<Document>(&result).expect("failed to deserialize in test"),
            expected
        );
    }

    #[test]
    fn test_panic_safe_exec_panic_unknown_command() {
        let command = bson::to_vec(&doc! {
            "command": "unknown",
            "options": {
                "test": false,
            },
        })
        .expect("failed to serialize bson");
        let result = panic_safe_exec(|| Command::new(&command).run());
        let result = bson::from_slice::<Document>(&result).expect("failed to deserialize in test");
        assert!(result.get("error").is_some());
        assert!(result
            .get("error_is_internal")
            .expect("error_is_internal is missing")
            .as_bool()
            .expect("error_is_internal is not a bool"));
        assert!(
            result
                .get("error")
                .expect("error is missing")
                .as_str()
                .expect("error is not a string").contains("Internal Error: report this to MongoDB: Deserializing the provided Bson::Document into `Command` data type failed."),
            "Unknown command"
        );
    }
}
