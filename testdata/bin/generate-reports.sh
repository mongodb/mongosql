#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    cd "$ARTIFACTS_DIR/out"

    echo "generating coverage reports..."
    for covfile in $(ls | grep 'coverage.out'); do
        go tool cover -func=$covfile -o "$ARTIFACTS_DIR/reports/${covfile%.out}.txt"
        go tool cover -html=$covfile -o "$ARTIFACTS_DIR/reports/${covfile%.out}.html"
    done
    echo "done generating coverage reports"

    echo "generating benchmark reports..."
    for benchfile in $(ls | grep 'benchmarks.out'); do
        cat $benchfile | \
            grep '^Benchmark' | \
            awk 'BEGIN{OFS="\t";print "Query","Minute/Op","MemAlloc/Op","ByteAlloc/Op","#Runs";}{gsub(/BenchmarkQuery/, "Q");gsub(/-16/, "");print $1, $3/60000000000, $5, $7, $2}'a | \
            sed -e :a -e '$d;N;2,2ba' -e 'P;D' \
            > "$ARTIFACTS_DIR/reports/${benchfile%.out}.txt"
    done
    echo "done generating benchmark reports"

    echo "generating test suite reports..."
    for suitefile in $(ls | grep 'suite.out'); do
        cat -v $suitefile | perl -ne 's/\^\[\[(0|31|32|33)m//g;print' > "$ARTIFACTS_DIR/reports/${suitefile%.out}.txt"
    done
    echo "done generating test suite reports"

    cd $PROJECT_DIR

    echo "generating artifact tarball..."
    rm -rf testdata/artifacts/mongodb testdata/artifacts/bin
    tar czf artifacts.tar.gz testdata/artifacts
    mv artifacts.tar.gz testdata/artifacts
    echo "done generating artifact tarball"

) > $LOG_FILE 2>&1

print_exit_msg
