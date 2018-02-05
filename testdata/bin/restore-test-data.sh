#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    echo "restoring test data..."

    cd "$PROJECT_DIR"

    go test -v \
        -run "TestIntegration/$SUITE/$^" \
        -timeout 4h \
        -tags="ssl $BUILD_TAGS" \
        $BUILD_FLAGS \
        $VERSION_FLAG \
        -automate data

    echo "done restoring test data"

) > $LOG_FILE 2>&1

print_exit_msg
