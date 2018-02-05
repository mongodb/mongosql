#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"
(
    set -o errexit
    echo "building mongosqld ($CURRENT_VERSION)..."
    echo "build flags are '$BUILD_FLAGS'"
    echo "build tags are '$BUILD_TAGS'"
    echo "ldflags are '$LD_FLAGS'"
    out="$ARTIFACTS_DIR/bin/mongosqld"
    main="$PROJECT_DIR/main/sqlproxy.go"
    go build -v $BUILD_FLAGS -tags="ssl $BUILD_TAGS" -ldflags="$LD_FLAGS" -o $out $main

    echo "done building mongosqld"

) > $LOG_FILE 2>&1

print_exit_msg
