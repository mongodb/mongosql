#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "running drdl connection test..."

    output=$($ARTIFACTS_DIR/bin/mongodrdl $DRDL_ARGS -d test 2>&1)
    code=$?

    set -o errexit

    output=$(echo "$output" | sed 's/\$//g')

    if [ "$code" != "$EXPECTED_STATUS" ]; then
        echo "expected connection to exit '$EXPECTED_STATUS', but it exited '$code'"
        echo "output: $output"
        exit 1
    fi

    if [ "$code" = "1" ]; then
        if [ "$output" != "$EXPECTED_ERROR" ]; then
            echo "expected error: '$EXPECTED_ERROR'"
            echo "actual error: '$output'"
            exit 1
        fi
    fi

    echo "done running drdl connection test"

) > $LOG_FILE 2>&1

print_exit_msg
