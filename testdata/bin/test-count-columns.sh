#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running column-count test..."

    num_columns=$(mysql --skip-column-names --silent $CLIENT_ARGS -e "select count(*) from information_schema.columns where table_name = 'column_count';" 2>&1)
    code=$?

    if [ "$code" != "0" ]; then
        echo "mysql command exited with code $code"
        echo "output: $output"
        exit 1
    fi

    if [ "$num_columns" != "$EXPECTED_NUM_COLUMNS" ]; then
        echo "expected $EXPECTED_NUM_COLUMNS columns, got $num_columns"
        exit 1
    fi

    echo "done running column-count test"

) > $LOG_FILE 2>&1

print_exit_msg
