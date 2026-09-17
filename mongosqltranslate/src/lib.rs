use command::Command;
use jni::{
    objects::{JByteArray, JClass},
    sys::jbyteArray,
    JNIEnv,
};
use panic_safe::panic_safe_exec;
use semver::VersionReq;
use std::sync::LazyLock;

mod command;
#[cfg(test)]
mod command_tests;
mod panic_safe;

static MONGOSQLTRANSLATE_VERSION: LazyLock<String> = LazyLock::new(|| {
    format!(
        "v{}.{}.{}",
        env!("CARGO_PKG_VERSION_MAJOR"),
        env!("CARGO_PKG_VERSION_MINOR"),
        env!("CARGO_PKG_VERSION_PATCH")
    )
});

// TODO: SQL-2309: Update minimum compatible driver versions in mongosqltranslate to the correct versions
static MINIMUM_COMPATIBLE_JDBC_VERSION: LazyLock<VersionReq> = LazyLock::new(|| {
    VersionReq::parse(">=3.0.0-alpha")
        .expect("Minimum compatible JDBC version could not be parsed.")
});
static MINIMUM_COMPATIBLE_ODBC_VERSION: LazyLock<VersionReq> = LazyLock::new(|| {
    VersionReq::parse(">=2.0.0-alpha")
        .expect("Minimum compatible ODBC version could not be parsed.")
});

static DEV_ODBC_VERSION: LazyLock<VersionReq> =
    LazyLock::new(|| VersionReq::parse("=0.0.0").expect("ODBC dev version could not be parsed."));

pub const DEV_JDBC_VERSION_SUFFIX: &str = "-dirty";

pub const SNAPSHOT_JDBC_VERSION_SUFFIX: &str = "-SNAPSHOT";

#[repr(C)]
pub struct OdbcCommand {
    data: *const u8,
    length: usize,
    capacity: usize,
}

/// This is the JDBC entry point for the library.
///
/// # Parameters
/// - `env`: The JNI environment.
/// - `_class`: The Java class, which is not used.
/// - `command`: The command to execute as a JByteArray.
#[allow(non_snake_case)]
#[unsafe(no_mangle)]
pub extern "C" fn Java_com_mongodb_jdbc_mongosql_MongoSQLTranslate_runCommand(
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
