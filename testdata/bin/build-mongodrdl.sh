#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"
(
    set -o errexit
    echo "building mongodrdl ($CURRENT_VERSION)..."
    out="$ARTIFACTS_DIR/bin/mongodrdl"
    main="$PROJECT_DIR/mongodrdl/main/mongodrdl.go"
    go build -v -ldflags="$LD_FLAGS" $BUILD_FLAGS -tags="$BUILD_TAGS" -o $out $main

    echo "done building mongodrdl"

) > $LOG_FILE 2>&1

print_exit_msg
