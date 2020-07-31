#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit

    echo "running schema availability test..."

    timeout="${TIMEOUT:-1}"
    iters=0

    while [ "$iters" != "$timeout" ]; do
        sleep 1
        ((iters++)) || true

        set +o errexit
        output=$(mysql $CLIENT_ARGS -e "use information_schema;" 2>&1)
        code=$?

        output=$(echo $output | sed 's/, system error: .*$//')
        set -o errexit

        if [ "$code" = "0" ]; then
            echo "done running schema availability test"
            exit 0
        fi

        if [ "$output" != "$SCHEMA_UNAVAILABLE_ERROR" ]; then
            echo "error waiting for schema: $output"
            exit 1
        fi
    done

    echo "schema not available after $timeout seconds"
    exit 1




