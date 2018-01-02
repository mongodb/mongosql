#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    benchtype=${TYPE:-integration}

    echo "generating benchmark reports..."

    cd "$PROJECT_DIR"
    file="$ARTIFACTS_DIR/perf.json"

    go run testdata/bin/parse-benchmark-results.go -type "$benchtype" > "$file"

    echo "done generating benchmark reports"

) > $LOG_FILE 2>&1

print_exit_msg
