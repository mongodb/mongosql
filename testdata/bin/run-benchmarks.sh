#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    benchtype=${TYPE:-queries}
    benchtimeout=${TIMEOUT:-16h}
    benchnames='^([^t]...|.[^p]..|..[^c].|...[^h])'

    if [ "$benchtype" = "tpch-micro" ]; then
        benchtype='queries'
        benchnames='^tpch_micro'
    elif [ "$benchtype" = "tpch-normalized" ]; then
        benchtype='queries'
        benchnames='^tpch_full_normalized'
    elif [ "$benchtype" = "tpch-denormalized" ]; then
        benchtype='queries'
        benchnames='^tpch_full_denormalized'
    elif [ "$benchtype" = "tpch-handwritten-denormalized" ]; then
        benchtype='queries'
        benchnames='^tpch_full_handwritten_denormalized'
    fi

    echo "running $benchtype benchmarks..."

    cd "$PROJECT_DIR"

    test_pipe="$ARTIFACTS_DIR/test_pipe"
    [ -e $test_pipe ] && rm $test_pipe
    mkfifo $test_pipe
    tee -a "$ARTIFACTS_DIR/out/benchmarks.out" < $test_pipe&

    benchtime="5s"
    benchcount="3"
    if [ "$VARIANT" != "" ]; then
        # spend more time benchmarking on evergreen
        benchtime="10s"
        benchcount="5"
    fi

    go test -v \
        -run $^ \
        -bench="BenchmarkIntegration/$benchtype/$benchnames" \
        -benchtime="$benchtime" \
        -count="$benchcount" \
        -automate data \
        -timeout="$benchtimeout" \
        -tags="ssl $BUILD_TAGS" \
        $BUILD_FLAGS \
        > $test_pipe

    rm $test_pipe

    echo "done running $benchtype benchmarks"

) > $LOG_FILE 2>&1

print_exit_msg
