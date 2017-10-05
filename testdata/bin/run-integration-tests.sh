#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running ${SUITE:-all integration} tests..."

    cd "$PROJECT_DIR"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/${SUITE}-suite.out" < $test_pipe&

    go test -v \
        -run "TestIntegration/$SUITE/$NAMES" \
        -automate data \
        -timeout 4h \
        $TEST_BUILD_FLAGS \
        $TEST_PARALLEL_FLAG \
        $VERSION_FLAG \
        > $test_pipe

    rm $test_pipe

    echo "done running ${SUITE:-all integration} tests"

) > $LOG_FILE 2>&1

print_exit_msg
