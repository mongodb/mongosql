#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running $SUITE benchmarks..."

    cd "$PROJECT_DIR"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/${SUITE}-benchmarks.out" < $test_pipe&

    go test -v \
        -run $^ \
        -timeout 4h \
        -restoreData "$SUITE" \
        -bench=${BENCHMARKS:-.} \
        -benchmem \
        -benchtime=5s \
        $BUILD_FLAGS \
        $VERSION_FLAG \
        > $test_pipe

    rm $test_pipe

    echo "done running $SUITE benchmarks"

) > $LOG_FILE 2>&1

print_exit_msg
