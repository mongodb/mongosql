
# This file should be sourced by any script in testdata/bin that uses an
# externally-defined environment variable.
# This file should be sourced after testdata/bin/platforms.sh

# use the provided script dir, if specified
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname $0)" && pwd -P)}"

# store commonly-used filepaths in variables
basename=${0##*/}
PROJECT_DIR="$(dirname "$(dirname $SCRIPT_DIR)")"
ARTIFACTS_DIR="$PROJECT_DIR/testdata/artifacts"
MONGODB_BINARIES="$ARTIFACTS_DIR/mongodb/bin"
LOG_FILE="$ARTIFACTS_DIR/log/${basename%.sh}.log"
KRB5_TRACE="$ARTIFACTS_DIR/log/krb5.log"
if [ "$VARIANT" = "" ]; then
    SQLPROXY_TEST_CACHE_DIR="$HOME/.sqlproxy-test-cache"
else
    SQLPROXY_TEST_CACHE_DIR="$PROJECT_DIR/.sqlproxy-test-cache"
fi

# set GOPATH, GOBIN
GOPATH="$(dirname $(dirname $(dirname $(dirname $PROJECT_DIR))))"
GOBIN="$GOPATH/bin/"

# set defaults for infrastructure config variables
if [ "$INFRASTRUCTURE_CONFIG" = "" ]; then
    INFRASTRUCTURE_CONFIG="default"
fi

MYSQL_PATH="$ARTIFACTS_DIR/mysql/bin"

# if on cygwin, convert paths as needed
if [ "Windows_NT" = "$OS" ]; then
    SCRIPT_DIR="$(cygpath -m $SCRIPT_DIR)"
    PROJECT_DIR="$(cygpath -m $PROJECT_DIR)"
    ARTIFACTS_DIR="$(cygpath -m $ARTIFACTS_DIR)"
    MYSQL_PATH="$(cygpath -m $MYSQL_PATH)"
    MONGODB_BINARIES="$(cygpath -m $MONGODB_BINARIES)"
    LOG_FILE="$(cygpath -m $LOG_FILE)"
    GOPATH="$(cygpath -m $GOPATH)"
    KRB5_TRACE="$(cygpath -m $KRB5_TRACE)"
fi

PATH="$GOBINDIR:$PYTHON_PATH:$MYSQL_PATH:$MONGODB_BINARIES:$PATH:$MINGW_PATH:$LIBRARY_PATH:$GOBIN:$GOROOT"

# source infrastructure config files
for infra_config in $( echo $INFRASTRUCTURE_CONFIG | sed "s/,/ /g" ); do
    . "$PROJECT_DIR/testdata/config/$infra_config"
    [ "$?" = "0" ] || exit 1
done

# set variables used for versioning
CURRENT_VERSION="${CURRENT_VERSION:-$(git describe)}"
GIT_SPEC="$(git rev-parse HEAD)"

# assemble various build tags into one variable
if [ "$BUILD_GSSAPI" = 'true' ]; then
    BUILD_TAGS="$BUILD_TAGS gssapi"
fi

# assemble linker flags for building the binaries
CONFIG_PATH="github.com/10gen/sqlproxy/internal/config"
LD_FLAGS="-X $CONFIG_PATH.VersionStr=$CURRENT_VERSION -X $CONFIG_PATH.Gitspec=$GIT_SPEC"

# assemble various build argument sets into one variable
BUILD_FLAGS="$BUILD_RACE_FLAG"

# assemble various sqlproxy argument sets into one variable
SQLPROXY_SSL_ARGS="$SQLPROXY_MONGO_MIN_CLIENT_TLS $SQLPROXY_SSLMODE $SQLPROXY_SSLCAFILE $SQLPROXY_SSLPEMKEYFILE $SQLPROXY_SSLPEMKEYPASSWORD $SQLPROXY_SSLALLOWINVALIDCERTIFICATES"
SQLPROXY_MONGO_SSL_ARGS="$SQLPROXY_MIN_MONGO_TLS $SQLPROXY_MONGO_SSL $SQLPROXY_MONGO_SSLCAFILE $SQLPROXY_MONGO_SSLPEMKEYFILE $SQLPROXY_MONGO_SSLPEMKEYPASSWORD $SQLPROXY_MONGO_SSLALLOWINVALIDCERTIFICATES $SQLPROXY_MONGO_SSLALLOWINVALIDHOSTNAMES $SQLPROXY_MONGO_SSLFIPSMODE $SQLPROXY_MONGO_SSLCRLFILE"
SQLPROXY_LOG_ARGS="$SQLPROXY_LOG_ROTATE $SQLPROXY_LOG_PATH"
SQLPROXY_SCHEMA_ARGS="$SQLPROXY_SCHEMA_SOURCE $SQLPROXY_SCHEMA_ALTER $SQLPROXY_SAMPLE_SIZE $SQLPROXY_SAMPLE_MODE $SQLPROXY_SAMPLE_INTERVAL $SQLPROXY_SAMPLE_NAMESPACES"
SQLPROXY_DEBUG_ARGS="$SQLPROXY_PROFILE"
SQLPROXY_SERVER_ARGS="$SQLPROXY_ADDR"
SQLPROXY_AUTH_ARGS="$SQLPROXY_AUTH_MECHANISM $SQLPROXY_AUTH $SQLPROXY_ADMIN_CREDS $SQLPROXY_ADMIN_AUTH_SOURCE $SQLPROXY_GSSAPI_HOST_NAME_ARG $SQLPROXY_GSSAPI_SERVICE_NAME_ARG"
SQLPROXY_ARGS="$SQLPROXY_AUTH_ARGS $SQLPROXY_SSL_ARGS $SQLPROXY_MONGO_SSL_ARGS $SQLPROXY_MONGO_URI_ARGS $SQLPROXY_LOG_ARGS $SQLPROXY_SCHEMA_ARGS $SQLPROXY_DEBUG_ARGS $SQLPROXY_SERVER_ARGS"

# assemble various mongo shell argument sets into one variable
MONGO_CLIENT_ARGS="$MONGO_CLIENT_AUTH_ARGS $MONGO_CLIENT_SSL_ARGS"

# assemble various mysql cli argument sets into one variable
CLIENT_CONNECT_ARGS="$CLIENT_HOST_ARG $CLIENT_PORT_ARG $CLIENT_PROTOCOL_ARG"
CLIENT_SSL_ARGS="$CLIENT_SSLMODE $CLIENT_SSLCA $CLIENT_SSLCERT"
CLIENT_AUTH_ARGS="$CLIENT_CREDS $CLIENT_CLEARTEXT $CLIENT_PLUGIN"
CLIENT_ARGS="$CLIENT_CONNECT_ARGS $CLIENT_SSL_ARGS $CLIENT_AUTH_ARGS $CLIENT_PROTOCOL"

# assemble various mysql cli argument sets into one variable
SECOND_CLIENT_CONNECT_ARGS="$CLIENT_HOST_ARG $CLIENT_PORT_ARG $CLIENT_PROTOCOL_ARG"
SECOND_CLIENT_SSL_ARGS="$CLIENT_SSLMODE $CLIENT_SSLCA $CLIENT_SSLCERT"
SECOND_CLIENT_AUTH_ARGS="$SECOND_CLIENT_CREDS $CLIENT_CLEARTEXT $CLIENT_PLUGIN"
SECOND_CLIENT_ARGS="$SECOND_CLIENT_CONNECT_ARGS $SECOND_CLIENT_SSL_ARGS $SECOND_CLIENT_AUTH_ARGS $CLIENT_PROTOCOL"

# assemble mysql CLI arguments for a second user, if one is provided, otherwise use the same arguments as the primary user.
if [ -z "$MONGO_OTHER_USER_CREDS" ]; then
    OTHER_CLIENT_AUTH_ARGS="$CLIENT_AUTH_ARGS"
    MONGO_OTHER_USER_PWD="$MYSQL_PWD"
else
    OTHER_CLIENT_AUTH_ARGS="$MONGO_OTHER_USER_CREDS $CLIENT_CLEARTEXT $CLIENT_PLUGIN"
fi
OTHER_CLIENT_ARGS="$CLIENT_CONNECT_ARGS $CLIENT_SSL_ARGS $OTHER_CLIENT_AUTH_ARGS $CLIENT_PROTOCOL"

# assemble various mongodrdl cli argument sets into one variable
DRDL_AUTH_ARGS="$DRDL_CREDS $DRDL_AUTH_SOURCE $DRDL_AUTH_MECHANISM"
DRDL_SSL_ARGS="$DRDL_SSL $DRDL_MIN_TLS"
DRDL_MONGO_HOST="$DRDL_MONGO_HOST"
DRDL_NAMESPACE="$DRDL_NAMESPACE"
DRDL_ARGS="$DRDL_NAMESPACE $DRDL_MONGO_HOST $DRDL_AUTH_ARGS $DRDL_SSL_ARGS"

# export any environment variables that will be needed by subprocesses
export CC
export GOPATH
export GORACE
export GOROOT
export JAVA_HOME
export KRB5_TRACE
export KRB5_CONFIG
export KRB5_KTNAME
export KRB5_CLIENT_KTNAME
export MYSQL_PWD
export SECOND_MYSQL_PWD
export PATH
export PKG_CONFIG_PATH
export SQLPROXY_AUTHTEST
export SQLPROXY_MEMORY_MANAGER_FAILPOINT_OFF
export SQLPROXY_SSLTEST

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
