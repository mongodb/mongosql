#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    echo "generating benchmark reports..."

    cd "$PROJECT_DIR"
    file="$ARTIFACTS_DIR/perf.json"

    # make sure this runs successfully
    go run testdata/bin/parse-benchmark-results.go > /dev/null

    echo "$(go run testdata/bin/parse-benchmark-results.go)" > $file

    echo "done generating benchmark reports"

) > $LOG_FILE 2>&1

print_exit_msg
