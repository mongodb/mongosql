#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "restoring test data..."
    set -o verbose
    cd "$PROJECT_DIR"

    if [ -z "$NO_FLUSH_SCHEMA" ]; then
        AUTOMATE_OPTS="data,schema"
    else
        AUTOMATE_OPTS="data"
    fi

    go test -v \
        -run "TestIntegration/$SUITE/$^" \
        -timeout 4h \
        -tags="ssl $BUILD_TAGS" \
        $BUILD_FLAGS \
        -automate $AUTOMATE_OPTS \
        $ADDITIONAL_TEST_FLAGS
    echo "done restoring test data"

) > $LOG_FILE 2>&1

print_exit_msg
