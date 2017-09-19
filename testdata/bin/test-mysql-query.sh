#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "running mysql query test..."

    result=$(mysql --skip-column-names --silent $CLIENT_ARGS -e "$QUERY" 2>&1)
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "mysql command exited with code $code"
        echo "output: $result"
        exit 1
    fi

    if [ "$result" != "$EXPECTED" ]; then
        echo "expected '$EXPECTED', got '$result'"
        exit 1
    fi

    echo "done running mysql query test"

) > $LOG_FILE 2>&1

print_exit_msg
