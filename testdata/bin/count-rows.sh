#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running mysql query test..."

    cmd="$(echo "$QUERY" | sed 's/,,/;/g')"

    set +o errexit
    result=$(mysql --skip-column-names --silent $CLIENT_ARGS -e "$cmd" | awk 'END{print NR}' 2>&1)
    code=$?
    set -o errexit

    if [ "$code" != "0" ]; then
        if [ "$result" = "$EXPECTED_ERROR" ]; then
            echo "done running mysql query test with correct error"
            exit 0
        fi
        echo "mysql command exited with code $code"
        echo "output: $result"
        if [ "$EXPECTED_ERROR" != "" ]; then
            echo "expected error: $EXPECTED_ERROR"
        fi
        exit 1
    fi

    if [ "$result" != "$EXPECTED" ]; then
        echo "expected '$EXPECTED', got '$result'"
        exit 1
    fi

    echo "done running mysql query test"

) > $LOG_FILE 2>&1

print_exit_msg
