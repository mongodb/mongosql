#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"
(
    set -o errexit
    echo "building mongodrdl ($CURRENT_VERSION)..."
    echo "build flags are '$BUILD_FLAGS'"
    echo "build tags are '$BUILD_TAGS'"
    echo "ldflags are '$LD_FLAGS'"
    out="$ARTIFACTS_DIR/bin/mongodrdl"
    main="$PROJECT_DIR/cmd/mongodrdl/mongodrdl.go"
    go build -v $BUILD_FLAGS -tags="$BUILD_TAGS" -ldflags="$LD_FLAGS" -o $out $main

    echo "done building mongodrdl"

) > $LOG_FILE 2>&1

print_exit_msg
