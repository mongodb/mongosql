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
        -tags="$BUILD_TAGS" \
        $BUILD_FLAGS \
        $COVER_FLAG \
        | tee -a "$ARTIFACTS_DIR/out/${SUITE}-suite.out"

    echo "done running $SUITE tests"
}

(
    set -o errexit

    for pkg in $(find . -name '*.go' | grep -v './vendor' | grep -v './testdata' | grep -v './release' | xargs -L1 dirname | uniq); do
        if [[ "$pkg" == "." ]]; then
            continue;
        fi
        echo "Running unit tests for package $pkg";
        run_unit_tests $pkg;
    done


) > $LOG_FILE 2>&1

print_exit_msg
