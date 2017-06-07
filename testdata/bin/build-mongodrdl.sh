#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"
trap "mv -f $PROJECT_DIR/internal/config/version.go.bak $PROJECT_DIR/internal/config/version.go" HUP EXIT

(
    set -o errexit
    echo "building mongodrdl ($CURRENT_VERSION)..."

    sed -i.bak -e "s/built-without-version-string/$CURRENT_VERSION/" \
        -e "s/built-without-git-spec/$GIT_SPEC/" \
        "$PROJECT_DIR/internal/config/version.go"

    out="$ARTIFACTS_DIR/bin/mongodrdl"
    main="$PROJECT_DIR/mongodrdl/main/mongodrdl.go"
    go build \
        $BUILD_FLAGS \
        $RACE_DETECTOR \
        -o $out $main

    echo "done building mongodrdl"

) > $LOG_FILE 2>&1

print_exit_msg
