#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit

    echo "parsing memory test reports..."

    cd "$SCRIPT_DIR/parse"
    go test -v -run TestMemoryLimits/${TEST_NAME}

    echo "done parsing memory test reports"




