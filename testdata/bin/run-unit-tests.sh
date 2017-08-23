#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

run_unit_tests() {
    set -o pipefail

    cd "$PROJECT_DIR/$1"
    SUITE=${PWD##*/}
    COVER_FLAG="-coverprofile=$ARTIFACTS_DIR/out/$SUITE-coverage.out"

    echo "running $SUITE tests..."

    go test -v \
        $TEST_BUILD_FLAGS \
        $COVER_FLAG \
        | tee -a "$ARTIFACTS_DIR/out/${SUITE}-suite.out"

    echo "done running $SUITE tests"
}

(
    set -o errexit

    run_unit_tests catalog
    run_unit_tests collation
    run_unit_tests internal/config
    run_unit_tests internal/json
    run_unit_tests internal/util
    run_unit_tests internal/util/bsonutil
    run_unit_tests mongodrdl
    run_unit_tests mongodrdl/mongo
    run_unit_tests mongodrdl/relational
    run_unit_tests options
    run_unit_tests variable
    run_unit_tests evaluator
    run_unit_tests parser
    run_unit_tests mongodb
    run_unit_tests server
    run_unit_tests schema
    run_unit_tests schema/mongo

    rm -rf "$PROJECT_DIR/mongodrdl/out/"

) > $LOG_FILE 2>&1

print_exit_msg
