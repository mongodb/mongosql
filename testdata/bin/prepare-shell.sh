
# This file should be sourced by any script in testdata/bin that uses an
# externally-defined environment variable. On evergreen, most of these
# variables are defined in the "fetch source" task; these defaults will make
# the scripts behave as expected when run locally.

scripts_dir="$(cd "$(dirname $0)" && pwd -P)"
PROJECT_DIR="${PROJECT_DIR:-$(dirname "$(dirname $scripts_dir)")}"
ARTIFACTS_DIR="$PROJECT_DIR/testdata/artifacts"
COVER_FLAG="-coverprofile=$ARTIFACTS_DIR/out/$SUITE-coverage.out"

if [ "$ON_EVERGREEN" != "true" ]; then
    CURRENT_VERSION="$(git describe)"
    GIT_SPEC="$(git rev-parse HEAD)"
    PUSH_ARCH="x86_64-ubuntu1404"
    PUSH_NAME="linux"
fi

if [ "$SSL" = "ssl" ]; then
    export SQLPROXY_SSLTEST=1
    export SQLPROXY_ARGS="--mongo-ssl --mongo-sslAllowInvalidCertificates --mongo-sslPEMKeyFile $PROJECT_DIR/testdata/resources/x509gen/client.pem"
    BUILD_TAGS="-tags ssl"
fi

BUILD_FLAGS="$BUILD_TAGS $BUILD_FLAGS"

export MONGO_ORCHESTRATION_HOME="$ARTIFACTS_DIR/orchestration"

MONGODB_BINARIES="$ARTIFACTS_DIR/mongodb/bin"
PATH="$MONGODB_BINARIES:$PATH"

basename=${0##*/}
LOG_FILE="$ARTIFACTS_DIR/log/${basename%.sh}.log"

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
