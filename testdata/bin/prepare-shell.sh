
# This file should be sourced by any script in testdata/bin that uses an
# externally-defined environment variable.
# This file should be sourced after testdata/bin/platforms.sh

# use the provided script dir, if specified
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname $0)" && pwd -P)}"

# store commonly-used filepaths in variables
basename=${0##*/}
PROJECT_DIR="$(dirname "$(dirname $SCRIPT_DIR)")"
ARTIFACTS_DIR="$PROJECT_DIR/testdata/artifacts"
MONGO_ORCHESTRATION_HOME="$ARTIFACTS_DIR/orchestration"
MONGODB_BINARIES="$ARTIFACTS_DIR/mongodb/bin"
LOG_FILE="$ARTIFACTS_DIR/log/${basename%.sh}.log"

# set GOPATH
GOPATH="$(dirname $(dirname $(dirname $(dirname $PROJECT_DIR))))"

# set PATH
PATH="$MONGODB_BINARIES:$PATH:$MINGW_PATH:$GOBINDIR:$LIBRARY_PATH"

# set variables used for versioning
CURRENT_VERSION="${CURRENT_VERSION:-$(git describe)}"
GIT_SPEC="$(git rev-parse HEAD)"

# if on cygwin, convert paths as needed
if [ "Windows_NT" = "$OS" ]; then
    SCRIPT_DIR="$(cygpath -m $SCRIPT_DIR)"
    PROJECT_DIR="$(cygpath -m $PROJECT_DIR)"
    ARTIFACTS_DIR="$(cygpath -m $ARTIFACTS_DIR)"
    MONGO_ORCHESTRATION_HOME="$(cygpath -m $MONGO_ORCHESTRATION_HOME)"
    MONGODB_BINARIES="$(cygpath -m $MONGODB_BINARIES)"
    LOG_FILE="$(cygpath -m $LOG_FILE)"
    GOPATH="$(cygpath -m $GOPATH)"
fi

# set sqlproxy schema args according to whether we are
# using a drdl file or a dynamically-sampled schema
if [ "$USE_DYNAMIC_SCHEMA" != "true" ]; then
    SQLPROXY_SCHEMA_ARGS="--schemaDirectory $PROJECT_DIR/testdata/resources/schema"
fi

if [ "$SSL" = "ssl" ]; then
    export SQLPROXY_SSLTEST=1
    SQLPROXY_MONGO_SSL_ARGS="--mongo-ssl --mongo-sslAllowInvalidCertificates --mongo-sslPEMKeyFile $PROJECT_DIR/testdata/resources/x509gen/client.pem"
    BUILD_TAGS="-tags ssl"
fi

if [ "$AUTH" = "auth" ]; then
    SQLPROXY_AUTH_ARGS="--auth"
fi

# assemble various sqlproxy argument sets into one variable
SQLPROXY_ARGS="$SQLPROXY_AUTH_ARGS $SQLPROXY_SSL_ARGS $SQLPROXY_MONGO_SSL_ARGS $SQLPROXY_SCHEMA_ARGS"

# assemble various golang build arguments into one variable
BUILD_FLAGS="$BUILD_TAGS $BUILD_FLAGS"

# export any environment variables that will be needed by subprocesses
export SQLPROXY_SSLTEST
export MONGO_ORCHESTRATION_HOME
export GOROOT
export GOPATH
export CC
export JAVA_HOME
export PATH

# define the function that prints the exit message at the end of each script
print_exit_msg() {
    exit_code=$?
    if [ "$exit_code" != "0" ]; then
        status=FAILURE
    else
        status=SUCCESS
    fi

    echo "$status: $basename" 1>&2
    if [ "$status" = "FAILURE" ]; then
        echo "printing log from failed script:" 1>&2
        cat $LOG_FILE 1>&2
    fi

    return $exit_code
}
