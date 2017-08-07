#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running logging test..."

    cmd="$(echo "$MYSQL_CMD" | sed 's/,/;/g')"
    output=$(mysql $CLIENT_ARGS -e "$cmd" 2>&1)
    code=$?

    num_files="$(ls $ARTIFACTS_DIR/log/mongosqld.log* | wc -l)"

    if [ "$code" != "0" ]; then
        echo "provided mysql command exited with code $code"
        echo "output: $output"
        exit 1
    fi

    if [ "$num_files" != "$EXPECTED_NUM_FILES" ]; then
        echo "expected $EXPECTED_NUM_FILES log files, got $num_files"
        exit 1
    fi

    sleep 1 # wait for logs to flush

    for file in $ARTIFACTS_DIR/log/mongosqld.log*; do
        lines="$(cat $file | wc -l)"
        if [ "$lines" = "0" ]; then
            echo "log file $file was empty"
            exit 1
        fi
    done

    if [ "$content" != "$EXPECTED_CONTENT" ]; then
        echo "content does not match expected content"
        echo "expected:"
        echo "$EXPECTED_CONTENT"
        echo
        echo "got:"
        echo "$content"
        echo
        echo "diff:"
        echo "$diff"
        exit 1
    fi

    echo "done running logging test"

) > $LOG_FILE 2>&1

print_exit_msg
