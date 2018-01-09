#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    package=${PACKAGE:-evaluator}

    echo "running $package benchmarks"

    cd "$PROJECT_DIR/$package"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/benchmarks-unit.out" < $test_pipe&

    benchtime="5s"
    benchcount="3"
    if [ "$VARIANT" != "" ]; then
        # spend more time benchmarking on evergreen
        benchtime="10s"
        benchcount="5"
    fi

    go test -v \
        -run $^ \
        -bench='.' \
        -benchtime="$benchtime" \
        -count="$benchcount" \
        -timeout 4h \
        $BUILD_FLAGS \
        $TEST_BUILD_FLAGS \
        > $test_pipe

    rm $test_pipe

    echo "done running $package benchmarks"

) > $LOG_FILE 2>&1

print_exit_msg
