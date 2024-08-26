use command::Command;
use jni::{
    objects::{JByteArray, JClass},
    sys::jbyteArray,
    JNIEnv,
};
use panic_safe::panic_safe_exec;

mod command;
mod panic_safe;

#[repr(C)]
pub struct OdbcCommand {
    data: *const u8,
    length: usize,
    capacity: usize,
}

/// This is the ODBC entry point for the library.
///
/// # Safety
/// Because this is an FFI function, it is unsafe. The caller must ensure that the input parameters are valid.
///
/// # Parameters
/// - `command`: a C struct that describes a decomposed BSON byte vector containing the command to execute. The three fields in the struct are data (*const u8), length (usize), and capacity (usize).
#[no_mangle]
pub unsafe extern "C" fn runCommand(command: OdbcCommand) -> OdbcCommand {
    let result = panic_safe_exec(|| {
        let data = Vec::from_raw_parts(command.data.cast_mut(), command.length, command.capacity);
        let command = Command::new(data.as_slice());
        command.run()
    });
    let length = result.len();
    let capacity = result.capacity();
    OdbcCommand {
        data: Box::into_raw(result.into_boxed_slice()).cast(),
        length,
        capacity,
    }
}

/// This is the JDBC entry point for the library.
///
/// # Parameters
/// - `env`: The JNI environment.
/// - `_class`: The Java class, which is not used.
/// - `command`: The command to execute as a JByteArray.
#[allow(non_snake_case)]
#[no_mangle]
pub extern "C" fn Java_com_mongodb_mongosql_runCommand(
    env: JNIEnv,
    _class: JClass,
    command: JByteArray,
) -> jbyteArray {
    let result = panic_safe_exec(|| {
        let command = env
            .convert_byte_array(command)
            .expect("Failed to convert command");
        let command = Command::new(&command);
        command.run()
    });

    env.byte_array_from_slice(&result)
        .expect("Failed to convert to byte array")
        .into_raw()
}

#[cfg(test)]
mod test {
    use super::*;
    use mongodb::bson::{doc, Document};

    #[test]
    fn test_run_command_odbc() {
        let command = doc! {
            "command": "test",
            "options": {
                "test": true
            }
        };
        let command = mongodb::bson::to_vec(&command).unwrap();
        let length = command.len();
        let capacity = command.capacity();
        let odbc_command = OdbcCommand {
            data: Box::into_raw(command.into_boxed_slice()).cast(),
            length,
            capacity,
        };
        let ffi_result = unsafe { runCommand(odbc_command) };
        let result: Document = unsafe {
            mongodb::bson::from_slice(
                Vec::from_raw_parts(
                    ffi_result.data.cast_mut(),
                    ffi_result.length,
                    ffi_result.capacity,
                )
                .as_slice(),
            )
            .expect("Failed to deserialize result")
        };
        let expected = doc! {
            "success": true,
        };
        assert_eq!(result, expected);
    }
}
