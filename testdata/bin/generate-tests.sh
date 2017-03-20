#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    suites=${SUITE:-integration}

    echo "generating $suites tests..."

    cd $PROJECT_DIR
    go run testdata/bin/generate.go -suites $suites

    echo "done generating $suites tests"

) > $LOG_FILE 2>&1

print_exit_msg
