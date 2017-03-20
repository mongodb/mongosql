#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"
trap "mv -f $PROJECT_DIR/common/version.go.bak $PROJECT_DIR/common/version.go" HUP EXIT

(
    set -o errexit
    echo "building mongosqld ($CURRENT_VERSION)..."

    sed -i.bak -e "s/built-without-version-string/$CURRENT_VERSION/" \
        -e "s/built-without-git-spec/$GIT_SPEC/" \
        "$PROJECT_DIR/common/version.go"

    out="$ARTIFACTS_DIR/bin/mongosqld"
    main="$PROJECT_DIR/main/sqlproxy.go"
    go build \
        $RACE_DETECTOR \
        $BUILD_FLAGS \
        -o $out $main

    echo "done building mongosqld"

) > $LOG_FILE 2>&1

print_exit_msg
