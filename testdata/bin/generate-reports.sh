#!/bin/bash

. "$(dirname $0)/platforms.sh"
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
