#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "running schema availability test..."

    timeout="${TIMEOUT:-1}"
    iters=0

    while [ "$iters" != "$timeout" ]; do
        sleep 1
        ((iters++))

        output=$(mysql $CLIENT_ARGS -e "use information_schema;" 2>&1)
        code=$?
        output=$(echo $output | sed 's/, system error: .*$//')

        if [ "$code" = "0" ]; then
            echo "done running schema availability test"
            exit 0
        fi

        schema_not_available_err="ERROR 1043 (08S01): MongoDB schema not yet available"
        if [ "$output" != "$schema_not_available_err" ]; then
            echo "error waiting for schema: $output"
            exit 1
        fi
    done

    echo "schema not available after $timeout seconds"
    exit 1

) > $LOG_FILE 2>&1

print_exit_msg
