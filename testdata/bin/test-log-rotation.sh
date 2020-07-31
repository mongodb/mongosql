#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit
    echo "running logging test..."

    cmd="$(echo "$MYSQL_CMD" | sed 's/,/;/g')"

    set +o errexit

    if [ "$ROTATION_METHOD" = "SIGUSR1" ]; then
        output=$(kill -s USR1 $(ps -A | grep -m1 mongosqld | awk '{print $1}'))
        code=$?
    else
        output=$(mysql $CLIENT_ARGS -e "$cmd" 2>&1)
        code=$?
    fi

    set -o errexit

    num_files="$(ls $ARTIFACTS_DIR/log/mongosqld.log* | wc -l | tr -d '[:space:]')"

    if [ "$code" != "0" ]; then
        echo "provided command exited with code $code"
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

    echo "done running logging test"




