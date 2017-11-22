
# This file should be sourced by any script in testdata/bin that uses an
# externally-defined environment variable.
# This file should be sourced after testdata/bin/platforms.sh

# use the provided script dir, if specified
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname $0)" && pwd -P)}"

# store commonly-used filepaths in variables
basename=${0##*/}
PROJECT_DIR="$(dirname "$(dirname $SCRIPT_DIR)")"
ARTIFACTS_DIR="$PROJECT_DIR/testdata/artifacts"
MONGO_ORCHESTRATION_HOME="$ARTIFACTS_DIR/mo"
MONGODB_BINARIES="$ARTIFACTS_DIR/mongodb/bin"
LOG_FILE="$ARTIFACTS_DIR/log/${basename%.sh}.log"

# set GOPATH
GOPATH="$(dirname $(dirname $(dirname $(dirname $PROJECT_DIR))))"

# set defaults for infrastructure config variables
if [ "$INFRASTRUCTURE_CONFIG" = "" ]; then
    INFRASTRUCTURE_CONFIG="default"
fi

# set PATH
MYSQL_PATH="$ARTIFACTS_DIR/mysql/bin"
PATH="$MYSQL_PATH:$MONGODB_BINARIES:$GOBINDIR:$PATH:$MINGW_PATH:$LIBRARY_PATH"

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

# source infrastructure config files
for infra_config in $( echo $INFRASTRUCTURE_CONFIG | sed "s/,/ /g" ); do
    . "$PROJECT_DIR/testdata/config/$infra_config"
    [ "$?" = "0" ] || exit 1
done

# set variables used for versioning
CURRENT_VERSION="${CURRENT_VERSION:-$(git describe)}"
GIT_SPEC="$(git rev-parse HEAD)"

# assemble various build argument sets into one variable
BUILD_FLAGS="$BUILD_TAGS $BUILD_RACE_FLAG"

# assemble various sqlproxy argument sets into one variable
SQLPROXY_SSL_ARGS="$SQLPROXY_SSLMODE $SQLPROXY_SSLCAFILE $SQLPROXY_SSLPEMKEYFILE $SQLPROXY_SSLPEMKEYPASSWORD $SQLPROXY_SSLALLOWINVALIDCERTIFICATES"
SQLPROXY_MONGO_SSL_ARGS="$SQLPROXY_MONGO_SSL $SQLPROXY_MONGO_SSLCAFILE $SQLPROXY_MONGO_SSLPEMKEYFILE $SQLPROXY_MONGO_SSLPEMKEYPASSWORD $SQLPROXY_MONGO_SSLALLOWINVALIDCERTIFICATES $SQLPROXY_MONGO_SSLALLOWINVALIDHOSTNAMES $SQLPROXY_MONGO_SSLFIPSMODE $SQLPROXY_MONGO_SSLCRLFILE"
SQLPROXY_LOG_ARGS="$SQLPROXY_LOG_ROTATE $SQLPROXY_LOG_PATH"
SQLPROXY_SCHEMA_ARGS="$SQLPROXY_SCHEMA_SOURCE $SQLPROXY_SCHEMA_ALTER $SQLPROXY_SAMPLE_SIZE $SQLPROXY_SAMPLE_CREDS $SQLPROXY_SAMPLE_AUTH_SOURCE $SQLPROXY_SAMPLE_MODE $SQLPROXY_SAMPLE_INTERVAL"
SQLPROXY_DEBUG_ARGS="$SQLPROXY_PROFILE"
SQLPROXY_ARGS="$SQLPROXY_AUTH_ARGS $SQLPROXY_SSL_ARGS $SQLPROXY_MONGO_SSL_ARGS $SQLPROXY_SCHEMA_ARGS $SQLPROXY_LOG_ARGS $SQLPROXY_SCHEMA_ARGS $SQLPROXY_DEBUG_ARGS"

# assemble various mongo shell argument sets into one variable
MONGO_CLIENT_ARGS="$MONGO_CLIENT_AUTH_ARGS $MONGO_CLIENT_SSL_ARGS"

# assemble various mysql cli argument sets into one variable
CLIENT_CONNECT_ARGS="$CLIENT_HOST_ARG $CLIENT_PORT_ARG $CLIENT_PROTOCOL_ARG"
CLIENT_SSL_ARGS="$CLIENT_SSLMODE $CLIENT_SSLCA $CLIENT_SSLCERT"
CLIENT_AUTH_ARGS="$CLIENT_CREDS $CLIENT_CLEARTEXT $CLIENT_PLUGIN"
CLIENT_ARGS="$CLIENT_CONNECT_ARGS $CLIENT_SSL_ARGS $CLIENT_AUTH_ARGS $CLIENT_PROTOCOL"

# assemble mysql CLI arguments for a second user, if one is provided, otherwise use the same arguments as the primary user.
if [ -z "$USER2_CREDS" ]; then
    USER2_AUTH_ARGS="$CLIENT_AUTH_ARGS"
    USER2_PWD="$MYSQL_PWD"
else
    USER2_AUTH_ARGS="$USER2_CREDS $CLIENT_CLEARTEXT $CLIENT_PLUGIN"
fi
USER2_CLIENT_ARGS="$CLIENT_CONNECT_ARGS $CLIENT_SSL_ARGS $USER2_AUTH_ARGS $CLIENT_PROTOCOL"

# assemble various mongodrdl cli argument sets into one variable
DRDL_AUTH_ARGS="$DRDL_CREDS $DRDL_AUTH_SOURCE"
DRDL_SSL_ARGS="$DRDL_SSL"
DRDL_ARGS="$DRDL_AUTH_ARGS $DRDL_SSL_ARGS"

# export any environment variables that will be needed by subprocesses
export SQLPROXY_SSLTEST
export SQLPROXY_AUTHTEST
export SQLPROXY_PUSHDOWN_OFF
export MONGO_ORCHESTRATION_HOME
export GOROOT
export GOPATH
export CC
export JAVA_HOME
export PATH
export MYSQL_PWD
export GORACE

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
