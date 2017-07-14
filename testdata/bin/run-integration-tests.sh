#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running $SUITE tests..."

    cd "$PROJECT_DIR"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/${SUITE}-suite.out" < $test_pipe&

    go test -v \
        -timeout 4h \
        $RACE_DETECTOR \
        $BUILD_FLAGS \
        $COVER_FLAG \
        $VERSION_FLAG \
        $INTEGRATION_TEST_FLAGS \
        > $test_pipe

    rm $test_pipe

    echo "done running $SUITE tests"

) > $LOG_FILE 2>&1

print_exit_msg
