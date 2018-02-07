#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "running simple connection test..."

    output=$(mysql $CLIENT_ARGS -e "use information_schema;" 2>&1)
    code=$?

    set -o errexit

    # some error messages returned when we fail to connect have nondeterministic components.
    # this section filters out those nondeterministic pieces so that our config tests can
    # be written to expect the same error message every time
    output=$(echo $output | sed 's/, system error: .*$//')
    output=$(echo $output | sed 's/database '\''.*'\''$/database <dbname>/')
    output=$(echo $output | sed 's/namespace ".*"\.".*"$/namespace <colname>/')

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

    echo "done running simple connection test"

) > $LOG_FILE 2>&1

print_exit_msg
