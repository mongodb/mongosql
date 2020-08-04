#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"
(
    set -o errexit
    echo "building mongotranslate..."
    echo "build flags are '$BUILD_FLAGS'"
    echo "build tags are '$BUILD_TAGS'"
    echo "ldflags are '$LD_FLAGS'"
    out="$ARTIFACTS_DIR/bin/mongotranslate"
    main="$PROJECT_DIR/cmd/mongotranslate/mongotranslate.go"
    go build -v $BUILD_FLAGS -tags="$BUILD_TAGS" -ldflags="$LD_FLAGS" -o $out $main

    echo "done building mongotranslate"

) > $LOG_FILE 2>&1

print_exit_msg
