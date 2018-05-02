#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "restoring test data..."
    set -o verbose
    cd "$PROJECT_DIR"

    go test -v \
        -run "TestIntegration/$SUITE/$^" \
        -timeout 4h \
        -tags="ssl $BUILD_TAGS" \
        $BUILD_FLAGS \
        -automate data
    echo "done restoring test data"

) > $LOG_FILE 2>&1

print_exit_msg
