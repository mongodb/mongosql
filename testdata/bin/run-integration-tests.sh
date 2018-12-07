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

    test_pushdown_flag=''
    if [ "$SQLPROXY_PUSHDOWN_OFF" != '' ]; then
        test_pushdown_flag='-nopushdown'
    fi

    go test -v \
        -run "TestIntegration/$SUITE/$NAMES" \
        -automate data,schema \
        $test_pushdown_flag \
        -timeout 4h \
        -tags="ssl $BUILD_TAGS" \
        $BUILD_FLAGS \
        $TEST_PARALLEL_FLAG \
        > $test_pipe

    rm $test_pipe

    echo "done running ${SUITE:-all integration} tests"

) > $LOG_FILE 2>&1

print_exit_msg
