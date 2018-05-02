#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "running flush test..."

    set +o errexit
    output=$(mysql $CLIENT_ARGS -e "flush sample;" 2>&1)
    code=$?

    set -o errexit

    if [ "$code" != "$EXPECTED_STATUS" ]; then
        echo "expected connection to exit '$EXPECTED_STATUS', but it exited '$code'"
        echo "output: $output"
        exit 1
    fi

    if [ "$code" != "0" ]; then
        if [ "$output" != "$EXPECTED_ERROR" ]; then
            echo "expected code: '$EXPECTED_STATUS'"
            echo "actual code: '$code'"
            echo "expected error: '$EXPECTED_ERROR'"
            echo "actual error: '$output'"
            exit 1
        fi
    fi

    echo "done running flush test"

) > $LOG_FILE 2>&1

print_exit_msg
