#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    table="${TABLE:-test1}"
    column="${COLUMN}"
    newcolumn="${NEW_COLUMN}"

    echo "renaming $table.$column to $newcolumn..."

    output=$(mysql $CLIENT_ARGS -e "use test; alter table $table change $column $newcolumn;" 2>&1)
    code=$?

    set -o errexit

    if [ "$code" != "$EXPECTED_STATUS" ]; then
        echo "expected rename command to exit '$EXPECTED_STATUS', but it exited '$code'"
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

    echo "done renaming"

) > $LOG_FILE 2>&1

print_exit_msg
