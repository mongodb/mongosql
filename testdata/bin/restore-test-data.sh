#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "restoring $SUITE data..."

    cd "$PROJECT_DIR"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/${SUITE}-suite.out" < $test_pipe&

    cd "$PROJECT_DIR"

    go test -v \
        -run $^ \
        -timeout 4h \
        $RACE_DETECTOR \
        $BUILD_FLAGS \
        $COVER_FLAG \
        $VERSION_FLAG \
        -restoreData "$SUITE"
        > $test_pipe

    rm $test_pipe

    echo "done restoring $SUITE data"

) > $LOG_FILE 2>&1

print_exit_msg
