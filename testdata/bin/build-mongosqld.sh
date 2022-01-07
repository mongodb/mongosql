#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"
(
    set -o errexit
    echo "building mongosqld ($CURRENT_VERSION)..."
    echo "$(go version)"
    echo "build flags are '$BUILD_FLAGS'"
    echo "build tags are '$BUILD_TAGS'"
    echo "ldflags are '$LD_FLAGS'"
    out="$ARTIFACTS_DIR/bin/mongosqld"
    main="$PROJECT_DIR/cmd/mongosqld/mongosqld.go"
    go build -v $BUILD_FLAGS -tags="$BUILD_TAGS" -ldflags="$LD_FLAGS" -o $out $main

    echo "done building mongosqld"

	echo "running 'mongosqld --version'"
	$out --version
	echo "done running 'mongosqld --version'"

) > $LOG_FILE 2>&1

print_exit_msg
