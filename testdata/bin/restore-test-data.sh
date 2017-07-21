#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    suites=${SUITE:-integration}

    echo "restoring $suites data..."

    cd "$PROJECT_DIR"

    go test -v \
        -run $^ \
        -timeout 4h \
        $TEST_BUILD_FLAGS \
        $COVER_FLAG \
        $VERSION_FLAG \
        -restoreData "$suites"

    echo "done restoring $suites data"

) > $LOG_FILE 2>&1

print_exit_msg
