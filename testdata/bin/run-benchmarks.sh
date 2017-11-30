#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    benchtype=${TYPE:-queries}

    echo "running $type benchmarks..."

    cd "$PROJECT_DIR"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/benchmarks.out" < $test_pipe&

    go test -v \
        -run $^ \
        -bench="BenchmarkIntegration/$benchtype" \
        -automate data \
        -timeout 4h \
        -benchmem \
        -benchtime=5s \
        $TEST_BUILD_FLAGS \
        $VERSION_FLAG \
        > $test_pipe

    rm $test_pipe

    echo "done running $benchtype benchmarks"

) > $LOG_FILE 2>&1

print_exit_msg
